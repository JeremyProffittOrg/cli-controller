param(
    [string]$OutputPath = "",
    [string]$NeuralVoice = "en-US-BrianNeural",
    [string]$NeuralRate = "+2%"
)

$ErrorActionPreference = "Stop"
Add-Type -AssemblyName System.Drawing

$videoDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$docsDir = Split-Path -Parent $videoDir
$repoRoot = Split-Path -Parent $docsDir
$imageDir = Join-Path $docsDir "images"
if (-not $OutputPath) {
    $OutputPath = Join-Path $docsDir "cli-controller-how-to.mp4"
}
$OutputPath = [IO.Path]::GetFullPath($OutputPath)

foreach ($tool in @("ffmpeg", "ffprobe", "edge-tts")) {
    if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
        throw "$tool is required"
    }
}

$screens = @{
    Controller = Join-Path $imageDir "settings-controller.png"
    Display = Join-Path $imageDir "settings-display.png"
    Knees = Join-Path $imageDir "settings-knees.png"
    Desk = Join-Path $imageDir "settings-desk.png"
}
foreach ($screen in $screens.Values) {
    if (-not (Test-Path -LiteralPath $screen)) {
        throw "missing screenshot $screen"
    }
}

$scenes = @(
    [pscustomobject]@{
        Title = "MEET CLI CONTROLLER"
        Kicker = "A tiny control deck for your command lines"
        Body = "FOCUS  /  TILE  /  STACK`n`nDial it. Tap it. Move it."
        Diagram = "system"
        Screenshot = ""
        Narration = "Meet CLI Controller: the tiny round control deck that can focus, tile, and stack your Windows command-line sessions. Use the Dial, optional knee sensors, or a quick desk gesture."
    },
    [pscustomobject]@{
        Title = "THREE WAYS TO CONTROL"
        Kicker = "Use the input that fits the moment"
        Body = "DIAL`nSpin to browse. Stop or press to activate.`n`nTOUCH`nChoose Tile or Stack, then confirm.`n`nMOTION`nOptional knee and desk gestures."
        Diagram = "flow"
        Screenshot = ""
        Narration = "Spin the Dial to browse, stop to activate, or press for instant selection. On the touch screen, choose Tile or Stack and confirm. Motion controls are optional, so the Dial works before any sensor is connected."
    },
    [pscustomobject]@{
        Title = "WIRE THE OPTIONAL PARTS"
        Kicker = "One bus. Five useful channels. Zero address conflicts."
        Body = "Port A`nGPIO13 SDA / GPIO15 SCL`n`nPCA9548`nAddress 0x70`n`nPower off before rewiring."
        Diagram = "wiring"
        Screenshot = ""
        Narration = "For motion control, connect Port A to a PCA ninety-five forty-eight multiplexer. Distance sensors go on channels zero through three. The desk accelerometer goes on channel four. Disconnect USB power before changing cables."
    },
    [pscustomobject]@{
        Title = "START ON CONTROLLER"
        Kicker = "Pick your CLIs, connection, and timing"
        Body = "1. Enable the CLI families you use.`n`n2. Start with automatic Dial discovery.`n`n3. Choose one activation delay for Dial and knees."
        Diagram = ""
        Screenshot = $screens.Controller
        Narration = "On the Controller tab, choose which command-line families to manage. Automatic connection is the easiest start. The activation delay is shared by the physical Dial and knee gestures."
    },
    [pscustomobject]@{
        Title = "MAKE THE DISPLAY YOURS"
        Kicker = "Classic, graphical, themed, and fully rotatable"
        Body = "Choose a view.`n`nPick a graphical theme.`n`nEnter any rotation from 0 to 359 degrees."
        Diagram = ""
        Screenshot = $screens.Display
        Narration = "On Display, choose the classic list or graphical Dial, pick a theme, and enter any rotation from zero through three hundred fifty-nine degrees. The same value keeps the display and touch map aligned."
    },
    [pscustomobject]@{
        Title = "TEACH YOUR KNEES"
        Kicker = "Assign channels, set thresholds, choose a mode"
        Body = "ARM THEN SELECT`nLeft sequence opens. Right moves. Idle activates.`n`nRIGHT THEN CONFIRM`nRight moves anytime. Left sequence activates."
        Diagram = ""
        Screenshot = $screens.Knees
        Narration = "On Knees, assign each detected channel to Left, Right, or Off. Arm then select uses the left sequence to open the overlay, then right raises move. Right then confirm lets right move at any time and uses the left sequence to activate."
    },
    [pscustomobject]@{
        Title = "GIVE THE DESK A GESTURE"
        Kicker = "Four directions. Three choices each."
        Body = "Match the board orientation.`n`nMap Left, Right, Forward, and Back.`n`nStart at 350 milli-g, then tune."
        Diagram = ""
        Screenshot = $screens.Desk
        Narration = "On Desk, enable the accelerometer, match its mounted orientation, then map each direction to Tile, Stack, or None. Start at three hundred fifty milli-g and raise sensitivity if bumps trigger actions."
    },
    [pscustomobject]@{
        Title = "READY TO ROLL"
        Kicker = "Commit, test, tune, and enjoy the spin"
        Body = "[OK] Commit your settings`n[OK] Test with low-risk windows`n[OK] Tune thresholds after mounting`n[OK] Not detected is safe"
        Diagram = "finish"
        Screenshot = ""
        Narration = "Commit your settings and try low-risk windows while tuning. Not detected is safe; optional devices reconnect automatically. You are ready to give your command lines a spin."
    }
)

