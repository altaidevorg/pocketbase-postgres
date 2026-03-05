$ErrorActionPreference = "Stop"
$base = "http://127.0.0.1:8090/api"

# Login
$body = @{ identity = "admin@example.com"; password = "1234567890" } | ConvertTo-Json
try {
    $resp = Invoke-RestMethod -Uri "$base/collections/_superusers/auth-with-password" -Method Post -Body $body -ContentType "application/json"
    $token = $resp.token
    Write-Host "Logged in. Token: $token"
} catch {
    Write-Host "Login failed: $_"
    exit 1
}

$headers = @{ Authorization = $token }

# Create Collection
$colBody = @{
    name = "posts"
    type = "base"
    schema = @(
        @{ name = "title"; type = "text"; required = $true }
    )
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri "$base/collections" -Method Post -Body $colBody -ContentType "application/json" -Headers $headers
    Write-Host "Collection 'posts' created."
} catch {
    # If exists, fine
    Write-Host "Collection creation note: $($_.Exception.Message)"
}

# Create Record
$recBody = @{ title = "Hello World" } | ConvertTo-Json
try {
    $rec = Invoke-RestMethod -Uri "$base/collections/posts/records" -Method Post -Body $recBody -ContentType "application/json" -Headers $headers
    Write-Host "Record created: $($rec.id)"
} catch {
    Write-Host "Record creation failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        Write-Host "Response Body: $($reader.ReadToEnd())"
    }
}

# List Records
try {
    $list = Invoke-RestMethod -Uri "$base/collections/posts/records" -Method Get -Headers $headers
    Write-Host "Found $($list.totalItems) records."
} catch {
    Write-Host "List records failed: $($_.Exception.Message)"
     if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        Write-Host "Response Body: $($reader.ReadToEnd())"
    }
}
