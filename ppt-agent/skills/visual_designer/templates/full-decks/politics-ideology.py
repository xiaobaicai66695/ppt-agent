TEMPLATE = {
    "name": "politics-ideology",
    "name_cn": "思政/团课",
    "description": "适合思政教育、团课培训、爱国主义教育等场景。政治性强，价值观明确，结构清晰。",
    "target_audience": "学生、共青团员、党员、干部",
    "typical_slides": 16,
    "typical_duration": "20-30分钟",
    "palette": "patriotic_blue",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{思政主题名称}",
            "subtitle": "{副标题}",
            "author": "{演讲者}",
            "date": "{日期}",
            "notes": "标题页庄重大气，体现政治严肃性",
            "filling_prompt": "必须填入真实内容：title 为本次思政教育的主题名称（如'新时代青年的使命与担当'），subtitle 为概括性副标题，author 为演讲者姓名，date 为实际日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  理论学习",
                "02  案例分析",
                "03  实践要求",
                "04  总结提升"
            ],
            "notes": "让学员快速了解课程结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "理论学习",
            "subtitle": "核心思想与理论基础",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "核心思想解读",
            "content_type": "content_slide",
            "notes": "解读核心思想的关键要点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供核心思想的3-5个关键要点解读，每条有标题和详细说明。禁止虚构理论内容。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "quote_slide",
            "title": "重要论述",
            "content_type": "quote_slide",
            "notes": "展示领导人重要论述",
            "filling_prompt": "必须填入真实内容：提供1-2条重要论述（需准确引用），注明出处（讲话名称、时间、场合）。禁止虚构论述。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "案例分析",
            "subtitle": "典型案例与榜样力量",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "image_text",
            "title": "典型案例：{人物/事件名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示典型案例，增强可信性和感染力",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'榜样力量'；title 中的 {人物/事件名称} 替换为真实人物姓名或事件名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX同志先进事迹'，不超过35字）；sub_header 为身份简介（不超过35字）；bullets 列出3-4条主要事迹或贡献，每条不超过35字。references 列出 URL。禁止虚构人物和事迹。"
        },
        {
            "index": 8,
            "type": "card_grid",
            "title": "榜样人物",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "展示2-4位榜样人物的简要事迹",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4位榜样人物，每人有 header（姓名+身份，不超过35字）和 body（一句话概括其主要事迹，不超过35字）。禁止虚构人物。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "实践要求",
            "subtitle": "如何将思想落实到行动",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "process_flow",
            "title": "实践路径",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "{路径1}", "desc": "{一句话描述}"},
                {"num": "02", "title": "{路径2}", "desc": "{一句话描述}"},
                {"num": "03", "title": "{路径3}", "desc": "{一句话描述}"},
                {"num": "04", "title": "{路径4}", "desc": "{一句话描述}"}
            ],
            "notes": "4个具体可行的实践路径",
            "filling_prompt": "必须填入真实内容：提供4个具体可操作的实践路径，每条有 title（行动名称）和 desc（具体做法，不超过30字）。禁止空泛口号。"
        },
        {
            "index": 11,
            "type": "content_slide",
            "title": "行动指南",
            "content_type": "content_slide",
            "notes": "具体的行动建议",
            "filling_prompt": "必须填入真实内容：提供3-5条具体可执行的行动建议，每条说明具体做什么、怎么做。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "04",
            "title": "总结提升",
            "subtitle": "巩固学习成果",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "学习要点回顾",
            "content_type": "content_slide",
            "notes": "回顾本节课的核心要点",
            "filling_prompt": "必须填入真实内容：列出3-5个本节课的核心学习要点。"
        },
        {
            "index": 14,
            "type": "content_slide",
            "title": "心得体会",
            "content_type": "content_slide",
            "notes": "引导学员思考和分享",
            "filling_prompt": "必须填入真实内容：提供3-5个引导性问题，帮助学员思考（如'通过今天的学习，我最深的体会是...'、'在今后的学习中，我将...'）。"
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "学习效果自评",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "自评 · 学习效果",
            "notes": "帮助学员自我评估学习效果",
            "filling_prompt": "必须填入真实内容：提供4个学习效果自评维度，每个有 value（如分数或等级）、label（维度名称）、delta（变化趋势）。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心要点1}",
                "02 {核心要点2}",
                "03 {行动承诺}"
            ],
            "thank_you": "感谢聆听",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心学习要点2条+行动承诺1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "政治表达要准确规范",
        "案例要有真实性和感染力",
        "实践建议要具体可操作",
        "避免空洞说教，注重以理服人",
        "设计要有仪式感和庄重感"
    ]
}
