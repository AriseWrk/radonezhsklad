param([switch]$SkipMovements, [switch]$SkipBalances)
$ErrorActionPreference = 'Stop'
[System.Threading.Thread]::CurrentThread.CurrentCulture = [System.Globalization.CultureInfo]::InvariantCulture

$root = Split-Path -Parent $PSScriptRoot
. (Join-Path $root 'devtools.ps1')

$utf8 = New-Object System.Text.UTF8Encoding $false
$TAB  = [char]9

function Esc-Text($v) {
    if ([string]::IsNullOrEmpty($v)) { return '\N' }
    $r = [string]$v
    $r = $r.Replace('\\', '\\\\')
    $r = $r.Replace("`t", '\t')
    $r = $r.Replace("`n", '\n')
    $r = $r.Replace("`r", '\r')
    return $r
}

# ============== PHASE 1: MOVEMENTS ==============
if (-not $SkipMovements) {
    Write-Host "[1/2] Rebuilding stock_movements from documents..."

    $sql = @"
BEGIN;
DELETE FROM stock_movements;

-- расход с warehouse
INSERT INTO stock_movements (warehouse_id, product_id, document_id, quantity_delta, created_at)
SELECT d.warehouse_id, di.product_id, d.id,
       CASE d.type
         WHEN 'receipt'  THEN  di.quantity
         WHEN 'shipment' THEN -di.quantity
         WHEN 'writeoff' THEN -di.quantity
         WHEN 'transfer' THEN -di.quantity
       END,
       COALESCE(d.posted_at, d.created_at)
FROM documents d
JOIN document_items di ON di.document_id = d.id
WHERE d.status='posted' AND di.product_id IS NOT NULL
  AND d.type IN ('receipt','shipment','writeoff','transfer')
  AND d.warehouse_id IS NOT NULL;

-- приход на target при transfer
INSERT INTO stock_movements (warehouse_id, product_id, document_id, quantity_delta, created_at)
SELECT d.target_warehouse_id, di.product_id, d.id, di.quantity,
       COALESCE(d.posted_at, d.created_at)
FROM documents d
JOIN document_items di ON di.document_id = d.id
WHERE d.status='posted' AND di.product_id IS NOT NULL
  AND d.type='transfer' AND d.target_warehouse_id IS NOT NULL;

SELECT COUNT(*) AS movements, COUNT(DISTINCT document_id) AS docs FROM stock_movements;
COMMIT;
"@
    Invoke-SqlQuery -Db radonezh_warehouse -Query $sql
}

