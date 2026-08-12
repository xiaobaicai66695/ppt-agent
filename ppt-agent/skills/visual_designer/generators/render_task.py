#!/usr/bin/env python3
"""Render one slide from tasks.json by task_id.

This script is the deterministic renderer used by the backend worker pool. It
keeps LLM work in planning and turns a SlideSpec into generator calls.
"""

from __future__ import annotations

import argparse
import inspect
import json
import re
import sys
from pathlib import Path
from typing import Any, Callable


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--work-dir", required=True)
    parser.add_argument("--skills-dir", required=True)
    parser.add_argument("--task-id", required=True)
    args = parser.parse_args()

    work_dir = Path(args.work_dir).resolve()
    skills_dir = Path(args.skills_dir).resolve()
    visual_designer_dir = skills_dir / "visual_designer"
    sys.path.insert(0, str(visual_designer_dir))

    from generators import (  # pylint: disable=import-error,import-outside-toplevel
        new_presentation,
        save_slide,
        generate_agenda,
        generate_card_grid,
        generate_case_study,
        generate_chart_slide,
        generate_comparison_table,
        generate_content_slide,
        generate_icon_grid,
        generate_image_text,
        generate_kpi_dashboard,
        generate_process_flow,
        generate_quote_slide,
        generate_section_divider,
        generate_stat_slide,
        generate_summary_slide,
        generate_three_column,
        generate_timeline,
        generate_title_slide,
        generate_two_column,
    )

    manifest = read_json(work_dir / "tasks.json")
    task = find_task(manifest, args.task_id)
    if not task:
        raise SystemExit(f"task_id {args.task_id} not found")

    palette = manifest.get("theme") or "ocean_soft"
    content_type = normalize_content_type(task.get("content_type", "content_slide"))
    background = task.get("background") or None
    output_file = task.get("output_file") or f"{task.get('page_index', args.task_id)}.pptx"
    output_path = work_dir / output_file

    generators: dict[str, Callable[..., Any]] = {
        "agenda": generate_agenda,
        "card_grid": generate_card_grid,
        "case_study": generate_case_study,
        "chart_slide": generate_chart_slide,
        "comparison_table": generate_comparison_table,
        "content_slide": generate_content_slide,
        "icon_grid": generate_icon_grid,
        "image_text": generate_image_text,
        "kpi_dashboard": generate_kpi_dashboard,
        "process_flow": generate_process_flow,
        "quote_slide": generate_quote_slide,
        "section_divider": generate_section_divider,
        "stat_slide": generate_stat_slide,
        "summary_slide": generate_summary_slide,
        "three_column": generate_three_column,
        "timeline": generate_timeline,
        "title_slide": generate_title_slide,
        "two_column": generate_two_column,
    }

    prs = new_presentation(palette=palette)
    generator = generators.get(content_type, generate_content_slide)
    params = build_params(content_type, task, manifest)
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
        if str(task.get("task_id")) == str(task_id) or str(task.get("page_index")) == str(task_id):
            return task
    return None


def normalize_content_type(content_type: str) -> str:
    aliases = {
        "bar_chart": "chart_slide",
        "line_chart": "chart_slide",
        "pie_chart": "chart_slide",
        "doughnut_chart": "chart_slide",
        "table": "comparison_table",
    }
    return aliases.get((content_type or "").strip(), content_type or "content_slide")


def accepted_params(func: Callable[..., Any], params: dict[str, Any]) -> dict[str, Any]:
    sig = inspect.signature(func)
    return {k: v for k, v in params.items() if k in sig.parameters}


