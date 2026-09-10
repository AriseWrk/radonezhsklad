# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1. Окружение
- [x] Этап 2. Структура, docker-compose
- [x] Этап 3. Docker-инфра
- [x] Этап 4. Auth: миграция
- [x] Этап 5. Auth: Go-код :8081
- [x] Этап 6. Auth: edge-кейсы
- [x] Этап 7. Shared-пакет
- [x] Этап 8. Product :8082
- [x] Этап 9. Gateway :8080
- [x] Этап 10. Frontend :5173 (логин/дашборд/товары/категории)
- [x] Этап 11. Warehouse :8083 (склады, остатки, документы приёмки/отгрузки/перемещения/инвентаризации, проводка и отмена, audit)

## 🚧 В работе
- [ ] Этап 12. Order (заказы)

## ⏭️ Далее
- [ ] Frontend-страницы для склада
- [ ] Роли и права (RequireRole)

## ⚠️ Порты
- Postgres 5433, Redis 6380, RabbitMQ 5672 (UI :15672)
- auth 8081, product 8082, warehouse 8083, order 8084, gateway 8080, web 5173

## 📌 Контекст для продолжения
Стек: Go 1.22 + Gin, PostgreSQL 16, Vue 3 + TS + Vite + Pinia.
Корень: D:\Radonezhsklad\projects\radonezhsklad
ОС: Windows, PowerShell.
Терминалы:
1) services\auth      — auth :8081
2) services\product   — product :8082
3) services\warehouse — warehouse :8083
4) services\gateway   — gateway :8080
5) web                — npm run dev :5173
6) корень             — тесты

## 🔌 Warehouse API (:8083, через gateway :8080)
- GET|POST         /api/v1/warehouses
- GET|PUT|DELETE   /api/v1/warehouses/:id
- GET              /api/v1/stock?warehouse_id=&product_id=
- GET|POST         /api/v1/documents
- GET              /api/v1/documents/:id
- POST             /api/v1/documents/:id/post    — проводка (двигает остатки)
- POST             /api/v1/documents/:id/cancel  — отмена (откат)

Типы документов: receipt / shipment / transfer / inventory
Статусы: draft / posted / cancelled

## 🗄️ Схема radonezh_warehouse
- warehouses (id, name, address, is_active, created_at, updated_at)
- documents (id, type, number, status, warehouse_id, target_warehouse_id, comment, created_by, created_at, updated_at, posted_at, cancelled_at)
- document_items (id, document_id FK CASCADE, product_id, quantity, price)
- stock_balances (id, warehouse_id, product_id, quantity, updated_at) UNIQUE(warehouse_id, product_id)
- stock_movements (id, warehouse_id, product_id, document_id, quantity_delta, created_at) — audit

## 🔑 Учётные данные dev
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
- Тестовый пользователь: admin@radonezh.local / qwerty123