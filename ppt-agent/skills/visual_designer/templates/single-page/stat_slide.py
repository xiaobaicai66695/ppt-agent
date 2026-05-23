TEMPLATE = {
    "type": "stat_slide",
    "name": "关键数字页",
    "description": "把一个或几个超高亮度数字放大居中，配合简短说明，冲击力强。适合展示核心指标、里程碑数据。",
    "layout_hint": "center",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "36pt",
            "font_weight": "bold",
            "alignment": "left",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "16pt",
            "color": "text_muted",
            "margin_bottom": "0.3in",
            "max_chars": 40
        },
        "stats": [
            {
                "number": "数字（60-72pt，超大加粗）",
                "unit": "单位（可选，如 %、倍、次）",
                "label": "简短说明（14pt，次要色）",
                "trend": "可选：变化趋势（↑ 30% 或 ↓ 50%）",
                "icon": "可选：数字前的小图标文字描述"
            }
        ],
        "stats_max": 3,
        "arrangement": "横向排列或纵向堆叠",
        "highlight_stat": {
            "enabled": True,
            "index": 0,
            "color": "accent"
        }
    },
    "visual_elements": [
        "数字用 primary 色或 accent 色，超大字号（60-72pt）",
        "数字下方用细线分隔（颜色 secondary，1pt，宽度为数字宽度）",
        "背景可用浅色色块衬托数字区域（light_bg）",
        "整体居中或偏左均可，留足上下留白",
        "可选：数字周围添加装饰性圆环（primary 色，透明度 10%）",
        "数字之间可用竖线分隔（secondary 色，1pt）",
        "每个数字下方可配合趋势标签（如「↑ 30%」用 secondary 色，「↓ 50%」用 accent 色）",
        "底部可添加数据来源注释（text_muted，10pt）",
        "可选：底部整体进度条（展示指标完成度）和汇总统计"
    ],
    "example": {
        "kicker": "年度成果",
        "title": "年度核心成果",
        "subtitle": "2025财年关键数据一览",
        "stats": [
            {
                "number": "99.99",
                "unit": "%",
                "label": "系统可用性",
                "trend": "↑ 0.3%"
            },
            {
                "number": "3.2",
                "unit": "倍",
                "label": "性能提升",
                "trend": "↑ 220%"
            },
            {
                "number": "500",
                "unit": "万+",
                "label": "服务用户数",
                "trend": "↑ 40%"
            }
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{主题名称}",
        "subtitle": "{副标题}",
        "stats": [
            {"number": "{数字1}", "unit": "{单位1}", "label": "{说明1}", "trend": "{趋势1}"},
            {"number": "{数字2}", "unit": "{单位2}", "label": "{说明2}", "trend": "{趋势2}"},
            {"number": "{数字3}", "unit": "{单位3}", "label": "{说明3}", "trend": "{趋势3}"}
        ]
    },
    "when_to_use": [
        "需要强视觉冲击的核心数据展示",
        "年度总结、里程碑汇报",
        "产品发布核心指标"
    ],
    "never": [
        "禁止数字字号小于48pt",
        "禁止超过3个大数字",
        "禁止数字没有足够留白衬托",
        "禁止数字下方说明文字字号大于16pt"
    ]
}
