"""MATLAB-style mechanical drawing of the Adafruit VL53L4CD (5396) ToF case.

Board numbers from Adafruit Eagle (Adafruit-VL53L4CD-PCB). Units: millimetres.
"""
from __future__ import annotations

from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
from matplotlib.patches import Arc, FancyBboxPatch, Rectangle, Wedge
from mpl_toolkits.mplot3d.art3d import Poly3DCollection

# --- Adafruit VL53L4CD breakout (product 5396) ---
PCB_X = 25.40
PCB_Z = 17.78
PCB_T = 1.60
HOLE_SPAN_X = 20.32
HOLE_SPAN_Z = 12.70
PCB_HOLE_D = 2.50
SENSOR_X = 4.40
SENSOR_Z = 2.40

# --- case ---
WALL = 3.00
FRONT_T = 3.00
FIT = 0.50
OUTER_X = 39.00
DEPTH = 12.00
CUT = 3.00
CUT_R = 0.60
CUT_BASE_R = 1.50
CUT_Y0 = 9.00  # from the front; the 3x3 sits on the back (print Z=0)
WINDOW_S = 10.00
WINDOW_R = 2.00
BOSS_H = 3.00
BOSS_OD = 5.00
BOSS_HOLE = 2.00
BOSS_HOLE_DEPTH = 4.00
WING_T = 4.00
WING_R = 12.00
WING_HOLE_D = 4.00
WING_HOLE_Y = 6.00  # centred in the 12 mm wing
WING_FILLET = 3.00
# drawing Y=0 is the FRONT; back (open, print bed) is Y=DEPTH
WING_Y0 = DEPTH - WING_T

INNER_X = OUTER_X - 2 * WALL
INNER_Z = PCB_Z + 2 * FIT
OUTER_Z = INNER_Z + 2 * WALL
MARGIN = (INNER_X - PCB_X) / 2

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
    t = np.linspace(0, 2 * np.pi, 64)
    ax.plot(xy[0] + r * np.cos(t), xy[1] + r * np.sin(t), **kw)


def _rect(ax, xy, w, h, **kw):
    ax.add_patch(Rectangle(xy, w, h, fill=False, **kw))


def _rrect(ax, xy, w, h, r, **kw):
    ax.add_patch(
        FancyBboxPatch(
            xy,
            w,
            h,
            boxstyle=f"round,pad=0,rounding_size={r}",
            mutation_aspect=1,
            fill=False,
            **kw,
        )
    )


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
        y + 1.4 * side,
        text,
        ha="center",
        va="bottom" if side > 0 else "top",
        fontsize=8,
        color="k",
        bbox=dict(boxstyle="square,pad=0.12", fc="white", ec="none"),
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
        x + 1.6 * side,
        (y1 + y2) / 2,
        text,
        ha="left" if side > 0 else "right",
        va="center",
        fontsize=8,
        color="k",
        bbox=dict(boxstyle="square,pad=0.12", fc="white", ec="none"),
        clip_on=False,
    )


