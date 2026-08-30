"""MATLAB-style mechanical drawing of the Adafruit PCA9548 (5626) I2C mux case.

Board numbers read from adafruit/Adafruit-PCA9548-PCB, "PCA9548A QT Board.brd"
(Eagle 9.6.2): layer-20 outline, the four mounting holes, and the nine JST_SH4
elements. Units: millimetres. Mirrors pca9548_case.m and pca9548_case.scad.
"""
from __future__ import annotations

from pathlib import Path

import matplotlib.pyplot as plt
from matplotlib.patches import Circle, FancyBboxPatch, Rectangle
from mpl_toolkits.mplot3d.art3d import Poly3DCollection

# --- Adafruit PCA9548 breakout (product 5626) ---
PCB_X = 40.64
PCB_Y = 20.32
PCB_T = 1.60
PCB_R = 2.54
HOLE_SPAN_X = 35.56
HOLE_SPAN_Y = 15.24
PCB_HOLE_D = 2.032
PORT_X = (-11.43, -3.81, 3.81, 11.43)  # JST SH4 centres, model coordinates
JST_W = 6.20   # SM04B-SRSS-TB body width, drawn for context only
JST_H = 2.90

# --- case ---
WALL = 3.00
FLOOR_T = 2.00
FIT = 0.50
H_STD = 12.00
H_TALL = 24.00
CUT = 3.00
CUT_R = 0.60
STAND_H = 3.00
STAND_OD = 5.00
STAND_HOLE = 2.00
STAND_HOLE_DEPTH = 4.00

INNER_X = PCB_X + 2 * FIT
INNER_Y = PCB_Y + 2 * FIT
OUTER_X = INNER_X + 2 * WALL
OUTER_Y = INNER_Y + 2 * WALL
INNER_R = PCB_R + FIT
OUTER_R = INNER_R + WALL

PCB_Z0 = FLOOR_T + STAND_H     # board underside, 5.00
PCB_Z1 = PCB_Z0 + PCB_T        # board top face, 6.60
CUT_Z0 = FLOOR_T               # 2.00
CUT_Z1 = CUT_Z0 + CUT          # 5.00, flush with the board underside

MATLAB_BLUE = (0.000, 0.447, 0.741)
MATLAB_ORANGE = (0.850, 0.325, 0.098)
MATLAB_GREEN = (0.466, 0.674, 0.188)
MATLAB_GREY = (0.400, 0.400, 0.420)

OUT = Path(__file__).resolve().parent


def _matlab_rc():
    plt.rcParams.update(
        {
            "figure.facecolor": "white",
            "axes.facecolor": "white",
            "axes.edgecolor": "black",
            "axes.grid": True,
            "grid.color": (0.85, 0.85, 0.85),
            "grid.linestyle": "-",
            "font.size": 9,
            "axes.titlesize": 10,
            "axes.labelsize": 9,
        }
    )


def rounded(ax, x, y, w, h, r, **kw):
    kw.setdefault("linewidth", 1.4)
    kw.setdefault("facecolor", "none")
    ax.add_patch(
        FancyBboxPatch(
            (x + r, y + r),
            w - 2 * r,
            h - 2 * r,
            boxstyle="round,pad=%f" % r,
            mutation_aspect=1,
            **kw,
        )
    )


def dimh(ax, x1, x2, y, label, colour="black"):
    ax.annotate(
        "",
        xy=(x1, y),
        xytext=(x2, y),
        arrowprops=dict(arrowstyle="<->", color=colour, linewidth=0.9, shrinkA=0, shrinkB=0),
    )
    ax.text(
        (x1 + x2) / 2,
        y,
        label,
        ha="center",
        va="center",
        fontsize=8,
        color=colour,
        bbox=dict(facecolor="white", edgecolor="none", pad=1.0),
    )


def dimv(ax, x, y1, y2, label, colour="black"):
    ax.annotate(
        "",
        xy=(x, y1),
        xytext=(x, y2),
        arrowprops=dict(arrowstyle="<->", color=colour, linewidth=0.9, shrinkA=0, shrinkB=0),
    )
    ax.text(
        x,
        (y1 + y2) / 2,
        label,
        ha="center",
        va="center",
        rotation=90,
        fontsize=8,
        color=colour,
        bbox=dict(facecolor="white", edgecolor="none", pad=1.0),
    )


