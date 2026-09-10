# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1. Установка окружения (VS Code, Git, Go, Node, Docker)
- [x] Этап 2. Создана структура проекта, docker-compose, .env.example, README

## 🚧 В работе
- [ ] Этап 3. Запуск docker compose up -d и проверка сервисов

## ⏭️ Далее по плану
- [ ] Этап 4. Auth-сервис: миграции БД
- [ ] Этап 5. Auth-сервис: Go-код (config, db, models, repo, service, handler, main)
- [ ] Этап 6. Запуск auth-сервиса и тест через curl
- [ ] Этап 7. Shared-пакет (logger, middleware, errors)
- [ ] Этап 8. Product-сервис
- [ ] Этап 9. API Gateway
- [ ] Этап 10. Frontend Vue 3

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
- Postgres: radonezh / radonezh_dev_pass (localhost:5432)
- Redis: localhost:6379 (без пароля)
- RabbitMQ: radonezh / radonezh_dev_pass (localhost:5672, UI :15672)