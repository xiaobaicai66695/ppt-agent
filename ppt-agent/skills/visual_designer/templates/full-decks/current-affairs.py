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
            "title": "中国数字经济发展报告",
            "subtitle": "2025年上半年运行态势与政策解读",
            "author": "国家信息中心",
            "date": "2025年5月",
            "notes": "标题页庄重大气，体现时政严肃性",
            "filling_prompt": "必须填入真实内容：title 为本次分享的主题名称，subtitle 为概括性副标题，author 为演讲者姓名或单位，date 为实际日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "报告目录",
            "kicker": "目录",
            "items": [
                "01  热点背景",
                "02  政策演进",
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
            "subtitle": "数字经济发展的宏观环境",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "数字经济背景分析",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景 · 宏观环境",
            "header": "数字经济成为经济增长核心引擎",
            "sub_header": "政策密集出台，产业加速布局",
            "paragraph": "2025年以来，中国数字经济迎来新一轮政策爆发期。国务院先后印发《数字经济发展"十四五"规划中期评估报告》和《数据要素市场化配置改革2025年行动方案》，明确了数据要素确权登记、公共数据授权运营、数据交易场所建设等关键任务的路线图和时间表。与此同时，工信部、发改委、网信办等部门也陆续出台人工智能、云计算、工业互联网等细分领域的专项支持政策。据中国信息通信研究院测算，2025年上半年数字经济核心产业增加值同比增长11.2%，增速显著高于同期GDP增速，数字经济对GDP增长的贡献率已超过45%，成为名副其实的宏观经济"压舱石"。",
            "references": [
                "国务院《数据要素市场化配置改革2025年行动方案》",
                "中国信息通信研究院《中国数字经济发展报告（2025年上半年）》",
                "国家统计局《2025年上半年国民经济运行情况》"
            ],
            "notes": "右侧配新闻截图/政策文件封面/数据图表，左侧分析背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含官方机构来源），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景核心概述（不超过35字）；sub_header 为事件重要性说明（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述热点事件的背景、发展过程和相关影响，用流畅的段落形式呈现，禁止罗列要点，必须包含具体时间、地点或数据。references 逐条列出 URL 并标注来源机构名称。禁止空洞描述。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "政策演进",
            "subtitle": "2025年上半年重要政策梳理",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "timeline",
            "title": "2025年数字经济政策演进",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "milestones": [
                {"date": "2025年1月", "event": "数据二十条正式落地", "desc": "数据资产确权、估值、交易等基础制度正式实施"},
                {"date": "2025年2月", "event": "数字中国建设峰会召开", "desc": "大会发布数字中国发展报告，明确年度重点任务"},
                {"date": "2025年3月", "event": "AI产业专项政策出台", "desc": "工信部发布人工智能产业创新发展三年行动计划"},
                {"date": "2025年4月", "event": "公共数据授权运营试点扩大", "desc": "新增12个城市开展公共数据授权运营试点"},
                {"date": "2025年5月", "event": "数字人民币跨境支付新规", "desc": "央行扩大数字人民币跨境支付试点范围至20个国家和地区"}
            ],
            "notes": "时间轴展示政策演进过程",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4-5个关键时间节点，每个节点有 date、event（一句话事件名称）、desc（一句话描述）。禁止虚构时间节点。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "各方反应",
            "subtitle": "政府、企业、学术界多维视角",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "two_column",
            "title": "各方观点与立场",
            "content_type": "two_column",
            "kicker": "各方反应 · 多角度分析",
            "left_header": "官方政策导向",
            "left_sections": {
                "key_points": [
                    "坚持自立自强，突破核心算法、芯片等关键技术瓶颈",
                    "推动数字经济与实体经济深度融合，赋能传统产业升级",
                    "加快数据要素市场化改革，构建全国统一数据市场",
                    "强化数字治理能力，完善数字经济监管体系"
                ],
                "analysis": [
                    "政策重点从基础设施建设转向应用创新与安全并重",
                    "区域协调发展战略中数字经济被赋予核心定位"
                ]
            },
            "right_header": "产业界响应",
            "right_sections": {
                "key_points": [
                    "互联网巨头加速AI大模型商业化落地，应用场景持续拓宽",
                    "传统制造业积极拥抱数字化转型，工业软件需求激增",
                    "云计算市场竞争加剧，价格战向差异化服务竞争转型",
                    "新能源与数字技术深度融合，智能网联汽车产业爆发"
                ],
                "data": [
                    "2025年上半年AI领域融资规模突破1200亿元",
                    "工业互联网平台连接设备数超过1.2亿台",
                    "智能网联汽车销量同比增长87%"
                ]
            },
            "references": [
                "国务院常务会议纪要（2025年3月）",
                "工信部《2025年上半年电子信息制造业运行情况》"
            ],
            "notes": "左右对比展示官方立场与产业响应的不同维度",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_header 为'官方政策导向'，left_sections.key_points 列出3-5条官方媒体或权威机构的表态要点，left_sections.analysis 列出2-3个分析维度；right_header 为'产业界响应'，right_sections.key_points 列出3-5条不同群体的反应要点，right_sections.data 列出2-3个具体数据指标。禁止虚构观点。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "04",
            "title": "影响分析",
            "subtitle": "数字经济对各领域的深度影响",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "case_study",
            "title": "数字经济多维影响",
            "content_type": "case_study",
            "kicker": "影响分析 · 典型案例",
            "cases": [
                {
                    "title": "制造业：智能工厂降本增效",
                    "desc": "海尔卡奥斯工业互联网平台接入企业超过100万家，平均生产效率提升23%，单位能耗下降18%。数字化改造使订单交付周期从平均15天缩短至8天，库存周转率提升35%。"
                },
                {
                    "title": "服务业：平台经济创造新就业",
                    "desc": "美团、饿了么等本地生活平台直接创造灵活就业岗位超过3000万个，平台注册骑手平均月收入达到6200元。数字零售带动快递业务量同比增长22%，日均处理包裹量突破5亿件。"
                },
                {
                    "title": "治理：数字政府提升效能",
                    "desc": "浙江省"浙里办"政务APP注册用户突破8000万，3000余项政务服务实现"一网通办"，平均办事时间从5.8天压缩至1.2天，群众满意度达到94.7%。数字技术赋能基层治理，网格员事件响应时间缩短60%。"
                }
            ],
            "notes": "用3个典型案例展示数字经济的广泛影响",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供3个典型案例，每个案例有 title（案例标题）和 desc（150-200字案例描述，包含具体数据）。案例应涵盖不同领域。禁止虚构案例。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "关键数据指标",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 影响指标",
            "kpis": [
                {"value": "56万亿元", "label": "数字经济规模", "delta": "占GDP比重41.5%", "baseline": "2024年全年50万亿元"},
                {"value": "6.5万亿元", "label": "数字贸易额", "delta": "同比增长18%", "baseline": "跨境电商占比持续提升"},
                {"value": "420万个", "label": "5G基站总数", "delta": "覆盖全部地级市", "baseline": "每万人18个基站"},
                {"value": "8000亿元", "label": "云计算市场规模", "delta": "同比增长27%", "baseline": "公有云占比超65%"}
            ],
            "notes": "用数据量化数字经济影响力",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个关键数据指标，每个有 value、label、delta、baseline。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "趋势展望",
            "subtitle": "数字经济未来发展方向研判",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "process_flow",
            "title": "未来趋势研判",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "kicker": "趋势展望",
            "steps": [
                {"num": "01", "title": "AI与产业深度融合", "desc": "大模型从通用走向行业垂直应用，工业、医疗、金融等领域AI渗透率将超过50%"},
                {"num": "02", "title": "数据要素市场化", "desc": "数据资产确权登记制度全面落地，数据交易所年交易规模有望突破万亿元"},
                {"num": "03", "title": "数字人民币全球化", "desc": "数字人民币跨境支付试点扩大，有望成为国际贸易结算的重要补充渠道"},
                {"num": "04", "title": "数字基础设施升级", "desc": "6G标准制定启动，卫星互联网与地面网络融合，万物互联时代加速到来"}
            ],
            "notes": "4个未来趋势展望",
            "filling_prompt": "必须填入真实内容：提供4个该时政主题的未来发展趋势，每条有 title（趋势名称）和 desc（一句话描述，不超过40字）。趋势必须有理有据，结合当前政策走向和产业动态。禁止虚构趋势。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "总结与启示",
            "key_points": [
                "01 数字经济已成为GDP增长主动力，2025年上半年规模达56万亿元，占GDP比重突破41%，对经济增长贡献率超过45%",
                "02 未来5年将迎来新一轮政策红利期，AI大模型应用、数据要素市场化、数字基础设施升级是三大核心赛道",
                "03 中小企业数字化转型是最大蓝海，政府需持续完善公共服务平台，降低中小企业上云用数赋智的门槛"
            ],
            "thank_you": "感谢聆听",
            "notes": "总结核心观点，提出启示",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心结论2条+启示建议1条），每条内容具体充实。结尾致谢语可保持固定。禁止保留花括号。"
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
