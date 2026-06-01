from __future__ import annotations

import math
import random
from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter, ImageFont


ROOT = Path(__file__).resolve().parents[1]
FONT_BOLD = "/System/Library/Fonts/Supplemental/Verdana Bold.ttf"

FIELD_LIGHT = "#cfe88f"
FIELD_MID = "#a8d76a"
FIELD_GREEN = "#07833f"
DEEP_GREEN = "#036735"
WHITE = "#ffffff"


def font(size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(FONT_BOLD, size)


def draw_pitch_background(size: int) -> Image.Image:
    img = Image.new("RGB", (size, size), FIELD_LIGHT)
    overlay = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(overlay)

    for y in range(size):
        ratio = y / max(1, size - 1)
        r = int(207 * (1 - ratio) + 168 * ratio)
        g = int(232 * (1 - ratio) + 215 * ratio)
        b = int(143 * (1 - ratio) + 106 * ratio)
        ImageDraw.Draw(img).line((0, y, size, y), fill=(r, g, b))

    line = max(4, size // 120)
    line_color = (255, 255, 255, 70)
    draw.rectangle((-size * 0.08, -size * 0.03, size * 0.42, size * 0.3), outline=line_color, width=line)
    draw.arc((size * 0.08, size * 0.12, size * 0.44, size * 0.48), 5, 185, fill=line_color, width=line)
    draw.rectangle((size * 0.78, size * 0.28, size * 1.18, size * 0.64), outline=line_color, width=line)
    draw.arc((size * 0.62, size * 0.44, size * 1.03, size * 0.83), 185, 355, fill=line_color, width=line)
    draw.line((-size * 0.08, size * 0.78, size * 0.38, size * 0.54), fill=line_color, width=line)
    draw.arc((-size * 0.05, size * 0.82, size * 0.38, size * 1.18), 190, 345, fill=line_color, width=line)

    dash_color = (255, 255, 255, 75)
    for start, end in [
        ((size * 0.02, size * 0.42), (size * 0.28, size * 0.35)),
        ((size * 0.58, size * 0.16), (size * 0.86, size * 0.11)),
        ((size * 0.62, size * 0.7), (size * 0.84, size * 0.88)),
        ((size * 0.06, size * 0.76), (size * 0.22, size * 0.62)),
    ]:
        draw_dashed_arrow(draw, start, end, dash_color, max(3, size // 160))

    img = Image.alpha_composite(img.convert("RGBA"), overlay)

    noise = Image.new("L", (size, size))
    rng = random.Random(42)
    noise.putdata([rng.randrange(0, 42) for _ in range(size * size)])
    noise = noise.filter(ImageFilter.GaussianBlur(radius=max(1, size // 380)))
    texture = Image.new("RGBA", (size, size), (255, 255, 255, 0))
    texture.putalpha(noise)
    img = Image.alpha_composite(img, texture)
    return img.convert("RGB")


def draw_dashed_arrow(
    draw: ImageDraw.ImageDraw,
    start: tuple[float, float],
    end: tuple[float, float],
    fill: tuple[int, int, int, int],
    width: int,
) -> None:
    sx, sy = start
    ex, ey = end
    dx = ex - sx
    dy = ey - sy
    length = math.hypot(dx, dy)
    if length == 0:
        return
    ux = dx / length
    uy = dy / length
    dash = length / 7
    gap = dash * 0.7
    pos = 0.0
    while pos < length - dash:
        x1 = sx + ux * pos
        y1 = sy + uy * pos
        x2 = sx + ux * min(length, pos + dash)
        y2 = sy + uy * min(length, pos + dash)
        draw.line((x1, y1, x2, y2), fill=fill, width=width)
        pos += dash + gap

    angle = math.atan2(dy, dx)
    head = width * 7
    p1 = (ex, ey)
    p2 = (ex - head * math.cos(angle - 0.55), ey - head * math.sin(angle - 0.55))
    p3 = (ex - head * math.cos(angle + 0.55), ey - head * math.sin(angle + 0.55))
    draw.line((p2, p1, p3), fill=fill, width=width)


def logo_layer(size: int) -> Image.Image:
    layer = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(layer)

    p_font = font(int(size * 0.56))
    text = "P"
    bbox = draw.textbbox((0, 0), text, font=p_font, stroke_width=0)
    x = (size - (bbox[2] - bbox[0])) / 2 - bbox[0] - size * 0.01
    y = (size - (bbox[3] - bbox[1])) / 2 - bbox[1] - size * 0.015

    shadow = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    shadow_draw = ImageDraw.Draw(shadow)
    shadow_draw.text(
        (x + size * 0.018, y + size * 0.024),
        text,
        font=p_font,
        fill=(0, 0, 0, 70),
        stroke_width=max(8, size // 56),
        stroke_fill=(0, 0, 0, 70),
    )
    shadow = shadow.filter(ImageFilter.GaussianBlur(radius=max(4, size // 80)))
    layer = Image.alpha_composite(layer, shadow)
    draw = ImageDraw.Draw(layer)

    draw.text(
        (x, y),
        text,
        font=p_font,
        fill=FIELD_GREEN,
        stroke_width=max(8, size // 54),
        stroke_fill=WHITE,
    )

    # Lower diagonal line, matching the old splash mark's football-field cue.
    line_width = max(8, size // 42)
    draw.line(
        (size * 0.28, size * 0.66, size * 0.52, size * 0.38),
        fill=WHITE,
        width=line_width,
    )
    draw.line(
        (size * 0.295, size * 0.655, size * 0.51, size * 0.405),
        fill=FIELD_GREEN,
        width=max(4, size // 78),
    )

    draw_soccer_ball(draw, (int(size * 0.57), int(size * 0.43)), int(size * 0.098))
    draw.ellipse(
        (
            size * 0.34,
            size * 0.64,
            size * 0.43,
            size * 0.73,
        ),
        outline=WHITE,
        width=max(6, size // 70),
    )
    draw.ellipse(
        (
            size * 0.375,
            size * 0.675,
            size * 0.395,
            size * 0.695,
        ),
        fill=WHITE,
    )

    return layer


def draw_soccer_ball(draw: ImageDraw.ImageDraw, center: tuple[int, int], radius: int) -> None:
    cx, cy = center
    draw.ellipse((cx - radius, cy - radius, cx + radius, cy + radius), fill=WHITE)
    pentagon = []
    inner = radius * 0.42
    for i in range(5):
        angle = -math.pi / 2 + i * math.tau / 5
        pentagon.append((cx + math.cos(angle) * inner, cy + math.sin(angle) * inner))
    draw.polygon(pentagon, fill=DEEP_GREEN)
    for px, py in pentagon:
        ex = cx + (px - cx) * 2.1
        ey = cy + (py - cy) * 2.1
        draw.line((px, py, ex, ey), fill=DEEP_GREEN, width=max(2, radius // 10))
    draw.arc((cx - radius, cy - radius, cx + radius, cy + radius), 295, 35, fill=DEEP_GREEN, width=max(3, radius // 8))
    draw.arc((cx - radius, cy - radius, cx + radius, cy + radius), 120, 205, fill=DEEP_GREEN, width=max(3, radius // 8))


def make_logo(size: int) -> Image.Image:
    img = draw_pitch_background(size).convert("RGBA")
    img = Image.alpha_composite(img, logo_layer(size))
    return img.convert("RGB")


def save_all() -> None:
    icon = make_logo(1024)
    splash = make_logo(1240)
    favicon = make_logo(512)
    og = make_logo(1200)

    outputs = {
        "frontend/assets/icon.png": icon,
        "frontend/assets/adaptive-icon.png": icon,
        "frontend/assets/splash-icon.png": icon,
        "frontend/assets/favicon.png": favicon,
        "frontend/assets/splash-palpite.png": splash,
        "landing/public/splash-palpite.png": splash,
        "landing/public/og-image.png": og,
    }

    for rel_path, image in outputs.items():
        out = ROOT / rel_path
        out.parent.mkdir(parents=True, exist_ok=True)
        image.save(out, optimize=True)


if __name__ == "__main__":
    save_all()
