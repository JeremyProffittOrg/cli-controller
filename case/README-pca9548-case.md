# Adafruit PCA9548 I2C multiplexer case

Open-top tray for the [Adafruit PCA9548 8-Channel STEMMA QT / Qwiic I2C
Multiplexer, product 5626](https://www.adafruit.com/product/5626). Two heights.
Nine 3 mm x 3 mm cable exits sit at the same Z in both heights.

| Standard, 12 mm | Tall, 24 mm |
|---|---|
| ![std isometric](pca9548_case_std_iso.png) | ![tall isometric](pca9548_case_tall_iso.png) |
| [Download STL](pca9548_case_std.stl) | [Download STL](pca9548_case_tall.stl) |

![Drawing sheet](pca9548_case_drawing.png)

## Board data

Every board number is read from Adafruit's own Eagle file,
[`adafruit/Adafruit-PCA9548-PCB`](https://github.com/adafruit/Adafruit-PCA9548-PCB)
→ `PCA9548A QT Board.brd` (Eagle 9.6.2). Nothing here is estimated from a photo
or from the marketing dimensions.

| Feature | Value | Source in the .brd |
|---|---|---|
| Outline | 40.64 x 20.32 mm, corner R2.54 | layer 20 wires and 90 deg curves |
| Thickness | 1.60 mm PCB, 4.8 mm with parts | Adafruit product page |
| Mounting holes | 4 x D2.032, span 35.56 x 15.24 mm | two `<hole>` + two `MOUNTINGHOLE_2.5_PLATED` |
| Channel ports | 8 x `JST_SH4`, board x 8.89 / 16.51 / 24.13 / 31.75, y 3.048 and 17.272 | `CONN1 CONN2 CONN5 CONN6 CONN7 CONN8 CONN9 CONN10` |
| Controller port | 1 x `JST_SH4`, board x 3.175, y 10.16 (left end, on centre) | `CONN4` |

The model is centred, so board coordinates are shifted by (-20.32, -10.16).
Port centres become X = -11.43 / -3.81 / +3.81 / +11.43, and the controller
port lands exactly on Y = 0.

## Case

| Parameter | Value |
|---|---|
| Outer | 47.64 x 27.32 mm, corner R6.04 |
| Height | 12.00 mm (`std`) or 24.00 mm (`tall`) |
| Wall / floor | 3.00 / 2.00 mm |
| Board clearance | 0.50 mm per side |
| Standoffs | 4 x D5.00, 3.00 mm off the floor top |
| Screw pilots | D2.00 x 4.00 mm deep, for M2 self-tapping screws |
| Board underside | Z = 5.00 mm |
| Board top face | Z = 6.60 mm |
| Cable exits | 9 x 3.00 x 3.00 mm, corner R0.60, Z 2.00 to 5.00 |

The board goes in **component-side down**. The four standoffs are 3.00 mm tall,
which clears the 2.9 mm JST SH shells, and the shells then sit inside the
cutout mouths. The top edge of every cutout is flush with the board underside
at Z 5.00, so each cable leaves the case at floor level with no bend.

Four cutouts sit in each long wall, one per JST SH channel port, and one sits
in the left end wall on the controller-side port. The right end wall is left
solid: the `JP1` 1x05 breadboard row ships unpopulated and the open top reaches
it, along with the `A0` / `A1` / `A2` address jumpers, `SW1`, and the power LED.

## What changes between the two heights

Only wall above the board. The floor, the standoffs, the screw pilots and all
nine cutouts are at identical Z in both files, which was checked by comparing
cross-section areas:

```
z= 1.0  std area 1257.5691  tall area 1257.5691  delta 1.36e-12
z= 2.5  std area  375.7123  tall area  375.7123  delta 5.68e-14
z= 3.5  std area  375.2417  tall area  375.2417  delta 5.68e-14
z= 4.9  std area  389.7395  tall area  389.7395  delta 1.14e-13
z= 6.0  std area  390.3154  tall area  390.3154  delta 1.82e-12
```

The tall version leaves 17.40 mm of clear space above the board for cable
slack, a stacked board, or a lid of your own.

## Printing

Print floor-down, open top up. No support is needed: the standoffs rise from
the floor and each cutout only bridges 3 mm of wall.

```powershell
& 'C:\Program Files\OpenSCAD\openscad.com' -o case\pca9548_case_std.stl  -D 'variant="std"'  case\pca9548_case.scad
& 'C:\Program Files\OpenSCAD\openscad.com' -o case\pca9548_case_tall.stl -D 'variant="tall"' case\pca9548_case.scad
```

## Regenerate the documentation images

```powershell
python .\case\draw_pca9548_case.py
python .\case\render_stl_views.py pca9548_case_std pca9548_case_tall
```

Run `pca9548_case.m` in MATLAB when a native MATLAB export is required.

## Model checks

Both files load as one triangle mesh and are watertight after normal vertex
processing:

| Model | Triangles | Bounds, mm | Volume, mm3 | Watertight |
|---|---:|---|---:|---|
| `pca9548_case_std.stl` | 7,504 | 47.64 x 27.32 x 12.00 | 6393.99 | Yes |
| `pca9548_case_tall.stl` | 7,504 | 47.64 x 27.32 x 24.00 | 11077.78 | Yes |

Twenty-eight point-containment probes confirmed that all nine cutouts are open
at Z 3.5, that the wall is solid between them and above them, that the right
end wall is solid, and that the four screw pilots are open on axis and closed
1.00 mm above the bed.

Review the slicer preview before printing. Printer calibration, material
shrinkage, cable bend radius, screw-head geometry, and board revisions can
change the final fit.
