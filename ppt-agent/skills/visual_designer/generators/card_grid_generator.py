"""Generator for card_grid - 卡片阵列 (4~8 cards)."""
from typing import Optional, List, Dict

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_round_rect, add_glass_panel,
    set_slide_background, add_text_in_shape,
    resolve_background, set_image_background,
)
from .asset_manager import add_local_icon, icon_id_from_text
from .layout_intelligence import balanced_band_top, body_font_size, card_layout_for_count, density_level, title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{能力标题}",
    cards: List[Dict[str, str]] = None,
    layout: str = "2x2",
    layout_variant: str = "",
    kicker: str = "",
    subtitle: str = "",
    background: str = None,
    glass_colors: dict = None,
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
        layout_variant: "equal_grid", "featured_card_plus_grid", or "masonry_cards".
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
    bg_path = resolve_background(background) if background else None
    if bg_path:
        colors = set_image_background(slide, bg_path, brightness=0.95, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    layout = card_layout_for_count(len(cards), layout)
    all_body = " ".join([c.get("body", "") for c in cards])
    level = density_level([c.get("header", "") for c in cards], title=title, body=all_body)

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
        font_size=title_font_size(title, base=36, sparse_boost=6, max_size=46), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
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
            colors=colors,
        )
        y_cards += 0.45

    variant = _resolve_layout_variant(layout_variant)
    if variant == "featured_card_plus_grid" and len(cards) >= 3:
        return _render_featured_cards(prs, slide, cards, palette, colors, source, y_cards, level)
    if variant == "masonry_cards":
        return _render_masonry_cards(prs, slide, cards, palette, colors, source, y_cards, level)

    # Parse layout
    rows, cols = map(int, layout.split("x"))
    total_cards = len(cards)
    card_rows = min(rows, (total_cards + cols - 1) // cols)

    # Grid dimensions
    margin_x = 0.6
    margin_top = y_cards
    gap = 0.25
    card_area_w = 12.133 - margin_x * 2
    card_area_h = 5.15
    card_w = (card_area_w - gap * (cols - 1)) / cols
    card_h = (card_area_h - gap * (card_rows - 1)) / card_rows
    used_grid_h = card_rows * card_h + max(card_rows - 1, 0) * gap
    margin_top = balanced_band_top(y_cards, card_area_h, used_grid_h, min_top=y_cards)

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
        add_glass_panel(
            slide,
            left=x, top=y, width=card_w, height=card_h,
            fill_color="light_bg", alpha=204, palette=palette,
        )

        # Card left accent bar
        add_rect(
            slide,
            left=x, top=y, width=0.06, height=card_h,
            fill_color="primary", palette=palette,
        )

        # Semantic icon badge
        badge_size = 0.5 if level != "dense" else 0.42
        badge_x = x + 0.2
        badge_y = y + 0.15
        icon_id = icon if icon and not icon.isdigit() else icon_id_from_text(header + " " + body, fallback="card")
        add_local_icon(slide, icon_id, left=badge_x, top=badge_y, size=badge_size, palette=palette, with_badge=True)

        # Card header (next to badge)
        add_text(
            slide,
            text=header,
            left=badge_x + badge_size + 0.1, top=badge_y + 0.05,
            width=card_w - badge_size - 0.5, height=0.45,
            font_size=17 if level == "sparse" else 15, bold=True,
            color="primary", alignment="left",
            palette=palette,
            colors=colors,
        )

        # Card body
        body_top = badge_y + badge_size + 0.15
        add_text(
            slide,
            text=body,
            left=x + 0.2, top=body_top,
            width=card_w - 0.4, height=card_h - (body_top - y) - 0.5,
            font_size=body_font_size([body], base=14), bold=False,
            color="text", alignment="left",
            vertical_alignment="middle" if level == "sparse" else "top",
            line_spacing=0.94,
            palette=palette,
            colors=colors,
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
                colors=colors,
            )

    add_source_line(slide, source, palette)
    return prs


def _resolve_layout_variant(layout_variant: str = "") -> str:
    key = (layout_variant or "").strip().lower().replace("-", "_")
    if key in {"featured_card_plus_grid", "featured", "featured_grid"}:
        return "featured_card_plus_grid"
    if key in {"masonry_cards", "masonry"}:
        return "masonry_cards"
    return "equal_grid"


def _draw_card(
    slide,
    card: Dict[str, str],
    idx: int,
    left: float,
    top: float,
    width: float,
    height: float,
    palette: str,
    colors: dict,
    level: str,
    featured: bool = False,
) -> None:
    header = card.get("header", "")
    body = card.get("body", "")
    icon = card.get("icon", f"{idx+1:02d}")
    footer = card.get("footer", "")
    fill = "background" if featured else "light_bg"
    add_glass_panel(
        slide,
        left=left, top=top, width=width, height=height,
        fill_color=fill,
        alpha=214 if featured else 204,
        palette=palette,
        line_color="divider",
        line_width=0.6,
    )
    add_rect(slide, left=left, top=top, width=0.06, height=height, fill_color="primary", palette=palette)
    badge_size = 0.62 if featured else (0.5 if level != "dense" else 0.42)
    badge_x = left + 0.22
    badge_y = top + 0.2
    icon_id = icon if icon and not icon.isdigit() else icon_id_from_text(header + " " + body, fallback="card")
    add_local_icon(slide, icon_id, left=badge_x, top=badge_y, size=badge_size, palette=palette, with_badge=True)
    add_text(
        slide,
        text=header,
        left=badge_x + badge_size + 0.13, top=badge_y + 0.04,
        width=width - badge_size - 0.62, height=0.48,
        font_size=19 if featured else (17 if level == "sparse" else 15),
        bold=True,
        color="primary",
        alignment="left",
        palette=palette,
        colors=colors,
    )
    body_top = badge_y + badge_size + (0.22 if featured else 0.15)
    add_text(
        slide,
        text=body,
        left=left + 0.22, top=body_top,
        width=width - 0.44, height=height - (body_top - top) - (0.58 if footer else 0.22),
        font_size=body_font_size([body], base=15 if featured else 14),
        bold=False,
        color="text",
        alignment="left",
        vertical_alignment="middle" if level == "sparse" else "top",
        line_spacing=0.94,
        palette=palette,
        colors=colors,
    )
    if footer:
        add_text(
            slide,
            text=footer,
            left=left + 0.22, top=top + height - 0.48,
            width=width - 0.44, height=0.34,
            font_size=12,
            bold=True,
            color="secondary",
            alignment="left",
            palette=palette,
            colors=colors,
        )


def _render_featured_cards(prs: Presentation, slide, cards, palette: str, colors: dict, source: str, y_cards: float, level: str) -> Presentation:
    visible = cards[:6]
    margin_x = 0.62
    area_h = 5.1
    top = balanced_band_top(y_cards, area_h, 4.78, min_top=y_cards)
    featured_w = 5.0
    gap = 0.25
    _draw_card(slide, visible[0], 0, margin_x, top, featured_w, 4.78, palette, colors, level, featured=True)
    side_x = margin_x + featured_w + gap
    side_w = 12.133 - margin_x * 2 - featured_w - gap
    side_cards = visible[1:]
    cols = 2
    rows = max(1, (len(side_cards) + cols - 1) // cols)
    card_h = (4.78 - gap * max(rows - 1, 0)) / rows
    card_w = (side_w - gap) / cols
    for i, card in enumerate(side_cards):
        row = i // cols
        col = i % cols
        _draw_card(slide, card, i + 1, side_x + col * (card_w + gap), top + row * (card_h + gap), card_w, card_h, palette, colors, level)
    add_source_line(slide, source, palette)
    return prs


def _render_masonry_cards(prs: Presentation, slide, cards, palette: str, colors: dict, source: str, y_cards: float, level: str) -> Presentation:
    visible = cards[:6]
    margin_x = 0.62
    gap = 0.24
    area_w = 12.133 - margin_x * 2
    col_w = (area_w - gap * 2) / 3
    top = balanced_band_top(y_cards, 5.1, 4.72, min_top=y_cards)
    heights = [2.22, 1.8, 2.22, 1.8, 2.22, 1.8]
    for idx, card in enumerate(visible):
        col = idx % 3
        row = idx // 3
        x = margin_x + col * (col_w + gap)
        y_offset = 0.0 if col != 1 else 0.28
        y = top + row * 2.34 + y_offset
        h = heights[idx]
        _draw_card(slide, card, idx, x, y, col_w, h, palette, colors, level)
    add_source_line(slide, source, palette)
    return prs
