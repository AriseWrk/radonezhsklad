# shared

Общие пакеты для всех сервисов RadonezhSklad.

## Пакеты

- **errors** — типизированные ошибки с HTTP-статусом и кодом.
- **logger** — настройка `log/slog` (dev — текст, prod — JSON).
- **middleware** — RequestID, Recovery, AccessLog, CORS, ErrorHandler для Gin.
- **httpx** — единый формат ответов: OK / Created / Error / Validation.
- **config** — хелперы для чтения env-переменных.

## Подключение в сервисе

В `go.mod` сервиса:

    require github.com/radonezhsklad/shared v0.0.0
    replace github.com/radonezhsklad/shared => ../../shared

Импорты:

    import (
        apperr "github.com/radonezhsklad/shared/errors"
        "github.com/radonezhsklad/shared/logger"
        mw "github.com/radonezhsklad/shared/middleware"
        "github.com/radonezhsklad/shared/httpx"
    )