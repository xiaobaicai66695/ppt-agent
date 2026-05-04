TEMPLATE = {
    "type": "summary_slide",
    "name": "总结页",
    "description": "回顾核心要点 + 感谢语。比 thank_you 更重内容，结尾页首选。",
    "layout_hint": "left",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "总结",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "36pt",
            "font_weight": "bold",
            "alignment": "left",
            "max_chars": 30
        },
        "key_points": {
            "font_size": "16pt",
            "items_max": 4,
            "char_per_item_max": 30,
            "bullet_style": "加粗序号 + 要点文字",
            "content_placeholder": [
                "01 {核心要点1}",
                "02 {核心要点2}",
                "03 {核心要点3}",
                "04 {核心要点4}"
            ]
        },
        "divider": {
            "style": "细线",
            "color": "secondary",
            "width": "1pt",
            "position": "above_thank_you"
        },
        "thank_you": {
            "position": "bottom_center",
            "font_size": "24pt",
            "font_weight": "bold",
            "color": "primary"
        },
        "contact": {
            "position": "below_thank_you",
            "font_size": "12pt",
            "color": "secondary",
            "content_placeholder": "联系邮箱: {邮箱} | {其他联系方式}"
        },
        "qr_code": {
            "enabled": False,
            "position": "bottom_right",
            "content": "可选：二维码占位「[请插入群聊/官网二维码]」",
            "size": "1in x 1in"
        }
    },
    "visual_elements": [
        "总结要点可用左侧竖线色块（primary 色，3pt）作为视觉锚点",
        "要点序号用大号字体（20pt）加粗 primary 色",
        "感谢语区域可用浅色色块背景衬托（light_bg）",
        "底部角落可添加 Logo 或联系方式",
        "整体背景可添加装饰性小圆点（accent 色，透明度 5-10%）",
        "可用色块从左下角延伸到右上角作为装饰（primary 色，透明度 5%）"
    ],
    "example": {
        "kicker": "总结",
        "title": "核心要点回顾",
        "key_points": [
            "01 深度学习从CNN到Transformer持续演进",
            "02 Transformer通过自注意力实现大一统",
            "03 大模型涌现能力推动AI应用爆发",
            "04 行业落地从单点突破走向系统融合"
        ],
        "thank_you": "感谢聆听",
        "contact": "联系邮箱: tech@company.com | 公众号: AI研究院"
    },
    "example_2": {
        "kicker": "总结",
        "title": "{总结标题}",
        "key_points": [
            "01 {核心要点1}",
            "02 {核心要点2}",
            "03 {核心要点3}",
            "04 {核心要点4}"
        ],
        "thank_you": "感谢聆听",
        "contact": "{联系方式}"
    },
    "when_to_use": [
        "PPT结尾页",
        "需要总结核心内容的场合"
    ],
    "never": [
        "禁止总结要点超过4条",
        "禁止感谢语字号小于20pt",
        "禁止总结页没有感谢语",
        "禁止在总结页加入新的大段内容"
    ]
}
