# RadonezhSklad — Прогресс разработки

## ✅ Завершено
- [x] Этап 1-9. Окружение, инфра, shared, auth, product, gateway
- [x] Этап 10. Frontend Vue 3: логин, дашборд, товары, категории
- [x] Этап 11. Warehouse :8083 (склады, остатки, документы, проводка)
- [x] Этап 12. Order :8084 (покупатели, заказы, интеграция с warehouse)
- [x] Этап 13. Frontend: склады, остатки, документы, покупатели, заказы

## 🚧 В работе
- [ ] Этап 14. Роли и права (RequireRole + UI)

## ⏭️ Далее
- [ ] Уведомления через RabbitMQ (брокер уже поднят)
- [ ] Отчёты (продажи за период, ABC-анализ)
- [ ] Печатные формы

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

## 🖥️ Frontend-страницы (web/src/views/)
- LoginView, DashboardView
- ProductsView, CategoriesView, CustomersView
- WarehousesView, StockView, DocumentsView
- OrdersView (с подтверждением/отгрузкой/отменой, создание документа в warehouse при отгрузке)

## 🔌 API через gateway (:8080)
- /api/v1/auth/*
- /api/v1/units, /api/v1/categories*, /api/v1/products*
- /api/v1/warehouses*, /api/v1/stock*, /api/v1/documents*
- /api/v1/customers*, /api/v1/orders*

## 🔗 Межсервисное взаимодействие
Order -> Warehouse (HTTP на :8083, минуя gateway):
  POST /api/v1/documents (type=shipment) + POST /api/v1/documents/:id/post
  JWT пробрасывается.

## 🔑 Учётные данные dev
- Postgres: radonezh / radonezh_dev_pass (localhost:5433)
- Тестовый пользователь: admin@radonezh.local / qwerty123