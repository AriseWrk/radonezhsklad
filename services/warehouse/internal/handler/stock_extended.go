package handler

import (
"math"
"strings"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/product"
apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
)

type extendedRow struct {
ProductID   uuid.UUID  `json:"product_id"`
ProductName string     `json:"product_name"`
SKU         *string    `json:"sku,omitempty"`
WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
Quantity    float64    `json:"quantity"`
MinStock    float64    `json:"min_stock"`
Reserve     float64    `json:"reserve"`
Incoming    float64    `json:"incoming"`
Available   float64    `json:"available"`
UnitShort   string     `json:"unit_short"`
DaysOnStock *int       `json:"days_on_stock,omitempty"`
CostPrice   float64    `json:"cost_price"`
CostTotal   float64    `json:"cost_total"`
SalePrice   float64    `json:"sale_price"`
SaleTotal   float64    `json:"sale_total"`
}

type extendedTotals struct {
Quantity  float64 `json:"quantity"`
MinStock  float64 `json:"min_stock"`
Reserve   float64 `json:"reserve"`
Incoming  float64 `json:"incoming"`
Available float64 `json:"available"`
CostTotal float64 `json:"cost_total"`
SaleTotal float64 `json:"sale_total"`
}

func (h *Handler) StockExtended(c *gin.Context) {
ctx := c.Request.Context()
token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

var whID *uuid.UUID
if s := c.Query("warehouse_id"); s != "" {
if id, err := uuid.Parse(s); err == nil { whID = &id }
}

rows, err := h.svc.StockExtended(ctx, whID)
if err != nil { c.Error(err); return }

seen := map[uuid.UUID]bool{}
ids := []uuid.UUID{}
for _, r := range rows {
if !seen[r.ProductID] { seen[r.ProductID] = true; ids = append(ids, r.ProductID) }
}

products := map[uuid.UUID]product.Product{}
unitsMap := map[uuid.UUID]string{}
if len(ids) > 0 {
prods, err := h.productClient.ListByIDs(ctx, token, ids)
if err != nil {
c.Error(apperr.Internal("fetch products", err)); return
}
for _, p := range prods { products[p.ID] = p }
units, err := h.productClient.ListUnits(ctx, token)
if err == nil {
for _, u := range units { unitsMap[u.ID] = u.ShortName }
}
}

incoming, err := h.svc.IncomingByProduct(ctx)
if err != nil { incoming = map[uuid.UUID]float64{} }

items := make([]extendedRow, 0, len(rows))
totals := extendedTotals{}
now := time.Now()

for _, r := range rows {
p := products[r.ProductID]
inc := incoming[r.ProductID]

var days *int
if r.LastMovement != nil {
d := int(math.Floor(now.Sub(*r.LastMovement).Hours() / 24))
days = &d
}

unitShort := ""
if p.UnitID != nil { unitShort = unitsMap[*p.UnitID] }

costTotal := r.Quantity * p.CostPrice
saleTotal := r.Quantity * p.Price
available := r.Quantity

items = append(items, extendedRow{
ProductID:   r.ProductID,
ProductName: p.Name,
SKU:         p.SKU,
WarehouseID: &r.WarehouseID,
Quantity:    r.Quantity,
MinStock:    p.MinStock,
Reserve:     0,
Incoming:    inc,
Available:   available,
UnitShort:   unitShort,
DaysOnStock: days,
CostPrice:   p.CostPrice,
CostTotal:   costTotal,
SalePrice:   p.Price,
SaleTotal:   saleTotal,
})

totals.Quantity += r.Quantity
totals.MinStock += p.MinStock
totals.Incoming += inc
totals.Available += available
totals.CostTotal += costTotal
totals.SaleTotal += saleTotal
}

httpx.OK(c, gin.H{
"items":  items,
"totals": totals,
"count":  len(items),
})
}