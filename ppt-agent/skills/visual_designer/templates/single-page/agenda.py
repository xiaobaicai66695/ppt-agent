TEMPLATE = {
    "type": "agenda",
    "name": "目录页",
    "description": "展示 PPT 章节结构的目录/大纲页。双栏排列编号章节，左栏 01-03，右栏 04-06。",
    "layout_hint": "two_column_numbered",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "目录",
            "margin_bottom": "0.1in"
        },
        "title": {
            "font_size": "36pt",
            "font_weight": "bold",
            "alignment": "left",
            "text": "内容概览",
            "margin_bottom": "0.3in",
            "max_chars": 20
        },
        "items": {
            "count": "4-8",
            "layout": "two columns（左栏奇数项，右栏偶数项）",
            "item_spec": {
                "number_font_size": "28pt",
                "number_color": "primary",
                "number_font_weight": "bold",
                "title_font_size": "16pt",
                "title_color": "text",
                "spacing": "0.7in",
                "char_per_item_max": 20
            },
            "format": "{序号}  {标题}",
            "example": [
                "01  什么是大模型",
                "02  技术发展历程",
                "03  核心原理",
                "04  能力与特点",
                "05  行业应用",
                "06  未来展望"
            ]
        },
        "progress_indicator": {
            "enabled": False,
            "content": "可选：当前章节高亮标记（如「进行中」标签）"
        }
    },
    "visual_elements": [
        "序号使用 primary 色加粗大字（28pt）",
        "两项之间用 divider 色细线分隔（secondary 色，0.5pt）",
        "双栏布局：左栏奇数项，右栏偶数项",
        "序号和标题之间留有一定间距（0.2in）",
        "每个条目的序号前可添加小圆点或方块装饰（primary 色）",
        "可用不同透明度区分已完成/当前/未开始章节（如已完成的用 primary 100%，未开始的用 50%）",
        "标题右侧可添加小箭头或页码提示（text_muted，12pt）",
        "页面底部可添加装饰性色块（primary 色，透明度 5-10%）"
    ],
    "example": {
        "kicker": "目录",
        "title": "内容概览",
        "items": [
            "01  什么是大模型",
            "02  技术发展历程",
            "03  核心原理与Transformer架构",
            "04  能力与特点",
            "05  行业应用案例",
            "06  未来展望与挑战"
        ]
    },
    "example_2": {
        "kicker": "目录",
        "title": "{目录标题}",
        "items": [
            "01  {章节1}",
            "02  {章节2}",
            "03  {章节3}",
            "04  {章节4}",
            "05  {章节5}",
            "06  {章节6}",
            "07  {章节7}",
            "08  {章节8}"
        ]
    },
    "when_to_use": [
        "作为 PPT 的第二页，给听众建立整体框架",
        "章节数在 4-8 个时效果最佳"
    ],
    "never": [
        "禁止使用纯 bullet 列表——必须用编号+分隔线的结构化布局",
        "禁止条目超过 8 个——超过时分两页或合并章节",
        "禁止标题叫'目录'但没有任何编号/视觉层级"
    ]
}
