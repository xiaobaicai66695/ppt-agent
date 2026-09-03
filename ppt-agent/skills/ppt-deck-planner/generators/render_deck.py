#!/usr/bin/env python3
"""Render every task in tasks.json into one PPTX deck."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any, Callable

import render_task
import validate_deck


DEFAULT_OUTPUT = "deck.pptx"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--work-dir", required=True)
    parser.add_argument("--skills-dir", required=True)
    parser.add_argument("--output", default=DEFAULT_OUTPUT)
    args = parser.parse_args()

    work_dir = Path(args.work_dir).resolve()
    skills_dir = Path(args.skills_dir).resolve()
    deck_planner_dir = skills_dir / "ppt-deck-planner"
    sys.path.insert(0, str(deck_planner_dir))

    manifest_path = work_dir / "tasks.json"
    contract_path = deck_planner_dir / "templates" / "component_contracts.json"
    validation = validate_deck.validate_manifest_file(manifest_path, contract_path, work_dir)
    if not validation["ok"]:
        print(json.dumps(validation, ensure_ascii=False, indent=2), file=sys.stderr)
        return 1

    from generators import (  # pylint: disable=import-error,import-outside-toplevel
        new_presentation,
        save_presentation,
        generate_agenda,
        generate_brand_focus,
        generate_card_grid,
        generate_chart_slide,
        generate_comparison_table,
        generate_content_slide,
        generate_image_hero,
        generate_image_text,
        generate_kanban,
        generate_kpi_dashboard,
        generate_quote_slide,
        generate_region_map,
        generate_section_divider,
        generate_swot_analysis,
        generate_timeline,
        generate_title_slide,
        generate_two_column,
    )

    generators: dict[str, Callable[..., Any]] = {
        "agenda": generate_agenda,
        "brand_focus": generate_brand_focus,
        "card_grid": generate_card_grid,
        "chart_slide": generate_chart_slide,
        "comparison_table": generate_comparison_table,
        "content_slide": generate_content_slide,
        "image_hero": generate_image_hero,
        "image_text": generate_image_text,
        "kanban": generate_kanban,
        "kpi_dashboard": generate_kpi_dashboard,
        "quote_slide": generate_quote_slide,
        "region_map": generate_region_map,
        "section_divider": generate_section_divider,
        "swot_analysis": generate_swot_analysis,
        "timeline": generate_timeline,
        "title_slide": generate_title_slide,
        "two_column": generate_two_column,
    }

    manifest = render_task.read_json(manifest_path)
    tasks = sorted(
        [task for task in manifest.get("tasks", []) if isinstance(task, dict)],
        key=lambda task: render_task.safe_int(task.get("page_index"), 9999),
    )
    if not tasks:
        raise SystemExit("tasks.json contains no renderable tasks")

    palette = render_task.manifest_palette(manifest)
    prs = new_presentation(palette=palette)
    rendered: list[dict[str, Any]] = []
    for index, task in enumerate(tasks, start=1):
        task_id = render_task.clean_text(task.get("task_id")) or f"slide_{index:02d}"
        content_type = render_task.normalize_content_type(task.get("content_type", "content_slide"))
        generator = generators.get(content_type, generate_content_slide)
        params = render_task.build_params(content_type, task, manifest, work_dir=work_dir)
        params.update({
            "prs": prs,
            "palette": palette,
            "background": render_task.background_from_task(task, work_dir),
        })
        generator(**render_task.accepted_params(generator, params))
        rendered.append({"task_id": task_id, "page_index": task.get("page_index"), "content_type": content_type})

    output_path = resolve_output_path(args.output, work_dir)
    save_presentation(prs, str(output_path))
    print(json.dumps({"ok": True, "output_file": str(output_path), "slides": rendered}, ensure_ascii=False))
    return 0


def resolve_output_path(output: str, work_dir: Path) -> Path:
    path = Path(output).expanduser()
    if not path.is_absolute():
        path = work_dir / path
    return path.resolve()


if __name__ == "__main__":
    raise SystemExit(main())
