"""Generator for content_slide - 普通内容页（兜底类型）."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_text_in_shape,
    set_slide_background, add_paragraph,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{页面标题}",
    section_header: str = "{小节标题}",
    bullets: List[str] = None,
    kicker: str = "",
    lede: str = "",
    highlight_stats: List[dict] = None,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a content slide with title, optional section header, and bullet list.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        section_header: Optional section header text.
        bullets: List of bullet items (max 5, each up to 20 Chinese chars).
        kicker: Small label above title (e.g. "要点 · 核心技术").
        lede: Optional lead-in paragraph below section header (1-2 sentences, ~50 chars).
        highlight_stats: Optional list of dicts with keys: value, label for right-side stat cards.
                        Example: [{"value": "98%", "label": "满意度"}, {"value": "3x", "label": "效率提升"}]

    Returns:
        The Presentation object.
    """
    if bullets is None:
        bullets = [
            "{要点1}",
            "{要点2}",
            "{要点3}",
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
    y_title = 0.75
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.7, top=0.3, width=11.5, height=0.35,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_title = 0.7

    # Left accent vertical bar
    add_rect(
        slide,
        left=0.5, top=y_title + 0.05, width=0.08, height=0.6,
        fill_color="primary", palette=palette,
    )

    # Title
    add_text(
        slide,
        text=title,
        left=0.7, top=y_title, width=11.5, height=0.7,
        font_size=36, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Section header
    y_offset = 1.7 if not kicker else 1.55
    if section_header:
        add_text(
            slide,
            text=section_header,
            left=0.7, top=y_offset, width=11.5, height=0.5,
            font_size=20, bold=True,
            color="primary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.6

    # Lede (lead-in paragraph between section_header and bullets)
    if lede:
        add_text(
            slide,
            text=lede,
            left=0.7, top=y_offset, width=11.5, height=0.5,
            font_size=14, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.6

    # Bullet list
    bullet_spacing = 0.65
    for i, item in enumerate(bullets[:6]):
        y = y_offset + i * bullet_spacing

        # Bullet dot
        dot = add_rect(
            slide,
            left=0.8, top=y + 0.12, width=0.12, height=0.12,
            fill_color="secondary", palette=palette,
        )

        # Bullet text
        add_text(
            slide,
            text=item,
            left=1.05, top=y, width=11.0, height=0.5,
            font_size=16, bold=False,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )

    # Right-side highlight stats (if provided)
    if highlight_stats:
        stats_x = 9.5
        stats_top = 2.5
        for i, stat in enumerate(highlight_stats[:2]):
            y = stats_top + i * 1.6
            value = stat.get("value", "")
            label = stat.get("label", "")

            # Stat card background
            add_rect(
                slide,
                left=stats_x, top=y, width=3.3, height=1.4,
                fill_color="light_bg", palette=palette,
            )
            # Left accent bar
            add_rect(
                slide,
                left=stats_x, top=y, width=0.06, height=1.4,
                fill_color="primary", palette=palette,
            )
            # Value
            add_text(
                slide,
                text=value,
                left=stats_x + 0.2, top=y + 0.2,
                width=2.9, height=0.7,
                font_size=28, bold=True,
                color="primary", alignment="left",
                palette=palette,
                colors=colors,
            )
            # Label
            add_text(
                slide,
                text=label,
                left=stats_x + 0.2, top=y + 0.9,
                width=2.9, height=0.4,
                font_size=12, bold=False,
                color="text_muted", alignment="left",
                palette=palette,
                colors=colors,
            )

    # Bottom-right decorative geometric shape
    add_rect(
        slide,
        left=11.8, top=6.5, width=1.2, height=0.8,
        fill_color="light_bg", palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs
