# RadonezhSklad

Функциональная копия сервиса складского учёта (аналог МойСклад).

## Стек

- Backend: Go 1.22+, Gin
- Frontend: Vue 3 + TypeScript + Vite
- БД: PostgreSQL 16
- Кэш: Redis 7
- Очереди: RabbitMQ 3.13

## Быстрый старт

Инфраструктура (postgres, redis, rabbitmq):

```
docker compose up -d
```

Dev-окружение (все сервисы + фронт, с логами):

```powershell
cd D:\Radonezhsklad\projects\radonezhsklad
. .\devtools.ps1; Set-ConsoleUtf8
.\dev-up.ps1                    # поднять всё
.\dev-up.ps1 -Rebuild           # пересобрать Go-сервисы
.\dev-up.ps1 -Only X -Rebuild   # один сервис: auth|product|warehouse|order|audit|gateway|front
.\dev-down.ps1                  # остановить
```

Frontend: http://localhost:5173 (admin@radonezh.local / qwerty123)

## Порты

| Сервис    | Порт |
|-----------|------|
| gateway   | 8080 |
| auth      | 8081 |
| product   | 8082 |
| warehouse | 8083 |
| order     | 8084 |
| audit     | 8085 |
| front     | 5173 |

## Тесты

Frontend (vitest):

```powershell
cd web; npm test
```

## Умный поиск номенклатуры

`web/src/utils/smartSearch.ts` — fuzzy-поиск по name / sku / barcode / external_id.
Работает в форме создания внутреннего заказа; задуман как переиспользуемый для любых
списков номенклатуры (остатки, документы, справочники).

Что умеет:

- нормализация: lower-case, схлопывание разделителей, гомоглифы (латиница ↔ кириллица:
  е/а/о/с/р/х/у/к/м/т/н/в), знаки умножения × ✕ ✖ * x х → один канонический символ;
- скелет [a-z0-9] с вырезанием x между цифрами: 1x2x0.75 == 1×2×0.75 == 1х2х0.75 == 1 2 0 75;
- токены — AND по токенам запроса в любом порядке;
- fuzzy — Levenshtein ≤ 2 (при длине ≥ 5), ловит опечатки;
- ранжирование: exact → startsWith code → includes code → startsWith name → includes name → skeleton → tokens-AND → fuzzy; тай-брейк по длине имени.

Тесты: `web/src/utils/smartSearch.spec.ts` — 11 кейсов.
