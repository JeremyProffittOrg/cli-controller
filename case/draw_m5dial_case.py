"""MATLAB-style mechanical drawing of the M5Stack Dial under-shelf case.

Dimensions match m5dial_shelf_case.scad. Units: millimetres.
"""
from __future__ import annotations

from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
from matplotlib.patches import Circle, FancyArrowPatch, Polygon, Rectangle
from mpl_toolkits.mplot3d.art3d import Poly3DCollection

# --- same numbers as the SCAD ---
WALL = 3.0
TOP_T = 5.0
DEPTH = 2 * 25.4
DIAL_BEZEL_D = 51.0
DIAL_HOLE_D = 45.2
M4_CLEAR_D = 4.5
M4_HEAD_D = 11.0
BOSS_INNER_D = M4_HEAD_D + 2.0
BOSS_WALL = 2.0
BOSS_OUTER_D = BOSS_INNER_D + 2 * BOSS_WALL
CABLE_SLOT_W = 20.0
CABLE_SLOT_H = 8.0
PED_T = 3.0
PED_HOLE_D = 8.2
PED_SPLAY = 15.0
PED_DOWN = 30.0
PED_SPAN = 14.0
PED_W = 34.0
PED_H = 18.0
BOTTOM_CLEAR = 3.0
WIDTH = 86.0
SCREW_X = 33.0
SCREW_Y = 18.0
BOTTOM_CHAMFER = 30.0
BOTTOM_CHAMFER2 = 10.0
BEZEL_R = DIAL_BEZEL_D / 2
CENTER_Z = WALL + BOTTOM_CLEAR + BEZEL_R
INNER_TOP = CENTER_Z + BEZEL_R
HEIGHT = INNER_TOP + TOP_T
CABLE_Z = INNER_TOP - 3.0 - CABLE_SLOT_H / 2
HW = WIDTH / 2
C2S = BOTTOM_CHAMFER2 / np.sqrt(2)
TOP_OPEN_W = 0.5 * WIDTH
TOP_OPEN_D = 0.8 * DEPTH
TOP_OPEN_CR = 4.0

MATLAB_BLUE = (0.000, 0.447, 0.741)
MATLAB_ORANGE = (0.850, 0.325, 0.098)
MATLAB_GREEN = (0.466, 0.674, 0.188)
MATLAB_PURPLE = (0.494, 0.184, 0.556)

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
            "font.family": "DejaVu Sans",
            "axes.titlesize": 11,
            "axes.labelsize": 9,
        }
    )


def _circle(ax, xy, r, **kw):
    ax.add_patch(Circle(xy, r, fill=False, **kw))


def _rect(ax, xy, w, h, **kw):
    ax.add_patch(Rectangle(xy, w, h, fill=False, **kw))


def _dim_h(ax, x1, x2, y, text, side=1):
    ax.annotate(
        "",
        xy=(x2, y),
        xytext=(x1, y),
        arrowprops=dict(arrowstyle="<->", color="k", lw=0.8),
        annotation_clip=False,
    )
    ax.text(
        (x1 + x2) / 2,
        y + 1.6 * side,
        text,
        ha="center",
        va="bottom" if side > 0 else "top",
        fontsize=8,
        color="k",
        bbox=dict(boxstyle="square,pad=0.15", fc="white", ec="none"),
        clip_on=False,
    )


def _dim_v(ax, x, y1, y2, text, side=1):
    ax.annotate(
        "",
        xy=(x, y2),
        xytext=(x, y1),
        arrowprops=dict(arrowstyle="<->", color="k", lw=0.8),
        annotation_clip=False,
    )
    ax.text(
        x + 2.0 * side,
        (y1 + y2) / 2,
        text,
        ha="left" if side > 0 else "right",
        va="center",
        fontsize=8,
        color="k",
        rotation=90 if abs(side) == 1 else 0,
        bbox=dict(boxstyle="square,pad=0.15", fc="white", ec="none"),
        clip_on=False,
    )


def _front_outline():
    hw = HW
    c = BOTTOM_CHAMFER
    c2 = BOTTOM_CHAMFER2
    s2 = C2S
    return [
        (-hw + c + c2, 0),
        (hw - c - c2, 0),
        (hw - c + s2, s2),
        (hw - s2, c - s2),
        (hw, c + c2),
        (hw, HEIGHT),
        (-hw, HEIGHT),
        (-hw, c + c2),
        (-hw + s2, c - s2),
        (-hw + c - s2, s2),
    ]


