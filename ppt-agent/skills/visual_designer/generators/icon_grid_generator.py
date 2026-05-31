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
            {"icon": "研", "label": "基础研究", "color": "primary"},
            {"icon": "算", "label": "算力平台", "color": "secondary"},
            {"icon": "数", "label": "数据治理", "color": "accent"},
            {"icon": "模", "label": "模型训练", "color": "primary"},
            {"icon": "工", "label": "工程落地", "color": "secondary"},
            {"icon": "安", "label": "安全合规", "color": "accent"},
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
        font_size=30, bold=True,
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
    icon_size = min(cell_width, cell_height) * 0.5
    icon_size = min(icon_size, 1.0)  # Cap at 1 inch

    for idx, icon_data in enumerate(visible_icons):
        row = idx // cols
        col = idx % cols

        cell_x = margin_x + col * (cell_width + gap_x)
        cell_y = margin_top + row * (cell_height + gap_y)

        icon_text = icon_data.get("icon", "?")
        label = icon_data.get("label", "")
        color_key = icon_data.get("color", ["primary", "secondary", "accent"][idx % 3])

        # Icon background circle
        icon_center_x = cell_x + cell_width / 2
        icon_center_y = cell_y + cell_height * 0.35

        add_round_rect(
            slide,
            left=icon_center_x - icon_size / 2,
            top=icon_center_y - icon_size / 2,
            width=icon_size,
            height=icon_size,
            fill_color=color_key, palette=palette,
        )

        # Icon text
        icon_font_size = int(icon_size * 24)  # Scale font to icon size
        icon_font_size = min(max(icon_font_size, 18), 36)  # Clamp

        add_text(
            slide,
            text=icon_text,
            left=icon_center_x - icon_size / 2,
            top=icon_center_y - icon_size / 2,
            width=icon_size,
            height=icon_size,
            font_size=icon_font_size,
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
