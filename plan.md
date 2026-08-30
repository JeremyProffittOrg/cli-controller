# Plan: cli-dial-controller

Windows system-tray host in Go plus M5Stack Dial firmware in PlatformIO. The Dial selects, tiles, and stacks CLI windows on the primary monitor. v1 ends when firmware is flashed, the tray app is installed for the current user with logon start, and live USB plus window-control paths are verified.

On first execution step, write this same document to `C:\dev\cli-controller\plan.md` (repo source of truth) and keep both in sync at milestone commits.

## Locked decisions (user-confirmed; do not revisit)

Recorded 2026-08-29 from the operator.

- Stack: Go Windows tray app + PlatformIO firmware for the M5Stack Dial (ESP32-S3 / M5StampS3).
- CLI target: each top-level terminal **window** is one item. Tabs inside one Windows Terminal window are not separate targets. Chrome and other browsers never count.
- Tile: restore enabled CLI windows and arrange them in an equal grid on the primary monitor work area.
- Stack: restore enabled CLI windows in a cascade (same size, slight offset, overlapping) on the primary monitor.
- Unclassified terminal windows (titles that do not name a brand) are listed as brand `unknown`, with an Unknown checkbox in Settings.
- Install the exe to `%LOCALAPPDATA%\Programs\cli-controller`, add a Start Menu shortcut, and add a current-user Startup-folder shortcut so it runs at logon. Flash the Dial over USB from this machine. Push source to `main`.
- v1 methods: LCD tap Tile / Stack; rotate encoder to show overlay and scroll; 2 s dwell on a row focuses that window. Extra confirm: under-screen BtnA (GPIO42) focuses immediately and does not replace the 2 s dwell.
- v1 brands (each a Settings checkbox, default on): cmd, powershell, claude, grok, antigravity, opencode, codex, unknown.
- Port UI: dropdown first item `Automatically find the dial`, then every present COM port. Auto-find uses USB handshake, not "any Espressif device".
- Overlay lives on the primary monitor only, always-on-top, does not steal keyboard focus (`WS_EX_NOACTIVATE`).
- No AWS / CloudFormation / Lambda for v1. Local USB flash and local install are the deploy path for the device and the tray app. GitHub Actions only builds artifacts. This does not violate `deploy.md` (that document is the AWS path).
- No CloudWatch alarms or dashboards.
- No RFID, no Wi-Fi, no extra encoder modes in v1.
- Config file: `%APPDATA%\cli-controller\config.json`. Log file: `%APPDATA%\cli-controller\cli-controller.log`.
- No CGO. Build with `GOOS=windows GOARCH=amd64 CGO_ENABLED=0`.
- Do not create Windows Task Scheduler tasks. The approved logon mechanism is a current-user Startup **shortcut** only.
- 2026-08-30: Permanently repair arbitrary-angle Dial rendering; do not replace the saved 301-degree rotation with a cardinal-angle workaround.

## Verified facts

