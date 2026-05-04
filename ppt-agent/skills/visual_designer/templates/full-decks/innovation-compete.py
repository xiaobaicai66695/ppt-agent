TEMPLATE = {
    "name": "innovation-compete",
    "name_cn": "科创竞赛",
    "description": "适合大创、挑战杯、互联网+等科创竞赛汇报。创新性强，数据支撑，展示潜力。",
    "target_audience": "评委、投资人、导师、参赛团队",
    "typical_slides": 16,
    "typical_duration": "10-15分钟（路演）",
    "palette": "civic_gold",
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
            "title": "{项目名称}",
            "subtitle": "{项目口号/一句话介绍}",
            "author": "{团队名称}",
            "date": "{参赛年份}",
            "notes": "标题页醒目，吸引眼球",
            "filling_prompt": "必须填入真实内容：title 为项目名称，subtitle 为一句概括性口号，author 为团队名称，date 为参赛年份。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "项目简介",
            "content_type": "content_slide",
            "notes": "用一句话和一页让评委快速了解项目",
            "filling_prompt": "必须填入真实内容：提供项目的一句话简介（不超过30字）、核心功能描述（2-3句话）、目标用户/应用场景。"
        },
        {
            "index": 3,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  市场分析",
                "02  产品介绍",
                "03  技术方案",
                "04  商业模式",
                "05  团队介绍",
                "06  发展规划"
            ],
            "notes": "让评委了解汇报结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 4,
            "type": "section_divider",
            "number": "01",
            "title": "市场分析",
            "subtitle": "需求与机会",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 5,
            "type": "image_text",
            "title": "市场需求分析",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示市场需求分析，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'市场分析'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为市场主题（不超过35字）；sub_header 为市场概述（不超过35字）；bullets 列出3-4条市场需求或痛点，每条不超过35字。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "stat_slide",
            "title": "市场规模",
            "content_type": "stat_slide",
            "notes": "大数字展示市场潜力",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供市场规模数据（如TAM、SAM、SOM），用大数字突出。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "02",
            "title": "产品介绍",
            "subtitle": "核心功能与亮点",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "card_grid",
            "title": "核心功能",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "4个核心功能卡片",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称）和 body（一句话描述功能价值）。功能要具体，不能是'功能1'等占位符。"
        },
        {
            "index": 9,
            "type": "image_text",
            "title": "应用场景：{场景名称}",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "notes": "用图文混排展示典型应用场景，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {场景} 替换为具体应用场景（如'智能客服'、'数据分析'）；title 中的 {场景名称} 替换为具体场景名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为场景标题（不超过35字）；sub_header 为场景简介（不超过35字）；bullets 列出3-4条应用效果或用户价值，每条不超过35字。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "03",
            "title": "技术方案",
            "subtitle": "技术架构与创新",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "deep_dive",
            "title": "核心技术方案",
            "content_type": "deep_dive",
            "kicker": "技术 · 核心创新",
            "lede": "项目的技术架构与创新点",
            "left_column": {
                "key_points": [
                    "{要点1}",
                    "{要点2}",
                    "{要点3}",
                    "{要点4}"
                ],
                "analysis": [
                    "{分析维度1}",
                    "{分析维度2}"
                ]
            },
            "right_column": {
                "case_example": [
                    "{案例要素1}",
                    "{案例要素2}",
                    "{案例要素3}",
                    "{案例要素4}"
                ],
                "data_evidence": [
                    "{数据指标1}",
                    "{数据指标2}",
                    "{数据指标3}"
                ]
            },
            "notes": "展示技术实力和创新点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_column.key_points 为技术要点（3-4条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 URL。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "04",
            "title": "商业模式",
            "subtitle": "盈利与发展",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "kpi_dashboard",
            "title": "运营数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 运营成果",
            "notes": "用数据证明项目可行性",
            "filling_prompt": "必须填入真实内容：提供4个运营指标（如用户数、活跃度、收入、增长率等），每个有 value、label、delta、baseline。如项目处于早期阶段，可用预期数据但需注明。禁止虚构已完成的指标数据。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "05",
            "title": "团队介绍",
            "subtitle": "核心成员",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 15,
            "type": "card_grid",
            "title": "团队成员",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "展示4位核心团队成员",
            "filling_prompt": "必须填入真实内容：提供4位核心成员信息，每人有 header（姓名+职位）和 body（教育背景、专业技能、主要贡献）。禁止虚构成员信息。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "发展规划",
            "key_points": [
                "01 {短期目标}",
                "02 {中期目标}",
                "03 {长期愿景}"
            ],
            "thank_you": "感谢聆听，欢迎交流！",
            "notes": "展示团队愿景和执行力",
            "filling_prompt": "必须填入真实内容：key_points 提供3个发展阶段（短期/中期/长期目标）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "竞赛路演要抓住评委注意力",
        "数据要真实有力，来源可查",
        "突出创新点和竞争优势",
        "团队展示要体现专业性和互补性",
        "PPT设计要有视觉冲击力"
    ]
}
