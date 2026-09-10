package config

import (
"strings"

shcfg "github.com/radonezhsklad/shared/config"
)

type Config struct {
Port         string
DatabaseURL  string
JWTSecret    string
CORSOrigins  []string
WarehouseURL string
}

func Load() *Config {
return &Config{
Port:         shcfg.GetString("PORT", "8084"),
DatabaseURL:  shcfg.GetString("DATABASE_URL", "postgres://radonezh:radonezh_dev_pass@localhost:5433/radonezh_order?sslmode=disable"),
JWTSecret:    shcfg.GetString("JWT_SECRET", "dev-secret-change-me"),
CORSOrigins:  strings.Split(shcfg.GetString("CORS_ORIGINS", "*"), ","),
WarehouseURL: shcfg.GetString("WAREHOUSE_URL", "http://localhost:8083"),
}
}