$colors = @{
    Background = [Drawing.Color]::FromArgb(7, 12, 22)
    Panel = [Drawing.Color]::FromArgb(15, 23, 42)
    Panel2 = [Drawing.Color]::FromArgb(30, 41, 59)
    Cyan = [Drawing.Color]::FromArgb(34, 211, 238)
    Sky = [Drawing.Color]::FromArgb(56, 189, 248)
    Green = [Drawing.Color]::FromArgb(52, 211, 153)
    Yellow = [Drawing.Color]::FromArgb(250, 204, 21)
    Violet = [Drawing.Color]::FromArgb(167, 139, 250)
    Text = [Drawing.Color]::FromArgb(226, 232, 240)
    Muted = [Drawing.Color]::FromArgb(148, 163, 184)
}

function New-Font([string]$name, [single]$size, [Drawing.FontStyle]$style) {
    try { return New-Object Drawing.Font($name, $size, $style, [Drawing.GraphicsUnit]::Pixel) }
    catch { return New-Object Drawing.Font("Segoe UI", $size, $style, [Drawing.GraphicsUnit]::Pixel) }
}

function Draw-Text([Drawing.Graphics]$graphics, [string]$text, [Drawing.Font]$font, [Drawing.Color]$color, [single]$x, [single]$y, [single]$width, [single]$height, [Drawing.StringAlignment]$alignment = [Drawing.StringAlignment]::Near) {
    $brush = New-Object Drawing.SolidBrush($color)
    $format = New-Object Drawing.StringFormat
    $format.Alignment = $alignment
    $format.LineAlignment = [Drawing.StringAlignment]::Near
    $format.Trimming = [Drawing.StringTrimming]::EllipsisWord
    $graphics.DrawString($text, $font, $brush, (New-Object Drawing.RectangleF($x, $y, $width, $height)), $format)
    $format.Dispose()
    $brush.Dispose()
}

function Draw-Box([Drawing.Graphics]$graphics, [string]$text, [single]$x, [single]$y, [single]$width, [single]$height, [Drawing.Color]$edge, [Drawing.Color]$fill) {
    $fillBrush = New-Object Drawing.SolidBrush($fill)
    $pen = New-Object Drawing.Pen($edge, 4)
    $graphics.FillRectangle($fillBrush, $x, $y, $width, $height)
    $graphics.DrawRectangle($pen, $x, $y, $width, $height)
    $font = New-Font "Bahnschrift" 30 ([Drawing.FontStyle]::Bold)
    $textBrush = New-Object Drawing.SolidBrush($colors.Text)
    $format = New-Object Drawing.StringFormat
    $format.Alignment = [Drawing.StringAlignment]::Center
    $format.LineAlignment = [Drawing.StringAlignment]::Center
    $graphics.DrawString($text, $font, $textBrush, (New-Object Drawing.RectangleF($x, $y, $width, $height)), $format)
    $format.Dispose(); $textBrush.Dispose(); $font.Dispose(); $pen.Dispose(); $fillBrush.Dispose()
}

