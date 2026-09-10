# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1. Установка окружения
- [x] Этап 2. Структура проекта, docker-compose, .env.example, README
- [x] Этап 3. Docker-инфраструктура запущена. 3 контейнера healthy. 4 БД созданы.
- [x] Этап 4. Auth: миграция 0001_init применена
- [x] Этап 5. Auth: Go-код написан, сервис работает на :8081
- [x] Этап 6. Auth: edge-кейсы проверены (409/401/400), git-commit

## 🚧 В работе
- [ ] Этап 7. Shared-пакет (logger, middleware, errors, http-client)

## ⏭️ Далее по плану
- [ ] Этап 8. Product-сервис (товары, категории, единицы измерения, цены)
- [ ] Этап 9. API Gateway (единая точка входа :8080)
- [ ] Этап 10. Frontend Vue 3
- [ ] Этап 11. Warehouse-сервис (склад, остатки, документы)
- [ ] Этап 12. Order-сервис

## ⚠️ Порты
- Postgres: 5433 (внутри 5432)
- Redis:    6380 (внутри 6379)
- RabbitMQ: 5672, UI http://localhost:15672 (radonezh / radonezh_dev_pass)

## 📌 Контекст для продолжения
Стек: Go 1.22 + Gin, PostgreSQL 16, Redis 7, RabbitMQ 3.13, Vue 3 + TS.
Корень: D:\Radonezhsklad\projects\radonezhsklad
ОС: Windows, PowerShell. Все команды — из корня проекта.
Порты сервисов: auth 8081, product 8082, warehouse 8083, order 8084, gateway 8080, web 5173.

## 🔌 API auth-сервиса
- GET  /api/v1/health               -> {status, service}
- POST /api/v1/auth/register        -> 201 User | 400 (валидация) | 409 (email exists)
- POST /api/v1/auth/login           -> 200 TokenPair | 401
- GET  /api/v1/auth/me              -> 200 {user_id, role} | 401

## 🗄️ Схема БД radonezh_auth
- users (id, email UNIQUE, password_hash, full_name, role, is_active, created_at, updated_at)
- refresh_tokens (id, user_id FK->users CASCADE, token_hash, expires_at, created_at)
- триггер trg_users_updated_at

## 🛠️ Структура auth-сервиса
- cmd/api/main.go
- internal/config/config.go
- internal/db/db.go
- internal/models/user.go, refresh_token.go
- internal/repository/user_repo.go, token_repo.go
- internal/service/auth_service.go
- internal/handler/auth_handler.go
- internal/middleware/auth.go

## 🔑 Учётные данные dev
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
- Redis:    localhost:6380
- RabbitMQ: radonezh / radonezh_dev_pass (localhost:5672)
- Тестовый пользователь: admin@radonezh.local / qwerty123