package exporter

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/radonezhsklad/warehouse/internal/repository"
	"github.com/xuri/excelize/v2"
)

func TestReconciliationXLSX(t *testing.T) {
	runID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()

	now := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	matchMethod := "external_id"
	sku := "TEST-001"
	localName := "Тестовый товар"
	msName := "Тестовый товар МС"
	warehouseName := "Основной склад"
	storeName := "Склад МС"
	msProductID := "ms-product-001"
	msStoreID := "ms-store-001"
	localPrice := 150.0
	msPrice := 140.0
	priceDifference := 10.0
	localStock := 5.0
	msStock := 3.0
	stockDifference := 2.0

	run := &repository.ReconciliationRun{
		ID:               runID,
		Status:           "completed",
		StartedAt:        now,
		FinishedAt:       &now,
		TriggeredByName:  "Тест",
		PriceTypeName:    "Цена продажи",
		TotalProducts:    1,
		MatchedProducts:  1,
		PriceDifferences: 1,
		StockDifferences: 1,
	}

	items := []repository.ReconciliationItem{
		{
			ID:                 uuid.New(),
			RunID:              runID,
			ItemType:           "price",
			ResultStatus:       "difference",
			MatchMethod:        &matchMethod,
			LocalProductID:     &productID,
			MSProductID:        &msProductID,
			SKU:                &sku,
			LocalProductName:   &localName,
			MSProductName:      &msName,
			LocalWarehouseID:   &warehouseID,
			MSStoreID:          &msStoreID,
			LocalWarehouseName: &warehouseName,
			MSStoreName:        &storeName,
			LocalPrice:         &localPrice,
			MSPrice:            &msPrice,
			PriceDifference:    &priceDifference,
			LocalStock:         &localStock,
			MSStock:            &msStock,
			StockDifference:    &stockDifference,
			Details:            []byte("{\"test\":true}"),
			CreatedAt:          now,
		},
	}

	data, err := ReconciliationXLSX(run, items)
	if err != nil {
		t.Fatalf("ReconciliationXLSX returned error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("generated XLSX is empty")
	}

	file, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("cannot reopen generated XLSX: %v", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) != 2 || sheets[0] != "Сводка" || sheets[1] != "Строки" {
		t.Fatalf("unexpected sheets: %#v", sheets)
	}

	checks := []struct {
		sheet string
		cell  string
		want  string
	}{
		{"Сводка", "A1", "Сверка с МойСклад"},
		{"Сводка", "B2", runID.String()},
		{"Сводка", "B7", "Цена продажи"},
		{"Строки", "A1", "Тип"},
		{"Строки", "B1", "Результат"},
		{"Строки", "D2", sku},
		{"Строки", "E2", localName},
		{"Строки", "F2", msName},
		{"Строки", "T2", "{\"test\":true}"},
	}

	for _, check := range checks {
		got, err := file.GetCellValue(check.sheet, check.cell)
		if err != nil {
			t.Fatalf("cannot read %s!%s: %v", check.sheet, check.cell, err)
		}
		if got != check.want {
			t.Errorf("%s!%s = %q, want %q", check.sheet, check.cell, got, check.want)
		}
	}
}
