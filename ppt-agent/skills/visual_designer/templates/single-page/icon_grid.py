TEMPLATE = {
    "type": "icon_grid",
    "name": "图标网格页",
    "description": "用图标+文字网格展示核心要点，视觉冲击力强。适合展示功能特性、服务内容、方法步骤等。6-9个图标均匀分布。",
    "layout_hint": "grid (2x3, 3x3, 3x2)",
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
            "margin_bottom": "0.15in",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "14pt",
            "color": "text_muted",
            "margin_bottom": "0.25in",
            "max_chars": 40
        },
        "icons": {
            "count": "6-9",
            "icon_style": "圆形或圆角方形背景 + 图标文字（如「研」「创」「智」等单字）或 emoji 风格文字",
            "icon_size": "0.8-1.0in",
            "label_below": "图标下方文字说明",
            "icon_colors": "使用 palette 的 primary/secondary/accent 循环"
        }
    },
    "visual_elements": [
        "图标用圆形或圆角矩形背景（light_bg 色）",
        "图标内部用大号加粗单字或文字符号（primary 色）",
        "图标下方有简短文字标签（text 色，14pt）",
        "图标网格均匀分布，间距 0.3-0.4in",
        "每个图标周围有足够的留白空间",
        "可交替使用不同大小的图标（中心大、边缘小）增加层次感",
        "标题区左侧有竖线装饰（primary 色，3pt）",
        "可选：图标之间用细线连接形成网络感"
    ],
    "example": {
        "kicker": "核心能力",
        "title": "六大技术支柱",
        "subtitle": "构建完整AI技术体系",
        "layout": "3x2",
        "icons": [
            {"icon": "研", "label": "基础研究", "color": "primary"},
            {"icon": "算", "label": "算力平台", "color": "secondary"},
            {"icon": "数", "label": "数据治理", "color": "accent"},
            {"icon": "模", "label": "模型训练", "color": "primary"},
            {"icon": "工", "label": "工程落地", "color": "secondary"},
            {"icon": "安", "label": "安全合规", "color": "accent"}
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{图标网格标题}",
        "subtitle": "{副标题}",
        "layout": "3x2 或 3x3",
        "icons": [
            {"icon": "{单字1}", "label": "{说明1}", "color": "primary"},
            {"icon": "{单字2}", "label": "{说明2}", "color": "secondary"},
            {"icon": "{单字3}", "label": "{说明3}", "color": "accent"},
            {"icon": "{单字4}", "label": "{说明4}", "color": "primary"},
            {"icon": "{单字5}", "label": "{说明5}", "color": "secondary"},
            {"icon": "{单字6}", "label": "{说明6}", "color": "accent"}
        ]
    },
    "when_to_use": [
        "展示多个同等重要的事项",
        "功能/服务/特性概览",
        "方法/步骤/要素展示",
        "需要强视觉冲击的场合"
    ],
    "never": [
        "禁止图标超过9个",
        "禁止图标太小（小于0.6in）",
        "禁止图标之间没有间距",
        "禁止图标文字过长",
        "禁止图标颜色混乱（保持规律循环）"
    ]
}
