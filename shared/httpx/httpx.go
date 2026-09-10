package httpx

import (
"net/http"

"github.com/gin-gonic/gin"

apperr "github.com/radonezhsklad/shared/errors"
)

// OK — 200 с телом.
func OK(c *gin.Context, data any) {
c.JSON(http.StatusOK, data)
}

// Created — 201.
func Created(c *gin.Context, data any) {
c.JSON(http.StatusCreated, data)
}

// NoContent — 204.
func NoContent(c *gin.Context) {
c.Status(http.StatusNoContent)
}

// Error — единый формат ошибки. Если err — *AppError, используем его статус и код.
func Error(c *gin.Context, err error) {
appErr := apperr.AsAppError(err)
c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
"error": gin.H{
"code":    appErr.Code,
"message": appErr.Message,
},
})
}

// Validation — 400 с перечнем ошибок валидации.
func Validation(c *gin.Context, details any) {
c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
"error": gin.H{
"code":    "validation_error",
"message": "request validation failed",
"details": details,
},
})
}