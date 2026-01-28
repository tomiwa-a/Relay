package controllers

import "encoding/json"

type CreateJobRequest struct {
	ParentJobID *int32          `json:"parent_job_id"`
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description"`
	Payload     json.RawMessage `json:"payload" binding:"required"`
	MaxRetries  int32           `json:"max_retries"`
}

type JobResponse struct {
	ID          int32           `json:"id"`
	ParentJobID *int32          `json:"parent_job_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Payload     json.RawMessage `json:"payload"`
	Status      string          `json:"status"`
	Retries     int32           `json:"retries"`
	MaxRetries  int32           `json:"max_retries"`
	CreatedAt   string          `json:"created_at"`
}
