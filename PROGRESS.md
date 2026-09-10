# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1. Установка окружения
- [x] Этап 2. Структура проекта, docker-compose, .env.example, README
- [x] Этап 3. Docker-инфраструктура запущена. 3 контейнера healthy. 4 БД созданы.
- [x] Этап 4. Auth: миграция 0001_init применена (users, refresh_tokens, триггер updated_at)

## 🚧 В работе
- [ ] Этап 5. Auth-сервис: Go-код (config, db, models, repo, service, handler, main)

## ⏭️ Далее по плану
- [ ] Этап 6. Запуск auth-сервиса и тест через curl
- [ ] Этап 7. Shared-пакет (logger, middleware, errors)
- [ ] Этап 8. Product-сервис
- [ ] Этап 9. API Gateway
- [ ] Этап 10. Frontend Vue 3

## ⚠️ ВАЖНО: сменённые порты
- Postgres: 5433 (внутри контейнера 5432)
- Redis:    6380 (внутри контейнера 6379)
- RabbitMQ: 5672, UI http://localhost:15672 (radonezh / radonezh_dev_pass)

## 📌 Контекст для продолжения
Стек: Go 1.22 + Gin, PostgreSQL 16, Redis 7, RabbitMQ 3.13, Vue 3 + TS.
Корень проекта: D:\Radonezhsklad\projects\radonezhsklad
ОС: Windows, PowerShell. Все команды — из корня проекта.
Порты сервисов: auth 8081, product 8082, warehouse 8083, order 8084, gateway 8080, web 5173.

## 🗄️ Схема БД (radonezh_auth)
- users (id UUID PK, email UNIQUE, password_hash, full_name, role, is_active, created_at, updated_at)
- refresh_tokens (id UUID PK, user_id FK->users CASCADE, token_hash, expires_at, created_at)
- триггер trg_users_updated_at: BEFORE UPDATE ON users -> set_updated_at()

## 🔑 Учётные данные dev-окружения
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
- Redis:    localhost:6380 (без пароля)
- RabbitMQ: radonezh / radonezh_dev_pass (localhost:5672, UI :15672)