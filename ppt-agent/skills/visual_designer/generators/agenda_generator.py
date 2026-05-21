"""Generator for agenda - 目录页 (自适应条目数)."""
from typing import List, Optional
from pptx import Presentation
from .base import (
    add_source_line, PALETTES, add_text, add_round_rect,
    new_presentation, set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "目录",
    title: str = "内容概览",
    items: Optional[List[str]] = None,
) -> Presentation:
    """Adaptive agenda: fewer items → larger text, single column. Many items → two columns."""
    if items is None:
        items = ["01  {章节1}", "02  {章节2}", "03  {章节3}", "04  {章节4}"]

    if prs is None:
        prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    n = len(items)
    # Clamp to reasonable range
    if n <= 3:
        n_cols, font_sz, row_h, start_y = 1, 22, 0.95, 1.6
    elif n <= 5:
        n_cols, font_sz, row_h, start_y = 1, 18, 0.8, 1.5
    else:
        n_cols, font_sz, row_h, start_y = 2, 16, 0.7, 1.5

    # ── Header ──
    add_text(slide, text=kicker,
        left=0.5, top=0.25, width=12.0, height=0.25,
        font_size=11, bold=False, color="text_muted", alignment="left", palette=palette)
    add_text(slide, text=title,
        left=0.5, top=0.45, width=12.0, height=0.55,
        font_size=32, bold=True, color="text", alignment="left", palette=palette)

    # ── Items ──
    col_w = 5.8 if n_cols == 2 else 11.5
    gap = 0.8

    # Accent colors cycle
    accents = ["primary", "secondary", "accent", "primary", "secondary"]

    for i, item in enumerate(items[:10]):
        if n_cols == 2:
            col = i % 2
            row_in_col = i // 2
            x = 0.5 + col * (col_w + gap)
        else:
            col = 0
            row_in_col = i
            x = 0.7

        y = start_y + row_in_col * row_h
        a = accents[i % len(accents)]

        # Small round dot accent
        add_round_rect(slide, left=x, top=y + 0.1, width=0.1, height=0.1,
            fill_color=a, palette=palette)

        add_text(slide, text=item,
            left=x + 0.25, top=y, width=col_w - 0.25, height=row_h,
            font_size=font_sz, bold=True, color="text", alignment="left", palette=palette)

        # Subtle divider (dashed feel via thin line)
        if row_in_col < ((len(items[:10]) - 1) // n_cols if n_cols == 1 else (len(items[:10]) + 1) // n_cols - 1):
            from .base import add_rect
            add_rect(slide, left=x + 0.05, top=y + row_h - 0.03, width=3.0, height=0.01,
                fill_color="divider", palette=palette)

    # Bottom-right soft ornament
    add_round_rect(slide, left=12.4, top=6.6, width=0.6, height=0.4,
        fill_color="light_bg", palette=palette)
    add_source_line(slide, source, palette)
    return prs
