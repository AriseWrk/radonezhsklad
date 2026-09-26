package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperr "github.com/radonezhsklad/shared/errors"
	"github.com/radonezhsklad/warehouse/internal/exporter"
	"github.com/radonezhsklad/warehouse/internal/product"
)

// Export — генерирует Excel «Инвентаризация», отдаёт файл.
func (h *InventoryHandler) Export(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperr.BadRequest("invalid id"))
		return
	}

	ctx := c.Request.Context()
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	inv, err := h.svc.Get(ctx, id)
	if err != nil {
		c.Error(err)
		return
	}

	if len(inv.Items) > 0 {
		ids := make([]uuid.UUID, 0, len(inv.Items))
		seen := map[uuid.UUID]bool{}
		for _, it := range inv.Items {
			if !seen[it.ProductID] {
				seen[it.ProductID] = true
				ids = append(ids, it.ProductID)
			}
		}
		prods, err := h.productCli.ListByIDs(ctx, token, ids)
		if err == nil {
			pmap := make(map[uuid.UUID]product.Product, len(prods))
			for _, p := range prods {
				pmap[p.ID] = p
			}
			units, _ := h.productCli.ListUnits(ctx, token)
			umap := make(map[uuid.UUID]string, len(units))
			for _, u := range units {
				umap[u.ID] = u.ShortName
			}
			for i := range inv.Items {
				if p, ok := pmap[inv.Items[i].ProductID]; ok {
					inv.Items[i].ProductName = p.Name
					if p.SKU != nil {
						inv.Items[i].ProductSKU = *p.SKU
					}
					if p.UnitID != nil {
						inv.Items[i].ProductUnitShort = umap[*p.UnitID]
					}
				}
			}
		}
	}

	data, err := exporter.InventoryXLSX(exporter.InventoryExportCtx{Inventory: inv})
	if err != nil {
		c.Error(apperr.Internal("export xlsx", err))
		return
	}

	filename := fmt.Sprintf("inventory-%s.xls", extractNumericTail(inv.Number))
	c.Header("Content-Type", "application/vnd.ms-excel")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("X-Filename", filename)
	c.Data(200, "application/vnd.ms-excel", data)
}
