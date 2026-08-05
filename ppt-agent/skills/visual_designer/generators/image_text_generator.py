"""Generator for image_text - 图文混排页."""

from pathlib import Path
from typing import List, Optional

from pptx import Presentation

from .asset_manager import add_cropped_photo, asset_path, photo_id_from_text, resolve_photo
from .base import (
    PALETTES,
    add_rect,
    add_source_line,
    add_text,
    new_presentation,
    resolve_background,
    set_slide_background,
)
from .layout_intelligence import balanced_band_top, weighted_text_len


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{页面标题}",
    layout: str = "right-image",
    layout_variant: str = "",
    header: str = "{子标题}",
    paragraph: str = "",
    bullets: List[str] = None,
    kicker: str = "",
    sub_header: str = "",
    background: str = None,
    glass_colors: dict = None,
    image_path: str = "",
) -> Presentation:
    """Generate a text explanation with one replaceable picture object.

    ``image_path`` accepts a local file, a registered photo id, or
    ``asset:<photo-id>``. When it is absent or invalid, the generator chooses a
    semantic local photo, then a planned background image, and finally the
    stable general workspace photo.
    """
    bullets = bullets or []
    if prs is None:
        prs = new_presentation(palette=palette)

    layout = _resolve_layout(layout=layout, layout_variant=layout_variant)
    slide = prs.slides.add_slide(prs.slide_layouts[6])
    set_slide_background(slide, palette)
    colors = glass_colors or PALETTES.get(palette, PALETTES["ocean_soft"])

    visual_text = " ".join([title, header, sub_header, paragraph, *bullets])
    content_image = _resolve_content_photo(
        image_path=image_path,
        background=background,
        visual_text=visual_text,
    )
    if not content_image:
        raise FileNotFoundError(
            "image_text requires a valid image_path or the packaged default photo assets"
        )

    _render_title(slide, palette, colors, title, kicker)
    if layout == "photo-strip":
        _render_photo_strip(
            slide=slide,
            palette=palette,
            colors=colors,
            image_path=content_image,
            header=header,
            paragraph=paragraph,
            bullets=bullets,
            sub_header=sub_header,
        )
    else:
        _render_side_photo(
            slide=slide,
            palette=palette,
            colors=colors,
            image_path=content_image,
            layout=layout,
            header=header,
            paragraph=paragraph,
            bullets=bullets,
            sub_header=sub_header,
        )

    add_source_line(slide, source, palette)
    return prs


def _resolve_content_photo(image_path: str, background: str, visual_text: str) -> str | None:
    explicit = resolve_photo(image_path=image_path, text="", fallback="")
    if explicit:
        return explicit

    semantic_id = photo_id_from_text(visual_text, fallback="")
    semantic = asset_path(semantic_id) if semantic_id else None
    if semantic:
        return semantic

    planned_background = resolve_background(background) if background else None
    if planned_background and Path(planned_background).is_file():
        return planned_background

    return resolve_photo(text="", fallback="photo_business_work")


def _render_title(slide, palette: str, colors: dict, title: str, kicker: str) -> None:
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.5,
            top=0.12,
            width=12.0,
            height=0.26,
            font_size=12,
            color="secondary",
            alignment="left",
            palette=palette,
            colors=colors,
        )
    add_text(
        slide,
        text=title,
        left=0.5,
        top=0.4,
        width=12.0,
        height=0.72,
        font_size=36,
        bold=True,
        color="text",
        alignment="left",
        palette=palette,
        colors=colors,
    )


