# Import-Contracts.ps1 — импорт договоров из Excel
param(
    [Parameter(Mandatory=$true)] [string]$XlsPath,
    [string]$ApiBase = "http://localhost:8080/api/v1",
    [string]$Email   = "admin@radonezh.local",
    [string]$Password = "qwerty123"
)

$ErrorActionPreference = "Stop"

Write-Host "→ Логин..." -ForegroundColor Cyan
$auth = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post `
    -ContentType "application/json" -Body (@{ email=$Email; password=$Password } | ConvertTo-Json)
$H = @{ Authorization = "Bearer $($auth.access_token)" }

Write-Host "→ Загрузка контрагентов..." -ForegroundColor Cyan
$custs = (Invoke-RestMethod -Uri "$ApiBase/customers" -Headers $H).items
$custByName = @{}
foreach ($c in $custs) { $custByName[$c.name.ToLower()] = $c.id }
Write-Host "  Контрагентов: $($custs.Count)"

Write-Host "→ Чтение Excel: $XlsPath" -ForegroundColor Cyan
$excel = New-Object -ComObject Excel.Application
$excel.Visible = $false
$excel.DisplayAlerts = $false

$rowsData = @()
try {
    $wb = $excel.Workbooks.Open($XlsPath, 0, $true)
    $ws = $wb.Sheets.Item(1)
    $maxRow = $ws.UsedRange.Rows.Count
    $maxCol = $ws.UsedRange.Columns.Count
    Write-Host "  Строк: $maxRow, столбцов: $maxCol"

    # Ищем заголовок по «Номер»
    $headerRow = 0
    for ($r = 1; $r -le [Math]::Min(30, $maxRow); $r++) {
        for ($c = 1; $c -le $maxCol; $c++) {
            $v = "$($ws.Cells.Item($r,$c).Value2)".Trim()
            if ($v -eq "Номер") { $headerRow = $r; break }
        }
        if ($headerRow -gt 0) { break }
    }
    if ($headerRow -eq 0) { throw "Не найден заголовок «Номер»" }
    Write-Host "  Заголовок: строка $headerRow"

    $col = @{}
    for ($c = 1; $c -le $maxCol; $c++) {
        $headerVal = "$($ws.Cells.Item($headerRow,$c).Value2)".Trim()
        if ($headerVal) { $col[$headerVal] = $c }
    }

    for ($r = $headerRow + 1; $r -le $maxRow; $r++) {
        $num = ""
        if ($col.ContainsKey("Номер")) {
            $num = "$($ws.Cells.Item($r, $col["Номер"]).Value2)".Trim()
        }
        if ([string]::IsNullOrWhiteSpace($num)) { continue }
        if ($num -eq "Итого:") { continue }

        $get = { param($n)
            if (-not $col.ContainsKey($n)) { return "" }
            return "$($ws.Cells.Item($r, $col[$n]).Value2)".Trim()
        }
        $getNum = { param($n)
            if (-not $col.ContainsKey($n)) { return 0 }
            $v = $ws.Cells.Item($r, $col[$n]).Value2
            if ($v -is [double] -or $v -is [int]) { return [double]$v }
            return 0
        }

        $rowsData += [PSCustomObject]@{
            number    = $num
            doc_date  = & $get "Время"
            customer  = & $get "Контрагент"
            amount    = & $getNum "Сумма"
            paid      = & $getNum "Оплачено"
            fulfilled = & $getNum "Выполнено"
            comment   = & $get "Комментарий"
        }
    }
    $wb.Close($false)
} finally {
    $excel.Quit() | Out-Null
    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($excel) | Out-Null
    [GC]::Collect()
}

Write-Host "  Готово к отправке: $($rowsData.Count)" -ForegroundColor Green

Write-Host "→ Отправка..." -ForegroundColor Cyan
$created = 0; $failed = 0
$i = 0
$firstErrors = @()
$codeCounter = 1

foreach ($row in $rowsData) {
    $i++
    if ($i % 200 -eq 0) { Write-Host "  ... $i / $($rowsData.Count)" }

    $customerId = ""
    if ($row.customer) {
        $key = $row.customer.ToLower()
        if ($custByName.ContainsKey($key)) { $customerId = $custByName[$key] }
    }

    $docDate = ""
    if ($row.doc_date) {
        try {
            $dt = [DateTime]::Parse($row.doc_date)
            $docDate = $dt.ToString("yyyy-MM-ddTHH:mm:sszzz")
        } catch {}
    }

    $body = @{
        number       = [string]$row.number
        code         = ("{0:D6}" -f $codeCounter)
        doc_date     = $docDate
        customer_id  = $customerId
        amount       = [double]$row.amount
        currency     = "RUB"
        paid         = [double]$row.paid
        fulfilled    = [double]$row.fulfilled
        comment      = [string]$row.comment
        archived     = $false
    }
    if (-not $customerId) { $body.Remove("customer_id") }
    if (-not $docDate)    { $body.Remove("doc_date") }

    $json = $body | ConvertTo-Json -Depth 3

    try {
        $null = Invoke-RestMethod -Uri "$ApiBase/contracts" -Method Post -Headers $H `
            -ContentType "application/json; charset=utf-8" -Body $json
        $created++
        $codeCounter++
    } catch {
        $failed++
        if ($firstErrors.Count -lt 5) {
            $errMsg = $_.ErrorDetails.Message
            if (-not $errMsg) { $errMsg = $_.Exception.Message }
            $firstErrors += "  [$($row.number)] → $errMsg"
        }
    }
}

Write-Host ""
if ($firstErrors.Count -gt 0) {
    Write-Host "Первые ошибки:" -ForegroundColor Yellow
    $firstErrors | ForEach-Object { Write-Host $_ -ForegroundColor Red }
    Write-Host ""
}
Write-Host "✅ Готово! Создано: $created, ошибок: $failed" -ForegroundColor Green