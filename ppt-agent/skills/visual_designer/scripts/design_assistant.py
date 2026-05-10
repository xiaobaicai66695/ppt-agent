#!/usr/bin/env python3
"""
Visual Designer Assistant for PowerPoint Presentations

基于 pptx 设计理念的专业幻灯片视觉设计辅助工具。

设计理念核心：
- 拒绝无聊：纯文字+白底无法打动任何人
- 主次分明：一种颜色占 60-70%，配合 1-2 种辅助色和锐利强调色
- 统一使用浅色基调：所有幻灯片统一使用浅色背景，营造明亮清爽的视觉效果
- 视觉主题贯穿始终：选择独特元素在每页重复使用
- 每页必须有视觉元素：图片、图表、图标或形状
- 禁止标题下装饰线（典型 AI 感特征）
- 模板优先：优先使用完整模板改编，而非从零设计
"""

import sys
import json
import os
import argparse
from typing import Dict, List, Optional, Any
from pathlib import Path

# ============================================================
# 模板资产路径（相对于脚本所在目录）
# ============================================================
_SCRIPT_DIR = Path(__file__).parent.resolve()
_TEMPLATE_DIR = _SCRIPT_DIR.parent / "templates"
_SINGLE_PAGE_DIR = _TEMPLATE_DIR / "single-page"
_FULL_DECK_DIR = _TEMPLATE_DIR / "full-decks"


def _load_py_template(path: Path) -> Optional[Dict[str, Any]]:
    """加载单个 Python 模板文件（从 TEMPLATE 变量），失败返回 None。"""
    try:
        namespace = {}
        with open(path, "r", encoding="utf-8") as f:
            exec(f.read(), namespace)
        return namespace.get("TEMPLATE")
    except Exception:
        return None


def load_all_single_page_templates() -> Dict[str, Dict[str, Any]]:
    """加载所有单页布局模板。"""
    templates = {}
    if _SINGLE_PAGE_DIR.exists():
        for f in sorted(_SINGLE_PAGE_DIR.glob("*.py")):
            t = _load_py_template(f)
            if t:
                templates[t.get("type", f.stem)] = t
    return templates


def load_all_full_deck_templates() -> Dict[str, Dict[str, Any]]:
    """加载所有完整 PPT 模板。"""
    templates = {}
    if _FULL_DECK_DIR.exists():
        for f in sorted(_FULL_DECK_DIR.glob("*.py")):
            t = _load_py_template(f)
            if t:
                templates[t.get("name", f.stem)] = t
    return templates


def load_single_page_template(page_type: str) -> Optional[Dict[str, Any]]:
    """加载指定类型的单页模板。"""
    path = _SINGLE_PAGE_DIR / f"{page_type}.py"
    return _load_py_template(path)


def load_full_deck_template(deck_name: str) -> Optional[Dict[str, Any]]:
    """加载指定名称的完整模板。"""
    path = _FULL_DECK_DIR / f"{deck_name}.py"
    return _load_py_template(path)


def recommend_full_deck_template(user_content: str) -> List[Dict[str, str]]:
    """根据用户需求推荐匹配的完整模板。

    分析用户输入中的关键词，返回最匹配的模板列表。
    """
    content_lower = user_content.lower()
    recommendations = []

    # 关键词匹配规则
    keyword_map = {
        "tech-sharing": ["技术分享", "技术培训", "架构", "技术文档", "原理", "源码", "开发", "工程师", "技术"],
        "tech-intro": ["技术介绍", "技术科普", "知识分享", "新技术", "行业科普", "概念普及"],
        "product-launch": ["产品发布", "新产品", "产品介绍", "功能演示", "发布会", "客户演示"],
        "weekly-report": ["周报", "月报", "工作汇报", "周总结", "月度总结", "工作进度"],
        "pitch-deck": ["商业计划", "路演", "融资", "创业", "投资", "商业", "vc", "创业计划", "商业画布"],
        "course-module": ["课程", "课件", "教学", "培训", "学习", "知识点", "教材", "课堂"],
    }

    for template_name, keywords in keyword_map.items():
        score = 0
        for kw in keywords:
            if kw.lower() in content_lower:
                score += 1
        if score > 0:
            deck = load_full_deck_template(template_name)
            if deck:
                recommendations.append({
                    "name": template_name,
                    "name_cn": deck.get("name_cn", template_name),
                    "description": deck.get("description", ""),
                    "score": score,
                    "target_audience": deck.get("target_audience", ""),
                })

    # 按分数排序
    recommendations.sort(key=lambda x: x["score"], reverse=True)
    return recommendations


