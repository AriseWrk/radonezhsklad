package handler

import (
"strconv"

"github.com/gin-gonic/gin"

apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
)

type salesAnalyticsRow struct {
ProductID   string   `json:"product_id"`
SoldQty     float64  `json:"sold_qty"`
SoldSum     float64  `json:"sold_sum"`
OrdersCount int      `json:"orders_count"`
FirstSoldAt *string  `json:"first_sold_at,omitempty"`
LastSoldAt  *string  `json:"last_sold_at,omitempty"`
}

func (h *Handler) SalesAnalytics(c *gin.Context) {
days := 14
if s := c.Query("days"); s != "" {
if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 365 { days = n }
}

rows, err := h.svc.SalesAnalytics(c.Request.Context(), days)
if err != nil {
c.Error(apperr.Internal("sales analytics", err)); return
}

out := make([]salesAnalyticsRow, 0, len(rows))
for _, r := range rows {
row := salesAnalyticsRow{
ProductID:   r.ProductID.String(),
SoldQty:     r.SoldQty,
SoldSum:     r.SoldSum,
OrdersCount: r.OrdersCount,
}
if r.FirstSoldAt != nil { s := r.FirstSoldAt.Format("2006-01-02T15:04:05Z07:00"); row.FirstSoldAt = &s }
if r.LastSoldAt != nil  { s := r.LastSoldAt.Format("2006-01-02T15:04:05Z07:00");  row.LastSoldAt = &s }
out = append(out, row)
}

httpx.OK(c, gin.H{
"days":  days,
"items": out,
"count": len(out),
})
}