#!/usr/bin/env python3
"""Validate bundled visual assets and smoke-test icon fallback portability."""

from __future__ import annotations

import argparse
import json
import os
import sys
import tempfile
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[1]
SKILL_ROOT = PROJECT_ROOT / "skills" / "visual_designer"
ASSET_ROOT = SKILL_ROOT / "assets"
MANIFEST = ASSET_ROOT / "manifest.json"


def image_errors(asset_id: str, path: Path) -> list[str]:
    try:
        from PIL import Image
    except ImportError:
        return ["Pillow is required to validate runtime image readability"]
    try:
        with Image.open(path) as image:
            image.verify()
        with Image.open(path).convert("RGBA") as image:
            if image.width < 8 or image.height < 8:
                return [f"image_too_small:{asset_id}:{image.size}"]
            if image.getchannel("A").getbbox() is None:
                return [f"image_fully_transparent:{asset_id}"]
    except Exception as error:  # noqa: BLE001 - report every unreadable asset.
        return [f"image_unreadable:{asset_id}:{error}"]
    return []


def validate_bundle() -> list[str]:
    if not MANIFEST.is_file():
        return [f"manifest_missing:{MANIFEST}"]
    data = json.loads(MANIFEST.read_text(encoding="utf-8"))
    assets = data.get("assets", [])
    errors: list[str] = []
    seen: set[str] = set()
    for asset in assets:
        asset_id = str(asset.get("id", "")).strip()
        relative = str(asset.get("path", "")).strip()
        if not asset_id or asset_id in seen:
            errors.append(f"duplicate_or_empty_id:{asset_id or '<empty>'}")
            continue
        seen.add(asset_id)
        target = (ASSET_ROOT / relative).resolve()
        try:
            target.relative_to(ASSET_ROOT.resolve())
        except ValueError:
            errors.append(f"path_outside_bundle:{asset_id}:{relative}")
            continue
        if not target.is_file():
            errors.append(f"missing:{asset_id}:{relative}")
            continue
        cursor = ASSET_ROOT
        for part in Path(relative).parts:
            if part not in {child.name for child in cursor.iterdir()}:
                errors.append(f"case_mismatch:{asset_id}:{relative}")
                break
            cursor /= part
        errors.extend(image_errors(asset_id, target))
    return errors


def smoke_icon_grid(output: Path) -> list[str]:
    sys.path.insert(0, str(PROJECT_ROOT))
    old_cwd = Path.cwd()
    try:
        with tempfile.TemporaryDirectory(prefix="ppt-assets-cwd-") as temp_cwd:
            try:
                os.chdir(temp_cwd)
                from skills.visual_designer.generators.asset_manager import validate_manifest
                from skills.visual_designer.generators.icon_grid_generator import generate

                errors = validate_manifest()
                if errors:
                    return errors
                presentation = generate(
                    title="视觉资源部署检查",
                    subtitle="本页包含已打包图标与语义 fallback",
                    icons=[
                        {"icon": "runtime", "label": "运行状态"},
                        {"icon": "timeline", "label": "执行轨迹"},
                        {"icon": "missing-semantic-icon", "label": "缺失图标降级"},
                        {"icon": "chart", "label": "图表"},
                        {"icon": "template", "label": "模板"},
                        {"icon": "file", "label": "文件"},
                    ],
                )
                output.parent.mkdir(parents=True, exist_ok=True)
                presentation.save(output)
                slide_text = "\n".join(
                    shape.text for shape in presentation.slides[0].shapes if hasattr(shape, "text_frame")
                )
                if "MI" not in slide_text:
                    return ["fallback_text_missing:missing-semantic-icon"]
            finally:
                os.chdir(old_cwd)
    except Exception as error:  # noqa: BLE001 - smoke failure must be visible.
        return [f"icon_grid_smoke_failed:{error}"]
    finally:
        os.chdir(old_cwd)
    return []


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--smoke-output", type=Path, help="Optional path to retain the smoke PPTX")
    args = parser.parse_args()
    errors = validate_bundle()
    if args.smoke_output:
        output = args.smoke_output.resolve()
        errors.extend(smoke_icon_grid(output))
    else:
        with tempfile.TemporaryDirectory(prefix="ppt-assets-smoke-") as temp_dir:
            errors.extend(smoke_icon_grid(Path(temp_dir) / "icon-grid-smoke.pptx"))
    if errors:
        print("Visual asset validation failed:")
        for error in errors:
            print(f"- {error}")
        return 1
    print("Visual asset validation passed: manifest, images, portable import, and icon fallback are valid.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
