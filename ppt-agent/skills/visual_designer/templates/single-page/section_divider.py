TEMPLATE = {
    "type": "section_divider",
    "name": "章节分隔页",
    "description": "以大号章节序号 + 简短标题为主，常用大面积色块，仪式感强。用于划分PPT章节。",
    "layout_hint": "center 或 left",
    "elements": {
        "section_number": {
            "position": "top_left 或 左侧偏中",
            "font_size": "72-96pt",
            "font_weight": "bold",
            "color": "accent",
            "opacity": "0.2-0.3（大号装饰数字）"
        },
        "section_title": {
            "position": "center 或 偏左",
            "font_size": "40-48pt",
            "font_weight": "bold",
            "color": "text",
            "max_chars": 20
        },
        "section_subtitle": {
            "position": "below_title",
            "font_size": "16pt",
            "color": "secondary",
            "max_chars": 30
        },
        "chapter_tag": {
            "position": "top_right",
            "font_size": "12pt",
            "color": "text_muted",
            "text": "CHAPTER {number}"
        }
    },
    "visual_elements": [
        "大面积 primary 色或 accent 色色块（占页面 40-60%）作为背景",
        "大号章节序号作为装饰性背景文字（透明度 20-30%）",
        "章节标题居中或偏左，与色块形成对比",
        "可用色块从左到右渐变（深→浅）暗示内容递进",
        "右侧可叠加装饰性几何图形（小圆形、斜切矩形，primary 色，透明度 10-15%）",
        "底部可添加细线装饰（secondary 色，0.5pt）",
        "可用左侧色条作为装饰（accent 色，3pt 宽，从上延伸至下）",
        "背景色块可使用斜向分割（对角线，左侧色块右侧留白）或横向分割（上色块下留白）"
    ],
    "example": {
        "number": "01",
        "title": "什么是大模型",
        "subtitle": "从原理到应用的全面解读",
        "chapter_tag": "CHAPTER 01"
    },
    "example_2": {
        "number": "{序号}",
        "title": "{章节名称}",
        "subtitle": "{章节副标题}",
        "chapter_tag": "CHAPTER {序号}"
    },
    "when_to_use": [
        "PPT章节转换",
        "划分大的内容段落",
        "演讲中的「章节页」概念"
    ],
    "never": [
        "禁止用深色背景",
        "禁止章节标题字号小于36pt",
        "禁止没有足够视觉区分度",
        "禁止与标题页视觉相似（需要明显区分）",
        "**重要**：生成器函数参数名是 number，不是 section_number；是 subtitle，不是 section_subtitle"
    ]
}
