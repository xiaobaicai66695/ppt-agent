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
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{公司/项目名称}",
            "subtitle": "{一句话价值主张}",
            "author": "{创始人/CEO}",
            "date": "{日期}",
            "filling_prompt": "必须填入真实内容：title 为公司或项目名称，subtitle 为一句有力的价值主张，author 为创始人姓名，date 为日期。禁止保留花括号。"
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
            "filling_prompt": "目录页为固定结构，无需额外填充。"
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
            "type": "content_slide",
            "title": "痛点",
            "content_type": "content_slide",
            "notes": "用数据和故事说明目标市场的痛点",
            "filling_prompt": "必须填入真实内容：提供3-4条具体痛点，每条要有数据支撑（如'XX行业每年因XX问题损失XX亿元'、'XX%用户因XX问题流失'）。禁止空泛描述。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "解决方案",
            "content_type": "content_slide",
            "notes": "一句话说清楚解决方案，如何解决上述痛点",
            "filling_prompt": "必须填入真实内容：一句话概括解决方案，并配合2-3条具体说明如何解决痛点。"
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
                {"value": "{TAM}", "label": "TAM 总可触达市场", "delta": "{增速}", "baseline": "{年份}"},
                {"value": "{SAM}", "label": "SAM 可服务市场", "delta": "{增速}", "baseline": "{年份}"},
                {"value": "{SOM}", "label": "SOM 目标市场", "delta": "{增速}", "baseline": "{年份}"},
                {"value": "{CAGR}", "label": "市场年复合增长率", "delta": "{增长}", "baseline": "{年份}"}
            ],
            "notes": "展示 TAM/SAM/SOM 三层市场数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个市场数据（TAM/SAM/SOM/CAGR），每个有 value（具体数字+单位）、label（说明）、delta（变化趋势）、baseline（数据年份或来源）。注明数据来源。references 列出 web_search 获取的 URL（至少2个）。"
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
            "notes": "如何赚钱，收入来源",
            "filling_prompt": "必须填入真实内容：说明具体收入模式（如'SaaS订阅制+交易抽成'），列出主要收入来源和定价策略。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "核心优势",
            "content_type": "card_grid",
            "layout_hint": "2x2",
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
            "notes": "用图文混排展示运营数据，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'运营数据'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为数据主题（不超过35字）；sub_header 为数据概述（不超过35字）；bullets 列出3-4条核心运营指标，每条不超过35字（如'MAU 50万'、'月增长率 30%'、'月收入 200万'），注明数据时间。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "增长策略",
            "content_type": "process_flow",
            "notes": "未来增长路径和策略",
            "filling_prompt": "必须填入真实内容：提供4-6个增长策略步骤，每步有名称和一句话描述（如'渠道拓展：入驻3个新平台'）。"
        },
        {
            "index": 14,
            "type": "content_slide",
            "title": "团队介绍",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "核心团队成员，背景和经历",
            "filling_prompt": "必须填入真实内容：介绍2-4位核心团队成员，每人有姓名、职位、相关背景（如'CEO：前XX公司技术VP，10年行业经验'）。"
        },
        {
            "index": 15,
            "type": "content_slide",
            "title": "融资计划",
            "content_type": "content_slide",
            "notes": "融资金额、估值、资金用途",
            "filling_prompt": "必须填入真实内容：说明融资轮次（如'本轮Pre-A'）、融资金额、估值、资金主要用途（2-3条）。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "愿景",
            "key_points": [
                "01 {一句话愿景}",
                "02 {联系方式}",
                "03 {期待交流}"
            ],
            "thank_you": "感谢聆听",
            "contact": "{邮箱} | {电话}",
            "notes": "结尾页，明确愿景和联系方式",
            "filling_prompt": "必须填入真实内容：key_points[0] 填入一句愿景；contact 填写真实联系方式（邮箱+电话）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "路演PPT要有说服力，数据是核心",
        "痛点→解决方案→市场→商业模式→竞争优势→数据→团队，逻辑严密",
        "数据要具体：用户数、增长率、市场规模等",
        "团队介绍要突出背景和相关性",
        "结尾要明确融资需求"
    ]
}
