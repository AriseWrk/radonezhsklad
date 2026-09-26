package exporter

import (
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/radonezhsklad/warehouse/internal/models"
)

type InventoryExportCtx struct {
	Inventory *models.Inventory
}

// InventoryXLSX генерирует файл «Инвентаризация» (структура как в отчёте МойСклад).
func InventoryXLSX(ctx InventoryExportCtx) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Лист1"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}

	_ = f.SetColWidth(sheet, "A", "A", 3)
	_ = f.SetColWidth(sheet, "B", "B", 6)
	_ = f.SetColWidth(sheet, "C", "C", 12)
	_ = f.SetColWidth(sheet, "D", "D", 55)
	_ = f.SetColWidth(sheet, "E", "E", 16)
	_ = f.SetColWidth(sheet, "F", "F", 16)
	_ = f.SetColWidth(sheet, "G", "G", 10)
	_ = f.SetColWidth(sheet, "H", "H", 8)
	_ = f.SetColWidth(sheet, "I", "I", 12)
	_ = f.SetColWidth(sheet, "J", "J", 18)

	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Size: 11, Bold: true}})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	cellLeft, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	numRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	numCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	inv := ctx.Inventory

	_ = f.SetCellValue(sheet, "B2", "Инвентаризация:")
	_ = f.SetCellStyle(sheet, "B2", "B2", titleStyle)
	_ = f.SetCellValue(sheet, "D2", inv.Number)

	_ = f.SetCellValue(sheet, "B3", "Дата проведения:")
	_ = f.SetCellStyle(sheet, "B3", "B3", titleStyle)
	_ = f.SetCellValue(sheet, "D3", inv.DocDate.Format("2006-01-02 15:04:05"))

	_ = f.SetCellValue(sheet, "B4", "Склад:")
	_ = f.SetCellStyle(sheet, "B4", "B4", titleStyle)
	_ = f.SetCellValue(sheet, "D4", inv.WarehouseName)

	headers := []struct{ col, val string }{
		{"B", "№"},
		{"C", "Код"},
		{"D", "Наименование"},
		{"E", "Расчетный остаток"},
		{"F", "Фактический остаток"},
		{"G", "Разница"},
		{"H", "Ед. изм."},
		{"I", "Цена"},
		{"J", "Избыток / недостача"},
	}
	for _, h := range headers {
		_ = f.SetCellValue(sheet, h.col+"6", h.val)
	}
	_ = f.SetCellStyle(sheet, "B6", "J6", headerStyle)

	row := 7
	var totalCalc, totalFact float64
	for i, it := range inv.Items {
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), i+1)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numCenter)

		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), it.ProductSKU)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), it.ProductName)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), it.CalculatedQuantity)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), it.Quantity)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), it.CorrectionAmount)
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), it.ProductUnitShort)
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", row), it.Price)
		_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", row), it.CorrectionSum)

		_ = f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("D%d", row), cellLeft)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("G%d", row), numRight)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), numCenter)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("I%d", row), fmt.Sprintf("J%d", row), numRight)

		totalCalc += it.CalculatedQuantity
		totalFact += it.Quantity
		row++
	}

	_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "Итого:")
	_ = f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), totalCalc)
	_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), totalFact)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("F%d", row), numRight)
	_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), "")
	_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "")
	_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", row), "")
	_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", row), "")
	_ = f.SetCellStyle(sheet, fmt.Sprintf("G%d", row), fmt.Sprintf("J%d", row), numRight)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