def generate_slides_from_full_deck(deck_name: str, user_content: str) -> Optional[Dict[str, Any]]:
    """基于完整模板生成幻灯片结构。

    根据用户内容填充模板中的占位符，返回完整的幻灯片列表。
    """
    deck = load_full_deck_template(deck_name)
    if not deck:
        return None

    slides = []
    for slide_def in deck.get("slide_structure", []):
        slide = slide_def.copy()
        slide["content_plan"] = slide_def.get("content_plan", {})
        slides.append(slide)

    return {
        "template": deck_name,
        "palette": deck.get("palette", "ocean_soft"),
        "typography": deck.get("typography", {}),
        "slides": slides,
        "design_tips": deck.get("design_tips", []),
    }


def get_template_catalog() -> Dict[str, Any]:
    """获取模板目录总览。"""
    single_pages = load_all_single_page_templates()
    full_decks = load_all_full_deck_templates()

    return {
        "single_page_count": len(single_pages),
        "full_deck_count": len(full_decks),
        "single_page_types": list(single_pages.keys()),
        "full_decks": [
            {
                "name": d.get("name"),
                "name_cn": d.get("name_cn"),
                "description": d.get("description"),
                "target_audience": d.get("target_audience"),
                "typical_slides": d.get("typical_slides"),
                "palette": d.get("palette"),
            }
            for d in full_decks.values()
        ],
        "palettes": list(COLOR_PALETTES.keys()),
    }


def suggest_page_layout(content_type: str, content: Dict[str, Any]) -> Dict[str, Any]:
    """基于内容和类型推荐页面布局细节。

    结合单页模板的规范，给出具体的布局建议。
    """
    template = load_single_page_template(content_type)
    if not template:
        # 回退到 content_slide
        template = load_single_page_template("content_slide")

    if not template:
        return {}

    # 合并模板规范和具体内容
    layout = {
        "type": content_type,
        "name": template.get("name", content_type),
        "elements": template.get("elements", {}),
        "visual_elements": template.get("visual_elements", []),
        "never": template.get("never", []),
        "example": content if content else template.get("example", {}),
    }

    return layout


def print_template_catalog():
    """打印模板目录。"""
    catalog = get_template_catalog()
    print("\n=== 模板资产目录 ===")
    print(f"单页模板: {catalog['single_page_count']} 种")
    print(f"完整模板: {catalog['full_deck_count']} 种")

    print("\n--- 单页模板类型 ---")
    for t in sorted(catalog["single_page_types"]):
        print(f"  - {t}")

    print("\n--- 完整 PPT 模板 ---")
    for deck in catalog["full_decks"]:
        print(f"  [{deck['name']}] {deck['name_cn']}")
        print(f"    描述: {deck['description']}")
        print(f"    受众: {deck['target_audience']}")
        print(f"    典型页数: {deck['typical_slides']} 页 | 配色: {deck['palette']}")
        print()


def print_template_recommendation(user_content: str):
    """打印模板推荐结果。"""
    recs = recommend_full_deck_template(user_content)
    if not recs:
        print("\n=== 模板推荐 ===")
        print("  未找到明确匹配的模板，将使用单页模板组合生成。")
        return

    print(f"\n=== 模板推荐 (共 {len(recs)} 个匹配) ===")
    for i, rec in enumerate(recs, 1):
        print(f"  {i}. [{rec['name']}] {rec['name_cn']}")
        print(f"     描述: {rec['description']}")
        print(f"     受众: {rec['target_audience']}")


def print_template_detail(deck_name: str):
    """打印指定模板的详细信息。"""
    deck = load_full_deck_template(deck_name)
    if not deck:
        print(f"\n未找到模板: {deck_name}")
        return

    print(f"\n=== 模板详情: {deck.get('name_cn', deck_name)} ===")
    print(f"描述: {deck.get('description')}")
    print(f"目标受众: {deck.get('target_audience')}")
    print(f"典型页数: {deck.get('typical_slides')} 页")
    print(f"配色: {deck.get('palette')}")
    print(f"\n幻灯片结构:")
    for slide in deck.get("slide_structure", []):
        idx = slide.get("index", "?")
        stype = slide.get("type", "?")
        title = slide.get("title", slide.get("name", ""))
        print(f"  {idx:2d}. [{stype}] {title}")

    print(f"\n设计建议:")
    for tip in deck.get("design_tips", []):
        print(f"  - {tip}")

