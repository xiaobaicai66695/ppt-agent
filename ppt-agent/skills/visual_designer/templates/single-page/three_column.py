TEMPLATE = {
    "type": "three_column",
    "name": "三栏并列",
    "description": "适合三个维度、三个选项或三个案例的对称排列。",
    "layout_hint": "grid",
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
        "columns": [
            {
                "position": "left_third",
                "width": "30%",
                "header": {
                    "font_size": "18pt",
                    "font_weight": "bold",
                    "color": "primary"
                },
                "header_bg": "light_bg",
                "bullets": {
                    "font_size": "14pt",
                    "items_max": 3,
                    "char_per_item_max": 25
                },
                "icon": {
                    "type": "数字序号或小图标文字描述",
                    "size": "24pt",
                    "color": "primary"
                }
            },
            {
                "position": "center_third",
                "width": "30%",
                "header": {
                    "font_size": "18pt",
                    "font_weight": "bold",
                    "color": "secondary"
                },
                "header_bg": "light_bg",
                "bullets": {
                    "font_size": "14pt",
                    "items_max": 3,
                    "char_per_item_max": 25
                },
                "icon": {
                    "type": "数字序号或小图标文字描述",
                    "size": "24pt",
                    "color": "secondary"
                }
            },
            {
                "position": "right_third",
                "width": "30%",
                "header": {
                    "font_size": "18pt",
                    "font_weight": "bold",
                    "color": "accent"
                },
                "header_bg": "light_bg",
                "bullets": {
                    "font_size": "14pt",
                    "items_max": 3,
                    "char_per_item_max": 25
                },
                "icon": {
                    "type": "数字序号或小图标文字描述",
                    "size": "24pt",
                    "color": "accent"
                }
            }
        ]
    },
    "visual_elements": [
        "每栏顶部用不同深浅的色块作为栏目标题背景（依次：primary淡、secondary淡、accent淡）",
        "三栏可用等宽卡片形式，间距均匀（0.2in）",
        "栏目标题可配合数字序号（如 01 / 02 / 03）置于标题左侧",
        "栏目标题下方用 2pt 强调线（与标题同色）",
        "每栏底部可添加小型装饰元素（如小圆点或短横线，accent 色）",
        "三栏整体可用浅底色背景衬托（light_bg），增加层次感"
    ],
    "example": {
        "kicker": "能力矩阵",
        "title": "产品核心能力",
        "col1": {
            "header": "01 智能问答",
            "bullets": ["多轮对话理解", "上下文记忆", "知识推理"]
        },
        "col2": {
            "header": "02 多模态",
            "bullets": ["文本图像理解", "视频内容分析", "语音交互"]
        },
        "col3": {
            "header": "03 个性化",
            "bullets": ["用户画像构建", "偏好学习", "推荐定制"]
        }
    },
    "example_2": {
        "kicker": "{主题标签}",
        "title": "{主题名称}",
        "col1_header": "01 {维度A}",
        "col1_bullets": ["{要点A1}", "{要点A2}", "{要点A3}"],
        "col2_header": "02 {维度B}",
        "col2_bullets": ["{要点B1}", "{要点B2}", "{要点B3}"],
        "col3_header": "03 {维度C}",
        "col3_bullets": ["{要点C1}", "{要点C2}", "{要点C3}"]
    },
    "when_to_use": [
        "三个平等维度/选项并列",
        "三个案例/行业/产品对比",
        "三角关系（优/劣/中）分析"
    ],
    "never": [
        "禁止某一栏内容明显比其他两栏多很多",
        "禁止三栏不等宽",
        "禁止超过3条要点每栏"
    ]
}
