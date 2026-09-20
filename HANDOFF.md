# RadonezhSklad — Handoff для нового чата

Полное состояние проекта — в PROGRESS.md в корне репо (1145 строк, Этап 38).
Этот файл — компактный стартовый контекст для нового чата.

---

## Стек и окружение

- Go 1.22 + Gin, PostgreSQL 16 (Docker rs_postgres, порт 5433), Redis 6380, RabbitMQ 5672
- Vue 3 + TS + Vite + Pinia + axios
- Windows + PowerShell 7.6.6, локальный корень: D:\Radonezhsklad\projects\radonezhsklad
- GitHub: https://github.com/AriseWrk/radonezhsklad (main)
- БД: radonezh_auth / radonezh_product / radonezh_warehouse / radonezh_order / radonezh_audit
- Порты: gateway 8080, auth 8081, product 8082, warehouse 8083, order 8084, audit 8085
- ВАЖНО: сервисы запускаются вручную через "go run ./cmd/api" в отдельных терминалах (НЕ docker-compose)

### Перед работой:
  cd D:\Radonezhsklad\projects\radonezhsklad
  . .\devtools.ps1
  Set-ConsoleUtf8

### Токен МойСклад:
D:\Radonezhsklad\.secrets\moysklad.token (40 hex, вне репо)

### Логин админки:
admin@radonezh.local / qwerty123

---

## Что сделано (последние этапы)

| Этап | Коммит | Что |
|---|---|---|
| 34 | 7fffcbd | Импорт справочников (контрагенты/организации/склады) |
| 35 | 428a618..0c49058 | Импорт 49 485 документов из МС |
| 36 | 6e68916 | Архивные товары (12 шт.) + missing позиции |
| 37 | d2c2282 | Пересчёт stock_balances и stock_movements |
| 38 | 6a7e36d | Раздел Инвентаризации — backend + frontend |

### Цифры в БД

- radonezh_product.products: 3568 (3556 + 12 архивных, source=moysklad)
- radonezh_warehouse.documents: receipt 4807, shipment 9, transfer 15554, writeoff 17276
- internal_orders: 11090 / items 45732
- inventories: 749 / items 46620
- stock_movements: 183731 (37467 док., 3421 товар, 189 складов)
- stock_balances: 3019 (85 складов, 453 отрицательных — как в МС)

---

## Инфраструктура импорта МС

### Функции в devtools.ps1

- Get-MsToken — читает токен
- MsApi-Get -Path <p> -Query @{...} — GET с retry (6 попыток, backoff 1s-30s, JSON-валидация)
- MsApi-GetAll -Path <p> -PageSize N — пагинация
- Invoke-SqlQuery -Db <db> -Query <sql> — psql через docker
- Invoke-SqlFile -Path <f> -Database <db>
- Write-Utf8File / Read-Utf8File — UTF-8 без BOM

### Скрипты (scripts/)

| Скрипт | Назначение |
|---|---|
| import-ms-docs.ps1 -Type X -DocType Y | supply/demand/move/enter/loss в documents |
| import-ms-internalorders.ps1 | internalorder в internal_orders |
| import-ms-inventory.ps1 | inventory в inventories |
| import-ms-archived-products.ps1 | архивные товары (filter=archived=true) |
| diag-missing-products.ps1 | диагностика product mappings |
| sync-stock.ps1 [-SkipMovements|-SkipBalances] | пересчёт остатков/движений |

SQL-шаблоны: scripts/sql/upsert_documents.sql, upsert_internalorders.sql, upsert_inventory.sql

### Паттерн импорта

1. MsApi-GetAll / пагинация с expand=positions (иначе rate limit)
2. Генерация TSV в формате text (бэкслеш-N = NULL, экранирование бэкслешей, табов, переводов строк)
3. docker cp tsv rs_postgres:/tmp/X.tsv
4. COPY tmp_X FROM tsv WITH (FORMAT text) — БЕЗ HEADER
5. INSERT ... SELECT с резолвом FK через external_id

---

## КРИТИЧНЫЕ GOTCHA (проверено на этапах 35-38)

