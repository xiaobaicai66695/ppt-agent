TEMPLATE = {
    "name": "short-class-talk",
    "name_cn": "课堂短时分享",
    "description": "适合课堂 5-10 分钟短时分享、课题介绍等场景。精简高效，重点突出，快速传达。",
    "target_audience": "老师、同学",
    "typical_slides": 6,
    "typical_duration": "5-10分钟",
    "palette": "simple_gray",
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
            "title": "{分享主题}",
            "subtitle": "{副标题（如有）}",
            "author": "{分享人}",
            "date": "{日期}",
            "notes": "标题页简洁，一目了然",
            "filling_prompt": "必须填入真实内容：title 为分享主题名称，subtitle 为副标题（如有），author 为分享人姓名，date 为分享日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "背景介绍",
            "content_type": "content_slide",
            "notes": "一句话介绍背景，快速切入主题",
            "filling_prompt": "必须填入真实内容：用 1-2 段话说明分享主题的背景，简洁明了，让听众快速了解为什么分享这个内容。"
        },
        {
            "index": 3,
            "type": "content_slide",
            "title": "核心内容",
            "content_type": "content_slide",
            "notes": "展示 3-4 个核心要点",
            "filling_prompt": "必须填入真实内容：列出 3-4 个核心要点，每条有标题和一句话说明。要点要精炼，每条不超过 30 字。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "案例/应用",
            "content_type": "content_slide",
            "notes": "用一个具体案例或应用场景加深理解",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少1个URL），再填入真实内容：提供一个具体案例，说明案例名称、背景和关键信息。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "要点回顾",
            "content_type": "content_slide",
            "notes": "快速回顾核心要点，加深印象",
            "filling_prompt": "必须填入真实内容：列出 3-4 个核心要点回顾，与第三页呼应，可用 checklist 形式。"
        },
        {
            "index": 6,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心收获1}",
                "02 {核心收获2}",
                "03 {行动建议（如有）}"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁结尾，可加 Q&A 提示",
            "filling_prompt": "必须填入真实内容：key_points 提供 2-3 个要点（核心收获 + 行动建议）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "课堂分享要精简，每页只讲一个要点",
        "文字要少，多用图表和关键词",
        "时间控制在 5-10 分钟，6-8 页为宜",
        "留出时间回答问题",
        "注意与听众的眼神交流"
    ]
}
