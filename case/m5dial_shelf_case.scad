// M5Stack Dial (SKU K130 / K130-V11) under-shelf case.
// Units: millimetres. Every edge is a 1 mm round.
// Print with the TOP face on the bed (accurate M4 screw holes).
// Layer 0.2 mm, 3+ perimeters, 30 % infill. Support the bottom plate.

$fn = 64;

// --- parameters (all mm) ---
wall          = 3.0;
top_t         = 5.0;
depth         = 2 * 25.4;     // 50.8 mm, 2.00 in
dial_bezel_d  = 51.0;
dial_hole_d   = 45.2;
dial_body_d   = 45.0;
dial_len      = 32.3;
m4_clear_d    = 4.5;          // shank through the 5 mm top
m4_head_d     = 11.0;         // M4 wafer head (sized large enough)
boss_inner_extra = 2.0;       // extra inner diameter around the head
boss_wall     = 2.0;
boss_inner_d  = m4_head_d + boss_inner_extra;   // 13
boss_outer_d  = boss_inner_d + 2 * boss_wall;   // 17
cable_slot_w  = 20.0;
cable_slot_h  = 8.0;
ped_t         = 3.0;
ped_hole_d    = 8.2;
ped_splay     = 15;
ped_down      = 30;           // tilt below the rear axis
ped_span      = 14.0;
ped_w         = 34.0;
ped_h         = 18.0;
bottom_clear  = 3.0;
width         = 86.0;
screw_x       = 33.0;
boss_edge     = 3.0;          // outer cylinder to front / back face
print_slop    = 0.2;
edge_r        = 1.0;
bottom_chamfer  = 30.0;       // 45 deg first cut, bottom-left and bottom-right
bottom_chamfer2 = 10.0;       // extra chamfer on both edges of each 30 mm cut
fn_round      = 24;
eps           = 0.2;

bezel_r   = dial_bezel_d / 2;
center_z  = wall + bottom_clear + bezel_r;
inner_top = center_z + bezel_r;
height    = inner_top + top_t;
inner_w   = width - 2 * wall;
inner_d   = depth - 2 * wall;
inner_h   = inner_top - wall;
cable_z   = inner_top - 3.0 - cable_slot_h / 2;
hw        = width / 2;
c2s       = bottom_chamfer2 / sqrt(2);
top_open_w    = 0.5 * width;  // 50 % of width, centered
top_open_d    = 0.8 * depth;  // 80 % of depth, centered
top_open_cr   = 4.0;          // corner radius of the top hatch
screw_y_front = boss_outer_d / 2 + boss_edge;                 // 11.5
screw_y_back  = depth - boss_outer_d / 2 - boss_edge;          // 39.3
screw_ys      = [screw_y_front, screw_y_back];

preview = false;
cutaway = false;

module round_r()
    sphere(r = edge_r, $fn = fn_round);

