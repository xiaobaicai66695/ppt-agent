"""Generator for kanban - 看板进度页."""
from typing import Optional, List
from pptx import Presentation
from .base import (
    add_source_line, new_presentation, PALETTES, add_text, add_rect, add_round_rect,
    set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "",
    title: str = "{看板标题}",
    subtitle: str = "",
    columns: Optional[List[dict]] = None,
    progress: int = 65,
    stats: str = "",
) -> Presentation:
    """Generate a kanban board slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        source: Data source annotation.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Subtitle below title.
        columns: List of column dicts, each with keys:
            - title: column header text
            - color: color key (text_muted, secondary, primary)
            - cards: list of card dicts with keys:
                - text: card title
                - tag: tag label
                - priority: priority level (high, medium, low, done)
        progress: Overall completion percentage (0-100).
        stats: Custom stats text. If empty, auto-generates from columns.

    Returns:
        The Presentation object.
    """
    if columns is None:
        columns = [
            {
                "title": "待办事项",
                "color": "text_muted",
                "cards": [
                    {"text": "任务1", "tag": "需求", "priority": "high"},
                    {"text": "任务2", "tag": "分析", "priority": "medium"},
                ]
            },
            {
                "title": "进行中",
                "color": "secondary",
                "cards": [
                    {"text": "任务3", "tag": "开发", "priority": "medium"},
                ]
            },
            {
                "title": "已完成",
                "color": "primary",
                "cards": [
                    {"text": "任务4", "tag": "管理", "priority": "done"},
                    {"text": "任务5", "tag": "需求", "priority": "done"},
                ]
            },
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    y_t = 0.3
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.2, width=12.0, height=0.2,
            font_size=12, bold=False, color="text_muted", alignment="left", palette=palette)
        y_t = 0.35
    add_text(slide, text=title, left=0.5, top=y_t, width=12.0, height=0.5,
        font_size=32, bold=True, color="text", alignment="left", palette=palette)
    if subtitle:
        add_text(slide, text=subtitle, left=0.5, top=y_t + 0.5, width=12.0, height=0.3,
            font_size=14, bold=False, color="text_muted", alignment="left", palette=palette)

    # Calculate column dimensions
    col_width = 3.8
    col_gap = 0.35
    start_x = 0.5
    card_start_y = 2.0

    # Draw columns
    for col_idx, col in enumerate(columns):
        x = start_x + col_idx * (col_width + col_gap)
        col_color = col.get("color", "secondary")
        col_title = col.get("title", "")
        cards = col.get("cards", [])

        # Column header
        add_rect(slide, left=x, top=1.35, width=col_width, height=0.5,
                 fill_color=col_color, palette=palette)
        add_text(slide, text=col_title,
                 left=x + 0.15, top=1.4, width=col_width - 0.7, height=0.4,
                 font_size=14, bold=True, color="background", palette=palette)

        # Count badge
        add_round_rect(slide, left=x + col_width - 0.6, top=1.42,
                       width=0.45, height=0.35,
                       fill_color="background", palette=palette)
        add_text(slide, text=str(len(cards)),
                 left=x + col_width - 0.6, top=1.43, width=0.45, height=0.35,
                 font_size=12, bold=True, color=col_color, alignment="center", palette=palette)

        # Cards
        card_y = card_start_y
        for card in cards:
            # Card background
            add_rect(slide, left=x, top=card_y, width=col_width, height=1.0,
                     fill_color="background", palette=palette)

            # Priority indicator (left edge)
            priority = card.get("priority", "medium")
            priority_colors = {"high": "primary", "medium": "secondary", "low": "accent", "done": "divider"}
            pc = priority_colors.get(priority, "secondary")
            add_rect(slide, left=x, top=card_y, width=0.06, height=1.0,
                     fill_color=pc, palette=palette)

            # Tag
            tag = card.get("tag", "")
            if tag:
                add_round_rect(slide, left=x + 0.15, top=card_y + 0.1,
                               width=0.8, height=0.28,
                               fill_color="light_bg", palette=palette)
                add_text(slide, text=tag,
                         left=x + 0.15, top=card_y + 0.1, width=0.8, height=0.28,
                         font_size=9, bold=False, color="text_muted",
                         alignment="center", palette=palette)

            # Card title
            card_text = card.get("text", "")
            add_text(slide, text=card_text,
                     left=x + 0.15, top=card_y + 0.45, width=col_width - 0.3, height=0.45,
                     font_size=12, bold=True, color="text", palette=palette)

            card_y += 1.1

    # Footer - progress bar
    footer_y = 6.15
    add_rect(slide, left=0.5, top=footer_y, width=12.333, height=0.08,
             fill_color="light_bg", palette=palette)
    add_rect(slide, left=0.5, top=footer_y, width=12.333 * progress / 100, height=0.08,
             fill_color="primary", palette=palette)

    # Auto-generate stats if not provided
    if not stats:
        todo = len(columns[0].get("cards", [])) if len(columns) > 0 else 0
        doing = len(columns[1].get("cards", [])) if len(columns) > 1 else 0
        done = len(columns[2].get("cards", [])) if len(columns) > 2 else 0
        stats = f"整体进度：{progress}%  |  待办 {todo} 项  |  进行中 {doing} 项  |  已完成 {done} 项"

    add_text(slide, text=stats,
             left=0.5, top=footer_y + 0.15, width=12.333, height=0.35,
             font_size=12, color="text_muted", alignment="center", palette=palette)

    add_source_line(slide, source, palette)
    return prs
