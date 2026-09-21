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
        $pairs = foreach ($k in $Query.Keys) {
            $val = [string]$Query[$k]
            "$([uri]::EscapeDataString($k))=$([uri]::EscapeDataString($val))"
        }
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

# === Extended MsApi-Get with retry (Stage 35) ===
# Redefines MsApi-Get: validates JSON, retries on transient errors.
function MsApi-Get {
    param(
        [Parameter(Mandatory)][string]$Path,
        [hashtable]$Query = @{},
        [int]$MaxRetries = 6
    )
    $token = Get-MsToken
    $qs = [string]::Empty
    if ($Query.Count -gt 0) {
        $pairs = foreach ($k in $Query.Keys) {
            $val = [string]$Query[$k]
            "$([uri]::EscapeDataString($k))=$([uri]::EscapeDataString($val))"
        }
        $qs = '?' + ($pairs -join '&')
    }
    $url = "https://api.moysklad.ru/api/remap/1.2$Path$qs"
    $attempt = 0
    $delay = 1000
    while ($true) {
        $attempt++
        $tmp = Join-Path $env:TEMP ("ms_" + [guid]::NewGuid().ToString('N') + ".json")
        $ok = $false
        $result = $null
        try {
            & curl.exe -s --compressed -o $tmp `
                -H "Authorization: Bearer $token" `
                -H "Accept: application/json;charset=utf-8" `
                -H "Accept-Encoding: gzip" `
                $url
            if ($LASTEXITCODE -ne 0) { throw "curl exit $LASTEXITCODE" }
            $raw = [IO.File]::ReadAllText($tmp, (New-Object System.Text.UTF8Encoding $false))
            $trim = $raw.TrimStart()
            if ($trim.Length -eq 0) { throw "empty response" }
            $c0 = [int][char]$trim[0]
            if ($c0 -ne 123 -and $c0 -ne 91) {
                $preview = $raw.Substring(0, [Math]::Min(200, $raw.Length))
                throw "non-JSON response: $preview"
            }
            $result = $raw | ConvertFrom-Json
            $ok = $true
        } catch {
            if ($attempt -ge $MaxRetries) { throw }
            $msg = $_.Exception.Message
            Write-Warning ("MS GET " + $Path + " failed (attempt " + $attempt + "/" + $MaxRetries + "): " + $msg + ". Retry in " + $delay + "ms")
            Start-Sleep -Milliseconds $delay
            $delay = [Math]::Min($delay * 2, 30000)
        } finally {
            Remove-Item $tmp -ErrorAction SilentlyContinue
        }
        if ($ok) { return $result }
    }
}

# === Универсальный слой SQL: docker (дома) / native psql (офис) ===
# Кэшируем выбор при первом вызове.
$script:RsSqlBackend = $null   # 'docker' | 'psql'
$script:RsPsqlLocal  = $null
$script:RsPgHost     = 'localhost'
$script:RsPgPass     = if ($env:PGPASSWORD) { $env:PGPASSWORD } else { 'radonezh_dev_pass' }

function Get-RsSqlBackend {
	# Явное переопределение через env (например, офис без Docker)
	if ($env:RS_SQL_BACKEND) {
		$script:RsSqlBackend = $env:RS_SQL_BACKEND
		Write-Host "rs-sql backend: $($script:RsSqlBackend) (env override)" -ForegroundColor DarkGray
		return $script:RsSqlBackend
	}

    if ($script:RsSqlBackend) { return $script:RsSqlBackend }

    # 1) ищем локальный psql
    $candidates = @(
        'C:\Program Files\PostgreSQL\16\bin\psql.exe',
        'C:\Program Files\PostgreSQL\15\bin\psql.exe',
        'C:\Program Files\PostgreSQL\14\bin\psql.exe',
        'C:\Program Files\PostgreSQL\17\bin\psql.exe'
    )
    $psql = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
    if (-not $psql) {
        $cmd = Get-Command psql -ErrorAction SilentlyContinue
        if ($cmd) { $psql = $cmd.Source }
    }
    if ($psql) {
        $script:RsPsqlLocal  = $psql
        $script:RsSqlBackend = 'psql'
        Write-Host "rs-sql backend: native psql ($psql)" -ForegroundColor DarkGray
        return 'psql'
    }

    # 2) fallback — docker
    $running = docker ps --filter "name=rs_postgres" --filter "status=running" --format "{{.Names}}" 2>$null
    if ($running -eq 'rs_postgres') {
        $script:RsSqlBackend = 'docker'
        Write-Host "rs-sql backend: docker exec rs_postgres" -ForegroundColor DarkGray
        return 'docker'
    }

    throw "Ни нативный psql, ни запущенный rs_postgres не найдены."
}

