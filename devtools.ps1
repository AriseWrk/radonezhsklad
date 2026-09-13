# devtools.ps1 — хелперы для работы с UTF-8 в PowerShell 5.1
# Подключение:  . .\devtools.ps1

$script:Utf8NoBom = New-Object System.Text.UTF8Encoding $false

function Write-Utf8File {
    param(
        [Parameter(Mandatory)] [string]$Path,
        [Parameter(Mandatory)] [string]$Content
    )
    [IO.File]::WriteAllText($Path, $Content, $script:Utf8NoBom)
}

function Read-Utf8File {
    param([Parameter(Mandatory)] [string]$Path)
    [IO.File]::ReadAllText($Path, $script:Utf8NoBom)
}

function Invoke-SqlFile {
    param(
        [Parameter(Mandatory)] [string]$Path,
        [Parameter(Mandatory)] [string]$Database,
        [string]$User      = 'radonezh',
        [string]$Container = 'rs_postgres'
    )
    Read-Utf8File $Path | docker exec -i $Container psql -U $User -d $Database
}

function Set-ConsoleUtf8 {
    chcp 65001 | Out-Null
    [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
    Write-Host "OK: консоль переключена в UTF-8"
}