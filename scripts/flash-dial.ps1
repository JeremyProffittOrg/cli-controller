param(
    [string]$Port = "COM10"
)
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Get-Process -Name "cli-controller" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 400
pio run -d (Join-Path $Root "firmware") -t upload --upload-port $Port
if ($LASTEXITCODE -ne 0) {
    throw "pio upload failed: $LASTEXITCODE"
}
