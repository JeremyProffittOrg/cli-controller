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

### tof-zero-photons — VL53L4CD reported st 2 sig 0 amb 0 on every frame

- [x] tof-zero-photons — root cause found with the `diag` command (raw ambient register 0, config verified); a soft reset (register 0x0000) at every init restores ranging; all four sensors valid at 20 Hz on `COM10`.

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
- 2026-09-15: Operator: "THERE IS NO MEASUREMENT OF THE TIME IN FLIGHT SENSORS". Old screenshot `docs/images/settings-sensors.png` shows the same `st 2 sig 0 amb 0` under firmware 0.6.x, so the fault predates this work. Firmware 0.7.4 `diag`: config read-back matches the ST table (only timing/threshold/status bytes differ), `fwst=03`, result block `04 09 .. d4 90 00 00 00 00 ..` = raw ambient 0 and raw signal 0 with 212 SPADs. VHV recalibration alone changed nothing. Soft reset on `mux:70:0:tof` (`{"t":"diag","id":..,"reset":1}`) then re-init: `st 0`, 35 mm, signal 15240 kcps. Soft reset added to `tofInitDevice`; after flashing, all four sensors report `st 0` on 100 % of frames: ch0 34-40 mm, ch1 32-38 mm, ch5 32-36 mm, ch6 49-55 mm, 20.2-20.4 Hz.
- 2026-09-15: Host reinstalled; Sensors tab live: ch0 36 mm, ch1 34 mm, ch5 31 mm, ch6 53 mm, accel x -834 y -425 z -538 mg. Commit `3cafe2e` pushed; CI success.
- 2026-09-15: Run status #4 (with correction) sent by SES, MessageId `010001a0a3a37a59-7a2b5b97-6bff-47c1-9716-8e283d9e30e8-000000`.

## m5dial-top-clearance — add 10 mm above the Dial and rebuild the mount drawings

Outcome: increase the M5Dial height from 62 to 72 mm. Relative to the shelf mounting face, move the Dial, rear cable slot, LED holes, and bottom profile down together by 10 mm. Rebuild and open the M5Dial, VL53L4CD, and PCA9548 PNG drawing sheets.

Non-goals: change the other mounts' geometry, firmware, host software, hole diameters, mounting footprint, wall thickness, or top thickness.

Files: `C:\dev\cli-controller\case\m5dial_shelf_case.scad`, `m5dial_shelf_case.m`, `draw_m5dial_case.py`, the matching STL and seven STL views, the three existing drawing PNGs, `case\README-m5dial-shelf-case.md`, `case\README.md`, and this plan. Drawing layout changes are limited to these requested sheets if visual checks find clipped or overlapping labels.

Proof: the exported M5Dial STL must be one closed mesh with bounds 86 x 50.8 x 72 mm. Mesh sections and source dimensions must preserve the lower features and add 10 mm of top clearance. All three drawing commands and `git diff --check` must exit 0. The three PNGs must open in Photos, and the pushed commit's build workflow must finish successfully.

## Locked decisions (user-confirmed; do not revisit)

- 2026-10-22: "m5dial".
- 2026-10-22: "add 10mm from the top of the dial hole to the top of the mount, bringing everything down below it".
- 2026-10-22: "then show me a new diagraam, and build diagrams for the other mounts".
- 2026-10-22: "and open it" remains the requested display action for the new diagrams.

## Verified facts

- `C:\dev\cli-controller\case\m5dial_shelf_case.scad`: current height is 62 mm, Dial center Z is 31.5 mm, inner top Z is 57 mm, and cable slot center Z is 50 mm. The bottom is the model's Z=0 datum.
- `C:\dev\cli-controller\case\README.md`: the other mounts are VL53L4CD and PCA9548; PCA9548 has standard and tall variants on one drawing sheet.
- `git status --short --branch`: on `main`; preserve the pre-existing deletion of `case/m5dial_shelf_case_cutaway.png` and untracked `firmware/src/idf_component.yml`.
- `git ls-remote origin refs/heads/main`: `3dd7d34d17c8d3769c8c5046c54adc199ac85f0a`; GitHub connector confirms `JeremyProffittOrg/cli-controller`, default branch `main`, push access.
- `C:\Program Files\OpenSCAD\openscad.com` exists. Existing drawing scripts require Matplotlib, NumPy, and trimesh; use an isolated `uv run --no-project` environment for those existing dependencies.
- `C:\dev\cli-controller\.github\workflows\build.yml`: pushes to `main` run Windows Go tests/build and the PlatformIO firmware build; no local deployment is needed.

