$ErrorActionPreference = "Continue"

Write-Host "Force Restarting Backend..."

# 1. Kill invalid processes
Write-Host "1. Cleaning up ports and processes..."
$port = 8080
try {
    $connections = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($connections) {
        foreach ($conn in $connections) {
            $pid_to_kill = $conn.OwningProcess
            if ($pid_to_kill -gt 0) {
                Write-Host "   Found process $($pid_to_kill) on port $($port). Killing..."
                Stop-Process -Id $pid_to_kill -Force -ErrorAction SilentlyContinue
            }
        }
    }
}
catch {
    Write-Host "   Warning: Error cleaning up port $($port): $($_.Exception.Message)"
}

# Also try by name
Stop-Process -Name "backend" -Force -ErrorAction SilentlyContinue
Stop-Process -Name "main" -Force -ErrorAction SilentlyContinue

# 2. Go Mod Tidy
Write-Host "2. Running go mod tidy..."
go mod tidy

# 3. Build Binary
Write-Host "3. Building backend binary..."
if (Test-Path "backend.exe") {
    Remove-Item "backend.exe" -Force -ErrorAction SilentlyContinue
}

# Ensure we stop on build failure
$buildOutput = go build -o backend.exe main.go 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed with exit code $($LASTEXITCODE)"
    Write-Host "Output: $($buildOutput)"
    exit 1
}

if (-not (Test-Path "backend.exe")) {
    Write-Host "Build failed: backend.exe not created."
    exit 1
}

# 4. Run Backend with Log Redirection
Write-Host "4. Starting backend..."
$logFile = "backend_force_restart.log"
if (Test-Path $logFile) { Remove-Item $logFile -Force }

# Start process in background
$proc = Start-Process -FilePath "./backend.exe" -RedirectStandardOutput $logFile -RedirectStandardError $logFile -PassThru -NoNewWindow

if ($proc) {
    Write-Host "   Backend started with PID: $($proc.Id)"
    Write-Host "   Logs are being written to: $($logFile)"
}
else {
    Write-Host "Failed to start backend process."
    exit 1
}

# 5. Wait for server to be ready
Write-Host "5. Waiting for server readiness..."
$timeout = 600
$timer = 0
$started = $false

while ($timer -lt $timeout) {
    if (Test-Path $logFile) {
        $content = Get-Content $logFile -Tail 20 -ErrorAction SilentlyContinue
        if ($content) {
            if ($content -match "Server running") {
                Write-Host "`nServer is UP and RUNNING!"
                $started = $true
                break
            }
            if ($content -match "FATAL" -or $content -match "panic:") {
                Write-Host "`nServer failed to start! Check logs."
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
    Write-Host "`nServer failed to start within $($timeout) seconds. Check $($logFile)."
    exit 1
}

# 6. Verify Health
Write-Host "`n6. Verifying Health Endpoint..."
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/health" -Method Get -ErrorAction Stop
    Write-Host "   Health Check: Success ($($response.status))"
}
catch {
    Write-Host "   Health Check Failed: $($_.Exception.Message)"
    exit 1
}

exit 0
