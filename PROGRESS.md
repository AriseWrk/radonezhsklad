# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1. Окружение
- [x] Этап 2. Структура, docker-compose
- [x] Этап 3. Docker-инфра: postgres/redis/rabbitmq
- [x] Этап 4. Auth: миграция 0001
- [x] Этап 5. Auth: Go-код на :8081
- [x] Этап 6. Auth: edge-кейсы
- [x] Этап 7. Shared-пакет
- [x] Этап 8. Product: CRUD на :8082
- [x] Этап 9. Gateway на :8080: прокси на auth и product, X-Request-ID, 502 при падении upstream

## 🚧 В работе
- [ ] Этап 10. Frontend Vue 3

## ⏭️ Далее по плану
- [ ] Этап 11. Warehouse
- [ ] Этап 12. Order

## ⚠️ Порты
- Postgres 5433, Redis 6380, RabbitMQ 5672 (UI :15672)
- auth :8081, product :8082, warehouse :8083, order :8084, gateway :8080, web :5173

## 📌 Контекст для продолжения
Стек: Go 1.22 + Gin, PostgreSQL 16, Redis 7, RabbitMQ 3.13, Vue 3 + TS.
Корень: D:\Radonezhsklad\projects\radonezhsklad
ОС: Windows, PowerShell.
Терминалы для работы:
1) services\auth — auth :8081
2) services\product — product :8082
3) services\gateway — gateway :8080
4) корень — тесты

## 📦 Shared (github.com/radonezhsklad/shared)
- errors: AppError, NotFound/Conflict/BadRequest/Unauthorized/Forbidden/Internal
- logger: Init(env)
- middleware: RequestID, Recovery, AccessLog, CORS, ErrorHandler, RequireJWT(secret), RequireRole, CurrentUserID, CurrentRole
- httpx: OK, Created, NoContent, Error, Validation
- config: GetString, GetInt, GetBool, GetStrings

## 🌐 Gateway (:8080)
- /api/v1/auth/*        -> :8081
- /api/v1/units*        -> :8082
- /api/v1/categories*   -> :8082
- /api/v1/products*     -> :8082
- /api/v1/warehouse*    -> :8083 (сервис ещё не создан)
- /api/v1/orders*       -> :8084 (сервис ещё не создан)
- /api/v1/health        -> отвечает gateway
Особенности: X-Request-ID прокидывается, клиентские X-Forwarded-* удаляются, 502 при недоступном upstream.

## 🔌 API
- POST /api/v1/auth/register | /api/v1/auth/login
- GET  /api/v1/auth/me
- GET  /api/v1/units
- CRUD /api/v1/categories[/:id]
- CRUD /api/v1/products[/:id] (+ ?include_archived=true&category_id=<uuid>)

## 🔑 Учётные данные dev
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
- Redis:    localhost:6380
- RabbitMQ: radonezh / radonezh_dev_pass (localhost:5672)
- Тестовый пользователь: admin@radonezh.local / qwerty123