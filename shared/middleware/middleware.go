package middleware

import (
"log/slog"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

apperr "github.com/radonezhsklad/shared/errors"
)

const HeaderRequestID = "X-Request-ID"
const CtxRequestID = "request_id"

// RequestID — читает X-Request-ID из заголовка или генерирует новый.
func RequestID() gin.HandlerFunc {
return func(c *gin.Context) {
rid := c.GetHeader(HeaderRequestID)
if rid == "" {
rid = uuid.NewString()
}
c.Set(CtxRequestID, rid)
c.Writer.Header().Set(HeaderRequestID, rid)
c.Next()
}
}

// Recovery — перехватывает паники, логирует, возвращает 500.
func Recovery() gin.HandlerFunc {
return func(c *gin.Context) {
defer func() {
if r := recover(); r != nil {
slog.Error("panic recovered",
"request_id", c.GetString(CtxRequestID),
"path", c.Request.URL.Path,
"method", c.Request.Method,
"panic", r,
)
c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
"error": gin.H{
"code":    "internal",
"message": "internal error",
},
})
}
}()
c.Next()
}
}

// AccessLog — структурированный лог запросов.
func AccessLog() gin.HandlerFunc {
return func(c *gin.Context) {
start := time.Now()
c.Next()
slog.Info("http",
"request_id", c.GetString(CtxRequestID),
"method", c.Request.Method,
"path", c.Request.URL.Path,
"status", c.Writer.Status(),
"latency_ms", time.Since(start).Milliseconds(),
"client_ip", c.ClientIP(),
)
}
}

// CORS — базовый CORS для dev и прод-домена.
func CORS(allowedOrigins []string) gin.HandlerFunc {
allow := map[string]bool{}
for _, o := range allowedOrigins {
allow[o] = true
}
return func(c *gin.Context) {
origin := c.GetHeader("Origin")
if origin != "" && (allow[origin] || allow["*"]) {
c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
c.Writer.Header().Set("Vary", "Origin")
c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
c.Writer.Header().Set("Access-Control-Max-Age", "86400")
}
if c.Request.Method == http.MethodOptions {
c.AbortWithStatus(http.StatusNoContent)
return
}
c.Next()
}
}

// ErrorHandler — превращает apperr в корректный HTTP-ответ и логирует.
func ErrorHandler() gin.HandlerFunc {
return func(c *gin.Context) {
c.Next()

if len(c.Errors) == 0 {
return
}
lastErr := c.Errors.Last().Err
appErr := apperr.AsAppError(lastErr)

if appErr.HTTPStatus >= 500 {
slog.Error("handler error",
"request_id", c.GetString(CtxRequestID),
"path", c.Request.URL.Path,
"error", appErr.Error(),
)
}

c.AbortWithStatusJSON(appErr.HTTPStatus, gin.H{
"error": gin.H{
"code":    appErr.Code,
"message": appErr.Message,
},
})
}
}