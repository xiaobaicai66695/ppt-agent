"""Generator for two_column - 双栏对比."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect,
    set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{对比标题}",
    left_header: str = "{左侧标题}",
    right_header: str = "{右侧标题}",
    left_bullets: List[str] = None,
    right_bullets: List[str] = None,
    kicker: str = "",
) -> Presentation:
    """
    Generate a two-column comparison slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        left_header: Left column header.
        right_header: Right column header.
        left_bullets: Left column bullet items.
        right_bullets: Right column bullet items.
        kicker: Small label above title (e.g. "方案对比").

    Returns:
        The Presentation object.
    """
    if left_bullets is None:
        left_bullets = [
            "{要点1}",
            "{要点2}",
            "{要点3}",
            "{要点4}",
        ]
    if right_bullets is None:
        right_bullets = [
            "{要点1}",
            "{要点2}",
            "{要点3}",
            "{要点4}",
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Kicker (above title)
    y_title = 0.4
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.5, top=0.1, width=12.0, height=0.3,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
        )
        y_title = 0.35

    # Title
    add_text(
        slide,
        text=title,
        left=0.5, top=y_title, width=12.0, height=0.7,
        font_size=36, bold=True,
        color="text", alignment="left",
        palette=palette,
    )

    col_w = 5.8
    col_h = 5.2
    gap = 0.7
    start_x = 0.5
    start_y = 1.4 if not kicker else 1.35

    # Left column background
    add_rect(
        slide,
        left=start_x, top=start_y, width=col_w, height=col_h,
        fill_color="light_bg", palette=palette,
    )
    # Left column top accent bar
    add_rect(
        slide,
        left=start_x, top=start_y, width=col_w, height=0.08,
        fill_color="primary", palette=palette,
    )

    # Right column background
    add_rect(
        slide,
        left=start_x + col_w + gap, top=start_y, width=col_w, height=col_h,
        fill_color="light_bg", palette=palette,
    )
    # Right column top accent bar
    add_rect(
        slide,
        left=start_x + col_w + gap, top=start_y, width=col_w, height=0.08,
        fill_color="accent", palette=palette,
    )

    # Left header
    add_text(
        slide,
        text=left_header,
        left=start_x + 0.2, top=start_y + 0.2, width=col_w - 0.4, height=0.6,
        font_size=18, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )

    # Left bullets
    for i, item in enumerate(left_bullets[:6]):
        add_text(
            slide,
            text="· " + item,
            left=start_x + 0.2, top=start_y + 1.0 + i * 0.9, width=col_w - 0.4, height=0.8,
            font_size=13, bold=False,
            color="text", alignment="left",
            palette=palette,
        )

    # Right header
    add_text(
        slide,
        text=right_header,
        left=start_x + col_w + gap + 0.2, top=start_y + 0.2, width=col_w - 0.4, height=0.6,
        font_size=18, bold=True,
        color="accent", alignment="left",
        palette=palette,
    )

    # Right bullets
    for i, item in enumerate(right_bullets[:6]):
        add_text(
            slide,
            text="· " + item,
            left=start_x + col_w + gap + 0.2, top=start_y + 1.0 + i * 0.9, width=col_w - 0.4, height=0.8,
            font_size=13, bold=False,
            color="text", alignment="left",
            palette=palette,
        )

    add_source_line(slide, source, palette)
    return prs