// Convex round: minkowski of a body shrunk by edge_r.
module outer_shrunk() {
    translate([-width / 2 + edge_r, edge_r, edge_r])
        cube([width - 2 * edge_r, depth - 2 * edge_r, height - 2 * edge_r]);
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

module filleted_slot(w, h, len) {
    // Stadium: width w, height h, fully round ends, 1 mm edge rounds.
    minkowski() {
        hull() {
            translate([-(w - h) / 2, 0, 0])
                cylinder(d = max(h - 2 * edge_r, 0.2), h = len);
            translate([(w - h) / 2, 0, 0])
                cylinder(d = max(h - 2 * edge_r, 0.2), h = len);
        }
        round_r();
    }
}

// Compound bottom-corner profile (XZ): 30 mm 45 deg plus extra chamfers
// of bottom_chamfer2 on the two edges of each 30 mm face.
module left_chamfer_2d() {
    c = bottom_chamfer;
    c2 = bottom_chamfer2;
    polygon([
        [-hw - 4, -4],
        [-hw + c + c2, -4],
        [-hw + c + c2, 0],
        [-hw + c - c2s, c2s],
        [-hw + c2s, c - c2s],
        [-hw, c + c2],
        [-hw - 4, c + c2]
    ]);
}

module right_chamfer_2d() {
    c = bottom_chamfer;
    c2 = bottom_chamfer2;
    polygon([
        [hw + 4, -4],
        [hw - c - c2, -4],
        [hw - c - c2, 0],
        [hw - c + c2s, c2s],
        [hw - c2s, c - c2s],
        [hw, c + c2],
        [hw + 4, c + c2]
    ]);
}

module chamfer_2d() {
    union() {
        left_chamfer_2d();
        right_chamfer_2d();
    }
}

// grow > 0 offsets the cut so an inner cavity keeps `grow` mm of wall.
module chamfer_cuts(grow = 0) {
    y0 = -2 - edge_r;
    len = depth + 8 + 2 * edge_r;
    minkowski() {
        rotate([90, 0, 0])
            translate([0, 0, -y0 - len])
                linear_extrude(len)
                    if (grow > 0)
                        offset(r = grow)
                            chamfer_2d();
                    else
                        chamfer_2d();
        round_r();
    }
}

module outer_solid() {
    difference() {
        minkowski() {
            outer_shrunk();
            round_r();
        }
        chamfer_cuts(0);
    }
}

module inner_cavity() {
    difference() {
        minkowski() {
            inner_shrunk();
            round_r();
        }
        chamfer_cuts(wall);
    }
}

module dial_hole() {
    translate([0, -2, center_z])
        rotate([-90, 0, 0])
            filleted_cyl(dial_hole_d + print_slop, wall + 4);
}

module screw_bosses() {
    // Overlap the 5 mm top so the union is not coplanar with the underside.
    for (s = [-1, 1], y = screw_ys)
        translate([s * screw_x, y, 0])
            cylinder(d = boss_outer_d, h = inner_top + 1);
}

module screw_head_bores() {
    for (s = [-1, 1], y = screw_ys)
        translate([s * screw_x, y, -4])
            filleted_cyl(boss_inner_d, inner_top + 2);
}

module screw_top_holes() {
    for (s = [-1, 1], y = screw_ys)
        translate([s * screw_x, y, inner_top - 2])
            filleted_cyl(m4_clear_d, top_t + 4);
}

module cable_slot() {
    translate([0, depth + 2, cable_z])
        rotate([90, 0, 0])
            filleted_slot(cable_slot_w, cable_slot_h, wall + 4);
}

module rounded_rect_cut(w, d, h, cr) {
    minkowski() {
        hull() {
            for (sx = [-1, 1], sy = [-1, 1])
                translate([sx * (w / 2 - cr), sy * (d / 2 - cr), 0])
                    cylinder(r = max(cr - edge_r, 0.2), h = max(h, 0.2));
        }
        round_r();
    }
}

module top_opening() {
    translate([0, depth / 2, inner_top - 2])
        rounded_rect_cut(top_open_w, top_open_d, top_t + 8, top_open_cr);
}

// Thickened inner back around the LEDs, full case width, flush outside.
module led_housing() {
    translate([-width / 2, depth - wall - ped_t, center_z - ped_h / 2])
        cube([width, wall + ped_t + 1, ped_h]);
}

module led_holes() {
    for (side = [-1, 1]) {
        a = side * ped_splay;
        translate([side * ped_span / 2, depth, center_z])
            rotate([0, 0, a])
                rotate([-90 - ped_down, 0, 0])
                    filleted_cyl(ped_hole_d, wall + ped_t + 12, center = true);
    }
}

module case_solid() {
    difference() {
        union() {
            difference() {
                outer_solid();
                inner_cavity();
            }
            // Bosses live in the cavity, trimmed to the outer (chamfered) skin
            // so they share the 3 mm side wall instead of adding extra OD.
            intersection() {
                union() {
                    screw_bosses();
                    led_housing();
                }
                outer_solid();
            }
        }
        dial_hole();
        screw_head_bores();
        screw_top_holes();
        cable_slot();
        led_holes();
        top_opening();
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
        cube([width, depth + 10, height + 10]);
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
