TEMPLATE = {
    "name": "weekly-report",
    "name_cn": "周报/月报",
    "description": "适合团队周报、项目月报、工作汇报等场景。简洁高效，重点突出，数据驱动。",
    "target_audience": "团队负责人、项目经理、管理层",
    "typical_slides": 12,
    "typical_duration": "5-10分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "reporting_principles": [
            "简洁：只汇报关键信息，不罗列细节",
            "数据：用数据说话，避免空泛描述",
            "重点：突出本周关键进展和问题",
            "计划：明确下周工作安排"
        ],
        "reader_mindset": "领导关心的是：进度如何？有什么问题？需要什么支持？"
    },
    "structure_guide": {
        "overview": "本周整体情况，一句话概括",
        "completed": "本周完成的工作",
        "in_progress": "进行中的工作",
        "issues": "遇到的问题和风险",
        "next_week": "下周工作计划"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "营销一部 第21周工作周报",
            "subtitle": "2025年5月19日—5月23日",
            "author": "张明远",
            "date": "2025年5月23日",
            "notes": "标题页简洁，注明时间段和汇报人",
            "filling_prompt": "必须填入真实内容：title 中填入实际部门名称和周次，subtitle 为具体时间段，author 为汇报人姓名，date 为实际日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "本周工作概览",
            "kicker": "目录",
            "items": [
                "01  核心指标",
                "02  完成事项",
                "03  进行中事项",
                "04  关键进展",
                "05  阶段深度分析",
                "06  问题与风险",
                "07  下周计划"
            ],
            "notes": "让观众快速了解本报告结构，每项一行即可",
            "filling_prompt": "目录页为固定结构，可根据实际内容增删章节。"
        },
        {
            "index": 3,
            "type": "kpi_dashboard",
            "title": "本周核心指标",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 周期概览",
            "kpis": [
                {"value": "33家", "label": "新增客户数", "delta": "+18%", "baseline": "上周28家"},
                {"value": "128万元", "label": "合同签约额", "delta": "+23%", "baseline": "上周104万元"},
                {"value": "47次", "label": "拜访客户数", "delta": "+12%", "baseline": "上周42次"},
                {"value": "78%", "label": "项目交付率", "delta": "-5%", "baseline": "上周83%"}
            ],
            "notes": "4个核心指标，展示本周期整体情况",
            "filling_prompt": "必须填入真实内容：提供4个本周期的核心指标数据，每个有 value（数字）、label（说明）、delta（变化趋势）、baseline（对比基准）。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "新签合同：智慧社区项目",
            "content_type": "example_detail",
            "kicker": "完成事项 · 合同签署",
            "lede": "成功签署绿城物业智慧社区系统项目，合同金额45万元，预计6月中旬启动实施。",
            "context_block": "绿城物业自2025年3月开始数字化选型，先后考察了5家供应商，我方在4月中旬进入最终候选名单。",
            "solution_block": "销售团队经过6轮商务谈判，针对对方提出的价格折扣、服务响应速度、系统集成能力三个核心诉求，提供了定制化方案并争取到公司特批折扣，最终双方达成合作。",
            "metrics": [
                {"value": "45万元", "label": "合同金额", "trend": "毛利率38%"},
                {"value": "6轮", "label": "商务谈判", "trend": "历时6周"},
                {"value": "1家", "label": "新增战略客户", "trend": "后续有望复制"}
            ],
            "takeaway": "大型客户需要耐心跟进，核心在于精准理解对方需求并提供差异化价值。",
            "notes": "用具体数字说明完成情况",
            "filling_prompt": "必须填入真实内容：lede 一句话概括本周期的核心成果；context_block 描述工作背景（1-2句话）；solution_block 具体说明完成工作的过程和关键动作（2-3句话）；metrics_grid 提供3个量化指标，每个有 value（数字）、label（说明）、trend（对比上周期）；takeaway 用一句话总结成果意义。禁止虚构数据。"
        },
        {
            "index": 5,
            "type": "process_flow",
            "title": "团队效能提升",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "kicker": "流程优化",
            "steps": [
                {"num": "01", "title": "市场线索获取", "desc": "通过展会、网络推广、渠道推荐获取潜在客户信息"},
                {"num": "02", "title": "电话初筛", "desc": "在48小时内完成首次电话沟通，筛选意向客户"},
                {"num": "03", "title": "上门拜访", "desc": "对意向客户进行需求调研，现场演示产品方案"},
                {"num": "04", "title": "方案报价", "desc": "根据需求定制方案，3个工作日内出具详细报价"},
                {"num": "05", "title": "合同签署", "desc": "完成商务谈判，签署销售合同并收取定金"}
            ],
            "notes": "从线索到签单的标准流程，5步闭环管理",
            "filling_prompt": "必须填入真实内容：提供从线索获取到签单的完整流程步骤，每步有 title 和 desc（不超过30字）。结合团队实际工作流程填写。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "华东区域市场拓展",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "关键进展",
            "header": "华东三城市场调研成果",
            "sub_header": "新增意向客户8家，锁定重点目标市场",
            "paragraph": "本周带领华东拓展小组赴上海、苏州、杭州三地开展为期5天的市场调研。实地走访了12家目标客户，涵盖物业管理、教育信息化、政府数字化三个重点行业。调研发现华东市场对智慧社区解决方案需求旺盛，当地政府也出台了相关数字化转型补贴政策。通过此次调研，新增意向客户8家，其中3家已进入方案报价阶段，预计6月份可完成至少2家合同签署。",
            "references": [
                "上海市住建委《智慧社区建设指南（2025版）》",
                "苏州市政府数字化转型三年行动计划"
            ],
            "notes": "用图文混排展示关键进展，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少1个URL，如无外部数据需求可填入内部系统URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为进展标题（不超过35字）；sub_header 为进展概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述项目的关键进展、突破点和实际成果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "当前项目进度",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "kicker": "进行中事项",
            "cards": [
                {
                    "title": "智慧社区",
                    "status": "签约定金",
                    "progress": 90,
                    "desc": "已完成合同签署，收到定金10万元，6月中旬启动开发。"
                },
                {
                    "title": "企业OA系统",
                    "status": "方案评审",
                    "progress": 65,
                    "desc": "方案已提交客户，预计本周完成内部评审，下周进入合同谈判。"
                },
                {
                    "title": "政府党建平台",
                    "status": "交付验收",
                    "progress": 88,
                    "desc": "系统已上线试运行，正在配合甲方进行数据迁移和用户培训。"
                },
                {
                    "title": "教育信息化",
                    "status": "需求确认",
                    "progress": 35,
                    "desc": "完成第一轮需求调研，正在编写详细解决方案和预算方案。"
                }
            ],
            "notes": "4个重点项目的进度一目了然",
            "filling_prompt": "必须填入真实内容：提供4个当前进行中的重点项目，每张卡片有 title（项目名称）、status（当前阶段名称）、progress（0-100的完成度百分比）、desc（一句话描述当前进展和下一步）。"
        },
        {
            "index": 8,
            "type": "stat_slide",
            "title": "本周关键成果",
            "content_type": "stat_slide",
            "kicker": "数据亮点",
            "stats": [
                {
                    "value": "3份",
                    "label": "签约合同",
                    "sub_label": "合同金额128万元",
                    "highlight": true
                },
                {
                    "value": "128万元",
                    "label": "合同金额",
                    "sub_label": "同比增长23%",
                    "highlight": true
                },
                {
                    "value": "33家",
                    "label": "新增客户",
                    "sub_label": "其中A类客户8家",
                    "highlight": false
                },
                {
                    "value": "78%",
                    "label": "项目交付率",
                    "sub_label": "较上周下降5个百分点",
                    "highlight": false
                }
            ],
            "notes": "用数据突出本周核心成果",
            "filling_prompt": "必须填入真实内容：提供4个关键数据，每条有 value（数字）、label（大指标名称）、sub_label（一句话补充说明）、highlight（是否重点突出）。数据必须与 KPI 看板保持一致。"
        },
        {
            "index": 9,
            "type": "deep_dive",
            "title": "Q2季度目标完成情况",
            "content_type": "deep_dive",
            "kicker": "深度分析",
            "header": "Q2销售目标进度追踪",
            "overview": "Q2季度销售目标600万元，截至本周累计完成372万元，目标完成率62%，落后时间进度8个百分点。",
            "analysis_blocks": [
                {
                    "title": "目标差距",
                    "content": "Q2剩余5周需完成228万元，平均每周需签单45.6万元，显著高于当前每周32万元的签单速度。"
                },
                {
                    "title": "结构分析",
                    "content": "已完成372万中，政府类项目占58%，企业类占30%，教育类占12%。政府类项目回款周期长，导致现金流压力较大。"
                },
                {
                    "title": "原因剖析",
                    "content": "落后原因有三：一是1名资深销售4月底离职导致客户流失；二是教育行业竞争加剧，平均报价下浮15%；三是部分大单签约周期超出预期。"
                }
            ],
            "recommendation": "建议Q3启动校园招聘补充销售力量，同步加强企业类客户的拓展力度，以改善客户结构。",
            "notes": "深入分析季度目标差距，为后续决策提供依据",
            "filling_prompt": "必须填入真实内容：overview 一句话概括整体情况；analysis_blocks 提供3个分析维度，每个有 title 和 content；recommendation 给出下一步建议。数据必须真实准确。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "问题与风险",
            "content_type": "content_slide",
            "issues": [
                {
                    "issue": "政府党建项目验收延迟",
                    "severity": "高",
                    "description": "甲方内部审批流程延长，原定5月底的验收可能推迟至6月中旬，影响Q2回款计划。",
                    "solution": "已与甲方项目负责人沟通，争取分阶段验收先行回款50%，同时推进内部审批加急。",
                    "status": "跟进中"
                },
                {
                    "issue": "人才招聘进度落后",
                    "severity": "中",
                    "description": "Q2计划招聘2名销售工程师，至今未找到合适人选，市场上符合要求的人才供不应求。",
                    "solution": "调整招聘渠道，增加猎头服务，同时考虑内部培养有潜力的初级销售。",
                    "status": "执行中"
                },
                {
                    "issue": "教育信息化报价竞争激烈",
                    "severity": "中",
                    "description": "3家竞争对手近期大幅压低报价，我方报价高于市场均值20%，导致多个教育项目丢单。",
                    "solution": "研究差异化竞争策略，强调服务质量和长期价值；探索与集成商合作的模式。",
                    "status": "分析中"
                }
            ],
            "notes": "列出当前面临的问题和风险，以及应对措施",
            "filling_prompt": "必须填入真实内容：列出1-3个当前面临的问题/风险，每条说明 severity（高/中/低）、description、solution、status。坦诚汇报，不回避问题。"
        },
        {
            "index": 11,
            "type": "content_slide",
            "title": "第22周工作计划",
            "content_type": "content_slide",
            "next_week_items": [
                {"task": "完成政府党建平台交付验收", "priority": "高", "due": "5月30日"},
                {"task": "完成企业OA系统合同签署", "priority": "高", "due": "5月29日"},
                {"task": "开展杭州客户见面会（3家意向客户）", "priority": "中", "due": "5月28日"},
                {"task": "完成Q2季度销售总结报告", "priority": "中", "due": "5月31日"}
            ],
            "notes": "下周的工作计划和目标",
            "filling_prompt": "必须填入真实内容：列出下周工作计划（3-5条），每条具体可执行，每项标注 priority（高/中/低）和 due（截止时间）。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "本周总结",
            "key_points": [
                "01 新签合同3份，合同金额128万元，同比增长23%，整体目标完成率62%",
                "02 下周全力冲刺政府党建平台验收，力争Q2回款不低于450万元",
                "03 华东市场拓展成效显著，新增意向客户8家，下月重点跟进转化"
            ],
            "thank_you": "感谢聆听",
            "notes": "简洁结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（本周期核心成果2条+下周工作重点1条），每条内容具体充实。结尾致谢语可保持固定。"
        }
    ],
    "design_tips": [
        "周报要简洁，不要堆砌文字",
        "用数据说话，核心指标突出",
        "问题与风险要坦诚，不要回避",
        "计划要具体可执行",
        "每项行动标注优先级和截止时间",
        "善用颜色区分紧急程度",
        "图表比文字更直观",
        "保持格式一致，方便对比"
    ],
    "common_metrics": {
        "progress": ["完成率", "进度百分比", "里程碑达成"],
        "quality": ["Bug数", "测试通过率", "缺陷密度"],
        "team": ["人力投入", "人均产出", "协作情况"],
        "risk": ["风险数量", "风险等级", "应对措施"]
    },
    "anti_patterns": [
        "流水账式记录，缺乏重点",
        "报喜不报忧，隐藏问题",
        "计划太虚，缺乏具体行动",
        "数据不准确，前后矛盾"
    ]
}
