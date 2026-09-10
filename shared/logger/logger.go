package logger

import (
"log/slog"
"os"
)

// Init настраивает глобальный slog-логгер.
// env = "dev" -> текстовый вывод с Debug-уровнем
// env = "prod" -> JSON с Info-уровнем
func Init(env string) *slog.Logger {
var handler slog.Handler
opts := &slog.HandlerOptions{}

if env == "prod" {
opts.Level = slog.LevelInfo
handler = slog.NewJSONHandler(os.Stdout, opts)
} else {
opts.Level = slog.LevelDebug
handler = slog.NewTextHandler(os.Stdout, opts)
}

l := slog.New(handler)
slog.SetDefault(l)
return l
}

// With — короткий хелпер для добавления полей.
func With(args ...any) *slog.Logger {
return slog.Default().With(args...)
}