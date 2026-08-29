$ErrorActionPreference = "Stop"
Get-Process -Name "cli-controller" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 400
$DestDir = Join-Path $env:LOCALAPPDATA "Programs\cli-controller"
$StartMenu = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\CLI Dial.lnk"
$Startup = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Startup\CLI Dial.lnk"
foreach ($p in @($StartMenu, $Startup)) {
    if (Test-Path $p) { Remove-Item -Force $p }
}
if (Test-Path $DestDir) { Remove-Item -Recurse -Force $DestDir }
Write-Host "uninstalled"
