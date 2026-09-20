# RadonezhSklad — Развёртывание на офисном ПК

Инструкция для переноса проекта с dev-машины на офисный ПК (для демонстрации).
Предполагается: Windows + RDP + интернет.

---

## 0. Что нужно установить на офисном ПК

| Компонент | Версия | Ссылка |
|---|---|---|
| Git | latest | https://git-scm.com/ |
| Go | 1.22 | https://go.dev/dl/ |
| Node.js | LTS (20+) | https://nodejs.org/ |
| Docker Desktop | latest | https://www.docker.com/products/docker-desktop/ |
| PowerShell 7 | 7.4+ | https://github.com/PowerShell/PowerShell/releases |

После установки Docker Desktop — перезагрузить и убедиться, что `docker ps` работает.

---

## 1. Распаковка пакета

Скопировать на офисный ПК (через RDP/облако):

- `radonezh-deploy.zip` (~100 MB)

Распаковать в `D:\Radonezhsklad\transfer\` (или другое место):

    Expand-Archive -Path radonezh-deploy.zip -DestinationPath D:\Radonezhsklad\transfer

Внутри:

    transfer/
    ├── dumps/                ← 5 дампов БД (101 MB)
    ├── services-env/         ← 6 .env файлов
    ├── .secrets/             ← moysklad.token
    ├── infra/
    └── docker-compose.yml

---

## 2. Клонирование репозитория

    cd D:\Radonezhsklad\projects
    git clone https://github.com/AriseWrk/radonezhsklad.git
    cd radonezhsklad

Репозиторий ~5 MB, клонируется за 10-20 сек.

---

## 3. Копирование .env и секретов

Из распакованного пакета:

    cd D:\Radonezhsklad\projects\radonezhsklad
    New-Item -ItemType Directory -Force -Path D:\Radonezhsklad\.secrets | Out-Null
    Copy-Item D:\Radonezhsklad\transfer\.secrets\moysklad.token D:\Radonezhsklad\.secrets\moysklad.token

    foreach ($svc in "auth","product","warehouse","order","gateway","audit") {
        Copy-Item "D:\Radonezhsklad\transfer\services-env\$svc.env" ".\services\$svc\.env"
    }

---

## 4. Запуск инфраструктуры

    docker compose up -d

Проверить:

    docker ps

Должны быть запущены `rs_postgres`, `rs_redis`, `rs_rabbitmq`.

Подождать 5-10 сек, пока Postgres инициализируется.

---

## 5. Восстановление баз из дампов

    cd D:\Radonezhsklad\projects\radonezhsklad
    . .\devtools.ps1
    Set-ConsoleUtf8

Скопировать дампы из пакета в репозиторий:

    New-Item -ItemType Directory -Force -Path dumps | Out-Null
    Copy-Item D:\Radonezhsklad\transfer\dumps\*.sql dumps\

Применить дампы:

    .\scripts\db-restore.ps1

Скрипт восстановит все 5 БД и выведет sanity-статистику:

    documents ~49 000, internal_orders ~11 091, inventories 749, projects 2733

---

## 6. Установка зависимостей Go (первый запуск)

    cd services\auth      && go mod download
    cd ..\product         && go mod download
    cd ..\warehouse       && go mod download
    cd ..\order           && go mod download
    cd ..\gateway         && go mod download
    cd ..\audit           && go mod download

Или одной командой из корня (если есть PowerShell-скрипт — см. ниже).

---

## 7. Установка фронтенда

    cd web
    npm install

---

## 8. Запуск сервисов (7 терминалов)

Открыть 7 отдельных окон PowerShell в `D:\Radonezhsklad\projects\radonezhsklad`.

В каждом:

    . .\devtools.ps1
    Set-ConsoleUtf8

Затем:

| # | Терминал | Команда |
|---|---|---|
| 1 | services\auth | `go run .\cmd\api` |
| 2 | services\product | `go run .\cmd\api` |
| 3 | services\warehouse | `go run .\cmd\api` |
| 4 | services\order | `go run .\cmd\api` |
| 5 | services\gateway | `go run .\cmd\api` |
| 6 | services\audit | `go run .\cmd\api` |
| 7 | web | `npm run dev` |

Открыть http://localhost:5173 — войти admin@radonezh.local / qwerty123.

---

## 9. Быстрая проверка после развёртывания

    # API health
    curl http://localhost:8080/api/v1/health

    # Заказы
    curl http://localhost:8080/api/v1/internal-orders (с Bearer)

    # В браузере
    /internal-orders → 11 091 заказ, сверху 02070, 02069, 02068…
    /inventories → 749 записей, pager работает
    /documents → 49 485 документов, есть карточка

---

## 10. Если что-то не работает

| Проблема | Решение |
|---|---|
| `rs_postgres` не запускается | `docker compose down -v && docker compose up -d` (внимание: удалит volume) |
| Ошибка «database does not exist» | Проверить, что init-скрипт отработал: `docker exec rs_postgres psql -U radonezh -lqt` |
| Русский бьётся в `?????` | Запустить `. .\devtools.ps1; Set-ConsoleUtf8` в текущем окне |
| Порт занят (5433, 6380, 5672, 8080-8085) | Убить процесс или изменить в docker-compose/.env |
| Дампы не восстанавливаются | Проверить размер файлов в `dumps/` — должно быть ~101 MB |

---

## 11. Чек-лист (кратко)

- [ ] Установлены: Git, Go 1.22, Node LTS, Docker Desktop, PowerShell 7
- [ ] `git clone` репозитория
- [ ] Скопированы `.env` (6 шт) и `.secrets/moysklad.token`
- [ ] `docker compose up -d` → 3 контейнера running
- [ ] Скопированы дампы в `dumps/`
- [ ] `.\scripts\db-restore.ps1` → sanity OK
- [ ] `cd web; npm install`
- [ ] 7 терминалов запущены, frontend на :5173
- [ ] Логин admin@radonezh.local / qwerty123 работает
- [ ] /internal-orders показывает 11 091 заказ

---

## Готово к демо.