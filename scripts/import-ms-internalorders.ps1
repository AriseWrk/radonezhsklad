param(
    [int]$Limit = 0,
    [switch]$KeepTsv
)
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
function Esc-Bool($v) {
    if ($null -eq $v) { return '\N' }
    if ($v) { return 'true' } else { return 'false' }
}
function Get-HrefUuid($obj, $kind) {
    if ($null -eq $obj) { return $null }
    $h = $null
    try { $h = $obj.meta.href } catch { return $null }
    if (-not $h) { return $null }
    if ($h -match "/(?:$kind)/([0-9a-f-]{36})") { return $Matches[1] }
    return $null
}

Write-Host "=== Import MS internalorder -> internal_orders (limit=$Limit) ==="

Write-Host "[1/5] Loading product map..."
$prodRaw = Invoke-SqlQuery -Db radonezh_product -Query "SELECT external_id::text || '|' || id::text FROM products WHERE external_id IS NOT NULL;"
$prodMap = @{}
foreach ($ln in ($prodRaw -split "`n")) {
    if ($ln -match '([0-9a-f-]{36})\|([0-9a-f-]{36})') { $prodMap[$Matches[1]] = $Matches[2] }
}
Write-Host "  loaded $($prodMap.Count) product mappings"

Write-Host "[2/5] Fetching documents (expand=positions)..."
$tsvDir   = Join-Path $env:TEMP ("ms_imp_" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tsvDir -Force | Out-Null
$docsTsv  = Join-Path $tsvDir 'docs.tsv'
$itemsTsv = Join-Path $tsvDir 'items.tsv'

$docLines  = New-Object System.Collections.Generic.List[string]
$itemLines = New-Object System.Collections.Generic.List[string]

$pageSize = 100
$offset = 0
$fetched = 0
$total = $null
$missing = 0
$extraPosCalls = 0

while ($true) {
    if ($Limit -gt 0 -and $fetched -ge $Limit) { break }
    $q = @{ limit = $pageSize; offset = $offset; expand = 'positions' }
    $page = MsApi-Get -Path '/entity/internalorder' -Query $q
    if ($null -eq $total) { $total = [int]$page.meta.size; Write-Host "  total: $total" }
    if (-not $page.rows -or $page.rows.Count -eq 0) { break }

    foreach ($d in $page.rows) {
        if ($Limit -gt 0 -and $fetched -ge $Limit) { break }
        $fetched++

        $storeU = Get-HrefUuid $d.store 'store'
        $orgU   = Get-HrefUuid $d.organization 'organization'
        $projU  = Get-HrefUuid $d.project 'project'

        $moment = $null
        if ($d.moment) { $moment = ($d.moment -replace '\.\d+$','') + '+03' }
        $planDate = $null
        if ($d.deliveryPlannedMoment) { $planDate = ($d.deliveryPlannedMoment -split ' ')[0] }
        $totalR = 0
        if ($null -ne $d.sum) { $totalR = [math]::Round(([double]$d.sum) / 100, 2) }

        $row = @(
            $d.id,
            (Esc-Text $d.name),
            (Esc-Text $moment),
            (Esc-Bool $d.applicable),
            (Esc-Num  $totalR),
            (Esc-Bool $d.vatEnabled),
            (Esc-Bool $d.vatIncluded),
            (Esc-Text $d.description),
            (Esc-Text $planDate),
            (Esc-Text $storeU),
            (Esc-Text $orgU),
            (Esc-Text $projU)
        ) -join $TAB
        $docLines.Add($row)

        $positions = @()
        if ($d.positions -and $d.positions.rows) { $positions = $d.positions.rows }
        $expected = 0
        if ($d.positions -and $d.positions.meta -and $d.positions.meta.size) { $expected = [int]$d.positions.meta.size }
        if ($expected -gt $positions.Count) {
            $extraPosCalls++
            $positions = MsApi-GetAll -Path "/entity/internalorder/$($d.id)/positions" -PageSize 1000
        }

        foreach ($p in $positions) {
            $prodMs = $null
            if ($p.assortment.meta.href -match '/(?:product|variant|consignment|service)/([0-9a-f-]{36})') { $prodMs = $Matches[1] }
            $prodId = $null
            if ($prodMs -and $prodMap.ContainsKey($prodMs)) { $prodId = $prodMap[$prodMs] }
            if (-not $prodId) { $missing++ }
            $priceR = 0
            if ($null -ne $p.price) { $priceR = [math]::Round(([double]$p.price) / 100, 2) }
            $sumR = 0
            if ($null -ne $p.price -and $null -ne $p.quantity) {
                $sumR = [math]::Round((([double]$p.price) * ([double]$p.quantity)) / 100, 2)
            }
            $irow = @(
                $d.id,
                $p.id,
                (Esc-Text $prodId),
                (Esc-Num  $p.quantity),
                (Esc-Num  $priceR),
                (Esc-Num  $p.vat),
                (Esc-Num  $sumR)
            ) -join $TAB
            $itemLines.Add($irow)
        }
    }

    $offset += $page.rows.Count
    Write-Host "  fetched $fetched / $total  (docs: $($docLines.Count), items: $($itemLines.Count))"
    if ($offset -ge $total) { break }
    Start-Sleep -Milliseconds 120
}

[IO.File]::WriteAllLines($docsTsv,  $docLines,  $utf8)
[IO.File]::WriteAllLines($itemsTsv, $itemLines, $utf8)
Write-Host "  docs rows:  $($docLines.Count)"
Write-Host "  item rows:  $($itemLines.Count)"
Write-Host "  missing product mappings: $missing"
Write-Host "  extra position calls:     $extraPosCalls"

Write-Host "[3/5] Copying TSVs to container..."
docker cp $docsTsv  rs_postgres:/tmp/ms_docs.tsv  | Out-Null
docker cp $itemsTsv rs_postgres:/tmp/ms_items.tsv | Out-Null

$tpl = Read-Utf8File -Path (Join-Path $root 'scripts\sql\upsert_internalorders.sql')
$sql = $tpl.Replace('{{DOCS_PATH}}',  '/tmp/ms_docs.tsv').Replace('{{ITEMS_PATH}}', '/tmp/ms_items.tsv')

$sqlLocal = Join-Path $env:TEMP ("ms_upsert_" + [guid]::NewGuid().ToString('N') + '.sql')
[IO.File]::WriteAllText($sqlLocal, $sql, $utf8)
docker cp $sqlLocal rs_postgres:/tmp/ms_upsert.sql | Out-Null

Write-Host "[4/5] Running upsert..."
& docker exec -i rs_postgres bash -c "psql -U radonezh -d radonezh_warehouse -f /tmp/ms_upsert.sql"

Write-Host "[5/5] Cleaning up..."
& docker exec -i rs_postgres rm -f /tmp/ms_docs.tsv /tmp/ms_items.tsv /tmp/ms_upsert.sql
Remove-Item $sqlLocal -ErrorAction SilentlyContinue
if (-not $KeepTsv) { Remove-Item -Recurse -Force $tsvDir -ErrorAction SilentlyContinue } else { Write-Host "TSVs: $tsvDir" }
Write-Host "=== DONE ==="
