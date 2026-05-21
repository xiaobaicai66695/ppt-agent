TEMPLATE = {
    "name": "research-report",
    "name_cn": "调研报告",
    "description": "适合市场调研、行业分析、可行性研究等场景。数据详实，逻辑严密，结论明确。",
    "target_audience": "决策层、项目组、客户、评审专家",
    "typical_slides": 16,
    "typical_duration": "20-30分钟",
    "palette": "report_green",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "research_objectives": "调研的核心目的是为决策提供依据",
        "key_principles": [
            "数据说话：用数据支撑观点",
            "逻辑严密：论证过程清晰可追溯",
            "客观中立：如实呈现调研发现",
            "结论明确：给出可操作的建议"
        ],
        "methodology_tips": [
            "明确调研范围和方法",
            "多渠道收集数据，确保代表性",
            "交叉验证数据，确保准确性",
            "区分事实与观点"
        ]
    },
    "report_structure": {
        "executive_summary": "执行摘要",
        "background": "调研背景",
        "methodology": "调研方法",
        "findings": "现状分析",
        "diagnosis": "问题诊断",
        "recommendations": "对策建议",
        "conclusion": "结论展望"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "调研报告名称",
            "subtitle": "概括性副标题",
            "author": "调研团队/负责人",
            "date": "报告完成日期",
            "notes": "标题页正式，注明调研时间和团队",
            "filling_prompt": "必须填入真实内容：title 为调研报告的名称，subtitle 为概括性副标题，author 为调研团队或负责人，date 为报告完成日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "example_detail",
            "title": "执行摘要",
            "content_type": "example_detail",
            "kicker": "实例 · 执行摘要",
            "lede": "一句话概括核心发现",
            "context_block": "描述调研背景和问题（1-2句话）。",
            "solution_block": "总结主要发现和建议方向（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "样本量指标", "trend": "覆盖范围"},
                {"value": "数字", "label": "发现数量", "trend": "说明"},
                {"value": "数字", "label": "建议数量", "trend": "说明"}
            ],
            "takeaway": "一句话总结调研对决策的意义。",
            "notes": "一页概括报告核心发现和建议",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心发现；context_block 描述调研背景和问题（1-2句话）；solution_block 总结主要发现和建议方向（2-3句话）；metrics_grid 提供3个调研指标，每个有 value、label、trend；takeaway 用一句话总结调研对决策的意义。"
        },
        {
            "index": 3,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  调研背景",
                "02  调研方法",
                "03  现状分析",
                "04  问题诊断",
                "05  对策建议",
                "06  结论展望"
            ],
            "notes": "清晰展示报告结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 4,
            "type": "section_divider",
            "number": "01",
            "title": "调研背景",
            "subtitle": "目的与范围",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "调研背景与目的",
            "content_type": "content_slide",
            "background": {
                "market_trend": "行业趋势描述",
                "business_need": "调研目的描述",
                "research_goal": "调研目标描述"
            },
            "notes": "说明为什么开展这次调研",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：说明调研的背景（行业趋势、政策环境等）、调研目的、调研范围和对象。references 列出 URL。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "调研方法",
            "subtitle": "数据来源与分析方法",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "调研方法说明",
            "content_type": "content_slide",
            "methods": [
                {"name": "方法名称", "desc": "方法描述"},
                {"name": "方法名称", "desc": "方法描述"},
                {"name": "方法名称", "desc": "方法描述"},
                {"name": "方法名称", "desc": "方法描述"}
            ],
            "notes": "说明数据来源和调研方法",
            "filling_prompt": "必须填入真实内容：说明调研采用的方法（如问卷调查、深度访谈、数据采集等）、样本量、抽样方法、数据分析工具。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "现状分析",
            "subtitle": "数据与发现",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "kpi_dashboard",
            "title": "关键发现",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 关键发现",
            "kpis": [
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"}
            ],
            "notes": "用数据呈现核心发现",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个关键数据指标，每个有 value、label、delta、baseline。这些数据要能直接支持调研结论。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "多维度调研发现",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "发现 · 数据分析",
            "header": "最核心的一个调研发现",
            "sub_header": "发现的影响说明",
            "paragraph": "详细描述调研的核心发现、数据支撑和实际意义，用流畅的段落形式呈现，禁止罗列要点，必须包含具体数字。",
            "references": [
                "https://权威来源URL1（标注机构名称）",
                "https://权威来源URL2（标注机构名称）",
                "https://权威来源URL3（标注机构名称）"
            ],
            "notes": "右侧配数据图表/行业报告截图，左侧呈现核心发现",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含行业报告来源），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最核心的一个调研发现（不超过35字）；sub_header 为发现的影响说明（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述调研的核心发现、数据支撑和实际意义，用流畅的段落形式呈现，禁止罗列要点，必须包含具体数字。references 逐条列出 URL 并标注报告来源机构名称。禁止模糊描述。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "问题诊断",
            "subtitle": "挑战与风险",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 12,
            "type": "image_text",
            "title": "主要问题与风险",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "诊断 · 问题识别",
            "header": "最核心的一个问题",
            "sub_header": "问题的严重程度评估",
            "paragraph": "详细描述问题的具体表现、产生原因和潜在影响，用流畅的段落形式呈现，禁止罗列要点，必须包含影响数据（如占比、损失金额、影响范围等）。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2",
                "https://权威来源URL3"
            ],
            "notes": "左侧配问题关系图/风险矩阵图，右侧分析问题详情",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最核心的一个问题（不超过35字）；sub_header 为问题的严重程度评估（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述问题的具体表现、产生原因和潜在影响，用流畅的段落形式呈现，禁止罗列要点，必须包含影响数据。references 逐条列出 URL 并标注来源。禁止模糊描述。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "对策建议",
            "subtitle": "解决方案与行动计划",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 14,
            "type": "process_flow",
            "title": "建议与行动计划",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "建议名称", "desc": "具体行动步骤"},
                {"num": "02", "title": "建议名称", "desc": "具体行动步骤"},
                {"num": "03", "title": "建议名称", "desc": "具体行动步骤"},
                {"num": "04", "title": "建议名称", "desc": "具体行动步骤"}
            ],
            "notes": "4个具体可执行的建议",
            "filling_prompt": "必须填入真实内容：提供4个具体可执行的建议，每个有 title（建议名称）和 desc（具体行动步骤）。建议要具体可操作，不能是空洞口号。"
        },
        {
            "index": 15,
            "type": "section_divider",
            "number": "06",
            "title": "结论展望",
            "subtitle": "总结与未来",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心结论1",
                "02 核心结论2",
                "03 主要建议：建议描述"
            ],
            "thank_you": "感谢聆听",
            "notes": "总结核心结论和主要建议",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心结论2条+主要建议1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "调研报告要数据说话，有据可查",
        "逻辑要严密，结论要有数据支撑",
        "建议要具体可执行",
        "引用来源要准确，注明出处",
        "问题诊断要客观，不回避",
        "图表要清晰，数据解读要到位",
        "结论要明确，便于决策参考"
    ],
    "data_visualization": {
        "recommended_charts": [
            "市场规模趋势图",
            "竞争格局饼图/条形图",
            "消费者画像雷达图",
            "问题分布帕累托图",
            "建议优先级矩阵图"
        ],
        "chart_design_tips": [
            "图表标题清晰，标注数据来源",
            "颜色统一，视觉协调",
            "数据标签清晰可读",
            "适当使用对比色突出重点"
        ]
    },
    "methodology_appendix": {
        "questionnaire_design": "问卷设计说明",
        "sampling_method": "抽样方法说明",
        "confidence_level": "置信水平说明",
        "data_validation": "数据验证方法说明"
    }
}
