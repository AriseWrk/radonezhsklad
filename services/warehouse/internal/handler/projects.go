package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperr "github.com/radonezhsklad/shared/errors"
	"github.com/radonezhsklad/shared/httpx"
	"github.com/radonezhsklad/warehouse/internal/service"
)

type ProjectHandler struct{ svc *service.ProjectService }

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type projectReq struct {
	Name     string `json:"name" binding:"required"`
	Archived bool   `json:"archived"`
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req projectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest(err.Error()))
		return
	}
	p, err := h.svc.Create(c.Request.Context(), service.ProjectInput{Name: req.Name})
	if err != nil {
		c.Error(err)
		return
	}
	httpx.Created(c, p)
}

func (h *ProjectHandler) List(c *gin.Context) {
	includeArchived := c.Query("include_archived") == "true"
	items, err := h.svc.List(c.Request.Context(), includeArchived)
	if err != nil {
		c.Error(err)
		return
	}
	httpx.OK(c, gin.H{"items": items})
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperr.BadRequest("invalid id"))
		return
	}
	p, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	httpx.OK(c, p)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperr.BadRequest("invalid id"))
		return
	}
	var req projectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest(err.Error()))
		return
	}
	p, err := h.svc.Update(c.Request.Context(), id, service.ProjectInput{Name: req.Name, Archived: req.Archived})
	if err != nil {
		c.Error(err)
		return
	}
	httpx.OK(c, p)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperr.BadRequest("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
