# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1. Окружение
- [x] Этап 2. Структура, docker-compose
- [x] Этап 3. Docker-инфа: postgres/redis/rabbitmq
- [x] Этап 4. Auth: миграция 0001
- [x] Этап 5. Auth: Go-код на :8081
- [x] Этап 6. Auth: edge-кейсы
- [x] Этап 7. Shared-пакет
- [x] Этап 8. Product: CRUD на :8082
- [x] Этап 9. Gateway на :8080
- [x] Этап 10. Frontend Vue 3 на :5173: логин, layout, дашборд, товары, категории

## 🚧 В работе
- [ ] Этап 11. Warehouse (склад, остатки, документы)

## ⏭️ Далее по плану
- [ ] Этап 12. Order
- [ ] (полировка) роли и права (RequireRole на фронте и бэке)

## ⚠️ Порты
- Postgres 5433, Redis 6380, RabbitMQ 5672 (UI :15672)
- auth :8081, product :8082, warehouse :8083, order :8084, gateway :8080, web :5173

## 📌 Контекст для продолжения
Стек: Go 1.22 + Gin, PostgreSQL 16, Redis 7, RabbitMQ 3.13, Vue 3 + TS + Vite + Pinia + vue-router + axios.
Корень: D:\Radonezhsklad\projects\radonezhsklad
ОС: Windows, PowerShell.
Терминалы:
1) services\auth    — auth :8081
2) services\product — product :8082
3) services\gateway — gateway :8080
4) web              — npm run dev :5173
5) корень           — тесты/утилиты

## 📦 Shared (github.com/radonezhsklad/shared)
errors, logger, middleware (в т.ч. RequireJWT/RequireRole), httpx, config

## 🌐 Gateway (:8080)
/api/v1/auth/*        -> :8081
/api/v1/units*        -> :8082
/api/v1/categories*   -> :8082
/api/v1/products*     -> :8082
/api/v1/warehouse*    -> :8083 (не создан)
/api/v1/orders*       -> :8084 (не создан)

## 🖥️ Frontend (web/)
- Vite + Vue 3 + TS + Pinia + vue-router + axios
- Прокси: /api -> http://localhost:8080
- localStorage: rs_access_token
- Маршруты: /login, / (dashboard), /products, /categories
- Layout: сайдбар + топбар + logout
- api/client.ts: axios-инстанс с Authorization и 401→/login
- stores/auth.ts: token, login(), logout(), fetchMe()
- views: Login, Dashboard (счётчики), Products (список+создание+архив), Categories (список+создание+удаление)

## 🔑 Учётные данные dev
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
- Redis:    localhost:6380
- RabbitMQ: radonezh / radonezh_dev_pass (localhost:5672)
- Тестовый пользователь: admin@radonezh.local / qwerty123