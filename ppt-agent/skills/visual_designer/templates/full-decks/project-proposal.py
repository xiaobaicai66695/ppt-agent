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
            "title": "智慧工厂数据中台建设项目提案",
            "subtitle": "打通数据孤岛，构建统一数据资产，赋能精益生产",
            "author": "数字化转型项目组",
            "date": "2025年6月",
            "notes": "封面页正式庄重，清晰标注项目名称和提案团队",
            "filling_prompt": "本模板为固定格式，title 填入项目名称，subtitle 填入项目定位或一句话说明，author 填入提案人或团队名称，date 填入提案日期。禁止保留花括号。"
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
            "notes": "让评审快速了解提案结构，六大章节一览无余",
            "filling_prompt": "目录页为固定结构，根据实际章节内容调整编号和章节名称。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "项目背景",
            "subtitle": "为什么需要这个项目",
            "notes": "章节分隔页，标注第一章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "业务痛点与立项背景",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景",
            "header": "数据孤岛林立，决策缺乏依据",
            "sub_header": "六大产线系统数据标准不一，信息流断点频发",
            "paragraph": "华东智能制造基地目前运行着ERP、MES、WMS、SCADA、QMS、TPM等六套核心系统，但各系统由不同供应商在不同年份建设，缺乏统一的数据标准和技术接口。生产部门、财务部门和采购部门同一份基础数据需要重复录入，数据一致性和时效性无法保障。2024年全年，因数据不一致导致的排产错误、库存呆滞和质量问题追溯成本合计超过1800万元。与此同时，管理层缺少实时的生产运营数据看板，决策依赖日报周报，响应滞后至少2-3天。",
            "references": [
                "https://www.mckinsey.com/business-functions/operations/our-insights  麦肯锡数字化制造白皮书2024",
                "https://www.gartner.com/en/digital-markets  Gartner工业数据平台魔力象限2024"
            ],
            "notes": "右侧配现有系统架构图或数据断点示意图，左侧呈现核心痛点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 填入最核心的一个背景问题（不超过35字）；sub_header 填入具体表现说明（不超过35字）；paragraph 填入300-450字的自然语言段落，详细描述项目背景、问题现状和立项必要性，包含具体数字，禁止罗列要点。references 逐条列出 URL 并标注来源。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "需求分析",
            "subtitle": "要解决什么问题",
            "notes": "章节分隔页，标注第二章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "card_grid",
            "title": "核心需求梳理",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "kicker": "需求 · 痛点分析",
            "cards": [
                {"header": "统一数据标准与主数据管理", "body": "各系统商品编码、物料编码、客户编码、供应商编码不统一，同一实体在不同系统中有多个编码版本。亟需建立企业级主数据管理平台，统一编码规则和数据标准，从源头解决数据不一致问题。预计涉及编码映射超过12万条，跨系统数据治理工作量巨大。"},
                {"header": "实时生产数据采集与可视化", "body": "现有SCADA系统数据仅保存在本地工控机，无法跨系统共享。产线OEE、良品率、设备稼动率等关键指标依赖人工统计，T+1才能呈现。需求实现各产线设备数据的实时采集、云端汇聚和可视化看板，支持从车间级到工厂级的层层穿透。"},
                {"header": "跨系统业务协同与流程打通", "body": "订单变更无法自动同步至生产计划，采购备料与生产排程脱节，质量异常无法联动追溯。亟需通过数据中台打通核心业务链路，实现以订单为驱动的自动排产、以库存为约束的智能采购、以批次为线索的全流程追溯。"},
                {"header": "历史数据资产化与智能分析", "body": "过去5年积累了超过2TB的生产数据，但分散存储在各自系统中，无人分析利用。需求建设历史数据湖，打通数据资产化最后一公里，为后续AI质量预测、设备预测性维护等高级分析场景奠定基础。"
                }
            ],
            "notes": "四个核心需求，痛点明确，需求具体",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心需求/痛点，每个有 header（需求名称，不超过35字）和 body（详细描述，100-120字，包含具体数据或场景）。禁止虚构或使用通用化描述。references 逐条列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "解决方案",
            "subtitle": "如何解决这个问题",
            "notes": "章节分隔页，标注第三章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "two_column",
            "title": "解决方案概述",
            "content_type": "two_column",
            "layout_hint": "left-right",
            "kicker": "方案",
            "header": "数据中台：统一底座，赋能业务",
            "left_title": "现状问题",
            "left_content": [
                "六套异构系统，数据标准各异",
                "数据分散存储，无法统一分析",
                "业务链路断点，协同效率低下",
                "人工统计报表，响应严重滞后"
            ],
            "right_title": "解决方案",
            "right_content": [
                "建设统一数据中台，制定数据标准",
                "汇聚全量数据，构建企业数据资产",
                "打通核心业务链路，实现数据驱动运营",
                "搭建实时数据看板，支持科学决策"
            ],
            "notes": "左右对比，清晰呈现现状问题和解决方案的对应关系",
            "filling_prompt": "必须填入真实内容：header 填入方案核心定位（不超过35字）；left_title 填入现状问题标题；left_content 填入4条现状痛点（每条不超过25字）；right_title 填入解决方案标题；right_content 填入4条对应解决方案（每条不超过25字）。左右内容一一对应，形成问题-解决方案对照。"
        },
        {
            "index": 9,
            "type": "image_text",
            "title": "技术架构设计",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "架构",
            "header": "四层数据架构，全链路可观测",
            "sub_header": "数据采集层 → 数据治理层 → 数据资产层 → 数据应用层",
            "paragraph": "数据中台采用分层架构设计。数据采集层通过标准API、消息队列和CDC变更数据捕获三种方式，实现对ERP、MES、WMS等业务系统的实时与准实时数据同步。数据治理层提供数据质量监控、元数据管理、数据血缘追踪和敏感数据分级分类四大能力，确保入湖数据质量可追溯、安全合规有保障。数据资产层构建统一的数据目录和标签体系，形成可复用、可组合的数据服务能力。数据应用层面向业务场景，提供自助取数平台、实时大屏、即席分析和开放API四类出口，支撑不同角色用户的数字化需求。",
            "references": [
                "https://aws.amazon.com/cn/data/lake-solution/  AWS数据湖架构最佳实践",
                "https://databricks.com/solutions/data-lakehouse  Databricks湖仓一体架构白皮书"
            ],
            "notes": "左侧配技术架构图，右侧配文字说明",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 填入架构核心理念（不超过35字）；sub_header 填入架构层次说明（不超过35字）；paragraph 填入300-450字的自然语言段落，详细描述技术架构的各层职责、技术选型和模块设计，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "项目计划",
            "subtitle": "时间表与里程碑",
            "notes": "章节分隔页，标注第四章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "timeline",
            "title": "实施路线与里程碑",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "kicker": "计划 · 里程碑",
            "milestones": [
                {"date": "2025年Q3", "event": "数据中台基础平台上线", "deliverable": "完成数据采集层和治理层建设，打通ERP与MES数据链路，发布V1.0数据目录"},
                {"date": "2025年Q4", "event": "核心业务数据汇聚完成", "deliverable": "完成六大系统全量数据接入，数据质量达标率超过95%，上线实时生产看板"},
                {"date": "2026年Q1", "event": "数据资产化与自助分析上线", "deliverable": "建成企业级数据资产目录，上线自助取数平台，数据服务API开放至20个以上"},
                {"date": "2026年Q2", "event": "高级分析与AI场景试点", "deliverable": "试点设备预测性维护和质量预测模型，沉淀3个以上可推广的AI应用场景"}
            ],
            "notes": "四个里程碑，时间跨度约12个月",
            "filling_prompt": "必须填入真实内容：提供4个里程碑节点，每个有 date（具体时间）、event（里程碑名称，不超过35字）和 deliverable（交付物说明，50-80字）。里程碑时间要合理，前后有逻辑递进关系。"
        },
        {
            "index": 12,
            "type": "process_flow",
            "title": "团队与资源",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "kicker": "团队 · 角色分工",
            "steps": [
                {"num": "01", "title": "项目经理", "desc": "统筹项目全局，协调内外部资源，控制进度与风险，向数字化负责人汇报"},
                {"num": "02", "title": "数据架构师", "desc": "负责数据中台整体架构设计，制定数据标准和建模规范，把控数据质量"},
                {"num": "03", "title": "ETL开发工程师", "desc": "负责数据采集、清洗、转换和加载脚本开发，确保数据按时保质入湖"},
                {"num": "04", "title": "BI工程师", "desc": "负责数据看板和报表开发，支撑业务部门自助取数和可视化分析需求"}
            ],
            "notes": "四类核心角色，覆盖项目全生命周期",
            "filling_prompt": "必须填入真实内容：提供4个核心角色，每个有 title（角色名称）和 desc（职责说明，50-80字）。角色要覆盖项目经理、数据架构师、开发实施和业务对接四类职能。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "预算估算",
            "subtitle": "资源需求",
            "notes": "章节分隔页，标注第五章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 14,
            "type": "kpi_dashboard",
            "title": "费用预算明细",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "预算 · 费用明细",
            "kpis": [
                {"value": "280万元", "label": "软件平台", "delta": "占比42%", "baseline": "数据中台软件许可：180万，数据治理工具：60万，BI工具：40万"},
                {"value": "160万元", "label": "实施服务", "delta": "占比24%", "baseline": "数据架构设计：40万，ETL开发：80万，报表开发：40万"},
                {"value": "120万元", "label": "硬件基础设施", "delta": "占比18%", "baseline": "服务器及存储扩容：90万，网络改造：30万"},
                {"value": "110万元", "label": "项目管理与培训", "delta": "占比16%", "baseline": "项目管理：30万，知识转移：40万，运维培训：40万"}
            ],
            "notes": "四个预算类别，总投资估算670万元",
            "filling_prompt": "必须填入真实内容：提供4个预算类别，每个有 value（金额）、label（类别名称）、delta（占总预算比例）、baseline（预算明细说明，30-50字）。各项之和等于总预算，各项比例之和为100%。"
        },
        {
            "index": 15,
            "type": "section_divider",
            "number": "06",
            "title": "预期效果",
            "subtitle": "项目价值",
            "notes": "章节分隔页，标注第六章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "项目总结与行动计划",
            "content_type": "summary_slide",
            "kicker": "总结",
            "key_points": [
                "01 核心价值：打通6大系统数据孤岛，建设统一数据资产平台，预计年化降本增效超过800万元",
                "02 关键数据：总投资670万元，预计18个月回本，IRR达36%，关键里程碑Q3首期上线、Q4全量贯通",
                "03 资源请求：申请项目立项批复，组建4人核心团队，Q3正式启动"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁总结，明确资源请求",
            "filling_prompt": "必须填入真实内容：key_points 提供3个核心要点，每条不超过60字。01为核心价值（项目能解决什么问题、带来什么收益）；02为关键数据（预算、ROI、时间节点等量化信息）；03为资源请求（需要什么审批、资源或支持）。禁止保留花括号。"
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
