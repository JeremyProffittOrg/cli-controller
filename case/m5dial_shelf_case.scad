// M5Stack Dial (SKU K130 / K130-V11) under-shelf case.
// Units: millimetres.
// Print with the TOP face on the bed (accurate 3 mm screw holes).
// Layer 0.2 mm, 3+ perimeters, 30 % infill. Support the bottom plate.

$fn = 64;

// --- parameters (all mm) ---
wall          = 3.0;
top_t         = 5.0;          // top plate, screw engagement
depth         = 2 * 25.4;     // 50.8 mm, 2.00 in minimum
usb_clear     = 15.0;         // air above the dial for a USB-C plug
dial_bezel_d  = 51.0;         // rotating ring, sits proud of the front
dial_hole_d   = 45.2;         // official panel hole 45 + 0.2
dial_body_d   = 45.0;
dial_len      = 32.3;
screw_d       = 3.0;          // top through-holes into the shelf
access_d      = 10.0;         // screwdriver holes in the bottom
cable_d       = 8.0;          // USB cable exit, back wall, near the top
ped_t         = 3.0;          // pedestal proud of the back
ped_hole_d    = 8.2;
ped_splay     = 15;           // deg each side of the rear axis; 30 deg included
ped_span      = 14.0;         // hole centre distance at the back wall
ped_w         = 34.0;
ped_h         = 18.0;
bottom_clear  = 3.0;          // air under the bezel, inside
width         = 86.0;
screw_x       = 33.0;         // left/right of the dial
screw_y       = 18.0;         // from the outer front, along depth
print_slop    = 0.2;          // added to through-holes that must pass a part
eps           = 0.2;

bezel_r  = dial_bezel_d / 2;
center_z = wall + bottom_clear + bezel_r;           // 31.5
inner_top = center_z + bezel_r + usb_clear;         // 72
height    = inner_top + top_t;                      // 77
inner_w   = width - 2 * wall;
inner_d   = depth - 2 * wall;
inner_h   = inner_top - wall;                       // cavity height
cable_z   = inner_top - 3.0 - cable_d / 2;          // 3 mm meat under the top plate

preview = false;   // true: orange dial ghost (do not export STL this way)
cutaway = false;   // true: drop +X half so the interior is visible

module outer_box() {
    translate([-width / 2, 0, 0])
        cube([width, depth, height]);
}

module inner_cavity() {
    translate([-inner_w / 2, wall, wall])
        cube([inner_w, inner_d, inner_h]);
}

module dial_hole() {
    translate([0, -eps, center_z])
        rotate([-90, 0, 0])
            cylinder(d = dial_hole_d + print_slop, h = wall + 2 * eps);
}

module screw_top_holes() {
    for (s = [-1, 1])
        translate([s * screw_x, screw_y, inner_top - eps])
            cylinder(d = screw_d, h = top_t + 2 * eps);
}

module screw_bottom_access() {
    for (s = [-1, 1])
        translate([s * screw_x, screw_y, -eps])
            cylinder(d = access_d, h = wall + 2 * eps);
}

module cable_exit() {
    translate([0, depth + eps, cable_z])
        rotate([90, 0, 0])
            cylinder(d = cable_d, h = wall + 2 * eps);
}

module pedestal() {
    translate([-ped_w / 2, depth - eps, center_z - ped_h / 2])
        cube([ped_w, ped_t + eps, ped_h]);
}

module pedestal_holes() {
    // Two 8.2 mm holes on the back pedestal.
    // Axes lie in the horizontal plane, ±15 deg from the rear (+Y) axis,
    // so they face away from each other with 30 deg included.
    for (side = [-1, 1]) {
        a = side * ped_splay;
        translate([side * ped_span / 2, depth + ped_t / 2, center_z])
            rotate([0, 0, a])
                rotate([-90, 0, 0])
                    cylinder(d = ped_hole_d, h = ped_t + wall + 8, center = true);
    }
}

module case_solid() {
    difference() {
        union() {
            outer_box();
            pedestal();
        }
        inner_cavity();
        dial_hole();
        screw_top_holes();
        screw_bottom_access();
        cable_exit();
        pedestal_holes();
    }
}

module dial_ghost() {
    color([1.0, 0.45, 0.1, 0.45])
        translate([0, -0.5, center_z])
            rotate([-90, 0, 0]) {
                cylinder(d = dial_bezel_d, h = 3.5);
                translate([0, 0, 3.5])
                    cylinder(d = dial_body_d, h = dial_len - 3.5);
            }
}

module cutaway_half() {
    translate([0, -5, -5])
        cube([width, depth + ped_t + 10, height + 10]);
}

if (cutaway) {
    difference() {
        case_solid();
        cutaway_half();
    }
} else {
    case_solid();
}

if (preview)
    dial_ghost();
