# CLI Controller user how-to video storyboard

Target: a friendly 60-90 second, 1920x1080 narrated guide for a first-time user.

## Scene 1 - Meet CLI Controller

Visual: bright title card with M5Dial, motion sensors, and Windows as three connected stations.

Narration: Meet CLI Controller: the tiny round control deck that can focus, tile, and stack your Windows command-line sessions. Use the Dial, optional knee sensors, or a quick desk gesture.

## Scene 2 - Three ways to control

Visual: encoder, touch, and motion paths converging on Focus, Tile, and Stack.

Narration: Spin the Dial to browse, stop to activate, or press for instant selection. On the touch screen, choose Tile or Stack and confirm. Motion controls are optional, so the Dial works before any sensor is connected.

## Scene 3 - Wire the optional parts

Visual: M5Dial Port A to PCA9548, then channels 0-3 to distance sensors and channel 4 to ADXL345.

Narration: For motion control, connect Port A to a PCA ninety-five forty-eight multiplexer. Distance sensors go on channels zero through three. The desk accelerometer goes on channel four. Disconnect USB power before changing cables.

## Scene 4 - Controller setup

Visual: real Controller settings screenshot with callouts for CLI selection, serial connection, and activation delay.

Narration: On the Controller tab, choose which command-line families to manage. Automatic connection is the easiest start. The activation delay is shared by the physical Dial and knee gestures.

## Scene 5 - Make the display yours

Visual: real Display settings screenshot with themed color accents and rotation callout.

Narration: On Display, choose the classic list or graphical Dial, pick a theme, and enter any rotation from zero through three hundred fifty-nine degrees. The same value keeps the display and touch map aligned.

## Scene 6 - Configure knee gestures

Visual: real Knees screenshot plus two compact mode flows.

Narration: On Knees, assign each detected channel to Left, Right, or Off. Arm then select uses the left sequence to open the overlay, then right raises move. Right then confirm lets right move at any time and uses the left sequence to activate.

## Scene 7 - Configure desk motion

Visual: real Desk screenshot with a four-arrow direction compass.

Narration: On Desk, enable the accelerometer, match its mounted orientation, then map each direction to Tile, Stack, or None. Start at three hundred fifty milli-g and raise sensitivity if bumps trigger actions.

## Scene 8 - Ready to roll

Visual: short checklist and repository video/build paths.

Narration: Commit your settings and try low-risk windows while tuning. Not detected is safe; optional devices reconnect automatically. You are ready to give your command lines a spin.
