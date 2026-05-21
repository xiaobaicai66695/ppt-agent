TEMPLATE = {
    "name": "project-proposal",
    "name_cn": "项目提案",
    "description": "适合新项目立项、项目申请、资源申请等场景。理由充分，方案可行，预算清晰。",
    "target_audience": "领导、评审委员会、项目审批方",
    "typical_slides": 16,
    "typical_duration": "15-20分钟",
    "palette": "charcoal_light",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "proposal_strategy": "提案要解决'为什么要做'和'为什么是你们'两个核心问题",
        "key_elements": [
            "问题足够痛吗？",
            "方案足够好吗？",
            "时机对吗？",
            "预算合理吗？",
            "能落地执行吗？"
        ],
        "persuasion_focus": [
            "用数据和事实说话",
            "展示团队执行能力",
            "明确风险和应对",
            "提供清晰的ROI"
        ]
    },
    "decision_criteria": {
        "strategic_fit": "项目与公司战略的契合度",
        "business_value": "预期收益和价值",
        "feasibility": "技术可行性和资源需求",
        "risk_level": "风险可控性",
        "roi": "投资回报率"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "项目名称（项目提案）",
            "subtitle": "项目定位或一句话说明",
            "author": "提案人/团队名称",
            "date": "提案日期",
            "notes": "标题页正式，清晰标注项目名称",
            "filling_prompt": "必须填入真实内容：title 为项目名称+项目提案，subtitle 为项目定位或一句话说明，author 为提案人或团队名称，date 为提案日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  项目背景",
                "02  需求分析",
                "03  解决方案",
                "04  项目计划",
                "05  预算估算",
                "06  预期效果"
            ],
            "notes": "让评审快速了解提案结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "项目背景",
            "subtitle": "为什么需要这个项目",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "项目背景",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景",
            "header": "背景主题",
            "sub_header": "背景说明",
            "paragraph": "详细描述项目背景、问题现状和立项必要性，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排说明项目背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景主题（不超过35字）；sub_header 为背景说明；paragraph 为300-450字的自然语言段落，详细描述项目背景、问题现状和立项必要性，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "需求分析",
            "subtitle": "要解决什么问题",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "card_grid",
            "title": "需求分析",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "需求名称（不超过35字）", "body": "详细描述该需求的具体表现和影响（100-120字），包含具体数据或场景。"},
                {"header": "需求名称（不超过35字）", "body": "详细描述该需求的具体表现和影响（100-120字），包含具体数据或场景。"},
                {"header": "需求名称（不超过35字）", "body": "详细描述该需求的具体表现和影响（100-120字），包含具体数据或场景。"},
                {"header": "需求名称（不超过35字）", "body": "详细描述该需求的具体表现和影响（100-120字），包含具体数据或场景。"}
            ],
            "notes": "列出4个主要需求或痛点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个需求/痛点，每个有 header（需求名称，不超过35字）和 body（详细描述该需求的具体表现和影响，100-120字，包含具体数据或场景）。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "解决方案",
            "subtitle": "如何解决这个问题",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "解决方案",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "方案",
            "header": "方案名称",
            "sub_header": "方案简介",
            "paragraph": "详细描述解决方案的技术架构、实施路径和预期效果，用流畅的段落形式呈现，禁止罗列要点。",
            "notes": "用图文混排说明解决方案",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为方案名称（不超过35字）；sub_header 为方案简介；paragraph 为300-450字的自然语言段落，详细描述解决方案的技术架构、实施路径和预期效果，用流畅的段落形式呈现，禁止罗列要点。"
        },
        {
            "index": 9,
            "type": "process_flow",
            "title": "实施路径",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "阶段名称", "desc": "阶段说明（不超过35字）"},
                {"num": "02", "title": "阶段名称", "desc": "阶段说明（不超过35字）"},
                {"num": "03", "title": "阶段名称", "desc": "阶段说明（不超过35字）"},
                {"num": "04", "title": "阶段名称", "desc": "阶段说明（不超过35字）"}
            ],
            "notes": "用流程图展示实施阶段",
            "filling_prompt": "必须填入真实内容：提供4个实施阶段，每个有 title（阶段名称，不超过35字）和 desc（阶段说明，不超过35字）。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "项目计划",
            "subtitle": "时间表与里程碑",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "timeline",
            "title": "项目里程碑",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "milestones": [
                {"date": "时间", "event": "里程碑名称", "deliverable": "交付物"},
                {"date": "时间", "event": "里程碑名称", "deliverable": "交付物"},
                {"date": "时间", "event": "里程碑名称", "deliverable": "交付物"},
                {"date": "时间", "event": "里程碑名称", "deliverable": "交付物"}
            ],
            "notes": "展示项目关键里程碑",
            "filling_prompt": "必须填入真实内容：提供4-5个里程碑节点，每个有具体时间、里程碑名称和交付物说明。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "预算估算",
            "subtitle": "资源需求",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "kpi_dashboard",
            "title": "预算概览",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "预算 · 费用明细",
            "kpis": [
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "总预算"},
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "总预算"},
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "总预算"},
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "总预算"}
            ],
            "notes": "展示预算分配",
            "filling_prompt": "必须填入真实内容：提供4个预算类别，每个有 value（金额）、label（类别名称）、delta（占比）、baseline（总预算）。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "06",
            "title": "预期效果",
            "subtitle": "项目价值",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 15,
            "type": "example_detail",
            "title": "预期收益",
            "content_type": "example_detail",
            "kicker": "实例 · 预期收益",
            "lede": "一句话概括最重要收益",
            "context_block": "描述实施前现状和痛点（1-2句话）。",
            "solution_block": "具体说明实施后的收益和价值（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "收益指标", "trend": "变化趋势"},
                {"value": "数字", "label": "收益指标", "trend": "变化趋势"},
                {"value": "数字", "label": "收益指标", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结核心价值。",
            "notes": "列出项目的预期收益",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：lede 一句话概括最重要收益；context_block 描述实施前现状和痛点（1-2句话）；solution_block 具体说明实施后的收益和价值（2-3句话）；metrics_grid 提供3个收益指标，每个有 value、label、trend；takeaway 用一句话总结核心价值。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心价值：描述",
                "02 关键数据：描述",
                "03 资源请求：描述"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁总结，明确资源请求",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心价值+关键数据+资源请求）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "提案理由要充分，数据说话",
        "方案要可行，避免过度承诺",
        "预算要合理，有明细有依据",
        "预期效果要可衡量可验证",
        "PPT设计正式稳重，体现专业性",
        "注意逻辑清晰，层次分明",
        "预判评审问题，准备好应答",
        "时间控制适当，留出答疑时间"
    ],
    "review_preparation": {
        "likely_questions": [
            "为什么不用现有方案？",
            "预算构成是否合理？",
            "项目风险如何控制？",
            "如何衡量项目成功？"
        ],
        "success_criteria": [
            "指标1",
            "指标2",
            "指标3",
            "指标4"
        ]
    }
}
