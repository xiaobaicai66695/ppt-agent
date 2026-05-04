TEMPLATE = {
    "type": "image_text",
    "name": "图文混排页",
    "description": "灵活组合图片和文字。常见形式：左图右文、右图左文、上图下文。可用于案例展示、产品介绍等。图片区域可留空让用户自行填入。",
    "layout_hint": "left-image 或 right-image 或 top-image",
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
        "image": {
            "position": "左侧60% 或 右侧40% 或 顶部50%",
            "aspect_ratio": "16:9 或 4:3 或 1:1",
            "border_radius": "0.1in（可选）",
            "placeholder": "【图片占位】留空让用户自行填入，标注如「[请插入产品截图]」「[请插入架构图]」「[请插入团队照片]」"
        },
        "text_content": {
            "position": "与图片对侧",
            "header": {
                "font_size": "20pt",
                "font_weight": "bold",
                "color": "primary",
                "max_chars": 20
            },
            "sub_header": {
                "font_size": "14pt",
                "color": "secondary",
                "max_chars": 30
            },
            "bullets": {
                "font_size": "14pt",
                "items_max": 4,
                "char_per_item_max": 25,
                "line_spacing": "1.4x"
            }
        }
    },
    "visual_elements": [
        "图片用 rounded rectangle 裁剪，边框 accent 色（2pt）",
        "图片可叠加半透明色块用于文字叠加（image_hero 变体）",
        "文字与图片间距至少 0.3 英寸",
        "图片区域可添加小型几何装饰（斜切角、小圆点）",
        "如果图片留空，用浅色虚线边框标注占位区域，内含文字说明期望的图片内容",
        "文字区域左侧可有竖线装饰（primary 色，3pt）",
        "图片和文字可使用卡片形式包裹，增加层次感"
    ],
    "example": {
        "kicker": "产品介绍",
        "title": "智能推荐引擎",
        "layout": "right-image",
        "image_placeholder": "[请插入产品界面截图]",
        "header": "个性化推荐系统",
        "sub_header": "基于深度学习的实时推荐引擎",
        "bullets": [
            "实时行为分析，秒级响应",
            "多维度用户画像，精准匹配",
            "A/B测试框架，效果可量化",
            "支持千万级商品毫秒级推荐"
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{案例/产品名称}",
        "layout": "right-image",
        "image_placeholder": "[请插入{图片内容描述}]",
        "header": "{核心技术/产品名称}",
        "sub_header": "{副标题}",
        "bullets": [
            "{优势1}",
            "{优势2}",
            "{优势3}",
            "{优势4}"
        ]
    },
    "when_to_use": [
        "案例/产品/技术介绍",
        "需要图片支撑的文字说明",
        "人物/团队/公司介绍"
    ],
    "never": [
        "禁止图片占比过小（小于30%）",
        "禁止图片和文字没有间距",
        "禁止文字超过4条要点"
    ]
}
