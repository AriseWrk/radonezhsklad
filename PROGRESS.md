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
- [x] Импорт товаров из МойСклад API (3556 позиций) — см. Этап 34
- [x] Импорт контрагентов из МойСклад API (146)
- [x] Импорт организаций из МойСклад API (1)
- [x] Импорт складов из МойСклад API (223)



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

| audit      | 8085 | внутренний         |

| gateway    | 8080 | единая точка входа |

| web (Vite) | 5173 |                    |



## 📌 Контекст для продолжения

- Стек: Go 1.22 + Gin, PostgreSQL 16, Vue 3 + TS + Vite + Pinia + axios

- Корень: `D:\Radonezhsklad\projects\radonezhsklad`

- ОС: Windows, PowerShell 7 (проверено на 7.6.6). Все SQL — через `Invoke-SqlQuery` из `devtools.ps1` (docker cp + UTF-8 без BOM).

- Терминалы:

  1. `services\auth`      — `go run .\cmd\api` (auth :8081)

  2. `services\product`   — `go run .\cmd\api` (product :8082)

  3. `services\warehouse` — `go run .\cmd\api` (warehouse :8083)

  4. `services\order`     — `go run .\cmd\api` (order :8084)

  5. `services\gateway`   — `go run .\cmd\api` (gateway :8080)

  6. `services\audit`     — `go run .\cmd\api` (audit :8085)

  7. `web`                — `npm run dev` (frontend :5173)

  8. корень               — тесты (PowerShell, сначала `. .\devtools.ps1`)



## ⚠️ Windows + PowerShell 7: работа с UTF-8


**Проблема:** PS 7 по умолчанию читает/пишет в CP1251 → русский бьётся в `?????`.



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

- `-replace` в PS 7 для многострочных замен — вместо этого перезаписывайте файл целиком через `Write-Utf8File`



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

| Inventory  | GET /inventory/prepare?warehouse_id=   | —                            |

| Suppliers  | GET/POST/PUT/DELETE /suppliers[/:id]   | admin, manager               |

| Organizations | GET /organizations   | admin, manager               |

| InternalOrders | GET/POST /internal-orders, GET /internal-orders/:id, POST /internal-orders/:id/{print,send}, GET /internal-orders/:id/export | admin, manager, warehouse |

| Contracts  | GET/POST/PUT/DELETE /contracts[/:id]   | admin, manager               |

| Analytics  | GET /analytics/sales?days=N, GET /analytics/sales/daily?days=N   | admin, manager               |

| Audit      | GET /audit (admin), POST /internal/audit (X-Internal-Token)   | admin / internal             |



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
  description, price, cost_price, min_stock, currency, is_archived, created_at, updated_at,
  external_id UNIQUE (partial, где NOT NULL), external_code, source (manual|moysklad),
  weight, volume



### radonezh_warehouse

- **warehouses** — id, name, address, is_active, external_id, external_code, source,
  created_at, updated_at

- **documents** — id, type (receipt/shipment/transfer/inventory), number,

  status (draft/posted/cancelled), warehouse_id, target_warehouse_id, comment,

  created_by, created_at, updated_at, posted_at, cancelled_at,

  supplier_id FK, organization_id FK, incoming_number, incoming_date,

  paid_amount, printed_at, sent_at

- **suppliers** — id, name, inn, kpp, ogrn, okpo, phone, email, address,
  legal_address, actual_address, fax, comment, counterparty_type, archived,
  external_id, external_code, source, created_at, updated_at

- **organizations** — id, name, inn, kpp, ogrn, okpo, legal_address, email,
  is_default, archived, external_id, external_code, source, created_at

- **internal_orders** — id, number, organization_id FK, warehouse_id FK,

  status, total, currency, comment, owner_id, owner_dept, printed_at, sent_at,

  shipped_amount, created_by, created_at, updated_at

- **internal_order_items** — id, internal_order_id FK CASCADE, product_id, quantity, price

- **document_items** — id, document_id FK CASCADE, product_id, quantity, price

- **stock_balances** — id, warehouse_id, product_id, quantity, updated_at,

  UNIQUE(warehouse_id, product_id)

- **stock_movements** — id, warehouse_id, product_id, document_id, quantity_delta,

  created_at (audit)



### radonezh_order

- **customers** — id, name, full_name, last_name, first_name, middle_name,
  phone, fax, email, address, legal_address, actual_address,
  inn, kpp, ogrn, okpo, external_code, external_id, source, counterparty_type,
  status, group_name, comment, archived, created_at, updated_at

- **orders** — id, number UNIQUE, customer_id FK, warehouse_id,

  status (draft/confirmed/shipped/cancelled), total, currency, comment,

  warehouse_doc_id, created_by, created_at, updated_at, confirmed_at,

  shipped_at, cancelled_at

