package exporter

import (
"bytes"
_ "embed"
"fmt"
"math"
"strings"

"github.com/xuri/excelize/v2"

"github.com/radonezhsklad/warehouse/internal/models"
)

//go:embed templates/inv19_template.xlsx
var inv19Template []byte

// Inv19ExportCtx — данные для генерации сличительной ведомости ИНВ-19 (ОКУД 317017).
type Inv19ExportCtx struct {
Inventory     *models.Inventory
Organization  *models.Organization
WarehouseName string
}

// InventoryInv19XLSX заполняет готовый шаблон сличительной ведомости ИНВ-19.
func InventoryInv19XLSX(ctx Inv19ExportCtx) ([]byte, error) {
f, err := excelize.OpenReader(bytes.NewReader(inv19Template))
if err != nil {
return nil, err
}
defer f.Close()

sheet := "Лист1"
inv := ctx.Inventory

// Шапка
_ = f.SetCellValue(sheet, "O15", inv.Number)
_ = f.SetCellValue(sheet, "R15", inv.DocDate.Format("2006-01-02 15:04:05"))

// Найти строку «Итого»: первая строка начиная с 35-й, где E пустая
startRow := 35
rows, _ := f.GetRows(sheet)
totalRow := 0
for r := startRow; r <= len(rows); r++ {
var eVal string
if r-1 < len(rows) && len(rows[r-1]) >= 5 {
eVal = strings.TrimSpace(rows[r-1][4])
}
if eVal == "" {
totalRow = r
break
}
}
if totalRow == 0 {
totalRow = len(rows) + 1
}

needed := len(inv.Items)
have := totalRow - startRow

// Стили эталонной строки позиций
var styles [11]int
cols := []string{"C", "E", "H", "J", "K", "Q", "S", "T", "W"}
for i, col := range cols {
styles[i], _ = f.GetCellStyle(sheet, fmt.Sprintf("%s%d", col, startRow))
}

for i := 0; i < have; i++ {
_ = f.RemoveRow(sheet, startRow)
}
if needed > 0 {
_ = f.InsertRows(sheet, startRow, needed)
}
totalRow = startRow + needed

var totalSurplusQty, totalSurplusSum, totalShortageQty, totalShortageSum float64
for i, it := range inv.Items {
r := startRow + i

for j, col := range cols {
_ = f.SetCellStyle(sheet, fmt.Sprintf("%s%d", col, r), fmt.Sprintf("%s%d", col, r), styles[j])
}

surplusQty := 0.0
surplusSum := 0.0
shortageQty := 0.0
shortageSum := 0.0
if it.CorrectionAmount > 0 {
surplusQty = it.CorrectionAmount
surplusSum = math.Abs(it.CorrectionSum)
} else if it.CorrectionAmount < 0 {
shortageQty = math.Abs(it.CorrectionAmount)
shortageSum = math.Abs(it.CorrectionSum)
}

okei := ""
if code, ok := okeiByShort[it.ProductUnitShort]; ok {
okei = code
}

_ = f.SetCellFormula(sheet, fmt.Sprintf("C%d", r), "")
_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", r), i+1)
_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", r), it.ProductName)
_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", r), it.ProductSKU)
_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", r), okei)
_ = f.SetCellValue(sheet, fmt.Sprintf("K%d", r), it.ProductUnitShort)
_ = f.SetCellValue(sheet, fmt.Sprintf("Q%d", r), surplusQty)
_ = f.SetCellValue(sheet, fmt.Sprintf("S%d", r), surplusSum)
_ = f.SetCellValue(sheet, fmt.Sprintf("T%d", r), shortageQty)
_ = f.SetCellValue(sheet, fmt.Sprintf("W%d", r), shortageSum)

totalSurplusQty += surplusQty
totalSurplusSum += surplusSum
totalShortageQty += shortageQty
totalShortageSum += shortageSum
}

// «Итого» — суммы Q/S/T/W
_ = f.SetCellValue(sheet, fmt.Sprintf("Q%d", totalRow), totalSurplusQty)
_ = f.SetCellValue(sheet, fmt.Sprintf("S%d", totalRow), totalSurplusSum)
_ = f.SetCellValue(sheet, fmt.Sprintf("T%d", totalRow), totalShortageQty)
_ = f.SetCellValue(sheet, fmt.Sprintf("W%d", totalRow), totalShortageSum)

buf, err := f.WriteToBuffer()
if err != nil {
return nil, err
}
return buf.Bytes(), nil
}