# start-local.ps1 — starts all three services as background jobs and tails their logs.
# Usage: .\scripts\start-local.ps1
# Stop everything: .\scripts\start-local.ps1 -Stop

param([switch]$Stop)

$GoExe  = "C:\Program Files\Go\bin\go.exe"
$Root   = Split-Path $PSScriptRoot -Parent

if ($Stop) {
    Get-Job -Name "user-svc","order-svc","gateway" -ErrorAction SilentlyContinue | Stop-Job | Remove-Job
    Write-Host "All services stopped." -ForegroundColor Yellow
    exit
}

# Kill any leftover jobs from a previous run
Get-Job -Name "user-svc","order-svc","gateway" -ErrorAction SilentlyContinue | Stop-Job | Remove-Job

Write-Host "`nStarting services..." -ForegroundColor Cyan

$userJob = Start-Job -Name "user-svc" -ScriptBlock {
    param($go, $root)
    & $go run "$root\cmd\user-svc" 2>&1
} -ArgumentList $GoExe, $Root

$orderJob = Start-Job -Name "order-svc" -ScriptBlock {
    param($go, $root)
    & $go run "$root\cmd\order-svc" 2>&1
} -ArgumentList $GoExe, $Root

# Give the backend services a moment to bind their ports
Start-Sleep -Seconds 3

$gwJob = Start-Job -Name "gateway" -ScriptBlock {
    param($go, $root)
    & $go run "$root\cmd\gateway" 2>&1
} -ArgumentList $GoExe, $Root

Start-Sleep -Seconds 2

# Generate a dev token
$TOKEN = & $GoExe run "$Root\cmd\gentoken" 2>$null
Write-Host "`n=== Dev JWT ===" -ForegroundColor Green
Write-Host $TOKEN
Write-Host ""

Write-Host "=== Quick test commands ===" -ForegroundColor Green
Write-Host ('$TOKEN = "' + $TOKEN + '"')
Write-Host 'curl.exe http://localhost:8080/healthz'
Write-Host 'curl.exe -X POST http://localhost:8080/users -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\"}"'
Write-Host ""
Write-Host "Press Ctrl+C to stop tailing logs. Services keep running as background jobs." -ForegroundColor DarkGray
Write-Host "Run:  .\scripts\start-local.ps1 -Stop   to kill all services.`n" -ForegroundColor DarkGray

# Tail logs from all three jobs
try {
    while ($true) {
        foreach ($job in @($userJob, $orderJob, $gwJob)) {
            $lines = Receive-Job $job -ErrorAction SilentlyContinue
            foreach ($line in $lines) {
                $color = switch ($job.Name) {
                    "user-svc"  { "Blue" }
                    "order-svc" { "Magenta" }
                    "gateway"   { "Cyan" }
                }
                Write-Host "[$($job.Name)] $line" -ForegroundColor $color
            }
        }
        Start-Sleep -Milliseconds 500
    }
} finally {
    # Ctrl+C lands here — leave jobs running
}
