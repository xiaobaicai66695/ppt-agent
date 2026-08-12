"""Generator for two_column - 双栏对比."""
from typing import Optional, List, Dict, Any

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_glass_panel,
    set_slide_background,
    resolve_background, set_image_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{对比标题}",
    left_header: str = "{左侧标题}",
    right_header: str = "{右侧标题}",
    left_bullets: List[str] = None,
    right_bullets: List[str] = None,
    kicker: str = "",
    # Rich content support
    left_intro: str = "",
    right_intro: str = "",
    left_sections: Dict[str, List[str]] = None,
    right_sections: Dict[str, List[str]] = None,
    left_items: List[Dict[str, Any]] = None,
    right_items: List[Dict[str, Any]] = None,
    layout_variant: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a two-column comparison slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        source: Data source citation.
        title: Slide title.
        left_header: Left column header.
        right_header: Right column header.
        left_bullets: Left column bullet items (simple text, fallback if no rich content).
        right_bullets: Right column bullet items (simple text, fallback if no rich content).
        kicker: Small label above title (e.g. "方案对比").
        left_intro: Opening paragraph for left column (richer context).
        right_intro: Opening paragraph for right column (richer context).
        left_sections: Dict of sections for left column. Keys:
            - "key_points": List of key point strings
            - "analysis": List of analysis strings
            - "data": List of data/evidence strings
            - "quote": List of quote strings (optional)
            Each section gets a bold label and the items below it.
        right_sections: Same structure as left_sections for right column.
        left_items: Left column rich items (per-item title+desc+metric). Each item:
            - title: Item title (bold)
            - desc: Item description (normal)
            - metric: Optional highlighted metric/value
        right_items: Right column rich items. Same structure as left_items.
        layout_variant: "balanced_cards", "split_table", or "mirror_emphasis".

    Returns:
        The Presentation object.
    """
    if left_sections is None and left_items is None and left_bullets is None:
        left_bullets = [
            "{左侧要点1}",
            "{左侧要点2}",
            "{左侧要点3}",
            "{左侧要点4}",
        ]
    if right_sections is None and right_items is None and right_bullets is None:
        right_bullets = [
            "{右侧要点1}",
            "{右侧要点2}",
            "{右侧要点3}",
            "{右侧要点4}",
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
    variant = _resolve_layout_variant(layout_variant)

    # Kicker (above title)
    y_title = 0.4
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
        y_title = 0.35

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

    col_w = 5.8
    col_h = 5.5
    gap = 0.7
    start_x = 0.5
    start_y = 1.25 if not kicker else 1.2

    if variant == "split_table":
        return _render_split_table(
            prs, slide, palette, colors, source, title, left_header, right_header,
            left_bullets or [], right_bullets or [], start_y,
        )
    if variant == "mirror_emphasis":
        return _render_mirror_emphasis(
            prs, slide, palette, colors, source, title, left_header, right_header,
            left_bullets or [], right_bullets or [], start_y,
        )

    # Left column background
    add_glass_panel(
        slide,
        left=start_x, top=start_y, width=col_w, height=col_h,
        fill_color="light_bg", alpha=84, palette=palette,
    )
    # Left column top accent bar
    add_rect(
        slide,
        left=start_x, top=start_y, width=col_w, height=0.08,
        fill_color="primary", palette=palette,
    )

    # Right column background
    add_glass_panel(
        slide,
        left=start_x + col_w + gap, top=start_y, width=col_w, height=col_h,
        fill_color="light_bg", alpha=84, palette=palette,
    )
    # Right column top accent bar
    add_rect(
        slide,
        left=start_x + col_w + gap, top=start_y, width=col_w, height=0.08,
        fill_color="accent", palette=palette,
    )

    # Left header
    add_text(
        slide,
        text=left_header,
        left=start_x + 0.2, top=start_y + 0.18, width=col_w - 0.4, height=0.5,
        font_size=18, bold=True,
        color="primary", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Left content
    if left_sections:
        _render_sections(
            slide, left_sections, start_x + 0.2, start_y, col_w, palette, colors, is_left=True
        )
    elif left_items:
        _render_rich_items(
            slide, left_items, start_x + 0.2, start_y, col_w, palette, colors, is_left=True
        )
    elif left_intro:
        add_text(
            slide,
            text=left_intro,
            left=start_x + 0.2, top=start_y + 0.75, width=col_w - 0.4, height=0.8,
            font_size=11, bold=False,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )
        for i, item in enumerate(left_bullets[:5]):
            add_text(
                slide,
                text="· " + item,
                left=start_x + 0.2, top=start_y + 1.55 + i * 0.75, width=col_w - 0.4, height=0.7,
                font_size=12, bold=False,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )
    elif left_bullets:
        for i, item in enumerate(left_bullets[:6]):
            add_text(
                slide,
                text="· " + item,
                left=start_x + 0.2, top=start_y + 0.85 + i * 0.85, width=col_w - 0.4, height=0.8,
                font_size=13, bold=False,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )

    # Right header
    add_text(
        slide,
        text=right_header,
        left=start_x + col_w + gap + 0.2, top=start_y + 0.18, width=col_w - 0.4, height=0.5,
        font_size=18, bold=True,
        color="accent", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Right content
    if right_sections:
        _render_sections(
            slide, right_sections, start_x + col_w + gap + 0.2, start_y, col_w, palette, colors, is_left=False
        )
    elif right_items:
        _render_rich_items(
            slide, right_items, start_x + col_w + gap + 0.2, start_y, col_w, palette, colors, is_left=False
        )
    elif right_intro:
        add_text(
            slide,
            text=right_intro,
            left=start_x + col_w + gap + 0.2, top=start_y + 0.75, width=col_w - 0.4, height=0.8,
            font_size=11, bold=False,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )
        for i, item in enumerate(right_bullets[:5]):
            add_text(
                slide,
                text="· " + item,
                left=start_x + col_w + gap + 0.2, top=start_y + 1.55 + i * 0.75, width=col_w - 0.4, height=0.7,
                font_size=12, bold=False,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )
    elif right_bullets:
        for i, item in enumerate(right_bullets[:6]):
            add_text(
                slide,
                text="· " + item,
                left=start_x + col_w + gap + 0.2, top=start_y + 0.85 + i * 0.85, width=col_w - 0.4, height=0.8,
                font_size=13, bold=False,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )

    add_source_line(slide, source, palette)
    return prs


def _resolve_layout_variant(layout_variant: str = "") -> str:
    key = (layout_variant or "").strip().lower().replace("-", "_")
    if key in {"split_table", "table"}:
        return "split_table"
    if key in {"mirror_emphasis", "mirror"}:
        return "mirror_emphasis"
    return "balanced_cards"


def _render_split_table(
    prs: Presentation,
    slide,
    palette: str,
    colors: dict,
    source: str,
    title: str,
    left_header: str,
    right_header: str,
    left_bullets: List[str],
    right_bullets: List[str],
    start_y: float,
) -> Presentation:
    table_x, table_y, table_w = 0.72, start_y + 0.15, 11.9
    row_h = 0.78
    add_rect(slide, left=table_x, top=table_y, width=table_w, height=0.58, fill_color="primary", palette=palette)
    add_text(slide, text=left_header, left=table_x + 0.24, top=table_y + 0.12, width=5.35, height=0.28, font_size=14, bold=True, color="background", alignment="left", palette=palette, colors=colors)
    add_text(slide, text=right_header, left=table_x + 6.12, top=table_y + 0.12, width=5.35, height=0.28, font_size=14, bold=True, color="background", alignment="left", palette=palette, colors=colors)
    rows = max(len(left_bullets), len(right_bullets), 3)
    for i in range(min(rows, 6)):
        y = table_y + 0.58 + i * row_h
        add_glass_panel(slide, left=table_x, top=y, width=table_w, height=row_h - 0.02, fill_color="light_bg", alpha=92, palette=palette)
        add_rect(slide, left=table_x + 5.95, top=y, width=0.02, height=row_h - 0.02, fill_color="divider", palette=palette)
        left = left_bullets[i] if i < len(left_bullets) else ""
        right = right_bullets[i] if i < len(right_bullets) else ""
        add_text(slide, text=left, left=table_x + 0.24, top=y + 0.12, width=5.35, height=0.46, font_size=12, color="text", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
        add_text(slide, text=right, left=table_x + 6.12, top=y + 0.12, width=5.35, height=0.46, font_size=12, color="text", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
    add_source_line(slide, source, palette)
    return prs


def _render_mirror_emphasis(
    prs: Presentation,
    slide,
    palette: str,
    colors: dict,
    source: str,
    title: str,
    left_header: str,
    right_header: str,
    left_bullets: List[str],
    right_bullets: List[str],
    start_y: float,
) -> Presentation:
    center_x = 6.666
    add_glass_panel(slide, left=0.66, top=start_y + 0.02, width=5.58, height=5.42, fill_color="light_bg", alpha=92, palette=palette)
    add_glass_panel(slide, left=7.08, top=start_y + 0.02, width=5.58, height=5.42, fill_color="light_bg", alpha=92, palette=palette)
    add_rect(slide, left=center_x - 0.03, top=start_y + 0.25, width=0.06, height=5.15, fill_color="accent", palette=palette)
    add_text(slide, text=left_header, left=0.7, top=start_y + 0.15, width=5.2, height=0.5, font_size=22, bold=True, color="primary", alignment="right", palette=palette, colors=colors)
    add_text(slide, text=right_header, left=7.42, top=start_y + 0.15, width=5.2, height=0.5, font_size=22, bold=True, color="accent", alignment="left", palette=palette, colors=colors)
    for i, item in enumerate(left_bullets[:5]):
        y = start_y + 0.9 + i * 0.86
        add_text(slide, text=item, left=0.95, top=y, width=4.85, height=0.58, font_size=12, color="text", alignment="right", vertical_alignment="middle", palette=palette, colors=colors)
        add_rect(slide, left=6.1, top=y + 0.18, width=0.16, height=0.16, fill_color="primary", palette=palette)
    for i, item in enumerate(right_bullets[:5]):
        y = start_y + 0.9 + i * 0.86
        add_rect(slide, left=7.08, top=y + 0.18, width=0.16, height=0.16, fill_color="accent", palette=palette)
        add_text(slide, text=item, left=7.42, top=y, width=4.85, height=0.58, font_size=12, color="text", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
    add_source_line(slide, source, palette)
    return prs


def _render_sections(
    slide,
    sections: Dict[str, List[str]],
    col_x: float,
    start_y: float,
    col_w: float,
    palette: str,
    colors: dict,
    is_left: bool,
) -> None:
    """
    Render a dict of labeled sections within a column.
    Each key is a section label (e.g. "核心要点"), each value is a list of items.
    """
    section_labels = {
        "key_points": "核心要点",
        "analysis": "深度分析",
        "data": "数据支撑",
        "quote": "观点引用",
        "points": "要点",
    }
    accent_color = "primary" if is_left else "accent"
    cur_y = start_y + 0.75

    for key, items in sections.items():
        if not items:
            continue
        label = section_labels.get(key, key)
        # Section label with a small left bar indicator
        add_text(
            slide,
            text=label,
            left=col_x, top=cur_y, width=col_w - 0.3, height=0.3,
            font_size=12, bold=True,
            color=accent_color, alignment="left",
            palette=palette,
            colors=colors,
        )
        cur_y += 0.32
        for item in items[:5]:
            add_text(
                slide,
                text=item,
                left=col_x, top=cur_y, width=col_w - 0.3, height=0.6,
                font_size=11, bold=False,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )
            cur_y += 0.55
        cur_y += 0.1


def _render_rich_items(
    slide,
    items: List[Dict[str, Any]],
    col_x: float,
    start_y: float,
    col_w: float,
    palette: str,
    colors: dict,
    is_left: bool,
) -> None:
    """
    Render rich items with title, description, and optional metric.
    Each item uses a card-like block:
      [metric] (optional, highlighted)
      title (bold)
      desc (normal)
    """
    item_start_y = start_y + 0.9
    item_height = 1.05

    for i, item in enumerate(items[:5]):
        y = item_start_y + i * item_height

        if item.get("metric"):
            metric_text = item.get("metric", "")
            add_text(
                slide,
                text=metric_text,
                left=col_x, top=y, width=col_w - 0.3, height=0.3,
                font_size=11, bold=True,
                color="accent" if not is_left else "primary",
                alignment="left",
                palette=palette,
                colors=colors,
            )
            title_y_offset = 0.28
        else:
            title_y_offset = 0

        title_text = item.get("title", "")
        if title_text:
            add_text(
                slide,
                text=title_text,
                left=col_x, top=y + title_y_offset, width=col_w - 0.3, height=0.35,
                font_size=13, bold=True,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )

        desc_text = item.get("desc", "")
        if desc_text:
            add_text(
                slide,
                text=desc_text,
                left=col_x, top=y + title_y_offset + 0.35, width=col_w - 0.3, height=0.65,
                font_size=11, bold=False,
                color="secondary", alignment="left",
                palette=palette,
                colors=colors,
            )
