"""Generator for image_text - 图文混排页."""
from typing import Optional, List

from pptx import Presentation

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_ellipse,
    set_slide_background,
)


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

    img_y = 1.3
    img_h = 5.0

    # Image placeholder box
    add_rect(
        slide,
        left=img_x, top=img_y, width=img_w, height=img_h,
        fill_color="light_bg", palette=palette,
    )
    # Image border accent
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

    # Image placeholder label
    add_text(
        slide,
        text="[图片占位]",
        left=img_x, top=img_y + img_h / 2 - 0.3, width=img_w, height=0.6,
        font_size=14, bold=False,
        color="text_muted", alignment="center",
        palette=palette,
    )

    # Section header
    add_text(
        slide,
        text=header,
        left=text_x, top=1.5, width=text_w, height=0.6,
        font_size=20, bold=True,
        color="primary", alignment="left",
        palette=palette,
    )

    # Sub header (between feature header and bullets)
    y_content_start = 2.5
    if sub_header:
        add_text(
            slide,
            text=sub_header,
            left=text_x, top=2.15, width=text_w, height=0.4,
            font_size=13, bold=False,
            color="secondary", alignment="left",
            palette=palette,
        )
        y_content_start = 2.6
        # Accent line under sub_header
        add_rect(
            slide,
            left=text_x, top=2.55, width=1.0, height=0.04,
            fill_color="divider", palette=palette,
        )
    else:
        # Accent line under header
        add_rect(
            slide,
            left=text_x, top=2.15, width=1.0, height=0.05,
            fill_color="primary", palette=palette,
        )

    # Content: paragraph or bullets
    if paragraph:
        # Render paragraph with natural line wrapping
        add_text(
            slide,
            text=paragraph,
            left=text_x, top=y_content_start, width=text_w, height=4.0,
            font_size=14, bold=False,
            color="text", alignment="left",
            palette=palette,
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
            )

    add_source_line(slide, source, palette)
    return prs
