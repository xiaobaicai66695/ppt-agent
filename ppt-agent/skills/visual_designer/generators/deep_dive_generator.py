"""Generator for deep_dive - 详解页 (上下布局：要点横排 → 案例数据全宽)."""
from typing import List, Optional
from pptx import Presentation
from .base import (
    add_source_line, PALETTES, add_round_rect, add_text,
    new_presentation, set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "",
    kicker: str = "{领域标签}", title: str = "{主题}", lede: str = "{一句话概括}",
    left_header: str = "核心要点", key_points: Optional[List[str]] = None,
    analysis: Optional[List[str]] = None,
    right_header: str = "案例", case_example: Optional[List[str]] = None,
    data_evidence: Optional[List[str]] = None,
    supplement: Optional[List[str]] = None,
) -> Presentation:
    if key_points is None: key_points = ["{要点1}", "{要点2}", "{要点3}"]
    if analysis is None: analysis = []
    if case_example is None: case_example = []
    if data_evidence is None: data_evidence = []
    if supplement is None: supplement = []

    if prs is None: prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]; slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    # ── Header ──
    y = 0.18
    add_text(slide, text=kicker, left=0.5, top=0.04, width=12.0, height=0.14,
        font_size=9, bold=False, color="text_muted", alignment="left", palette=palette)
    add_text(slide, text=title, left=0.5, top=0.18, width=12.0, height=0.35,
        font_size=24, bold=True, color="text", alignment="left", palette=palette)
    add_text(slide, text=lede, left=0.5, top=0.5, width=12.0, height=0.20,
        font_size=12, bold=False, color="text_muted", alignment="left", palette=palette, italic=True)

    body_y = 0.8

    # ── Section 1: Key Points (horizontal card strip) ──
    n_kp = min(len(key_points), 5)
    card_gap = 0.15
    card_w = (12.333 - (n_kp - 1) * card_gap) / n_kp
    card_h = 0.75
    add_text(slide, text=left_header, left=0.5, top=body_y - 0.18, width=3.0, height=0.18,
        font_size=10, bold=True, color="primary", alignment="left", palette=palette)

    accents = ["primary", "secondary", "accent", "primary", "secondary"]
    for i, pt in enumerate(key_points[:n_kp]):
        x = 0.5 + i * (card_w + card_gap)
        add_round_rect(slide, left=x, top=body_y, width=card_w, height=card_h,
            fill_color="light_bg", palette=palette)
        add_round_rect(slide, left=x, top=body_y, width=card_w, height=0.03,
            fill_color=accents[i % len(accents)], palette=palette)
        add_text(slide, text=pt, left=x + 0.1, top=body_y + 0.12, width=card_w - 0.2, height=card_h - 0.2,
            font_size=9, bold=False, color="text", alignment="left", palette=palette)

    # ── Section 2: Case + Data (full width, stacked vertically) ──
    sec2_y = body_y + card_h + 0.2
    case_w = 12.333; case_h = 1.55

    add_round_rect(slide, left=0.5, top=sec2_y, width=case_w, height=case_h,
        fill_color="light_bg", palette=palette)
    add_round_rect(slide, left=0.5, top=sec2_y, width=case_w, height=0.03,
        fill_color="accent", palette=palette)
    add_text(slide, text=right_header, left=0.65, top=sec2_y + 0.08, width=case_w - 0.3, height=0.2,
        font_size=11, bold=True, color="accent", alignment="left", palette=palette)

    # Case items in 2 columns within the case panel
    n_cases = min(len(case_example), 6)
    cols_c = 2
    case_item_w = (case_w - 0.6 - 0.3) / cols_c
    for ci, c_item in enumerate(case_example[:n_cases]):
        col = ci % cols_c; row_in = ci // cols_c
        cx = 0.65 + col * (case_item_w + 0.3)
        cy = sec2_y + 0.35 + row_in * 0.32
        add_text(slide, text=c_item, left=cx, top=cy, width=case_item_w, height=0.28,
            font_size=9, bold=False, color="text", alignment="left", palette=palette)

    # Data evidence below case panel
    if data_evidence:
        data_y = sec2_y + case_h + 0.12
        data_w = 12.333; data_h = 0.7
        add_round_rect(slide, left=0.5, top=data_y, width=data_w, height=data_h,
            fill_color="light_bg", palette=palette)
        add_text(slide, text="数据", left=0.65, top=data_y + 0.04, width=2.0, height=0.18,
            font_size=10, bold=True, color="text_muted", alignment="left", palette=palette)
        n_data = min(len(data_evidence), 4)
        d_item_w = (data_w - 0.6) / n_data
        for di, d in enumerate(data_evidence[:n_data]):
            dx = 0.65 + di * d_item_w
            add_text(slide, text=d, left=dx, top=data_y + 0.28, width=d_item_w, height=0.35,
                font_size=9, bold=False, color="text", alignment="left", palette=palette)

    # Soft dot decor
    add_round_rect(slide, left=12.15, top=0.06, width=0.06, height=0.06,
        fill_color="accent", palette=palette)
    add_round_rect(slide, left=12.3, top=0.14, width=0.04, height=0.04,
        fill_color="secondary", palette=palette)

    add_source_line(slide, source, palette)
    return prs
