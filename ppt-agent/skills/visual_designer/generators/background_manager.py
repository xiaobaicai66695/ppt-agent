"""Background theme catalog and local image resolver."""

from __future__ import annotations

from functools import lru_cache
import json
from pathlib import Path
import random
from typing import Optional

from PIL import Image


BACKGROUND_ROOT = Path(__file__).resolve().parents[1] / "background_templates"
BACKGROUND_MANIFEST_PATH = BACKGROUND_ROOT / "manifest.json"


@lru_cache(maxsize=1)
def load_background_manifest() -> dict:
    if not BACKGROUND_MANIFEST_PATH.is_file():
        return {"version": 0, "themes": []}
    data = json.loads(BACKGROUND_MANIFEST_PATH.read_text(encoding="utf-8"))
    return data if isinstance(data, dict) else {"version": 0, "themes": []}


def _manifest_theme_mapping() -> dict[str, dict]:
    result: dict[str, dict] = {}
    for theme in load_background_manifest().get("themes", []):
        theme_id = str(theme.get("id", "")).strip()
        if not theme_id:
            continue
        result[theme_id] = {
            "name_cn": theme.get("name_cn", theme_id),
            "scenarios": theme.get("scenarios", []),
            "priority": int(theme.get("priority", 0)),
        }
    return result


def _manifest_palette_mapping() -> dict[str, str]:
    return {
        str(theme.get("id")): str(theme.get("recommended_palette"))
        for theme in load_background_manifest().get("themes", [])
        if theme.get("id") and theme.get("recommended_palette")
    }


# Kept as exported compatibility constants; their source of truth is manifest.json.
THEME_MAPPING = _manifest_theme_mapping()
BACKGROUND_PALETTE_MAP: dict[str, str] = _manifest_palette_mapping()


def get_background_dir() -> Path:
    return BACKGROUND_ROOT


def scan_backgrounds() -> list[dict]:
    """Return manifest-backed themes, falling back to legacy directory scan."""
    themes = load_background_manifest().get("themes", [])
    if themes:
        results: list[dict] = []
        for theme in themes:
            images: list[str] = []
            image_records: list[dict] = []
            for record in theme.get("images", []):
                relative = str(record.get("path", "")).replace("\\", "/").lstrip("/")
                target = _safe_background_path(relative)
                if target and target.is_file():
                    images.append(str(target))
                    image_records.append(record)
            if images:
                results.append({
                    "theme": str(theme.get("id", "")),
                    "name_cn": str(theme.get("name_cn", theme.get("id", ""))),
                    "scenarios": list(theme.get("scenarios", [])),
                    "priority": int(theme.get("priority", 0)),
                    "recommended_palette": str(theme.get("recommended_palette", "ocean_soft")),
                    "images": images,
                    "image_records": image_records,
                })
        return sorted(results, key=lambda item: -item["priority"])
    return _scan_legacy_backgrounds()


def _scan_legacy_backgrounds() -> list[dict]:
    results: list[dict] = []
    if not BACKGROUND_ROOT.exists():
        return results
    for theme_dir in BACKGROUND_ROOT.iterdir():
        if not theme_dir.is_dir():
            continue
        images_dir = theme_dir / "images"
        images = [str(path) for path in sorted(images_dir.glob("*.jpg"))] if images_dir.exists() else []
        root_bg = theme_dir / "background.jpg"
        if root_bg.is_file():
            images.append(str(root_bg))
        if not images:
            continue
        mapping = THEME_MAPPING.get(theme_dir.name, {})
        results.append({
            "theme": theme_dir.name,
            "name_cn": mapping.get("name_cn", theme_dir.name),
            "scenarios": mapping.get("scenarios", []),
            "priority": mapping.get("priority", 0),
            "recommended_palette": BACKGROUND_PALETTE_MAP.get(theme_dir.name, "ocean_soft"),
            "images": images,
            "image_records": [],
        })
    return sorted(results, key=lambda item: -item["priority"])


def _safe_background_path(relative: str) -> Path | None:
    if not relative or ".." in Path(relative).parts:
        return None
    target = (BACKGROUND_ROOT / relative).resolve()
    try:
        target.relative_to(BACKGROUND_ROOT.resolve())
    except ValueError:
        return None
    return target


def validate_background_manifest(min_images_per_theme: int = 4) -> list[str]:
    errors: list[str] = []
    data = load_background_manifest()
    if data.get("version") != 1:
        errors.append(f"unsupported_version:{data.get('version', '<missing>')}")
    seen: set[str] = set()
    for theme in data.get("themes", []):
        theme_id = str(theme.get("id", "")).strip()
        if not theme_id or theme_id in seen:
            errors.append(f"duplicate_or_empty_theme:{theme_id or '<empty>'}")
            continue
        seen.add(theme_id)
        images = theme.get("images", [])
        if len(images) < min_images_per_theme:
            errors.append(f"insufficient_images:{theme_id}:{len(images)}")
        if not theme.get("recommended_palette"):
            errors.append(f"missing_palette:{theme_id}")
        for record in images:
            relative = str(record.get("path", ""))
            filename = Path(relative).name
            if not Path(filename).stem.isdigit():
                errors.append(f"non_numeric_filename:{theme_id}:{relative}")
            target = _safe_background_path(relative)
            if not target or not target.is_file():
                errors.append(f"missing:{theme_id}:{relative or '<empty>'}")
                continue
            expected = record.get("dimensions")
            if not (isinstance(expected, list) and len(expected) == 2):
                errors.append(f"invalid_dimensions:{theme_id}:{relative}")
                continue
            try:
                with Image.open(target) as image:
                    if [image.width, image.height] != expected:
                        errors.append(f"dimension_mismatch:{theme_id}:{relative}")
                    ratio = image.width / max(image.height, 1)
                    if image.width < 1280 or image.height < 720 or abs(ratio - (16 / 9)) > 0.02:
                        errors.append(f"not_16_9_compatible:{theme_id}:{relative}")
            except Exception as exc:
                errors.append(f"unreadable_image:{theme_id}:{relative}:{type(exc).__name__}")
            source_id = str(record.get("source_id", ""))
            if source_id and not source_id.startswith("project-"):
                for field in ("source_url", "download_url", "license", "attribution"):
                    if not record.get(field):
                        errors.append(f"missing_external_metadata:{theme_id}:{relative}:{field}")
    return errors


def get_background(
    theme: Optional[str] = None,
    scenario: Optional[str] = None,
    random_select: bool = False,
) -> Optional[str]:
    backgrounds = scan_backgrounds()
    if not backgrounds:
        return None

    candidates = backgrounds
    if theme:
        candidates = [item for item in backgrounds if item["theme"] == theme]
        if not candidates:
            candidates = [
                item
                for item in backgrounds
                if theme in item["theme"] or theme in item["name_cn"]
            ]
    if not candidates and scenario:
        scenario_lower = scenario.lower()
        candidates = [
            item
            for item in backgrounds
            if scenario_lower in item["name_cn"].lower()
            or any(scenario_lower in value.lower() for value in item["scenarios"])
        ]
    if not candidates:
        candidates = backgrounds

    selected = random.choice(candidates) if random_select else candidates[0]
    images = selected.get("images", [])
    if not images:
        return None
    return random.choice(images) if random_select else images[0]


def list_themes() -> list[dict]:
    return scan_backgrounds()


def get_palette_for_background(background_theme: str) -> str:
    return BACKGROUND_PALETTE_MAP.get(background_theme, "ocean_soft")
