TEMPLATE = {
    "name": "activity-plan",
    "name_cn": "活动策划",
    "description": "适合团建活动、校园活动、节日策划等场景。活泼有创意，执行清晰，预算可控。",
    "target_audience": "领导、团队成员、活动参与者",
    "typical_slides": 14,
    "typical_duration": "10-15分钟（策划汇报）",
    "palette": "activity_orange",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "planning_principles": [
            "活动要有明确的主题和目标",
            "执行计划要具体，时间节点清晰",
            "预算要合理，有明细有依据",
            "风险预案要全面，有备无患"
        ],
        "engagement_tips": [
            "设置互动环节，提高参与度",
            "考虑不同人群的参与意愿",
            "留下足够自由交流的时间",
            "活动后收集反馈，持续改进"
        ]
    },
    "activity_types": {
        "team_building": "团建活动",
        "celebration": "节日庆祝",
        "training": "培训拓展",
        "social": "社交联谊",
        "volunteer": "志愿服务"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "活动名称",
            "subtitle": "活动口号/主题标语",
            "author": "策划团队名称",
            "date": "计划举办日期",
            "notes": "标题页有活力，吸引眼球",
            "filling_prompt": "必须填入真实内容：title 为活动名称，subtitle 为一句概括性口号，author 为策划团队，date 为计划举办日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  活动概述",
                "02  活动详情",
                "03  人员分工",
                "04  预算安排",
                "05  风险预案"
            ],
            "notes": "让观众了解策划框架",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "活动概述",
            "subtitle": "背景与目标",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "活动背景与目标",
            "content_type": "example_detail",
            "kicker": "实例 · 活动背景",
            "lede": "通过活动，增强团队凝聚力，激发员工活力",
            "context_block": "描述活动策划的背景、团队现状和预期效果（1-2句话）。",
            "solution_block": "具体说明活动核心创意和执行亮点（2-3句话）。",
            "metrics": [
                {"value": "人数", "label": "预计参与人数", "trend": "参与范围"},
                {"value": "时长", "label": "活动时长", "trend": "时间安排"},
                {"value": "预算", "label": "总预算", "trend": "人均预算"}
            ],
            "takeaway": "一句话总结活动意义。",
            "notes": "说明为什么举办这个活动",
            "filling_prompt": "必须填入真实内容：lede 一句话概括活动核心理念；context_block 描述策划背景和预期效果（1-2句话）；solution_block 具体说明活动核心创意和执行亮点（2-3句话）；metrics_grid 提供3个活动指标；takeaway 用一句话总结活动意义。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "活动详情",
            "subtitle": "时间地点与流程",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "timeline",
            "title": "活动流程",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "milestones": [
                {"time": "时间", "event": "环节名称", "duration": "持续时长", "location": "地点"},
                {"time": "时间", "event": "环节名称", "duration": "持续时长", "location": "地点"},
                {"time": "时间", "event": "环节名称", "duration": "持续时长", "location": "地点"}
            ],
            "notes": "展示活动当天的流程安排",
            "filling_prompt": "必须填入真实内容：提供活动当天的流程安排，列出4-6个关键时间节点，每个有时间+环节名称+持续时长+负责人。"
        },
        {
            "index": 7,
            "type": "image_text",
            "title": "活动亮点",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "亮点 · 核心创意",
            "header": "核心亮点名称",
            "sub_header": "亮点预期效果",
            "paragraph": "详细描述活动亮点的创意来源、实施方式和预期效果，用流畅的段落形式呈现，禁止罗列要点。",
            "notes": "右侧配往期活动照片/场地效果图/创意展示图，左侧列亮点",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最具特色的亮点名称（不超过35字）；sub_header 为该亮点的预期效果（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述活动亮点的创意来源、实施方式和预期效果，用流畅的段落形式呈现，禁止罗列要点。如果参考了外部案例，通过 web_search 获取相关资料并在 references 列出 URL。"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "活动项目安排",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "环节名称", "body": "环节描述和价值"},
                {"header": "环节名称", "body": "环节描述和价值"},
                {"header": "环节名称", "body": "环节描述和价值"},
                {"header": "环节名称", "body": "环节描述和价值"}
            ],
            "notes": "展示活动项目安排",
            "filling_prompt": "必须填入真实内容：提供4个活动环节，每个有 header（环节名称）和 body（环节描述和价值）。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "人员分工",
            "subtitle": "组织架构与职责",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "组织架构与分工",
            "content_type": "content_slide",
            "structure": [
                {"role": "角色名称", "name": "负责人姓名", "responsibility": "职责描述"},
                {"role": "角色名称", "name": "负责人姓名", "responsibility": "职责描述"}
            ],
            "notes": "展示团队分工和职责",
            "filling_prompt": "必须填入真实内容：说明活动组织架构，列出各小组及每个小组的职责和负责人。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "预算安排",
            "subtitle": "费用预算与资源",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 12,
            "type": "kpi_dashboard",
            "title": "预算概览",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "预算 · 费用明细",
            "kpis": [
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "说明"},
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "说明"},
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "说明"},
                {"value": "金额", "label": "预算类别", "delta": "占比", "baseline": "说明"}
            ],
            "notes": "展示预算分配",
            "filling_prompt": "必须填入真实内容：提供4个预算类别，每个有 value（金额）、label（类别名称）、delta（占比）、baseline（总预算）。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "风险预案",
            "subtitle": "预案与措施",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 14,
            "type": "example_detail",
            "title": "风险识别与预案",
            "content_type": "example_detail",
            "kicker": "实例 · 风险预案",
            "lede": "提前识别风险，有备无患，确保活动顺利进行",
            "context_block": "描述风险具体表现和发生场景（1-2句话）。",
            "solution_block": "针对每个风险提供预防措施和应急预案（2-3句话）。",
            "metrics": [
                {"value": "已购买", "label": "保险覆盖", "trend": "覆盖情况"},
                {"value": "数量", "label": "备用资源", "trend": "保障情况"},
                {"value": "人数", "label": "安全员配置", "trend": "配置情况"}
            ],
            "takeaway": "一句话总结风险管理核心原则。",
            "notes": "识别风险并准备预案",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最需要关注的风险点；context_block 描述风险具体表现和发生场景（1-2句话）；solution_block 针对每个风险提供预防措施和应急预案（2-3句话）；metrics_grid 提供3个风险相关指标；takeaway 用一句话总结风险管理核心原则。"
        },
        {
            "index": 15,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心主题：活动名称",
                "02 核心亮点：亮点描述",
                "03 预期效果：效果描述"
            ],
            "thank_you": "期待您的支持！",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心理念+核心亮点+预期效果）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "活动策划要有创意，亮点突出",
        "执行计划要具体，时间节点清晰",
        "预算要合理，有明细有依据",
        "风险预案要全面，有备无患",
        "PPT设计可以活泼一些，体现活动氛围",
        "多使用活动现场照片或效果图",
        "考虑不同参与者的需求和意愿",
        "提前测试活动项目，确保可行性"
    ],
    "checklist": {
        "before_event": [
            "确认场地预订",
            "确认交通安排",
            "确认餐饮预订",
            "购买保险",
            "收集参与者健康信息",
            "准备物资清单",
            "确认人员分工",
            "发布活动通知"
        ],
        "during_event": [
            "签到确认",
            "安全提醒",
            "流程把控",
            "应急处理",
            "拍照记录",
            "现场协调"
        ],
        "after_event": [
            "安全返程",
            "活动总结",
            "费用结算",
            "反馈收集",
            "资料归档"
        ]
    }
}