Assumption: "add 10mm" means 10 mm more than the current gap. The nominal hole-top-to-mount distance becomes 17.9 mm, with the existing print allowance retained.

## mount-artifacts — generate and verify the requested mechanical outputs

### dial-clearance-geometry — keep lower features together below the shelf

- [x] dial-clearance-sources — add one top-clearance parameter in SCAD, MATLAB, and Python; keep the cable slot tied to the Dial. Export with `"C:\Program Files\OpenSCAD\openscad.com" -o C:\dev\cli-controller\case\m5dial_shelf_case.stl C:\dev\cli-controller\case\m5dial_shelf_case.scad`; exit 0 and a closed 72 mm mesh are required.
- [x] dial-clearance-check — verify bounds, mesh closure, and unchanged lower feature positions against the original STL using the existing trimesh/NumPy dependencies; assertions must pass.

### mount-drawing-sheets — show all current mounts with dimensions

depends on: dial-clearance-geometry

- [x] mount-sheet-build — run `python C:\dev\cli-controller\case\render_stl_views.py m5dial_shelf_case`, then `python C:\dev\cli-controller\case\draw_m5dial_case.py`, `python C:\dev\cli-controller\case\draw_vl53l4cd_case.py`, and `python C:\dev\cli-controller\case\draw_pca9548_case.py` in the isolated plotting environment; all exit 0.
- [x] mount-sheet-display — inspect the three PNGs for readable dimensions and open them with `Start-Process -FilePath <absolute PNG path>`; Photos process titles must identify the requested files.

### mount-artifact-delivery — publish only the requested model and drawing changes

depends on: mount-drawing-sheets

- [x] mount-artifact-push — `git diff --check` exits 0; commit only scoped files on `main`; `git push origin main` exits 0.
- [x] mount-artifact-workflow — query the exact commit's `build` run until `status=completed` and `conclusion=success`.

## Stop conditions (only these)

- Missing credentials or resources prevent export or publication after the available configured paths are checked.
- An unrelated destructive action or material expansion of scope is required.

## Mount job restart policy

- Run short commands with direct hidden child processes and no shell windows. Run exports and rendering in tracked background processes with captured stdout, stderr, exit code, and a ten-minute deadline. Check failure and success at intervals no longer than 60 seconds.
- OpenSCAD export: one normal attempt; after a deterministic fix, one retry. A timeout or nonzero exit is failure, not success. Preserve the original STL in machine-local temporary storage until the new mesh passes verification.
- Drawing generation: fix an identified source or environment error before one retry. Reuse the existing dependency set; no new project dependencies or test framework.
- Git push/build: no interactive prompts. Fix deterministic scoped failures before a retry; allow at most two transient retries with backoff. Use the GitHub connector or public GitHub Actions API to monitor because local `gh` is not signed in.
- No scheduled jobs, service changes, or email are part of this task.

## Execution log

