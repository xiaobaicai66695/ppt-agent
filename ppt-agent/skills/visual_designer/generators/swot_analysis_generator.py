"""SWOT analysis generator for SWOT analysis slide type."""
from typing import Optional, Dict, List
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
    kicker: str = "",
    title: str = "{SWOT分析标题}",
    subtitle: str = "",
    swot: Optional[Dict] = None,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a SWOT analysis slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Optional subtitle.
        swot: Dict with strengths, weaknesses, opportunities, threats.
              Each contains label, label_full, color, items (list).

    Returns:
        The Presentation object.
    """
    if swot is None:
        swot = {
            "strengths": {
                "label": "S",
                "label_full": "优势",
                "items": ["技术领先，算法准确率高", "团队经验丰富", "数据积累深厚"]
            },
            "weaknesses": {
                "label": "W",
                "label_full": "劣势",
                "items": ["成本高，定价缺乏竞争力", "品牌知名度不足", "销售渠道单一"]
            },
            "opportunities": {
                "label": "O",
                "label_full": "机会",
                "items": ["市场需求快速增长", "政策支持AI发展", "行业标准尚未成熟"]
            },
            "threats": {
                "label": "T",
                "label_full": "威胁",
                "items": ["大厂入局，竞争加剧", "技术迭代快", "数据合规要求趋严"]
            }
        }

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
    y_offset = 0.3
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
        y_offset = 0.25

    # Title
    add_rect(
        slide,
        left=0.4, top=y_offset + 0.05, width=0.08, height=0.5,
        fill_color="primary", palette=palette,
    )
    add_text(
        slide,
        text=title,
        left=0.55, top=y_offset, width=11.5, height=0.6,
        font_size=28, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )
    y_offset += 0.6

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.55, top=y_offset, width=11.5, height=0.35,
            font_size=14, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.35

    # SWOT Grid layout
    grid_top = y_offset + 0.25
    grid_left = 0.5
    grid_width = 12.0
    grid_height = 5.2
    gap = 0.1

    cell_width = (grid_width - gap) / 2
    cell_height = (grid_height - gap) / 2

    # Define quadrants: (row, col) -> key
    quadrant_map = [
        [("strengths", "S", "优势", "primary"),
         ("weaknesses", "W", "劣势", "secondary")],
        [("opportunities", "O", "机会", "accent"),
         ("threats", "T", "威胁", "warm")]
    ]

    swot_colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    # 玻璃模式下，象限色块用深蓝版本（_fill 变体）保证对比度
    if "primary_fill" in colors:
        swot_colors["primary"] = colors.get("primary_fill", swot_colors.get("primary"))
        swot_colors["secondary"] = colors.get("secondary_fill", swot_colors.get("secondary"))
        swot_colors["accent"] = colors.get("accent_fill", swot_colors.get("accent"))

    for row in range(2):
        for col in range(2):
            key, letter, label_full, color_key = quadrant_map[row][col]
            raw = swot.get(key, {"items": []})
            # Support both dict (expected) and list (LLM may pass raw items)
            if isinstance(raw, list):
                items = raw[:5]
            else:
                items = raw.get("items", []) if isinstance(raw, dict) else []

            cell_x = grid_left + col * (cell_width + gap)
            cell_y = grid_top + row * (cell_height + gap)

            # Background
            if color_key == "warm":
                bg_color = swot_colors.get("text", "8B4513")  # Brownish for threats
            else:
                bg_color = swot_colors.get(color_key, swot_colors.get("primary", "5A8AA8"))

            # Draw quadrant background
            add_rect(
                slide,
                left=cell_x, top=cell_y,
                width=cell_width, height=cell_height,
                fill_color=bg_color, palette=palette,
            )

            # Letter (huge)
            add_text(
                slide,
                text=letter,
                left=cell_x + 0.2, top=cell_y + 0.15,
                width=1.0, height=0.9,
                font_size=60, bold=True,
                color="background", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Full label
            add_text(
                slide,
                text=label_full,
                left=cell_x + 1.1, top=cell_y + 0.4,
                width=1.5, height=0.5,
                font_size=14, bold=True,
                color="background", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Items
            item_start_y = cell_y + 1.2
            item_spacing = 0.5
            for i, item in enumerate(items[:5]):
                # Bullet point
                add_rect(
                    slide,
                    left=cell_x + 0.25, top=item_start_y + i * item_spacing + 0.08,
                    width=0.12, height=0.12,
                    fill_color="background", palette=palette,
                )
                # Item text
                add_text(
                    slide,
                    text=item,
                    left=cell_x + 0.45, top=item_start_y + i * item_spacing,
                    width=cell_width - 0.6, height=0.45,
                    font_size=13, bold=False,
                    color="background", alignment="left",
                    palette=palette,
                    colors=colors,
                )

    add_source_line(slide, source, palette)
    return prs
