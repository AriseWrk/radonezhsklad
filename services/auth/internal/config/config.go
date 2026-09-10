package config

import (
"os"
"strconv"
)

type Config struct {
Port          string
DatabaseURL   string
JWTSecret     string
AccessTTLMin  int
RefreshTTLDay int
}

func Load() *Config {
return &Config{
Port:          getEnv("PORT", "8081"),
DatabaseURL:   getEnv("DATABASE_URL", "postgres://radonezh:radonezh_dev_pass@localhost:5433/radonezh_auth?sslmode=disable"),
JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
AccessTTLMin:  getEnvInt("ACCESS_TTL_MIN", 15),
RefreshTTLDay: getEnvInt("REFRESH_TTL_DAY", 30),
}
}

func getEnv(key, def string) string {
if v := os.Getenv(key); v != "" {
return v
}
return def
}

func getEnvInt(key string, def int) int {
if v := os.Getenv(key); v != "" {
if n, err := strconv.Atoi(v); err == nil {
return n
}
}
return def
}