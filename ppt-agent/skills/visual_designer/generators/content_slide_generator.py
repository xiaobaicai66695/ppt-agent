"""Generator for content_slide - 普通内容页（兜底类型）."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_round_rect,
    set_slide_background, add_paragraph,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{页面标题}",
    section_header: str = "",
    bullets: List[str] = None,
    kicker: str = "",
    lede: str = "",
) -> Presentation:
    """Generate a content slide with title and staggered bullet list.

    Bullets alternate between left and center-left positions to break
    the monotony of a single vertical column. Supports 1-8 items.
    """
    if bullets is None:
        bullets = ["{要点1}", "{要点2}", "{要点3}"]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # ── Header area ──
    y_head = 0.35
    if kicker:
        add_text(slide, text=kicker,
            left=0.7, top=0.12, width=11.5, height=0.28,
            font_size=11, bold=False, color="text_muted", alignment="left",
            palette=palette)
        y_head = 0.3

    # Title with slim left accent bar
    add_rect(slide, left=0.5, top=y_head + 0.08, width=0.05, height=0.5,
        fill_color="primary", palette=palette)
    add_text(slide, text=title,
        left=0.7, top=y_head, width=11.8, height=0.7,
        font_size=34, bold=True, color="text", alignment="left",
        palette=palette)

    # Section header
    y_body = 1.25 if not kicker else 1.15
    if section_header:
        add_text(slide, text=section_header,
            left=0.7, top=y_body, width=11.5, height=0.45,
            font_size=20, bold=True, color="primary", alignment="left",
            palette=palette)
        y_body += 0.55

    # Lede paragraph
    if lede:
        add_text(slide, text=lede,
            left=0.7, top=y_body, width=11.5, height=0.45,
            font_size=13, bold=False, color="text_muted", alignment="left",
            palette=palette)
        y_body += 0.55

    # ── Staggered bullet list ──
    n = len(bullets[:8])
    if n == 0:
        add_source_line(slide, source, palette)
        return prs

    # 1-3 items: use larger font, wider spacing
    if n <= 3:
        row_h = 0.9
        for i, item in enumerate(bullets[:n]):
            y = y_body + i * row_h
            _bullet_row(slide, palette, colors, item, y, i, n, row_h, font_size=18, dot_size=0.14)
    # 4-6 items: standard
    elif n <= 6:
        row_h = 0.7
        for i, item in enumerate(bullets[:n]):
            y = y_body + i * row_h
            _bullet_row(slide, palette, colors, item, y, i, n, row_h, font_size=15, dot_size=0.12)
    # 7-8 items: compact
    else:
        row_h = 0.58
        for i, item in enumerate(bullets[:n]):
            y = y_body + i * row_h
            _bullet_row(slide, palette, colors, item, y, i, n, row_h, font_size=13, dot_size=0.10)

    # ── Soft corner decoration ──
    # Right-bottom: gentle accent shape (rounded rect instead of sharp)
    add_round_rect(slide, left=12.0, top=6.3, width=1.1, height=0.7,
        fill_color="light_bg", palette=palette)
    # Top-right: tiny dot cluster
    add_round_rect(slide, left=12.2, top=0.18, width=0.08, height=0.08,
        fill_color="accent", palette=palette)
    add_round_rect(slide, left=12.4, top=0.28, width=0.06, height=0.06,
        fill_color="secondary", palette=palette)

    add_source_line(slide, source, palette)
    return prs


def _bullet_row(slide, palette, colors, item, y, i, n, row_h, font_size=15, dot_size=0.12):
    """Render a single bullet with staggered left offset."""
    # Stagger: even items shift right slightly
    base_left = 0.7
    if i % 2 == 1:
        base_left += 0.35  # 偶数行右移营造错落感

    # Round dot (replaces the square dot)
    add_round_rect(slide,
        left=base_left, top=y + 0.07, width=dot_size, height=dot_size,
        fill_color="secondary" if i % 2 == 0 else "accent", palette=palette)

    # Item text
    text_w = 10.8 - (base_left - 0.7)  # narrower when shifted right
    add_text(slide, text=item,
        left=base_left + 0.25, top=y, width=text_w, height=row_h,
        font_size=font_size, bold=False, color="text", alignment="left",
        palette=palette)

    # Subtle horizontal rule between items (soft divider)
    if i < n - 1:
        add_rect(slide,
            left=0.7, top=y + row_h - 0.02, width=4.5, height=0.01,
            fill_color="divider", palette=palette)
