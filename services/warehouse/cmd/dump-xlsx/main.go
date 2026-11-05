package main

import (
"fmt"
"os"
"strings"

"github.com/xuri/excelize/v2"
)

func main() {
if len(os.Args) < 2 {
fmt.Println("usage: dump-xlsx <file.xlsx>")
os.Exit(1)
}
f, err := excelize.OpenFile(os.Args[1])
if err != nil {
fmt.Println("open:", err)
os.Exit(1)
}
defer f.Close()
for _, sheet := range f.GetSheetList() {
fmt.Println("=== SHEET:", sheet, "===")
rows, err := f.GetRows(sheet)
if err != nil {
continue
}
for ri, row := range rows {
for ci, cell := range row {
if strings.TrimSpace(cell) == "" {
continue
}
col, _ := excelize.ColumnNumberToName(ci + 1)
fmt.Printf("%s%d: %q\n", col, ri+1, cell)
}
}
}
}
