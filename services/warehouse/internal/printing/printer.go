package printing

import (
"fmt"
"os"
"path/filepath"
"strings"

"github.com/radonezhsklad/warehouse/internal/models"
)

// Printer — фасад для печати.
type Printer struct {
TemplatesDir string
}

func New(templatesDir string) *Printer {
return &Printer{TemplatesDir: templatesDir}
}

// Render — рендерит документ по шаблону.
func (p *Printer) Render(doc *models.Document, templateName, wh, twh, org string) ([]byte, error) {
if strings.Contains(templateName, "..") || strings.ContainsAny(templateName, "/\\") {
return nil, fmt.Errorf("invalid template name")
}
tpl := filepath.Join(p.TemplatesDir, templateName)
if _, err := os.Stat(tpl); err != nil {
return nil, fmt.Errorf("template not found: %s", templateName)
}
pc := BuildContext(doc, wh, twh, org, "")
return Render(tpl, pc)
}