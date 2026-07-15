$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Docs = Join-Path $Root "docs"
$Port = 8080

& (Join-Path $Root "scripts\build-web.ps1")

Write-Host "Serving $Docs at http://localhost:$Port"
Write-Host "Press Ctrl+C to stop."

Push-Location $Docs
try {
    python -m http.server $Port
} finally {
    Pop-Location
}
