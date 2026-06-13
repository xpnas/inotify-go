$ErrorActionPreference = "Stop"

Write-Host "Starting backend on http://127.0.0.1:8000"
Start-Process powershell -WindowStyle Hidden -ArgumentList "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "cd '$PSScriptRoot\..\backend'; go run ./cmd/inotify"

Write-Host "Starting frontend on http://127.0.0.1:9000"
Set-Location "$PSScriptRoot\..\frontend"
npm run dev
