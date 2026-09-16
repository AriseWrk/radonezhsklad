# Restore-Stock.ps1 — восстановление остатков из XLS МойСклад
param(
    [Parameter(Mandatory=$true)] [string]$XlsPath,
    [string]$ApiBase     = "http://localhost:8080/api/v1",
    [string]$Email       = "admin@radonezh.local",
    [string]$Password    = "qwerty123",
    [string]$Warehouse   = "Основной склад",
    [switch]$WipeExisting,
    [switch]$CreateMissingProducts
)

$ErrorActionPreference = "Stop"

Write-Host "-> Логин..." -ForegroundColor Cyan
$auth = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post `
    -ContentType "application/json" -Body (@{ email=$Email; password=$Password } | ConvertTo-Json)
$H = @{ Authorization = "Bearer $($auth.access_token)" }

Write-Host "-> Получение склада..." -ForegroundColor Cyan
$whs = (Invoke-RestMethod -Uri "$ApiBase/warehouses" -Headers $H).items
$wh  = $whs | Where-Object { $_.name -eq $Warehouse } | Select-Object -First 1
if (-not $wh) { $wh = $whs | Select-Object -First 1 }
if (-not $wh) { throw "Нет ни одного склада" }
Write-Host "  Склад: $($wh.name) ($($wh.id))"

Write-Host "-> Загрузка каталога товаров..." -ForegroundColor Cyan
$products = (Invoke-RestMethod -Uri "$ApiBase/products?include_archived=true" -Headers $H).items
$idBySku = @{}
foreach ($p in $products) {
    if ($p.sku) { $idBySku[$p.sku] = $p.id }
}
Write-Host "  Товаров в каталоге: $($products.Count) (с SKU: $($idBySku.Count))"

Write-Host "-> Чтение Excel: $XlsPath" -ForegroundColor Cyan
$excel = New-Object -ComObject Excel.Application
$excel.Visible = $false
$excel.DisplayAlerts = $false

$rows = New-Object System.Collections.Generic.List[object]
try {
    $wb = $excel.Workbooks.Open($XlsPath, 0, $true)
    $ws = $wb.Sheets.Item(1)
    $maxRow = $ws.UsedRange.Rows.Count
    $maxCol = $ws.UsedRange.Columns.Count
    Write-Host "  Строк: $maxRow, столбцов: $maxCol"

    $headerRow = 0
    for ($r = 1; $r -le [Math]::Min(30, $maxRow); $r++) {
        for ($c = 1; $c -le $maxCol; $c++) {
            $v = "$($ws.Cells.Item($r,$c).Value2)".Trim()
            if ($v -eq "Наименование" -or $v -eq "Код") { $headerRow = $r; break }
        }
        if ($headerRow -gt 0) { break }
    }
    if ($headerRow -eq 0) { throw "Не найден заголовок" }
    Write-Host "  Заголовок: строка $headerRow"

    $col = @{}
    for ($c = 1; $c -le $maxCol; $c++) {
        $hv = "$($ws.Cells.Item($headerRow,$c).Value2)".Trim()
        if ($hv) { $col[$hv] = $c }
    }

    function GetCell {
        param($row, $colName)
        if (-not $col.ContainsKey($colName)) { return $null }
        return $ws.Cells.Item($row, $col[$colName]).Value2
    }

    for ($r = $headerRow + 1; $r -le $maxRow; $r++) {
        $code = GetCell $r "Код"
        $name = GetCell $r "Наименование"

        if ($null -eq $code -and $null -eq $name) { continue }
        $name = if ($name) { "$name".Trim() } else { "" }
        if ($name -eq "Итого:" -or $name -like "Итого:*") { continue }

        $sku = if ($code) { "$code".Trim() } else { "" }
        if ([string]::IsNullOrWhiteSpace($sku) -and [string]::IsNullOrWhiteSpace($name)) { continue }

        $qtyV   = GetCell $r "Доступно"
        $costV  = GetCell $r "Себестоимость"
        $priceV = GetCell $r "Цена продажи"

        $qty = 0.0
        if ($qtyV -is [double] -or $qtyV -is [int]) { $qty = [double]$qtyV }
        elseif ($qtyV -is [string] -and $qtyV -match '^-?\d+([.,]\d+)?$') {
            $qty = [double]($qtyV -replace ',', '.')
        }

        $cost = 0.0
        if ($costV -is [double] -or $costV -is [int]) { $cost = [double]$costV }
        elseif ($costV -is [string] -and $costV -match '^-?\d+([.,]\d+)?$') {
            $cost = [double]($costV -replace ',', '.')
        }

        $price = 0.0
        if ($priceV -is [double] -or $priceV -is [int]) { $price = [double]$priceV }
        elseif ($priceV -is [string] -and $priceV -match '^-?\d+([.,]\d+)?$') {
            $price = [double]($priceV -replace ',', '.')
        }

        if ($qty -le 0) { continue }

        $rows.Add([PSCustomObject]@{
            Sku   = $sku
            Name  = $name
            Qty   = $qty
            Cost  = $cost
            Price = $price
        })
    }

    $wb.Close($false)
} finally {
    $excel.Quit() | Out-Null
    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($excel) | Out-Null
    [GC]::Collect()
}

Write-Host "  Строк с положительным остатком: $($rows.Count)" -ForegroundColor Green

$missing = $rows | Where-Object { -not $idBySku.ContainsKey($_.Sku) }
if ($missing.Count -gt 0) {
    Write-Host "-> Отсутствующих в каталоге: $($missing.Count)" -ForegroundColor Yellow
    if ($CreateMissingProducts) {
        $created = 0
        foreach ($m in $missing) {
            try {
                $body = @{
                    name       = [string]$m.Name
                    sku        = [string]$m.Sku
                    price      = [double]$m.Price
                    cost_price = [double]$m.Cost
                    min_stock  = 0
                    currency   = "RUB"
                } | ConvertTo-Json
                $p = Invoke-RestMethod -Uri "$ApiBase/products" -Method Post -Headers $H `
                    -ContentType "application/json; charset=utf-8" -Body $body
                $idBySku[$m.Sku] = $p.id
                $created++
            } catch {
                try {
                    $found = (Invoke-RestMethod -Uri "$ApiBase/products?include_archived=true" -Headers $H).items |
                        Where-Object { $_.sku -eq $m.Sku } | Select-Object -First 1
                    if ($found) { $idBySku[$m.Sku] = $found.id; $created++ }
                } catch {}
            }
        }
        Write-Host "  Создано товаров: $created" -ForegroundColor Green
    }
}