def draw_front(ax):
    ox, oz = OUTER_X / 2, OUTER_Z / 2
    _rect(ax, (-ox, -oz), OUTER_X, OUTER_Z, ec="k", lw=1.5)
    _rect(ax, (-INNER_X / 2, -INNER_Z / 2), INNER_X, INNER_Z, ec=(0.55, 0.55, 0.55), lw=0.8, ls=":")
    _rect(ax, (-PCB_X / 2, -PCB_Z / 2), PCB_X, PCB_Z, ec=(0.25, 0.25, 0.25), lw=0.9, ls="--")
    _rrect(ax, (-WINDOW_S / 2, -WINDOW_S / 2), WINDOW_S, WINDOW_S, WINDOW_R, ec=MATLAB_ORANGE, lw=1.8)
    _rect(ax, (-SENSOR_X / 2, -SENSOR_Z / 2), SENSOR_X, SENSOR_Z, ec=MATLAB_ORANGE, lw=0.7, ls="--")
    for sx in (-1, 1):
        for sz in (-1, 1):
            _circle(
                ax,
                (sx * HOLE_SPAN_X / 2, sz * HOLE_SPAN_Z / 2),
                BOSS_OD / 2,
                color=MATLAB_BLUE,
                lw=1.0,
                ls=":",
            )
            _circle(
                ax,
                (sx * HOLE_SPAN_X / 2, sz * HOLE_SPAN_Z / 2),
                BOSS_HOLE / 2,
                color=MATLAB_BLUE,
                lw=1.3,
            )
        # half-circle wings
        wedge = Wedge(
            (0, sx * oz),
            WING_R,
            0 if sx > 0 else 180,
            180 if sx > 0 else 360,
            facecolor="none",
            edgecolor=MATLAB_PURPLE,
            lw=1.6,
        )
        ax.add_patch(wedge)
        _circle(ax, (0, sx * (oz + WING_HOLE_Y)), WING_HOLE_D / 2, color=MATLAB_PURPLE, lw=1.5)
        # convex root rounds going out
        _circle(ax, (WING_R, sx * oz), WING_FILLET, color=MATLAB_PURPLE, lw=1.0)
        _circle(ax, (-WING_R, sx * oz), WING_FILLET, color=MATLAB_PURPLE, lw=1.0)
    ax.text(0, 0, "window", ha="center", va="center", fontsize=8, color=MATLAB_ORANGE)
    ax.text(
        0,
        oz + WING_R + 3.5,
        "back wings  r=12  t=4   4 mm hole centred  outward root r=3",
        ha="center",
        fontsize=8,
        color=MATLAB_PURPLE,
        clip_on=False,
    )
    _dim_h(ax, -ox, ox, oz + WING_R + 8, f"{OUTER_X:.0f}")
    _dim_h(ax, -PCB_X / 2, PCB_X / 2, -oz - 5, f"PCB  {PCB_X:.2f}", side=-1)
    _dim_v(ax, -ox - 7, -oz, oz, f"{OUTER_Z:.2f}", side=-1)
    _dim_v(ax, ox + 6, -WINDOW_S / 2, WINDOW_S / 2, "10 sq")
    _dim_v(ax, 8, oz, oz + WING_R, "12")
    ax.set_title("Front  (X-Z)  sensor face")
    ax.set_xlabel("X")
    ax.set_ylabel("Z")
    ax.set_aspect("equal")
    ax.set_xlim(-ox - 16, ox + 18)
    ax.set_ylim(-oz - WING_R - 10, oz + WING_R + 14)


def draw_top(ax):
    ox = OUTER_X / 2
    _rect(ax, (-ox, 0), OUTER_X, DEPTH, ec="k", lw=1.5)
    _rect(
        ax,
        (-INNER_X / 2, FRONT_T),
        INNER_X,
        DEPTH - FRONT_T,
        ec=(0.55, 0.55, 0.55),
        lw=0.8,
        ls=":",
    )
    ax.plot([-ox, ox], [FRONT_T, FRONT_T], color=MATLAB_BLUE, lw=1.0, ls="--")
    ax.plot([-ox, ox], [WING_Y0, WING_Y0], color=MATLAB_PURPLE, lw=1.0, ls=":")
    _rect(ax, (-PCB_X / 2, FRONT_T + BOSS_H), PCB_X, PCB_T, ec=(0.25, 0.25, 0.25), lw=0.9, ls="--")
    _rrect(ax, (-WINDOW_S / 2, -0.15), WINDOW_S, 0.3, 0.05, ec=MATLAB_ORANGE, lw=2.0)
    for s in (-1, 1):
        ax.add_patch(
            Rectangle(
                (s * ox - (WALL if s > 0 else 0), CUT_Y0),
                WALL,
                CUT,
                fill=True,
                fc=(0.80, 0.93, 0.70),
                ec=MATLAB_GREEN,
                lw=1.3,
            )
        )
        _circle(ax, (s * ox, DEPTH), CUT_BASE_R, color=MATLAB_GREEN, lw=1.1)
        _circle(ax, (s * HOLE_SPAN_X / 2, FRONT_T + BOSS_H / 2), BOSS_OD / 2, color=MATLAB_BLUE, lw=1.0)
        _circle(ax, (s * HOLE_SPAN_X / 2, FRONT_T + BOSS_H / 2), BOSS_HOLE / 2, color=MATLAB_BLUE, lw=1.1)
    ax.text(0, DEPTH + 1.4, "OPEN BACK", ha="center", fontsize=8)
    ax.text(
        0,
        CUT_Y0 - 1.6,
        "3 mm cutout on the base (z=0), flare out r=1.5",
        ha="center",
        fontsize=8,
        color=MATLAB_GREEN,
    )
    _dim_h(ax, -ox, ox, -4.5, f"{OUTER_X:.0f}")
    _dim_v(ax, ox + 6, 0, DEPTH, f"{DEPTH:.0f}")
    _dim_v(ax, -ox - 6, 0, FRONT_T, f"{FRONT_T:.0f}", side=-1)
    _dim_v(ax, ox + 12, CUT_Y0, CUT_Y0 + CUT, "3")
    ax.set_title("Top  (X-Y)  front at Y=0, back open")
    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_aspect("equal")
    ax.set_xlim(-ox - 16, ox + 20)
    ax.set_ylim(-8, DEPTH + 8)


