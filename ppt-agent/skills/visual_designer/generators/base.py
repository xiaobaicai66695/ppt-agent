"""PPT generator base module - shared utilities and palette definitions."""
from __future__ import annotations

import os
from copy import deepcopy
from typing import Literal

from pptx import Presentation
from pptx.dml.color import RGBColor
from pptx.enum.shapes import MSO_SHAPE, MSO_CONNECTOR_TYPE
from pptx.enum.text import MSO_ANCHOR, PP_ALIGN
from pptx.util import Inches, Pt

# ---------------------------------------------------------------------------
# Palette definitions (浅色基调，饱和度 5%-45%，亮度 80%-95%)
# ---------------------------------------------------------------------------

PALETTES: dict[str, dict[str, str]] = {
    "ocean_soft": {
        "name": "雾霾蓝",
        "primary": "5A8AA8",
        "secondary": "7BA3B8",
        "accent": "A8C4D4",
        "background": "F5F8FA",
        "text": "3A4A52",
        "text_muted": "7A8A92",
        "light_bg": "E8F0F5",
        "divider": "C8D8E5",
    },
    "sage_calm": {
        "name": "鼠尾草绿",
        "primary": "6A8E85",
        "secondary": "84B59F",
        "accent": "A8C5BD",
        "background": "FAF5F2",
        "text": "3A4A46",
        "text_muted": "8A9A92",
        "light_bg": "E8F0EC",
        "divider": "C8D8D0",
    },
    "warm_terracotta": {
        "name": "陶土橙",
        "primary": "C47060",
        "secondary": "D4A574",
        "accent": "E8DDD0",
        "background": "FAF5F2",
        "text": "4A3A32",
        "text_muted": "9A8A82",
        "light_bg": "F0E8E0",
        "divider": "D8C8B8",
    },
    "charcoal_light": {
        "name": "浅炭灰",
        "primary": "5A6A75",
        "secondary": "8A9AA5",
        "accent": "C8D0D5",
        "background": "F8F8F8",
        "text": "3A4248",
        "text_muted": "8A9298",
        "light_bg": "E8ECEE",
        "divider": "C0C8CC",
    },
    "berry_cream": {
        "name": "玫瑰灰粉",
        "primary": "8D5A6B",
        "secondary": "B89098",
        "accent": "E0D0D4",
        "background": "FAF5F5",
        "text": "4A3238",
        "text_muted": "9A8288",
        "light_bg": "F0E4E8",
        "divider": "D8C0C8",
    },
    "lavender_mist": {
        "name": "薰衣草灰",
        "primary": "7A6A8A",
        "secondary": "9A8AAA",
        "accent": "C8B8D4",
        "background": "F8F5FA",
        "text": "3A3248",
        "text_muted": "8A7A98",
        "light_bg": "EDE4F4",
        "divider": "D0C0E0",
    },
    "forest_moss": {
        "name": "苔藓绿",
        "primary": "5A8E5A",
        "secondary": "7EBF7E",
        "accent": "A8D4A8",
        "background": "F5FAF5",
        "text": "1A2A1A",
        "text_muted": "5A7A5A",
        "light_bg": "E4F4E4",
        "divider": "A8C8A8",
    },
    "sunset_peach": {
        "name": "杏色",
        "primary": "C4A080",
        "secondary": "D8B898",
        "accent": "E8D8C8",
        "background": "FAF5F2",
        "text": "3A2010",
        "text_muted": "8A6A4A",
        "light_bg": "F0E4D8",
        "divider": "D8C8B0",
    },

    # ── 中国场景配色 ──
    "government_red": {
        "name": "政务红",
        "primary": "8B1A1A",
        "secondary": "B87070",
        "accent": "D4A0A0",
        "background": "FDF8F5",
        "text": "3A1A1A",
        "text_muted": "8A6A6A",
        "light_bg": "F5E8E5",
        "divider": "D8C0B8",
    },
    "patriotic_blue": {
        "name": "爱国蓝",
        "primary": "2B5797",
        "secondary": "4A7AB0",
        "accent": "A0B8D4",
        "background": "F5F8FC",
        "text": "1A2A48",
        "text_muted": "5A6A8A",
        "light_bg": "E4EAF4",
        "divider": "B8C8D8",
    },
    "civic_gold": {
        "name": "公民金",
        "primary": "B8860B",
        "secondary": "C8A040",
        "accent": "E8D890",
        "background": "FAF8F0",
        "text": "3A2A0A",
        "text_muted": "8A7A4A",
        "light_bg": "F0E8D8",
        "divider": "D8C8A0",
    },
    "debate_purple": {
        "name": "答辩紫",
        "primary": "5A4A7A",
        "secondary": "8A74AA",
        "accent": "C8B8D4",
        "background": "F8F5FA",
        "text": "2A1A3A",
        "text_muted": "7A6A8A",
        "light_bg": "EDE4F4",
        "divider": "C8B8D8",
    },
    "activity_orange": {
        "name": "活力橙",
        "primary": "C06030",
        "secondary": "C89060",
        "accent": "E8C8A0",
        "background": "FAF5F2",
        "text": "3A2010",
        "text_muted": "8A6A4A",
        "light_bg": "F0E0D0",
        "divider": "D8C0A8",
    },
    "report_green": {
        "name": "报告绿",
        "primary": "3A6A4A",
        "secondary": "5A8A6A",
        "accent": "A0C0B0",
        "background": "F5F8F5",
        "text": "1A2A1A",
        "text_muted": "5A7A5A",
        "light_bg": "E4F0E8",
        "divider": "A8C0B0",
    },
    "simple_gray": {
        "name": "简约灰",
        "primary": "4A4A4A",
        "secondary": "6A6A6A",
        "accent": "C0C0C0",
        "background": "F8F8F8",
        "text": "2A2A2A",
        "text_muted": "8A8A8A",
        "light_bg": "E8E8E8",
        "divider": "C0C0C0",
    },
}


