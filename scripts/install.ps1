$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$Src = Join-Path $Root "cli-controller.exe"
if (-not (Test-Path $Src)) {
    throw "missing $Src - build with go build -ldflags=-H windowsgui -o cli-controller.exe ./cmd/cli-controller"
}
Get-Process -Name "cli-controller" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 400
$DestDir = Join-Path $env:LOCALAPPDATA "Programs\cli-controller"
New-Item -ItemType Directory -Force -Path $DestDir | Out-Null
$Dest = Join-Path $DestDir "cli-controller.exe"
Copy-Item -Force $Src $Dest
$Wsh = New-Object -ComObject WScript.Shell
$StartMenu = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\CLI Dial.lnk"
$Startup = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Startup\CLI Dial.lnk"
foreach ($lnkPath in @($StartMenu, $Startup)) {
    $lnk = $Wsh.CreateShortcut($lnkPath)
    $lnk.TargetPath = $Dest
    $lnk.WorkingDirectory = $DestDir
    $lnk.WindowStyle = 7
    $lnk.Description = "CLI Dial"
    $lnk.Save()
}
Start-Process -FilePath $Dest
Write-Host "installed $Dest"