def draw_front(ax):
    ax.add_patch(Polygon(_front_outline(), fill=False, ec="k", lw=1.4, closed=True))
    _rect(ax, (-WIDTH / 2, INNER_TOP), WIDTH, TOP_T, ec=MATLAB_BLUE, lw=1.0, ls="--")
    _circle(ax, (0, CENTER_Z), DIAL_HOLE_D / 2, ec=MATLAB_ORANGE, lw=1.8)
    _circle(ax, (0, CENTER_Z), DIAL_BEZEL_D / 2, ec=MATLAB_ORANGE, lw=0.8, ls="--")
    for s in (-1, 1):
        ax.plot([s * SCREW_X, s * SCREW_X], [INNER_TOP, HEIGHT], color=MATLAB_BLUE, lw=1.4)
        _circle(ax, (s * SCREW_X, HEIGHT - TOP_T / 2), M4_CLEAR_D / 2, ec=MATLAB_BLUE, lw=1.0)
        _circle(ax, (s * SCREW_X, CENTER_Z), BOSS_INNER_D / 2, ec=MATLAB_BLUE, lw=0.8, ls=":")
    ax.text(
        0,
        CENTER_Z,
        "dial hole\n45.2 mm",
        ha="center",
        va="center",
        fontsize=8,
        color=MATLAB_ORANGE,
    )
    ax.text(
        0,
        HEIGHT + 3,
        "M4 wafer  4.5 mm through top  13 mm boss ID",
        ha="center",
        fontsize=8,
        color=MATLAB_BLUE,
        clip_on=False,
    )
    _dim_h(ax, -WIDTH / 2, WIDTH / 2, HEIGHT + 10, f"{WIDTH:.0f}")
    _dim_v(ax, -WIDTH / 2 - 10, 0, HEIGHT, f"{HEIGHT:.0f}", side=-1)
    _dim_v(ax, WIDTH / 2 + 8, CENTER_Z - DIAL_HOLE_D / 2, CENTER_Z + DIAL_HOLE_D / 2, "45.2")
    _dim_h(ax, -HW, -HW + BOTTOM_CHAMFER, -6, f"{BOTTOM_CHAMFER:.0f} + {BOTTOM_CHAMFER2:.0f} chamfer", side=-1)
    ax.set_title("Front  (X-Z)")
    ax.set_xlabel("X")
    ax.set_ylabel("Z")
    ax.set_aspect("equal")
    ax.set_xlim(-WIDTH / 2 - 18, WIDTH / 2 + 18)
    ax.set_ylim(-8, HEIGHT + 18)


def draw_top(ax):
    _rect(ax, (-WIDTH / 2, 0), WIDTH, DEPTH, ec="k", lw=1.4)
    y0 = (DEPTH - TOP_OPEN_D) / 2
    _rect(ax, (-TOP_OPEN_W / 2, y0), TOP_OPEN_W, TOP_OPEN_D, ec=MATLAB_ORANGE, lw=1.4, ls="--")
    ax.text(0, DEPTH / 2, "top hatch\n50% W x 80% D", ha="center", va="center", fontsize=8, color=MATLAB_ORANGE)
    ax.plot([-DIAL_HOLE_D / 2, DIAL_HOLE_D / 2], [0, 0], color=MATLAB_ORANGE, lw=2.4)
    ax.plot([-DIAL_BEZEL_D / 2, DIAL_BEZEL_D / 2], [-0.8, -0.8], color=MATLAB_ORANGE, lw=1.0, ls="--")
    for s in (-1, 1):
        _circle(ax, (s * SCREW_X, SCREW_Y), M4_CLEAR_D / 2, ec=MATLAB_BLUE, lw=1.5)
        _circle(ax, (s * SCREW_X, SCREW_Y), BOSS_INNER_D / 2, ec=MATLAB_BLUE, lw=0.8, ls=":")
        _circle(ax, (s * SCREW_X, SCREW_Y), BOSS_OUTER_D / 2, ec=MATLAB_BLUE, lw=0.8)
        a = np.deg2rad(s * PED_SPLAY)
        x0, y0h = s * PED_SPAN / 2, DEPTH
        x1, y1 = x0 + 16 * np.sin(a), y0h + 10 * np.cos(a)
        ax.annotate(
            "",
            xy=(x1, y1),
            xytext=(x0, y0h),
            arrowprops=dict(arrowstyle="-|>", color=MATLAB_GREEN, lw=1.4),
        )
        _circle(ax, (x0, y0h), PED_HOLE_D / 2, ec=MATLAB_GREEN, lw=1.2)
    ax.text(0, SCREW_Y + 12, "M4  13 mm ID boss  2 mm wall", ha="center", fontsize=8, color=MATLAB_BLUE)
    ax.text(
        0,
        DEPTH + 9,
        "8.2 mm flush in back wall   +/-15 deg splay   30 deg down",
        ha="center",
        fontsize=8,
        color=MATLAB_GREEN,
        clip_on=False,
    )
    _dim_h(ax, -WIDTH / 2, WIDTH / 2, -10, f"{WIDTH:.0f}")
    _dim_v(ax, WIDTH / 2 + 10, 0, DEPTH, f"{DEPTH:.1f}  (2.00 in)")
    _dim_h(ax, -SCREW_X, SCREW_X, SCREW_Y - 8, f"{2 * SCREW_X:.0f}  (screw span)", side=-1)
    ax.set_title("Top  (X-Y)  looking down through the shelf")
    ax.set_xlabel("X")
    ax.set_ylabel("Y  (front=0, back=+Y)")
    ax.set_aspect("equal")
    ax.set_xlim(-WIDTH / 2 - 18, WIDTH / 2 + 22)
    ax.set_ylim(-16, DEPTH + 16)


