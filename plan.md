# Robust sensor readings — firmware 0.7.0 sensor layer review

## Outcome

Every I2C sensor path in `C:\dev\cli-controller\firmware\src\main.cpp` is bounded, concurrent, and self-healing. The Windows host consumes the 20 Hz VL53L4CD and 50 Hz ADXL345 streams with rate-independent filtering. The Dial on `COM10` runs the new firmware, the host build passes its tests, the change is pushed to `main`, and the operator receives an HTML report by SES mail.

## Locked decisions (user-confirmed; do not revisit)

- 2026-09-14: "do an exhaustive review of the esp32 source code specifically around the sensors and validate and update any code which does not provide the most robust and capable sensor readings possible".
- 2026-09-14: "when done, flash the esp32 and then validate the software is leveraging those new readings properly".
- 2026-09-14: "EMAIL ME A FULL REPORT WHEN FINISHED".
- Delivery is a direct push to `main` (`C:\dev\cli-controller\CLAUDE.md`). Flashing `COM10` is authorised by the flash instruction above.

## Verified facts

- `firmware/.pio/libdeps/m5stack-stamps3/STM32duino VL53L4CD/src/platform.cpp` `VL53L4CD_I2CRead` loops `do { ... } while (status != 0)` with no bound: a sensor unplugged mid-read hangs the firmware. `VL53L4CD_StartRanging` blocks up to 1 s with 1 ms delays.
- Firmware 0.6.6 ranged one VL53L4CD at a time (stop/start on every poll) and blocked the main loop up to 80 ms per poll. ADXL345 was read on a 20 ms timer with no DATA_READY gate, range +/-2 g.
- Hardware on `COM10` (measured 2026-09-14): PCA9548 at `0x70` on Port B (`sda 2`, `scl 1`), VL53L4CD on channels 0, 1, 5, 6, ADXL345 `0x53` on channel 3.
- Host protocol: `internal/protocol/protocol.go` `DeviceMsg` ignores unknown JSON fields; `internal/app/app_windows.go` feeds `tof` samples to `motion.Distance` only when `st == 0`.
- `internal/motion/motion.go` used a per-sample baseline EMA of 0.02 (rate dependent) before this run.
- arduino-esp32 3.3.9 `HWCDC::write` blocks up to 20 x `tx_timeout_ms` under host back-pressure; default 100 ms.
- Operator mail: HTML via SES `send-raw-email`, template `C:\Users\Jeremy\.claude\templates\email-status.html`, sender `@jeremy.ninja`, region `us-east-1`, recipient `proffitt.jeremy@gmail.com`.

## Workstreams

### firmware-sensor-layer — bounded, concurrent, self-healing sensor I/O

- [x] tof-register-driver — in-tree VL53L4CD driver replaces the STM32duino library; `pio run -d firmware` exits 0 with no warnings.
- [x] concurrent-ranging — every OK VL53L4CD streams at ~20 Hz at the same time; measured 20.5 Hz x 4 on `COM10`.
- [x] block-read-validation — 15-byte result block decode equals single-register reads on hardware (`xcheck` log line, 2026-09-14).
- [x] accel-data-ready — ADXL345 FULL_RES +/-16 g, 50 Hz, frames gated on `INT_SOURCE.DATA_READY`, format byte re-checked every read; measured 48.7 Hz.
- [x] mux-verify-cache — PCA9548 control register read back after every write; channel cache avoids redundant writes.
- [x] bus-recovery — six consecutive failures on known-good devices trigger a 9-clock SDA release plus STOP and a driver restart.
- [x] stall-watchdog — a ranging sensor with no data for 600 ms is re-initialised; ADXL345 likewise.
- [x] hot-plug-scan — background probe of one bus position every 40 ms; empty bus re-picks the Grove port every 3 s.
- [x] root-address-move — root VL53L4CD moved to `0x2A`; ADXL345 alt address `0x1D` scanned; conflicts fenced and logged.
- [x] usb-tx-bound — `Serial.setTxTimeoutMs(5)` so a stalled host reader cannot freeze the loop.

### host-consumption — host uses the new stream correctly

