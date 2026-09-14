# Unbounded I2C sensor scan and controls

## Outcome

Firmware and the Windows host scan PCA9548 muxes at `0x70`-`0x77` plus the root I2C bus for VL53L4CD and ADXL345 devices, list them in Settings, and let the operator add an unbounded number of controls. Deliver a Sensors-tab screenshot by SES mail.

## Locked decisions (user-confirmed; do not revisit)

- 2026-09-13: Scan on the PCA9548 and off it (root bus).
- 2026-09-13: List discovered sensors and add controls in the UI.
- 2026-09-13: Support an arbitrary number of sensors (muxes `0x70`-`0x77`, all eight channels, plus root).
- 2026-09-13: Build it and send an email with the Sensors configuration screenshot.

## Verified facts

- `C:\dev\cli-controller\firmware\src\main.cpp` firmware 0.5.0 hard-codes mux `0x70`, VL53L4CD on channels 0-3, ADXL345 on channel 4.
- Host `internal/config/config.go` stores four `KneeChannels` and one desk sensor.
- Settings UI has Controller, Display, Knees, Desk tabs in `internal/settings/settings_windows.go`.
- Protocol version 1 JSON lines; host max line 512 bytes.
- Delivery is push to `main` (`deploy.md`). User-owned `case` deletions must not be staged.
- Operator mail: HTML via SES `send-raw-email`, template `C:\Users\Jeremy\.claude\templates\email-status.html`, sender `@jeremy.ninja`.

## Workstreams

### firmware-scan — discover muxes, channels, and root sensors

- [ ] firmware-scan — `pio run -d firmware` exits 0; hello `fw` is `0.6.0`.

### host-config-motion — unbounded SensorControl slice and ID-keyed motion

- [ ] host-config-motion — `go test ./internal/config ./internal/motion ./internal/protocol` exits 0.

### settings-ui — Sensors tab with inventory, scan, add/remove controls

- [ ] settings-ui — Settings builds; Sensors tab lists inventory and bindings.

### verify-ship — test, build, screenshot, mail, push

- [ ] verify-ship — `go test ./...` and Windows GUI build exit 0; SES message id recorded; GitHub Actions terminal success.

## Stop conditions (only these)

- Irreversible flash or deploy authorization withdrawn.
- SES or GitHub credentials missing after retry ceiling.
- Firmware upload fails two automatic attempts and one G0-bootloader attempt.

## Retry policy

- Firmware upload: two automatic attempts, then G0.
- GitHub Actions: three unchanged transient retries; fix deterministic failures before re-push.
- Subagent email: one retry, then send inline.

## Execution log

- 2026-09-13: Read firmware, protocol, config, motion, settings, deploy.md.
- 2026-09-13: Firmware 0.6.0 scans muxes 0x70-0x77, all 8 channels, and the root bus. Host Sensors tab lists inventory and unbounded controls.
- 2026-09-13: `go test ./...` passed. `pio run -d firmware` passed. Flashed COM10. Measured Dial scan `{"v":1,"t":"scan","n":0}` on Port A I2C0 SDA13/SCL15.
- 2026-09-13: Screenshot `docs/images/settings-sensors.png`.
