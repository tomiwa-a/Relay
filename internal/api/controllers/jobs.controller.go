package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
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

		job, err := application.Repository.CreateJob(c.Request.Context(), repository.CreateJobParams{
			ParentJobID: parentID,
			Title:       req.Title,
			Description: description,
			Payload:     req.Payload,
			MaxRetries:  maxRetries,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "job created successfully",
			"data":    job,
		})
	}
}