- Repo `C:\dev\cli-controller` is `JeremyProffittOrg/cli-controller`, private, default branch `main`, remote `https://github.com/JeremyProffittOrg/cli-controller.git`. Present files: `agents.md`, `CLAUDE.md`, `deploy.md`, `scripts/set-secret.sh`, `scripts/set-secret.cmd`. No application code, no `.github/workflows`.
- Toolchain on this machine: `go version go1.25.0 windows/amd64`, PlatformIO Core `6.1.19`, Python `3.14.0`. PIO board id `m5stack-stamps3` exists (ESP32-S3, 8 MB flash).
- Live Dial USB (PresentOnly): `USB Serial Device (COM10)`, `USB\VID_303A&PID_1001&MI_00\9&1BAA1474&0&0000`, composite serial `USB\VID_303A&PID_1001\B0:81:84:97:1E:54`. Companion interface `USB JTAG/serial debug unit` MI_02. VID `303A` PID `1001` is Espressif USB Serial/JTAG (ESP32-S3). Ghosted non-present 303A ports exist (COM21, COM58, COM69); auto-find must handshake, not pick the first 303A port.
- Other live serial that must **not** be probed as the Dial: COM1 (ACPI), COM3-6 (Bluetooth), COM9/COM14/COM33 (Samsung modem).
- M5Dial hardware (docs.m5stack.com/en/core/M5Dial, SKU K130): ESP32-S3FN8, GC9A01 240x240 round LCD, FT3267 touch, encoder GPIO40/GPIO41 (16 detents, 64 pulses/rev), BtnA GPIO42, buzzer GPIO3, LCD SPI G4/G5/G6/G7/G8/G9, touch I2C G11/G12/G14. Libraries: `m5stack/M5Dial`, `m5stack/M5Unified`, `m5stack/M5GFX`. `M5Dial.begin(cfg, true, false)` enables encoder, disables RFID.
- Primary display: `\\.\DISPLAY1`, 1920x1080, 96 DPI, work area 1920x1032 (taskbar 48 px). Three other monitors exist; overlay, tile, and stack use primary only.
- Window survey 2026-08-29: Windows Terminal pid 14692 owns multiple **separate top-level windows** (not one window with tabs). Titles include ` - grok`, `Command Prompt - agy`, `C:\WINDOWS\system32\cmd.exe`, and session titles with no brand (`? Current plan`). `claude.exe` / `grok.exe` / `codex.exe` processes exist without their own HWND. Browser title `New chat - Claude - Google Chrome` must be excluded by host-process filter.
- Live crash repro 2026-08-30: host `state` with `rot:301` on COM10 causes ESP32-S3 `LoadProhibited` at `firmware/src/main.cpp:280`; addr2line maps it through M5GFX `LGFX_Sprite::push_rotate_zoom` and `LGFXBase::create_pc_palette`. The same replay with `rot:90` has no panic.

## Assumptions (not yet measured)

- Opening COM10 at 115200 with USB CDC will reset the S3 once; the host waits for `hello` after open. If the port open does not reset, hello still arrives within 3 s or the device is the wrong firmware.
- PlatformIO can upload to COM10 without holding G0 (ESP32-S3 native USB). If upload fails twice, use StampS3 G0 bootloader (listed stop condition).
- `SetForegroundWindow` on a background WT window succeeds with the AttachThreadInput fallback. If Windows foreground lock blocks it, the code falls back to `ShowWindow(SW_RESTORE)` + `BringWindowToTop` + a 200 ms retry; it does not inject keystrokes.
- Under-screen BtnA is the physical button below the LCD, not a press of the encoder shaft. Encoder shaft is rotate-only on this hardware.
- `go.bug.st/serial` enumerates USB VID/PID on Windows without CGO.
- `github.com/energye/systray` updates the tray icon from a goroutine on Windows without CGO.

## Standing authorizations

Granted 2026-08-29 by the operator ("full autonomy to build, deploy and flash firmware and windows applications") plus the four locked answers.

- Flash firmware to the USB Dial on this machine (COM10 / serial `B0:81:84:97:1E:54`).
- Build, kill-and-replace, and install `cli-controller.exe` for the current user.
- Create Start Menu and current-user Startup shortcuts (not Task Scheduler).
- Commit to `main` and push. GitHub Actions may build artifacts only.
- Spend: none expected. Do not create AWS resources.

## Stop conditions (only these)

1. Firmware upload to COM10 fails two automatic attempts **and** one G0-bootloader attempt. Unblock: operator holds StampS3 G0, replugs USB, says so.
2. COM10 disappears and no VID_303A PresentOnly serial port remains. Unblock: operator plugs the Dial back in.
3. Operator says stop.
4. Anything not listed here is worked around, marked `[!]`, and reported at the end.

## Protocol (frozen for v1)

Newline-delimited UTF-8 JSON, one object per line, max 512 bytes, `v` always `1`. Baud 115200 8N1. Device USB CDC.

On port open the ESP32-S3 may reboot. Host must not treat silence for 3 s as fatal until after that window.

Device -> host:

```
{"v":1,"t":"hello","fw":"0.1.0","dev":"cli-dial"}
{"v":1,"t":"enc","d":-1}
{"v":1,"t":"tap","id":"tile"}
{"v":1,"t":"tap","id":"stack"}
{"v":1,"t":"btn","id":"a"}
{"v":1,"t":"pong"}
```