if ($WipeExisting) {
    Write-Host "-> Очищаем остатки на складе $($wh.name)..." -ForegroundColor Cyan
    $sql = "DELETE FROM stock_movements WHERE warehouse_id = '$($wh.id)'; DELETE FROM stock_balances WHERE warehouse_id = '$($wh.id)';"
    Invoke-SqlQuery -Database radonezh_warehouse -Query $sql
    Write-Host "  Очищено" -ForegroundColor Green
}

Write-Host "-> Формирование документа(ов) приёмки..." -ForegroundColor Cyan

$items = New-Object System.Collections.Generic.List[object]
$skipped = 0
foreach ($row in $rows) {
    $productId = $idBySku[$row.Sku]
    if (-not $productId) { $skipped++; continue }
    $items.Add(@{
        product_id = $productId
        quantity   = [double]$row.Qty
        price      = [double]$row.Cost
    })
}
Write-Host "  Позиций: $($items.Count)  (пропущено: $skipped)"
if ($items.Count -eq 0) { Write-Host "Нечего заливать."; return }

$chunkSize = 300
$chunks    = [Math]::Ceiling($items.Count / $chunkSize)
$stamp     = Get-Date -Format "yyyyMMdd-HHmmss"

Write-Host "  Разбивка на $chunks документ(ов)"
for ($c = 0; $c -lt $chunks; $c++) {
    $from  = $c * $chunkSize
    $to    = [Math]::Min($from + $chunkSize, $items.Count) - 1
    $slice = $items[$from..$to]

    $number = if ($chunks -gt 1) { "REST-$stamp-$($c+1)" } else { "REST-$stamp" }
    $body = @{
        type         = "receipt"
        number       = $number
        warehouse_id = $wh.id
        comment      = "Восстановление остатков (часть $($c+1)/$chunks)"
        items        = $slice
    } | ConvertTo-Json -Depth 6 -Compress

    Write-Host "  [$($c+1)/$chunks] POST /documents ($($slice.Count) поз.)..." -ForegroundColor Cyan
    $doc = Invoke-RestMethod -Uri "$ApiBase/documents" -Method Post -Headers $H `
        -ContentType "application/json; charset=utf-8" -Body $body

    Write-Host "    Проводка $($doc.number)..." -ForegroundColor Cyan
    $posted = Invoke-RestMethod -Uri "$ApiBase/documents/$($doc.id)/post" -Method Post -Headers $H
    Write-Host "    OK: $($posted.status)" -ForegroundColor Green
}

Write-Host ""
Write-Host "ГОТОВО! Документов: $chunks, позиций: $($items.Count)" -ForegroundColor Green