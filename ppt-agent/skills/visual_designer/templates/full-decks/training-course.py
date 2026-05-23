TEMPLATE = {
    "name": "training-course",
    "name_cn": "培训课件",
    "description": "适合内部培训、新人入职培训、技能培训等场景。知识系统，讲解清晰，互动引导。",
    "target_audience": "员工、新人、需要学习相关技能的人员",
    "typical_slides": 20,
    "typical_duration": "45-90分钟",
    "palette": "education_blue",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "training_philosophy": "培训的目标是让学员'学会'而非'听过'",
        "key_principles": [
            "内容要由浅入深，循序渐进",
            "概念讲解要准确，举例要贴切",
            "图文并茂，增强理解",
            "留出互动和练习时间"
        ],
        "adult_learning": {
            "relevance": "培训内容要与工作实际相关",
            "experience": "善于利用学员已有经验",
            "problem_orientation": "聚焦解决实际问题",
            "autonomy": "给予学员一定选择权"
        }
    },
    "training_objectives": {
        "knowledge": "知识目标：理解数据驱动决策的核心概念与方法论",
        "skill": "技能目标：能够使用常用数据分析工具完成业务数据解读",
        "attitude": "态度目标：建立用数据说话、用数据决策的工作习惯"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "数据驱动决策：从理念到实践",
            "subtitle": "基于业务数据的分析思维与决策方法训练",
            "author": "数据分析部 · 李明",
            "date": "2026年5月",
            "notes": "标题页正式，注明培训主题和讲师",
            "filling_prompt": "必须填入真实内容：title 为培训主题名称，subtitle 为培训副标题，author 为培训讲师姓名和部门，date 为培训日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "培训目标",
            "kicker": "学习目标",
            "section_header": "本次培训的三大目标",
            "bullets": [
                "能够识别业务场景中的关键数据指标",
                "掌握常用的数据描述方法与图表选择",
                "理解从数据到洞察、从洞察到决策的完整链路",
                "能够在日常工作中应用数据思维做判断"
            ],
            "highlight_stats": [
                {"value": "4大", "label": "核心能力模块"},
                {"value": "3次", "label": "实战练习"}
            ],
            "notes": "明确培训要达到的目标",
            "filling_prompt": "必须填入真实内容：列出3-5条培训目标，每条说明学员学完本次培训后能够做什么（用动词开头，如'能够...'、'掌握...'、'理解...'）。highlight_stats 可选提供右侧数据卡片。"
        },
        {
            "index": 3,
            "type": "agenda",
            "title": "课程内容",
            "kicker": "目录",
            "items": [
                "01  基础知识：数据驱动决策概述",
                "02  核心技能：常用数据分析方法",
                "03  实战演练：真实业务场景分析",
                "04  总结回顾：知识巩固与行动计划"
            ],
            "notes": "让学员了解培训结构",
            "filling_prompt": "目录页为固定结构，可根据实际内容调整章节。items 格式为 'XX  章节名称'，共4个章节。"
        },
        {
            "index": 4,
            "type": "section_divider",
            "number": "01",
            "title": "基础知识",
            "subtitle": "数据驱动决策概述",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "什么是数据驱动决策",
            "kicker": "核心概念",
            "section_header": "Data-Driven Decision Making",
            "bullets": [
                "数据驱动决策：用客观数据替代主观经验进行业务判断",
                "核心三要素：数据采集 → 数据分析 → 决策行动",
                "与传统经验决策相比：更客观、更可量化、更可追溯",
                "适用于：运营优化、产品迭代、市场策略、风险管理等场景",
                "关键前提：数据质量、分析能力、业务理解三者缺一不可"
            ],
            "notes": "讲解核心概念和定义",
            "filling_prompt": "必须填入真实内容：title 为页面标题，kicker 为标签（如'核心概念'），section_header 为英文小节标题，bullets 列出4-5条，每条不超过35字。禁止保留花括号。"
        },
        {
            "index": 6,
            "type": "deep_dive",
            "title": "数据驱动 vs 经验驱动",
            "kicker": "方法论对比",
            "lede": "两种决策方式各有优劣，数据驱动并非万能，但能让决策更有据可依",
            "left_header": "数据驱动决策",
            "key_points": [
                "优势：客观量化、可重复验证、适合规模化",
                "优势：能发现经验盲区，挖掘隐藏规律",
                "局限：依赖数据质量和完整性",
                "局限：需要分析能力和工具支撑",
                "局限：复杂场景下数据可能滞后"
            ],
            "analysis": [
                "适用场景：高频重复决策、有大量历史数据支撑",
                "不适用：全新领域、数据缺失、需快速直觉响应"
            ],
            "right_header": "经验驱动决策",
            "case_example": [
                "优势：灵活快速，适合不确定性和创新探索",
                "优势：可综合隐性知识和直觉判断",
                "优势：对突发情况和非结构化问题更有效",
                "局限：主观偏差，难以规模化复制",
                "局限：知识无法沉淀和传承"
            ],
            "data_evidence": [
                "实践表明：两者结合效果最佳",
                "建议比例：70%数据参考 + 30%经验判断",
                "关键：明确何时用数据、何时用经验"
            ],
            "notes": "图文详解，左要点右案例",
            "filling_prompt": "必须填入真实内容：kicker 为领域标签，title 为主题，lede 为一句话概括，left_header 和 right_header 为左右栏标题，key_points 和 case_example 各列出3-5条，analysis 列出2条，data_evidence 列出2-3条。禁止保留花括号。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "02",
            "title": "核心技能",
            "subtitle": "常用数据分析方法",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "process_flow",
            "title": "数据分析六步法",
            "kicker": "方法论",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "明确问题", "desc": "确定业务目标与核心问题"},
                {"num": "02", "title": "定义指标", "desc": "拆解为可量化的数据指标"},
                {"num": "03", "title": "数据采集", "desc": "获取相关数据，确保质量"},
                {"num": "04", "title": "数据分析", "desc": "描述统计、对比分析、趋势分析"},
                {"num": "05", "title": "洞察提炼", "desc": "从数据中发现规律和机会"},
                {"num": "06", "title": "决策行动", "desc": "形成建议，推进落地执行"}
            ],
            "notes": "用流程图展示操作步骤",
            "filling_prompt": "必须填入真实内容：提供6个步骤，每个步骤有 title（步骤名称，不超过20字）和 desc（操作说明，不超过25字）。"
        },
        {
            "index": 9,
            "type": "example_detail",
            "title": "用户留存率下降分析",
            "kicker": "实例 · 运营分析",
            "lede": "通过数据拆解找到留存率下降的根因，制定针对性提升策略",
            "context_block": "某产品月活跃用户留存率从65%下降至52%，直接影响收入增长。产品团队初步判断为竞品冲击，但具体原因不明。",
            "solution_block": "采用漏斗分析方法，逐层拆解用户流失节点。关键发现：流失主要集中在首次使用后的第3-7天，主要原因是'功能复杂、上手困难'。据此设计新用户引导流程，留存率3个月后回升至68%。",
            "metrics": [
                {"value": "13%", "label": "留存率降幅", "trend": "↓"},
                {"value": "68%", "label": "回升后留存率", "trend": "↑"},
                {"value": "3月", "label": "恢复周期", "trend": ""},
                {"value": "5步", "label": "新用户引导", "trend": "优化"}
            ],
            "takeaway": "数据拆解是找到真因的关键，避免凭直觉做判断",
            "notes": "案例背景、方案、成果完整呈现",
            "filling_prompt": "必须填入真实内容：kicker 为领域标签，title 为案例名称，lede 为一句话概括（不超过30字），context_block 描述背景（1-2句话），solution_block 描述方案（2-3句话），metrics 提供4个指标（value、label、trend），takeaway 用一句话总结。禁止保留花括号。"
        },
        {
            "index": 10,
            "type": "three_column",
            "title": "常用数据分析方法",
            "kicker": "技能工具箱",
            "columns": [
                {"header": "描述性分析", "bullets": ["对比分析：同比/环比/竞品", "分布分析：均值/中位数/分位数", "结构分析：构成比例与排名"]},
                {"header": "诊断性分析", "bullets": ["漏斗分析：转化率与流失节点", "归因分析：多因素贡献度", "相关性分析：变量间关联强度"]},
                {"header": "预测性分析", "bullets": ["趋势预测：时间序列模型", "分类预测：用户分层与标签", "异常检测：异常值自动识别"]}
            ],
            "notes": "三栏并列展示多维度信息",
            "filling_prompt": "必须填入真实内容：title 为页面标题，kicker 为标签，columns 提供3个栏，每栏有 header（栏标题，不超过15字）和 bullets（3个要点，每条不超过20字）。禁止保留花括号。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "03",
            "title": "实战演练",
            "subtitle": "真实业务场景分析",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 12,
            "type": "kanban",
            "title": "练习任务看板",
            "kicker": "实战练习",
            "subtitle": "小组协作完成真实业务数据分析",
            "columns": [
                {"title": "待完成", "color": "text_muted", "cards": [
                    {"text": "任务一：GMV下降归因分析", "tag": "分析", "priority": "high"},
                    {"text": "任务二：用户分群画像", "tag": "分群", "priority": "medium"},
                    {"text": "任务三：活动效果评估", "tag": "评估", "priority": "high"}
                ]},
                {"title": "进行中", "color": "secondary", "cards": [
                    {"text": "任务四：竞品数据对标", "tag": "对标", "priority": "medium"},
                    {"text": "任务五：A/B测试设计", "tag": "实验", "priority": "high"}
                ]},
                {"title": "已完成", "color": "primary", "cards": [
                    {"text": "任务六：数据看板搭建", "tag": "看板", "priority": "done"},
                    {"text": "任务七：指标体系设计", "tag": "体系", "priority": "done"}
                ]}
            ],
            "progress": 45,
            "notes": "任务看板展示练习安排",
            "filling_prompt": "必须填入真实内容：title 为页面标题，subtitle 为副标题，columns 提供3个栏，每栏有 title、color、cards（每张卡片有 text、tag、priority），progress 为整体进度百分比（0-100）。禁止保留花括号。"
        },
        {
            "index": 13,
            "type": "case_study",
            "title": "某电商平台大促活动效果评估",
            "kicker": "案例 · 运营",
            "context": "双十一大促期间，平台投入营销费用5000万元，包含优惠券、限时折扣、直播引流等组合策略。活动结束后需评估整体ROI和各渠道贡献。",
            "problem": "传统统计方式难以区分各渠道交叉影响，GMV数据包含凑单退款等噪音，且未考虑品牌长期价值贡献，单一ROI指标无法全面评价活动效果。",
            "solution": "建立多维度评估体系：短期指标（GMV、订单量、新客转化）+ 中期指标（复购率、客单价）+ 长期指标（品牌搜索热度、NPS）。通过归因模型分配各渠道贡献，区分直接转化与协同促进作用。",
            "results": [
                {"metric": "整体ROI", "value": "3.2倍", "comparison": "vs 目标2.5倍"},
                {"metric": "新客占比", "value": "38%", "comparison": "vs 平时15%"},
                {"metric": "30日复购率", "value": "42%", "comparison": "vs 平时35%"},
                {"metric": "品牌搜索指数", "value": "+65%", "comparison": "vs 活动前"}
            ],
            "notes": "背景-痛点-方案-成果四段式",
            "filling_prompt": "必须填入真实内容：kicker 为领域标签，title 为案例名称，context 为背景（1-2句话），problem 为痛点（1-2句话），solution 为解决方案（2-3句话），results 提供4个成果指标（metric、value、comparison）。禁止保留花括号。"
        },
        {
            "index": 14,
            "type": "two_column",
            "title": "常用数据分析工具对比",
            "kicker": "工具选型",
            "left_header": "Excel + SQL",
            "right_header": "Python + BI工具",
            "left_bullets": [
                "适合：日常报表、快速验证想法",
                "优点：门槛低、灵活性高、人人会用",
                "优点：无需编程，快速出图",
                "局限：百万级以上数据处理较慢",
                "局限：复杂分析逻辑编写困难",
                "适用人群：业务分析师、产品经理"
            ],
            "right_bullets": [
                "适合：大规模数据、复杂模型",
                "优点：处理速度快，支持自动化",
                "优点：可视化丰富，报告美观",
                "局限：有一定学习曲线",
                "局限：需要基础编程能力",
                "适用人群：数据分析师、数据科学家"
            ],
            "source": "内部调研 · 2026年Q1",
            "notes": "左右对比，突显差异化",
            "filling_prompt": "必须填入真实内容：title 为对比标题，left_header 和 right_header 为左右栏标题，left_bullets 和 right_bullets 各列出5-6条对比要点，每条不超过25字。source 为数据来源。禁止保留花括号。"
        },
        {
            "index": 15,
            "type": "section_divider",
            "number": "04",
            "title": "总结回顾",
            "subtitle": "知识巩固与行动计划",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 16,
            "type": "image_text",
            "title": "学习资源推荐",
            "kicker": "资源 · 推荐",
            "layout": "left-image",
            "header": "持续学习，数据能力进阶路径",
            "paragraph": "数据分析能力提升需要持续学习与实践。建议从夯实统计基础开始，逐步掌握SQL取数、Excel高级分析、BI可视化等工具，最终形成数据思维与业务洞察的综合能力。推荐从工作中实际项目入手，边学边练，快速迭代。",
            "bullets": [
                "内部课程：数据分析基础（已完成）",
                "进阶课程：SQL高级查询与数据建模",
                "工具课程：Tableau/PowerBI可视化实战",
                "实战项目：加入数据驱动专项小组"
            ],
            "sub_header": "学习路径建议：基础 → 工具 → 实战 → 沉淀",
            "notes": "左图右文，展示学习资源",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为资源主题（不超过35字）；sub_header 为副标题；paragraph 为300-450字的自然语言段落，用流畅的段落形式呈现，禁止罗列要点；bullets 为推荐学习资源列表（3-5条）。禁止保留花括号。"
        },
        {
            "index": 17,
            "type": "stat_slide",
            "title": "培训效果数据",
            "kicker": "数据 · 回顾",
            "subtitle": "历次同类培训效果统计",
            "stats": [
                {"number": "92", "unit": "%", "label": "知识掌握率", "trend": "↑ 8%"},
                {"number": "85", "unit": "%", "label": "满意度", "trend": "↑ 12%"},
                {"number": "3周", "unit": "", "label": "平均见效周期", "trend": "↓ 1周"}
            ],
            "notes": "用关键数据展示培训效果",
            "filling_prompt": "必须填入真实内容：提供3个关键指标，每个有 number（数字）、unit（单位）、label（指标说明）、trend（趋势）。数字使用真实数据。禁止保留花括号。"
        },
        {
            "index": 18,
            "type": "kpi_dashboard",
            "title": "学习效果自评",
            "kicker": "自评 · 效果",
            "subtitle": "对照培训目标评估学习成果",
            "layout_hint": "2x2",
            "kpis": [
                {"value": "掌握", "label": "数据驱动决策概念", "delta": "已达标", "baseline": "目标：理解核心概念"},
                {"value": "熟悉", "label": "数据分析方法", "delta": "进行中", "baseline": "目标：掌握六步法"},
                {"value": "应用", "label": "实际工作应用", "delta": "计划中", "baseline": "目标：完成1个实践项目"},
                {"value": "持续", "label": "后续学习计划", "delta": "制定中", "baseline": "推荐：进阶课程"}
            ],
            "notes": "学习效果多维度自评",
            "filling_prompt": "必须填入真实内容：title 为页面标题，subtitle 为副标题，kicker 为标签，kpis 提供4个学习效果自评维度，每个有 value（如'掌握'、'熟悉'等）、label（维度名称）、delta（当前状态）、baseline（目标说明）。"
        },
        {
            "index": 19,
            "type": "brand_focus",
            "title": "核心理念",
            "kicker": "理念倡导",
            "subtitle": "用数据说话，让决策有据可依",
            "center_text": "数据\n思维",
            "surrounding_points": [
                {"title": "客观", "description": "用数据替代直觉做判断", "color": "secondary", "angle": 45},
                {"title": "量化", "description": "一切皆可度量才有优化可能", "color": "accent", "angle": 135},
                {"title": "迭代", "description": "持续验证假设，不断改进", "color": "secondary", "angle": 225},
                {"title": "协作", "description": "数据团队与业务团队紧密配合", "color": "accent", "angle": 315}
            ],
            "principles": [
                {"title": "先问问题，再找数据", "description": "避免数据挖掘中的目标漂移"},
                {"title": "相关性不等于因果性", "description": "谨慎推导，避免错误归因"},
                {"title": "数据会说谎", "description": "关注数据来源和口径，警惕统计陷阱"},
                {"title": "行动才是终点", "description": "分析最终要为决策和行动服务"}
            ],
            "notes": "核心理念以视觉化方式呈现",
            "filling_prompt": "必须填入真实内容：title 为页面标题，subtitle 为副标题，kicker 为标签，center_text 为中心圆圈文字（用换行分隔两行），surrounding_points 提供4个围绕中心的价值点（title、description、angle），principles 提供4条核心理念（title、description）。禁止保留花括号。"
        },
        {
            "index": 20,
            "type": "summary_slide",
            "title": "课程总结",
            "key_points": [
                "01  数据驱动决策是用客观数据替代主观经验进行业务判断的核心方法论",
                "02  数据分析六步法：明确问题→定义指标→数据采集→数据分析→洞察提炼→决策行动",
                "03  实战练习：GMV归因分析、用户分群、活动效果评估是常见应用场景",
                "04  行动计划：选择一个实际业务问题，运用本次所学方法完成分析报告"
            ],
            "thank_you": "感谢聆听！",
            "contact": "李明  |  数据分析部  |  liming@company.com",
            "notes": "简洁总结，提出后续学习建议",
            "filling_prompt": "必须填入真实内容：key_points 提供4个要点（格式为 'XX  核心内容'，前3条为知识总结，第4条为行动计划），thank_you 为结束语，contact 为联系方式。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "培训内容要由浅入深，循序渐进",
        "概念讲解要准确，举例要贴切",
        "图文并茂，增强理解",
        "留出互动和练习时间",
        "结尾要有后续学习指引",
        "配套测试题，检验学习效果",
        "收集反馈，持续改进培训内容",
        "提供学习资料，方便后续复习"
    ],
    "training_materials": {
        "handouts": [
            "培训课件PDF",
            "数据分析六步法操作手册",
            "常用指标定义词典",
            "练习数据集与参考答案"
        ],
        "online_resources": [
            "内部知识库 · 数据分析专栏",
            "学习视频中心 · 进阶课程入口",
            "导师答疑群 · 扫码入群"
        ]
    },
    "assessment_methods": {
        "during_training": "课堂随堂测验、练习表现打分",
        "after_training": "在线测试（满分100，70分及格）",
        "on_job": "1个月后提交实际业务分析报告"
    }
}
