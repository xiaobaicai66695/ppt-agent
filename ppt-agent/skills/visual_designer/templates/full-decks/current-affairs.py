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
            "title": "2025年两会政策解读",
            "subtitle": "聚焦新质生产力与高质量发展",
            "author": "党委宣传部",
            "date": "2025年3月8日",
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
            "header": "2025年两会：承上启下的关键年份",
            "sub_header": "'十四五'规划收官与'十五五'规划谋篇之年",
            "paragraph": "2025年全国两会是在全面贯彻党的二十大和二十届二中、三中全会精神关键时刻召开的重要会议。今年是'十四五'规划收官之年，也是'十五五'规划谋篇之年。两会聚焦发展新质生产力、推动高质量发展、深化改革开放、保障和改善民生等重大议题，传递出中国经济社会发展的新信号、新动向。GDP增长目标、CPI预期目标、就业目标等关键指标的设定，体现了稳中求进的工作总基调。",
            "references": [
                "https://www.xinhuanet.com/",
                "https://www.gov.cn/",
                "https://www.ccps.gov.cn/"
            ],
            "notes": "右侧配新闻截图/地图/时间线图，左侧分析背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含官方机构来源），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景核心概述（不超过35字）；sub_header 为事件重要性说明（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述热点事件的背景、发展过程和相关影响，用流畅的段落形式呈现，禁止罗列要点，必须包含具体时间、地点或数据。references 逐条列出 URL 并标注来源机构名称。禁止空洞描述，必须用具体事实填充。"
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
                {"date": "2025.3.4", "event": "全国政协开幕", "desc": "全国政协十四届三次会议在北京召开"},
                {"date": "2025.3.5", "event": "人大会议开幕", "desc": "国务院总理作政府工作报告"},
                {"date": "2025.3.5", "event": "GDP目标发布", "desc": "GDP增长目标定为5%左右"},
                {"date": "2025.3.10", "event": "重要法案通过", "desc": "审议通过多部重要法律草案"},
                {"date": "2025.3.11", "event": "闭幕", "desc": "两会完成各项议程，胜利闭幕"}
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
                    "发展质量：GDP目标5%左右，体现稳中求进工作总基调",
                    "改革力度：重点领域改革深化，民营经济发展环境持续改善",
                    "民生导向：就业、收入、医疗等领域部署明确、措施务实",
                    "开放姿态：制度型开放稳步推进，高水平对外开放持续深化"
                ],
                "analysis": [
                    "权威性：官方媒体一致传递积极信号，提振发展信心",
                    "一致性：各部委政策解读口径统一，政策导向明确连贯"
                ]
            },
            "right_header": "社会反响",
            "right_sections": {
                "key_points": [
                    "企业界：新质生产力布局受关注，减税降费政策利好持续释放",
                    "科技圈：大模型、人工智能等领域创业者对政策支持充满期待",
                    "基层干部：民生关切得到积极回应，工作方向更加明确",
                    "普通民众：就业形势、收入增长、社会保障等问题关注度高"
                ],
                "data": [
                    "企业信心指数：环比回升3.2个百分点",
                    "民生领域财政支出：同比增长6.1%",
                    "资本市场：两会期间A股整体走势平稳"
                ]
            },
            "references": [
                "https://www.xinhuanet.com/",
                "https://www.gov.cn/"
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
            "header": "新质生产力成为发展新引擎",
            "sub_header": "科技创新和产业创新深度融合",
            "paragraph": "两会关于新质生产力的部署，对经济社会发展将产生深远影响。首先，科技创新的战略地位更加突出，科技创新引领现代化产业体系建设成为重点任务。其次，产业升级方向更加明确，传统产业的数字化智能化改造将加速推进。再次，民营经济发展环境持续改善，各种所有制企业将获得更加公平的发展机会。这些政策部署为经济社会发展指明了方向，也为企业转型发展提供了重要指引。",
            "references": [
                "https://www.xinhuanet.com/",
                "https://www.forbeschina.com/",
                "https://www.caict.ac.cn/"
            ],
            "notes": "右侧配数据图表/新闻截图，左侧列举影响维度",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含官方机构和权威媒体），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最核心的一个影响维度（不超过35字）；sub_header 概括影响程度（不超过35字）；paragraph 为300-450字的自然语言段落，详细分析热点事件的多维度影响和深层原因，用流畅的段落形式呈现，禁止罗列要点，必须包含具体数字或事实。references 逐条列出 URL（标记来源机构名称）。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "关键数据指标",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 影响指标",
            "kpis": [
                {"value": "5%左右", "label": "GDP增长目标", "delta": "稳中求进", "baseline": "2024年5%"},
                {"value": "约1200万", "label": "城镇新增就业目标", "delta": "大于2024年", "baseline": "2024年1200万+"},
                {"value": "3%", "label": "CPI预期目标", "delta": "温和通胀", "baseline": "2024年3%"},
                {"value": "超2万亿", "label": "超长期特别国债", "delta": "积极财政", "baseline": "持续发行"}
            ],
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
                {"num": "01", "title": "科技创新", "desc": "新质生产力引领高质量发展"},
                {"num": "02", "title": "产业升级", "desc": "传统产业数字化智能化转型加速"},
                {"num": "03", "title": "民生改善", "desc": "就业、收入、医疗等领域持续发力"},
                {"num": "04", "title": "改革开放", "desc": "重点领域改革深化，开放水平提升"}
            ],
            "notes": "4个未来趋势展望",
            "filling_prompt": "必须填入真实内容：提供4个该时政主题的未来发展趋势，每条有 title（趋势名称）和 desc（一句话描述，不超过30字）。禁止虚构趋势。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "总结与启示",
            "key_points": [
                "01 两会聚焦新质生产力，高质量发展成为主旋律",
                "02 政策导向明确，科技创新和产业升级是重点",
                "03 坚定信心，立足本职，为发展贡献力量"
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
