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
        Write-Host "Exception: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            # Try to read status code if available
            $statusCode = $_.Exception.Response.StatusCode
            Write-Host "Status Code: $statusCode"
             
            # Try to read response body
            try {
                $reader = New-Object System.IO.StreamReader $_.Exception.Response.GetResponseStream()
                $responseBody = $reader.ReadToEnd()
                Write-Host "Response Body: $responseBody"
            }
            catch {
                Write-Host "Could not read response body."
            }
        }
        return $null
    }
}

Write-Host "Checking Health..."
# Retry health check loop
for ($i = 0; $i -lt 30; $i++) {
    try {
        $health = Invoke-RestMethod -Uri "http://localhost:8080/api/health" -ErrorAction Stop
        if ($health) {
            Write-Host "Health Check Passed: $($health | ConvertTo-Json -Depth 2)"
            break
        }
    }
    catch {
        Write-Host "Waiting for backend..."
        Start-Sleep -Seconds 2
    }
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

    Write-Host "`nFetching Stories..."
    $stories = Test-Endpoint -Uri "http://localhost:8080/api/stories" -Headers $headers -Method "GET"
    if ($stories -ne $null) {
        Write-Host "Stories Fetch Successful!"
        Write-Host "Stories: $($stories | ConvertTo-Json -Depth 2)"
    }
    else {
        Write-Error "Failed to fetch stories."
    }

}
else {
    Write-Error "Login Failed."
}