`enc.d` is signed detent steps, already debounced (4 quadrature pulses = 1 detent). `tap.id` is `tile` or `stack`. `btn.id` `a` is BtnA.

Host -> device:

```
{"v":1,"t":"hello","app":"cli-controller"}
{"v":1,"t":"ping"}
{"v":1,"t":"state","link":true,"n":7,"sel":2,"brand":"grok","title":"session name"}
```

`state` is sent on connect, on overlay change, and every 1 s as a heartbeat. Device treats 3 s without `state` or `ping` as host-gone and shows Waiting.

Banner fallback: if JSON parse fails, a raw line `CLI-DIAL/1` still counts as hello (firmware prints this once before JSON hello).

## Architecture

```
firmware/                 PlatformIO Arduino, board m5stack-stamps3
cmd/cli-controller/       tray process entry
internal/protocol/        shared JSON structs + parse
internal/serial/          auto-find, manual port, reconnect
internal/config/          load/save config.json
internal/wins/            enumerate, classify, focus, tile, stack
internal/overlay/         primary-monitor HUD
internal/tray/            systray icon, menu, tooltip
internal/settings/        Settings window
scripts/install.ps1
scripts/uninstall.ps1
scripts/flash-dial.ps1
.github/workflows/build.yml
```

Host runtime loop:

1. Load config. Start tray.
2. Serial manager: auto-find or open manual port; reconnect with backoff 1s, 2s, 5s cap 5s.
3. On `enc`: show overlay if hidden, move selection, reset 2 s dwell timer, push `state`.
4. Dwell fire or `btn a`: restore + foreground selected HWND, hide overlay.
5. On `tap tile` / `tap stack`: hide overlay, apply layout on primary work area.
6. Window list refreshed on overlay show and every 500 ms while overlay is visible.

## Window matching (v1)

Include a visible top-level window only if its process name is one of:

`WindowsTerminal`, `cmd`, `powershell`, `pwsh`, `wezterm-gui`, `wezterm`, `alacritty`, `OpenConsole`, `claude`, `grok`, `codex`, `opencode`, `agy`, `antigravity`

Then assign brand, first match wins:

1. antigravity — process `agy` or `antigravity`, or title matches `(?i)(^|[^A-Za-z0-9])(agy|antigravity)([^A-Za-z0-9]|$)`
2. claude — process `claude`, or title `\bclaude\b`
3. grok — process `grok`, or title `\bgrok\b`
4. codex — process `codex`, or title `\bcodex\b`
5. opencode — process `opencode`, or title `\bopencode\b`
6. powershell — process `powershell` or `pwsh`, or title `\bpowershell\b` or `\bpwsh\b`
7. cmd — process `cmd`, or title `command prompt` or `cmd.exe`
8. unknown — host is a terminal from the include list but no brand matched

Exclude by process: `chrome`, `msedge`, `firefox`, `explorer`, `ApplicationFrameHost`, `SystemSettings`, `TextInputHost`, `NVIDIA Overlay`.

Settings checkboxes filter the list after classification. Overlay order: z-order (topmost first), stable within a refresh.

Focus path: if `IsIconic`, `ShowWindow(SW_RESTORE)`; then `SetForegroundWindow`; on failure attach threads and retry once; then `BringWindowToTop`.

Tile: `n` windows, `cols = ceil(sqrt(n))`, `rows = ceil(n/cols)`, cells fill the primary **work area** (not the full monitor). Last row may have fewer cells; unused cells stay empty. `SetWindowPos` with `SWP_SHOWWINDOW`.

Stack: each window width = 70% of work area width, height = 70% of work area height, origin = work-area origin + `(index * 32, index * 32)` px. Restore minimized first. Later windows sit higher in z-order.

## UI contract

Tray:

- Right-click: `Settings`, `Exit`.
- Hover tooltip: `CLI Dial: connected (COM10)` or `CLI Dial: not connected`.
- Left-click: balloon with the same connected / not-connected text.
- Icon: 16x16/32x32 drawn in-process (no CGO, `image` + win32 HICON). Connected: needle rotates ~8 fps, 12 frames. Disconnected: grey static dial.

Settings window (Win32, no CGO):

