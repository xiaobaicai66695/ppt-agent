TEMPLATE = {
    "name": "meeting-minutes",
    "name_cn": "会议纪要",
    "description": "适合会议记录、工作例会、项目评审会等场景。结构清晰，要点明确，行动明确。",
    "target_audience": "参会人员、项目经理、领导",
    "typical_slides": 16,
    "typical_duration": "5-10分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "documentation_standards": "纪要的核心价值是：备忘、传递、追踪",
        "key_principles": [
            "客观记录：如实记录，不添加个人解读",
            "要点清晰：突出关键决策和行动项",
            "责任明确：每项行动有负责人和截止时间",
            "及时分发：会议结束后24小时内发出"
        ],
        "quality_check": [
            "是否记录了所有关键决策？",
            "是否明确了每项行动的负责人？",
            "是否有明确的截止时间？",
            "是否分发了所有相关人员？"
        ]
    },
    "minutes_template": {
        "basic_info": "会议基本信息（时间、地点、参会人等）",
        "agenda": "会议议程",
        "discussion": "讨论内容摘要",
        "decisions": "会议决议",
        "action_items": "行动项清单",
        "next_meeting": "下次会议安排"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "Q3产品规划评审会",
            "subtitle": "2025年7月15日 14:00-16:00 · 总部大厦A座3层会议室",
            "author": "张文静",
            "date": "2025年7月15日",
            "notes": "标题页简洁，注明会议基本信息（主题、时间地点、记录人）",
            "filling_prompt": "必须填入真实内容：title 为会议名称，subtitle 为会议时间和地点，author 为记录人姓名，date 为纪要编写日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  会议背景",
                "02  讨论要点",
                "03  决议事项",
                "04  行动项"
            ],
            "notes": "清晰展示纪要结构，让参会者快速了解会议脉络",
            "filling_prompt": "目录页为固定结构，无需额外填充。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "会议背景",
            "subtitle": "会议目的与议程",
            "notes": "章节分隔页，仪式感强，用于划分不同内容板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "会议背景与目的",
            "content_type": "example_detail",
            "kicker": "实例 · 会议背景",
            "lede": "Q3产品规划需在7月底前完成终版确认，当前存在功能优先级分歧和技术资源约束。",
            "context_block": "本次会议由产品委员会发起，针对Q3产品路线图进行集中评审。产品团队于7月8日提交了初版规划，但技术团队对部分功能的实现周期存在异议，需在本次会议中达成共识。",
            "solution_block": "会议目标一是明确Q3核心功能清单，目标二是确定各功能的交付优先级，目标三是确认技术团队的资源投入计划。与会各方需在会议结束时签署终版规划确认书，确保后续执行有据可依。",
            "metrics": [
                {"value": "9人", "label": "参会人数", "trend": "含3位部门负责人"},
                {"value": "6项", "label": "议程议题", "trend": "覆盖产品/技术/运营"},
                {"value": "4项", "label": "预期决议", "trend": "需形成书面记录"}
            ],
            "takeaway": "本次会议将为Q3产品开发定调，决议结果直接影响下季度资源分配和技术排期。",
            "notes": "说明召开会议的背景和目的，让读者理解为何开这个会",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心议题和决策焦点；context_block 描述会议背景原因和组织动机（1-2句话）；solution_block 具体说明要解决的核心问题和预期产出（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结会议对项目/工作的影响。"
        },
        {
            "index": 5,
            "type": "kpi_dashboard",
            "title": "会议数据概览",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 会议统计",
            "kpis": [
                {"value": "9人", "label": "实到人数", "delta": "1人缺席", "baseline": "计划10人"},
                {"value": "2小时", "label": "会议时长", "delta": "按计划进行", "baseline": "计划2小时"},
                {"value": "6项", "label": "讨论议题", "delta": "全部覆盖", "baseline": "计划6项"},
                {"value": "4项", "label": "形成决议", "delta": "1项待跟进", "baseline": "目标4项"}
            ],
            "notes": "用数据快速呈现会议整体情况",
            "filling_prompt": "必须填入真实内容：提供4个会议相关指标数据，每个有 value（数字）、label（说明）、delta（变化趋势）、baseline（对比基准）。如实反映会议实际进行情况。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "讨论要点",
            "subtitle": "会议讨论内容摘要",
            "notes": "章节分隔页，用于划分讨论内容板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "two_column",
            "title": "核心议题讨论",
            "content_type": "two_column",
            "kicker": "议题对比 · 讨论摘要",
            "left_header": "功能优先级争议",
            "left_sections": {
                "analysis": [
                    "智能推荐模块 vs 基础搜索优化：产品认为推荐是核心差异化竞争力，技术认为搜索优化投入产出比更高",
                    "移动端优先策略：移动端用户占比已达78%，但部分核心功能PC端体验仍优于移动端",
                    "A/B测试资源冲突：运营团队提出的新实验与Q3主功能开发争抢同一批测试资源"
                ],
                "data": [
                    "推荐模块预估工时：8人周",
                    "搜索优化预估工时：3人周",
                    "当前A/B实验队列：12个"
                ]
            },
            "right_header": "达成共识",
            "right_sections": {
                "key_points": [
                    "优先级确认：Q3以搜索优化为最高优先级，推荐模块延后至Q4第一周启动",
                    "移动端策略：建立移动端体验基线，任何新功能必须通过基线测试方可上线",
                    "资源协调：测试资源按功能优先级分配，运营实验队列超过8个时需产品VP审批"
                ],
                "data": [
                    "搜索优化目标：7月底完成",
                    "移动端基线：下周一出初版标准",
                    "测试上限：8个并行实验"
                ]
            },
            "notes": "左右对比展示议题讨论的对立观点与最终共识，每个维度用数据支撑",
            "filling_prompt": "必须填入真实内容：left_header 为争议焦点，left_sections.analysis 列出2-3个具体讨论问题并说明各方观点，left_sections.data 列出对应的量化数据；right_header 为达成的共识，right_sections.key_points 列出2-3条对应的决议内容，right_sections.data 列出对应的行动数据。注意左右数据形成对比关系。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "决议事项",
            "subtitle": "会议决定",
            "notes": "章节分隔页，用于划分决议板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "content_slide",
            "title": "会议决议清单",
            "content_type": "content_slide",
            "kicker": "决议 · 正式决定",
            "decisions": [
                {
                    "num": "决议01",
                    "decision": "Q3最高优先级功能确认为「搜索体验优化」，目标7月31日前完成v1.0上线",
                    "rationale": "搜索是当前用户使用频率最高的功能，优化后预计可提升15%转化率",
                    "stakeholders": "产品部、技术部"
                },
                {
                    "num": "决议02",
                    "decision": "智能推荐模块Q3暂缓，Q4第一周启动，预计11月15日前完成灰度测试",
                    "rationale": "技术评估推荐模块需8人周，当前资源无法支撑，需等Q3搜索项目交付后释放人力",
                    "stakeholders": "产品部、技术部"
                },
                {
                    "num": "决议03",
                    "decision": "建立移动端体验标准，移动端基线测试通过率需达到90%以上方可上线",
                    "rationale": "移动端用户占比78%，体验质量直接影响留存和转化",
                    "stakeholders": "产品部、测试部"
                },
                {
                    "num": "决议04",
                    "decision": "测试资源池并行实验上限设为8个，超出需产品VP审批方可新增",
                    "rationale": "当前12个并行实验造成资源争抢，影响核心功能测试质量",
                    "stakeholders": "运营部、测试部"
                }
            ],
            "notes": "列出会议形成的所有正式决议，每条包含决议内容、决策依据和负责部门",
            "filling_prompt": "必须填入真实内容：列出会议形成的4项正式决议，每条包含 num（编号）、decision（决议内容）、rationale（决策依据）、stakeholders（涉及部门）。内容要具体明确，避免模糊表述。"
        },
        {
            "index": 10,
            "type": "quote_slide",
            "title": "关键引言",
            "content_type": "quote_slide",
            "kicker": "金句",
            "quote": "我们不能什么都做，必须有所取舍。与其做十个平庸的功能，不如把三个核心功能做到极致。",
            "attribution": "李明辉 · 产品VP",
            "context": "在讨论功能优先级时，产品VP李明辉强调聚焦的重要性",
            "notes": "摘录会议中具有指导意义的关键引言，增强纪要可读性",
            "filling_prompt": "必须填入真实内容：quote 为会议中某位重要发言人的关键引言（1-2句话），attribution 为发言人姓名和职位，context 为引言的背景说明。如会议中无特别有价值的引言，可跳过此页或如实填写'无'。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "行动项",
            "subtitle": "待办事项与负责人",
            "notes": "章节分隔页，用于划分行动项板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 12,
            "type": "card_grid",
            "title": "行动项清单",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {
                    "header": "输出Q3产品规划终版文档",
                    "body": "负责人：王思远 | 截止：2025年7月18日 | 根据会议决议，更新Q3产品路线图，确认功能优先级和时间节点，输出终版规划文档，经产品VP审批后同步至全公司。"
                },
                {
                    "header": "制定移动端体验基线标准",
                    "body": "负责人：陈晓琳 | 截止：2025年7月22日 | 测试部牵头，联合产品部制定移动端体验基线标准，包含性能、交互、兼容性三大类指标，确保新功能上线前必须通过基线测试。"
                },
                {
                    "header": "协调测试资源分配方案",
                    "body": "负责人：赵海峰 | 截止：2025年7月17日 | 测试部制定详细的测试资源分配方案，明确Q3主功能的测试优先级和时间表，并建立实验队列超限的预警和审批机制。"
                },
                {
                    "header": "通知推荐模块延后决定",
                    "body": "负责人：张文静 | 截止：2025年7月16日 | 纪要分发后24小时内，通过邮件和即时通讯告知相关部门（市场部、客服部、客户成功团队）推荐模块Q3暂缓的决定，并说明Q4计划。"
                }
            ],
            "notes": "明确每项行动的责任人和截止时间，确保决议可执行可追踪",
            "filling_prompt": "必须填入真实内容：提供4个行动项，每个有 header（任务名称，不超过35字）和 body（详细描述责任人、截止时间、具体要求和预期成果，100-120字）。禁止保留花括号占位符。"
        },
        {
            "index": 13,
            "type": "process_flow",
            "title": "决议执行计划",
            "content_type": "process_flow",
            "kicker": "流程 · 执行步骤",
            "direction": "horizontal",
            "steps": [
                {"num": "1", "title": "纪要分发", "desc": "会议结束后24小时内发出"},
                {"num": "2", "title": "部门传达", "desc": "各负责人在48小时内完成部门内传达"},
                {"num": "3", "title": "文档输出", "desc": "终版规划文档7月18日前完成"},
                {"num": "4", "title": "VP审批", "desc": "规划文档经产品VP审批确认"},
                {"num": "5", "title": "全公司同步", "desc": "审批通过后向全公司发布"}
            ],
            "notes": "用流程图展示决议从形成到落地的执行步骤",
            "filling_prompt": "必须填入真实内容：提供5个执行步骤，每步有名称和一句话描述，展示决议落地的完整流程。"
        },
        {
            "index": 14,
            "type": "content_slide",
            "title": "下次会议安排",
            "content_type": "content_slide",
            "kicker": "后续 · 会议预告",
            "next_meeting": {
                "title": "Q3规划执行跟进会",
                "date": "2025年7月29日（周二）",
                "time": "14:00-15:30",
                "location": "总部大厦A座3层会议室",
                "host": "王思远",
                "attendees": ["产品委员会全体成员", "技术部负责人", "测试部负责人"],
                "agenda": [
                    "Q3规划执行进度汇报（每人5分钟）",
                    "移动端基线标准制定进展",
                    "测试资源分配方案评审",
                    "Q4推荐模块启动筹备"
                ]
            },
            "notes": "明确下次会议的时间、地点、主持人和主要议程",
            "filling_prompt": "必须填入真实内容：next_meeting 填写下次会议的完整信息，包括 title、date、time、location、host、attendees、agenda。如暂无下次会议安排，可填写'待定'，并说明需要跟进的事项。"
        },
        {
            "index": 15,
            "type": "stat_slide",
            "title": "会议要点速览",
            "content_type": "stat_slide",
            "kicker": "要点 · 一页速览",
            "stats": [
                {"value": "4项", "label": "形成决议", "desc": "功能优先级、推荐模块延后、移动端标准、测试资源管理"},
                {"value": "4项", "label": "行动项", "desc": "规划文档、体验基线、资源方案、部门通知"},
                {"value": "3个部门", "label": "涉及范围", "desc": "产品部、技术部、测试部、运营部"},
                {"value": "7月18日", "label": "关键节点", "desc": "Q3终版规划文档提交截止"}
            ],
            "notes": "用一页数字速览会议核心成果，方便快速回顾",
            "filling_prompt": "必须填入真实内容：提供4个核心统计数字，每个有 value、label、desc，简洁概括会议要点。统计数字要与会议实际内容一致。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心决议：Q3搜索优化最高优先，推荐模块Q4启动",
                "02 关键行动：7月18日前输出终版规划，7月22日前制定移动端基线",
                "03 下次会议：7月29日Q3规划执行跟进会"
            ],
            "thank_you": "感谢参与",
            "notes": "简洁总结会议核心决议和行动项，预告下次会议",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心决议+关键行动项+下次会议安排）。禁止保留花括号占位符。"
        }
    ],
    "design_tips": [
        "会议纪要要客观准确，不添加个人解读",
        "决议和行动项要明确，避免模糊表述",
        "每项行动要有明确的负责人和截止时间",
        "结构清晰，便于后续追踪",
        "及时分发，确保相关人员知晓",
        "使用统一模板，保持格式一致",
        "重点突出，关键信息一目了然",
        "存档规范，便于查询历史"
    ],
    "distribution_checklist": [
        "所有参会人员",
        "项目干系人",
        "相关领导",
        "PMO（如有）"
    ],
    "archival_guidelines": {
        "naming_convention": "会议纪要_项目名称_日期",
        "storage_location": "项目文档库/会议纪要文件夹",
        "retention_period": "项目结束+1年"
    }
}