def draw_right(ax):
    oz = OUTER_Z / 2
    _rect(ax, (0, -oz), DEPTH, OUTER_Z, ec="k", lw=1.5)
    ax.plot([FRONT_T, FRONT_T], [-oz, oz], color=MATLAB_BLUE, lw=1.0, ls="--")
    # wings on the BACK (Y = 8 to 12)
    _rect(ax, (WING_Y0, oz), WING_T, WING_R, ec=MATLAB_PURPLE, lw=1.4)
    _rect(ax, (WING_Y0, -oz - WING_R), WING_T, WING_R, ec=MATLAB_PURPLE, lw=1.4)
    _circle(ax, (WING_Y0 + WING_T / 2, oz + WING_HOLE_Y), WING_HOLE_D / 2, color=MATLAB_PURPLE, lw=1.4)
    _circle(ax, (WING_Y0 + WING_T / 2, -oz - WING_HOLE_Y), WING_HOLE_D / 2, color=MATLAB_PURPLE, lw=1.4)
    # outward round at wall-to-wing
    _circle(ax, (WING_Y0, oz), WING_FILLET, color=MATLAB_PURPLE, lw=1.0, ls=":")
    _circle(ax, (WING_Y0, -oz), WING_FILLET, color=MATLAB_PURPLE, lw=1.0, ls=":")
    _rect(ax, (FRONT_T, -PCB_Z / 2), PCB_T, PCB_Z, ec=(0.25, 0.25, 0.25), lw=0.9, ls="--")
    # bosses 3 mm tall from inner front
    for s in (-1, 1):
        _rect(
            ax,
            (FRONT_T, s * HOLE_SPAN_Z / 2 - BOSS_OD / 2),
            BOSS_H,
            BOSS_OD,
            ec=MATLAB_BLUE,
            lw=1.2,
        )
        ax.plot(
            [FRONT_T + BOSS_H - BOSS_HOLE_DEPTH, FRONT_T + BOSS_H],
            [s * HOLE_SPAN_Z / 2, s * HOLE_SPAN_Z / 2],
            color=MATLAB_BLUE,
            lw=1.6,
        )
    _rrect(ax, (CUT_Y0, -CUT / 2), CUT, CUT, CUT_R, ec=MATLAB_GREEN, lw=1.7)
    _circle(ax, (DEPTH, -CUT / 2), CUT_BASE_R, color=MATLAB_GREEN, lw=1.2)
    _circle(ax, (DEPTH,  CUT / 2), CUT_BASE_R, color=MATLAB_GREEN, lw=1.2)
    ax.text(
        CUT_Y0 + CUT / 2,
        -CUT / 2 - 3.2,
        "3x3 on z=0  outward at base",
        ha="center",
        fontsize=8,
        color=MATLAB_GREEN,
    )
    ax.plot([DEPTH, DEPTH], [-oz, oz], color="k", lw=1.0, ls="--")
    ax.text(DEPTH + 1.0, 0, "OPEN", rotation=90, ha="center", va="bottom", fontsize=8)
    _dim_v(ax, -7, -oz, oz, f"{OUTER_Z:.2f}", side=-1)
    _dim_h(ax, 0, DEPTH, oz + WING_R + 5, f"{DEPTH:.0f}")
    _dim_v(ax, DEPTH + 6, oz, oz + WING_R, "12")
    _dim_h(ax, WING_Y0, DEPTH, -oz - WING_R - 4, "4", side=-1)
    ax.set_title("Right  (Y-Z)  connector-side wall")
    ax.set_xlabel("Y")
    ax.set_ylabel("Z")
    ax.set_aspect("equal")
    ax.set_xlim(-14, DEPTH + 14)
    ax.set_ylim(-oz - WING_R - 10, oz + WING_R + 10)


