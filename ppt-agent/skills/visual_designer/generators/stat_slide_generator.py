"""Generator for stat_slide - 关键数字页 (紧凑)."""
from typing import Optional, List, Dict
from pptx import Presentation
from .base import (
    add_source_line, new_presentation, PALETTES, add_text, add_round_rect,
    set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "",
    title: str = "{指标标题}", stats: List[Dict[str, str]] = None,
    kicker: str = "", subtitle: str = "", analysis: str = "",
) -> Presentation:
    if stats is None: stats = [{"number":"{数字}","unit":"{单位}","label":"{说明}"}] * 3
    n = max(2, min(5, len(stats))); stats = stats[:n]
    if prs is None: prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]; slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)
    has_analysis = bool(analysis)

    y_t = 0.22
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.04, width=12.0, height=0.16,
            font_size=9, bold=False, color="text_muted", alignment="left", palette=palette)
        y_t = 0.18
    add_text(slide, text=title, left=0.5, top=y_t, width=12.0, height=0.36,
        font_size=22, bold=True, color="text", alignment="left", palette=palette)

    # Compact panel
    panel_h = 2.6 if n <= 3 else 3.2
    stats_y = y_t + 0.52
    if subtitle:
        add_text(slide, text=subtitle, left=0.5, top=y_t + 0.34, width=12.0, height=0.18,
            font_size=10, bold=False, color="secondary", alignment="left", palette=palette)
        stats_y = y_t + 0.62

    add_round_rect(slide, left=0.5, top=stats_y, width=12.333, height=panel_h,
        fill_color="light_bg", palette=palette)

    stat_w = 12.333 / n
    number_sizes = {2: 40, 3: 34, 4: 30, 5: 26}

    for i, stat in enumerate(stats):
        x_c = 0.5 + stat_w * i + stat_w / 2
        num, unit, label = stat.get("number",""), stat.get("unit",""), stat.get("label","")

        add_text(slide, text=num,
            left=x_c - stat_w/2 + 0.1, top=stats_y + 0.3, width=stat_w - 0.2, height=0.7,
            font_size=number_sizes.get(n, 24), bold=True, color="primary",
            alignment="center", palette=palette)
        if unit:
            add_text(slide, text=unit,
                left=x_c - stat_w/2 + 0.1, top=stats_y + 1.0, width=stat_w - 0.2, height=0.2,
                font_size=14, bold=False, color="secondary", alignment="center", palette=palette)
            ly = stats_y + 1.22
        else:
            ly = stats_y + 1.05
        add_text(slide, text=label,
            left=x_c - stat_w/2 + 0.1, top=ly, width=stat_w - 0.2, height=0.35,
            font_size=11, bold=False, color="text", alignment="center", palette=palette)
        if i < n - 1:
            add_round_rect(slide, left=x_c + stat_w/2 + 0.01, top=stats_y + 0.5,
                width=0.015, height=1.4, fill_color="divider", palette=palette)

    if analysis:
        ay = stats_y + panel_h + 0.12
        add_round_rect(slide, left=0.5, top=ay, width=0.04, height=0.25,
            fill_color="primary", palette=palette)
        add_text(slide, text=analysis, left=0.7, top=ay - 0.02, width=11.5, height=0.45,
            font_size=10, bold=False, color="text", alignment="left", palette=palette)

    add_round_rect(slide, left=12.3, top=0.08, width=0.05, height=0.05,
        fill_color="accent", palette=palette)
    add_source_line(slide, source, palette)
    return prs
