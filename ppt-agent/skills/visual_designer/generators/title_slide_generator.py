"""Generator for title_slide -开场标题页."""
import os
from typing import Optional

from pptx import Presentation
from pptx.util import Inches

from .base import (
    add_source_line,
    PALETTES, rgb, add_text, add_ellipse, add_rect,
    set_slide_background, add_text_in_shape,
    new_presentation, resolve_background,
    set_image_background,
)
from .asset_manager import apply_asset_background
from .layout_intelligence import balanced_band_top, title_font_size


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    title: str = "{演讲主题}",
    subtitle: str = "{副标题}",
    author: str = "{演讲者}",
    date: str = "{日期}",
    kicker: str = "",
    background: str = None,
    background_brightness: float = 0.25,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a title slide.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        title: Main title text.
        subtitle: Subtitle text.
        author: Author/department name (bottom left).
        date: Date text (bottom right).
        kicker: Small label above title (e.g. "产品发布 · 2025").
        background: 背景配置，支持：
                   - 主题名: "party_government", "minimalist_blue"
                   - 场景描述: "党建汇报", "商务演示"
                   - 图片路径: "D:/path/to/image.jpg"
                   - 默认: None (纯色背景)
        background_brightness: 背景亮度 (0.0-1.0)

    Returns:
        The Presentation object.
    """
    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)

    # 设置背景
    colors = apply_asset_background(slide, background, palette, role="title", brightness=max(background_brightness, 0.92))
    if colors:
        pass
    else:
        bg_path = resolve_background(background) if background else None
        if bg_path:
            colors = set_image_background(slide, bg_path, brightness=background_brightness, palette=palette)
        else:
            set_slide_background(slide, palette)
            colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    title_size = title_font_size(title, base=44, sparse_boost=8, max_size=56)

    # Editorial side panel makes sparse title pages feel intentionally composed.
    add_rect(
        slide,
        left=0, top=0, width=0.22, height=7.5,
        fill_color="primary", palette=palette,
    )
    add_rect(
        slide,
        left=0.45, top=5.95, width=2.8, height=0.12,
        fill_color="accent", palette=palette,
    )

    content_h = (0.42 if kicker else 0.0) + 1.35 + (0.72 if subtitle else 0.0)
    group_top = balanced_band_top(1.45, 3.75, content_h, min_top=1.55)
    cur_y = group_top

    # Kicker (above title)
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=1.0, top=cur_y, width=11.333, height=0.4,
            font_size=14, bold=False,
            color="secondary", alignment="center",
            palette=palette,
            colors=colors,
        )
        cur_y += 0.42

    # Low-key editorial accents; avoid the old generic corner blobs.
    add_ellipse(slide, left=11.25, top=0.35, width=0.95, height=0.95, fill_color="secondary", palette=palette)
    add_ellipse(slide, left=11.75, top=6.25, width=0.7, height=0.7, fill_color="accent", palette=palette)

    # Main title - centered, large
    add_text(
        slide,
        text=title,
        left=1.0, top=cur_y, width=11.333, height=1.35,
        font_size=title_size, bold=True,
        color="text", alignment="center",
        vertical_alignment="middle",
        palette=palette,
        colors=colors,
    )
    cur_y += 1.45

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=1.7, top=cur_y, width=9.933, height=0.65,
            font_size=22 if len(subtitle) <= 26 else 18, bold=False,
            color="secondary", alignment="center",
            vertical_alignment="middle",
            palette=palette,
            colors=colors,
        )

    # Author (bottom left)
    add_text(
        slide,
        text=author,
        left=0.55, top=6.52, width=4.0, height=0.4,
        font_size=14, bold=False,
        color="primary", alignment="left",
        palette=palette,
        colors=colors,
    )

    # Date (bottom right)
    add_text(
        slide,
        text=date,
        left=9.0, top=6.5, width=3.5, height=0.4,
        font_size=12, bold=False,
        color="text_muted", alignment="right",
        palette=palette,
        colors=colors,
    )

    add_source_line(slide, source, palette)
    return prs
