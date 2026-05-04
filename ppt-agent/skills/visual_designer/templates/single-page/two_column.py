TEMPLATE = {
    "type": "two_column",
    "name": "双栏对比",
    "description": "常见于 A vs B 分析，左右并置。可用于方案对比、功能对比、时间段对比等。",
    "layout_hint": "split",
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
        "left_column": {
            "position": "left_half",
            "width": "45%",
            "header": {
                "font_size": "20pt",
                "font_weight": "bold",
                "color": "primary"
            },
            "header_bg": "light_bg",
            "bullets": {
                "font_size": "14pt",
                "items_max": 4,
                "char_per_item_max": 25,
                "bullet_style": "实心点或短横线"
            },
            "highlight_box": {
                "enabled": True,
                "label": "推荐",
                "color": "secondary"
            }
        },
        "right_column": {
            "position": "right_half",
            "width": "45%",
            "header": {
                "font_size": "20pt",
                "font_weight": "bold",
                "color": "accent"
            },
            "header_bg": "light_bg",
            "bullets": {
                "font_size": "14pt",
                "items_max": 4,
                "char_per_item_max": 25,
                "bullet_style": "实心点或短横线"
            }
        },
        "divider": "两栏之间可用细线（secondary 色，1pt）或留白分隔（0.3in）"
    },
    "visual_elements": [
        "两栏顶部用色块条作为栏目标题背景（left: primary 色淡底，right: accent 色淡底）",
        "栏目标题下方用 2pt 强调线（与标题同色）",
        "可选：在某一栏右上角添加「推荐」标签徽章（secondary 色背景，白色文字）",
        "两栏背景均可使用 light_bg 浅底色，栏间留白 0.3in",
        "底部可添加小型装饰几何图形（primary 色，透明度 10%）",
        "如果对比的是时间维度，可在顶部添加小图标（时钟、箭头等文字描述）"
    ],
    "example": {
        "kicker": "方案对比",
        "title": "技术方案选型",
        "left_header": "传统方案",
        "left_bullets": [
            "部署周期 5-7 天",
            "运维人力 10 人",
            "故障恢复 > 2 小时",
            "扩容需提前申请"
        ],
        "left_highlight": "推荐",
        "right_header": "云原生方案",
        "right_bullets": [
            "部署周期 < 1 天",
            "运维人力 2 人",
            "故障恢复 < 5 分钟",
            "弹性扩容秒级响应"
        ]
    },
    "example_2": {
        "kicker": "{对比主题标签}",
        "title": "{对比主题}",
        "left_header": "{方案A名称}",
        "left_bullets": [
            "{方案A特点1}",
            "{方案A特点2}",
            "{方案A特点3}",
            "{方案A特点4}"
        ],
        "right_header": "{方案B名称}",
        "right_bullets": [
            "{方案B特点1}",
            "{方案B特点2}",
            "{方案B特点3}",
            "{方案B特点4}"
        ]
    },
    "when_to_use": [
        "A vs B 类对比分析",
        "两个方案/产品/观点并列呈现",
        "时间维度前后对比"
    ],
    "never": [
        "禁止两栏内容完全不对等（一边很多一边很少）",
        "禁止不用色块区分两栏，导致视觉模糊",
        "禁止每栏超过4条要点"
    ]
}