def draw_top(ax):
    """Plan view, X-Y, looking into the open top."""
    rounded(ax, -OUTER_X / 2, -OUTER_Y / 2, OUTER_X, OUTER_Y, OUTER_R, edgecolor="black")
    rounded(ax, -INNER_X / 2, -INNER_Y / 2, INNER_X, INNER_Y, INNER_R,
            edgecolor=MATLAB_GREY, linestyle="--", linewidth=1.0)
    rounded(ax, -PCB_X / 2, -PCB_Y / 2, PCB_X, PCB_Y, PCB_R,
            edgecolor=MATLAB_GREEN, linewidth=1.6)

    # four standoffs and their M2 pilot holes
    for sx in (-1, 1):
        for sy in (-1, 1):
            cx, cy = sx * HOLE_SPAN_X / 2, sy * HOLE_SPAN_Y / 2
            ax.add_patch(Circle((cx, cy), STAND_OD / 2, facecolor=(0.88, 0.93, 0.98),
                                edgecolor=MATLAB_BLUE, linewidth=1.3))
            ax.add_patch(Circle((cx, cy), STAND_HOLE / 2, facecolor="white",
                                edgecolor=MATLAB_BLUE, linewidth=1.1))

    # eight long-side cable exits, drawn through the wall band
    for sy in (-1, 1):
        y0 = sy * INNER_Y / 2
        y1 = sy * OUTER_Y / 2
        for px in PORT_X:
            ax.add_patch(Rectangle((px - CUT / 2, min(y0, y1)), CUT, WALL,
                                   facecolor=(1.0, 0.90, 0.84), edgecolor=MATLAB_ORANGE,
                                   linewidth=1.3))
            ax.add_patch(Rectangle((px - JST_W / 2, sy * (PCB_Y / 2 - 3.6)), JST_W, 3.6,
                                   facecolor="none", edgecolor=MATLAB_GREEN,
                                   linewidth=0.7, linestyle=":"))
    # one end cable exit, controller-side STEMMA QT
    ax.add_patch(Rectangle((-OUTER_X / 2, -CUT / 2), WALL, CUT,
                           facecolor=(1.0, 0.90, 0.84), edgecolor=MATLAB_ORANGE, linewidth=1.3))

    dimh(ax, -OUTER_X / 2, OUTER_X / 2, OUTER_Y / 2 + 5.0, "%.2f" % OUTER_X)
    dimh(ax, -PCB_X / 2, PCB_X / 2, -OUTER_Y / 2 - 4.0, "PCB %.2f" % PCB_X, MATLAB_GREEN)
    dimv(ax, -OUTER_X / 2 - 5.0, -OUTER_Y / 2, OUTER_Y / 2, "%.2f" % OUTER_Y)
    dimh(ax, PORT_X[0], PORT_X[1], OUTER_Y / 2 + 1.4, "7.62", MATLAB_ORANGE)
    dimh(ax, -HOLE_SPAN_X / 2, HOLE_SPAN_X / 2, 3.2, "%.2f" % HOLE_SPAN_X, MATLAB_BLUE)
    dimv(ax, -6.5, -HOLE_SPAN_Y / 2, HOLE_SPAN_Y / 2, "%.2f" % HOLE_SPAN_Y, MATLAB_BLUE)

    ax.set_title("Plan  (X-Y)   open top, board goes in component-side down")
    ax.set_xlabel("X mm")
    ax.set_ylabel("Y mm")
    ax.set_xlim(-OUTER_X / 2 - 9, OUTER_X / 2 + 6)
    ax.set_ylim(-OUTER_Y / 2 - 8, OUTER_Y / 2 + 8)
    ax.set_aspect("equal")