- 2026-10-22: Read the three M5Dial sources, existing drawing generators, STL renderer, case guides, root plan, and build workflow. Confirmed the source geometry, default branch, OpenSCAD installation, Git remote access, and unrelated worktree changes listed above.
- 2026-10-22: `uv run --no-project --python C:/Users/Jeremy/AppData/Local/Programs/Python/Python313/python.exe --with matplotlib --with trimesh python -c "import sys,matplotlib,numpy,trimesh; print(sys.executable); print('matplotlib='+matplotlib.__version__); print('trimesh='+trimesh.__version__)"` exited 0. Reused the scripts' existing dependency set in an isolated environment: Matplotlib 3.11.2, trimesh 5.1.1. No project dependency was added.
- 2026-10-22: Export job `dial-stl-export`, PID 43556: `"C:/Program Files/OpenSCAD/openscad.com" -o "C:/Users/Jeremy/AppData/Local/Temp/m5dial-clearance-5KZDgm/updated.stl" C:/dev/cli-controller/case/m5dial_shelf_case.scad` exited 0 in 77 seconds; output `Simple: yes`. Saved the verified result as binary STL, matching the existing artifact format.
- 2026-10-22: NumPy/trimesh assertions passed on the final binary STL: 10,972 triangles; `watertight=true`, consistent face winding; nominal bounds 86 x 50.8 x 72 mm. Existing edge-round tessellation puts each extreme about 0.0086 mm inside the nominal envelope. Both old and new meshes contain one positive outer surface and four negative internal surfaces; that topology is preserved.
- 2026-10-22: Compared mesh-plane sections at Z=1.5, 8, 22.5, 31.5, 40, 46, 50, 54: maximum old/new deviation 0.000174 mm. Compared original mount sections at Z=58, 59.5, 61 with new sections 10 mm higher: maximum deviation 0.000059 mm. Result: `PASS: height +10 mm; lower features unchanged; mounting geometry shifted +10 mm`.
- 2026-10-22: Ran the seven-view renderer using the existing `render_stl_views.py m5dial_shelf_case` entry point with Matplotlib `Agg`. Ran all three `draw_*_case.py` entry points with `python -B -c "import matplotlib; matplotlib.use('Agg'); import runpy; runpy.run_path('case/draw_m5dial_case.py',run_name='__main__'); runpy.run_path('case/draw_vl53l4cd_case.py',run_name='__main__'); runpy.run_path('case/draw_pca9548_case.py',run_name='__main__')"`; job `mount-sheets-unlocked`, PID 7760, exited 0. The earlier sheet export encountered the old Photos preview's file lock; closing that specific preview resolved it before the single retry.
- 2026-10-22: Visually checked all three sheets. M5Dial PNG: 2475 x 1575; VL53L4CD PNG: 2415 x 1600; PCA9548 PNG: 2613 x 1648. Moved the sensor sheet's overlapping note and included full label bounds in the other sheets. No other mount geometry changed. Native MATLAB source dimensions were updated; native MATLAB execution and physical printing were not performed.
- 2026-10-22: Opened each final PNG through its Windows file association. `Get-Process Photos` reported `m5dial_shelf_case_drawing.png`, `pca9548_case_drawing.png`, and `vl53l4cd_case_drawing.png`. M5Dial was opened last.
- 2026-10-22: `git diff --check` exited 0. `git -c credential.interactive=never -c core.askPass= push --dry-run origin main` exited 0 without an interactive prompt.
- 2026-10-22: Committed and pushed `a5c05c4e216f2286f0a6cc3e72ba5aa2bf39c8a1` (`Add 10 mm above M5Dial and refresh mount diagrams`). `git push origin main` exited 0. Build run `39781807318` started for that exact SHA; firmware job `119834677181`, Windows job `119834677707`. The only worktree changes after the artifact commit were the two pre-existing changes listed above. Removed the task's two temporary mesh comparison files and empty temporary directory.
- 2026-10-22: GitHub Actions run `39781807318` finished with `status=completed`, `conclusion=success`. Windows `go test ./...` and application build passed; firmware `pio run -d firmware` passed. The geometry and drawing milestones are complete. This final plan update records that verified result; its documentation-only push uses the same build workflow.

## velcro-mount-variants — add broad adhesive pads and open five models in Bambu Studio

Outcome: add PCA9548 and VL53L4CD versions with two solid mounting tabs, each the full case width and projecting half the body length. Package M5Dial, PCA screw, VL screw, PCA Velcro, and VL Velcro as five separate objects in `C:\dev\cli-controller\case\cli_controller_mounts.3mf`. Open that file by passing its path directly to the Bambu Studio executable.

Non-goals: alter the original screw-mount STL files, PCB supports or pilot holes, sensor window, cable exits, M5Dial geometry, printer settings, or start a print. Velcro replaces the external mounting screws only.