1. MS rate limit: 45 запросов за 3 сек (X-RateLimit-Limit: 45, X-Lognex-Retry-TimeInterval: 3000).
   Превышение — HTML вместо JSON. Решение: expand=positions + Start-Sleep 120ms.
2. expand=positions возвращает позиции в d.positions.rows.
3. FORMAT csv в COPY ломается на кавычках в текстах. Только FORMAT text.
4. FORMAT text НЕ поддерживает HEADER — TSV без заголовков.
5. [math]::Round в ru-RU сериализует 3783,6 — ломает COPY.
   В начале PS-скрипта: [Threading.Thread]::CurrentThread.CurrentCulture = [InvariantCulture].
6. partial UNIQUE не работает с ON CONFLICT — только full UNIQUE.
   Миграции: 0007_unique_external_id_full.sql (warehouse), 0004_unique_external_id_full.sql (product).
7. vatIncluded/applicable могут отсутствовать в ответе — COALESCE(..., false) в SQL.
8. MS скрывает архивные товары из /entity/product — нужен filter=archived=true.
9. Обратный слэш и обратные кавычки в PS: backtick — это escape-символ.
   Go-теги json-нотации писать через одинарные кавычки или через файл.
10. PSReadLine интерпретирует Tab внутри команд как автокомплит — вставлять пробелы вручную.
11. Апострофы в одинарных PS-строках экранируются как два апострофа (двойной апостроф).
    Пример: вместо it's — it''s.

---

## Правила работы в PS 7

1. Все правки файлов — через Write-Utf8File или:
   [IO.File]::WriteAllText($abs, $s, (New-Object System.Text.UTF8Encoding $false))
2. Path через Join-Path (Get-Location).Path ... — иначе CWD-баг.
3. НЕ использовать -replace для многострочных замен — построчный обход или .Replace().
4. Большие here-string через терминал бьются — писать через [string[]]@(...) + WriteAllLines.
5. Перед каждой правкой файла — показать превью / содержимое, потом писать.
6. После каждой значимой правки — git add / commit / push + обновить PROGRESS.md.

---

## Что осталось (TODO)

### Небольшие хвосты

- Карточка инвентаризации: pager 1 из 749 захардкожен, prev/next ведут на список.
  Кнопки Изменить / Создать документ / Печать / табы Связанные документы / Задачи / Файлы — заглушки.
- Каркас для Документов: сейчас DocumentsView показывает список по типам
  (receipt/shipment/transfer/writeoff), но без карточки. Нужно DocumentCardView.vue.
- Внутренние заказы: read-only просмотр импортированных из МС — ЗАКРЫТО в Этапе 39.

### Большие темы (обсудить)

- C) Сверка сумм: SUM(documents.total) vs МС по типам — валидация импорта.
- E) Регулярная синхронизация: cron/queue для sync-stock.ps1 и импортёров.
- F) Дополнительные справочники: uom (у нас 7, в МС 62), productFolder, project, contract, expenseItem.
- B2) UI для движения товара: /stock/product/:id уже есть в backend, но не в UI.

---

## Формат ответов

- Терминальная команда одним блоком, с проверками.
- Всё, что пишем в файл — сразу UTF-8 без BOM.
- После каждого блока — жду вывод, разбираю, даю следующий.
- Коммиты с осмысленными сообщениями (feat/fix/docs: ...).
- Диагностика сначала (что реально в файле/БД), потом патч.

---

## С чего начать завтра

Вариант 1 — Добить карточку инвентаризации (10-15 мин):
- Pager (1 из 749) через query-параметр или отдельный endpoint
- Prev/next через sort + offset
- Кнопка Печать через window.print()

Вариант 2 — DocumentCardView.vue для документов (30-60 мин):
- Роут /documents/:id
- Карточка по образцу InventoryCardView
- Кнопки: Изменить (для draft), Провести, Отменить, Печать

Вариант 3 — Расширить InternalOrdersView карточкой для импортированных (30-60 мин):
- Проверить, что InternalOrderCardView работает для read-only
- Добавить ссылки из списка

Начать с диагностики: что сейчас в UI реально отображается, где кнопки, какие actions доступны.
