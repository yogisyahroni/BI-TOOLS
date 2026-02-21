$ErrorActionPreference = "Stop"

function Test-Endpoint {
    param($Uri, $Method = "GET", $Headers = @{}, $Body = $null)
    try {
        if ($Body) {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -Headers $Headers -Body $Body -ContentType "application/json"
        }
        else {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -Headers $Headers
        }
        return $response
    }
    catch {
        Write-Host "Error calling $Uri"
        Write-Host $_.Exception.Message
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader $_.Exception.Response.GetResponseStream()
            $responseBody = $reader.ReadToEnd()
            Write-Host "Response Body: $responseBody"
        }
        return $null
    }
}

Write-Host "Checking Health..."
$health = Test-Endpoint -Uri "http://localhost:8080/api/health"
if ($health) {
    Write-Host "Health Check Passed: $($health | ConvertTo-Json -Depth 2)"
}
else {
    Write-Error "Health Check Failed. Backend might not be running."
}

Write-Host "`nLogging in..."
$loginBody = @{
    email    = "yogisyahroni766.ysr@gmai.com"
    password = "Namakamu766!!"
} | ConvertTo-Json

$loginResponse = Test-Endpoint -Uri "http://localhost:8080/api/auth/login" -Method "POST" -Body $loginBody
if ($loginResponse -and $loginResponse.token) {
    $token = $loginResponse.token
    Write-Host "Login Successful. Token received."
    
    $headers = @{
        Authorization = "Bearer $token"
    }

    Write-Host "`nFetching Alerts..."
    $alerts = Test-Endpoint -Uri "http://localhost:8080/api/alerts" -Headers $headers
    if ($alerts) {
        Write-Host "Alerts Fetch Successful!"
        Write-Host "Alerts Count: $($alerts.total)"
    }
    else {
        Write-Error "Failed to fetch alerts."
    }

    Write-Host "`nFetching Audit Logs (just to check table existence)..."
    # Note: /api/audit-logs might not exist or require different permissions, but let's try if route exists
    # Assuming standard route /api/audit-logs or similar. If not found, ignore.
    try {
        $auditLogs = Invoke-RestMethod -Uri "http://localhost:8080/api/audit-logs" -Method "GET" -Headers $headers
        Write-Host "Audit Logs Fetch Successful!"
    }
    catch {
        Write-Host "Audit Logs fetch failed (expected if route not defined or 403): $($_.Exception.Message)"
    }

}
else {
    Write-Error "Login Failed."
}
