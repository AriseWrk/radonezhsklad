package config

import (
"strings"

shcfg "github.com/radonezhsklad/shared/config"
)

type Config struct {
Port        string
DatabaseURL string
JWTSecret   string
CORSOrigins []string
ProductURL  string

// --- МойСклад (обратная синхронизация: push наших документов в МС) ---
// MS_PUSH_ENABLED = true — включаем пуш; по умолчанию выкл.
MSPushEnabled bool
MSAPIBase     string
MSToken       string
MSTokenFile   string

// Путь к папке scripts/ (для запуска PS-скриптов синхронизации).
ScriptsDir string

// Директория для шаблонов печати и загруженных файлов.
UploadsDir string
}

func Load() *Config {
return &Config{
Port:        shcfg.GetString("PORT", "8083"),
DatabaseURL: shcfg.GetString("DATABASE_URL", "postgres://radonezh:radonezh_dev_pass@localhost:5433/radonezh_warehouse?sslmode=disable"),
JWTSecret:   shcfg.GetString("JWT_SECRET", "dev-secret-change-me"),
CORSOrigins: strings.Split(shcfg.GetString("CORS_ORIGINS", "*"), ","),
ProductURL:  shcfg.GetString("PRODUCT_URL", "http://localhost:8082"),

MSPushEnabled: shcfg.GetBool("MS_PUSH_ENABLED", false),
MSAPIBase:     shcfg.GetString("MS_API_BASE", "https://api.moysklad.ru/api/remap/1.2"),
MSToken:       shcfg.GetString("MS_TOKEN", ""),
MSTokenFile:   shcfg.GetString("MS_TOKEN_FILE", ""),

ScriptsDir: shcfg.GetString("SCRIPTS_DIR", "../../scripts"),
UploadsDir: shcfg.GetString("UPLOADS_DIR", "uploads/templates"),
}
}