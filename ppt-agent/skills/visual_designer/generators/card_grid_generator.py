"""Generator for card_grid - 卡片阵列 (2-8 cards, auto-layout)."""
from typing import Optional, List, Dict

from pptx import Presentation

from .base import (
    add_source_line, new_presentation,
    PALETTES, add_text, add_round_rect,
    set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{能力标题}",
    cards: List[Dict[str, str]] = None,
    layout: str = "auto",
    kicker: str = "",
    subtitle: str = "",
) -> Presentation:
    """Generate a card grid slide with rounded cards.

    layout="auto" picks the best grid for the card count:
      2 cards → 2x1  3 cards → 3x1  4 cards → 2x2
      5 cards → 3+2  6 cards → 3x2  7 cards → 4+3  8 cards → 4x2
    """
    if cards is None:
        cards = [
            {"header": "{能力1}", "body": "{能力描述}"},
            {"header": "{能力2}", "body": "{能力描述}"},
            {"header": "{能力3}", "body": "{能力描述}"},
            {"header": "{能力4}", "body": "{能力描述}"},
        ]

    n = max(2, min(8, len(cards)))
    cards = cards[:n]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    # ── Layout auto-selection ──
    if layout == "auto":
        if n <= 3:           cols = n;  rows = 1
        elif n == 4:         cols = 2;  rows = 2
        elif n == 5:         cols = 3;  rows = 2
        elif n == 6:         cols = 3;  rows = 2
        elif n == 7:         cols = 4;  rows = 2
        else:                cols = 4;  rows = 2
    else:
        rows, cols = map(int, layout.split("x"))

    # ── Header ──
    y_head = 0.28
    if kicker:
        add_text(slide, text=kicker,
            left=0.5, top=0.08, width=12.0, height=0.22,
            font_size=11, bold=False, color="text_muted", alignment="left",
            palette=palette)
        y_head = 0.22

    add_text(slide, text=title,
        left=0.5, top=y_head, width=12.0, height=0.55,
        font_size=30, bold=True, color="text", alignment="left",
        palette=palette)

    card_start_y = y_head + 0.7
    if subtitle:
        add_text(slide, text=subtitle,
            left=0.5, top=y_head + 0.5, width=12.0, height=0.25,
            font_size=13, bold=False, color="secondary", alignment="left",
            palette=palette)
        card_start_y += 0.3

    # ── Grid ──
    gap_x, gap_y = 0.2, 0.2
    margin_x = 0.5
    available_w = 12.4
    available_h = 6.2 - card_start_y
    card_w = (available_w - gap_x * (cols - 1)) / cols
    card_h = (available_h - gap_y * (rows - 1)) / rows

    # Accent colors for variety
    accent_colors = ["primary", "secondary", "accent", "primary", "secondary", "accent", "primary", "secondary"]

    for idx, card in enumerate(cards):
        if n == 5 and idx >= 3:  # bottom row of 3+2
            row = 1
            col = idx - 3
            bot_cols = 2
            offset_x = (available_w - (bot_cols * card_w + gap_x * (bot_cols - 1))) / 2
            x = margin_x + offset_x + col * (card_w + gap_x)
        elif n == 7 and idx >= 4:  # bottom row of 4+3
            row = 1
            col = idx - 4
            bot_cols = 3
            offset_x = (available_w - (bot_cols * card_w + gap_x * (bot_cols - 1))) / 2
            x = margin_x + offset_x + col * (card_w + gap_x)
        else:
            row = idx // cols
            col = idx % cols
            x = margin_x + col * (card_w + gap_x)

        y = card_start_y + row * (card_h + gap_y)
        accent = accent_colors[idx % len(accent_colors)]

        # Rounded card background
        add_round_rect(slide,
            left=x, top=y, width=card_w, height=card_h,
            fill_color="light_bg", palette=palette)

        # Top accent strip
        add_round_rect(slide,
            left=x, top=y, width=card_w, height=0.04,
            fill_color=accent, palette=palette)

        # Card header
        add_text(slide, text=card.get("header", ""),
            left=x + 0.18, top=y + 0.15, width=card_w - 0.36, height=0.4,
            font_size=15, bold=True, color="primary", alignment="left",
            palette=palette)

        # Card body
        add_text(slide, text=card.get("body", ""),
            left=x + 0.18, top=y + 0.6, width=card_w - 0.36, height=card_h - 0.8,
            font_size=12, bold=False, color="text", alignment="left",
            palette=palette)

    # Soft decoration
    add_round_rect(slide, left=12.2, top=0.12, width=0.06, height=0.06,
        fill_color="accent", palette=palette)

    add_source_line(slide, source, palette)
    return prs
