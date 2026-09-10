# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1-9. Окружение, инфра, shared, auth, product, gateway
- [x] Этап 10. Frontend Vue 3 на :5173
- [x] Этап 11. Warehouse :8083 (склады, остатки, документы, проводка)
- [x] Этап 12. Order :8084 (покупатели, заказы, статусы, интеграция с warehouse при отгрузке)

## 🚧 В работе
- [ ] Этап 13. Frontend-страницы для склада и заказов

## ⏭️ Далее
- [ ] Роли и права (RequireRole)
- [ ] Уведомления через RabbitMQ
- [ ] Отчёты

## ⚠️ Порты
- Postgres 5433, Redis 6380, RabbitMQ 5672
- auth 8081, product 8082, warehouse 8083, order 8084, gateway 8080, web 5173

## 📌 Контекст для продолжения
Стек: Go 1.22 + Gin, PostgreSQL 16, Vue 3 + TS + Vite + Pinia.
Корень: D:\Radonezhsklad\projects\radonezhsklad
ОС: Windows, PowerShell.
Терминалы:
1) services\auth      — auth :8081
2) services\product   — product :8082
3) services\warehouse — warehouse :8083
4) services\order     — order :8084
5) services\gateway   — gateway :8080
6) web                — npm run dev :5173
7) корень             — тесты

## 🔌 Order API (:8084, через gateway :8080)
- GET|POST        /api/v1/customers
- GET|DELETE      /api/v1/customers/:id
- GET|POST        /api/v1/orders
- GET             /api/v1/orders/:id
- POST            /api/v1/orders/:id/confirm  — подтвердить
- POST            /api/v1/orders/:id/ship     — отгрузить (создаёт shipment в warehouse)
- POST            /api/v1/orders/:id/cancel   — отменить

Статусы: draft / confirmed / shipped / cancelled

## 🗄️ Схема radonezh_order
- customers (id, name, phone, email, address, created_at, updated_at)
- orders (id, number UNIQUE, customer_id, warehouse_id, status, total, currency, comment, warehouse_doc_id, created_by, created_at, updated_at, confirmed_at, shipped_at, cancelled_at)
- order_items (id, order_id FK CASCADE, product_id, quantity, price, created_at)

## 🔗 Межсервисное взаимодействие
Order -> Warehouse (HTTP, прямой вызов на :8083, минуя gateway):
  POST /api/v1/documents (type=shipment) + POST /api/v1/documents/:id/post
  JWT пробрасывается от клиента.

## 🔑 Учётные данные dev
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
- Тестовый пользователь: admin@radonezh.local / qwerty123