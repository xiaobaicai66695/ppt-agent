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
            "title": "智护康养",
            "subtitle": "AI+物联网智慧养老解决方案",
            "author": "星辰创新团队",
            "date": "2025年5月",
            "notes": "标题页醒目，吸引眼球",
            "filling_prompt": "必须填入真实内容：title 为项目名称，subtitle 为一句概括性口号，author 为团队名称，date 为参赛年份。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "example_detail",
            "title": "项目简介",
            "content_type": "example_detail",
            "kicker": "实例 · 项目简介",
            "lede": "用AI和物联网技术，为独居老人提供24小时智能监护服务",
            "context_block": "中国60岁以上人口已达2.8亿，其中独居老人超过2600万。子女工作繁忙，无法时刻陪伴在父母身边；传统养老方式成本高、覆盖有限；老人突发意外时，往往无法及时发现和救助。",
            "solution_block": "我们开发了一套'智护康养'系统，通过智能穿戴设备+AI算法+云平台，实现对老人健康状况的实时监测和异常预警。系统可监测心率、血压、位置等数据，当检测到异常时自动通知家属和社区。目前已在3个社区试点，服务老人200+。",
            "metrics": [
                {"value": "3项", "label": "核心创新点", "trend": "技术+模式+场景"},
                {"value": "2600万", "label": "目标用户群", "trend": "市场空间巨大"},
                {"value": "95%", "label": "预警准确率", "trend": "行业领先"}
            ],
            "takeaway": "启示：用科技守护亲情，让老人安享晚年，让子女安心工作",
            "notes": "用一句话和一页让评委快速了解项目",
            "filling_prompt": "必须填入真实内容：lede 一句话概括项目核心价值；context_block 描述项目诞生的背景和团队动机（1-2句话）；solution_block 具体说明项目核心功能和目标用户（2-3句话）；metrics_grid 提供3个指标（如创新点数、目标用户规模、预期效果），每个有 value、label、trend；takeaway 用一句话总结参赛价值。禁止虚构数据。"
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
            "header": "老龄化加速，养老需求井喷",
            "sub_header": "政策+市场+技术三重驱动",
            "paragraph": "中国正加速进入老龄化社会：60岁以上人口占比已达19.8%，预计2035年将突破30%。与此同时，传统养老面临诸多挑战：机构养老床位紧张、民办养老价格高昂、居家养老缺乏专业监护。社会迫切需要一种普惠、便捷、智能的养老解决方案。国家和地方出台多项政策支持智慧养老产业发展，5G、AI、物联网技术的成熟为智慧养老提供了技术支撑。",
            "references": [
                "https://www.stats.gov.cn/",
                "https://www.mca.gov.cn/"
            ],
            "notes": "用图文混排展示市场需求分析，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'市场分析'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为市场主题（不超过35字）；sub_header 为市场概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细分析市场规模、用户需求和痛点，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "stat_slide",
            "title": "市场规模",
            "content_type": "stat_slide",
            "stats": [
                {"value": "4.5万亿", "label": "养老产业总规模", "note": "2025年预测"},
                {"value": "25%", "label": "年复合增长率", "note": "持续高速增长"},
                {"value": "5000亿", "label": "智慧养老市场", "note": "2025年预测"}
            ],
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
            "cards": [
                {"header": "健康监测", "body": "24小时监测心率、血压、血氧等生命体征"},
                {"header": "智能预警", "body": "AI算法识别异常，自动通知家属和社区"},
                {"header": "位置追踪", "body": "实时定位，防止老人走失"},
                {"header": "用药提醒", "body": "智能提醒服药，避免漏服错服"}
            ],
            "notes": "4个核心功能卡片",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称）和 body（一句话描述功能价值）。功能要具体，不能是'功能1'等占位符。"
        },
        {
            "index": 9,
            "type": "image_text",
            "title": "应用场景：独居老人监护",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "应用场景",
            "header": "子女不在身边，也能守护父母安全",
            "sub_header": "智能守护，让爱无距离",
            "paragraph": "李奶奶今年78岁，独居在老城区，儿子在外地工作。以前儿子总担心妈妈的身体状况，每天要打好几个电话确认。现在，李奶奶戴上了我们的智能手环，系统实时监测她的健康数据。当血压异常时，系统会自动给儿子发消息提醒。一天晚上，李奶奶不慎摔倒，系统立即检测到异常体位变化，5分钟内通知了儿子和社区工作者，及时得到救助。李奶奶说：'有这个小东西，儿子放心多了，我也安心多了。'",
            "references": [
                "https://techcrunch.com/",
                "https://www.who.int/"
            ],
            "notes": "用图文混排展示典型应用场景，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {场景} 替换为具体应用场景（如'智能客服'、'数据分析'）；title 中的 {场景名称} 替换为具体场景名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为场景标题（不超过35字）；sub_header 为场景简介（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该应用场景的具体情况、用户需求和使用效果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
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
            "lede": "自研AI算法+轻量化设计，打造高性价比解决方案",
            "left_column": {
                "key_points": [
                    "自研心电AI算法：准确识别16种心律异常",
                    "多源数据融合：穿戴+环境+行为多维度感知",
                    "边缘计算：本地处理保护隐私，降低云端依赖",
                    "低功耗设计：续航长达7天"
                ],
                "analysis": [
                    "相比竞品：算法准确率提升15%，功耗降低30%",
                    "技术创新点：首次将ECG应用于居家养老场景"
                ]
            },
            "right_column": {
                "case_example": [
                    "佩戴智能手环，自动采集数据",
                    "边缘侧AI预处理，过滤无效数据",
                    "异常数据上传云端，深度分析",
                    "触发预警，推送家属和社区"
                ],
                "data_evidence": [
                    "预警准确率：95%",
                    "误报率：<3%",
                    "响应时间：<5分钟"
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
                {"value": "200+", "label": "服务老人数", "delta": "↑ 持续增长", "baseline": "3个社区试点"},
                {"value": "95%", "label": "预警准确率", "delta": "行业领先", "baseline": "vs 竞品80%"},
                {"value": "4.8分", "label": "用户满意度", "delta": "↑ 用户好评", "baseline": "满分5分"},
                {"value": "500元/月", "label": "客单价", "delta": "普惠定价", "baseline": "含设备+服务"}
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
                {"header": "张明（队长）", "body": "计算机学院，负责人工智能算法研发，曾获国家级竞赛一等奖"},
                {"header": "李华", "body": "物联网学院，负责硬件设计与嵌入式开发，已发表SCI论文1篇"},
                {"header": "王芳", "body": "管理学院，负责商业模式设计与市场推广，有创业经验"},
                {"header": "陈强", "body": "医学院，负责健康数据解读与医学顾问，附属医院实习"}
            ],
            "notes": "展示4位核心团队成员",
            "filling_prompt": "必须填入真实内容：提供4位核心成员信息，每人有 header（姓名+职位）和 body（教育背景、专业技能、主要贡献）。禁止虚构成员信息。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "发展规划",
            "key_points": [
                "01 短期：完成产品优化，扩展至10个社区",
                "02 中期：打通B端渠道，与养老机构合作",
                "03 长期愿景：成为中国智慧养老行业领导者"
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
