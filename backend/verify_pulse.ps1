$ErrorActionPreference = "Stop"
$baseUrl = "http://localhost:8080/api"

# Define user credentials
$Email = "yogisyahroni766.ysr@gmai.com"
$Password = "Namakamu766!!"

Write-Host "1. Logging in..."
$loginBody = @{ email = $Email; password = $Password } | ConvertTo-Json
try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "   Success! Token obtained."
}
catch {
    Write-Error "   Login Failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $respBody = $reader.ReadToEnd()
        Write-Error "   Server Response: $respBody"
    }
    exit 1
}

$headers = @{ Authorization = "Bearer $token" }

Write-Host "2. Creating Collection..."
$colBody = @{ name = "Test Collection $(Get-Random)"; description = "Pulse Test" } | ConvertTo-Json
try {
    $col = Invoke-RestMethod -Uri "$baseUrl/collections" -Method Post -Headers $headers -Body $colBody -ContentType "application/json"
    $colId = $col.id
    Write-Host "   Success! Collection ID: $colId"
}
catch {
    $err = "Create Collection Failed: $($_.Exception.Message)"
    Write-Error $err
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $respBody = $reader.ReadToEnd()
        Write-Error "   Server Response: $respBody"
    }
    exit 1
}

Write-Host "3. Creating Dashboard..."
$dashBody = @{ name = "Pulse Dashboard"; description = "Pulse Test Dash"; collectionId = $colId; layout = @{} } | ConvertTo-Json
try {
    $dash = Invoke-RestMethod -Uri "$baseUrl/dashboards" -Method Post -Headers $headers -Body $dashBody -ContentType "application/json"
    $dashId = $dash.id
    Write-Host "   Success! Dashboard ID: $dashId"
}
catch {
    $err = "Create Dashboard Failed: $($_.Exception.Message)"
    Write-Error $err
    $err | Out-File "dashboard_error.txt"
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $respBody = $reader.ReadToEnd()
        Write-Error "   Server Response: $respBody"
        $respBody | Out-File "dashboard_response.txt"
    }
    exit 1
}

Write-Host "4. Creating Pulse..."
$pulseBody = @{
    name         = "Integration Test Pulse";
    dashboard_id = $dashId;
    schedule     = "0 9 * * 1";
    channel_type = "slack";
    config       = @{ channel = "#test-alerts" };
    is_active    = $true
} | ConvertTo-Json -Depth 5

try {
    $pulse = Invoke-RestMethod -Uri "$baseUrl/pulses" -Method Post -Headers $headers -Body $pulseBody -ContentType "application/json"
    $pulseId = $pulse.id
    Write-Host "   Success! Pulse ID: $pulseId"
}
catch {
    $err = "Create Pulse Failed: $($_.Exception.Message)"
    Write-Error $err
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $respBody = $reader.ReadToEnd()
        Write-Error "   Server Response: $respBody"
    }
    exit 1
}

Write-Host "5. Triggering Pulse..."
try {
    $trigger = Invoke-RestMethod -Uri "$baseUrl/pulses/$pulseId/trigger" -Method Post -Headers $headers -ContentType "application/json"
    Write-Host "   Success! Pulse Triggered."
}
catch {
    Write-Error "   Trigger Pulse Failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $respBody = $reader.ReadToEnd()
        Write-Error "   Server Response: $respBody"
    }
    exit 1
}

Write-Host "VERIFICATION COMPLETE: Pulse feature behaves correctly."
