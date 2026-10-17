# dev-down.ps1 — остановить всё, что поднял dev-up.ps1
#
#   .\dev-down.ps1              — только бэкенды + фронт (Docker оставить)
#   .\dev-down.ps1 -Only warehouse[,gateway] — погасить только указанные сервисы
#   .\dev-down.ps1 -Docker      — ещё и docker compose down (игнорируется с -Only)
#   .\dev-down.ps1 -Clean       — удалить логи и .run (игнорируется с -Only)

[CmdletBinding()]
param(
    [switch]$Docker,
    [switch]$Clean,
    [string[]]$Only
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

$all = [ordered]@{}
$raw = Get-Content $pidsFile -Raw
if ($raw) {
    $obj = $raw | ConvertFrom-Json
    foreach ($prop in $obj.PSObject.Properties) { $all[$prop.Name] = $prop.Value }
}

$targets = @($all.Keys)
if ($Only) {
    foreach ($o in $Only) {
        if (-not $all.Contains($o)) { Write-Host "== ${o}: нет в pids.json (не запущен?)" -ForegroundColor DarkGray }
    }
    $targets = @($targets | Where-Object { $_ -in $Only })
}
if (-not $targets) { Write-Host 'Нечего останавливать' -ForegroundColor Yellow; return }

foreach ($svc in $targets) {
    $svcPid = $all[$svc]
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
    $all.Remove($svc)
}

$stopFront = (-not $Only) -or ('front' -in $Only)
if ($stopFront) {
    Get-Process node -ErrorAction SilentlyContinue | ForEach-Object {
        try { Stop-Process -Id $_.Id -Force -ErrorAction Stop } catch {}
    }
    $all.Remove('front')
}

if ($all.Count -gt 0) {
    $all | ConvertTo-Json | Set-Content -Path $pidsFile -Encoding UTF8
} else {
    Remove-Item $pidsFile -ErrorAction SilentlyContinue
}

if ($Docker -and -not $Only) {
    Write-Host '==> docker compose down' -ForegroundColor Cyan
    docker compose down
}

if ($Clean -and -not $Only) {
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $logDir, $runDir
    Write-Host 'Логи и .run удалены' -ForegroundColor Green
}

Write-Host ''
Write-Host 'Готово.' -ForegroundColor Green