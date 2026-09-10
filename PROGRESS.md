# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1. Установка окружения (VS Code, Git, Go, Node, Docker)
- [x] Этап 2. Создана структура проекта, docker-compose, .env.example, README
- [x] Этап 3. Docker-инфраструктура запущена. 3 контейнера healthy. 4 БД созданы.

## 🚧 В работе
- [ ] Этап 4. Auth-сервис: миграции БД

## ⏭️ Далее по плану
- [ ] Этап 5. Auth-сервис: Go-код (config, db, models, repo, service, handler, main)
- [ ] Этап 6. Запуск auth-сервиса и тест через curl
- [ ] Этап 7. Shared-пакет (logger, middleware, errors)
- [ ] Этап 8. Product-сервис
- [ ] Этап 9. API Gateway
- [ ] Этап 10. Frontend Vue 3

## ⚠️ ВАЖНО: сменённые порты (из-за локальных служб Windows)
- Postgres: НЕ 5432, а 5433  (внутри контейнера 5432)
- Redis:    НЕ 6379, а 6380  (внутри контейнера 6379)
- RabbitMQ: 5672 (без изменений)
- RabbitMQ UI: http://localhost:15672 (radonezh / radonezh_dev_pass)

Причина: порт 5432 занят нативным postgres.exe, порт 6379 — WSL-релеем Docker.

## 📌 Контекст для продолжения
Стек: Go 1.22 + Gin, PostgreSQL 16, Redis 7, RabbitMQ 3.13, Vue 3 + TS.
Корень проекта: D:\Radonezhsklad\projects\radonezhsklad
ОС разработки: Windows, PowerShell
Все команды — из корня проекта.

## 📂 Структура
- services/auth — аутентификация (порт 8081)
- services/product — товары (порт 8082)
- services/warehouse — склад (порт 8083)
- services/order — заказы (порт 8084)
- services/gateway — API Gateway (порт 8080)
- web — Vue 3 фронтенд (порт 5173)

## 🔑 Учётные данные dev-окружения
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
  - Базы: radonezh_auth, radonezh_product, radonezh_warehouse, radonezh_order
- Redis:    localhost:6380 (без пароля)
- RabbitMQ: radonezh / radonezh_dev_pass (localhost:5672, UI :15672)