# ============================================================
# 配色方案（浅色基调，统一使用柔和浅色）
# ============================================================
# 所有主题统一使用浅色背景，禁止深色背景
# 主色使用柔和的灰感色彩，亮度适中，不刺眼

COLOR_PALETTES = {
    # Sage Calm - 鼠尾草绿，平静治愈
    "sage_calm": {
        "primary": "#6A8E85",
        "secondary": "#84B59F",
        "accent": "#A8C5BD",
        "background": "#FAF5F2",
        "text": "#4A5A55",
        "light_bg": "#E8F0EC",
    },
    # Forest & Moss - 森林苔藓，自然清新
    "forest_moss": {
        "primary": "#4A7C4E",
        "secondary": "#6A9A6E",
        "accent": "#97BC62",
        "background": "#FAF8F5",
        "text": "#3A5A3E",
        "light_bg": "#E0EAD8",
    },
    # Warm Terracotta - 陶土暖调，温暖有质感
    "warm_terracotta": {
        "primary": "#C47060",
        "secondary": "#D4A574",
        "accent": "#E8C8A8",
        "background": "#FAF5F2",
        "text": "#6A4A40",
        "light_bg": "#F0E4DC",
    },
    # Berry & Cream - 玫瑰灰粉，优雅柔和
    "berry_cream": {
        "primary": "#8D5A6B",
        "secondary": "#A88090",
        "accent": "#C8A8B0",
        "background": "#FAF5F5",
        "text": "#5A3A45",
        "light_bg": "#EDE0E4",
    },
    # Ocean Soft - 雾霾蓝，清新专业
    "ocean_soft": {
        "primary": "#5A8AA8",
        "secondary": "#7BA3B8",
        "accent": "#A8C4D4",
        "background": "#F5F8FA",
        "text": "#3A5A6A",
        "light_bg": "#E0EAF0",
    },
    # Charcoal Light - 浅炭灰，现代简洁
    "charcoal_light": {
        "primary": "#5A6A75",
        "secondary": "#8A9AA5",
        "accent": "#B8C4CC",
        "background": "#F8F8F8",
        "text": "#3A454D",
        "light_bg": "#E8EAEC",
    },
    # Lavender Mist - 薰衣草灰，文艺清新
    "lavender_mist": {
        "primary": "#7A6A8A",
        "secondary": "#9A8AAA",
        "accent": "#C8B8D4",
        "background": "#F8F5FA",
        "text": "#5A4A6A",
        "light_bg": "#EDE4F0",
    },
    # Sunset Peach - 杏色日落，温暖柔和
    "sunset_peach": {
        "primary": "#C4A080",
        "secondary": "#D8B898",
        "accent": "#E8D0B8",
        "background": "#FAF5F2",
        "text": "#6A5040",
        "light_bg": "#F0E4D8",
    },
    # 专业商务默认
    "professional": {
        "primary": "#5A7A8A",
        "secondary": "#7A9AAA",
        "accent": "#A8C4D4",
        "background": "#FAF5F5",
        "text": "#4A5A65",
        "light_bg": "#E8ECEE",
    },
    # 创意主题
    "creative": {
        "primary": "#8A7A9A",
        "secondary": "#A89AB8",
        "accent": "#C8C0D8",
        "background": "#FAF8FA",
        "text": "#5A4A6A",
        "light_bg": "#EEE8F0",
    },
    # 科技主题
    "tech": {
        "primary": "#5A8AA8",
        "secondary": "#7A9AB8",
        "accent": "#A8C4D4",
        "background": "#F8FAFB",
        "text": "#4A5A65",
        "light_bg": "#E4ECEE",
    },
    # 商务主题
    "business": {
        "primary": "#5A6A75",
        "secondary": "#7A8A95",
        "accent": "#A8B8C4",
        "background": "#FAF8F8",
        "text": "#4A555D",
        "light_bg": "#E8EAEC",
    },
}

