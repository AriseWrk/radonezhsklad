# RadonezhSklad — Web

SPA на Vue 3 + TypeScript + Vite. Pinia для состояния, vue-router, axios для API.

## Команды

```bash
npm install     # установить зависимости
npm run dev     # dev-сервер (http://localhost:5173)
npm run build   # production-сборка
npm test        # unit-тесты (vitest, run once)
```

## Структура

- `src/api/`         — типизированные клиенты API (products, internalOrders, stock, …)
- `src/views/`       — страницы (StockView, InternalOrderCardView, InternalOrdersView, …)
- `src/components/`  — переиспользуемые компоненты (MsButton, MsSmartSelect, …)
- `src/utils/`       — утилиты; `smartSearch.ts` — fuzzy-поиск номенклатуры
- `src/layouts/`, `src/router/`, `src/stores/` — каркас приложения

## Умный поиск номенклатуры

`src/utils/smartSearch.ts` — нормализация + скелет + токены + fuzzy + ранжирование.
Используется в форме создания внутреннего заказа. Экспортирует:

- `normalize(s)` — lower-case + гомоглифы + разделители + знаки умножения
- `skeleton(s)` — [a-z0-9] с вырезанием x между цифрами
- `tokenize(s)` — split по разделителям
- `levenshtein(a, b, max?)` — расстояние с early-exit
- `smartSearch(items, query, opts)` — ранжированный список + hits для подсветки
- `smartScoreOne(query, item, opts)` — единичный кандидат

## Тесты

```bash
npm test
```

Спека: `src/utils/smartSearch.spec.ts`.