- **order_items** — id, order_id FK CASCADE, product_id, quantity, price

- **contracts** — id, number, code, doc_date, customer_id FK, organization_id,

  amount, currency, paid, fulfilled, comment, printed_at, sent_at,

  contract_type, archived, created_at, updated_at



### radonezh_audit

- **audit_logs** — id, user_id, user_email, method, path, resource, resource_id,

  status, request_body, client_ip, request_id, created_at



## 🔗 Межсервисное взаимодействие

- **warehouse → product** (HTTP :8082, GET /products/list?ids=..., GET /units):

  для расширенной таблицы остатков

- **order → warehouse** (HTTP :8083, POST /documents + POST /documents/:id/post):

  при отгрузке заказа создаётся и проводится документ shipment

- **gateway → audit** (HTTP :8085, POST /internal/audit с X-Internal-Token):

  асинхронно логирует все POST/PUT/PATCH/DELETE с кодом 2xx

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

    │   ├── audit/               ← :8085, radonezh_audit

    │   │   ├── cmd/api/main.go

    │   │   ├── internal/{config,db,models,repository,handler}/

    │   │   └── migrations/{0001_init,0002_suppliers,0003_internal_orders,0004_internal_orders_ext}.sql

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

    │   │   └── migrations/{0001_init,0002_crm,0003_contracts,0004_contract_type}.sql

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

## Этап 25. Внутренние заказы — редизайн 1:1 с МойСклад

- [x] Миграция 0004: shipped_amount, sent_at, printed_at, owner_id, owner_dept

- [x] Backend: модель обновлена, select и scan, методы MarkPrinted/MarkSent

- [x] API: POST /internal-orders/:id/print, POST /internal-orders/:id/send

- [x] Frontend: полностью переделан InternalOrdersView.vue:

      toolbar (Заказ, Фильтр, поиск, счётчик выбранных, Изменить, Статус,

      Создать, Печать, настройки),

      расширенная фильтр-панель в 3 ряда (Период, Товар/группа, Склад, Проект,

      Организация, Статус, Проведено, Напечатано, Отправлено, Владелец,

      Общий доступ, Когда изменен, Кто изменил),

      таблица с колонками № / Время / Организация / Сумма / Отгружено / Отправлено / Напечатано / Комментарий,

      чекбоксы, бейджи Напечатан/Отправлен, футер с «Показать итоги»

## Этап 26. Печать внутреннего заказа в Excel

- [x] Backend: exporter.InternalOrderXLSX через github.com/xuri/excelize/v2

      — лист «Заявка», заголовок, склад, объект, шапка таблицы,

      позиции, ИТОГО, три строки подписей

- [x] Backend: GET /api/v1/internal-orders/:id/export — генерирует xls

      (xlsx-контент), отдаёт файл с именем intorder1-{номер}.xls,

      попутно ставит printed_at

- [x] Backend: OrganizationRepo.Get, SupplierService.GetOrganization

- [x] Backend: InternalOrderHandler получил зависимости warehouseSvc, supplierSvc, productCli

- [x] Frontend: exportInternalOrder() с responseType: 'blob', парсинг Content-Disposition

- [x] Frontend: кнопка «Печать» в карточке заказа скачивает файл

## ⚠️ КРИТИЧНО: Docker Desktop на Windows ломает UTF-8 в pipe



**Симптом:** `docker exec -i rs_postgres psql -c "SELECT name..."` выводит `?????`, и **запись через pipe тоже искажается** — в БД попадают байты 0x3F вместо UTF-8.



**Доказательство:** `encode(name::bytea, 'hex')` возвращает `3f3f3f...` вместо `d0a8d182d183d0bad0b0` (Штука).



**Причина:** pipe `docker exec` на Windows конвертирует stdin/stdout через кодовую страницу консоли (CP866/CP1251), теряя не-ASCII байты.



**Единственный надёжный способ — через docker cp:**

1. Записать SQL в локальный файл (UTF-8 без BOM)

2. `docker cp file.sql rs_postgres:/tmp/file.sql`

3. `docker exec rs_postgres psql -f /tmp/file.sql > /tmp/out.txt`

4. `docker cp rs_postgres:/tmp/out.txt out.txt`

5. Прочитать локально как UTF-8



**Реализовано в devtools.ps1** функциями `Invoke-SqlQuery` и `Invoke-SqlFile`. **Никогда** не использовать `Get-Content file | docker exec -i psql` и не использовать `docker exec -i psql -c "..."` для русских строк.



Backend/API/браузер работают корректно — проблема только на входе/выходе контейнера через pipe.

## Этап 28. Страница «Товары» — редизайн 1:1 с МойСклад

