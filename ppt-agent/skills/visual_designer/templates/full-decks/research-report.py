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
            "title": "中国智能家居市场发展调研报告",
            "subtitle": "2024-2025年市场趋势与竞争格局分析",
            "author": "战略研究部",
            "date": "2025年3月",
            "notes": "标题页正式，注明调研时间和团队",
            "filling_prompt": "必须填入真实内容：title 为调研报告的名称，subtitle 为概括性副标题，author 为调研团队或负责人，date 为报告完成日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "example_detail",
            "title": "执行摘要",
            "content_type": "example_detail",
            "kicker": "实例 · 执行摘要",
            "lede": "中国智能家居市场进入快速增长期，预计2025年市场规模将突破8000亿元",
            "context_block": "本次调研历时3个月，覆盖全国一线至四线城市，样本量达5000+，通过问卷调查、深度访谈、行业数据收集等多种方式，全面分析了智能家居市场的发展现状、竞争格局和未来趋势。",
            "solution_block": "核心发现：1）市场需求旺盛，但渗透率仍较低，仅为13.7%；2）全屋智能成为新趋势，预计2025年增速达35%；3）国产品牌崛起，华为、小米、海尔占据头部位置；4）渠道多元化，线上线下融合加速。建议企业加大全屋智能投入，优化渠道布局。",
            "metrics": [
                {"value": "5000+", "label": "调研样本量", "trend": "覆盖各级城市"},
                {"value": "50+", "label": "深度访谈数", "trend": "行业专家+消费者"},
                {"value": "30+", "label": "数据来源", "trend": "权威机构+一手数据"}
            ],
            "takeaway": "启示：智能家居市场机遇与挑战并存，把握全屋智能趋势将是制胜关键",
            "notes": "一页概括报告核心发现和建议",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心发现；context_block 描述调研背景和问题（1-2句话）；solution_block 总结主要发现和建议方向（2-3句话）；metrics_grid 提供3个调研指标（如样本量、发现数量、建议数量），每个有 value、label、trend；takeaway 用一句话总结调研对决策的意义。"
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
                "market_trend": "智能家居市场持续增长，消费者接受度提升",
                "business_need": "公司计划进入智能家居领域，需要了解市场现状和竞争格局",
                "research_goal": "为战略决策提供数据支撑和洞察"
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
                {"name": "问卷调查", "desc": "线上+线下，覆盖5000+消费者"},
                {"name": "深度访谈", "desc": "50+场，包含行业专家、企业高管、经销商"},
                {"name": "数据分析", "desc": "行业报告、企业财报、电商数据"},
                {"name": "竞品分析", "desc": "20+品牌产品体验和对比"}
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
                {"value": "7500亿", "label": "2024年市场规模", "delta": "↑ 18%", "baseline": "vs 2023年"},
                {"value": "13.7%", "label": "家庭渗透率", "delta": "↑ 3.2%", "baseline": "vs 2023年"},
                {"value": "35%", "label": "全屋智能增速", "delta": "↑ 显著", "baseline": "vs 整体市场"},
                {"value": "58%", "label": "消费者认知度", "delta": "↑ 12%", "baseline": "vs 2023年"}
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
            "header": "全屋智能成为消费升级新方向",
            "sub_header": "从单品智能向全屋智能演进",
            "paragraph": "调研发现，智能家居市场正在经历从单品智能向全屋智能的升级。消费者不再满足于单个智能产品的体验，而是期望实现全屋设备的互联互通。数据显示，已有15%的智能家居用户表示会在未来1-2年内升级为全屋智能方案，这一比例在年轻消费群体（25-35岁）中更高，达到23%。全屋智能的客单价是单品的5-8倍，但用户满意度和复购意愿显著更高。这为智能家居企业提供了从产品销售向解决方案服务转型的重要机遇。",
            "references": [
                "https://www.caict.ac.cn/",
                "https://www.iresearch.com.cn/",
                "https://www.gartner.com/"
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
            "header": "标准不统一、体验碎片化成最大痛点",
            "sub_header": "42%用户因互联互通问题放弃智能家居",
            "paragraph": "尽管智能家居市场增长迅速，但调研也发现了制约行业发展的核心问题。首要是标准不统一问题：当前市场存在多个生态体系（如小米、华为、苹果HomeKit等），不同品牌设备难以互联互通。42%的受访者表示，这是他们不愿意使用智能家居的主要原因。其次是用户体验问题：产品操作复杂、学习成本高、售后响应慢等问题也较为突出。再次是数据安全隐患：35%的用户对智能家居的数据安全表示担忧。这些问题如不解决，将制约市场的进一步发展。",
            "references": [
                "https://www.cyberpolice.cn/",
                "https://www.isc.org.cn/",
                "https://www.miit.gov.cn/"
            ],
            "notes": "左侧配问题关系图/风险矩阵图，右侧分析问题详情",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最核心的一个问题（不超过35字）；sub_header 为问题的严重程度评估（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述问题的具体表现、产生原因和潜在影响，用流畅的段落形式呈现，禁止罗列要点，必须包含影响数据（如占比、损失金额、影响范围等）。references 逐条列出 URL 并标注来源。禁止模糊描述'存在一些问题'等空泛表述。"
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
                {"num": "01", "title": "产品策略", "desc": "聚焦全屋智能解决方案"},
                {"num": "02", "title": "技术投入", "desc": "加强Matter协议兼容性"},
                {"num": "03", "title": "渠道建设", "desc": "线上线下融合体验"},
                {"num": "04", "title": "服务升级", "desc": "完善安装和售后服务"}
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
                "01 中国智能家居市场潜力巨大，全屋智能是未来趋势",
                "02 标准统一和用户体验优化是行业发展的关键",
                "03 建议聚焦全屋智能，加强生态兼容和服务能力"
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
        "questionnaire_design": "问卷包含20道选择题，涵盖消费行为、使用体验、购买意愿等维度",
        "sampling_method": "配额抽样，确保性别、年龄、城市级别分布均衡",
        "confidence_level": "95%置信水平，抽样误差±3%",
        "data_validation": "交叉验证、异常值处理、缺失值填补"
    }
}
