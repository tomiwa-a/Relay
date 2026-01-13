package customerrors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tomiwa-a/Relay/internal/api/app"
	"github.com/tomiwa-a/Relay/internal/api/utils"
)

func LogError(app *app.Application, c *gin.Context, err error) {
	app.Logger.Println(err)
}

func ErrorResponse(c *gin.Context, status int, message interface{}) {
	env := utils.Envelope{"error": message}
	c.JSON(status, env)
}

func BadRequestResponse(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusBadRequest, err.Error())
}

func ServerErrorResponse(app *app.Application, c *gin.Context, err error) {
	LogError(app, c, err)
	message := "the server encountered a problem and could not process your request"
	ErrorResponse(c, http.StatusInternalServerError, message)
}

func EditConflictResponse(c *gin.Context) {
	message := "unable to update the record due to an edit conflict, please try again"
	ErrorResponse(c, http.StatusInternalServerError, message)
}

func NotFoundResponse(c *gin.Context) {
	message := "the requested resource could not be found"
	ErrorResponse(c, http.StatusNotFound, message)
}

func MethodNotAllowedResponse(c *gin.Context) {
	message := fmt.Sprintf("the %s method is not supported for this resource", c.Request.Method)
	ErrorResponse(c, http.StatusMethodNotAllowed, message)
}

func FailedValidationResponse(c *gin.Context, errors map[string]string) {
	for key, value := range errors {
		ErrorResponse(c, http.StatusUnprocessableEntity, fmt.Sprintf("%s %s", key, value))
		return
	}
}

func RateLimitExceededResponse(c *gin.Context) {
	message := "rate limit exceeded"
	ErrorResponse(c, http.StatusTooManyRequests, message)
}

func InvalidCredentialsResponse(c *gin.Context) {
	message := "invalid authentication credentials"
	ErrorResponse(c, http.StatusUnauthorized, message)
}

func InvalidAuthenticationTokenResponse(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	ErrorResponse(c, http.StatusUnauthorized, message)
}

func AuthenticationRequiredResponse(c *gin.Context) {
	message := "you must be authenticated to access this resource"
	ErrorResponse(c, http.StatusUnauthorized, message)
}

func InactiveAccountResponse(c *gin.Context) {
	message := "your user account must be activated to access this resource"
	ErrorResponse(c, http.StatusForbidden, message)
}

func NotPermittedResponse(c *gin.Context) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	ErrorResponse(c, http.StatusForbidden, message)
}