function Draw-Line([Drawing.Graphics]$graphics, [single]$x1, [single]$y1, [single]$x2, [single]$y2, [Drawing.Color]$color) {
    $pen = New-Object Drawing.Pen($color, 7)
    $graphics.DrawLine($pen, $x1, $y1, $x2, $y2)
    $end = New-Object Drawing.SolidBrush($color)
    $graphics.FillEllipse($end, $x2 - 7, $y2 - 7, 14, 14)
    $end.Dispose(); $pen.Dispose()
}

function Draw-Diagram([Drawing.Graphics]$graphics, [string]$kind) {
    if ($kind -eq "system") {
        Draw-Box $graphics "M5 DIAL" 885 380 260 160 $colors.Cyan $colors.Panel
        Draw-Box $graphics "WINDOWS`nAPP" 1450 380 300 160 $colors.Green $colors.Panel
        Draw-Box $graphics "MOTION`nSENSORS" 1065 710 300 160 $colors.Violet $colors.Panel
        Draw-Line $graphics 1145 460 1450 460 $colors.Cyan
        Draw-Line $graphics 1215 710 1030 540 $colors.Violet
        Draw-Line $graphics 1365 790 1600 540 $colors.Green
        return
    }
    if ($kind -eq "flow") {
        Draw-Box $graphics "DIAL" 780 345 250 130 $colors.Cyan $colors.Panel
        Draw-Box $graphics "TOUCH" 780 535 250 130 $colors.Sky $colors.Panel
        Draw-Box $graphics "MOTION" 780 725 250 130 $colors.Violet $colors.Panel
        Draw-Box $graphics "FOCUS" 1330 345 300 130 $colors.Yellow $colors.Panel
        Draw-Box $graphics "TILE" 1330 535 300 130 $colors.Sky $colors.Panel
        Draw-Box $graphics "STACK" 1330 725 300 130 $colors.Green $colors.Panel
        Draw-Line $graphics 1030 410 1330 410 $colors.Cyan
        Draw-Line $graphics 1030 600 1330 600 $colors.Sky
        Draw-Line $graphics 1030 790 1330 790 $colors.Violet
        return
    }
    if ($kind -eq "wiring") {
        Draw-Box $graphics "M5 DIAL`nPORT A" 700 465 250 160 $colors.Cyan $colors.Panel
        Draw-Box $graphics "PCA9548`n0x70" 1100 465 250 160 $colors.Yellow $colors.Panel
        Draw-Box $graphics "CH 0-3`nVL53L4CD" 1510 320 300 150 $colors.Sky $colors.Panel
        Draw-Box $graphics "CH 4`nADXL345" 1510 650 300 150 $colors.Violet $colors.Panel
        Draw-Line $graphics 950 545 1100 545 $colors.Cyan
        Draw-Line $graphics 1350 515 1510 395 $colors.Sky
        Draw-Line $graphics 1350 575 1510 725 $colors.Violet
        return
    }
    if ($kind -eq "finish") {
        $font = New-Font "Bahnschrift" 50 ([Drawing.FontStyle]::Bold)
        Draw-Text $graphics "SPIN" $font $colors.Cyan 1020 405 280 80 ([Drawing.StringAlignment]::Center)
        Draw-Text $graphics "FOCUS" $font $colors.Yellow 1280 540 300 80 ([Drawing.StringAlignment]::Center)
        Draw-Text $graphics "CREATE" $font $colors.Green 1030 675 300 80 ([Drawing.StringAlignment]::Center)
        $pen = New-Object Drawing.Pen($colors.Cyan, 12)
        $graphics.DrawEllipse($pen, 940, 330, 720, 500)
        $pen.Dispose(); $font.Dispose()
    }
}

