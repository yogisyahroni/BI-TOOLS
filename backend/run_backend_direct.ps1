$ErrorActionPreference = "Continue"

Write-Host "Running Backend with CMD Wrapper..."

# 1. Kill invalid processes
Write-Host "1. Cleaning up ports..."
$port = 8080
try {
    $connections = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($connections) {
        foreach ($conn in $connections) {
            $pid_to_kill = $conn.OwningProcess
            if ($pid_to_kill -gt 0) {
                Stop-Process -Id $pid_to_kill -Force -ErrorAction SilentlyContinue
            }
        }
    }
}
catch {
    Write-Host "   Cleanup warning: $($_.Exception.Message)"
}
Stop-Process -Name "backend" -Force -ErrorAction SilentlyContinue
Stop-Process -Name "main" -Force -ErrorAction SilentlyContinue

# 2. Run Go via CMD to handle redirection properly
Write-Host "2. Starting backend..."
$logFile = "backend_direct.log"
if (Test-Path $logFile) { Remove-Item $logFile -Force }

# Use cmd /c to handle redirection > 2>&1
$proc = Start-Process -FilePath "cmd" -ArgumentList "/c go run main.go > $logFile 2>&1" -PassThru -NoNewWindow

if ($proc) {
    Write-Host "   Backend started with PID: $($proc.Id)"
    Write-Host "   Logs: $logFile"
}
else {
    Write-Error "Failed to start backend."
    exit 1
}

# 3. Wait for server
Write-Host "3. Waiting for startup..."
$timeout = 600
$timer = 0
$started = $false

while ($timer -lt $timeout) {
    if (Test-Path $logFile) {
        $content = Get-Content $logFile -Tail 20 -ErrorAction SilentlyContinue
        if ($content) {
            if ($content -match "Server running") {
                Write-Host "`nServer is UP!"
                $started = $true
                break
            }
            if ($content -match "panic:") {
                Write-Error "`nServer PANIC!"
                Get-Content $logFile -Tail 20
                exit 1
            }
        }
    }
    Start-Sleep -Seconds 1
    $timer++
    if ($timer % 5 -eq 0) { Write-Host -NoNewline "." }
}

if (-not $started) {
    Write-Error "Timeout waiting for server."
    exit 1
}

# 4. Verify Health
Write-Host "`n4. Metrics Check..."
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/health" -Method Get -ErrorAction Stop
    Write-Host "   Health: OK ($($response.status))"
}
catch {
    Write-Host "   Health Check Warn: $($_.Exception.Message)"
}

exit 0
