TEMPLATE = {
    "type": "timeline",
    "name": "时间轴",
    "description": "水平或垂直时间轴，按时间点标记事件。适合发展历程、项目里程碑等。",
    "layout_hint": "horizontal",
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
        "timeline_axis": {
            "direction": "horizontal 或 vertical",
            "line_color": "primary",
            "line_width": "2pt",
            "line_style": "实线"
        },
        "nodes": [
            {
                "position": "节点位置（轴上方或下方交替）",
                "year": "时间标签（14pt，次要色）",
                "event": "事件描述（16pt，正文色）",
                "icon": "可选：数字序号或图标（24pt）"
            }
        ],
        "nodes_max": 5,
        "spacing": "节点之间均匀分布",
        "connector": "节点与轴之间用细线连接（secondary 色，1pt）"
    },
    "visual_elements": [
        "时间轴主线用 primary 色，2-3pt 粗线",
        "节点用 accent 色圆点（半径 0.15in）或圆角矩形",
        "节点内部可放数字序号（白色，bold）",
        "时间标签在轴上方或下方交替排列，避免文字重叠",
        "节点之间可用虚线或箭头连接（secondary 色，1pt）",
        "时间轴两端可添加装饰性小圆点（primary 色，透明度 30%）",
        "每个节点下方/上方可用细线（1pt）连接至时间轴",
        "整体背景可添加极淡的网格线（secondary 色，透明度 5%）衬托",
        "可用不同颜色区分不同类型的时间节点（如：突破性事件用 accent 色，常规事件用 secondary 色）"
    ],
    "example": {
        "kicker": "技术演进",
        "title": "AI技术发展里程碑",
        "subtitle": "从深度学习到大模型的时代跨越",
        "direction": "horizontal",
        "nodes": [
            {
                "year": "2012",
                "event": "AlexNet\nImageNet突破",
                "icon": "01",
                "highlight": True
            },
            {
                "year": "2017",
                "event": "Transformer\n注意力机制",
                "icon": "02",
                "highlight": True
            },
            {
                "year": "2020",
                "event": "GPT-3\n超大模型涌现",
                "icon": "03",
                "highlight": True
            },
            {
                "year": "2023",
                "event": "GPT-4\n多模态融合",
                "icon": "04",
                "highlight": False
            },
            {
                "year": "2025",
                "event": "Agent\n自主执行",
                "icon": "05",
                "highlight": False
            }
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{主题名称}",
        "subtitle": "{副标题}",
        "direction": "horizontal",
        "nodes": [
            {"year": "{年份1}", "event": "{事件1}", "icon": "01"},
            {"year": "{年份2}", "event": "{事件2}", "icon": "02"},
            {"year": "{年份3}", "event": "{事件3}", "icon": "03"}
        ]
    },
    "when_to_use": [
        "有明确时间顺序的发展历程",
        "项目里程碑展示",
        "技术演进路线"
    ],
    "never": [
        "禁止超过5个时间节点",
        "禁止时间标签字号过小（小于12pt）",
        "禁止时间轴线太细（小于1.5pt）",
        "禁止节点之间间距不均匀"
    ]
}
