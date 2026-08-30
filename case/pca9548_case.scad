// Adafruit PCA9548 8-channel STEMMA QT / Qwiic I2C multiplexer (product 5626) case.
// Units: millimetres. Matches pca9548_case.m / draw_pca9548_case.py.
//
// Open-top tray. Board drops in component-side DOWN onto four standoffs and is
// screwed down from above. Nine 3 mm x 3 mm cable exits sit on the floor:
// four in each long wall on the eight JST SH port centres, one in the left end
// wall on the controller-side STEMMA QT port.
//
// Two variants, selected with -D variant="std" or -D variant="tall".
// The tall variant only adds wall above the board: the floor, the standoffs and
// all nine 3x3 cutouts stay at exactly the same Z.
//
// Board geometry is taken from adafruit/Adafruit-PCA9548-PCB,
// "PCA9548A QT Board.brd" (Eagle 9.6.2), layer 20 outline plus the four
// MOUNTINGHOLE_2.5_PLATED / <hole> entries and the nine JST_SH4 elements.
// Board origin in that file is the lower-left corner; this model is centred,
// so every board coordinate below is shifted by (-20.32, -10.16).

$fn = 96;
eps = 0.2;

variant = "std";                  // "std" | "tall"

// --- Adafruit PCA9548 breakout, measured from the Eagle board file ---
pcb_x        = 40.64;             // layer-20 outline, x 0 .. 40.64
pcb_y        = 20.32;             // layer-20 outline, y 0 .. 20.32
pcb_t        = 1.60;
pcb_r        = 2.54;              // four 90 deg outline corner arcs
pcb_comp_h   = 3.00;              // tallest top-side part: JST SH shell, 2.9 mm
hole_span_x  = 35.56;             // holes at board x = 2.54 and 38.10
hole_span_y  = 15.24;             // holes at board y = 2.54 and 17.78
pcb_hole_d   = 2.032;             // Eagle drill

// JST SH4 port centres, board x, shifted to the centred model
port_x = [-11.43, -3.81, 3.81, 11.43];   // board x = 8.89, 16.51, 24.13, 31.75
// CONN4, the controller-side port, is on the left end wall at board y = 10.16,
// which is the centred model y = 0.

// --- case ---
wall        = 3.00;
floor_t     = 2.00;
fit         = 0.50;               // per side, board to inner wall
h_std       = 12.00;
h_tall      = 24.00;
cut         = 3.00;               // the 3 mm x 3 mm cable exits
cut_r       = 0.60;
stand_h     = 3.00;               // clears the 2.9 mm JST SH shells under the board
stand_od    = 5.00;
stand_hole  = 2.00;               // M2 self-tapping
stand_hole_depth = 4.00;          // 3 mm standoff + 1 mm floor, 1 mm floor remains

inner_x = pcb_x + 2 * fit;
inner_y = pcb_y + 2 * fit;
outer_x = inner_x + 2 * wall;
outer_y = inner_y + 2 * wall;
inner_r = pcb_r + fit;
outer_r = inner_r + wall;

case_h  = (variant == "tall") ? h_tall : h_std;

pcb_z0  = floor_t + stand_h;      // board underside
pcb_z1  = pcb_z0 + pcb_t;         // board top face
cut_z0  = floor_t;                // cable exits sit on the floor
cut_z1  = cut_z0 + cut;           // = pcb_z0, flush with the board underside

module rounded_rect(w, d, r) {
    offset(r = r)
        square([w - 2 * r, d - 2 * r], center = true);
}

// 3 mm x 3 mm cable exit, in (across-wall, Z). Bottom edge on the floor top face.
module cut_profile() {
    translate([0, cut_z0 + cut / 2])
        offset(r = cut_r)
            square(cut - 2 * cut_r, center = true);
}

// Prism in (X, Z) driven through a long wall along Y.
module cut_through_y() {
    rotate([90, 0, 0])
        linear_extrude(wall + 4, center = true)
            cut_profile();
}

// Prism in (Y, Z) driven through an end wall along X.
module cut_through_x() {
    rotate([0, 0, 90])
        rotate([90, 0, 0])
            linear_extrude(wall + 4, center = true)
                cut_profile();
}

module shell() {
    difference() {
        linear_extrude(case_h)
            rounded_rect(outer_x, outer_y, outer_r);
        translate([0, 0, floor_t])
            linear_extrude(case_h - floor_t + eps)
                rounded_rect(inner_x, inner_y, inner_r);
    }
}

module standoffs() {
    for (sx = [-1, 1])
        for (sy = [-1, 1])
            translate([sx * hole_span_x / 2, sy * hole_span_y / 2, floor_t - eps])
                cylinder(d = stand_od, h = stand_h + eps);
}

module case_body() {
    difference() {
        union() {
            shell();
            standoffs();
        }

        // four cable exits in each long wall, on the eight JST SH port centres
        for (sy = [-1, 1])
            for (px = port_x)
                translate([px, sy * (outer_y / 2 - wall / 2), 0])
                    cut_through_y();

        // one cable exit in the left end wall, on the controller-side port
        translate([-(outer_x / 2 - wall / 2), 0, 0])
            cut_through_x();

        // M2 pilot holes down through the standoffs into the floor
        for (sx = [-1, 1])
            for (sy = [-1, 1])
                translate([sx * hole_span_x / 2, sy * hole_span_y / 2,
                           pcb_z0 - stand_hole_depth])
                    cylinder(d = stand_hole, h = stand_hole_depth + eps);
    }
}

// Print orientation: floor on the bed at Z=0, open top up. No support needed.
case_body();
