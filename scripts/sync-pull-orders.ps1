# sync-pull-orders.ps1
# Синхронизация internal_orders из МС в нашу БД.
# Тянет только то, что изменилось в МС с прошлого успешного запуска.
#
# Использование:
#   .\scripts\sync-pull-orders.ps1             # с прошлой метки (или -7 дней, если метки нет)
#   .\scripts\sync-pull-orders.ps1 -From "2026-09-15 00:00:00"   # явно с даты
#   .\scripts\sync-pull-orders.ps1 -Days 30    # за последние N дней (перезаписывает метку)
#   .\scripts\sync-pull-orders.ps1 -Reset      # сбросить метку
#
# Выводит строки "[PROGRESS] fetched=N total=M" — их парсит Go-обёртка (кнопка в UI).
# В конце — "[DONE] ok" или "[FAIL] <текст>".

param(
    [string]$From = '',
    [int]$Days = 0,
    [switch]$Reset
)

$ErrorActionPreference = 'Stop'

# Офис без Docker: если есть нативный psql и не задан RS_SQL_BACKEND — использовать psql
if (-not $env:RS_SQL_BACKEND) {
    $psqlCandidates = @(
        'C:\Program Files\PostgreSQL\17\bin\psql.exe',
        'C:\Program Files\PostgreSQL\16\bin\psql.exe',
        'C:\Program Files\PostgreSQL\15\bin\psql.exe',
        'C:\Program Files\PostgreSQL\14\bin\psql.exe'
    )
    foreach ($c in $psqlCandidates) {
        if (Test-Path $c) { $env:RS_SQL_BACKEND = 'psql'; break }
    }
}
[System.Threading.Thread]::CurrentThread.CurrentCulture = [System.Globalization.CultureInfo]::InvariantCulture

$root = Split-Path -Parent $PSScriptRoot
. (Join-Path $root 'devtools.ps1')

$stateDir = Join-Path $root 'dumps'
if (-not (Test-Path $stateDir)) { New-Item -ItemType Directory -Path $stateDir -Force | Out-Null }
$stateFile = Join-Path $stateDir '.last-pull-orders'

if ($Reset -and (Test-Path $stateFile)) {
    Remove-Item -LiteralPath $stateFile -Force
    Write-Host "[INFO] метка сброшена"
}

$since = $null
if ($From) {
    $since = $From
    Write-Host "[INFO] источник: параметр -From $since"
} elseif ($Days -gt 0) {
    $since = (Get-Date).AddDays(-$Days).ToString('yyyy-MM-dd HH:mm:ss')
    Write-Host "[INFO] источник: -Days $Days => since=$since"
} elseif (Test-Path $stateFile) {
    $since = (Get-Content -LiteralPath $stateFile -Raw).Trim()
    Write-Host "[INFO] источник: метка $since"
} else {
    $since = (Get-Date).AddDays(-7).ToString('yyyy-MM-dd HH:mm:ss')
    Write-Host "[INFO] метки нет — берём за последние 7 дней (since=$since)"
}

if (-not $since) { throw "не удалось определить since" }

$startedAt = Get-Date
$startedAtStr = $startedAt.ToString('yyyy-MM-dd HH:mm:ss')
Write-Host "[INFO] sync started at $startedAtStr, since=$since"

try {
    & (Join-Path $PSScriptRoot 'import-ms-internalorders.ps1') -Since $since
    if ($LASTEXITCODE -ne 0) { throw "import-ms-internalorders.ps1 exit $LASTEXITCODE" }
} catch {
    Write-Host "[FAIL] $_"
    exit 1
}

# Метку пишем ДО начала — так, чтобы повторный запуск не потерял изменения,
# произошедшие во время импорта. Оптимизация: можно и на startedAt.
[IO.File]::WriteAllText($stateFile, $startedAtStr, (New-Object System.Text.UTF8Encoding $false))
Write-Host "[INFO] метка обновлена: $startedAtStr"

Write-Host "[DONE] ok"