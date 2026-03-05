$ErrorActionPreference = "Stop"

# Login as superuser
$loginUrl = "http://127.0.0.1:8090/api/collections/_superusers/auth-with-password"
$body = @{
    identity = "admin@example.com"
    password = "1234567890"
}
$response = Invoke-RestMethod -Uri $loginUrl -Method Post -Body $body -ContentType "application/x-www-form-urlencoded"
$token = $response.token

Write-Host "Logged in. Token acquired."

# Function to test list records
function Test-ListRecords {
    param (
        [string]$Sort
    )
    $listUrl = "http://127.0.0.1:8090/api/collections/posts/records?page=1&perPage=30&sort=$Sort"
    Write-Host "Testing list with sort: $Sort"
    try {
        $headers = @{ "Authorization" = $token }
        $listResponse = Invoke-RestMethod -Uri $listUrl -Method Get -Headers $headers
        Write-Host "Success! Found $($listResponse.items.Count) records."
        if ($listResponse.items.Count -gt 0) {
            Write-Host "First item: $($listResponse.items[0] | ConvertTo-Json -Depth 5)"
        }
    }
    catch {
        Write-Host "Failed!" -ForegroundColor Red
        Write-Host $_.Exception.Message
        if ($_.Exception.Response) {
             $reader = New-Object System.IO.StreamReader $_.Exception.Response.GetResponseStream()
             $responseBody = $reader.ReadToEnd()
             Write-Host "Response Body: $responseBody"
        }
    }
}

Test-ListRecords -Sort "-created"
Test-ListRecords -Sort "-updated"
Test-ListRecords -Sort "id"
# Admin UI sometimes uses @rowid if not specified? 
# Test-ListRecords -Sort "-@rowid" 
