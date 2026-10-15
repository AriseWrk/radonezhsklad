# dev-test.ps1 — запуск unit-тестов фронта (vitest)
#   .\dev-test.ps1          — один прогон (npm test)
#   .\dev-test.ps1 -Watch   — watch-режим (npm run test:watch)
[CmdletBinding()]
param(
    [switch]$Watch
)
$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot
$webDir = Join-Path $root 'web'

if (-not (Test-Path (Join-Path $webDir 'node_modules'))) {
    Write-Host '==> npm install (первый запуск, подожди)' -ForegroundColor Cyan
    Push-Location $webDir; try { npm install } finally { Pop-Location }
}

Push-Location $webDir
try {
    if ($Watch) {
        Write-Host '==> vitest watch' -ForegroundColor Cyan
        npm run test:watch
    } else {
        Write-Host '==> vitest run' -ForegroundColor Cyan
        npm test
    }
} finally {
    Pop-Location
}