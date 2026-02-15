# Simple Auth API Test
Write-Host "=== Testing Auth APIs ===" -ForegroundColor Cyan

# 1. Health Check
Write-Host "`nTest 1: Health Check" -ForegroundColor Yellow
try {
    $r = Invoke-RestMethod -Uri "http://localhost:8080/api/health" -Method GET
    Write-Host "✅ SUCCESS" -ForegroundColor Green
    Write-Host ($r | ConvertTo-Json)
}
catch {
    Write-Host "❌ FAILED: $($_)" -ForegroundColor Red
}

# 2. Registration
Write-Host "`nTest 2: User Registration" -ForegroundColor Yellow
$regBody = @{
    email    = "test$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
    password = "SecurePass123!"
    name     = "Test User"
}
try {
    $r = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/register" -Method POST -Body ($regBody | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✅ SUCCESS" -ForegroundColor Green
    Write-Host ($r | ConvertTo-Json)
}
catch {
    Write-Host "❌ FAILED: Status=$($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    Write-Host $_.ErrorDetails.Message
}

# 3. Login Test
Write-Host "`nTest 3: Login" -ForegroundColor Yellow
$loginBody = @{
    email    = "demo@spectra.id"
    password = "demo123"
}
try {
    $r = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body ($loginBody | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✅ SUCCESS - Token received" -ForegroundColor Green
    Write-Host "Token: $($r.token.Substring(0,30))..."
}
catch {
    Write-Host "❌ FAILED: Status=$($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    Write-Host $_.ErrorDetails.Message
}

Write-Host "`n=== Tests Complete ===" -ForegroundColor Cyan
