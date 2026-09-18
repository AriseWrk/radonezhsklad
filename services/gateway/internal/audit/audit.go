package audit

import (
"bytes"
"encoding/json"
"io"
"log/slog"
"net/http"
"strings"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

mw "github.com/radonezhsklad/shared/middleware"
)

type Logger struct {
baseURL       string
internalToken string
client        *http.Client
}

func NewLogger(baseURL, internalToken string) *Logger {
return &Logger{
baseURL:       baseURL,
internalToken: internalToken,
client:        &http.Client{Timeout: 5 * time.Second},
}
}

type event struct {
UserID      string `json:"user_id,omitempty"`
UserEmail   string `json:"user_email,omitempty"`
Method      string `json:"method"`
Path        string `json:"path"`
Resource    string `json:"resource,omitempty"`
ResourceID  string `json:"resource_id,omitempty"`
Status      int    `json:"status"`
RequestBody string `json:"request_body,omitempty"`
ClientIP    string `json:"client_ip,omitempty"`
RequestID   string `json:"request_id,omitempty"`
}

// Middleware пишет в audit все POST/PUT/DELETE/PATCH с кодом 2xx.
func (l *Logger) Middleware() gin.HandlerFunc {
return func(c *gin.Context) {
method := c.Request.Method

// читаем тело только для мутаций
var bodyCopy []byte
if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
if c.Request.Body != nil {
bodyCopy, _ = io.ReadAll(c.Request.Body)
c.Request.Body = io.NopCloser(bytes.NewReader(bodyCopy))
}
}

c.Next()

if bodyCopy == nil {
return
}
if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
return
}

// кто
var uidStr, email string
if uid := mw.CurrentUserID(c); uid != uuid.Nil {
uidStr = uid.String()
}
// email берём из заголовка X-User-Email (у нас его нет, оставим пустым)

resource, resourceID := parseResource(c.Request.URL.Path)

// не логируем сами audit-запросы
if strings.HasPrefix(c.Request.URL.Path, "/api/v1/audit") {
return
}

ev := event{
UserID:      uidStr,
UserEmail:   email,
Method:      method,
Path:        c.Request.URL.Path,
Resource:    resource,
ResourceID:  resourceID,
Status:      c.Writer.Status(),
RequestBody: truncate(string(sanitizeBody(bodyCopy)), 4000),
ClientIP:    c.ClientIP(),
RequestID:   c.GetString(mw.CtxRequestID),
}

// асинхронно
go l.send(ev)
}
}

func (l *Logger) send(ev event) {
payload, err := json.Marshal(ev)
if err != nil {
return
}

req, err := http.NewRequest(http.MethodPost, l.baseURL+"/api/v1/internal/audit", bytes.NewReader(payload))
if err != nil {
return
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-Internal-Token", l.internalToken)

resp, err := l.client.Do(req)
if err != nil {
slog.Warn("audit send failed", "error", err)
return
}
defer resp.Body.Close()
if resp.StatusCode >= 300 {
slog.Warn("audit service rejected", "status", resp.StatusCode)
}
}

// parseResource: /api/v1/products/abc-123 → ("products", "abc-123")
func parseResource(path string) (string, string) {
parts := strings.Split(strings.Trim(path, "/"), "/")
// api, v1, products, abc-123
if len(parts) < 3 {
return "", ""
}
resource := parts[2]
id := ""
if len(parts) >= 4 {
id = parts[3]
// если id — не UUID, оставляем как есть
}
return resource, id
}

func truncate(s string, n int) string {
if len(s) <= n {
return s
}
return s[:n] + "...[truncated]"
}

// sensitiveKeys — поля, значения которых нельзя логировать в открытом виде.
var sensitiveKeys = map[string]bool{
"password":      true,
"password_hash": true,
"token":         true,
"refresh_token": true,
"access_token":  true,
"secret":        true,
}

// sanitizeBody рекурсивно маскирует значения чувствительных полей в JSON.
// Если тело не является JSON — возвращает его как есть.
func sanitizeBody(raw []byte) []byte {
if len(raw) == 0 {
return raw
}
var obj any
if err := json.Unmarshal(raw, &obj); err != nil {
return raw
}
sanitizeValue(obj)
out, err := json.Marshal(obj)
if err != nil {
return raw
}
return out
}

func sanitizeValue(v any) {
switch t := v.(type) {
case map[string]any:
for k, val := range t {
if sensitiveKeys[strings.ToLower(k)] {
t[k] = "***"
continue
}
sanitizeValue(val)
}
case []any:
for _, item := range t {
sanitizeValue(item)
}
}
}