def draw_back(ax):
    ox, oz = OUTER_X / 2, OUTER_Z / 2
    _rect(ax, (-ox, -oz), OUTER_X, OUTER_Z, ec="k", lw=1.5)
    _rect(ax, (-INNER_X / 2, -INNER_Z / 2), INNER_X, INNER_Z, ec=(0.55, 0.55, 0.55), lw=1.0, ls=":")
    _rrect(ax, (-WINDOW_S / 2, -WINDOW_S / 2), WINDOW_S, WINDOW_S, WINDOW_R, ec=MATLAB_ORANGE, lw=1.6)
    for sx in (-1, 1):
        for sz in (-1, 1):
            _circle(
                ax,
                (sx * HOLE_SPAN_X / 2, sz * HOLE_SPAN_Z / 2),
                BOSS_OD / 2,
                color=MATLAB_BLUE,
                lw=1.1,
            )
            _circle(
                ax,
                (sx * HOLE_SPAN_X / 2, sz * HOLE_SPAN_Z / 2),
                BOSS_HOLE / 2,
                color=MATLAB_BLUE,
                lw=1.3,
            )
        wedge = Wedge(
            (0, sx * oz),
            WING_R,
            0 if sx > 0 else 180,
            180 if sx > 0 else 360,
            facecolor="none",
            edgecolor=MATLAB_PURPLE,
            lw=1.4,
        )
        ax.add_patch(wedge)
        _circle(ax, (0, sx * (oz + WING_HOLE_Y)), WING_HOLE_D / 2, color=MATLAB_PURPLE, lw=1.4)
        _rrect(
            ax,
            (sx * (ox - WALL / 2) - CUT / 2, -CUT / 2),
            CUT,
            CUT,
            CUT_R,
            ec=MATLAB_GREEN,
            lw=1.2,
        )
    ax.text(0, 0, "open back", ha="center", va="center", fontsize=8)
    ax.text(
        0,
        -oz - 4,
        "bosses 3 mm tall, 2 mm hole x 4 mm deep, not through the front",
        ha="center",
        fontsize=8,
        color=MATLAB_BLUE,
        clip_on=False,
    )
    ax.set_title("Back  (X-Z)  looking in")
    ax.set_xlabel("X")
    ax.set_ylabel("Z")
    ax.set_aspect("equal")
    ax.set_xlim(-ox - 8, ox + 8)
    ax.set_ylim(-oz - WING_R - 8, oz + WING_R + 8)


def _box_faces(x0, x1, y0, y1, z0, z1):
    return [
        [(x0, y0, z0), (x1, y0, z0), (x1, y0, z1), (x0, y0, z1)],
        [(x0, y1, z0), (x1, y1, z0), (x1, y1, z1), (x0, y1, z1)],
        [(x0, y0, z0), (x0, y1, z0), (x0, y1, z1), (x0, y0, z1)],
        [(x1, y0, z0), (x1, y1, z0), (x1, y1, z1), (x1, y0, z1)],
        [(x0, y0, z1), (x1, y0, z1), (x1, y1, z1), (x0, y1, z1)],
        [(x0, y0, z0), (x1, y0, z0), (x1, y1, z0), (x0, y1, z0)],
    ]


