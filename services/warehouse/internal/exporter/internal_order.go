package exporter

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

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

// Ширины столбцов (в условных символах Calibri 11).
const (
	colWidthA float64 = 8
	colWidthB float64 = 60
	colWidthC float64 = 12
	colWidthD float64 = 12
)

// Суммарная ширина merge-области A:D — для оценки числа строк при wrap.
const mergedADWidth = colWidthA + colWidthB + colWidthC + colWidthD

// InternalOrderXLSX генерирует файл «Заявка» (структура как во внутреннем заказе МойСклад).
func InternalOrderXLSX(ctx OrderExportCtx) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Заявка"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}

	_ = f.SetColWidth(sheet, "A", "A", colWidthA)
	_ = f.SetColWidth(sheet, "B", "B", colWidthB)
	_ = f.SetColWidth(sheet, "C", "C", colWidthC)
	_ = f.SetColWidth(sheet, "D", "D", colWidthD)

	bigBold, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 14, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
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
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
	})
	totalNum, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		NumFmt:    2,
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
	})
	bordered, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
	})
	// textWrap — для длинных текстовых ячеек: перенос по словам, выравнивание по верху.
	textWrap, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "top", WrapText: true},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
	})
	// infoWrap — без рамки, для merge-строк «Склад: …», «Объект: …» и подписей.
	infoWrap, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "top", WrapText: true},
	})

	o := ctx.Order

	title := fmt.Sprintf("Внутренний заказ № %s от %s", o.Number, formatDateRu(o.DocDate))
	_ = f.SetCellValue(sheet, "A1", title)
	_ = f.SetCellStyle(sheet, "A1", "D1", bigBold)
	_ = f.MergeCell(sheet, "A1", "D1")
	_ = f.SetRowHeight(sheet, 1, rowHeightForLines(estimateLines(title, mergedADWidth)))

	warehouseLine := "Склад: "
	if ctx.WarehouseName != "" {
		warehouseLine += ctx.WarehouseName
	}
	_ = f.SetCellValue(sheet, "A3", warehouseLine)
	_ = f.SetCellStyle(sheet, "A3", "D3", infoWrap)
	_ = f.MergeCell(sheet, "A3", "D3")
	_ = f.SetRowHeight(sheet, 3, rowHeightForLines(estimateLines(warehouseLine, mergedADWidth)))

	objLine := "Объект:" + ctx.ProjectName
	_ = f.SetCellValue(sheet, "A5", objLine)
	_ = f.SetCellStyle(sheet, "A5", "D5", infoWrap)
	_ = f.MergeCell(sheet, "A5", "D5")
	_ = f.SetRowHeight(sheet, 5, rowHeightForLines(estimateLines(objLine, mergedADWidth)))

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
		_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), textWrap)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), unit)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), it.Quantity)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), centerStyle)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("D%d", row), centerStyle)
		totalQty += it.Quantity
		row++
	}

	_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "ИТОГО:")
	_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), totalQty)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), bordered)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), totalBold)
	_ = f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalNum)

	sign1 := "Оборудование согласно перечня сдал: _____________________________ Лушкина Н.Г."
	sign2 := "Оборудование согласно перечня принял: ________________________"
	sign3 := `Ген. Директор ООО "ПТС" / Изюмская Е.А./ _______________________________`

	setMergedInfoLine(f, sheet, row+4, sign1, infoWrap)
	setMergedInfoLine(f, sheet, row+6, sign2, infoWrap)
	setMergedInfoLine(f, sheet, row+9, sign3, infoWrap)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatDateRu(t time.Time) string {
	return t.Format("02.01.2006")
}

// setMergedInfoLine пишет текст в объединённую ячейку A:D указанной строки,
// включает перенос и подгоняет высоту строки.
func setMergedInfoLine(f *excelize.File, sheet string, row int, text string, style int) {
	cellA := fmt.Sprintf("A%d", row)
	cellD := fmt.Sprintf("D%d", row)
	_ = f.SetCellValue(sheet, cellA, text)
	_ = f.SetCellStyle(sheet, cellA, cellD, style)
	_ = f.MergeCell(sheet, cellA, cellD)
	_ = f.SetRowHeight(sheet, row, rowHeightForLines(estimateLines(text, mergedADWidth)))
}

// estimateLines — грубая оценка числа строк текста при заданной ширине (в символах Calibri 11).
// Учитывает явные \n и поправку 0.95 на кириллицу (шире латиницы).
func estimateLines(text string, widthChars float64) int {
	if strings.TrimSpace(text) == "" {
		return 1
	}
	if widthChars <= 0 {
		widthChars = 60
	}
	total := 0
	for _, part := range strings.Split(text, "\n") {
		n := utf8.RuneCountInString(part)
		if n == 0 {
			total++
			continue
		}
		lines := int(math.Ceil(float64(n) * 1.5 / widthChars))
		if lines < 1 {
			lines = 1
		}
		total += lines
	}
	if total < 1 {
		total = 1
	}
	return total
}

// rowHeightForLines — высота строки в пунктах: 15pt на строку, минимум 15pt.
func rowHeightForLines(lines int) float64 {
	if lines < 1 {
		lines = 1
	}
	return float64(lines) * 15.0
}