def rgb(hex_color: str) -> RGBColor:
    """Convert hex string like '5A8AA8' to RGBColor."""
    hex_color = hex_color.lstrip("#")
    return RGBColor(
        int(hex_color[0:2], 16),
        int(hex_color[2:4], 16),
        int(hex_color[4:6], 16),
    )


# ---------------------------------------------------------------------------
# Presentation helpers
# ---------------------------------------------------------------------------

def new_presentation(
    width_in: float = 13.333,
    height_in: float = 7.5,
    palette: str = "ocean_soft",
) -> Presentation:
    """Create a new 16:9 presentation. No slide is created — generators add slides."""
    prs = Presentation()
    prs.slide_width = Inches(width_in)
    prs.slide_height = Inches(height_in)
    return prs


def set_slide_background(slide, palette: str = "ocean_soft"):
    """Set solid background color for a slide."""
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    bg = slide.background
    fill = bg.fill
    fill.solid()
    fill.fore_color.rgb = rgb(colors["background"])


def save_presentation(prs: Presentation, output_path: str):
    """Save the presentation to a file path."""
    os.makedirs(os.path.dirname(output_path) or ".", exist_ok=True)
    prs.save(output_path)


def _clone_xml(el):
    """Clone an lxml element by round-tripping through string serialization.
    This produces a clean copy detached from the source document tree,
    avoiding namespace conflicts that deepcopy can cause across different Presentations.
    """
    from lxml import etree

    return etree.fromstring(etree.tostring(el, encoding="unicode"))


def save_slide(slide, output_path: str):
    """
    Save a single slide as a standalone PPTX file.
    Creates a new presentation, clones the slide into it, and saves.
    """
    from pptx.oxml.ns import qn

    new_prs = Presentation()
    new_prs.slide_width = slide.part.package.presentation_part.presentation.slide_width
    new_prs.slide_height = slide.part.package.presentation_part.presentation.slide_height

    blank_layout = None
    for lo in new_prs.slide_layouts:
        if "blank" in lo.name.lower():
            blank_layout = lo
            break
    if blank_layout is None:
        blank_layout = new_prs.slide_layouts[6]

    dest_slide = new_prs.slides.add_slide(blank_layout)

    # Remove placeholder shapes from the blank layout
    for shape in list(dest_slide.shapes):
        sp = shape.element
        sp.getparent().remove(sp)

    # Clone shapes from source slide. Use _clone_xml instead of deepcopy
    # to avoid namespace-reference issues across different lxml trees.
    spTree = dest_slide.shapes._spTree
    extLst = spTree.find(qn("p:extLst"))

    for shape in slide.shapes:
        el = _clone_xml(shape.element)
        if extLst is not None:
            extLst.addprevious(el)
        else:
            spTree.append(el)

    # Clone background if present
    bg_src = slide.background
    if bg_src is not None:
        try:
            bg_el = _clone_xml(bg_src.element)
            p_cSld = dest_slide._element.find(qn("p:cSld"))
            existing_bg = p_cSld.find(qn("p:bg"))
            if existing_bg is not None:
                p_cSld.replace(existing_bg, bg_el)
            else:
                p_cSld.insert(0, bg_el)
        except Exception:
            pass

    save_presentation(new_prs, output_path)


