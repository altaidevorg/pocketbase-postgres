$ErrorActionPreference = "Stop"
$base = "http://127.0.0.1:8090/api"

# Login
$body = @{ identity = "admin@example.com"; password = "1234567890" } | ConvertTo-Json
try {
    $resp = Invoke-RestMethod -Uri "$base/collections/_superusers/auth-with-password" -Method Post -Body $body -ContentType "application/json"
    $token = $resp.token
    Write-Host "Logged in."
} catch {
    Write-Host "Login failed: $_"
    exit 1
}

$headers = @{ Authorization = $token }

# Try to delete posts collection if exists
try {
    $existing = Invoke-RestMethod -Uri "$base/collections/posts" -Method Get -Headers $headers -ErrorAction SilentlyContinue
    if ($existing) {
        Write-Host "Deleting existing posts collection..."
        Invoke-RestMethod -Uri "$base/collections/posts" -Method Delete -Headers $headers
    }
} catch {
    Write-Host "Collection check ignored: $_"
}

# Create Collection
$colBody = @{
    name = "posts"
    type = "base"
    fields = @(
        @{ name = "title"; type = "text"; required = $true }
    )
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri "$base/collections" -Method Post -Body $colBody -ContentType "application/json" -Headers $headers
    Write-Host "Collection 'posts' created."
} catch {
    Write-Host "Collection creation failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
         $reader = New-Object System.IO.StreamReader $_.Exception.Response.GetResponseStream()
         Write-Host "Response Body: $($reader.ReadToEnd())"
    }
    exit 1
}

# Create Record
$recBody = @{ title = "Clean Test" } | ConvertTo-Json
try {
    $rec = Invoke-RestMethod -Uri "$base/collections/posts/records" -Method Post -Body $recBody -ContentType "application/json" -Headers $headers
    Write-Host "Record created: $($rec.id)"
} catch {
    Write-Host "Record creation failed: $($_.Exception.Message)"
    exit 1
}

# List Records
try {
    $list = Invoke-RestMethod -Uri "$base/collections/posts/records?sort=-created" -Method Get -Headers $headers
    Write-Host "Found $($list.totalItems) records."
    if ($list.items.Count -gt 0) {
        Write-Host "First item: $($list.items[0] | ConvertTo-Json -Depth 5)"
    }
} catch {
    Write-Host "List records failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
         $reader = New-Object System.IO.StreamReader $_.Exception.Response.GetResponseStream()
         Write-Host "Response Body: $($reader.ReadToEnd())"
    }
}
