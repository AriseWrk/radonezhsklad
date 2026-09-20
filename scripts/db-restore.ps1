$ErrorActionPreference = 'Stop'
$utf8 = [Text.UTF8Encoding]::new($false)
$root = Split-Path $PSScriptRoot -Parent
$dumpsDir = Join-Path $root "deploy-package\dumps"
if (-not (Test-Path $dumpsDir)) { $dumpsDir = Join-Path $root "dumps" }
if (-not (Test-Path $dumpsDir)) { throw "dumps dir not found" }

$databases = @(
    @{ db = 'radonezh_auth';      file = 'radonezh_auth.sql' },
    @{ db = 'radonezh_product';   file = 'radonezh_product.sql' },
    @{ db = 'radonezh_warehouse'; file = 'radonezh_warehouse.sql' },
    @{ db = 'radonezh_order';     file = 'radonezh_order.sql' },
    @{ db = 'radonezh_audit';     file = 'radonezh_audit.sql' }
)

Write-Host "checking rs_postgres ..." -ForegroundColor Cyan
$running = docker ps --filter "name=rs_postgres" --filter "status=running" --format "{{.Names}}"
if (-not $running) { throw "rs_postgres is not running. Start docker compose first." }
Write-Host "  OK" -ForegroundColor Green

foreach ($item in $databases) {
    $db = $item.db; $fn = $item.file
    $file = Join-Path $dumpsDir $fn
    if (-not (Test-Path $file)) { Write-Host "SKIP $db (no dump $fn)"; continue }
    $size = [Math]::Round((Get-Item $file).Length / 1MB, 2)
    Write-Host "restoring $db ($size MB) ... " -ForegroundColor Cyan -NoNewline
    $tmp = "/tmp/restore-$db.sql"
    docker cp $file "rs_postgres:$tmp" | Out-Null
    $log = docker exec rs_postgres psql -U radonezh -d $db -f $tmp 2>&1
    if ($LASTEXITCODE -ne 0) { Write-Host "FAIL" -ForegroundColor Red; $log | Select-Object -Last 20; throw "restore failed: $db" }
    Write-Host "OK" -ForegroundColor Green
}

Write-Host "`n=== sanity ===" -ForegroundColor Cyan
docker exec rs_postgres psql -U radonezh -d radonezh_warehouse -c "SELECT (SELECT COUNT(*) FROM documents) documents, (SELECT COUNT(*) FROM internal_orders) internal_orders, (SELECT COUNT(*) FROM inventories) inventories, (SELECT COUNT(*) FROM projects) projects;"
docker exec rs_postgres psql -U radonezh -d radonezh_product -c "SELECT COUNT(*) products FROM products;"
docker exec rs_postgres psql -U radonezh -d radonezh_auth -c "SELECT COUNT(*) users FROM users;"