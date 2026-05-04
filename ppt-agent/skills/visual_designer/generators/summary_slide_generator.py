"""Generator for summary_slide - 总结页."""
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
    title: str = "{总结标题}",
    key_points: List[str] = None,
    thank_you: str = "感谢聆听",
    contact: str = "{联系方式}",
    kicker: str = "",
) -> Presentation:
    """
    Generate a summary slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        key_points: List of key takeaway strings (max 4).
        thank_you: Thank you message.
        contact: Contact information.
        kicker: Small label above title (e.g. "总结").

    Returns:
        The Presentation object.
    """
    if key_points is None:
        key_points = [
            "01 {要点1}",
            "02 {要点2}",
            "03 {要点3}",
            "04 {要点4}",
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

    # Key points area background
    add_rect(
        slide,
        left=0.5, top=1.3, width=8.5, height=4.5,
        fill_color="light_bg", palette=palette,
    )

    # Key points with left accent bar
    for i, point in enumerate(key_points[:4]):
        y = 1.5 + i * 1.0

        # Left accent bar
        add_rect(
            slide,
            left=0.65, top=y + 0.05, width=0.06, height=0.55,
            fill_color="primary", palette=palette,
        )

        add_text(
            slide,
            text=point,
            left=0.85, top=y, width=7.8, height=0.7,
            font_size=16, bold=False,
            color="text", alignment="left",
            palette=palette,
        )

    # Thank you area background
    add_rect(
        slide,
        left=9.3, top=1.3, width=3.5, height=4.5,
        fill_color="primary", palette=palette,
    )

    # Thank you text
    add_text(
        slide,
        text=thank_you,
        left=9.3, top=2.8, width=3.5, height=0.8,
        font_size=24, bold=True,
        color="background", alignment="center",
        palette=palette,
    )

    # Contact
    add_text(
        slide,
        text=contact,
        left=9.3, top=3.7, width=3.5, height=0.6,
        font_size=12, bold=False,
        color="accent", alignment="center",
        palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs
