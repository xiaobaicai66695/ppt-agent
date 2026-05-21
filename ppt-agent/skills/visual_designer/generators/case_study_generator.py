"""Generator for case_study - 案例解析 (纵向流：背景→痛点→方案→成果)."""
from typing import List, Optional
from pptx import Presentation
from .base import (
    add_source_line, new_presentation, PALETTES, add_text, add_round_rect,
    set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "",
    kicker: str = "案例 · {领域}", title: str = "{案例名称}",
    context: str = "{背景}", problem: str = "{痛点}", solution: str = "{解决方案}",
    results: Optional[List[dict]] = None,
) -> Presentation:
    if results is None:
        results = [{"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"}] * 3
    if prs is None: prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]; slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    # ── Header ──
    y = 0.5  # extra top margin so content doesn't crowd the top
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.3, width=12.0, height=0.16,
            font_size=9, bold=False, color="text_muted", alignment="left", palette=palette)
        y = 0.42
    add_text(slide, text=title, left=0.5, top=y, width=12.0, height=0.35,
        font_size=22, bold=True, color="text", alignment="left", palette=palette)

    # ── Sequential vertical flow ──
    y += 0.55; left_x = 0.5; right_x = 7.0; full_w = 12.333
    line_h = 0.85  # row height per section

    # Row 1: 背景 (left) → 痛点 (right)
    add_round_rect(slide, left=left_x, top=y, width=6.0, height=line_h,
        fill_color="light_bg", palette=palette)
    add_text(slide, text="背景", left=left_x + 0.15, top=y + 0.04, width=1.5, height=0.22,
        font_size=11, bold=True, color="primary", alignment="left", palette=palette)
    add_text(slide, text=context, left=left_x + 0.15, top=y + 0.3, width=5.7, height=line_h - 0.35,
        font_size=9, bold=False, color="text", alignment="left", palette=palette)

    add_round_rect(slide, left=right_x, top=y, width=6.0, height=line_h,
        fill_color="light_bg", palette=palette)
    add_text(slide, text="痛点", left=right_x + 0.15, top=y + 0.04, width=1.5, height=0.22,
        font_size=11, bold=True, color="accent", alignment="left", palette=palette)
    add_text(slide, text=problem, left=right_x + 0.15, top=y + 0.3, width=5.7, height=line_h - 0.35,
        font_size=9, bold=False, color="text", alignment="left", palette=palette)

    # Row 2: 解决方案 (full width)
    y += line_h + 0.15; sol_h = 1.05
    add_round_rect(slide, left=left_x, top=y, width=full_w, height=sol_h,
        fill_color="light_bg", palette=palette)
    add_text(slide, text="解决方案", left=left_x + 0.15, top=y + 0.04, width=3.0, height=0.22,
        font_size=11, bold=True, color="secondary", alignment="left", palette=palette)
    add_text(slide, text=solution, left=left_x + 0.15, top=y + 0.3, width=full_w - 0.3, height=sol_h - 0.35,
        font_size=9, bold=False, color="text", alignment="left", palette=palette)

    # Row 3: 成果 (compact cards)
    y += sol_h + 0.15; n = min(len(results), 5)
    rcard_w = (full_w - (n - 1) * 0.15) / n; rcard_h = 0.78
    for j, r in enumerate(results[:n]):
        rx = left_x + j * (rcard_w + 0.15)
        is_first = (j == 0)
        add_round_rect(slide, left=rx, top=y, width=rcard_w, height=rcard_h,
            fill_color="primary" if is_first else "light_bg", palette=palette)
        add_text(slide, text=r.get("metric", ""),
            left=rx + 0.08, top=y + 0.06, width=rcard_w - 0.16, height=0.28,
            font_size=12, bold=True,
            color="background" if is_first else "primary", alignment="left", palette=palette)
        desc = r.get("value", r.get("desc", ""))
        add_text(slide, text=desc,
            left=rx + 0.08, top=y + 0.36, width=rcard_w - 0.16, height=rcard_h - 0.42,
            font_size=7, bold=False, color="background" if is_first else "text_muted",
            alignment="left", palette=palette)

    add_source_line(slide, source, palette)
    return prs