function New-Slide($scene, [int]$index, [int]$count, [string]$path) {
    $bitmap = New-Object Drawing.Bitmap(1920, 1080)
    $graphics = [Drawing.Graphics]::FromImage($bitmap)
    $graphics.SmoothingMode = [Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.InterpolationMode = [Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $graphics.TextRenderingHint = [Drawing.Text.TextRenderingHint]::AntiAliasGridFit
    $graphics.Clear($colors.Background)

    $panelBrush = New-Object Drawing.SolidBrush($colors.Panel)
    $graphics.FillRectangle($panelBrush, 0, 0, 1920, 220)
    $panelBrush.Dispose()
    $cyanBrush = New-Object Drawing.SolidBrush($colors.Cyan)
    $graphics.FillRectangle($cyanBrush, 0, 214, 1920, 6)
    $graphics.FillEllipse($cyanBrush, 1685, 35, 150, 150)
    $cyanBrush.Dispose()
    $violetBrush = New-Object Drawing.SolidBrush([Drawing.Color]::FromArgb(70, 167, 139, 250))
    $graphics.FillEllipse($violetBrush, 1780, 760, 320, 320)
    $violetBrush.Dispose()

    $titleFont = New-Font "Bahnschrift" 64 ([Drawing.FontStyle]::Bold)
    $kickerFont = New-Font "Segoe UI" 29 ([Drawing.FontStyle]::Regular)
    $bodyFont = New-Font "Segoe UI" 33 ([Drawing.FontStyle]::Regular)
    $smallFont = New-Font "Bahnschrift" 17 ([Drawing.FontStyle]::Bold)
    Draw-Text $graphics $scene.Title $titleFont $colors.Cyan 90 48 1500 85
    Draw-Text $graphics $scene.Kicker $kickerFont $colors.Muted 94 139 1500 55
    Draw-Text $graphics "CLI`nCONTROLLER" $smallFont $colors.Background 1685 77 150 62 ([Drawing.StringAlignment]::Center)
    Draw-Text $graphics ("SCENE {0:D2}" -f ($index + 1)) $smallFont $colors.Muted 1835 90 75 30 ([Drawing.StringAlignment]::Center)

    $bodyWidth = 700
    if (-not $scene.Screenshot -and -not $scene.Diagram) { $bodyWidth = 1700 }
    Draw-Text $graphics $scene.Body $bodyFont $colors.Text 95 285 $bodyWidth 650

    if ($scene.Screenshot) {
        $image = [Drawing.Image]::FromFile($scene.Screenshot)
        try {
            $maxW = 1010.0; $maxH = 735.0
            $scale = [Math]::Min($maxW / $image.Width, $maxH / $image.Height)
            $drawW = [single]($image.Width * $scale); $drawH = [single]($image.Height * $scale)
            $x = [single](850 + (($maxW - $drawW) / 2)); $y = [single](265 + (($maxH - $drawH) / 2))
            $shadow = New-Object Drawing.SolidBrush([Drawing.Color]::FromArgb(100, 0, 0, 0))
            $graphics.FillRectangle($shadow, $x + 18, $y + 18, $drawW, $drawH)
            $shadow.Dispose()
            $graphics.DrawImage($image, $x, $y, $drawW, $drawH)
            $border = New-Object Drawing.Pen($colors.Cyan, 4)
            $graphics.DrawRectangle($border, $x, $y, $drawW, $drawH)
            $border.Dispose()
        } finally { $image.Dispose() }
    } elseif ($scene.Diagram) {
        Draw-Diagram $graphics $scene.Diagram
    }

    $track = New-Object Drawing.SolidBrush($colors.Panel2)
    $progress = New-Object Drawing.SolidBrush($colors.Green)
    $graphics.FillRectangle($track, 0, 1055, 1920, 25)
    $graphics.FillRectangle($progress, 0, 1055, [single](1920 * (($index + 1) / $count)), 25)
    $progress.Dispose(); $track.Dispose()

    $bitmap.Save($path, [Drawing.Imaging.ImageFormat]::Png)
    $smallFont.Dispose(); $bodyFont.Dispose(); $kickerFont.Dispose(); $titleFont.Dispose()
    $graphics.Dispose(); $bitmap.Dispose()
}

$tempRoot = [IO.Path]::GetFullPath($env:TEMP)
$work = Join-Path $tempRoot ("cli-controller-how-to-video-" + [Guid]::NewGuid().ToString("N"))
$work = [IO.Path]::GetFullPath($work)
if (-not $work.StartsWith($tempRoot, [StringComparison]::OrdinalIgnoreCase)) {
    throw "unsafe temporary path $work"
}
New-Item -ItemType Directory -Path $work | Out-Null

try {
    $segments = [Collections.Generic.List[string]]::new()
    for ($i = 0; $i -lt $scenes.Count; $i++) {
        $slide = Join-Path $work ("slide-{0:D2}.png" -f $i)
        $audio = Join-Path $work ("audio-{0:D2}.mp3" -f $i)
        $segment = Join-Path $work ("segment-{0:D2}.mp4" -f $i)
        New-Slide $scenes[$i] $i $scenes.Count $slide
        & edge-tts --voice $NeuralVoice "--rate=$NeuralRate" --text $scenes[$i].Narration --write-media $audio
        if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $audio)) { throw "neural narration failed for scene $i" }

        $durationText = (& ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 $audio).Trim()
        $audioDuration = [double]::Parse($durationText, [Globalization.CultureInfo]::InvariantCulture)
        $duration = $audioDuration + 0.8
        $fadeOut = [Math]::Max(0.5, $duration - 0.45)
        $audioFade = [Math]::Max(0.5, $audioDuration + 0.35)
        $durationArg = $duration.ToString("0.000", [Globalization.CultureInfo]::InvariantCulture)
        $fadeOutArg = $fadeOut.ToString("0.000", [Globalization.CultureInfo]::InvariantCulture)
        $audioFadeArg = $audioFade.ToString("0.000", [Globalization.CultureInfo]::InvariantCulture)
        $filter = "[0:v]zoompan=z='min(zoom+0.00018,1.025)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':d=1:s=1920x1080:fps=30,fade=t=in:st=0:d=0.35,fade=t=out:st=${fadeOutArg}:d=0.45[v];[1:a]loudnorm=I=-16:TP=-1.5:LRA=11,apad=pad_dur=0.8,afade=t=in:st=0:d=0.15,afade=t=out:st=${audioFadeArg}:d=0.4[a]"
        & ffmpeg -hide_banner -loglevel error -y -loop 1 -i $slide -i $audio -filter_complex $filter -map "[v]" -map "[a]" -t $durationArg -c:v libx264 -preset medium -crf 20 -pix_fmt yuv420p -c:a aac -b:a 160k -ar 48000 -movflags +faststart $segment
        if ($LASTEXITCODE -ne 0) { throw "ffmpeg failed for scene $i" }
        $segments.Add($segment)
        Write-Host ("rendered scene {0}/{1} with $NeuralVoice" -f ($i + 1), $scenes.Count)
    }

    $concat = Join-Path $work "concat.txt"
    $concatLines = $segments | ForEach-Object { "file '$($_.Replace('\', '/'))'" }
    [IO.File]::WriteAllLines($concat, $concatLines, [Text.Encoding]::ASCII)
    $outputDir = Split-Path -Parent $OutputPath
    if (-not (Test-Path -LiteralPath $outputDir)) { New-Item -ItemType Directory -Force -Path $outputDir | Out-Null }
    & ffmpeg -hide_banner -loglevel error -y -f concat -safe 0 -i $concat -c copy -movflags +faststart $OutputPath
    if ($LASTEXITCODE -ne 0) { throw "ffmpeg concat failed" }
    Write-Host "built $OutputPath"
} finally {
    if ((Test-Path -LiteralPath $work) -and $work.StartsWith($tempRoot, [StringComparison]::OrdinalIgnoreCase)) {
        Remove-Item -LiteralPath $work -Recurse -Force
    }
}
