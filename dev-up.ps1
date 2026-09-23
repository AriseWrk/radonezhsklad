# dev-up.ps1 — поднять весь стек RadonezhSklad в фоне (Windows / PS 7 или 5.1)
#
#   .\dev-up.ps1              — инфра + все бэкенды + фронт
#   .\dev-up.ps1 -NoFront     — только инфра + бэкенды
#   .\dev-up.ps1 -Only warehouse,gateway — конкретные сервисы
#   .\dev-up.ps1 -Rebuild     — пересобрать Go-бинарники
#
# Логи: logs\<service>.log
# PID:  .run\pids.json

[CmdletBinding()]
param(
    [switch]$NoFront,
    [switch]$Rebuild,
    [string[]]$Only
)

$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot
Set-Location $root

$logDir = Join-Path $root 'logs'
$runDir = Join-Path $root '.run'
New-Item -ItemType Directory -Force -Path $logDir, $runDir | Out-Null

function Write-Step($msg) { Write-Host $msg -ForegroundColor Cyan }
function Write-Ok($msg)   { Write-Host $msg -ForegroundColor Green }
function Write-Warn($msg) { Write-Host $msg -ForegroundColor Yellow }

# --- Docker-инфра ---
Write-Step '==> Docker infra (postgres, redis, rabbitmq)...'
docker compose up -d | Out-Null
Start-Sleep -Seconds 2

# --- Go-сервисы ---
$services = @('auth', 'product', 'warehouse', 'order', 'audit', 'gateway')
if ($Only) { $services = $services | Where-Object { $_ -in $Only } }

$existing = @{}
$pidsFile = Join-Path $runDir 'pids.json'
if (Test-Path $pidsFile) {
    $raw = Get-Content $pidsFile -Raw
    if ($raw) {
        $obj = $raw | ConvertFrom-Json
        foreach ($prop in $obj.PSObject.Properties) { $existing[$prop.Name] = $prop.Value }
    }
}

foreach ($svc in $services) {
    $svcDir = Join-Path $root "services\$svc"
    $binDir = Join-Path $svcDir 'bin'
    $bin    = Join-Path $binDir 'api.exe'
    $envFile = Join-Path $svcDir '.env'
    if (-not (Test-Path $envFile)) { Write-Warn "skip ${svc}: нет .env"; continue }

    if ($Rebuild -or -not (Test-Path $bin)) {
        Write-Step "==> build $svc"
        New-Item -ItemType Directory -Force -Path $binDir | Out-Null
        Push-Location $svcDir
        try {
            & go build -o $bin ./cmd/api
            if ($LASTEXITCODE -ne 0) { throw "go build failed for $svc" }
        } finally { Pop-Location }
    }

    $pid_old = $existing[$svc]
    if ($pid_old) {
        $proc = Get-Process -Id $pid_old -ErrorAction SilentlyContinue
        if ($proc) {
            Write-Warn "== ${svc}: уже запущен (PID $pid_old), пропускаю"
            continue
        }
    }

    # Читаем .env вручную — чтобы сервис получил те же переменные
    $envLines = Get-Content $envFile
    foreach ($line in $envLines) {
        $trim = $line.Trim()
        if (-not $trim -or $trim.StartsWith('#')) { continue }
        $eq = $trim.IndexOf('=')
        if ($eq -lt 1) { continue }
        $k = $trim.Substring(0, $eq).Trim()
        $v = $trim.Substring($eq + 1).Trim().Trim('"')
        [Environment]::SetEnvironmentVariable($k, $v, 'Process')
    }

    $logOut = Join-Path $logDir "$svc.log"
    $logErr = Join-Path $logDir "$svc.err.log"
    $proc = Start-Process -FilePath $bin -WorkingDirectory $svcDir `
        -RedirectStandardOutput $logOut -RedirectStandardError $logErr `
        -PassThru -WindowStyle Hidden
    Start-Sleep -Milliseconds 300
    if ($proc.HasExited) {
        Write-Warn "== ${svc}: упал сразу, смотри $logErr"
        continue
    }
    $existing[$svc] = $proc.Id
    Write-Ok "== ${svc}: запущен (PID $($proc.Id))  log: $logOut"
}

$existing | ConvertTo-Json | Set-Content -Path $pidsFile -Encoding UTF8

# --- Frontend (vite) ---
if (-not $NoFront) {
    $webDir = Join-Path $root 'web'
    if (-not (Test-Path (Join-Path $webDir 'node_modules'))) {
        Write-Step '==> npm install (первый запуск, подожди)'
        Push-Location $webDir; try { npm install } finally { Pop-Location }
    }
    $frontPid = $existing['front']
    if ($frontPid -and (Get-Process -Id $frontPid -ErrorAction SilentlyContinue)) {
        Write-Warn "== front: уже запущен (PID $frontPid)"
    } else {
        $logOut = Join-Path $logDir 'front.log'
        $logErr = Join-Path $logDir 'front.err.log'
        $proc = Start-Process -FilePath 'cmd.exe' -ArgumentList '/c','npm run dev' `
            -WorkingDirectory $webDir `
            -RedirectStandardOutput $logOut -RedirectStandardError $logErr `
            -PassThru -WindowStyle Hidden
        Start-Sleep -Milliseconds 800
        $existing['front'] = $proc.Id
        $existing | ConvertTo-Json | Set-Content -Path $pidsFile -Encoding UTF8
        Write-Ok "== front (vite): запущен (PID $($proc.Id))  log: $logOut"
        Write-Host '   фронт обычно на http://localhost:5173' -ForegroundColor Gray
    }
}

Write-Host ''
Write-Ok 'Готово. Остановить: .\dev-down.ps1'