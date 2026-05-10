TEMPLATE = {
    "name": "personal-summary",
    "name_cn": "述职报告",
    "description": "适合个人总结、述职报告、年终总结等场景。重点突出，成果可见，计划明确。",
    "target_audience": "领导、同事、评审",
    "typical_slides": 12,
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
            "subtitle": "张明 | 产品部产品经理",
            "author": "张明",
            "date": "2025年1月15日",
            "notes": "标题页简洁正式",
            "filling_prompt": "必须填入真实内容：title 为时间段（如'2024年度'）+ 述职报告，subtitle 为姓名和岗位信息，author 为姓名，date 为述职日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  工作概述",
                "02  主要成果",
                "03  问题反思",
                "04  未来计划"
            ],
            "notes": "让领导快速了解汇报结构",
            "filling_prompt": "目录页为固定结构。"
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
            "title": "工作概览",
            "content_type": "example_detail",
            "kicker": "实例 · 工作概览",
            "lede": "负责B端核心产品，全年交付3个重大版本，支撑业务增长",
            "context_block": "本人负责B端核心产品'智云CRM'的产品规划和迭代工作。产品服务于1500+企业客户，年收入贡献超过5000万元。本人在团队中承担核心模块负责人角色，同时负责3名初级产品经理的带教工作。",
            "solution_block": "2024年主要工作包括：1）完成V3.0至V3.3三个大版本的规划与交付；2）主导AI功能模块从0到1的建设；3）优化需求评审流程，提升需求通过率30%；4）带教3名新入职产品经理，2人已独立承接模块。",
            "metrics": [
                {"value": "3个", "label": "完成大版本", "trend": "vs 计划2个"},
                {"value": "100%", "label": "需求完成率", "trend": "vs KPI 95%"},
                {"value": "8人", "label": "团队协作", "trend": "跨5个部门"}
            ],
            "takeaway": "启示：聚焦核心价值，平衡短期交付与长期建设",
            "notes": "用数据展示工作量和成果",
            "filling_prompt": "必须填入真实内容：lede 一句话概括本周期工作整体情况；context_block 描述述职人所在岗位职责和主要工作范围（1-2句话）；solution_block 具体说明本周期内完成的主要工作和取得的成绩（2-3句话）；metrics_grid 提供3个量化指标（如完成项目数、关键成果、团队协作次数），每个有 value、label、trend；takeaway 用一句话总结本周期工作的价值。禁止虚构数据。"
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
            "title": "核心业绩：AI功能模块建设",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "核心业绩",
            "header": "AI功能从0到1，赋能销售效率提升",
            "sub_header": "智能录入+成交预测+流失预警三大核心功能",
            "paragraph": "2024年，我主导了AI功能模块的建设工作，实现了该领域从0到1的突破。通过深入调研客户需求、对标竞品分析、反复验证MVP，最终交付了智能录入、成交预测、流失预警三大核心功能。上线6个月，功能使用率超过65%，客户反馈'大大提升了工作效率'。该模块预计明年贡献收入增量超过1000万元。",
            "notes": "用图文混排展示核心业绩，增强可信性",
            "filling_prompt": "必须填入真实内容：kicker 为'核心业绩'；title 中的 {项目名称} 替换为具体项目或工作名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为项目标题（不超过35字）；sub_header 为项目概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述项目的实施过程、克服的困难和取得的成果，用流畅的段落形式呈现，禁止罗列要点。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "其他成果",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "流程优化", "body": "优化需求评审流程，需求通过率从60%提升至78%"},
                {"header": "客户调研", "body": "完成50+场客户深访，形成3份专项洞察报告"},
                {"header": "团队带教", "body": "带教3名新人，2人已独立承接模块"},
                {"header": "知识沉淀", "body": "沉淀10+篇产品方法论文档，组织内部分享3场"}
            ],
            "notes": "4个其他重要成果",
            "filling_prompt": "必须填入真实内容：提供4个其他工作成果，每个有 header（成果名称）和 body（一句话描述成果内容）。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "问题反思",
            "subtitle": "不足与改进",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "two_column",
            "title": "反思与改进",
            "content_type": "two_column",
            "kicker": "自我剖析 · 持续成长",
            "left_header": "不足与反思",
            "left_sections": {
                "key_points": [
                    "技术深度欠缺：与研发团队沟通技术方案时，有时理解不够深入，影响需求评审质量和推进效率",
                    "跨部门协调能力有待提升：涉及多部门协作的复杂项目，资源争取和进度协调中存在明显短板",
                    "文档规范性不足：部分需求文档在边界条件、异常处理等细节上描述不够完善，给开发带来返工"
                ],
                "analysis": [
                    "根因分析：上述不足均源于经验积累不够和对关联领域知识储备不足",
                    "影响评估：技术理解不深导致需求颗粒度不够细，跨部门沟通效率低增加了项目延期的风险"
                ]
            },
            "right_header": "改进计划",
            "right_sections": {
                "key_points": [
                    "技术提升：系统学习后端基础课程（数据库、API设计、性能优化），已报名内部技术训练营，下季度完成3门核心课程",
                    "协作技巧：向有丰富跨部门项目经验的同事请教，梳理出高效沟通的话术模板，在下个项目中使用并迭代优化",
                    "文档规范：建立需求文档 Checklist，涵盖前置条件、正常流程、异常流程、接口契约、边界值等要素，完成后评审通过率提升至90%"
                ],
                "analysis": [
                    "预期效果：通过上述改进，需求评审一次通过率从65%提升至80%以上",
                    "成长周期：预计在下一个半年周期内实现上述改进目标"
                ]
            },
            "notes": "坦诚分析不足，给出具体改进计划，展示自我驱动的成长意愿",
            "filling_prompt": "必须填入真实内容：left_header 为'不足与反思'，left_sections.key_points 列出2-4条工作中的不足（每条说明具体表现和影响），left_sections.analysis 列出2条根因分析；right_header 为'改进计划'，right_sections.key_points 列出对应的改进措施（每条说明具体行动、计划和时间节点），right_sections.analysis 列出预期效果。注意不足和改进一一对应。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "未来计划",
            "subtitle": "下阶段目标与规划",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "process_flow",
            "title": "下阶段计划",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "产品升级", "desc": "完成V4.0版本规划，主打智能化升级"},
                {"num": "02", "title": "能力提升", "desc": "系统学习数据分析，提升数据驱动能力"},
                {"num": "03", "title": "团队建设", "desc": "培养2名初级PM独立承接完整模块"},
                {"num": "04", "title": "行业深耕", "desc": "深入研究行业头部客户，形成方法论"}
            ],
            "notes": "4个下阶段工作目标",
            "filling_prompt": "必须填入真实内容：提供4个下阶段工作目标，每个有 title（目标名称）和 desc（具体行动）。目标要具体可衡量。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心业绩：完成3个大版本交付，主导AI功能从0到1",
                "02 主要收获：产品方法论沉淀，团队带教经验积累",
                "03 努力方向：提升技术深度，加强跨部门协调能力"
            ],
            "thank_you": "感谢聆听，请领导批评指正！",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心成果+主要收获+努力方向）。禁止保留花括号。"
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
