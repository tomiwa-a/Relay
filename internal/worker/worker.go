package worker

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	job, err := w.app.Repository.GetJob(ctx, jobID)
	if err != nil {
		w.app.Logger.Printf("error fetching job [%d]: %v", jobID, err)
		return
	}

	if job.Status.JobStatus != repository.JobStatusPending {
		w.app.Logger.Printf("job [%d] is not pending, skipping (status: %s)", jobID, job.Status.JobStatus)
		return
	}

	w.app.Logger.Printf("processing job [%d]: %s", job.ID, job.Title)

	_, err = w.app.Repository.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
		ID:     job.ID,
		Status: repository.NullJobStatus{JobStatus: repository.JobStatusInProgress, Valid: true},
	})
	if err != nil {
		w.app.Logger.Printf("error updating job [%d] to in_progress: %v", job.ID, err)
		return
	}

	fmt.Printf("PAYLOAD for Job [%d]: %s\n", job.ID, string(job.Payload))
	time.Sleep(2 * time.Second)

	_, err = w.app.Repository.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
		ID:     job.ID,
		Status: repository.NullJobStatus{JobStatus: repository.JobStatusCompleted, Valid: true},
	})
	if err != nil {
		w.app.Logger.Printf("error updating job [%d] to completed: %v", job.ID, err)
	} else {
		w.app.Logger.Printf("job [%d] completed successfully", job.ID)
	}
}
