"""Generator for stat_slide - 关键数字页."""
import random
from typing import Optional, List, Dict

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_ellipse,
    set_slide_background,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{指标标题}",
    stats: List[Dict[str, str]] = None,
    kicker: str = "",
    subtitle: str = "",
    show_progress: bool = True,
    progress_value: int = 75,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a stat slide with large highlighted numbers.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        stats: List of dicts with keys: number, unit (optional), label, trend (optional).
               Example: [{"number": "99.99", "unit": "%", "label": "系统可用性", "trend": "↑ 0.3%"}]
        kicker: Small label above title (e.g. "年度成果").
        subtitle: Optional subtitle below title.
        show_progress: Whether to show a progress bar at the bottom.
        progress_value: Progress percentage (0-100) for the bottom progress bar.

    Returns:
        The Presentation object.
    """
    if stats is None:
        stats = [
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
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
    y_title = 0.5
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
        y_title = 0.45

    # Title
    add_text(
        slide,
        text=title,
        left=0.5, top=y_title, width=12.0, height=0.7,
        font_size=36, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Subtitle
    y_stats_bg = 1.85
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
        y_stats_bg = 2.15

    # Stats area background
    add_rect(
        slide,
        left=0.5, top=y_stats_bg, width=12.333, height=4.5,
        fill_color="light_bg", palette=palette,
    )

    # Distribute stats horizontally
    n = len(stats)
    available_width = 12.333
    stat_width = available_width / n

    # 根据 stat 数量动态调整字号
    if n <= 2:
        number_font_size = 72
        unit_font_size = 24
        number_height = 1.5
    elif n == 3:
        number_font_size = 56
        unit_font_size = 20
        number_height = 1.4
    else:  # 4+
        number_font_size = 44
        unit_font_size = 16
        number_height = 1.2

    for i, stat in enumerate(stats[:4]):
        x_center = 0.5 + stat_width * i + stat_width / 2

        number = stat.get("number", "")
        unit = stat.get("unit", "")
        label = stat.get("label", "")
        trend = stat.get("trend", "")

        # Decorative circle behind number
        add_ellipse(
            slide,
            left=x_center - stat_width * 0.35, top=y_stats_bg + 0.3,
            width=stat_width * 0.7, height=stat_width * 0.7,
            fill_color="background", palette=palette,
        )

        # Large number
        add_text(
            slide,
            text=number,
            left=x_center - stat_width * 0.45, top=y_stats_bg + 0.5,
            width=stat_width * 0.9, height=number_height,
            font_size=number_font_size, bold=True,
            color="primary", alignment="center",
            palette=palette,
            colors=colors,
        )

        # Unit (if any)
        if unit:
            add_text(
                slide,
                text=unit,
                left=x_center - stat_width * 0.45, top=y_stats_bg + 1.9,
                width=stat_width * 0.9, height=0.5,
                font_size=unit_font_size, bold=False,
                color="secondary", alignment="left",
                palette=palette,
                colors=colors,
            )

        # Trend label
        if trend:
            trend_color = "secondary" if trend.startswith("↑") else "accent"
            add_rect(
                slide,
                left=x_center - 0.6, top=y_stats_bg + 2.45,
                width=1.2, height=0.35,
                fill_color="background", palette=palette,
            )
            add_text(
                slide,
                text=trend,
                left=x_center - 0.6, top=y_stats_bg + 2.45,
                width=1.2, height=0.35,
                font_size=13, bold=True,
                color=trend_color, alignment="center",
                palette=palette,
                colors=colors,
            )

        # Label
        add_text(
            slide,
            text=label,
            left=x_center - stat_width * 0.45, top=y_stats_bg + 3.0,
            width=stat_width * 0.9, height=0.5,
            font_size=14, bold=False,
            color="text_muted", alignment="center",
            palette=palette,
            colors=colors,
        )

        # Divider line between stats (except last)
        if i < n - 1:
            add_rect(
                slide,
                left=x_center + stat_width / 2 - 0.01, top=y_stats_bg + 0.7, width=0.02, height=2.5,
                fill_color="divider", palette=palette,
            )

    # Bottom progress bar (hide when too many stats to avoid overlap)
    if show_progress and n <= 3:
        progress_y = 6.1
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
            slide,
            text=f"完成度：{progress_value}%",
            left=0.5, top=progress_y + 0.15, width=12.333, height=0.3,
            font_size=11, bold=False,
            color="text_muted", alignment="center",
            palette=palette,
            colors=colors,
        )

    add_source_line(slide, source, palette)
    return prs
