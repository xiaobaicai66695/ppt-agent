"""Component-first slide layout engine.

This module replaces rigid single-page drawing logic with semantic components.
Generator entry points keep their public function signatures, but internally
they convert keyword arguments into components and render them through this
shared engine.
"""
from __future__ import annotations

import math
import re
from typing import Any

from pptx import Presentation

from .base import (
    PALETTES,
    add_glass_panel,
    add_arrow,
    add_ellipse,
    add_line,
    add_rect,
    add_round_rect,
    add_source_line,
    add_text,
    new_presentation,
    resolve_background,
    set_image_background,
    set_slide_background,
)
from .asset_manager import add_cropped_photo, resolve_photo

SLIDE_W = 13.333
SLIDE_H = 7.5

TITLE_TYPES = {"headline", "subheadline", "section_marker", "eyebrow", "deck_title"}
CARD_TYPES = {
    "feature_card",
    "key_point",
    "callout",
    "fact_card",
    "insight",
    "recommendation",
    "risk_item",
    "opportunity_item",
    "toc_item",
    "case_snapshot",
    "decision_item",
}
LIST_TYPES = {"argument_block", "paragraph", "text_block", "bullet_list", "evidence_list", "list", "numbered_list"}
METRIC_TYPES = {"kpi_metric", "stat", "number_callout"}
FLOW_TYPES = {"timeline_node", "process_step", "milestone"}
MEDIA_TYPES = {"image", "map", "diagram"}
PRIMITIVE_TYPES = {"divider", "icon", "tag", "shape", "arrow", "architecture_box"}
TABLE_TYPES = {"table", "comparison_matrix"}
QUOTE_TYPES = {"quote_block"}
SOURCE_TYPES = {"source_note"}


