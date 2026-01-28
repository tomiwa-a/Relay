package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/kafka-go"
	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/repository"
)

func GetAllJobs(application *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs, err := application.Repository.ListJobs(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch jobs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "jobs fetched successfully",
			"data":    jobs,
		})
	}
}

func GetSingleJob(application *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		{
			jobID := c.Param("id")
			jobIDInt, err := strconv.Atoi(jobID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
				return
			}
			job, err := application.Repository.GetJob(c.Request.Context(), int32(jobIDInt))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch job"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "job fetched successfully",
				"data":    job,
			})
		}
	}
}

func AddJob(application *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateJobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		parentID := pgtype.Int4{}
		if req.ParentJobID != nil {
			parentID = pgtype.Int4{Int32: *req.ParentJobID, Valid: true}
		}

		description := pgtype.Text{}
		if req.Description != "" {
			description = pgtype.Text{String: req.Description, Valid: true}
		}

		maxRetries := pgtype.Int4{Int32: 3, Valid: true}
		if req.MaxRetries > 0 {
			maxRetries = pgtype.Int4{Int32: req.MaxRetries, Valid: true}
		}

		timeoutSeconds := pgtype.Int4{Int32: 30, Valid: true} // Default 30 seconds
		if req.TimeoutSeconds > 0 {
			timeoutSeconds = pgtype.Int4{Int32: req.TimeoutSeconds, Valid: true}
		}

		job, err := application.Repository.CreateJob(c.Request.Context(), repository.CreateJobParams{
			ParentJobID:    parentID,
			Title:          req.Title,
			Description:    description,
			Payload:        req.Payload,
			MaxRetries:     maxRetries,
			TimeoutSeconds: timeoutSeconds,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
			return
		}

		msg := kafka.Message{
			Key:   []byte(strconv.Itoa(int(job.ID))),
			Value: []byte(strconv.Itoa(int(job.ID))),
		}

		err = application.KafkaWriter.WriteMessages(c.Request.Context(), msg)
		if err != nil {
			application.Logger.Printf("failed to push job [%d] to kafka: %v", job.ID, err)
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "job created successfully",
			"data":    job,
		})
	}
}