def add_source_line(
    slide,
    source: str,
    palette: str = "ocean_soft",
):
    """在幻灯片底部渲染数据来源/参考信息。
    当 source 非空时，在底部添加灰色小字来源行。
    """
    if not source:
        return

    # 底部淡灰色分割线
    add_rect(
        slide,
        left=0.5, top=6.85, width=12.0, height=0.01,
        fill_color="text_muted", palette=palette,
    )

    # 来源文字（小号灰色）
    add_text(
        slide,
        text=source,
        left=0.5, top=6.9, width=12.0, height=0.35,
        font_size=9, bold=False,
        color="text_muted", alignment="left",
        palette=palette,
    )


# ---------------------------------------------------------------------------
# Text helpers
# ---------------------------------------------------------------------------

def add_text(
    slide,
    text: str,
    left: float,
    top: float,
    width: float,
    height: float,
    font_size: float = 14,
    bold: bool = False,
    color: str = "text",
    alignment: Literal["left", "center", "right"] = "left",
    palette: str = "ocean_soft",
    font_name: str = None,
    italic: bool = False,
) -> "pptx.shapes.shapetree.Shape":
    """Add a text box to the slide with consistent styling."""
    txbox = slide.shapes.add_textbox(Inches(left), Inches(top), Inches(width), Inches(height))
    tf = txbox.text_frame
    tf.word_wrap = True

    p = tf.paragraphs[0]
    p.text = text
    p.font.size = Pt(font_size)
    p.font.bold = bold
    p.font.italic = italic

    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    p.font.color.rgb = rgb(colors.get(color, color))
    p.alignment = {
        "left": PP_ALIGN.LEFT,
        "center": PP_ALIGN.CENTER,
        "right": PP_ALIGN.RIGHT,
    }.get(alignment, PP_ALIGN.LEFT)

    if font_name:
        p.font.name = font_name

    return txbox


def add_paragraph(
    tf,
    text: str,
    font_size: float = 14,
    bold: bool = False,
    color: str = "text",
    palette: str = "ocean_soft",
    alignment: Literal["left", "center", "right"] = "left",
    space_before: float = 0,
    space_after: float = 0,
    italic: bool = False,
    font_name: str = None,
):
    """Add a paragraph to an existing text frame."""
    p = tf.add_paragraph()
    p.text = text
    p.font.size = Pt(font_size)
    p.font.bold = bold
    p.font.italic = italic
    p.space_before = Pt(space_before)
    p.space_after = Pt(space_after)
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    p.font.color.rgb = rgb(colors.get(color, color))
    p.alignment = {
        "left": PP_ALIGN.LEFT,
        "center": PP_ALIGN.CENTER,
        "right": PP_ALIGN.RIGHT,
    }.get(alignment, PP_ALIGN.LEFT)
    if font_name:
        p.font.name = font_name
    return p


# ---------------------------------------------------------------------------
# Shape helpers
# ---------------------------------------------------------------------------

def add_rect(
    slide,
    left: float,
    top: float,
    width: float,
    height: float,
    fill_color: str,
    palette: str = "ocean_soft",
    line_color: str = None,
    line_width: float = 0,
) -> "pptx.shapes.shapetree.Shape":
    """Add a filled rectangle."""
    shape = slide.shapes.add_shape(
        MSO_SHAPE.RECTANGLE,
        Inches(left), Inches(top), Inches(width), Inches(height),
    )
    fill = shape.fill
    fill.solid()
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    fill.fore_color.rgb = rgb(colors.get(fill_color, fill_color))

    line = shape.line
    if line_color:
        line.color.rgb = rgb(colors.get(line_color, line_color))
        line.width = Pt(line_width)
    else:
        line.fill.background()

    return shape