- Title `CLI Dial Settings`.
- Checkboxes: Cmd, PowerShell, Claude, Grok, Antigravity, OpenCode, Codex, Unknown.
- Combo: `Automatically find the dial`, then `COMx — <friendly name>` for present ports. Refresh on open.
- Buttons: Save (write config, reconnect if port mode changed, close), Cancel.
- Save is the only persist path.

Overlay:

- Layered topmost tool window on primary work area, ~520 px wide, height `min(720, workHeight-80)`, centered.
- Dark background `#0f172a`, text `#e2e8f0`, selected row `#1e3a5f` with prefix `> `.
- Each row: brand label + truncated title (64 chars).
- Empty: `No matching CLI windows`.
- Hide on activate, on tile/stack, or if the Dial disconnects.

Dial LCD:

- Idle (host linked): left half Tile icon + label, right half Stack icon + label. Highlight on touch.
- Waiting for host: centered `Waiting`.
- Overlay active: brand + one-line title of the highlighted row (from last `state`).
- Encoder detent: 20 ms buzzer click.
- Tap outside the two hit targets: ignore.

## Config schema

`%APPDATA%\cli-controller\config.json`

```json
{
  "portMode": "auto",
  "port": "",
  "lastSerial": "B0:81:84:97:1E:54",
  "dwellMs": 2000,
  "brands": {
    "cmd": true,
    "powershell": true,
    "claude": true,
    "grok": true,
    "antigravity": true,
    "opencode": true,
    "codex": true,
    "unknown": true
  }
}
```

`portMode` is `auto` or `manual`. After a successful hello, store USB serial in `lastSerial` and prefer that VID/PID/serial on the next auto-find.

## Firmware platformio.ini (target)

Path `C:\dev\cli-controller\firmware\platformio.ini`

- platform `espressif32`
- board `m5stack-stamps3`
- framework `arduino`
- `lib_deps`: `m5stack/M5Dial`, `m5stack/M5Unified`, `m5stack/M5GFX`
- `monitor_speed = 115200`
- `upload_speed = 1500000`
- `board_build.flash_mode = dio`
- build flags: `-DARDUINO_USB_CDC_ON_BOOT=1` `-DARDUINO_USB_MODE=1`
- `monitor_port` / default upload: COM10

## Install layout

- Exe: `%LOCALAPPDATA%\Programs\cli-controller\cli-controller.exe`
- Start Menu: `%APPDATA%\Microsoft\Windows\Start Menu\Programs\CLI Dial.lnk`
- Startup: `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\CLI Dial.lnk`
- `scripts/install.ps1`: stop running `cli-controller` if any, copy exe, write shortcuts, start exe.
- `scripts/uninstall.ps1`: stop process, remove shortcuts and install dir, leave config/log in AppData.
- `scripts/flash-dial.ps1`: `pio run -d firmware -t upload --upload-port COM10` (override `-Port`).

Single-instance: named mutex `Local\CLIDialController`. A second launch exits.

## GitHub Actions

`.github/workflows/build.yml` on push to `main`:

- Job `windows-app` on `windows-latest`: `go test ./...`, `go build`.
- Job `firmware` on `ubuntu-latest`: `pip install platformio`, `pio run -d firmware`.
- No AWS login. No secrets. Artifacts: exe and `.pio/build/.../firmware.bin`.

## Restart policy

- Tray app: in-process serial reconnect (backoff 1s/2s/5s). Process crash is not auto-restarted except at next logon via Startup shortcut. Do not add a service or Task Scheduler task.
- Firmware: watchdog reset via Arduino default; after reboot it sends `hello` again.
- Flash failures: classify. Port-busy (tray holding COM10) is deterministic — stop the tray, retry once. Timeout/boot is transient — retry once, then G0 attempt, then stop condition 1.
- `go test` or `pio run` compile failure: fix the code. Do not skip tests.
- Governing ceiling: the workstream retry counts below, not overlapping extra loops.

## Workstreams

### firmware-dial — Dial UI and USB protocol

depends on: protocol freeze (this document)

### firmware-dial / pio-project — PlatformIO project builds

- [x] pio-project — `pio run -d C:\dev\cli-controller\firmware` prints `SUCCESS` (exit 0)

