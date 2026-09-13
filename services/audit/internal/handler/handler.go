package handler

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/audit/internal/models"
"github.com/radonezhsklad/audit/internal/repository"
apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
)

type Handler struct {
repo          *repository.Repo
internalToken string
}

func New(repo *repository.Repo, internalToken string) *Handler {
return &Handler{repo: repo, internalToken: internalToken}
}

type logReq struct {
UserID      string `json:"user_id"`
UserEmail   string `json:"user_email"`
Method      string `json:"method"`
Path        string `json:"path"`
Resource    string `json:"resource"`
ResourceID  string `json:"resource_id"`
Status      int    `json:"status"`
RequestBody string `json:"request_body"`
ClientIP    string `json:"client_ip"`
RequestID   string `json:"request_id"`
}

func (h *Handler) InternalLog(c *gin.Context) {
if c.GetHeader("X-Internal-Token") != h.internalToken {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid internal token"}); return
}

var req logReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
}

l := &models.AuditLog{
Method: req.Method,
Path:   req.Path,
Status: req.Status,
}
if req.UserID != "" {
if id, err := uuid.Parse(req.UserID); err == nil { l.UserID = &id }
}
if req.UserEmail != "" { l.UserEmail = &req.UserEmail }
if req.Resource != "" { l.Resource = &req.Resource }
if req.ResourceID != "" { l.ResourceID = &req.ResourceID }
if req.RequestBody != "" { l.RequestBody = &req.RequestBody }
if req.ClientIP != "" { l.ClientIP = &req.ClientIP }
if req.RequestID != "" { l.RequestID = &req.RequestID }

if err := h.repo.Create(c.Request.Context(), l); err != nil {
c.Error(apperr.Internal("create audit", err)); return
}
c.JSON(http.StatusCreated, l)
}

func (h *Handler) List(c *gin.Context) {
f := repository.ListFilters{ Limit: 100, Offset: 0 }
if s := c.Query("user_id"); s != "" {
if id, err := uuid.Parse(s); err == nil { f.UserID = &id }
}
if s := c.Query("method"); s != "" { f.Method = &s }
if s := c.Query("resource"); s != "" { f.Resource = &s }
if s := c.Query("date_from"); s != "" { f.DateFrom = &s }
if s := c.Query("date_to"); s != "" { f.DateTo = &s }
if s := c.Query("limit"); s != "" {
if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 500 { f.Limit = n }
}
if s := c.Query("offset"); s != "" {
if n, err := strconv.Atoi(s); err == nil && n >= 0 { f.Offset = n }
}

items, total, err := h.repo.List(c.Request.Context(), f)
if err != nil { c.Error(apperr.Internal("list audit", err)); return }
httpx.OK(c, gin.H{"items": items, "total": total, "limit": f.Limit, "offset": f.Offset})
}