# Копирует файл в БД-сторону (docker cp или ничего для native psql)
# Возвращает строку, которую надо подставить в '\copy ... FROM ''<path>'''
function Copy-ToPostgres {
    param(
        [Parameter(Mandatory)][string]$LocalPath,
        [Parameter(Mandatory)][string]$RemoteName
    )
    $abs = (Get-Item -LiteralPath $LocalPath).FullName
    $backend = Get-RsSqlBackend
    if ($backend -eq 'docker') {
        $remote = "/tmp/$RemoteName"
        docker cp $abs "rs_postgres:$remote" | Out-Null
        return $remote
    } else {
        # native psql: \copy читает локальный файл; путь с прямыми слешами
        return ($abs -replace '\\', '/')
    }
}

# Выполнить SQL-файл. Возвращает вывод psql.
function Invoke-PsqlFile {
    param(
        [Parameter(Mandatory)][string]$Database,
        [Parameter(Mandatory)][string]$File
    )
    $abs = (Get-Item -LiteralPath $File).FullName
    $backend = Get-RsSqlBackend
    if ($backend -eq 'docker') {
        $tmp = "/tmp/rs_$(Get-Random).sql"
        docker cp $abs "rs_postgres:$tmp" | Out-Null
        $out = docker exec rs_postgres psql -U radonezh -d $Database -v ON_ERROR_STOP=1 --set=client_min_messages=error -f $tmp 2>&1
        if ($LASTEXITCODE -ne 0) { Write-Host ($out | Out-String) -ForegroundColor Red; throw "psql failed: $File" }
        return $out
    } else {
        $prev = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
        try {
            $env:PGPASSWORD = $script:RsPgPass
            $env:PGCLIENTENCODING = 'UTF8'
            $out = & $script:RsPsqlLocal -U radonezh -h $script:RsPgHost -d $Database -v ON_ERROR_STOP=1 --set=client_min_messages=error -f $abs 2>&1
        } finally { $ErrorActionPreference = $prev }
        if ($LASTEXITCODE -ne 0) { Write-Host ($out | Out-String) -ForegroundColor Red; throw "psql failed: $File" }
        return $out
    }
}

# Удалить временный файл в контейнере (для native psql — no-op)
function Remove-PostgresTemp {
    param([string]$RemotePath)
    $backend = Get-RsSqlBackend
    if ($backend -eq 'docker' -and $RemotePath) {
        docker exec rs_postgres rm -f $RemotePath 2>$null | Out-Null
    }
}

# === Backend-aware Invoke-SqlQuery / Invoke-SqlFile (override старых docker-only версий) ===
function Invoke-SqlQuery {
    param(
        [Parameter(Mandatory)][string]$Db,
        [Parameter(Mandatory)][string]$Query
    )
    $backend = Get-RsSqlBackend
    if ($backend -eq 'docker') {
        $out = docker exec rs_postgres psql -U radonezh -d $Db -v ON_ERROR_STOP=1 -t -A --set=client_min_messages=error -c $Query 2>&1
    } else {
        $prev = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
        try {
            $env:PGPASSWORD = $script:RsPgPass
            $env:PGCLIENTENCODING = 'UTF8'
            $out = & $script:RsPsqlLocal -U radonezh -h $script:RsPgHost -d $Db -v ON_ERROR_STOP=1 -t -A --set=client_min_messages=error -c $Query 2>&1
        } finally { $ErrorActionPreference = $prev }
    }
    if ($LASTEXITCODE -ne 0) { throw "psql failed in $Db : $($out | Out-String)" }
    $clean = @(); foreach ($l in $out) { if ($l -isnot [System.Management.Automation.ErrorRecord]) { $clean += $l } }
    return ($clean -join "`n")
}

function Invoke-SqlFile {
    param(
        [Parameter(Mandatory)][string]$Database,
        [Parameter(Mandatory)][string]$File
    )
    return Invoke-PsqlFile -Database $Database -File $File
}