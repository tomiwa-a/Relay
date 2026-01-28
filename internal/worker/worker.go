package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/repository"
)

type Worker struct {
	app *app.Application
}

func NewWorker(app *app.Application) *Worker {
	return &Worker{app: app}
}

func (w *Worker) Start(ctx context.Context) {
	w.app.Logger.Println("starting background worker...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.app.Logger.Println("stopping background worker...")
			return
		case <-ticker.C:
			w.processPendingJobs(ctx)
		}
	}
}

func (w *Worker) processPendingJobs(ctx context.Context) {
	jobs, err := w.app.Repository.GetPendingJobs(ctx)
	if err != nil {
		w.app.Logger.Printf("error fetching pending jobs: %v", err)
		return
	}

	for _, job := range jobs {
		w.app.Logger.Printf("processing job [%d]: %s", job.ID, job.Title)

		// Transition to in_progress
		_, err = w.app.Repository.UpdateJobStatus(ctx, repository.UpdateJobStatusParams{
			ID:     job.ID,
			Status: repository.NullJobStatus{JobStatus: repository.JobStatusInProgress, Valid: true},
		})
		if err != nil {
			w.app.Logger.Printf("error updating job [%d] to in_progress: %v", job.ID, err)
			continue
		}

		// Simulate work
		fmt.Printf("PAYLOAD for Job [%d]: %s\n", job.ID, string(job.Payload))
		time.Sleep(2 * time.Second)

		// Transition to completed
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
}
