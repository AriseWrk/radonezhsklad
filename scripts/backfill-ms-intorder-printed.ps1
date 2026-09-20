$ErrorActionPreference = 'Stop'
[Threading.Thread]::CurrentThread.CurrentCulture = [System.Globalization.CultureInfo]::InvariantCulture
$utf8 = [Text.UTF8Encoding]::new($false)
$root = (Get-Location).Path

$sb = New-Object System.Text.StringBuilder
$pageSize = 1000
$offset = 0
$cnt = 0
while ($true) {
    $r = MsApi-Get '/entity/internalorder' @{ limit = $pageSize; offset = $offset }
    $rows = $r.rows
    if (-not $rows -or $rows.Count -eq 0) { break }
    foreach ($d in $rows) {
        $pr = if ($d.printed) { 't' } else { 'f' }
        [void]$sb.AppendLine("$($d.id)`t$pr")
        $cnt++
    }
    if ($rows.Count -lt $pageSize) { break }
    $offset += $pageSize
    Start-Sleep -Milliseconds 120
}
Write-Host "MS: $cnt rows, TSV $([Math]::Round($sb.Length/1KB,1)) KB" -ForegroundColor Cyan

$tmp = Join-Path $root "scripts\_printed.tsv"
[IO.File]::WriteAllText($tmp, $sb.ToString(), $utf8)

$sql = @(
  'CREATE TEMP TABLE tmp_printed (external_id UUID, is_printed TEXT);',
  '\copy tmp_printed FROM ''/tmp/printed.tsv'' WITH (FORMAT text)',
  'UPDATE internal_orders io',
  '   SET is_printed = (t.is_printed = ''t'')',
  '  FROM tmp_printed t',
  ' WHERE io.external_id = t.external_id;',
  'DROP TABLE tmp_printed;'
) -join "`n"

$sqlFile = Join-Path $root "scripts\_printed.sql"
[IO.File]::WriteAllText($sqlFile, $sql, $utf8)

& docker cp $tmp rs_postgres:/tmp/printed.tsv
& docker cp $sqlFile rs_postgres:/tmp/printed.sql
& docker exec rs_postgres psql -U radonezh -d radonezh_warehouse -f /tmp/printed.sql 2>&1 | Select-Object -Last 5

Remove-Item $tmp, $sqlFile -ErrorAction SilentlyContinue