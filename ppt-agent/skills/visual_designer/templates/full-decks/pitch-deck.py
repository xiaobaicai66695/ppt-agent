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
            "title": "智医问诊",
            "subtitle": "AI赋能基层医疗，让每个人都能享受优质医疗服务",
            "author": "张磊 | CEO & 创始人",
            "date": "2025年3月",
            "filling_prompt": "必须填入真实内容：title 为公司或项目名称，subtitle 为一句有力的价值主张，author 为创始人姓名，date 为日期。禁止保留花括号。",
            "visual_suggestions": "简洁有力的Logo，配合一句核心Slogan"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  痛点与机会",
                "02  解决方案",
                "03  市场分析",
                "04  商业模式",
                "05  竞争优势",
                "06  运营数据",
                "07  增长策略",
                "08  团队融资"
            ],
            "notes": "让投资人快速了解报告结构",
            "filling_prompt": "目录页为固定结构，无需额外填充。",
            "timing_hint": "快速翻过，约10秒"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "痛点与机会",
            "subtitle": "为什么现在",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "痛点",
            "content_type": "example_detail",
            "kicker": "实例 · 市场痛点",
            "lede": "中国基层医疗资源严重不足，患者看病难、看病贵问题亟待解决",
            "context_block": "中国有9亿农村和基层人口，但只占医疗资源的30%。基层医疗机构面临'缺医少药'困境：全科医生严重匮乏，平均每千人仅0.9名医生（远低于发达国家的3名）；基层误诊率高达40%，患者不得不辗转到大医院就诊。",
            "solution_block": "这些问题导致的后果是：1）患者看病成本大幅增加，平均就医成本是基层的5-8倍；2）大医院人满为患，医生超负荷工作；3）错过最佳诊疗时机，小病拖成大病。我们看到了巨大的改善空间和商业机会。",
            "metrics": [
                {"value": "9亿", "label": "基层服务人口", "trend": "占全国65%"},
                {"value": "40%", "label": "基层误诊率", "trend": "vs 大医院10%"},
                {"value": "5-8倍", "label": "就医成本差距", "trend": "亟需改善"}
            ],
            "takeaway": "启示：AI技术成熟+政策支持+市场需求三重叠加，窗口期已到",
            "notes": "用数据和故事说明目标市场的痛点",
            "filling_prompt": "必须填入真实内容（通过 web_search 获取权威数据，至少2个URL）：lede 一句话概括核心挑战；context_block 描述目标市场的普遍困境（1-2句话）；solution_block 具体说明这些困境导致的后果和对用户的影响（2-3句话）；metrics_grid 提供3个量化指标（如'XX行业每年因XX问题损失XX亿元'、'XX%用户因XX原因流失'），每个有 value（数字）、label（说明）、trend（变化趋势）；takeaway 用一句话总结机会。禁止空泛描述。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "example_detail",
            "title": "解决方案",
            "content_type": "example_detail",
            "kicker": "实例 · 解决方案",
            "lede": "AI辅助诊断系统，让基层医生也能拥有'专家级'诊疗能力",
            "context_block": "针对基层医疗的核心痛点——医生诊疗能力不足，我们开发了一套AI辅助诊断系统。",
            "solution_block": "我们的解决方案包括：1）AI问诊助手：模拟专家问诊流程，智能引导患者描述症状，生成结构化病历；2）智能辅诊：基于5000万+临床病例训练的AI模型，提供诊断建议和鉴别诊断；3）用药助手：智能审核处方，规避用药风险；4）转诊建议：判断是否需要向上转诊。系统已获国家二类医疗器械注册证。",
            "metrics": [
                {"value": "95%+", "label": "常见病诊断准确率", "trend": "vs 基层医生70%"},
                {"value": "60%", "label": "降低误诊率", "trend": "显著提升基层能力"},
                {"value": "3分钟", "label": "单次问诊时间", "trend": "效率提升5倍"}
            ],
            "takeaway": "启示：技术赋能+合规资质=商业化壁垒",
            "notes": "一句话说清楚解决方案，如何解决上述痛点",
            "filling_prompt": "必须填入真实内容：lede 一句话概括解决方案的核心价值；context_block 简要说明解决方案针对的痛点（1-2句话）；solution_block 详细展开解决方案的核心机制、技术路线和差异化优势（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结核心竞争力。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "市场分析",
            "subtitle": "规模与机会",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 7,
            "type": "kpi_dashboard",
            "title": "市场规模",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 市场空间",
            "kpis": [
                {"value": "1.2万亿", "label": "TAM 医疗AI总市场", "delta": "↑ 25% CAGR", "baseline": "2025年预测"},
                {"value": "800亿", "label": "SAM 基层医疗AI", "delta": "↑ 30% CAGR", "baseline": "2025年预测"},
                {"value": "200亿", "label": "SOM 辅诊细分市场", "delta": "↑ 35% CAGR", "baseline": "2025年预测"},
                {"value": "28%", "label": "年复合增长率", "delta": "↑ 持续增长", "baseline": "2024-2028"}
            ],
            "notes": "展示 TAM/SAM/SOM 三层市场数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个市场数据（TAM/SAM/SOM/CAGR），每个有 value（具体数字+单位）、label（说明）、delta（变化趋势）、baseline（数据年份或来源）。注明数据来源。references 列出 web_search 获取的 URL（至少2个）。",
            "data_source": "数据来源：国家卫健委、艾瑞咨询、国际医疗AI报告"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "商业模式",
            "subtitle": "如何盈利",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 9,
            "type": "content_slide",
            "title": "商业模式",
            "content_type": "content_slide",
            "revenue_model": {
                "B2B_SaaS": {
                    "description": "面向基层医疗机构，按年收取SaaS服务费",
                    "pricing": "5000-50000元/年/机构",
                    "target": "基层卫生院、社区卫生服务中心"
                },
                "B2G": {
                    "description": "参与政府AI医疗采购项目",
                    "pricing": "单项目50-500万",
                    "target": "卫健委、医保局"
                },
                "B2C": {
                    "description": "面向个人的健康管理订阅服务",
                    "pricing": "99元/月",
                    "target": "C端慢病患者"
                }
            },
            "notes": "如何赚钱，收入来源",
            "filling_prompt": "必须填入真实内容：说明具体收入模式（如'SaaS订阅制+交易抽成'），列出主要收入来源和定价策略。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "核心优势",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "专有数据集", "body": "5000万+脱敏临床数据，覆盖1000+病种"},
                {"header": "二类医疗器械证", "body": "国内首批获证的AI辅诊系统"},
                {"header": "三甲专家背书", "body": "协和、301等顶级专家参与研发"},
                {"header": "基层实战验证", "body": "已覆盖500+基层机构，1年+实际应用"}
            ],
            "notes": "4个核心护城河/竞争优势",
            "filling_prompt": "必须填入真实内容：提供4个核心竞争优势，每个有 header（优势名称）和 body（一句话描述）。优势要具体，如'自研核心算法'、'独家数据资产'。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "运营数据",
            "subtitle": "增长与验证",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 12,
            "type": "image_text",
            "title": "运营数据",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "运营数据",
            "header": "商业化验证成果",
            "sub_header": "ARR突破2000万，净推荐值NPS 68",
            "paragraph": "过去18个月，我们完成了产品打磨和商业化验证。核心成果包括：1）服务基层医疗机构500+，月均活跃机构320+；2）累计辅助诊断超过500万次，日均问诊量2万+；3）客户续费率达92%，净推荐值NPS 68；4）ARR突破2000万元，实现盈亏平衡。在河南、湖北试点区域，已进入卫健委推荐目录。",
            "references": [
                "https://www.nhc.gov.cn/",
                "https://www.iresearch.com.cn/"
            ],
            "notes": "用图文混排展示运营数据，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'运营数据'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为数据主题（不超过35字）；sub_header 为数据概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细解读运营数据的含义、变化趋势和业务启示，用流畅的段落形式呈现，禁止罗列要点，必须包含具体指标数值。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "增长策略",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "深耕试点", "desc": "河南、湖北区域全覆盖"},
                {"num": "02", "title": "复制扩张", "desc": "拓展至5省50市"},
                {"num": "03", "title": "产品矩阵", "desc": "慢病管理+用药安全"},
                {"num": "04", "title": "生态合作", "desc": "牵手药企、险企"},
                {"num": "05", "title": "数据变现", "desc": "真实世界研究"}
            ],
            "notes": "未来增长路径和策略",
            "filling_prompt": "必须填入真实内容：提供4-6个增长策略步骤，每步有名称和一句话描述（如'渠道拓展：入驻3个新平台'）。"
        },
        {
            "index": 14,
            "type": "content_slide",
            "title": "团队介绍",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "张磊 | CEO", "body": "前腾讯医疗产品总监，15年互联网经验，连续创业者"},
                {"header": "王华 | CTO", "body": "前微软亚洲研究院高级研究员，AI博士，发表顶会论文20+篇"},
                {"header": "李明 | CMO", "body": "前阿里健康运营总监，10年医疗行业经验"},
                {"header": "陈静 | COO", "body": "前平安好医生VP，丰富的基层医疗资源"}
            ],
            "notes": "核心团队成员，背景和经历",
            "filling_prompt": "必须填入真实内容：介绍2-4位核心团队成员，每人有姓名、职位、相关背景（如'CEO：前XX公司技术VP，10年行业经验'）。"
        },
        {
            "index": 15,
            "type": "content_slide",
            "title": "融资计划",
            "content_type": "content_slide",
            "round": "本轮Pre-A",
            "amount": "融资3000万元",
            "valuation": "投前估值1.5亿",
            "use_of_funds": {
                "研发": "40%（产品迭代+算法优化）",
                "市场": "35%（区域扩张+渠道建设）",
                "运营": "15%（团队扩充+运营优化）",
                "合规": "10%（资质申请+合规建设）"
            },
            "notes": "融资金额、估值、资金用途",
            "filling_prompt": "必须填入真实内容：说明融资轮次（如'本轮Pre-A'）、融资金额、估值、资金主要用途（2-3条）。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "愿景",
            "key_points": [
                "01 使命：让8亿基层人民享受优质医疗服务",
                "02 联系方式：hello@zhiyi.ai | 官网：www.zhiyi.ai",
                "03 期待与您合作，共同改变基层医疗现状"
            ],
            "thank_you": "感谢聆听",
            "contact": "张磊 138-xxxx-xxxx | zhang@zhiliyi.ai",
            "notes": "结尾页，明确愿景和联系方式",
            "filling_prompt": "必须填入真实内容：key_points[0] 填入一句愿景；contact 填写真实联系方式（邮箱+电话）。禁止保留花括号。"
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
            "CAC（用户获取成本）",
            "LTV（用户生命周期价值）",
            "NPS（净推荐值）",
            "月环比增长率",
            "续费率/流失率"
        ]
    }
}
