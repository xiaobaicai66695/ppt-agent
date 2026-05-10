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
            "title": "CRM系统升级项目周例会纪要",
            "subtitle": "2025年3月10日 14:00-15:30 | 会议室A",
            "author": "王芳",
            "date": "2025年3月10日",
            "notes": "标题页简洁，注明会议基本信息",
            "filling_prompt": "必须填入真实内容：title 为会议名称，subtitle 为会议时间和地点，author 为记录人姓名，date 为纪要编写日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "会议概览",
            "content_type": "content_slide",
            "meeting_info": {
                "meeting_name": "CRM系统升级项目周例会",
                "time": "2025年3月10日 14:00-15:30",
                "location": "会议室A",
                "host": "张经理（项目经理）",
                "recorder": "王芳",
                "attendees": [
                    "张经理（项目经理）",
                    "李工（开发负责人）",
                    "王芳（产品经理）",
                    "刘工（测试负责人）",
                    "赵总（客户代表）"
                ],
                "absent": ["陈工（UI设计师）- 请假"]
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
            "lede": "跟进CRM升级项目进展，解决当前阻塞问题",
            "context_block": "CRM升级项目自2月启动以来，整体进展顺利，但本周遇到几个关键阻塞问题需要决策：1）第三方接口对接方案需要确认；2）客户要求变更需求范围；3）测试环境部署遇到问题。",
            "solution_block": "本次会议的主要目标是：1）明确第三方接口对接的技术方案；2）讨论需求变更的影响和应对；3）协调测试环境部署问题；4）确认下阶段里程碑计划。",
            "metrics": [
                {"value": "5人", "label": "参会人数", "trend": "含客户代表"},
                {"value": "3个", "label": "讨论议题数", "trend": "关键阻塞项"},
                {"value": "3项", "label": "预期决议数", "trend": "全部明确"}
            ],
            "takeaway": "启示：及时解决阻塞问题是保障项目进度的关键",
            "notes": "说明召开会议的背景和目的",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心议题和决策焦点；context_block 描述会议背景原因和组织动机（1-2句话）；solution_block 具体说明要解决的核心问题和预期产出（2-3句话）；metrics_grid 提供3个量化指标（如参会人数、议题数量、预期决策数），每个有 value、label、trend；takeaway 用一句话总结会议对项目/工作的影响。"
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
                    "topic": "议题一：第三方接口对接方案",
                    "discussion": "李工介绍了两种对接方案：方案A采用直连方式，开发周期短但扩展性差；方案B采用中间件方式，开发周期长但扩展性好。团队倾向方案B，但需要额外2周时间。",
                    "conclusion": "决定采用方案B，增加2周开发周期。"
                },
                {
                    "topic": "议题二：客户需求变更",
                    "discussion": "客户提出增加'数据导出'功能，影响范围涉及前端、后台和数据库设计。预估影响2周进度。",
                    "conclusion": "将'数据导出'功能纳入下一版本，本期保持原范围。"
                },
                {
                    "topic": "议题三：测试环境部署",
                    "discussion": "测试环境因服务器资源问题无法部署，影响测试进度。刘工建议使用云测试环境。",
                    "conclusion": "采用云测试环境方案，本周完成部署。"
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
            "lede": "明确三大决议，保障项目稳步推进",
            "context_block": "经过充分讨论，会议就以下关键事项达成一致：",
            "solution_block": "会议决议如下：1）第三方接口采用方案B（中间件方式），开发周期延长2周，总工期调整为12周；2）客户需求变更'数据导出'功能移至下版本，本期保持原范围；3）测试环境改用云测试环境，本周完成部署；4）项目里程碑顺延2周。",
            "metrics": [
                {"value": "4项", "label": "形成的决议数", "trend": "100%通过"},
                {"value": "2周", "label": "整体延期", "trend": "需更新计划"},
                {"value": "5人", "label": "决策参与人数", "trend": "含客户"}
            ],
            "takeaway": "启示：及时决策，避免项目阻塞",
            "notes": "明确会议形成的决定",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心决议；context_block 描述会议讨论的核心过程（1-2句话）；solution_block 具体说明决议内容和执行要求（2-3句话）；metrics_grid 提供3个量化指标（如决议数量、执行时限、参与人数），每个有 value、label、trend；takeaway 用一句话总结决议的指导意义。"
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
                {"header": "接口方案设计文档", "body": "负责人：李工 | 截止：3月12日 | 要求：输出详细设计文档，包含接口规范和中间件选型建议"},
                {"header": "需求变更通知客户", "body": "负责人：王芳 | 截止：3月11日 | 要求：撰写变更说明邮件，与客户确认并获取书面确认"},
                {"header": "云测试环境部署", "body": "负责人：刘工 | 截止：3月14日 | 要求：完成环境搭建和基础测试数据准备"},
                {"header": "更新项目计划", "body": "负责人：张经理 | 截止：3月12日 | 要求：基于新里程碑更新项目计划，发给全员"}
            ],
            "notes": "明确每项行动的责任人和截止时间",
            "filling_prompt": "必须填入真实内容：提供4个行动项，每个有 header（任务名称，不超过35字）和 body（详细描述责任人、截止时间、具体要求和预期成果，100-120字）。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心决议：采用方案B、需求变更延期、云测试环境",
                "02 关键行动：本周需完成方案设计、变更通知、环境部署",
                "03 下次会议：3月17日14:00，会议室A"
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
