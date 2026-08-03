#!/usr/bin/env python3
"""Generate the first offline local asset pack for visual_designer."""

from __future__ import annotations

import json
import math
import random
from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter


ROOT = Path(__file__).resolve().parents[1]
ASSET_ROOT = ROOT / "assets"
ICON_DIR = ASSET_ROOT / "icons" / "core"
BG_DIR = ASSET_ROOT / "backgrounds" / "editorial"
PATTERN_DIR = ASSET_ROOT / "patterns" / "subtle"

INK = (77, 116, 138, 255)
INK_DARK = (37, 54, 70, 255)
ACCENT = (181, 151, 105, 255)
MUTED = (116, 145, 159, 255)
SOFT = (229, 237, 240, 255)


ICON_TAGS = {
    "runtime": ["runtime", "observability", "status"],
    "timeline": ["timeline", "event", "history"],
    "tool": ["tool", "execution", "operation"],
    "llm": ["llm", "model", "agent"],
    "file": ["file", "artifact", "output"],
    "report": ["report", "summary", "runtime"],
    "contract": ["contract", "schema", "template"],
    "capacity": ["capacity", "density", "layout"],
    "layout": ["layout", "composition", "template"],
    "split": ["split", "overflow", "pagination"],
    "align": ["align", "center", "balance"],
    "density": ["density", "content", "text"],
    "chart": ["chart", "data", "analysis"],
    "kpi": ["kpi", "metric", "dashboard"],
    "trend": ["trend", "growth", "line"],
    "source": ["source", "citation", "data"],
    "warning": ["warning", "risk", "budget"],
    "fix": ["fix", "repair", "continue"],
    "template": ["template", "slide", "layout"],
    "background": ["background", "visual", "hero"],
    "primitive": ["primitive", "shape", "icon"],
    "card": ["card", "grid", "module"],
    "flow": ["flow", "process", "step"],
    "review": ["review", "qa", "check"],
}


def ensure_dirs() -> None:
    for path in [ICON_DIR, BG_DIR, PATTERN_DIR]:
        path.mkdir(parents=True, exist_ok=True)


