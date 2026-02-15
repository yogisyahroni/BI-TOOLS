# Detailed Auth Test
Write-Host "=== Detailed Auth Test ===" -ForegroundColor Cyan

# Test Registration with full error details
Write-Host "`nTesting Registration..." -ForegroundColor Yellow
$regBody = @{
    email    = "uniquetest$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
    password = "SecurePass123!"
    name     = "Test User"
} | ConvertTo-Json

Write-Host "Request Body: $regBody" -ForegroundColor Gray

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/auth/register" -Method POST -Body $regBody -ContentType "application/json" -UseBasicParsing
    Write-Host "✅ SUCCESS - Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)" -ForegroundColor Green
}
catch {
    Write-Host "❌ FAILED" -ForegroundColor Red
    Write-Host "Status Code: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Yellow
    Write-Host "Status Description: $($_.Exception.Response.StatusDescription)" -ForegroundColor Yellow
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $reader.BaseStream.Position = 0
    $reader.DiscardBufferedData()
    $responseBody = $reader.ReadToEnd()
    Write-Host "Response Body: $responseBody" -ForegroundColor Yellow
}

Write-Host "`n=== Complete ===" -ForegroundColor Cyan