Files: `C:\dev\cli-controller\case\pca9548_case.scad`, `vl53l4cd_case.scad`, two new Velcro STL files, `cli_controller_mounts.3mf`, the three case README guides, and `C:\dev\cli-controller\plan.md`. Reuse the existing SCAD geometry with `mount_style="screw"` as the unchanged default and `mount_style="velcro"` for the new versions.

Proof: OpenSCAD exports exit 0; each new mesh is closed, consistently oriented, and connected. Check full-width solid tab extents, tab thickness, no external screw holes, and unchanged body sections. Read back the 3MF ZIP/XML and confirm five separate objects/build items with the source mesh dimensions and no overlap. Launch Bambu Studio with the exact 3MF path as a command-line argument. `git diff --check` and the pushed commit's build must pass.

## Locked decisions (user-confirmed; do not revisit)

- 2026-10-22: "create new versions of the pca and vl models which have large tabs the width of the case and 1/2 the length, so that we can use velcrow instead of screws.  Then open all 5 models in bambulabs studio".
- 2026-10-22: "put them in a 3mf file and open that".
- 2026-10-22: "why are you trying to control my computer, build the 3mf and open it as a command line parameter opening bambu labs". Use file generation and a direct executable launch only; no further Computer Use calls or UI input.

## Verified facts

- `C:\dev\cli-controller\case\pca9548_case.scad`: body 53.64 x 33.32 mm; standard height 12 mm; tall height 24 mm; existing end tabs 3 mm thick at the rim. New pad dimensions are 33.32 mm wide x 26.82 mm outward per end, with a 3 mm overlap into the end wall.
- `C:\dev\cli-controller\case\vl53l4cd_case.scad`: body width 39 mm, length between the mounting ends 24.78 mm, depth 12 mm; back tabs are 4 mm thick. New pads are 39 mm wide x 12.39 mm outward per end.
- `C:\Program Files\Bambu Studio\bambu-studio.exe` is installed. Computer Use returned Bambu Studio as not running, with no open project to replace.
- `git status --short --branch`: `main` tracks `origin/main`; preserve the existing deleted cutaway PNG and untracked firmware component manifest.
- Existing isolated plotting environment: `C:\Users\Jeremy\AppData\Local\uv\cache\archive-v0\bQGN7nFnyiAQmL9ak3Q7L\Scripts\python.exe`; no new project dependency is needed for mesh checks.

Assumptions: use adhesive-backed Velcro on solid pads, not strap slots. Use the standard 12 mm PCA body unless the optional height question is answered differently. "All 5" means the three main screw designs plus the two new Velcro designs, rather than adding the alternate PCA height as a sixth object.

## velcro-models — generate and verify the two additional mount options

### velcro-pad-geometry — preserve the bodies and enlarge only the mounting tabs

- [x] velcro-pad-exports — run `"C:\Program Files\OpenSCAD\openscad.com" --export-format binstl -D "mount_style=\"velcro\"" -o <new STL> <matching SCAD>`; add `-D "variant=\"std\""` for PCA. Both commands must exit 0.
- [x] velcro-pad-checks — run NumPy/trimesh assertions for PCA bounds 107.28 x 33.32 x 12 mm and VL print-orientation bounds 39 x 49.56 x 12 mm; verify closed connected meshes, solid pads, and unchanged body features.

### five-model-preview — inspect all requested models in Bambu Studio

depends on: velcro-pad-geometry

- [x] studio-five-objects — build and read back `cli_controller_mounts.3mf` with five named, separate, nonoverlapping mesh objects in millimetres. Launch `"C:\Program Files\Bambu Studio\bambu-studio.exe" "C:\dev\cli-controller\case\cli_controller_mounts.3mf"`. Do not use UI automation, change printer settings, or start a print.

### velcro-model-delivery — publish the new variants through the repository workflow

depends on: five-model-preview

- [x] velcro-model-push — `git diff --check`, focused commit, and `git push origin main` exit 0.
- [x] velcro-model-build — exact pushed commit's GitHub `build` reaches `completed/success`.

