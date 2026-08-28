#!/usr/bin/env python3
"""Validate a standalone ppt-deck-planner tasks.json before rendering."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


PATH_FIELDS = {"local_path", "asset_path", "image_path", "path"}
IMAGE_HINT_FIELDS = PATH_FIELDS | {"asset_id", "asset_query", "asset_subject", "image_url", "preview_url"}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--work-dir", required=True)
    parser.add_argument("--skills-dir", required=True)
    parser.add_argument("--manifest", default="tasks.json")
    parser.add_argument("--json", action="store_true", dest="as_json")
    args = parser.parse_args()

    work_dir = Path(args.work_dir).resolve()
    skills_dir = Path(args.skills_dir).resolve()
    manifest_path = (work_dir / args.manifest).resolve()
    contract_path = skills_dir / "ppt-deck-planner" / "templates" / "component_contracts.json"

    result = validate_manifest_file(manifest_path, contract_path, work_dir)
    if args.as_json:
        print(json.dumps(result, ensure_ascii=False, indent=2))
    else:
        print_human(result)
    return 0 if result["ok"] else 1


def validate_manifest_file(manifest_path: Path, contract_path: Path, work_dir: Path) -> dict[str, Any]:
    errors: list[dict[str, Any]] = []
    warnings: list[dict[str, Any]] = []

    manifest = read_json(manifest_path, errors, "manifest")
    contract = read_json(contract_path, errors, "contract")
    if not isinstance(manifest, dict) or not isinstance(contract, dict):
        return result(errors, warnings)

    component_types = set((contract.get("component_types") or {}).keys())
    content_types = contract.get("content_types") if isinstance(contract.get("content_types"), dict) else {}
    forbidden = set(
        ((contract.get("planning_rules") or {}).get("forbidden_component_fields") or [])
    )

    tasks = manifest.get("tasks")
    if not isinstance(tasks, list) or not tasks:
        add_error(errors, "manifest.tasks", "missing_tasks", "manifest must contain a non-empty tasks array")
        return result(errors, warnings)

    seen_task_ids: set[str] = set()
    seen_page_indexes: set[int] = set()
    for index, task in enumerate(tasks):
        task_path = f"tasks[{index}]"
        if not isinstance(task, dict):
            add_error(errors, task_path, "invalid_task", "task must be an object")
            continue
        validate_task(task, task_path, content_types, component_types, forbidden, work_dir, errors, warnings)

        task_id = clean(task.get("task_id"))
        if task_id:
            if task_id in seen_task_ids:
                add_error(errors, task_path + ".task_id", "duplicate_task_id", f"duplicate task_id: {task_id}")
            seen_task_ids.add(task_id)

        page_index = task.get("page_index")
        if isinstance(page_index, int):
            if page_index in seen_page_indexes:
                add_error(errors, task_path + ".page_index", "duplicate_page_index", f"duplicate page_index: {page_index}")
            seen_page_indexes.add(page_index)

    find_legacy_asset_ids(manifest, "manifest", errors)
    return result(errors, warnings)


def validate_task(
    task: dict[str, Any],
    task_path: str,
    content_types: dict[str, Any],
    component_types: set[str],
    forbidden: set[str],
    work_dir: Path,
    errors: list[dict[str, Any]],
    warnings: list[dict[str, Any]],
) -> None:
    _ = warnings
    title = clean(task.get("title"))
    if not title:
        add_error(errors, task_path + ".title", "missing_title", "task.title is required")

    content_type = clean(task.get("content_type")) or "content_slide"
    content_contract = content_types.get(content_type)
    if not content_contract:
        add_error(errors, task_path + ".content_type", "invalid_content_type", f"unknown content_type: {content_type}")
    else:
        variants = content_contract.get("variants") if isinstance(content_contract, dict) else []
        layout_variant = clean(task.get("layout_variant"))
        if layout_variant and layout_variant not in set(variants or []):
            add_error(
                errors,
                task_path + ".layout_variant",
                "invalid_layout_variant",
                f"{content_type} does not support layout_variant {layout_variant}",
            )

    page_index = task.get("page_index")
    if not isinstance(page_index, int) or page_index < 1:
        add_error(errors, task_path + ".page_index", "invalid_page_index", "page_index must be a positive integer")

    plan = task.get("content_plan")
    if not isinstance(plan, dict):
        add_error(errors, task_path + ".content_plan", "missing_content_plan", "content_plan object is required")
        return

    components = plan.get("components")
    if not isinstance(components, list):
        add_error(errors, task_path + ".content_plan.components", "missing_components", "components must be an array")
        components = []

    if content_contract:
        capacity = content_contract.get("capacity") if isinstance(content_contract, dict) else {}
        max_components = capacity.get("max_components") if isinstance(capacity, dict) else None
        if isinstance(max_components, int) and len(components) > max_components:
            add_error(
                errors,
                task_path + ".content_plan.components",
                "too_many_components",
                f"{content_type} allows at most {max_components} components, got {len(components)}",
            )

    visual_intent = plan.get("visual_intent")
    if isinstance(visual_intent, dict):
        validate_image_like_item(visual_intent, task_path + ".content_plan.visual_intent", work_dir, errors)

    for component_index, component in enumerate(components):
        component_path = f"{task_path}.content_plan.components[{component_index}]"
        if not isinstance(component, dict):
            add_error(errors, component_path, "invalid_component", "component must be an object")
            continue
        validate_component(component, component_path, component_types, forbidden, work_dir, errors)


def validate_component(
    component: dict[str, Any],
    component_path: str,
    component_types: set[str],
    forbidden: set[str],
    work_dir: Path,
    errors: list[dict[str, Any]],
) -> None:
    component_type = clean(component.get("type"))
    if not component_type:
        add_error(errors, component_path + ".type", "missing_component_type", "component.type is required")
    elif component_type not in component_types:
        add_error(errors, component_path + ".type", "invalid_component_type", f"unknown component type: {component_type}")

    for field in sorted(forbidden.intersection(component.keys())):
        add_error(errors, component_path + f".{field}", "forbidden_visual_field", f"component must not specify visual field {field}")
    data = component.get("data")
    if isinstance(data, dict):
        for field in sorted(forbidden.intersection(data.keys())):
            add_error(errors, component_path + f".data.{field}", "forbidden_visual_field", f"component data must not specify visual field {field}")

    if component_type == "image" or has_image_hints(component):
        validate_image_like_item(component, component_path, work_dir, errors, require_path=component_type == "image")
    if component_type == "kpi_metric":
        validate_kpi_metric(component, component_path, errors)
    if component_type == "chart":
        validate_chart(component, component_path, errors)


def validate_image_like_item(
    item: dict[str, Any],
    path: str,
    work_dir: Path,
    errors: list[dict[str, Any]],
    require_path: bool = False,
) -> None:
    path_value = first_path_value(item)
    if require_path and not path_value:
        add_error(errors, path, "missing_image_path", "image component requires local_path/image_path/asset_path/path")
        return
    if path_value:
        validate_local_path(path_value, path, work_dir, errors)


def validate_kpi_metric(component: dict[str, Any], path: str, errors: list[dict[str, Any]]) -> None:
    data = component.get("data") if isinstance(component.get("data"), dict) else component
    for field in ("value", "label"):
        if not clean(data.get(field)):
            add_error(errors, path + f".{field}", "missing_kpi_field", f"kpi_metric requires {field}")
    if not clean(data.get("delta")) and not clean(data.get("baseline")):
        add_error(errors, path, "missing_kpi_context", "kpi_metric requires delta or baseline")


def validate_chart(component: dict[str, Any], path: str, errors: list[dict[str, Any]]) -> None:
    data = component.get("data") if isinstance(component.get("data"), dict) else {}
    if not clean(data.get("chart_type")):
        add_error(errors, path + ".data.chart_type", "missing_chart_type", "chart.data.chart_type is required")
    labels = data.get("labels")
    if not isinstance(labels, list) or not labels:
        add_error(errors, path + ".data.labels", "missing_chart_labels", "chart.data.labels must be a non-empty array")
    datasets = data.get("datasets")
    if not isinstance(datasets, list) or not datasets:
        add_error(errors, path + ".data.datasets", "missing_chart_datasets", "chart.data.datasets must be a non-empty array")
    else:
        for index, dataset in enumerate(datasets):
            if not isinstance(dataset, dict) or not isinstance(dataset.get("values"), list) or not dataset.get("values"):
                add_error(errors, f"{path}.data.datasets[{index}].values", "missing_chart_values", "each dataset requires non-empty values")


def has_image_hints(item: dict[str, Any]) -> bool:
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    return any(clean(item.get(key) or data.get(key)) for key in IMAGE_HINT_FIELDS)


def first_path_value(item: dict[str, Any]) -> str:
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    for key in PATH_FIELDS:
        value = clean(item.get(key) or data.get(key))
        if value:
            return value
    return ""


def validate_local_path(value: str, path: str, work_dir: Path, errors: list[dict[str, Any]]) -> None:
    if value.startswith("asset:"):
        add_error(errors, path, "legacy_asset_id", f"legacy asset id is unsupported: {value}")
        return
    candidate = Path(value).expanduser()
    if not candidate.is_absolute():
        candidate = work_dir / candidate
    if not candidate.is_file():
        add_error(errors, path, "missing_local_image", f"local image path does not exist: {value}")


def find_legacy_asset_ids(value: Any, path: str, errors: list[dict[str, Any]]) -> None:
    if isinstance(value, str):
        if value.strip().startswith("asset:"):
            add_error(errors, path, "legacy_asset_id", f"legacy asset id is unsupported: {value}")
        return
    if isinstance(value, list):
        for index, item in enumerate(value):
            find_legacy_asset_ids(item, f"{path}[{index}]", errors)
    elif isinstance(value, dict):
        for key, item in value.items():
            if key in PATH_FIELDS:
                continue
            find_legacy_asset_ids(item, f"{path}.{key}", errors)


def read_json(path: Path, errors: list[dict[str, Any]], label: str) -> Any:
    if not path.is_file():
        add_error(errors, str(path), f"missing_{label}", f"{label} file not found: {path}")
        return None
    try:
        return json.loads(path.read_text(encoding="utf-8-sig"))
    except json.JSONDecodeError as exc:
        add_error(errors, str(path), f"invalid_{label}_json", f"invalid JSON: {exc}")
        return None


def add_error(errors: list[dict[str, Any]], path: str, code: str, message: str) -> None:
    errors.append({"path": path, "code": code, "message": message})


def result(errors: list[dict[str, Any]], warnings: list[dict[str, Any]]) -> dict[str, Any]:
    return {"ok": not errors, "error_count": len(errors), "warning_count": len(warnings), "errors": errors, "warnings": warnings}


def print_human(result_value: dict[str, Any]) -> None:
    if result_value["ok"]:
        print("DeckSpec validation passed")
        return
    print(f"DeckSpec validation failed: {result_value['error_count']} errors")
    for error in result_value["errors"]:
        print(f"- [{error['code']}] {error['path']}: {error['message']}")


def clean(value: Any) -> str:
    return str(value or "").strip()


if __name__ == "__main__":
    raise SystemExit(main())
