"""Strict local-image helpers for ppt-deck-planner generators."""

from __future__ import annotations

from pathlib import Path

from PIL import Image
from pptx.util import Inches


def resolve_photo(image_path: str = "") -> str | None:
    """Resolve an explicit local image path or fail loudly.

    Empty input means the caller did not provide an image. Any non-empty value
    must be a real local file path; legacy ``asset:`` ids are unsupported in the
    standalone skill.
    """
    value = (image_path or "").strip()
    if not value:
        return None
    if value.startswith("asset:"):
        raise ValueError(f"legacy asset id is unsupported: {value}")
    candidate = Path(value).expanduser()
    if not candidate.is_file():
        raise FileNotFoundError(f"image path does not exist: {value}")
    return str(candidate.resolve())


def add_cropped_photo(
    slide,
    image_path: str,
    left: float,
    top: float,
    width: float,
    height: float,
):
    """Add a center-cropped PowerPoint picture to a fixed frame."""
    path_text = resolve_photo(image_path)
    if width <= 0 or height <= 0:
        raise ValueError(f"invalid image frame size: {width}x{height}")

    path = Path(path_text)
    with Image.open(path) as image:
        image_ratio = image.width / image.height
    frame_ratio = width / height

    picture = slide.shapes.add_picture(
        str(path),
        Inches(left),
        Inches(top),
        width=Inches(width),
        height=Inches(height),
    )
    if image_ratio > frame_ratio:
        visible = frame_ratio / image_ratio
        picture.crop_left = picture.crop_right = max(0.0, (1.0 - visible) / 2.0)
    elif image_ratio < frame_ratio:
        visible = image_ratio / frame_ratio
        picture.crop_top = picture.crop_bottom = max(0.0, (1.0 - visible) / 2.0)
    picture.name = f"Replaceable photo - {path.stem}"
    return picture
