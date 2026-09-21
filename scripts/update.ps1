<#
.SYNOPSIS
    Обновляет проект с GitHub и применяет миграции.

.DESCRIPTION
    1. git fetch + сравнение HEAD с origin/main
    2. если изменений нет — выход
    3. git pull
    4. если менялись миграции — запуск migrate.ps1
    5. если менялся web/package.json — npm install
    6. если менялись *.go — напоминание перезапустить сервисы

.PARAMETER Force
    Выполнить, даже если изменений нет.
#>
param([switch]$Force)

$ErrorActionPreference = 'Stop'
$root = (Get-Location).Path

Write-Host "`n=== RadonezhSklad: update ===" -ForegroundColor Magenta

# --- 1. Проверка чистоты рабочей копии ---
$dirty = git status --short
if ($dirty) {
    Write-Host "`n⚠ Есть незакоммиченные изменения:" -ForegroundColor Yellow
    Write-Host $dirty
    Write-Host "`nЧто сделать: git stash или git commit, потом повторить." -ForegroundColor Yellow
    return
}

# --- 2. Fetch + сравнение ---
Write-Host "`n[1/5] git fetch ..." -ForegroundColor Cyan
git fetch origin --quiet
$local  = git rev-parse HEAD
$remote = git rev-parse origin/main

if ($local -eq $remote -and -not $Force) {
    Write-Host "  Изменений нет. HEAD = origin/main = $($local.Substring(0,7))" -ForegroundColor Green
    return
}

Write-Host "  local  : $($local.Substring(0,7))"
Write-Host "  remote : $($remote.Substring(0,7))"

# --- 3. Что изменилось ---
Write-Host "`n[2/5] Изменённые файлы:" -ForegroundColor Cyan
$changed = git diff --name-only HEAD origin/main
$changed | ForEach-Object { "  $_" }

# --- 4. git pull ---
Write-Host "`n[3/5] git pull ..." -ForegroundColor Cyan
git pull origin main
$newHead = git rev-parse HEAD
Write-Host "  new HEAD: $($newHead.Substring(0,7))" -ForegroundColor Green

# --- 5. Миграции ---
$migrationsChanged = $changed | Where-Object { $_ -match 'migrations/.*\.sql$' }
if ($migrationsChanged) {
    Write-Host "`n[4/5] Обнаружены новые миграции — применяю ..." -ForegroundColor Cyan
    & (Join-Path $root "scripts\migrate.ps1")
} else {
    Write-Host "`n[4/5] Миграции не менялись." -ForegroundColor DarkGray
}

# --- 6. npm install? ---
$npmChanged = $changed | Where-Object { $_ -match '^web/(package\.json|package-lock\.json)$' }
if ($npmChanged) {
    Write-Host "`n[5/5] web/package.json менялся — npm install ..." -ForegroundColor Cyan
    Push-Location (Join-Path $root "web")
    npm install
    Pop-Location
} else {
    Write-Host "`n[5/5] Frontend-зависимости не менялись." -ForegroundColor DarkGray
}

# --- Что делать дальше ---
$goChanged = $changed | Where-Object { $_ -match '\.go$' }
$webChanged = $changed | Where-Object { $_ -match '^web/' }

Write-Host "`n=== Что дальше ===" -ForegroundColor Magenta
if ($goChanged) {
    Write-Host "⚠ Менялись Go-файлы — надо перезапустить сервисы:" -ForegroundColor Yellow
    Write-Host "  В окнах auth/product/warehouse/order/gateway/audit: Ctrl+C, затем 'go run .\cmd\api'" -ForegroundColor Yellow
} else {
    Write-Host "✓ Go-код не менялся — сервисы перезапускать не нужно" -ForegroundColor Green
}
if ($webChanged) {
    Write-Host "✓ Vite подхватит изменения web/ автоматически (HMR)" -ForegroundColor Green
}

Write-Host "`nГотово." -ForegroundColor Green