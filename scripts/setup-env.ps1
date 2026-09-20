$ErrorActionPreference = 'Stop'
$utf8 = [Text.UTF8Encoding]::new($false)
$root = Split-Path $PSScriptRoot -Parent
$pkgEnv = Join-Path $root "deploy-package\services-env"
if (-not (Test-Path $pkgEnv)) { throw "deploy-package/services-env not found. Unpack ZIP first." }

foreach ($svc in @('auth','product','warehouse','order','gateway','audit')) {
    $src = Join-Path $pkgEnv "$svc.env"
    $dst = Join-Path $root "services\$svc\.env"
    if (-not (Test-Path $src)) { Write-Host "SKIP $svc"; continue }
    Copy-Item $src $dst -Force
    Write-Host "  $svc -> services\$svc\.env" -ForegroundColor Green
}

$tokenSrc = Join-Path $root "deploy-package\.secrets\moysklad.token"
$tokenDst = "D:\Radonezhsklad\.secrets\moysklad.token"
if (Test-Path $tokenSrc) {
    $dstDir = Split-Path $tokenDst -Parent
    if (-not (Test-Path $dstDir)) { New-Item -ItemType Directory -Path $dstDir | Out-Null }
    Copy-Item $tokenSrc $tokenDst -Force
    Write-Host "  moysklad.token -> $tokenDst" -ForegroundColor Green
}

Write-Host "`nDone." -ForegroundColor Cyan