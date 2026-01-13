package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tomiwa-a/Relay/internal/api/app"
)

func GetAllJobs(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "jobs fetched successfully",
		})
	}
}
