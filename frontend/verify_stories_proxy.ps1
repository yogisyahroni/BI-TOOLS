$proxyUrl = "http://localhost:3000/api/stories"
$maxRetries = 60
$retryInterval = 2

Write-Host "Verifying Story Builder proxy at $proxyUrl..."

for ($i = 0; $i -lt $maxRetries; $i++) {
    try {
        $response = Invoke-RestMethod -Uri $proxyUrl -Method Get -ErrorAction Stop
        Write-Host "Success: Got 200 OK (Unexpected without token, but reachable)"
        exit 0
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 401) {
            Write-Host "Success: Got 401 Unauthorized (Reachable via Proxy)"
            exit 0
        }
        elseif ($statusCode -eq 404) {
            # If we get 404, it might be the proxy isn't loaded yet or the rule is wrong.
            # But since next.js returns 404 for unknown routes, it means server IS up. 
            # If next.config.mjs was not applied, it would still return 404.
            # We should wait a bit more to be sure it's not just "starting up".
            Write-Host "Waiting... Got 404 (Server up, but route might be missing or compiling) - Attempt $($i+1)/$maxRetries"
        }
        else {
            if ($statusCode -eq 500) {
                Write-Host "Backend returned 500, but proxy is working."
                exit 0
            }
            Write-Host "Waiting... Status: $statusCode - Attempt $($i+1)/$maxRetries"
        }
    }
    Start-Sleep -Seconds $retryInterval
}
Write-Host "Failed: Timed out or persistent 404."
exit 1