- [x] Левая панель категорий со счётчиками, клик фильтрует таблицу

- [x] Toolbar: + Товар / ¥ Услуга / ⚙ Комплект / 🗂 Группа / Фильтр / поиск / счётчик / Изменить / Печать / Импорт / Экспорт / ⚙

- [x] Таблица с колонками: чекбокс / Наименование / Код / Артикул / Ед. изм. / Цена продажи

- [x] Сортировка по клику на заголовок (name, sku, price)

- [x] Футер с пагинацией и итогами (позиций, сумма)

- [x] Модалка создания/редактирования расширена: category, unit, price, cost_price, min_stock, barcode, description

- [x] Экспорт в CSV

## Этап 29. Страница «Аудит» — редизайн 1:1 с МойСклад

- [x] Таблица из 3 колонок: Время | Сотрудник | Событие

- [x] Сотрудник: аватар с инициалами + «Фамилия И. О.»

- [x] Человекочитаемые события на русском (matchSpecial, detectResource, verbFor)

      — Создан товар «...», Проведён документ, Распечатан внутренний заказ, Отгружен заказ и т.д.

- [x] Из request_body вытягивается number/name для ссылки-идентификатора

- [x] Фильтры: период, сотрудник, поиск по описанию события

- [x] Пагинация с 4 кнопками и «1–100 из N»

- [x] Иконка ⓘ возле заголовка, ↻ обновление

## Этап 31. CRM → Договоры + импорт из Excel

- [x] Миграция 0003_contracts: таблица contracts (number, code, doc_date, customer_id, organization_id, amount, currency, paid, fulfilled, comment, printed_at, sent_at, archived)

- [x] Backend order: models.Contract, ContractRepo, ContractService, ContractHandler

- [x] API: GET/POST /contracts, GET/PUT/DELETE /contracts/:id

- [x] Gateway: проксирует /contracts

- [x] Frontend: api/contracts.ts, ContractsView.vue (тулбар, чекбоксы, сортировка, фильтры, жёлтая подсветка неоплаченных, экспорт)

- [x] Router: /contracts, подтаб Договоры в разделе CRM

- [x] Import-Contracts.ps1 — импорт из XLS, матчит контрагентов по имени

## Этап 32. Безопасность аудита — маскирование чувствительных полей

- [x] gateway/internal/audit: добавлена sanitizeBody — рекурсивно маскирует

      значения полей password, password_hash, token, refresh_token,

      access_token, secret в request_body перед отправкой в audit-сервис

- [x] Очищены накопленные записи audit_logs (regexp_replace по password)

## Этап 33. Фикс аналитики продаж (500 → 200)

- [x] order/internal/repository: заменено ($1::text || ' days')::interval

      на make_interval(days => $1) в SalesAnalytics и SalesDaily

- [x] Причина: pgx v5 не мог согласовать тип параметра $1 при ::text-касте,

      падал до выполнения запроса (в psql тот же SQL работал корректно)


## Этап 34. Импорт справочников из МойСклад API

### Что сделано (товары)

- [x] Клиент МойСклад API в devtools.ps1: Get-MsToken, MsApi-Get, MsApi-GetAll
- [x] Токен вне репо: D:\Radonezhsklad\.secrets\moysklad.token (40 hex)
- [x] Импортировано 3556 товаров из /entity/product в radonezh_product.products
- [x] Схема products расширена: +external_id, +external_code, +source, +weight, +volume
- [x] В units добавлен km (Километр)

### Ключевые особенности API МойСклад

- Accept-Encoding: gzip обязателен, иначе 415 Unsupported Media Type
  Accept: application/json;charset=utf-8, используем curl.exe --compressed
- Цены в минорных единицах (копейки): JSON 5139.0 = 51.39 руб, делим на 100
- Даты без TZ, по факту МСК (UTC+3), при парсинге добавляем +03
- Все связи через meta-объекты (meta.href), UUID из хвоста URL
- code у товаров уникален (3556/3556) - используем как наш sku для UPSERT
- externalCode тоже уникален - пишем в external_code
- id (uuid) - пишем в external_id, ключ для идемпотентного upsert
- Лимиты API: 45 запросов / 3 секунды на аккаунт

### Проблема cp866 в stdout curl.exe

PS 7 на русской Windows читает stdout внешних процессов в cp866,
а curl отдаёт UTF-8. Решение: curl.exe -o <tempfile> +
[IO.File]::ReadAllText($tmp, UTF8NoBOM). Уже инкапсулировано в MsApi-Get.

### Примеры использования

cd D:\Radonezhsklad\projects\radonezhsklad
. .\devtools.ps1 ; Set-ConsoleUtf8
$org = MsApi-Get '/entity/organization' @{limit=1}
$products = MsApi-GetAll '/entity/product' -PageSize 1000