def draw_long_elevation(ax, height, label):
    """Long-side elevation, X-Z. The 3x3 cutouts sit on the floor in both variants."""
    ax.add_patch(Rectangle((-OUTER_X / 2, 0), OUTER_X, height,
                           facecolor=(0.94, 0.96, 0.99), edgecolor="black", linewidth=1.4))
    ax.plot([-OUTER_X / 2, OUTER_X / 2], [FLOOR_T, FLOOR_T],
            color=MATLAB_GREY, linewidth=1.0, linestyle="--")

    for px in PORT_X:
        rounded(ax, px - CUT / 2, CUT_Z0, CUT, CUT, CUT_R,
                edgecolor=MATLAB_ORANGE, facecolor="white", linewidth=1.5)

    # board and its top-side parts, hidden behind the wall
    ax.add_patch(Rectangle((-PCB_X / 2, PCB_Z0), PCB_X, PCB_T,
                           facecolor=(0.90, 0.96, 0.88), edgecolor=MATLAB_GREEN,
                           linewidth=1.2, linestyle=(0, (5, 2))))
    for px in PORT_X:
        ax.add_patch(Rectangle((px - JST_W / 2, PCB_Z0 - JST_H), JST_W, JST_H,
                               facecolor="none", edgecolor=MATLAB_GREEN,
                               linewidth=0.8, linestyle=":"))

    for sx in (-1, 1):
        cx = sx * HOLE_SPAN_X / 2
        ax.add_patch(Rectangle((cx - STAND_OD / 2, FLOOR_T), STAND_OD, STAND_H,
                               facecolor=(0.88, 0.93, 0.98), edgecolor=MATLAB_BLUE,
                               linewidth=1.1))
        ax.plot([cx, cx], [PCB_Z0 - STAND_HOLE_DEPTH, PCB_Z0],
                color=MATLAB_BLUE, linewidth=1.6)

    dimv(ax, OUTER_X / 2 + 3.0, 0, height, "%.2f" % height)
    dimv(ax, -OUTER_X / 2 - 3.0, CUT_Z0, CUT_Z1, "%.2f" % CUT, MATLAB_ORANGE)
    dimv(ax, -OUTER_X / 2 - 7.5, 0, FLOOR_T, "%.2f" % FLOOR_T, MATLAB_GREY)
    ax.text(0, height + 1.2, "board top face Z=%.2f" % PCB_Z1, ha="center", fontsize=8,
            color=MATLAB_GREEN)

    ax.set_title(label)
    ax.set_xlabel("X mm")
    ax.set_ylabel("Z mm")
    ax.set_xlim(-OUTER_X / 2 - 11, OUTER_X / 2 + 8)
    ax.set_ylim(-2, max(H_TALL, height) + 4)
    ax.set_aspect("equal")


def draw_end_elevation(ax):
    """Left end elevation, Y-Z. One 3x3 on the controller-side STEMMA QT port."""
    ax.add_patch(Rectangle((-OUTER_Y / 2, 0), OUTER_Y, H_STD,
                           facecolor=(0.94, 0.96, 0.99), edgecolor="black", linewidth=1.4))
    ax.plot([-OUTER_Y / 2, OUTER_Y / 2], [FLOOR_T, FLOOR_T],
            color=MATLAB_GREY, linewidth=1.0, linestyle="--")
    rounded(ax, -CUT / 2, CUT_Z0, CUT, CUT, CUT_R,
            edgecolor=MATLAB_ORANGE, facecolor="white", linewidth=1.5)
    ax.add_patch(Rectangle((-PCB_Y / 2, PCB_Z0), PCB_Y, PCB_T,
                           facecolor=(0.90, 0.96, 0.88), edgecolor=MATLAB_GREEN,
                           linewidth=1.2, linestyle=(0, (5, 2))))
    for sy in (-1, 1):
        cy = sy * HOLE_SPAN_Y / 2
        ax.add_patch(Rectangle((cy - STAND_OD / 2, FLOOR_T), STAND_OD, STAND_H,
                               facecolor=(0.88, 0.93, 0.98), edgecolor=MATLAB_BLUE,
                               linewidth=1.1))
    dimh(ax, -OUTER_Y / 2, OUTER_Y / 2, H_STD + 2.2, "%.2f" % OUTER_Y)
    dimv(ax, OUTER_Y / 2 + 3.0, 0, H_STD, "%.2f" % H_STD)
    dimv(ax, -CUT / 2 - 2.2, CUT_Z0, CUT_Z1, "%.2f" % CUT, MATLAB_ORANGE)
    ax.set_title("Left end  (Y-Z)   controller-side port exit, on centre")
    ax.set_xlabel("Y mm")
    ax.set_ylabel("Z mm")
    ax.set_xlim(-OUTER_Y / 2 - 7, OUTER_Y / 2 + 8)
    ax.set_ylim(-2, H_STD + 6)
    ax.set_aspect("equal")


