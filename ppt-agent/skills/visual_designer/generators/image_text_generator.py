"""Generator for image_text - 图文混排页."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_ellipse, add_round_rect,
    set_slide_background,
    resolve_background, set_image_background,
)
from .asset_manager import add_local_icon, add_pattern_overlay, icon_id_from_text
from .layout_intelligence import balanced_band_top, focal_font_size, weighted_text_len


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{页面标题}",
    layout: str = "right-image",
    header: str = "{子标题}",
    paragraph: str = "",
    bullets: List[str] = None,
    kicker: str = "",
    sub_header: str = "",
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate an image-text slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Slide title.
        layout: "left-image" or "right-image".
        header: Section/feature header text.
        paragraph: A single paragraph of text (300-450 chars) for detailed description.
        bullets: List of bullet items (legacy, use paragraph instead).
        kicker: Small label above title (e.g. "功能 · 核心").
        sub_header: Secondary header between title and feature header (e.g. "能力亮点").

    Returns:
        The Presentation object.
    """
    if bullets is None:
        bullets = []

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    bg_path = resolve_background(background) if background else None
    if bg_path:
        colors = set_image_background(slide, bg_path, brightness=0.95, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Kicker (above title)
    y_title = 0.4
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.5, top=0.1, width=12.0, height=0.3,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_title = 0.4

    # Title
    add_text(
        slide,
        text=title,
        left=0.5, top=y_title, width=12.0, height=0.7,
        font_size=36, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )

    if layout == "left-image":
        img_x, img_w = 0.5, 7.5
        text_x = 8.3
        text_w = 4.5
    else:
        # right-image (default)
        img_x, img_w = 7.0, 5.8
        text_x = 0.5
        text_w = 6.2

    img_y = 1.35
    img_h = 4.95

    # Visual summary panel: local assets make the default state feel finished.
    add_rect(
        slide,
        left=img_x, top=img_y, width=img_w, height=img_h,
        fill_color="light_bg", palette=palette,
    )
    add_pattern_overlay(slide, "pattern_grid", left=img_x + 0.1, top=img_y + 0.1, width=img_w - 0.2, height=img_h - 0.2, opacity_backdrop=False, palette=palette)
    add_rect(
        slide,
        left=img_x, top=img_y, width=img_w, height=0.06,
        fill_color="accent", palette=palette,
    )
    add_rect(
        slide,
        left=img_x, top=img_y, width=0.06, height=img_h,
        fill_color="accent", palette=palette,
    )

    visual_text = " ".join([title, header, sub_header, paragraph] + bullets)
    icon_id = icon_id_from_text(visual_text, fallback="layout")
    add_local_icon(
        slide,
        icon_id,
        left=img_x + img_w * 0.5 - 0.62,
        top=img_y + 0.78,
        size=1.24,
        palette=palette,
        with_badge=True,
    )
    add_round_rect(
        slide,
        left=img_x + img_w * 0.2,
        top=img_y + 2.35,
        width=img_w * 0.6,
        height=0.78,
        fill_color="background",
        palette=palette,
        line_color="divider",
        line_width=0.6,
    )
    add_text(
        slide,
        text=header,
        left=img_x + img_w * 0.23,
        top=img_y + 2.48,
        width=img_w * 0.54,
        height=0.38,
        font_size=focal_font_size(header, base=22, max_size=30, min_size=16),
        bold=True,
        color="primary",
        alignment="center",
        palette=palette,
        colors=colors,
    )
    visual_notes = [
        sub_header or "结构化内容",
        "动态字号" if weighted_text_len(paragraph) < 500 else "容量控制",
        "本地素材",
    ]
    note_w = (img_w - 1.0) / 3
    for i, note in enumerate(visual_notes):
        x = img_x + 0.5 + i * note_w
        add_rect(slide, left=x, top=img_y + 3.62, width=note_w - 0.18, height=0.5, fill_color="background", palette=palette)
        add_text(
            slide,
            text=note,
            left=x + 0.08,
            top=img_y + 3.71,
            width=note_w - 0.34,
            height=0.24,
            font_size=11,
            bold=True,
            color="secondary",
            alignment="center",
            palette=palette,
            colors=colors,
        )

    text_region_top = 1.42
    text_region_h = 4.9
    paragraph_h = 3.55 if paragraph else min(4.2, max(1.2, len(bullets[:6]) * 0.68))
    content_h = 0.55 + (0.42 if sub_header else 0.15) + paragraph_h
    text_top = balanced_band_top(text_region_top, text_region_h, content_h, min_top=1.35)

    # Section header
    add_text(
        slide,
        text=header,
        left=text_x, top=text_top, width=text_w, height=0.58,
        font_size=20, bold=True,
        color="primary", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Sub header (between feature header and bullets)
    y_content_start = text_top + 1.0
    if sub_header:
        add_text(
            slide,
            text=sub_header,
            left=text_x, top=text_top + 0.65, width=text_w, height=0.36,
            font_size=13, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_content_start = text_top + 1.14
        # Accent line under sub_header
        add_rect(
            slide,
            left=text_x, top=text_top + 1.05, width=1.0, height=0.04,
            fill_color="divider", palette=palette,
        )
    else:
        # Accent line under header
        add_rect(
            slide,
            left=text_x, top=text_top + 0.68, width=1.0, height=0.05,
            fill_color="primary", palette=palette,
        )

    # Content: paragraph or bullets
    if paragraph:
        # Render paragraph with natural line wrapping
        add_text(
            slide,
            text=paragraph,
            left=text_x, top=y_content_start, width=text_w, height=4.0,
            font_size=15 if weighted_text_len(paragraph) < 360 else 14, bold=False,
            color="text", alignment="left",
            vertical_alignment="middle" if weighted_text_len(paragraph) < 260 else "top",
            line_spacing=0.95,
            palette=palette,
            colors=colors,
        )
    else:
        # Legacy: bullet list
        for i, item in enumerate(bullets[:6] if bullets else []):
            y = y_content_start + i * 0.85

            # Bullet dot
            add_rect(
                slide,
                left=text_x + 0.05, top=y + 0.12, width=0.12, height=0.12,
                fill_color="secondary", palette=palette,
            )

            add_text(
                slide,
                text=item,
                left=text_x + 0.28, top=y, width=text_w - 0.3, height=0.75,
                font_size=14, bold=False,
                color="text", alignment="left",
                palette=palette,
                colors=colors,
            )

    add_source_line(slide, source, palette)
    return prs
