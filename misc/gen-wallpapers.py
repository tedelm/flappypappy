"""Generate color variants of pub_wall.png via HSV hue rotation."""

from colorsys import hsv_to_rgb, rgb_to_hsv
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
SRC = ROOT / "src" / "internal" / "game" / "sprite"
SOURCE = SRC / "pub_wall.png"

# Hue shifts (degrees) for variants 2..6 — tuned for distinct pub themes.
HUE_SHIFTS = {
    2: 95,   # forest green
    3: 200,  # deep navy blue
    4: 280,  # royal purple
    5: 25,   # burnt rust/orange
    6: 160,  # teal
}


def rotate_hue(img: Image.Image, degrees: float) -> Image.Image:
    shift = degrees / 360.0
    rgba = img.convert("RGBA")
    out = Image.new("RGBA", rgba.size)
    px_in = rgba.load()
    px_out = out.load()
    w, h = rgba.size

    for y in range(h):
        for x in range(w):
            r, g, b, a = px_in[x, y]
            if a == 0:
                px_out[x, y] = (r, g, b, a)
                continue
            h_, s, v = rgb_to_hsv(r / 255.0, g / 255.0, b / 255.0)
            h_ = (h_ + shift) % 1.0
            nr, ng, nb = hsv_to_rgb(h_, s, v)
            px_out[x, y] = (
                int(nr * 255),
                int(ng * 255),
                int(nb * 255),
                a,
            )
    return out


def main() -> None:
    if not SOURCE.exists():
        raise SystemExit(f"source not found: {SOURCE}")

    base = Image.open(SOURCE)
    for num, degrees in HUE_SHIFTS.items():
        out_path = SRC / f"pub_wall_{num}.png"
        variant = rotate_hue(base, degrees)
        variant.save(out_path)
        print(f"wrote {out_path} (hue +{degrees}°)")


if __name__ == "__main__":
    main()
