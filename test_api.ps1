# Read the base64 content from the file and clean it
$base64Content = Get-Content "base64.txt" -Raw
$base64Content = $base64Content.Trim() -replace "`n|`r", ""

Write-Host "Base64 content length: $($base64Content.Length)"

# Create the JSON payload
$body = @{
    pdf_base64 = $base64Content
} | ConvertTo-Json

Write-Host "JSON payload length: $($body.Length)"

# Test the API
try {
    Write-Host "Sending request to parse-transcript API..."
    $response = Invoke-WebRequest -Uri "http://localhost:4000/parse-transcript" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body
    Write-Host "Status Code: $($response.StatusCode)"
    Write-Host "Response: $($response.Content)"
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response Body: $responseBody"
    }
} 