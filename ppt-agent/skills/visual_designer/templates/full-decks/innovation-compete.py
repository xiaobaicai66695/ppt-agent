TEMPLATE = {
    "name": "innovation-compete",
    "name_cn": "科创竞赛",
    "description": "适合大创、挑战杯、互联网+等科创竞赛汇报。创新性强，数据支撑，展示潜力。",
    "target_audience": "评委、投资人、导师、参赛团队",
    "typical_slides": 16,
    "typical_duration": "10-15分钟（路演）",
    "palette": "civic_gold",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "competition_strategy": "竞赛路演的核心是展示：创新性+可行性+团队力",
        "key_elements": [
            "项目创新性：技术/模式/应用的创新点",
            "项目可行性：技术实现+商业模式验证",
            "团队执行力：背景互补+分工明确",
            "市场潜力：市场规模+增长预期"
        ],
        "judging_criteria": {
            "innovation": "创新性：30%",
            "feasibility": "可行性：25%",
            "commercial": "商业价值：20%",
            "team": "团队能力：15%",
            "presentation": "展示效果：10%"
        }
    },
    "pitch_tips": {
        "opening": "30秒内抓住评委注意力",
        "problem": "清晰阐述问题和市场机会",
        "solution": "用一句话说清楚你的解决方案",
        "traction": "展示已有的进展和成果",
        "ask": "明确说明你需要什么"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "智谱文献——基于知识图谱的智能学术推荐系统",
            "subtitle": "让科研更高效，让发现更简单",
            "author": "智创未来团队",
            "date": "2026年5月",
            "notes": "标题页醒目，吸引眼球",
            "filling_prompt": "固定结构，内容已填入：title 为项目名称，subtitle 为一句概括性口号，author 为团队名称，date 为参赛年份。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "内容概览",
            "kicker": "目录",
            "items": [
                "01  市场分析",
                "02  产品介绍",
                "03  技术方案",
                "04  商业模式",
                "05  团队介绍",
                "06  未来展望"
            ],
            "notes": "让评委了解汇报结构",
            "filling_prompt": "固定结构，目录页内容已填入。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "市场分析",
            "subtitle": "需求与机会",
            "notes": "第一章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "学术文献检索的困境",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "市场分析",
            "header": "科研人员面临的信息过载难题",
            "sub_header": "传统检索方式已无法满足深度学术研究需求",
            "paragraph": "随着全球科研论文数量年均增长超过8%，学术研究者平均每天需要浏览超过50篇文献摘要。然而，传统关键词检索返回的结果相关性普遍低于40%，研究者平均花费2.3小时才能找到一篇真正有用的参考文献。更严重的是，孤立检索无法发现跨学科的潜在关联，导致大量有价值的研究线索被遗漏。本项目正是针对这一痛点，利用知识图谱技术实现语义级的智能推荐。",
            "references": [
                "https://www.scimagojr.com/"
            ],
            "notes": "用图文混排展示市场需求分析",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料，再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为市场主题（不超过35字）；sub_header 为市场概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细分析市场规模、用户需求和痛点，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "stat_slide",
            "title": "市场规模",
            "content_type": "stat_slide",
            "stats": [
                {"value": "4.8亿", "label": "全球科研人员数量", "note": "Web of Science 2025"},
                {"value": "1.2万亿", "label": "学术出版市场规模（美元）", "note": "STM报告2025"},
                {"value": "340亿美元", "label": "科研工具软件市场", "note": "Grand View Research 2025"},
                {"value": "38%", "label": "文献检索效率提升空间", "note": "Nature调查数据"}
            ],
            "notes": "大数字展示市场潜力",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料，再填入真实内容：提供市场规模数据（大数字突出）。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "产品介绍",
            "subtitle": "核心功能与亮点",
            "notes": "第二章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "四大核心功能",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "语义级智能推荐", "body": "基于知识图谱理解研究意图，推荐相关性提升65%"},
                {"header": "跨学科关联发现", "body": "自动识别跨领域引用关系，发现潜在创新点"},
                {"header": "文献知识网络", "body": "可视化呈现文献间的引用、衍生、对立关系"},
                {"header": "个性化追踪", "body": "一键订阅作者、机构、研究方向动态更新"}
            ],
            "notes": "展示4个核心功能",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称）和 body（一句话描述功能价值）。功能要具体。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "技术方案",
            "subtitle": "技术架构与创新",
            "notes": "第三章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 9,
            "type": "deep_dive",
            "title": "核心技术方案",
            "content_type": "deep_dive",
            "kicker": "技术 · 核心创新",
            "lede": "知识图谱+大语言模型双引擎驱动，实现学术文献的语义理解与智能推荐",
            "left_column": {
                "key_points": [
                    "学术实体识别：论文、作者、机构、方法的自动化抽取",
                    "关系抽取：引用、衍生、方法论、实验数据的关系建模",
                    "语义嵌入：将文献投影到高维向量空间计算相似度",
                    "增量更新：实时接入arXiv、PubMed等预印本平台数据"
                ],
                "analysis": [
                    "技术壁垒：自建学术知识图谱包含2300万实体节点",
                    "创新点：首次将LLM与知识图谱深度融合用于推荐"
                ]
            },
            "right_column": {
                "case_example": [
                    "用户输入研究主题「Transformer在生物信息学的应用」",
                    "系统自动构建子图查询，识别核心概念节点",
                    "LLM生成查询扩展，结合知识图谱路径推理",
                    "返回相关论文及跨学科引用关系，按创新潜力排序"
                ],
                "data_evidence": [
                    "推荐准确率：82.3%（对比传统方法提升42%）",
                    "平均检索时间：缩短至3.2分钟（传统方式需28分钟）",
                    "跨学科发现率：每位用户平均发现2.7篇关联文献"
                ]
            },
            "notes": "展示技术实力和创新点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料，再填入真实内容：left_column.key_points 为技术要点（3-4条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 URL。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "商业模式",
            "subtitle": "盈利与发展",
            "notes": "第四章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "运营数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 运营成果",
            "kpis": [
                {"value": "12,800", "label": "注册用户数", "delta": "+23%", "baseline": "上线首月1,050"},
                {"value": "67%", "label": "月活跃率", "delta": "+8pp", "baseline": "行业平均42%"},
                {"value": "3.8分钟", "label": "平均单次使用时长", "delta": "+62%", "baseline": "传统工具1.4分钟"},
                {"value": "4.6分", "label": "用户满意度（满分5）", "delta": "+0.4", "baseline": "Beta测试4.2分"}
            ],
            "notes": "用数据证明项目可行性",
            "filling_prompt": "必须填入真实内容：提供4个运营指标（如用户数、活跃度、收入、增长率等），每个有 value、label、delta、baseline。如项目处于早期阶段，可用预期数据但需注明。禁止虚构已完成的指标数据。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "团队介绍",
            "subtitle": "核心成员",
            "notes": "第五章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 13,
            "type": "card_grid",
            "title": "核心成员",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "张明远 | 项目负责人", "body": "计算机学院硕士，研究方向为知识图谱与NLP，发表CCF-A类论文2篇"},
                {"header": "李雨桐 | 技术负责人", "body": "人工智能学院博士，专注大模型微调与知识推理，主导图谱构建工作"},
                {"header": "王思齐 | 产品负责人", "body": "信息管理学院硕士，有3段互联网产品实习经验，负责用户体验设计"},
                {"header": "陈晓阳 | 商务负责人", "body": "经济管理学院本科生，校创业协会副会长，负责高校市场拓展"}
            ],
            "notes": "展示4位核心团队成员",
            "filling_prompt": "必须填入真实内容：提供4位核心成员信息，每人有 header（姓名+职位）和 body（教育背景、专业技能、主要贡献）。禁止虚构成员信息。"
        },
        {
            "index": 14,
            "type": "case_study",
            "title": "典型应用场景",
            "content_type": "case_study",
            "kicker": "案例 · 高校图书馆",
            "context": "某985高校图书馆年均借阅纸质文献12万册次，数字资源访问量超过800万次，但读者反馈查找效率低。",
            "problem": "学生平均每次检索耗时超过15分钟，仅有约20%的检索能直接找到目标文献，大量时间消耗在筛选和排除上。",
            "solution": "为该校部署智谱文献系统，嵌入图书馆OPAC系统。学生检索时可一键切换智能推荐模式，系统基于其历史行为和知识图谱推理进行个性化推荐，并展示跨学科关联论文。",
            "results": [
                {"metric": "检索耗时", "value": "-58%", "comparison": "从15分钟降至6.3分钟"},
                {"metric": "目标达成率", "value": "+73%", "comparison": "从20%提升至34.6%"},
                {"metric": "用户好评率", "value": "91%", "comparison": "较原有系统提升28pp"},
                {"metric": "跨学科发现", "value": "3.2篇/人", "comparison": "人均发现3.2篇关联文献"}
            ],
            "notes": "展示典型应用场景案例",
            "filling_prompt": "必须填入真实内容：context 描述应用场景背景；problem 描述核心痛点；solution 描述解决方案；results 提供4个量化指标（metric/value/comparison）。"
        },
        {
            "index": 15,
            "type": "process_flow",
            "title": "发展规划",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "kicker": "未来路线",
            "subtitle": "从概念验证到商业化落地的三阶段路径",
            "steps": [
                {"num": "01", "title": "校园深耕", "desc": "完成5所985高校部署，收集反馈优化产品"},
                {"num": "02", "title": "企业拓展", "desc": "与3家科技企业签署API合作协议"},
                {"num": "03", "title": "生态构建", "desc": "推出开放平台，吸引第三方开发者接入"},
                {"num": "04", "title": "国际化", "desc": "拓展英文市场，对接Web of Science数据"}
            ],
            "notes": "展示团队愿景和执行力",
            "filling_prompt": "固定结构，已填入4个发展规划阶段，每步有 num、title、desc。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "感谢聆听",
            "kicker": "竞赛答辩总结",
            "key_points": [
                "01  创新性：知识图谱+LLM双引擎，学术推荐准确率提升42%",
                "02  可行性：已完成高校部署验证，12,800用户，67%月活",
                "03  团队力：跨学科背景互补，技术与商务能力兼备",
                "04  合作诉求：寻求种子轮融资200万元，用于研发与市场拓展"
            ],
            "thank_you": "感谢聆听，欢迎交流！",
            "contact": "项目官网: zhituwenxian.edu.cn  |  团队邮箱: team@zhituwenxian.edu.cn",
            "notes": "展示团队愿景和联系方式",
            "filling_prompt": "必须填入真实内容：key_points 提供3-4个核心亮点总结；contact 填写真实联系方式。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "竞赛路演要抓住评委注意力",
        "数据要真实有力，来源可查",
        "突出创新点和竞争优势",
        "团队展示要体现专业性和互补性",
        "PPT设计要有视觉冲击力",
        "控制时间，每页只讲一个重点",
        "准备Q&A，预判评委可能的问题",
        "展示Passion，评委喜欢有热情的团队"
    ],
    "competition_prep": {
        "common_judging_questions": [
            "你的创新点是什么？",
            "技术壁垒在哪里？",
            "如何获取用户？",
            "项目的局限性是什么？",
            "团队如何分工？"
        ],
        "material_checklist": [
            "商业计划书",
            "产品Demo视频",
            "技术文档",
            "用户调研报告",
            "财务报表（如有）"
        ]
    }
}
