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
	TemplatePath     string
}

// InternalOrderXLSX заполняет эталонный шаблон внутреннего заказа.
func InternalOrderXLSX(ctx OrderExportCtx) ([]byte, error) {
	if ctx.TemplatePath == "" {
		return nil, fmt.Errorf("не задан путь к шаблону внутреннего заказа")
	}

	f, err := excelize.OpenFile(ctx.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("открыть шаблон внутреннего заказа %q: %w", ctx.TemplatePath, err)
	}
	defer f.Close()

	const sheet = "Заявка"
	found := false
	for _, name := range f.GetSheetList() {
		if name == sheet {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("в шаблоне %q отсутствует лист %q", ctx.TemplatePath, sheet)
	}
	if ctx.Order == nil {
		return nil, fmt.Errorf("не передан внутренний заказ")
	}

	o := ctx.Order

	setCell := func(cell string, value interface{}) error {
		if err := f.SetCellValue(sheet, cell, value); err != nil {
			return fmt.Errorf("заполнить ячейку %s: %w", cell, err)
		}
		return nil
	}

	title := fmt.Sprintf("Внутренний заказ № %s от %s", o.Number, formatDateRu(o.DocDate))
	if err := setCell("A1", title); err != nil {
		return nil, err
	}

	warehouseLine := "Склад: " + ctx.WarehouseName
	if err := setCell("A3", warehouseLine); err != nil {
		return nil, err
	}

	if err := setCell("A5", "Объект:"+ctx.ProjectName); err != nil {
		return nil, err
	}

	// В шаблоне строка 9 — маркер начала цикла,
	// строка 10 — образец строки товара, строка 11 — маркер конца цикла.
	// Удаляем верхний маркер: образец перемещается на строку 9.
	if err := f.RemoveRow(sheet, 9); err != nil {
		return nil, fmt.Errorf("удалить верхний маркер шаблона: %w", err)
	}

	itemCount := len(o.Items)

	if itemCount == 0 {
		// Без товаров удаляем строку-образец и нижний маркер.
		if err := f.RemoveRow(sheet, 9); err != nil {
			return nil, fmt.Errorf("удалить пустую строку товара: %w", err)
		}
		if err := f.RemoveRow(sheet, 9); err != nil {
			return nil, fmt.Errorf("удалить нижний маркер шаблона: %w", err)
		}
	} else {
		// Размножаем строку-образец для дополнительных товаров.
		for i := 1; i < itemCount; i++ {
			targetRow := 9 + i
			if err := f.DuplicateRowTo(sheet, 9, targetRow); err != nil {
				return nil, fmt.Errorf("создать строку товара %d: %w", i+1, err)
			}
		}

		// Нижний маркер находится сразу после всех строк товаров.
		markerRow := 9 + itemCount
		if err := f.RemoveRow(sheet, markerRow); err != nil {
			return nil, fmt.Errorf("удалить нижний маркер шаблона: %w", err)
		}
	}

	var totalQty float64
	for i, it := range o.Items {
		row := 9 + i
		p := ctx.Products[it.ProductID]

		unit := ""
		if p.UnitID != nil {
			unit = ctx.Units[*p.UnitID]
		}

		if err := setCell(fmt.Sprintf("A%d", row), i+1); err != nil {
			return nil, err
		}
		if err := setCell(fmt.Sprintf("B%d", row), p.Name); err != nil {
			return nil, err
		}
		if err := setCell(fmt.Sprintf("C%d", row), unit); err != nil {
			return nil, err
		}
		if err := setCell(fmt.Sprintf("D%d", row), it.Quantity); err != nil {
			return nil, err
		}

		totalQty += it.Quantity
	}

	totalRow := 9 + itemCount
	if err := setCell(fmt.Sprintf("C%d", totalRow), "ИТОГО:"); err != nil {
		return nil, err
	}
	if err := setCell(fmt.Sprintf("D%d", totalRow), totalQty); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("сохранить заполненный шаблон: %w", err)
	}
	return buf.Bytes(), nil
}

func formatDateRu(t time.Time) string {
	return t.Format("02.01.2006")
}
