package exporter

import (
"fmt"
"time"

"github.com/google/uuid"
"github.com/xuri/excelize/v2"

"github.com/radonezhsklad/warehouse/internal/models"
"github.com/radonezhsklad/warehouse/internal/product"
)

type OrderExportCtx struct {
Order            *models.InternalOrder
Products         map[uuid.UUID]product.Product
Units            map[uuid.UUID]string
WarehouseName    string
OrganizationName string
ProjectName      string
}

// InternalOrderXLSX генерирует файл «Заявка» (структура как во внутреннем заказе МойСклад).
func InternalOrderXLSX(ctx OrderExportCtx) ([]byte, error) {
f := excelize.NewFile()
defer f.Close()

sheet := "Заявка"
if err := f.SetSheetName("Sheet1", sheet); err != nil {
return nil, err
}

_ = f.SetColWidth(sheet, "A", "A", 8)
_ = f.SetColWidth(sheet, "B", "B", 60)
_ = f.SetColWidth(sheet, "C", "C", 12)
_ = f.SetColWidth(sheet, "D", "D", 12)

bigBold, _ := f.NewStyle(&excelize.Style{
Font: &excelize.Font{Size: 14, Bold: true},
})
headerBold, _ := f.NewStyle(&excelize.Style{
Font:      &excelize.Font{Size: 12, Bold: true},
Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
Border: []excelize.Border{
{Type: "left", Color: "000000", Style: 1},
{Type: "right", Color: "000000", Style: 1},
{Type: "top", Color: "000000", Style: 1},
{Type: "bottom", Color: "000000", Style: 1},
},
})
totalBold, _ := f.NewStyle(&excelize.Style{
Font:      &excelize.Font{Size: 11, Bold: true},
Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
})
totalNum, _ := f.NewStyle(&excelize.Style{
Font:      &excelize.Font{Size: 11, Bold: true},
Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
NumFmt:    2,
})
centerStyle, _ := f.NewStyle(&excelize.Style{
Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
})

o := ctx.Order

title := fmt.Sprintf("Внутренний заказ № %s от %s", o.Number, formatDateRu(o.DocDate))
_ = f.SetCellValue(sheet, "A1", title)
_ = f.SetCellStyle(sheet, "A1", "D1", bigBold)
_ = f.MergeCell(sheet, "A1", "D1")

warehouseLine := "Склад: "
if ctx.WarehouseName != "" {
warehouseLine += ctx.WarehouseName
}
_ = f.SetCellValue(sheet, "A3", warehouseLine)

_ = f.SetCellValue(sheet, "A5", "Объект:"+ctx.ProjectName)

_ = f.SetCellValue(sheet, "A7", "№ п.п.")
_ = f.SetCellValue(sheet, "B7", "Наименование")
_ = f.SetCellValue(sheet, "C7", "Ед. изм.")
_ = f.SetCellValue(sheet, "D7", "Кол-во")
_ = f.SetCellStyle(sheet, "A7", "D7", headerBold)

row := 9
var totalQty float64
for i, it := range o.Items {
p := ctx.Products[it.ProductID]
unit := ""
if p.UnitID != nil {
unit = ctx.Units[*p.UnitID]
}

_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), p.Name)
_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), unit)
_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), it.Quantity)
_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), centerStyle)
_ = f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("D%d", row), centerStyle)
totalQty += it.Quantity
row++
}

_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "ИТОГО:")
_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), totalQty)
_ = f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), totalBold)
_ = f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalNum)

_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row+4),
"Оборудование согласно перечня сдал: _____________________________ Лушкина Н.Г.")
_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row+6),
"Оборудование согласно перечня принял: ________________________")
_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row+9),
`Ген. Директор ООО "ПТС" / Изюмская Е.А./ _______________________________`)

buf, err := f.WriteToBuffer()
if err != nil {
return nil, err
}
return buf.Bytes(), nil
}

func formatDateRu(t time.Time) string {
return t.Format("02.01.2006")
}