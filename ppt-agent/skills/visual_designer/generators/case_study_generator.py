"""Generator for case_study - 案例研究页.

Structured Context→Problem→Solution→Results layout with
four distinct card regions, each with labeled header and accent treatment.
"""
from typing import List, Optional

from pptx import Presentation

from .base import (
    add_source_line,
    PALETTES,
    add_rect,
    add_round_rect,
    add_text,
    new_presentation,
    set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "案例 · {领域}",
    title: str = "{案例名称}",
    context: str = "{背景}",
    problem: str = "{痛点}",
    solution: str = "{解决方案}",
    results: Optional[List[dict]] = None,
) -> Presentation:
    """Generate a case_study slide with Context/Problem/Solution/Results."""
    if results is None:
        results = [
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Kicker
    add_text(
        slide, text=kicker,
        left=0.5, top=0.2, width=12.0, height=0.3,
        font_size=12, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
    )

    # Title
    add_text(
        slide, text=title,
        left=0.5, top=0.45, width=12.0, height=0.55,
        font_size=30, bold=True,
        color="text", alignment="left",
        palette=palette,
    )

    # ── Four stacked cards ──
    card_left = 0.5
    card_w = 12.333
    label_w = 2.0
    content_left = card_left + label_w + 0.3
    content_w = card_w - label_w - 0.5

    cur_y = 1.15

    # Card 1: Context
    card_h = 0.85
    add_round_rect(
        slide,
        left=card_left, top=cur_y, width=card_w, height=card_h,
        fill_color="light_bg", palette=palette,
    )
    add_text(
        slide, text="Context · 背景",
        left=card_left + 0.2, top=cur_y + 0.08, width=label_w, height=0.35,
        font_size=13, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )
    add_text(
        slide, text=context,
        left=content_left, top=cur_y + 0.08, width=content_w, height=card_h - 0.15,
        font_size=13, bold=False,
        color="text", alignment="left",
        palette=palette,
    )
    cur_y += card_h + 0.12

    # Card 2: Problem (accent border)
    card_h = 0.85
    add_round_rect(
        slide,
        left=card_left, top=cur_y, width=card_w, height=card_h,
        fill_color="light_bg", palette=palette,
        line_color="secondary", line_width=1.5,
    )
    add_text(
        slide, text="Problem · 痛点",
        left=card_left + 0.2, top=cur_y + 0.08, width=label_w, height=0.35,
        font_size=13, bold=True,
        color="secondary", alignment="left",
        palette=palette,
    )
    add_text(
        slide, text=problem,
        left=content_left, top=cur_y + 0.08, width=content_w, height=card_h - 0.15,
        font_size=13, bold=True,
        color="text", alignment="left",
        palette=palette,
    )
    cur_y += card_h + 0.12

    # Card 3: Solution
    card_h = 1.1
    add_round_rect(
        slide,
        left=card_left, top=cur_y, width=card_w, height=card_h,
        fill_color="light_bg", palette=palette,
    )
    add_text(
        slide, text="Solution · 方案",
        left=card_left + 0.2, top=cur_y + 0.08, width=label_w, height=0.35,
        font_size=13, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )
    add_text(
        slide, text=solution,
        left=content_left, top=cur_y + 0.08, width=content_w, height=card_h - 0.15,
        font_size=13, bold=False,
        color="text", alignment="left",
        palette=palette,
    )
    cur_y += card_h + 0.12

    # Card 4: Results — horizontal metric cards
    add_text(
        slide, text="Results · 成果",
        left=card_left + 0.2, top=cur_y + 0.05, width=label_w, height=0.35,
        font_size=13, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )

    n = min(len(results), 4)
    metric_w = (card_w - (n - 1) * 0.2) / n
    for i, r in enumerate(results[:4]):
        mx = card_left + i * (metric_w + 0.2)
        my = cur_y + 0.45
        mh = 1.2
        add_round_rect(
            slide,
            left=mx, top=my, width=metric_w, height=mh,
            fill_color="light_bg", palette=palette,
        )
        add_rect(
            slide,
            left=mx, top=my, width=metric_w, height=0.05,
            fill_color="primary", palette=palette,
        )
        add_text(
            slide, text=r.get("value", ""),
            left=mx + 0.12, top=my + 0.15, width=metric_w - 0.24, height=0.45,
            font_size=36, bold=True,
            color="primary", alignment="left",
            palette=palette,
        )
        add_text(
            slide, text=r.get("metric", ""),
            left=mx + 0.12, top=my + 0.6, width=metric_w - 0.24, height=0.25,
            font_size=11, bold=True,
            color="text", alignment="left",
            palette=palette,
        )
        add_text(
            slide, text=r.get("comparison", ""),
            left=mx + 0.12, top=my + 0.85, width=metric_w - 0.24, height=0.25,
            font_size=10, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
        )

    add_source_line(slide, source, palette)
    return prs