def draw_icon_base() -> tuple[Image.Image, ImageDraw.ImageDraw]:
    img = Image.new("RGBA", (256, 256), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    return img, draw


def line(draw: ImageDraw.ImageDraw, points, fill=INK, width=12) -> None:
    draw.line(points, fill=fill, width=width, joint="curve")


def rect(draw: ImageDraw.ImageDraw, xy, outline=INK, width=12, radius=22) -> None:
    draw.rounded_rectangle(xy, radius=radius, outline=outline, width=width)


def icon_runtime(draw):
    draw.ellipse((54, 54, 202, 202), outline=INK, width=12)
    line(draw, [(128, 80), (128, 132), (166, 154)], INK_DARK, 12)
    draw.arc((34, 34, 222, 222), 210, 310, fill=ACCENT, width=10)


def icon_timeline(draw):
    line(draw, [(42, 132), (214, 132)], INK, 10)
    for x, y in [(56, 132), (112, 98), (166, 132), (210, 88)]:
        draw.ellipse((x - 16, y - 16, x + 16, y + 16), fill=SOFT, outline=INK_DARK, width=8)
    line(draw, [(112, 114), (112, 176), (178, 176)], ACCENT, 8)


def icon_tool(draw):
    line(draw, [(72, 184), (150, 106)], INK_DARK, 18)
    draw.arc((128, 42, 214, 128), 30, 250, fill=INK, width=16)
    draw.rounded_rectangle((50, 168, 92, 210), radius=10, fill=ACCENT)


def icon_llm(draw):
    rect(draw, (58, 68, 198, 174), INK, 12, 26)
    for x in [88, 128, 168]:
        draw.ellipse((x - 10, 112, x + 10, 132), fill=ACCENT)
    line(draw, [(98, 174), (84, 210), (128, 174)], INK, 10)


def icon_file(draw):
    draw.polygon([(72, 42), (154, 42), (198, 86), (198, 216), (72, 216)], outline=INK, fill=None)
    line(draw, [(154, 42), (154, 88), (198, 88)], INK, 10)
    for y in [122, 154, 186]:
        line(draw, [(96, y), (170, y)], MUTED, 8)


def icon_report(draw):
    rect(draw, (54, 44, 202, 212), INK, 10, 16)
    for i, h in enumerate([62, 92, 42]):
        x = 86 + i * 34
        draw.rounded_rectangle((x, 174 - h, x + 18, 174), radius=5, fill=[INK, ACCENT, MUTED][i])
    line(draw, [(82, 76), (174, 76)], INK_DARK, 8)


def icon_contract(draw):
    rect(draw, (50, 52, 206, 204), INK, 10, 18)
    for y in [92, 126, 160]:
        draw.ellipse((76, y - 8, 92, y + 8), fill=ACCENT)
        line(draw, [(106, y), (176, y)], INK_DARK, 7)


def icon_capacity(draw):
    rect(draw, (48, 70, 208, 186), INK, 10, 16)
    line(draw, [(82, 128), (174, 128)], ACCENT, 12)
    line(draw, [(132, 92), (174, 128), (132, 164)], ACCENT, 12)


def icon_layout(draw):
    rect(draw, (44, 52, 212, 204), INK, 10, 16)
    draw.rounded_rectangle((66, 78, 124, 132), radius=10, fill=SOFT, outline=INK_DARK, width=6)
    draw.rounded_rectangle((138, 78, 190, 176), radius=10, fill=SOFT, outline=ACCENT, width=6)
    draw.rounded_rectangle((66, 146, 124, 176), radius=10, fill=SOFT, outline=MUTED, width=6)


def icon_split(draw):
    line(draw, [(128, 42), (128, 214)], INK, 10)
    line(draw, [(72, 82), (44, 110), (72, 138)], ACCENT, 12)
    line(draw, [(184, 82), (212, 110), (184, 138)], ACCENT, 12)
    rect(draw, (58, 162, 198, 206), MUTED, 8, 14)


def icon_align(draw):
    for i, w in enumerate([132, 86, 132, 66]):
        y = 68 + i * 34
        line(draw, [(64, y), (64 + w, y)], [INK, MUTED, INK, ACCENT][i], 10)
    line(draw, [(128, 46), (128, 212)], INK_DARK, 5)


def icon_density(draw):
    sizes = [14, 10, 18, 8, 12, 16, 10, 14, 8]
    for i, s in enumerate(sizes):
        x = 62 + (i % 3) * 54
        y = 68 + (i // 3) * 54
        draw.rounded_rectangle((x, y, x + s * 2, y + s * 2), radius=6, fill=[INK, MUTED, ACCENT][i % 3])


def icon_chart(draw):
    line(draw, [(58, 194), (204, 194)], INK_DARK, 10)
    line(draw, [(58, 194), (58, 64)], INK_DARK, 10)
    for i, h in enumerate([46, 84, 118]):
        x = 88 + i * 38
        draw.rounded_rectangle((x, 194 - h, x + 24, 194), radius=5, fill=[MUTED, ACCENT, INK][i])


def icon_kpi(draw):
    for i, xy in enumerate([(48, 60, 116, 124), (140, 60, 208, 124), (48, 148, 116, 208), (140, 148, 208, 208)]):
        draw.rounded_rectangle(xy, radius=14, fill=SOFT, outline=[INK, MUTED, ACCENT, INK_DARK][i], width=7)


def icon_trend(draw):
    line(draw, [(54, 182), (98, 142), (136, 154), (202, 78)], INK, 12)
    line(draw, [(170, 78), (202, 78), (202, 112)], ACCENT, 12)
    line(draw, [(54, 204), (210, 204)], INK_DARK, 8)


def icon_source(draw):
    draw.ellipse((54, 54, 202, 202), outline=INK, width=10)
    line(draw, [(82, 132), (174, 132)], MUTED, 8)
    line(draw, [(128, 66), (128, 190)], MUTED, 8)
    draw.arc((86, 54, 170, 202), 70, 290, fill=ACCENT, width=8)
    draw.arc((86, 54, 170, 202), -110, 110, fill=ACCENT, width=8)


def icon_warning(draw):
    draw.polygon([(128, 46), (218, 202), (38, 202)], outline=INK, fill=SOFT)
    line(draw, [(128, 94), (128, 152)], ACCENT, 12)
    draw.ellipse((120, 170, 136, 186), fill=ACCENT)


def icon_fix(draw):
    line(draw, [(62, 190), (122, 130), (150, 158), (204, 104)], INK, 13)
    draw.ellipse((104, 112, 166, 174), outline=ACCENT, width=10)
    line(draw, [(174, 82), (204, 52), (218, 66), (188, 96)], MUTED, 10)


def icon_template(draw):
    rect(draw, (46, 50, 210, 206), INK, 10, 18)
    draw.rounded_rectangle((70, 76, 186, 116), radius=10, fill=SOFT, outline=ACCENT, width=6)
    for x in [70, 116, 162]:
        draw.rounded_rectangle((x, 138, x + 24, 174), radius=6, fill=MUTED)


def icon_background(draw):
    rect(draw, (46, 64, 210, 192), INK, 10, 20)
    draw.polygon([(60, 176), (108, 122), (140, 156), (166, 128), (202, 176)], fill=SOFT, outline=ACCENT)
    draw.ellipse((158, 82, 184, 108), fill=ACCENT)


def icon_primitive(draw):
    draw.ellipse((52, 68, 116, 132), outline=INK, width=9)
    draw.rounded_rectangle((142, 64, 206, 128), radius=12, outline=ACCENT, width=9)
    draw.polygon([(86, 188), (128, 146), (170, 188)], outline=MUTED, width=9)


def icon_card(draw):
    for i, xy in enumerate([(48, 68, 122, 140), (134, 68, 208, 140), (90, 152, 166, 212)]):
        draw.rounded_rectangle(xy, radius=16, fill=SOFT, outline=[INK, MUTED, ACCENT][i], width=8)
        line(draw, [(xy[0] + 18, xy[1] + 24), (xy[2] - 18, xy[1] + 24)], [INK, MUTED, ACCENT][i], 6)


def icon_flow(draw):
    for x, label_color in [(54, INK), (128, ACCENT), (202, MUTED)]:
        draw.ellipse((x - 24, 104, x + 24, 152), fill=SOFT, outline=label_color, width=8)
    line(draw, [(78, 128), (104, 128)], INK_DARK, 8)
    line(draw, [(152, 128), (178, 128)], INK_DARK, 8)


def icon_review(draw):
    rect(draw, (52, 54, 204, 202), INK, 10, 18)
    line(draw, [(82, 132), (112, 160), (174, 92)], ACCENT, 14)


ICON_DRAWERS = {
    "runtime": icon_runtime,
    "timeline": icon_timeline,
    "tool": icon_tool,
    "llm": icon_llm,
    "file": icon_file,
    "report": icon_report,
    "contract": icon_contract,
    "capacity": icon_capacity,
    "layout": icon_layout,
    "split": icon_split,
    "align": icon_align,
    "density": icon_density,
    "chart": icon_chart,
    "kpi": icon_kpi,
    "trend": icon_trend,
    "source": icon_source,
    "warning": icon_warning,
    "fix": icon_fix,
    "template": icon_template,
    "background": icon_background,
    "primitive": icon_primitive,
    "card": icon_card,
    "flow": icon_flow,
    "review": icon_review,
}


def save_icons(manifest: list[dict]) -> None:
    for icon_id, drawer in ICON_DRAWERS.items():
        img, draw = draw_icon_base()
        drawer(draw)
        out = ICON_DIR / f"{icon_id}.png"
        img.save(out)
        manifest.append({
            "id": icon_id,
            "type": "icon",
            "path": str(out.relative_to(ASSET_ROOT)).replace("\\", "/"),
            "tags": ICON_TAGS[icon_id],
            "style": "soft-line",
            "recommended_templates": ["icon_grid", "card_grid", "content_slide"],
            "license": "project-generated",
        })


def lerp(a: int, b: int, t: float) -> int:
    return int(a + (b - a) * t)


def gradient_background(name: str, colors: list[tuple[int, int, int]], accent: tuple[int, int, int]) -> Path:
    w, h = 1920, 1080
    img = Image.new("RGB", (w, h), colors[0])
    px = img.load()
    for y in range(h):
        t = y / max(h - 1, 1)
        c1 = colors[0]
        c2 = colors[-1]
        for x in range(w):
            wave = (math.sin((x / w) * math.pi * 1.5 + t * 1.8) + 1) * 0.07
            tt = min(1, max(0, t + wave))
            px[x, y] = tuple(lerp(c1[i], c2[i], tt) for i in range(3))
    draw = ImageDraw.Draw(img, "RGBA")
    random.seed(42 + len(name))
    for i in range(5):
        x = random.randint(-200, w - 100)
        y = random.randint(-120, h - 60)
        ww = random.randint(360, 780)
        hh = random.randint(120, 320)
        alpha = random.randint(26, 58)
        draw.rounded_rectangle((x, y, x + ww, y + hh), radius=70, fill=(*accent, alpha))
    for i in range(12):
        x = random.randint(0, w)
        y = random.randint(0, h)
        r = random.randint(3, 9)
        draw.ellipse((x - r, y - r, x + r, y + r), fill=(*accent, 34))
    img = img.filter(ImageFilter.GaussianBlur(radius=0.4))
    out = BG_DIR / f"{name}.png"
    img.save(out, quality=95)
    return out


def pattern_background(name: str, mode: str) -> Path:
    w, h = 1920, 1080
    img = Image.new("RGBA", (w, h), (247, 249, 248, 255))
    draw = ImageDraw.Draw(img, "RGBA")
    if mode == "grid":
        for x in range(0, w, 80):
            line(draw, [(x, 0), (x, h)], (77, 116, 138, 34), 2)
        for y in range(0, h, 80):
            line(draw, [(0, y), (w, y)], (77, 116, 138, 30), 2)
    elif mode == "dots":
        for y in range(44, h, 74):
            for x in range(44, w, 74):
                draw.ellipse((x - 3, y - 3, x + 3, y + 3), fill=(77, 116, 138, 44))
    elif mode == "waves":
        for y in range(80, h, 120):
            pts = []
            for x in range(0, w + 20, 20):
                pts.append((x, y + math.sin(x / 95) * 22))
            draw.line(pts, fill=(181, 151, 105, 50), width=4)
    else:
        for x in range(-h, w, 110):
            draw.line([(x, h), (x + h, 0)], fill=(77, 116, 138, 34), width=5)
    out = PATTERN_DIR / f"{name}.png"
    img.convert("RGB").save(out, quality=95)
    return out


def save_backgrounds(manifest: list[dict]) -> None:
    specs = [
        ("editorial_blue", [(242, 247, 249), (222, 235, 240)], (77, 116, 138), ["title", "section", "hero"]),
        ("editorial_sage", [(246, 248, 244), (225, 236, 226)], (104, 137, 119), ["title", "quote", "summary"]),
        ("editorial_gold", [(249, 247, 241), (238, 229, 207)], (181, 151, 105), ["section", "summary"]),
        ("editorial_ink", [(239, 243, 244), (218, 226, 230)], (37, 54, 70), ["title", "quote"]),
        ("editorial_coral", [(249, 245, 242), (237, 222, 215)], (174, 112, 92), ["title", "section"]),
        ("editorial_lavender", [(247, 246, 250), (229, 225, 241)], (124, 111, 156), ["quote", "summary"]),
    ]
    for bg_id, colors, accent, templates in specs:
        out = gradient_background(bg_id, colors, accent)
        manifest.append({
            "id": bg_id,
            "type": "background",
            "path": str(out.relative_to(ASSET_ROOT)).replace("\\", "/"),
            "tags": templates + ["editorial", "soft"],
            "style": "soft-editorial",
            "recommended_templates": templates,
            "license": "project-generated",
        })

    for pattern_id, mode in [
        ("pattern_grid", "grid"),
        ("pattern_dots", "dots"),
        ("pattern_waves", "waves"),
        ("pattern_diagonal", "diagonal"),
    ]:
        out = pattern_background(pattern_id, mode)
        manifest.append({
            "id": pattern_id,
            "type": "pattern",
            "path": str(out.relative_to(ASSET_ROOT)).replace("\\", "/"),
            "tags": [mode, "subtle", "clean"],
            "style": "subtle-pattern",
            "recommended_templates": ["content_slide", "card_grid", "chart_slide"],
            "license": "project-generated",
        })


def main() -> None:
    ensure_dirs()
    manifest: list[dict] = []
    save_icons(manifest)
    save_backgrounds(manifest)
    (ASSET_ROOT / "manifest.json").write_text(
        json.dumps({"version": 1, "assets": manifest}, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    print(f"generated {len(manifest)} assets in {ASSET_ROOT}")


if __name__ == "__main__":
    main()
