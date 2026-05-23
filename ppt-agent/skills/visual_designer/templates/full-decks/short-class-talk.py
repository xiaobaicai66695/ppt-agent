TEMPLATE = {
    "name": "short-class-talk",
    "name_cn": "课堂短时分享",
    "description": "适合课堂5-10分钟短时分享、课题介绍等场景。精简高效，重点突出，快速传达。",
    "target_audience": "老师、同学",
    "typical_slides": 6,
    "typical_duration": "5-10分钟",
    "palette": "education_blue",
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
            "title": "Goroutine与并发基础",
            "subtitle": "5分钟快速入门Go语言并发",
            "author": "李华",
            "date": "2026年5月23日",
            "notes": "标题页简洁，一目了然，标题有视觉冲击力",
            "filling_prompt": "必须填入真实内容：title为分享主题名称，subtitle为副标题，author为分享人姓名，date为分享日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "example_detail",
            "title": "为什么需要并发",
            "content_type": "example_detail",
            "kicker": "实例 · 背景",
            "lede": "单线程顺序执行导致现代多核CPU利用率不足30%",
            "context_block": "在多核处理器时代，顺序执行的程序只能使用一个核心，导致其他核心闲置。实际测试表明：一个处理10个独立任务的顺序程序，在8核CPU上利用率仅约12%。",
            "solution_block": "并发编程让多个任务同时执行，充分挖掘多核潜力。Go语言的Goroutine使得并发编程变得极其简单，只需一个go关键字即可启动并发任务。",
            "metrics": [
                {"value": "<30%", "label": "顺序程序CPU利用率", "trend": "多核严重闲置"},
                {"value": "10万+", "label": "Go单进程并发上限", "trend": "goroutine轻松支撑"},
                {"value": "1行代码", "label": "启动并发的语法", "trend": "go func()即可"}
            ],
            "takeaway": "并发是充分利用多核CPU、提升程序效率的必由之路。",
            "notes": "用具体数字说明并发的必要性，快速切入主题",
            "filling_prompt": "必须填入真实数据：lede一句话概括分享主题的核心价值；context_block描述背景（1-2句）；solution_block说明为什么这个主题值得分享（2-3句）；metrics_grid提供3个背景数据指标；takeaway一句话总结听众的收获。禁止保留花括号。"
        },
        {
            "index": 3,
            "type": "quote_slide",
            "title": "大师金句",
            "content_type": "quote_slide",
            "quote": "Concurrency is about dealing with lots of things at once. Parallelism is about doing lots of things at once.",
            "attribution": "Rob Pike，Go语言联合创始人",
            "kicker": "金句 · 并发哲学",
            "notes": "用一句引言过渡到核心概念，增加分享深度",
            "filling_prompt": "必须填入真实内容：quote为经典引言，attribution为出处。引言应准确反映并发与并行的区别。禁止虚构。"
        },
        {
            "index": 4,
            "type": "card_grid",
            "title": "快速开始并发编程",
            "content_type": "card_grid",
            "layout_hint": "1x3",
            "kicker": "要点 · 三步上手",
            "cards": [
                {"header": "第一步：启动协程", "body": "使用go关键字：go func(){ /* 并发执行的代码 */ }()", "icon": "01", "footer": "一行代码搞定并发"},
                {"header": "第二步：通信协调", "body": "用Channel在goroutine之间传递数据：ch := make(chan int)", "icon": "02", "footer": "线程安全 · 无需锁"},
                {"header": "第三步：等待完成", "body": "用WaitGroup等待所有协程结束：wg.Add() / wg.Done() / wg.Wait()", "icon": "03", "footer": "优雅同步"}
            ],
            "notes": "3个步骤的卡片网格，每步清晰简洁，配合简短代码片段",
            "filling_prompt": "必须填入真实内容：提供3个步骤卡片，每个卡片有header（步骤名称）、body（一句话说明+简短代码示例）、footer（效果标注）。步骤要精炼，代码示例要真实可运行。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "核心要点回顾",
            "content_type": "content_slide",
            "kicker": "要点 · 核心总结",
            "section_header": "记住这3点",
            "lede": "Go并发只需要掌握三个核心概念：goroutine、channel、waitgroup",
            "bullets": [
                "goroutine：用go func()启动轻量级并发，由Go运行时自动调度",
                "channel：用ch <- val和<-ch进行goroutine间的安全通信",
                "waitgroup：用wg.Add/Done/Wait管理协程的生命周期"
            ],
            "notes": "3个精炼要点，1分钟内讲完",
            "filling_prompt": "必须填入真实内容：lede一句话概括核心；bullets列出3个精炼要点，每条不超过30字，配合一句话说明。禁止保留花括号。"
        },
        {
            "index": 6,
            "type": "summary_slide",
            "title": "总结",
            "kicker": "总结",
            "key_points": [
                "01 并发是多核时代的必由之路，让程序充分利用硬件",
                "02 Go语言用goroutine让并发编程变得简单高效",
                "03 记住三剑客：go启动、channel通信、waitgroup等待"
            ],
            "thank_you": "感谢聆听！",
            "contact": "欢迎提问 | 课后答疑",
            "notes": "简洁结尾，可加Q&A提示",
            "filling_prompt": "必须填入真实内容：key_points提供3个要点（核心收获），thank_you为感谢语，contact为联系方式或Q&A提示。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "课堂分享要精简，每页只讲一个要点",
        "文字要少，多用图表和关键词",
        "时间控制在5-10分钟，6页为宜",
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