def build_params(content_type: str, task: dict[str, Any], manifest: dict[str, Any]) -> dict[str, Any]:
    plan = task.get("content_plan") or {}
    title = task.get("title") or plan.get("summary") or "页面"
    summary = plan.get("summary") or task.get("description") or title
    source = extract_source(plan)
    items = extract_items(plan, task)
    cards = extract_cards(plan, items)
    layout_variant = task.get("layout_variant") or nested_get(plan, "visual_intent", "preferred_variant") or ""

    if content_type == "title_slide":
        return {
            "title": title,
            "subtitle": summary,
            "kicker": plan.get("kicker", ""),
            "author": plan.get("author", ""),
            "date": plan.get("date", ""),
            "source": source,
            "layout_variant": layout_variant,
        }
    if content_type == "section_divider":
        return {
            "number": section_number(task, manifest),
            "title": title,
            "subtitle": summary,
            "kicker": plan.get("kicker", "章节"),
            "source": source,
            "layout_variant": layout_variant,
        }
    if content_type == "agenda":
        agenda_items = items or [
            f"{int(t.get('page_index') or i + 1):02d}  {t.get('title', '')}"
            for i, t in enumerate(manifest.get("tasks", [])[1:7])
            if t.get("title")
        ]
        return {"title": title, "items": agenda_items[:8], "kicker": "目录", "source": source}
    if content_type == "summary_slide":
        return {
            "title": title,
            "key_points": ensure_items(items, summary, 4),
            "thank_you": plan.get("thank_you", "感谢聆听"),
            "contact": plan.get("contact", ""),
            "kicker": plan.get("kicker", "总结"),
            "source": source,
        }
    if content_type == "quote_slide":
        quote = first_by_type(plan, "quote") or summary
        return {
            "quote": quote,
            "attribution": plan.get("attribution") or " ",
            "kicker": plan.get("kicker", "观点"),
            "source": source,
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
        }
    if content_type == "kpi_dashboard":
        return {
            "title": title,
            "subtitle": summary,
            "kpis": kpis(plan, cards),
            "kicker": plan.get("kicker", "指标"),
            "source": source,
        }
    if content_type == "comparison_table":
        return comparison_params(title, summary, plan, items, source)
    if content_type in {"process_flow", "timeline", "icon_grid", "stat_slide", "case_study"}:
        return generic_structured_params(content_type, title, summary, plan, items, cards, source)
    return {
        "title": title,
        "section_header": first_card_title(cards) or "核心要点",
        "bullets": ensure_items(items, summary, 5),
        "lede": summary,
        "layout_variant": layout_variant,
        "kicker": plan.get("kicker", "要点"),
        "source": source,
    }


def extract_items(plan: dict[str, Any], task: dict[str, Any]) -> list[str]:
    result: list[str] = []
    for element in plan.get("elements") or []:
        if not isinstance(element, dict):
            result.append(clean_text(element))
            continue
        for item in element.get("items") or []:
            text = clean_text(item)
            if text:
                result.append(text)
        for key in ("text", "description"):
            text = clean_text(element.get(key))
            if text:
                result.append(text)
    if not result:
        result = split_text(plan.get("summary") or task.get("description") or task.get("title") or "")
    return [x for x in result if x][:8]


def extract_cards(plan: dict[str, Any], items: list[str]) -> list[dict[str, str]]:
    cards: list[dict[str, str]] = []
    for element in plan.get("elements") or []:
        if not isinstance(element, dict):
            continue
        title = clean_text(element.get("title"))
        body = clean_text(element.get("description") or element.get("text"))
        if title or body:
            cards.append({"header": title or body[:16], "body": body or title, "icon": ""})
    if len(cards) < 3:
        for item in items:
            header, body = split_header_body(item)
            cards.append({"header": header, "body": body, "icon": ""})
    return cards[:6] or [{"header": "核心要点", "body": plan.get("summary", "围绕主题展开分析"), "icon": ""}]


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
    headers = [clean_text(e.get("title")) for e in plan.get("elements", []) if isinstance(e, dict) and e.get("title")]
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
    for element in plan.get("elements") or []:
        if isinstance(element, dict) and element.get("type") == element_type:
            return clean_text(element.get("text") or element.get("description") or element.get("title"))
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
        task.get("section_number")
        or plan.get("section_number")
        or plan.get("number")
        or nested_get(plan if isinstance(plan, dict) else {}, "visual_intent", "section_number")
    )
    if clean_text(explicit):
        return normalize_section_number(clean_text(explicit))

    current_page = int(task.get("page_index") or 0)
    count = 0
    for candidate in sorted(manifest.get("tasks", []), key=lambda item: int(item.get("page_index") or 0)):
        if normalize_content_type(candidate.get("content_type", "")) != "section_divider":
            continue
        count += 1
        if str(candidate.get("task_id")) == str(task.get("task_id")) or int(candidate.get("page_index") or 0) == current_page:
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
        body = clean_text(value.get("description") or value.get("text") or value.get("value"))
        if title and body and title != body:
            return f"{title}: {body}"
        return title or body
    if isinstance(value, list):
        return " / ".join(clean_text(x) for x in value if clean_text(x))
    return str(value).strip()


def extract_source(plan: dict[str, Any]) -> str:
    source = plan.get("source") or plan.get("sources") or ""
    if isinstance(source, list):
        return "；".join(clean_text(x) for x in source if clean_text(x))
    return clean_text(source)


def chart_data(plan: dict[str, Any], items: list[str]) -> dict[str, Any]:
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
