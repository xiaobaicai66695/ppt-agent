TEMPLATE = {
    "type": "comparison_table",
    "name": "对比表格页",
    "description": "多项目多维度对比表格，清晰展示不同方案/产品/选项的优劣。适合选型对比、功能对比、方案对比等场景。",
    "layout_hint": "table with header",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "30pt",
            "font_weight": "bold",
            "alignment": "left",
            "margin_bottom": "0.2in",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "14pt",
            "color": "text_muted",
            "margin_bottom": "0.25in",
            "max_chars": 40
        },
        "table": {
            "header_row": {
                "columns": ["{对比项}", "{方案A}", "{方案B}", "{方案C}"],
                "bg_color": "primary",
                "text_color": "background"
            },
            "data_rows": [
                {
                    "cells": ["{指标1}", "{A表现}", "{B表现}", "{C表现}"],
                    "highlight": True  # 是否高亮该行
                }
            ],
            "footer_row": {
                "cells": ["综合评分", "{评分A}", "{评分B}", "{评分C}"],
                "bg_color": "light_bg"
            }
        }
    },
    "visual_elements": [
        "表格标题行使用 primary 色背景，白色加粗字",
        "数据行交替使用 background 色和 light_bg 色（斑马纹）",
        "需要高亮的行使用浅 accent 色背景",
        "单元格内容居中对齐，表头居中对齐",
        "表格边框使用 divider 色（0.5pt），保持简洁",
        "行高适中（约 0.5in），保证内容可读性",
        "表格下方可添加小结或推荐结论",
        "标题区左侧有竖线装饰（primary 色，3pt）"
    ],
    "example": {
        "kicker": "选型对比",
        "title": "AI平台选型对比",
        "subtitle": "三大云厂商AI能力全面对比",
        "table": {
            "headers": ["对比维度", "AWS SageMaker", "Google Vertex AI", "Azure ML"],
            "rows": [
                ["模型训练速度", "快", "最快", "中等"],
                ["价格", "较高", "中等", "较低"],
                ["易用性", "中等", "简单", "简单"],
                ["生态丰富度", "丰富", "一般", "一般"],
                ["中国区支持", "有限", "有限", "良好"]
            ],
            "recommendation": "综合考虑，建议选择 Azure ML"
        }
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{对比表格标题}",
        "subtitle": "{副标题}",
        "table": {
            "headers": ["{维度1}", "{选项A}", "{选项B}", "{选项C}"],
            "rows": [
                ["{指标1}", "{A}", "{B}", "{C}"],
                ["{指标2}", "{A}", "{B}", "{C}"],
                ["{指标3}", "{A}", "{B}", "{C}"]
            ],
            "recommendation": "{推荐结论}"
        }
    },
    "when_to_use": [
        "方案/产品选型对比",
        "功能特性对比",
        "供应商/服务商对比",
        "前后对比"
    ],
    "never": [
        "禁止对比项超过6个",
        "禁止对比选项超过4个",
        "禁止单元格内容过于冗长",
        "禁止表格没有边框或边框过粗",
        "禁止缺少表头"
    ]
}
