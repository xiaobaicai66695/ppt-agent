"""
示例：使用背景图片生成PPT

运行此脚本生成不同背景风格的PPT示例。
"""
import os
import sys

# 添加父目录路径
current_dir = os.path.dirname(os.path.abspath(__file__))
parent_dir = os.path.dirname(current_dir)
sys.path.insert(0, parent_dir)

from generators.base import (
    new_presentation, save_presentation,
    resolve_background, set_image_background, set_slide_background,
    PALETTES, rgb, add_text, add_rect, add_ellipse,
)
from generators.background_manager import list_themes


def generate_simple_slide_with_bg(title: str, subtitle: str, background: str, output_path: str):
    """生成一张简单的带背景的幻灯片"""
    prs = new_presentation()
    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)

    # 设置背景 - 保持原图清晰
    bg_path = resolve_background(background)
    if bg_path:
        set_image_background(slide, bg_path, brightness=0.95)
    else:
        set_slide_background(slide, "ocean_soft")

    # 文字样式 - 根据背景选择深色或浅色
    if bg_path:
        # 有背景时用白色文字
        title_color = "FFFFFF"
        subtitle_color = "E0E0E0"
    else:
        # 无背景时用深色文字
        title_color = "text"
        subtitle_color = "secondary"

    # 标题
    add_text(
        slide, text=title,
        left=0.5, top=3.0, width=12.333, height=1.2,
        font_size=48, bold=True,
        color=title_color, alignment="center",
        palette="ocean_soft",
    )

    # 副标题
    add_text(
        slide, text=subtitle,
        left=0.5, top=4.3, width=12.333, height=0.6,
        font_size=22, color=subtitle_color, alignment="center",
        palette="ocean_soft",
    )

    save_presentation(prs, output_path)
    print(f"已生成: {output_path}")


def main():
    # 输出目录
    output_dir = os.path.join(current_dir, "examples_output")
    os.makedirs(output_dir, exist_ok=True)

    # 获取所有可用主题
    themes = list_themes()
    print("可用背景主题:")
    for t in themes:
        print(f"  - {t['theme']}: {t['name_cn']} ({len(t['images'])}张图片)")

    print("\n" + "="*60)
    print("生成示例PPT...")
    print("="*60 + "\n")

    # 示例1: 不使用背景
    print("1. 不使用背景")
    generate_simple_slide_with_bg(
        title="通用演示模板",
        subtitle="不使用背景，纯色风格",
        background="",
        output_path=os.path.join(output_dir, "01_no_background.pptx")
    )

    # 示例2: 党政办公
    print("\n2. 党政办公风格")
    generate_simple_slide_with_bg(
        title="党建工作汇报",
        subtitle="2026年第一季度总结",
        background="party_government",
        output_path=os.path.join(output_dir, "02_party_gov.pptx")
    )

    # 示例3: 简约蓝白
    print("\n3. 简约蓝白风格")
    generate_simple_slide_with_bg(
        title="产品发布会",
        subtitle="新一代智能终端引领未来",
        background="minimalist_blue",
        output_path=os.path.join(output_dir, "03_minimalist.pptx")
    )

    # 示例4: 水墨山水
    print("\n4. 水墨山水风格")
    generate_simple_slide_with_bg(
        title="中国传统文化专题",
        subtitle="传承经典，创新未来",
        background="ink_wash_mountain",
        output_path=os.path.join(output_dir, "04_ink_wash.pptx")
    )

    # 示例5: 复古中国风
    print("\n5. 复古中国风")
    generate_simple_slide_with_bg(
        title="古韵新生",
        subtitle="传统文化与现代设计的融合",
        background="vintage_chinese",
        output_path=os.path.join(output_dir, "05_vintage_chinese.pptx")
    )

    # 示例6: 雪山风景
    print("\n6. 雪山风景风格")
    generate_simple_slide_with_bg(
        title="自然之美",
        subtitle="探索大自然的壮丽景观",
        background="snowy_mountain",
        output_path=os.path.join(output_dir, "06_snowy_mountain.pptx")
    )

    print("\n" + "="*60)
    print(f"所有示例已生成到: {output_dir}")
    print("="*60)


if __name__ == "__main__":
    main()