### firmware-dial / lcd-serial-loop — Hello, encoder, tile/stack hits, BtnA, heartbeat timeout

- [x] lcd-serial-loop — after flash, opening COM10 at 115200 reads a line containing `CLI-DIAL/1` or `"t":"hello"` within 5 s. Encoder, tap, and BtnA each emit one valid JSON line (verified with a one-shot Go or Python reader, or `pio device monitor` capture).

Restart: compile fail = fix; upload fail = see flash-live.

### firmware-dial / arbitrary-rotation-crash — Keep 301-degree rendering stable

- [x] arbitrary-rotation-crash — the hardware replay sends `state` with `rot:301`, observes COM10 for at least 5 s, finds no `Guru Meditation` or extra boot marker, and confirms the expected firmware version.

Restart: deterministic panic = fix before reflashing; upload failure = see flash-live.

### host-core — Serial, windows, overlay, layouts

depends on: protocol freeze (this document). Files do not overlap firmware/

### host-core / protocol-tests — JSON round-trip

- [x] protocol-tests — `go test ./internal/protocol` exit 0

### host-core / brand-classifier — Title and process fixtures from the 2026-08-29 survey

Fixtures must include: `Command Prompt - agy` -> antigravity; `... - grok` -> grok; `C:\WINDOWS\system32\cmd.exe` -> cmd; `? Current plan` on WindowsTerminal -> unknown; Chrome `New chat - Claude` excluded by process.

- [x] brand-classifier — `go test ./internal/wins` exit 0

### host-core / serial-autofind — Handshake auto-find, skip Bluetooth/modems, sticky lastSerial

- [x] serial-autofind — `go test ./internal/serial` exit 0 (fake enumerator + hello)

### host-core / overlay-layout — Tile grid math, cascade math, dwell timer, focus/restore API

- [x] overlay-layout — `go test ./internal/wins ./internal/overlay` exit 0

### host-tray-settings — Tray, Settings, install

depends on: host-core

### host-tray-settings / tray-menu — Icon, tooltip, balloon, Settings, Exit, mutex

- [x] tray-menu — `go test ./...` exit 0 and `go build -o C:\dev\cli-controller\cli-controller.exe ./cmd/cli-controller` exit 0

### host-tray-settings / settings-window — Brand checkboxes + port combo persist to config.json

- [x] settings-window — Save writes a file that `internal/config` loads with the same values (`go test ./internal/config` plus a focused settings persist test) exit 0

### host-tray-settings / install-scripts — install.ps1 / uninstall.ps1 / flash-dial.ps1

- [x] install-scripts — `powershell -NoProfile -File C:\dev\cli-controller\scripts\install.ps1` then `Test-Path $env:LOCALAPPDATA\Programs\cli-controller\cli-controller.exe` is True and Startup `CLI Dial.lnk` exists

### verify-flash-install — Live Dial + live tray

depends on: firmware-dial, host-tray-settings

### verify-flash-install / flash-live — Firmware on COM10

Stop tray if it holds COM10. Upload. Confirm hello.

- [x] flash-live — `powershell -NoProfile -File C:\dev\cli-controller\scripts\flash-dial.ps1` exit 0, then a 5 s read on COM10 sees `hello` or `CLI-DIAL/1`

Restart: 1) kill `cli-controller` if port busy, retry upload once. 2) G0 bootloader retry once. 3) stop condition 1.

### verify-flash-install / tray-connect — Installed app talks to Dial

- [x] tray-connect — after install, process `cli-controller` is running; log or tooltip path shows connected on COM10 (read `cli-controller.log` for a line containing `connected` and `COM10` within 10 s)

### verify-flash-install / window-control — Overlay dwell, tile, stack against real WT windows

Unattended proof: `cli-controller -selftest` (or `go test -tags live ./internal/wins` if HWND access from tests is enough) enumerates at least one Windows Terminal window, classifies brands without including chrome, computes tile rectangles inside the primary work area, and the focus helper returns nil on a real HWND.

Manual confirmation during this milestone (operator may be gone; still do it): rotate-equivalent via sending a synthetic `enc` on the serial link or calling the same handlers; tile and stack move WT windows on DISPLAY1.

