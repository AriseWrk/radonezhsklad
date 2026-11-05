package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperr "github.com/radonezhsklad/shared/errors"
	"github.com/radonezhsklad/shared/httpx"
	"github.com/radonezhsklad/warehouse/internal/service"
)

type SupplierHandler struct{ svc *service.SupplierService }

func NewSupplierHandler(svc *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{svc: svc}
}

type supplierReq struct {
	Name    string `json:"name"    binding:"required"`
	INN     string `json:"inn"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

func (h *SupplierHandler) Create(c *gin.Context) {
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest(err.Error()))
		return
	}
	s, err := h.svc.Create(c.Request.Context(), service.SupplierInput{
		Name: req.Name, INN: strPtr(req.INN), Phone: strPtr(req.Phone),
		Email: strPtr(req.Email), Address: strPtr(req.Address),
	})
	if err != nil {
		c.Error(err)
		return
	}
	httpx.Created(c, s)
}

func (h *SupplierHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	httpx.OK(c, gin.H{"items": items})
}

func (h *SupplierHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperr.BadRequest("invalid id"))
		return
	}
	s, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	httpx.OK(c, s)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperr.BadRequest("invalid id"))
		return
	}
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperr.BadRequest(err.Error()))
		return
	}
	s, err := h.svc.Update(c.Request.Context(), id, service.SupplierInput{
		Name: req.Name, INN: strPtr(req.INN), Phone: strPtr(req.Phone),
		Email: strPtr(req.Email), Address: strPtr(req.Address),
	})
	if err != nil {
		c.Error(err)
		return
	}
	httpx.OK(c, s)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
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

type organizationReq struct {
Name         string `json:"name" binding:"required"`
FullName     string `json:"full_name"`
INN          string `json:"inn"`
KPP          string `json:"kpp"`
OGRN         string `json:"ogrn"`
OKPO         string `json:"okpo"`
LegalAddress string `json:"legal_address"`
IsDefault    bool   `json:"is_default"`
}

func (h *SupplierHandler) CreateOrganization(c *gin.Context) {
var req organizationReq
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(apperr.BadRequest(err.Error()))
return
}
org, err := h.svc.CreateOrganization(c.Request.Context(), service.OrganizationInput{
Name:         req.Name,
FullName:     strPtr(req.FullName),
INN:          strPtr(req.INN),
KPP:          strPtr(req.KPP),
OGRN:         strPtr(req.OGRN),
OKPO:         strPtr(req.OKPO),
LegalAddress: strPtr(req.LegalAddress),
IsDefault:    req.IsDefault,
})
if err != nil {
c.Error(err)
return
}
httpx.Created(c, org)
}
func (h *SupplierHandler) UpdateOrganization(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(apperr.BadRequest("invalid id"))
return
}
var req organizationReq
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(apperr.BadRequest(err.Error()))
return
}
org, err := h.svc.UpdateOrganization(c.Request.Context(), id, service.OrganizationInput{
Name:         req.Name,
FullName:     strPtr(req.FullName),
INN:          strPtr(req.INN),
KPP:          strPtr(req.KPP),
OGRN:         strPtr(req.OGRN),
OKPO:         strPtr(req.OKPO),
LegalAddress: strPtr(req.LegalAddress),
IsDefault:    req.IsDefault,
})
if err != nil {
c.Error(err)
return
}
httpx.OK(c, org)
}

func (h *SupplierHandler) DeleteOrganization(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(apperr.BadRequest("invalid id"))
return
}
if err := h.svc.DeleteOrganization(c.Request.Context(), id); err != nil {
c.Error(err)
return
}
c.Status(http.StatusNoContent)
}
func (h *SupplierHandler) ListOrganizations(c *gin.Context) {
	items, err := h.svc.ListOrganizations(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	httpx.OK(c, gin.H{"items": items})
}
