TEMPLATE = {
    "type": "kpi_dashboard",
    "name": "KPI 仪表盘页",
    "description": "展示一组关键绩效指标的仪表盘视图，4 个 KPI 卡片 2x2 排列。每卡片含大数值、指标名、趋势箭头、对比基准。",
    "layout_hint": "2x2",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "数据 · {主题标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "font_size": "30pt",
            "font_weight": "bold",
            "alignment": "left",
            "text": "{仪表盘标题}",
            "margin_bottom": "0.3in",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "14pt",
            "color": "text_muted",
            "margin_bottom": "0.25in",
            "content_placeholder": "{可选：补充说明，如「2025财年Q3数据」}"
        },
        "kpi_cards": {
            "count": "4",
            "layout": "2x2 grid",
            "card_spec": {
                "width": "flexible",
                "height": "1.5in",
                "padding": "0.15in",
                "value_font_size": "40pt",
                "value_color": "primary",
                "value_font_weight": "bold",
                "label_font_size": "12pt",
                "label_color": "text",
                "delta_font_size": "13pt",
                "delta_up_color": "secondary",
                "delta_down_color": "accent",
                "baseline_font_size": "10pt",
                "baseline_color": "text_muted",
                "card_bg": "light_bg",
                "card_radius": "0.1in"
            },
            "item_format": {
                "value": "核心数值（如 99.99%）",
                "label": "指标名称（如 模型准确率）",
                "delta": "变化趋势+幅度（如 ↑ 0.3% 或 ↓ 50%）",
                "baseline": "对比基准（如 vs 旧方案 95%）",
                "icon": "可选：指标类型图标文字描述（如「用户」「收入」「性能」）"
            }
        },
        "footer": {
            "enabled": False,
            "content": "可选：数据来源和时间（如「数据来源：内部统计 | 截至 2025年12月」）",
            "font_size": "10pt",
            "color": "text_muted"
        }
    },
    "visual_elements": [
        "每张 KPI 卡片使用 light_bg 浅底色，圆角边框（radius 0.1in）",
        "卡片顶部 primary 色强调线（3pt）",
        "数值使用 primary 色加粗大字体（40pt）",
        "趋势向上（↑）用 secondary 色，趋势向下（↓）用 accent/warm 色",
        "对比基准用小字 text_muted 色置于卡片底部",
        "四张卡片整体形成 2x2 网格，间距均匀（0.2in）",
        "整体背景可添加极淡的网格线（secondary 色，透明度 3-5%）衬托仪表盘感",
        "如果数据为同比/环比变化，可在数值旁边添加小箭头图标",
        "底部可添加数据来源注释（text_muted，10pt）"
    ],
    "example": {
        "kicker": "数据 · 年度总结",
        "title": "2025年度核心指标",
        "subtitle": "业务线关键绩效数据",
        "kpis": [
            {
                "value": "99.99%",
                "label": "系统可用性",
                "delta": "↑ 0.3%",
                "baseline": "vs 2024: 99.96%"
            },
            {
                "value": "5.2亿",
                "label": "年活跃用户",
                "delta": "↑ 40%",
                "baseline": "vs 2024: 3.7亿"
            },
            {
                "value": "¥12亿",
                "label": "年度营收",
                "delta": "↑ 28%",
                "baseline": "vs 2024: ¥9.4亿"
            },
            {
                "value": "8.6分",
                "label": "客户满意度",
                "delta": "↑ 0.5",
                "baseline": "vs 2024: 8.1分"
            }
        ],
        "footer": "数据来源：内部统计 | 截至 2025年12月"
    },
    "example_2": {
        "kicker": "数据 · {主题标签}",
        "title": "{仪表盘标题}",
        "subtitle": "{补充说明}",
        "kpis": [
            {"value": "{数值}", "label": "{指标名称}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
            {"value": "{数值}", "label": "{指标名称}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
            {"value": "{数值}", "label": "{指标名称}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
            {"value": "{数值}", "label": "{指标名称}", "delta": "{变化趋势}", "baseline": "{对比基准}"}
        ]
    },
    "when_to_use": [
        "展示多项目/多业务线的效果数据对比",
        "季度/年度汇报中的关键指标汇总",
        "需要让听众快速抓住 3-6 个核心数据时"
    ],
    "never": [
        "禁止不是 4 个指标——必须是 2x2 排列，4 个 KPI",
        "禁止缺少 delta 或 baseline——每个 KPI 必须有对比参照",
        "禁止指标之间无关联——KPI 必须围绕同一主题"
    ]
}