# ============================================================
# 字体配对方案（参考 pptx SKILL.md）
# ============================================================
# 选择有特色的标题字体搭配干净的正文/正文字体
# 避免默认使用 Arial

TYPOGRAPHY_PAIRINGS = {
    "classic": {
        "header": "Georgia",
        "body": "Calibri",
        "style": "经典优雅",
    },
    "modern": {
        "header": "Arial Black",
        "body": "Arial",
        "style": "现代简洁",
    },
    "clean": {
        "header": "Calibri",
        "body": "Calibri Light",
        "style": "商务干净",
    },
    "traditional": {
        "header": "Cambria",
        "body": "Calibri",
        "style": "传统专业",
    },
    "tech": {
        "header": "Trebuchet MS",
        "body": "Calibri",
        "style": "科技感",
    },
    "literary": {
        "header": "Palatino",
        "body": "Garamond",
        "style": "文学气质",
    },
}

# 中文字体
CHINESE_FONTS = {
    "header": ["Microsoft YaHei", "SimHei", "PingFang SC"],
    "body": ["Microsoft YaHei", "SimSun", "PingFang SC"],
}

# 字号规范
FONT_SIZE_SPEC = {
    "title": {"min": 36, "max": 44, "unit": "pt"},
    "section_header": {"min": 20, "max": 24, "unit": "pt"},
    "body": {"min": 14, "max": 16, "unit": "pt"},
    "caption": {"min": 10, "max": 12, "unit": "pt"},
}

# ============================================================
# 布局模板（参考 pptx SKILL.md）
# ============================================================
# 布局由内容性质驱动，而非要点数量决定
# 每页必须有视觉元素

