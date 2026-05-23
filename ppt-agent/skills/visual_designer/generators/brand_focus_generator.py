"""Generator for brand_focus - 品牌价值聚焦页."""
import math
import random
from typing import Optional, List
from pptx import Presentation
from .base import (
    add_source_line, new_presentation, PALETTES, add_text, add_rect, add_round_rect,
    add_ellipse, add_line, set_slide_background,
)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "",
    title: str = "{品牌标题}",
    subtitle: str = "",
    center_text: str = "核心\n理念",
    surrounding_points: Optional[List[dict]] = None,
    principles: Optional[List[dict]] = None,
) -> Presentation:
    """Generate a brand focus slide with center circle and surrounding points.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        source: Data source annotation.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Subtitle below title.
        center_text: Text to display in the center circle.
        surrounding_points: List of dicts with keys: title, description, color, angle
        principles: List of dicts with keys: title, description

    Returns:
        The Presentation object.
    """
    if surrounding_points is None:
        surrounding_points = [
            {"title": "创新", "description": "持续突破技术边界", "color": "secondary", "angle": 45},
            {"title": "品质", "description": "精益求精的产品", "color": "accent", "angle": 135},
            {"title": "服务", "description": "专业贴心的支持", "color": "secondary", "angle": 225},
            {"title": "责任", "description": "可持续发展的承诺", "color": "accent", "angle": 315},
        ]

    if principles is None:
        principles = [
            {"title": "以用户为中心", "description": "每一项决策都基于用户真实需求"},
            {"title": "长期主义", "description": "坚持做正确的事，而非容易的事"},
            {"title": "开放创新", "description": "拥抱变化，持续学习和迭代"},
            {"title": "共赢合作", "description": "与伙伴共同成长，共享成果"},
        ]

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    set_slide_background(slide, palette)

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    y_t = 0.3
    if kicker:
        add_text(slide, text=kicker, left=0.5, top=0.2, width=12.0, height=0.2,
            font_size=12, bold=False, color="text_muted", alignment="left", palette=palette)
        y_t = 0.35
    add_text(slide, text=title, left=0.5, top=y_t, width=12.0, height=0.5,
        font_size=32, bold=True, color="text", alignment="left", palette=palette)
    if subtitle:
        add_text(slide, text=subtitle, left=0.5, top=y_t + 0.5, width=12.0, height=0.3,
            font_size=14, bold=False, color="text_muted", alignment="left", palette=palette)

    # Center focus area
    center_x, center_y = 3.2, 4.2

    # Outermost layer - decorative circles
    for i in range(3):
        add_ellipse(
            slide,
            left=center_x - 1.0 - i * 0.25,
            top=center_y - 1.0 - i * 0.25,
            width=2.0 + i * 0.5,
            height=2.0 + i * 0.5,
            fill_color="light_bg", palette=palette,
        )

    # Middle layer - accent
    add_ellipse(
        slide,
        left=center_x - 0.9,
        top=center_y - 0.9,
        width=1.8,
        height=1.8,
        fill_color="accent", palette=palette,
    )

    # Core - primary circle
    add_ellipse(
        slide,
        left=center_x - 0.6,
        top=center_y - 0.6,
        width=1.2,
        height=1.2,
        fill_color="primary", palette=palette,
    )

    # Center text
    add_text(
        slide, text=center_text,
        left=center_x - 0.5,
        top=center_y - 0.4,
        width=1.0,
        height=0.9,
        font_size=14, bold=True, color="background",
        alignment="center", palette=palette,
    )

    # Surrounding points
    random.seed(42)
    for point in surrounding_points:
        angle = math.radians(point.get("angle", 0))
        r = 2.2
        vx = center_x + r * math.cos(angle)
        vy = center_y + r * math.sin(angle)

        # Connection line
        line_end_x = center_x + 1.1 * math.cos(angle)
        line_end_y = center_y + 1.1 * math.sin(angle)

        for s in range(10):
            t1 = s / 10
            t2 = (s + 0.5) / 10
            sx1 = line_end_x + (vx - line_end_x) * t1 + random.uniform(-0.02, 0.02)
            sy1 = line_end_y + (vy - line_end_y) * t1 + random.uniform(-0.02, 0.02)
            sx2 = line_end_x + (vx - line_end_x) * t2 + random.uniform(-0.02, 0.02)
            sy2 = line_end_y + (vy - line_end_y) * t2 + random.uniform(-0.02, 0.02)
            if s % 2 == 0:
                add_line(slide, sx1, sy1, sx2, sy2,
                         color="divider", width=1.2, palette=palette)

        # Value point circle
        point_color = point.get("color", "secondary")
        add_ellipse(
            slide,
            left=vx - 0.5,
            top=vy - 0.4,
            width=1.0,
            height=0.8,
            fill_color=point_color, palette=palette,
        )
        add_text(
            slide, text=point.get("title", ""),
            left=vx - 0.5,
            top=vy - 0.25,
            width=1.0,
            height=0.5,
            font_size=11, bold=True, color="background",
            alignment="center", palette=palette,
        )

    # Right panel
    add_rect(slide, left=7.0, top=1.4, width=5.8, height=5.4,
             fill_color="background", palette=palette)
    add_rect(slide, left=7.0, top=1.4, width=5.8, height=0.08,
             fill_color="primary", palette=palette)

    add_text(
        slide, text="核心理念",
        left=7.2, top=1.6, width=5.4, height=0.4,
        font_size=16, bold=True, color="text", palette=palette,
    )

    for i, p in enumerate(principles):
        y = 2.2 + i * 1.05
        add_rect(slide, left=7.2, top=y, width=5.4, height=0.9,
                 fill_color="light_bg", palette=palette)
        add_rect(slide, left=7.2, top=y, width=0.06, height=0.9,
                 fill_color="primary", palette=palette)

        add_text(
            slide, text=f"{i+1:02d}",
            left=7.4, top=y + 0.2, width=0.5, height=0.5,
            font_size=16, bold=True, color="primary", palette=palette,
        )
        add_text(
            slide, text=p.get("title", ""),
            left=8.0, top=y + 0.1, width=4.4, height=0.35,
            font_size=13, bold=True, color="text", palette=palette,
        )
        add_text(
            slide, text=p.get("description", ""),
            left=8.0, top=y + 0.45, width=4.4, height=0.35,
            font_size=11, color="text_muted", palette=palette,
        )
        if i < len(principles) - 1:
            add_rect(slide, left=7.2, top=y + 0.7, width=5.4, height=0.01,
                     fill_color="divider", palette=palette)

    add_source_line(slide, source, palette)
    return prs
