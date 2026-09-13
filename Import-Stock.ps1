# Import-Stock.ps1 — импорт товаров и остатков из XLS МойСклад в RadonezhSklad
param(
    [Parameter(Mandatory=$true)] [string]$XlsPath,
    [string]$ApiBase = "http://localhost:8080/api/v1",
    [string]$Email   = "admin@radonezh.local",
    [string]$Password = "qwerty123",
    [string]$WarehouseName = "Основной склад"
)

$ErrorActionPreference = "Stop"

# ---------- 1. Логин ----------
Write-Host "→ Логин..." -ForegroundColor Cyan
$auth = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post `
    -ContentType "application/json" -Body (@{ email=$Email; password=$Password } | ConvertTo-Json)
$H = @{ Authorization = "Bearer $($auth.access_token)" }

# ---------- 2. Склад ----------
Write-Host "→ Получение склада..." -ForegroundColor Cyan
$whs = (Invoke-RestMethod -Uri "$ApiBase/warehouses" -Headers $H).items
$wh = $whs | Where-Object { $_.name -eq $WarehouseName } | Select-Object -First 1
if (-not $wh) { $wh = $whs | Select-Object -First 1 }
if (-not $wh) { throw "Нет ни одного склада. Создайте склад сначала." }
$warehouseId = $wh.id
Write-Host "  Склад: $($wh.name) ($warehouseId)"

# ---------- 3. Единицы ----------
Write-Host "→ Получение справочника единиц..." -ForegroundColor Cyan
$units = (Invoke-RestMethod -Uri "$ApiBase/units" -Headers $H).items
$unitByShort = @{}
foreach ($u in $units) { $unitByShort[$u.short_name] = $u.id }
Write-Host "  Единиц: $($units.Count)"

# ---------- 4. Чтение XLS ----------
Write-Host "→ Чтение Excel: $XlsPath" -ForegroundColor Cyan
$excel = New-Object -ComObject Excel.Application
$excel.Visible = $false
$excel.DisplayAlerts = $false
try {
    $wb = $excel.Workbooks.Open($XlsPath, 0, $true)
    $ws = $wb.Sheets.Item(1)
    $maxRow = $ws.UsedRange.Rows.Count
    Write-Host "  Всего строк: $maxRow"

    # Ищем заголовок (по слову «Наименование»)
    $headerRow = 0
    for ($r = 1; $r -le [Math]::Min(30, $maxRow); $r++) {
        for ($c = 1; $c -le 15; $c++) {
            if ("$($ws.Cells.Item($r,$c).Value2)".Trim() -eq "Наименование") {
                $headerRow = $r; break
            }
        }
        if ($headerRow -gt 0) { break }
    }
    if ($headerRow -eq 0) { throw "Не найден заголовок «Наименование»" }
    Write-Host "  Заголовок в строке: $headerRow"

    $rows = New-Object System.Collections.Generic.List[object]
    $seen = @{}   # dedupe SKU

    for ($r = $headerRow + 1; $r -le $maxRow; $r++) {
        $sku   = "$($ws.Cells.Item($r,2).Value2)".Trim()   # B — Код
        $name  = "$($ws.Cells.Item($r,4).Value2)".Trim()   # D — Наименование
        $unit  = "$($ws.Cells.Item($r,5).Value2)".Trim()   # E — Ед.изм.
        $qtyV  = $ws.Cells.Item($r,6).Value2                # F — Доступно
        $costV = $ws.Cells.Item($r,10).Value2               # J — Себестоимость
        $priceV= $ws.Cells.Item($r,12).Value2               # L — Цена продажи

        if ([string]::IsNullOrWhiteSpace($name)) { continue }
        if ($name -eq "Итого:" -or $name -like "Итого*") { continue }

        $qty   = if ($qtyV -is [double] -or $qtyV -is [int]) { [double]$qtyV } else { 0 }
        $cost  = if ($costV -is [double] -or $costV -is [int]) { [double]$costV } else { 0 }
        $price = if ($priceV -is [double] -or $priceV -is [int]) { [double]$priceV } else { 0 }

        if ([string]::IsNullOrWhiteSpace($sku)) {
            $sku = "IMP-" + ($r.ToString("00000"))
        }
        # Дубликаты SKU — первый выигрывает
        if ($seen.ContainsKey($sku)) { continue }
        $seen[$sku] = $true

        $rows.Add([PSCustomObject]@{
            Sku = $sku; Name = $name; Unit = $unit
            Qty = $qty; Cost = $cost; Price = $price
        })
    }

    $wb.Close($false)
} finally {
    $excel.Quit() | Out-Null
    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($excel) | Out-Null
    [GC]::Collect()
}

Write-Host "  Уникальных товаров: $($rows.Count)" -ForegroundColor Green

# ---------- 5. Создание товаров ----------
Write-Host "→ Создание товаров (может занять 1-2 минуты)..." -ForegroundColor Cyan
$idBySku = @{}
$created = 0; $updated = 0; $failed = 0
$i = 0

foreach ($row in $rows) {
    $i++
    if ($i % 50 -eq 0) { Write-Host "  ... $i / $($rows.Count)" }

    $unitId = if ($row.Unit -and $unitByShort.ContainsKey($row.Unit)) { $unitByShort[$row.Unit] } else { $null }

    $body = @{
        name       = $row.Name
        sku        = $row.Sku
        price      = $row.Price
        cost_price = $row.Cost
        min_stock  = 0
        currency   = "RUB"
    }
    if ($unitId) { $body.unit_id = $unitId }
    $json = $body | ConvertTo-Json

    try {
        $p = Invoke-RestMethod -Uri "$ApiBase/products" -Method Post -Headers $H `
            -ContentType "application/json" -Body $json
        $idBySku[$row.Sku] = $p.id
        $created++
    } catch {
        # возможно, SKU уже существует — ищем ID
        try {
            $found = (Invoke-RestMethod -Uri "$ApiBase/products?include_archived=true" -Headers $H).items |
                Where-Object { $_.sku -eq $row.Sku } | Select-Object -First 1
            if ($found) {
                $idBySku[$row.Sku] = $found.id
                $updated++
            } else { $failed++ }
        } catch { $failed++ }
    }
}

