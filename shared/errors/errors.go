package errors

import (
"errors"
"fmt"
"net/http"
)

// AppError — типизированная ошибка с HTTP-статусом и кодом.
type AppError struct {
HTTPStatus int    `json:"-"`
Code       string `json:"code"`
Message    string `json:"message"`
Err        error  `json:"-"`
}

func (e *AppError) Error() string {
if e.Err != nil {
return fmt.Sprintf("%s: %v", e.Message, e.Err)
}
return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// Конструкторы

func NotFound(msg string) *AppError {
return &AppError{HTTPStatus: http.StatusNotFound, Code: "not_found", Message: msg}
}

func Conflict(msg string) *AppError {
return &AppError{HTTPStatus: http.StatusConflict, Code: "conflict", Message: msg}
}

func BadRequest(msg string) *AppError {
return &AppError{HTTPStatus: http.StatusBadRequest, Code: "bad_request", Message: msg}
}

func Unauthorized(msg string) *AppError {
return &AppError{HTTPStatus: http.StatusUnauthorized, Code: "unauthorized", Message: msg}
}

func Forbidden(msg string) *AppError {
return &AppError{HTTPStatus: http.StatusForbidden, Code: "forbidden", Message: msg}
}

func Internal(msg string, err error) *AppError {
return &AppError{HTTPStatus: http.StatusInternalServerError, Code: "internal", Message: msg, Err: err}
}

// AsAppError извлекает *AppError из цепочки ошибок или оборачивает в Internal.
func AsAppError(err error) *AppError {
var appErr *AppError
if errors.As(err, &appErr) {
return appErr
}
return Internal("internal error", err)
}