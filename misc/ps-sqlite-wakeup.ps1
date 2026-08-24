<#
Wake the self-hosted score API with GET /health.
#>

param (
  [Parameter(Mandatory = $false)][string]$ApiUrl = $env:FLAPPY_SCORE_API_URL
)

if (-not $ApiUrl) {
  throw "Pass -ApiUrl or set FLAPPY_SCORE_API_URL"
}

$ApiUrl = $ApiUrl.TrimEnd("/")
Invoke-RestMethod -Uri "$ApiUrl/health" -Method Get
