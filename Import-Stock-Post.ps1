# Import-Stock-Post.ps1 — только формирование и проводка документа «Ввод остатков»
param(
    [Parameter(Mandatory=$true)] [string]$XlsPath,
    [string]$ApiBase = "http://localhost:8080/api/v1",
    [string]$Email   = "admin@radonezh.local",
    [string]$Password = "qwerty123",
    [string]$WarehouseName = "Основной склад"
)

$ErrorActionPreference = "Stop"

Write-Host "→ Логин..." -ForegroundColor Cyan
$auth = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post `
    -ContentType "application/json" -Body (@{ email=$Email; password=$Password } | ConvertTo-Json)
$H = @{ Authorization = "Bearer $($auth.access_token)" }

Write-Host "→ Получение склада..." -ForegroundColor Cyan
$whs = (Invoke-RestMethod -Uri "$ApiBase/warehouses" -Headers $H).items
$wh = $whs | Where-Object { $_.name -eq $WarehouseName } | Select-Object -First 1
if (-not $wh) { $wh = $whs | Select-Object -First 1 }
if (-not $wh) { throw "Нет ни одного склада" }
$warehouseId = $wh.id
Write-Host "  Склад: $($wh.name)"

Write-Host "→ Загрузка товаров из API..." -ForegroundColor Cyan
$allProducts = (Invoke-RestMethod -Uri "$ApiBase/products?include_archived=true" -Headers $H).items
$idBySku = @{}
foreach ($p in $allProducts) {
    if ($p.sku) { $idBySku[$p.sku] = $p.id }
}
Write-Host "  Загружено товаров: $($idBySku.Count)"

Write-Host "→ Чтение XLS: $XlsPath" -ForegroundColor Cyan
$excel = New-Object -ComObject Excel.Application
$excel.Visible = $false
$excel.DisplayAlerts = $false
try {
    $wb = $excel.Workbooks.Open($XlsPath, 0, $true)
    $ws = $wb.Sheets.Item(1)
    $maxRow = $ws.UsedRange.Rows.Count

    $headerRow = 0
    for ($r = 1; $r -le [Math]::Min(30, $maxRow); $r++) {
        for ($c = 1; $c -le 15; $c++) {
            if ("$($ws.Cells.Item($r,$c).Value2)".Trim() -eq "Наименование") {
                $headerRow = $r; break
            }
        }
        if ($headerRow -gt 0) { break }
    }

    $items = New-Object System.Collections.Generic.List[object]
    $missing = 0
    $zeroQty = 0
    $seenSku = @{}

    for ($r = $headerRow + 1; $r -le $maxRow; $r++) {
        $sku   = "$($ws.Cells.Item($r,2).Value2)".Trim()
        $name  = "$($ws.Cells.Item($r,4).Value2)".Trim()
        $qtyV  = $ws.Cells.Item($r,6).Value2
        $costV = $ws.Cells.Item($r,10).Value2

        if ([string]::IsNullOrWhiteSpace($name)) { continue }
        if ($name -like "Итого*") { continue }

        $qty  = if ($qtyV  -is [double] -or $qtyV  -is [int]) { [double]$qtyV  } else { 0 }
        $cost = if ($costV -is [double] -or $costV -is [int]) { [double]$costV } else { 0 }

        if ([string]::IsNullOrWhiteSpace($sku)) { $sku = "IMP-" + ($r.ToString("00000")) }
        if ($seenSku.ContainsKey($sku)) { continue }
        $seenSku[$sku] = $true

        if ($qty -le 0) { $zeroQty++; continue }

        $productId = $idBySku[$sku]
        if (-not $productId) { $missing++; continue }

        $items.Add(@{
            product_id = $productId
            quantity   = $qty
            price      = $cost
        })
    }

    $wb.Close($false)
} finally {
    $excel.Quit() | Out-Null
    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($excel) | Out-Null
    [GC]::Collect()
}

Write-Host "  Позиций для документа: $($items.Count)" -ForegroundColor Green
Write-Host "  Пропущено (нулевой остаток): $zeroQty" -ForegroundColor Gray
Write-Host "  Не найдено в БД: $missing" -ForegroundColor Yellow

if ($items.Count -eq 0) { Write-Host "Нечего проводить."; return }

$importNumber = "IMP-" + (Get-Date -Format "yyyyMMdd-HHmmss")
Write-Host "→ Создание документа $importNumber..." -ForegroundColor Cyan

# Разбиваем на порции по 300 позиций (сервер может ругаться на слишком большие тела)
$chunkSize = 300
$total = $items.Count
$chunks = [Math]::Ceiling($total / $chunkSize)
Write-Host "  Разбивка на $chunks документ(ов)"

for ($c = 0; $c -lt $chunks; $c++) {
    $from = $c * $chunkSize
    $to = [Math]::Min($from + $chunkSize, $total) - 1
    $slice = $items[$from..$to]

    $num = if ($chunks -gt 1) { "$importNumber-$($c+1)" } else { $importNumber }
    $docBody = @{
        type         = "receipt"
        number       = $num
        warehouse_id = $warehouseId
        comment      = "Начальный ввод остатков из МойСклад (часть $($c+1)/$chunks)"
        items        = $slice
    } | ConvertTo-Json -Depth 6 -Compress

    Write-Host "  [$($c+1)/$chunks] POST /documents ($($slice.Count) поз., ~$([Math]::Round($docBody.Length/1024,1)) KB)..." -ForegroundColor Cyan
    $doc = Invoke-RestMethod -Uri "$ApiBase/documents" -Method Post -Headers $H `
        -ContentType "application/json" -Body $docBody

    Write-Host "    Проводка $($doc.number)..." -ForegroundColor Cyan
    $posted = Invoke-RestMethod -Uri "$ApiBase/documents/$($doc.id)/post" -Method Post -Headers $H
    Write-Host "    OK: статус $($posted.status)" -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ Готово! Документов: $chunks, позиций: $total" -ForegroundColor Green
Write-Host "Проверьте: Склад → Остатки" -ForegroundColor Green