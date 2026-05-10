TEMPLATE = {
    "name": "product-intro",
    "name_cn": "产品介绍",
    "description": "适合产品介绍、客户演示、功能展示等场景。突出价值，展示功能，增强信任。",
    "target_audience": "客户、合作伙伴、销售团队",
    "typical_slides": 15,
    "typical_duration": "15-20分钟",
    "palette": "warm_terracotta",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "opening_hook": "用一个客户痛点或成功故事开场，建立情感连接",
        "key_moments": [
            "产品价值：强调能为客户解决什么问题",
            "核心功能：用具体场景展示功能价值",
            "客户案例：用真实故事增加可信度",
            "竞争优势：差异化要清晰明确"
        ],
        "closing_strength": "明确的行动号召（预约演示/获取试用）"
    },
    "value_proposition_tips": {
        "focus_on_outcomes": "聚焦客户收益，而非产品功能",
        "use_stories": "用故事而非数据来建立情感连接",
        "quantify_impact": "尽可能量化业务影响",
        "address_objections": "预见并回应潜在疑虑"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "企业级智能客服系统",
            "subtitle": "让每一次客户交互都成为增长机会",
            "author": "王芳 | 售前顾问",
            "date": "2025年3月15日",
            "notes": "标题页醒目，展示产品名称和价值主张",
            "filling_prompt": "必须填入真实内容：title 为产品名称，subtitle 为产品定位或一句话价值主张，author 为主讲人姓名，date 为演示日期。禁止保留花括号。",
            "visual_suggestions": "产品Logo居中，配合抽象的客户交互图形"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  产品概述",
                "02  核心功能",
                "03  应用场景",
                "04  客户案例",
                "05  产品优势"
            ],
            "notes": "让听众了解产品介绍结构",
            "filling_prompt": "目录页为固定结构。",
            "timing_hint": "约15秒，快速翻过"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "产品概述",
            "subtitle": "产品定位与价值",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "产品概览",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "产品定位",
            "header": "新一代智能客服解决方案",
            "sub_header": "AI驱动 • 全渠道接入 • 敏捷部署",
            "paragraph": "企业级智能客服系统是一套基于大语言模型的智能客服解决方案，旨在帮助企业提升客户服务质量、降低运营成本。系统支持网页、APP、微信、企微等多渠道接入，7×24小时为客户提供即时服务。通过自然语言理解和知识图谱技术，系统能够准确理解客户意图，自动解答80%以上的常见问题；对于复杂问题，系统会智能识别并转接人工客服，同时推送相关上下文信息，确保服务无缝衔接。目前已服务超过2000家企业，客户满意度平均提升35%。",
            "notes": "用图文混排展示产品主视觉和核心价值",
            "filling_prompt": "必须填入真实内容：kicker 为产品领域标签；title 为产品名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为一句话核心价值主张（不超过35字）；sub_header 为产品定位说明；paragraph 为300-450字的自然语言段落，详细描述产品的核心价值、功能特点和适用场景，用流畅的段落形式呈现，禁止罗列要点。",
            "visual_tip": "右侧可展示产品界面截图或功能演示图"
        },
        {
            "index": 5,
            "type": "kpi_dashboard",
            "title": "产品数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 产品表现",
            "kpis": [
                {"value": "2000+", "label": "服务企业数", "delta": "↑ 持续增长", "baseline": "截至2025年Q1"},
                {"value": "80%+", "label": "问题自动解答率", "delta": "↑ 行业领先", "baseline": "vs 行业平均60%"},
                {"value": "35%", "label": "客户满意度提升", "delta": "↑ 显著", "baseline": "vs 实施前"},
                {"value": "< 5秒", "label": "平均响应时间", "delta": "↓ 90%", "baseline": "vs 人工客服"}
            ],
            "notes": "用数据展示产品市场表现",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个产品相关指标（如用户数、功能数、覆盖率、客户满意度等），每个有 value、label、delta、baseline。references 列出 URL。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "核心功能",
            "subtitle": "主要功能模块",
            "filling_prompt": "固定内容，无需额外填充。",
            "transition_phrase": "接下来，让我为大家详细介绍我们的核心功能。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "核心功能",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {
                    "header": "智能问答",
                    "body": "基于大模型技术，准确理解客户问题，自动生成专业回答，支持多轮对话和上下文记忆"
                },
                {
                    "header": "知识库管理",
                    "body": "可视化知识库配置，支持文档自动学习，支持FAQ、结构化知识等多种知识形态"
                },
                {
                    "header": "人机协作",
                    "body": "智能工单分配，话术推荐，情绪识别，确保复杂问题无缝转接人工"
                },
                {
                    "header": "数据分析",
                    "body": "实时监控客服数据，多维度报表分析，洞察客户需求，优化服务策略"
                }
            ],
            "notes": "展示4个核心功能模块",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称，不超过35字）和 body（详细描述该功能的核心价值和应用场景，100-120字，包含具体使用效果或数据）。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "功能详解：智能问答引擎",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "功能详解",
            "header": "AI驱动的智能问答引擎",
            "sub_header": "准确率95%以上的智能理解能力",
            "paragraph": "智能问答引擎是产品的核心能力，它基于最新的大语言模型技术，能够准确理解客户自然语言表达的意图。与传统关键词匹配不同，我们的引擎能够理解上下文语义、处理复杂句式、甚至识别用户的潜在需求。系统支持99个行业的专属知识库预训练，上线首日即可达到85%的回答准确率，经过一周的自学习，准确率可提升至95%以上。引擎还支持意图识别、实体提取、情绪分析等高级功能，为后续的精准服务提供数据支撑。",
            "notes": "用图文混排详解一个核心功能",
            "filling_prompt": "必须填入真实内容：kicker 为'功能详解'；title 中的 {功能名称} 替换为具体功能；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为功能名称（不超过35字）；sub_header 为功能简介；paragraph 为300-450字的自然语言段落，详细描述该功能的工作原理、使用步骤、关键优势和实际效果，用流畅的段落形式呈现，禁止罗列要点。",
            "feature_highlights": ["准确率95%+", "行业知识预训练", "7×24小时服务", "多轮对话支持"]
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "应用场景",
            "subtitle": "典型使用场景",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "典型应用场景",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "应用场景",
            "header": "全行业客户支持场景",
            "sub_header": "电商、金融、政务、教育全覆盖",
            "paragraph": "我们的产品已在多个行业得到成功应用。在电商场景，我们帮助某头部电商平台实现了客服效率提升300%，大促期间从容应对10倍流量峰值。在金融场景，我们为某股份制银行打造的智能客服，单日处理咨询量超过50万次，问题解答准确率达97%。在政务场景，我们承建的12345智能热线，日均服务群众10万人次，群众满意度提升至92%。产品支持快速私有化部署，满足企业对数据安全的要求。",
            "notes": "用图文混排展示应用场景",
            "filling_prompt": "必须填入真实内容：kicker 为'应用场景'；title 为场景名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为场景标题（不超过35字）；sub_header 为场景描述；paragraph 为300-450字的自然语言段落，详细描述该应用场景的具体情况、用户痛点、解决方案和使用效果，用流畅的段落形式呈现，禁止罗列要点。",
            "industry_coverage": ["电商零售", "金融服务", "政务民生", "教育培训"]
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "客户案例",
            "subtitle": "成功案例展示",
            "filling_prompt": "固定内容，无需额外填充。",
            "transition_phrase": "下面，让我分享一个具体的客户案例。"
        },
        {
            "index": 12,
            "type": "image_text",
            "title": "客户案例：某头部电商平台",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "电商行业",
            "header": "某头部电商平台智能客服升级",
            "sub_header": "日均订单500万+的大促保障",
            "paragraph": "该电商平台原有客服系统采用传统FAQ+关键词匹配，日均服务客户30万人次，人工客服坐席2000人，但大促期间仍出现严重排队，平均等待时间超过15分钟，客户投诉率居高不下。引入我们的智能客服后，系统自动承接80%的常见问题，人工客服只需处理复杂问题。实测数据显示：大促期间平均响应时间从15分钟降至30秒，客户满意度从75%提升至92%，人工客服需求减少60%，每年节省人力成本约1200万元。",
            "references": [
                "https://www.aliyun.com/",
                "https://www Forrester.com/"
            ],
            "notes": "用图文混排展示一个成功案例",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {行业} 替换为客户所在行业；title 中的 {客户名称} 替换为真实客户名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为客户名称（不超过35字）；sub_header 为合作项目名称；paragraph 为300-450字的自然语言段落，详细描述客户背景、合作过程、实施效果和客户评价，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 13,
            "type": "kpi_dashboard",
            "title": "客户成效",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 客户成效",
            "kpis": [
                {"value": "↓ 95%", "label": "平均响应时间降低", "delta": "↓ 15分钟→30秒", "baseline": "vs 实施前"},
                {"value": "↑ 23%", "label": "客户满意度提升", "delta": "↑ 23%", "baseline": "vs 实施前"},
                {"value": "↓ 60%", "label": "人工客服需求减少", "delta": "↓ 60%", "baseline": "vs 实施前"},
                {"value": "1200万/年", "label": "年节省人力成本", "delta": "↓ 40%", "baseline": "vs 实施前"}
            ],
            "notes": "展示客户实施后的具体成效数据",
            "filling_prompt": "必须填入真实内容：提供4个客户成效指标，每个有 value、label、delta、baseline。",
            "visual_tip": "可配合对比柱状图展示前后变化"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "05",
            "title": "产品优势",
            "subtitle": "为什么选择我们",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 15,
            "type": "example_detail",
            "title": "核心优势",
            "content_type": "example_detail",
            "kicker": "实例 · 核心优势",
            "lede": "自研大模型+行业知识图谱，提供业内领先的智能客服解决方案",
            "context_block": "当前市场上智能客服产品同质化严重，大多基于通用大模型，缺乏行业深度定制能力。企业客户在选择时，往往面临准确率不够、私有化部署复杂、运维成本高等问题。",
            "solution_block": "我们的核心优势在于：1）自研客服领域大模型，经过2000+企业数据训练，专为客服场景优化；2）首创知识图谱+大模型融合架构，回答准确率比纯大模型方案提升30%；3）支持SaaS和私有化一键部署，最快3天上线；4）7×24小时专属客服，成功率98%以上。",
            "metrics": [
                {"value": "95%+", "label": "回答准确率", "trend": "vs 竞品80%"},
                {"value": "3天", "label": "最快上线周期", "trend": "vs 行业30天"},
                {"value": "98%+", "label": "客服成功率", "trend": "行业领先"}
            ],
            "takeaway": "启示：选择我们，就是选择专业、安心、省心",
            "notes": "总结产品核心竞争优势",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心优势；context_block 简要说明目标市场和用户需求（1-2句话）；solution_block 详细展开核心优势和技术亮点（2-3句话）；metrics_grid 提供3个优势指标；takeaway 用一句话总结差异化竞争壁垒。禁止保留花括号。",
            "competitive_differentiators": ["自研领域大模型", "知识图谱融合", "极速部署上线", "专属服务支持"]
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 智能客服：效率提升300%，满意度提升23%",
                "02 行业领先：准确率95%+，最快3天上线",
                "03 立即行动：扫码预约产品演示，获取专属方案"
            ],
            "thank_you": "感谢聆听！",
            "contact": "官网：www.example.com | 热线：400-xxx-xxxx | 扫码咨询",
            "notes": "简洁有力的结尾，附上联系方式",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心价值+关键优势+联系方式/下一步）。禁止保留花括号。",
            "cta_suggestion": "提供二维码方便观众扫码预约演示"
        }
    ],
    "design_tips": [
        "产品介绍要突出价值而非功能",
        "案例要真实可信，注明来源",
        "图文混排增强说服力",
        "数据说话，展示市场认可",
        "结尾要有明确的行动号召",
        "用客户语言而非技术语言",
        "强调ROI和业务价值",
        "准备好应对常见异议"
    ],
    "presentation_flow": {
        "opening": {
            "duration": "1-2分钟",
            "goal": "建立连接，引起兴趣",
            "tip": "用一个客户痛点或成功故事开场"
        },
        "body": {
            "duration": "12-15分钟",
            "goal": "展示价值，建立信任",
            "tip": "产品概述→功能详解→案例验证→优势总结"
        },
        "closing": {
            "duration": "2-3分钟",
            "goal": "促成行动",
            "tip": "明确CTA，提供联系方式"
        }
    },
    "objection_handling": {
        "common_objections": [
            "价格太高 → 强调ROI和长期价值",
            "实施周期长 → 强调3天上线能力",
            "数据安全 → 强调私有化部署和数据隔离",
            "准确率不够 → 展示测试数据和客户案例"
        ],
        "proof_points": [
            "第三方评测报告",
            "客户推荐信",
            "POC测试数据",
            "行业白皮书"
        ]
    }
}
