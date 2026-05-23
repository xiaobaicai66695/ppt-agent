"""Generator for card_grid - 卡片阵列 (4~8 cards)."""
from typing import Optional, List, Dict

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_round_rect,
    set_slide_background, add_text_in_shape,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{能力标题}",
    cards: List[Dict[str, str]] = None,
    layout: str = "2x2",
    kicker: str = "",
    subtitle: str = "",
) -> Presentation:
    """
    Generate a card grid slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        cards: List of dicts with keys: header, body, icon (optional), footer (optional).
               Example: [{"header": "智能问答", "body": "基于大模型的自然语言交互", "icon": "01", "footer": "↑ 3倍"}]
        layout: Grid layout, e.g. "2x2", "2x3", "3x2".
        kicker: Small label above title (e.g. "能力 · 核心模块").
        subtitle: Optional subtitle below title.

    Returns:
        The Presentation object.
    """
    if cards is None:
        cards = [
            {"header": "{能力1}", "body": "{能力描述}", "icon": "01"},
            {"header": "{能力2}", "body": "{能力描述}", "icon": "02"},
            {"header": "{能力3}", "body": "{能力描述}", "icon": "03"},
            {"header": "{能力4}", "body": "{能力描述}", "icon": "04"},
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

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
    )

    # Subtitle
    y_cards = 1.3 if not kicker else 1.15
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.5, top=y_title + 0.7, width=12.0, height=0.4,
            font_size=16, bold=False,
            color="secondary", alignment="left",
            palette=palette,
        )
        y_cards += 0.45

    # Parse layout
    rows, cols = map(int, layout.split("x"))
    total_cards = len(cards)
    card_rows = min(rows, (total_cards + cols - 1) // cols)

    # Grid dimensions
    margin_x = 0.6
    margin_top = y_cards
    gap = 0.25
    card_area_w = 12.133 - margin_x * 2
    card_area_h = 5.2
    card_w = (card_area_w - gap * (cols - 1)) / cols
    card_h = (card_area_h - gap * (card_rows - 1)) / card_rows

    # Clamp card count based on grid
    visible_cards = cards[:rows * cols]

    for idx, card in enumerate(visible_cards):
        row = idx // cols
        col = idx % cols

        x = margin_x + col * (card_w + gap)
        y = margin_top + row * (card_h + gap)

        header = card.get("header", "")
        body = card.get("body", "")
        icon = card.get("icon", f"{idx+1:02d}")
        footer = card.get("footer", "")

        # Card background
        add_rect(
            slide,
            left=x, top=y, width=card_w, height=card_h,
            fill_color="light_bg", palette=palette,
        )

        # Card left accent bar
        add_rect(
            slide,
            left=x, top=y, width=0.06, height=card_h,
            fill_color="primary", palette=palette,
        )

        # Circular number badge
        badge_size = 0.45
        badge_x = x + 0.2
        badge_y = y + 0.15
        add_round_rect(
            slide,
            left=badge_x, top=badge_y, width=badge_size, height=badge_size,
            fill_color="primary", palette=palette,
        )
        add_text(
            slide,
            text=icon,
            left=badge_x, top=badge_y + 0.05,
            width=badge_size, height=badge_size - 0.05,
            font_size=14, bold=True,
            color="background", alignment="center",
            palette=palette,
        )

        # Card header (next to badge)
        add_text(
            slide,
            text=header,
            left=badge_x + badge_size + 0.1, top=badge_y + 0.05,
            width=card_w - badge_size - 0.5, height=0.45,
            font_size=16, bold=True,
            color="primary", alignment="left",
            palette=palette,
        )

        # Card body
        body_top = badge_y + badge_size + 0.15
        add_text(
            slide,
            text=body,
            left=x + 0.2, top=body_top,
            width=card_w - 0.4, height=card_h - (body_top - y) - 0.5,
            font_size=14, bold=False,
            color="text", alignment="left",
            palette=palette,
        )

        # Card footer (trend label)
        if footer:
            add_text(
                slide,
                text=footer,
                left=x + 0.2, top=y + card_h - 0.45,
                width=card_w - 0.4, height=0.35,
                font_size=12, bold=True,
                color="secondary", alignment="left",
                palette=palette,
            )

    add_source_line(slide, source, palette)
    return prs
