# RadonezhSklad — Прогресс разработки

## ✅ Завершено

### Инфраструктура (Этапы 1-3)
- [x] VS Code, Git, Go 1.22, Node, Docker
- [x] Структура монорепо: services/, shared/, web/, infra/, docs/
- [x] Docker: postgres:16 (5433), redis:7 (6380), rabbitmq:3.13 (5672, UI :15672)
- [x] 4 БД: radonezh_auth, radonezh_product, radonezh_warehouse, radonezh_order

### Auth-сервис :8081 (Этапы 4-6, 14, 16)
- [x] Миграции 0001 (users, refresh_tokens), 0002 (роли), 0003 (ФИО/телефон/логин/описание)
- [x] bcrypt, JWT HS256, refresh-токены
- [x] API: register, login, me, GET/POST /users, PUT /users/:id, PUT /users/:id/role, PUT /users/:id/active
- [x] Роли: admin / manager / warehouse / user

### Shared-пакет (Этап 7)
- [x] shared/errors      — AppError + конструкторы
- [x] shared/logger      — slog (dev-текст / prod-JSON)
- [x] shared/middleware  — RequestID, Recovery, AccessLog, CORS, ErrorHandler, RequireJWT, RequireRole, CurrentUserID
- [x] shared/httpx       — OK, Created, NoContent, Error, Validation
- [x] shared/config      — GetString / GetInt / GetBool / GetStrings

### Product-сервис :8082 (Этапы 8, 15)
- [x] Миграции 0001 (categories, units, products), 0002 (min_stock, cost_price)
- [x] API: units, categories CRUD, products CRUD (soft-archive)
- [x] /products/list?ids= для внутренних вызовов
- [x] Роли: write — admin/manager, read — все авторизованные

### Warehouse-сервис :8083 (Этапы 11, 15)
- [x] Миграция 0001 (warehouses, documents, document_items, stock_balances, stock_movements)
- [x] API: склады CRUD, остатки, документы (receipt/shipment/transfer/inventory), проводка/отмена
- [x] /stock/extended — агрегация с product-сервисом (min_stock, cost_price, incoming из черновиков)
- [x] product-клиент (HTTP на :8082) — internal/product/client.go
- [x] Роли: склады — admin/warehouse, документы — admin/manager/warehouse

### Order-сервис :8084 (Этап 12)
- [x] Миграция 0001 (customers, orders, order_items)
- [x] API: покупатели CRUD, заказы CRUD, confirm/ship/cancel
- [x] Интеграция с warehouse: ship создаёт и проводит shipment
- [x] Роли: write — admin/manager

### Gateway :8080 (Этап 9)
- [x] Reverse proxy на auth / product / warehouse / order
- [x] X-Request-ID прокидывается, клиентские X-Forwarded-* удаляются
- [x] 502 при недоступном upstream
- [x] /api/v1/users → auth

### Frontend :5173 (Этапы 10, 13, 14, 15, 16)
- [x] Vite + Vue 3 + TS + Pinia + vue-router + axios
- [x] Прокси /api → :8080
- [x] MainLayout в стиле МойСклад: синий топбар + подтабы + пользователь справа
- [x] Раздел «Компания»: Показатели, Сотрудники
- [x] Раздел «Товары»: Товары, Категории
- [x] Раздел «Закупки»: Документы, Склады
- [x] Раздел «Продажи»: Заказы, Покупатели
- [x] Раздел «Склад»: Остатки, Документы, Склады
- [x] Страницы: Login, Dashboard, Users (Сотрудники в стиле МойСклад), Products,
      Categories, Customers, Warehouses, Stock (14 колонок + фильтры + футер),
      Documents, Orders
- [x] RBAC на фронте: meta.roles в роутере + auth.can() в меню + бейдж роли
- [x] CSS-классы ms-table, filter-panel, page-title-bar, view-switch, ms-footer

## 🚧 В работе
- [ ] Тонкая настройка страницы «Сотрудники» (данные заполняются через UI)

## ⏭️ Возможные следующие этапы
- [ ] Резерв в остатках (интеграция с order через HTTP)
- [ ] Карточка товара с полным редактированием (min_stock, cost_price)
- [ ] Отчёты: продажи за период, ABC-анализ, движение по складу
- [ ] Печатные формы: накладная, счёт, ТОРГ-12
- [ ] Экспорт / импорт CSV
- [ ] Уведомления через RabbitMQ
- [ ] Группировка товаров по категориям в остатках
- [ ] Поставщики и приходные документы от них

