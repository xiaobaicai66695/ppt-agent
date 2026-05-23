TEMPLATE = {
    "name": "tech-sharing",
    "name_cn": "技术分享",
    "description": "适合内部技术分享、技术培训、架构讲解等场景。结构清晰，有章节划分，注重内容深度。",
    "target_audience": "工程师、技术管理者、技术爱好者",
    "typical_slides": 18,
    "typical_duration": "30-45分钟",
    "palette": "ocean_soft",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "opening_hook": "用一个实际的技术问题或挑战开场，引起工程师共鸣",
        "key_moments": [
            "问题分析时：展示具体的问题场景和代码",
            "原理讲解时：配合架构图和数据流图",
            "实践环节：展示真实的代码示例和运行结果",
            "Q&A环节：预留时间解答技术细节"
        ],
        "closing_strength": "总结核心知识点，提供延伸学习资源"
    },
    "prerequisites": {
        "required_knowledge": "本次分享假设听众具备基本的编程能力和相关技术基础",
        "optional_prep": "建议听众提前阅读相关技术文档",
        "materials_provided": "演讲结束后将分享演讲稿和代码示例"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "Go语言并发编程",
            "subtitle": "Goroutine调度机制与实战技巧",
            "author": "张明 | 基础架构部",
            "date": "2026年5月23日",
            "notes": "开场标题页，标题字体有重量感，副标题说明技术范围",
            "filling_prompt": "title填写真实技术主题名称，subtitle填副标题（如技术范围或亮点），author填演讲者姓名和部门，date填实际日期。禁止保留花括号占位符。",
            "visual_suggestions": "可添加抽象的集群拓扑图或Go语言gopher吉祥物作为背景装饰"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "内容概览",
            "kicker": "目录",
            "items": [
                "01  背景与问题",
                "02  核心原理",
                "03  架构设计",
                "04  实践案例",
                "05  总结与展望"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "目录页为固定结构，无需额外填充。items数组中填写各章节标题。",
            "timing_hint": "约30秒"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "背景与问题",
            "subtitle": "为什么需要Go并发",
            "notes": "章节分隔页，仪式感强，左侧彩色块配大号数字",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "design_notes": "章节标题使用大字号（44pt），与正文形成对比"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "传统并发模型的痛点",
            "content_type": "example_detail",
            "kicker": "实例 · 问题背景",
            "lede": "单线程等待导致CPU利用率不足40%，成为系统瓶颈",
            "context_block": "在Go出现之前，主流服务端语言依赖操作系统线程处理并发，每个线程占用2MB栈空间，万级别并发需要消耗大量内存，且上下文切换成本高昂。",
            "solution_block": "Go从语言层面引入Goroutine，每个仅占用2KB栈空间，由GMP调度器在用户态完成轻量级调度，可轻松支持10万级并发协程。",
            "metrics": [
                {"value": "2KB", "label": "Goroutine栈大小", "trend": "vs 线程2MB"},
                {"value": "10万+", "label": "单进程并发量", "trend": "轻松支撑"},
                {"value": "<1μs", "label": "上下文切换", "trend": "用户态完成"}
            ],
            "takeaway": "Goroutine让并发编程从奢侈品变为日常工具，开发者无需关注底层调度细节。",
            "notes": "用具体数字对比传统方案与Go协程的差异，不要空泛描述",
            "filling_prompt": "必须填入真实数据（通过 web_search 获取权威数据，至少2个URL）：lede一句话说明问题严重性；context_block描述行业普遍痛点；solution_block具体说明Go并发方案的优势；metrics_grid提供3个量化指标；takeaway一句话总结启示。references列出URL。禁止虚构数据。"
        },
        {
            "index": 5,
            "type": "two_column",
            "title": "方案对比",
            "content_type": "two_column",
            "kicker": "方案对比 · 技术选型",
            "left_header": "传统方案局限性",
            "left_sections": {
                "key_points": [
                    "线程创建成本高：每次创建约1MB开销",
                    "上下文切换慢：需进入内核态，约2-5μs",
                    "内存消耗大：万级并发需消耗20GB+内存",
                    "编程复杂：需手动管理线程池和队列"
                ],
                "data": [
                    "CPU利用率：通常30%-50%",
                    "内存占用：2MB/线程",
                    "调度延迟：2-5μs/次"
                ]
            },
            "right_header": "Go协程优势",
            "right_sections": {
                "key_points": [
                    "创建成本极低：仅2KB初始栈，可按需增长",
                    "用户态调度：GMP模型在用户态完成切换",
                    "高效内存：10万协程仅需约200MB",
                    "简洁语法：go func()即可启动并发"
                ],
                "data": [
                    "CPU利用率：可达80%以上",
                    "内存占用：2KB/协程（初始）",
                    "调度延迟：<1μs"
                ]
            },
            "notes": "左右对比传统方案的不足与Go协程的优势，每个维度用数据支撑",
            "filling_prompt": "必须填入真实内容：left_header为'传统方案局限性'，left_sections.key_points列出3-4个具体问题，left_sections.data列出量化指标；right_header为'Go协程优势'，right_sections.key_points列出对应的改进点，right_sections.data列出改进后指标。注意左右数据形成直观对比。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "核心原理",
            "subtitle": "Goroutine调度机制",
            "notes": "章节分隔页，数字'02'使用160pt大字号",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "transition_phrase": "接下来，让我们深入理解Goroutine的调度原理。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "核心概念解析",
            "content_type": "content_slide",
            "kicker": "要点 · 核心概念",
            "section_header": "GMP调度模型",
            "lede": "Go运行时维护三级调度体系：Goroutine（G）挂载在Processor（P）上，由Machine（M）执行",
            "bullets": [
                "G（Goroutine）：轻量级执行单元，保存执行栈和状态",
                "P（Processor）：执行上下文，最多=GOMAXPROCS个",
                "M（Machine）：操作系统线程，真正执行者",
                "全局运行队列GRQ与本地LRQ协同调度"
            ],
            "highlight_stats": [
                {"value": "GOMAXPROCS", "label": "默认=P核数"},
                {"value": "work-stealing", "label": "窃取调度算法"}
            ],
            "notes": "用简洁的语言解释GMP模型，配合文字示意图",
            "filling_prompt": "必须填入真实内容：lede一句话概括GMP模型；bullets列出3-4个核心概念，每条配合一句话说明；highlight_stats提供2个辅助指标。",
            "visual_suggestions": "可用文字描述GMP三者关系：M从P的LRQ取G执行，LRQ空时从GRQ或别的P偷取"
        },
        {
            "index": 8,
            "type": "process_flow",
            "title": "调度流程",
            "content_type": "process_flow",
            "kicker": "原理 · 调度机制",
            "direction": "horizontal",
            "steps": [
                {"num": "01", "title": "创建Goroutine", "desc": "go func()创建G，放入P本地队列"},
                {"num": "02", "title": "入队就绪", "desc": "G标记为runnable，加入LRQ或GRQ"},
                {"num": "03", "title": "调度获取", "desc": "M从P的LRQ取出G，准备执行"},
                {"num": "04", "title": "执行运行", "desc": "G在M上执行用户代码"},
                {"num": "05", "title": "阻塞/让出", "desc": "系统调用时M阻塞，调度器唤醒新M接手"}
            ],
            "notes": "用流程图展示Goroutine从创建到执行的完整生命周期",
            "filling_prompt": "必须填入真实内容：提供5个核心步骤，每步有title和一句话desc，展示Goroutine的调度流程。",
            "technical_detail_level": "overview"
        },
        {
            "index": 9,
            "type": "stat_slide",
            "title": "Go并发关键数据",
            "content_type": "stat_slide",
            "kicker": "数据 · 核心指标",
            "stats": [
                {"number": "10万+", "unit": "协程", "label": "单进程最大并发数", "trend": "↑ 远超线程"},
                {"number": "<1", "unit": "μs", "label": "Goroutine创建耗时", "trend": "↑ 比线程快千倍"},
                {"number": "80%+", "unit": "", "label": "压测CPU利用率", "trend": "↑ 远超高"}
            ],
            "notes": "3个大数字突出展示Go并发相比线程的核心优势",
            "filling_prompt": "必须填入真实数据：3个核心指标，每个有number（具体数字）、unit（单位，可为空）、label（效果说明）、trend（变化趋势）。指标要具体且有代表性。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "03",
            "title": "架构设计",
            "subtitle": "并发程序的架构模式",
            "notes": "章节分隔页，数字'03'使用160pt大字号",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "image_text",
            "title": "整体架构",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "架构 · 整体设计",
            "header": "并发程序分层架构",
            "sub_header": "组件关系与数据流",
            "paragraph": "并发程序通常采用分层架构设计。入口层负责接收请求并进行初步校验；业务逻辑层通过goroutine池并发处理核心任务；数据访问层使用连接池管理数据库资源，避免过度创建连接；监控层实时采集运行指标，通过channel与主流程解耦。整个架构强调：无锁共享、channel通信、context取消。",
            "notes": "用文字描述分层架构（组件+关系），不要求真实图片",
            "filling_prompt": "必须填入真实内容：用文字描述系统整体分层架构（组件名称+组件之间的关系）。paragraph为300-450字的自然语言段落，详细描述各层职责和交互关系，禁止罗列要点。",
            "visual_placeholder": "架构图区域：展示分层架构组件及其交互关系"
        },
        {
            "index": 12,
            "type": "card_grid",
            "title": "核心模块详解",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "kicker": "模块 · 核心组件",
            "cards": [
                {"header": "Goroutine池", "body": "复用协程而非每次创建，避免过度调度导致的开销", "icon": "01", "footer": "↑ 减少50%调度开销"},
                {"header": "Channel通道", "body": "goroutine间的安全通信机制，支持带缓冲和无缓冲模式", "icon": "02", "footer": "线程安全 · 无锁"},
                {"header": "Context上下文", "body": "携带截止时间、取消信号和请求域数据，贯穿调用链", "icon": "03", "footer": "超时控制 · 优雅取消"},
                {"header": "sync同步原语", "body": "Mutex、WaitGroup、Once等提供细粒度并发控制", "icon": "04", "footer": "原子操作 · 高效锁"}
            ],
            "notes": "用卡片展示4个核心模块，每个模块有header和一句话说明",
            "filling_prompt": "必须填入真实内容：提供4个核心模块，每个卡片有header（模块名称）、body（一句话说明功能）、footer（可选的效果标注）。模块名称要具体，描述要精准。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "04",
            "title": "实践案例",
            "subtitle": "真实项目中的应用与效果",
            "notes": "章节分隔页，数字'04'使用160pt大字号",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 14,
            "type": "case_study",
            "title": "微服务并发优化实战",
            "content_type": "case_study",
            "kicker": "案例 · 电商秒杀系统",
            "context": "某电商平台秒杀接口在双十一峰值时QPS骤降，接口超时率高达60%，用户体验极差。压测发现瓶颈在于订单处理线程池容量不足。",
            "problem": "Tomcat线程池固定为200，无法应对突发流量；数据库连接池耗尽导致大量等待；同步处理造成资源浪费。",
            "solution": "采用Go重写核心逻辑，使用goroutine池处理秒杀请求，Channel实现订单缓冲，Redis分布式锁保证库存一致性，异步落库降低数据库压力。",
            "results": [
                {"metric": "QPS", "value": "12,000/s", "comparison": "↑ 6倍"},
                {"metric": "延迟P99", "value": "45ms", "comparison": "↓ 80%"},
                {"metric": "成功率", "value": "99.7%", "comparison": "↑ 39.7%"},
                {"metric": "资源成本", "value": "-65%", "comparison": "服务器节省"}
            ],
            "notes": "Context→Problem→Solution→Results四段式，数据驱动",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：context描述背景（1-2句）；problem说明痛点（2-3句）；solution描述技术方案（2-3句）；results提供4个量化指标。references列出URL。"
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "关键数据看板",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 核心指标",
            "kpis": [
                {"value": "12,000/s", "label": "峰值QPS", "delta": "↑ 6倍", "baseline": "vs Java方案 2,000/s"},
                {"value": "45ms", "label": "P99延迟", "delta": "↓ 80%", "baseline": "vs Java方案 230ms"},
                {"value": "99.7%", "label": "请求成功率", "delta": "↑ 39.7%", "baseline": "vs Java方案 60%"},
                {"value": "-65%", "label": "服务器成本", "delta": "↓ 成本", "baseline": "硬件投入减少"}
            ],
            "notes": "4个核心指标，delta为变化比例，baseline为对比基准",
            "filling_prompt": "必须填入真实数据：提供4个核心数据指标，每个有value（具体数字）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。指标要具体且有代表性。"
        },
        {
            "index": 16,
            "type": "quote_slide",
            "title": "大师金句",
            "content_type": "quote_slide",
            "quote": "Don't communicate by sharing memory, share memory by communicating.",
            "attribution": "Rob Pike, Go语言之父",
            "kicker": "金句 · 并发哲学",
            "notes": "用一句经典引言收尾，增加分享的深度感",
            "filling_prompt": "必须填入真实内容：quote为经典引言（建议来自技术大师或官方文档），attribution为出处。禁止虚构。"
        },
        {
            "index": 17,
            "type": "section_divider",
            "number": "05",
            "title": "总结与展望",
            "subtitle": "核心要点与未来方向",
            "notes": "章节分隔页，数字'05'使用160pt大字号",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "transition_phrase": "最后，让我们总结一下本次分享的核心要点。"
        },
        {
            "index": 18,
            "type": "summary_slide",
            "title": "总结",
            "kicker": "总结",
            "key_points": [
                "01 GMP调度器实现用户态高效调度，支撑10万+并发协程",
                "02 Channel与Mutex各有适用场景，按需选择同步机制",
                "03 真实案例验证：Go并发可将系统吞吐量提升6倍"
            ],
            "thank_you": "感谢聆听",
            "contact": "联系邮箱：zhangming@company.com | 技术交流群：Go语言实战",
            "notes": "结尾页，核心回顾 + 感谢",
            "filling_prompt": "必须填入真实内容：key_points提供3个核心回顾要点；contact填写真实联系方式。禁止保留花括号。",
            "q_and_a_hint": "预留10-15分钟Q&A，欢迎提问",
            "materials_note": "演讲材料和代码示例将在会后通过邮件发送"
        }
    ],
    "design_tips": [
        "技术分享要注重内容深度，不要堆砌文字",
        "每章节用 section_divider 清晰划分",
        "用数字说明效果比文字描述更有说服力",
        "架构图用文字描述组件关系即可，不需要真实图片",
        "代码示例要精选，突出重点，避免大段代码",
        "预留Q&A时间，技术问题当场解答效果更好",
        "提供延伸学习资料，满足不同深度的学习需求",
        "结合自身实践经验，增加分享的真实性和说服力"
    ],
    "presentation_flow": {
        "opening": {
            "duration": "3-5分钟",
            "goal": "建立问题意识，引起工程师共鸣",
            "tip": "用一个实际的技术挑战或痛点开场"
        },
        "body": {
            "duration": "25-35分钟",
            "goal": "深入讲解原理和实践",
            "tip": "每个章节控制在5-8分钟，穿插互动问题"
        },
        "closing": {
            "duration": "5-10分钟",
            "goal": "总结要点，展望未来",
            "tip": "预留足够Q&A时间，解答技术细节"
        }
    },
    "code_examples_guidance": {
        "best_practices": [
            "代码示例要简短，突出核心逻辑",
            "添加注释解释关键步骤",
            "展示实际运行结果",
            "对比优化前后的差异"
        ],
        "common_pitfalls": [
            "避免一次性展示大量代码",
            "不要展示敏感的配置信息",
            "确保代码在演示环境下可以运行"
        ]
    }
}
