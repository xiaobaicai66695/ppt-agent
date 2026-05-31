TEMPLATE = {
    "type": "chart_slide",
    "name": "图表专页",
    "description": "专业数据可视化页面，支持柱状图、饼图、折线图、环形图等多种图表类型。适合展示数据对比、趋势变化、占比分析等场景。",
    "layout_hint": "chart + legend + optional analysis panel",
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
        "chart": {
            "types": ["bar", "pie", "line", "doughnut", "stacked_bar"],
            "bar_direction": "vertical 或 horizontal",
            "colors": "使用 palette 中的 primary/secondary/accent 循环",
            "data_labels": "显示数值标签在柱子/扇区上",
            "legend_position": "底部 或 右侧",
            "axis_config": "Y轴自适应范围，避免刻度过密"
        },
        "chart_data": {
            "labels": ["{标签1}", "{标签2}", "{标签3}"],
            "datasets": [
                {
                    "name": "{数据集名称}",
                    "values": [100, 200, 300],
                    "color": "primary"
                }
            ]
        },
        "analysis": {
            "position": "图表右侧（当提供时）",
            "width": "页面35%",
            "background": "light_bg 色圆角矩形",
            "header": "数据分析（14pt加粗，primary色）",
            "content": "分析文字（11pt）",
            "max_chars": 300
        }
    },
    "visual_elements": [
        "图表区域占页面 60-70%（有分析时）或 100%（无分析时）",
        "图表使用 primary/secondary/accent 三色循环填充",
        "图例位于图表下方或右侧，使用小色块+文字",
        "数据标签紧贴图表元素（柱顶、扇区外）",
        "网格线使用 divider 色（0.5pt），保持简洁",
        "标题区左侧有竖线装饰（primary 色，3pt）",
        "有analysis时：图表右侧显示数据分析面板（圆角矩形，light_bg背景）",
        "数据分析面板包含标题+分割线+分析内容",
        "背景使用 background 色，保持干净"
    ],
    "example": {
        "kicker": "数据分析",
        "title": "季度营收对比",
        "subtitle": "2025年各季度营收数据一览",
        "chart_type": "bar",
        "data": {
            "labels": ["Q1", "Q2", "Q3", "Q4"],
            "datasets": [
                {"name": "2025年", "values": [1200, 1500, 1800, 2200]},
                {"name": "2024年", "values": [900, 1100, 1300, 1600]}
            ]
        },
        "analysis": "• Q4表现最佳，营收达2200万，同比增长37.5%\n• 2025年整体呈上升趋势，环比增长稳健\n• 下半年增速明显加快，建议关注持续性"
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{图表标题}",
        "subtitle": "{副标题}",
        "chart_type": "bar 或 pie 或 line",
        "data": {
            "labels": ["{类别1}", "{类别2}", "{类别3}", "{类别4}"],
            "datasets": [
                {"name": "{系列1}", "values": [100, 200, 300, 400]},
                {"name": "{系列2}", "values": [80, 150, 250, 350]}
            ]
        },
        "analysis": "{可选的数据分析文字}"
    },
    "when_to_use": [
        "数据对比分析",
        "趋势变化展示",
        "占比分析",
        "多维度数据呈现"
    ],
    "never": [
        "禁止图表过于复杂（超过3个数据系列）",
        "禁止数据标签遮挡图表",
        "禁止缺少图例说明",
        "禁止使用过多颜色（保持3色以内）",
        "禁止图表占比过小"
    ]
}