## ⚠️ Порты
| Сервис     | Порт | Примечание         |
| ---------- | ---- | ------------------ |
| postgres   | 5433 | внутри 5432        |
| redis      | 6380 | внутри 6379        |
| rabbitmq   | 5672 | UI :15672          |
| auth       | 8081 |                    |
| product    | 8082 |                    |
| warehouse  | 8083 |                    |
| order      | 8084 |                    |
| gateway    | 8080 | единая точка входа |
| web (Vite) | 5173 |                    |

## 📌 Контекст для продолжения
- Стек: Go 1.22 + Gin, PostgreSQL 16, Vue 3 + TS + Vite + Pinia + axios
- Корень: `D:\Radonezhsklad\projects\radonezhsklad`
- ОС: Windows, PowerShell 5.1 (важно: файлы через `[IO.File]::WriteAllText` с `UTF8Encoding $false`)
- Терминалы:
  1. `services\auth`      — `go run .\cmd\api` (auth :8081)
  2. `services\product`   — `go run .\cmd\api` (product :8082)
  3. `services\warehouse` — `go run .\cmd\api` (warehouse :8083)
  4. `services\order`     — `go run .\cmd\api` (order :8084)
  5. `services\gateway`   — `go run .\cmd\api` (gateway :8080)
  6. `web`                — `npm run dev` (frontend :5173)
  7. корень               — тесты (PowerShell, сначала `. .\devtools.ps1`)

## ⚠️ Windows + PowerShell 5.1: работа с UTF-8

**Проблема:** PS 5.1 по умолчанию читает/пишет в CP1251 → русский бьётся в `?????`.

**Решение:** хелпер `devtools.ps1` в корне проекта.

Подключение в новой сессии:

    cd D:\Radonezhsklad\projects\radonezhsklad
    . .\devtools.ps1
    Set-ConsoleUtf8

Функции:

- `Write-Utf8File -Path <p> -Content <s>` — записать файл в UTF-8 без BOM
- `Read-Utf8File -Path <p>` — прочитать файл в UTF-8
- `Invoke-SqlFile -Path <p> -Database <db>` — применить SQL в UTF-8
- `Set-ConsoleUtf8` — переключить консоль в UTF-8

**Никогда не используйте:**

- `Get-Content file -Raw | docker exec ... psql` на файлах с русским текстом
- `-replace` в PS 5.1 для многострочных замен — вместо этого перезаписывайте файл целиком через `Write-Utf8File`

## 🔌 API через gateway (:8080)

Все эндпоинты требуют `Authorization: Bearer <jwt>` (кроме /health, /auth/register, /auth/login).

| Группа     | Эндпоинты                                                                              | Кто может писать             |
| ---------- | -------------------------------------------------------------------------------------- | ---------------------------- |
| Health     | GET /api/v1/health                                                                     | все                          |
| Auth       | POST /auth/register, POST /auth/login, GET /auth/me                                    | —                            |
| Users      | GET/POST /users, PUT /users/:id, PUT /users/:id/role, PUT /users/:id/active            | admin                        |
| Units      | GET /units                                                                             | —                            |
| Categories | GET/POST/PUT/DELETE /categories[/:id]                                                  | admin, manager               |
| Products   | GET/POST/PUT/DELETE /products[/:id], GET /products/list?ids=                            | admin, manager               |
| Warehouses | GET/POST/PUT/DELETE /warehouses[/:id]                                                  | admin, warehouse             |
| Stock      | GET /stock, GET /stock/extended                                                        | —                            |
| Documents  | GET/POST /documents, GET /documents/:id, POST /documents/:id/post, POST /documents/:id/cancel | admin, manager, warehouse    |
| Customers  | GET/POST/DELETE /customers[/:id]                                                       | admin, manager               |
| Orders     | GET/POST /orders, GET /orders/:id, POST /orders/:id/{confirm,ship,cancel}              | admin, manager               |

## 🗄️ Схемы БД

### radonezh_auth
- **users** — id, email UNIQUE, password_hash, full_name, last_name, first_name, middle_name,
  phone, login, description, role (CHECK admin/manager/warehouse/user), is_active, created_at, updated_at
- **refresh_tokens** — id, user_id FK CASCADE, token_hash, expires_at, created_at
- Триггер: `trg_users_updated_at`

### radonezh_product
- **categories** — id, name, parent_id FK SET NULL, created_at, updated_at
- **units** — id, code UNIQUE, name, short_name, created_at
- **products** — id, name, sku UNIQUE, barcode, category_id FK, unit_id FK RESTRICT,
  description, price, cost_price, min_stock, currency, is_archived, created_at, updated_at

