TEMPLATE = {
    "name": "product-launch",
    "name_cn": "产品发布",
    "description": "适合新产品发布会、产品宣讲、客户演示等场景。强调价值主张、核心功能、差异化优势。",
    "target_audience": "客户、投资者、合作伙伴、媒体",
    "typical_slides": 14,
    "typical_duration": "15-20分钟",
    "palette": "warm_terracotta",
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
            "title": "{产品名称}",
            "subtitle": "{一句Slogan}",
            "author": "{公司名称}",
            "date": "{日期}",
            "notes": "开场标题页要有冲击力，Slogan要简短有力",
            "filling_prompt": "必须填入真实内容：title 为实际产品名称，subtitle 为一句有冲击力的Slogan，author 为公司名称，date 为日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  市场痛点",
                "02  解决方案",
                "03  核心功能",
                "04  产品优势",
                "05  客户案例"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "目录页为固定结构，无需额外填充。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "市场痛点",
            "subtitle": "当前面临的挑战",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "痛点分析",
            "content_type": "content_slide",
            "notes": "用数据说明当前市场的痛点和问题，唤起共鸣",
            "filling_prompt": "必须填入真实内容：提供3-4条具体痛点，每条要有数据支撑（如'企业平均每年在XX环节损失X万元'、'用户流失率高达XX%因为XX原因'）。禁止空泛描述。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "解决方案",
            "subtitle": "我们如何解决",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 6,
            "type": "content_slide",
            "title": "解决方案",
            "content_type": "content_slide",
            "notes": "一句话说清楚我们的解决方案是什么",
            "filling_prompt": "必须填入真实内容：一句话概括解决方案的核心价值主张，并配合2-3条支撑要点。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "核心功能",
            "subtitle": "产品能做什么",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "核心功能概览",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "4个核心功能，用图标+标题+一句话描述",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称）和 body（一句话描述功能价值）。功能名称要具体，如'智能推荐'、'实时监控'。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "04",
            "title": "产品优势",
            "subtitle": "为什么选择我们",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 10,
            "type": "kpi_dashboard",
            "title": "效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 产品价值",
            "kpis": [
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"}
            ],
            "notes": "展示4个核心效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心效果指标，每个 KPI 有 value（具体数值）、label（效果说明）、delta（变化趋势，如'↑ 30%'或'↓ 50%'）、baseline（对比基准，如'vs 传统方案'）。指标要具体，如'处理效率提升 3 倍'、'成本降低 40%'。references 列出 web_search 获取的 URL（至少2个）。禁止保留花括号；禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "05",
            "title": "客户案例",
            "subtitle": "真实客户的成功故事",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 12,
            "type": "image_text",
            "title": "客户案例：{客户名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示客户成功案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {行业} 替换为具体行业（如'金融'、'电商'、'医疗'）；title 中的 {客户名称} 替换为真实客户名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX公司数字化转型'，不超过35字）；sub_header 为合作项目名称（不超过35字）；bullets 列出3-4条应用效果或客户评价，每条不超过35字。references 列出 web_search 获取的 URL（至少2个）。禁止使用'某公司'等匿名实体；禁止虚构数据。"
        },
        {
            "index": 13,
            "type": "image_text",
            "title": "实施效果：{项目名称}",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "notes": "用图文混排展示实施效果，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'实施效果'；title 中的 {项目名称} 替换为具体项目名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为项目标题（不超过35字）；sub_header 为项目简介（不超过35字）；bullets 列出3-4条实施效果或量化成果，每条不超过35字。references 列出 web_search 获取的 URL（至少2个）。禁止虚构数据。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "联系我们",
            "key_points": [
                "01 {行动号召1}",
                "02 {行动号召2}"
            ],
            "thank_you": "感谢关注",
            "contact": "{联系方式}",
            "notes": "结尾页，明确行动号召和联系方式",
            "filling_prompt": "必须填入真实内容：key_points 提供2个行动号召（如'立即申请试用'、'预约产品演示'）；contact 填写真实联系方式（官网、邮箱、微信等）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "产品发布要有冲击力，视觉上突出产品和价值",
        "痛点→方案→功能→优势→案例，逻辑链条清晰",
        "数据说话：转化率、效率提升、成本降低等",
        "结尾要明确行动号召（CTA）"
    ]
}
