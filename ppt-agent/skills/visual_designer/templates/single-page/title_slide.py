TEMPLATE = {
    "type": "title_slide",
    "name": "标题页",
    "description": "开场标题页，核心信息 + 视觉冲击，留白要大，标题字体要有重量感。",
    "layout_hint": "center",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{开场标签/领域标签}",
            "margin_bottom": "0.15in",
            "position": "above_title"
        },
        "title": {
            "position": "center",
            "font_size": "44pt",
            "font_weight": "bold",
            "alignment": "center",
            "line_spacing": "1.2x",
            "max_chars": 30
        },
        "subtitle": {
            "position": "below_title",
            "font_size": "20pt",
            "alignment": "center",
            "color": "secondary",
            "line_spacing": "1.4x",
            "max_chars": 40
        },
        "author": {
            "position": "bottom_left",
            "font_size": "14pt",
            "color": "text_muted",
            "margin_left": "0.5in",
            "margin_bottom": "0.4in"
        },
        "date": {
            "position": "bottom_right",
            "font_size": "12pt",
            "color": "text_muted",
            "margin_right": "0.5in",
            "margin_bottom": "0.4in"
        }
    },
    "visual_elements": [
        "底部或角落的几何色块装饰（primary 色，半透明 20-30%，占页面 15-20%）",
        "大面积浅色背景（background 色）",
        "左侧竖向色条（primary 色，3pt 宽，从顶部延伸至底部，作为装饰锚点）",
        "右侧可叠加小圆形或方形装饰元素（accent 色，透明度 10-15%）",
        "标题文字上方可添加极细的分隔线（secondary 色，0.5pt）",
        "如果 kicker 存在，其下方用细线分隔（secondary 色，0.5pt）"
    ],
    "example": {
        "kicker": "技术分享",
        "title": "云原生架构实战",
        "subtitle": "从容器化到服务网格的演进之路",
        "author": "张明 / 基础架构部",
        "date": "2026年5月",
        "palette": "ocean_soft"
    },
    "example_2": {
        "kicker": "{领域/标签}",
        "title": "{主题名称}",
        "subtitle": "{一句副标题/Slogan}",
        "author": "{演讲者/机构}",
        "date": "{日期}"
    },
    "when_to_use": [
        "PPT第一页，开场",
        "需要强视觉冲击力的场合"
    ],
    "never": [
        "禁止在标题下加装饰线",
        "禁止使用深色背景",
        "禁止标题字号小于36pt"
    ]
}
