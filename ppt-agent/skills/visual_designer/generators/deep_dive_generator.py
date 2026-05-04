"""Generator for deep_dive - 双栏详解页.

Dual-column layout: left column explains key points,
right column shows case examples and data evidence.
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
    kicker: str = "{领域标签}",
    title: str = "{主题}",
    lede: str = "{一句话概括}",
    # Left column
    left_header: str = "核心要点",
    key_points: Optional[List[str]] = None,
    analysis: Optional[List[str]] = None,
    # Right column
    right_header: str = "案例/数据",
    case_example: Optional[List[str]] = None,
    data_evidence: Optional[List[str]] = None,
    supplement: Optional[List[str]] = None,
) -> Presentation:
    """Generate a deep_dive slide with dual-column layout."""
    if key_points is None:
        key_points = [
            "{要点1}",
            "{要点2}",
            "{要点3}",
            "{要点4}",
        ]
    if analysis is None:
        analysis = [
            "{分析维度1}：{结论}",
            "{分析维度2}：{结论}",
        ]
    if case_example is None:
        case_example = [
            "{案例要素1}",
            "{案例要素2}",
            "{案例要素3}",
        ]
    if data_evidence is None:
        data_evidence = [
            "{指标1}：{数值}",
            "{指标2}：{数值}",
            "{指标3}：{数值}",
        ]
    if supplement is None:
        supplement = []

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
        left=0.5, top=0.5, width=12.0, height=0.55,
        font_size=30, bold=True,
        color="text", alignment="left",
        palette=palette,
    )

    # Lede
    add_text(
        slide, text=lede,
        left=0.5, top=1.0, width=12.0, height=0.35,
        font_size=16, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
        italic=True,
    )

    # ── Two-column layout ──
    col_w = 5.8
    col_h = 5.6
    gap = 0.7
    left_x = 0.5
    right_x = left_x + col_w + gap
    col_y = 1.5

    # === Left column: Key Points ===
    add_rect(
        slide,
        left=left_x, top=col_y, width=col_w, height=col_h,
        fill_color="light_bg", palette=palette,
    )
    add_rect(
        slide,
        left=left_x, top=col_y, width=col_w, height=0.06,
        fill_color="primary", palette=palette,
    )

    # Left header
    add_text(
        slide, text=left_header,
        left=left_x + 0.2, top=col_y + 0.15, width=col_w - 0.4, height=0.4,
        font_size=18, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )

    cur_y = col_y + 0.65

    # Key points
    add_text(
        slide, text="核心要点",
        left=left_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.3,
        font_size=13, bold=True,
        color="text", alignment="left",
        palette=palette,
    )
    cur_y += 0.35
    for point in key_points[:5]:
        add_text(
            slide, text=point,
            left=left_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.4,
            font_size=11, bold=False,
            color="text", alignment="left",
            palette=palette,
        )
        cur_y += 0.4

    # Analysis
    if analysis:
        cur_y += 0.1
        add_text(
            slide, text="深度分析",
            left=left_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.3,
            font_size=13, bold=True,
            color="text", alignment="left",
            palette=palette,
        )
        cur_y += 0.35
        for item in analysis[:3]:
            add_text(
                slide, text=item,
                left=left_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.4,
                font_size=11, bold=False,
                color="text", alignment="left",
                palette=palette,
            )
            cur_y += 0.4

    # === Right column: Case & Data ===
    add_rect(
        slide,
        left=right_x, top=col_y, width=col_w, height=col_h,
        fill_color="light_bg", palette=palette,
    )
    add_rect(
        slide,
        left=right_x, top=col_y, width=col_w, height=0.06,
        fill_color="accent", palette=palette,
    )

    # Right header
    add_text(
        slide, text=right_header,
        left=right_x + 0.2, top=col_y + 0.15, width=col_w - 0.4, height=0.4,
        font_size=18, bold=True,
        color="accent", alignment="left",
        palette=palette,
    )

    cur_y = col_y + 0.65

    # Case example
    add_text(
        slide, text="案例说明",
        left=right_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.3,
        font_size=13, bold=True,
        color="text", alignment="left",
        palette=palette,
    )
    cur_y += 0.35
    for item in case_example[:4]:
        add_text(
            slide, text=item,
            left=right_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.35,
            font_size=11, bold=False,
            color="text", alignment="left",
            palette=palette,
        )
        cur_y += 0.35

    # Data evidence
    if data_evidence:
        cur_y += 0.1
        add_text(
            slide, text="数据支撑",
            left=right_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.3,
            font_size=13, bold=True,
            color="text", alignment="left",
            palette=palette,
        )
        cur_y += 0.35
        for item in data_evidence[:3]:
            add_text(
                slide, text=item,
                left=right_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.35,
                font_size=11, bold=False,
                color="text", alignment="left",
                palette=palette,
            )
            cur_y += 0.35

    # Supplement
    if supplement:
        cur_y += 0.1
        add_text(
            slide, text="补充信息",
            left=right_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.3,
            font_size=13, bold=True,
            color="text", alignment="left",
            palette=palette,
        )
        cur_y += 0.35
        for item in supplement[:2]:
            add_text(
                slide, text=item,
                left=right_x + 0.2, top=cur_y, width=col_w - 0.4, height=0.35,
                font_size=11, bold=False,
                color="text", alignment="left",
                palette=palette,
            )
            cur_y += 0.35

    add_source_line(slide, source, palette)
    return prs