def _box(ax, x0, x1, y0, y1, z0, z1, colour, alpha):
    faces = [
        [(x0, y0, z0), (x1, y0, z0), (x1, y1, z0), (x0, y1, z0)],
        [(x0, y0, z1), (x1, y0, z1), (x1, y1, z1), (x0, y1, z1)],
        [(x0, y0, z0), (x1, y0, z0), (x1, y0, z1), (x0, y0, z1)],
        [(x0, y1, z0), (x1, y1, z0), (x1, y1, z1), (x0, y1, z1)],
        [(x0, y0, z0), (x0, y1, z0), (x0, y1, z1), (x0, y0, z1)],
        [(x1, y0, z0), (x1, y1, z0), (x1, y1, z1), (x1, y0, z1)],
    ]
    ax.add_collection3d(
        Poly3DCollection(faces, facecolor=colour, edgecolor=(0.1, 0.1, 0.1),
                         linewidths=0.5, alpha=alpha)
    )


def draw_iso(ax, height, label):
    ox, oy = OUTER_X / 2, OUTER_Y / 2
    ix, iy = INNER_X / 2, INNER_Y / 2
    _box(ax, -ox, ox, -oy, oy, 0, FLOOR_T, MATLAB_BLUE, 0.30)          # floor
    _box(ax, -ox, ox, iy, oy, FLOOR_T, height, MATLAB_BLUE, 0.18)      # +Y wall
    _box(ax, -ox, ox, -oy, -iy, FLOOR_T, height, MATLAB_BLUE, 0.18)    # -Y wall
    _box(ax, -ox, -ix, -iy, iy, FLOOR_T, height, MATLAB_BLUE, 0.18)    # -X wall
    _box(ax, ix, ox, -iy, iy, FLOOR_T, height, MATLAB_BLUE, 0.18)      # +X wall
    for px in PORT_X:                                                   # cable exits
        for sy in (-1, 1):
            lo, hi = sorted((sy * iy, sy * oy))
            _box(ax, px - CUT / 2, px + CUT / 2, lo, hi, CUT_Z0, CUT_Z1,
                 MATLAB_ORANGE, 0.85)
    _box(ax, -ox, -ix, -CUT / 2, CUT / 2, CUT_Z0, CUT_Z1, MATLAB_ORANGE, 0.85)
    for sx in (-1, 1):                                                  # standoffs
        for sy in (-1, 1):
            cx, cy = sx * HOLE_SPAN_X / 2, sy * HOLE_SPAN_Y / 2
            _box(ax, cx - STAND_OD / 2, cx + STAND_OD / 2, cy - STAND_OD / 2,
                 cy + STAND_OD / 2, FLOOR_T, PCB_Z0, MATLAB_BLUE, 0.55)
    _box(ax, -PCB_X / 2, PCB_X / 2, -PCB_Y / 2, PCB_Y / 2, PCB_Z0, PCB_Z1,
         MATLAB_GREEN, 0.45)

    ax.set_title(label)
    ax.set_xlabel("X mm")
    ax.set_ylabel("Y mm")
    ax.set_zlabel("Z mm")
    ax.set_xlim(-ox - 2, ox + 2)
    ax.set_ylim(-oy - 2, oy + 2)
    ax.set_zlim(0, H_TALL + 2)
    ax.set_box_aspect((OUTER_X, OUTER_Y, H_TALL))
    ax.view_init(elev=24, azim=-52)