### radonezh_warehouse
- **warehouses** — id, name, address, is_active, created_at, updated_at
- **documents** — id, type (receipt/shipment/transfer/inventory), number,
  status (draft/posted/cancelled), warehouse_id, target_warehouse_id, comment,
  created_by, created_at, updated_at, posted_at, cancelled_at
- **document_items** — id, document_id FK CASCADE, product_id, quantity, price
- **stock_balances** — id, warehouse_id, product_id, quantity, updated_at,
  UNIQUE(warehouse_id, product_id)
- **stock_movements** — id, warehouse_id, product_id, document_id, quantity_delta,
  created_at (audit)

### radonezh_order
- **customers** — id, name, phone, email, address, created_at, updated_at
- **orders** — id, number UNIQUE, customer_id FK, warehouse_id,
  status (draft/confirmed/shipped/cancelled), total, currency, comment,
  warehouse_doc_id, created_by, created_at, updated_at, confirmed_at,
  shipped_at, cancelled_at
- **order_items** — id, order_id FK CASCADE, product_id, quantity, price

## 🔗 Межсервисное взаимодействие
- **warehouse → product** (HTTP :8082, GET /products/list?ids=..., GET /units):
  для расширенной таблицы остатков
- **order → warehouse** (HTTP :8083, POST /documents + POST /documents/:id/post):
  при отгрузке заказа создаётся и проводится документ shipment
- JWT пробрасывается от клиента вниз по цепочке

