# devtools.ps1 — хелперы RadonezhSklad
# ВАЖНО: обход бага Docker Desktop на Windows — все SQL идут через docker cp,
# НЕ через pipe `docker exec -i`. Pipe теряет не-ASCII байты.

function Get-Utf8NoBom {
    return New-Object System.Text.UTF8Encoding $false
}

function Write-Utf8File {
    param([Parameter(Mandatory)][string]$Path, [Parameter(Mandatory)][string]$Content)
    $utf8 = New-Object System.Text.UTF8Encoding $false
    [IO.File]::WriteAllText($Path, $Content, $utf8)
}

function Read-Utf8File {
    param([Parameter(Mandatory)][string]$Path)
    $utf8 = New-Object System.Text.UTF8Encoding $false
    return [IO.File]::ReadAllText($Path, $utf8)
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