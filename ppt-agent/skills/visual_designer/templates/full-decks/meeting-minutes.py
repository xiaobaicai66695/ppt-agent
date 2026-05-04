TEMPLATE = {
    "name": "meeting-minutes",
    "name_cn": "会议纪要",
    "description": "适合会议记录、工作例会、项目评审会等场景。结构清晰，要点明确，行动明确。",
    "target_audience": "参会人员、项目经理、领导",
    "typical_slides": 8,
    "typical_duration": "5-10分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{会议名称}会议纪要",
            "subtitle": "{会议时间} | {会议地点}",
            "author": "{记录人}",
            "date": "{纪要日期}",
            "notes": "标题页简洁，注明会议基本信息",
            "filling_prompt": "必须填入真实内容：title 为会议名称，subtitle 为会议时间和地点，author 为记录人姓名，date 为纪要编写日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "会议概览",
            "content_type": "content_slide",
            "notes": "快速了解会议基本信息",
            "filling_prompt": "必须填入真实内容：说明会议主题、参会人员名单、主持人、记录人等基本信息。"
        },
        {
            "index": 3,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  会议背景",
                "02  讨论内容",
                "03  决议事项",
                "04  行动项"
            ],
            "notes": "清晰展示纪要结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 4,
            "type": "section_divider",
            "number": "01",
            "title": "会议背景",
            "subtitle": "会议目的与议程",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "会议背景与目的",
            "content_type": "content_slide",
            "notes": "说明召开会议的背景和目的",
            "filling_prompt": "必须填入真实内容：说明会议召开的背景（1-2段话）和本次会议的目的（2-3条）。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "讨论内容",
            "subtitle": "会议讨论要点",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "讨论要点汇总",
            "content_type": "content_slide",
            "notes": "列出会议讨论的主要问题",
            "filling_prompt": "必须填入真实内容：列出会议讨论的3-5个主要议题，每个议题说明讨论背景、各方观点和主要结论。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "决议事项",
            "subtitle": "会议决定",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "content_slide",
            "title": "会议决议",
            "content_type": "content_slide",
            "notes": "明确会议形成的决定",
            "filling_prompt": "必须填入真实内容：列出会议形成的3-5条决议，每条说明决议内容和通过情况。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "行动项",
            "subtitle": "待办事项与负责人",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "card_grid",
            "title": "行动项清单",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "明确每项行动的责任人和截止时间",
            "filling_prompt": "必须填入真实内容：提供4个行动项，每个有 header（任务名称，不超过35字）和 body（责任人+截止时间+具体要求）。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心决议}",
                "02 {关键行动项}",
                "03 {下次会议安排}"
            ],
            "thank_you": "感谢参与",
            "notes": "简洁总结，预告下次会议",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心决议+关键行动项+下次会议安排）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "会议纪要要客观准确，不添加个人解读",
        "决议和行动项要明确，避免模糊表述",
        "每项行动要有明确的负责人和截止时间",
        "结构清晰，便于后续追踪",
        "及时分发，确保相关人员知晓"
    ]
}
