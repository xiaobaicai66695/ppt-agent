#!/usr/bin/env python3
"""Offline PPT Agent quality evaluator.

The evaluator checks observable outcomes in a generated task work directory.
It intentionally avoids LLM calls so it can run in CI and local regression
loops. A future judge step can consume the JSON output from this script.
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


DONE_STATUSES = {"done", "qa_done", "fixed"}


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def score_case(case: dict[str, Any], work_dir: Path) -> dict[str, Any]:
    expected = case.get("expected", {})
    rubric = case.get("rubric", {})
    manifest_path = work_dir / "tasks.json"
    issues: list[str] = []
    points = 0
    max_points = sum(int(v) for v in rubric.values())

    if not manifest_path.exists():
        return {
            "case_id": case.get("id"),
            "pass": False,
            "score": 0,
            "max_score": max_points,
            "issues": [f"missing manifest: {manifest_path}"],
        }

    manifest = load_json(manifest_path)
    tasks = manifest.get("tasks") or []
    required_fields = expected.get("required_manifest_fields") or []

    missing_fields = []
    for idx, task in enumerate(tasks):
        for field in required_fields:
            if task.get(field) in (None, ""):
                missing_fields.append(f"slide[{idx}].{field}")
    if missing_fields:
        issues.append("missing manifest fields: " + ", ".join(missing_fields[:12]))
    else:
        points += int(rubric.get("manifest_consistency", 0))

    slide_count = len(tasks)
    min_slides = int(expected.get("min_slides", 0))
    max_slides = int(expected.get("max_slides", 10**9))
    if min_slides <= slide_count <= max_slides:
        points += int(rubric.get("page_count_fit", 0))
    else:
        issues.append(f"slide count {slide_count} outside [{min_slides}, {max_slides}]")

    content_types = {task.get("content_type") for task in tasks}
    missing_types = [t for t in expected.get("required_content_types", []) if t not in content_types]
    if missing_types:
        issues.append("missing required content types: " + ", ".join(missing_types))
    else:
        points += int(rubric.get("layout_coverage", 0))

    missing_outputs = []
    for task in tasks:
        output = task.get("output_file")
        if output and not (work_dir / output).exists():
            missing_outputs.append(output)
    if missing_outputs:
        issues.append("missing output files: " + ", ".join(missing_outputs[:12]))
    else:
        points += int(rubric.get("file_outputs", 0))

    pending = [task.get("task_id") for task in tasks if task.get("status") not in DONE_STATUSES]
    if pending:
        issues.append("unfinished tasks: " + ", ".join(str(x) for x in pending[:12]))
    else:
        points += int(rubric.get("no_pending_status", 0))

    qa_text = "\n".join(str(task.get("qa_report") or "") for task in tasks).lower()
    veto_hits = [kw for kw in expected.get("qa_veto_keywords", []) if kw.lower() in qa_text]
    if veto_hits:
        issues.append("qa veto keywords found: " + ", ".join(veto_hits))
    else:
        points += int(rubric.get("qa_cleanliness", 0))

    return {
        "case_id": case.get("id"),
        "pass": not issues and points == max_points,
        "score": points,
        "max_score": max_points,
        "issues": issues,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--cases", default="../docs/eval/ppt_quality_cases.json")
    parser.add_argument("--work-dir", required=True)
    parser.add_argument("--case-id", required=True)
    args = parser.parse_args()

    cases_path = Path(args.cases).resolve()
    work_dir = Path(args.work_dir).resolve()
    cases = load_json(cases_path)
    case = next((c for c in cases if c.get("id") == args.case_id), None)
    if case is None:
        raise SystemExit(f"case not found: {args.case_id}")

    result = score_case(case, work_dir)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if result["pass"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
