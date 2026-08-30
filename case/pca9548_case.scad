// Adafruit PCA9548 8-channel STEMMA QT / Qwiic I2C multiplexer (product 5626) case.
// Units: millimetres. Matches pca9548_case.m / draw_pca9548_case.py.
//
// Open-top tray. The board drops in component-side down onto four standoffs and
// is screwed down from above. A 3 mm buffer surrounds the board on every side.
//
// Eight 3 mm x 3 mm cable exits are notched into the TOP of the two long walls,
// from Z = max(z) - 3 up to the rim, on the eight JST SH channel-port centres.
// The two short ends carry mounting tabs instead of cutouts: each tab reaches
// 10 mm out, is 3 mm high in the same max(z) - 3 band, has a 4 mm hole midway,
// and is rounded at its outer end.
//
// Two variants, selected with -D variant="std" or -D variant="tall".
// The cutouts and the tabs are referenced to the rim, so they move up with the
// wall: Z 9-12 in the std variant, Z 21-24 in the tall variant.
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

// JST SH4 channel-port centres, board x, shifted to the centred model
port_x = [-11.43, -3.81, 3.81, 11.43];   // board x = 8.89, 16.51, 24.13, 31.75

// --- case ---
wall        = 3.00;
floor_t     = 2.00;
fit         = 0.50;               // print clearance, board to inner wall
buffer      = 3.00;               // extra cavity buffer around the board, per side
h_std       = 12.00;
h_tall      = 24.00;
cut         = 3.00;               // the 3 mm x 3 mm cable exits
cut_r       = 0.60;               // round on the two lower corners only
stand_h     = 3.00;               // clears the 2.9 mm JST SH shells under the board
stand_od    = 5.00;
stand_hole  = 2.00;               // M2 self-tapping
stand_hole_depth = 4.00;          // 3 mm standoff + 1 mm floor, 1 mm floor remains

// --- end mounting tabs ---
tab_out     = 10.00;              // projection beyond the end wall
tab_h       = 3.00;               // thickness, sitting in the max(z) - 3 band
tab_r       = 5.00;               // rounded outer end, so the tab is 10 mm wide
tab_hole_d  = 4.00;
tab_hole_out = 5.00;              // midway along the 10 mm projection

gap     = fit + buffer;           // 3.50 mm from board edge to inner wall
inner_x = pcb_x + 2 * gap;
inner_y = pcb_y + 2 * gap;
outer_x = inner_x + 2 * wall;
outer_y = inner_y + 2 * wall;
inner_r = pcb_r + gap;
outer_r = inner_r + wall;

case_h  = (variant == "tall") ? h_tall : h_std;

pcb_z0  = floor_t + stand_h;      // board underside
pcb_z1  = pcb_z0 + pcb_t;         // board top face
cut_z1  = case_h;                 // the exits open onto the rim
cut_z0  = case_h - cut;           // max(z) - 3
tab_z0  = case_h - tab_h;         // tabs sit in the same band

module rounded_rect(w, d, r) {
    offset(r = r)
        square([w - 2 * r, d - 2 * r], center = true);
}

// 3 mm x 3 mm cable exit, in (across-wall, Z). Open to the rim, so only the
// two lower corners are rounded. Printed floor-down this needs no bridging.
module cut_profile() {
    translate([0, cut_z0])
        hull() {
            translate([-cut / 2 + cut_r, cut_r]) circle(r = cut_r);
            translate([ cut / 2 - cut_r, cut_r]) circle(r = cut_r);
            translate([-cut / 2, cut_r]) square([cut, cut - cut_r + eps]);
        }
}

// Prism in (X, Z) driven through a long wall along Y.
module cut_through_y() {
    rotate([90, 0, 0])
        linear_extrude(wall + 4, center = true)
            cut_profile();
}

// Mounting tab plan shape, reaching +X from the end wall face at the origin.
module tab_2d() {
    hull() {
        translate([-eps, -tab_r])
            square([tab_out - tab_r + eps, 2 * tab_r]);
        translate([tab_out - tab_r, 0])
            circle(r = tab_r);
    }
}

module tabs() {
    for (sx = [-1, 1])
        translate([sx * outer_x / 2, 0, tab_z0])
            rotate([0, 0, sx > 0 ? 0 : 180])
                linear_extrude(tab_h)
                    tab_2d();
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
            tabs();
        }

        // four cable exits notched into the top of each long wall,
        // on the eight JST SH channel-port centres
        for (sy = [-1, 1])
            for (px = port_x)
                translate([px, sy * (outer_y / 2 - wall / 2), 0])
                    cut_through_y();

        // 4 mm hole midway along each tab
        for (sx = [-1, 1])
            translate([sx * (outer_x / 2 + tab_hole_out), 0, tab_z0 - eps])
                cylinder(d = tab_hole_d, h = tab_h + 2 * eps);

        // M2 pilot holes down through the standoffs into the floor
        for (sx = [-1, 1])
            for (sy = [-1, 1])
                translate([sx * hole_span_x / 2, sy * hole_span_y / 2,
                           pcb_z0 - stand_hole_depth])
                    cylinder(d = stand_hole, h = stand_hole_depth + eps);
    }
}

// Print orientation: floor on the bed at Z=0, open top up. The rim notches need
// no bridging. The two end tabs are cantilevered at the rim and do need support.
case_body();
