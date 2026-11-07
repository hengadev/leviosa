"""Generate the favicons and PWA icons into static/: a rounded "L" on a
rounded square, drawn like germinal's "G" (same stroke, same colours).

    python3 scripts/generate-icons.py   (from frontend/, needs Pillow)
"""

from pathlib import Path

from PIL import Image, ImageDraw

DARK = "#1a1a1a"
WHITE = "#ffffff"
# The staff app keeps the slate of its manifest theme colour.
SLATE = "#475569"

# Drawn this many times larger, then scaled down, for smooth edges.
SUPERSAMPLE = 8

STATIC = Path(__file__).resolve().parent.parent / "static"


def icon(size: int, background: str, letter: str) -> Image.Image:
    s = size * SUPERSAMPLE
    image = Image.new("RGBA", (s, s), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)
    draw.rounded_rectangle((0, 0, s - 1, s - 1), radius=s * 0.15, fill=background)

    # The "L": down then right, with round caps and a round corner.
    width = s * 0.08
    top, bottom = (s * 0.35, s * 0.22), (s * 0.35, s * 0.78)
    right = (s * 0.66, s * 0.78)
    draw.line([top, bottom, right], fill=letter, width=round(width), joint="curve")
    r = width / 2
    for x, y in (top, bottom, right):
        draw.ellipse((x - r, y - r, x + r, y + r), fill=letter)

    return image.resize((size, size), Image.LANCZOS)


def main() -> None:
    for size in (192, 512):
        icon(size, WHITE, DARK).save(STATIC / f"pwa-{size}x{size}.png")
        icon(size, DARK, WHITE).save(STATIC / f"pwa-admin-{size}x{size}.png")
        icon(size, SLATE, WHITE).save(STATIC / f"pwa-staff-{size}x{size}.png")

    for size in (16, 32):
        icon(size, WHITE, DARK).save(STATIC / f"favicon-{size}x{size}.png")
    icon(48, WHITE, DARK).save(STATIC / "favicon.ico", sizes=[(16, 16), (32, 32), (48, 48)])

    print(f"Icons written to {STATIC}")


if __name__ == "__main__":
    main()
