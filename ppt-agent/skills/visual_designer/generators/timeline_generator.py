"""Timeline generator for timeline slide type."""
from typing import Optional, List
from pptx import Presentation

from .base import add_source_line, new_presentation, PALETTES, add_text, add_rect, add_ellipse, set_slide_background, resolve_background, set_image_background
from .asset_manager import add_local_icon, asset_path, icon_id_from_text
from .layout_intelligence import title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "",
    title: str = "{页面标题}",
    subtitle: str = "",
    direction: str = "horizontal",
    nodes: Optional[List[dict]] = None,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a timeline slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Subtitle below title.
        direction: "horizontal" or "vertical".
        nodes: List of dicts with keys: year, event, icon.

    Returns:
        The Presentation object.
    """
    if nodes is None:
        nodes = [
            {"year": "2020", "event": "里程碑1", "icon": "01"},
            {"year": "2022", "event": "里程碑2", "icon": "02"},
            {"year": "2024", "event": "里程碑3", "icon": "03"},
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
    y_offset = 0.5
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.7, top=0.3, width=11.5, height=0.35,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset = 0.7

    # Title
    add_text(
        slide,
        text=title,
        left=0.7, top=y_offset, width=11.5, height=0.7,
        font_size=title_font_size(title, base=36, sparse_boost=5, max_size=44), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )
    y_offset += 0.75

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.7, top=y_offset, width=11.5, height=0.4,
            font_size=16, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.5

    # Timeline axis, centered in the remaining content band.
    content_top = y_offset + 0.55
    content_bottom = 6.45
    axis_y = content_top + (content_bottom - content_top) * 0.46
    add_rect(
        slide,
        left=0.7, top=axis_y, width=11.5, height=0.05,
        fill_color="primary", palette=palette,
    )

    # Timeline nodes
    node_count = max(1, len(nodes))
    spacing = 11.5 / (node_count + 1)

    for i, node in enumerate(nodes):
        node_center_x = 0.7 + spacing * (i + 1)
        node_x = node_center_x - 0.15
        year = node.get("year", "")
        event = node.get("event", "")
        icon = node.get("icon", str(i + 1))

        # Node marker
        add_ellipse(
            slide,
            left=node_center_x - 0.11, top=axis_y - 0.11, width=0.22, height=0.22,
            fill_color="accent", palette=palette,
        )

        # Year label above
        add_text(
            slide,
            text=year,
            left=node_center_x - 0.55, top=axis_y - 0.76, width=1.1, height=0.36,
            font_size=15, bold=True,
            color="primary", alignment="center",
            palette=palette,
            colors=colors,
        )

        # Event description below, with room for natural wrapping.
        add_text(
            slide,
            text=event,
            left=node_center_x - 0.72, top=axis_y + 0.42, width=1.44, height=1.08,
            font_size=12.5, bold=False,
            color="text", alignment="center",
            line_spacing=0.92,
            palette=palette,
            colors=colors,
        )

        # Stable icon/number badge; avoid tiny text squeezed into a square.
        add_rect(
            slide,
            left=node_center_x - 0.24, top=axis_y - 0.28, width=0.48, height=0.48,
            fill_color="light_bg", palette=palette,
            line_color="divider",
            line_width=0.6,
        )
        if icon and asset_path(icon):
            add_local_icon(slide, icon, left=node_center_x - 0.17, top=axis_y - 0.21, size=0.34, palette=palette)
        else:
            icon_text = str(icon or i + 1)
            if not icon_text.isdigit() and len(icon_text) > 2:
                semantic = icon_id_from_text(icon_text + event, fallback="")
                if semantic and asset_path(semantic):
                    add_local_icon(slide, semantic, left=node_center_x - 0.17, top=axis_y - 0.21, size=0.34, palette=palette)
                    continue
            add_text(
                slide,
                text=f"{i + 1:02d}" if icon_text.isdigit() else icon_text[:2],
                left=node_center_x - 0.22, top=axis_y - 0.25, width=0.44, height=0.38,
                font_size=10.5, bold=True,
                color="primary", alignment="center",
                vertical_alignment="middle",
                palette=palette,
                colors=colors,
            )

    add_source_line(slide, source, palette)
    return prs
