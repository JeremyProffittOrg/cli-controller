# Leg and desk motion controls

## Outcome

Add optional PCA9548-connected VL53L4CD leg sensors and an ADXL345 desk-motion sensor to the M5Dial firmware and Windows application. Flash firmware 0.5.0 to `COM10`, install the application, push `main`, verify GitHub Actions, capture all four settings tabs, and email the screenshots.

Non-goals: case-design changes, scheduled automation, AWS infrastructure changes, and physical sensor verification before the hardware is available.

## Locked decisions (user-confirmed; do not revisit)

- 2026-08-30: Use the Adafruit PCA9548 eight-channel multiplexer.
- 2026-08-30: Provide both knee modes as selectable settings.
- 2026-08-30: Default to `arm` (Arm then select).
- 2026-08-30: In arm mode, a completed left gesture with no right movement activates the current item after the idle delay.
- 2026-08-30: Provide four configurable ADXL345 directions.
- 2026-08-30: Verify the no-hardware path now and defer physical gesture validation to `backlog.md`.
- 2026-08-30: Email one screenshot of every settings tab after deployment.
- 2026-08-30: Deployment, firmware flashing, installation, required restarts, pushing `main`, and bounded deployment retries are authorized.

## Verified facts

- `C:\dev\cli-controller\deploy.md` specifies deployment only by pushing `main` and watching GitHub Actions.
- `main` tracks `origin/main`.
- The worktree has user-owned `case` deletions and untracked case files. They must not be changed or staged.
- The target firmware port is `COM10`.
- The GitHub repository is `JeremyProffittOrg/cli-controller`.
- External M5Dial I2C uses GPIO13 SDA and GPIO15 SCL.
- PCA9548 is `0x70`; VL53L4CD is `0x29`; ADXL345 is `0x53`.

## Interfaces and defaults

- PCA9548 channels 0-3 host optional VL53L4CD sensors. Channel 4 hosts the optional ADXL345.
- Protocol version 1 adds `sensor`, `tof`, and `accel` messages without changing existing messages.
- Knee defaults: mode `arm`, two left raises, right direction `1`, channel 0 Left, channel 1 Right, other channels Off, 75 mm threshold, and 2000 ms activation delay.
- Desk defaults: disabled, 350 milli-g sensitivity, zero-degree orientation, Left=Tile, Right=Stack, Forward=None, Back=None.
- Old configuration files remain valid through normalization.

## Workstreams

- [x] Plan and protocol contract
  - [x] Create and commit `C:\dev\cli-controller\plan.md` before implementation.
  - Done command: `Test-Path C:\dev\cli-controller\plan.md`; `git show --stat HEAD` must contain only `plan.md`.
- [x] Firmware sensor transport
  - [x] Add isolated external I2C, optional sensor discovery, nonblocking samples and recovery, and firmware 0.5.0.
  - Done command: `pio run -d C:\dev\cli-controller\firmware`.
- [x] Host motion engine
  - [x] Add protocol parsing, normalized configuration, knee filtering/latching/modes, and desk high-pass gesture actions.
  - Done command: focused Go tests for config, protocol, gesture, and app packages.
- [x] Four-tab settings dialog
  - [x] Add Controller, Display, Knees, and Desk native tabs with no-hardware status and unchanged Commit/Abort behavior.
  - Done command: config round-trip tests and visual inspection of all four tabs.
- [ ] Verification, installation, and delivery
  - [ ] Run all Go and firmware checks, flash and verify COM10, install and verify the application, capture four screenshots, update `backlog.md`, commit and push only task files, watch GitHub Actions to success, and email the screenshots through SES.
  - Done commands: `go test ./...`; `go build -ldflags="-H windowsgui" -o cli-controller.exe ./cmd/cli-controller`; `pio run -d firmware`; flash and verify scripts; `git diff --check`; `gh run watch <run-id>`; SES send returning a message ID.

## Retry policy

- Firmware upload: three attempts. Re-enumerate COM10 after a transient failure; fix deterministic failures before retry.
- Installation: two attempts after checking and stopping the installed process.
- GitHub Actions: three unchanged retries only for transient runner failures. Fix deterministic failures and push a new commit.
- SES: three bounded attempts with backoff. A delegated send without a message ID fails and requires inline fallback.

## Stop conditions (only these)

- COM10, GitHub authentication, or SES credentials remain unavailable after bounded recovery.
- The work requires an unrelated destructive data change or a material scope expansion.
- A required external service or runner remains unavailable after its retry ceiling.

## Execution log

- 2026-08-30: Read `C:\dev\cli-controller\deploy.md`; confirmed GitHub Actions on `main` is the only deployment path.
- 2026-08-30: `git status --short --branch` confirmed only the listed user-owned `case` changes before task work.
- 2026-08-30: Commit `7f58fe5` added only `plan.md`.
- 2026-08-30: `go test ./internal/config ./internal/protocol ./internal/motion ./internal/settings ./internal/app` passed after host and settings implementation.
- 2026-08-30: `pio run -d firmware` succeeded for firmware 0.5.0; RAM use 8.6%, flash use 17.7%.