def add_round_rect(
    slide,
    left: float,
    top: float,
    width: float,
    height: float,
    fill_color: str,
    palette: str = "ocean_soft",
    line_color: str = None,
    line_width: float = 0,
) -> "pptx.shapes.shapetree.Shape":
    """Add a rounded rectangle."""
    shape = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(left), Inches(top), Inches(width), Inches(height),
    )
    fill = shape.fill
    fill.solid()
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    fill.fore_color.rgb = rgb(colors.get(fill_color, fill_color))

    line = shape.line
    if line_color:
        line.color.rgb = rgb(colors.get(line_color, line_color))
        line.width = Pt(line_width)
    else:
        line.fill.background()

    return shape


def add_ellipse(
    slide,
    left: float,
    top: float,
    width: float,
    height: float,
    fill_color: str,
    palette: str = "ocean_soft",
    line_color: str = None,
    line_width: float = 0,
) -> "pptx.shapes.shapetree.Shape":
    """Add an ellipse/circle."""
    shape = slide.shapes.add_shape(
        MSO_SHAPE.OVAL,
        Inches(left), Inches(top), Inches(width), Inches(height),
    )
    fill = shape.fill
    fill.solid()
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    fill.fore_color.rgb = rgb(colors.get(fill_color, fill_color))

    line = shape.line
    if line_color:
        line.color.rgb = rgb(colors.get(line_color, line_color))
        line.width = Pt(line_width)
    else:
        line.fill.background()

    return shape


def add_line(
    slide,
    x1: float,
    y1: float,
    x2: float,
    y2: float,
    color: str = "primary",
    width: float = 1.5,
    palette: str = "ocean_soft",
) -> "pptx.shapes.shapetree.Shape":
    """Add a straight line connector."""
    connector = slide.shapes.add_connector(
        MSO_CONNECTOR_TYPE.STRAIGHT,
        Inches(x1), Inches(y1), Inches(x2), Inches(y2),
    )
    connector.line.color.rgb = rgb(PALETTES.get(palette, PALETTES["ocean_soft"]).get(color, color))
    connector.line.width = Pt(width)
    return connector


def add_arrow(
    slide,
    x1: float,
    y1: float,
    x2: float,
    y2: float,
    color: str = "secondary",
    width: float = 1.5,
    palette: str = "ocean_soft",
) -> "pptx.shapes.shapetree.Shape":
    """Add a right-pointing arrow."""
    shape = slide.shapes.add_shape(
        MSO_SHAPE.RIGHT_ARROW,
        Inches(x1), Inches(y1), Inches(x2 - x1), Inches(y2 - y1),
    )
    shape.fill.solid()
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    shape.fill.fore_color.rgb = rgb(colors.get(color, color))
    shape.line.fill.background()
    return shape


def add_text_in_shape(
    shape,
    text: str,
    font_size: float = 14,
    bold: bool = False,
    color: str = "text",
    palette: str = "ocean_soft",
    alignment: Literal["left", "center", "right"] = "left",
    vertical_alignment: str = "top",
    font_name: str = None,
    italic: bool = False,
):
    """Set text inside an existing shape."""
    tf = shape.text_frame
    tf.word_wrap = True

    # Clear existing paragraphs except first
    for i in range(len(tf.paragraphs) - 1, 0, -1):
        p = tf.paragraphs[i]._p
        p.getparent().remove(p)

    p = tf.paragraphs[0]
    p.text = text
    p.font.size = Pt(font_size)
    p.font.bold = bold
    p.font.italic = italic
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    p.font.color.rgb = rgb(colors.get(color, color))
    p.alignment = {
        "left": PP_ALIGN.LEFT,
        "center": PP_ALIGN.CENTER,
        "right": PP_ALIGN.RIGHT,
    }.get(alignment, PP_ALIGN.LEFT)

    if font_name:
        p.font.name = font_name

    # Vertical alignment
    tf.anchor = {
        "top": MSO_ANCHOR.TOP,
        "middle": MSO_ANCHOR.MIDDLE,
        "bottom": MSO_ANCHOR.BOTTOM,
        "center": MSO_ANCHOR.MIDDLE,
    }.get(vertical_alignment, MSO_ANCHOR.TOP)
