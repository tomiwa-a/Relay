package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/api/customerrors"
)

func HandleRequests(r *gin.Engine, app *app.Application) {

	r.NoRoute(customerrors.NotFoundResponse)
	r.NoMethod(customerrors.MethodNotAllowedResponse)

	RegisterJobRoutes(r, app)

}
