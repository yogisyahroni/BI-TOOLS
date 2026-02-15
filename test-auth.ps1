# Auth API Testing Script
$baseUrl = "http://localhost:8080/api"

Write-Host "=== Testing InsightEngine Auth APIs ===" -ForegroundColor Cyan

# Test 1: Health Check
Write-Host "`n[1] Testing Health Endpoint..." -ForegroundColor Yellow
$response = try {
    Invoke-WebRequest -Uri "$baseUrl/health" -Method GET -UseBasicParsing
    $response | Select-Object StatusCode, Content | Format-List
}
catch {
    Write-Host "ERROR: $_" -ForegroundColor Red
}


# Test 2: User Registration
Write-Host "`n[2] Testing User Registration..." -ForegroundColor Yellow
$registerBody = @{
    email    = "test_$(Get-Date - Format 'yyyyMMddHHmmss')@example.com"
    password = "SecurePass123!"
    name     = "Test User"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/register" -Method POST -Body $registerBody -ContentType "application/json"
    $response | Select-Object StatusCode, Content | Format-List
}
catch {
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    Write-Host "Response: $($_.ErrorDetails.Message)" -ForegroundColor Red
}

# Test 3: Login (with demo user if exists)
Write-Host "`n[3] Testing Login Endpoint..." -ForegroundColor Yellow
$loginBody = @{
    email    = "demo@spectra.id"
    password = "demo123"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $response | Select-Object StatusCode, Content | Format-List
    
    # Extract token if successful
    if ($response.StatusCode -eq 200) {
        $token = ($response.Content | ConvertFrom-Json).token
        Write-Host "Token received: $($token.Substring(0, 20))..." -ForegroundColor Green
    }
}
catch {
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    Write-Host "Response: $($_.ErrorDetails.Message)" -ForegroundColor Red
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan
