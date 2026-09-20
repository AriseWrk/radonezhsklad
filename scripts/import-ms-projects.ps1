$ErrorActionPreference = 'Stop'
[Threading.Thread]::CurrentThread.CurrentCulture = [System.Globalization.CultureInfo]::InvariantCulture
$utf8 = [Text.UTF8Encoding]::new($false)
$root = (Get-Location).Path

# Забираем все проекты (limit<=100 чтобы positions не рубились, но тут без positions — можно 1000)
$pageSize = 1000
$offset = 0
$all = @()
while ($true) {
    $r = MsApi-Get '/entity/project' @{ limit = $pageSize; offset = $offset }
    $rows = $r.rows
    if (-not $rows -or $rows.Count -eq 0) { break }
    foreach ($d in $rows) { $all += $d }
    if ($rows.Count -lt $pageSize) { break }
    $offset += $pageSize
    Start-Sleep -Milliseconds 120
}
Write-Host "MS projects: $($all.Count)" -ForegroundColor Cyan

# Генерируем SQL с UPSERT
$sql = New-Object System.Text.StringBuilder
[void]$sql.AppendLine('BEGIN;')
foreach ($p in $all) {
    $id = $p.id
    $name = if ($p.name) { $p.name } else { '' }
    # Экранирование одинарных кавычек
    $nameEsc = $name.Replace("'", "''")
    $arch = if ($p.archived) { 'TRUE' } else { 'FALSE' }
    [void]$sql.AppendLine("INSERT INTO projects (external_id, name, archived) VALUES ('$id', '$nameEsc', $arch) ON CONFLICT (external_id) DO UPDATE SET name = EXCLUDED.name, archived = EXCLUDED.archived, updated_at = NOW();")
}
[void]$sql.AppendLine('COMMIT;')

$tmp = Join-Path $root "scripts\_projects.sql"
[IO.File]::WriteAllText($tmp, $sql.ToString(), $utf8)
Write-Host "SQL written: $tmp  ($([Math]::Round((Get-Item $tmp).Length/1KB,1)) KB)"

# COPY в контейнер
& docker cp $tmp rs_postgres:/tmp/projects.sql
& docker exec rs_postgres psql -U radonezh -d radonezh_warehouse -f /tmp/projects.sql 2>&1 | Select-Object -Last 3

Remove-Item $tmp -ErrorAction SilentlyContinue