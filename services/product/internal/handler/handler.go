package handler

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/product/internal/repository"
"github.com/radonezhsklad/product/internal/service"
"github.com/radonezhsklad/shared/httpx"
)

type Handler struct {
svc *service.Service
}

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

func parseUUID(s string) *uuid.UUID {
if s == "" {
return nil
}
id, err := uuid.Parse(s)
if err != nil {
return nil
}
return &id
}

func strPtr(s string) *string {
if s == "" {
return nil
}
return &s
}

// ---------- categories ----------

type categoryReq struct {
Name     string `json:"name"      binding:"required"`
ParentID string `json:"parent_id"`
}

func (h *Handler) CreateCategory(c *gin.Context) {
var req categoryReq
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(errBadRequest(err.Error()))
return
}
cat, err := h.svc.CreateCategory(c.Request.Context(), req.Name, parseUUID(req.ParentID))
if err != nil {
c.Error(err)
return
}
httpx.Created(c, cat)
}

func (h *Handler) ListCategories(c *gin.Context) {
cats, err := h.svc.ListCategories(c.Request.Context())
if err != nil {
c.Error(err)
return
}
httpx.OK(c, gin.H{"items": cats})
}

func (h *Handler) GetCategory(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(errBadRequest("invalid id"))
return
}
cat, err := h.svc.GetCategory(c.Request.Context(), id)
if err != nil {
c.Error(err)
return
}
httpx.OK(c, cat)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(errBadRequest("invalid id"))
return
}
var req categoryReq
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(errBadRequest(err.Error()))
return
}
cat, err := h.svc.UpdateCategory(c.Request.Context(), id, req.Name, parseUUID(req.ParentID))
if err != nil {
c.Error(err)
return
}
httpx.OK(c, cat)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(errBadRequest("invalid id"))
return
}
if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
c.Error(err)
return
}
c.Status(http.StatusNoContent)
}

// ---------- units ----------

func (h *Handler) ListUnits(c *gin.Context) {
units, err := h.svc.ListUnits(c.Request.Context())
if err != nil {
c.Error(err)
return
}
httpx.OK(c, gin.H{"items": units})
}

// ---------- products ----------

type productReq struct {
Name        string  `json:"name"     binding:"required"`
SKU         string  `json:"sku"`
Barcode     string  `json:"barcode"`
CategoryID  string  `json:"category_id"`
UnitID      string  `json:"unit_id"`
Description string  `json:"description"`
Price       float64 `json:"price"`
Currency    string  `json:"currency"`
}

func (r productReq) toInput() repository.ProductInput {
return repository.ProductInput{
Name:        r.Name,
SKU:         strPtr(r.SKU),
Barcode:     strPtr(r.Barcode),
CategoryID:  parseUUID(r.CategoryID),
UnitID:      parseUUID(r.UnitID),
Description: strPtr(r.Description),
Price:       r.Price,
Currency:    r.Currency,
}
}

func (h *Handler) CreateProduct(c *gin.Context) {
var req productReq
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(errBadRequest(err.Error()))
return
}
p, err := h.svc.CreateProduct(c.Request.Context(), req.toInput())
if err != nil {
c.Error(err)
return
}
httpx.Created(c, p)
}

func (h *Handler) ListProducts(c *gin.Context) {
includeArchived := c.Query("include_archived") == "true"
catID := parseUUID(c.Query("category_id"))
items, err := h.svc.ListProducts(c.Request.Context(), includeArchived, catID)
if err != nil {
c.Error(err)
return
}
httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) GetProduct(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(errBadRequest("invalid id"))
return
}
p, err := h.svc.GetProduct(c.Request.Context(), id)
if err != nil {
c.Error(err)
return
}
httpx.OK(c, p)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(errBadRequest("invalid id"))
return
}
var req productReq
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(errBadRequest(err.Error()))
return
}
p, err := h.svc.UpdateProduct(c.Request.Context(), id, req.toInput())
if err != nil {
c.Error(err)
return
}
httpx.OK(c, p)
}

func (h *Handler) ArchiveProduct(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil {
c.Error(errBadRequest("invalid id"))
return
}
if err := h.svc.ArchiveProduct(c.Request.Context(), id); err != nil {
c.Error(err)
return
}
c.Status(http.StatusNoContent)
}