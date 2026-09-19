param()
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
function Esc-Num($v) {
    if ($null -eq $v) { return '\N' }
    return ([string]$v)
}

$uomMap = @{
    '19f1edc0-fc42-4001-94cb-c9ec9c62ec10' = '08b31d0d-3ff2-4cff-af71-f732dc6c4fa0'
    'dfe54549-ec98-4fd7-ad1b-215162a9199a' = '1880831f-7fef-4b5d-a544-033d78436913'
    '2ec1170c-3f69-4409-87bb-c68e0011b275' = '59907d09-f042-48d6-a57e-8f03781f73c1'
    '250d5161-f815-4524-85b6-fd6afebc1ad6' = '1265637b-2352-4cf3-ab7f-73eebdcf0f23'
    '33f85108-3b2a-11ed-0a80-0d23000b1cc8' = '2ab1a02d-5bbc-442d-90a9-6a8aece3ef6a'
    'b8ed17f7-82b5-11e9-9109-f8fc000e6b18' = '5900f3c5-eebf-4e93-92d9-71f8cb5295b6'
    '4265ca4f-59c4-11f1-0a80-0e09000ddd0f' = '75cd817a-a57d-4c64-9d90-223187f0d78d'
}

Write-Host "=== Import MS archived products ==="

Write-Host "[1/4] Fetching archived products..."
$prods = MsApi-GetAll -Path '/entity/product' -Query @{ filter = 'archived=true' } -PageSize 100
Write-Host "  archived products: $($prods.Count)"

Write-Host "[2/4] Building TSV..."
$tsvDir = Join-Path $env:TEMP ("ms_arch_" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tsvDir -Force | Out-Null
$tsvPath = Join-Path $tsvDir 'products.tsv'

$lines2 = New-Object System.Collections.Generic.List[string]
foreach ($p in $prods) {
    $uomU = $null
    try { if ($p.uom.meta.href -match '/uom/([0-9a-f-]{36})') { $uomU = $Matches[1] } } catch {}
    $unitId = $null
    if ($uomU -and $uomMap.ContainsKey($uomU)) { $unitId = $uomMap[$uomU] }

    $salePrice = 0
    if ($p.salePrices -and $p.salePrices.Count -gt 0 -and $null -ne $p.salePrices[0].value) {
        $salePrice = [math]::Round(([double]$p.salePrices[0].value) / 100, 2)
    }
    $buyPrice = 0
    if ($p.buyPrice -and $null -ne $p.buyPrice.value) {
        $buyPrice = [math]::Round(([double]$p.buyPrice.value) / 100, 2)
    }

    $weight = 0; if ($null -ne $p.weight) { $weight = $p.weight }
    $volume = 0; if ($null -ne $p.volume) { $volume = $p.volume }

    $row = @(
        $p.id,
        (Esc-Text $p.name),
        (Esc-Text $p.code),
        (Esc-Text $p.description),
        (Esc-Num  $salePrice),
        (Esc-Num  $buyPrice),
        (Esc-Text $unitId),
        (Esc-Num  $weight),
        (Esc-Num  $volume),
        (Esc-Text $p.externalCode)
    ) -join $TAB
    $lines2.Add($row)
}
[IO.File]::WriteAllLines($tsvPath, $lines2, $utf8)
Write-Host "  rows: $($lines2.Count)"

$sql = @"
\set ON_ERROR_STOP on
BEGIN;
CREATE TEMP TABLE tmp_p (
    ms_id         uuid,
    name          varchar,
    code          varchar,
    description   text,
    price         numeric,
    cost_price    numeric,
    unit_id       uuid,
    weight        numeric,
    volume        numeric,
    external_code varchar
) ON COMMIT DROP;
\copy tmp_p FROM '/tmp/ms_arch.tsv' WITH (FORMAT text)

INSERT INTO products (
    name, sku, description, price, currency, is_archived,
    min_stock, cost_price, external_id, external_code, source, weight, volume, unit_id
)
SELECT
    t.name, NULLIF(t.code,''), NULLIF(t.description,''), COALESCE(t.price,0), 'RUB', true,
    0, COALESCE(t.cost_price,0), t.ms_id, t.external_code, 'moysklad',
    COALESCE(t.weight,0), COALESCE(t.volume,0), t.unit_id
FROM tmp_p t
ON CONFLICT (external_id) DO UPDATE SET
    name          = EXCLUDED.name,
    sku           = EXCLUDED.sku,
    description   = EXCLUDED.description,
    price         = EXCLUDED.price,
    cost_price    = EXCLUDED.cost_price,
    is_archived   = EXCLUDED.is_archived,
    external_code = EXCLUDED.external_code,
    unit_id       = EXCLUDED.unit_id,
    weight        = EXCLUDED.weight,
    volume        = EXCLUDED.volume,
    updated_at    = NOW();

SELECT COUNT(*) AS total_products FROM products WHERE source='moysklad';
SELECT COUNT(*) AS archived_products FROM products WHERE source='moysklad' AND is_archived=true;
COMMIT;
"@

Write-Host "[3/4] Copying TSV + SQL..."
docker cp $tsvPath rs_postgres:/tmp/ms_arch.tsv | Out-Null
$sqlLocal = Join-Path $env:TEMP ("ms_arch_" + [guid]::NewGuid().ToString('N') + '.sql')
[IO.File]::WriteAllText($sqlLocal, $sql, $utf8)
docker cp $sqlLocal rs_postgres:/tmp/ms_arch.sql | Out-Null

Write-Host "[4/4] Running upsert..."
& docker exec -i rs_postgres bash -c "psql -U radonezh -d radonezh_product -f /tmp/ms_arch.sql"

& docker exec -i rs_postgres rm -f /tmp/ms_arch.tsv /tmp/ms_arch.sql
Remove-Item $sqlLocal -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force $tsvDir -ErrorAction SilentlyContinue
Write-Host "=== DONE ==="
