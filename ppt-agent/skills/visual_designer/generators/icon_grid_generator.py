"""Icon grid generator for icon grid slide type."""
from typing import Optional, List, Dict
from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_round_rect,
    set_slide_background,
    resolve_background, set_image_background,
)
from .asset_manager import add_local_icon, asset_path, icon_id_from_text
from .layout_intelligence import title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "",
    title: str = "{图标网格标题}",
    subtitle: str = "",
    layout: str = "3x2",
    icons: Optional[List[Dict]] = None,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate an icon grid slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Optional subtitle.
        layout: Grid layout, e.g. "3x2", "3x3", "2x3".
        icons: List of dicts with keys: icon (single char/text), label, color (optional).

    Returns:
        The Presentation object.
    """
    if icons is None:
        icons = [
            {"icon": "source", "label": "基础研究", "color": "primary"},
            {"icon": "runtime", "label": "算力平台", "color": "secondary"},
            {"icon": "density", "label": "数据治理", "color": "accent"},
            {"icon": "llm", "label": "模型训练", "color": "primary"},
            {"icon": "tool", "label": "工程落地", "color": "secondary"},
            {"icon": "review", "label": "安全合规", "color": "accent"},
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

    # Kicker
    y_offset = 0.4
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.6, top=0.15, width=12.0, height=0.3,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset = 0.35

    # Title with left accent bar
    add_rect(
        slide,
        left=0.5, top=y_offset + 0.05, width=0.08, height=0.55,
        fill_color="primary", palette=palette,
    )
    add_text(
        slide,
        text=title,
        left=0.7, top=y_offset, width=11.5, height=0.65,
        font_size=title_font_size(title, base=32, sparse_boost=6, max_size=42), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )
    y_offset += 0.7

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.7, top=y_offset, width=11.5, height=0.35,
            font_size=14, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.4

    # Parse layout
    rows, cols = map(int, layout.split("x"))
    visible_icons = icons[:rows * cols]

    # Grid dimensions
    margin_x = 0.8
    margin_top = y_offset + 0.3
    grid_width = 12.0 - margin_x * 2
    grid_height = 5.0
    gap_x = 0.4
    gap_y = 0.5

    cell_width = (grid_width - gap_x * (cols - 1)) / cols
    cell_height = (grid_height - gap_y * (rows - 1)) / rows

    # Icon and label dimensions
    icon_size = min(cell_width, cell_height) * 0.62
    icon_size = min(icon_size, 1.22)  # Cap at 1.22 inch

    for idx, icon_data in enumerate(visible_icons):
        row = idx // cols
        col = idx % cols

        cell_x = margin_x + col * (cell_width + gap_x)
        cell_y = margin_top + row * (cell_height + gap_y)

        icon_text = icon_data.get("icon", "?")
        label = icon_data.get("label", "")
        color_key = icon_data.get("color", ["primary", "secondary", "accent"][idx % 3])

        # Icon background
        icon_center_x = cell_x + cell_width / 2
        icon_center_y = cell_y + cell_height * 0.35
        icon_id = icon_text if icon_text and asset_path(icon_text) else icon_id_from_text(label, fallback="primitive")
        if not add_local_icon(
            slide,
            icon_id,
            left=icon_center_x - icon_size / 2,
            top=icon_center_y - icon_size / 2,
            size=icon_size,
            palette=palette,
            with_badge=True,
        ):
            add_round_rect(
                slide,
                left=icon_center_x - icon_size / 2,
                top=icon_center_y - icon_size / 2,
                width=icon_size,
                height=icon_size,
                fill_color=color_key, palette=palette,
            )
            add_text(
                slide,
                text=icon_text[:2],
                left=icon_center_x - icon_size / 2,
                top=icon_center_y - icon_size / 2,
                width=icon_size,
                height=icon_size,
                font_size=24,
                bold=True,
                color="background", alignment="center",
                palette=palette,
                colors=colors,
            )

        # Label below icon
        label_y = icon_center_y + icon_size / 2 + 0.15
        add_text(
            slide,
            text=label,
            left=cell_x,
            top=label_y,
            width=cell_width,
            height=0.4,
            font_size=14,
            bold=False,
            color="text", alignment="center",
            palette=palette,
            colors=colors,
        )

    add_source_line(slide, source, palette)
    return prs
