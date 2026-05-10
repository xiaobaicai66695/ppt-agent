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
            "title": "智能客服系统建设项目提案",
            "subtitle": "提升客户服务质量，降低运营成本",
            "author": "产品研发部",
            "date": "2025年3月15日",
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
            "header": "客户服务面临巨大挑战",
            "sub_header": "人工客服效率低、成本高、体验差",
            "paragraph": "随着公司业务规模快速增长，客户服务部门面临越来越大的压力。现有人工客服团队100人，年人工成本约1200万元，但客户满意度持续下降，NPS评分从年初的45分下滑至38分。问题根源在于：1）高峰期咨询量暴增，人工难以承接；2）常见问题重复解答，占用大量人力；3）服务时间受限，无法提供7×24小时支持。我们希望通过引入智能客服系统，从根本上解决这些问题。",
            "references": [
                "https://www.forrester.com/",
                "https://www.gartner.com/"
            ],
            "notes": "用图文混排说明项目背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'背景'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景主题（不超过35字）；sub_header 为背景说明；paragraph 为300-450字的自然语言段落，详细描述项目背景、问题现状和立项必要性，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
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
                {"header": "提升响应效率", "body": "当前平均响应时间15分钟，高峰期超过30分钟，客户投诉率上升50%"},
                {"header": "降低运营成本", "body": "人工客服成本年增长20%，需要通过自动化减少人力投入"},
                {"header": "扩展服务时间", "body": "目前仅支持9:00-21:00服务，夜间的咨询无法及时响应"},
                {"header": "统一服务标准", "body": "人工客服水平参差不齐，服务质量难以标准化"}
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
            "header": "智能客服系统整体方案",
            "sub_header": "AI驱动+人机协作",
            "paragraph": "本项目计划建设一套智能客服系统，核心能力包括：1）智能问答：基于大模型技术，自动解答80%以上的常见问题；2）人机协作：复杂问题智能转人工，并推送上下文信息；3）知识库管理：可视化知识库配置，支持快速更新；4）数据分析：实时监控服务数据，洞察客户需求。系统采用云原生架构，支持弹性扩缩容，部署周期预计3个月。",
            "notes": "用图文混排说明解决方案",
            "filling_prompt": "必须填入真实内容：kicker 为'方案'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为方案名称（不超过35字）；sub_header 为方案简介；paragraph 为300-450字的自然语言段落，详细描述解决方案的技术架构、实施路径和预期效果，用流畅的段落形式呈现，禁止罗列要点。"
        },
        {
            "index": 9,
            "type": "process_flow",
            "title": "实施路径",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "需求调研", "desc": "业务调研、场景梳理、知识收集", "timeline": "第1-2周"},
                {"num": "02", "title": "系统开发", "desc": "核心功能开发、模型训练、接口对接", "timeline": "第3-8周"},
                {"num": "03", "title": "测试验收", "desc": "功能测试、性能测试、用户验收", "timeline": "第9-10周"},
                {"num": "04", "title": "上线运营", "desc": "灰度发布、全量上线、持续优化", "timeline": "第11-12周"}
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
                {"date": "Week 2", "event": "需求确认", "deliverable": "需求文档签字确认"},
                {"date": "Week 6", "event": "核心功能完成", "deliverable": "V1.0版本发布"},
                {"date": "Week 10", "event": "UAT完成", "deliverable": "用户验收通过"},
                {"date": "Week 12", "event": "正式上线", "deliverable": "全量发布，生产运行"},
                {"date": "Month 6", "event": "运营优化", "deliverable": "首期运营优化完成"}
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
                {"value": "180万", "label": "系统建设费", "delta": "占比55%", "baseline": "总预算328万"},
                {"value": "80万", "label": "年度运维费", "delta": "占比25%", "baseline": "含云资源和人力"},
                {"value": "40万", "label": "培训推广费", "delta": "占比12%", "baseline": "含培训和推广"},
                {"value": "28万", "label": "风险储备金", "delta": "占比8%", "baseline": "总预算328万"}
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
            "lede": "智能客服系统预计年节省成本500万元，投资回报期仅8个月",
            "context_block": "当前客服团队100人，年成本1200万元。即使引入系统后需要保留50人，也能节省一半人力成本。同时，响应速度提升将显著改善客户体验，带来潜在的收入增长。",
            "solution_block": "项目上线后预期收益包括：1）人力成本节省：减少50个客服坐席，年节省600万元；2）效率提升：机器人承接80%咨询，人工仅处理20%复杂问题；3）收入增长：客户满意度提升预计带来5%复购增长；4）品牌提升：7×24小时服务提升品牌形象。综合ROI达到320%。",
            "metrics": [
                {"value": "600万/年", "label": "人力成本节省", "trend": "vs 投入328万"},
                {"value": "320%", "label": "3年ROI", "trend": "显著"},
                {"value": "92分", "label": "预期满意度", "trend": "vs 当前38分"}
            ],
            "takeaway": "启示：项目投入产出比优秀，能够快速实现价值交付",
            "notes": "列出项目的预期收益",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：lede 一句话概括最重要收益；context_block 描述实施前现状和痛点（1-2句话）；solution_block 具体说明实施后的收益和价值（2-3句话）；metrics_grid 提供3个收益指标，每个有 value、label、trend；takeaway 用一句话总结核心价值。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 智能客服系统能够解决当前服务困境",
                "02 项目预算328万，ROI 320%，回报期8个月",
                "03 请领导批准立项，我们有信心交付成功"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁总结，明确资源请求",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心价值+关键优势+资源请求）。禁止保留花括号。"
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
            "机器人解答率≥80%",
            "客户满意度≥85分",
            "成本节省≥40%",
            "系统可用性≥99.9%"
        ]
    }
}
