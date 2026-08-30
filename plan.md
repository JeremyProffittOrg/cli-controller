# README and user how-to video

## Outcome

Create an extensive repository README with part and wiring diagrams, runtime flow diagrams, real settings screenshots, installation and use instructions, configuration and protocol references, troubleshooting, and contributor checks. Create a short, fun, narrated MP4 that teaches a user how to connect, configure, and use the controller.

Non-goals: firmware or application behavior changes, new dependencies in the product, case-design changes, physical sensor validation, scheduled automation, and infrastructure changes.

## Locked decisions (user-confirmed; do not revisit)

- 2026-08-30: Generate an extensive README with part diagrams, flow diagrams, and screenshots.
- 2026-08-30: Also build a fun user how-to video.
- 2026-08-30: Replace the poor local text-to-speech narration and fix the clipped `CLI CONTROLLER` circle badge.
- 2026-08-30: Approved installing `edge-tts` and sending only the narration text to Microsoft's neural speech service.

## Verified facts

- `C:\dev\cli-controller\deploy.md` requires delivery through a push to `main` and a terminal GitHub Actions result.
- The current product is firmware 0.5.0 plus a Windows Go application.
- M5Dial Port A uses GPIO13 SDA and GPIO15 SCL for the optional external I2C bus.
- PCA9548 channels 0-3 support VL53L4CD sensors; channel 4 supports ADXL345.
- The firmware polls VL53L4CD devices at 20 Hz and ADXL345 at 50 Hz.
- Four verified tab screenshots exist in `C:\Users\Jeremy\AppData\Local\Temp\cli-controller-screenshots`.
- FFmpeg 8.1.2 and FFprobe are installed.
- User-owned `case` deletions and untracked case files are present. They must not be changed or staged.

## Workstreams

- [x] Plan
  - [x] Record the documentation and video scope before implementation.
  - Done: `git show --stat HEAD` contains only `plan.md`.
- [x] Documentation assets
  - [x] Copy the four verified settings screenshots into `docs/images`.
  - [x] Add a reproducible video storyboard and build script without changing product dependencies.
  - Done: all local README image links resolve and the video builder exits 0.
- [x] Extensive README
  - [x] Document purpose, features, parts, wiring, setup, controls, knee modes, desk motion, settings tabs, architecture, protocol, configuration, development, troubleshooting, safety, and limitations.
  - [x] Use Mermaid for the wiring and runtime flow diagrams and real images for settings screens.
  - Done: `README.md` has no broken local links and contains all required sections and diagrams.
- [x] User how-to video
  - [x] Build a narrated 1080p H.264/AAC MP4 with title, system overview, wiring, four settings tabs, gesture use, and closing guidance.
  - Done: `ffprobe` reports a playable 1920x1080 H.264 video with AAC audio and nonzero duration.
- [~] Video quality correction
  - [ ] Replace local System.Speech with an explicit Microsoft neural voice and normalized audio.
  - [ ] Enlarge the circular badge and keep the full `CLI CONTROLLER` label inside it.
  - [ ] Re-render, inspect every scene, and open the corrected video.
  - Done: contact-sheet review shows the full badge, audio is present at a suitable level, and the user can view the corrected MP4.
- [x] Verification and delivery
  - [x] Run documentation link checks, `go test ./...`, Windows build, and firmware build.
  - [x] Stage only documentation task files, commit, push `main`, and watch the triggered workflow to success.
  - Done: local checks pass, `git diff --check` passes, and GitHub Actions concludes `success`.

## Retry policy

- Video rendering: fix deterministic script or codec errors before retrying; three render attempts maximum.
- GitHub Actions: retry an unchanged transient runner failure at most three times; fix deterministic failures before pushing again.

## Stop conditions (only these)

- A required codec or media tool remains unavailable after a reversible fallback.
- Delivery requires unrelated destructive work or a material scope expansion.
- GitHub authentication or required runners remain unavailable after the retry ceiling.

## Execution log

- 2026-08-30: Read `deploy.md`, the workflow, firmware, configuration, motion engine, protocol, installation scripts, and preview entry points.
- 2026-08-30: Confirmed FFmpeg 8.1.2, FFprobe, and all four screenshot sources.
- 2026-08-30: Confirmed the pre-existing user-owned `case` work remains outside task scope.
- 2026-08-30: Commit `a4d7d28` recorded only the documentation and video plan.
- 2026-08-30: Copied four verified screenshots into `docs/images` and added `README.md`, `docs/video/storyboard.md`, and the reproducible video builder.
- 2026-08-30: The first video render stopped on an FFmpeg filter timestamp parse error; the builder was fixed before retry.
- 2026-08-30: The second render built `docs/cli-controller-how-to.mp4`: 105.506 seconds, 1920x1080, H.264 at 30 fps, AAC mono at 48 kHz, 3,178,064 bytes.
- 2026-08-30: Visual contact-sheet review covered all eight scenes. Audio measured mean -20.7 dB and peak -2.3 dB.
- 2026-08-30: README link check passed with six Mermaid diagrams and four screenshot links.
- 2026-08-30: `go test ./...`, the Windows GUI build, and `pio run -d firmware` passed.
- 2026-08-30: Commit `41685fb88ce3906d4510dd73de9566dae2c10b2e` added only the documentation, verified screenshots, video source, rendered video, and plan update.
- 2026-08-30: Pushed `main`; GitHub Actions run `33335139734` concluded `success`. Firmware job `99320594914` and windows-app job `99320595056` both passed.
- 2026-08-30: User rejected the local text-to-speech quality and reported that the circular `CLI CONTROLLER` label was clipped.
- 2026-08-30: Installed approved `edge-tts` 7.2.8 for the current user and selected `en-US-BrianNeural` for the correction.
