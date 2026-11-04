package handler

import (
"io"
"net/http"
"os"
"path/filepath"
"strings"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
mw "github.com/radonezhsklad/shared/middleware"
"github.com/radonezhsklad/warehouse/internal/repository"
)

type FilesHandler struct {
repo *repository.FileRepo
}

func NewFilesHandler(repo *repository.FileRepo) *FilesHandler {
return &FilesHandler{repo: repo}
}

func (h *FilesHandler) List(c *gin.Context) {
items, err := h.repo.List()
if err != nil {
c.Error(apperr.Internal("list files", err))
return
}
httpx.OK(c, gin.H{"items": items})
}

func (h *FilesHandler) Upload(c *gin.Context) {
file, err := c.FormFile("file")
if err != nil {
c.Error(apperr.BadRequest("file is required"))
return
}
category := c.PostForm("category")
if category == "" {
category = "Другое"
}

f, err := file.Open()
if err != nil {
c.Error(apperr.Internal("open upload", err))
return
}
defer f.Close()
content, err := io.ReadAll(f)
if err != nil {
c.Error(apperr.Internal("read upload", err))
return
}

name := filepath.Base(file.Filename)
name = strings.ReplaceAll(name, " ", "_")

var uid *uuid.UUID
if id := mw.CurrentUserID(c); id != uuid.Nil {
uid = &id
}
userName := c.GetString("user_name")

m, err := h.repo.Save(name, category, content, uid, userName)
if err != nil {
c.Error(apperr.BadRequest(err.Error()))
return
}
httpx.Created(c, m)
}

func (h *FilesHandler) Download(c *gin.Context) {
name := c.Param("name")
path, err := h.repo.Get(name)
if err != nil {
if os.IsNotExist(err) {
c.Error(apperr.NotFound("file not found"))
return
}
c.Error(apperr.BadRequest(err.Error()))
return
}
c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
c.File(path)
}

func (h *FilesHandler) Delete(c *gin.Context) {
name := c.Param("name")
if err := h.repo.Delete(name); err != nil {
c.Error(apperr.BadRequest(err.Error()))
return
}
c.Status(http.StatusNoContent)
}
func (h *FilesHandler) UpdateCategory(c *gin.Context) {
name := c.Param("name")
var req struct {
Category string `json:"category"`
}
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(apperr.BadRequest("invalid json"))
return
}
if req.Category == "" {
c.Error(apperr.BadRequest("category is required"))
return
}
if err := h.repo.UpdateCategory(name, req.Category); err != nil {
c.Error(apperr.BadRequest(err.Error()))
return
}
c.Status(http.StatusNoContent)
}