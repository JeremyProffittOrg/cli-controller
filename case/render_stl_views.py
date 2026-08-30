"""Render consistent multi-angle documentation images from the checked-in STL files."""

from __future__ import annotations

from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import trimesh
from mpl_toolkits.mplot3d.art3d import Poly3DCollection


ROOT = Path(__file__).resolve().parent
BACKGROUND = np.array([7, 12, 22], dtype=float) / 255.0
CYAN = np.array([34, 211, 238], dtype=float) / 255.0
TEXT = np.array([226, 232, 240], dtype=float) / 255.0
MUTED = np.array([148, 163, 184], dtype=float) / 255.0


MODELS = {
    "m5dial_shelf_case": {
        "title": "M5Stack Dial under-shelf case",
        "rotate_x": 0,
        "views": {
            "m5dial_shelf_case_iso.png": (24, -48, "Isometric"),
            "m5dial_shelf_case_backiso.png": (22, 132, "Rear isometric"),
            "m5dial_shelf_case_front.png": (0, -90, "Front"),
            "m5dial_shelf_case_back.png": (0, 90, "Rear"),
            "m5dial_shelf_case_top.png": (90, -90, "Top"),
            "m5dial_shelf_case_bottom.png": (-90, -90, "Bottom"),
            "m5dial_shelf_case_side.png": (12, 0, "Right side"),
        },
    },
    "vl53l4cd_case": {
        "title": "Adafruit VL53L4CD sensor case",
        "rotate_x": 90,
        "views": {
            "vl53l4cd_case_stl_iso.png": (24, -48, "Isometric"),
            "vl53l4cd_case_stl_rear_iso.png": (22, 132, "Rear isometric"),
            "vl53l4cd_case_stl_front.png": (0, -90, "Front"),
            "vl53l4cd_case_stl_back.png": (0, 90, "Rear"),
            "vl53l4cd_case_stl_top.png": (90, -90, "Top"),
            "vl53l4cd_case_stl_bottom.png": (-90, -90, "Bottom"),
            "vl53l4cd_case_stl_side.png": (12, 0, "Right side"),
        },
    },
}


def face_colors(mesh: trimesh.Trimesh) -> np.ndarray:
    light = np.array([0.35, -0.45, 0.82], dtype=float)
    light /= np.linalg.norm(light)
    strength = 0.34 + 0.66 * np.clip(mesh.face_normals @ light, 0.0, 1.0)
    colors = CYAN[None, :] * strength[:, None]
    colors += np.array([0.015, 0.025, 0.045])[None, :]
    return np.clip(colors, 0.0, 1.0)


def render(mesh: trimesh.Trimesh, title: str, elev: float, azim: float, label: str, output: Path) -> None:
    fig = plt.figure(figsize=(10, 7.5), dpi=140, facecolor=BACKGROUND)
    ax = fig.add_subplot(111, projection="3d", facecolor=BACKGROUND)
    triangles = mesh.vertices[mesh.faces]
    poly = Poly3DCollection(
        triangles,
        facecolors=face_colors(mesh),
        edgecolors=(0.02, 0.07, 0.10, 0.28),
        linewidths=0.18,
        antialiased=True,
    )
    ax.add_collection3d(poly)

    bounds = mesh.bounds
    center = bounds.mean(axis=0)
    span = bounds[1] - bounds[0]
    radius = max(span) * 0.60
    ax.set_xlim(center[0] - radius, center[0] + radius)
    ax.set_ylim(center[1] - radius, center[1] + radius)
    ax.set_zlim(center[2] - radius, center[2] + radius)
    ax.set_box_aspect((1, 1, 1))
    ax.set_proj_type("ortho")
    ax.view_init(elev=elev, azim=azim)
    ax.set_axis_off()

    size = " x ".join(f"{value:.1f}" for value in span)
    fig.text(0.055, 0.935, title, color=TEXT, fontsize=22, fontweight="bold", ha="left", va="top")
    fig.text(0.055, 0.885, f"STL VIEW  /  {label.upper()}", color=CYAN, fontsize=11, fontweight="bold", ha="left")
    fig.text(0.945, 0.935, f"{size} mm", color=MUTED, fontsize=10, ha="right", va="top")
    fig.text(0.945, 0.055, f"{len(mesh.faces):,} triangles", color=MUTED, fontsize=9, ha="right")
    fig.subplots_adjust(left=0.01, right=0.99, bottom=0.01, top=0.99)
    fig.savefig(output, facecolor=BACKGROUND, bbox_inches="tight", pad_inches=0.08)
    plt.close(fig)
    print(f"wrote {output}")


def main() -> None:
    for stem, spec in MODELS.items():
        path = ROOT / f"{stem}.stl"
        mesh = trimesh.load_mesh(path, process=False)
        if not isinstance(mesh, trimesh.Trimesh):
            raise TypeError(f"{path} did not load as one triangle mesh")
        if spec["rotate_x"]:
            mesh.apply_transform(
                trimesh.transformations.rotation_matrix(
                    np.radians(spec["rotate_x"]), (1.0, 0.0, 0.0)
                )
            )
        for filename, (elev, azim, label) in spec["views"].items():
            render(mesh, spec["title"], elev, azim, label, ROOT / filename)


if __name__ == "__main__":
    main()
