# devtools.ps1 — хелперы RadonezhSklad
# ВАЖНО: обход бага Docker Desktop на Windows — все SQL идут через docker cp,
# НЕ через pipe `docker exec -i`. Pipe теряет не-ASCII байты.

function Get-Utf8NoBom {
    return New-Object System.Text.UTF8Encoding $false
}

function Invoke-SqlQuery {
    param(
        [Parameter(Mandatory)][string]$Database,
        [Parameter(Mandatory)][string]$Query,
        [string]$User = 'radonezh',
        [string]$Container = 'rs_postgres'
    )
    $stamp = [guid]::NewGuid().ToString('N')
    $localSql   = Join-Path $env:TEMP "rs_${stamp}.sql"
    $localOut   = Join-Path $env:TEMP "rs_${stamp}.out"
    $containerIn  = "/tmp/rs_${stamp}.sql"
    $containerOut = "/tmp/rs_${stamp}.out"

    try {
        Write-Utf8File -Path $localSql -Content $Query
        docker cp $localSql "${Container}:${containerIn}" | Out-Null
        docker exec -i $Container bash -c "psql -U $User -d $Database -f $containerIn > $containerOut"
        docker cp "${Container}:${containerOut}" $localOut | Out-Null
        return Read-Utf8File -Path $localOut
    } finally {
        docker exec -i $Container rm -f $containerIn $containerOut 2>$null
        Remove-Item $localSql, $localOut -ErrorAction SilentlyContinue
    }
}

function Invoke-SqlFile {
    param(
        [Parameter(Mandatory)][string]$Path,
        [Parameter(Mandatory)][string]$Database,
        [string]$User = 'radonezh',
        [string]$Container = 'rs_postgres'
    )
    $sql = Read-Utf8File -Path $Path
    Invoke-SqlQuery -Database $Database -Query $sql -User $User -Container $Container
}

function Set-ConsoleUtf8 {
    if ($PSVersionTable.PSVersion.Major -ge 7) {
        Write-Host "OK: PowerShell $($PSVersionTable.PSVersion) — UTF-8 нативно"
        return
    }
    chcp 65001 | Out-Null
    [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
    [Console]::InputEncoding  = [System.Text.UTF8Encoding]::new($false)
    $global:OutputEncoding    = [System.Text.UTF8Encoding]::new($false)
    Write-Host "OK: PS 5.1 — UTF-8 обход применён"
}
# === Moysklad API (api.moysklad.ru/remap/1.2) ===
# Token file: D:\Radonezhsklad\.secrets\moysklad.token (outside repo)
# Headers: Authorization + Accept: application/json;charset=utf-8 + Accept-Encoding: gzip
# Without Accept-Encoding: gzip the server returns 415 Unsupported Media Type.

function Get-MsToken {
    $path = 'D:\Radonezhsklad\.secrets\moysklad.token'
    if (-not (Test-Path $path)) { throw "No MS token file: $path" }
    $utf8 = New-Object System.Text.UTF8Encoding $false
    $tok = ([IO.File]::ReadAllText($path, $utf8)).Trim()
    if ([string]::IsNullOrWhiteSpace($tok)) { throw "Empty token in $path" }
    return $tok
}


function MsApi-Get {
    param(
        [Parameter(Mandatory)][string]$Path,
        [hashtable]$Query = @{}
    )
    $token = Get-MsToken
    $qs = [string]::Empty
    if ($Query.Count -gt 0) {
        $pairs = foreach ($k in $Query.Keys) { "$k=$($Query[$k])" }
        $qs = '?' + ($pairs -join '&')
    }
    $url = "https://api.moysklad.ru/api/remap/1.2$Path$qs"
    $tmp = Join-Path $env:TEMP ("ms_" + [guid]::NewGuid().ToString('N') + ".json")
    try {
        & curl.exe -s --compressed -o $tmp `
            -H "Authorization: Bearer $token" `
            -H "Accept: application/json;charset=utf-8" `
            -H "Accept-Encoding: gzip" `
            $url
        if ($LASTEXITCODE -ne 0) { throw "curl exit $LASTEXITCODE for $url" }
        $raw = [IO.File]::ReadAllText($tmp, (New-Object System.Text.UTF8Encoding $false))
        return ($raw | ConvertFrom-Json)
    } finally {
        Remove-Item $tmp -ErrorAction SilentlyContinue
    }
}

function MsApi-GetAll {
    param(
        [Parameter(Mandatory)][string]$Path,
        [hashtable]$Query = @{},
        [int]$PageSize = 100,
        [int]$MaxItems = 0
    )
    $all = New-Object System.Collections.Generic.List[object]
    $offset = 0
    while ($true) {
        $q = @{}
        foreach ($k in $Query.Keys) { $q[$k] = $Query[$k] }
        $q['limit']  = $PageSize
        $q['offset'] = $offset
        $page = MsApi-Get -Path $Path -Query $q
        if (-not $page.rows) { break }
        foreach ($r in $page.rows) { $all.Add($r) }
        $offset += $page.rows.Count
        $total = [int]$page.meta.size
        Write-Host "  received $($all.Count) / $total"
        if ($MaxItems -gt 0 -and $all.Count -ge $MaxItems) { break }
        if ($offset -ge $total) { break }
        Start-Sleep -Milliseconds 100
    }
    return $all
}


# ─── Fix CWD-резолюции относительных путей ─────────────────────────
# PS 7 хранит [Environment]::CurrentDirectory отдельно от $PWD,
# а [IO.File]::* резолвит относительный путь против Environment.CurrentDirectory.
# Используем PS-провайдер, чтобы путь резолвился от $PWD вызывающего.
function Write-Utf8File {
    param([Parameter(Mandatory)][string]$Path, [Parameter(Mandatory)][string]$Content)
    $abs = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Path)
    $utf8 = New-Object System.Text.UTF8Encoding $false
    [IO.File]::WriteAllText($abs, $Content, $utf8)
}

function Read-Utf8File {
    param([Parameter(Mandatory)][string]$Path)
    $abs = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Path)
    $utf8 = New-Object System.Text.UTF8Encoding $false
    return [IO.File]::ReadAllText($abs, $utf8)
}
