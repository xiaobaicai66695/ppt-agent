#!/usr/bin/env python3
"""
PPT Generator - Main entry point.

Usage:
    python generator.py

Generates a complete demo PPT using python-pptx.
All slides follow the visual_designer skill guidelines.
"""
from pathlib import Path
import sys

# generators package is at .../visual_designer/generators/
# add the parent (visual_designer/) so 'from generators import ...' works
sys.path.insert(0, str(Path(__file__).parent.parent))

from pptx import Presentation

from generators import (
    PALETTES,
    new_presentation,
    generate_title_slide,
    generate_section_divider,
    generate_content_slide,
    generate_stat_slide,
    generate_quote_slide,
    generate_card_grid,
    generate_timeline,
    generate_process_flow,
    generate_two_column,
    generate_three_column,
    generate_summary_slide,
    generate_image_text,
    generate_example_detail,
    generate_deep_dive,
    generate_agenda,
    generate_case_study,
    generate_kpi_dashboard,
)


def generate_demo_deck(palette: str = "ocean_soft", output_path: str = "output/demo.pptx"):
    """
    Generate a complete demo PPT deck showcasing various slide types.

    Args:
        palette: Color palette name (see PALETTES keys).
        output_path: Output .pptx file path.
    """
    prs = new_presentation(palette=palette)

    # ── Slide 1: Title ────────────────────────────────────────────────
    generate_title_slide(
        prs,
        palette=palette,
        title="{主题名称}",
        subtitle="{副标题}",
        author="{演讲者}",
        date="{日期}",
    )

    # ── Slide 2: Table of Contents ────────────────────────────────────
    generate_agenda(
        prs,
        palette=palette,
        kicker="目录",
        title="内容概览",
        items=[
            "01  {章节1}",
            "02  {章节2}",
            "03  {章节3}",
            "04  {章节4}",
            "05  {章节5}",
            "06  {章节6}",
        ],
    )

    # ── Slide 3: Section Divider ────────────────────────────────────
    generate_section_divider(
        prs, palette=palette,
        number="01", title="{章节1}",
        subtitle="{章节副标题}",
    )

    # ── Slide 4: Definition ──────────────────────────────────────────
    generate_content_slide(
        prs,
        palette=palette,
        title="{主题}的定义",
        section_header="{核心概念}",
        bullets=[
            "{要点1}",
            "{要点2}",
            "{要点3}",
            "{要点4}",
        ],
    )

    # ── Slide 5: Scale Stats ────────────────────────────────────────
    generate_stat_slide(
        prs,
        palette=palette,
        title="{指标标题}",
        stats=[
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
        ],
    )

    # ── Slide 6: Section Divider ────────────────────────────────────
    generate_section_divider(
        prs, palette=palette,
        number="02", title="{章节2}",
        subtitle="{章节副标题}",
    )

    # ── Slide 7: Timeline ───────────────────────────────────────────
    generate_timeline(
        prs,
        palette=palette,
        title="{发展历程标题}",
        direction="horizontal",
        nodes=[
            {"year": "{年份}", "event": "{事件描述}", "icon": "01"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "02"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "03"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "04"},
            {"year": "{年份}", "event": "{事件描述}", "icon": "05"},
        ],
    )

    # ── Slide 8: Section Divider ────────────────────────────────────
    generate_section_divider(
        prs, palette=palette,
        number="03", title="{章节3}",
        subtitle="{章节副标题}",
    )

    # ── Slide 9: Core Concept ────────────────────────────────────────
    generate_content_slide(
        prs,
        palette=palette,
        title="{核心概念}",
        section_header="{子标题}",
        bullets=[
            "{要点1}",
            "{要点2}",
            "{要点3}",
            "{要点4}",
        ],
    )

    # ── Slide 10: Capabilities ────────────────────────────────────────
    generate_card_grid(
        prs,
        palette=palette,
        title="{能力标题}",
        layout="2x2",
        cards=[
            {"header": "{能力1}", "body": "{能力描述}"},
            {"header": "{能力2}", "body": "{能力描述}"},
            {"header": "{能力3}", "body": "{能力描述}"},
            {"header": "{能力4}", "body": "{能力描述}"},
        ],
    )

    # ── Slide 11: Section Divider ───────────────────────────────────
    generate_section_divider(
        prs, palette=palette,
        number="04", title="{章节4}",
        subtitle="{章节副标题}",
    )

    # ── Slide 12: Core Capabilities ────────────────────────────────
    generate_card_grid(
        prs,
        palette=palette,
        title="{能力标题}",
        layout="2x2",
        cards=[
            {"header": "{能力1}", "body": "{能力描述}"},
            {"header": "{能力2}", "body": "{能力描述}"},
            {"header": "{能力3}", "body": "{能力描述}"},
            {"header": "{能力4}", "body": "{能力描述}"},
        ],
    )

    # ── Slide 13: Section Divider ───────────────────────────────────
    generate_section_divider(
        prs, palette=palette,
        number="05", title="{章节5}",
        subtitle="{章节副标题}",
    )

    # ── Slide 14: Industry Cases ───────────────────────────────────
    generate_three_column(
        prs,
        palette=palette,
        title="{案例标题}",
        columns=[
            {
                "header": "01  {领域1}",
                "bullets": ["{要点1}", "{要点2}", "{要点3}"],
            },
            {
                "header": "02  {领域2}",
                "bullets": ["{要点1}", "{要点2}", "{要点3}"],
            },
            {
                "header": "03  {领域3}",
                "bullets": ["{要点1}", "{要点2}", "{要点3}"],
            },
        ],
    )

    # ── Slide 15: Application Metrics ───────────────────────────────
    generate_stat_slide(
        prs,
        palette=palette,
        title="{效果数据标题}",
        stats=[
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
            {"number": "{数字}", "unit": "{单位}", "label": "{指标说明}"},
        ],
    )

    # ── Slide 16: Section Divider ───────────────────────────────────
    generate_section_divider(
        prs, palette=palette,
        number="06", title="{章节6}",
        subtitle="{章节副标题}",
    )

    # ── Slide 17: Development Trends ───────────────────────────────
    generate_process_flow(
        prs,
        palette=palette,
        title="{趋势标题}",
        direction="horizontal_zigzag",
        steps=[
            {"num": "01", "title": "{趋势1}", "desc": "{描述}"},
            {"num": "02", "title": "{趋势2}", "desc": "{描述}"},
            {"num": "03", "title": "{趋势3}", "desc": "{描述}"},
            {"num": "04", "title": "{趋势4}", "desc": "{描述}"},
            {"num": "05", "title": "{趋势5}", "desc": "{描述}"},
            {"num": "06", "title": "{趋势6}", "desc": "{描述}"},
        ],
    )

    # ── Slide 18: Example Detail ──────────────────────────────────────
    generate_example_detail(
        prs,
        palette=palette,
        kicker="实例 · {领域}",
        title="{案例名称}: {一句话总结}",
        lede="{核心数据或价值}",
        context_block="{背景描述}",
        solution_block="{解决方案}",
        metrics=[
            {"value": "{数值}", "label": "{指标}", "trend": "{趋势}"},
            {"value": "{数值}", "label": "{指标}", "trend": "{趋势}"},
            {"value": "{数值}", "label": "{指标}", "trend": "{趋势}"},
        ],
        takeaway="{启示}",
    )

    # ── Slide 19: Deep Dive ─────────────────────────────────────────
    generate_deep_dive(
        prs,
        palette=palette,
        kicker="详解 · {概念}",
        title="{技术名称}",
        lede="{一句话总结核心价值}",
        steps=[
            "{步骤1}",
            "{步骤2}",
            "{步骤3}",
            "{步骤4}",
            "{步骤5}",
        ],
        design_decisions=[
            "{设计决策1}",
            "{设计决策2}",
        ],
        comparisons=[
            "{对比说明}",
        ],
        code_block=(
            "# {伪代码/示例代码}\n"
            "def example():\n"
            "    pass"
        ),
        architecture_text=(
            "{组件1} → {组件2} → {组件3}"
        ),
        performance_data=[
            "{性能指标1}",
            "{性能指标2}",
            "{性能指标3}",
        ],
    )

    # ── Slide 20: Case Study ──────────────────────────────────────────
    generate_case_study(
        prs,
        palette=palette,
        kicker="案例 · {领域}",
        title="{案例名称}",
        context="{背景}",
        problem="{痛点}",
        solution="{解决方案}",
        results=[
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
            {"metric": "{指标}", "value": "{数值}", "comparison": "{对比}"},
        ],
    )

    # ── Slide 21: KPI Dashboard ──────────────────────────────────────
    generate_kpi_dashboard(
        prs,
        palette=palette,
        kicker="数据 · {效果}",
        title="{指标标题}",
        kpis=[
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
            {"value": "{数值}", "label": "{说明}", "delta": "{趋势}", "baseline": "{基准}"},
        ],
    )

    # ── Slide 22: Quote ─────────────────────────────────────────────
    generate_quote_slide(
        prs,
        palette=palette,
        quote="{金句}",
        attribution="{来源}",
    )

    # ── Slide 23: Summary ───────────────────────────────────────────
    generate_summary_slide(
        prs,
        palette=palette,
        title="核心要点回顾",
        key_points=[
            "01 {要点1}",
            "02 {要点2}",
            "03 {要点3}",
            "04 {要点4}",
        ],
        thank_you="感谢聆听",
        contact="{联系方式}",
    )

    # Save
    prs.save(output_path)
    print(f"Saved: {output_path}")
    return prs


if __name__ == "__main__":
    import os
    os.makedirs("output", exist_ok=True)

    # Generate with default palette (ocean_soft)
    generate_demo_deck(palette="ocean_soft", output_path="output/demo_ocean.pptx")

    # Also generate a warm palette variant
    generate_demo_deck(palette="warm_terracotta", output_path="output/demo_warm.pptx")
    print("Done.")
