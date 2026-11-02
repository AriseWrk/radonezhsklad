package exporter

import (
"bytes"
_ "embed"
"fmt"
"strings"

"github.com/xuri/excelize/v2"

"github.com/radonezhsklad/warehouse/internal/models"
)

//go:embed templates/inventory_template.xlsx
var inventoryTemplate []byte

type InventoryExportCtx struct {
Inventory *models.Inventory
}

// InventoryXLSX заполняет готовый шаблон «Инвентаризация» (структура МойСклад).
func InventoryXLSX(ctx InventoryExportCtx) ([]byte, error) {
f, err := excelize.OpenReader(bytes.NewReader(inventoryTemplate))
if err != nil {
return nil, err
}
defer f.Close()

sheet := "Лист1"
inv := ctx.Inventory

// Шапка
_ = f.SetCellValue(sheet, "D2", inv.Number)
_ = f.SetCellValue(sheet, "D3", inv.DocDate.Format("2006-01-02 15:04:05"))
_ = f.SetCellValue(sheet, "D4", inv.WarehouseName)

// Найти строку «Итого:» по колонке D
rows, _ := f.GetRows(sheet)
totalRow := 0
for idx, row := range rows {
if len(row) >= 4 && strings.TrimSpace(row[3]) == "Итого:" {
totalRow = idx + 1
break
}
}
if totalRow == 0 {
return nil, fmt.Errorf("inventory template: 'Итого:' row not found")
}

startRow := 7
needed := len(inv.Items)
have := totalRow - startRow

// Сохранить стили строки startRow (эталон для вставленных строк)
var styles [10]int
for c := 1; c <= 10; c++ {
col, _ := excelize.ColumnNumberToName(c)
styles[c-1], _ = f.GetCellStyle(sheet, fmt.Sprintf("%s%d", col, startRow))
}

// Удалить шаблонные позиции
for i := 0; i < have; i++ {
_ = f.RemoveRow(sheet, startRow)
}
if needed > 0 {
_ = f.InsertRows(sheet, startRow, needed)
}
totalRow = startRow + needed

var totalCalc, totalFact float64
for i, it := range inv.Items {
r := startRow + i
for c := 1; c <= 10; c++ {
col, _ := excelize.ColumnNumberToName(c)
_ = f.SetCellStyle(sheet, fmt.Sprintf("%s%d", col, r), fmt.Sprintf("%s%d", col, r), styles[c-1])
}
_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", r), i+1)
_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", r), it.ProductSKU)
_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", r), it.ProductName)
_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", r), it.CalculatedQuantity)
_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", r), it.Quantity)
_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", r), it.CorrectionAmount)
_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", r), it.ProductUnitShort)
_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", r), it.Price)
_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", r), it.CorrectionSum)
totalCalc += it.CalculatedQuantity
totalFact += it.Quantity
}

// «Итого:» — суммы E/F, остальные колонки оставляем как в шаблоне
_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), totalCalc)
_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", totalRow), totalFact)

buf, err := f.WriteToBuffer()
if err != nil {
return nil, err
}
return buf.Bytes(), nil
}