- [x] protocol-fields — `sg`, `addr`, `msg` parsed; `LiveText` shows sigma; `go test ./internal/protocol` exits 0.
- [x] motion-time-baseline — baseline EMA is wall-clock based (tau 6 s); `go test ./internal/motion` exits 0.
- [x] dial-log-passthrough — firmware `log` lines land in the host log as `dial: ...`.
- [x] host-live-validation — installed host connects to firmware 0.7.0 and the Sensors tab shows live readings for all five devices.

### ship — flash, commit, push, CI, mail

- [x] flash-com10 — `pio run -d firmware -t upload --upload-port COM10` exits 0; hello reports `fw 0.7.0`.
- [x] commit-push — focused commit on `main`; `gh run watch` reaches success.
- [x] operator-report — SES message id recorded in the execution log.

### accel-fifo — ADXL345 FIFO stream so a slow host tick never drops a frame

- [x] accel-fifo — FIFO_CTL stream mode; every pass drains FIFO_STATUS entries; measured rate on `COM10` stays ~50 Hz with zero dropped frames over 15 s.

### scan-selftest — sensor report carries the chip identity

- [x] scan-selftest — `sensor` lines include `chip` (VL53L4CD `ebaa`, ADXL345 `e5`); host shows it in the Sensors tab.

### tof-calibration — offset calibration from Settings, persisted on the Dial

- [x] tof-calibration — host `{"t":"cal","id":..,"mm":N}` runs a 20-sample offset calibration, stores the offset in NVS, re-applies it at init, and answers `{"t":"cal",...}`; Settings gets a "Calibrate at N mm" button; `go test ./...` and `pio run -d firmware` exit 0; flashed and verified on `COM10`.

### adopt-discovered — one click replaces stale default controls with the devices the Dial found

- [x] adopt-discovered — Settings > Sensors "Use discovered" drops controls whose id is not in the inventory and adds one per discovered device; `go test ./...` exits 0; verified on the installed app.

### error-counters — per-sensor transaction errors and re-inits visible in the Sensors tab

- [x] error-counters — `sensor` lines carry `err` and `init`; Sensors tab shows them when non-zero; flashed and verified on `COM10`.

### settings-ui-starvation — Settings dialog "Not Responding" with live sensor stream

- [x] settings-ui-starvation — live readouts are flushed at most every 125 ms and only changed texts are written; the dialog stays responsive (20 s watch, 0 hung samples) with all sensors streaming.
- [x] phantom-slot-guard — firmware 0.7.3 creates a slot only when the part answers with its model id; host logs every new inventory id.

## Stop conditions (only these)

- Flash fails two automatic attempts and one G0-bootloader attempt.
- SES or GitHub credentials missing after one retry.

## Retry policy

- Firmware upload: two automatic attempts, then G0.
- GitHub Actions: three unchanged transient retries; fix deterministic failures before re-push.
- Subagent email: one retry, then send inline.

## Execution log

