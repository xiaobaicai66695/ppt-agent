TEMPLATE = {
    "name": "short-class-talk",
    "name_cn": "课堂短时分享",
    "description": "适合课堂5-10分钟短时分享、课题介绍等场景。精简高效，重点突出，快速传达。",
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
            "title": "分享主题名称",
            "subtitle": "副标题（如有）",
            "author": "分享人姓名",
            "date": "分享日期",
            "notes": "标题页简洁，一目了然",
            "filling_prompt": "必须填入真实内容：title 为分享主题名称，subtitle 为副标题（如有），author 为分享人姓名，date 为分享日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "example_detail",
            "title": "背景介绍",
            "content_type": "example_detail",
            "kicker": "实例 · 背景",
            "lede": "一句话概括分享主题的核心价值",
            "context_block": "描述分享主题在当前环境/行业中的背景（1-2句话）。",
            "solution_block": "具体说明为什么这个主题值得分享（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "背景数据1", "trend": "说明"},
                {"value": "数字", "label": "背景数据2", "trend": "说明"},
                {"value": "数字", "label": "背景数据3", "trend": "说明"}
            ],
            "takeaway": "一句话总结听众的收获。",
            "notes": "一句话介绍背景，快速切入主题",
            "filling_prompt": "必须填入真实内容：lede 一句话概括分享主题的核心价值；context_block 描述分享主题在当前环境/行业中的背景（1-2句话）；solution_block 具体说明为什么这个主题值得分享（2-3句话）；metrics_grid 提供3个背景数据指标，每个有 value、label、trend；takeaway 用一句话总结听众的收获。禁止保留花括号。"
        },
        {
            "index": 3,
            "type": "example_detail",
            "title": "核心内容",
            "content_type": "example_detail",
            "kicker": "实例 · 核心内容",
            "lede": "一句话概括最核心的信息",
            "context_block": "说明为什么这个内容重要（1-2句话）。",
            "solution_block": "详细展开核心内容的具体内涵（2-3句话）。",
            "metrics": [
                {"value": "要点1", "label": "核心要点1", "trend": "关联效果"},
                {"value": "要点2", "label": "核心要点2", "trend": "关联效果"},
                {"value": "要点3", "label": "核心要点3", "trend": "关联效果"}
            ],
            "takeaway": "一句话总结实践意义。",
            "notes": "展示3-4个核心要点",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的信息；context_block 说明为什么这个内容重要（1-2句话）；solution_block 详细展开核心内容的具体内涵（2-3句话）；metrics_grid 提供3个核心要点指标，每个有 value（要点名称）、label（解释说明）、trend（关联效果）；takeaway 用一句话总结实践意义。要点要精炼，每条不超过30字。禁止保留花括号。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "案例/应用",
            "content_type": "example_detail",
            "kicker": "实例 · 案例应用",
            "lede": "一句话概括案例的核心成果",
            "context_block": "描述案例背景和应用场景（1-2句话）。",
            "solution_block": "具体说明案例的具体做法和关键过程（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "成果指标1", "trend": "对比基准"},
                {"value": "数字", "label": "成果指标2", "trend": "对比基准"},
                {"value": "数字", "label": "成果指标3", "trend": "对比基准"}
            ],
            "takeaway": "一句话总结借鉴意义。",
            "notes": "用一个具体案例或应用场景加深理解",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少1个URL），再填入真实内容：lede 一句话概括案例的核心成果；context_block 描述案例背景和应用场景（1-2句话）；solution_block 具体说明案例的具体做法和关键过程（2-3句话）；metrics_grid 提供3个成果指标，每个有 value（数字）、label（说明）、trend（对比基准）；takeaway 用一句话总结借鉴意义。references 列出 URL。禁止虚构数据；禁止保留花括号。"
        },
        {
            "index": 5,
            "type": "example_detail",
            "title": "要点回顾",
            "content_type": "example_detail",
            "kicker": "实例 · 要点回顾",
            "lede": "一句话概括核心收获",
            "context_block": "回顾分享中的关键信息和亮点（1-2句话）。",
            "solution_block": "总结核心要点和实践建议（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "核心要点数", "trend": "覆盖核心"},
                {"value": "数字", "label": "可行动建议", "trend": "可操作性"},
                {"value": "数字", "label": "延伸阅读", "trend": "推荐资源"}
            ],
            "takeaway": "一句话总结听众下一步可以做什么。",
            "notes": "快速回顾核心要点，加深印象",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心收获；context_block 回顾分享中的关键信息和亮点（1-2句话）；solution_block 总结核心要点和实践建议（2-3句话）；metrics_grid 提供3个回顾指标，每个有 value、label、trend；takeaway 用一句话总结听众下一步可以做什么。禁止保留花括号。"
        },
        {
            "index": 6,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心收获1",
                "02 核心收获2",
                "03 行动建议：描述"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁结尾，可加Q&A提示",
            "filling_prompt": "必须填入真实内容：key_points 提供2-3个要点（核心收获+行动建议）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "课堂分享要精简，每页只讲一个要点",
        "文字要少，多用图表和关键词",
        "时间控制在5-10分钟，6-8页为宜",
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