## 🔑 Учётные данные dev
- Postgres: `radonezh / radonezh_dev_pass` (localhost:5433)
- Redis: `localhost:6380` (без пароля)
- RabbitMQ: `radonezh / radonezh_dev_pass` (localhost:5672, UI http://localhost:15672)
- Главный админ: `admin@radonezh.local / qwerty123`
- Остальные создаются через UI: **Компания → Сотрудники → 👤 Сотрудник**

## 📂 Структура проекта

    radonezhsklad/
    ├── docker-compose.yml
    ├── devtools.ps1              ← хелпер UTF-8
    ├── PROGRESS.md               ← этот файл
    ├── README.md
    ├── .gitignore
    ├── infra/postgres/init/01-databases.sql
    ├── services/
    │   ├── auth/
    │   │   ├── cmd/api/main.go
    │   │   ├── internal/{config,db,models,repository,service,handler,middleware}/
    │   │   └── migrations/{0001_init,0002_roles,0003_employee_fields}.sql
    │   ├── product/
    │   │   ├── cmd/api/main.go
    │   │   ├── internal/{config,db,models,repository,service,handler}/
    │   │   └── migrations/{0001_init,0002_min_cost}.sql
    │   ├── warehouse/
    │   │   ├── cmd/api/main.go
    │   │   ├── internal/{config,db,models,repository,service,handler,product}/
    │   │   │   ├── handler/handler.go
    │   │   │   ├── handler/stock_extended.go
    │   │   │   └── product/client.go
    │   │   └── migrations/0001_init.sql
    │   ├── order/
    │   │   ├── cmd/api/main.go
    │   │   ├── internal/{config,db,models,repository,service,handler,warehouse}/
    │   │   │   └── warehouse/client.go
    │   │   └── migrations/0001_init.sql
    │   └── gateway/
    │       ├── cmd/api/main.go
    │       └── internal/{config,proxy}/
    ├── shared/
    │   ├── errors/
    │   ├── logger/
    │   ├── middleware/  (middleware.go, jwt.go)
    │   ├── httpx/
    │   └── config/
    └── web/
        └── src/
            ├── api/     (client, auth, users, units, categories, products,
            │             warehouses, stock, documents, customers, orders)
            ├── stores/  (auth)
            ├── router/  (index)
            ├── layouts/ (MainLayout)
            └── views/   (Login, Dashboard, Users, Products, Categories,
                          Customers, Warehouses, Stock, Documents, Orders)
## Этап 17. Страница «Документы» в стиле МойСклад
- [x] Backend: ListDocuments и GetDocument теперь возвращают `items_count` и `total`
- [x] Frontend: DocumentsView переписана — вкладки типов/статусов с счётчиками,
      расширенная панель фильтров (период, склад, статус, поиск, суммы),
      таблица ms-table с колонками №/Дата/Тип/Склад/Комментарий/Позиций/Сумма/Статус/Действия,
      пагинация, футер с итогами
- [x] Стили `.doc-tabs`, `.doc-tab`, `.tab-count`, `.doc-info`, `.actions-col` добавлены в style.css
## Этап 18. Инвентаризация и подтабы склада
- [x] Backend: GET /api/v1/inventory/prepare?warehouse_id= — список товаров на складе с учётным остатком
- [x] Backend: BookStockForInventory в repository/service, маршрут в warehouse и gateway
- [x] Frontend: InventoryView.vue — выбор склада, таблица товаров, ввод фактов, подсветка расхождений, создание inventory-документа
- [x] Frontend: маршрут /inventory с ролями admin/manager/warehouse
- [x] Frontend: подтабы Склад и Закупки разбиты по типам документов (Оприходования/Списания/Перемещения/Инвентаризации/Остатки/Склады)
- [x] Frontend: DocumentsView принимает ?type= из URL
## Этап 19. Аудит действий
- [x] Отдельный сервис audit :8085, БД radonezh_audit, таблица audit_logs
- [x] API: POST /internal/audit (внутренний приём с X-Internal-Token), GET /audit (только admin)
- [x] Gateway middleware: логирует все POST/PUT/PATCH/DELETE с кодом 2xx, асинхронно шлёт в audit
- [x] Gateway: UserContext парсит JWT и кладёт user_id в контекст (для аудита)
- [x] Frontend: AuditView.vue — таблица событий, фильтры (метод/ресурс/период), пагинация, модалка деталей с JSON
- [x] Подтаб «Аудит» в разделе «Компания» (только admin)
## Этап 20. Закупки → Приёмки
- [x] Миграция 0002: таблицы suppliers, organizations, + 6 полей в documents (supplier_id, organization_id, incoming_number, incoming_date, paid_amount, printed_at, sent_at)
- [x] Backend: SupplierRepo, OrganizationRepo, SupplierService, SupplierHandler
- [x] Backend: обновлены DocumentInput, documentSelect, ListDocuments (DocumentFilters), CreateDocument
- [x] API: GET/POST/PUT/DELETE /suppliers[/:id], GET /organizations, через gateway
- [x] Frontend: suppliers.ts, обновлён documents.ts (новые поля)
- [x] Frontend: ReceiptsView.vue со всеми колонками как в МойСклад (№ / Время / Склад / Контрагент / Организация / Сумма / Оплачено / Входящая дата / Входящий номер / Отправлено / Напечатано / Комментарий)
- [x] Frontend: маршрут /purchases/receipts, подтабы раздела Закупки
## Этап 21. Управление закупками (планирование)
- [x] Backend: SalesAnalytics в order — агрегация проданного за N дней по shipped-заказам
- [x] API: GET /analytics/sales?days=N через gateway (order-сервис)
- [x] Frontend: analytics.ts
- [x] Frontend: PurchasesPlanningView.vue — 18 колонок (продажи + остатки + рекомендации),
      прогноз на N дней, сортировка, фильтры, футер с итогами
- [x] Frontend: маршрут /purchases/planning, подтаб «Управление закупками» в разделе Закупки

Формулы:
- avgDailySales = sold_qty / N
- days_of_stock = available / avgDailySales
- expectedDemand = avgDailySales × N
- supply = available − expectedDemand
- to_order = max(0, expectedDemand − available)
## Этап 22. Продажи → Аналитика продаж
- [x] Backend: /analytics/sales/daily?days=N — агрегация по дням (DATE(shipped_at))
- [x] Frontend: analytics.ts — salesDaily()
- [x] Frontend: SalesAnalyticsView.vue — KPI-карточки (сумма/чек/прибыль/топ),
      bar-chart по дням, таблица по товарам с сортировкой, экспорт CSV,
      селектор периода (7/14/30/90/180/365)
- [x] Frontend: маршрут /sales/analytics, подтаб «Аналитика» в разделе Продажи
## Этап 23. Карточка товара в Остатках (правая панель)
- [x] Backend: GET /api/v1/stock/product/:id — детализация остатков по товару
      (сводка + разбивка по складам + движения с документами)
- [x] Backend: repository.MovementsByProduct + BalancesByProduct,
      service-обёртки, handler.ProductStockDetail
- [x] Frontend: stock.ts — productStockDetail()
- [x] Frontend: StockView.vue — клик по строке открывает правую панель:
      название, код/артикул, 10 метрик, таблица «Себестоимость» с разбивкой
      по складам и движениям (тип, номер, даты, кол-во, дней, себест., сумма)
- [x] Анимация выезда, кнопка закрытия, подсветка выбранной строки