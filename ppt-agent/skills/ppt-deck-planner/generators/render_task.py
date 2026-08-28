#!/usr/bin/env python3
"""Render one slide from tasks.json by task_id.

This deterministic renderer keeps LLM work in planning and turns a SlideSpec
into generator calls. Callers may provide ``output_file``; otherwise the script
derives ``<task_id>.pptx``.
"""

from __future__ import annotations

import argparse
import inspect
import json
import re
import sys
from pathlib import Path
from typing import Any, Callable


DEFAULT_PALETTE = "ocean_soft"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--work-dir", required=True)
    parser.add_argument("--skills-dir", required=True)
    parser.add_argument("--task-id", required=True)
    args = parser.parse_args()

    work_dir = Path(args.work_dir).resolve()
    skills_dir = Path(args.skills_dir).resolve()
    deck_planner_dir = skills_dir / "ppt-deck-planner"
    sys.path.insert(0, str(deck_planner_dir))

    from generators import (  # pylint: disable=import-error,import-outside-toplevel
        new_presentation,
        save_slide,
        generate_agenda,
        generate_brand_focus,
        generate_card_grid,
        generate_case_study,
        generate_chart_slide,
        generate_comparison_table,
        generate_content_slide,
        generate_deep_dive,
        generate_example_detail,
        generate_icon_grid,
        generate_image_hero,
        generate_image_text,
        generate_kanban,
        generate_kpi_dashboard,
        generate_process_flow,
        generate_quote_slide,
        generate_region_map,
        generate_section_divider,
        generate_stat_slide,
        generate_summary_slide,
        generate_swot_analysis,
        generate_three_column,
        generate_timeline,
        generate_title_slide,
        generate_two_column,
    )

    manifest = read_json(work_dir / "tasks.json")
    task = find_task(manifest, args.task_id)
    if not task:
        raise SystemExit(f"task_id {args.task_id} not found")

    palette = manifest_palette(manifest)
    content_type = normalize_content_type(task.get("content_type", "content_slide"))
    background = background_from_task(task, work_dir)
    output_file = clean_text(task.get("output_file")) or default_output_file(args.task_id)
    output_path = work_dir / output_file

    generators: dict[str, Callable[..., Any]] = {
        "agenda": generate_agenda,
        "brand_focus": generate_brand_focus,
        "card_grid": generate_card_grid,
        "case_study": generate_case_study,
        "chart_slide": generate_chart_slide,
        "comparison_table": generate_comparison_table,
        "content_slide": generate_content_slide,
        "deep_dive": generate_deep_dive,
        "example_detail": generate_example_detail,
        "icon_grid": generate_icon_grid,
        "image_hero": generate_image_hero,
        "image_text": generate_image_text,
        "kanban": generate_kanban,
        "kpi_dashboard": generate_kpi_dashboard,
        "process_flow": generate_process_flow,
        "quote_slide": generate_quote_slide,
        "region_map": generate_region_map,
        "section_divider": generate_section_divider,
        "stat_slide": generate_stat_slide,
        "summary_slide": generate_summary_slide,
        "swot_analysis": generate_swot_analysis,
        "three_column": generate_three_column,
        "timeline": generate_timeline,
        "title_slide": generate_title_slide,
        "two_column": generate_two_column,
    }

    prs = new_presentation(palette=palette)
    generator = generators.get(content_type, generate_content_slide)
    params = build_params(content_type, task, manifest, work_dir=work_dir)
    params.update({"prs": prs, "palette": palette, "background": background})
    params = accepted_params(generator, params)
    generated = generator(**params)
    save_slide(generated.slides[-1], str(output_path))
    print(json.dumps({
        "ok": True,
        "task_id": task.get("task_id"),
        "content_type": content_type,
        "output_file": output_file,
    }, ensure_ascii=False))
    return 0


