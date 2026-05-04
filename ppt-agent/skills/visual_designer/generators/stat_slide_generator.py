"""Generator for stat_slide - 关键数字页."""
from typing import Optional, List, Dict

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect,
    set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{指标标题}",
    stats: List[Dict[str, str]] = None,
    kicker: str = "",
    subtitle: str = "",
) -> Presentation:
    """
    Generate a stat slide with large highlighted numbers.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        stats: List of dicts with keys: number, unit (optional), label.
               Example: [{"number": "99.99", "unit": "%", "label": "系统可用性"}]
        kicker: Small label above title (e.g. "年度成果").
        subtitle: Optional subtitle below title.

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
    )

    # Subtitle
    y_stats_bg = 1.5
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.5, top=y_title + 0.65, width=12.0, height=0.4,
            font_size=16, bold=False,
            color="secondary", alignment="left",
            palette=palette,
        )
        y_stats_bg = 1.85

    # Stats area background
    add_rect(
        slide,
        left=0.5, top=y_stats_bg, width=12.333, height=5.2,
        fill_color="light_bg", palette=palette,
    )

    # Distribute stats horizontally
    n = len(stats)
    available_width = 12.333
    stat_width = available_width / n

    for i, stat in enumerate(stats[:3]):
        x_center = 0.5 + stat_width * i + stat_width / 2

        number = stat.get("number", "")
        unit = stat.get("unit", "")
        label = stat.get("label", "")

        # Large number
        add_text(
            slide,
            text=number,
            left=x_center - stat_width / 2 + 0.2, top=y_stats_bg + 0.5, width=stat_width - 0.4, height=1.5,
            font_size=72, bold=True,
            color="primary", alignment="center",
            palette=palette,
        )

        # Unit (if any)
        if unit:
            add_text(
                slide,
                text=unit,
                left=x_center - stat_width / 2 + 0.2, top=y_stats_bg + 1.9, width=stat_width - 0.4, height=0.5,
                font_size=24, bold=False,
                color="secondary", alignment="left",
                palette=palette,
            )

        # Label
        add_text(
            slide,
            text=label,
            left=x_center - stat_width / 2 + 0.2, top=y_stats_bg + 2.7, width=stat_width - 0.4, height=0.5,
            font_size=14, bold=False,
            color="text_muted", alignment="center",
            palette=palette,
        )

        # Divider line between stats (except last)
        if i < n - 1:
            add_rect(
                slide,
                left=x_center + stat_width / 2 - 0.01, top=y_stats_bg + 0.7, width=0.02, height=2.5,
                fill_color="divider", palette=palette,
            )

    add_source_line(slide, source, palette)
    return prs
