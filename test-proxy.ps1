# Test Frontend Proxy
Write-Host "=== Testing Frontend API Proxy ===" -ForegroundColor Cyan

# Test via Next.js proxy (this is how the frontend calls it)
Write-Host "`nTesting via Frontend Proxy (/api/go/auth/register)..." -ForegroundColor Yellow
$regBody = @{
    email    = "proxytest$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
    password = "SecurePass123!"
    name     = "Proxy Test User"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri "http://localhost:3000/api/go/auth/register" -Method POST -Body $regBody -ContentType "application/json" -UseBasicParsing
    Write-Host "✅ SUCCESS - Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)" -ForegroundColor Green
}
catch {
    Write-Host "❌ FAILED - Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    try {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $reader.DiscardBufferedData()
        Write-Host "Error: $($reader.ReadToEnd())" -ForegroundColor Yellow
    }
    catch {
        Write-Host "Could not read error body" -ForegroundColor Gray
    }
}

# Test direct backend (for comparison)
Write-Host "`nTesting Direct Backend (/api/auth/register)..." -ForegroundColor Yellow
$ regBody = @{
    email    = "directtest$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
    password = "SecurePass123!"
    name     = "Direct Test User"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/auth/register" -Method POST -Body $regBody -ContentType "application/json" -UseBasicParsing
    Write-Host "✅ SUCCESS - Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)" -ForegroundColor Green
}
catch {
    Write-Host "❌ FAILED - Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
}

Write-Host "`n=== Complete ===" -ForegroundColor Cyan
