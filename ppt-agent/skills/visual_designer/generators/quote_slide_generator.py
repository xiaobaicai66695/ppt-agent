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
from .asset_manager import apply_asset_background, add_local_icon
from .layout_intelligence import balanced_band_top, focal_font_size


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
    colors = apply_asset_background(slide, background, palette, role="quote", brightness=0.96)
    if colors:
        pass
    else:
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

    add_local_icon(slide, "quote", left=11.0, top=0.7, size=0.75, palette=palette)

    quote_box_h = 3.35
    quote_group_h = quote_box_h + 0.68
    quote_top = balanced_band_top(y_start + 0.55, 5.25 - y_start, quote_group_h, min_top=y_start + 0.9)

    # Large decorative opening quote mark (top left)
    add_text(
        slide,
        text="\u201C",
        left=0.5, top=quote_top - 1.12, width=2.0, height=2.0,
        font_size=120, bold=False,
        color="accent", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Quote background box
    add_rect(
        slide,
        left=1.35, top=quote_top, width=10.633, height=quote_box_h,
        fill_color="light_bg", palette=palette,
    )

    # Quote text (centered)
    add_text(
        slide,
        text=quote,
        left=1.85, top=quote_top + 0.22, width=9.633, height=2.9,
        font_size=focal_font_size(quote, base=30, max_size=42, min_size=22), bold=False,
        color="text", alignment="center",
        vertical_alignment="middle",
        line_spacing=0.95,
        palette=palette,
        colors=colors,
    )

    # Attribution
    add_text(
        slide,
        text=f"—— {attribution}",
        left=5.0, top=quote_top + quote_box_h + 0.18, width=7.5, height=0.5,
        font_size=14, bold=False,
        color="secondary", alignment="right",
        palette=palette,
        colors=colors,
    )

    # Bottom-right editorial block
    add_rect(
        slide,
        left=11.35, top=6.1, width=1.35, height=0.92,
        fill_color="accent", palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs
