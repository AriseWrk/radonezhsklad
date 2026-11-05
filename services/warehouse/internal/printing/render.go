package printing

import (
"bytes"
"fmt"
"strconv"
"strings"

"github.com/xuri/excelize/v2"
)

const maxCols = 40 // сколько колонок сканировать

// Render — открывает шаблон и рендерит с контекстом.
func Render(templatePath string, pc *PrintContext) ([]byte, error) {
f, err := excelize.OpenFile(templatePath)
if err != nil {
return nil, err
}
defer f.Close()

for _, sheet := range f.GetSheetList() {
blk, err := FindForEach(f, sheet)
if err != nil {
return nil, err
}
if blk != nil {
if err := renderLoop(f, sheet, blk, pc); err != nil {
return nil, err
}
}
// После рендера цикла — скалярные подстановки (везде, кроме служебных строк уже удалены).
if err := replaceScalar(f, sheet, pc); err != nil {
return nil, err
}
// Формулы: $[SUM(X23)@0] -> =SUM(X23:XN) с учётом кол-ва строк
if blk != nil {
if err := fixFormulas(f, sheet, blk, pc); err != nil {
return nil, err
}
}
}

var buf bytes.Buffer
if err := f.Write(&buf); err != nil {
return nil, err
}
return buf.Bytes(), nil
}

// renderLoop обрабатывает <jx:forEach>.
// Логика:
//   1) Удаляем все строки между Start+1 и End-1 (шаблонные), запомнив их содержимое.
//   2) Вставляем len(Items) строк после StartRow (вместо удалённых).
//   3) Заполняем эти строки по очереди.
//   4) Удаляем строку StartRow (с <jx:forEach>) и EndRow (с </jx:forEach>) — они сдвигаются.
func renderLoop(f *excelize.File, sheet string, blk *ForEachBlock, pc *PrintContext) error {
n := len(pc.Items)
if n == 0 {
n = 1 // хотя бы одна строка, чтобы не сломать шаблон
}

// Сохраняем данные шаблонных строк (по колонкам).
type cellData struct {
val string
}
tmplRows := blk.TemplateRows
if len(tmplRows) == 0 {
return fmt.Errorf("empty foreach block")
}
// Считаем, что данные — в последней строке блока; выше — заголовки (не трогаем).
dataRow := tmplRows[len(tmplRows)-1]
headerRows := tmplRows[:len(tmplRows)-1]

// Сохраняем данные dataRow по колонкам.
saved := make([]string, maxCols+1)
for c := 1; c <= maxCols; c++ {
colName, _ := excelize.ColumnNumberToName(c)
v, _ := f.GetCellValue(sheet, fmt.Sprintf("%s%d", colName, dataRow))
saved[c] = v
}

// Удаляем все шаблонные строки блока (сверху вниз).
for r := blk.EndRow - 1; r >= blk.StartRow+1; r-- {
if err := f.RemoveRow(sheet, r); err != nil {
return err
}
}

// После удаления: StartRow — на месте, шаблонные строки удалены, EndRow сместился.
insertAt := blk.StartRow + 1 // куда вставлять первую строку данных

// Вставляем n-1 пустых строк (первая совпадёт с insertAt сама, если бы мы писали в StartRow+1).
// Проще: вставляем n строк ВСЕГДА (или n-1, если место уже есть).
if n > 0 {
// Первая строка будет на insertAt. Сдвигаем всё ниже на n-1 строк вставкой.
if n > 1 {
if err := f.InsertRows(sheet, insertAt+1, n-1); err != nil {
return err
}
}
}

// Заполняем n строк начиная с insertAt.
for i := 0; i < n; i++ {
rowIdx := insertAt + i
it := pc.Items[0]
if i < len(pc.Items) {
it = pc.Items[i]
}
for c := 1; c <= maxCols; c++ {
v := saved[c]
if v == "" {
continue
}
nv := ReplacePlaceholders(v, func(expr string) string {
return renderItemExpr(expr, it)
})
colName, _ := excelize.ColumnNumberToName(c)
_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, rowIdx), nv)
}
}

