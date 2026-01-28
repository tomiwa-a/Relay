package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/kafka-go"
	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/repository"
)

type Worker struct {
	app         *app.Application
	kafkaReader *kafka.Reader
}

func NewWorker(app *app.Application, reader *kafka.Reader) *Worker {
	return &Worker{
		app:         app,
		kafkaReader: reader,
	}
}

func (w *Worker) logJob(ctx context.Context, jobID int32, level repository.LogLevel, message string) {
	w.app.Logger.Printf("[Job %d] [%s] %s", jobID, level, message)

	_, _ = w.app.Repository.CreateJobLog(ctx, repository.CreateJobLogParams{
		JobID:    jobID,
		Level:    level,
		Message:  message,
		Stdout:   pgtype.Text{Valid: false},
		Stderr:   pgtype.Text{Valid: false},
		ExitCode: pgtype.Int4{Valid: false},
	})
}

func (w *Worker) Start(ctx context.Context) {
	w.app.Logger.Println("starting background worker...")

	for {
		m, err := w.kafkaReader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.app.Logger.Printf("error reading message from kafka: %v", err)
			continue
		}

		jobID, err := strconv.Atoi(string(m.Value))
		if err != nil {
			w.app.Logger.Printf("invalid job ID in kafka message: %s", string(m.Value))
			continue
		}

		w.processJob(ctx, int32(jobID))
	}
}

func (w *Worker) processJob(ctx context.Context, jobID int32) {
	lockKey := fmt.Sprintf("job:lock:%d", jobID)
	val, err := w.app.Redis.SetNX(ctx, lockKey, "locked", w.app.Config.Redis.LockTTL).Result()
	if err != nil {
		w.app.Logger.Printf("error acquiring lock for job [%d]: %v", jobID, err)
		return
	}

	if !val {
		w.app.Logger.Printf("job [%d] is already being processed by another worker, skipping", jobID)
		return
	}

	// Watchdog logic
	if w.app.Config.Redis.UseWatchdog {
		watchdogCtx, stopWatchdog := context.WithCancel(ctx)
		defer stopWatchdog()

		go func() {
			ticker := time.NewTicker(w.app.Config.Redis.LockTTL / 2)
			defer ticker.Stop()

			for {
				select {
				case <-watchdogCtx.Done():
					w.app.Redis.Del(ctx, lockKey)
					return
				case <-ticker.C:
					w.app.Redis.Expire(ctx, lockKey, w.app.Config.Redis.LockTTL)
				}
			}
		}()
	} else {
		defer w.app.Redis.Del(ctx, lockKey)
	}

	job, err := w.app.Repository.GetJob(ctx, jobID)
	if err != nil {
		w.app.Logger.Printf("error fetching job [%d]: %v", jobID, err)
		return
	}

	if job.Status.JobStatus != repository.JobStatusPending {
		w.app.Logger.Printf("job [%d] is not pending, skipping (status: %s)", jobID, job.Status.JobStatus)
		return
	}

	w.logJob(ctx, job.ID, repository.LogLevelINFO, fmt.Sprintf("processing job: %s", job.Title))

	_, err = w.app.Repository.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
		ID:      job.ID,
		Status:  repository.NullJobStatus{JobStatus: repository.JobStatusInProgress, Valid: true},
		Retries: job.Retries,
	})
	if err != nil {
		w.app.Logger.Printf("error updating job [%d] to in_progress: %v", job.ID, err)
		return
	}

	// Job Execution with Timeout
	execTimeout := 30 * time.Second
	if job.TimeoutSeconds.Valid && job.TimeoutSeconds.Int32 > 0 {
		execTimeout = time.Duration(job.TimeoutSeconds.Int32) * time.Second
	}

	execCtx, cancelExec := context.WithTimeout(ctx, execTimeout)
	defer cancelExec()

	// Channel to capture execution result
	done := make(chan error, 1)

	go func() {
		var payload map[string]interface{}
		_ = json.Unmarshal(job.Payload, &payload)

		// Simulate failure if "fail": true is in payload
		if fail, ok := payload["fail"].(bool); ok && fail {
			done <- errors.New("simulated job failure")
			return
		}

		// Simulate long running job if "sleep": N is in payload
		sleepTime := 2 * time.Second
		if s, ok := payload["sleep"].(float64); ok {
			sleepTime = time.Duration(s) * time.Second
		}

		time.Sleep(sleepTime)
		done <- nil
	}()

	select {
	case <-execCtx.Done():
		err = execCtx.Err()
		if err == context.DeadlineExceeded {
			err = fmt.Errorf("job execution timed out after %v", execTimeout)
		}
	case execErr := <-done:
		err = execErr
	}

	if err != nil {
		w.handleFailure(ctx, job, err)
		return
	}

	_, err = w.app.Repository.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
		ID:      job.ID,
		Status:  repository.NullJobStatus{JobStatus: repository.JobStatusCompleted, Valid: true},
		Retries: job.Retries,
	})
	if err != nil {
		w.app.Logger.Printf("error updating job [%d] to completed: %v", job.ID, err)
	} else {
		w.logJob(ctx, job.ID, repository.LogLevelINFO, "job completed successfully")
	}
}

func (w *Worker) handleFailure(ctx context.Context, job repository.Job, execErr error) {
	w.app.Logger.Printf("job [%d] failed: %v", job.ID, execErr)

	if job.Retries.Int32 < job.MaxRetries.Int32 {
		nextRetry := job.Retries.Int32 + 1
		backoff := time.Duration(math.Pow(2, float64(nextRetry))) * time.Second

		w.app.Logger.Printf("retrying job [%d] in %v (attempt %d/%d)", job.ID, backoff, nextRetry, job.MaxRetries.Int32)

		_, err := w.app.Repository.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
			ID:      job.ID,
			Status:  repository.NullJobStatus{JobStatus: repository.JobStatusPending, Valid: true},
			Retries: pgtype.Int4{Int32: nextRetry, Valid: true},
		})
		if err != nil {
			w.app.Logger.Printf("error updating job [%d] for retry: %v", job.ID, err)
			return
		}

		// Wait for backoff then re-push to Kafka to trigger again
		go func() {
			time.Sleep(backoff)
			msg := kafka.Message{
				Key:   []byte(strconv.Itoa(int(job.ID))),
				Value: []byte(strconv.Itoa(int(job.ID))),
			}
			if err := w.app.KafkaWriter.WriteMessages(ctx, msg); err != nil {
				w.app.Logger.Printf("failed to re-push job [%d] to kafka: %v", job.ID, err)
			}
		}()
	} else {
		w.logJob(ctx, job.ID, repository.LogLevelERROR, fmt.Sprintf("job has reached max retries (%d), marking as dead", job.MaxRetries.Int32))
		_, _ = w.app.Repository.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
			ID:      job.ID,
			Status:  repository.NullJobStatus{JobStatus: repository.JobStatusDead, Valid: true}, // DLQ: Marked as dead
			Retries: job.Retries,
		})
	}
}
