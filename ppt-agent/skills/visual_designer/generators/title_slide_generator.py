"""Generator for title_slide -开场标题页."""
from typing import Optional

from pptx import Presentation
from pptx.util import Inches

from .base import (
    add_source_line,
    PALETTES, rgb, add_text, add_ellipse, add_rect,
    set_slide_background, add_text_in_shape,
    new_presentation,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{演讲主题}",
    subtitle: str = "{副标题}",
    author: str = "{演讲者}",
    date: str = "{日期}",
    kicker: str = "",
) -> Presentation:
    """
    Generate a title slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Main title text.
        subtitle: Subtitle text.
        author: Author/department name (bottom left).
        date: Date text (bottom right).
        kicker: Small label above title (e.g. "产品发布 · 2025").

    Returns:
        The Presentation object.
    """
    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Kicker (above title)
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=1.0, top=1.9, width=11.333, height=0.4,
            font_size=14, bold=False,
            color="secondary", alignment="center",
            palette=palette,
        )

    # Decorative bottom-left corner rounded rect
    add_rect(
        slide,
        left=-0.3, top=6.2, width=3.5, height=1.8,
        fill_color="primary", palette=palette,
    )

    # Decorative bottom-right small circle
    add_ellipse(
        slide,
        left=11.8, top=6.0, width=1.2, height=1.2,
        fill_color="accent", palette=palette,
    )

    # Decorative top-right small circle
    add_ellipse(
        slide,
        left=11.5, top=0.2, width=0.8, height=0.8,
        fill_color="secondary", palette=palette,
    )

    # Main title - centered, large
    add_text(
        slide,
        text=title,
        left=1.0, top=2.35, width=11.333, height=1.2,
        font_size=44, bold=True,
        color="text", alignment="center",
        palette=palette,
    )

    # Subtitle
    add_text(
        slide,
        text=subtitle,
        left=1.5, top=3.65, width=10.333, height=0.6,
        font_size=20, bold=False,
        color="secondary", alignment="center",
        palette=palette,
    )

    # Author (bottom left)
    add_text(
        slide,
        text=author,
        left=0.5, top=6.5, width=4.0, height=0.4,
        font_size=14, bold=False,
        color="background", alignment="left",
        palette=palette,
    )

    # Date (bottom right)
    add_text(
        slide,
        text=date,
        left=9.0, top=6.5, width=3.5, height=0.4,
        font_size=12, bold=False,
        color="text_muted", alignment="right",
        palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs
