package exporter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/radonezhsklad/warehouse/internal/repository"
	"github.com/xuri/excelize/v2"
)

func ReconciliationXLSX(
	run *repository.ReconciliationRun,
	items []repository.ReconciliationItem,
) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	if err := f.SetSheetName("Sheet1", "Сводка"); err != nil {
		return nil, err
	}
	if _, err := f.NewSheet("Строки"); err != nil {
		return nil, err
	}

	summary := "Сводка"
	rows := "Строки"

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	if err != nil {
		return nil, err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D9EAF7"}, Pattern: 1},
	})
	if err != nil {
		return nil, err
	}

	summaryValues := [][]any{
		{"Сверка с МойСклад"},
		{"ID запуска", run.ID.String()},
		{"Статус", run.Status},
		{"Начало", run.StartedAt},
		{"Завершение", ptrTime(run.FinishedAt)},
		{"Инициатор", run.TriggeredByName},
		{"Тип цены", run.PriceTypeName},
		{"Всего локальных товаров", run.TotalProducts},
		{"Совпавших товаров", run.MatchedProducts},
		{"Расхождений цен", run.PriceDifferences},
		{"Расхождений остатков", run.StockDifferences},
		{"Нет в МойСклад", run.MissingLocalProducts},
		{"Нет локально", run.MissingMSProducts},
		{"Неоднозначных совпадений", run.AmbiguousProducts},
		{"Несопоставленных складов", run.UnmatchedWarehouses},
		{"Позиций с нулевым остатком", run.ZeroStockItems},
	}
	if run.Error != nil {
		summaryValues = append(summaryValues, []any{"Ошибка", *run.Error})
	}

	for i, values := range summaryValues {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if len(values) == 1 {
			_ = f.SetCellValue(summary, cell, values[0])
			_ = f.SetCellStyle(summary, cell, cell, titleStyle)
			continue
		}
		_ = f.SetCellValue(summary, cell, values[0])
		valueCell, _ := excelize.CoordinatesToCellName(2, i+1)
		_ = f.SetCellValue(summary, valueCell, values[1])
	}
	_ = f.SetColWidth(summary, "A", "A", 34)
	_ = f.SetColWidth(summary, "B", "B", 48)

	headers := []string{
		"Тип", "Результат", "Способ сопоставления",
		"Артикул / SKU", "Локальный товар", "Товар МойСклад",
		"ID локального товара", "ID товара МойСклад",
		"Локальный склад", "Склад МойСклад",
		"ID локального склада", "ID склада МойСклад",
		"Локальная цена", "Цена МойСклад", "Разница цены",
		"Локальный остаток", "Остаток МойСклад", "Разница остатка",
		"Нулевой остаток", "Детали",
	}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(rows, cell, header)
		_ = f.SetCellStyle(rows, cell, cell, headerStyle)
	}

	for i, item := range items {
		row := i + 2
		values := []any{
			item.ItemType,
			item.ResultStatus,
			ptrString(item.MatchMethod),
			ptrString(item.SKU),
			ptrString(item.LocalProductName),
			ptrString(item.MSProductName),
			ptrUUID(item.LocalProductID),
			ptrString(item.MSProductID),
			ptrString(item.LocalWarehouseName),
			ptrString(item.MSStoreName),
			ptrUUID(item.LocalWarehouseID),
			ptrString(item.MSStoreID),
			ptrFloat(item.LocalPrice),
			ptrFloat(item.MSPrice),
			ptrFloat(item.PriceDifference),
			ptrFloat(item.LocalStock),
			ptrFloat(item.MSStock),
			ptrFloat(item.StockDifference),
			item.IsZeroStock,
			string(item.Details),
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			_ = f.SetCellValue(rows, cell, value)
		}
	}

	_ = f.SetColWidth(rows, "A", "C", 22)
	_ = f.SetColWidth(rows, "D", "F", 32)
	_ = f.SetColWidth(rows, "G", "L", 38)
	_ = f.SetColWidth(rows, "M", "S", 18)
	_ = f.SetColWidth(rows, "T", "T", 55)
	_ = f.SetPanes(rows, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})
	_ = f.AutoFilter(rows, fmt.Sprintf("A1:T%d", len(items)+1), nil)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ptrString(v *string) any {
	if v == nil {
		return ""
	}
	return *v
}

func ptrUUID(v *uuid.UUID) any {
	if v == nil {
		return ""
	}
	return v.String()
}

func ptrTime(v *time.Time) any {
	if v == nil {
		return ""
	}
	return v.Format("2006-01-02 15:04:05")
}

func ptrFloat(v *float64) any {
	if v == nil {
		return ""
	}
	return *v
}
