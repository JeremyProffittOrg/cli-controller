// M5Stack Dial (SKU K130 / K130-V11) under-shelf case.
// Units: millimetres. Every edge is a 1 mm round.
// Print with the TOP face on the bed (accurate 3 mm screw holes).
// Layer 0.2 mm, 3+ perimeters, 30 % infill. Support the bottom plate.

$fn = 64;

// --- parameters (all mm) ---
wall          = 3.0;
top_t         = 5.0;
depth         = 2 * 25.4;     // 50.8 mm, 2.00 in
usb_clear     = 15.0;
dial_bezel_d  = 51.0;
dial_hole_d   = 45.2;
dial_body_d   = 45.0;
dial_len      = 32.3;
screw_d       = 3.0;
access_d      = 10.0;
cable_d       = 8.0;
ped_t         = 3.0;
ped_hole_d    = 8.2;
ped_splay     = 15;
ped_span      = 14.0;
ped_w         = 34.0;
ped_h         = 18.0;
bottom_clear  = 3.0;
width         = 86.0;
screw_x       = 33.0;
screw_y       = 18.0;
print_slop    = 0.2;
edge_r        = 1.0;          // round on every edge
bottom_chamfer = 30.0;        // 45 deg cut on bottom-left and bottom-right
fn_round      = 24;
eps           = 0.2;

bezel_r   = dial_bezel_d / 2;
center_z  = wall + bottom_clear + bezel_r;
inner_top = center_z + bezel_r + usb_clear;
height    = inner_top + top_t;
inner_w   = width - 2 * wall;
inner_d   = depth - 2 * wall;
inner_h   = inner_top - wall;
cable_z   = inner_top - 3.0 - cable_d / 2;

preview = false;
cutaway = false;

module round_r()
    sphere(r = edge_r, $fn = fn_round);

// Convex round: minkowski of a body shrunk by edge_r.
module outer_shrunk() {
    translate([-width / 2 + edge_r, edge_r, edge_r])
        cube([width - 2 * edge_r, depth - 2 * edge_r, height - 2 * edge_r]);
    // Pedestal overlaps the shrunk back face so the union is one solid.
    translate([-ped_w / 2 + edge_r, depth - edge_r - 0.5, center_z - ped_h / 2 + edge_r])
        cube([ped_w - 2 * edge_r, ped_t + 0.5, ped_h - 2 * edge_r]);
}

module inner_shrunk() {
    translate([-inner_w / 2 + edge_r, wall + edge_r, wall + edge_r])
        cube([inner_w - 2 * edge_r, inner_d - 2 * edge_r, inner_h - 2 * edge_r]);
}

module filleted_cyl(d, h, center = false) {
    minkowski() {
        cylinder(d = max(d - 2 * edge_r, 0.2), h = h, center = center);
        round_r();
    }
}

module dial_hole() {
    translate([0, -2, center_z])
        rotate([-90, 0, 0])
            filleted_cyl(dial_hole_d + print_slop, wall + 4);
}

module screw_top_holes() {
    for (s = [-1, 1])
        translate([s * screw_x, screw_y, inner_top - 2])
            filleted_cyl(screw_d, top_t + 4);
}

module screw_bottom_access() {
    for (s = [-1, 1])
        translate([s * screw_x, screw_y, -2])
            filleted_cyl(access_d, wall + 4);
}

module cable_exit() {
    translate([0, depth + 2, cable_z])
        rotate([90, 0, 0])
            filleted_cyl(cable_d, wall + 4);
}

module pedestal_holes() {
    for (side = [-1, 1]) {
        a = side * ped_splay;
        translate([side * ped_span / 2, depth + ped_t / 2, center_z])
            rotate([0, 0, a])
                rotate([-90, 0, 0])
                    filleted_cyl(ped_hole_d, ped_t + wall + 8, center = true);
    }
}

// 45 deg prism in +X/+Z, extruded along +Y.
module corner_chamfer_prism(c, len) {
    rotate([90, 0, 0])
        translate([0, 0, -len])
            linear_extrude(len)
                polygon([[0, 0], [c, 0], [0, c]]);
}

// Large chamfers on the two bottom edges left and right of the dial opening.
// Minkowski keeps the new edges at the 1 mm round.
module bottom_chamfers() {
    minkowski() {
        union() {
            translate([-width / 2, -2, 0])
                corner_chamfer_prism(bottom_chamfer, depth + 4);
            translate([width / 2, -2, 0])
                mirror([1, 0, 0])
                    corner_chamfer_prism(bottom_chamfer, depth + 4);
        }
        round_r();
    }
}

module case_solid() {
    difference() {
        minkowski() {
            outer_shrunk();
            round_r();
        }
        minkowski() {
            inner_shrunk();
            round_r();
        }
        dial_hole();
        screw_top_holes();
        screw_bottom_access();
        cable_exit();
        pedestal_holes();
        bottom_chamfers();
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
