package handler

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/repository"
"github.com/radonezhsklad/warehouse/internal/service"
apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
mw "github.com/radonezhsklad/shared/middleware"
)

type Handler struct{ svc *service.Service }

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

func parseUUIDOpt(s string) *uuid.UUID {
if s == "" { return nil }
id, err := uuid.Parse(s)
if err != nil { return nil }
return &id
}

func strPtr(s string) *string {
if s == "" { return nil }
return &s
}

// ---------- warehouses ----------

type warehouseReq struct {
Name     string `json:"name"     binding:"required"`
Address  string `json:"address"`
IsActive *bool  `json:"is_active"`
}

func (h *Handler) CreateWarehouse(c *gin.Context) {
var req warehouseReq
if err := c.ShouldBindJSON(&req); err != nil {
c.Error(apperr.BadRequest(err.Error())); return
}
w, err := h.svc.CreateWarehouse(c.Request.Context(), req.Name, strPtr(req.Address))
if err != nil { c.Error(err); return }
httpx.Created(c, w)
}

func (h *Handler) ListWarehouses(c *gin.Context) {
items, err := h.svc.ListWarehouses(c.Request.Context())
if err != nil { c.Error(err); return }
httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) GetWarehouse(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
w, err := h.svc.GetWarehouse(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, w)
}

func (h *Handler) UpdateWarehouse(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
var req warehouseReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
isActive := true
if req.IsActive != nil { isActive = *req.IsActive }
w, err := h.svc.UpdateWarehouse(c.Request.Context(), id, req.Name, strPtr(req.Address), isActive)
if err != nil { c.Error(err); return }
httpx.OK(c, w)
}

func (h *Handler) DeleteWarehouse(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
if err := h.svc.DeleteWarehouse(c.Request.Context(), id); err != nil { c.Error(err); return }
c.Status(http.StatusNoContent)
}

// ---------- stock ----------

func (h *Handler) ListStock(c *gin.Context) {
whID := parseUUIDOpt(c.Query("warehouse_id"))
prodID := parseUUIDOpt(c.Query("product_id"))
items, err := h.svc.ListStock(c.Request.Context(), whID, prodID)
if err != nil { c.Error(err); return }
httpx.OK(c, gin.H{"items": items})
}

// ---------- documents ----------

type docItemReq struct {
ProductID string  `json:"product_id" binding:"required"`
Quantity  float64 `json:"quantity"   binding:"required"`
Price     float64 `json:"price"`
}

type docReq struct {
Type              string       `json:"type"       binding:"required"`
Number            string       `json:"number"`
WarehouseID       string       `json:"warehouse_id" binding:"required"`
TargetWarehouseID string       `json:"target_warehouse_id"`
Comment           string       `json:"comment"`
Items             []docItemReq `json:"items"      binding:"required,min=1,dive"`
}

func (h *Handler) CreateDocument(c *gin.Context) {
var req docReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }

whID, err := uuid.Parse(req.WarehouseID)
if err != nil { c.Error(apperr.BadRequest("invalid warehouse_id")); return }

var createdBy *uuid.UUID
if uid := mw.CurrentUserID(c); uid != uuid.Nil {
createdBy = &uid
}

items := make([]repository.DocItemInput, 0, len(req.Items))
for _, it := range req.Items {
pid, err := uuid.Parse(it.ProductID)
if err != nil { c.Error(apperr.BadRequest("invalid product_id")); return }
if it.Quantity <= 0 { c.Error(apperr.BadRequest("quantity must be > 0")); return }
items = append(items, repository.DocItemInput{ProductID: pid, Quantity: it.Quantity, Price: it.Price})
}

d, err := h.svc.CreateDocument(c.Request.Context(), repository.DocumentInput{
Type:              req.Type,
Number:            req.Number,
WarehouseID:       whID,
TargetWarehouseID: parseUUIDOpt(req.TargetWarehouseID),
Comment:           strPtr(req.Comment),
CreatedBy:         createdBy,
Items:             items,
})
if err != nil { c.Error(err); return }
httpx.Created(c, d)
}

func (h *Handler) ListDocuments(c *gin.Context) {
var typeF, statusF *string
if v := c.Query("type"); v != "" { typeF = &v }
if v := c.Query("status"); v != "" { statusF = &v }
whID := parseUUIDOpt(c.Query("warehouse_id"))
items, err := h.svc.ListDocuments(c.Request.Context(), typeF, statusF, whID)
if err != nil { c.Error(err); return }
httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) GetDocument(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
d, err := h.svc.GetDocument(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, d)
}

func (h *Handler) PostDocument(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
d, err := h.svc.PostDocument(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, d)
}

func (h *Handler) CancelDocument(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
d, err := h.svc.CancelDocument(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, d)
}