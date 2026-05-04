"""Generator for timeline - 时间轴."""
from typing import Optional, List, Dict

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_ellipse,
    set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{发展历程标题}",
    direction: str = "horizontal",
    nodes: List[Dict[str, str]] = None,
    kicker: str = "",
    subtitle: str = "",
) -> Presentation:
    """
    Generate a timeline slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        direction: "horizontal" or "vertical".
        nodes: List of dicts with keys: year, event, icon (optional).
        kicker: Small label above title (e.g. "技术演进").
        subtitle: Optional subtitle below title.

    Returns:
        The Presentation object.
    """
    if nodes is None:
        nodes = [
            {"year": "{年份}", "event": "{事件描述}", "icon": "01"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "02"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "03"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "04"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "05"},
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
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
        )
        y_title = 0.35

    # Title
    add_text(
        slide,
        text=title,
        left=0.5, top=y_title, width=12.0, height=0.7,
        font_size=36, bold=True,
        color="text", alignment="left",
        palette=palette,
    )

    # Subtitle
    y_axis = 4.0
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.5, top=y_title + 0.65, width=12.0, height=0.4,
            font_size=16, bold=False,
            color="secondary", alignment="left",
            palette=palette,
        )
        y_axis = 4.5

    n = len(nodes)
    visible_nodes = nodes[:5]

    if direction == "horizontal":
        # Timeline horizontal axis
        axis_y = y_axis
        axis_x_start = 1.2
        axis_x_end = 12.0
        spacing = (axis_x_end - axis_x_start) / (n - 1) if n > 1 else 0

        # Main axis line
        add_rect(
            slide,
            left=axis_x_start, top=axis_y - 0.02, width=axis_x_end - axis_x_start, height=0.04,
            fill_color="primary", palette=palette,
        )

        for i, node in enumerate(visible_nodes):
            x = axis_x_start + i * spacing if n > 1 else (axis_x_start + axis_x_end) / 2
            year = node.get("year", "")
            event = node.get("event", "")
            icon = node.get("icon", "")

            # Circle node on axis
            add_ellipse(
                slide,
                left=x - 0.15, top=axis_y - 0.15, width=0.3, height=0.3,
                fill_color="accent", palette=palette,
            )

            # Alternating: odd nodes above, even nodes below
            if i % 2 == 0:
                # Year label above
                add_text(
                    slide,
                    text=year,
                    left=x - 0.5, top=axis_y - 1.3, width=1.0, height=0.4,
                    font_size=14, bold=True,
                    color="primary", alignment="center",
                    palette=palette,
                )
                # Event below
                add_text(
                    slide,
                    text=event,
                    left=x - 0.7, top=axis_y + 0.4, width=1.4, height=1.2,
                    font_size=13, bold=False,
                    color="text", alignment="center",
                    palette=palette,
                )
                # Connector line
                add_rect(
                    slide,
                    left=x - 0.01, top=axis_y - 1.0, width=0.02, height=0.85,
                    fill_color="divider", palette=palette,
                )
            else:
                # Year label below
                add_text(
                    slide,
                    text=year,
                    left=x - 0.5, top=axis_y + 0.3, width=1.0, height=0.4,
                    font_size=14, bold=True,
                    color="primary", alignment="center",
                    palette=palette,
                )
                # Event above
                add_text(
                    slide,
                    text=event,
                    left=x - 0.7, top=axis_y - 2.0, width=1.4, height=1.2,
                    font_size=13, bold=False,
                    color="text", alignment="center",
                    palette=palette,
                )
                # Connector line
                add_rect(
                    slide,
                    left=x - 0.01, top=axis_y - 0.7, width=0.02, height=0.85,
                    fill_color="divider", palette=palette,
                )
    else:
        # Vertical timeline
        axis_x = 2.0
        axis_y_start = 1.5
        axis_y_end = 6.5
        spacing = (axis_y_end - axis_y_start) / (n - 1) if n > 1 else 0

        # Main axis line
        add_rect(
            slide,
            left=axis_x - 0.02, top=axis_y_start, width=0.04, height=axis_y_end - axis_y_start,
            fill_color="primary", palette=palette,
        )

        for i, node in enumerate(visible_nodes):
            y = axis_y_start + i * spacing if n > 1 else (axis_y_start + axis_y_end) / 2
            year = node.get("year", "")
            event = node.get("event", "")

            # Circle node
            add_ellipse(
                slide,
                left=axis_x - 0.15, top=y - 0.15, width=0.3, height=0.3,
                fill_color="accent", palette=palette,
            )

            # Year label left
            add_text(
                slide,
                text=year,
                left=0.5, top=y - 0.25, width=1.2, height=0.5,
                font_size=14, bold=True,
                color="primary", alignment="right",
                palette=palette,
            )

            # Event text right
            add_text(
                slide,
                text=event,
                left=axis_x + 0.4, top=y - 0.3, width=9.5, height=0.8,
                font_size=14, bold=False,
                color="text", alignment="left",
                palette=palette,
            )

    add_source_line(slide, source, palette)
    return prs
