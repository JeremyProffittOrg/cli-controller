param(
    [string]$Port = "COM10",
    [int]$Rotation = 301,
    [int]$ObserveSeconds = 6,
    [string]$ExpectedFirmware = ""
)

$ErrorActionPreference = "Stop"
$hostProcess = Get-Process -Name "cli-controller" -ErrorAction SilentlyContinue | Select-Object -First 1
$hostExe = if ($hostProcess) { $hostProcess.Path } else { $null }
$serial = $null
$lines = [System.Collections.Generic.List[string]]::new()
$stateSent = $false
$lineCountAtState = 0

try {
    if ($hostProcess) {
        Stop-Process -Id $hostProcess.Id -Force
        Start-Sleep -Milliseconds 600
    }

    $serial = [System.IO.Ports.SerialPort]::new($Port, 115200, "None", 8, "One")
    $serial.ReadTimeout = 150
    $serial.DtrEnable = $true
    $serial.Open()

    $helloDeadline = [DateTime]::UtcNow.AddSeconds(4)
    while ([DateTime]::UtcNow -lt $helloDeadline -and -not $stateSent) {
        try {
            $line = $serial.ReadLine().Trim()
            if (-not $line) {
                continue
            }
            $lines.Add($line)
            if ($line -eq "CLI-DIAL/1") {
                $serial.WriteLine('{"v":1,"t":"hello","app":"rotation-verifier"}')
                $state = '{"v":1,"t":"state","link":true,"n":2,"sel":0,"brand":"codex","title":"rotation-verifier","rot":' + $Rotation + '}'
                $serial.WriteLine($state)
                $stateSent = $true
                $lineCountAtState = $lines.Count
            }
        } catch [System.TimeoutException] {
        }
    }

    if (-not $stateSent) {
        throw "No Dial hello received on $Port within 4 seconds"
    }

    $observeDeadline = [DateTime]::UtcNow.AddSeconds($ObserveSeconds)
    while ([DateTime]::UtcNow -lt $observeDeadline) {
        try {
            $line = $serial.ReadLine().Trim()
            if ($line) {
                $lines.Add($line)
            }
        } catch [System.TimeoutException] {
        }
    }

    $afterState = @($lines | Select-Object -Skip $lineCountAtState)
    $panics = @($afterState | Where-Object { $_ -match "Guru Meditation|panic'ed" })
    $reboots = @($afterState | Where-Object { $_ -eq "CLI-DIAL/1" })
    $firmwareLines = @($lines | Where-Object { $_ -match '"t":"hello"' })

    "ROTATION=$Rotation"
    "STATE_SENT=$stateSent"
    "PANICS_AFTER_STATE=$($panics.Count)"
    "REBOOTS_AFTER_STATE=$($reboots.Count)"
    "FIRMWARE_HELLO=$($firmwareLines | Select-Object -First 1)"

    if ($panics.Count -ne 0 -or $reboots.Count -ne 0) {
        throw "Dial crashed after rendering rotation $Rotation"
    }
    if ($ExpectedFirmware -and -not ($firmwareLines -match ('"fw":"' + [regex]::Escape($ExpectedFirmware) + '"'))) {
        throw "Expected firmware $ExpectedFirmware was not observed"
    }
} finally {
    if ($serial -and $serial.IsOpen) {
        $serial.Close()
    }
    if ($hostExe) {
        Start-Process -FilePath $hostExe
        Start-Sleep -Seconds 4
    }
}
