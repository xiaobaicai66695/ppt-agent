TEMPLATE = {
    "name": "short-class-talk",
    "name_cn": "课堂短时分享",
    "description": "适合课堂 5-10 分钟短时分享、课题介绍等场景。精简高效，重点突出，快速传达。",
    "target_audience": "老师、同学",
    "typical_slides": 6,
    "typical_duration": "5-10分钟",
    "palette": "simple_gray",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "time_management": "5-10分钟，每页控制在1-2分钟",
        "key_principles": [
            "精简：只讲一个核心观点",
            "聚焦：聚焦听众最关心的内容",
            "生动：用故事或案例吸引注意力",
            "互动：适当设置提问，引发思考"
        ],
        "content_strategy": [
            "开场：30秒内切入主题",
            "展开：2-3分钟核心内容",
            "案例：1-2分钟案例说明",
            "收尾：30秒总结和互动"
        ]
    },
    "speaking_tips": {
        "opening": "用一个问题或现象开场，引发好奇心",
        "body": "只讲3个以内的要点，每点1分钟",
        "closing": "用一句话总结，预留Q&A"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "人工智能如何改变我们的未来",
            "subtitle": "从ChatGPT看AI技术发展趋势",
            "author": "张三",
            "date": "2025年3月10日",
            "notes": "标题页简洁，一目了然",
            "filling_prompt": "必须填入真实内容：title 为分享主题名称，subtitle 为副标题（如有），author 为分享人姓名，date 为分享日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "example_detail",
            "title": "背景介绍",
            "content_type": "example_detail",
            "kicker": "实例 · 背景",
            "lede": "ChatGPT在2个月内用户破亿，AI正在以前所未有的速度改变世界",
            "context_block": "2022年11月，OpenAI发布ChatGPT，2个月内用户突破1亿，成为史上增长最快的应用。2023年被业界称为'AI元年'，大模型技术持续突破，应用场景不断拓展。",
            "solution_block": "今天我想和大家探讨：AI技术发展到了什么程度？它将如何改变我们的生活和工作？我们应该如何看待和应对这场技术变革？",
            "metrics": [
                {"value": "1亿", "label": "ChatGPT 2个月用户数", "trend": "史上最快"},
                {"value": "100+", "label": "大模型数量", "trend": "持续增长"},
                {"value": "60%", "label": "500强已部署AI", "trend": "vs 2022年20%"}
            ],
            "takeaway": "启示：理解AI，是把握未来机遇的关键",
            "notes": "一句话介绍背景，快速切入主题",
            "filling_prompt": "必须填入真实内容：lede 一句话概括分享主题的核心价值；context_block 描述分享主题在当前环境/行业中的背景（1-2句话）；solution_block 具体说明为什么这个主题值得分享（2-3句话）；metrics_grid 提供3个背景数据指标，每个有 value、label、trend；takeaway 用一句话总结听众的收获。禁止保留花括号。"
        },
        {
            "index": 3,
            "type": "example_detail",
            "title": "核心内容",
            "content_type": "example_detail",
            "kicker": "实例 · 核心内容",
            "lede": "AI不会取代人类，但会用AI的人会取代不会用的人",
            "context_block": "面对AI的快速发展，有人恐慌，觉得自己会被取代；有人轻视，觉得AI不过如此。两种态度都不可取。",
            "solution_block": "正确的态度是：1）理解AI的能力边界——它擅长模式识别和生成，但缺乏真正的理解和创造力；2）掌握与AI协作的能力——学会提问、验证、整合；3）聚焦人类独特优势——批判性思维、情感沟通、复杂决策。未来的竞争力，在于人机协作的能力。",
            "metrics": [
                {"value": "80%", "label": "AI可完成重复性工作", "trend": "效率革命"},
                {"value": "不可替代", "label": "人类独特能力", "trend": "创造力+情感"},
                {"value": "必备", "label": "AI协作能力", "trend": "未来核心竞争力"}
            ],
            "takeaway": "启示：拥抱AI，学会与它协作，而非恐惧它",
            "notes": "展示3-4个核心要点",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的信息；context_block 说明为什么这个内容重要（1-2句话）；solution_block 详细展开核心内容的具体内涵（2-3句话）；metrics_grid 提供3个核心要点指标，每个有 value（要点名称）、label（解释说明）、trend（关联效果）；takeaway 用一句话总结实践意义。要点要精炼，每条不超过30字。禁止保留花括号。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "案例/应用",
            "content_type": "example_detail",
            "kicker": "实例 · 案例应用",
            "lede": "AI赋能各行各业的真实案例",
            "context_block": "AI不是遥不可及的技术，它已经深入到我们生活的方方面面。",
            "solution_block": "分享几个身边的AI应用案例：1）编程辅助——程序员用Copilot提效50%以上；2）内容创作——编辑用AI辅助写作，产量翻倍；3）数据分析——非技术人员用AI工具自主分析数据；4）学习助手——学生用AI辅导功课，查漏补缺。这些案例告诉我们：AI是工具，善用它可以大大提升效率。",
            "metrics": [
                {"value": "↑ 50%", "label": "编程效率提升", "trend": "vs 不用AI"},
                {"value": "↑ 100%", "label": "内容产量提升", "trend": "vs 传统方式"},
                {"value": "人人可用", "label": "AI普惠化", "trend": "门槛降低"}
            ],
            "takeaway": "启示：AI就在身边，从今天开始尝试使用它",
            "notes": "用一个具体案例或应用场景加深理解",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少1个URL），再填入真实内容：lede 一句话概括案例的核心成果；context_block 描述案例背景和应用场景（1-2句话）；solution_block 具体说明案例的具体做法和关键过程（2-3句话）；metrics_grid 提供3个成果指标，每个有 value（数字）、label（说明）、trend（对比基准）；takeaway 用一句话总结借鉴意义。references 列出 URL。禁止虚构数据；禁止保留花括号。"
        },
        {
            "index": 5,
            "type": "example_detail",
            "title": "要点回顾",
            "content_type": "example_detail",
            "kicker": "实例 · 要点回顾",
            "lede": "理解AI、善用AI、拥抱AI",
            "context_block": "今天的分享，我们探讨了AI的发展趋势和对我们生活的影响。",
            "solution_block": "核心要点回顾：1）AI技术正在快速发展，已经深入各行各业；2）面对AI，既不恐慌也不轻视，正确的态度是学习和协作；3）从今天开始，尝试在工作中使用AI工具，提升效率。记住：AI是工具，而工具的价值在于使用者的智慧。",
            "metrics": [
                {"value": "3个", "label": "核心要点数", "trend": "覆盖核心"},
                {"value": "3条", "label": "可行动建议", "trend": "可操作性"},
                {"value": "延伸阅读", "label": "推荐资源", "trend": "deepseek/ChatGPT"}
            ],
            "takeaway": "启示：今天就开始，用AI解决一个你的实际问题",
            "notes": "快速回顾核心要点，加深印象",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心收获；context_block 回顾分享中的关键信息和亮点（1-2句话）；solution_block 总结核心要点和实践建议（2-3句话）；metrics_grid 提供3个回顾指标，每个有 value、label、trend；takeaway 用一句话总结听众下一步可以做什么。禁止保留花括号。"
        },
        {
            "index": 6,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 理解AI：它是工具，不是威胁",
                "02 善用AI：学会与AI协作",
                "03 拥抱AI：从今天开始尝试"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁结尾，可加 Q&A 提示",
            "filling_prompt": "必须填入真实内容：key_points 提供 2-3 个要点（核心收获 + 行动建议）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "课堂分享要精简，每页只讲一个要点",
        "文字要少，多用图表和关键词",
        "时间控制在 5-10 分钟，6-8 页为宜",
        "留出时间回答问题",
        "注意与听众的眼神交流",
        "用故事或案例吸引注意力",
        "开场要抓人，结尾要有号召",
        "PPT设计简洁，留白充足"
    ],
    "time_allocation": {
        "opening": "30秒 - 1分钟",
        "core_content": "3 - 5分钟",
        "case_study": "1 - 2分钟",
        "summary": "30秒 - 1分钟",
        "qa": "根据时间灵活安排"
    },
    "quick_prep_tips": [
        "提前演练，控制时间",
        "准备几个可能被问到的问题",
        "带上PPT打印版作为备用",
        "开场前深呼吸，保持自信"
    ]
}
