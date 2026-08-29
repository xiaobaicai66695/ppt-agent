#!/usr/bin/env python3
"""Validate visual asset planning and materialization before deck rendering."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


PATH_FIELDS = ("local_path", "image_path", "asset_path", "path")
QUERY_FIELDS = ("asset_query", "asset_subject", "image_url", "preview_url")
IMAGE_REQUIRED_CONTENT_TYPES = {"image_text", "image_hero"}
VALID_MODES = {"required", "optional", "none"}
MATERIALIZED_STATUSES = {"downloaded", "generated"}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--work-dir", required=True)
    parser.add_argument("--skills-dir", required=True)
    parser.add_argument("--manifest", default="tasks.json")
    parser.add_argument("--min-image-pages", type=int, default=None)
    parser.add_argument("--json", action="store_true", dest="as_json")
    args = parser.parse_args()

    work_dir = Path(args.work_dir).resolve()
    manifest_path = (work_dir / args.manifest).resolve()
    result_value = validate_visual_manifest_file(manifest_path, work_dir, args.min_image_pages)

    if args.as_json:
        print(json.dumps(result_value, ensure_ascii=False, indent=2))
    else:
        print_human(result_value)
    return 0 if result_value["ok"] else 1


def validate_visual_manifest_file(manifest_path: Path, work_dir: Path, min_image_pages: int | None = None) -> dict[str, Any]:
    errors: list[dict[str, Any]] = []
    warnings: list[dict[str, Any]] = []

    manifest = read_json(manifest_path, errors, "manifest")
    if not isinstance(manifest, dict):
        return result(errors, warnings, image_page_count=0)

    policy = manifest.get("visual_policy")
    if not isinstance(policy, dict):
        add_error(errors, "manifest.visual_policy", "missing_visual_policy", "deck must declare visual_policy.mode")
        mode = "required"
        policy = {}
    else:
        mode = clean(policy.get("mode")) or "required"

    if mode not in VALID_MODES:
        add_error(errors, "manifest.visual_policy.mode", "invalid_visual_policy", f"visual_policy.mode must be one of {sorted(VALID_MODES)}")
        mode = "required"

    if mode == "required":
        if not isinstance(policy.get("min_image_pages"), int) or policy.get("min_image_pages") < 1:
            add_error(errors, "manifest.visual_policy.min_image_pages", "missing_min_image_pages", "required visual_policy requires min_image_pages >= 1")
        required_roles = policy.get("required_roles")
        if not isinstance(required_roles, list) or not required_roles:
            add_error(errors, "manifest.visual_policy.required_roles", "missing_required_roles", "required visual_policy requires non-empty required_roles")

    tasks = manifest.get("tasks")
    if not isinstance(tasks, list) or not tasks:
        add_error(errors, "manifest.tasks", "missing_tasks", "manifest must contain a non-empty tasks array")
        return result(errors, warnings, image_page_count=0)

    image_page_count = 0
    planned_but_unmaterialized = 0

    for index, task in enumerate(tasks):
        if not isinstance(task, dict):
            continue
        task_path = f"tasks[{index}]"
        content_type = clean(task.get("content_type")) or "content_slide"
        visual_items = collect_visual_items(task, task_path)
        task_has_materialized_image = False

        for item_path, item in visual_items:
            if not isinstance(item, dict):
                continue
            status = clean(item.get("search_status"))
            path_value = first_path_value(item)
            has_query = has_query_hint(item)
            component_type = clean(item.get("type"))

            if status and status not in {"planned", "downloaded", "generated", "skipped"}:
                add_error(errors, item_path + ".search_status", "invalid_search_status", "search_status must be planned, downloaded, generated or skipped")

            if path_value:
                validate_local_path(path_value, item_path, work_dir, errors)
                task_has_materialized_image = True
                if not has_attribution(item):
                    add_warning(warnings, item_path, "missing_attribution", "materialized image should include source_url, attribution or provider")
                continue

            if status == "skipped":
                if not clean(item.get("skip_reason")):
                    add_error(errors, item_path + ".skip_reason", "missing_skip_reason", "skipped visual plan requires skip_reason")
                continue

            if component_type == "image" or has_query or status == "planned":
                planned_but_unmaterialized += 1
                add_error(errors, item_path, "unmaterialized_visual_asset", "planned visual asset must be materialized to a local path or marked skipped with skip_reason before rendering")

        if task_has_materialized_image:
            image_page_count += 1

        if content_type in IMAGE_REQUIRED_CONTENT_TYPES and not task_has_materialized_image:
            add_error(errors, task_path, "missing_required_slide_image", f"{content_type} requires a materialized local image")

    required_min = min_image_pages
    if required_min is None and mode == "required":
        required_min = policy.get("min_image_pages") if isinstance(policy.get("min_image_pages"), int) else 1

    if mode == "required" and isinstance(required_min, int) and image_page_count < required_min:
        add_error(
            errors,
            "manifest.visual_policy.min_image_pages",
            "too_few_image_pages",
            f"visual_policy requires at least {required_min} image pages, got {image_page_count}",
        )
    elif mode == "optional" and image_page_count == 0 and planned_but_unmaterialized == 0:
        add_warning(warnings, "manifest.visual_policy", "no_optional_images", "optional visual_policy has no materialized or planned image pages")

    return result(errors, warnings, image_page_count=image_page_count)


def collect_visual_items(task: dict[str, Any], task_path: str) -> list[tuple[str, dict[str, Any]]]:
    items: list[tuple[str, dict[str, Any]]] = []
    plan = task.get("content_plan") if isinstance(task.get("content_plan"), dict) else {}
    visual_intent = plan.get("visual_intent")
    if isinstance(visual_intent, dict):
        items.append((task_path + ".content_plan.visual_intent", visual_intent))
    components = plan.get("components") if isinstance(plan.get("components"), list) else []
    for index, component in enumerate(components):
        if isinstance(component, dict) and is_visual_component(component):
            items.append((f"{task_path}.content_plan.components[{index}]", component))
    return items


def is_visual_component(component: dict[str, Any]) -> bool:
    return clean(component.get("type")) == "image" or first_path_value(component) or has_query_hint(component) or clean(component.get("search_status"))


def first_path_value(item: dict[str, Any]) -> str:
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    for key in PATH_FIELDS:
        value = clean(item.get(key) or data.get(key))
        if value:
            return value
    return ""


def has_query_hint(item: dict[str, Any]) -> bool:
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    return any(clean(item.get(key) or data.get(key)) for key in QUERY_FIELDS)


def has_attribution(item: dict[str, Any]) -> bool:
    data = item.get("data") if isinstance(item.get("data"), dict) else {}
    return any(clean(item.get(key) or data.get(key)) for key in ("source_url", "attribution", "provider"))


def validate_local_path(value: str, path: str, work_dir: Path, errors: list[dict[str, Any]]) -> None:
    if value.startswith("asset:"):
        add_error(errors, path, "legacy_asset_id", f"legacy asset id is unsupported: {value}")
        return
    candidate = Path(value).expanduser()
    if not candidate.is_absolute():
        candidate = work_dir / candidate
    if not candidate.is_file():
        add_error(errors, path, "missing_local_image", f"local image path does not exist: {value}")


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


def add_warning(warnings: list[dict[str, Any]], path: str, code: str, message: str) -> None:
    warnings.append({"path": path, "code": code, "message": message})


def result(errors: list[dict[str, Any]], warnings: list[dict[str, Any]], image_page_count: int) -> dict[str, Any]:
    return {
        "ok": not errors,
        "image_page_count": image_page_count,
        "error_count": len(errors),
        "warning_count": len(warnings),
        "errors": errors,
        "warnings": warnings,
    }


def print_human(result_value: dict[str, Any]) -> None:
    if result_value["ok"]:
        suffix = f" ({result_value['warning_count']} warnings)" if result_value["warning_count"] else ""
        print(f"Visual asset validation passed: {result_value['image_page_count']} image pages{suffix}")
    else:
        print(f"Visual asset validation failed: {result_value['error_count']} errors")
    for warning in result_value["warnings"]:
        print(f"- [warning:{warning['code']}] {warning['path']}: {warning['message']}")
    for error in result_value["errors"]:
        print(f"- [{error['code']}] {error['path']}: {error['message']}")


def clean(value: Any) -> str:
    return str(value or "").strip()


if __name__ == "__main__":
    raise SystemExit(main())