Write-Host "  Создано: $created, найдено: $updated, ошибок: $failed" -ForegroundColor Green

# ---------- 6. Документ «Ввод остатков» ----------
Write-Host "→ Формирование документа приёмки..." -ForegroundColor Cyan

$items = New-Object System.Collections.Generic.List[object]
foreach ($row in $rows) {
    $productId = $idBySku[$row.Sku]
    if (-not $productId) { continue }
    if ($row.Qty -eq 0) { continue }
    $items.Add(@{
        product_id = $productId
        quantity   = [double]$row.Qty
        price      = [double]$row.Cost
    })
}

Write-Host "  Позиций в документе: $($items.Count)"

if ($items.Count -eq 0) {
    Write-Host "Нет ненулевых остатков — документ не создаём." -ForegroundColor Yellow
    return
}

$importNumber = "IMP-" + (Get-Date -Format "yyyyMMdd-HHmmss")
$docBody = @{
    type         = "receipt"
    number       = $importNumber
    warehouse_id = $warehouseId
    comment      = "Начальный ввод остатков из МойСклад"
    items        = $items
} | ConvertTo-Json -Depth 6 -Compress

Write-Host "  POST /documents (body ~$([Math]::Round($docBody.Length/1024,1)) KB)..." -ForegroundColor Cyan
$doc = Invoke-RestMethod -Uri "$ApiBase/documents" -Method Post -Headers $H `
    -ContentType "application/json" -Body $docBody
Write-Host "  Создан документ: $($doc.number) ($($doc.id))" -ForegroundColor Green

# ---------- 7. Проводка ----------
Write-Host "→ Проводка документа (stock_balances + stock_movements)..." -ForegroundColor Cyan
$posted = Invoke-RestMethod -Uri "$ApiBase/documents/$($doc.id)/post" -Method Post -Headers $H
Write-Host "  Проведён: $($posted.number), статус: $($posted.status)" -ForegroundColor Green

Write-Host ""
Write-Host "✅ Готово! Загружено товаров: $($rows.Count), движений: $($items.Count)" -ForegroundColor Green
Write-Host "Открывайте: Склад → Остатки, Склад → Документы" -ForegroundColor Green