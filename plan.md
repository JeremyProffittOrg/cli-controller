# README and user how-to video

## Outcome

Create an extensive repository README with part and wiring diagrams, runtime flow diagrams, real settings screenshots, installation and use instructions, configuration and protocol references, troubleshooting, and contributor checks. Create a short, fun, narrated MP4 that teaches a user how to connect, configure, and use the controller.

Non-goals: firmware or application behavior changes, new dependencies in the product, case-design changes, physical sensor validation, scheduled automation, and infrastructure changes.

## Locked decisions (user-confirmed; do not revisit)

- 2026-08-30: Generate an extensive README with part diagrams, flow diagrams, and screenshots.
- 2026-08-30: Also build a fun user how-to video.
- 2026-08-30: Replace the poor local text-to-speech narration and fix the clipped `CLI CONTROLLER` circle badge.
- 2026-08-30: Approved installing `edge-tts` and sending only the narration text to Microsoft's neural speech service.
- 2026-08-30: Feature the printable case STL files near the top of the main README.
- 2026-08-30: Give each case its own README with MATLAB drawing sources and multi-angle images rendered from the STL.

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
- [x] Video quality correction
  - [x] Replace local System.Speech with an explicit Microsoft neural voice and normalized audio.
  - [x] Enlarge the circular badge and keep the full `CLI CONTROLLER` label inside it.
  - [x] Re-render, inspect every scene, and open the corrected video.
  - Done: contact-sheet review shows the full badge, audio is present at a suitable level, and the user can view the corrected MP4.
- [x] README video hero
  - [x] Generate a lightweight animated preview from the corrected MP4.
  - [x] Place the preview and full-video link above the README introduction.
  - [x] Verify GitHub's public rendered README shows the hero and resolves the MP4 link.
  - Done: the public repository page displays the animated hero above the fold and its link returns the full video.
- [x] Printable case documentation
  - [x] Add a prominent top-level case link near the README video hero.
  - [x] Create a case index and one README for the M5Dial shelf case and one for the VL53L4CD sensor case.
  - [x] Retain and document each OpenSCAD, STL, MATLAB, and supporting drawing source.
  - [x] Render consistent front, rear, top, side, and isometric images from both STL meshes.
  - [x] Regenerate the MATLAB-style mechanical drawing images and verify every local case link.
  - Done: both case guides resolve their model, source, drawing, and multi-angle image links, and the main README links the case index above the fold.
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
- 2026-08-30: Replaced System.Speech with `en-US-BrianNeural` at +2% rate and FFmpeg loudness normalization targeting -16 LUFS and -1.5 dB peak.
- 2026-08-30: Enlarged the cyan badge from 110 to 150 pixels and centered the complete two-line `CLI CONTROLLER` label inside it.
- 2026-08-30: Re-rendered all eight scenes. Corrected MP4 is 112.435 seconds and 3,865,384 bytes; measured audio is -16.3 dB mean and -1.5 dB peak.
- 2026-08-30: Contact-sheet and individual Desk/final-frame review confirmed the badge is fully readable. Opened the corrected MP4 in the default Windows player.
- 2026-08-30: PowerShell parser, README local-link check, `git diff --check`, and media probes passed.
- 2026-08-30: Commit `efb2583b062934b17563192b69ffb07e3392d291` replaced the narration, fixed the badge, updated the builder and README, and replaced the MP4.
- 2026-08-30: GitHub Actions run `33335532130` concluded `success`; firmware job `99321638360` and windows-app job `99321638522` passed. The Windows job emitted a non-fatal setup-go cache warning after its tests, build, and artifact upload succeeded.
- 2026-08-30: User requested that the video be embedded and featured as the README hero.
- 2026-08-30: GitHub's Markdown API rendered a native HTML `<video>` block as an empty paragraph, confirming that README video tags are stripped. Selected an animated linked preview plus direct full-video call-to-action.
- 2026-08-30: User added printable-case documentation and requested the case link near the top of the README.
- 2026-08-30: Confirmed two STL models: `m5dial_shelf_case.stl` (10,972 faces) and `vl53l4cd_case.stl` (4,136 faces). Both have OpenSCAD, MATLAB, and Python drawing sources; MATLAB/Octave is not installed, while Matplotlib and trimesh are available.
- 2026-08-30: Added a 14.01-second, 878,602-byte animated hero generated from the corrected MP4 and made the full video and printable cases the two primary calls to action.
- 2026-08-30: Added `case/README.md` plus dedicated M5Dial and VL53L4CD case guides with downloads, dimensions, print guidance, MATLAB source links, drawing sheets, and six-view STL galleries.
- 2026-08-30: Added `case/render_stl_views.py` and generated seven consistent STL views per model. Corrected the print-oriented VL53L4CD mesh by rotating height from Y into documentation Z before rendering.
- 2026-08-30: Regenerated both MATLAB-style drawing sheets from the matching dimensional parameters. Native `.m` sources remain checked in for MATLAB export.
- 2026-08-30: Recursive documentation link checks, Python source checks, Git diff check, GIF probe, and watertight mesh checks passed. GitHub's Markdown API confirmed hero, MP4, and case links render above the first section.
- 2026-08-30: Commit `9e6a77a6ef7f8de611e53c38fdaa5a14c9135a8f` published the README hero and printable case guides. The public page showed the hero and case link, and the MP4 returned HTTP 200.
- 2026-08-30: GitHub Actions run `33337069303` could not start because the public repository reported zero available self-hosted runners. Canceled the impossible queued run and changed the workflow to public-safe `windows-latest` and `ubuntu-latest` GitHub-hosted runners.
