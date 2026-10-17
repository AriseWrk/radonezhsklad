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

//go:embed templates/inv3_template.xlsx
var inv3Template []byte

// Inv3ExportCtx — данные для генерации формы ИНВ-3 (ОКУД 0317004).
type Inv3ExportCtx struct {
	Inventory     *models.Inventory
	Organization  *models.Organization
	WarehouseName string
}

// okeiByShort — маппинг кратких обозначений единиц на код по ОКЕИ.
var okeiByShort = map[string]string{
	"шт":    "796",
	"м":     "006",
	"км":    "008",
	"кг":    "166",
	"л":     "112",
	"компл": "839",
	"упак":  "778",
}

// InventoryInv3XLSX генерирует форму ИНВ-3, заполняя готовый шаблон.
func InventoryInv3XLSX(ctx Inv3ExportCtx) ([]byte, error) {
	f, err := excelize.OpenReader(bytes.NewReader(inv3Template))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// ---------- стр1 ----------
	s1 := "стр1"
	orgName := ""
	okpo := ""
	legal := ""
	if ctx.Organization != nil {
		orgName = ctx.Organization.Name
		if ctx.Organization.FullName != nil && *ctx.Organization.FullName != "" {
			orgName = *ctx.Organization.FullName
		}
		if ctx.Organization.LegalAddress != nil {
			legal = *ctx.Organization.LegalAddress
		}
		if ctx.Organization.OKPO != nil {
			okpo = *ctx.Organization.OKPO
		}
	}
	orgLine := orgName
	if legal != "" {
		if orgLine != "" {
			orgLine += ", "
		}
		orgLine += legal
	}

	// Заполняем ТОЛЬКО данные. Ячейка E18 ("ИНВЕНТАРИЗАЦИОННАЯ ОПИСЬ") уже отформатирована в шаблоне, её не трогаем.
	_ = f.SetCellValue(s1, "A6", orgLine)
	_ = f.SetCellValue(s1, "R6", okpo)
	_ = f.SetCellValue(s1, "A8", ctx.WarehouseName)
	_ = f.SetCellValue(s1, "M18", ctx.Inventory.Number)
	_ = f.SetCellValue(s1, "R18", ctx.Inventory.DocDate.Format("02.01.2006"))

	// ---------- стр2 ----------
	s2 := "стр2"
	// Ищем строку "Итого"
	rows, _ := f.GetRows(s2)
	totalRow := 0
	for idx, row := range rows {
		for _, cell := range row {
			if strings.TrimSpace(cell) == "Итого" {
				totalRow = idx + 1
				break
			}
		}
		if totalRow > 0 {
			break
		}
	}

	startRow := 4 // Данные начинаются с 4-й строки (1-3 шапка)
	if totalRow <= startRow {
		totalRow = startRow + 1
	}

	// Сохраняем стили 4-й строки ДО удаления, чтобы данные выглядели как таблица
	var dataStyles [13]int
	for c := 1; c <= 13; c++ {
		col, _ := excelize.ColumnNumberToName(c)
		dataStyles[c-1], _ = f.GetCellStyle(s2, fmt.Sprintf("%s%d", col, startRow))
	}

	// Удаляем все старые пустые строки из шаблона между шапкой и "Итого"
	for i := startRow; i < totalRow; i++ {
		_ = f.RemoveRow(s2, startRow)
	}

	// Вставляем нужное количество новых строк
	needed := len(ctx.Inventory.Items)
	if needed > 0 {
		_ = f.InsertRows(s2, startRow, needed)
	}
	totalRow = startRow + needed // Позиция "Итого" сдвинулась вниз

	// Заполняем данные
	var totalQtyFact, totalSumFact, totalQtyCalc, totalSumCalc float64
	for i, it := range ctx.Inventory.Items {
		r := startRow + i

		// Применяем стили
		for c := 1; c <= 13; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			_ = f.SetCellStyle(s2, fmt.Sprintf("%s%d", col, r), fmt.Sprintf("%s%d", col, r), dataStyles[c-1])
		}

		_ = f.SetCellValue(s2, fmt.Sprintf("A%d", r), i+1)
		_ = f.SetCellValue(s2, fmt.Sprintf("B%d", r), "")
		_ = f.SetCellValue(s2, fmt.Sprintf("C%d", r), it.ProductName)
		_ = f.SetCellValue(s2, fmt.Sprintf("D%d", r), it.ProductSKU)
		okei := ""
		if code, ok := okeiByShort[it.ProductUnitShort]; ok {
			okei = code
		}
		_ = f.SetCellValue(s2, fmt.Sprintf("E%d", r), okei)
		_ = f.SetCellValue(s2, fmt.Sprintf("F%d", r), it.ProductUnitShort)
		_ = f.SetCellValue(s2, fmt.Sprintf("G%d", r), it.Price)
		_ = f.SetCellValue(s2, fmt.Sprintf("H%d", r), "")
		_ = f.SetCellValue(s2, fmt.Sprintf("I%d", r), "")
		factSum := it.Quantity * it.Price
		calcSum := it.CalculatedQuantity * it.Price
		_ = f.SetCellValue(s2, fmt.Sprintf("J%d", r), it.Quantity)
		_ = f.SetCellValue(s2, fmt.Sprintf("K%d", r), factSum)
		_ = f.SetCellValue(s2, fmt.Sprintf("L%d", r), it.CalculatedQuantity)
		_ = f.SetCellValue(s2, fmt.Sprintf("M%d", r), calcSum)
		totalQtyFact += it.Quantity
		totalSumFact += factSum
		totalQtyCalc += it.CalculatedQuantity
		totalSumCalc += calcSum
	}

	// Заполняем "Итого"
	_ = f.SetCellValue(s2, fmt.Sprintf("J%d", totalRow), totalQtyFact)
	_ = f.SetCellValue(s2, fmt.Sprintf("K%d", totalRow), totalSumFact)
	_ = f.SetCellValue(s2, fmt.Sprintf("L%d", totalRow), totalQtyCalc)
	_ = f.SetCellValue(s2, fmt.Sprintf("M%d", totalRow), totalSumCalc)

	// ---------- стр3 ----------
	s3 := "стр3"
	n := len(ctx.Inventory.Items)
	rub := int64(totalSumFact)
	kop := int64(math.Round((totalSumFact - float64(rub)) * 100))
	sumWords := inv3NumToWordsRu(rub, false) + " рублей " + fmt.Sprintf("%02d", kop) + " копеек"

	// Заполняем только значения, не трогая объединения шаблона
	_ = f.SetCellValue(s3, "L3", inv3NumToWordsRu(int64(n), false))
	_ = f.SetCellValue(s3, "L5", inv3NumToWordsRu(int64(totalQtyFact), true))
	_ = f.SetCellValue(s3, "I7", sumWords)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// inv3NumToWordsRu — число прописью (0..999 999 999 999).
