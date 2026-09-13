package handler

import (
"strings"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/product"
apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
)

type inventoryPrepareRow struct {
ProductID    uuid.UUID `json:"product_id"`
ProductName  string    `json:"product_name"`
SKU          *string   `json:"sku,omitempty"`
UnitShort    string    `json:"unit_short"`
BookQuantity float64   `json:"book_quantity"`
CostPrice    float64   `json:"cost_price"`
SalePrice    float64   `json:"sale_price"`
}

func (h *Handler) InventoryPrepare(c *gin.Context) {
ctx := c.Request.Context()
token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

whID, err := uuid.Parse(c.Query("warehouse_id"))
if err != nil { c.Error(apperr.BadRequest("invalid warehouse_id")); return }

rows, err := h.svc.BookStockForInventory(ctx, whID)
if err != nil { c.Error(apperr.Internal("book stock", err)); return }

seen := map[uuid.UUID]bool{}
ids := []uuid.UUID{}
for _, r := range rows {
if !seen[r.ProductID] { seen[r.ProductID] = true; ids = append(ids, r.ProductID) }
}

productsMap := map[uuid.UUID]product.Product{}
unitsMap := map[uuid.UUID]string{}
if len(ids) > 0 {
prods, err := h.productClient.ListByIDs(ctx, token, ids)
if err == nil {
for _, p := range prods { productsMap[p.ID] = p }
}
units, err := h.productClient.ListUnits(ctx, token)
if err == nil {
for _, u := range units { unitsMap[u.ID] = u.ShortName }
}
}

items := make([]inventoryPrepareRow, 0, len(rows))
for _, r := range rows {
p := productsMap[r.ProductID]
unitShort := ""
if p.UnitID != nil { unitShort = unitsMap[*p.UnitID] }

items = append(items, inventoryPrepareRow{
ProductID:    r.ProductID,
ProductName:  p.Name,
SKU:          p.SKU,
UnitShort:    unitShort,
BookQuantity: r.BookQuantity,
CostPrice:    p.CostPrice,
SalePrice:    p.Price,
})
}

httpx.OK(c, gin.H{
"warehouse_id": whID,
"items":        items,
"count":        len(items),
})
}