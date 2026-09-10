package handler

import (
"net/http"
"strings"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/order/internal/repository"
"github.com/radonezhsklad/order/internal/service"
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

// ---------- customers ----------

type customerReq struct {
Name    string `json:"name" binding:"required"`
Phone   string `json:"phone"`
Email   string `json:"email"`
Address string `json:"address"`
}

func (h *Handler) CreateCustomer(c *gin.Context) {
var req customerReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
item, err := h.svc.CreateCustomer(c.Request.Context(), req.Name, strPtr(req.Phone), strPtr(req.Email), strPtr(req.Address))
if err != nil { c.Error(err); return }
httpx.Created(c, item)
}

func (h *Handler) ListCustomers(c *gin.Context) {
items, err := h.svc.ListCustomers(c.Request.Context())
if err != nil { c.Error(err); return }
httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) GetCustomer(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
item, err := h.svc.GetCustomer(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, item)
}

func (h *Handler) DeleteCustomer(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
if err := h.svc.DeleteCustomer(c.Request.Context(), id); err != nil { c.Error(err); return }
c.Status(http.StatusNoContent)
}

// ---------- orders ----------

type orderItemReq struct {
ProductID string  `json:"product_id" binding:"required"`
Quantity  float64 `json:"quantity"   binding:"required"`
Price     float64 `json:"price"`
}

type orderReq struct {
Number      string         `json:"number"`
CustomerID  string         `json:"customer_id"  binding:"required"`
WarehouseID string         `json:"warehouse_id" binding:"required"`
Comment     string         `json:"comment"`
Items       []orderItemReq `json:"items" binding:"required,min=1,dive"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
var req orderReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }

custID, err := uuid.Parse(req.CustomerID)
if err != nil { c.Error(apperr.BadRequest("invalid customer_id")); return }
whID, err := uuid.Parse(req.WarehouseID)
if err != nil { c.Error(apperr.BadRequest("invalid warehouse_id")); return }

var createdBy *uuid.UUID
if uid := mw.CurrentUserID(c); uid != uuid.Nil { createdBy = &uid }

items := make([]repository.OrderItemInput, 0, len(req.Items))
for _, it := range req.Items {
pid, err := uuid.Parse(it.ProductID)
if err != nil { c.Error(apperr.BadRequest("invalid product_id")); return }
if it.Quantity <= 0 { c.Error(apperr.BadRequest("quantity must be > 0")); return }
items = append(items, repository.OrderItemInput{ProductID: pid, Quantity: it.Quantity, Price: it.Price})
}

o, err := h.svc.CreateOrder(c.Request.Context(), repository.OrderInput{
Number:      req.Number,
CustomerID:  custID,
WarehouseID: whID,
Comment:     strPtr(req.Comment),
CreatedBy:   createdBy,
Items:       items,
})
if err != nil { c.Error(err); return }
httpx.Created(c, o)
}

func (h *Handler) ListOrders(c *gin.Context) {
var statusF *string
if v := c.Query("status"); v != "" { statusF = &v }
custID := parseUUIDOpt(c.Query("customer_id"))
items, err := h.svc.ListOrders(c.Request.Context(), statusF, custID)
if err != nil { c.Error(err); return }
httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) GetOrder(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
o, err := h.svc.GetOrder(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

func (h *Handler) ConfirmOrder(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
o, err := h.svc.ConfirmOrder(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

// ShipOrder — берёт Bearer-токен из заголовка и передаёт в warehouse.
func (h *Handler) ShipOrder(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }

auth := c.GetHeader("Authorization")
token := strings.TrimPrefix(auth, "Bearer ")

o, err := h.svc.ShipOrder(c.Request.Context(), id, token)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}

func (h *Handler) CancelOrder(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
o, err := h.svc.CancelOrder(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, o)
}