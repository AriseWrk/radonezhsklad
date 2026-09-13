package handler

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/product"
"github.com/radonezhsklad/warehouse/internal/repository"
"github.com/radonezhsklad/warehouse/internal/service"
apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
mw "github.com/radonezhsklad/shared/middleware"
)

type Handler struct {
svc           *service.Service
productClient *product.Client
}

func New(svc *service.Service, pc *product.Client) *Handler {
return &Handler{svc: svc, productClient: pc}
}

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
SupplierID        string       `json:"supplier_id"`
OrganizationID    string       `json:"organization_id"`
IncomingNumber    string       `json:"incoming_number"`
IncomingDate      string       `json:"incoming_date"`
Comment           string       `json:"comment"`
Items             []docItemReq `json:"items"      binding:"required,min=1,dive"`
}

func parseDate(s string) *time.Time {
if s == "" { return nil }
t, err := time.Parse("2006-01-02", s)
if err != nil { return nil }
return &t
}

func (h *Handler) CreateDocument(c *gin.Context) {
var req docReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }

whID, err := uuid.Parse(req.WarehouseID)
if err != nil { c.Error(apperr.BadRequest("invalid warehouse_id")); return }

var createdBy *uuid.UUID
if uid := mw.CurrentUserID(c); uid != uuid.Nil { createdBy = &uid }

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
SupplierID:        parseUUIDOpt(req.SupplierID),
OrganizationID:    parseUUIDOpt(req.OrganizationID),
IncomingNumber:    strPtr(req.IncomingNumber),
IncomingDate:      parseDate(req.IncomingDate),
Comment:           strPtr(req.Comment),
CreatedBy:         createdBy,
Items:             items,
})
if err != nil { c.Error(err); return }
httpx.Created(c, d)
}

func (h *Handler) ListDocuments(c *gin.Context) {
f := repository.DocumentFilters{}
if v := c.Query("type"); v != "" { f.Type = &v }
if v := c.Query("status"); v != "" { f.Status = &v }
f.WarehouseID = parseUUIDOpt(c.Query("warehouse_id"))
f.SupplierID = parseUUIDOpt(c.Query("supplier_id"))
items, err := h.svc.ListDocuments(c.Request.Context(), f)
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