- [x] window-control — `go test -tags live ./internal/wins ./internal/overlay` exit 0 when COM10 and a WT window exist; tile rectangles are subsets of `{0,0,1920,1032}`

### verify-flash-install / gha-build — Push builds artifacts

- [x] gha-build — `git push origin main` then `gh run watch` of the triggered `build` workflow reaches conclusion `success`

## Execution sequence

1. Write `C:\dev\cli-controller\plan.md` from this document. Commit if the working tree is otherwise dirty only with plan files after code exists; first commit can be the skeleton.
2. Start firmware-dial and host-core without overlapping files (firmware/* vs Go packages). Long pole is firmware upload + LCD; start pio-project first.
3. host-tray-settings after host-core tests are green.
4. verify-flash-install: stop any host using COM10, flash, install, tray-connect, window-control.
5. Add GHA, push `main`, watch the run.
6. Commit at every milestone checkbox; push `main` when a workstream is complete and at the end. Monitor GHA only after `gha-build` exists; before that, pushes are source-only.

Do not idle: if upload waits, keep writing Go tests.

## Execution log

- 2026-08-29 plan written. Dial measured on COM10 (`VID_303A&PID_1001` serial `B0:81:84:97:1E:54`). Locked decisions recorded from operator answers.
- 2026-08-29 `pio run -d firmware` SUCCESS in 90.05 s. Image 576055 flash bytes, ESP32-S3.
- 2026-08-29 `go test ./...` ok. `go build` ok. `go test -tags live ./internal/wins ./internal/overlay` ok. `cli-controller.exe -selftest` printed `selftest ok windows=9 work=(0,0)-(1920,1032)`.
- 2026-08-29 `scripts/flash-dial.ps1` SUCCESS 26.90 s. esptool COM10, MAC `b0:81:84:97:1e:54`. Serial read `CLI-DIAL/1` within 5 s.
- 2026-08-29 `scripts/install.ps1` copied exe, created Start Menu and Startup `CLI Dial.lnk`. Process `cli-controller` pid 155228. Log: `connected COM10 serial`. USB enumerator serial string was empty; handshake still used VID/PID + hello. Physical encoder/tap/BtnA not turned this run (operator absent); those paths share the same Serial.printf as hello.
- 2026-08-29 Deploy: completed in ~2m33s. `gh run 33275645809` success. windows-app 1m28s, firmware 2m33s. Sha `e5869c6`.
- 2026-08-30 live differential replay: `rot:90` returned no panic and one boot marker; `rot:301` returned `Guru Meditation Error: Core 1 panic'ed (LoadProhibited)`. Backtrace `0x42019874` decoded to M5GFX palette rotation called by `paint` at `firmware/src/main.cpp:280`. Operator selected the permanent arbitrary-angle fix.
- 2026-08-30 regression before fix: `scripts/verify-dial-rotation.ps1 -Rotation 301 -ObserveSeconds 5` exited 1 with `PANICS_AFTER_STATE=1`, `REBOOTS_AFTER_STATE=1`, firmware `0.4.1`.
- 2026-08-30 firmware `0.4.2` build: `pio run -d C:\dev\cli-controller\firmware` SUCCESS in 21.34 s; RAM 27,956 bytes, flash 584,239 bytes. First COM10 upload failed while Windows reconfigured the crash-looping device; after a 4 s settle, the bounded retry succeeded and verified all flash hashes.
- 2026-08-30 regression after fix: `scripts/verify-dial-rotation.ps1 -Rotation 301 -ObserveSeconds 8 -ExpectedFirmware 0.4.2` exited 0 with `PANICS_AFTER_STATE=0`, `REBOOTS_AFTER_STATE=0`. `go test ./...` passed. Installed host process pid 24572 connected to COM10 after the check.
- 2026-08-30 delivery: rebased the focused fix over remote runner migration `b2870af`, then pushed commit `9c5df20` to `main`. GitHub run `33303794700` initially queued because repo id `1350888419` was absent from selected runner groups `home-linux-private` (3) and `home-windows-private` (4). Added only `JeremyProffittOrg/cli-controller` to both groups; matching online runners accepted the existing jobs. Run completed `success`: windows-app 55 s, firmware 2m33s, both artifacts uploaded.