def read_json(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8-sig") as f:
        return json.load(f)


def find_task(manifest: dict[str, Any], task_id: str) -> dict[str, Any] | None:
    for task in manifest.get("tasks", []):
        if str(task.get("task_id")) == str(task_id):
            return task
    return None


def default_output_file(task_id: str) -> str:
    safe = re.sub(r"[^A-Za-z0-9._-]+", "_", clean_text(task_id)).strip("._-")
    return f"{safe or 'slide'}.pptx"


def manifest_palette(manifest: dict[str, Any]) -> str:
    return clean_text(manifest.get("theme") or manifest.get("palette")) or DEFAULT_PALETTE


def normalize_content_type(content_type: str) -> str:
    return (content_type or "").strip() or "content_slide"


def auto_image_text_variant(task: dict[str, Any]) -> str:
    variants = ["image_left", "image_right", "image_top_band", "image_bottom_band"]
    page_index = safe_int(task.get("page_index"), 1)
    return variants[(max(1, page_index) - 1) % len(variants)]


def accepted_params(func: Callable[..., Any], params: dict[str, Any]) -> dict[str, Any]:
    sig = inspect.signature(func)
    if any(param.kind == inspect.Parameter.VAR_KEYWORD for param in sig.parameters.values()):
        return params
    return {k: v for k, v in params.items() if k in sig.parameters}


def build_params(content_type: str, task: dict[str, Any], manifest: dict[str, Any], work_dir: Path | None = None) -> dict[str, Any]:
    plan = task.get("content_plan") or {}
    title = task.get("title") or plan.get("summary") or "页面"
    summary = plan.get("summary") or plan.get("slide_intent") or title
    source = extract_source(plan)
    items = extract_items(plan, task)
    cards = extract_cards(plan, items)
    components = semantic_components(plan, work_dir=work_dir)
    layout_variant = task.get("layout_variant") or nested_get(plan, "visual_intent", "preferred_variant") or ""
    if content_type == "image_text" and not layout_variant:
        layout_variant = auto_image_text_variant(task)

    if content_type == "title_slide":
        return {
            "title": title,
            "subtitle": summary,
            "kicker": plan.get("kicker", ""),
            "author": plan.get("author", ""),
            "date": plan.get("date", ""),
            "source": source,
            "layout_variant": layout_variant,
            "components": components,
        }
    if content_type == "section_divider":
        return {
            "number": section_number(task, manifest),
            "title": title,
            "subtitle": summary,
            "kicker": plan.get("kicker", "章节"),
            "source": source,
            "layout_variant": layout_variant,
            "components": components,
        }
    if content_type == "agenda":
        agenda_items = agenda_items_from_manifest(manifest, task) or items
        agenda_components = [
            {"type": "toc_item", "title": agenda_item_title(item), "body": item}
            for item in agenda_items[:8]
        ]
        return {"title": title, "items": agenda_items[:8], "kicker": "目录", "source": source, "components": agenda_components or components}
    if content_type == "summary_slide":
        return {
            "title": title,
            "key_points": ensure_items(items, summary, 4),
            "thank_you": plan.get("thank_you", "感谢聆听"),
            "contact": plan.get("contact", ""),
            "kicker": plan.get("kicker", "总结"),
            "source": source,
            "components": components,
        }
    if content_type == "quote_slide":
        quote = first_by_type(plan, "quote") or summary
        return {
            "quote": quote,
            "attribution": plan.get("attribution") or " ",
            "kicker": plan.get("kicker", "观点"),
            "source": source,
            "components": components,
        }
    if content_type == "card_grid":
        return {
            "title": title,
            "cards": cards[:6],
            "layout": "2x2" if len(cards) <= 4 else "3x2",
            "layout_variant": layout_variant,
            "subtitle": summary,
            "kicker": plan.get("kicker", "要点"),
            "source": source,
            "components": components,
        }
    if content_type == "two_column":
        left, right = split_items(items, summary)
        headers = extract_headers(plan, ["左侧", "右侧"])
        return {
            "title": title,
            "left_header": headers[0],
            "right_header": headers[1],
            "left_bullets": left,
            "right_bullets": right,
            "layout_variant": layout_variant,
            "kicker": plan.get("kicker", "对比"),
            "source": source,
            "components": components,
        }
    if content_type == "three_column":
        groups = split_three(items, summary)
        headers = extract_headers(plan, ["一", "二", "三"])
        return {
            "title": title,
            "columns": [
                {"header": headers[0], "bullets": groups[0]},
                {"header": headers[1], "bullets": groups[1]},
                {"header": headers[2], "bullets": groups[2]},
            ],
            "kicker": plan.get("kicker", "三维分析"),
            "source": source,
            "components": components,
        }
    if content_type == "image_text":
        return {
            "title": title,
            "header": first_card_title(cards) or "核心观察",
            "paragraph": paragraph_from(items, summary),
            "layout": "right-image",
            "layout_variant": layout_variant,
            "kicker": plan.get("kicker", "图文解读"),
            "sub_header": plan.get("sub_header", ""),
            "source": source,
            "components": components,
        }
    if content_type == "chart_slide":
        return {
            "title": title,
            "subtitle": summary,
            "chart_type": plan.get("chart_type") or "bar",
            "data": chart_data(plan, items),
            "analysis": summary,
            "show_legend": True,
            "kicker": plan.get("kicker", "数据"),
            "source": source,
            "components": components,
        }
    if content_type == "kpi_dashboard":
        return {
            "title": title,
            "subtitle": summary,
            "kpis": kpis(plan, cards),
            "kicker": plan.get("kicker", "指标"),
            "source": source,
            "components": components,
        }
    if content_type == "comparison_table":
        params = comparison_params(title, summary, plan, items, source)
        params["components"] = components
        return params
    if content_type in {"process_flow", "timeline", "icon_grid", "stat_slide", "case_study"}:
        params = generic_structured_params(content_type, title, summary, plan, items, cards, source)
        params["components"] = components
        return params
    if content_type in {"image_hero", "example_detail", "deep_dive", "swot_analysis", "kanban", "brand_focus", "region_map"}:
        return {
            "title": title,
            "subtitle": summary,
            "lede": summary,
            "kicker": plan.get("kicker", ""),
            "layout_variant": layout_variant,
            "source": source,
            "components": components,
        }
    return {
        "title": title,
        "section_header": first_card_title(cards) or "核心要点",
        "bullets": ensure_items(items, summary, 5),
        "lede": summary,
        "layout_variant": layout_variant,
        "kicker": plan.get("kicker", "要点"),
        "source": source,
        "components": components,
    }


def extract_items(plan: dict[str, Any], task: dict[str, Any]) -> list[str]:
    result: list[str] = []
    for component in semantic_components(plan):
        component_type = component.get("type", "")
        if component_type in {"feature_card", "kpi_metric", "chart", "table", "image"}:
            continue
        for item in component.get("items") or []:
            text = clean_text(item)
            if text:
                result.append(text)
        for key in ("text", "body", "title"):
            text = clean_text(component.get(key))
            if text:
                result.append(text)
    if not result:
        result = split_text(plan.get("summary") or plan.get("slide_intent") or task.get("title") or "")
    return [x for x in result if x][:8]


def agenda_items_from_manifest(manifest: dict[str, Any], current_task: dict[str, Any]) -> list[str]:
    tasks = sorted(
        [t for t in manifest.get("tasks", []) if isinstance(t, dict)],
        key=lambda item: safe_int(item.get("page_index"), 9999),
    )
    current_id = str(current_task.get("task_id"))

    def usable(task: dict[str, Any]) -> bool:
        if str(task.get("task_id")) == current_id:
            return False
        if not clean_text(task.get("title")):
            return False
        return normalize_content_type(task.get("content_type", "")) not in {"title_slide", "agenda"}

    sections = [task for task in tasks if usable(task) and normalize_content_type(task.get("content_type", "")) == "section_divider"]
    candidates = sections if len(sections) >= 2 else [task for task in tasks if usable(task)]
    result: list[str] = []
    for index, task in enumerate(candidates[:8], start=1):
        title = clean_text(task.get("title"))
        # Agenda numbers represent chapter order, never the absolute slide
        # index. This keeps the TOC stable when cover, notes, or divider pages
        # are inserted before a section.
        result.append(f"{index:02d}  {title}")
    return result


def agenda_item_title(item: str) -> str:
    text = clean_text(item)
    text = re.sub(r"^\d+\s*", "", text).strip()
    return text[:18] or "目录项"


def safe_int(value: Any, fallback: int) -> int:
    try:
        return int(value)
    except (TypeError, ValueError):
        return fallback


def extract_cards(plan: dict[str, Any], items: list[str]) -> list[dict[str, str]]:
    cards: list[dict[str, str]] = []
    for component in semantic_components(plan, {"feature_card", "key_point", "callout"}):
        title = clean_text(component.get("title") or component.get("text"))
        body = clean_text(component.get("body") or component.get("text"))
        if title or body:
            cards.append({
                "header": title or body[:16],
                "body": body or title,
                "icon": "",
                "emphasis": clean_text(component.get("emphasis")),
            })
    if len(cards) < 3:
        for item in items:
            header, body = split_header_body(item)
            cards.append({"header": header, "body": body, "icon": ""})
    return cards[:6] or [{"header": "核心要点", "body": plan.get("summary", "围绕主题展开分析"), "icon": ""}]


def semantic_components(plan: dict[str, Any], types: set[str] | None = None, work_dir: Path | None = None) -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = []
    for component in plan.get("components") or []:
        if not isinstance(component, dict):
            continue
        component_type = clean_text(component.get("type"))
        if types is not None and component_type not in types:
            continue
        normalized = dict(component)
        normalized["type"] = component_type
        normalize_component_asset_paths(normalized, work_dir)
        result.append(normalized)
    visual_intent = plan.get("visual_intent") if isinstance(plan.get("visual_intent"), dict) else {}
    image_component = visual_intent_to_image_component(visual_intent, work_dir)
    if image_component and (types is None or image_component.get("type") in types):
        result.append(image_component)
    return result


def visual_intent_to_image_component(visual_intent: dict[str, Any], work_dir: Path | None) -> dict[str, Any]:
    image_fields = ("local_path", "asset_path", "image_path", "asset_id", "asset_query", "image_url", "preview_url")
    if not any(clean_text(visual_intent.get(key)) for key in image_fields):
        return {}
    image_component = {
        "id": "visual_intent_image",
        "type": "image",
        "role": clean_text(visual_intent.get("role")),
        "asset_purpose": clean_text(visual_intent.get("asset_purpose")),
        "asset_subject": clean_text(visual_intent.get("asset_subject")),
        "asset_query": clean_text(visual_intent.get("asset_query")),
        "composition": clean_text(visual_intent.get("composition")),
        "caption": clean_text(visual_intent.get("caption")),
        "asset_id": clean_text(visual_intent.get("asset_id")),
        "local_path": clean_text(visual_intent.get("local_path")),
        "image_url": clean_text(visual_intent.get("image_url")),
        "preview_url": clean_text(visual_intent.get("preview_url")),
        "source_url": clean_text(visual_intent.get("source_url")),
        "attribution": clean_text(visual_intent.get("attribution")),
    }
    normalize_component_asset_paths(image_component, work_dir)
    return {key: value for key, value in image_component.items() if value not in ("", None, [], {})}


def normalize_component_asset_paths(component: dict[str, Any], work_dir: Path | None) -> None:
    data = component.get("data") if isinstance(component.get("data"), dict) else {}
    for key in ("local_path", "asset_path", "image_path", "path"):
        value = clean_text(component.get(key) or data.get(key))
        if not value:
            continue
        component[key] = resolve_workdir_path(value, work_dir)
        if key != "local_path" and not clean_text(component.get("local_path")):
            component["local_path"] = component[key]


def resolve_workdir_path(value: str, work_dir: Path | None) -> str:
    value = clean_text(value)
    if not value:
        return value
    if value.startswith("asset:"):
        raise ValueError(f"legacy asset id is unsupported: {value}")
    candidate = Path(value).expanduser()
    if candidate.is_absolute():
        return str(candidate)
    if work_dir is not None:
        return str((work_dir / candidate).resolve())
    return value


def background_from_task(task: dict[str, Any], work_dir: Path | None = None) -> str | None:
    plan = task.get("content_plan") if isinstance(task.get("content_plan"), dict) else {}
    visual_intent = plan.get("visual_intent") if isinstance(plan.get("visual_intent"), dict) else {}
    candidate = background_path_from_visual_item(visual_intent, work_dir)
    if candidate:
        return candidate
    for component in plan.get("components") or []:
        if not isinstance(component, dict):
            continue
        candidate = background_path_from_visual_item(component, work_dir)
        if candidate:
            return candidate
    return None


def background_path_from_visual_item(item: dict[str, Any], work_dir: Path | None = None) -> str:
    if not is_background_visual_item(item):
        return ""
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    for key in ("local_path", "asset_path", "image_path", "path"):
        value = clean_text(item.get(key) or data.get(key))
        if value:
            return resolve_workdir_path(value, work_dir)
    return ""


def is_background_visual_item(item: dict[str, Any]) -> bool:
    if not isinstance(item, dict):
        return False
    purpose = clean_text(item.get("asset_purpose")).lower()
    position = clean_text(item.get("image_position")).lower()
    role = clean_text(item.get("role")).lower()
    return purpose == "background" or position == "background" or role in {"background", "hero_photo"}


def split_header_body(text: str) -> tuple[str, str]:
    text = clean_text(text)
    if ":" in text:
        left, right = text.split(":", 1)
        return left.strip()[:18], right.strip()
    if "：" in text:
        left, right = text.split("：", 1)
        return left.strip()[:18], right.strip()
    return text[:18], text


def split_items(items: list[str], fallback: str) -> tuple[list[str], list[str]]:
    values = ensure_items(items, fallback, 6)
    mid = max(1, len(values) // 2)
    return values[:mid], values[mid:] or values[:mid]


def split_three(items: list[str], fallback: str) -> list[list[str]]:
    values = ensure_items(items, fallback, 6)
    return [values[0::3] or values[:1], values[1::3] or values[:1], values[2::3] or values[:1]]


def extract_headers(plan: dict[str, Any], fallback: list[str]) -> list[str]:
    headers = [
        clean_text(component.get("title"))
        for component in semantic_components(plan)
        if component.get("title")
    ]
    headers = [h for h in headers if h]
    while len(headers) < len(fallback):
        headers.append(fallback[len(headers)])
    return headers[: len(fallback)]


def ensure_items(items: list[str], fallback: str, count: int) -> list[str]:
    values = [clean_text(x) for x in items if clean_text(x)]
    if not values:
        values = split_text(fallback)
    while len(values) < min(count, 3):
        values.append(fallback)
    return values[:count]


def split_text(text: str) -> list[str]:
    text = clean_text(text)
    parts = [p.strip(" ，。；;") for p in re.split(r"[。；;\n]", text) if p.strip()]
    return parts or ([text] if text else [])


def paragraph_from(items: list[str], summary: str) -> str:
    paragraph = "。".join([clean_text(x).rstrip("。") for x in items if clean_text(x)])
    if not paragraph:
        paragraph = clean_text(summary)
    if len(paragraph) < 160:
        paragraph = f"{paragraph}。{clean_text(summary)}"
    return paragraph


def first_by_type(plan: dict[str, Any], element_type: str) -> str:
    for component in semantic_components(plan):
        if component.get("type") == element_type:
            return clean_text(component.get("text") or component.get("body") or component.get("title"))
    return ""


def first_card_title(cards: list[dict[str, str]]) -> str:
    return cards[0].get("header", "") if cards else ""


def nested_get(value: dict[str, Any], *keys: str) -> str:
    current: Any = value
    for key in keys:
        if not isinstance(current, dict):
            return ""
        current = current.get(key)
    return clean_text(current)


def section_number(task: dict[str, Any], manifest: dict[str, Any]) -> str:
    plan = task.get("content_plan") or {}
    explicit = (
        plan.get("section_number")
        or plan.get("number")
        or nested_get(plan if isinstance(plan, dict) else {}, "visual_intent", "section_number")
    )
    if clean_text(explicit):
        return normalize_section_number(clean_text(explicit))

    count = 0
    for candidate in sorted(manifest.get("tasks", []), key=lambda item: int(item.get("page_index") or 0)):
        if normalize_content_type(candidate.get("content_type", "")) != "section_divider":
            continue
        count += 1
        if str(candidate.get("task_id")) == str(task.get("task_id")):
            return f"{count:02d}"
    return "01"


def normalize_section_number(value: str) -> str:
    value = clean_text(value)
    match = re.search(r"\d+", value)
    if match:
        return f"{int(match.group(0)):02d}"
    return value[:4] or "01"


def clean_text(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value.strip()
    if isinstance(value, dict):
        title = clean_text(value.get("title") or value.get("label") or value.get("name"))
        body = clean_text(value.get("body") or value.get("text") or value.get("value"))
        if title and body and title != body:
            return f"{title}: {body}"
        return title or body
    if isinstance(value, list):
        return " / ".join(clean_text(x) for x in value if clean_text(x))
    return str(value).strip()


def extract_source(plan: dict[str, Any]) -> str:
    source = plan.get("source") or plan.get("sources") or ""
    component_sources = [
        clean_text(component.get("source"))
        for component in semantic_components(plan)
        if clean_text(component.get("source"))
    ]
    if component_sources:
        return "；".join(component_sources)
    if isinstance(source, list):
        return "；".join(clean_text(x) for x in source if clean_text(x))
    return clean_text(source)


def chart_data(plan: dict[str, Any], items: list[str]) -> dict[str, Any]:
    for component in semantic_components(plan, {"chart"}):
        data = component.get("data")
        if isinstance(data, dict) and data.get("labels") and data.get("datasets"):
            return data
    data = plan.get("data") or plan.get("chart_data")
    if isinstance(data, dict) and data.get("labels") and data.get("datasets"):
        return data
    labels = plan.get("labels")
    datasets = plan.get("datasets")
    if isinstance(labels, list) and isinstance(datasets, list):
        return {"labels": labels, "datasets": datasets}
    labels = [f"项{i + 1}" for i in range(max(3, min(5, len(items) or 4)))]
    values = [max(10, 100 - i * 12) for i in range(len(labels))]
    return {"labels": labels, "datasets": [{"name": "指标", "values": values}]}


def kpis(plan: dict[str, Any], cards: list[dict[str, str]]) -> list[dict[str, str]]:
    metric_components = semantic_components(plan, {"kpi_metric"})
    if metric_components:
        result = []
        for component in metric_components[:5]:
            data = component.get("data") if isinstance(component.get("data"), dict) else {}
            result.append({
                "value": clean_text(data.get("value") or component.get("text") or component.get("title")),
                "label": clean_text(data.get("label") or component.get("title") or component.get("body")),
                "delta": clean_text(data.get("delta") or component.get("emphasis") or "稳步推进"),
                "baseline": clean_text(data.get("baseline") or component.get("body")),
            })
        return result
    raw = plan.get("kpis") or plan.get("metrics")
    if isinstance(raw, list) and raw:
        return raw[:5]
    result = []
    for i, card in enumerate(cards[:4]):
        result.append({
            "value": f"{i + 1}",
            "label": card.get("header") or f"指标{i + 1}",
            "delta": "稳步推进",
            "baseline": card.get("body") or "",
        })
    return result or [{"value": "1", "label": "核心指标", "delta": "持续改善", "baseline": plan.get("summary", "")}]

def comparison_params(title: str, summary: str, plan: dict[str, Any], items: list[str], source: str) -> dict[str, Any]:
    headers = plan.get("headers")
    rows = plan.get("rows")
    for component in semantic_components(plan, {"table"}):
        data = component.get("data")
        if isinstance(data, dict) and isinstance(data.get("headers"), list) and isinstance(data.get("rows"), list):
            headers = data.get("headers")
            rows = data.get("rows")
            break
    if not isinstance(headers, list) or not isinstance(rows, list):
        left, right = split_items(items, summary)
        headers = ["维度", "方案A", "方案B"]
        rows = [[f"维度{i + 1}", left[i] if i < len(left) else "", right[i] if i < len(right) else ""] for i in range(max(len(left), len(right)))]
    return {
        "title": title,
        "headers": headers,
        "rows": rows,
        "recommendation": plan.get("recommendation") or summary,
        "kicker": plan.get("kicker", "对比"),
        "subtitle": summary,
        "source": source,
    }


def generic_structured_params(
    content_type: str,
    title: str,
    summary: str,
    plan: dict[str, Any],
    items: list[str],
    cards: list[dict[str, str]],
    source: str,
) -> dict[str, Any]:
    if content_type == "process_flow":
        return {
            "title": title,
            "steps": [
                {"num": f"{i + 1:02d}", "title": c.get("header", ""), "desc": c.get("body", "")}
                for i, c in enumerate(cards[:6])
            ],
            "kicker": plan.get("kicker", "流程"),
            "subtitle": summary,
            "source": source,
        }
    if content_type == "timeline":
        return {
            "title": title,
            "nodes": [
                {"year": c.get("header", f"{i + 1:02d}"), "event": c.get("body", ""), "icon": ""}
                for i, c in enumerate(cards[:6])
            ],
            "kicker": plan.get("kicker", "时间线"),
            "subtitle": summary,
            "source": source,
        }
    if content_type == "icon_grid":
        return {
            "title": title,
            "icons": [{"icon": c.get("header", ""), "label": c.get("body", "") or c.get("header", "")} for c in cards[:6]],
            "subtitle": summary,
            "kicker": plan.get("kicker", "能力"),
            "source": source,
        }
    if content_type == "stat_slide":
        return {
            "title": title,
            "stats": [
                {"number": f"{i + 1}", "unit": "", "label": c.get("header", ""), "trend": c.get("body", "")}
                for i, c in enumerate(cards[:4])
            ],
            "subtitle": summary,
            "kicker": plan.get("kicker", "数据"),
            "source": source,
        }
    if content_type == "case_study":
        return {
            "title": title,
            "context": summary,
            "problem": items[0] if items else summary,
            "solution": "；".join(items[1:4]) if len(items) > 1 else summary,
            "results": [{"metric": "结果", "value": str(i + 1), "comparison": item} for i, item in enumerate(ensure_items(items, summary, 4))],
            "kicker": plan.get("kicker", "案例"),
            "source": source,
        }
    return {"title": title, "bullets": ensure_items(items, summary, 5), "source": source}


if __name__ == "__main__":
    raise SystemExit(main())
