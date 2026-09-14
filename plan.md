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