func inv3NumToWordsRu(n int64, female bool) string {
	if n == 0 {
		return "ноль"
	}
	onesM := []string{"", "один", "два", "три", "четыре", "пять", "шесть", "семь", "восемь", "девять"}
	onesF := []string{"", "одна", "две", "три", "четыре", "пять", "шесть", "семь", "восемь", "девять"}
	teens := []string{"десять", "одиннадцать", "двенадцать", "тринадцать", "четырнадцать", "пятнадцать", "шестнадцать", "семнадцать", "восемнадцать", "девятнадцать"}
	tens := []string{"", "", "двадцать", "тридцать", "сорок", "пятьдесят", "шестьдесят", "семьдесят", "восемьдесят", "девяносто"}
	hundreds := []string{"", "сто", "двести", "триста", "четыреста", "пятьсот", "шестьсот", "семьсот", "восемьсот", "девятьсот"}

	three := func(x int, fem bool) string {
		var parts []string
		h := x / 100
		t := (x / 10) % 10
		o := x % 10
		if h > 0 {
			parts = append(parts, hundreds[h])
		}
		if t == 1 {
			parts = append(parts, teens[o])
		} else {
			if t > 1 {
				parts = append(parts, tens[t])
			}
			if o > 0 {
				if fem {
					parts = append(parts, onesF[o])
				} else {
					parts = append(parts, onesM[o])
				}
			}
		}
		return strings.Join(parts, " ")
	}
	plural := func(x int, one, few, many string) string {
		x = x % 100
		if x >= 11 && x <= 19 {
			return many
		}
		switch x % 10 {
		case 1:
			return one
		case 2, 3, 4:
			return few
		}
		return many
	}

	m := n / 1000000
	th := (n / 1000) % 1000
	un := n % 1000
	var parts []string
	if m > 0 {
		parts = append(parts, three(int(m), false), plural(int(m), "миллион", "миллиона", "миллионов"))
	}
	if th > 0 {
		parts = append(parts, three(int(th), true), plural(int(th), "тысяча", "тысячи", "тысяч"))
	}
	if un > 0 {
		parts = append(parts, three(int(un), female))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
