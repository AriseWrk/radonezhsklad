package config

import (
"os"
"strconv"
)

// GetString читает переменную окружения или возвращает default.
func GetString(key, def string) string {
if v := os.Getenv(key); v != "" {
return v
}
return def
}

// GetInt — то же, но для int.
func GetInt(key string, def int) int {
if v := os.Getenv(key); v != "" {
if n, err := strconv.Atoi(v); err == nil {
return n
}
}
return def
}

// GetBool — то же, но для bool.
func GetBool(key string, def bool) bool {
if v := os.Getenv(key); v != "" {
if b, err := strconv.ParseBool(v); err == nil {
return b
}
}
return def
}

// GetStrings — парсит строку через запятую в срез.
func GetStrings(key, def string) []string {
raw := GetString(key, def)
if raw == "" {
return nil
}
parts := []string{}
for _, p := range splitAndTrim(raw, ',') {
if p != "" {
parts = append(parts, p)
}
}
return parts
}

func splitAndTrim(s string, sep rune) []string {
var out []string
var cur []rune
for _, ch := range s {
if ch == sep {
out = append(out, string(cur))
cur = nil
continue
}
cur = append(cur, ch)
}
out = append(out, string(cur))
// trim spaces
for i := range out {
out[i] = trimSpace(out[i])
}
return out
}

func trimSpace(s string) string {
start, end := 0, len(s)
for start < end && (s[start] == ' ' || s[start] == '\t') {
start++
}
for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
end--
}
return s[start:end]
}