def draw_notes(ax):
    ax.axis("off")
    lines = [
        "BOARD  Adafruit PCA9548 8-ch STEMMA QT / Qwiic mux, product 5626",
        "  outline           40.64 x 20.32 mm, corner R2.54, t 1.60 mm",
        "  mounting holes    4 x D%.3f, span %.2f x %.2f mm" % (PCB_HOLE_D, HOLE_SPAN_X, HOLE_SPAN_Y),
        "  ports             8 x JST SH4 on the long edges, pitch 7.62 mm",
        "                    port X = -11.43 / -3.81 / +3.81 / +11.43",
        "                    1 x JST SH4 controller side, left end, on Y=0",
        "  source            adafruit/Adafruit-PCA9548-PCB",
        "                    PCA9548A QT Board.brd, layer 20 outline + holes",
        "",
        "CASE   open-top tray, two heights",
        "  outer             %.2f x %.2f mm, corner R%.2f" % (OUTER_X, OUTER_Y, OUTER_R),
        "  height std        %.2f mm      pca9548_case_std.stl" % H_STD,
        "  height tall       %.2f mm      pca9548_case_tall.stl" % H_TALL,
        "  wall / floor      %.2f / %.2f mm" % (WALL, FLOOR_T),
        "  board clearance   %.2f mm per side" % FIT,
        "  standoffs         4 x D%.2f, h %.2f mm off the floor" % (STAND_OD, STAND_H),
        "  screw pilots      D%.2f x %.2f deep, M2 self-tap" % (STAND_HOLE, STAND_HOLE_DEPTH),
        "  board underside   Z = %.2f mm    top face Z = %.2f mm" % (PCB_Z0, PCB_Z1),
        "",
        "CABLE EXITS  9 x %.0f x %.0f mm, corner R%.2f" % (CUT, CUT, CUT_R),
        "  Z %.2f to %.2f mm in BOTH variants" % (CUT_Z0, CUT_Z1),
        "  4 in each long wall on the eight JST SH port centres",
        "  1 in the left end wall on the controller-side port",
        "  top edge is flush with the board underside, Z %.2f" % PCB_Z0,
        "  the tall variant adds wall ABOVE the board only",
        "",
        "PRINT  floor on the bed, open top up, no support needed.",
        "  Board goes in component-side down, so the JST shells sit in the",
        "  cutout mouths and every cable leaves at floor level.",
    ]
    ax.text(0.0, 1.0, "\n".join(lines), ha="left", va="top", fontsize=8.2,
            family="DejaVu Sans Mono", transform=ax.transAxes)


def main():
    _matlab_rc()
    fig = plt.figure(figsize=(17.5, 11.0), dpi=130)
    fig.suptitle(
        "Adafruit PCA9548 (5626) 8-channel I2C mux case   "
        "%.2f x %.2f mm   h %.0f mm and h %.0f mm   "
        "9 x %.0fx%.0f mm cable exits at Z %.0f-%.0f in both"
        % (OUTER_X, OUTER_Y, H_STD, H_TALL, CUT, CUT, CUT_Z0, CUT_Z1),
        fontsize=12,
        fontweight="medium",
        y=0.975,
    )
    gs = fig.add_gridspec(2, 3, left=0.045, right=0.985, top=0.905, bottom=0.05,
                          wspace=0.26, hspace=0.30)
    draw_top(fig.add_subplot(gs[0, 0]))
    draw_long_elevation(fig.add_subplot(gs[0, 1]), H_STD, "Long side  (X-Z)   std, 12.00 mm")
    draw_long_elevation(fig.add_subplot(gs[0, 2]), H_TALL, "Long side  (X-Z)   tall, 24.00 mm")
    draw_end_elevation(fig.add_subplot(gs[1, 0]))
    draw_iso(fig.add_subplot(gs[1, 1], projection="3d"), H_TALL,
             "Isometric   tall variant, board in place")
    draw_notes(fig.add_subplot(gs[1, 2]))
    png = OUT / "pca9548_case_drawing.png"
    fig.savefig(png, dpi=150)
    print("wrote %s" % png)
    return png


if __name__ == "__main__":
    main()
