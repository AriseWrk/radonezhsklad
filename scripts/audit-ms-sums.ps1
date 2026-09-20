$ErrorActionPreference = 'Stop'
[Threading.Thread]::CurrentThread.CurrentCulture = [System.Globalization.CultureInfo]::InvariantCulture

$root = (Get-Location).Path
function Pth($rel) { Join-Path $root $rel }

$msTypes = @(
    @{ ms = 'supply'; our = 'receipt'  },
    @{ ms = 'enter';  our = 'receipt'  },
    @{ ms = 'demand'; our = 'shipment' },
    @{ ms = 'move';   our = 'transfer' },
    @{ ms = 'loss';   our = 'writeoff' }
)

# ============ 1) MS-суммы ============
$msByOur = @{}

foreach ($pair in $msTypes) {
    $t = $pair.ms
    Write-Host "MS: /entity/$t …" -ForegroundColor Cyan -NoNewline
    $pageSize = 1000
    $offset = 0
    $sumKop = [int64]0
    $cnt = 0
    $pages = 0
    while ($true) {
        $r = MsApi-Get "/entity/$t" @{ limit = $pageSize; offset = $offset }
        $rows = $r.rows
        if (-not $rows -or $rows.Count -eq 0) { break }
        foreach ($d in $rows) {
            if ($d.sum) { $sumKop += [int64]$d.sum }
            $cnt++
        }
        $pages++
        if ($rows.Count -lt $pageSize) { break }
        $offset += $pageSize
        Start-Sleep -Milliseconds 120
    }
    $rub = [Math]::Round($sumKop / 100.0, 2)
    Write-Host "  docs=$cnt  pages=$pages  sum=$rub" -ForegroundColor Green

    if (-not $msByOur.ContainsKey($pair.our)) {
        $msByOur[$pair.our] = @{ docs = 0; sum = [double]0 }
    }
    $msByOur[$pair.our].docs += $cnt
    $msByOur[$pair.our].sum  += $rub
}

# ============ 2) Наши суммы ============
Write-Host "`nНаши суммы из radonezh_warehouse …" -ForegroundColor Cyan

$ourByType = @{}
$ourRaw = Invoke-SqlQuery -Db radonezh_warehouse -Query @"
SELECT d.type,
       COUNT(DISTINCT d.id)::text,
       ROUND(COALESCE(SUM(di.quantity * di.price), 0)::numeric, 2)::text
FROM documents d
LEFT JOIN document_items di ON di.document_id = d.id
WHERE d.external_id IS NOT NULL
GROUP BY d.type
ORDER BY d.type;
"@

foreach ($line in ($ourRaw -split "`n")) {
    if ($line -match '^\s*(receipt|shipment|transfer|writeoff)\s*\|\s*(\d+)\s*\|\s*([\d\.]+)\s*$') {
        $tp = $matches[1]
        $docs = [int]$matches[2]
        $sum = [double]$matches[3]
        $ourByType[$tp] = @{ docs = $docs; sum = $sum }
    }
}

# ============ 3) Сверка ============
Write-Host "`n=== СВЕРКА ===" -ForegroundColor Cyan
""
"{0,-10} | {1,8} | {2,8} | {3,16} | {4,16} | {5,14}" -f `
    'type','ms docs','our docs','ms sum','our sum','diff'
"{0,-10} | {1,8} | {2,8} | {3,16} | {4,16} | {5,14}" -f `
    ('-'*10), ('-'*8), ('-'*8), ('-'*16), ('-'*16), ('-'*14)

$totalMs = 0.0; $totalOur = 0.0
foreach ($tp in @('receipt','shipment','transfer','writeoff')) {
    $ms = if ($msByOur.ContainsKey($tp)) { $msByOur[$tp] } else { @{ docs = 0; sum = 0.0 } }
    $our = if ($ourByType.ContainsKey($tp)) { $ourByType[$tp] } else { @{ docs = 0; sum = 0.0 } }
    $diff = [Math]::Round($our.sum - $ms.sum, 2)
    $totalMs += $ms.sum; $totalOur += $our.sum

    "{0,-10} | {1,8} | {2,8} | {3,16:N2} | {4,16:N2} | {5,14:N2}" -f `
        $tp, $ms.docs, $our.docs, $ms.sum, $our.sum, $diff
}
"{0,-10} | {1,8} | {2,8} | {3,16:N2} | {4,16:N2} | {5,14:N2}" -f `
    'ИТОГО', '', '', $totalMs, $totalOur, ([Math]::Round($totalOur - $totalMs, 2))