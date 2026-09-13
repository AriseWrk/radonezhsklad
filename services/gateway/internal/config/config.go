package config

import (
"strings"

shcfg "github.com/radonezhsklad/shared/config"
)

type Config struct {
Port          string
AuthURL       string
ProductURL    string
WarehouseURL  string
OrderURL      string
CORSOrigins   []string
AuditURL      string
InternalToken string
}

func Load() *Config {
return &Config{
Port:          shcfg.GetString("PORT", "8080"),
AuthURL:       shcfg.GetString("AUTH_URL", "http://localhost:8081"),
ProductURL:    shcfg.GetString("PRODUCT_URL", "http://localhost:8082"),
WarehouseURL:  shcfg.GetString("WAREHOUSE_URL", "http://localhost:8083"),
OrderURL:      shcfg.GetString("ORDER_URL", "http://localhost:8084"),
CORSOrigins:   strings.Split(shcfg.GetString("CORS_ORIGINS", "*"), ","),
AuditURL:      shcfg.GetString("AUDIT_URL", "http://localhost:8085"),
InternalToken: shcfg.GetString("INTERNAL_TOKEN", "dev-audit-internal-token"),
}
}