def draw_right(ax):
    _rect(ax, (0, 0), DEPTH, HEIGHT, ec="k", lw=1.4)
    y0 = (DEPTH - TOP_OPEN_D) / 2
    _rect(ax, (y0, INNER_TOP), TOP_OPEN_D, TOP_T, ec=MATLAB_ORANGE, lw=1.2, ls="--")
    _rect(ax, (0, INNER_TOP), DEPTH, TOP_T, ec=MATLAB_BLUE, lw=1.0, ls="--")
    ax.plot([0, WALL], [CENTER_Z - DIAL_HOLE_D / 2] * 2, color=MATLAB_ORANGE, lw=1.6)
    ax.plot([0, WALL], [CENTER_Z + DIAL_HOLE_D / 2] * 2, color=MATLAB_ORANGE, lw=1.6)
    slot = Rectangle(
        (DEPTH - WALL, CABLE_Z - CABLE_SLOT_H / 2),
        WALL,
        CABLE_SLOT_H,
        fill=False,
        ec="k",
        lw=1.4,
    )
    ax.add_patch(slot)
    ax.plot([SCREW_Y, SCREW_Y], [INNER_TOP, HEIGHT], color=MATLAB_BLUE, lw=1.4)
    _circle(ax, (SCREW_Y, INNER_TOP / 2), BOSS_INNER_D / 2, ec=MATLAB_BLUE, lw=0.8, ls=":")
    ax.plot([0, DEPTH], [INNER_TOP, INNER_TOP], color=MATLAB_BLUE, lw=0.6, ls=":")
    ax.text(DEPTH + PED_T + 2, CABLE_Z, "20 mm\nslot", fontsize=8, va="center", color="k")
    _dim_v(ax, DEPTH + PED_T + 14, 0, HEIGHT, f"{HEIGHT:.0f}")
    _dim_h(ax, 0, DEPTH, HEIGHT + 8, f"{DEPTH:.1f}")
    _dim_v(ax, WALL + 6, INNER_TOP, HEIGHT, "5", side=1)
    ax.set_title("Right  (Y-Z)")
    ax.set_xlabel("Y")
    ax.set_ylabel("Z")
    ax.set_aspect("equal")
    ax.set_xlim(-16, DEPTH + PED_T + 24)
    ax.set_ylim(-8, HEIGHT + 16)


def draw_back(ax):
    _rect(ax, (-WIDTH / 2, 0), WIDTH, HEIGHT, ec="k", lw=1.4)
    _rect(ax, (-PED_W / 2, CENTER_Z - PED_H / 2), PED_W, PED_H, ec=MATLAB_GREEN, lw=1.5)
    ax.add_patch(
        Rectangle(
            (-CABLE_SLOT_W / 2, CABLE_Z - CABLE_SLOT_H / 2),
            CABLE_SLOT_W,
            CABLE_SLOT_H,
            fill=False,
            ec="k",
            lw=1.6,
            joinstyle="round",
        )
    )
    ax.text(0, CABLE_Z + CABLE_SLOT_H / 2 + 4, "20 x 8 mm slot, round ends", ha="center", fontsize=8)
    for s in (-1, 1):
        _circle(ax, (s * PED_SPAN / 2, CENTER_Z), PED_HOLE_D / 2, ec=MATLAB_GREEN, lw=1.6)
        a = np.deg2rad(s * PED_SPLAY)
        ax.annotate(
            "",
            xy=(s * PED_SPAN / 2 + 14 * np.sin(a), CENTER_Z - 10),
            xytext=(s * PED_SPAN / 2, CENTER_Z),
            arrowprops=dict(arrowstyle="-|>", color=MATLAB_GREEN, lw=1.3),
        )
        _circle(ax, (s * SCREW_X, HEIGHT - TOP_T / 2), M4_CLEAR_D / 2, ec=MATLAB_BLUE, lw=1.0)
    ax.text(
        0,
        CENTER_Z - PED_H / 2 - 7,
        "two 8.2 mm holes   flush with back   +/-15 deg   30 deg down",
        ha="center",
        fontsize=8,
        color=MATLAB_GREEN,
    )
    ax.set_title("Back  (X-Z)  LED holes in the rear wall")
    ax.set_xlabel("X")
    ax.set_ylabel("Z")
    ax.set_aspect("equal")
    ax.set_xlim(-WIDTH / 2 - 8, WIDTH / 2 + 8)
    ax.set_ylim(-8, HEIGHT + 10)


