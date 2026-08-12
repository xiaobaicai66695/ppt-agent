"""Component-first slide layout engine.

This module replaces rigid single-page drawing logic with semantic components.
Generator entry points keep their public function signatures, but internally
they convert keyword arguments into components and render them through this
shared engine.
"""
from __future__ import annotations

from typing import Any

from pptx import Presentation

from .base import (
    PALETTES,
    add_glass_panel,
    add_rect,
    add_round_rect,
    add_source_line,
    add_text,
    new_presentation,
    resolve_background,
    set_image_background,
    set_slide_background,
)

SLIDE_W = 13.333
SLIDE_H = 7.5

TITLE_TYPES = {"headline", "subheadline", "section_marker"}
CARD_TYPES = {"feature_card", "key_point", "callout"}
LIST_TYPES = {"bullet_list", "paragraph"}
METRIC_TYPES = {"kpi_metric"}
FLOW_TYPES = {"timeline_node", "process_step"}


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
        _render_cards(slide, colors, palette, callouts, 0.86, 5.08, 11.5, 1.1, compact=True)


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

    body_components = [c for c in components if c.get("type") not in TITLE_TYPES]
    charts = [c for c in body_components if c.get("type") == "chart"]
    tables = [c for c in body_components if c.get("type") == "table"]
    metrics = [c for c in body_components if c.get("type") in METRIC_TYPES]
    flows = [c for c in body_components if c.get("type") in FLOW_TYPES]
    cards = [c for c in body_components if c.get("type") in CARD_TYPES]
    lists = [c for c in body_components if c.get("type") in LIST_TYPES or c.get("items")]

    top = 1.72
    height = 5.25
    if content_type in {"chart_slide"} and charts:
        _render_chart_placeholder(slide, colors, palette, charts[0], 0.65, top, 7.45, height)
        _render_cards(slide, colors, palette, metrics + cards + lists, 8.35, top, 4.25, height, compact=True)
    elif content_type in {"kpi_dashboard", "stat_slide"} and metrics:
        _render_metric_grid(slide, colors, palette, metrics, 0.65, top, 11.95, height)
    elif content_type in {"timeline", "process_flow"} and flows:
        _render_flow(slide, colors, palette, flows, 0.72, 3.0, 11.7, 2.1)
    elif content_type in {"two_column", "comparison_table"}:
        left = body_components[0::2]
        right = body_components[1::2]
        _render_cards(slide, colors, palette, left, 0.65, top, 5.75, height, compact=True)
        _render_cards(slide, colors, palette, right, 6.9, top, 5.75, height, compact=True)
    elif content_type in {"image_text", "case_study", "example_detail", "deep_dive"}:
        add_glass_panel(slide, 0.65, top, 4.0, height, palette=palette, fill_color="light_bg", alpha=210)
        image_text = component_text(first_component(body_components, "image")) or "主题视觉"
        add_text(slide, image_text, 0.95, top + 1.9, 3.4, 0.8, 22, True, "primary", "center", palette=palette, colors=colors)
        _render_cards(slide, colors, palette, [c for c in body_components if c.get("type") != "image"], 5.0, top, 7.6, height, compact=True)
    else:
        ordered = []
        seen = set()
        for item in cards + metrics + flows + lists + body_components:
            marker = id(item)
            if marker in seen:
                continue
            seen.add(marker)
            ordered.append(item)
        _render_cards(slide, colors, palette, ordered, 0.65, top, 11.95, height)


def _render_cards(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float, compact: bool = False):
    components = [c for c in components if component_text(c) or clean(c.get("title")) or clean(c.get("body"))][:8]
    if not components:
        components = [component("paragraph", text="围绕主题展开结构化说明")]
    cols = 1 if compact or width < 6.0 else (3 if len(components) >= 5 else 2)
    rows = (len(components) + cols - 1) // cols
    gap = 0.18
    card_w = (width - gap * (cols - 1)) / cols
    card_h = min(1.42 if compact else 1.62, (height - gap * (rows - 1)) / max(rows, 1))
    for i, item in enumerate(components):
        row = i // cols
        col = i % cols
        x = left + col * (card_w + gap)
        y = top + row * (card_h + gap)
        add_glass_panel(slide, x, y, card_w, card_h, palette=palette, fill_color="light_bg", alpha=210)
        add_rect(slide, x, y, 0.06, card_h, "accent" if item.get("emphasis") == "primary" else "primary", palette=palette)
        label = clean(item.get("title") or item.get("label") or item.get("text")) or f"{i + 1:02d}"
        body = clean(item.get("body") or item.get("description") or item.get("text") or item.get("items"))
        add_text(slide, label, x + 0.18, y + 0.13, card_w - 0.35, 0.32, 15 if compact else 17, True, "text", palette=palette, colors=colors)
        add_text(slide, body, x + 0.18, y + 0.52, card_w - 0.35, card_h - 0.62, 11 if compact else 12, color="secondary", palette=palette, colors=colors)


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
        add_text(slide, value, x + 0.28, y + 0.28, card_w - 0.55, 0.72, 34, True, "primary", palette=palette, colors=colors)
        add_text(slide, label, x + 0.3, y + 1.05, card_w - 0.6, 0.55, 13, color="text", palette=palette, colors=colors)


def _render_flow(slide, colors: dict, palette: str, components: list[dict[str, Any]], left: float, top: float, width: float, height: float):
    count = max(1, min(len(components), 6))
    gap = 0.18
    step_w = (width - gap * (count - 1)) / count
    for i, item in enumerate(components[:count]):
        x = left + i * (step_w + gap)
        add_round_rect(slide, x, top, step_w, height, "light_bg", palette=palette, line_color="divider", line_width=0.4)
        add_text(slide, f"{i + 1:02d}", x + 0.14, top + 0.16, step_w - 0.28, 0.38, 15, True, "primary", "center", palette=palette, colors=colors)
        add_text(slide, clean(item.get("title") or item.get("text")), x + 0.16, top + 0.68, step_w - 0.32, 0.44, 14, True, "text", "center", palette=palette, colors=colors)
        add_text(slide, clean(item.get("body") or item.get("description")), x + 0.16, top + 1.18, step_w - 0.32, 0.62, 10.5, color="secondary", alignment="center", palette=palette, colors=colors)


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
