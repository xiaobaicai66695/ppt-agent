"""Generator for example_detail - 实例详解页.

Displays a real-world case study with kicker, title, lede, context,
solution, metrics grid, and takeaway.
"""
from typing import List, Optional

from pptx import Presentation
from pptx.util import Pt

from .base import (
    add_source_line,
    PALETTES,
    add_rect,
    add_round_rect,
    add_text,
    new_presentation,
    set_slide_background,
)

# card_spec: metric card dimensions
CARD_W = 3.8
CARD_H = 1.4
CARD_Y = 3.6
CARD_GAP = 0.3


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "实例 · {领域}",
    title: str = "{案例名称}: {一句话总结}",
    lede: str = "{核心数据或价值}",
    context_block: str = "{背景描述}",
    solution_block: str = "{解决方案}",
    metrics: Optional[List[dict]] = None,
    takeaway: str = "{启示}",
) -> Presentation:
    """Generate an example_detail slide with a named, quantified case study."""
    if metrics is None:
        metrics = [
            {"value": "{数值}", "label": "{指标}", "trend": "{趋势}"},
            {"value": "{数值}", "label": "{指标}", "trend": "{趋势}"},
            {"value": "{数值}", "label": "{指标}", "trend": "{趋势}"},
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Decorative left accent line through title area
    add_rect(
        slide,
        left=0.3, top=0.4, width=0.06, height=2.4,
        fill_color="primary", palette=palette,
    )

    # Kicker
    add_text(
        slide, text=kicker,
        left=0.5, top=0.4, width=12.0, height=0.35,
        font_size=12, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
    )

    # Title
    add_text(
        slide, text=title,
        left=0.5, top=0.7, width=12.0, height=0.65,
        font_size=30, bold=True,
        color="text", alignment="left",
        palette=palette,
    )

    # Lede
    add_text(
        slide, text=lede,
        left=0.5, top=1.25, width=12.0, height=0.45,
        font_size=16, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
        italic=True,
    )

    # Context block
    add_text(
        slide, text="背景",
        left=0.5, top=1.75, width=2.0, height=0.35,
        font_size=13, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )
    add_text(
        slide, text=context_block,
        left=2.3, top=1.75, width=10.2, height=0.55,
        font_size=13, bold=False,
        color="text", alignment="left",
        palette=palette,
    )

    # Solution block
    add_text(
        slide, text="方案",
        left=0.5, top=2.4, width=2.0, height=0.35,
        font_size=13, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )
    add_text(
        slide, text=solution_block,
        left=2.3, top=2.4, width=10.2, height=0.65,
        font_size=13, bold=False,
        color="text", alignment="left",
        palette=palette,
    )

    # Metrics grid — up to 4 cards in a row
    n = min(len(metrics), 4)
    total_w = n * CARD_W + (n - 1) * CARD_GAP
    start_x = (13.333 - total_w) / 2

    for i, m in enumerate(metrics[:4]):
        cx = start_x + i * (CARD_W + CARD_GAP)
        # Card background
        add_round_rect(
            slide,
            left=cx, top=CARD_Y, width=CARD_W, height=CARD_H,
            fill_color="light_bg", palette=palette,
        )
        # Top accent strip
        add_rect(
            slide,
            left=cx, top=CARD_Y, width=CARD_W, height=0.06,
            fill_color="primary", palette=palette,
        )
        # Value
        add_text(
            slide, text=m.get("value", ""),
            left=cx + 0.15, top=CARD_Y + 0.2, width=CARD_W - 0.3, height=0.55,
            font_size=36, bold=True,
            color="primary", alignment="left",
            palette=palette,
        )
        # Label
        add_text(
            slide, text=m.get("label", ""),
            left=cx + 0.15, top=CARD_Y + 0.75, width=CARD_W - 0.3, height=0.3,
            font_size=11, bold=False,
            color="text", alignment="left",
            palette=palette,
        )
        # Trend
        trend = m.get("trend", "")
        if trend:
            add_text(
                slide, text=trend,
                left=cx + 0.15, top=CARD_Y + 1.05, width=CARD_W - 0.3, height=0.25,
                font_size=11, bold=True,
                color="secondary", alignment="left",
                palette=palette,
            )

    # Takeaway — bottom strip
    add_rect(
        slide,
        left=0.5, top=5.5, width=12.333, height=0.65,
        fill_color="light_bg", palette=palette,
    )
    add_text(
        slide, text=takeaway,
        left=0.7, top=5.55, width=11.933, height=0.55,
        font_size=14, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs
