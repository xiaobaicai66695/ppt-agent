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
            "title": "\"凝心聚力 逐梦前行\"春季团建活动",
            "subtitle": "挑战自我，熔炼团队，共创辉煌",
            "author": "人力资源部",
            "date": "2025年4月15日",
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
            "lede": "通过户外拓展活动，增强团队凝聚力，激发员工活力",
            "context_block": "公司成立5年来，团队规模从20人发展到150人，但跨部门协作仍有提升空间。通过调研发现，65%的员工表示希望有更多团建活动来增进同事间的了解。同时，Q2季度项目压力大，员工普遍反映工作压力大、缺乏放松机会。",
            "solution_block": "本次活动策划主题为'凝心聚力 逐梦前行'，目标包括：1）增进不同部门员工间的了解，打破沟通壁垒；2）释放工作压力，提升员工幸福感；3）培养团队协作精神，增强团队凝聚力；4）留下美好回忆，增强员工归属感。",
            "metrics": [
                {"value": "150人", "label": "预计参与人数", "trend": "全员参与"},
                {"value": "1天", "label": "活动时长", "trend": "周六整天"},
                {"value": "8万", "label": "总预算", "trend": "人均约530元"}
            ],
            "takeaway": "启示：一次好的团建活动，能够激发团队活力，增强组织粘性",
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
                {"time": "08:00-08:30", "event": "集合出发", "duration": "30分钟", "location": "公司楼下"},
                {"time": "08:30-09:30", "event": "车程+热身", "duration": "1小时", "location": "大巴车上"},
                {"time": "09:30-10:00", "event": "开场仪式", "duration": "30分钟", "location": "营地"},
                {"time": "10:00-12:00", "event": "团队挑战赛", "duration": "2小时", "location": "营地"},
                {"time": "12:00-13:30", "event": "午餐+休息", "duration": "1.5小时", "location": "餐厅"},
                {"time": "13:30-16:30", "event": "自由活动+颁奖", "duration": "3小时", "location": "营地"},
                {"time": "16:30-17:30", "event": "返程", "duration": "1小时", "location": "大巴车"}
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
            "header": "团队挑战赛：密室逃脱+定向越野",
            "sub_header": "智慧与体力的双重挑战",
            "paragraph": "本次活动最独特的亮点是团队挑战赛环节——'拯救总裁'密室逃脱+定向越野。参与者在规定时间内，通过解谜、寻找线索、定向任务等方式，完成挑战并'拯救'被困的'总裁'。这个设计既考验团队的智慧，也考验协作能力，同时增加了趣味性和话题性。往期类似活动反馈显示，这种沉浸式体验深受员工欢迎，参与满意度达95%以上。",
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
                {"header": "破冰环节", "body": "分组+团建小游戏，快速打破尴尬氛围"},
                {"header": "团队挑战赛", "body": "密室逃脱+定向越野，考验团队协作"},
                {"header": "野炊/BBQ", "body": "自己动手做午餐，体验团队合作的乐趣"},
                {"header": "自由交流", "body": "下午茶时间，自由交流，增进感情"}
            ],
            "notes": "展示活动项目安排"
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
                {"role": "总指挥", "name": "张三（HR总监）", "responsibility": "整体统筹，协调资源"},
                {"role": "策划组", "name": "李四（HR经理）+2人", "responsibility": "活动方案设计，流程策划"},
                {"role": "执行组", "name": "王五（行政主管）+4人", "responsibility": "现场执行，物资管理"},
                {"role": "后勤组", "name": "赵六（行政专员）+3人", "responsibility": "车辆、餐饮、住宿安排"},
                {"role": "安全组", "name": "保安队长+2人", "responsibility": "安全保障，紧急预案"}
            ],
            "notes": "展示团队分工和职责",
            "filling_prompt": "必须填入真实内容：说明活动组织架构，列出各小组（如策划组、执行组、后勤组、宣传组等）及每个小组的职责和负责人。"
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
                {"value": "3.2万", "label": "场地及活动费", "delta": "占比40%", "baseline": "含教练+物资"},
                {"value": "2.4万", "label": "餐饮费", "delta": "占比30%", "baseline": "含午餐+下午茶"},
                {"value": "1.6万", "label": "交通费", "delta": "占比20%", "baseline": "含大巴+保险"},
                {"value": "8000", "label": "其他费用", "delta": "占比10%", "baseline": "含奖品+备用金"}
            ],
            "notes": "展示预算分配",
            "filling_prompt": "必须填入真实内容：提供4个预算类别（如场地费用、餐饮费用、物料费用、人员费用等），每个有 value（金额）、label（类别名称）、delta（占比）、baseline（总预算）。"
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
            "context_block": "户外活动存在多种潜在风险，包括天气变化、人员安全、突发状况等。需要提前做好预案，确保活动安全顺利进行。",
            "solution_block": "主要风险及应对措施包括：1）天气风险：关注天气预报，提前准备雨具；如遇恶劣天气，延期举行。2）人员安全：购买意外保险，准备急救药箱；活动区域设置安全员。3）健康风险：提前收集参与者的健康信息；避免高强度运动。4）紧急预案：提前联系最近的医院；安排备用车辆。",
            "metrics": [
                {"value": "已购买", "label": "意外保险", "trend": "全员覆盖"},
                {"value": "2家", "label": "备用医院", "trend": "15分钟车程"},
                {"value": "5人", "label": "安全员配置", "trend": "活动全程"}
            ],
            "takeaway": "启示：安全是活动的底线，风险预案是安全的保障",
            "notes": "识别风险并准备预案",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最需要关注的风险点；context_block 描述风险具体表现和发生场景（1-2句话）；solution_block 针对每个风险提供预防措施和应急预案（2-3句话）；metrics_grid 提供3个风险相关指标；takeaway 用一句话总结风险管理核心原则。"
        },
        {
            "index": 15,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 主题：凝心聚力 逐梦前行",
                "02 亮点：沉浸式团队挑战赛",
                "03 预期效果：增进了解，提升凝聚力"
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
