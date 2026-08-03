"""Generator for three_column - 三栏并列."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect,
    set_slide_background,
    resolve_background, set_image_background,
)
from .layout_intelligence import balanced_band_top, title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{三栏标题}",
    columns: List[dict] = None,
    kicker: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a three-column slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        columns: List of 3 dicts with keys: header, bullets.
        kicker: Small label above title (e.g. "能力矩阵").

    Returns:
        The Presentation object.
    """
    if columns is None:
        columns = [
            {
                "header": "01  {领域1}",
                "bullets": ["{要点1}", "{要点2}", "{要点3}"],
            },
            {
                "header": "02  {领域2}",
                "bullets": ["{要点1}", "{要点2}", "{要点3}"],
            },
            {
                "header": "03  {领域3}",
                "bullets": ["{要点1}", "{要点2}", "{要点3}"],
            },
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
            colors=colors,
        )
        y_title = 0.35

    # Title
    add_text(
        slide,
        text=title,
        left=0.5, top=y_title, width=12.0, height=0.7,
        font_size=title_font_size(title, base=36, sparse_boost=5, max_size=44), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    col_w = 3.8
    gap = 0.3
    start_x = 0.6
    col_h = 5.0
    start_y = balanced_band_top(1.25 if kicker else 1.35, 5.15, col_h, min_top=1.35)

    header_colors = ["primary", "secondary", "accent"]

    for i, col in enumerate(columns[:3]):
        x = start_x + i * (col_w + gap)
        header = col.get("header", "")
        bullets = col.get("bullets", [])

        # Column background
        add_rect(
            slide,
            left=x, top=start_y, width=col_w, height=col_h,
            fill_color="light_bg", palette=palette,
        )

        # Header background bar
        add_rect(
            slide,
            left=x, top=start_y, width=col_w, height=0.7,
            fill_color=header_colors[i], palette=palette,
        )

        # Header text
        add_text(
            slide,
            text=header,
            left=x + 0.15, top=start_y + 0.15, width=col_w - 0.3, height=0.5,
            font_size=16, bold=True,
            color="background" if i == 0 else "text", alignment="left",
            palette=palette,
            colors=colors,
        )

        # Bullet items
        for j, item in enumerate(bullets[:5]):
            # Bullet dot
            add_rect(
                slide,
                left=x + 0.2, top=start_y + 1.1 + j * 0.9 + 0.08, width=0.12, height=0.12,
                fill_color=header_colors[i], palette=palette,
            )
            add_text(
                slide,
                text=item,
                left=x + 0.45, top=start_y + 1.0 + j * 0.9, width=col_w - 0.6, height=0.6,
                font_size=14, bold=False,
                color="text", alignment="left",
                vertical_alignment="middle",
                palette=palette,
                colors=colors,
            )

    add_source_line(slide, source, palette)
    return prs
