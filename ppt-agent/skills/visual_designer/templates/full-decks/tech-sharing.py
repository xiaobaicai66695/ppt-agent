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
            "title": "技术主题名称",
            "subtitle": "副标题",
            "author": "演讲者姓名 | 部门",
            "date": "实际日期",
            "notes": "开场标题页，留白充足，标题字体有重量感",
            "filling_prompt": "必须填入真实内容：title 为本次分享的实际技术主题名称，author 为演讲者姓名，date 为实际日期。禁止保留花括号占位符。",
            "visual_suggestions": "可添加抽象的集群拓扑图或容器示意图作为背景"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  背景与问题",
                "02  核心原理",
                "03  架构设计",
                "04  实践案例",
                "05  总结与展望"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "目录页为固定结构，无需额外填充。",
            "timing_hint": "约30秒"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "背景与问题",
            "subtitle": "为什么需要这项技术",
            "notes": "章节分隔页，仪式感强",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "design_notes": "章节标题使用大字号，与正文形成对比"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "问题背景",
            "content_type": "example_detail",
            "kicker": "实例 · 问题背景",
            "lede": "一句话说明问题的严重性",
            "context_block": "描述行业普遍痛点（1-2句话）。",
            "solution_block": "具体说明该痛点导致的后果或损失（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "指标1", "trend": "变化趋势"},
                {"value": "数字", "label": "指标2", "trend": "变化趋势"},
                {"value": "数字", "label": "指标3", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结启示。",
            "notes": "用具体数字说明痛点，不要空泛描述",
            "filling_prompt": "必须填入真实内容（通过 web_search 获取权威数据，至少2个URL）：lede 一句话说明问题的严重性；context_block 描述行业普遍痛点（1-2句话）；solution_block 具体说明该痛点导致的后果或损失（2-3句话）；metrics_grid 提供3个量化指标，每个有 value（数字）、label（说明）、trend（趋势）；takeaway 用一句话总结启示。禁止空泛描述。references 列出 URL。禁止保留花括号。"
        },
        {
            "index": 5,
            "type": "two_column",
            "title": "现有方案分析",
            "content_type": "two_column",
            "kicker": "方案对比 · 技术选型",
            "left_header": "传统方案局限性",
            "left_sections": {
                "analysis": [
                    "现有方案问题1及影响",
                    "现有方案问题2及影响",
                    "现有方案问题3及影响"
                ],
                "data": [
                    "效果指标1",
                    "效果指标2",
                    "效果指标3"
                ]
            },
            "right_header": "改进方向",
            "right_sections": {
                "key_points": [
                    "改进方案1及效果",
                    "改进方案2及效果",
                    "改进方案3及效果",
                    "改进方案4及效果"
                ],
                "data": [
                    "效果指标1",
                    "效果指标2",
                    "效果指标3"
                ]
            },
            "notes": "左右对比传统方案的不足与改进方向，每个维度用数据支撑",
            "filling_prompt": "必须填入真实内容：left_header 为'传统方案局限性'，left_sections.analysis 列出2-3个具体问题并说明其影响，left_sections.data 列出2-3个量化指标；right_header 为'改进方向'，right_sections.key_points 列出2-3条对应的改进方案并说明效果，right_sections.data 列出对应的改进后指标。注意左右数据形成对比。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "核心原理",
            "subtitle": "技术核心概念",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "transition_phrase": "接下来，让我们深入理解核心概念。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "核心概念",
            "content_type": "content_slide",
            "concepts": [
                {"name": "概念名称", "desc": "概念说明"},
                {"name": "概念名称", "desc": "概念说明"},
                {"name": "概念名称", "desc": "概念说明"},
                {"name": "概念名称", "desc": "概念说明"}
            ],
            "notes": "用简洁的语言解释核心概念，配合示意图（文字描述即可）",
            "filling_prompt": "必须填入真实内容：用通俗语言解释3-4个核心概念，每条配合一句话说明。可用文字描述示意图内容。",
            "visual_suggestions": "配合简单的架构示意图效果更佳"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "调度流程",
            "content_type": "process_flow",
            "direction": "horizontal",
            "steps": [
                {"num": "1", "title": "步骤1", "desc": "步骤描述"},
                {"num": "2", "title": "步骤2", "desc": "步骤描述"},
                {"num": "3", "title": "步骤3", "desc": "步骤描述"},
                {"num": "4", "title": "步骤4", "desc": "步骤描述"},
                {"num": "5", "title": "步骤5", "desc": "步骤描述"}
            ],
            "notes": "用流程图展示核心步骤，3-5步为宜",
            "filling_prompt": "必须填入真实内容：提供3-5个核心步骤，每步有名称和一句话描述，展示该技术的工作流程。",
            "technical_detail_level": "overview"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "架构设计",
            "subtitle": "系统整体架构与模块划分",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "整体架构",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "architecture_description": "用文字描述系统整体架构（组件名称+组件之间的关系）。",
            "notes": "用文字描述架构图（组件+关系），不要求真实图片",
            "filling_prompt": "必须填入真实内容：用文字描述系统整体架构（组件名称+组件之间的关系，如'API Server接收请求 → Scheduler分配节点 → Kubelet执行 → 状态同步至etcd'）。",
            "visual_placeholder": "架构图区域：展示核心组件及其交互关系"
        },
        {
            "index": 11,
            "type": "content_slide",
            "title": "核心模块详解",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "模块名称", "body": "模块功能描述"},
                {"header": "模块名称", "body": "模块功能描述"},
                {"header": "模块名称", "body": "模块功能描述"},
                {"header": "模块名称", "body": "模块功能描述"}
            ],
            "notes": "用卡片展示核心模块，每个模块一句话说明",
            "filling_prompt": "必须填入真实内容：提供4个核心模块，每个模块有 header（模块名称）和 body（一句话说明功能）。模块名称要具体。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "04",
            "title": "实践案例",
            "subtitle": "真实项目中的应用与效果",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "image_text",
            "title": "应用案例",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "行业",
            "header": "案例标题",
            "sub_header": "项目名称",
            "paragraph": "详细描述案例背景、技术方案、实施过程和应用效果，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示具体应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为项目名称；paragraph 为300-450字的自然语言段落，详细描述案例背景、技术方案、实施过程和应用效果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 14,
            "type": "kpi_dashboard",
            "title": "关键数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 核心指标",
            "kpis": [
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 迁移前"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 迁移前"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 迁移前"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 迁移前"}
            ],
            "notes": "4个核心指标，delta 为变化比例，baseline 为对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心数据指标，每个有 value（具体数字）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。指标要具体且有代表性。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 15,
            "type": "section_divider",
            "number": "05",
            "title": "总结与展望",
            "subtitle": "核心要点与未来方向",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "transition_phrase": "最后，让我们总结一下本次分享的核心要点。"
        },
        {
            "index": 16,
            "type": "example_detail",
            "title": "核心要点",
            "content_type": "example_detail",
            "kicker": "实例 · 核心要点",
            "lede": "一句话概括最核心的信息",
            "context_block": "回顾分享的核心内容（1-2句话）。",
            "solution_block": "总结核心要点和关键结论（2-3句话）。",
            "metrics": [
                {"value": "程度", "label": "知识点覆盖", "trend": "vs 分享前"},
                {"value": "程度", "label": "架构认知", "trend": "vs 分享前"},
                {"value": "程度", "label": "动手能力", "trend": "vs 分享前"}
            ],
            "takeaway": "一句话总结如何将分享应用到实际工作中。",
            "notes": "3-4条核心要点，用加粗序号",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的信息；context_block 回顾分享的核心内容（1-2句话）；solution_block 总结核心要点和关键结论（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结如何将分享应用到实际工作中。禁止保留花括号。"
        },
        {
            "index": 17,
            "type": "example_detail",
            "title": "未来方向",
            "content_type": "example_detail",
            "kicker": "实例 · 未来方向",
            "lede": "一句话概括最重要演进趋势",
            "context_block": "说明当前现状和局限（1-2句话）。",
            "solution_block": "详细描述未来演进方向和突破点（2-3句话）。",
            "metrics": [
                {"value": "频率", "label": "版本更新", "trend": "每季度大版本"},
                {"value": "数量", "label": "生态工具增长", "trend": "趋势"},
                {"value": "百分比", "label": "企业采纳率", "trend": "趋势"}
            ],
            "takeaway": "一句话总结如何把握未来趋势。",
            "notes": "技术演进方向或后续规划",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最重要演进趋势；context_block 说明当前现状和局限（1-2句话）；solution_block 详细描述未来演进方向和突破点（2-3句话）；metrics_grid 提供3个趋势指标；takeaway 用一句话总结如何把握未来趋势。禁止保留花括号。"
        },
        {
            "index": 18,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心回顾1",
                "02 核心回顾2",
                "03 核心回顾3"
            ],
            "thank_you": "感谢聆听",
            "contact": "联系方式：邮箱 | 技术交流群",
            "notes": "结尾页，核心回顾 + 感谢",
            "filling_prompt": "必须填入真实内容：key_points 提供3个核心回顾要点；contact 填写真实联系方式。禁止保留花括号。",
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