LAYOUT_TEMPLATES = {
    "title_slide": {
        "name": "标题页",
        "elements": ["title", "subtitle", "author", "date"],
        "alignment": "center",
        "spacing": "large",
        "bg_style": "浅色背景，可加装饰色块点缀",
        "visual_required": True,
    },
    "content_slide": {
        "name": "内容页",
        "elements": ["title", "bullet_points", "optional_image"],
        "alignment": "left",
        "spacing": "medium",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
    "two_column": {
        "name": "双栏布局",
        "elements": ["title", "left_column", "right_column"],
        "content_modes": [
            "纯要点：left_bullets + right_bullets（向后兼容）",
            "引言+要点：left_intro + left_bullets + right_intro + right_bullets",
            "多区块（推荐）：left_sections{key_points/analysis/data} + right_sections{key_points/analysis/data}"
        ],
        "alignment": "split",
        "spacing": "medium",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
    "icon_text_rows": {
        "name": "图标+文字行",
        "elements": ["title", "icon_circle", "bold_header", "description"],
        "alignment": "left",
        "spacing": "medium",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
    "grid_2x2": {
        "name": "2x2 网格",
        "elements": ["title", "grid_items"],
        "alignment": "grid",
        "spacing": "medium",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
    "half_bleed": {
        "name": "半出血图片",
        "elements": ["title", "full_bleed_image", "content_overlay"],
        "alignment": "overlay",
        "spacing": "large",
        "bg_style": "图片+半透明叠加",
        "visual_required": True,
    },
    "stat_callout": {
        "name": "大数字强调",
        "elements": ["title", "big_number", "small_label"],
        "alignment": "center",
        "spacing": "large",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
    "comparison": {
        "name": "对比列",
        "elements": ["title", "left_side", "right_side"],
        "alignment": "split",
        "spacing": "medium",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
    "timeline": {
        "name": "时间线/流程",
        "elements": ["title", "numbered_steps", "arrows"],
        "alignment": "horizontal",
        "spacing": "medium",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
    "section_divider": {
        "name": "章节分隔页",
        "elements": ["section_title", "section_number"],
        "alignment": "center",
        "spacing": "large",
        "bg_style": "浅色背景，可加深色色块点缀",
        "visual_required": False,
    },
    "quote_slide": {
        "name": "金句/引言页",
        "elements": ["quote_text", "attribution"],
        "alignment": "center",
        "spacing": "large",
        "bg_style": "浅色背景，可加装饰色块",
        "visual_required": False,
    },
    "summary_slide": {
        "name": "总结页",
        "elements": ["title", "key_points", "thank_you"],
        "alignment": "left",
        "spacing": "medium",
        "bg_style": "浅色背景",
        "visual_required": True,
    },
}

# ============================================================
# 排版间距规范
# ============================================================
SPACING_SPEC = {
    "margins": {
        "min": 0.5,  # 英寸
        "recommended": 0.5,
    },
    "content_gap": {
        "min": 0.3,
        "max": 0.5,
        "recommended": 0.3,
    },
    "text_to_shape": {
        "margin": 0,  # 文本框内边距设为0以精确对齐
    },
}

# ============================================================
# 视觉检查清单（QA）
# ============================================================
VISUAL_QA_CHECKLIST = [
    {"check": "overlapping_elements", "desc": "重叠元素（文字穿透形状、线条穿过文字）"},
    {"check": "text_overflow", "desc": "文字溢出或被边缘/形状边界截断"},
    {"check": "decorative_line_mismatch", "desc": "装饰线位置为单行文字设计，但标题换行成两行"},
    {"check": "footer_collision", "desc": "来源引用或页脚与上方内容碰撞"},
    {"check": "insufficient_gap", "desc": "元素间距太小（< 0.3英寸）"},
    {"check": "uneven_gaps", "desc": "间距不均（一处大面积空白，另一处拥挤）"},
    {"check": "margin_violation", "desc": "页边距不足（距边缘 < 0.5英寸）"},
    {"check": "alignment_inconsistency", "desc": "列或类似元素对齐不一致"},
    {"check": "low_text_contrast", "desc": "文字对比度低（如浅灰文字配奶油色背景）"},
    {"check": "low_icon_contrast", "desc": "图标对比度低（如浅色背景上的浅色图标无对比圆形衬托）"},
    {"check": "textbox_too_narrow", "desc": "文本框太窄导致换行过多"},
    {"check": "placeholder_remaining", "desc": "残留的占位符内容"},
    {"check": "title_underline", "desc": "标题下有装饰线（AI感典型特征，禁止）"},
]


def generate_color_palette(theme: str) -> Dict[str, str]:
    """生成指定主题的配色方案。"""
    if theme not in COLOR_PALETTES:
        theme = "professional"
    return COLOR_PALETTES[theme]


def get_typography_pairing(style: str = "clean") -> Dict:
    """获取字体配对方案。"""
    if style not in TYPOGRAPHY_PAIRINGS:
        style = "clean"
    return TYPOGRAPHY_PAIRINGS[style]


def suggest_layouts(slides_count: int, content: str) -> List[Dict]:
    """为演示文稿建议最优布局分布。

    布局由内容性质驱动，而非要点数量决定。
    每页必须有视觉元素。
    """
    layouts = []

    # 始终以标题页开始
    layouts.append({
        "slide_number": 1,
        "layout": "title_slide",
        "reason": "开场标题页，带标题和概述",
    })

    # 计算内容页数量
    content_slides = slides_count - 2  # 减去标题页和总结页

    if content_slides > 0:
        for i in range(content_slides):
            slide_num = i + 2
            # 每3页插入章节分隔页
            if i % 3 == 0 and slides_count > 5:
                layouts.append({
                    "slide_number": slide_num,
                    "layout": "section_divider",
                    "reason": f"第{i // 3 + 1}章介绍",
                })
            # 交替使用不同布局，避免单调
            elif i % 5 == 1:
                layouts.append({
                    "slide_number": slide_num,
                    "layout": "two_column",
                    "reason": "双栏对比或详细视图",
                })
            elif i % 5 == 2:
                layouts.append({
                    "slide_number": slide_num,
                    "layout": "icon_text_rows",
                    "reason": "图标+文字行，适合多要点展示",
                })
            elif i % 5 == 3:
                layouts.append({
                    "slide_number": slide_num,
                    "layout": "grid_2x2",
                    "reason": "2x2网格，适合并列信息展示",
                })
            else:
                layouts.append({
                    "slide_number": slide_num,
                    "layout": "content_slide",
                    "reason": "主要文字内容页",
                })

    # 始终以总结页结尾
    layouts.append({
        "slide_number": slides_count,
        "layout": "summary_slide",
        "reason": "结论和关键要点",
    })

    return layouts


def analyze_visual_hierarchy(content: str) -> Dict:
    """分析内容并建议视觉层级。"""
    words = content.split()

    hierarchy = {
        "title_level": "主标题 - 最大号、加粗、主色",
        "section_level": "章节标题 - 中号、辅色",
        "bullet_level": "要点文字 - 常规、正文色",
        "note_level": "脚注 - 小号、淡化色",
    }

    recommendations = []

    if len(words) > 100:
        recommendations.append({
            "issue": "文字密度过高",
            "suggestion": "考虑拆分为多页或添加视觉元素",
            "priority": "high",
        })

    if any(char in content for char in ["#", "*", "-"]):
        recommendations.append({
            "issue": "可能有未格式化的列表",
            "suggestion": "转换为规范的要点列表以提高可读性",
            "priority": "medium",
        })

    # 40%文字、30%视觉、30%留白的平衡目标
    recommendations.append({
        "issue": "视觉平衡",
        "suggestion": "目标比例：40% 文字，30% 视觉元素，30% 留白",
        "priority": "general",
    })

    return {
        "hierarchy_levels": hierarchy,
        "recommendations": recommendations,
    }


def suggest_spacing(content: str, slides_count: int) -> Dict:
    """提供间距和布局优化建议。"""
    word_count = len(content.split())
    words_per_slide = word_count / max(slides_count, 1)

    spacing = {
        "margins": "0.5 英寸（所有边）",
        "title_to_content": "0.3-0.5 英寸",
        "content_gap": "0.3-0.5 英寸",
        "bullet_spacing": "1.2-1.5x 行高",
    }

    optimization = []

    if words_per_slide > 40:
        optimization.append({
            "metric": f"约 {int(words_per_slide)} 字/页",
            "status": "high",
            "advice": "考虑减少内容或添加视觉元素",
        })
    elif words_per_slide > 25:
        optimization.append({
            "metric": f"约 {int(words_per_slide)} 字/页",
            "status": "optimal",
            "advice": "可读性良好，可考虑增加视觉元素",
        })
    else:
        optimization.append({
            "metric": f"约 {int(words_per_slide)} 字/页",
            "status": "low",
            "advice": "考虑添加更多内容或更大的视觉元素",
        })

    return {
        "spacing_guidelines": spacing,
        "optimization": optimization,
    }


def get_visual_qa_checklist() -> List[Dict]:
    """获取视觉 QA 检查清单。"""
    return VISUAL_QA_CHECKLIST


def generate_full_analysis(theme: str, content: str, slides_count: int, font_style: str = "clean") -> Dict:
    """生成完整的视觉设计分析。"""
    return {
        "color_palette": generate_color_palette(theme),
        "typography": {
            "pairing": get_typography_pairing(font_style),
            "chinese": CHINESE_FONTS,
            "size_spec": FONT_SIZE_SPEC,
        },
        "spacing": SPACING_SPEC,
        "layout_suggestions": suggest_layouts(slides_count, content),
        "visual_hierarchy": analyze_visual_hierarchy(content),
        "qa_checklist": get_visual_qa_checklist(),
        "design_principles": [
            "每页必须有视觉元素：图片、图表、图标或形状",
            "禁止标题下装饰线（典型 AI 感特征）",
            "布局由内容性质驱动，而非要点数量决定",
            "主色占 60-70% 视觉权重，配合 1-2 种辅助色",
            "统一使用浅色基调：所有幻灯片使用浅色背景，营造明亮清爽的视觉效果",
        ],
    }


def print_palette(palette: Dict):
    """以可读格式打印配色方案。"""
    print("\n=== 配色方案 ===")
    for name, color in palette.items():
        print(f"  {name}: {color}")


def print_layouts(layouts: List[Dict]):
    """打印布局建议。"""
    print("\n=== 建议的布局分布 ===")
    for layout in layouts:
        print(f"  第 {layout['slide_number']} 页: {layout['layout']}")
        print(f"    -> {layout['reason']}")


def print_hierarchy(hierarchy: Dict):
    """打印视觉层级指南。"""
    print("\n=== 视觉层级 ===")
    print("  层级定义:")
    for level, desc in hierarchy["hierarchy_levels"].items():
        print(f"    {level}: {desc}")

    print("\n  建议:")
    for rec in hierarchy["recommendations"]:
        print(f"    [{rec['priority'].upper()}] {rec['issue']}: {rec['suggestion']}")


def print_spacing(spacing: Dict):
    """打印间距指南。"""
    print("\n=== 间距指南 ===")
    print("  推荐间距:")
    for key, value in spacing["spacing_guidelines"].items():
        print(f"    {key}: {value}")

    print("\n  优化状态:")
    for opt in spacing["optimization"]:
        status_icon = "OK" if opt["status"] == "optimal" else "WARN"
        print(f"    [{status_icon}] {opt['metric']} - {opt['advice']}")


def print_qa_checklist():
    """打印 QA 检查清单。"""
    print("\n=== 视觉 QA 检查清单 ===")
    for i, item in enumerate(VISUAL_QA_CHECKLIST, 1):
        print(f"  {i}. {item['desc']}")


def main():
    parser = argparse.ArgumentParser(
        description="PPT 视觉设计辅助工具（支持模板系统）"
    )
    parser.add_argument(
        "--mode",
        choices=["palette", "layout", "hierarchy", "spacing", "qa", "full",
                 "template-catalog", "template-recommend", "template-detail"],
        default="full",
        help="分析模式"
    )
    parser.add_argument(
        "--theme",
        choices=[
            "midnight_executive", "forest_moss", "warm_terracotta",
            "ocean_gradient", "charcoal_minimal", "berry_cream",
            "sage_calm", "cherry_bold",
            "professional", "creative", "tech", "business",
        ],
        default="professional",
        help="演示文稿主题"
    )
    parser.add_argument(
        "--font-style",
        choices=["classic", "modern", "clean", "traditional", "tech", "literary"],
        default="clean",
        help="字体配对风格"
    )
    parser.add_argument(
        "--content",
        type=str,
        default="",
        help="要分析的幻灯片内容"
    )
    parser.add_argument(
        "--slides-count",
        type=int,
        default=5,
        help="幻灯片数量"
    )
    parser.add_argument(
        "--output",
        choices=["text", "json"],
        default="text",
        help="输出格式"
    )
    parser.add_argument(
        "--template-name",
        type=str,
        default="",
        help="模板名称（用于 template-detail 模式）"
    )

    args = parser.parse_args()

    # 模板相关模式
    if args.mode == "template-catalog":
        print_template_catalog()
        return

    if args.mode == "template-recommend":
        content = args.content or ""
        print_template_recommendation(content)
        return

    if args.mode == "template-detail":
        name = args.template_name or ""
        print_template_detail(name)
        return

    result = {}

    if args.mode == "palette":
        result = {"color_palette": generate_color_palette(args.theme)}
        if args.output == "text":
            print_palette(result["color_palette"])
        else:
            print(json.dumps(result, indent=2, ensure_ascii=False))
        return

    if args.mode == "layout":
        result = {"layouts": suggest_layouts(args.slides_count, args.content)}
        if args.output == "text":
            print_layouts(result["layouts"])
        else:
            print(json.dumps(result, indent=2, ensure_ascii=False))
        return

    if args.mode == "hierarchy":
        result = {"hierarchy": analyze_visual_hierarchy(args.content)}
        if args.output == "text":
            print_hierarchy(result["hierarchy"])
        else:
            print(json.dumps(result, indent=2, ensure_ascii=False))
        return

    if args.mode == "spacing":
        result = {"spacing": suggest_spacing(args.content, args.slides_count)}
        if args.output == "text":
            print_spacing(result["spacing"])
        else:
            print(json.dumps(result, indent=2, ensure_ascii=False))
        return

    if args.mode == "qa":
        if args.output == "text":
            print_qa_checklist()
        else:
            print(json.dumps({"qa_checklist": get_visual_qa_checklist()}, indent=2, ensure_ascii=False))
        return

    # full 模式：包含模板信息
    result = generate_full_analysis(args.theme, args.content, args.slides_count, args.font_style)

    # 添加模板系统信息
    catalog = get_template_catalog()
    result["template_system"] = {
        "available_templates": catalog["full_decks"],
        "single_page_types": catalog["single_page_types"],
    }

    if args.output == "text":
        print_palette(result["color_palette"])
        print_layouts(result["layout_suggestions"])
        print_hierarchy(result["visual_hierarchy"])
        print_spacing(result["spacing"])
        print_qa_checklist()
        print_template_catalog()
    else:
        print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
