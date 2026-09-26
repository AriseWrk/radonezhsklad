# dev-down.ps1 — остановить всё, что поднял dev-up.ps1
#
#   .\dev-down.ps1              — только бэкенды + фронт (Docker оставить)
#   .\dev-down.ps1 -Docker      — ещё и docker compose down
#   .\dev-down.ps1 -Clean       — удалить логи и .run

[CmdletBinding()]
param(
    [switch]$Docker,
    [switch]$Clean
)

$ErrorActionPreference = 'Continue'
$root = $PSScriptRoot
Set-Location $root

$runDir = Join-Path $root '.run'
$logDir = Join-Path $root 'logs'
$pidsFile = Join-Path $runDir 'pids.json'

if (-not (Test-Path $pidsFile)) {
    Write-Host 'Нечего останавливать (нет .run\pids.json)' -ForegroundColor Yellow
    return
}

$raw = Get-Content $pidsFile -Raw
if ($raw) {
    $obj = $raw | ConvertFrom-Json
    foreach ($prop in $obj.PSObject.Properties) {
        $svc = $prop.Name
        $svcPid = $prop.Value
        $proc = Get-Process -Id $svcPid -ErrorAction SilentlyContinue
        if ($proc) {
            try {
                Stop-Process -Id $svcPid -Force -ErrorAction Stop
                Write-Host "== ${svc}: остановлен (PID $svcPid)" -ForegroundColor Green
            } catch {
                Write-Host "== ${svc}: не удалось остановить PID $svcPid — $($_.Exception.Message)" -ForegroundColor Yellow
            }
        } else {
            Write-Host "== ${svc}: уже не работает (PID $svcPid)" -ForegroundColor DarkGray
        }
    }
}

# Vite spawn'ит детей — добьём по дереву node
Get-Process node -ErrorAction SilentlyContinue | ForEach-Object {
    try { Stop-Process -Id $_.Id -Force -ErrorAction Stop } catch {}
}

Remove-Item $pidsFile -ErrorAction SilentlyContinue

if ($Docker) {
    Write-Host '==> docker compose down' -ForegroundColor Cyan
    docker compose down
}

if ($Clean) {
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $logDir, $runDir
    Write-Host 'Логи и .run удалены' -ForegroundColor Green
}

Write-Host ''
Write-Host 'Готово.' -ForegroundColor Green