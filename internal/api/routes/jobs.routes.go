package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/api/controllers"
)

func RegisterJobRoutes(r *gin.Engine, app *app.Application) {

	jobs := r.Group("jobs")

	jobs.GET("", controllers.GetAllJobs(app))
}
