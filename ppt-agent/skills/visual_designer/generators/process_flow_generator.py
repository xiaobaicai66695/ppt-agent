"""Generator for process_flow - 步骤流程图."""
import random
from typing import Optional, List, Dict

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_ellipse, add_line,
    set_slide_background,
    resolve_background, set_image_background,
)
from .layout_intelligence import balanced_band_top, title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{流程标题}",
    direction: str = "horizontal_zigzag",
    steps: List[Dict[str, str]] = None,
    kicker: str = "",
    subtitle: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a process flow slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        direction: "horizontal", "horizontal_zigzag", or "vertical".
        steps: List of dicts with keys: num, title, desc.
        kicker: Small label above title (e.g. "工程实践").
        subtitle: Optional subtitle below title.

    Returns:
        The Presentation object.
    """
    if steps is None:
        steps = [
            {"num": "01", "title": "{步骤1}", "desc": "{描述}"},
            {"num": "02", "title": "{步骤2}", "desc": "{描述}"},
            {"num": "03", "title": "{步骤3}", "desc": "{描述}"},
            {"num": "04", "title": "{步骤4}", "desc": "{描述}"},
            {"num": "05", "title": "{步骤5}", "desc": "{描述}"},
            {"num": "06", "title": "{步骤6}", "desc": "{描述}"},
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
        font_size=title_font_size(title, base=36, sparse_boost=5, max_size=44), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Subtitle
    y_flow = 1.8
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.5, top=y_title + 0.65, width=12.0, height=0.4,
            font_size=16, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_flow = 2.3
    else:
        y_flow = 1.8

    visible_steps = steps[:6]

    if direction == "horizontal":
        n = len(visible_steps)
        box_w = 1.8
        box_h = 1.4
        gap = 0.3
        total_w = n * box_w + (n - 1) * gap
        start_x = (13.333 - total_w) / 2
        axis_y = balanced_band_top(y_flow, 4.45, box_h, min_top=y_flow) + box_h / 2

        for i, step in enumerate(visible_steps):
            x = start_x + i * (box_w + gap)
            num = step.get("num", "")
            step_title = step.get("title", "")
            step_desc = step.get("desc", "")

            # Step box
            add_rect(
                slide,
                left=x, top=axis_y - box_h / 2, width=box_w, height=box_h,
                fill_color="light_bg", palette=palette,
            )
            # Left accent bar
            add_rect(
                slide,
                left=x, top=axis_y - box_h / 2, width=0.08, height=box_h,
                fill_color="primary", palette=palette,
            )

            # Step number
            add_text(
                slide,
                text=num,
                left=x + 0.2, top=axis_y - box_h / 2 + 0.15, width=0.6, height=0.4,
                font_size=18, bold=True,
                color="primary", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Step title
            add_text(
                slide,
                text=step_title,
                left=x + 0.1, top=axis_y - box_h / 2 + 0.55, width=box_w - 0.2, height=0.4,
                font_size=14, bold=True,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Step desc
            add_text(
                slide,
                text=step_desc,
                left=x + 0.1, top=axis_y - box_h / 2 + 0.95, width=box_w - 0.2, height=0.4,
                font_size=11, bold=False,
                color="text_muted", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Arrow between steps
            if i < n - 1:
                arrow_x = x + box_w + 0.05
                arrow_y = axis_y - 0.15
                add_text(
                    slide,
                    text=">",
                    left=arrow_x, top=arrow_y, width=gap - 0.1, height=0.3,
                    font_size=18, bold=True,
                    color="secondary", alignment="center",
                    palette=palette,
                    colors=colors,
                )

    elif direction == "horizontal_zigzag":
        # Two rows of 3 steps
        n = len(visible_steps)
        cols = 3
        box_w = 3.0
        box_h = 1.2
        gap_x = 0.5
        gap_y = 0.6
        start_x = 1.0
        total_h = box_h * 2 + gap_y
        start_y1 = balanced_band_top(y_flow, 4.2, total_h, min_top=y_flow)
        start_y2 = start_y1 + box_h + gap_y

        row1 = visible_steps[:3]
        row2 = visible_steps[3:6]

        def draw_step(step_obj, x, y, idx):
            num = step_obj.get("num", "")
            step_title = step_obj.get("title", "")
            step_desc = step_obj.get("desc", "")

            # Box
            add_rect(
                slide,
                left=x, top=y, width=box_w, height=box_h,
                fill_color="light_bg", palette=palette,
            )
            # Left accent bar
            add_rect(
                slide,
                left=x, top=y, width=0.08, height=box_h,
                fill_color="primary", palette=palette,
            )

            # Number
            add_text(
                slide,
                text=num,
                left=x + 0.2, top=y + 0.1, width=0.6, height=0.4,
                font_size=18, bold=True,
                color="primary", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Title
            add_text(
                slide,
                text=step_title,
                left=x + 0.8, top=y + 0.1, width=box_w - 0.9, height=0.4,
                font_size=14, bold=True,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Desc
            add_text(
                slide,
                text=step_desc,
                left=x + 0.2, top=y + 0.6, width=box_w - 0.3, height=0.5,
                font_size=12, bold=False,
                color="text_muted", alignment="left",
                palette=palette,
                colors=colors,
            )

        # Row 1 (left to right)
        for i, step in enumerate(row1):
            x = start_x + i * (box_w + gap_x)
            draw_step(step, x, start_y1, i)

        # Arrows for row 1 (left to right)
        for i in range(len(row1) - 1):
            ax = start_x + (i + 1) * (box_w + gap_x) - gap_x - 0.05
            add_text(
                slide,
                text=">",
                left=ax, top=start_y1 + box_h / 2 - 0.2, width=gap_x, height=0.4,
                font_size=16, bold=True,
                color="secondary", alignment="center",
                palette=palette,
                colors=colors,
            )

        # Connector down arrow
        add_text(
            slide,
            text="v",
            left=start_x + 2 * (box_w + gap_x) + box_w / 2 - 0.15,
            top=start_y1 + box_h + 0.05,
            width=0.3, height=gap_y - 0.1,
            font_size=16, bold=True,
            color="secondary", alignment="center",
            palette=palette,
            colors=colors,
        )

        # Row 2 (right to left - zigzag)
        for i, step in enumerate(row2):
            x = start_x + (cols - 1 - i) * (box_w + gap_x)
            draw_step(step, x, start_y2, i + 3)

        # Arrows for row 2 (right to left)
        for i in range(len(row2) - 1):
            ax = start_x + (cols - 2 - i) * (box_w + gap_x) + box_w + 0.05
            add_text(
                slide,
                text="<",
                left=ax, top=start_y2 + box_h / 2 - 0.2, width=gap_x, height=0.4,
                font_size=16, bold=True,
                color="secondary", alignment="center",
                palette=palette,
                colors=colors,
            )

    else:
        # Vertical
        n = len(visible_steps)
        box_w = 9.0
        box_h = 0.9
        gap = 0.25
        start_x = 2.0
        total_h = n * box_h + max(n - 1, 0) * gap
        start_y = balanced_band_top(y_flow, 4.55, total_h, min_top=y_flow)

        for i, step in enumerate(visible_steps):
            y = start_y + i * (box_h + gap)
            num = step.get("num", "")
            step_title = step.get("title", "")
            step_desc = step.get("desc", "")

            # Box
            add_rect(
                slide,
                left=start_x, top=y, width=box_w, height=box_h,
                fill_color="light_bg", palette=palette,
            )
            # Left accent bar
            add_rect(
                slide,
                left=start_x, top=y, width=0.08, height=box_h,
                fill_color="primary", palette=palette,
            )

            # Number
            add_text(
                slide,
                text=num,
                left=start_x + 0.2, top=y + 0.15, width=0.6, height=0.4,
                font_size=16, bold=True,
                color="primary", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Title
            add_text(
                slide,
                text=step_title,
                left=start_x + 0.8, top=y + 0.15, width=3.0, height=0.4,
                font_size=14, bold=True,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Desc
            add_text(
                slide,
                text=step_desc,
                left=start_x + 4.0, top=y + 0.15, width=4.8, height=0.6,
                font_size=12, bold=False,
                color="text_muted", alignment="left",
                palette=palette,
                colors=colors,
            )

            # Down arrow
            if i < n - 1:
                add_text(
                    slide,
                    text="v",
                    left=start_x + box_w / 2 - 0.1, top=y + box_h + 0.02, width=0.3, height=gap - 0.04,
                    font_size=14, bold=True,
                    color="secondary", alignment="center",
                    palette=palette,
                    colors=colors,
                )

    add_source_line(slide, source, palette)
    return prs
