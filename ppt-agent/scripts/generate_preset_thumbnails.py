#!/usr/bin/env python3
"""Generate real local preview images for every full-deck preset."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import tempfile
from pathlib import Path

try:
    from PIL import Image, ImageOps
except ImportError as exc:
    raise SystemExit("Pillow is required: run this script with the PPT generator Python environment") from exc


PROJECT_ROOT = Path(__file__).resolve().parents[1]
VISUAL_DESIGNER_DIR = PROJECT_ROOT / "skills" / "visual_designer"
PRESETS_DIR = VISUAL_DESIGNER_DIR / "templates" / "full-decks"
OUTPUT_DIR = PROJECT_ROOT / "frontend" / "public" / "templates" / "thumbs"
CONVERTER = PROJECT_ROOT / "backend" / "pkg" / "tools" / "qa" / "pptx_qa_converter.py"

sys.path.insert(0, str(VISUAL_DESIGNER_DIR))

from generators import generate_title_slide, new_presentation, save_presentation  # noqa: E402


CATEGORY_LABELS = {
    "biz": "商业演示",
    "edu": "教育培训",
    "gov": "政务思政",
    "other": "通用场景",
    "tech": "技术分享",
    "work": "工作汇报",
}

PALETTE_BACKGROUNDS = {
    "activity_orange": "editorial_coral",
    "charcoal_light": "editorial_ink",
    "civic_gold": "editorial_gold",
    "education_blue": "editorial_blue",
    "government_red": "editorial_coral",
    "ocean_soft": "editorial_blue",
    "report_green": "editorial_sage",
    "warm_terracotta": "editorial_coral",
}


def load_presets() -> list[dict]:
    presets = []
    for path in sorted(PRESETS_DIR.glob("*.json")):
        with path.open("r", encoding="utf-8") as file:
            preset = json.load(file)
        preset["_path"] = path
        presets.append(preset)
    return presets


def preview_title(preset: dict) -> str:
    slides = preset.get("default_slides") or []
    if slides:
        first_title = str(slides[0].get("title", "")).strip()
        if first_title and first_title not in {"封面", "封面页"}:
            return first_title
    return preset["display_name"]


def build_preview_pptx(preset: dict, output_path: Path) -> None:
    palette = preset.get("default_palette") or "ocean_soft"
    title = preview_title(preset)
    display_name = preset["display_name"]
    category = CATEGORY_LABELS.get(preset.get("category", ""), "演示模板")
    slide_count = preset.get("slide_count") or len(preset.get("default_slides") or [])

    presentation = new_presentation(palette=palette)
    generate_title_slide(
        prs=presentation,
        palette=palette,
        title=title,
        subtitle=preset.get("description", ""),
        author="PPT Agent 模板库",
        date=f"{slide_count} 页结构",
        kicker=f"{category} · {display_name}",
        background=PALETTE_BACKGROUNDS.get(palette, "editorial_blue"),
    )
    save_presentation(presentation, str(output_path))


def target_path(preset: dict) -> Path:
    thumbnail = str(preset.get("thumbnail", ""))
    prefix = "/templates/thumbs/"
    if not thumbnail.startswith(prefix):
        raise ValueError(f"{preset['_path'].name}: invalid thumbnail URL {thumbnail!r}")
    return OUTPUT_DIR / thumbnail.removeprefix(prefix)


def convert_previews(pptx_dir: Path, image_dir: Path, filenames: list[str]) -> None:
    command = [
        sys.executable,
        str(CONVERTER),
        "--pptx-dir",
        str(pptx_dir),
        "--output-dir",
        str(image_dir),
        "--dpi",
        "110",
        "--files",
        *filenames,
    ]
    result = subprocess.run(command, check=False, capture_output=True, text=True, encoding="utf-8")
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or result.stdout.strip() or "thumbnail conversion failed")
    try:
        report = json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise RuntimeError(result.stdout.strip() or "thumbnail converter returned invalid output") from exc
    if report.get("error"):
        raise RuntimeError(str(report["error"]))
    expected = [image_dir / f"{Path(filename).stem}.jpg" for filename in filenames]
    missing = [path.name for path in expected if not path.exists()]
    if missing:
        raise RuntimeError(
            f"converter missed {', '.join(missing)}; report={json.dumps(report, ensure_ascii=False)}"
        )


def optimize_preview(source: Path, target: Path) -> None:
    if not source.exists():
        raise FileNotFoundError(f"converter did not produce {source.name}")
    target.parent.mkdir(parents=True, exist_ok=True)
    with Image.open(source) as image:
        image = ImageOps.fit(image.convert("RGB"), (800, 450), method=Image.Resampling.LANCZOS)
        image.save(target, format="JPEG", quality=84, optimize=True, progressive=True)


def validate_outputs(presets: list[dict]) -> None:
    expected = {target_path(preset).resolve() for preset in presets}
    missing = [path for path in sorted(expected) if not path.is_file()]
    if missing:
        raise FileNotFoundError("missing preset previews: " + ", ".join(path.name for path in missing))

    unexpected = [path for path in OUTPUT_DIR.glob("*") if path.is_file() and path.resolve() not in expected]
    if unexpected:
        raise ValueError("unreferenced preset previews: " + ", ".join(path.name for path in unexpected))


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="only validate that every preset URL resolves")
    args = parser.parse_args()
    presets = load_presets()

    if not args.check:
        with tempfile.TemporaryDirectory(prefix="ppt-preset-thumbs-") as temp_dir:
            temp_root = Path(temp_dir)
            pptx_dir = temp_root / "pptx"
            image_dir = temp_root / "images"
            pptx_dir.mkdir()
            image_dir.mkdir()

            filenames = []
            for preset in presets:
                filename = f"{preset['name']}.pptx"
                build_preview_pptx(preset, pptx_dir / filename)
                filenames.append(filename)

            convert_previews(pptx_dir, image_dir, filenames)
            for preset in presets:
                optimize_preview(image_dir / f"{preset['name']}.jpg", target_path(preset))

    validate_outputs(presets)
    print(f"Validated {len(presets)} preset previews in {OUTPUT_DIR}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
