$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Src = Join-Path $Root "src"
$Web = Join-Path $Root "web"
$Docs = Join-Path $Root "docs"

if (Test-Path $Docs) {
    Get-ChildItem $Docs -Exclude ".gitkeep" | Remove-Item -Recurse -Force
} else {
    New-Item -ItemType Directory -Path $Docs | Out-Null
}

Push-Location $Src
try {
    $env:GOOS = "js"
    $env:GOARCH = "wasm"
    go build -o (Join-Path $Docs "flappy.wasm") ./cmd
} finally {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Pop-Location
}

$goroot = go env GOROOT
$wasmExec = Join-Path $goroot "lib\wasm\wasm_exec.js"
if (-not (Test-Path $wasmExec)) {
    $wasmExec = Join-Path $goroot "misc\wasm\wasm_exec.js"
}
Copy-Item $wasmExec (Join-Path $Docs "wasm_exec.js")

Copy-Item (Join-Path $Web "index.html") $Docs
Copy-Item (Join-Path $Web "manifest.webmanifest") $Docs
Copy-Item (Join-Path $Web "sw.js") $Docs
Copy-Item (Join-Path $Web "icons") (Join-Path $Docs "icons") -Recurse

Write-Host "Web build complete: $Docs"
