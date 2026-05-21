"""Generator for process_flow - 流程图 (圆角卡片 + 曲线连接)."""
from typing import Optional, List, Dict
from pptx import Presentation
from .base import (
    add_source_line, new_presentation,
    PALETTES, add_text, add_round_rect, add_ellipse,
    set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "",
    title: str = "{流程标题}", direction: str = "horizontal_zigzag",
    steps: List[Dict[str, str]] = None, kicker: str = "", subtitle: str = "",
) -> Presentation:
    if steps is None:
        steps = [{"num":"01","title":"{步骤}","desc":"{描述}"}] * 5
    if prs is None: prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]; slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    n = min(len(steps), 6)
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    y_t = 0.22
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.06, width=12.0, height=0.16,
            font_size=9, bold=False, color="text_muted", alignment="left", palette=palette)
        y_t = 0.18
    add_text(slide, text=title, left=0.5, top=y_t, width=12.0, height=0.38,
        font_size=24, bold=True, color="text", alignment="left", palette=palette)

    flow_y = y_t + 0.55
    if subtitle:
        add_text(slide, text=subtitle, left=0.5, top=y_t + 0.34, width=12.0, height=0.2,
            font_size=10, bold=False, color="secondary", alignment="left", palette=palette)
        flow_y = y_t + 0.65

    accent_list = ["primary", "secondary", "accent", "primary", "secondary", "accent"]

    if direction == "horizontal_zigzag" or direction == "horizontal":
        # Two rows, rounded cards linked by curved paths (ellipse dots)
        row1_n = (n + 1) // 2  # top row
        row2_n = n - row1_n   # bottom row

        box_w = 3.5; box_h = 1.0
        gap_x = 0.45; gap_y = 0.55
        row1_w = row1_n * box_w + (row1_n - 1) * gap_x
        row2_w = row2_n * box_w + (row2_n - 1) * gap_x
        start_x1 = (13.333 - row1_w) / 2
        start_x2 = (13.333 - row2_w) / 2
        y1 = flow_y; y2 = flow_y + box_h + gap_y

        for i, step in enumerate(steps[:row1_n]):
            x = start_x1 + i * (box_w + gap_x)
            _draw_card(slide, palette, x, y1, box_w, box_h, step, accent_list[i % len(accent_list)])
            # Connector dot between cards
            if i < row1_n - 1:
                cx = x + box_w + gap_x / 2; cy = y1 + box_h / 2
                add_ellipse(slide, left=cx - 0.07, top=cy - 0.07, width=0.14, height=0.14,
                    fill_color="accent", palette=palette)

        # Curved transition: arc of 3 small dots from row1 end to row2 start
        last_x1 = start_x1 + (row1_n - 1) * (box_w + gap_x)
        tx = last_x1 + box_w / 2; ty = y1 + box_h
        bx = start_x2 + box_w / 2; by = y2
        # Vertical dots forming a soft S-curve guide
        for dot_i in range(3):
            frac = (dot_i + 1) / 4
            dx = tx + (bx - tx) * frac
            dy = ty + (by - ty) * frac
            add_ellipse(slide, left=dx - 0.05, top=dy - 0.05, width=0.10, height=0.10,
                fill_color="divider", palette=palette)

        for i, step in enumerate(steps[row1_n:n]):
            x = start_x2 + i * (box_w + gap_x)
            _draw_card(slide, palette, x, y2, box_w, box_h, step, accent_list[(row1_n + i) % len(accent_list)])
            if i < row2_n - 1:
                cx = x + box_w + gap_x / 2; cy = y2 + box_h / 2
                add_ellipse(slide, left=cx - 0.07, top=cy - 0.07, width=0.14, height=0.14,
                    fill_color="accent", palette=palette)

    else:  # vertical
        box_w = 9.5; box_h = 0.78; gap = 0.18
        start_x = 2.0
        for i, step in enumerate(steps[:n]):
            y = flow_y + i * (box_h + gap)
            _draw_card(slide, palette, start_x, y, box_w, box_h, step, accent_list[i % len(accent_list)])
            if i < n - 1:
                cy = y + box_h + gap / 2
                add_ellipse(slide, left=start_x + box_w / 2 - 0.05, top=cy - 0.05,
                    width=0.10, height=0.10, fill_color="accent", palette=palette)

    add_source_line(slide, source, palette)
    return prs


def _draw_card(slide, palette, x, y, w, h, step, accent_c):
    """Draw a single rounded process step card."""
    add_round_rect(slide, left=x, top=y, width=w, height=h,
        fill_color="light_bg", palette=palette)
    # Soft accent strip
    add_round_rect(slide, left=x, top=y, width=w, height=0.04,
        fill_color=accent_c, palette=palette)

    num = step.get("num", "")
    title_text = step.get("title", "")
    desc_text = step.get("desc", "")

    add_text(slide, text=num,
        left=x + 0.12, top=y + 0.12, width=0.5, height=0.35,
        font_size=16, bold=True, color=accent_c, alignment="left", palette=palette)
    add_text(slide, text=title_text,
        left=x + 0.6, top=y + 0.12, width=w - 0.8, height=0.35,
        font_size=13, bold=True, color="text", alignment="left", palette=palette)
    add_text(slide, text=desc_text,
        left=x + 0.15, top=y + 0.5, width=w - 0.3, height=h - 0.55,
        font_size=9, bold=False, color="text_muted", alignment="left", palette=palette)
