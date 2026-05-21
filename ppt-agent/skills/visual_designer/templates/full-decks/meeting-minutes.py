TEMPLATE = {
    "name": "meeting-minutes",
    "name_cn": "会议纪要",
    "description": "适合会议记录、工作例会、项目评审会等场景。结构清晰，要点明确，行动明确。",
    "target_audience": "参会人员、项目经理、领导",
    "typical_slides": 12,
    "typical_duration": "5-10分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "documentation_standards": "纪要的核心价值是：备忘、传递、追踪",
        "key_principles": [
            "客观记录：如实记录，不添加个人解读",
            "要点清晰：突出关键决策和行动项",
            "责任明确：每项行动有负责人和截止时间",
            "及时分发：会议结束后24小时内发出"
        ],
        "quality_check": [
            "是否记录了所有关键决策？",
            "是否明确了每项行动的负责人？",
            "是否有明确的截止时间？",
            "是否分发了所有相关人员？"
        ]
    },
    "minutes_template": {
        "basic_info": "会议基本信息（时间、地点、参会人等）",
        "agenda": "会议议程",
        "discussion": "讨论内容摘要",
        "decisions": "会议决议",
        "action_items": "行动项清单",
        "next_meeting": "下次会议安排"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "会议名称",
            "subtitle": "会议时间和地点",
            "author": "记录人姓名",
            "date": "纪要编写日期",
            "notes": "标题页简洁，注明会议基本信息",
            "filling_prompt": "必须填入真实内容：title 为会议名称，subtitle 为会议时间和地点，author 为记录人姓名，date 为纪要编写日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "会议概览",
            "content_type": "content_slide",
            "meeting_info": {
                "meeting_name": "会议名称",
                "time": "时间",
                "location": "地点",
                "host": "主持人",
                "recorder": "记录人",
                "attendees": [
                    "参会人1",
                    "参会人2"
                ],
                "absent": ["缺席人1"]
            },
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
            "type": "example_detail",
            "title": "会议背景与目的",
            "content_type": "example_detail",
            "kicker": "实例 · 会议背景",
            "lede": "一句话概括核心议题和决策焦点",
            "context_block": "描述会议背景原因和组织动机（1-2句话）。",
            "solution_block": "具体说明要解决的核心问题和预期产出（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "参会人数", "trend": "说明"},
                {"value": "数字", "label": "讨论议题数", "trend": "说明"},
                {"value": "数字", "label": "预期决议数", "trend": "说明"}
            ],
            "takeaway": "一句话总结会议对项目/工作的影响。",
            "notes": "说明召开会议的背景和目的",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心议题和决策焦点；context_block 描述会议背景原因和组织动机（1-2句话）；solution_block 具体说明要解决的核心问题和预期产出（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结会议对项目/工作的影响。"
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
            "topics": [
                {
                    "topic": "议题名称",
                    "discussion": "讨论背景、各方观点",
                    "conclusion": "结论"
                },
                {
                    "topic": "议题名称",
                    "discussion": "讨论背景、各方观点",
                    "conclusion": "结论"
                },
                {
                    "topic": "议题名称",
                    "discussion": "讨论背景、各方观点",
                    "conclusion": "结论"
                }
            ],
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
            "type": "example_detail",
            "title": "会议决议",
            "content_type": "example_detail",
            "kicker": "实例 · 决议事项",
            "lede": "一句话概括核心决议",
            "context_block": "描述会议讨论的核心过程（1-2句话）。",
            "solution_block": "具体说明决议内容和执行要求（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "形成的决议数", "trend": "情况"},
                {"value": "数字", "label": "影响范围", "trend": "说明"},
                {"value": "数字", "label": "决策参与人数", "trend": "说明"}
            ],
            "takeaway": "一句话总结决议的指导意义。",
            "notes": "明确会议形成的决定",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心决议；context_block 描述会议讨论的核心过程（1-2句话）；solution_block 具体说明决议内容和执行要求（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结决议的指导意义。"
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
            "cards": [
                {"header": "任务名称（不超过35字）", "body": "责任人 | 截止时间 | 要求和预期成果（100-120字）。"},
                {"header": "任务名称（不超过35字）", "body": "责任人 | 截止时间 | 要求和预期成果（100-120字）。"},
                {"header": "任务名称（不超过35字）", "body": "责任人 | 截止时间 | 要求和预期成果（100-120字）。"},
                {"header": "任务名称（不超过35字）", "body": "责任人 | 截止时间 | 要求和预期成果（100-120字）。"}
            ],
            "notes": "明确每项行动的责任人和截止时间",
            "filling_prompt": "必须填入真实内容：提供4个行动项，每个有 header（任务名称，不超过35字）和 body（详细描述责任人、截止时间、具体要求和预期成果，100-120字）。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心决议：描述",
                "02 关键行动：描述",
                "03 下次会议：时间地点"
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
        "及时分发，确保相关人员知晓",
        "使用统一模板，保持格式一致",
        "重点突出，关键信息一目了然",
        "存档规范，便于查询历史"
    ],
    "distribution_checklist": [
        "所有参会人员",
        "项目干系人",
        "相关领导",
        "PMO（如有）"
    ],
    "archival_guidelines": {
        "naming_convention": "会议纪要_项目名称_日期",
        "storage_location": "项目文档库/会议纪要文件夹",
        "retention_period": "项目结束+1年"
    }
}
