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
            "title": "公司/项目名称",
            "subtitle": "一句有力的价值主张",
            "author": "创始人姓名 | CEO & 创始人",
            "date": "日期",
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
            "lede": "一句话概括核心挑战",
            "context_block": "描述目标市场的普遍困境（1-2句话）。",
            "solution_block": "具体说明这些困境导致的后果和对用户的影响（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "指标1", "trend": "变化趋势"},
                {"value": "数字", "label": "指标2", "trend": "变化趋势"},
                {"value": "数字", "label": "指标3", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结机会。",
            "notes": "用数据和故事说明目标市场的痛点",
            "filling_prompt": "必须填入真实内容（通过 web_search 获取权威数据，至少2个URL）：lede 一句话概括核心挑战；context_block 描述目标市场的普遍困境（1-2句话）；solution_block 具体说明这些困境导致的后果和对用户的影响（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结机会。禁止空泛描述。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "example_detail",
            "title": "解决方案",
            "content_type": "example_detail",
            "kicker": "实例 · 解决方案",
            "lede": "一句话概括解决方案的核心价值",
            "context_block": "简要说明解决方案针对的痛点（1-2句话）。",
            "solution_block": "详细展开解决方案的核心机制、技术路线和差异化优势（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "指标1", "trend": "变化趋势"},
                {"value": "数字", "label": "指标2", "trend": "变化趋势"},
                {"value": "数字", "label": "指标3", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结核心竞争力。",
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
                {"value": "数字+单位", "label": "TAM 总市场", "delta": "变化趋势", "baseline": "数据年份"},
                {"value": "数字+单位", "label": "SAM 可服务市场", "delta": "变化趋势", "baseline": "数据年份"},
                {"value": "数字+单位", "label": "SOM 目标市场", "delta": "变化趋势", "baseline": "数据年份"},
                {"value": "百分比", "label": "年复合增长率", "delta": "变化趋势", "baseline": "年份区间"}
            ],
            "notes": "展示 TAM/SAM/SOM 三层市场数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个市场数据（TAM/SAM/SOM/CAGR），每个有 value（具体数字+单位）、label（说明）、delta（变化趋势）、baseline（数据年份或来源）。注明数据来源。references 列出 URL。"
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
                "模式A": {
                    "description": "模式描述",
                    "pricing": "定价策略",
                    "target": "目标客户"
                },
                "模式B": {
                    "description": "模式描述",
                    "pricing": "定价策略",
                    "target": "目标客户"
                }
            },
            "notes": "如何赚钱，收入来源",
            "filling_prompt": "必须填入真实内容：说明具体收入模式，列出主要收入来源和定价策略。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "核心优势",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "优势名称", "body": "优势描述"},
                {"header": "优势名称", "body": "优势描述"},
                {"header": "优势名称", "body": "优势描述"},
                {"header": "优势名称", "body": "优势描述"}
            ],
            "notes": "4个核心护城河/竞争优势",
            "filling_prompt": "必须填入真实内容：提供4个核心竞争优势，每个有 header（优势名称）和 body（一句话描述）。优势要具体。"
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
            "header": "数据主题标题",
            "sub_header": "数据概述",
            "paragraph": "详细解读运营数据的含义、变化趋势和业务启示，用流畅的段落形式呈现，禁止罗列要点，必须包含具体指标数值。",
            "references": [
                "https://权威数据来源URL1",
                "https://权威数据来源URL2"
            ],
            "notes": "用图文混排展示运营数据，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为数据主题（不超过35字）；sub_header 为数据概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细解读运营数据的含义、变化趋势和业务启示，用流畅的段落形式呈现，禁止罗列要点，必须包含具体指标数值。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "增长策略",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "策略步骤", "desc": "步骤描述"},
                {"num": "02", "title": "策略步骤", "desc": "步骤描述"},
                {"num": "03", "title": "策略步骤", "desc": "步骤描述"},
                {"num": "04", "title": "策略步骤", "desc": "步骤描述"}
            ],
            "notes": "未来增长路径和策略",
            "filling_prompt": "必须填入真实内容：提供4-6个增长策略步骤，每步有名称和一句话描述。"
        },
        {
            "index": 14,
            "type": "content_slide",
            "title": "团队介绍",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "姓名 | 职位", "body": "背景描述"},
                {"header": "姓名 | 职位", "body": "背景描述"},
                {"header": "姓名 | 职位", "body": "背景描述"},
                {"header": "姓名 | 职位", "body": "背景描述"}
            ],
            "notes": "核心团队成员，背景和经历",
            "filling_prompt": "必须填入真实内容：介绍2-4位核心团队成员，每人有姓名、职位、相关背景。"
        },
        {
            "index": 15,
            "type": "content_slide",
            "title": "融资计划",
            "content_type": "content_slide",
            "round": "本轮融资轮次",
            "amount": "融资金额",
            "valuation": "投前估值",
            "use_of_funds": {
                "用途1": "占比和说明",
                "用途2": "占比和说明",
                "用途3": "占比和说明"
            },
            "notes": "融资金额、估值、资金用途",
            "filling_prompt": "必须填入真实内容：说明融资轮次、融资金额、估值、资金主要用途。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "愿景",
            "key_points": [
                "01 使命：愿景描述",
                "02 联系方式：真实联系方式",
                "03 合作诉求：诉求描述"
            ],
            "thank_you": "感谢聆听",
            "contact": "邮箱 | 电话 | 官网",
            "notes": "结尾页，明确愿景和联系方式",
            "filling_prompt": "必须填入真实内容：key_points[0] 填入一句愿景；contact 填写真实联系方式。禁止保留花括号。"
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