# ============== PHASE 2: BALANCES from MS report ==============
if (-not $SkipBalances) {
    Write-Host "[2/2] Fetching stock balances from MS /report/stock/all..."

    # 1) маппинг ms-uuid -> our uuid
    Write-Host "  [1/4] Loading maps..."
    $prodRaw = Invoke-SqlQuery -Db radonezh_product -Query "SELECT external_id::text || '|' || id::text FROM products WHERE external_id IS NOT NULL;"
    $prodMap = @{}
    foreach ($ln in ($prodRaw -split "`n")) {
        if ($ln -match '([0-9a-f-]{36})\|([0-9a-f-]{36})') { $prodMap[$Matches[1]] = $Matches[2] }
    }
    $whRaw = Invoke-SqlQuery -Db radonezh_warehouse -Query "SELECT external_id::text || '|' || id::text FROM warehouses WHERE external_id IS NOT NULL;"
    $whMap = @{}
    foreach ($ln in ($whRaw -split "`n")) {
        if ($ln -match '([0-9a-f-]{36})\|([0-9a-f-]{36})') { $whMap[$Matches[1]] = $Matches[2] }
    }
    Write-Host "    products: $($prodMap.Count)  warehouses: $($whMap.Count)"

    # 2) обходим склады, тянем отчёт
    Write-Host "  [2/4] Fetching per-warehouse stock..."
    $tsvDir = Join-Path $env:TEMP ("ms_stock_" + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tsvDir -Force | Out-Null
    $tsvPath = Join-Path $tsvDir 'balances.tsv'
    $rows = New-Object System.Collections.Generic.List[string]

    $scanned = 0; $withData = 0; $totalRows = 0; $missingProd = 0
    foreach ($msWhUuid in $whMap.Keys) {
        $scanned++
        $ourWhId = $whMap[$msWhUuid]
        $href = "https://api.moysklad.ru/api/remap/1.2/entity/store/$msWhUuid"
        $pageSize = 1000
        $offset = 0
        $whRowCount = 0
        while ($true) {
            $q = @{ limit = $pageSize; offset = $offset; filter = "store=$href" }
            $page = MsApi-Get -Path '/report/stock/all' -Query $q
            if (-not $page.rows -or $page.rows.Count -eq 0) { break }

            foreach ($r in $page.rows) {
                $msProdUuid = $null
                if ($r.meta.href -match '/product/([0-9a-f-]{36})') { $msProdUuid = $Matches[1] }
                if (-not $msProdUuid -or -not $prodMap.ContainsKey($msProdUuid)) { $missingProd++; continue }
                $ourProdId = $prodMap[$msProdUuid]
                $qty = 0
                if ($null -ne $r.stock) { $qty = $r.stock }
                $rows.Add(@($ourWhId, $ourProdId, $qty) -join $TAB)
                $whRowCount++
                $totalRows++
            }
            if ($page.rows.Count -lt $pageSize) { break }
            $offset += $page.rows.Count
            Start-Sleep -Milliseconds 120
        }
        if ($whRowCount -gt 0) { $withData++ }
        if ($scanned % 20 -eq 0) { Write-Host "    scanned $scanned / $($whMap.Count)  rows=$totalRows  nonEmpty=$withData" }
        Start-Sleep -Milliseconds 100
    }

    Write-Host "  rows: $totalRows  nonEmpty warehouses: $withData / $scanned  missing products: $missingProd"
    [IO.File]::WriteAllLines($tsvPath, $rows, $utf8)

    # 3) COPY + INSERT
    Write-Host "  [3/4] Copying TSV..."
    docker cp $tsvPath rs_postgres:/tmp/ms_balances.tsv | Out-Null

    $sql = @"
\set ON_ERROR_STOP on
BEGIN;
DELETE FROM stock_balances;

CREATE TEMP TABLE tmp_b (
    warehouse_id uuid,
    product_id   uuid,
    quantity     numeric
) ON COMMIT DROP;
\copy tmp_b FROM '/tmp/ms_balances.tsv' WITH (FORMAT text)

INSERT INTO stock_balances (warehouse_id, product_id, quantity, updated_at)
SELECT warehouse_id, product_id, SUM(quantity), NOW()
FROM tmp_b
GROUP BY warehouse_id, product_id
HAVING SUM(quantity) <> 0;

SELECT
  (SELECT COUNT(*) FROM stock_balances) AS balances,
  (SELECT COUNT(*) FROM stock_balances WHERE quantity < 0) AS negative,
  (SELECT COALESCE(SUM(quantity),0) FROM stock_balances) AS total_qty;
COMMIT;
"@
    $sqlLocal = Join-Path $env:TEMP ("ms_stock_" + [guid]::NewGuid().ToString('N') + '.sql')
    [IO.File]::WriteAllText($sqlLocal, $sql, $utf8)
    docker cp $sqlLocal rs_postgres:/tmp/ms_stock.sql | Out-Null

    Write-Host "  [4/4] Running upsert..."
    & docker exec -i rs_postgres bash -c "psql -U radonezh -d radonezh_warehouse -f /tmp/ms_stock.sql"

    & docker exec -i rs_postgres rm -f /tmp/ms_balances.tsv /tmp/ms_stock.sql
    Remove-Item $sqlLocal -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force $tsvDir -ErrorAction SilentlyContinue
}

Write-Host "=== DONE ==="
