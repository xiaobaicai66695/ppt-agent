"""Timeline generator for timeline slide type."""
from typing import Optional, List
from pptx import Presentation

from .base import add_source_line, new_presentation, PALETTES, add_text, add_rect, set_slide_background, resolve_background, set_image_background


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
        font_size=36, bold=True,
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

    # Timeline axis
    axis_y = y_offset + 1.5
    add_rect(
        slide,
        left=0.7, top=axis_y, width=11.5, height=0.05,
        fill_color="primary", palette=palette,
    )

    # Timeline nodes
    node_count = len(nodes)
    spacing = 11.5 / (node_count + 1)

    for i, node in enumerate(nodes):
        node_x = 0.7 + spacing * (i + 1) - 0.15
        year = node.get("year", "")
        event = node.get("event", "")
        icon = node.get("icon", str(i + 1))

        # Node circle
        add_rect(
            slide,
            left=node_x, top=axis_y - 0.1, width=0.25, height=0.25,
            fill_color="accent", palette=palette,
        )

        # Year label above
        add_text(
            slide,
            text=year,
            left=node_x - 0.3, top=axis_y - 0.55, width=0.85, height=0.35,
            font_size=14, bold=True,
            color="primary", alignment="center",
            palette=palette,
            colors=colors,
        )

        # Event description below
        add_text(
            slide,
            text=event,
            left=node_x - 0.5, top=axis_y + 0.35, width=1.25, height=0.8,
            font_size=12, bold=False,
            color="text", alignment="center",
            palette=palette,
            colors=colors,
        )

        # Icon number
        add_rect(
            slide,
            left=node_x - 0.05, top=axis_y - 0.08, width=0.2, height=0.2,
            fill_color="light_bg", palette=palette,
        )
        add_text(
            slide,
            text=icon,
            left=node_x - 0.05, top=axis_y - 0.08, width=0.2, height=0.2,
            font_size=10, bold=True,
            color="text", alignment="center",
            palette=palette,
            colors=colors,
        )

    add_source_line(slide, source, palette)
    return prs
