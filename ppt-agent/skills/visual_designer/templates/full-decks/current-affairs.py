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
    "scene_guidance": {
        "sharing_principles": [
            "客观准确：引用权威来源，不传谣不信谣",
            "理性分析：多角度解读，避免片面",
            "数据支撑：用数据说话，有据可查",
            "观点鲜明：有自己的判断和见解"
        ],
        "audience_considerations": [
            "时政分享需要庄重严谨的风格",
            "避免情绪化表达，保持理性客观",
            "注意政策表述的准确性",
            "尊重不同观点，求同存异"
        ]
    },
    "content_structure": {
        "background": "热点背景：事件发生的宏观环境",
        "process": "事件经过：核心事实梳理",
        "reactions": "各方反应：多角度分析立场",
        "impact": "影响分析：多维度评估影响",
        "outlook": "趋势展望：未来发展研判"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "时政主题名称",
            "subtitle": "概括性副标题",
            "author": "演讲者姓名/单位",
            "date": "实际日期",
            "notes": "标题页庄重大气，体现时政严肃性",
            "filling_prompt": "必须填入真实内容：title 为本次分享的时政主题名称，subtitle 为概括性副标题，author 为演讲者姓名或单位，date 为实际日期。禁止保留花括号。"
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
            "header": "背景核心概述",
            "sub_header": "事件重要性说明",
            "paragraph": "详细描述热点事件的背景、发展过程和相关影响，用流畅的段落形式呈现，禁止罗列要点，必须包含具体时间、地点或数据。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2",
                "https://权威来源URL3"
            ],
            "notes": "右侧配新闻截图/地图/时间线图，左侧分析背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含官方机构来源），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景核心概述（不超过35字）；sub_header 为事件重要性说明（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述热点事件的背景、发展过程和相关影响，用流畅的段落形式呈现，禁止罗列要点，必须包含具体时间、地点或数据。references 逐条列出 URL 并标注来源机构名称。禁止空洞描述。"
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
            "milestones": [
                {"date": "时间", "event": "事件名称", "desc": "事件描述"},
                {"date": "时间", "event": "事件名称", "desc": "事件描述"},
                {"date": "时间", "event": "事件名称", "desc": "事件描述"},
                {"date": "时间", "event": "事件名称", "desc": "事件描述"},
                {"date": "时间", "event": "事件名称", "desc": "事件描述"}
            ],
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
            "kicker": "各方反应 · 多角度分析",
            "left_header": "官方解读",
            "left_sections": {
                "key_points": [
                    "要点1",
                    "要点2",
                    "要点3",
                    "要点4"
                ],
                "analysis": [
                    "分析维度1",
                    "分析维度2"
                ]
            },
            "right_header": "社会反响",
            "right_sections": {
                "key_points": [
                    "要点1",
                    "要点2",
                    "要点3",
                    "要点4"
                ],
                "data": [
                    "数据指标1",
                    "数据指标2",
                    "数据指标3"
                ]
            },
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "左右对比展示官方立场与社会反应的不同维度",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_header 为'官方解读'，left_sections.key_points 列出3-5条官方媒体或权威机构的表态要点，left_sections.analysis 列出2-3个分析维度；right_header 为'社会反响'，right_sections.key_points 列出3-5条不同群体的反应要点，right_sections.data 列出2-3个具体数据指标。禁止虚构观点。references 列出 URL。"
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
            "header": "最核心的一个影响维度",
            "sub_header": "影响程度说明",
            "paragraph": "详细分析热点事件的多维度影响和深层原因，用流畅的段落形式呈现，禁止罗列要点，必须包含具体数字或事实。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2",
                "https://权威来源URL3"
            ],
            "notes": "右侧配数据图表/新闻截图，左侧列举影响维度",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含官方机构和权威媒体），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最核心的一个影响维度（不超过35字）；sub_header 概括影响程度（不超过35字）；paragraph 为300-450字的自然语言段落，详细分析热点事件的多维度影响和深层原因，用流畅的段落形式呈现，禁止罗列要点，必须包含具体数字或事实。references 逐条列出 URL。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "关键数据指标",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 影响指标",
            "kpis": [
                {"value": "数字", "label": "指标名称", "delta": "趋势说明", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "趋势说明", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "趋势说明", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "趋势说明", "baseline": "对比基准"}
            ],
            "notes": "用数据量化事件影响",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个关键数据指标，每个有 value、label、delta、baseline。references 列出 URL。禁止虚构数据。"
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
                {"num": "01", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "02", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "03", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "04", "title": "趋势名称", "desc": "趋势描述（不超过30字）"}
            ],
            "notes": "4个未来趋势展望",
            "filling_prompt": "必须填入真实内容：提供4个该时政主题的未来发展趋势，每条有 title（趋势名称）和 desc（一句话描述，不超过30字）。禁止虚构趋势。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "总结与启示",
            "key_points": [
                "01 核心结论1",
                "02 核心结论2",
                "03 启示建议：描述"
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
        "语言严谨，避免情绪化表达",
        "PPT设计庄重大气，契合主题",
        "适当使用图表展示数据",
        "注意政策表述的准确性"
    ],
    "source_categorization": {
        "official": ["新华社", "人民日报", "中国政府网", "央视新闻"],
        "authoritative": ["求是", "瞭望", "半月谈", "学习强国"],
        "academic": ["人民日报理论版", "求是杂志", "中国社会科学院"]
    },
    "anti_rumors": [
        "不传播未经证实的消息",
        "不片面解读政策",
        "不煽动情绪",
        "及时纠正错误信息"
    ],
    "engagement_tips": [
        "设置互动问题，引导思考",
        "结合实际案例，帮助理解",
        "鼓励提问和讨论",
        "提供延伸阅读材料"
    ]
}
