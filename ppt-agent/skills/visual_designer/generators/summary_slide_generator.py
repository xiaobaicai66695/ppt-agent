"""Generator for summary_slide - 总结页."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect,
    set_slide_background,
    resolve_background, set_image_background,
)
from .asset_manager import apply_asset_background, add_local_icon, icon_id_from_text
from .layout_intelligence import balanced_band_top, body_font_size, title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{总结标题}",
    key_points: List[str] = None,
    thank_you: str = "感谢聆听",
    contact: str = "{联系方式}",
    kicker: str = "",
    background: str = None,
    glass_colors: dict = None,
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
    colors = apply_asset_background(slide, background, palette, role="summary", brightness=0.96)
    if colors:
        pass
    else:
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
        font_size=title_font_size(title, base=36, sparse_boost=6, max_size=46), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Key points area background
    add_rect(
        slide,
        left=0.5, top=1.3, width=8.5, height=4.5,
        fill_color="light_bg", palette=palette,
    )

    # Key points with left accent bar
    point_font = body_font_size(key_points[:4], base=16)
    points = key_points[:4]
    point_gap = 0.9 if len(points) >= 4 else 0.98
    points_h = len(points) * 0.7 + max(len(points) - 1, 0) * (point_gap - 0.7)
    points_top = balanced_band_top(1.55, 3.95, points_h, min_top=1.5)
    for i, point in enumerate(points):
        y = points_top + i * point_gap

        add_local_icon(
            slide,
            icon_id_from_text(point, fallback="check"),
            left=0.72, top=y + 0.04, size=0.36,
            palette=palette,
        )

        add_text(
            slide,
            text=point,
            left=1.18, top=y, width=7.35, height=0.7,
            font_size=point_font, bold=False,
            color="text", alignment="left",
            vertical_alignment="middle",
            palette=palette,
            colors=colors,
        )

    # Thank you area background
    add_rect(
        slide,
        left=9.3, top=1.3, width=3.5, height=4.5,
        fill_color="primary", palette=palette,
    )

    add_local_icon(
        slide,
        "thanks",
        left=10.72, top=1.78, size=0.66,
        palette=palette,
        with_badge=True,
        badge_color="background",
    )

    # Thank you text
    add_text(
        slide,
        text=thank_you,
        left=9.3, top=2.8, width=3.5, height=0.8,
        font_size=28 if len(thank_you) <= 10 else 22, bold=True,
        color="background", alignment="center",
        vertical_alignment="middle",
        palette=palette,
        colors=colors,
    )

    # Contact
    add_text(
        slide,
        text=contact,
        left=9.3, top=3.7, width=3.5, height=0.6,
        font_size=12, bold=False,
        color="accent", alignment="center",
        vertical_alignment="middle",
        palette=palette,
        colors=colors,
    )

    add_source_line(slide, source, palette)
    return prs
