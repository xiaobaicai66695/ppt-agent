TEMPLATE = {
    "name": "current-affairs",
    "name_cn": "时政分享",
    "description": "适合时政热点分析、政策解读、国际形势分析等场景。稳重专业，信息密集，数据支撑强。",
    "target_audience": "政府机关、企事业单位、党团组织、关心时政的公众",
    "typical_slides": 14,
    "typical_duration": "15-20分钟",
    "palette": "government_red",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{时政主题名称}",
            "subtitle": "{副标题}",
            "author": "{演讲者}",
            "date": "{日期}",
            "notes": "标题页庄重大气，体现时政严肃性",
            "filling_prompt": "必须填入真实内容：title 为本次分享的时政主题名称（如'2025年两会政策解读'），subtitle 为概括性副标题，author 为演讲者姓名或单位，date 为实际日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  热点背景",
                "02  事件经过",
                "03  各方反应",
                "04  影响分析",
                "05  趋势展望"
            ],
            "notes": "让观众快速了解分享框架",
            "filling_prompt": "目录页为固定结构，根据实际主题调整章节名称。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "热点背景",
            "subtitle": "事件发生的宏观环境",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "热点背景分析",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景 · 宏观环境",
            "notes": "右侧配新闻截图/地图/时间线图，左侧分析背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含官方机构来源），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景核心概述（不超过35字）；sub_header 为事件重要性说明（不超过35字）；bullets 列出3-4条背景要点，每条不超过35字且包含具体时间、地点或数据。references 逐条列出 URL 并标注来源机构名称。禁止空洞描述，必须用具体事实填充。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "事件经过",
            "subtitle": "核心事实梳理",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "timeline",
            "title": "事件发展脉络",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "notes": "时间轴展示事件发展过程",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4-5个关键时间节点，每个节点有时间+事件名称+一句话描述。禁止虚构时间节点。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "各方反应",
            "subtitle": "多角度分析各方立场",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "two_column",
            "title": "各方观点与立场",
            "content_type": "two_column",
            "notes": "对比展示不同立场的观点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_header 为'官方立场'，left_bullets 列出3-5条官方表态要点；right_header 为'民间/国际反应'，right_bullets 列出3-5条反应要点。禁止虚构观点。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "04",
            "title": "影响分析",
            "subtitle": "多维度评估事件影响",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "多维影响分析",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "影响 · 深度分析",
            "notes": "右侧配数据图表/新闻截图，左侧列举影响维度",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含官方机构和权威媒体），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最核心的一个影响维度（不超过35字）；sub_header 概括影响程度（不超过35字）；bullets 列出3-4条具体影响，每条不超过35字且包含具体数字或事实。references 逐条列出 URL（标记来源机构名称）。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "关键数据指标",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 影响指标",
            "notes": "用数据量化事件影响",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个关键数据指标（如相关政策数量、涉及金额、影响人数、市场变化等），每个有 value、label、delta、baseline。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "趋势展望",
            "subtitle": "未来发展研判",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "未来趋势研判",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "{趋势1}", "desc": "{一句话描述}"},
                {"num": "02", "title": "{趋势2}", "desc": "{一句话描述}"},
                {"num": "03", "title": "{趋势3}", "desc": "{一句话描述}"},
                {"num": "04", "title": "{趋势4}", "desc": "{一句话描述}"}
            ],
            "notes": "4个未来趋势展望",
            "filling_prompt": "必须填入真实内容：提供4个该时政主题的未来发展趋势，每条有 title（趋势名称）和 desc（一句话描述，不超过30字）。禁止虚构趋势。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "总结与启示",
            "key_points": [
                "01 {核心结论1}",
                "02 {核心结论2}",
                "03 {启示与建议}"
            ],
            "thank_you": "感谢聆听",
            "notes": "总结核心观点，提出启示",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心结论2条+启示建议1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "时政分析要客观，数据说话",
        "引用权威来源，注明出处",
        "多维度分析，避免片面",
        "趋势研判要有理有据",
        "语言严谨，避免情绪化表达"
    ]
}
