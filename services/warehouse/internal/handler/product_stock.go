package handler

import (
"math"
"strings"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
)

type movementJSON struct {
ID                uuid.UUID  `json:"id"`
DocumentID        *uuid.UUID `json:"document_id,omitempty"`
DocumentType      string     `json:"document_type"`
DocumentNumber    string     `json:"document_number"`
DocumentCreatedAt *time.Time `json:"document_created_at,omitempty"`
MovementAt        time.Time  `json:"movement_at"`
QuantityDelta     float64    `json:"quantity_delta"`
DaysOnStock       int        `json:"days_on_stock"`
CostPrice         float64    `json:"cost_price"`
CostSum           float64    `json:"cost_sum"`
}

type warehouseGroupJSON struct {
WarehouseID   uuid.UUID      `json:"warehouse_id"`
WarehouseName string         `json:"warehouse_name"`
Quantity      float64        `json:"quantity"`
CostPrice     float64        `json:"cost_price"`
CostSum       float64        `json:"cost_sum"`
Movements     []movementJSON `json:"movements"`
}

func (h *Handler) ProductStockDetail(c *gin.Context) {
ctx := c.Request.Context()
token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

productID, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid product id")); return }

prods, err := h.productClient.ListByIDs(ctx, token, []uuid.UUID{productID})
if err != nil {
c.Error(apperr.Internal("fetch product", err)); return
}
if len(prods) == 0 {
c.Error(apperr.NotFound("product not found")); return
}
p := prods[0]

var unitShort string
if p.UnitID != nil {
if units, err := h.productClient.ListUnits(ctx, token); err == nil {
for _, u := range units {
if u.ID == *p.UnitID { unitShort = u.ShortName; break }
}
}
}

balances, err := h.svc.BalancesByProduct(ctx, productID)
if err != nil { c.Error(apperr.Internal("balances", err)); return }

balanceByWH := map[uuid.UUID]float64{}
warehouseNames := map[uuid.UUID]string{}
for _, b := range balances {
balanceByWH[b.WarehouseID] = b.Quantity
warehouseNames[b.WarehouseID] = b.WarehouseName
}

movements, err := h.svc.MovementsByProduct(ctx, productID)
if err != nil { c.Error(apperr.Internal("movements", err)); return }

groups := map[uuid.UUID]*warehouseGroupJSON{}
order := []uuid.UUID{}
now := time.Now()

for _, m := range movements {
g, ok := groups[m.WarehouseID]
if !ok {
g = &warehouseGroupJSON{
WarehouseID:   m.WarehouseID,
WarehouseName: m.WarehouseName,
Quantity:      balanceByWH[m.WarehouseID],
CostPrice:     p.CostPrice,
CostSum:       balanceByWH[m.WarehouseID] * p.CostPrice,
Movements:     []movementJSON{},
}
groups[m.WarehouseID] = g
order = append(order, m.WarehouseID)
}
days := int(math.Floor(now.Sub(m.MovementAt).Hours() / 24))

var docType, docNumber string
if m.DocumentType != nil { docType = *m.DocumentType }
if m.DocumentNumber != nil { docNumber = *m.DocumentNumber }

g.Movements = append(g.Movements, movementJSON{
ID:                m.ID,
DocumentID:        m.DocumentID,
DocumentType:      docType,
DocumentNumber:    docNumber,
DocumentCreatedAt: m.DocumentCreatedAt,
MovementAt:        m.MovementAt,
QuantityDelta:     m.QuantityDelta,
DaysOnStock:       days,
CostPrice:         p.CostPrice,
CostSum:           m.QuantityDelta * p.CostPrice,
})
}

for whID, qty := range balanceByWH {
if _, ok := groups[whID]; !ok {
groups[whID] = &warehouseGroupJSON{
WarehouseID:   whID,
WarehouseName: warehouseNames[whID],
Quantity:      qty,
CostPrice:     p.CostPrice,
CostSum:       qty * p.CostPrice,
Movements:     []movementJSON{},
}
order = append(order, whID)
}
}

var totalQty float64
for _, g := range groups { totalQty += g.Quantity }
groupsList := make([]warehouseGroupJSON, 0, len(groups))
for _, id := range order {
groupsList = append(groupsList, *groups[id])
}

httpx.OK(c, gin.H{
"product": gin.H{
"id":         p.ID,
"name":       p.Name,
"sku":        p.SKU,
"unit_short": unitShort,
"cost_price": p.CostPrice,
"price":      p.Price,
"min_stock":  p.MinStock,
"currency":   p.Currency,
},
"summary": gin.H{
"available":  totalQty,
"reserve":    0,
"incoming":   0,
"quantity":   totalQty,
"min_stock":  p.MinStock,
"cost_price": p.CostPrice,
"cost_sum":   totalQty * p.CostPrice,
"sale_price": p.Price,
"sale_sum":   totalQty * p.Price,
},
"warehouses": groupsList,
})
}