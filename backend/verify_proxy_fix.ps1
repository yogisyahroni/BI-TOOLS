$ErrorActionPreference = "Stop"

# Verify Proxy Connection
Write-Host "Verifying Proxy Connection (Port 3000 -> 8080)..."

try {
    # 1. Health Check via Proxy
    Write-Host "1. Checking /api/go/health..."
    $health = Invoke-RestMethod -Uri "http://localhost:3000/api/go/health" -Method Get
    Write-Host "   Health: OK ($($health.status))"
}
catch {
    Write-Host "   Health Check FAILED: $($_.Exception.Message)"
    exit 1
}

try {
    # 2. Alerts Endpoint via Proxy
    Write-Host "2. Checking /api/go/alerts..."
    $alerts = Invoke-RestMethod -Uri "http://localhost:3000/api/go/alerts" -Method Get
    Write-Host "   Alerts Count: $($alerts.Count)"
    
    if ($alerts.Count -ge 0) {
        Write-Host "✅ ALERTS API PROXY VERIFIED!"
    }
    else {
        Write-Error "Invalid response format from proxy."
    }
}
catch {
    Write-Host "❌ ALERTS PROXY FAILED: $($_.Exception.Message)"
    # Start-Process -FilePath "curl" -ArgumentList "-v http://localhost:3000/api/go/alerts"
    exit 1
}