def _add_poly(ax, faces, color, alpha):
    poly = Poly3DCollection(
        faces, facecolors=[color] * len(faces), edgecolors="k", linewidths=0.55, alpha=alpha
    )
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


def _semi3(ax, center, radius, normal, out, color, lw=1.6, n=32):
    nrm = np.asarray(normal, dtype=float)
    nrm = nrm / np.linalg.norm(nrm)
    outv = np.asarray(out, dtype=float)
    outv = outv / np.linalg.norm(outv)
    a = np.cross(nrm, outv)
    a = a / np.linalg.norm(a)
    t = np.linspace(-np.pi / 2, np.pi / 2, n)
    p = np.asarray(center) + radius * (np.outer(np.cos(t), outv) + np.outer(np.sin(t), a))
    ax.plot(p[:, 0], p[:, 1], p[:, 2], color=color, lw=lw)


def _rounded_square3(ax, origin, size, radius, normal, color, lw=1.6):
    nrm = np.asarray(normal, dtype=float)
    nrm = nrm / np.linalg.norm(nrm)
    a = np.cross(nrm, (0.0, 0.0, 1.0) if abs(nrm[2]) < 0.9 else (0.0, 1.0, 0.0))
    a = a / np.linalg.norm(a)
    b = np.cross(nrm, a)
    inner = size / 2.0 - radius
    segs = []
    for i, (cx, cy) in enumerate([(inner, inner), (-inner, inner), (-inner, -inner), (inner, -inner)]):
        t0 = i * np.pi / 2
        for ang in np.linspace(t0, t0 + np.pi / 2, 10):
            segs.append((cx + radius * np.cos(ang), cy + radius * np.sin(ang)))
    p = np.asarray(origin) + np.array([aa * a + bb * b for aa, bb in segs])
    p = np.vstack([p, p[0]])
    ax.plot(p[:, 0], p[:, 1], p[:, 2], color=color, lw=lw)


def draw_iso(ax):
    x0, x1 = -OUTER_X / 2, OUTER_X / 2
    z0, z1 = -OUTER_Z / 2, OUTER_Z / 2
    faces = _box_faces(x0, x1, 0, DEPTH, z0, z1)
    _add_poly(ax, [faces[0], faces[2], faces[3], faces[4], faces[5]], MATLAB_BLUE, 0.28)
    # wings on the back
    _add_poly(ax, _box_faces(-WING_R, WING_R, WING_Y0, DEPTH, z1, z1 + WING_R), MATLAB_PURPLE, 0.45)
    _add_poly(ax, _box_faces(-WING_R, WING_R, WING_Y0, DEPTH, z0 - WING_R, z0), MATLAB_PURPLE, 0.45)
    ix, iz = INNER_X / 2, INNER_Z / 2
    ax.plot([-ix, ix, ix, -ix, -ix], [DEPTH] * 5, [-iz, -iz, iz, iz, -iz], color="k", lw=1.1)
    _rounded_square3(ax, (0, 0, 0), WINDOW_S, WINDOW_R, (0, 1, 0), MATLAB_ORANGE, lw=1.8)
    for sx in (-1, 1):
        for sz in (-1, 1):
            _circ3(
                ax,
                (sx * HOLE_SPAN_X / 2, FRONT_T + BOSS_H, sz * HOLE_SPAN_Z / 2),
                BOSS_HOLE / 2,
                (0, 1, 0),
                MATLAB_BLUE,
                lw=1.3,
            )
            _circ3(
                ax,
                (sx * HOLE_SPAN_X / 2, FRONT_T + BOSS_H, sz * HOLE_SPAN_Z / 2),
                BOSS_OD / 2,
                (0, 1, 0),
                MATLAB_BLUE,
                lw=1.0,
            )
        _rounded_square3(
            ax,
            (sx * OUTER_X / 2, CUT_Y0 + CUT / 2, 0),
            CUT,
            CUT_R,
            (1, 0, 0),
            MATLAB_GREEN,
            lw=1.6,
        )
        _circ3(
            ax,
            (0, DEPTH, sx * (OUTER_Z / 2 + WING_HOLE_Y)),
            WING_HOLE_D / 2,
            (0, 1, 0),
            MATLAB_PURPLE,
            lw=1.5,
        )
        _semi3(ax, (0, DEPTH, sx * OUTER_Z / 2), WING_R, (0, 1, 0), (0, 0, sx), MATLAB_PURPLE, lw=1.6)
    ax.view_init(elev=22, azim=-52)
    span_z = OUTER_Z + 2 * WING_R
    ax.set_box_aspect((max(OUTER_X, 2 * WING_R), DEPTH, span_z))
    ax.set_xlabel("X")
    ax.set_ylabel("Y  (back open)")
    ax.set_zlabel("Z")
    ax.set_title("Isometric")


