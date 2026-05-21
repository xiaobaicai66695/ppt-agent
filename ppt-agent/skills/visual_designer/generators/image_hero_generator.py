"""Generator for image_hero - 视觉冲击页 (no harsh lines, soft color blocks)."""
from typing import Optional
from pptx import Presentation
from .base import (
    add_source_line, new_presentation, PALETTES, add_text, add_rect, add_round_rect,
    set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "",
    title: str = "{大标题}", subtitle: str = "", description: str = "",
    overlay_color: str = "primary",
) -> Presentation:
    if prs is None: prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    # ── Large color block (no internal lines) ──
    add_rect(slide, left=0, top=0, width=13.333, height=5.0,
        fill_color=overlay_color, palette=palette)

    # Bottom light area
    add_rect(slide, left=0, top=5.0, width=13.333, height=2.5,
        fill_color="light_bg", palette=palette)

    # ── Title: bold, centered on dark background ──
    add_text(slide, text=title, left=1.2, top=1.5, width=10.933, height=1.3,
        font_size=44, bold=True, color="background", alignment="center", palette=palette)

    if subtitle:
        add_text(slide, text=subtitle, left=2.2, top=3.0, width=8.933, height=0.45,
            font_size=18, bold=False, color="background", alignment="center", palette=palette)

    if description:
        add_text(slide, text=description, left=2.2, top=5.4, width=8.933, height=0.9,
            font_size=14, bold=False, color="text", alignment="center", palette=palette)

    # ── Organic corner dots (scattered, imperfect spacing) ──
    add_round_rect(slide, left=1.2, top=0.25, width=0.12, height=0.12,
        fill_color="accent", palette=palette)
    add_round_rect(slide, left=1.5, top=0.3, width=0.07, height=0.07,
        fill_color="secondary", palette=palette)
    add_round_rect(slide, left=11.8, top=6.3, width=0.09, height=0.09,
        fill_color="accent", palette=palette)

    add_source_line(slide, source, palette)
    return prs