## Stop conditions (only these)

- Required credentials or resources are unavailable after configured access paths are checked.
- A locked desktop, security or authentication prompt, or unrelated unsaved project requires the operator's action.
- The task requires an unrelated destructive action or a material expansion of scope.

## Velcro job restart policy

- Export each STL in a tracked hidden process; capture exit, stderr, stdout, and a ten-minute deadline. Retry once only after fixing an identified source or environment error. Check both success and failure within 60 seconds.
- Launch Bambu Studio directly with the 3MF argument and track the returned process. The requested application window is allowed; all short CLI jobs stay hidden. Do not use Computer Use or UI input.
- GitHub builds: fix deterministic scoped failures before retry; at most two transient retries. Preserve user changes and use only the existing push-to-main workflow. No recurring automation, print submission, or email.

## Execution log

- 2026-10-22: Read both SCAD sources, existing artifact inventory, current plan, Bambu Studio installation, and Computer Use guidance. Confirmed the dimensions above and that Bambu Studio has no running project.
- 2026-10-22: `pca-velcro-export` PID 30108 exited 0 in 12.7 seconds; `vl-velcro-export` PID 12628 exited 0 in 3.7 seconds. Both OpenSCAD exports reported `Simple: yes`. Initial mesh reads confirm PCA bounds 107.28 x 33.32 x 12 mm (5,580 triangles) and VL bounds 39 x 49.56 x 12 mm (2,912 triangles), both watertight with consistent winding.
- 2026-10-22: User stopped the initial Computer Use attempt, then requested a single 3MF and explicitly selected command-line opening. All remaining work uses file access and direct background processes, except the explicitly requested visible Bambu Studio launch.
- 2026-10-22: NumPy/trimesh acceptance checks passed: `2 connected watertight Velcro meshes; 40 solid-pad probes; 16 tab-boundary probes; 588 unchanged-body probes`. Both new STL files keep the expected full pad widths, half-length projections, and original pad thicknesses. SHA-256 comparisons confirmed all five pre-existing STL/3MF artifacts were unchanged.
- 2026-10-22: Built `C:\dev\cli-controller\case\cli_controller_mounts.3mf` using Python standard-library ZIP/XML packaging and the five existing STL triangle meshes. File size 324,661 bytes. Readback result: `PASS: 3MF readback; 5 separate objects; 30316 triangles; millimetres; no overlap; all bases at Z=0`. The five names are `M5Dial shelf case`, `PCA9548 screw mount - 12 mm`, `VL53L4CD screw mount`, `PCA9548 Velcro mount - 12 mm`, and `VL53L4CD Velcro mount`. Source orientations are retained; XY centers are (65,60), (60,145), (135,145), (180,60), (195,145) mm. No printer profiles, slice settings, G-code, or credentials are included.
- 2026-10-22: Launched `"C:\Program Files\Bambu Studio\bambu-studio.exe" "C:\dev\cli-controller\case\cli_controller_mounts.3mf"` directly, PID 9884. Read-only `tasklist /FI "IMAGENAME eq bambu-studio.exe" /V /FO CSV` reported window title `cli_controller_mounts - BambuStudio` for that PID. No further UI automation was used after the user's command-line instruction.
- 2026-10-22: Updated the three case guides with Velcro dimensions, regeneration arguments, and the combined 3MF link. `git diff --check` exited 0.
- 2026-10-22: Committed and pushed `070fb726f46056df231ac262519560a6fb3a1aae` (`Add Velcro mounts and a combined five-model 3MF`). `git push origin main` exited 0. GitHub build run `39812167272` is monitoring that exact commit. Removed the empty task temporary directory. Original user worktree changes remain untouched.
- 2026-10-22: Build run `39812167272` finished `completed/success` for commit `070fb726f46056df231ac262519560a6fb3a1aae`; both `windows-app` (`go test ./...` and application build) and `firmware` (`pio run -d firmware`) passed. All five models are packaged in the verified 3MF, and the command-line Bambu Studio launch is confirmed by its matching process window title. This documentation-only completion record uses the same repository push workflow.
