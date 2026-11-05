package printing

import (
"regexp"
"strings"

"github.com/xuri/excelize/v2"
)

var (
reForEachStart = regexp.MustCompile(`<jx:forEach\s+items="\$\{([^}]+)\}"\s+var="([^"]+)">`)
reForEachEnd   = regexp.MustCompile(`</jx:forEach>`)
rePlaceholder  = regexp.MustCompile(`\$\{([^}]+)\}`)
reMrpHolder    = regexp.MustCompile(`\{\{\s*\.?([A-Za-z0-9_.]+)\s*\}\}`)
reSumFormula   = regexp.MustCompile(`\$\[SUM\(([A-Z]+)(\d+)\)@0\]`)
)

// ForEachBlock — блок цикла в шаблоне.
type ForEachBlock struct {
StartRow     int      // строка с <jx:forEach>
EndRow       int      // строка с </jx:forEach>
ItemsExpr    string   // например "o.positions"
Var          string   // например "shipment"
TemplateRows []int    // строки данных между StartRow и EndRow (1-based)
}

// FindForEach — ищет блок <jx:forEach> на листе. Возвращает nil, если блока нет.
func FindForEach(f *excelize.File, sheet string) (*ForEachBlock, error) {
rows, err := f.GetRows(sheet)
if err != nil {
return nil, err
}
var blk *ForEachBlock
for ri, row := range rows {
for _, cell := range row {
if blk == nil {
if m := reForEachStart.FindStringSubmatch(cell); m != nil {
blk = &ForEachBlock{StartRow: ri + 1, ItemsExpr: m[1], Var: m[2]}
}
continue
}
if reForEachEnd.MatchString(cell) {
blk.EndRow = ri + 1
for r := blk.StartRow + 1; r < blk.EndRow; r++ {
blk.TemplateRows = append(blk.TemplateRows, r)
}
return blk, nil
}
}
}
return nil, nil
}

// ReplacePlaceholders — заменяет ${...} и {{...}} через callback.
func ReplacePlaceholders(text string, fn func(expr string) string) string {
text = rePlaceholder.ReplaceAllStringFunc(text, func(s string) string {
m := rePlaceholder.FindStringSubmatch(s)
if len(m) < 2 {
return s
}
return fn(m[1])
})
text = reMrpHolder.ReplaceAllStringFunc(text, func(s string) string {
m := reMrpHolder.FindStringSubmatch(s)
if len(m) < 2 {
return s
}
return fn(m[1])
})
return text
}

// IsFormula — ячейка содержит формулу МС $[SUM(...)@0].
func IsFormula(text string) bool {
return strings.HasPrefix(strings.TrimSpace(text), "$[")
}

// ConvertFormula — конвертирует $[SUM(O23)@0] в =SUM(O23), т.е. Excel-совместимую.
// @0 в МС-формате — это «начальная строка», диапазон до текущей.
// Простая реализация: SUM(O23) -> SUM(O23:O23).
func ConvertFormula(text string) string {
m := reSumFormula.FindStringSubmatch(text)
if m == nil {
return text
}
col, row := m[1], m[2]
return "=SUM(" + col + row + ":" + col + row + ")"
}