### Маппинг MS - products

| МойСклад | products | Примечание |
|---|---|---|
| id (uuid) | external_id | UPSERT-ключ |
| code | sku | UNIQUE |
| externalCode | external_code | UNIQUE в МС |
| name | name | |
| archived | is_archived | |
| uom.meta.href (uuid) | unit_id | через VALUES-маппинг uuid - code |
| salePrices[0].value / 100 | price | тип "Цена продажи" |
| buyPrice.value / 100 | cost_price | |
| barcodes[0].ean13 | barcode | только первый штрихкод |
| weight, volume | weight, volume | у всех 0 |
| (нет) | source | moysklad |
| productFolder | category_id | NULL, у товаров нет папок |

### Что сделано (справочники)

- [x] Контрагенты: 146 записей
      customers: +external_id, +source, UNIQUE external_code (full)
      suppliers: те же 146, роль поставщика
- [x] Организации: 1 запись (ООО ЧОО АБ "РАДОНЕЖ")
      +kpp, +ogrn, +okpo, +legal_address, +email, +archived
- [x] Склады: 223 записи

Миграции:
- services/product/migrations/0003_moysklad.sql
- services/order/migrations/0005_moysklad.sql
- services/warehouse/migrations/0005_moysklad.sql

### Маппинг MS - наши таблицы

| Сущность МС | Таблица | UPSERT-ключ |
|---|---|---|
| counterparty | radonezh_order.customers | external_code |
| counterparty | radonezh_warehouse.suppliers | external_code |
| organization | radonezh_warehouse.organizations | external_code |
| store | radonezh_warehouse.warehouses | external_code |

UNIQUE-индекс на external_code должен быть full, не partial — иначе
ON CONFLICT (external_code) падает с "no unique or exclusion constraint".
### Что дальше

- [x] Контрагенты (/entity/counterparty) — customers + suppliers
- [x] Организации (/entity/organization) — organizations
- [x] Склады (/entity/store) — warehouses
- [ ] Документы (опционально): приёмки, отгрузки, внутренние заказы

---

## Этап 35: Импорт документов из МойСклад

### Разведка

Объёмы по типам (49485 документов, ~229k позиций):

| Тип | Кол-во | Позиций | Куда |
|---|---:|---:|---|
| supply | 4789 | 13444 | documents (receipt) |
| demand | 9 | 15 | documents (shipment) |
| move | 15554 | ~23300 | documents (transfer) |
| enter | 18 | 38 | documents (receipt) |
| loss | 17276 | ~27600 | documents (writeoff) |
| inventory | 749 | ~103700 | inventories + inventory_items |
| internalorder | 11090 | ~68800 | internal_orders |

МС используется только как складской контур: платежей/кассы/заказов покупателя нет.

### Маппинг UUID

Наши справочники имеют external_id (МС UUID) и external_code (короткий хэш МС):

| Наш | external_id | external_code |
|---|---|---|
| products | product.id (UUID) | product.externalCode |
| warehouses | store.id | store.externalCode |
| suppliers | counterparty.id | counterparty.externalCode |
| organizations | organization.id | organization.externalCode |

Резолв FK: JOIN по external_id (UUID). Резолв product_id — в PowerShell,
т.к. products в другой БД (radonezh_product), cross-DB JOIN невозможен.

### Миграции

- 0006_moysklad_documents.sql — external_id/external_code/doc_date/total/vat_enabled/vat_included
  в documents; external_id/vat_rate/discount/sum в document_items;
  external_id/external_code в internal_orders; external_id в internal_order_items;
  таблицы inventories + inventory_items
- 0007_unique_external_id_full.sql — full UNIQUE на external_id (partial не работает с ON CONFLICT)

### Инфраструктура импорта

