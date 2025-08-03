$base64Content = Get-Content base64.txt -Raw
$json = '{"pdf_base64":"' + $base64Content + '"}'
Write-Host "Making API request..."
$response = Invoke-WebRequest -Uri "http://localhost:4000/parse-transcript" -Method POST -Headers @{"Content-Type"="application/json"} -Body $json
Write-Host "Response:"
Write-Host $response.Content
Write-Host "Response length:" $response.Content.Length
$response.Content | Out-File -FilePath "response.json" -Encoding UTF8
Write-Host "Full response saved to response.json" 