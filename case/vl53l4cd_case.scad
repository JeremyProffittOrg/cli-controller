// Adafruit VL53L4CD (product 5396) time-of-flight case.
// Units: millimetres. Matches vl53l4cd_case.m / draw_vl53l4cd_case.py.
// Open back. Four 3 mm bosses with 2 mm x 4 mm holes (not through the front).
// Half-circle wings on the BACK. Print with the BACK on the bed (Z=0).

$fn = 64;
eps = 0.2;

// --- Adafruit VL53L4CD breakout ---
pcb_x       = 25.40;
pcb_z       = 17.78;
hole_span_x = 20.32;
hole_span_z = 12.70;

// --- case ---
wall      = 3.00;
front_t   = 3.00;
fit       = 0.50;
outer_x   = 39.00;
depth     = 12.00;
cut       = 3.00;
cut_r     = 0.60;
cut_base_r = 1.50;      // outward round where the 3x3 meets the base (z=0)
window_s  = 10.00;
window_r  = 2.00;
boss_h    = 3.00;
boss_od   = 5.00;
boss_hole = 2.00;
boss_hole_depth = 4.00;
wing_t    = 4.00;
wing_r    = 12.00;
wing_hole_d = 4.00;
wing_fillet = 3.00;     // convex round at wing-to-body, going out

inner_x = outer_x - 2 * wall;
inner_z = pcb_z + 2 * fit;
outer_z = inner_z + 2 * wall;
front_y = depth;                  // outer front
inner_front_y = depth - front_t;  // inner face of the front plate
wing_hole_out = wing_r / 2;

module rounded_square(s, r) {
    offset(r = r)
        square(s - 2 * r, center = true);
}

// 3x3 mouse-hole in (Z, Y): sits on Y=0 (print Z=0), flares out at the base.
module side_cut_2d() {
    union() {
        hull() {
            translate([-cut / 2 + cut_r, cut - cut_r]) circle(r = cut_r);
            translate([ cut / 2 - cut_r, cut - cut_r]) circle(r = cut_r);
            translate([-cut / 2, -eps])
                square([cut, 0.4]);
        }
        translate([-cut / 2, 0]) circle(r = cut_base_r);
        translate([ cut / 2, 0]) circle(r = cut_base_r);
    }
}

// 2D wing in X (along the 39 mm side) and Y (out from the body).
// offset(-r) offset(r) rounds the two convex 90 deg roots going out along the outline.
module wing_2d() {
    offset(r = -wing_fillet) offset(r = wing_fillet)
        union() {
            translate([-outer_x / 2, -wall - 2])
                square([outer_x, wall + 2.4]);
            intersection() {
                circle(r = wing_r);
                translate([-wing_r - 1, 0])
                    square([2 * wing_r + 2, wing_r + 1]);
            }
        }
}

// Fill the L where the wall rises off the wing, r going out into the empty corner.
module wing_step_fillet(sz) {
    r = wing_fillet;
    translate([0, wing_t, sz * outer_z / 2])
        difference() {
            translate([-wing_r, 0, min(0, sz * r)])
                cube([2 * wing_r, r, abs(r)]);
            translate([0, r, sz * r])
                rotate([0, 90, 0])
                    cylinder(r = r, h = 2 * wing_r + 2, center = true);
        }
}

module wings() {
    translate([0, wing_t, outer_z / 2])
        rotate([90, 0, 0])
            linear_extrude(wing_t)
                wing_2d();
    translate([0, wing_t, -outer_z / 2])
        rotate([90, 0, 0])
            linear_extrude(wing_t)
                mirror([0, 1, 0])
                    wing_2d();
}

module bosses() {
    for (sx = [-1, 1])
        for (sz = [-1, 1])
            translate([sx * hole_span_x / 2, inner_front_y + 0.2, sz * hole_span_z / 2])
                rotate([90, 0, 0])
                    cylinder(d = boss_od, h = boss_h + 0.2);
}

module shell() {
    difference() {
        union() {
            translate([-outer_x / 2, 0, -outer_z / 2])
                cube([outer_x, depth, outer_z]);
            wings();
        }
        // cavity open at the back (y=0)
        translate([-inner_x / 2, -eps, -inner_z / 2])
            cube([inner_x, inner_front_y + eps, inner_z]);

        // rounded-square window through the front plate
        translate([0, inner_front_y - eps, 0])
            rotate([-90, 0, 0])
                linear_extrude(front_t + 2 * eps)
                    rounded_square(window_s, window_r);

        // 3 mm squares THROUGH the side walls, on the base (print Z=0),
        // with outward rounds where they meet the bed.
        for (sx = [-1, 1])
            translate([sx * (outer_x / 2 - wall / 2), 0, 0])
                rotate([0, 90, 0])
                    linear_extrude(wall + 2, center = true)
                        side_cut_2d();

        // 4 mm holes, centred in the 12 mm wings
        for (sz = [-1, 1])
            translate([0, -eps, sz * (outer_z / 2 + wing_hole_out)])
                rotate([-90, 0, 0])
                    cylinder(d = wing_hole_d, h = wing_t + 2 * eps);
    }
}

module case_body() {
    difference() {
        union() {
            shell();
            bosses();
        }
        // 2 mm holes, 4 mm deep, from the boss tip toward the front.
        // 3 mm through the boss + 1 mm into the front plate; 2 mm remains to the face.
        for (sx = [-1, 1])
            for (sz = [-1, 1])
                translate([sx * hole_span_x / 2, inner_front_y - boss_h - eps, sz * hole_span_z / 2])
                    rotate([-90, 0, 0])
                        cylinder(d = boss_hole, h = boss_hole_depth + 2 * eps);
    }
}

// Print orientation: back (open, with wings) is Z=0, on the bed.
rotate([90, 0, 0])
    case_body();