def _box_faces(x0, x1, y0, y1, z0, z1):
    return [
        [(x0, y0, z0), (x1, y0, z0), (x1, y0, z1), (x0, y0, z1)],
        [(x0, y1, z0), (x1, y1, z0), (x1, y1, z1), (x0, y1, z1)],
        [(x0, y0, z0), (x0, y1, z0), (x0, y1, z1), (x0, y0, z1)],
        [(x1, y0, z0), (x1, y1, z0), (x1, y1, z1), (x1, y0, z1)],
        [(x0, y0, z1), (x1, y0, z1), (x1, y1, z1), (x0, y1, z1)],
        [(x0, y0, z0), (x1, y0, z0), (x1, y1, z0), (x0, y1, z0)],
    ]


def _add_poly(ax, faces, color, alpha, zsort="average"):
    poly = Poly3DCollection(
        faces, facecolors=[color] * len(faces), edgecolors="k", linewidths=0.6, alpha=alpha
    )
    poly.set_zsort(zsort)
    ax.add_collection3d(poly)


def _circ3(ax, origin, radius, normal, color, lw=1.6, n=48):
    nrm = np.asarray(normal, dtype=float)
    nrm = nrm / np.linalg.norm(nrm)
    a = np.cross(nrm, (1.0, 0.0, 0.0) if abs(nrm[0]) < 0.9 else (0.0, 1.0, 0.0))
    a = a / np.linalg.norm(a)
    b = np.cross(nrm, a)
    t = np.linspace(0, 2 * np.pi, n)
    p = np.asarray(origin) + radius * (np.outer(np.cos(t), a) + np.outer(np.sin(t), b))
    ax.plot(p[:, 0], p[:, 1], p[:, 2], color=color, lw=lw)


def draw_iso(ax):
    x0, x1 = -WIDTH / 2, WIDTH / 2
    _add_poly(ax, _box_faces(x0, x1, 0, DEPTH, 0, HEIGHT), MATLAB_BLUE, 0.28)
    _add_poly(
        ax,
        _box_faces(-PED_W / 2, PED_W / 2, DEPTH - WALL, DEPTH, CENTER_Z - PED_H / 2, CENTER_Z + PED_H / 2),
        MATLAB_GREEN,
        0.55,
    )
    # top plate tint
    _add_poly(ax, [_box_faces(x0, x1, 0, DEPTH, INNER_TOP, HEIGHT)[4]], (0.2, 0.2, 0.3), 0.45)
    _circ3(ax, (0, 0, CENTER_Z), DIAL_HOLE_D / 2, (0, 1, 0), MATLAB_ORANGE, lw=2.0)
    _circ3(ax, (0, 0, CENTER_Z), DIAL_BEZEL_D / 2, (0, 1, 0), MATLAB_ORANGE, lw=0.8)
    _circ3(ax, (0, DEPTH, CABLE_Z), CABLE_SLOT_H / 2, (0, 1, 0), "k", lw=1.6)
    for s in (-1, 1):
        _circ3(ax, (s * SCREW_X, SCREW_Y, HEIGHT), M4_CLEAR_D / 2, (0, 0, 1), MATLAB_BLUE, lw=1.6)
        _circ3(ax, (s * SCREW_X, SCREW_Y, INNER_TOP / 2), BOSS_INNER_D / 2, (0, 0, 1), MATLAB_BLUE, lw=1.0)
        a = np.deg2rad(s * PED_SPLAY)
        down = np.deg2rad(PED_DOWN)
        nrm = (np.sin(a) * np.cos(down), np.cos(a) * np.cos(down), -np.sin(down))
        _circ3(
            ax,
            (s * PED_SPAN / 2, DEPTH, CENTER_Z),
            PED_HOLE_D / 2,
            nrm,
            MATLAB_GREEN,
            lw=1.6,
        )
        ax.plot(
            [s * PED_SPAN / 2, s * PED_SPAN / 2 + 12 * nrm[0]],
            [DEPTH, DEPTH + 12 * nrm[1]],
            [CENTER_Z, CENTER_Z + 12 * nrm[2]],
            color=MATLAB_GREEN,
            lw=1.4,
        )
    ax.view_init(elev=22, azim=-52)
    ax.set_box_aspect((WIDTH, DEPTH, HEIGHT))
    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_zlabel("Z")
    ax.set_title("Isometric (MATLAB view(3) style)")


