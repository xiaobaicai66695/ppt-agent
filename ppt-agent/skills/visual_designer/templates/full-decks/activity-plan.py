TEMPLATE = {
    "name": "activity-plan",
    "name_cn": "活动策划",
    "description": "适合团建活动、校园活动、节日策划等场景。活泼有创意，执行清晰，预算可控。",
    "target_audience": "领导、团队成员、活动参与者",
    "typical_slides": 15,
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
            "title": "「凝心聚力」年度团队拓展活动策划方案",
            "subtitle": "挑战自我 · 协同共进 · 共创辉煌",
            "author": "人力资源部 — 团建策划组",
            "date": "2026年6月15日（周日）",
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
            "lede": "打破部门壁垒，促进跨部门协作，建立更深厚的团队信任",
            "context_block": "随着公司业务快速扩张，各部门之间的沟通协作成本显著上升。据2026年Q1内部调研，约67%的员工表示希望有更多跨部门交流机会，团队归属感评分较去年同期下降了5个百分点。",
            "solution_block": "本次「凝心聚力」团建活动以户外拓展为载体，设计了多组需要跨部门协作才能完成的挑战项目。不同于传统团建，本次活动采用「积分对抗」机制，各部门混合组队，确保每个人都有机会认识新同事，在挑战中建立信任与默契。",
            "metrics": [
                {"value": "120人", "label": "预计参与人数", "trend": "覆盖各部门"},
                {"value": "1天", "label": "活动时长", "trend": "09:00-18:00"},
                {"value": "300元/人", "label": "人均预算", "trend": "总预算3.6万"}
            ],
            "takeaway": "通过沉浸式团队挑战，打破沟通壁垒，让每位成员感受到自己是团队不可或缺的一部分。",
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
                {"time": "08:30", "event": "集合签到", "duration": "30分钟", "location": "公司大楼前广场"},
                {"time": "09:00", "event": "开幕仪式", "duration": "30分钟", "location": "基地主广场"},
                {"time": "09:30", "event": "破冰分组", "duration": "30分钟", "location": "基地主广场"},
                {"time": "10:00", "event": "团队挑战赛（上）", "duration": "2.5小时", "location": "拓展场地A/B/C"},
                {"time": "12:30", "event": "午餐休息", "duration": "90分钟", "location": "基地餐厅"},
                {"time": "14:00", "event": "团队挑战赛（下）", "duration": "2.5小时", "location": "拓展场地D/E"},
                {"time": "16:30", "event": "颁奖与闭幕", "duration": "60分钟", "location": "基地主广场"},
                {"time": "17:30", "event": "返程", "duration": "30分钟", "location": "基地大门"}
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
            "header": "「城市猎人」真人CS对抗赛",
            "sub_header": "跨部门组队 · 策略协作 · 激发竞争活力",
            "paragraph": "「城市猎人」是本次团建的核心亮点项目。我们将还原真实的城市巷战场景，每支队伍由来自不同部门的6-8名成员组成，需在限定时间内完成情报收集、据点攻防、物资运送等多重任务。与传统CS不同，本项目引入了「技能卡」机制——每位成员拥有独特的技能（如医疗、狙击、突击等），只有团队成员互相配合、合理调度技能卡，才能在对抗中取得优势。这一设计确保了即使体能不占优势的成员也能在团队中发挥关键作用，真正实现「人人都是主角」的团建目标。",
            "notes": "右侧配往期活动照片/场地效果图/创意展示图，左侧列亮点",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最具特色的亮点名称（不超过35字）；sub_header 为该亮点的预期效果（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述活动亮点的创意来源、实施方式和预期效果，用流畅的段落形式呈现，禁止罗列要点。如果参考了外部案例，通过 web_search 获取相关资料并在 references 列出 URL。"
        },
        {
            "index": 8,
            "type": "card_grid",
            "title": "活动项目",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "破冰分组 — 「疯狂卡片」", "body": "通过趣味心理测试和肢体互动游戏，将120人随机混编为12支跨部门小队，每队10人，在轻松氛围中快速破冰。"},
                {"header": "「城市猎人」真人CS", "body": "在专业场地进行3轮对抗赛，每轮30分钟。融入技能卡、资源管理等策略元素，考验团队战术配合与执行力。"},
                {"header": "「信任背摔」高空挑战", "body": "每位成员依次从1.5米高台向后倒下，由团队其他成员用手臂接住。克服心理恐惧，建立对队友的绝对信任。"},
                {"header": "「共绘蓝图」团队创意", "body": "各队使用提供的材料，在60分钟内共同完成一幅巨型拼画。最终12幅画作将拼接成公司愿景图，作为企业文化资产永久留存。"}
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
            "type": "two_column",
            "title": "组织架构与分工",
            "content_type": "two_column",
            "layout_hint": "left-responsibilities right-team",
            "left_header": "职责分工",
            "left_items": [
                {"role": "总指挥", "name": "王海峰（HR总监）", "responsibility": "统筹协调全程，重大事项决策，资源调配"},
                {"role": "副总指挥", "name": "李晓燕（HR主管）", "responsibility": "现场执行调度，各组工作协调，进度把控"},
                {"role": "后勤保障组", "name": "张明（行政部）", "responsibility": "餐饮、交通、保险、医疗应急物资准备"},
                {"role": "安全督导组", "name": "赵强（安全专员）", "responsibility": "全程安全监督，高风险项目现场保护，处理突发情况"},
                {"role": "宣传记录组", "name": "陈思（品牌部）", "responsibility": "活动摄影、录像，制作活动回顾视频和推文"}
            ],
            "right_header": "执行团队",
            "right_items": [
                "主教练：孙磊（国家认证拓展培训师，8年经验）",
                "助理教练：4人（负责各场地项目引导与安全保障）",
                "计时计分组：2人（负责各环节成绩记录与积分统计）",
                "医疗保障：1人（持证急救员，随队配备急救箱）",
                "车辆调度：1人（负责大巴点对点接送）"
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
                {"value": "¥36,000", "label": "总预算", "delta": "300元/人", "baseline": "120人 × 300元"},
                {"value": "¥14,400", "label": "场地与项目", "delta": "40%", "baseline": "拓展基地租赁 + CS场地费"},
                {"value": "¥10,800", "label": "餐饮与交通", "delta": "30%", "baseline": "自助午餐 + 大巴往返接送"},
                {"value": "¥7,200", "label": "物资与保障", "delta": "20%", "baseline": "服装/横幅/保险/医疗/奖品"},
                {"value": "¥3,600", "label": "预留应急", "delta": "10%", "baseline": "应对突发情况与临时需求"}
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
            "type": "case_study",
            "title": "风险识别与预案",
            "content_type": "case_study",
            "kicker": "案例 · 风险预案",
            "case_title": "户外拓展活动常见风险应对",
            "context": "户外拓展活动面临多种潜在风险，包括天气突变、人员意外伤害、设备故障、人员突发疾病等。根据行业数据，户外团建活动的事故中，约45%与天气相关，30%与运动损伤相关，15%与突发疾病相关。",
            "challenges": [
                {"title": "恶劣天气", "detail": "雷雨、大风等突发天气可能导致户外项目无法正常进行，强行进行存在安全隐患。"},
                {"title": "运动损伤", "detail": "真人CS和拓展项目中可能发生擦伤、扭伤等轻微运动损伤，需现场快速处理。"},
                {"title": "人员走散", "detail": "120人的大型活动，在开放式场地中可能出现人员走散或掉队的情况。"},
                {"title": "突发疾病", "detail": "高温环境下长时间活动，可能有成员出现中暑、脱水等身体不适。"}
            ],
            "solutions": [
                {"title": "天气预案", "solution": "提前3天关注天气预报，如预报有雨，提前通知改期至备用日期（6月22日）。如活动当天突发雷雨，立即启动室内备选方案——「密室逃脱」主题团队挑战，所有物料已提前对接备选供应商。"},
                {"title": "医疗保障", "solution": "场地配备持证急救员一名，急救箱包含创可贴、碘伏、绷带、冰袋、防暑药品等。如发生较严重损伤，现场判断后送至距基地8公里的区人民医院（车程约15分钟）。"},
                {"title": "人员管理", "solution": "每支队伍配备一名带队教练，实时清点人数。活动前要求各部门上报参与者健康信息，建立微信群实时共享位置（自愿原则），设置集合点标识牌。"},
                {"title": "高温防护", "solution": "避开12:00-14:00高温时段进行户外项目，在阴凉处设置补水站，每30分钟强制补水休息。提供遮阳帽和冰袖，防止晒伤和中暑。"}
            ],
            "notes": "识别风险并准备预案",
            "filling_prompt": "必须填入真实内容：context 描述风险整体背景；challenges 列出4个具体风险及其描述；solutions 针对每个风险提供预防措施和应急预案。禁止虚构场景。"
        },
        {
            "index": 15,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心理念：打破部门壁垒，通过沉浸式挑战让跨部门成员建立深度信任",
                "02 核心亮点：「城市猎人」真人CS + 「共绘蓝图」创意活动，实现趣味与意义的双重目标",
                "03 预期效果：活动后团队归属感提升15%，跨部门协作效率改善，沉淀可复用的团建模式"
            ],
            "thank_you": "感谢各位领导的支持，期待与大家共同创造难忘的团队记忆！",
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