def clean(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value.strip()
    if isinstance(value, list):
        return " / ".join(clean(item) for item in value if clean(item))
    if isinstance(value, dict):
        title = clean(value.get("title") or value.get("label") or value.get("name"))
        body = clean(value.get("body") or value.get("text") or value.get("description") or value.get("value"))
        if title and body and title != body:
            return f"{title}: {body}"
        return title or body
    return str(value).strip()


def component(component_type: str, **kwargs: Any) -> dict[str, Any]:
    data = {"type": component_type}
    for key, value in kwargs.items():
        if value not in (None, "", [], {}):
            data[key] = value
    return data


def prefer_explicit_components(kwargs: dict[str, Any], fallback: list[dict[str, Any]]) -> list[dict[str, Any]]:
    explicit = kwargs.get("components")
    if isinstance(explicit, list) and explicit:
        return list(explicit)
    return fallback


def display_component_type(component_type: str) -> str:
    labels = {
        "fact_card": "事实",
        "insight": "洞察",
        "recommendation": "建议",
        "risk_item": "风险",
        "opportunity_item": "机会",
        "case_snapshot": "案例",
        "decision_item": "决策",
        "toc_item": "目录",
        "stat": "指标",
        "number_callout": "数字",
        "milestone": "节点",
        "argument_block": "论述",
        "text_block": "说明",
        "list": "列表",
        "numbered_list": "步骤",
        "divider": "分隔",
        "icon": "图标",
        "tag": "标签",
        "shape": "强调",
        "arrow": "关系",
        "architecture_box": "模块",
    }
    return labels.get(component_type, "")


def clamp_text(text: str, max_chars: int) -> str:
    text = re.sub(r"\s+", " ", clean(text)).strip()
    if max_chars <= 0 or len(text) <= max_chars:
        return text
    return text[: max(1, max_chars - 1)].rstrip("，,。；;:： ") + "…"


def text_limit(width: float, height: float, font_size: float, ratio: float = 0.92) -> int:
    # Conservative CJK-friendly estimate. The renderer still auto-fits fonts;
    # this prevents pathological component text from escaping its card.
    chars_per_line = max(6, int(width * 72 / max(font_size, 1) * 1.65))
    lines = max(1, int(height * 72 / max(font_size * 1.18, 1)))
    return max(8, int(chars_per_line * lines * ratio))


def setup_slide(prs: Presentation | None, palette: str, background: str | None):
    if prs is None:
        prs = new_presentation(palette=palette)
    slide = prs.slides.add_slide(prs.slide_layouts[6])
    bg_path = resolve_background(background) if background else None
    if bg_path:
        colors = set_image_background(slide, bg_path, brightness=0.95, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    return prs, slide, colors


def render_component_slide(
    prs: Presentation | None = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "",
    subtitle: str = "",
    kicker: str = "",
    components: list[dict[str, Any]] | None = None,
    content_type: str = "content_slide",
    layout_variant: str = "",
    background: str | None = None,
) -> Presentation:
    prs, slide, colors = setup_slide(prs, palette, background)
    components = normalize_components(components or [], title=title, subtitle=subtitle, content_type=content_type)

    if content_type == "title_slide":
        _render_title(slide, colors, palette, title, subtitle, kicker, components)
    elif content_type == "section_divider":
        _render_section(slide, colors, palette, title, subtitle, kicker, components)
    elif content_type == "quote_slide":
        _render_quote(slide, colors, palette, title, subtitle, components)
    else:
        _render_workbench(slide, colors, palette, title, subtitle, kicker, components, content_type, layout_variant)

    add_source_line(slide, source, palette)
    return prs


def normalize_components(
    components: list[dict[str, Any]],
    title: str = "",
    subtitle: str = "",
    content_type: str = "content_slide",
) -> list[dict[str, Any]]:
    normalized: list[dict[str, Any]] = []
    if title:
        normalized.append(component("headline", text=title, role="main_point"))
    if subtitle:
        normalized.append(component("subheadline", text=subtitle))
    for index, item in enumerate(components):
        if not isinstance(item, dict):
            normalized.append(component("paragraph", text=clean(item), id=f"paragraph_{index + 1}"))
            continue
        item = dict(item)
        item["type"] = clean(item.get("type")) or infer_component_type(item, content_type)
        item.setdefault("id", f"{item['type']}_{index + 1}")
        normalized.append(item)
    return normalized


def infer_component_type(item: dict[str, Any], content_type: str) -> str:
    if content_type in {"card_grid", "icon_grid"}:
        return "feature_card"
    if content_type in {"kpi_dashboard", "stat_slide"}:
        return "kpi_metric"
    if content_type in {"timeline"}:
        return "timeline_node"
    if content_type in {"process_flow"}:
        return "process_step"
    if content_type in {"chart_slide"}:
        return "chart"
    if content_type in {"comparison_table"}:
        return "table"
    if item.get("items"):
        return "bullet_list"
    return "paragraph"


def _render_title(slide, colors: dict, palette: str, title: str, subtitle: str, kicker: str, components: list[dict[str, Any]]):
    add_rect(slide, 0, 0, SLIDE_W, SLIDE_H, "background", palette=palette)
    add_rect(slide, 0, 0, 0.22, SLIDE_H, "primary", palette=palette)
    if kicker:
        add_text(slide, kicker, 0.82, 1.05, 4.6, 0.36, 13, color="secondary", palette=palette, colors=colors)
    add_text(slide, title or component_text(first_component(components, "headline")), 0.8, 1.68, 10.9, 1.35, 46, True, "text", palette=palette, colors=colors)
    add_rect(slide, 0.84, 3.25, 1.35, 0.07, "accent", palette=palette)
    add_text(slide, subtitle or component_text(first_component(components, "subheadline")), 0.86, 3.55, 9.8, 0.85, 20, color="secondary", palette=palette, colors=colors)
    callouts = [c for c in components if c.get("type") in CARD_TYPES | METRIC_TYPES][:3]
    if callouts:
        _render_cards(slide, colors, palette, callouts, 0.86, 5.08, 11.5, 1.52, align_y="middle")


def _render_section(slide, colors: dict, palette: str, title: str, subtitle: str, kicker: str, components: list[dict[str, Any]]):
    number = clean(first_value(components, "section_marker")) or "01"
    add_rect(slide, 0, 0, 5.0, SLIDE_H, "primary", palette=palette)
    add_text(slide, kicker or "SECTION", 0.55, 0.82, 3.8, 0.35, 13, color="background", palette=palette, colors=colors)
    add_text(slide, number, 0.48, 2.45, 4.25, 1.45, 102, True, "background", palette=palette, colors=colors)
    add_text(slide, title, 5.75, 2.34, 6.7, 0.95, 42, True, "text", palette=palette, colors=colors)
    add_rect(slide, 5.78, 3.45, 1.18, 0.07, "accent", palette=palette)
    add_text(slide, subtitle, 5.78, 3.82, 6.55, 0.68, 17, color="secondary", palette=palette, colors=colors)


def _render_quote(slide, colors: dict, palette: str, title: str, subtitle: str, components: list[dict[str, Any]]):
    quote = component_text(first_component(components, "quote_block")) or subtitle or title
    attribution = clean(first_component(components, "source_note").get("text") if first_component(components, "source_note") else "")
    add_text(slide, "“", 0.86, 0.86, 1.1, 0.9, 60, True, "accent", palette=palette, colors=colors)
    add_text(slide, quote, 1.65, 1.68, 10.0, 2.2, 34, True, "text", "center", palette=palette, colors=colors)
    if attribution:
        add_text(slide, attribution, 2.2, 4.22, 8.8, 0.45, 15, color="secondary", alignment="center", palette=palette, colors=colors)
    add_rect(slide, 5.78, 5.05, 1.7, 0.06, "accent", palette=palette)


def _render_workbench(
    slide,
    colors: dict,
    palette: str,
    title: str,
    subtitle: str,
    kicker: str,
    components: list[dict[str, Any]],
    content_type: str,
    layout_variant: str,
):
    if kicker:
        add_text(slide, kicker, 0.55, 0.18, 11.8, 0.28, 12, color="secondary", palette=palette, colors=colors)
    add_text(slide, title or component_text(first_component(components, "headline")), 0.55, 0.52, 11.9, 0.62, 30, True, "text", palette=palette, colors=colors)
    if subtitle:
        add_text(slide, subtitle, 0.58, 1.13, 10.8, 0.36, 13, color="secondary", palette=palette, colors=colors)
    add_rect(slide, 0.57, 1.48, 1.1, 0.05, "accent", palette=palette)

    body_components = [c for c in components if c.get("type") not in TITLE_TYPES | SOURCE_TYPES]
    architecture_boxes = [c for c in body_components if c.get("type") == "architecture_box"]
    primitive_helpers = [c for c in body_components if c.get("type") in PRIMITIVE_TYPES and c.get("type") != "architecture_box"]
    charts = [c for c in body_components if c.get("type") == "chart"]
    tables = [c for c in body_components if c.get("type") in TABLE_TYPES]
    metrics = [c for c in body_components if c.get("type") in METRIC_TYPES]
    flows = [c for c in body_components if c.get("type") in FLOW_TYPES]
    cards = [c for c in body_components if c.get("type") in CARD_TYPES]
    lists = [c for c in body_components if c.get("type") in LIST_TYPES or c.get("items")]

    top = 1.72
    height = 5.25
    if primitive_helpers:
        _render_primitive_strip(slide, colors, palette, primitive_helpers, 8.05, 0.26, 4.45)
    if content_type in {"chart_slide"} and charts:
        _render_chart_placeholder(slide, colors, palette, charts[0], 0.65, top, 7.45, height)
        _render_cards(slide, colors, palette, metrics + cards + lists, 8.35, top, 4.25, height, compact=True, align_y="middle")
    elif content_type == "agenda":
        _render_agenda(slide, colors, palette, body_components, 0.72, top, 11.85, height)
    elif content_type in {"kpi_dashboard", "stat_slide"} and metrics:
        _render_metric_grid(slide, colors, palette, metrics, 0.65, top, 11.95, height)
    elif architecture_boxes:
        side_components = [c for c in body_components if c.get("type") not in PRIMITIVE_TYPES | SOURCE_TYPES]
        if side_components:
            _render_architecture_diagram(slide, colors, palette, architecture_boxes, primitive_helpers, 0.65, top, 7.2, height)
            _render_cards(slide, colors, palette, side_components, 8.15, top, 4.45, height, compact=True, align_y="middle")
        else:
            _render_architecture_diagram(slide, colors, palette, architecture_boxes, primitive_helpers, 0.65, top, 11.95, height)
    elif content_type in {"timeline", "process_flow"} and flows:
        _render_flow(slide, colors, palette, flows, 0.72, 3.0, 11.7, 2.1)
    elif content_type in {"swot_analysis"}:
        _render_quadrant(slide, colors, palette, body_components, 0.65, top, 11.95, height)
    elif content_type in {"kanban"}:
        _render_kanban(slide, colors, palette, body_components, 0.65, top, 11.95, height)
    elif content_type in {"brand_focus"}:
        _render_brand_focus(slide, colors, palette, body_components, 0.65, top, 11.95, height)
    elif content_type in {"region_map"}:
        _render_region_map(slide, colors, palette, body_components, 0.65, top, 11.95, height)
    elif content_type in {"image_hero"}:
        _render_image_hero(slide, colors, palette, body_components, 0.65, top, 11.95, height)
    elif content_type == "comparison_table" and tables:
        recommendations = [c for c in body_components if c.get("type") == "recommendation"]
        _render_table(slide, colors, palette, tables[0], recommendations, 0.65, top, 11.95, height)
    elif content_type in {"two_column"}:
        left = body_components[0::2]
        right = body_components[1::2]
        _render_cards(slide, colors, palette, left, 0.65, top, 5.75, height, compact=True, align_y="middle")
        _render_cards(slide, colors, palette, right, 6.9, top, 5.75, height, compact=True, align_y="middle")
    elif content_type in {"image_text", "case_study", "example_detail", "deep_dive"}:
        image_component = first_component(body_components, "image")
        image_path = component_image_path(image_component)
        if image_path:
            photo = resolve_photo(image_path=image_path, text=component_text(image_component))
        else:
            photo = None
        if photo and add_cropped_photo(slide, photo, 0.65, top, 4.0, height):
            caption = clean(image_component.get("caption") or image_component.get("attribution") or component_text(image_component))
            if caption:
                add_glass_panel(slide, 0.88, top + height - 0.82, 3.54, 0.52, palette=palette, fill_color="background", alpha=224)
                add_text(slide, clamp_text(caption, text_limit(3.25, 0.26, 9.0, 0.95)), 1.02, top + height - 0.68, 3.25, 0.24, 9.0, color="secondary", palette=palette, colors=colors, min_font_size=7, max_font_size=False)
        else:
            add_glass_panel(slide, 0.65, top, 4.0, height, palette=palette, fill_color="light_bg", alpha=210)
            image_text = component_text(image_component) or "主题视觉"
            add_text(slide, image_text, 0.95, top + 1.9, 3.4, 0.8, 22, True, "primary", "center", palette=palette, colors=colors)
        text_side = [c for c in body_components if c.get("type") not in MEDIA_TYPES]
        if has_narrative_components(text_side):
            _render_narrative_panel(slide, colors, palette, text_side, 5.0, top, 7.6, height)
        else:
            _render_cards(slide, colors, palette, text_side, 5.0, top, 7.6, height, compact=True, align_y="middle")
    elif content_type in {"content_slide", "summary_slide"} and has_narrative_components(body_components):
        _render_narrative_panel(slide, colors, palette, body_components, 0.65, top, 11.95, height)
    else:
        ordered = []
        seen = set()
        for item in cards + metrics + flows + lists + [c for c in body_components if c.get("type") not in {"divider", "shape", "arrow"}]:
            marker = id(item)
            if marker in seen:
                continue
            seen.add(marker)
            ordered.append(item)
        _render_cards(slide, colors, palette, ordered, 0.65, top, 11.95, height, align_y="middle")


def _render_cards(
    slide,
    colors: dict,
    palette: str,
    components: list[dict[str, Any]],
    left: float,
    top: float,
    width: float,
    height: float,
    compact: bool = False,
    align_y: str = "top",
):
    components = [c for c in components if component_text(c) or clean(c.get("title")) or clean(c.get("body"))][:8]
    if not components:
        components = [component("paragraph", text="围绕主题展开结构化说明")]
    count = len(components)
    cols = choose_card_columns(count, width, compact)
    rows = max(1, math.ceil(count / cols))
    gap = 0.2 if not compact else 0.18
    card_w = (width - gap * (cols - 1)) / cols
    max_card_h = 1.55 if compact else 1.72
    min_card_h = 0.82 if compact else 1.02
    raw_card_h = (height - gap * (rows - 1)) / rows
    card_h = min(max_card_h, max(min_card_h, raw_card_h))
    total_h = rows * card_h + (rows - 1) * gap
    if total_h > height:
        card_h = max(0.68 if compact else 0.86, raw_card_h)
        total_h = rows * card_h + (rows - 1) * gap
    y_offset = max(0, (height - total_h) / 2) if align_y in {"middle", "center"} else 0
    for i, item in enumerate(components):
        row = i // cols
        col = i % cols
        row_count = min(cols, count - row * cols)
        row_w = row_count * card_w + (row_count - 1) * gap
        row_left = left + max(0, (width - row_w) / 2)
        x = row_left + col * (card_w + gap)
        y = top + y_offset + row * (card_h + gap)
        add_glass_panel(slide, x, y, card_w, card_h, palette=palette, fill_color="light_bg", alpha=210)
        add_rect(slide, x, y, 0.06, card_h, component_accent(item), palette=palette)
        label_raw = component_label(item, i + 1)
        body_raw = component_body(item)
        label_font = 14 if compact else 16
        body_font = 10.5 if compact else 11.5
        label = clamp_text(label_raw, text_limit(card_w - 0.35, 0.34, label_font, 0.98))
        body = clamp_text(body_raw, text_limit(card_w - 0.35, card_h - 0.66, body_font, 0.88))
        add_text(slide, label, x + 0.18, y + 0.13, card_w - 0.35, 0.34, label_font, True, "text", palette=palette, colors=colors, min_font_size=10, max_font_size=False)
        add_text(slide, body, x + 0.18, y + 0.54, card_w - 0.35, card_h - 0.65, body_font, color="secondary", palette=palette, colors=colors, min_font_size=8.5, max_font_size=False, line_spacing=0.92)


def has_narrative_components(components: list[dict[str, Any]]) -> bool:
    return any(c.get("type") in {"argument_block", "paragraph", "list", "numbered_list", "bullet_list", "evidence_list"} for c in components)


def _render_agenda(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    items: list[tuple[str, str]] = []
    for component_item in components:
        if component_item.get("type") not in {"toc_item", "feature_card", "key_point", "paragraph", "text_block"}:
            continue
        for raw in split_agenda_text(component_body(component_item) or component_text(component_item)):
            title = clean(re.sub(r"^\d+\s*", "", raw)).strip()
            if not title:
                continue
            number_match = re.match(r"^(\d{1,2})", raw.strip())
            number = f"{int(number_match.group(1)):02d}" if number_match else f"{len(items) + 1:02d}"
            items.append((number, title))
    if not items:
        items = [("01", "核心内容"), ("02", "关键分析"), ("03", "结论建议")]

    items = items[:8]
    count = len(items)
    cols = 2 if count > 4 else 1
    rows = math.ceil(count / cols)
    gap_x = 0.38
    gap_y = 0.18 if rows >= 4 else 0.24
    panel_w = (width - gap_x * (cols - 1)) / cols
    row_h = min(0.92 if rows >= 4 else 1.05, (height - gap_y * (rows - 1)) / rows)
    total_h = rows * row_h + (rows - 1) * gap_y
    start_y = top + max(0.0, (height - total_h) / 2)

    for index, (number, title) in enumerate(items):
        col = index // rows
        row = index % rows
        x = left + col * (panel_w + gap_x)
        y = start_y + row * (row_h + gap_y)
        add_glass_panel(slide, x, y, panel_w, row_h, palette=palette, fill_color="light_bg", alpha=214)
        add_rect(slide, x, y, 0.08, row_h, "accent" if index == 0 else "primary", palette=palette)
        add_text(slide, number, x + 0.26, y + 0.22, 0.68, 0.34, 15, True, "primary", "center", palette=palette, colors=colors, min_font_size=9, max_font_size=False)
        add_text(slide, clamp_text(title, text_limit(panel_w - 1.35, row_h - 0.2, 13.6, 1.04)), x + 1.05, y + 0.18, panel_w - 1.35, row_h - 0.25, 13.6, True, "text", palette=palette, colors=colors, min_font_size=9, max_font_size=False, vertical_alignment="middle", line_spacing=0.95)


def split_agenda_text(text: str) -> list[str]:
    text = clean(text)
    if not text:
        return []
    parts = [part.strip(" /，,;；") for part in re.split(r"\s*/\s*|[；;]\s*|\n+", text) if part.strip(" /，,;；")]
    if len(parts) > 1:
        return parts
    return [text]


def _render_narrative_panel(
    slide,
    colors: dict,
    palette: str,
    components: list[dict[str, Any]],
    left: float,
    top: float,
    width: float,
    height: float,
):
    narrative = first_component(components, "argument_block") or first_component(components, "paragraph") or first_component(components, "text_block")
    lists = [c for c in components if c.get("type") in {"list", "numbered_list", "bullet_list", "evidence_list"} or c.get("items")]
    list_ids = {id(c) for c in lists}
    supporting = [c for c in components if c is not narrative and id(c) not in list_ids and c.get("type") not in {"divider", "shape", "arrow"}]

    if supporting and width >= 9.0:
        main_w = width * 0.66
        side_left = left + main_w + 0.28
        side_w = width - main_w - 0.28
    else:
        main_w = width
        side_left = left
        side_w = width

    add_glass_panel(slide, left, top, main_w, height, palette=palette, fill_color="light_bg", alpha=214)
    inner_left = left + 0.35
    inner_w = main_w - 0.7
    y = top + 0.34
    bottom = top + height - 0.34

    if narrative:
        heading = clean(narrative.get("title") or narrative.get("label"))
        if heading:
            add_text(slide, clamp_text(heading, text_limit(inner_w, 0.34, 15.5, 0.96)), inner_left, y, inner_w, 0.34, 15.5, True, "primary", palette=palette, colors=colors, min_font_size=10, max_font_size=False)
            y += 0.46
        list_count = min(2, len(lists))
        reserved_for_lists = min(2.05, max(1.12, list_count * 0.92)) if list_count else 0
        body_h = max(1.4, bottom - y - reserved_for_lists - (0.24 if list_count else 0))
        body = clean(narrative.get("body") or narrative.get("text") or narrative.get("description")) or component_body(narrative)
        is_argument = narrative.get("type") == "argument_block"
        body_font = 12.2 if is_argument else 12.8
        body_ratio = 2.05 if is_argument else 1.72
        add_text(slide, clamp_text(body, text_limit(inner_w, body_h, body_font, body_ratio)), inner_left, y, inner_w, body_h, body_font, color="text", palette=palette, colors=colors, min_font_size=9.5, max_font_size=False, vertical_alignment="top", line_spacing=0.98)
        y += body_h + 0.24

    for item in lists[:3]:
        if y >= bottom - 0.58:
            break
        list_h = min(1.32, max(0.74, bottom - y))
        _render_list_block(slide, colors, palette, item, inner_left, y, inner_w, list_h)
        y += list_h + 0.16

    if supporting and width >= 9.0:
        _render_cards(slide, colors, palette, supporting[:4], side_left, top, side_w, height, compact=True, align_y="middle")


def _render_list_block(slide, colors: dict, palette: str, item: dict[str, Any], left: float, top: float, width: float, height: float):
    title = clean(item.get("title") or item.get("label"))
    items = component_items(item)
    numbered = item.get("type") == "numbered_list"
    if not items and component_body(item):
        items = [component_body(item)]
    if title:
        add_text(slide, clamp_text(title, text_limit(width, 0.26, 12.5, 0.96)), left, top, width, 0.26, 12.5, True, "primary", palette=palette, colors=colors, min_font_size=8.5, max_font_size=False)
        body_top = top + 0.32
        body_h = max(0.3, height - 0.32)
    else:
        body_top = top
        body_h = height
    prefix_items = []
    for i, value in enumerate(items[:5]):
        prefix = f"{i + 1}. " if numbered else "• "
        prefix_items.append(prefix + value)
    add_text(slide, "\n".join(prefix_items), left + 0.02, body_top, width - 0.04, body_h, 11.0, color="secondary", palette=palette, colors=colors, min_font_size=8.5, max_font_size=False, vertical_alignment="top", line_spacing=0.92)


def _render_table(
    slide,
    colors: dict,
    palette: str,
    table_component: dict[str, Any],
    recommendations: list[dict[str, Any]],
    left: float,
    top: float,
    width: float,
    height: float,
):
    data = table_component.get("data") if isinstance(table_component.get("data"), dict) else {}
    headers = data.get("headers") if isinstance(data.get("headers"), list) else []
    rows = data.get("rows") if isinstance(data.get("rows"), list) else []
    headers = [clean(header) for header in headers if clean(header)]
    normalized_rows: list[list[str]] = []
    for row in rows:
        if isinstance(row, list):
            normalized_rows.append([clean(cell) for cell in row])
        elif isinstance(row, dict):
            normalized_rows.append([clean(row.get(header)) for header in headers])
        elif clean(row):
            normalized_rows.append([clean(row)])

    if not headers and normalized_rows:
        headers = [f"列{i + 1}" for i in range(max(len(row) for row in normalized_rows))]
    if not headers:
        headers = ["维度", "说明"]
        normalized_rows = [["待补充", component_body(table_component) or "围绕关键维度补充对比信息"]]

    col_count = max(1, min(len(headers), 5))
    row_count = max(1, min(len(normalized_rows), 6))
    headers = headers[:col_count]
    normalized_rows = [(row + [""] * col_count)[:col_count] for row in normalized_rows[:row_count]]

    recommendation = recommendations[0] if recommendations else {}
    rec_text = component_body(recommendation) or clean(data.get("recommendation") or table_component.get("recommendation"))
    table_h = height - (0.78 if rec_text else 0)
    add_glass_panel(slide, left, top, width, table_h, palette=palette, fill_color="light_bg", alpha=214)

    header_h = 0.55
    gap = 0.0
    row_h = max(0.54, min(0.78, (table_h - header_h - 0.42) / row_count))
    start_x = left + 0.3
    start_y = top + 0.25
    table_w = width - 0.6
    col_w = table_w / col_count
    highlight_column = data.get("highlight_column")
    if isinstance(highlight_column, str) and highlight_column in headers:
        highlight_column = headers.index(highlight_column)
    if not isinstance(highlight_column, int):
        highlight_column = -1

    for col, header in enumerate(headers):
        x = start_x + col * col_w
        fill = "accent" if col == highlight_column else "primary"
        add_rect(slide, x, start_y, col_w - gap, header_h, fill, palette=palette)
        add_text(slide, clamp_text(header, text_limit(col_w - 0.18, 0.28, 10.8, 0.96)), x + 0.09, start_y + 0.14, col_w - 0.18, 0.22, 10.8, True, "background", "center", palette=palette, colors=colors, min_font_size=7.5, max_font_size=False)

    for row_index, row in enumerate(normalized_rows):
        y = start_y + header_h + row_index * row_h
        for col, cell in enumerate(row):
            x = start_x + col * col_w
            fill = "background" if row_index % 2 == 0 else "light_bg"
            line = "accent" if col == highlight_column else "divider"
            add_round_rect(slide, x, y, col_w - gap, row_h, fill, palette=palette, line_color=line, line_width=0.5)
            text_color = "text" if col == 0 else "secondary"
            add_text(slide, clamp_text(cell, text_limit(col_w - 0.2, row_h - 0.12, 9.6, 0.9)), x + 0.1, y + 0.08, col_w - 0.2, row_h - 0.12, 9.6, col == 0, text_color, "center" if col > 0 else "left", palette=palette, colors=colors, min_font_size=7, max_font_size=False, line_spacing=0.9)

    if rec_text:
        rec_top = top + table_h + 0.22
        add_glass_panel(slide, left, rec_top, width, 0.56, palette=palette, fill_color="background", alpha=220)
        add_rect(slide, left, rec_top, 0.08, 0.56, "accent", palette=palette)
        add_text(slide, clamp_text(rec_text, text_limit(width - 0.55, 0.34, 11.0, 0.95)), left + 0.26, rec_top + 0.12, width - 0.55, 0.32, 11.0, True, "text", palette=palette, colors=colors, min_font_size=8, max_font_size=False)


def _render_primitive_strip(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float):
    tags = [c for c in components if c.get("type") in {"tag", "icon"} and (component_text(c) or clean(c.get("icon")))]
    tags = tags[:4]
    if not tags:
        if any(c.get("type") == "divider" for c in components):
            add_line(slide, left, top + 0.2, left + width, top + 0.2, "divider", 0.9, palette=palette)
        return
    pill_w = min(1.35, max(0.85, (width - 0.12 * (len(tags) - 1)) / max(len(tags), 1)))
    for index, item in enumerate(tags):
        x = left + index * (pill_w + 0.12)
        label = clean(item.get("icon") or item.get("title") or item.get("text") or item.get("role"))
        label = clamp_text(label, 8)
        if item.get("type") == "icon":
            add_ellipse(slide, x, top, 0.34, 0.34, "light_bg", palette=palette, line_color="divider", line_width=0.45)
            add_text(slide, label[:2], x + 0.02, top + 0.06, 0.3, 0.18, 8.5, True, "primary", "center", palette=palette, colors=colors, min_font_size=6, max_font_size=False)
            if pill_w > 0.92:
                add_text(slide, label, x + 0.42, top + 0.05, pill_w - 0.42, 0.22, 8.5, color="secondary", palette=palette, colors=colors, min_font_size=6.5, max_font_size=False)
        else:
            add_round_rect(slide, x, top, pill_w, 0.34, "light_bg", palette=palette, line_color="divider", line_width=0.35)
            add_text(slide, label, x + 0.08, top + 0.06, pill_w - 0.16, 0.18, 8.5, True, "primary", "center", palette=palette, colors=colors, min_font_size=6.5, max_font_size=False)


def _render_architecture_diagram(
    slide,
    colors: dict,
    palette: str,
    boxes: list[dict[str, Any]],
    primitives: list[dict[str, Any]],
    left: float,
    top: float,
    width: float,
    height: float,
):
    boxes = boxes[:6]
    add_glass_panel(slide, left, top, width, height, palette=palette, fill_color="light_bg", alpha=206)
    if not boxes:
        return
    count = len(boxes)
    cols = 1 if count == 1 else (2 if count <= 4 and width < 8 else min(3, count))
    rows = max(1, math.ceil(count / cols))
    gap_x = 0.28
    gap_y = 0.32
    box_w = (width - 0.9 - gap_x * (cols - 1)) / cols
    box_h = min(1.28, max(0.98, (height - 0.85 - gap_y * (rows - 1)) / rows))
    total_h = rows * box_h + (rows - 1) * gap_y
    start_x = left + 0.45
    start_y = top + max(0.42, (height - total_h) / 2)
    centers: list[tuple[float, float]] = []
    for index, item in enumerate(boxes):
        row = index // cols
        col = index % cols
        row_count = min(cols, count - row * cols)
        row_w = row_count * box_w + (row_count - 1) * gap_x
        x = start_x + max(0, (width - 0.9 - row_w) / 2) + col * (box_w + gap_x)
        y = start_y + row * (box_h + gap_y)
        centers.append((x + box_w / 2, y + box_h / 2))
        add_round_rect(slide, x, y, box_w, box_h, "background", palette=palette, line_color="divider", line_width=0.55)
        layer = clean(item.get("role") or item.get("relation"))
        if layer:
            add_text(slide, clamp_text(layer, 12), x + 0.15, y + 0.12, box_w - 0.3, 0.2, 8.5, True, "accent", palette=palette, colors=colors, min_font_size=6.5, max_font_size=False)
            title_y = y + 0.34
        else:
            title_y = y + 0.18
        title = clamp_text(clean(item.get("title") or item.get("text") or item.get("label")), text_limit(box_w - 0.35, 0.32, 13.5, 0.96))
        body = clamp_text(clean(item.get("body") or item.get("description")), text_limit(box_w - 0.35, box_h - 0.62, 9.8, 0.86))
        add_text(slide, title or f"模块 {index + 1}", x + 0.18, title_y, box_w - 0.35, 0.3, 13.5, True, "text", "center", palette=palette, colors=colors, min_font_size=8, max_font_size=False)
        add_text(slide, body, x + 0.18, y + box_h - 0.46, box_w - 0.35, 0.34, 9.2, color="secondary", alignment="center", palette=palette, colors=colors, min_font_size=7, max_font_size=False, line_spacing=0.92)

    relation_labels = [component_text(c) for c in primitives if c.get("type") == "arrow" and component_text(c)]
    for index in range(len(centers) - 1):
        x1, y1 = centers[index]
        x2, y2 = centers[index + 1]
        same_row = abs(y1 - y2) < 0.12
        if same_row:
            arrow_left = x1 + box_w / 2 - 0.04
            arrow_right = x2 - box_w / 2 + 0.04
            if arrow_right - arrow_left > 0.2:
                add_arrow(slide, arrow_left, y1 - 0.045, arrow_right, y1 + 0.045, "accent", 1.0, palette=palette)
                if index < len(relation_labels):
                    add_text(slide, clamp_text(relation_labels[index], 12), arrow_left, y1 - 0.33, arrow_right - arrow_left, 0.18, 7.5, color="secondary", alignment="center", palette=palette, colors=colors, min_font_size=5.5, max_font_size=False)
        else:
            add_line(slide, x1, y1 + box_h / 2 - 0.04, x1, y2 - box_h / 2 + 0.04, "divider", 0.85, palette=palette)
            add_line(slide, x1, y2 - box_h / 2 + 0.04, x2 - box_w / 2 + 0.04, y2 - box_h / 2 + 0.04, "divider", 0.85, palette=palette)


def choose_card_columns(count: int, width: float, compact: bool) -> int:
    if compact or width < 6.0:
        return 1
    if count <= 1:
        return 1
    if count in {2, 4}:
        return 2
    if count in {3, 5, 6}:
        return 3
    return 4 if width >= 10.5 and count >= 7 else 3


def component_accent(item: dict[str, Any]) -> str:
    component_type = clean(item.get("type"))
    if item.get("emphasis") == "primary" or component_type in {"recommendation", "insight", "decision_item"}:
        return "accent"
    if component_type in {"risk_item", "opportunity_item"}:
        return "secondary"
    return "primary"


def component_label(item: dict[str, Any], index: int) -> str:
    component_type = clean(item.get("type"))
    explicit = clean(item.get("title") or item.get("label"))
    if explicit:
        return explicit
    if component_type in LIST_TYPES:
        return display_component_type(component_type) or "要点"
    if component_type in QUOTE_TYPES:
        return "引用"
    if component_type in SOURCE_TYPES:
        return "来源"
    return display_component_type(component_type) or clean(item.get("text")) or f"{index:02d}"


def component_body(item: dict[str, Any]) -> str:
    component_type = clean(item.get("type"))
    if component_type in LIST_TYPES:
        return clean(item.get("items") or item.get("text") or item.get("body") or item.get("description"))
    if component_type in METRIC_TYPES:
        data = item.get("data") if isinstance(item.get("data"), dict) else {}
        return clean(item.get("body") or item.get("description") or data.get("baseline") or data.get("delta"))
    if component_type in QUOTE_TYPES:
        return clean(item.get("text") or item.get("body") or item.get("description"))
    return clean(item.get("body") or item.get("description") or item.get("text") or item.get("items"))


def component_items(item: dict[str, Any]) -> list[str]:
    raw = item.get("items")
    if isinstance(raw, list):
        return [clean(value) for value in raw if clean(value)]
    text = clean(raw or item.get("body") or item.get("text") or item.get("description"))
    if not text:
        return []
    lines = [line.strip(" -•\t") for line in re.split(r"[\n\r]+", text) if line.strip(" -•\t")]
    if len(lines) > 1:
        return lines
    return [text]


def _render_metric_grid(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    cols = 2 if len(components) <= 4 else 3
    rows = (len(components[:6]) + cols - 1) // cols
    gap = 0.24
    card_w = (width - gap * (cols - 1)) / cols
    card_h = (height - gap * (rows - 1)) / max(rows, 1)
    for i, item in enumerate(components[:6]):
        data = item.get("data") if isinstance(item.get("data"), dict) else {}
        value = clean(data.get("value") or item.get("text") or item.get("title") or f"{i + 1}")
        label = clean(data.get("label") or item.get("body") or item.get("description") or item.get("title"))
        x = left + (i % cols) * (card_w + gap)
        y = top + (i // cols) * (card_h + gap)
        add_glass_panel(slide, x, y, card_w, card_h, palette=palette, fill_color="light_bg", alpha=214)
        add_text(slide, clamp_text(value, text_limit(card_w - 0.55, 0.72, 34, 0.95)), x + 0.28, y + 0.28, card_w - 0.55, 0.72, 34, True, "primary", palette=palette, colors=colors, min_font_size=20, max_font_size=False)
        add_text(slide, clamp_text(label, text_limit(card_w - 0.6, 0.55, 13, 0.9)), x + 0.3, y + 1.05, card_w - 0.6, 0.55, 13, color="text", palette=palette, colors=colors, min_font_size=9, max_font_size=False)


def _render_flow(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    count = max(1, min(len(components), 6))
    gap = 0.18
    step_w = (width - gap * (count - 1)) / count
    for i, item in enumerate(components[:count]):
        x = left + i * (step_w + gap)
        add_round_rect(slide, x, top, step_w, height, "light_bg", palette=palette, line_color="divider", line_width=0.4)
        add_text(slide, f"{i + 1:02d}", x + 0.14, top + 0.16, step_w - 0.28, 0.38, 15, True, "primary", "center", palette=palette, colors=colors)
        title = clamp_text(clean(item.get("title") or item.get("text")), text_limit(step_w - 0.32, 0.44, 14, 0.95))
        body = clamp_text(clean(item.get("body") or item.get("description")), text_limit(step_w - 0.32, 0.62, 10.5, 0.88))
        add_text(slide, title, x + 0.16, top + 0.68, step_w - 0.32, 0.44, 14, True, "text", "center", palette=palette, colors=colors, min_font_size=9, max_font_size=False)
        add_text(slide, body, x + 0.16, top + 1.18, step_w - 0.32, 0.62, 10.5, color="secondary", alignment="center", palette=palette, colors=colors, min_font_size=8, max_font_size=False)


def _render_quadrant(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    buckets = [
        ("优势", "fact_card", "primary"),
        ("机会", "opportunity_item", "accent"),
        ("风险", "risk_item", "secondary"),
        ("建议", "recommendation", "accent"),
    ]
    gap = 0.22
    card_w = (width - gap) / 2
    card_h = (height - gap) / 2
    used: set[int] = set()
    for index, (fallback_title, preferred_type, accent) in enumerate(buckets):
        item = next((c for c in components if id(c) not in used and c.get("type") == preferred_type), None)
        if item is None:
            item = next((c for c in components if id(c) not in used), component(preferred_type, title=fallback_title, body="围绕该维度补充关键判断"))
        used.add(id(item))
        x = left + (index % 2) * (card_w + gap)
        y = top + (index // 2) * (card_h + gap)
        add_glass_panel(slide, x, y, card_w, card_h, palette=palette, fill_color="light_bg", alpha=212)
        add_rect(slide, x, y, card_w, 0.08, accent, palette=palette)
        title = clean(item.get("title") or item.get("label") or fallback_title)
        body = clean(item.get("body") or item.get("description") or item.get("text") or item.get("items"))
        add_text(slide, clamp_text(title, text_limit(card_w - 0.55, 0.45, 18, 0.95)), x + 0.28, y + 0.3, card_w - 0.55, 0.45, 18, True, "text", palette=palette, colors=colors, min_font_size=12, max_font_size=False)
        add_text(slide, clamp_text(body, text_limit(card_w - 0.55, card_h - 0.95, 12, 0.9)), x + 0.28, y + 0.9, card_w - 0.55, card_h - 1.05, 12, color="secondary", palette=palette, colors=colors, min_font_size=8.5, max_font_size=False, line_spacing=0.95)


def _render_kanban(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    columns = [("待判断", "fact_card"), ("推进中", "milestone"), ("需关注", "risk_item")]
    gap = 0.22
    col_w = (width - gap * 2) / 3
    grouped: list[list[dict[str, Any]]] = [[] for _ in columns]
    for index, item in enumerate(components[:9]):
        target = 2 if item.get("type") == "risk_item" else (1 if item.get("type") in {"milestone", "decision_item"} else index % 3)
        grouped[target].append(item)
    for index, (label, _) in enumerate(columns):
        x = left + index * (col_w + gap)
        add_glass_panel(slide, x, top, col_w, height, palette=palette, fill_color="light_bg", alpha=204)
        add_text(slide, label, x + 0.2, top + 0.2, col_w - 0.4, 0.38, 15, True, "primary", palette=palette, colors=colors, min_font_size=10, max_font_size=False)
        _render_cards(slide, colors, palette, grouped[index], x + 0.18, top + 0.78, col_w - 0.36, height - 0.98, compact=True, align_y="top")


def _render_brand_focus(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    center = first_component(components, "key_point") or first_component(components, "callout") or (components[0] if components else component("key_point", title="核心主张", body="用一句话明确本页最重要的判断"))
    cx = left + width / 2
    cy = top + height / 2
    add_ellipse(slide, cx - 1.42, cy - 1.0, 2.84, 2.0, "light_bg", palette=palette, line_color="divider", line_width=0.6)
    add_text(slide, clamp_text(clean(center.get("title") or center.get("text")), text_limit(2.2, 0.48, 18, 0.95)), cx - 1.1, cy - 0.44, 2.2, 0.48, 18, True, "primary", "center", palette=palette, colors=colors, min_font_size=11, max_font_size=False)
    add_text(slide, clamp_text(clean(center.get("body") or center.get("description")), text_limit(2.2, 0.5, 10.5, 0.9)), cx - 1.1, cy + 0.1, 2.2, 0.5, 10.5, color="secondary", alignment="center", palette=palette, colors=colors, min_font_size=8, max_font_size=False)
    satellites = [c for c in components if c is not center][:6]
    positions = [
        (left + 0.25, top + 0.2),
        (left + width - 3.05, top + 0.2),
        (left + 0.25, top + height - 1.65),
        (left + width - 3.05, top + height - 1.65),
        (left + 0.9, cy - 0.72),
        (left + width - 3.7, cy - 0.72),
    ]
    for item, (x, y) in zip(satellites, positions):
        add_line(slide, cx, cy, x + 1.35, y + 0.62, "divider", 0.8, palette=palette)
        _render_cards(slide, colors, palette, [item], x, y, 2.7, 1.25, compact=True, align_y="middle")


def _render_region_map(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    map_item = first_component(components, "map") or first_component(components, "image")
    panel_w = width * 0.55
    add_glass_panel(slide, left, top, panel_w, height, palette=palette, fill_color="light_bg", alpha=204)
    add_text(slide, clamp_text(clean(map_item.get("title") or "区域关系"), text_limit(panel_w - 0.8, 0.42, 18, 0.95)), left + 0.4, top + 0.42, panel_w - 0.8, 0.42, 18, True, "primary", "center", palette=palette, colors=colors, min_font_size=11, max_font_size=False)
    add_text(slide, clamp_text(clean(map_item.get("body") or map_item.get("description") or "以空间位置、路线或业务版图解释各区域之间的关系"), text_limit(panel_w - 1.2, 0.78, 13, 0.9)), left + 0.6, top + height / 2 - 0.4, panel_w - 1.2, 0.78, 13, color="secondary", alignment="center", palette=palette, colors=colors, min_font_size=9, max_font_size=False)
    for i, dot in enumerate([0.22, 0.45, 0.67, 0.82]):
        add_ellipse(slide, left + panel_w * dot, top + height * (0.3 + 0.12 * (i % 2)), 0.16, 0.16, "accent" if i == 1 else "primary", palette=palette)
    side_components = [c for c in components if c.get("type") not in MEDIA_TYPES][:5]
    _render_cards(slide, colors, palette, side_components, left + panel_w + 0.35, top, width - panel_w - 0.35, height, compact=True, align_y="middle")


def _render_image_hero(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    image_component = first_component(components, "image")
    image_path = component_image_path(image_component)
    photo = resolve_photo(image_path=image_path, text=component_text(image_component)) if image_path else None
    if photo:
        add_cropped_photo(slide, photo, left, top, width, height)
        add_glass_panel(slide, left + width - 4.55, top + 0.35, 4.35, height - 0.7, palette=palette, fill_color="background", alpha=222)
    callouts = [c for c in components if c.get("type") in CARD_TYPES | METRIC_TYPES | LIST_TYPES]
    if not callouts:
        callouts = [component("callout", title="核心画面", body="用背景图建立主题氛围，再用少量文字完成判断")]
    _render_cards(slide, colors, palette, callouts[:3], left + width - 4.35, top + 0.55, 4.0, height - 1.1, compact=True, align_y="middle")


def _render_chart_placeholder(slide, colors: dict, palette: str, item: dict[str, Any], left: float, top: float, width: float, height: float):
    add_glass_panel(slide, left, top, width, height, palette=palette, fill_color="light_bg", alpha=214)
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    labels = data.get("labels") if isinstance(data.get("labels"), list) else ["A", "B", "C"]
    datasets = data.get("datasets") if isinstance(data.get("datasets"), list) else [{"values": [3, 5, 4]}]
    values = datasets[0].get("values", [3, 5, 4]) if datasets and isinstance(datasets[0], dict) else [3, 5, 4]
    max_value = max([float(v) for v in values if isinstance(v, (int, float))] or [1])
    chart_left = left + 0.55
    chart_bottom = top + height - 0.62
    bar_w = max(0.22, (width - 1.4) / max(len(labels), 1) * 0.56)
    slot = (width - 1.4) / max(len(labels), 1)
    for i, label in enumerate(labels[:8]):
        value = values[i] if i < len(values) and isinstance(values[i], (int, float)) else max_value * 0.55
        bar_h = (height - 1.5) * float(value) / max_value
        x = chart_left + i * slot + (slot - bar_w) / 2
        y = chart_bottom - bar_h
        add_rect(slide, x, y, bar_w, bar_h, "primary" if i == 0 else "accent", palette=palette)
        add_text(slide, clean(label), x - 0.18, chart_bottom + 0.08, bar_w + 0.36, 0.28, 9, color="secondary", alignment="center", palette=palette, colors=colors)


def first_component(components: list[dict[str, Any]], component_type: str) -> dict[str, Any]:
    for item in components:
        if item.get("type") == component_type:
            return item
    return {}


def first_value(components: list[dict[str, Any]], component_type: str) -> str:
    return component_text(first_component(components, component_type))


def component_text(item: dict[str, Any]) -> str:
    return clean(item.get("text") or item.get("body") or item.get("description") or item.get("title") or item.get("items"))


def component_image_path(item: dict[str, Any]) -> str:
    if not item:
        return ""
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    return clean(
        item.get("local_path")
        or item.get("asset_path")
        or item.get("image_path")
        or item.get("path")
        or data.get("local_path")
        or data.get("asset_path")
        or data.get("image_path")
        or data.get("path")
        or item.get("asset_id")
    )
