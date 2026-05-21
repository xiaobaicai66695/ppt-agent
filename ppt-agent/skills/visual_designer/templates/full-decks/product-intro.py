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
            "title": "产品名称",
            "subtitle": "产品定位或一句话价值主张",
            "author": "主讲人姓名",
            "date": "演示日期",
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
            "header": "一句话核心价值主张",
            "sub_header": "产品定位说明",
            "paragraph": "详细描述产品的核心价值、功能特点和适用场景，用流畅的段落形式呈现，禁止罗列要点。",
            "notes": "用图文混排展示产品主视觉和核心价值",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为一句话核心价值主张（不超过35字）；sub_header 为产品定位说明；paragraph 为300-450字的自然语言段落，详细描述产品的核心价值、功能特点和适用场景，用流畅的段落形式呈现，禁止罗列要点。"
        },
        {
            "index": 5,
            "type": "kpi_dashboard",
            "title": "产品数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 产品表现",
            "kpis": [
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"}
            ],
            "notes": "用数据展示产品市场表现",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个产品相关指标，每个有 value、label、delta、baseline。references 列出 URL。"
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
                {"header": "功能名称（不超过35字）", "body": "详细描述该功能的核心价值和应用场景（100-120字），包含具体使用效果或数据。"},
                {"header": "功能名称（不超过35字）", "body": "详细描述该功能的核心价值和应用场景（100-120字），包含具体使用效果或数据。"},
                {"header": "功能名称（不超过35字）", "body": "详细描述该功能的核心价值和应用场景（100-120字），包含具体使用效果或数据。"},
                {"header": "功能名称（不超过35字）", "body": "详细描述该功能的核心价值和应用场景（100-120字），包含具体使用效果或数据。"}
            ],
            "notes": "展示4个核心功能模块",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称，不超过35字）和 body（详细描述该功能的核心价值和应用场景，100-120字，包含具体使用效果或数据）。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "功能详解",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "功能详解",
            "header": "功能名称",
            "sub_header": "功能简介",
            "paragraph": "详细描述该功能的工作原理、使用步骤、关键优势和实际效果，用流畅的段落形式呈现，禁止罗列要点。",
            "notes": "用图文混排详解一个核心功能",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为功能名称（不超过35字）；sub_header 为功能简介；paragraph 为300-450字的自然语言段落，详细描述该功能的工作原理、使用步骤、关键优势和实际效果，用流畅的段落形式呈现，禁止罗列要点。"
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
            "header": "场景标题",
            "sub_header": "场景描述",
            "paragraph": "详细描述该应用场景的具体情况、用户痛点、解决方案和使用效果，用流畅的段落形式呈现，禁止罗列要点。",
            "notes": "用图文混排展示应用场景",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为场景标题（不超过35字）；sub_header 为场景描述；paragraph 为300-450字的自然语言段落，详细描述该应用场景的具体情况、用户痛点、解决方案和使用效果，用流畅的段落形式呈现，禁止罗列要点。"
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
            "title": "客户案例",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "行业",
            "header": "客户名称或项目名称",
            "sub_header": "合作项目名称",
            "paragraph": "详细描述客户背景、合作过程、实施效果和客户评价，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示一个成功案例",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为客户名称（不超过35字）；sub_header 为合作项目名称；paragraph 为300-450字的自然语言段落，详细描述客户背景、合作过程、实施效果和客户评价，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 13,
            "type": "kpi_dashboard",
            "title": "客户成效",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 客户成效",
            "kpis": [
                {"value": "数字", "label": "效果指标", "delta": "变化幅度", "baseline": "vs 实施前"},
                {"value": "数字", "label": "效果指标", "delta": "变化幅度", "baseline": "vs 实施前"},
                {"value": "数字", "label": "效果指标", "delta": "变化幅度", "baseline": "vs 实施前"},
                {"value": "数字", "label": "成本/收益", "delta": "变化幅度", "baseline": "vs 实施前"}
            ],
            "notes": "展示客户实施后的具体成效数据",
            "filling_prompt": "必须填入真实内容：提供4个客户成效指标，每个有 value、label、delta、baseline。"
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
            "lede": "一句话概括核心优势",
            "context_block": "简要说明目标市场和用户需求（1-2句话）。",
            "solution_block": "详细展开核心优势和技术亮点（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "优势指标", "trend": "变化趋势"},
                {"value": "数字", "label": "优势指标", "trend": "变化趋势"},
                {"value": "数字", "label": "优势指标", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结差异化竞争壁垒。",
            "notes": "总结产品核心竞争优势",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心优势；context_block 简要说明目标市场和用户需求（1-2句话）；solution_block 详细展开核心优势和技术亮点（2-3句话）；metrics_grid 提供3个优势指标；takeaway 用一句话总结差异化竞争壁垒。禁止保留花括号。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心价值：描述",
                "02 关键优势：描述",
                "03 行动号召：描述"
            ],
            "thank_you": "感谢聆听！",
            "contact": "官网 | 热线 | 扫码咨询",
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
            "异议1 → 回应方式",
            "异议2 → 回应方式",
            "异议3 → 回应方式",
            "异议4 → 回应方式"
        ],
        "proof_points": [
            "第三方评测报告",
            "客户推荐信",
            "POC测试数据",
            "行业白皮书"
        ]
    }
}
