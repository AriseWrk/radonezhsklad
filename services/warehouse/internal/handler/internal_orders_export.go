package handler

import (
"fmt"
"strings"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/exporter"
"github.com/radonezhsklad/warehouse/internal/product"
apperr "github.com/radonezhsklad/shared/errors"
)

// Export — генерирует Excel «Заявка», отмечает заказ как напечатанный, отдаёт файл.
func (h *InternalOrderHandler) Export(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }

ctx := c.Request.Context()
token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

o, err := h.svc.Get(ctx, id)
if err != nil { c.Error(err); return }

ids := []uuid.UUID{}
for _, it := range o.Items { ids = append(ids, it.ProductID) }

productsMap := map[uuid.UUID]product.Product{}
if len(ids) > 0 {
prods, err := h.productCli.ListByIDs(ctx, token, ids)
if err != nil { c.Error(apperr.Internal("fetch products", err)); return }
for _, p := range prods { productsMap[p.ID] = p }
}

unitsMap := map[uuid.UUID]string{}
if us, err := h.productCli.ListUnits(ctx, token); err == nil {
for _, u := range us { unitsMap[u.ID] = u.ShortName }
}

warehouseName := ""
if o.WarehouseID != nil {
if w, err := h.warehouseSvc.GetWarehouse(ctx, *o.WarehouseID); err == nil && w != nil {
warehouseName = w.Name
}
}

orgName := ""
if o.OrganizationID != nil {
if org, err := h.supplierSvc.GetOrganization(ctx, *o.OrganizationID); err == nil && org != nil {
orgName = org.Name
}
}

project := ""
if o.Project != nil { project = *o.Project }

// отметим как напечатанный
_ = h.svc.MarkPrinted(ctx, id)

data, err := exporter.InternalOrderXLSX(exporter.OrderExportCtx{
Order:            o,
Products:         productsMap,
Units:            unitsMap,
WarehouseName:    warehouseName,
OrganizationName: orgName,
ProjectName:      project,
})
if err != nil { c.Error(apperr.Internal("export xlsx", err)); return }

filename := fmt.Sprintf("intorder1-%s.xls", extractNumericTail(o.Number))

c.Header("Content-Type", "application/vnd.ms-excel")
c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
c.Header("X-Filename", filename)
c.Data(200, "application/vnd.ms-excel", data)
}

// extractNumericTail: "ВЗ-1" → "1", "02036" → "02036", "ORD-12-A" → "12"
func extractNumericTail(s string) string {
parts := strings.Split(s, "-")
if len(parts) == 0 { return s }
last := parts[len(parts)-1]
if last != "" { return last }
return s
}