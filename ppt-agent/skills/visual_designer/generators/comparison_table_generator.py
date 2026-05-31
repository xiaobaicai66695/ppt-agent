"""Comparison table generator for comparison table slide type."""
from typing import Optional, List, Dict
from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect,
    set_slide_background,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "",
    title: str = "{对比表格标题}",
    subtitle: str = "",
    headers: Optional[List[str]] = None,
    rows: Optional[List[List[str]]] = None,
    recommendation: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a comparison table slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Optional subtitle.
        headers: List of column headers.
        rows: List of data rows (each row is a list of cell values).
        recommendation: Optional recommendation text at the bottom.

    Returns:
        The Presentation object.
    """
    if headers is None:
        headers = ["对比维度", "方案A", "方案B", "方案C"]
    if rows is None:
        rows = [
            ["功能丰富度", "★★★☆☆", "★★★★☆", "★★★☆☆"],
            ["易用性", "★★★☆☆", "★★★★★", "★★★★☆"],
            ["性能表现", "★★★★☆", "★★★★★", "★★★☆☆"],
            ["价格", "较高", "中等", "较低"],
            ["支持服务", "企业级", "专业级", "基础级"],
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
    y_offset = 0.35
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
        y_offset = 0.3

    # Title with left accent bar
    add_rect(
        slide,
        left=0.4, top=y_offset + 0.05, width=0.08, height=0.5,
        fill_color="primary", palette=palette,
    )
    add_text(
        slide,
        text=title,
        left=0.55, top=y_offset, width=11.5, height=0.55,
        font_size=28, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )
    y_offset += 0.6

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.55, top=y_offset, width=11.5, height=0.3,
            font_size=14, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.35

    # Table layout
    table_top = y_offset + 0.3
    num_cols = len(headers)
    num_rows = len(rows)
    col_widths = [2.5] + [2.5] * (num_cols - 1)  # First col wider
    total_width = sum(col_widths)
    table_left = (13.33 - total_width) / 2

    row_height = 0.55
    header_height = 0.5

    # Draw header row
    x = table_left
    for col_idx, header in enumerate(headers):
        w = col_widths[col_idx]

        # Header background
        add_rect(
            slide,
            left=x, top=table_top,
            width=w, height=header_height,
            fill_color="primary", palette=palette,
        )

        # Header text
        add_text(
            slide,
            text=header,
            left=x, top=table_top,
            width=w, height=header_height,
            font_size=13, bold=True,
            color="background", alignment="center",
            palette=palette,
            colors=colors,
        )

        # Border
        add_rect(
            slide,
            left=x, top=table_top + header_height,
            width=w, height=0.01,
            fill_color="divider", palette=palette,
        )

        x += w

    # Draw data rows
    for row_idx, row in enumerate(rows):
        y = table_top + header_height + row_idx * row_height
        is_even = row_idx % 2 == 1
        bg_color = "light_bg" if is_even else "background"

        x = table_left
        for col_idx, cell in enumerate(row):
            w = col_widths[col_idx]

            # Cell background
            add_rect(
                slide,
                left=x, top=y,
                width=w, height=row_height,
                fill_color=bg_color, palette=palette,
            )

            # Cell text
            is_first_col = col_idx == 0
            add_text(
                slide,
                text=cell,
                left=x + 0.1, top=y,
                width=w - 0.2, height=row_height,
                font_size=12,
                bold=is_first_col,
                color="text" if is_first_col else "text",
                alignment="center" if not is_first_col else "left",
                palette=palette,
                colors=colors,
            )

            # Border
            add_rect(
                slide,
                left=x, top=y + row_height,
                width=w, height=0.01,
                fill_color="divider", palette=palette,
            )
            add_rect(
                slide,
                left=x, top=y,
                width=0.01, height=row_height,
                fill_color="divider", palette=palette,
            )

            x += w

    # Right border for last column
    add_rect(
        slide,
        left=table_left + total_width - 0.01,
        top=table_top,
        width=0.01,
        height=header_height + num_rows * row_height,
        fill_color="divider", palette=palette,
    )

    # Recommendation
    if recommendation:
        rec_y = table_top + header_height + num_rows * row_height + 0.3
        add_rect(
            slide,
            left=table_left,
            top=rec_y,
            width=total_width,
            height=0.5,
            fill_color="accent", palette=palette,
        )
        add_text(
            slide,
            text=f"建议: {recommendation}",
            left=table_left,
            top=rec_y,
            width=total_width,
            height=0.5,
            font_size=13, bold=True,
            color="text", alignment="center",
            palette=palette,
            colors=colors,
        )

    add_source_line(slide, source, palette)
    return prs
