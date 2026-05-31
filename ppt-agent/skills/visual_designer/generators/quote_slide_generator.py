"""Generator for quote_slide - 金句/引言页."""
from typing import Optional

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect,
    set_slide_background,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    quote: str = "{金句}",
    attribution: str = "{来源}",
    kicker: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a quote slide with centered quote and attribution.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        quote: The quote text.
        attribution: Attribution/source of the quote.
        kicker: Small label above quote mark area (e.g. "金句").

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

    # Kicker (above everything)
    y_start = 0.15
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.5, top=y_start, width=12.0, height=0.3,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_start += 0.35

    # Large decorative opening quote mark (top left)
    add_text(
        slide,
        text="\u201C",
        left=0.5, top=y_start + 0.3, width=2.0, height=2.0,
        font_size=120, bold=False,
        color="accent", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Quote background box
    add_rect(
        slide,
        left=1.5, top=y_start + 1.5, width=10.333, height=3.2,
        fill_color="light_bg", palette=palette,
    )

    # Quote text (centered)
    add_text(
        slide,
        text=quote,
        left=1.8, top=y_start + 1.8, width=9.733, height=2.8,
        font_size=28, bold=False,
        color="text", alignment="center",
        palette=palette,
        colors=colors,
    )

    # Attribution
    add_text(
        slide,
        text=f"—— {attribution}",
        left=5.0, top=y_start + 4.8, width=7.5, height=0.5,
        font_size=14, bold=False,
        color="secondary", alignment="right",
        palette=palette,
        colors=colors,
    )

    # Bottom-right decorative circle
    add_rect(
        slide,
        left=11.5, top=6.2, width=1.5, height=1.0,
        fill_color="accent", palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs
