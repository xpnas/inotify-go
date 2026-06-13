$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot\..\backend"
go test ./...
Pop-Location

Push-Location "$PSScriptRoot\..\frontend"
npm run lint
npm run build
Pop-Location