def _render_side_photo(
    slide,
    palette: str,
    colors: dict,
    image_path: str,
    layout: str,
    header: str,
    paragraph: str,
    bullets: List[str],
    sub_header: str,
) -> None:
    if layout == "left-image":
        img_x, img_w = 0.5, 6.55
        text_x, text_w = 7.4, 5.4
    else:
        text_x, text_w = 0.5, 5.85
        img_x, img_w = 6.7, 6.1

    img_y, img_h = 1.35, 5.45
    add_rect(
        slide,
        left=img_x - 0.06,
        top=img_y - 0.06,
        width=img_w + 0.12,
        height=img_h + 0.12,
        fill_color="accent",
        palette=palette,
    )
    add_cropped_photo(slide, image_path, img_x, img_y, img_w, img_h)

    body_weight = weighted_text_len(paragraph) if paragraph else sum(weighted_text_len(item) for item in bullets)
    body_height = 3.55 if body_weight > 300 else 3.05 if body_weight > 180 else 2.45
    content_height = 0.62 + (0.44 if sub_header else 0.16) + body_height
    text_top = balanced_band_top(1.45, 5.1, content_height, min_top=1.35)

    add_text(
        slide,
        text=header,
        left=text_x,
        top=text_top,
        width=text_w,
        height=0.62,
        font_size=22,
        bold=True,
        color="primary",
        alignment="left",
        palette=palette,
        colors=colors,
    )
    add_rect(
        slide,
        left=text_x,
        top=text_top + 0.7,
        width=0.72,
        height=0.05,
        fill_color="primary",
        palette=palette,
    )

    body_top = text_top + 0.92
    if sub_header:
        add_text(
            slide,
            text=sub_header,
            left=text_x,
            top=body_top,
            width=text_w,
            height=0.4,
            font_size=16,
            bold=True,
            color="secondary",
            alignment="left",
            palette=palette,
            colors=colors,
        )
        body_top += 0.52

    _render_body(
        slide=slide,
        palette=palette,
        colors=colors,
        paragraph=paragraph,
        bullets=bullets,
        left=text_x,
        top=body_top,
        width=text_w,
        height=min(body_height, 6.65 - body_top),
    )


def _render_photo_strip(
    slide,
    palette: str,
    colors: dict,
    image_path: str,
    header: str,
    paragraph: str,
    bullets: List[str],
    sub_header: str,
) -> None:
    img_x, img_y, img_w, img_h = 0.5, 1.32, 12.3, 2.72
    add_rect(
        slide,
        left=img_x - 0.05,
        top=img_y - 0.05,
        width=img_w + 0.1,
        height=img_h + 0.1,
        fill_color="accent",
        palette=palette,
    )
    add_cropped_photo(slide, image_path, img_x, img_y, img_w, img_h)

    text_x, text_w = 0.72, 11.85
    add_text(
        slide,
        text=header,
        left=text_x,
        top=4.3,
        width=text_w,
        height=0.52,
        font_size=22,
        bold=True,
        color="primary",
        alignment="left",
        palette=palette,
        colors=colors,
    )
    add_rect(
        slide,
        left=text_x,
        top=4.91,
        width=0.72,
        height=0.05,
        fill_color="primary",
        palette=palette,
    )

    body_top = 5.06
    if sub_header:
        add_text(
            slide,
            text=sub_header,
            left=text_x,
            top=body_top,
            width=text_w,
            height=0.34,
            font_size=16,
            bold=True,
            color="secondary",
            alignment="left",
            palette=palette,
            colors=colors,
        )
        body_top += 0.42

    _render_body(
        slide=slide,
        palette=palette,
        colors=colors,
        paragraph=paragraph,
        bullets=bullets,
        left=text_x,
        top=body_top,
        width=text_w,
        height=max(0.9, 6.72 - body_top),
    )


def _render_body(
    slide,
    palette: str,
    colors: dict,
    paragraph: str,
    bullets: List[str],
    left: float,
    top: float,
    width: float,
    height: float,
) -> None:
    if paragraph:
        length = weighted_text_len(paragraph)
        font_size = 17 if length < 220 else 16 if length < 360 else 15
        add_text(
            slide,
            text=paragraph,
            left=left,
            top=top,
            width=width,
            height=height,
            font_size=font_size,
            color="text",
            alignment="left",
            vertical_alignment="middle" if length < 260 else "top",
            line_spacing=1.0,
            palette=palette,
            colors=colors,
        )
        return

    visible = bullets[:5]
    if not visible:
        return
    row_height = height / len(visible)
    for index, item in enumerate(visible):
        y = top + index * row_height
        add_rect(
            slide,
            left=left,
            top=y + 0.14,
            width=0.11,
            height=0.11,
            fill_color="secondary",
            palette=palette,
        )
        add_text(
            slide,
            text=item,
            left=left + 0.24,
            top=y,
            width=width - 0.24,
            height=max(0.45, row_height - 0.05),
            font_size=16,
            color="text",
            alignment="left",
            vertical_alignment="middle",
            palette=palette,
            colors=colors,
        )


def _resolve_layout(layout: str, layout_variant: str = "") -> str:
    """Normalize variant metadata to a generator layout."""
    key = (layout_variant or layout or "").strip().lower().replace("_", "-")
    mapping = {
        "left-photo": "left-image",
        "left-image": "left-image",
        "right-photo": "right-image",
        "right-image": "right-image",
        "photo-strip": "photo-strip",
        "strip": "photo-strip",
    }
    return mapping.get(key, "right-image")
