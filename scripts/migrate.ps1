<#
.SYNOPSIS
    Применяет неприменённые миграции к БД проекта RadonezhSklad.
#>
param(
    [switch]$Init,
    [string]$Db,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

$psqlCandidates = @(
    'C:\Program Files\PostgreSQL\16\bin\psql.exe',
    'C:\Program Files\PostgreSQL\15\bin\psql.exe',
    'C:\Program Files\PostgreSQL\14\bin\psql.exe'
)
$psql = $psqlCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $psql) {
    $cmd = Get-Command psql -ErrorAction SilentlyContinue
    if ($cmd) { $psql = $cmd.Source } else { throw 'psql not found.' }
}
Write-Host "psql: $psql" -ForegroundColor DarkGray

$env:PGPASSWORD = 'radonezh_dev_pass'
$env:PGCLIENTENCODING = 'UTF8'

$root = (Get-Location).Path

$svcDbMap = [ordered]@{
    'auth'      = 'radonezh_auth'
    'product'   = 'radonezh_product'
    'warehouse' = 'radonezh_warehouse'
    'order'     = 'radonezh_order'
    'audit'     = 'radonezh_audit'
}

# --- Обёртка: временно ставит ErrorActionPreference=Continue,
#     чтобы NOTICE от psql в stderr не бросал NativeCommandError.
function Invoke-Psql-Command {
    param([string]$db, [string]$sql, [switch]$AsFile)
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        if ($AsFile) {
            $out = & $psql -U radonezh -h localhost -d $db -v ON_ERROR_STOP=1 --set=client_min_messages=error -f $sql 2>&1
        } else {
            $out = & $psql -U radonezh -h localhost -d $db -v ON_ERROR_STOP=1 -t -A --set=client_min_messages=error -c $sql 2>&1
        }
        $ec = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $prev
    }
    # Отфильтруем ErrorRecord (stderr) — оставим только строки stdout
    $clean = @()
    foreach ($line in $out) {
        if ($line -is [System.Management.Automation.ErrorRecord]) { continue }
        $clean += $line
    }
    return @{ ExitCode = $ec; Output = ($clean -join "`n") }
}

function Process-Service {
    param([string]$svc, [string]$db)

    Write-Host "`n=== $svc -> $db ===" -ForegroundColor Cyan

    $dir = Join-Path $root "services\$svc\migrations"
    if (-not (Test-Path $dir)) {
        Write-Host "  нет папки $dir — пропускаем" -ForegroundColor DarkGray
        return
    }

    $r = Invoke-Psql-Command $db "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"
    if ($r.ExitCode -ne 0) { throw "create schema_migrations failed in $db : $($r.Output)" }

    $applied = @{}
    $r = Invoke-Psql-Command $db "SELECT version FROM schema_migrations;"
    if ($r.ExitCode -ne 0) { throw "select schema_migrations failed in $db : $($r.Output)" }
    foreach ($v in ($r.Output -split "`n")) {
        $v = $v.Trim()
        if ($v) { $applied[$v] = $true }
    }

    $files = Get-ChildItem $dir -Filter '*.sql' | Sort-Object Name
    if ($files.Count -eq 0) {
        Write-Host "  нет файлов миграций" -ForegroundColor DarkGray
        return
    }

    if ($Init) {
        Write-Host "  режим -Init: помечаем все как применённые" -ForegroundColor Yellow
        foreach ($f in $files) {
            if ($applied.ContainsKey($f.Name)) {
                Write-Host "    skip  $($f.Name)" -ForegroundColor DarkGray
                continue
            }
            if ($DryRun) {
                Write-Host "    would mark: $($f.Name)" -ForegroundColor Yellow
                continue
            }
            $ver = $f.Name.Replace("'", "''")
            $r = Invoke-Psql-Command $db "INSERT INTO schema_migrations(version) VALUES ('$ver') ON CONFLICT DO NOTHING;"
            if ($r.ExitCode -ne 0) { throw "insert failed: $($r.Output)" }
            Write-Host "    mark  $($f.Name)" -ForegroundColor DarkYellow
        }
        return
    }

    $newCount = 0
    foreach ($f in $files) {
        if ($applied.ContainsKey($f.Name)) {
            Write-Host "    skip  $($f.Name)" -ForegroundColor DarkGray
            continue
        }
        if ($DryRun) {
            Write-Host "    WOULD APPLY: $($f.Name)" -ForegroundColor Yellow
            $newCount++
            continue
        }
        Write-Host "    apply $($f.Name) ..." -ForegroundColor Green -NoNewline
        $r = Invoke-Psql-Command $db $f.FullName -AsFile
        if ($r.ExitCode -ne 0) {
            Write-Host " FAIL" -ForegroundColor Red
            Write-Host $r.Output -ForegroundColor Red
            throw "migration failed: $($f.FullName)"
        }
        $ver = $f.Name.Replace("'", "''")
        $r = Invoke-Psql-Command $db "INSERT INTO schema_migrations(version) VALUES ('$ver') ON CONFLICT DO NOTHING;"
        if ($r.ExitCode -ne 0) { throw "insert failed: $($r.Output)" }
        Write-Host " OK" -ForegroundColor Green
        $newCount++
    }

    if ($newCount -eq 0) {
        Write-Host "  всё применено, ничего нового" -ForegroundColor DarkGray
    } else {
        Write-Host "  применено новых: $newCount" -ForegroundColor Cyan
    }
}

Write-Host "`nmode:  $(if($Init){'INIT'}elseif($DryRun){'DRY RUN'}else{'APPLY'})" -ForegroundColor Magenta
Write-Host "root:  $root" -ForegroundColor DarkGray

foreach ($svc in $svcDbMap.Keys) {
    $targetDb = $svcDbMap[$svc]
    if ($Db -and $Db -ne $targetDb) { continue }
    Process-Service $svc $targetDb
}

Write-Host "`nГотово." -ForegroundColor Green