def draw_notes(ax):
    ax.axis("off")
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.set_title("Numbers")
    text = (
        "Board  Adafruit VL53L4CD  5396\n"
        f"  PCB     {PCB_X:.2f} x {PCB_Z:.2f} mm\n"
        f"  holes   4 x {PCB_HOLE_D:.1f} mm at {HOLE_SPAN_X:.2f} x {HOLE_SPAN_Z:.2f}\n"
        "\n"
        "Case\n"
        f"  outer X     {OUTER_X:.0f} mm  (wings attach here)\n"
        f"  depth       {DEPTH:.0f} mm, open back\n"
        f"  outer Z     {OUTER_Z:.2f} mm  + {WING_R:.0f} mm wings each side\n"
        f"  window      {WINDOW_S:.0f} mm rounded square, r={WINDOW_R:.0f}\n"
        f"  bosses      4 x {BOSS_H:.0f} mm tall, OD {BOSS_OD:.0f}\n"
        f"              {BOSS_HOLE:.0f} mm hole x {BOSS_HOLE_DEPTH:.0f} mm deep\n"
        "              does not break the front face\n"
        f"  QT cutout   {CUT:.0f} x {CUT:.0f} mm on the base (print z=0)\n"
        f"              outward rounds r={CUT_BASE_R:.1f} where they meet the bed\n"
        f"  wings       on BACK, half-circle r={WING_R:.0f}, t={WING_T:.0f}\n"
        f"              4 mm hole centred ({WING_HOLE_Y:.0f} mm out)\n"
        f"              root rounds going out r={WING_FILLET:.0f}\n"
        "  print       back flat on the bed"
    )
    ax.text(
        0.04,
        0.96,
        text,
        ha="left",
        va="top",
        fontsize=9,
        family="DejaVu Sans Mono",
        transform=ax.transAxes,
    )


def main():
    _matlab_rc()
    fig = plt.figure(figsize=(16.5, 10.5), dpi=130)
    fig.suptitle(
        "Adafruit VL53L4CD (5396) ToF case   "
        f"{OUTER_X:.0f} x {DEPTH:.0f} x {OUTER_Z:.2f} mm + 12 mm back wings   "
        "through side cutouts   3 mm bosses   print back-down",
        fontsize=12,
        fontweight="medium",
        y=0.98,
    )
    gs = fig.add_gridspec(2, 3, left=0.05, right=0.98, top=0.91, bottom=0.05, wspace=0.28, hspace=0.32)
    ax_front = fig.add_subplot(gs[0, 0])
    ax_top = fig.add_subplot(gs[0, 1])
    ax_right = fig.add_subplot(gs[0, 2])
    ax_back = fig.add_subplot(gs[1, 0])
    ax_iso = fig.add_subplot(gs[1, 1], projection="3d")
    ax_notes = fig.add_subplot(gs[1, 2])
    draw_front(ax_front)
    draw_top(ax_top)
    draw_right(ax_right)
    draw_back(ax_back)
    draw_iso(ax_iso)
    draw_notes(ax_notes)
    png = OUT / "vl53l4cd_case_drawing.png"
    fig.savefig(png, dpi=150)
    print(f"wrote {png}")
    return png


if __name__ == "__main__":
    main()
