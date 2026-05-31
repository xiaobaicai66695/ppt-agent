"""Generator for section_divider - 章节分隔页."""
from typing import Optional

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, rgb, add_text, add_rect,
    set_slide_background,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    number: str = "01",
    title: str = "{章节标题}",
    subtitle: str = "{章节副标题}",
    kicker: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a section divider slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        number: Section number (e.g. "01").
        title: Section title text.
        subtitle: Optional subtitle text.
        kicker: Small label above number (e.g. "第三章").

    Returns:
        The Presentation object.
    """
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

    # Large background color block (left 40% of slide)
    add_rect(
        slide,
        left=0, top=0, width=5.2, height=7.5,
        fill_color="primary", palette=palette,
    )

    # Kicker (above number)
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.4, top=0.6, width=4.5, height=0.5,
            font_size=14, bold=False,
            color="background", alignment="left",
            palette=palette,
            colors=colors,
        )

    # Large decorative number (right side of color block)
    add_text(
        slide,
        text=number,
        left=0.4, top=1.2, width=4.5, height=3.0,
        font_size=160, bold=True,
        color="background", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Section title (on the right, white area)
    add_text(
        slide,
        text=title,
        left=5.8, top=3.1, width=6.8, height=1.2,
        font_size=44, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=5.8, top=4.4, width=6.8, height=0.6,
            font_size=16, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )

    # Accent line under title
    add_rect(
        slide,
        left=5.8, top=4.3, width=1.2, height=0.06,
        fill_color="accent", palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs
