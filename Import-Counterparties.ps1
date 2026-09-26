# Import-Counterparties.ps1 — импорт контрагентов из Excel МойСклад
param(
    [Parameter(Mandatory=$true)] [string]$XlsPath,
    [string]$ApiBase = "http://localhost:8080/api/v1",
    [string]$Email   = "admin@radonezh.local",
    [string]$Password = "qwerty123"
)

$ErrorActionPreference = "Stop"

# ---- Логин ----
Write-Host "→ Логин..." -ForegroundColor Cyan
$auth = Invoke-RestMethod -Uri "$ApiBase/auth/login" -Method Post `
    -ContentType "application/json" -Body (@{ email=$Email; password=$Password } | ConvertTo-Json)
$H = @{ Authorization = "Bearer $($auth.access_token)" }

# ---- Чтение Excel ----
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

    # Ищем заголовок
    $headerRow = 0
    for ($r = 1; $r -le [Math]::Min(30, $maxRow); $r++) {
        for ($c = 1; $c -le $maxCol; $c++) {
            $v = "$($ws.Cells.Item($r,$c).Value2)".Trim()
            if ($v -eq "Наименование") { $headerRow = $r; break }
        }
        if ($headerRow -gt 0) { break }
    }
    if ($headerRow -eq 0) { throw "Не найден заголовок «Наименование»" }
    Write-Host "  Заголовок: строка $headerRow"

    # Карта колонок
    $col = @{}
    for ($c = 1; $c -le $maxCol; $c++) {
        $headerVal = "$($ws.Cells.Item($headerRow,$c).Value2)".Trim()
        if ($headerVal) { $col[$headerVal] = $c }
    }

    # Собираем данные построчно в объекты (сырые строки)
    for ($r = $headerRow + 1; $r -le $maxRow; $r++) {
        $name = ""
        if ($col.ContainsKey("Наименование")) {
            $name = "$($ws.Cells.Item($r, $col["Наименование"]).Value2)".Trim()
        }
        if ([string]::IsNullOrWhiteSpace($name)) { continue }

        $get = { param($n)
            if (-not $col.ContainsKey($n)) { return "" }
            return "$($ws.Cells.Item($r, $col[$n]).Value2)".Trim()
        }

        $rowsData += [PSCustomObject]@{
            name              = $name
            full_name         = & $get "Полное наименование"
            last_name         = & $get "Фамилия (для ИП и физ. лиц)"
            first_name        = & $get "Имя (для ИП и физ. лиц)"
            middle_name       = & $get "Отчество (для ИП и физ. лиц)"
            phone             = & $get "Телефон"
            fax               = & $get "Факс"
            email             = & $get "E-mail"
            legal_address     = & $get "Юридический адрес"
            actual_address    = & $get "Фактический адрес"
            inn               = & $get "ИНН"
            kpp               = & $get "КПП"
            ogrn              = & $get "ОГРН"
            okpo              = & $get "ОКПО"
            external_code     = & $get "Внешний код"
            counterparty_type = & $get "Тип контрагента"
            status            = & $get "Статус"
            comment           = & $get "Комментарий"
            archived          = ((& $get "Архивный") -eq "да")
        }
    }

    $wb.Close($false)
} finally {
    $excel.Quit() | Out-Null
    [System.Runtime.InteropServices.Marshal]::ReleaseComObject($excel) | Out-Null
    [GC]::Collect()
}

Write-Host "  Готово к отправке: $($rowsData.Count)" -ForegroundColor Green

# ---- Отправка ----
Write-Host "→ Отправка контрагентов..." -ForegroundColor Cyan
$created = 0; $failed = 0
$i = 0
$firstErrors = @()

foreach ($row in $rowsData) {
    $i++
    if ($i % 50 -eq 0) { Write-Host "  ... $i / $($rowsData.Count)" }

    $statusVal = $row.status
    if ([string]::IsNullOrWhiteSpace($statusVal)) { $statusVal = "Новый" }

    # Хеш-таблица строится ЯВНО, без inline if
    $body = @{
        name              = [string]$row.name
        full_name         = [string]$row.full_name
        last_name         = [string]$row.last_name
        first_name        = [string]$row.first_name
        middle_name       = [string]$row.middle_name
        phone             = [string]$row.phone
        fax               = [string]$row.fax
        email             = [string]$row.email
        legal_address     = [string]$row.legal_address
        actual_address    = [string]$row.actual_address
        inn               = [string]$row.inn
        kpp               = [string]$row.kpp
        ogrn              = [string]$row.ogrn
        okpo              = [string]$row.okpo
        external_code     = [string]$row.external_code
        counterparty_type = [string]$row.counterparty_type
        status            = [string]$statusVal
        comment           = [string]$row.comment
        archived          = [bool]$row.archived
    }

    $json = $body | ConvertTo-Json -Depth 3

    try {
        $null = Invoke-RestMethod -Uri "$ApiBase/customers" -Method Post -Headers $H `
            -ContentType "application/json; charset=utf-8" -Body $json
        $created++
    } catch {
        $failed++
        if ($firstErrors.Count -lt 3) {
            $errMsg = $_.ErrorDetails.Message
            if (-not $errMsg) { $errMsg = $_.Exception.Message }
            $firstErrors += "  [$($row.name)] → $errMsg"
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