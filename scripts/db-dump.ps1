$ErrorActionPreference = 'Stop'
$utf8 = [Text.UTF8Encoding]::new($false)
$root = (Get-Location).Path
$dumpDir = Join-Path $root "dumps"
if (-not (Test-Path $dumpDir)) { New-Item -ItemType Directory -Path $dumpDir | Out-Null }

$databases = @(
    'radonezh_auth',
    'radonezh_product',
    'radonezh_warehouse',
    'radonezh_order',
    'radonezh_audit'
)

foreach ($db in $databases) {
    $out = Join-Path $dumpDir "$db.sql"
    Write-Host "Dumping $db ..." -ForegroundColor Cyan -NoNewline

    # pg_dump внутри контейнера + docker cp
    & docker exec rs_postgres bash -c "pg_dump -U radonezh --no-owner --no-acl --clean --if-exists $db" > $out

    if ($LASTEXITCODE -ne 0) { throw "pg_dump failed for $db" }
    $size = [Math]::Round((Get-Item $out).Length / 1MB, 2)
    Write-Host "  $size MB" -ForegroundColor Green
}

Write-Host "`n=== Итог ===" -ForegroundColor Cyan
Get-ChildItem $dumpDir -Filter *.sql | Select-Object Name, @{n='MB';e={[Math]::Round($_.Length/1MB,2)}} | Format-Table -AutoSize
$total = (Get-ChildItem $dumpDir -Filter *.sql | Measure-Object -Property Length -Sum).Sum
Write-Host "Total: $([Math]::Round($total/1MB,2)) MB" -ForegroundColor Yellow