package printing

import (
"github.com/radonezhsklad/warehouse/internal/models"
)

// PrintContext — контекст для рендера шаблона.
type PrintContext struct {
Doc             *models.Document
Items           []PrintItem
Warehouse       string
TargetWarehouse string
Organization    string
Contract        string
DateStr         string
}

// PrintItem — одна позиция в шаблоне.
type PrintItem struct {
Name     string
SKU      string
Unit     string
UnitCode string
Quantity float64
Price    float64
Sum      float64
}

// BuildContext — собрать контекст печати из документа.
func BuildContext(doc *models.Document, wh, twh, org, contract string) *PrintContext {
items := make([]PrintItem, 0, len(doc.Items))
for _, it := range doc.Items {
name := it.ProductName
if name == "" && len(it.ProductID.String()) >= 8 {
name = it.ProductID.String()[:8]
}
items = append(items, PrintItem{
Name:     name,
SKU:      it.ProductSKU,
Quantity: it.Quantity,
Price:    it.Price,
Sum:      it.Quantity * it.Price,
})
}
return &PrintContext{
Doc:             doc,
Items:           items,
Warehouse:       wh,
TargetWarehouse: twh,
Organization:    org,
Contract:        contract,
DateStr:         doc.CreatedAt.Format("02.01.2006 15:04"),
}
}