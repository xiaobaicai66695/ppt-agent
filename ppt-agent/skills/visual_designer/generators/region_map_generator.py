"""Generator for region_map - 区域版图页."""
from typing import Optional, List
from pptx import Presentation
from .base import (
    add_source_line, new_presentation, PALETTES, add_text, add_rect,
    set_slide_background,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "",
    title: str = "{版图标题}",
    subtitle: str = "",
    regions: Optional[List[dict]] = None,
    regions_detail: Optional[List[dict]] = None,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """Generate a region map slide with abstract map and data panel.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        source: Data source annotation.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Subtitle below title.
        regions: List of region dicts with keys: name, value, trend
        regions_detail: Detailed list with keys: name, value, trend, detail

    Returns:
        The Presentation object.
    """
    if regions is None:
        regions = [
            {"name": "华东", "value": "35%", "trend": "+18%"},
            {"name": "华北", "value": "28%", "trend": "+12%"},
            {"name": "华南", "value": "18%", "trend": "+25%"},
        ]

    if regions_detail is None:
        regions_detail = [
            {"name": "华东地区", "value": "35%", "trend": "+18%", "detail": "长三角核心区"},
            {"name": "华北地区", "value": "28%", "trend": "+12%", "detail": "京津冀协同发展"},
            {"name": "华南地区", "value": "18%", "trend": "+25%", "detail": "粤港澳大湾区"},
            {"name": "西南地区", "value": "6%", "trend": "+10%", "detail": "成渝双城经济圈"},
            {"name": "东北地区", "value": "5%", "trend": "+3%", "detail": "老工业基地"},
            {"name": "华中地区", "value": "8%", "trend": "+15%", "detail": "中部崛起战略"},
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    bg_path = resolve_background(background) if background else None
    if bg_path:
        set_image_background(slide, bg_path, brightness=0.95, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    y_t = 0.3
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.2, width=12.0, height=0.2,
            font_size=12, bold=False, color="text_muted", alignment="left", palette=palette, colors=colors)
        y_t = 0.35
    add_text(slide, text=title, left=0.5, top=y_t, width=12.0, height=0.5,
        font_size=32, bold=True, color="text", alignment="left", palette=palette, colors=colors)
    if subtitle:
        add_text(slide, text=subtitle, left=0.5, top=y_t + 0.5, width=12.0, height=0.3,
            font_size=14, bold=False, color="text_muted", alignment="left", palette=palette, colors=colors)

    # Define map regions with positions (abstract simplified map centered on left side)
    # Map occupies left ~60% of slide, right panel is for data
    # Using clean rectangles instead of jittered freeforms
    map_x = 0.4
    map_y = 1.35
    map_w = 6.2
    map_h = 5.2

    # Simplified regional blocks - clean rectangles with labels
    region_blocks = [
        # (left, top, width, height, fill_color, label)
        # 华北 (top-left)
        (map_x + 0.1, map_y + 0.1, 2.4, 1.6, "primary", "华北"),
        # 东北 (top-right)
        (map_x + 2.7, map_y + 0.2, 1.6, 1.3, "divider", "东北"),
        # 华东 (center)
        (map_x + 1.0, map_y + 1.8, 2.2, 1.6, "secondary", "华东"),
        # 华中 (center-right)
        (map_x + 3.4, map_y + 1.6, 1.8, 1.8, "accent", "华中"),
        # 华南 (bottom-right)
        (map_x + 3.3, map_y + 3.5, 2.4, 1.5, "light_bg", "华南"),
        # 西南 (bottom-left)
        (map_x + 0.3, map_y + 3.4, 2.6, 1.5, "accent", "西南"),
    ]

    for (rx, ry, rw, rh, fill, label) in region_blocks:
        add_rect(
            slide,
            left=rx, top=ry, width=rw, height=rh,
            fill_color=fill, palette=palette,
        )
        # Label centered in block
        text_color = "background" if fill in ["primary", "secondary"] else "text"
        add_text(
            slide, text=label,
            left=rx, top=ry + rh / 2 - 0.15, width=rw, height=0.3,
            font_size=12, bold=True, color=text_color,
            alignment="center", palette=palette, colors=colors,
        )

    # Right panel
    add_rect(slide, left=7.0, top=1.4, width=5.8, height=5.4,
             fill_color="background", palette=palette)
    add_rect(slide, left=7.0, top=1.4, width=5.8, height=0.08,
             fill_color="primary", palette=palette)

    add_text(
        slide, text="区域业绩概览",
        left=7.2, top=1.6, width=5.4, height=0.4,
        font_size=16, bold=True, color="text", palette=palette, colors=colors,
    )

    for i, reg in enumerate(regions_detail[:6]):
        y = 2.15 + i * 0.78
        add_rect(slide, left=7.2, top=y, width=5.4, height=0.7,
                 fill_color="light_bg", palette=palette)
        add_rect(slide, left=7.2, top=y, width=0.06, height=0.7,
                 fill_color="primary", palette=palette)

        add_text(
            slide, text=reg.get("name", ""),
            left=7.4, top=y + 0.1, width=1.5, height=0.3,
            font_size=12, bold=True, color="text", palette=palette, colors=colors,
        )
        add_text(
            slide, text=reg.get("value", ""),
            left=9.0, top=y + 0.08, width=1.0, height=0.35,
            font_size=14, bold=True, color="primary", palette=palette, colors=colors,
        )
        add_text(
            slide, text=reg.get("trend", ""),
            left=10.1, top=y + 0.1, width=0.8, height=0.3,
            font_size=11, bold=True, color="secondary", palette=palette, colors=colors,
        )
        add_text(
            slide, text=reg.get("detail", ""),
            left=7.4, top=y + 0.4, width=4.8, height=0.25,
            font_size=10, color="text_muted", palette=palette, colors=colors,
        )

    add_source_line(slide, source, palette)
    return prs
