TEMPLATE = {
    "type": "brand_focus",
    "name": "品牌价值聚焦",
    "description": "中心聚焦 + 周围价值点布局，适合展示品牌核心价值、企业理念或核心能力。左侧中心区域用多层圆形叠加营造聚焦感，右侧信息面板展示核心理念。",
    "layout_hint": "left_focus + right_panel",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "32pt",
            "font_weight": "bold",
            "alignment": "left",
            "margin_bottom": "0.2in",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "14pt",
            "color": "text_muted",
            "margin_bottom": "0.3in",
            "max_chars": 40
        },
        "center_focus": {
            "layers": [
                {"shape": "circle", "size": "2.5in", "color": "light_bg"},
                {"shape": "circle", "size": "2.0in", "color": "accent"},
                {"shape": "circle", "size": "1.4in", "color": "primary"},
                {"shape": "text", "content": "核心\n理念", "font_size": "14pt", "color": "background", "bold": True}
            ],
            "surrounding_points": [
                {
                    "title": "{价值点1}",
                    "description": "{描述1}",
                    "color": "secondary",
                    "angle": 45
                },
                {
                    "title": "{价值点2}",
                    "description": "{描述2}",
                    "color": "accent",
                    "angle": 135
                },
                {
                    "title": "{价值点3}",
                    "description": "{描述3}",
                    "color": "secondary",
                    "angle": 225
                },
                {
                    "title": "{价值点4}",
                    "description": "{描述4}",
                    "color": "accent",
                    "angle": 315
                }
            ]
        },
        "right_panel": {
            "title": "核心理念",
            "principles": [
                {
                    "number": "01",
                    "title": "{理念1}",
                    "description": "{理念1的说明}"
                },
                {
                    "number": "02",
                    "title": "{理念2}",
                    "description": "{理念2的说明}"
                },
                {
                    "number": "03",
                    "title": "{理念3}",
                    "description": "{理念3的说明}"
                },
                {
                    "number": "04",
                    "title": "{理念4}",
                    "description": "{理念4的说明}"
                }
            ]
        }
    },
    "visual_elements": [
        "左侧中心区域用多层圆形叠加：最外层 light_bg（最大），中层 accent，中间层 primary（最小），营造聚焦感",
        "中心圆形内放核心主题文字，白色加粗居中",
        "周围四个价值点用小椭圆或圆角矩形表示，围绕中心分布",
        "价值点与中心之间用细线连接（divider色，手绘风格）",
        "右侧面板用 background 色背景，顶部 primary 色强调线（3pt）",
        "右侧面板内每条理念用卡片式设计：左侧数字编号（primary色大号），右侧标题+描述",
        "理念之间用细分隔线（divider色）",
        "整体布局左右比例约 55%:45%"
    ],
    "example": {
        "kicker": "品牌战略",
        "title": "品牌价值主张",
        "subtitle": "以用户为中心的核心价值体系",
        "center_text": "用户\n至上",
        "surrounding_points": [
            {"title": "创新", "description": "持续突破技术边界", "color": "secondary"},
            {"title": "品质", "description": "精益求精的产品", "color": "accent"},
            {"title": "服务", "description": "专业贴心的支持", "color": "secondary"},
            {"title": "责任", "description": "可持续发展的承诺", "color": "accent"}
        ],
        "principles": [
            {"title": "以用户为中心", "description": "每一项决策都基于用户真实需求"},
            {"title": "长期主义", "description": "坚持做正确的事，而非容易的事"},
            {"title": "开放创新", "description": "拥抱变化，持续学习和迭代"},
            {"title": "共赢合作", "description": "与伙伴共同成长，共享成果"}
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{品牌标题}",
        "subtitle": "{副标题}",
        "center_text": "{核心\n主题}",
        "surrounding_points": [
            {"title": "{价值1}", "description": "{描述1}", "color": "secondary"},
            {"title": "{价值2}", "description": "{描述2}", "color": "accent"},
            {"title": "{价值3}", "description": "{描述3}", "color": "secondary"},
            {"title": "{价值4}", "description": "{描述4}", "color": "accent"}
        ],
        "principles": [
            {"title": "{理念1}", "description": "{说明1}"},
            {"title": "{理念2}", "description": "{说明2}"},
            {"title": "{理念3}", "description": "{说明3}"},
            {"title": "{理念4}", "description": "{说明4}"}
        ]
    },
    "when_to_use": [
        "品牌价值展示",
        "企业文化介绍",
        "核心能力呈现",
        "企业战略宣讲"
    ],
    "never": [
        "禁止中心圆形过于花哨",
        "禁止周围价值点超过6个",
        "禁止右侧面板内容过长",
        "禁止布局左右比例失调"
    ]
}
