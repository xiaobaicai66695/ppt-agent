"""
Background template manager - 背景图片路径获取工具

按主题或场景获取背景图片路径，支持动态扩充。
"""
import os
import random
from pathlib import Path
from typing import Optional


# 主题到中文名称和场景关键词的映射
THEME_MAPPING = {
    "party_government": {
        "name_cn": "党政办公",
        "scenarios": ["党建", "政府", "政务", "机关", "党委", "党支部", "红色"],
        "priority": 10,
    },
    "minimalist_blue": {
        "name_cn": "简约蓝白",
        "scenarios": ["商务", "企业", "科技", "现代", "简约", "专业", "会议", "方案", "产品"],
        "priority": 5,
    },
    "vintage_chinese": {
        "name_cn": "复古中国风",
        "scenarios": ["中国风", "传统", "文化", "国风", "古风", "复古", "文艺"],
        "priority": 7,
    },
    "ink_wash_mountain": {
        "name_cn": "水墨山水",
        "scenarios": ["水墨", "山水", "艺术", "自然"],
        "priority": 8,
    },
    "snowy_mountain": {
        "name_cn": "雪山风景",
        "scenarios": ["自然", "风景", "户外", "雪山", "山川", "旅行", "环保"],
        "priority": 4,
    },
    "artistic": {
        "name_cn": "艺术涂鸦",
        "scenarios": ["艺术", "创意", "涂鸦", "个性", "时尚", "现代艺术"],
        "priority": 6,
    },
}


def get_background_dir() -> Path:
    """获取 background_templates 目录路径"""
    current_dir = Path(__file__).parent.parent
    return current_dir / "background_templates"


def scan_backgrounds() -> list[dict]:
    """
    扫描 background_templates 目录，返回所有可用背景。

    Returns:
        [
            {
                "theme": "party_government",
                "name_cn": "党政办公",
                "scenarios": ["党建", "政府", ...],
                "priority": 10,
                "images": [
                    "D:/path/to/1.jpg",
                    "D:/path/to/2.jpg",
                ]
            },
            ...
        ]
    """
    bg_dir = get_background_dir()
    results = []

    if not bg_dir.exists():
        return results

    for theme_dir in bg_dir.iterdir():
        if not theme_dir.is_dir():
            continue

        theme_name = theme_dir.name
        mapping = THEME_MAPPING.get(theme_name, {})
        images = []

        # 扫描 images 子目录
        images_dir = theme_dir / "images"
        if images_dir.exists():
            for img in sorted(images_dir.glob("*.jpg")):
                images.append(str(img))

        # 扫描根目录的 background.jpg
        root_bg = theme_dir / "background.jpg"
        if root_bg.exists() and str(root_bg) not in images:
            images.append(str(root_bg))

        if images:
            results.append({
                "theme": theme_name,
                "name_cn": mapping.get("name_cn", theme_name),
                "scenarios": mapping.get("scenarios", []),
                "priority": mapping.get("priority", 0),
                "images": images,
            })

    return sorted(results, key=lambda x: -x["priority"])


def get_background(
    theme: Optional[str] = None,
    scenario: Optional[str] = None,
    random_select: bool = False,
) -> Optional[str]:
    """
    获取背景图片路径。

    Args:
        theme: 主题标识 (如 "party_government", "minimalist_blue")
        scenario: 场景关键词 (如 "党建汇报", "商务演示")
        random_select: 是否随机选择匹配的主题

    Returns:
        背景图片的完整路径，或 None

    示例:
        get_background(theme="party_government")
        get_background(scenario="党建")
        get_background(scenario="商务", random_select=True)
    """
    backgrounds = scan_backgrounds()
    if not backgrounds:
        return None

    candidates = backgrounds

    # 优先按 theme 匹配
    if theme:
        candidates = [b for b in backgrounds if b["theme"] == theme]
        if not candidates:
            # 模糊匹配 theme
            candidates = [b for b in backgrounds if theme in b["theme"] or theme in b["name_cn"]]

    # 按 scenario 匹配
    if not candidates and scenario:
        scenario_lower = scenario.lower()
        candidates = [
            b for b in backgrounds
            if scenario_lower in b["name_cn"].lower() or
               any(scenario_lower in s.lower() for s in b["scenarios"])
        ]

    if not candidates:
        candidates = backgrounds

    # 随机或优先级
    if random_select:
        selected = random.choice(candidates)
    else:
        selected = candidates[0]

    images = selected.get("images", [])
    if images:
        return random.choice(images) if random_select else images[0]

    return None


def list_themes() -> list[dict]:
    """列出所有可用主题及图片"""
    return scan_backgrounds()
