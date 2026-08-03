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
    resolve_background, set_image_background,
)
from .layout_intelligence import balanced_band_top, title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "数据 · {效果}",
    title: str = "{指标标题}",
    kpis: Optional[List[dict]] = None,
    subtitle: str = "",
    show_progress: bool = True,
    progress_value: int = 80,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """Generate a kpi_dashboard slide with metric cards in 2x2 grid layout.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        kicker: Small label above title.
        title: Slide title.
        kpis: List of dicts with keys: value, label, delta, baseline.
        subtitle: Optional subtitle below title.
        show_progress: Whether to show progress bar at bottom.
        progress_value: Progress percentage (0-100).
    """
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
    bg_path = resolve_background(background) if background else None
    if bg_path:
        colors = set_image_background(slide, bg_path, brightness=0.95, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Kicker
    add_text(
        slide, text=kicker,
        left=0.5, top=0.25, width=12.0, height=0.3,
        font_size=12, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Title
    add_text(
        slide, text=title,
        left=0.5, top=0.55, width=12.0, height=0.55,
        font_size=title_font_size(title, base=30, sparse_boost=4, max_size=38), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Subtitle
    card_start_y = 1.55
    if subtitle:
        add_text(
            slide, text=subtitle,
            left=0.5, top=1.1, width=12.0, height=0.3,
            font_size=13, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        card_start_y = 1.8

    # 2x2 grid dimensions
    cols = 2
    rows = 2
    gap_x = 0.35
    gap_y = 0.3
    margin_x = 0.5
    card_area_w = 13.333 - margin_x * 2
    card_w = (card_area_w - gap_x) / cols
    card_h = 1.82 if show_progress else 2.0
    cards_total_h = rows * card_h + (rows - 1) * gap_y
    card_start_y = balanced_band_top(card_start_y, 4.2, cards_total_h, min_top=card_start_y)

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
            left=x + 0.2, top=y + 0.12, width=card_w - 0.4, height=0.62,
            font_size=40 if show_progress else 44, bold=True,
            color="primary", alignment="left",
            vertical_alignment="middle",
            palette=palette,
            colors=colors,
        )

        # Label
        add_text(
            slide, text=kpi.get("label", ""),
            left=x + 0.2, top=y + 0.78, width=card_w - 0.4, height=0.36,
            font_size=13, bold=True,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )

        # Delta — colored by direction
        delta = kpi.get("delta", "")
        delta_color = "secondary"
        if delta.startswith("\u2193"):  # down arrow
            delta_color = "accent"

        if delta:
            add_text(
                slide, text=delta,
                left=x + 0.2, top=y + 1.18, width=card_w - 0.4, height=0.28,
                font_size=14, bold=True,
                color=delta_color, alignment="left",
                palette=palette,
                colors=colors,
            )

        # Baseline
        baseline = kpi.get("baseline", "")
        if baseline:
            add_text(
                slide, text=baseline,
                left=x + 0.2, top=y + 1.48, width=card_w - 0.4, height=0.28,
                font_size=10, bold=False,
                color="text_muted", alignment="left",
                palette=palette,
                colors=colors,
            )

    # Bottom progress bar
    if show_progress:
        progress_y = 6.12
        add_rect(
            slide,
            left=0.5, top=progress_y, width=12.333, height=0.1,
            fill_color="background", palette=palette,
        )
        add_rect(
            slide,
            left=0.5, top=progress_y, width=12.333 * progress_value / 100, height=0.1,
            fill_color="primary", palette=palette,
        )
        add_text(
            slide, text=f"整体完成度：{progress_value}%",
            left=0.5, top=progress_y + 0.18, width=12.333, height=0.3,
            font_size=11, bold=False,
            color="text_muted", alignment="center",
            palette=palette,
            colors=colors,
        )

    add_source_line(slide, source, palette)
    return prs
