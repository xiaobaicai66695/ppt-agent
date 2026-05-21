"""Generator for kpi_dashboard - KPI 看板 (居中 | 丰富分析)."""
from typing import List, Optional
from pptx import Presentation
from .base import (
    add_source_line, PALETTES, add_rect, add_round_rect, add_text,
    new_presentation, set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "",
    kicker: str = "数据 · {效果}", title: str = "{指标标题}",
    kpis: Optional[List[dict]] = None, subtitle: str = "",
    analysis: str = "",
) -> Presentation:
    """KPI cards (compact) + rich analysis. Cards center when few, analysis fills space."""
    if kpis is None:
        kpis = [{"value":"{值}","label":"{说明}","delta":"{趋势}","baseline":"{基准}"}] * 4
    if prs is None: prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]; slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)
    n = max(2, min(6, len(kpis)))

    # ── Header ──
    y_h = 0.18
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.04, width=12.0, height=0.15,
            font_size=9, bold=False, color="text_muted", alignment="left", palette=palette)
        y_h = 0.16
    add_text(slide, text=title, left=0.5, top=y_h, width=12.0, height=0.35,
        font_size=22, bold=True, color="text", alignment="left", palette=palette)

    # Layout
    if n <= 3: cols = n; rows = 1
    elif n == 4: cols = 2; rows = 2
    elif n == 5: cols = 3; rows = 2
    else: cols = 3; rows = 2

    gap_x, gap_y = 0.16, 0.14; margin_x = 0.5
    avail_w = 13.333 - margin_x * 2
    card_w = (avail_w - gap_x * (cols - 1)) / cols
    card_h = 1.45

    # Total card block height
    total_card_h = rows * card_h + (rows - 1) * gap_y
    has_analysis = bool(analysis)

    if has_analysis:
        # When analysis is present, cards go top, analysis fills remaining
        card_y = y_h + 0.5
    else:
        # No analysis: vertically center the card grid
        avail_space = 7.5 - (y_h + 0.5)
        card_y = y_h + 0.5 + max(0, (avail_space - total_card_h) / 2)

    card_accents = ["primary", "secondary", "accent", "primary", "secondary", "primary"]

    for idx, kpi in enumerate(kpis[:n]):
        col = idx % cols; row = idx // cols
        if n == 5 and row == 1:
            x_off = (avail_w - (2 * card_w + gap_x)) / 2
            x = margin_x + x_off + col * (card_w + gap_x)
        else:
            x = margin_x + col * (card_w + gap_x)
        y = card_y + row * (card_h + gap_y)
        ac = card_accents[idx % len(card_accents)]

        add_round_rect(slide, left=x, top=y, width=card_w, height=card_h,
            fill_color="light_bg", palette=palette)
        add_round_rect(slide, left=x, top=y, width=card_w, height=0.04,
            fill_color=ac, palette=palette)
        add_text(slide, text=kpi.get("value",""),
            left=x + 0.15, top=y + 0.12, width=card_w - 0.3, height=0.42,
            font_size=28 if n <= 4 else 22, bold=True, color="primary",
            alignment="left", palette=palette)
        lbl = kpi.get("label","")
        if lbl:
            add_text(slide, text=lbl, left=x + 0.15, top=y + 0.58,
                width=card_w - 0.3, height=0.24, font_size=9, bold=True,
                color="text", alignment="left", palette=palette)
        delta = kpi.get("delta",""); baseline = kpi.get("baseline","")
        dy = y + 0.82
        if delta:
            add_text(slide, text=delta, left=x + 0.15, top=dy, width=card_w - 0.3, height=0.18,
                font_size=10, bold=True,
                color="secondary" if any(c in delta for c in ("↑","+")) else "accent",
                alignment="left", palette=palette)
        if baseline:
            add_text(slide, text=baseline, left=x + 0.15, top=dy + 0.2 if delta else dy,
                width=card_w - 0.3, height=0.16, font_size=7, bold=False,
                color="text_muted", alignment="left", palette=palette)

    # ── Analysis area: fills remaining height ──
    analysis_y = card_y + total_card_h + 0.15
    if analysis:
        rem_h = 7.35 - analysis_y
        if rem_h > 0.5:
            add_round_rect(slide, left=0.5, top=analysis_y, width=0.04, height=0.3,
                fill_color="primary", palette=palette)
            add_text(slide, text="分析解读", left=0.75, top=analysis_y + 0.02,
                width=11.5, height=0.2, font_size=10, bold=True, color="text_muted",
                alignment="left", palette=palette)
            add_text(slide, text=analysis, left=0.75, top=analysis_y + 0.26,
                width=11.5, height=rem_h - 0.3, font_size=10, bold=False, color="text",
                alignment="left", palette=palette)

    add_source_line(slide, source, palette)
    return prs
