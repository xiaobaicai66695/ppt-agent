TEMPLATE = {
    "name": "pitch-deck",
    "name_cn": "商业计划/路演",
    "description": "适合创业路演、投资人演示、商业计划展示等场景。结构清晰，逻辑严密，数据驱动，说服力强。",
    "target_audience": "投资人、VC、潜在合作伙伴",
    "typical_slides": 16,
    "typical_duration": "10-15分钟",
    "palette": "charcoal_light",
    "typography": {
        "header": "Arial Black",
        "body": "Arial",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "investor_mindset": "投资人关注：市场规模、团队能力、商业模式、增长潜力、退出路径",
        "key_questions": [
            "市场规模有多大？（TAM/SAM/SOM）",
            "为什么是现在？（timing）",
            "为什么是你们？（team）",
            "如何赚钱？（model）",
            "护城河是什么？（moat）"
        ],
        "persuasion_tips": [
            "用数据说话，避免空泛描述",
            "讲真实故事，增加情感连接",
            "突出差异化，展现独特价值",
            "展现团队热情和执行力"
        ]
    },
    "deck_structure_logic": {
        "problem": "痛点足够大吗？",
        "solution": "方案足够好吗？",
        "market": "市场足够大吗？",
        "model": "能赚钱吗？",
        "traction": "有人在用吗？",
        "team": "团队能成吗？"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "智研科技——AI驱动的临床试验数据分析平台",
            "subtitle": "用AI加速新药研发，让患者更早获得有效治疗",
            "author": "刘思远 | CEO & 联合创始人",
            "date": "2026年5月",
            "notes": "简洁有力的封面页，突出核心Slogan",
            "filling_prompt": "固定结构，内容已填入：title 为公司或项目名称，subtitle 为一句有力的价值主张，author 为创始人姓名，date 为日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "内容概览",
            "kicker": "目录",
            "items": [
                "01  痛点与机会",
                "02  解决方案",
                "03  市场分析",
                "04  商业模式",
                "05  竞争优势",
                "06  运营数据",
                "07  团队融资"
            ],
            "notes": "让投资人快速了解报告结构",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "痛点与机会",
            "subtitle": "为什么现在",
            "notes": "第一章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "临床试验数据管理的行业痛点",
            "content_type": "example_detail",
            "kicker": "实例 · 市场痛点",
            "lede": "一款新药从研发到上市平均需要12年，数据分析效率低下是核心瓶颈之一",
            "context_block": "临床试验产生的海量数据（患者入组记录、药物反应、不良事件报告等）分散在多个CRO（合同研究组织）和医院系统中，数据格式不统一、标准各异。统计师平均每周花费超过20小时在数据清洗和标准化上，真正用于分析建模的时间不足30%。",
            "solution_block": "这一现状导致临床试验周期被人为拉长。以肿瘤领域为例，II期到III期试验的数据锁库（Database Lock）平均耗时长达6-8周，每延迟一天药企损失约100-300万美元。更严重的是，人工处理增加了数据错误风险，约有8%的关键数据需要返工核实。",
            "metrics": [
                {"value": "12年", "label": "新药研发平均周期", "trend": "近10年未显著缩短"},
                {"value": "20小时/周", "label": "统计师数据清洗耗时", "trend": "仅30%时间用于核心分析"},
                {"value": "100-300万", "label": "试验延期每日损失（美元）", "trend": "III期试验平均损失"}
            ],
            "takeaway": "AI驱动的数据分析可将数据锁库时间缩短70%，为药企节省数百万美元研发成本。",
            "notes": "用数据和故事说明目标市场的痛点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料，再填入真实内容：lede 一句话概括核心挑战；context_block 描述目标市场的普遍困境（1-2句话）；solution_block 具体说明这些困境导致的后果和对用户的影响（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结机会。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "解决方案",
            "subtitle": "产品与价值",
            "notes": "第二章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 6,
            "type": "example_detail",
            "title": "智研AI临床数据分析平台",
            "content_type": "example_detail",
            "kicker": "实例 · 解决方案",
            "lede": "一键自动完成数据清洗、标准化与质量核查，将数据锁库时间从6周缩短至10天",
            "context_block": "智研科技构建了一套专为临床试验设计的垂直领域大模型，训练语料包含超过2800万条脱敏临床记录、FDA申报文件和医学文献。与通用AI不同，我们的系统理解GCP（药物临床试验质量管理规范）和21 CFR Part 11合规要求。",
            "solution_block": "平台支持多源数据（EDC系统、CRO报表、实验室数据）自动接入，利用自然语言处理技术自动识别和映射不同数据标准（CDISC SDTM/ADaM）。AI同时进行实时数据质量监控，在数据录入阶段即发现异常值和逻辑冲突，将数据错误率降低85%。统计师可在统一界面完成数据审核、方案偏离判断和统计建模的全流程。",
            "metrics": [
                {"value": "70%", "label": "数据锁库时间缩短", "trend": "从42天降至12天"},
                {"value": "85%", "label": "数据错误率降低", "trend": "从8%降至1.2%"},
                {"value": "3.2x", "label": "统计师分析效率提升", "trend": "每周有效分析时间提升3.2倍"}
            ],
            "takeaway": "智研平台是唯一同时满足监管合规要求并集成垂直领域AI的临床数据分析解决方案。",
            "notes": "一句话说清楚解决方案，如何解决上述痛点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料，再填入真实内容：lede 一句话概括解决方案的核心价值；context_block 简要说明解决方案针对的痛点（1-2句话）；solution_block 详细展开解决方案的核心机制、技术路线和差异化优势（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结核心竞争力。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "市场分析",
            "subtitle": "规模与机会",
            "notes": "第三章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 8,
            "type": "stat_slide",
            "title": "市场规模",
            "content_type": "stat_slide",
            "stats": [
                {"value": "650亿美元", "label": "TAM 全球临床试验服务市场", "note": "Grand View Research 2025"},
                {"value": "180亿美元", "label": "SAM 制药企业数据服务市场", "note": "公司可服务细分市场"},
                {"value": "28亿美元", "label": "SOM AI驱动的数据分析市场", "note": "5年目标市场"},
                {"value": "26%", "label": "年复合增长率（CAGR）", "note": "AI临床服务赛道，2025-2030"}
            ],
            "notes": "展示 TAM/SAM/SOM 三层市场数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料，再填入真实内容：提供4个市场数据（TAM/SAM/SOM/CAGR），每个有 value（具体数字+单位）、label（说明）、note（数据来源）。禁止虚构数据。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "04",
            "title": "商业模式",
            "subtitle": "如何盈利",
            "notes": "第四章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 10,
            "type": "card_grid",
            "title": "四大核心优势",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "监管合规内置", "body": "系统内置FDA/EMA申报合规检查模块，数据可直接用于NDA申报，无需二次加工"},
                {"header": "垂直领域AI", "body": "2800万条临床记录训练的垂直大模型，理解GCP/CDISC专业术语和逻辑"},
                {"header": "多源数据融合", "body": "支持50+种EDC系统和CRO数据格式一键接入，打破数据孤岛"},
                {"header": "客户成功体系", "body": "配备专属临床数据科学家全程支持，平均7天完成客户系统集成"}
            ],
            "notes": "展示4个核心护城河/竞争优势",
            "filling_prompt": "必须填入真实内容：提供4个核心竞争优势，每个有 header（优势名称）和 body（一句话描述）。优势要具体、真实。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "运营数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 增长与验证",
            "kpis": [
                {"value": "18家", "label": "签约药企客户", "delta": "+80%", "baseline": "2025年初10家"},
                {"value": "420万", "label": "ARR（年度经常性收入）", "delta": "+135%", "baseline": "2025年初180万"},
                {"value": "98%", "label": "客户续费率", "delta": "+5pp", "baseline": "2025年93%"},
                {"value": "28家", "label": "在跑试验项目数", "delta": "+155%", "baseline": "2025年初11个"}
            ],
            "notes": "用数据证明产品市场契合度",
            "filling_prompt": "必须填入真实内容：提供4个运营指标，每个有 value、label、delta、baseline。数据必须真实，禁止虚构。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "竞争优势",
            "subtitle": "客户验证",
            "notes": "第五章分隔页",
            "filling_prompt": "固定结构，无需额外填充。"
        },
        {
            "index": 13,
            "type": "case_study",
            "title": "标杆客户案例：某头部创新药企",
            "content_type": "case_study",
            "kicker": "案例 · 肿瘤药物研发",
            "context": "该客户专注肿瘤创新药研发，拥有5条临床在研管线，其中2条进入III期试验。公司此前使用传统CRO模式，数据管理依赖人工核查，每项试验配备8名数据管理员，年均数据管理支出超过3000万元。",
            "problem": "II期试验数据锁库耗时52天，比预期延误18天，直接影响III期启动时间。同时，人工核查发现3.2%的数据需要返工核实，返工平均耗时3.5天/次，造成大量隐性成本。",
            "solution": "该客户引入智研平台后，将数据管理团队从8人缩减至3人（AI承担80%核查工作），数据锁库时间缩短至16天。AI实时质控系统在数据录入后2小时内即可发现异常值，将返工率从3.2%降至0.6%。III期试验提前4个月完成首例患者给药。",
            "results": [
                {"metric": "锁库时间", "value": "-69%", "comparison": "从52天降至16天"},
                {"metric": "人力成本", "value": "-62%", "comparison": "数据管理团队从8人减至3人"},
                {"metric": "返工率", "value": "-81%", "comparison": "从3.2%降至0.6%"},
                {"metric": "试验启动", "value": "提前4个月", "comparison": "III期试验提前4个月进入试验阶段"}
            ],
            "notes": "展示典型客户案例与量化效果",
            "filling_prompt": "必须填入真实内容：context 描述客户背景；problem 描述核心痛点；solution 描述解决方案和实施效果；results 提供4个量化指标（metric/value/comparison）。"
        },
        {
            "index": 14,
            "type": "card_grid",
            "title": "核心团队",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "刘思远 | CEO & 联合创始人", "body": "北京大学药学硕士，前恒瑞医药临床运营总监，12年新药研发经验，主导过3个NDA申报"},
                {"header": "赵文博 | CTO & 联合创始人", "body": "清华大学计算机博士，前字节跳动AI Lab算法负责人，在NeurIPS/ICML发表论文12篇"},
                {"header": "孙悦 | CPO 产品负责人", "body": "哥伦比亚大学生物统计硕士，前IQVIA高级统计师，主导过20+项临床试验数据分析"},
                {"header": "周健 | VP 商务负责人", "body": "中欧国际工商学院MBA，前Medidata中国区销售总监，8年生命科学软件销售经验"}
            ],
            "notes": "展示核心团队成员的背景和互补性",
            "filling_prompt": "必须填入真实内容：介绍4位核心团队成员，每人有 header（姓名+职位）和 body（教育背景、相关经历）。"
        },
        {
            "index": 15,
            "type": "process_flow",
            "title": "融资计划",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "kicker": "本轮融资",
            "subtitle": "A轮融资用途与里程碑规划",
            "steps": [
                {"num": "01", "title": "产品深化", "desc": "投入40%资金用于AI模型持续训练与合规功能升级"},
                {"num": "02", "title": "市场拓展", "desc": "投入30%资金扩展至美国/欧洲市场，对接TOP50药企"},
                {"num": "03", "title": "团队建设", "desc": "投入20%资金招募临床AI工程师与BD团队"},
                {"num": "04", "title": "生态合作", "desc": "投入10%资金与CRO、药企建立数据战略合作"}
            ],
            "notes": "展示融资轮次、金额与资金用途",
            "filling_prompt": "固定结构，已填入4个资金用途方向，每步有 num、title、desc。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "愿景",
            "kicker": "让AI加速治愈",
            "key_points": [
                "01  使命：降低新药研发成本30%，让有效治疗更早惠及患者",
                "02  目标：2028年服务全球TOP100药企，成为临床数据分析行业标准",
                "03  愿景：每一条临床数据，都值得被智能理解"
            ],
            "thank_you": "感谢聆听",
            "contact": "刘思远  |  CEO  |  siyuan.liu@zhiyantech.com  |  186-1234-5678",
            "notes": "结尾页，明确愿景和联系方式",
            "filling_prompt": "必须填入真实内容：key_points[0] 填入一句使命愿景；contact 填写真实联系方式。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "路演PPT要有说服力，数据是核心",
        "痛点→解决方案→市场→商业模式→竞争优势→数据→团队，逻辑严密",
        "数据要具体：用户数、增长率、市场规模等",
        "团队介绍要突出背景和相关性",
        "结尾要明确融资需求",
        "每页只讲一个核心观点",
        "PPT控制在10-15页，演讲10-15分钟",
        "准备Q&A，预判投资人可能的问题"
    ],
    "investor_qa_prep": {
        "common_questions": [
            "为什么是你们做这个？",
            "竞争对手是谁？你们有什么不同？",
            "如何获客？CAC是多少？",
            "什么时候能盈利？",
            "如果大厂入场怎么办？"
        ],
        "key_metrics_to_prepare": [
            "CAC（客户获取成本）",
            "LTV（客户生命周期价值）",
            "NPS（净推荐值）",
            "月环比增长率",
            "续费率/流失率"
        ]
    }
}
