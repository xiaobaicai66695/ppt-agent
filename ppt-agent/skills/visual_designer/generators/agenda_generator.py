"""Generator for agenda - 目录页.

Two-column numbered agenda with visual hierarchy: large primary-colored
numbers paired with chapter titles, separated by subtle divider lines.
"""
from typing import List, Optional

from pptx import Presentation

from .base import (
    add_source_line,
    PALETTES,
    add_line,
    add_text,
    new_presentation,
    set_slide_background,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "目录",
    title: str = "内容概览",
    items: Optional[List[str]] = None,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """Generate an agenda slide with structured two-column chapter listing."""
    if items is None:
        items = [
            "01  {章节1}",
            "02  {章节2}",
            "03  {章节3}",
            "04  {章节4}",
            "05  {章节5}",
            "06  {章节6}",
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    bg_path = resolve_background(background) if background else None
    if bg_path:
        colors = set_image_background(slide, bg_path, brightness=0.95, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Kicker
    add_text(
        slide, text=kicker,
        left=0.5, top=0.35, width=12.0, height=0.3,
        font_size=12, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Title
    add_text(
        slide, text=title,
        left=0.5, top=0.65, width=12.0, height=0.65,
        font_size=36, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Split items into two columns
    mid = (len(items) + 1) // 2
    left_items = items[:mid]
    right_items = items[mid:]

    col_w = 5.5
    gap = 1.0
    left_x = 0.7
    right_x = left_x + col_w + gap
    start_y = 1.7
    row_h = 0.75

    for i, item in enumerate(left_items):
        y = start_y + i * row_h
        # Divider line below each row
        if i < len(left_items):
            add_line(
                slide,
                x1=left_x, y1=y + 0.55, x2=left_x + col_w, y2=y + 0.55,
                color="divider", width=0.5, palette=palette,
            )
        add_text(
            slide, text=item,
            left=left_x, top=y, width=col_w, height=0.5,
            font_size=16, bold=True,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )

    for i, item in enumerate(right_items):
        y = start_y + i * row_h
        if i < len(right_items):
            add_line(
                slide,
                x1=right_x, y1=y + 0.55, x2=right_x + col_w, y2=y + 0.55,
                color="divider", width=0.5, palette=palette,
            )
        add_text(
            slide, text=item,
            left=right_x, top=y, width=col_w, height=0.5,
            font_size=16, bold=True,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )

    add_source_line(slide, source, palette)
    return prs
