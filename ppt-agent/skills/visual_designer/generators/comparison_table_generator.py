"""Generator for comparison_table - 结构对比表 (70%×60% 居中)."""
from typing import Optional, List
from pptx import Presentation
from .base import (
    add_source_line, new_presentation, PALETTES, add_text, add_rect, add_round_rect,
    set_slide_background,
)

def generate(
    prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "",
    title: str = "{对比标题}", headers: List[str] = None, rows: List[List[str]] = None,
    kicker: str = "", subtitle: str = "",
) -> Presentation:
    if headers is None: headers = ["维度", "A", "B"]
    if rows is None: rows = [["1","a1","b1"],["2","a2","b2"],["3","a3","b3"]]
    if prs is None: prs = new_presentation(palette=palette)
    blank_layout = prs.slide_layouts[6]; slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)
    n_cols = len(headers); n_rows = min(len(rows), 8)

    y_t = 0.2
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.04, width=12.0, height=0.16,
            font_size=9, bold=False, color="text_muted", alignment="left", palette=palette)
        y_t = 0.16
    add_text(slide, text=title, left=0.5, top=y_t, width=12.0, height=0.38,
        font_size=24, bold=True, color="text", alignment="left", palette=palette)

    # Table: 70% width, 60% height, centered
    tbl_w = 13.333 * 0.70
    tbl_h = 7.5 * 0.60
    tbl_x = (13.333 - tbl_w) / 2
    header_h = 0.38; row_h = (tbl_h - header_h) / n_rows
    tbl_y = y_t + 0.5 + max(0, (7.5 - (y_t + 0.5) - tbl_h) / 2)

    col_weights = [1.8] + [1.0] * (n_cols - 1); total_w = sum(col_weights)
    col_widths = [tbl_w * w / total_w for w in col_weights]

    add_round_rect(slide, left=tbl_x, top=tbl_y, width=tbl_w, height=header_h,
        fill_color="primary", palette=palette)
    cx = tbl_x
    for c, h in enumerate(headers):
        cw = col_widths[c]
        add_text(slide, text=h, left=cx + 0.1, top=tbl_y + 0.05, width=cw - 0.2, height=header_h - 0.1,
            font_size=10, bold=True, color="background", alignment="left" if c == 0 else "center", palette=palette)
        cx += cw
    for r, row in enumerate(rows[:n_rows]):
        ry = tbl_y + header_h + r * row_h
        bg = "light_bg" if r % 2 == 0 else "background"
        add_rect(slide, left=tbl_x, top=ry, width=tbl_w, height=row_h, fill_color=bg, palette=palette)
        cx = tbl_x
        for c, cell in enumerate(row[:n_cols]):
            cw = col_widths[c]
            add_text(slide, text=cell, left=cx + 0.1, top=ry + 0.04, width=cw - 0.2, height=row_h - 0.08,
                font_size=9, bold=c == 0, color="text", alignment="left" if c == 0 else "center", palette=palette)
            cx += cw
        if r == n_rows - 1:
            add_rect(slide, left=tbl_x, top=ry + row_h, width=tbl_w, height=0.02,
                fill_color="primary", palette=palette)
    add_source_line(slide, source, palette)
    return prs
