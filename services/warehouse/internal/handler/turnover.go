package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperr "github.com/radonezhsklad/shared/errors"
	"github.com/radonezhsklad/shared/httpx"
)

type turnoverRowJSON struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	SKU         string    `json:"sku,omitempty"`
	UnitShort   string    `json:"unit_short"`
	Opening     float64   `json:"opening"`
	Income      float64   `json:"income"`
	Outcome     float64   `json:"outcome"`
	Closing     float64   `json:"closing"`
}

func parseDateOr(s string, def time.Time) time.Time {
	if s == "" {
		return def
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return def
	}
	return t
}

func (h *Handler) TurnoverReport(c *gin.Context) {
	ctx := c.Request.Context()
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	var whID *uuid.UUID
	if v := c.Query("warehouse_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.Error(apperr.BadRequest("invalid warehouse_id"))
			return
		}
		whID = &id
	}

	from := parseDateOr(c.Query("from"), time.Now().AddDate(0, 0, -30))
	to := parseDateOr(c.Query("to"), time.Now())
	fromT := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	toT := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, time.Local)

	rows, err := h.svc.TurnoverReport(ctx, whID, fromT, toT)
	if err != nil {
		c.Error(apperr.Internal("turnover", err))
		return
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ProductID)
	}
	nameByID := map[uuid.UUID]string{}
	skuByID := map[uuid.UUID]string{}
	unitByID := map[uuid.UUID]string{}
	if len(ids) > 0 {
		if prods, err := h.productClient.ListByIDs(ctx, token, ids); err == nil {
			for _, p := range prods {
				nameByID[p.ID] = p.Name
				if p.SKU != nil {
					skuByID[p.ID] = *p.SKU
				}
			}
			var unitIDs []uuid.UUID
			for _, p := range prods {
				if p.UnitID != nil {
					unitIDs = append(unitIDs, *p.UnitID)
				}
			}
			if len(unitIDs) > 0 {
				if units, err := h.productClient.ListUnits(ctx, token); err == nil {
					unitNameByID := map[uuid.UUID]string{}
					for _, u := range units {
						unitNameByID[u.ID] = u.ShortName
					}
					for _, p := range prods {
						if p.UnitID != nil {
							unitByID[p.ID] = unitNameByID[*p.UnitID]
						}
					}
				}
			}
		}
	}

	out := make([]turnoverRowJSON, 0, len(rows))
	for _, r := range rows {
		out = append(out, turnoverRowJSON{
			ProductID:   r.ProductID,
			ProductName: nameByID[r.ProductID],
			SKU:         skuByID[r.ProductID],
			UnitShort:   unitByID[r.ProductID],
			Opening:     r.Opening,
			Income:      r.Income,
			Outcome:     r.Outcome,
			Closing:     r.Closing,
		})
	}
	httpx.OK(c, gin.H{"items": out, "count": len(out)})
}