def draw_cutaway(ax):
    """Quarter-cut isometric of the interior: USB gap, 5 mm top, dial hole."""
    x0, x1 = -WIDTH / 2, 0.0  # keep -X half
    _add_poly(ax, _box_faces(x0, x1, 0, DEPTH, 0, HEIGHT), MATLAB_BLUE, 0.22)
    # inner cavity walls (cut face)
    inner_x0 = -WIDTH / 2 + WALL
    faces = [
        # cut plane at x=0 looking into cavity
        [
            (0, WALL, WALL),
            (0, DEPTH - WALL, WALL),
            (0, DEPTH - WALL, INNER_TOP),
            (0, WALL, INNER_TOP),
        ],
        # inner top underside
        [
            (inner_x0, WALL, INNER_TOP),
            (0, WALL, INNER_TOP),
            (0, DEPTH - WALL, INNER_TOP),
            (inner_x0, DEPTH - WALL, INNER_TOP),
        ],
    ]
    _add_poly(ax, faces, (0.93, 0.69, 0.13), 0.35)
    # USB band on the cut face
    _circ3(ax, (0, 0, CENTER_Z), DIAL_HOLE_D / 2, (0, 1, 0), MATLAB_ORANGE, lw=2.0)
    ax.plot([-SCREW_X, -SCREW_X], [SCREW_Y, SCREW_Y], [0, HEIGHT], color=MATLAB_BLUE, lw=1.4)
    ax.plot([0, 0], [DEPTH, DEPTH], [CABLE_Z - CABLE_SLOT_H / 2, CABLE_Z + CABLE_SLOT_H / 2], color="k", lw=2)
    ax.text(2, SCREW_Y, HEIGHT + 3, "M4 through 5 mm top, 13 mm boss", color=MATLAB_BLUE, fontsize=8)
    ax.view_init(elev=18, azim=-40)
    ax.set_box_aspect((WIDTH / 2, DEPTH, HEIGHT))
    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_zlabel("Z")
    ax.set_title("Cutaway  (-X half)  interior USB gap")


def main():
    _matlab_rc()
    fig = plt.figure(figsize=(16.5, 10.5), dpi=130)
    fig.suptitle(
        "M5Stack Dial under-shelf case   "
        f"{WIDTH:.0f} x {DEPTH:.1f} x {HEIGHT:.0f} mm    "
        "walls 3 mm   top 5 mm   50x80% hatch   LED holes flush, +/-15 / 30 down",
        fontsize=13,
        fontweight="medium",
        y=0.98,
    )
    gs = fig.add_gridspec(2, 3, left=0.05, right=0.98, top=0.92, bottom=0.05, wspace=0.28, hspace=0.32)
    ax_front = fig.add_subplot(gs[0, 0])
    ax_top = fig.add_subplot(gs[0, 1])
    ax_right = fig.add_subplot(gs[0, 2])
    ax_back = fig.add_subplot(gs[1, 0])
    ax_iso = fig.add_subplot(gs[1, 1])
    ax_cut = fig.add_subplot(gs[1, 2])
    draw_front(ax_front)
    draw_top(ax_top)
    draw_right(ax_right)
    draw_back(ax_back)
    for ax, name, title in (
        (ax_iso, "m5dial_shelf_case_iso.png", "Isometric  (OpenSCAD)"),
        (ax_cut, "m5dial_shelf_case_backiso.png", "Rear isometric  — 3 mm pedestal, 8.2 mm @ ±15°"),
    ):
        img_path = OUT / name
        if img_path.exists():
            ax.imshow(plt.imread(img_path))
        ax.set_title(title)
        ax.axis("off")
    png = OUT / "m5dial_shelf_case_drawing.png"
    fig.savefig(png, dpi=150)
    print(f"wrote {png}")
    return png


if __name__ == "__main__":
    main()
