"""Generator for kpi_dashboard - KPI 仪表盘页.

Displays a set of KPI metric cards in a 2x2 grid layout.
Each card contains a large value, metric label, delta trend, and baseline.
"""
from typing import List, Optional

from pptx import Presentation

from .base import (
    add_source_line,
    PALETTES,
    add_rect,
    add_text,
    new_presentation,
    set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "数据 · {效果}",
    title: str = "{指标标题}",
    kpis: Optional[List[dict]] = None,
    subtitle: str = "",
) -> Presentation:
    """Generate a kpi_dashboard slide with metric cards in 2x2 grid layout."""
    if kpis is None:
        kpis = [
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
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
        left=0.5, top=0.25, width=12.0, height=0.3,
        font_size=12, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
    )

    # Title
    add_text(
        slide, text=title,
        left=0.5, top=0.55, width=12.0, height=0.55,
        font_size=28, bold=True,
        color="text", alignment="left",
        palette=palette,
    )

    # Subtitle
    card_start_y = 1.25
    if subtitle:
        add_text(
            slide, text=subtitle,
            left=0.5, top=1.1, width=12.0, height=0.3,
            font_size=13, bold=False,
            color="secondary", alignment="left",
            palette=palette,
        )
        card_start_y = 1.5

    # 2x2 grid dimensions
    cols = 2
    rows = 2
    gap_x = 0.3
    gap_y = 0.25
    margin_x = 0.5
    card_area_w = 13.333 - margin_x * 2
    card_w = (card_area_w - gap_x) / cols
    card_h = 2.35

    # Build the 2x2 position list (row-major: top-left, top-right, bottom-left, bottom-right)
    positions = []
    for r in range(rows):
        for c in range(cols):
            x = margin_x + c * (card_w + gap_x)
            y = card_start_y + r * (card_h + gap_y)
            positions.append((x, y))

    # Draw up to 4 KPI cards
    for idx, kpi in enumerate(kpis[:4]):
        x, y = positions[idx]

        # Card background (use light_bg)
        add_rect(
            slide,
            left=x, top=y, width=card_w, height=card_h,
            fill_color="light_bg", palette=palette,
        )

        # Left accent bar
        add_rect(
            slide,
            left=x, top=y, width=0.06, height=card_h,
            fill_color="primary", palette=palette,
        )

        # Value — large and prominent
        add_text(
            slide, text=kpi.get("value", ""),
            left=x + 0.2, top=y + 0.15, width=card_w - 0.4, height=0.7,
            font_size=44, bold=True,
            color="primary", alignment="left",
            palette=palette,
        )

        # Label
        add_text(
            slide, text=kpi.get("label", ""),
            left=x + 0.2, top=y + 0.88, width=card_w - 0.4, height=0.4,
            font_size=13, bold=True,
            color="text", alignment="left",
            palette=palette,
        )

        # Delta — colored by direction
        delta = kpi.get("delta", "")
        delta_color = "secondary"
        if delta.startswith("\u2193"):  # down arrow
            delta_color = "accent"

        if delta:
            add_text(
                slide, text=delta,
                left=x + 0.2, top=y + 1.32, width=card_w - 0.4, height=0.3,
                font_size=14, bold=True,
                color=delta_color, alignment="left",
                palette=palette,
            )

        # Baseline
        baseline = kpi.get("baseline", "")
        if baseline:
            add_text(
                slide, text=baseline,
                left=x + 0.2, top=y + 1.65, width=card_w - 0.4, height=0.35,
                font_size=10, bold=False,
                color="text_muted", alignment="left",
                palette=palette,
            )

    add_source_line(slide, source, palette)
    return prs
