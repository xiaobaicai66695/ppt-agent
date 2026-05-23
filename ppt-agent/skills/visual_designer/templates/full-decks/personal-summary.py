TEMPLATE = {
    "name": "personal-summary",
    "name_cn": "述职报告",
    "description": "适合个人总结、述职报告、年终总结等场景。重点突出，成果可见，计划明确。",
    "target_audience": "领导、同事、评审",
    "typical_slides": 16,
    "typical_duration": "10-15分钟",
    "palette": "report_green",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "summary_strategy": "述职的核心是展示：你做了什么、做得怎么样，下一步怎么做",
        "key_principles": [
            "成果要用数据说话",
            "问题要坦诚面对",
            "计划要具体可执行",
            "态度要谦逊但不卑微"
        ],
        "evaluation_criteria": {
            "performance": "业绩达成情况",
            "capability": "能力提升情况",
            "collaboration": "团队协作情况",
            "growth": "成长与反思"
        }
    },
    "self_evaluation_tips": {
        "strengths": "突出自己的核心优势和亮点",
        "improvements": "诚实地分析不足，但不要自我贬低",
        "learnings": "强调从经历中学到了什么"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "2024年度述职报告",
            "subtitle": "张伟 | 产品研发部 | 高级产品经理",
            "author": "张伟",
            "date": "2025年1月",
            "notes": "标题页简洁正式，信息完整",
            "filling_prompt": "必须填入真实内容：title 为时间段（如'2024年度'）+ 述职报告，subtitle 为姓名、部门、岗位信息，author 为姓名，date 为述职日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "CONTENTS",
            "items": [
                "01  工作概述",
                "02  主要成果",
                "03  能力成长",
                "04  问题反思",
                "05  未来规划"
            ],
            "notes": "5个章节让领导快速了解汇报结构",
            "filling_prompt": "目录页为固定结构，可根据实际情况调整章节编号和名称。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "工作概述",
            "subtitle": "岗位职责与工作回顾",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "岗位职责与年度目标",
            "content_type": "example_detail",
            "kicker": "实例 · 角色定位",
            "lede": "负责智能家居产品线规划与落地，带领5人产品团队完成3条产品线从0到1建设",
            "context_block": "高级产品经理，主要负责智能家居产品线的规划与生命周期管理。团队规模5人，对接研发、设计、运营、销售四个部门，直接汇报给产品总监。年初承接公司战略目标，主导智能家居生态平台的产品规划。",
            "solution_block": "本年度重点推进了三件事：一是完成智能门锁Pro产品线从立项到量产的全流程管理，产品上市首月销量突破2万台；二是主导智能家居中控屏2.0版本升级，用户 NPS 提升15个点；三是搭建产品数据监控体系，实现核心指标的实时追踪与预警机制。",
            "metrics": [
                {"value": "3条", "label": "产品线管理", "trend": "新增2条产品线"},
                {"value": "2.3亿", "label": "GMV贡献", "trend": "同比+47%"},
                {"value": "5人", "label": "团队管理", "trend": "培养2名骨干"}
            ],
            "takeaway": "从执行者向规划者的转型初见成效，为公司智能家居战略落地提供了产品支撑。",
            "notes": "左侧职责描述，右侧量化指标卡片，数据要真实",
            "filling_prompt": "必须填入真实内容：lede 一句话概括本周期工作整体情况；context_block 描述述职人所在岗位职责和主要工作范围（1-2句话）；solution_block 具体说明本周期内完成的主要工作和取得的成绩（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结本周期工作的价值。禁止虚构数据。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "主要成果",
            "subtitle": "核心业绩与贡献",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "核心项目业绩",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "核心项目",
            "header": "智能门锁Pro：从立项到品类TOP3",
            "sub_header": "9个月完成产品从0到月销2万台的全流程管理",
            "paragraph": "智能门锁Pro是我今年主导的核心项目。从市场调研、需求分析开始，我带领团队历经9个月完成了产品定义、工业设计对接、供应链协调、定价策略制定、上市推广全流程。项目最大的挑战是平衡安全性、便捷性与成本三者关系。通过引入双核MCU架构和银行级安全芯片，产品通过了BCTC金融级安全认证；同时优化结构设计，将成本控制在目标范围内，上市定价1999元，首发当日销售额突破3800万元。项目上线后持续稳居天猫智能门锁热销榜TOP3，累计销量突破15万台，为公司创造了2.1亿元GMV。",
            "notes": "右侧配产品实物图、发布会图或用户使用场景图",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为项目标题（不超过35字）；sub_header 为项目概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述项目的实施过程、克服的困难和取得的成果，用流畅的段落形式呈现，禁止罗列要点。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "四项关键成果",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "kicker": "工作成果",
            "cards": [
                {
                    "header": "智能中控屏2.0升级",
                    "body": "主导UI/UX重构和交互逻辑优化，引入场景联动引擎，用户单次使用时长达28分钟，较旧版提升67%，NPS净推荐值从32提升至47。"
                },
                {
                    "header": "产品数据体系搭建",
                    "body": "从0到1搭建产品数据监控平台，打通6个数据源，覆盖DAU、留存、转化率等18个核心指标，实现数据日报自动化，运营效率提升3倍。"
                },
                {
                    "header": "供应链降本攻关",
                    "body": "联合采购和研发团队优化BOM成本，通过规格重设计和供应商谈判，实现单台成本下降12%，年度节省成本约580万元。"
                },
                {
                    "header": "新品类市场切入",
                    "body": "完成智能窗帘电机品类规划，上市首季完成销售预期120%，成功验证了「高端智能窗帘」细分市场的可行性。"
                }
            ],
            "notes": "2x2卡片展示四个核心成果，每个卡片有标题和描述",
            "filling_prompt": "必须填入真实内容：提供4个其他工作成果，每个有 header（成果名称，不超过20字）和 body（成果描述，60-100字）。数据需真实可信。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "能力成长",
            "subtitle": "专业提升与经验沉淀",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "timeline",
            "title": "专业技能提升路径",
            "content_type": "timeline",
            "kicker": "能力成长",
            "items": [
                {
                    "period": "Q1 2024",
                    "title": "产品战略与规划",
                    "desc": "参加公司「产品战略工作坊」，系统学习市场分析、竞品研究、路线图制定方法论，输出智能家居3年产品规划报告。"
                },
                {
                    "period": "Q2 2024",
                    "title": "用户研究方法升级",
                    "desc": "主导引入「设计思维」工作坊，组织12场用户深访，提炼出6类核心用户画像，产出《智能家居用户需求白皮书》。"
                },
                {
                    "period": "Q3 2024",
                    "title": "数据分析能力深化",
                    "desc": "自学SQL和Python数据分析，完成数据分析师认证，构建智能家居产品数据看板，实现核心指标覆盖率从40%提升至95%。"
                },
                {
                    "period": "Q4 2024",
                    "title": "团队管理与领导力",
                    "desc": "参加「MTP中层管理者培训」，学习目标管理、绩效辅导、跨部门协作技巧，带领团队完成3个关键项目交付。"
                }
            ],
            "notes": "时间线从上到下或从左到右排列，每个节点标注季度、能力领域和学习成果",
            "filling_prompt": "必须填入真实内容：提供4个季度的能力提升记录，每个节点有 period（季度）、title（能力领域）、desc（具体学习内容和成果，40-80字）。确保内容真实，体现持续成长的过程。"
        },
        {
            "index": 10,
            "type": "kpi_dashboard",
            "title": "核心能力雷达评估",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "自我评估",
            "kpis": [
                {
                    "value": "4.2/5.0",
                    "label": "产品规划能力",
                    "delta": "+0.5",
                    "baseline": "vs 上年 (3.7/5.0)"
                },
                {
                    "value": "4.0/5.0",
                    "label": "数据分析能力",
                    "delta": "+1.0",
                    "baseline": "vs 上年 (3.0/5.0)"
                },
                {
                    "value": "4.5/5.0",
                    "label": "跨部门协作",
                    "delta": "+0.3",
                    "baseline": "vs 上年 (4.2/5.0)"
                },
                {
                    "value": "3.8/5.0",
                    "label": "团队管理能力",
                    "delta": "+0.6",
                    "baseline": "vs 上年 (3.2/5.0)"
                }
            ],
            "notes": "2x2布局展示4项核心能力的自评分数，每个指标标注提升幅度",
            "filling_prompt": "必须填入真实内容：提供4个能力评估指标，每个有 value（评分）、label（能力名称）、delta（同比变化）、baseline（上年分数）。数据来自实际自评或360评估。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "问题反思",
            "subtitle": "不足与改进",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 12,
            "type": "two_column",
            "title": "不足分析与改进计划",
            "content_type": "two_column",
            "kicker": "自我剖析 · 持续成长",
            "left_header": "不足与反思",
            "left_sections": {
                "analysis": [
                    "技术深度不足：对硬件技术和嵌入式系统的理解停留在表面，影响了与研发团队的高效沟通，有时难以准确评估技术方案的可行性和工作量。",
                    "跨部门影响力有限：在资源协调和跨部门推动时，往往依赖上级支持，缺乏独立推动多方达成共识的能力，导致部分项目进度受制于部门墙。",
                    "创新突破不够：日常工作被项目交付占满，缺乏对前沿技术和用户趋势的系统性研究，在产品创新上的突破性想法较少，战略视野有待拓宽。"
                ]
            },
            "right_header": "改进计划",
            "right_sections": {
                "key_points": [
                    "技术补强：每季度完成1门硬件技术课程学习，主动参与2次研发技术评审会议，记录并复盘技术决策过程，Q4前输出《智能硬件产品经理技术手册》初稿。",
                    "影响力提升：主动发起跨部门例会，建立与研发、设计、供应链的固定沟通机制，Q2前完成智能家居产品研发流程优化方案并推动落地。",
                    "创新机制：建立「创新灵感库」，每周收集3个行业创新案例，每月输出1篇竞品或行业趋势分析报告，Q3前孵化1个创新产品概念进入立项评审。"
                ]
            },
            "notes": "左右对比，左侧3个不足，右侧3个对应改进措施，一一对应形成闭环",
            "filling_prompt": "必须填入真实内容：left_header 为'不足与反思'，left_sections.analysis 列出3条工作中的不足（每条说明具体表现和影响，60-100字）；right_header 为'改进计划'，right_sections.key_points 列出对应的改进措施（每条说明具体行动、计划和时间节点）。注意不足和改进一一对应。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "未来规划",
            "subtitle": "下阶段目标与规划",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 14,
            "type": "process_flow",
            "title": "2025年度工作规划",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "kicker": "年度规划",
            "steps": [
                {"num": "Q1", "title": "产品力夯实", "desc": "完成智能门锁产品线年度迭代规划，推动3款新品立项；启动全屋智能解决方案2.0规划，整合现有产品线形成场景化套餐"},
                {"num": "Q2", "title": "能力升级", "desc": "完成硬件技术专项培训，建立技术顾问咨询机制；主导推动跨部门协作流程优化，减少项目卡点，提升研发效率20%"},
                {"num": "Q3", "title": "创新突破", "desc": "完成1个创新产品概念孵化并进入立项评审；搭建用户共创体系，组织至少6场用户共创活动，沉淀产品创新素材"},
                {"num": "Q4", "title": "团队建设", "desc": "培养2名初级产品经理，输出标准化的产品工作方法论；完成产品团队OKR体系建设，实现团队目标对齐与高效执行"}
            ],
            "notes": "4个季度的目标与行动计划，横向或锯齿形流程图展示",
            "filling_prompt": "必须填入真实内容：提供4个季度的工作目标，每个有 num（季度）、title（目标名称）、desc（具体行动描述，50-80字）。目标要具体可衡量，行动计划要切实可行。"
        },
        {
            "index": 15,
            "type": "content_slide",
            "title": "团队协作与跨部门贡献",
            "content_type": "content_slide",
            "kicker": "协作贡献",
            "layout_hint": "icon_text_list",
            "header": "跨部门协作与组织贡献",
            "sections": [
                {
                    "title": "研发协作",
                    "icon": "code",
                    "body": "与研发团队建立「产品技术双周会」机制，需求评审效率提升40%，技术方案返工率从35%降至12%，成为跨部门协作标杆案例。"
                },
                {
                    "title": "设计共创",
                    "icon": "pen-tool",
                    "body": "主导产品设计评审流程优化，建立「设计走查Checklist」标准，与设计团队协作完成3个产品线的UI规范统一，用户满意度调研显示界面美观度评分提升22%。"
                },
                {
                    "title": "供应链协同",
                    "icon": "truck",
                    "body": "推动「研产协同」机制，将产品需求提前3个月同步给供应链团队，样品交付周期从45天缩短至28天，有效支撑了新品快速上市节奏。"
                },
                {
                    "title": "销售赋能",
                    "icon": "trending-up",
                    "body": "为一线销售团队提供产品培训12场，输出标准化销售话术和竞品对比手册，协助销售攻克3个重点客户，带来直接订单超过800万元。"
                }
            ],
            "notes": "带图标的列表式布局，展示4个跨部门协作亮点",
            "filling_prompt": "必须填入真实内容：header 为页面标题（不超过35字）；sections 列出4个跨部门协作贡献，每项有 title（协作领域）、icon（图标名称）、body（40-80字描述）。确保描述真实，体现对组织的价值贡献。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "年度总结",
            "key_points": [
                "01 核心成果：主导3条产品线管理，GMV贡献2.3亿元，同比+47%，核心产品智能门锁Pro上市首月破2万台",
                "02 主要收获：从执行者向规划者转型，数据分析能力和团队管理能力显著提升，4项核心能力评分均有进步",
                "03 努力方向：补强技术深度、提升跨部门影响力、 建立创新机制，为公司智能家居战略贡献更大价值"
            ],
            "thank_you": "感谢各位领导和同事的支持与帮助，请大家批评指正！",
            "notes": "简洁有力的结尾，3条总结要清晰有力",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心成果+主要收获+努力方向），每条不超过60字。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "述职要实事求是，成果和问题都要有",
        "数据要真实，用具体数字说话",
        "反思要诚恳，改进计划要可行",
        "计划要具体，有时间节点",
        "PPT要简洁，不要堆砌文字",
        "用数据展示成果，而非罗列工作内容",
        "适当使用对比图展示进步",
        "准备Q&A，预判领导可能的问题"
    ],
    "self_review_template": {
        "performance": {
            "kpis": "KPI完成情况",
            "projects": "重点项目参与",
            "innovation": "创新贡献"
        },
        "capability": {
            "professional": "专业能力",
            "management": "管理能力",
            "leadership": "领导力"
        },
        "collaboration": {
            "teamwork": "团队协作",
            "cross_function": "跨部门协作",
            "stakeholder": "干系人管理"
        }
    },
    "common_questions": {
        "achievements": "最满意的成就是什么？",
        "challenges": "遇到的最大挑战是什么？",
        "improvements": "觉得自己还需要提升什么？",
        "next_year": "对下一年有什么规划？"
    }
}
