TEMPLATE = {
    "type": "image_text",
    "name": "图文混排页",
    "description": "灵活组合可替换图片和文字。常见形式：左图右文、右图左文、横向图片条带。未提供图片时自动插入语义相关的本地默认图，用户可在 PowerPoint 中直接更换。",
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
            "image_path": "可选：本地文件路径、photo 素材 id 或 asset:<photo-id>；为空时使用语义默认图",
            "fallback": "必须插入真实 Picture 对象，不得使用图标、纹理面板或占位文案伪装图片"
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
            "paragraph": {
                "font_size": "14pt",
                "char_range": "300-450",
                "line_spacing": "1.5x",
                "description": "用一段自然语言详细描述，避免罗列要点。内容应包含背景、细节、影响等完整信息。"
            }
        }
    },
    "visual_elements": [
        "图片用 rounded rectangle 裁剪，边框 accent 色（2pt）",
        "图片可叠加半透明色块用于文字叠加（image_hero 变体）",
        "文字与图片间距至少 0.3 英寸",
        "图片区域可添加小型几何装饰（斜切角、小圆点）",
        "如果图片留空，自动插入本地默认图，保持为可替换的 PowerPoint Picture 对象",
        "文字区域左侧可有竖线装饰（primary 色，3pt）",
        "图片和文字可使用卡片形式包裹，增加层次感"
    ],
    "example": {
        "kicker": "产品介绍",
        "title": "智能推荐引擎",
        "layout": "right-image",
        "image_path": "asset:photo_technology_device",
        "header": "个性化推荐系统",
        "sub_header": "基于深度学习的实时推荐引擎",
        "paragraph": "该推荐引擎通过深度学习算法实时分析用户行为数据，构建多维度用户画像，实现千人千面的个性化推荐。系统支持秒级响应，能够在用户浏览商品的瞬间完成推荐计算，同时通过A/B测试框架持续优化推荐效果，目前已在电商、内容平台等多个场景验证，显著提升用户点击率和转化率。"
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{案例/产品名称}",
        "layout": "right-image",
        "image_path": "",
        "header": "{核心技术/产品名称}",
        "sub_header": "{副标题}",
        "paragraph": "300-450字的自然语言段落"
    },
    "when_to_use": [
        "案例/产品/技术介绍",
        "需要图片支撑的文字说明",
        "人物/团队/公司介绍"
    ],
    "never": [
        "禁止图片占比过小（小于30%）",
        "禁止图片和文字没有间距",
        "禁止用图标、纹理块、虚线框或内部实现标签代替图片",
        "禁止将段落拆分为要点列表——必须用自然语言段落",
        "禁止段落字数少于300字或多于450字"
    ]
}