devtools.ps1 — MsApi-Get переопределён с retry:
- валидация JSON (первый символ { или [)
- 6 попыток, backoff 1s → 30s
- логирование preview при ошибке

scripts/import-ms-docs.ps1 — параметризованный импортёр:
- -Type supply|demand|move|enter|loss
- -DocType receipt|shipment|transfer|writeoff
- -Limit N — ограничение для пилота
- -KeepTsv — сохранить TSV для диагностики

Алгоритм:
1. product_map: ms_uuid → our_uuid (одним SELECT)
2. пагинация /entity/$Type с expand=positions (100/стр, sleep 120ms) — 48 запросов на 4789 док.
3. TSV в формате text (\t разделитель, \N = NULL, экранирование \\, \t, \n, \r)
4. docker cp + \copy ... WITH (FORMAT text) в TEMP-таблицы
5. INSERT documents c JOIN warehouses/suppliers/organizations по external_id
6. INSERT document_items с готовыми product_id

scripts/sql/upsert_documents.sql — SQL-шаблон с плейсхолдерами
{{DOCS_PATH}} / {{ITEMS_PATH}} / {{DOC_TYPE}}.

### Подводные камни этапа

1. MS rate limit: 45 запросов за 3 сек (X-RateLimit-Limit: 45,
   X-Lognex-Retry-TimeInterval: 3000). Превышение → HTML вместо JSON.
   Решение: expand=positions (48 запросов вместо 4837) + sleep 120ms.
2. expand=positions: позиции в d.positions.rows (все), meta.size для проверки.
3. FORMAT csv в \copy ломается на кавычках в текстах. Только FORMAT text
   (\N = NULL, экранирование \\, \t, \n, \r).
4. FORMAT text не поддерживает HEADER — TSV без заголовков.
5. [math]::Round в ru-RU сериализует 3783,6 — ломает COPY. В начале скрипта:
   [Threading.Thread]::CurrentThread.CurrentCulture = [InvariantCulture].
6. partial UNIQUE не работает с ON CONFLICT — только full UNIQUE (NULL разрешены многократно).
7. vatIncluded/applicable могут отсутствовать в ответе — COALESCE(..., false) в SQL.
8. docker cp + psql \copy: psql интерпретирует \N как NULL в FORMAT text.

### Результат

- supply → documents.type=receipt: 4789 док. / 13444 поз. / 0 missing products
  время: ~2:45 (fetch 2:30 + COPY/INSERT 15 сек)


### Дополнение (move, loss, enter)

| МС тип | DocType | docs | items |
|---|---|---:|---:|
| supply | receipt | 4789 | 13444 |
| enter | receipt | 18 | 35 |
| demand | shipment | 9 | 15 |
| move | transfer | 15554 | 54796 |
| loss | writeoff | 17276 | 60041 |
| **итого** | | **37646** | **128331** |

Gotcha: в move/loss/enter позициях нет vat/discount — COALESCE(...,0) в SQL.

### Что осталось

- internalorder: 11090 док. / ~68800 поз. → internal_orders + internal_order_items
- inventory: 749 док. / ~103700 поз. → inventories + inventory_items


### Финализация: internalorder + inventory

| МС тип | Таблица | docs | items |
|---|---|---:|---:|
| internalorder | internal_orders + internal_order_items | 11090 | 45714 |
| inventory | inventories + inventory_items | 749 | 46619 |

Новые скрипты:
- scripts/import-ms-internalorders.ps1
- scripts/import-ms-inventory.ps1
- scripts/sql/upsert_internalorders.sql
- scripts/sql/upsert_inventory.sql

Особенности:
- internalorder: plan_date из deliveryPlannedMoment, project UUID сохраняется как текст
  (в internal_orders.project VARCHAR), store/org через external_id
- inventory: page size 50 (позиций много: некоторые доки >1000 поз., нужен fallback
  MsApi-GetAll), поля calculatedQuantity/correctionAmount/correctionSum
- missing product mappings: 18 (internalorder) + 1 (inventory) — товары, которых нет
  в нашей БД; документы импортированы, эти позиции пропущены (product_id NOT NULL)

### Итог Этапа 35

Всего импортировано 49485 документов / ~220000 позиций:

| Категория | docs | items |
|---|---:|---:|
| receipt (supply+enter) | 4807 | 13479 |
| shipment (demand) | 9 | 15 |
| transfer (move) | 15554 | 54796 |
| writeoff (loss) | 17276 | 60041 |
| internalorder | 11090 | 45714 |
| inventory | 749 | 46619 |
| **ВСЕГО** | **49485** | **220664** |

Что НЕ импортировано (в МС 0 записей): receipt, customerorder, paymentin/out, cashin/out.

Пуш на GitHub: 428a618, f921da6, aab105e (затем финальный коммит этапа).


### Этап 36: дотянуть архивные товары (задача A)

Причина 19 missing product mappings: в МС 12 товаров с `archived=true`,
которые MS API не отдаёт в `/entity/product` по умолчанию. 5 из них
реально использованы в позициях internalorder, 1 — в inventory.

Что сделано:
- `services/product/migrations/0004_unique_external_id_full.sql` — full UNIQUE на external_id
  (partial не работает с ON CONFLICT)
- `scripts/import-ms-archived-products.ps1` — импорт `filter=archived=true`
  с маппингом uom (19f1edc0→pcs, dfe54549→m, +5 других) и ценами из salePrices/buyPrice
- Перезапущены `import-ms-internalorders.ps1` и `import-ms-inventory.ps1` (UPSERT)

Результат:
- products: 3556 → 3568 (+12 архивных, is_archived=true)
- internal_order_items: 45714 → 45732 (+18)
- inventory_items: 46619 → 46620 (+1)
- missing product mappings: 0


### Этап 37: пересчёт stock_balances / stock_movements (задача B)

Было: 1421 балансов и 1421 движения от seed (Restore-Stock.ps1, 6 документов).

Алгоритм:
1. Movements — из всех posted-документов (receipt +, shipment/writeoff/transfer -)
   + transfer: приход на target_warehouse_id
2. Balances — из /report/stock/all с фильтром `filter=store=<href>` по каждому
   складу (223 шт.). Это истина из МС, включая отрицательные остатки.

Скрипт: scripts/sync-stock.ps1 (-SkipMovements / -SkipBalances).

Результат:
- stock_movements: 183 731 (по 37 467 документам, 3421 товар, 189 складов)
  receipt 14 895 + shipment 18 + transfer 109 288 + writeoff 59 530 = 183 731
- stock_balances: 3 019 (85 складов, 453 отрицательных — как в МС, 0 missing products)

Также почищены seed-документы (5 receipt + 3 shipment + 1 transfer, external_id=NULL) —
они не попадают в movements, т.к. их type вне ''receipt''/''shipment''/''writeoff''/''transfer''
и/или status != posted.


---

## Этап 38: Раздел «Инвентаризации» (frontend + backend)

### Backend (services/warehouse)

Новые файлы:
- internal/repository/inventories.go — InventoryRepo (List/Get с JOIN на warehouses)
- internal/service/inventories.go — InventoryService
- internal/handler/inventories.go — Handler с обогащением имён через productClient.ListByIDs

Изменения:
- internal/models/models.go — +Inventory, +InventoryItem (ProductName, ProductSKU)
- cmd/api/main.go — invRepo/invSvc/invH + routes:
  - GET /inventories?warehouse_id=&from=&to=
  - GET /inventories/:id (с items + product_name/sku)

Gateway (services/gateway/cmd/api/main.go):
- проксирование /inventories и /inventories/*path -> warehouse

### Frontend (web/)

Новые:
- src/api/inventories.ts — listInventories, getInventory + типы Inventory/InventoryItem
- src/views/InventoriesView.vue — список с фильтром (склад/период) + пагинация
- src/views/InventoryCardView.vue — карточка 1:1 по образцу InternalOrderCardView

Изменения:
- src/router/index.ts — routes /inventories и /inventories/:id
- src/layouts/MainLayout.vue — пункт меню «Инвентаризации» -> /inventories
- src/views/InventoriesView.vue — клик по строке -> router.push(карточка)

Заодно починены накопленные TS-ошибки:
- CounterpartiesView.vue — убран c.okrn
- InternalOrderCardView.vue — убран неиспользуемый cancelInternalOrder/blurTimer
- InternalOrdersView.vue — убран неиспользуемый warehouseName

### Проверки

- API: GET /api/v1/inventories?limit=3 возвращает 500 записей
- GET /api/v1/inventories/:id возвращает items с product_name/sku
- npm run build — успешно, InventoryCardView-*.js собран

### Известные хвосты

- В карточке: pager «1 из 749» захардкожен, prev/next ведут на список
- Кнопки «Изменить», «Создать документ», «Печать», таб «Связанные документы»,
  «Задачи», «Файлы» — заглушки (disabled)
- Аналогичный каркас надо будет сделать для «Документов» (там сейчас модалок нет —
  пользуются DocumentsView с типами receipt/shipment/transfer/writeoff)


---

## Этап 39: Внутренние заказы — read-only просмотр импортированных из МС

### Проблема

Импортированные из МойСклад внутренние заказы (11090 шт.) имеют status='draft',
из-за чего фронт показывал их редактируемыми (canEdit = isNew || status === 'draft')
и разрешал проведение через чекбокс «Проведено». Это ломало данные: пользователь
мог случайно изменить/провести документ МС.

### Backend (services/warehouse)

- internal/models/models.go: + поле ExternalID *uuid.UUID `json:"external_id,omitempty"`
- internal/repository/internal_orders.go: intOrderSelect + scanIntOrder + List
  — добавлен o.external_id (&o.ExternalID в scan)
- Коммит ef46603

### Frontend (web/)

- src/api/internalOrders.ts: + external_id?: string | null в InternalOrder
- src/views/InternalOrderCardView.vue:
  - новый ref externalId; заполняется в load() из o.external_id
  - canEdit = isNew || (status === 'draft' && !externalId)
  - бейдж «МОЙСКЛАД» в шапке карточки (title — «только просмотр»)
  - CSS .ext-badge
- Коммит 8410545

### Проверки

- API: GET /internal-orders/:id отдаёт external_id для импортированных,
  пусто (omitempty) — для CRUD.
- vue-tsc --noEmit — OK.
- В браузере: импортированный → все поля серые, кнопка «Сохранить» disabled,
  чекбокс «Проведено» disabled, бейдж «МОЙСКЛАД»; CRUD-заказ — редактируемый.


---

## Этап 40: Pager в карточке инвентаризации

### Проблема

В InventoryCardView.vue было захардкожено «1 из 749», prev() и next() оба вели
на список /inventories. Нельзя было листать карточки подряд.

### Backend (services/warehouse)

- repository/inventories.go: +InventoryNeighbors + метод Neighbors(ctx, id) —
  оконная функция ROW_NUMBER() OVER (ORDER BY doc_date DESC, id DESC) + COUNT(*) OVER ()
  Возвращает prev_id/next_id (или null на краях), position, total.
- service/inventories.go: обёртка Neighbors.
- handler/inventories.go: GET /inventories/:id/neighbors → {prev_id, next_id, position, total}
- cmd/api/main.go: маршрут read.GET("/inventories/:id/neighbors", invH.Neighbors)

### Frontend (web/)

- src/api/inventories.ts: +InventoryNeighbors + getInventoryNeighbors(id) —
  однострочный URL /inventories//neighbors (многострочный template literal
  ломал путь и gateway отдавал 404 из-за двойного слеша).
- src/views/InventoryCardView.vue:
  - refs prevId/nextId/position/total
  - load() параллельно с getInventory тянет getInventoryNeighbors
  - кнопки ‹/› теперь :disabled на краях и переходят на /inventories/:id
  - счётчик {{ position }} из {{ total }}
  - watch(() => route.params.id, load) — Vue Router переиспользует компонент
    при смене только params.id, без watch карточка не перезагружалась.

### Проверки

- API /inventories/:id/neighbors: position=1/total=749/prev_id=null на первом,
  position=749/next_id=null на последнем.
- vue-tsc --noEmit — OK.
- В браузере: prev/next листают карточки, счётчик меняется, крайние кнопки disabled.

### Известные хвосты

- Кнопки «Изменить», «Создать документ», табы «Связанные документы»/«Задачи»/«Файлы»
  в карточке — по-прежнему заглушки (disabled).
- Кнопка «Печать» работает через window.print() (браузерный диалог).


---

## Этап 41: Карточка документа (receipt / shipment / transfer / writeoff)

### Проблема

DocumentsView показывал список по типам, но открытие документа было модалкой
с базовой информацией без наименований товаров. Карточки как таковой не было.

### Backend (services/warehouse)

- models.go: + Document.ExternalID, + DocItem.ProductName/ProductSKU
- repository.go: documentSelect + d.external_id; scanDocument и ListDocuments
  сканируют &d.ExternalID
- handler.go:
  - + import "strings"
  - GetDocument теперь обогащает items через productClient.ListByIDs —
    product_name + product_sku (тот же паттерн, что InventoryHandler.Get)

### Frontend (web/)

- api/documents.ts: + Document.external_id, + DocItem.product_name/product_sku
- router/index.ts: + /documents/:id → DocumentCardView.vue
- DocumentsView.vue:
  - useRouter, функция openCard(d) → router.push('/documents/'+d.id)
  - кнопка «Открыть» переведена с модалки на карточку
- views/DocumentCardView.vue (новый, ~250 строк):
  - шапка: тип, №, дата, статус-пилюля, бейдж «МОЙСКЛАД» для импортированных
  - поля: склад, склад-получатель (transfer), организация, входящий №/дата,
    оплачено, проведён/отменён
  - таблица позиций: наименование + артикул + кол-во + цена + сумма + итого
  - комментарий (pre-wrap)
  - кнопки: Провести (только draft + не импортированный), Отменить (posted + не импортированный),
    Печать (window.print()), Закрыть
  - watch(() => route.params.id, load) — перезагрузка при переходе между документами

### Проверки

- API: /documents/:id для writeoff/receipt/transfer/shipment возвращает
  external_id и product_name/product_sku в items.
- vue-tsc --noEmit — OK.
- В браузере: клик по строке открывает карточку, у импортированных бейдж «МОЙСКЛАД»
  и скрыты кнопки Провести/Отменить.

### Известные хвосты

- «Провести» / «Отменить» доступны для CRUD-документов (draft/posted без external_id);
  на реальных импортированных данных не тестировалось (бета-тест).
- Редактирование draft-документов (изменение позиций) — не реализовано.


---

## Этап 42: Создание корректирующего документа из инвентаризации

### Проблема

В карточке инвентаризации кнопки «Сохранить», «Изменить», «Создать документ»
были заглушками, табы «Задачи»/«Файлы» — тоже. Все 749 инвентаризаций импортированы
из МС (external_id NOT NULL), CRUD-редактирование не имеет смысла.
Полезная функция — сгенерировать writeoff (списание недостач) или receipt
(оприходование избытков) на основе коррекций инвентаризации.

### Backend (services/warehouse)

- migrations/0008_document_source_inventory.sql: documents.source_inventory_id UUID
  REFERENCES inventories(id) ON DELETE SET NULL + partial index.
- models.go: Document.SourceInventoryID *uuid.UUID.
- repository.go:
  - documentSelect + d.source_inventory_id; scan в scanDocument и ListDocuments;
  - DocumentInput + SourceInventoryID;
  - CreateDocument INSERT дополнен полем (11 параметров);
  - DocumentExistsForInventory(ctx, inventoryID, docType) — защита от дубля.
- service/inventories.go:
  - InventoryService получил docRepo *repository.Repo (второй аргумент конструктора);
  - CreateCorrection(ctx, invID, kind) — shortage → writeoff, surplus → receipt;
    фильтрует items по знаку correction_amount, создаёт документ draft с номером
    <номер-инвентаризации>-СП / -ОП и комментарием.
- handler/inventories.go: CreateCorrection — POST /inventories/:id/create-correction
  с {kind: shortage|surplus}, отвечает 201 или 400.
- main.go: подключение docRepo к InventoryService + роут docWrite.

### Frontend (web/)

- api/inventories.ts: + CorrectionKind, + CorrectionDocument, + createInventoryCorrection.
- views/InventoryCardView.vue:
  - тулбар упрощён (убраны «Сохранить», «Изменить», «Задачи», «Файлы»);
  - активная кнопка «Создать документ» (primary);
  - модалка с 2 кнопками («Списать недостачи» / «Оприходовать избытки»);
  - после успеха — редирект на карточку созданного документа;
  - при 400 (уже создан) — красное сообщение в модалке.

### Проверки

- POST /inventories/<id>/create-correction {"kind":"shortage"} → 201, документ
  00064-СП writeoff, 5 позиций, total 5881.25, source_inventory_id установлен.
- Повторный вызов → 400 «корректирующий документ уже создан».
- vue-tsc --noEmit — OK.
- warehouse build — OK.

### Известные хвосты

- Таб «Связанные документы» показывает заглушку — расширить до списка
  документов с этим source_inventory_id.
- Проведение созданного документа — вручную через карточку документа.


---

## Этап 43: Сверка сумм с МойСклад + VAT-логика в total

### Проблема

Требовалась валидация импорта 49 485 документов: сравнение SUM(documents.total)
с суммой из МС по каждому типу. При сверке выяснилось расхождение receipt на
-1.76 млн руб.

### Причина

В МС у документа есть флаги vatEnabled и vatIncluded. Когда vatEnabled=true
и vatIncluded=false, MS добавляет НДС СВЕРХУ к сумме qty×price (цены в позициях
без НДС). Мы же всегда считали total = SUM(qty×price).
Затронуто 140 supply-документов из 4807.

### Скрипт сверки

scripts/audit-ms-sums.ps1 — пагинирует /entity/{supply,enter,demand,move,loss}
(limit=100, expand=positions), суммирует .sum/100, сравнивает с нашей SQL-суммой.
ВНИМАНИЕ: при limit>100 MS молча НЕ отдаёт positions — использовать limit<=100.

Результат сверки:
- transfer, shipment, writeoff: расхождение 0.00 (в копейках — округление)
- receipt: -1 758 585.40 руб из-за VAT
- ИТОГО MS: 850 265 304.49 руб; наши: 848 506 719.92 руб

### Фикс (services/warehouse)

repository.go: documentSelect — формула total дополнена VAT-условием:

    SUM(CASE
      WHEN d.vat_enabled AND NOT d.vat_included AND di.vat_rate > 0
        THEN di.quantity * di.price * (100 + di.vat_rate) / 100
      ELSE di.quantity * di.price
    END)

Поля vat_enabled/vat_included уже импортированы в documents; vat_rate — в
document_items. Доп. миграций не потребовалось.

### Проверки после фикса

- Классификация 4781 supply (скрипт): F_noVat=4645, F_withVat=136, F_other=0.
- Сумма MS = 292 627 821.88; наша по новой формуле = 292 627 821.87.
- По t/f документам расхождение 0.02–13 руб на документ — округление до копейки
  по каждой позиции (MS), у нас — один ROUND в конце.
- API теперь отдаёт корректный total для всех типов.

### Известные хвосты

- Для 140 t/f supply-документов расхождение до 13 руб (округление). Не критично,
  но если нужна точность до копейки — округлять в SQL по каждой позиции.
