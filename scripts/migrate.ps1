<#
.SYNOPSIS
    Применяет неприменённые миграции к БД проекта RadonezhSklad.
    Работает через локальный psql (если есть) или через docker exec rs_postgres.
#>
param(
    [switch]$Init,
    [string]$Db,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

# --- Режим: local psql или docker ---
$psqlCandidates = @(
    'C:\Program Files\PostgreSQL\16\bin\psql.exe',
    'C:\Program Files\PostgreSQL\15\bin\psql.exe',
    'C:\Program Files\PostgreSQL\14\bin\psql.exe'
)
$psqlLocal = $psqlCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $psqlLocal) {
    $cmd = Get-Command psql -ErrorAction SilentlyContinue
    if ($cmd) { $psqlLocal = $cmd.Source }
}

$useDocker = $false
if (-not $psqlLocal) {
    $running = docker ps --filter "name=rs_postgres" --filter "status=running" --format "{{.Names}}" 2>$null
    if ($running -eq 'rs_postgres') { $useDocker = $true; Write-Host "mode: docker exec rs_postgres" -ForegroundColor DarkGray }
    else { throw 'psql not found and rs_postgres container not running.' }
} else {
    Write-Host "mode: local psql $psqlLocal" -ForegroundColor DarkGray
    $env:PGPASSWORD = 'radonezh_dev_pass'
    $env:PGCLIENTENCODING = 'UTF8'
}

$root = (Get-Location).Path

$svcDbMap = [ordered]@{
    'auth'      = 'radonezh_auth'
    'product'   = 'radonezh_product'
    'warehouse' = 'radonezh_warehouse'
    'order'     = 'radonezh_order'
    'audit'     = 'radonezh_audit'
}

# --- Выполнить SQL (строкой) ---
function Invoke-Sql {
    param([string]$db, [string]$sql)
    if ($useDocker) {
        $out = docker exec rs_postgres psql -U radonezh -d $db -v ON_ERROR_STOP=1 -t -A --set=client_min_messages=error -c $sql 2>&1
    } else {
        $prev = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
        try { $out = & $psqlLocal -U radonezh -h localhost -d $db -v ON_ERROR_STOP=1 -t -A --set=client_min_messages=error -c $sql 2>&1 }
        finally { $ErrorActionPreference = $prev }
    }
    if ($LASTEXITCODE -ne 0) { throw "psql failed in $db : $($out | Out-String)" }
    $clean = @(); foreach ($l in $out) { if ($l -isnot [System.Management.Automation.ErrorRecord]) { $clean += $l } }
    return ($clean -join "`n")
}

# --- Выполнить SQL-файл ---
function Invoke-SqlFile {
    param([string]$db, [string]$file)
    if ($useDocker) {
        $tmp = "/tmp/mig-$(Get-Random).sql"
        docker cp $file "rs_postgres:$tmp" | Out-Null
        $out = docker exec rs_postgres psql -U radonezh -d $db -v ON_ERROR_STOP=1 --set=client_min_messages=error -f $tmp 2>&1
        if ($LASTEXITCODE -ne 0) { Write-Host ($out | Out-String) -ForegroundColor Red; throw "migration failed: $file" }
    } else {
        $prev = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
        try { $out = & $psqlLocal -U radonezh -h localhost -d $db -v ON_ERROR_STOP=1 --set=client_min_messages=error -f $file 2>&1 }
        finally { $ErrorActionPreference = $prev }
        if ($LASTEXITCODE -ne 0) { Write-Host ($out | Out-String) -ForegroundColor Red; throw "migration failed: $file" }
    }
}

function Process-Service {
    param([string]$svc, [string]$db)
    Write-Host "`n=== $svc -> $db ===" -ForegroundColor Cyan

    $dir = Join-Path $root "services\$svc\migrations"
    if (-not (Test-Path $dir)) { Write-Host "  нет папки $dir" -ForegroundColor DarkGray; return }

    Invoke-Sql $db "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());" | Out-Null

    $applied = @{}
    $rows = Invoke-Sql $db "SELECT version FROM schema_migrations;"
    foreach ($r in ($rows -split "`n")) { $v = $r.Trim(); if ($v) { $applied[$v] = $true } }

    $files = Get-ChildItem $dir -Filter '*.sql' | Sort-Object Name
    if ($files.Count -eq 0) { Write-Host "  нет миграций" -ForegroundColor DarkGray; return }

    if ($Init) {
        Write-Host "  -Init: помечаем всё как применённое" -ForegroundColor Yellow
        foreach ($f in $files) {
            if ($applied.ContainsKey($f.Name)) { Write-Host "    skip  $($f.Name)" -ForegroundColor DarkGray; continue }
            if ($DryRun) { Write-Host "    would mark: $($f.Name)" -ForegroundColor Yellow; continue }
            $ver = $f.Name.Replace("'", "''")
            Invoke-Sql $db "INSERT INTO schema_migrations(version) VALUES ('$ver') ON CONFLICT DO NOTHING;" | Out-Null
            Write-Host "    mark  $($f.Name)" -ForegroundColor DarkYellow
        }
        return
    }

    $newCount = 0
    foreach ($f in $files) {
        if ($applied.ContainsKey($f.Name)) { Write-Host "    skip  $($f.Name)" -ForegroundColor DarkGray; continue }
        if ($DryRun) { Write-Host "    WOULD APPLY: $($f.Name)" -ForegroundColor Yellow; $newCount++; continue }
        Write-Host "    apply $($f.Name) ..." -ForegroundColor Green -NoNewline
        Invoke-SqlFile $db $f.FullName
        $ver = $f.Name.Replace("'", "''")
        Invoke-Sql $db "INSERT INTO schema_migrations(version) VALUES ('$ver') ON CONFLICT DO NOTHING;" | Out-Null
        Write-Host " OK" -ForegroundColor Green
        $newCount++
    }
    if ($newCount -eq 0) { Write-Host "  всё применено" -ForegroundColor DarkGray }
    else { Write-Host "  применено новых: $newCount" -ForegroundColor Cyan }
}

Write-Host "`nmode:  $(if($Init){'INIT'}elseif($DryRun){'DRY RUN'}else{'APPLY'})" -ForegroundColor Magenta
Write-Host "root:  $root" -ForegroundColor DarkGray

foreach ($svc in $svcDbMap.Keys) {
    $targetDb = $svcDbMap[$svc]
    if ($Db -and $Db -ne $targetDb) { continue }
    Process-Service $svc $targetDb
}

Write-Host "`nГотово." -ForegroundColor Green