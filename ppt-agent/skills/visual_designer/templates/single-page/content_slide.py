TEMPLATE = {
    "type": "content_slide",
    "name": "普通内容页",
    "description": "通用兜底类型。用清晰的小标题 + 3~5 条要点，配合适度留白。适合左侧内容+右侧装饰图形的布局。",
    "layout_hint": "left_content + right_decoration",
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
            "margin_bottom": "0.2in",
            "max_chars": 30
        },
        "section_header": {
            "font_size": "20pt",
            "font_weight": "bold",
            "color": "primary",
            "margin_bottom": "0.15in",
            "max_chars": 20
        },
        "bullet_list": {
            "font_size": "16pt",
            "line_spacing": "1.5x",
            "bullet_indent": "0.3in",
            "items_max": 5,
            "char_per_item_max": 20,
            "bullet_style": "实心圆点或竖线色块",
            "content_placeholder": [
                "{要点1}（{具体说明}）",
                "{要点2}（{具体说明}）",
                "{要点3}（{具体说明}）"
            ]
        },
        "notes": "左对齐；bullet 最多5条，每条不超过20个中文字符"
    },
    "visual_elements": [
        "左侧竖向色条装饰（primary 色，3pt 宽，从标题区延伸至底部）",
        "标题前用竖线色块作为视觉锚点（primary 色，4pt 宽，0.2in 高）",
        "右侧可添加装饰性图形（如信息卡片、统计数字块）",
        "bullet 点可用 accent 色小方块代替实心圆，区分度更强",
        "要点之间可用细线分隔（secondary 色，0.5pt），增加层次感",
        "可选：右侧添加小统计卡片（如「98%满意」「3x效率」）",
        "可选：底部添加整体进度条或汇总统计"
    ],
    "example": {
        "kicker": "技术背景",
        "title": "微服务架构的优势",
        "section_header": "三大核心价值",
        "bullets": [
            "独立部署（团队协作效率提升 3 倍）",
            "弹性伸缩（峰值处理能力线性扩展）",
            "技术异构（不同服务可选最优技术栈）",
            "故障隔离（单点故障不引发级联崩溃）",
            "快速迭代（发布周期从月缩短到天）"
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{技术名称}发展历程",
        "section_header": "三大技术里程碑",
        "bullets": [
            "{里程碑1}（{年份}）：{一句话描述}",
            "{里程碑2}（{年份}）：{一句话描述}",
            "{里程碑3}（{年份}）：{一句话描述}"
        ]
    },
    "when_to_use": [
        "没有特殊内容特征时的兜底选择",
        "需要清晰传递多条信息的场合"
    ],
    "never": [
        "禁止超过5条bullet",
        "禁止每条bullet超过20个中文字符",
        "禁止正文居中对齐",
        "禁止在标题下加装饰线"
    ]
}
