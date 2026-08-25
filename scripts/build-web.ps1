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
$configSrc = Join-Path $Web "config.js"
if (Test-Path $configSrc) {
    Copy-Item $configSrc $Docs
} else {
    @(
        'window.FLAPPY_SCORE_API_URL = "";'
        'window.FLAPPY_SCORE_API_KEY = "";'
    ) | Set-Content (Join-Path $Docs "config.js")
}

$BuildId = (Get-Date).ToUniversalTime().ToString("yyyyMMddHHmmss")
Push-Location $Root
try {
    $gitHash = & git rev-parse --short HEAD 2>$null
    if ($LASTEXITCODE -eq 0 -and $gitHash) {
        $BuildId = $gitHash
    }
} finally {
    Pop-Location
}
$swTemplate = Get-Content (Join-Path $Web "sw.js") -Raw
$swTemplate.Replace("__BUILD_ID__", $BuildId) | Set-Content (Join-Path $Docs "sw.js") -NoNewline

Copy-Item (Join-Path $Web "icons") (Join-Path $Docs "icons") -Recurse
Copy-Item (Join-Path $Web "favicon.ico") (Join-Path $Docs "favicon.ico")

Write-Host "Web build complete: $Docs (cache: flappy-beer-$BuildId)"
