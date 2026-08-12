"""Generator for content_slide - 普通内容页（兜底类型）."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_text_in_shape,
    set_slide_background, add_paragraph,
    resolve_background, set_image_background,
)
from .asset_manager import add_local_icon, icon_id_from_text
from .layout_intelligence import (
    alignment_for_density,
    body_font_size,
    density_level,
    focal_font_size,
    balanced_band_top,
    short_items,
    title_font_size,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{页面标题}",
    section_header: str = "",
    bullets: List[str] = None,
    kicker: str = "",
    lede: str = "",
    highlight_stats: List[dict] = None,
    layout_variant: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a content slide with title, optional section header, and bullet list.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        section_header: Optional section header text.
        bullets: List of bullet items (max 5, each up to 20 Chinese chars).
        kicker: Small label above title (e.g. "要点 · 核心技术").
        lede: Optional lead-in paragraph below section header (1-2 sentences, ~50 chars).
        highlight_stats: Optional list of dicts with keys: value, label for right-side stat cards.
                        Example: [{"value": "98%", "label": "满意度"}, {"value": "3x", "label": "效率提升"}]

    Returns:
        The Presentation object.
    """
    title = _clean_placeholder(title)
    section_header = _clean_placeholder(section_header)
    kicker = _clean_placeholder(kicker)
    lede = _clean_placeholder(lede)
    if bullets is None:
        bullets = [
            "{要点1}",
            "{要点2}",
            "{要点3}",
        ]
    bullets = [_clean_placeholder(item) for item in bullets]
    bullets = [item for item in bullets if item]
    if not title:
        title = "Key Takeaways"
    if not section_header:
        section_header = _infer_section_header(title, bullets, lede)
    bullets = short_items(bullets, 6)
    level = density_level(bullets, title=title, body=lede)
    is_sparse = level == "sparse" and not highlight_stats
    variant = _resolve_layout_variant(layout_variant, level, bool(highlight_stats))

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    bg_path = resolve_background(background) if background else None
    if bg_path:
        colors = set_image_background(slide, bg_path, brightness=0.98, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    if is_sparse or variant == "icon_focus":
        icon_id = icon_id_from_text(" ".join([title, section_header, lede] + bullets), fallback="overview")
        add_local_icon(slide, icon_id, left=5.78, top=1.32, size=1.35, palette=palette, with_badge=True)
        if kicker:
            add_text(
                slide,
                text=kicker,
                left=1.3, top=0.48, width=10.733, height=0.3,
                font_size=12, bold=False,
                color="secondary", alignment="center",
                palette=palette,
                colors=colors,
            )
        add_text(
            slide,
            text=title,
            left=1.2, top=2.88, width=10.933, height=0.75,
            font_size=title_font_size(title, base=40, sparse_boost=8, max_size=50), bold=True,
            color="text", alignment="center",
            palette=palette,
            colors=colors,
        )
        lead_text = lede or section_header
        if lead_text and not lead_text.startswith("{"):
            add_text(
                slide,
                text=lead_text,
                left=2.1, top=3.72, width=9.133, height=0.55,
                font_size=18, bold=False,
                color="secondary", alignment="center",
                palette=palette,
                colors=colors,
            )
        if bullets:
            box_w = min(3.4, 10.0 / max(len(bullets), 1))
            total_w = len(bullets) * box_w + max(len(bullets) - 1, 0) * 0.25
            start_x = (13.333 - total_w) / 2
            for i, item in enumerate(bullets[:3]):
                x = start_x + i * (box_w + 0.25)
                add_rect(slide, left=x, top=4.75, width=box_w, height=0.95, fill_color="light_bg", palette=palette)
                add_text(
                    slide,
                    text=item,
                    left=x + 0.18, top=4.98, width=box_w - 0.36, height=0.45,
                    font_size=15, bold=True,
                    color="text", alignment="center",
                    palette=palette,
                    colors=colors,
                )
        add_source_line(slide, source, palette)
        return prs

    if variant == "numbered_cards" and bullets:
        if kicker:
            add_text(
                slide,
                text=kicker,
                left=0.72, top=0.35, width=11.8, height=0.3,
                font_size=12, color="secondary", alignment="left",
                palette=palette,
                colors=colors,
            )
        add_text(
            slide,
            text=title,
            left=0.7, top=0.78, width=11.8, height=0.65,
            font_size=title_font_size(title, base=34, sparse_boost=4, max_size=42), bold=True,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )
        if lede:
            add_text(
                slide,
                text=lede,
                left=0.72, top=1.42, width=10.8, height=0.45,
                font_size=13, color="text_muted", alignment="left",
                palette=palette,
                colors=colors,
            )
        cols = 2
        card_w = 5.82
        card_h = 1.06 if len(bullets) <= 4 else 0.88
        start_x = 0.74
        start_y = 2.08
        gap_x = 0.25
        gap_y = 0.22
        for i, item in enumerate(bullets[:6]):
            col = i % cols
            row = i // cols
            x = start_x + col * (card_w + gap_x)
            y = start_y + row * (card_h + gap_y)
            add_rect(slide, left=x, top=y, width=card_w, height=card_h, fill_color="light_bg", palette=palette)
            add_text(slide, text=f"{i + 1:02d}", left=x + 0.18, top=y + 0.18, width=0.58, height=0.28, font_size=13, bold=True, color="primary", alignment="left", palette=palette, colors=colors)
            add_text(slide, text=item, left=x + 0.86, top=y + 0.16, width=card_w - 1.1, height=card_h - 0.22, font_size=12 if len(item) > 42 else 13, color="text", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
        add_source_line(slide, source, palette)
        return prs

    if variant == "side_panel":
        add_rect(slide, left=0.0, top=0.0, width=3.55, height=7.5, fill_color="light_bg", palette=palette)
        add_rect(slide, left=3.55, top=0.0, width=0.08, height=7.5, fill_color="accent", palette=palette)
        if kicker:
            add_text(slide, text=kicker, left=0.52, top=0.58, width=2.45, height=0.32, font_size=12, color="secondary", alignment="left", palette=palette, colors=colors)
        add_text(slide, text=section_header or "核心要点", left=0.5, top=1.34, width=2.45, height=1.3, font_size=title_font_size(section_header or title, base=28, sparse_boost=4, max_size=34), bold=True, color="primary", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
        if lede:
            add_text(slide, text=lede, left=0.52, top=3.0, width=2.35, height=1.15, font_size=12, color="text_muted", alignment="left", palette=palette, colors=colors)
        add_text(slide, text=title, left=4.15, top=0.72, width=8.35, height=0.72, font_size=title_font_size(title, base=34, sparse_boost=4, max_size=42), bold=True, color="text", alignment="left", palette=palette, colors=colors)
        y_offset = balanced_band_top(1.78, 4.9, len(bullets[:6]) * 0.58, min_top=1.78)
        for i, item in enumerate(bullets[:6]):
            y = y_offset + i * 0.72
            add_rect(slide, left=4.18, top=y + 0.12, width=0.11, height=0.11, fill_color="secondary", palette=palette)
            add_text(slide, text=item, left=4.46, top=y, width=7.55, height=0.5, font_size=body_font_size(bullets, base=15), color="text", alignment="left", vertical_alignment="middle", palette=palette, colors=colors)
        add_source_line(slide, source, palette)
        return prs

    # Kicker (above title)
    y_title = 0.75
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.7, top=0.3, width=11.5, height=0.35,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_title = 0.7

    if bg_path:
        panel_width = 9.35 if not highlight_stats else 11.75
        add_rect(
            slide,
            left=0.55, top=1.38, width=panel_width, height=4.72,
            fill_color=(255, 255, 255, 176),
            line_color="divider",
            line_width=0.4,
            palette=palette,
        )

    # Left accent vertical bar
    add_rect(
        slide,
        left=0.5, top=y_title + 0.05, width=0.08, height=0.6,
        fill_color="primary", palette=palette,
    )

    # Title
    add_text(
        slide,
        text=title,
        left=0.7, top=y_title, width=11.5, height=0.7,
        font_size=title_font_size(title, base=36, sparse_boost=4, max_size=44), bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Section header
    y_offset = 1.7 if not kicker else 1.55
    text_width = 8.75 if bg_path and not highlight_stats else 11.5
    bullet_width = 8.45 if bg_path and not highlight_stats else 11.0
    if section_header:
        add_text(
            slide,
            text=section_header,
            left=0.7, top=y_offset, width=text_width, height=0.5,
            font_size=20, bold=True,
            color="primary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.6

    # Lede (lead-in paragraph between section_header and bullets)
    if lede:
        add_text(
            slide,
            text=lede,
            left=0.7, top=y_offset, width=text_width, height=0.5,
            font_size=14, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.6

    # Bullet list
    bullet_spacing = 0.68 if level != "dense" else 0.58
    item_font = body_font_size(bullets, base=16)
    if bullets and not highlight_stats:
        bullets_h = len(bullets[:6]) * 0.5 + max(len(bullets[:6]) - 1, 0) * (bullet_spacing - 0.5)
        y_offset = balanced_band_top(y_offset, 5.85 - y_offset, bullets_h, min_top=y_offset)
    for i, item in enumerate(bullets[:6]):
        y = y_offset + i * bullet_spacing

        # Bullet dot
        add_rect(
            slide,
            left=0.8, top=y + 0.12, width=0.12, height=0.12,
            fill_color="secondary", palette=palette,
        )

        # Bullet text
        add_text(
            slide,
            text=item,
            left=1.05, top=y, width=bullet_width, height=0.5,
            font_size=item_font, bold=False,
            color="text", alignment="left",
            vertical_alignment="middle",
            palette=palette,
            colors=colors,
        )

    # Right-side highlight stats (if provided)
    if highlight_stats:
        stats_x = 9.5
        stats_top = 2.5
        for i, stat in enumerate(highlight_stats[:2]):
            y = stats_top + i * 1.6
            value = stat.get("value", "")
            label = stat.get("label", "")

            # Stat card background
            add_rect(
                slide,
                left=stats_x, top=y, width=3.3, height=1.4,
                fill_color="light_bg", palette=palette,
            )
            # Left accent bar
            add_rect(
                slide,
                left=stats_x, top=y, width=0.06, height=1.4,
                fill_color="primary", palette=palette,
            )
            # Value
            add_text(
                slide,
                text=value,
                left=stats_x + 0.2, top=y + 0.2,
                width=2.9, height=0.7,
                font_size=28, bold=True,
                color="primary", alignment="left",
                palette=palette,
                colors=colors,
            )
            # Label
            add_text(
                slide,
                text=label,
                left=stats_x + 0.2, top=y + 0.9,
                width=2.9, height=0.4,
                font_size=12, bold=False,
                color="text_muted", alignment="left",
                palette=palette,
                colors=colors,
            )

    # Bottom-right decorative geometric shape
    add_rect(
        slide,
        left=11.8, top=6.5, width=1.2, height=0.8,
        fill_color="light_bg", palette=palette,
    )

    add_source_line(slide, source, palette)
    return prs


def _clean_placeholder(value: str) -> str:
    value = str(value or "").strip()
    if not value:
        return ""
    normalized = value.replace(" ", "")
    if normalized.startswith("{") and normalized.endswith("}"):
        return ""
    if "{" in normalized and "}" in normalized:
        return ""
    placeholder_values = {"页面标题", "小节标题", "副标题", "章节副标题"}
    if normalized in placeholder_values or normalized.startswith("要点") and len(normalized) <= 4:
        return ""
    return value


def _infer_section_header(title: str, bullets: List[str], lede: str) -> str:
    text = " ".join([title or "", lede or ""] + (bullets or []))
    lower = text.lower()
    keyword_map = [
        ("reliability", "Reliability Principles"),
        ("manifest", "Manifest Robustness"),
        ("runtime", "Runtime Guardrails"),
        ("tool", "Tool-Call Safety"),
        ("summary", "Executive Summary"),
        ("takeaway", "Actionable Takeaways"),
        ("可靠", "可靠性原则"),
        ("运行时", "运行时护栏"),
        ("工具", "工具调用安全"),
        ("清单", "清单稳健性"),
        ("总结", "核心回顾"),
    ]
    for keyword, header in keyword_map:
        if keyword in lower or keyword in text:
            return header
    return "核心观点"


def _resolve_layout_variant(layout_variant: str, level: str, has_highlight_stats: bool) -> str:
    key = (layout_variant or "").strip().lower().replace("-", "_")
    if key in {"numbered_cards", "cards", "card_list"}:
        return "numbered_cards"
    if key in {"side_panel", "sidebar", "left_panel"}:
        return "side_panel"
    if key in {"icon_focus", "focus", "center_focus"}:
        return "icon_focus"
    if level == "sparse" and not has_highlight_stats:
        return "icon_focus"
    return "classic_bullets"
