param(
    [string[]]$Entities = @('internalorder','inventory')
)
$ErrorActionPreference = 'Stop'
[System.Threading.Thread]::CurrentThread.CurrentCulture = [System.Globalization.CultureInfo]::InvariantCulture

$root = Split-Path -Parent $PSScriptRoot
. (Join-Path $root 'devtools.ps1')

# 1) все ms-uuid товаров из нашей БД
Write-Host "[1/2] Loading known product ms-uuids..."
$raw = Invoke-SqlQuery -Db radonezh_product -Query "SELECT external_id::text FROM products WHERE external_id IS NOT NULL;"
$known = New-Object System.Collections.Generic.HashSet[string]
foreach ($ln in ($raw -split "`n")) {
    if ($ln -match '([0-9a-f-]{36})') { $null = $known.Add($Matches[1]) }
}
Write-Host "  known: $($known.Count)"

# 2) проход по каждому entity
$missing = New-Object System.Collections.Generic.List[object]

foreach ($ent in $Entities) {
    Write-Host "[2/2] Scanning /entity/$ent ..."
    $pageSize = 100
    $offset = 0
    $total = $null
    $fetched = 0
    $seen = New-Object System.Collections.Generic.HashSet[string]

    while ($true) {
        $q = @{ limit = $pageSize; offset = $offset; expand = 'positions' }
        $page = MsApi-Get -Path "/entity/$ent" -Query $q
        if ($null -eq $total) { $total = [int]$page.meta.size; Write-Host "  total: $total" }
        if (-not $page.rows -or $page.rows.Count -eq 0) { break }

        foreach ($d in $page.rows) {
            $fetched++
            $positions = @()
            if ($d.positions -and $d.positions.rows) { $positions = $d.positions.rows }
            $expected = 0
            if ($d.positions -and $d.positions.meta -and $d.positions.meta.size) { $expected = [int]$d.positions.meta.size }
            if ($expected -gt $positions.Count) {
                $positions = MsApi-GetAll -Path "/entity/$ent/$($d.id)/positions" -PageSize 1000
            }

            foreach ($p in $positions) {
                $prodMs = $null
                if ($p.assortment.meta.href -match '/(?:product|variant|consignment|service|bundle)/([0-9a-f-]{36})') { $prodMs = $Matches[1] }
                if ($prodMs -and -not $known.Contains($prodMs) -and -not $seen.Contains($prodMs)) {
                    $null = $seen.Add($prodMs)
                    $missing.Add([pscustomobject]@{
                        Entity   = $ent
                        DocId    = $d.id
                        DocName  = $d.name
                        ItemId   = $p.id
                        ProdMsId = $prodMs
                        ProdHref = $p.assortment.meta.href
                        ProdType = $p.assortment.meta.type
                        Qty      = $p.quantity
                    })
                }
            }
        }

        $offset += $page.rows.Count
        Write-Host "  fetched $fetched / $total  (missing so far: $($missing.Count))"
        if ($offset -ge $total) { break }
        Start-Sleep -Milliseconds 120
    }
}

Write-Host ""
Write-Host "=== MISSING PRODUCTS: $($missing.Count) ==="
foreach ($m in $missing) {
    Write-Host "  [$($m.Entity)] doc=$($m.DocName) item=$($m.ItemId) prodMs=$($m.ProdMsId) type=$($m.ProdType)"
}

