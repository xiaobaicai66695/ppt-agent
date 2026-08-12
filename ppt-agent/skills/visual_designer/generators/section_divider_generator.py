"""Generator for section_divider - 章节分隔页."""
from typing import Optional

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, rgb, add_text, add_rect, add_round_rect,
    set_slide_background,
    resolve_background, set_image_background,
)
from .asset_manager import apply_asset_background, add_local_icon, add_pattern_overlay, icon_id_from_text
from .layout_intelligence import balanced_band_top, title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    number: str = "01",
    title: str = "{章节标题}",
    subtitle: str = "{章节副标题}",
    kicker: str = "",
    layout_variant: str = "",
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
        layout_variant: "number_sidebar", "quiet_title", or "photo_band".

    Returns:
        The Presentation object.
    """
    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    colors = apply_asset_background(slide, background, palette, role="section", brightness=0.96)
    if colors:
        pass
    else:
        bg_path = resolve_background(background) if background else None
        if bg_path:
            colors = set_image_background(slide, bg_path, brightness=0.95, palette=palette)
        else:
            set_slide_background(slide, palette)
            colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    variant = _resolve_layout_variant(layout_variant)
    if variant == "photo_band":
        return _render_photo_band(prs, slide, palette, colors, source, number, title, subtitle, kicker)
    if variant == "quiet_title":
        return _render_quiet_title(prs, slide, palette, colors, source, number, title, subtitle, kicker)

    # Large background color block (left 40% of slide)
    add_rect(
        slide,
        left=0, top=0, width=5.2, height=7.5,
        fill_color="primary", palette=palette,
    )
    add_local_icon(
        slide,
        icon_id_from_text(title + " " + subtitle, fallback="section"),
        left=3.95, top=5.8, size=0.78, palette=palette,
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
        left=0.4, top=balanced_band_top(1.1, 4.3, 3.0), width=4.5, height=3.0,
        font_size=160, bold=True,
        color="background", alignment="left",
        vertical_alignment="middle",
        palette=palette,
        colors=colors,
    )

    right_h = 1.2 + (0.72 if subtitle else 0.0)
    right_top = balanced_band_top(2.2, 3.2, right_h, min_top=2.55)

    # Section title (on the right, white area)
    add_text(
        slide,
        text=title,
        left=5.8, top=right_top, width=6.8, height=1.2,
        font_size=title_font_size(title, base=44, sparse_boost=6, max_size=52), bold=True,
        color="text", alignment="left",
        vertical_alignment="middle",
        palette=palette,
        colors=colors,
    )

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=5.8, top=right_top + 1.28, width=6.8, height=0.6,
            font_size=16, bold=False,
            color="secondary", alignment="left",
            vertical_alignment="middle",
            palette=palette,
            colors=colors,
        )

    # Accent line under title
    add_rect(
        slide,
        left=5.8, top=right_top + 1.18, width=1.2, height=0.06,
        fill_color="accent", palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs


def _resolve_layout_variant(layout_variant: str = "") -> str:
    key = (layout_variant or "").strip().lower().replace("-", "_")
    if key in {"photo_band", "band"}:
        return "photo_band"
    if key in {"quiet_title", "quiet"}:
        return "quiet_title"
    return "number_sidebar"


def _render_photo_band(
    prs: Presentation,
    slide,
    palette: str,
    colors: dict,
    source: str,
    number: str,
    title: str,
    subtitle: str,
    kicker: str,
) -> Presentation:
    """Render a section divider with a strong horizontal visual band."""
    band_y, band_h = 1.05, 2.15
    add_rect(slide, left=0, top=band_y, width=13.333, height=band_h, fill_color="light_bg", palette=palette)
    add_pattern_overlay(slide, "pattern_grid", left=0.2, top=band_y + 0.12, width=12.9, height=band_h - 0.24, opacity_backdrop=False, palette=palette)
    add_rect(slide, left=0, top=band_y, width=13.333, height=0.08, fill_color="primary", palette=palette)
    add_rect(slide, left=0, top=band_y + band_h - 0.08, width=13.333, height=0.08, fill_color="accent", palette=palette)

    add_round_rect(slide, left=0.78, top=band_y + 0.5, width=1.35, height=0.78, fill_color="background", palette=palette, line_color="divider", line_width=0.6)
    add_text(slide, text=number, left=0.86, top=band_y + 0.62, width=1.18, height=0.42, font_size=24, bold=True, color="primary", alignment="center", palette=palette, colors=colors)
    add_local_icon(slide, icon_id_from_text(title + " " + subtitle, fallback="section"), left=11.45, top=band_y + 0.58, size=0.82, palette=palette, with_badge=True)

    if kicker:
        add_text(slide, text=kicker, left=0.82, top=4.12, width=11.8, height=0.34, font_size=13, color="secondary", alignment="left", palette=palette, colors=colors)
    title_top = balanced_band_top(4.1, 1.7, 0.95 + (0.5 if subtitle else 0), min_top=4.45)
    add_text(slide, text=title, left=0.82, top=title_top, width=11.8, height=0.88, font_size=title_font_size(title, base=40, sparse_boost=6, max_size=48), bold=True, color="text", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
    if subtitle:
        add_text(slide, text=subtitle, left=0.86, top=title_top + 0.98, width=10.8, height=0.45, font_size=16, color="secondary", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
    add_source_line(slide, source, palette)
    return prs


def _render_quiet_title(
    prs: Presentation,
    slide,
    palette: str,
    colors: dict,
    source: str,
    number: str,
    title: str,
    subtitle: str,
    kicker: str,
) -> Presentation:
    """Render a quiet divider with centered typography and modest visual anchor."""
    add_local_icon(slide, icon_id_from_text(title + " " + subtitle, fallback="section"), left=6.05, top=1.15, size=0.86, palette=palette, with_badge=True)
    add_text(slide, text=kicker or f"SECTION {number}", left=1.0, top=2.18, width=11.333, height=0.35, font_size=13, color="secondary", alignment="center", palette=palette, colors=colors)
    add_text(slide, text=title, left=1.25, top=2.72, width=10.85, height=1.0, font_size=title_font_size(title, base=42, sparse_boost=6, max_size=50), bold=True, color="text", alignment="center", vertical_alignment="middle", palette=palette, colors=colors)
    if subtitle:
        add_text(slide, text=subtitle, left=2.0, top=3.86, width=9.33, height=0.5, font_size=16, color="secondary", alignment="center", vertical_alignment="middle", palette=palette, colors=colors)
    add_rect(slide, left=5.86, top=4.72, width=1.6, height=0.07, fill_color="accent", palette=palette)
    add_source_line(slide, source, palette)
    return prs
