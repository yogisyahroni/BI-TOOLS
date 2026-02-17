$ErrorActionPreference = "Stop"

# 0. Cleanup Port
Write-Host "Killing process on port 8080..."
cmd /c "npx -y kill-port 8080"
Start-Sleep -Seconds 2

# 1. Start Backend in Background
Write-Host "Starting Backend..."
$backendProcess = Start-Process -FilePath "./insight-engine-backend.exe" -PassThru -NoNewWindow -RedirectStandardOutput "backend.log" -RedirectStandardError "backend_err.log"
Start-Sleep -Seconds 5

# 1.5. Register (Try to register, ignore if exists)
$registerUrl = "http://localhost:8080/api/auth/register"
$registerPayload = @{
    email    = "verify_admin@example.com"
    password = "password123"
    username = "verifyadmin"
    fullName = "Mock Admin"
}
$registerJson = $registerPayload | ConvertTo-Json

Write-Host "Registering Mock Admin..."
try {
    $response = Invoke-RestMethod -Uri $registerUrl -Method Post -Body $registerJson -ContentType "application/json"
    Write-Host "Registration successful."
}
catch {
    $e = $_.Exception
    Write-Host "Registration failed. Status: $($e.Response.StatusCode.value__)"
    if ($e.Response) {
        $reader = New-Object System.IO.StreamReader($e.Response.GetResponseStream())
        $body = $reader.ReadToEnd()
        Write-Host "Details: $body"
    }
}

# 2. Login
Write-Host "Authenticating..."
$loginUrl = "http://localhost:8080/api/auth/login"
$loginPayload = @{
    email    = "verify_admin@example.com"
    password = "password123"
} | ConvertTo-Json -Compress
try {
    $loginBox = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method Post -Body $loginPayload -ContentType "application/json"
    $token = $loginBox.token
    Write-Host "Logged in! Token received."
}
catch {
    Write-Host "Login failed. Details:"
    $stream = $_.Exception.Response.GetResponseStream()
    if ($stream) {
        $reader = New-Object IO.StreamReader($stream)
        Write-Host $reader.ReadToEnd()
        $reader.Close()
    }
    Stop-Process -Id $backendProcess.Id -Force
    exit 1
}

# 3. Create Connection
Write-Host "Creating Mock Connection..."
$connPayload = @{
    name     = "TestDB-FuncVerify"
    type     = "postgres"
    host     = "mock"
    port     = 5432
    username = "mock"
    password = "mock"
    database = "mock"
} | ConvertTo-Json -Compress
$connJson = $connPayload # Assuming $connJson is intended to be $connPayload
$headers = @{Authorization = "Bearer $token" } # Assuming $headers is intended to be this
$connResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/connections" -Method Post -Headers $headers -Body $connJson -ErrorAction Stop -ContentType "application/json"
$connId = $connResponse.data.id
Write-Host "Connection Created: $connId"

Write-Host "Connection Created: $connId"

# 3.5 Check Persistence
Write-Host "Listing Connections..."
try {
    $listResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/connections" -Method Get -Headers @{Authorization = "Bearer $token" }
    Write-Host "Found $($listResponse.data.Count) connections"
    $found = $listResponse.data | Where-Object { $_.id -eq $connId }
    if ($found) {
        Write-Host "Connection $connId FOUND in list."
    }
    else {
        Write-Error "Connection $connId NOT FOUND in list!"
        Write-Host "Available IDs: $($listResponse.data.id -join ', ')"
    }
}
catch {
    Write-Host "Failed to list connections. Status: $($_.Exception.Response.StatusCode.value__)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "Details: $($reader.ReadToEnd())"
    }
}

# 4. Test Connection
Write-Host "Testing Connection..."
try {
    $testResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/connections/$connId/test" -Method Post -Headers @{Authorization = "Bearer $token" }
    if ($testResponse.status -eq "success") {
        Write-Host "Connection Test Passed (Mocked)"
    }
    else {
        Write-Error "Connection Test Failed: $($testResponse.message)"
    }
}
catch {
    Write-Host "Connection Test Failed. Status: $($_.Exception.Response.StatusCode.value__)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "Details: $($reader.ReadToEnd())"
    }
    Stop-Process -Id $backendProcess.Id -Force
    exit 1
}

# 5. Schema Discovery
Write-Host "Discovering Schema..."
try {
    $schemaResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/connections/$connId/schema" -Method Get -Headers @{Authorization = "Bearer $token" }
    $tables = $schemaResponse.data
    if ($tables.Count -ge 3 -and $tables[0].name -eq "mock_users") {
        Write-Host "Schema Discovery Passed (Found $($tables.Count) mock tables)"
    }
    else {
        Write-Error "Schema Discovery Failed"
    }
}
catch {
    Write-Host "Schema Discovery Failed. Status: $($_.Exception.Response.StatusCode.value__)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "Details: $($reader.ReadToEnd())"
    }
    Stop-Process -Id $backendProcess.Id -Force
    exit 1
}

# 6. Query Execution
Write-Host "Executing Query..."
$queryPayload = @{
    sql          = "SELECT * FROM mock_users"
    connectionId = $connId
} | ConvertTo-Json -Compress
try {
    $queryResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/queries/execute" -Method Post -Body $queryPayload -Headers @{Authorization = "Bearer $token" } -ContentType "application/json"
    if ($queryResponse.data.rowCount -eq 2) {
        Write-Host "Query Executed Successfully (Returned $($queryResponse.data.rowCount) rows)"
    }
    else {
        Write-Error "Query Execution Failed"
    }
}
catch {
    Write-Host "Query Execution Failed. Status: $($_.Exception.Response.StatusCode.value__)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "Details: $($reader.ReadToEnd())"
    }
    Stop-Process -Id $backendProcess.Id -Force
    exit 1
}

# Cleanup
Stop-Process -Id $backendProcess.Id -Force
Write-Host "Verification Complete!"