// Теперь удаляем строку с <jx:forEach> (StartRow) и </jx:forEach> (сдвинут вниз).
// Строка с </jx:forEach> изначально была blk.EndRow.
// После удаления (EndRow - StartRow - 1) шаблонных строк и вставки n строк её позиция:
//   oldEnd - (blk.EndRow - blk.StartRow - 1) + (n - len(tmplRows))
// Упростим: ищем строку с </jx:forEach> заново.
_ = headerRows // пока не используем
if err := removeRowContaining(f, sheet, "</jx:forEach>"); err != nil {
return err
}
if err := removeRowContaining(f, sheet, "<jx:forEach"); err != nil {
return err
}
return nil
}

func removeRowContaining(f *excelize.File, sheet, substr string) error {
rows, err := f.GetRows(sheet)
if err != nil {
return err
}
for ri, row := range rows {
for _, cell := range row {
if strings.Contains(cell, substr) {
return f.RemoveRow(sheet, ri+1)
}
}
}
return nil
}

// replaceScalar — заменяет ${...} в одиночных ячейках.
func replaceScalar(f *excelize.File, sheet string, pc *PrintContext) error {
rows, err := f.GetRows(sheet)
if err != nil {
return err
}
for ri, row := range rows {
for ci, cell := range row {
if cell == "" || IsFormula(cell) {
continue
}
nv := ReplacePlaceholders(cell, func(expr string) string {
return renderScalarExpr(expr, pc)
})
if nv != cell {
col, _ := excelize.ColumnNumberToName(ci + 1)
_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, ri+1), nv)
}
}
}
return nil
}

// fixFormulas — заменяет $[SUM(X23)@0] на =SUM(X23:Y) с учётом n позиций.
// Конкретно: если в шаблоне было SUM(O23), а позиций n, то формула =SUM(O23:O(n+22)).
// Проще: SUM(O23) -> =SUM(O23:O23+n-1).
func fixFormulas(f *excelize.File, sheet string, blk *ForEachBlock, pc *PrintContext) error {
n := len(pc.Items)
if n == 0 {
n = 1
}
rows, err := f.GetRows(sheet)
if err != nil {
return err
}
baseRow := blk.StartRow + 1 // первая строка данных
lastRow := baseRow + n - 1
for ri, row := range rows {
for ci, cell := range row {
if !IsFormula(cell) {
continue
}
m := reSumFormula.FindStringSubmatch(cell)
if m == nil {
continue
}
col := m[1]
formula := fmt.Sprintf("=SUM(%s%d:%s%d)", col, baseRow, col, lastRow)
colName, _ := excelize.ColumnNumberToName(ci + 1)
_ = f.SetCellFormula(sheet, fmt.Sprintf("%s%d", colName, ri+1), formula)
}
}
return nil
}

// renderScalarExpr — подстановка ${o.X} / ${o.X.Y}.
func renderScalarExpr(expr string, pc *PrintContext) string {
switch expr {
case "o.name":
return pc.Doc.Number
case "o.moment":
return pc.DateStr
case "o.sourceStore.name":
return pc.Warehouse
case "o.targetStore.name":
return pc.TargetWarehouse
case "o.project.name":
return pc.Contract
case "o.sum.sum":
return formatRub(pc.Doc.Total)
case "o.sourceAgentRequisite.legalTitle",
"o.sourceAgentRequisite.agent.name":
return pc.Organization
}
if strings.Contains(expr, "printAmount") {
return formatRub(pc.Doc.Total)
}
if strings.Contains(expr, "getExcelDate") {
return pc.DateStr
}
if strings.Contains(expr, "printIfElse") {
return pc.Organization
}
return ""
}

// renderItemExpr — подстановка внутри цикла.
func renderItemExpr(expr string, it PrintItem) string {
e := expr
// отрезаем префикс вида `shipment.` (или любое имя переменной)
if idx := strings.Index(e, "."); idx > 0 {
e = e[idx+1:]
}
switch e {
case "printName":
return it.Name
case "consignment.modificationCode":
return it.SKU
case "consignment.good.uom.name":
return it.Unit
case "consignment.good.uom.code":
return it.UnitCode
case "quantity":
return formatQty(it.Quantity)
case "price.sum":
return formatRub(it.Price)
}
if strings.Contains(expr, "shipment.price.sum * shipment.quantity") ||
strings.Contains(expr, "round(") {
return formatRub(it.Sum)
}
return ""
}

func formatRub(v float64) string {
return strconv.FormatFloat(v, 'f', 2, 64)
}

func formatQty(v float64) string {
if v == float64(int64(v)) {
return strconv.FormatInt(int64(v), 10)
}
return strconv.FormatFloat(v, 'f', -1, 64)
}