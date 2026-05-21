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
    "scene_guidance": {
        "competition_strategy": "竞赛路演的核心是展示：创新性+可行性+团队力",
        "key_elements": [
            "项目创新性：技术/模式/应用的创新点",
            "项目可行性：技术实现+商业模式验证",
            "团队执行力：背景互补+分工明确",
            "市场潜力：市场规模+增长预期"
        ],
        "judging_criteria": {
            "innovation": "创新性：30%",
            "feasibility": "可行性：25%",
            "commercial": "商业价值：20%",
            "team": "团队能力：15%",
            "presentation": "展示效果：10%"
        }
    },
    "pitch_tips": {
        "opening": "30秒内抓住评委注意力",
        "problem": "清晰阐述问题和市场机会",
        "solution": "用一句话说清楚你的解决方案",
        "traction": "展示已有的进展和成果",
        "ask": "明确说明你需要什么"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "项目名称",
            "subtitle": "项目一句话价值主张",
            "author": "团队名称",
            "date": "参赛年份",
            "notes": "标题页醒目，吸引眼球",
            "filling_prompt": "必须填入真实内容：title 为项目名称，subtitle 为一句概括性口号，author 为团队名称，date 为参赛年份。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "example_detail",
            "title": "项目简介",
            "content_type": "example_detail",
            "kicker": "实例 · 项目简介",
            "lede": "用一句话概括项目核心价值",
            "context_block": "描述项目诞生的背景和团队动机（1-2句话）。",
            "solution_block": "具体说明项目核心功能和目标用户（2-3句话）。",
            "metrics": [
                {"value": "数量", "label": "核心创新点", "trend": "创新维度"},
                {"value": "数量", "label": "目标用户规模", "trend": "市场空间"},
                {"value": "百分比", "label": "效果指标", "trend": "性能指标"}
            ],
            "takeaway": "一句话总结参赛价值。",
            "notes": "用一句话和一页让评委快速了解项目",
            "filling_prompt": "必须填入真实内容：lede 一句话概括项目核心价值；context_block 描述项目诞生的背景和团队动机（1-2句话）；solution_block 具体说明项目核心功能和目标用户（2-3句话）；metrics_grid 提供3个指标，每个有 value、label、trend；takeaway 用一句话总结参赛价值。禁止虚构数据。"
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
            "kicker": "市场分析",
            "header": "市场主题标题",
            "sub_header": "市场概述",
            "paragraph": "详细分析市场规模、用户需求和痛点，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威数据来源URL1",
                "https://权威数据来源URL2"
            ],
            "notes": "用图文混排展示市场需求分析，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为市场主题（不超过35字）；sub_header 为市场概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细分析市场规模、用户需求和痛点，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "stat_slide",
            "title": "市场规模",
            "content_type": "stat_slide",
            "stats": [
                {"value": "数字", "label": "指标名称", "note": "数据来源"},
                {"value": "数字", "label": "指标名称", "note": "数据来源"},
                {"value": "数字", "label": "指标名称", "note": "数据来源"}
            ],
            "notes": "大数字展示市场潜力",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供市场规模数据（如TAM/SAM/SOM），用大数字突出。references 列出 URL。禁止虚构数据。"
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
            "cards": [
                {"header": "功能名称", "body": "功能描述和价值"},
                {"header": "功能名称", "body": "功能描述和价值"},
                {"header": "功能名称", "body": "功能描述和价值"},
                {"header": "功能名称", "body": "功能描述和价值"}
            ],
            "notes": "4个核心功能卡片",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称）和 body（一句话描述功能价值）。功能要具体，不能是'功能1'等占位符。"
        },
        {
            "index": 9,
            "type": "image_text",
            "title": "应用场景",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "应用场景",
            "header": "场景标题",
            "sub_header": "场景简介",
            "paragraph": "详细描述该应用场景的具体情况、用户需求和使用效果，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威数据来源URL1",
                "https://权威数据来源URL2"
            ],
            "notes": "用图文混排展示典型应用场景，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为场景标题（不超过35字）；sub_header 为场景简介（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该应用场景的具体情况、用户需求和使用效果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
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
            "lede": "一句话概括技术核心价值",
            "left_column": {
                "key_points": [
                    "技术要点1",
                    "技术要点2",
                    "技术要点3",
                    "技术要点4"
                ],
                "analysis": [
                    "与竞品对比分析",
                    "技术创新点"
                ]
            },
            "right_column": {
                "case_example": [
                    "案例步骤1",
                    "案例步骤2",
                    "案例步骤3",
                    "案例步骤4"
                ],
                "data_evidence": [
                    "效果指标1",
                    "效果指标2",
                    "效果指标3"
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
            "kpis": [
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"}
            ],
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
            "cards": [
                {"header": "姓名（职位）", "body": "教育背景、专业技能、主要贡献"},
                {"header": "姓名（职位）", "body": "教育背景、专业技能、主要贡献"},
                {"header": "姓名（职位）", "body": "教育背景、专业技能、主要贡献"},
                {"header": "姓名（职位）", "body": "教育背景、专业技能、主要贡献"}
            ],
            "notes": "展示4位核心团队成员",
            "filling_prompt": "必须填入真实内容：提供4位核心成员信息，每人有 header（姓名+职位）和 body（教育背景、专业技能、主要贡献）。禁止虚构成员信息。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "发展规划",
            "key_points": [
                "01 短期：目标描述",
                "02 中期：目标描述",
                "03 长期愿景：愿景描述"
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
        "PPT设计要有视觉冲击力",
        "控制时间，每页只讲一个重点",
        "准备Q&A，预判评委可能的问题",
        "展示Passion，评委喜欢有热情的团队"
    ],
    "competition_prep": {
        "common_judging_questions": [
            "你的创新点是什么？",
            "技术壁垒在哪里？",
            "如何获取用户？",
            "项目的局限性是什么？",
            "团队如何分工？"
        ],
        "material_checklist": [
            "商业计划书",
            "产品Demo视频",
            "技术文档",
            "用户调研报告",
            "财务报表（如有）"
        ]
    }
}