- 2026-09-14: Read firmware 0.6.6, STM32duino VL53L4CD 1.0.5 sources, host protocol/app/motion/settings, arduino-esp32 3.3.9 HWCDC and Wire.
- 2026-09-14: Rewrote the sensor layer (firmware 0.7.0). `pio run -d firmware` SUCCESS, RAM 9.4 %, Flash 17.7 %.
- 2026-09-14: Flashed `COM10`. 12 s capture: `{"v":1,"t":"i2c","port":"b","sda":2,"scl":1,"n":2}`, `{"v":1,"t":"scan","n":5}`, tof 20.5 Hz x 4 (ch 0, 1, 5, 6), accel 47.4 Hz (ch 3). All tof frames `st 2 sig 0 amb 0 spad 211-212 sg 10`: no optical return at any sensor (faces dark), consistent with the pre-existing "no target" behaviour.
- 2026-09-14: Cross-check build: `xcheck mux:70:0:tof block st=2 spad=212 sig=0 amb=0 sg=10 mm=0 | single st=4 spad=212 sig=0 amb=0 sg=10 mm=0` (raw status 4 maps to 2). Block decode verified; debug build replaced by the final build.
- 2026-09-14: Final flash. 15 s capture: tof 20.5/20.5/20.5/20.4 Hz, accel 48.7 Hz, x -772 y -159 z -744 mg at rest.
- 2026-09-14: Host: `go test ./...` passes after updating the LiveText expectation; `go build -ldflags="-H windowsgui"` OK.
- 2026-09-14: Replayed the 15 s capture through `protocol.ParseDeviceLine` and `motion.Engine`: 1966 lines, 0 parse errors, 0 motion events (all tof `st 2`, accel at rest).
- 2026-09-14: Installed host build to `C:\Users\Jeremy\AppData\Local\Programs\cli-controller\cli-controller.exe`; log `connected COM10`; tray shows `CLI Dial: connected (COM10)`; Sensors tab rows: `mux 0x70 ch 6  VL53L4CD 0x29  mux:70:6:tof  Detected  no target  st 2  sig 0  amb 0  sigma 10`, `mux 0x70 ch 3  ADXL345 0x53  mux:70:3:accel  Detected  x -768  y -163  z -741 mg`, plus ch 0, 1, 5 tof rows.
- 2026-09-14: Commit `3847aa1` pushed to `main`; `gh run watch` started (background task bxu5z8wb1).
- 2026-09-14: GitHub Actions run 34833974108 completed: firmware success, windows-app success.
- 2026-09-14: Operator report sent by SES, MessageId `010001a09f8090ed-3a44575c-84c1-4778-81cc-421f9696259b-000000`.
- 2026-09-14: Operator said "continue": start accel-fifo, scan-selftest, tof-calibration.
- 2026-09-14: Firmware 0.7.1 (`pio run` SUCCESS, RAM 9.8 %, Flash 17.8 %) flashed to `COM10`. 16 s capture: accel 49.7 Hz via FIFO (796 frames, max gap 51 ms, none dropped), tof 20.4-20.6 Hz x 4, `sensor` lines carry `chip ebaa` / `chip e5`. `cal mux:70:0:tof 100` answered `ok:false n:0` after the 5 s timeout (no target in view); `cal nope:tof` answered `ok:false` at once. Host `go test ./...` passes; installed with `scripts/install.ps1`.
- 2026-09-14: Settings > Sensors shows `id ebaa` / `id e5` and the `Calibrate mm` control. Pressed it for `mux:70:1:tof`: host log `calibrate mux:70:1:tof at 100 mm` then `dial cal mux:70:1:tof ok=false offset=0 avg=0 n=0`; dialog shows the failure note. Commit `3bb7798` pushed; CI run 34838690072 in progress.
- 2026-09-14: GitHub Actions run 34838690072 completed: firmware success, windows-app success.
- 2026-09-14: Run status #2 sent by SES, MessageId `010001a09fb4112f-739ea822-5abb-46bf-8c39-444294b472b3-000000`.
- 2026-09-14: Operator said continue again: start adopt-discovered and error-counters.
- 2026-09-14: Firmware 0.7.2 flashed to `COM10` (RAM 10.0 %, Flash 17.8 %). 10 s capture: tof 20.4-20.5 Hz x 4, accel 51.8 Hz, `sensor` lines carry `err 0 init 0`. Host reinstalled, `connected COM10`. In the live Sensors tab, `Use discovered` reported `Controls now match the hardware: 3 added, 3 removed` and the list became ch 0 left, ch 1 right, ch 3 accel, ch 5 off, ch 6 off; aborted without saving so the operator config is unchanged.
- 2026-09-14: GitHub Actions run 34840172347 success. Run status #3 sent by SES, MessageId `010001a09fc3a993-a7b60044-12af-40d7-bae8-d3184ed5bac1-000000`.
- 2026-09-15: Operator reported the Settings dialog "Not Responding" with "9 device(s)". Cause: every sensor frame (~130/s) rewrote every list row (LB_DELETESTRING+LB_INSERTSTRING) and every live label, starving the UI thread. Fix: WM_REFRESH posts coalesced, live readouts flushed from the 125 ms timer, texts written only when changed. Firmware 0.7.3 refuses phantom slots (model id required) and the host logs `sensor new ...` so the 9-device growth is traceable next time. Verified: dialog open on Sensors tab for 20 s, hung samples 0, `5 device(s), 5 connected, 5 streaming`.
