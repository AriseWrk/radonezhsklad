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

type InternalOrderHandler struct {
svc          *service.InternalOrderService
warehouseSvc *service.Service
supplierSvc  *service.SupplierService
productCli   *product.Client
}

func NewInternalOrderHandler(
svc *service.InternalOrderService,
warehouseSvc *service.Service,
supplierSvc *service.SupplierService,
productCli *product.Client,
) *InternalOrderHandler {
return &InternalOrderHandler{
svc:          svc,
warehouseSvc: warehouseSvc,
supplierSvc:  supplierSvc,
productCli:   productCli,
}
}

type intOrderItemReq struct {
ProductID string  `json:"product_id" binding:"required"`
Quantity  float64 `json:"quantity"   binding:"required"`
Price     float64 `json:"price"`
VatRate   float64 `json:"vat_rate"`
}

type intOrderReq struct {
Number         string            `json:"number"`
OrganizationID string            `json:"organization_id"`
WarehouseID    string            `json:"warehouse_id"`
PlanDate       string            `json:"plan_date"`
Project        string            `json:"project"`
Comment        string            `json:"comment"`
VatEnabled     *bool             `json:"vat_enabled"`
VatIncluded    *bool             `json:"vat_included"`
Items          []intOrderItemReq `json:"items" binding:"required,min=1,dive"`
}

func parsePlanDate(s string) *time.Time {
if s == "" { return nil }
t, err := time.Parse("2006-01-02", s)
if err != nil { return nil }
return &t
}

func (h *InternalOrderHandler) toInput(c *gin.Context, req intOrderReq) (repository.InternalOrderInput, error) {
var whID *uuid.UUID
if req.WarehouseID != "" {
if id, err := uuid.Parse(req.WarehouseID); err == nil { whID = &id }
}
var orgID *uuid.UUID
if req.OrganizationID != "" {
if id, err := uuid.Parse(req.OrganizationID); err == nil { orgID = &id }
}
items := make([]repository.InternalOrderItemInput, 0, len(req.Items))
for _, it := range req.Items {
pid, err := uuid.Parse(it.ProductID)
if err != nil { return repository.InternalOrderInput{}, err }
if it.Quantity <= 0 { return repository.InternalOrderInput{}, apperr.BadRequest("quantity must be > 0") }
vat := it.VatRate
if vat == 0 { vat = 20 }
items = append(items, repository.InternalOrderItemInput{
ProductID: pid, Quantity: it.Quantity, Price: it.Price, VatRate: vat,
})
}

var createdBy *uuid.UUID
if uid := mw.CurrentUserID(c); uid != uuid.Nil { createdBy = &uid }

vatEnabled := true
if req.VatEnabled != nil { vatEnabled = *req.VatEnabled }
vatIncluded := true
if req.VatIncluded != nil { vatIncluded = *req.VatIncluded }

return repository.InternalOrderInput{
Number:         req.Number,
OrganizationID: orgID,
WarehouseID:    whID,
PlanDate:       parsePlanDate(req.PlanDate),
Project:        strPtr(req.Project),
Comment:        strPtr(req.Comment),
VatEnabled:     vatEnabled,
VatIncluded:    vatIncluded,
CreatedBy:      createdBy,
Items:          items,
}, nil
}

func (h *InternalOrderHandler) Create(c *gin.Context) {
var req intOrderReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
in, err := h.toInput(c, req)
if err != nil { c.Error(err); return }
o, err := h.svc.Create(c.Request.Context(), in)
if err != nil { c.Error(err); return }
httpx.Created(c, o)
}

func (h *InternalOrderHandler) List(c *gin.Context) {
var statusF *string
if v := c.Query("status"); v != "" { statusF = &v }
var whID *uuid.UUID
if v := c.Query("warehouse_id"); v != "" {
if id, err := uuid.Parse(v); err == nil { whID = &id }
}
items, err := h.svc.List(c.Request.Context(), repository.InternalOrderFilters{
Status: statusF, WarehouseID: whID,
})
if err != nil { c.Error(err); return }
httpx.OK(c, gin.H{"items": items})
}

func (h *InternalOrderHandler) Get(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
o, err := h.svc.Get(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

func (h *InternalOrderHandler) Update(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
var req intOrderReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
in, err := h.toInput(c, req)
if err != nil { c.Error(err); return }
o, err := h.svc.Update(c.Request.Context(), id, in)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

func (h *InternalOrderHandler) Post(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
o, err := h.svc.Post(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

func (h *InternalOrderHandler) Cancel(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
o, err := h.svc.Cancel(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

func (h *InternalOrderHandler) Delete(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
if err := h.svc.Delete(c.Request.Context(), id); err != nil { c.Error(err); return }
c.Status(http.StatusNoContent)
}

func (h *InternalOrderHandler) NextNumber(c *gin.Context) {
n, err := h.svc.NextNumber(c.Request.Context())
if err != nil { c.Error(apperr.Internal("next number", err)); return }
httpx.OK(c, gin.H{"number": n})
}

func (h *InternalOrderHandler) Print(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
if err := h.svc.MarkPrinted(c.Request.Context(), id); err != nil {
c.Error(apperr.Internal("mark printed", err)); return
}
o, err := h.svc.Get(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

func (h *InternalOrderHandler) Send(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
if err := h.svc.MarkSent(c.Request.Context(), id); err != nil {
c.Error(apperr.Internal("mark sent", err)); return
}
o, err := h.svc.Get(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}