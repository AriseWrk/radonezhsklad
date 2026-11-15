package errors

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
)

// Middleware converts pending c.Errors into a JSON error response
// when the handler has not already written one. Handlers that use
// c.Error(apperr.X(...)) + return rely on this to produce a proper
// HTTP status + body; without it gin answers 200 with an empty body.
func Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) == 0 || c.Writer.Written() {
            return
        }

        appErr := AsAppError(c.Errors.Last().Err)
        if appErr.HTTPStatus >= http.StatusInternalServerError {
            log.Printf("[error] %s %s: %s (cause: %v)", c.Request.Method, c.Request.URL.Path, appErr.Message, appErr.Err)
        }
        c.JSON(appErr.HTTPStatus, gin.H{
            "error": gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
            },
        })
    }
}
