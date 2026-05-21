TEMPLATE = {
    "name": "training-course",
    "name_cn": "培训课件",
    "description": "适合内部培训、新人入职培训、技能培训等场景。知识系统，讲解清晰，互动引导。",
    "target_audience": "员工、新人、需要学习相关技能的人员",
    "typical_slides": 17,
    "typical_duration": "30-60分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "training_philosophy": "培训的目标是让学员'学会'而非'听过'",
        "key_principles": [
            "内容要由浅入深，循序渐进",
            "概念讲解要准确，举例要贴切",
            "图文并茂，增强理解",
            "留出互动和练习时间"
        ],
        "adult_learning": {
            "relevance": "培训内容要与工作实际相关",
            "experience": "善于利用学员已有经验",
            "problem_orientation": "聚焦解决实际问题",
            "autonomy": "给予学员一定选择权"
        }
    },
    "training_objectives": {
        "knowledge": "知识目标：理解XXX的概念和原理",
        "skill": "技能目标：能够独立完成XXX操作",
        "attitude": "态度目标：建立对XXX的正确认识"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "培训主题名称",
            "subtitle": "培训副标题",
            "author": "培训讲师姓名",
            "date": "培训日期",
            "notes": "标题页正式，注明培训主题和讲师",
            "filling_prompt": "必须填入真实内容：title 为培训主题名称，subtitle 为培训副标题，author 为培训讲师姓名，date 为培训日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "培训目标",
            "content_type": "content_slide",
            "objectives": [
                "培训目标1",
                "培训目标2",
                "培训目标3",
                "培训目标4",
                "培训目标5"
            ],
            "notes": "明确培训要达到的目标",
            "filling_prompt": "必须填入真实内容：列出3-5条培训目标，每条说明学员学完本次培训后能够做什么（用动词开头，如'能够...'、'掌握...'、'理解...'）。"
        },
        {
            "index": 3,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  基础知识",
                "02  核心内容",
                "03  实战演练",
                "04  总结回顾"
            ],
            "notes": "让学员了解培训结构",
            "filling_prompt": "目录页为固定结构，可根据实际内容调整章节。"
        },
        {
            "index": 4,
            "type": "section_divider",
            "number": "01",
            "title": "基础知识",
            "subtitle": "概念与背景",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "核心概念",
            "content_type": "content_slide",
            "concepts": [
                {"name": "概念名称", "desc": "概念定义/说明（不超过35字）"},
                {"name": "概念名称", "desc": "概念定义/说明（不超过35字）"},
                {"name": "概念名称", "desc": "概念定义/说明（不超过35字）"},
                {"name": "概念名称", "desc": "概念定义/说明（不超过35字）"}
            ],
            "notes": "讲解核心概念和定义",
            "filling_prompt": "必须填入真实内容：列出3-5个核心概念，每个有名称和定义/说明（每条说明不超过35字）。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "背景介绍",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景",
            "header": "背景主题",
            "sub_header": "背景说明",
            "paragraph": "详细描述培训内容的背景、重要性和发展趋势，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排讲解背景知识",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景主题（不超过35字）；sub_header 为背景说明；paragraph 为300-450字的自然语言段落，详细描述培训内容的背景、重要性和发展趋势，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "02",
            "title": "核心内容",
            "subtitle": "重点与难点",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "example_detail",
            "title": "重点知识",
            "content_type": "example_detail",
            "kicker": "实例 · 重点知识",
            "lede": "一句话概括最核心的知识点",
            "context_block": "描述常见误区或薄弱环节（1-2句话）。",
            "solution_block": "详细讲解核心要点和实践技巧（2-3句话）。",
            "metrics": [
                {"value": "时间", "label": "预计掌握时间", "trend": "vs 传统方式"},
                {"value": "频率", "label": "实际应用场景", "trend": "日常工作"},
                {"value": "程度", "label": "预期效果", "trend": "提升效果"}
            ],
            "takeaway": "一句话总结核心价值。",
            "notes": "讲解本次培训的重点内容",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的知识点；context_block 描述常见误区或薄弱环节（1-2句话）；solution_block 详细讲解核心要点和实践技巧（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结核心价值。禁止保留花括号。"
        },
        {
            "index": 9,
            "type": "process_flow",
            "title": "操作流程",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "步骤1", "desc": "步骤说明（不超过35字）"},
                {"num": "02", "title": "步骤2", "desc": "步骤说明（不超过35字）"},
                {"num": "03", "title": "步骤3", "desc": "步骤说明（不超过35字）"},
                {"num": "04", "title": "步骤4", "desc": "步骤说明（不超过35字）"},
                {"num": "05", "title": "步骤5", "desc": "步骤说明（不超过35字）"}
            ],
            "notes": "用流程图展示操作步骤",
            "filling_prompt": "必须填入真实内容：提供5个操作步骤，每个步骤有 title（步骤名称，不超过35字）和 desc（操作说明，不超过35字）。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "案例讲解",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "案例",
            "header": "案例标题",
            "sub_header": "案例简介",
            "paragraph": "详细描述案例的背景、具体情况、处理过程和经验启示，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排讲解一个具体案例",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为案例简介；paragraph 为300-450字的自然语言段落，详细描述案例的背景、具体情况、处理过程和经验启示，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "03",
            "title": "实战演练",
            "subtitle": "动手练习",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 12,
            "type": "example_detail",
            "title": "练习任务",
            "content_type": "example_detail",
            "kicker": "实例 · 练习任务",
            "lede": "一句话概括练习核心目标",
            "context_block": "描述练习目的和预期效果（1-2句话）。",
            "solution_block": "具体说明操作步骤和关键要点（2-3句话）。",
            "metrics": [
                {"value": "时长", "label": "预计完成时间", "trend": "含讨论和分享"},
                {"value": "形式", "label": "评判标准", "trend": "说明"},
                {"value": "百分比", "label": "预计正确率", "trend": "vs 培训前"}
            ],
            "takeaway": "一句话总结价值。",
            "notes": "布置课堂练习任务",
            "filling_prompt": "必须填入真实内容：lede 一句话概括练习核心目标；context_block 描述练习目的和预期效果（1-2句话）；solution_block 具体说明操作步骤和关键要点（2-3句话）；metrics_grid 提供3个练习指标；takeaway 用一句话总结价值。"
        },
        {
            "index": 13,
            "type": "example_detail",
            "title": "常见问题",
            "content_type": "example_detail",
            "kicker": "实例 · 常见问题",
            "lede": "一句话概括最易遇到的问题",
            "context_block": "描述问题具体表现（1-2句话）。",
            "solution_block": "针对每个问题提供解决方法（2-3句话）。",
            "metrics": [
                {"value": "百分比", "label": "问题发生率", "trend": "vs 老员工"},
                {"value": "时间", "label": "平均解决时间", "trend": "提前预防"},
                {"value": "程度", "label": "问题影响级别", "trend": "说明"}
            ],
            "takeaway": "一句话总结如何避免。",
            "notes": "列出常见问题和解决方法",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最易遇到的问题；context_block 描述问题具体表现（1-2句话）；solution_block 针对每个问题提供解决方法（2-3句话）；metrics_grid 提供3个问题相关指标；takeaway 用一句话总结如何避免。禁止保留花括号。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "04",
            "title": "总结回顾",
            "subtitle": "知识巩固",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 15,
            "type": "content_slide",
            "title": "知识回顾",
            "content_type": "content_slide",
            "highlights": [
                {"title": "知识点1", "desc": "简要说明"},
                {"title": "知识点2", "desc": "简要说明"},
                {"title": "知识点3", "desc": "简要说明"},
                {"title": "知识点4", "desc": "简要说明"},
                {"title": "知识点5", "desc": "简要说明"}
            ],
            "notes": "回顾本次培训的核心知识点",
            "filling_prompt": "必须填入真实内容：列出4-5个核心知识点回顾，每条有标题（不超过35字）和简要说明（不超过35字）。"
        },
        {
            "index": 16,
            "type": "kpi_dashboard",
            "title": "效果自评",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "自评 · 学习效果",
            "kpis": [
                {"value": "分数", "label": "知识掌握程度", "delta": "vs 培训前+X分", "baseline": "满分10分"},
                {"value": "分数", "label": "技能熟练程度", "delta": "vs 培训前+X分", "baseline": "满分10分"},
                {"value": "分数", "label": "认同程度", "delta": "vs 培训前+X分", "baseline": "满分10分"},
                {"value": "计划", "label": "下一步计划", "delta": "制定中", "baseline": "待确定"}
            ],
            "notes": "帮助学员自评学习效果",
            "filling_prompt": "必须填入真实内容：提供4个学习效果自评维度，每个有 value（如分数或完成度）、label（维度名称）、delta（提升情况）。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心知识点1",
                "02 核心知识点2",
                "03 后续学习：描述"
            ],
            "thank_you": "感谢聆听！",
            "contact": "HR邮箱 | 新人导师联系方式",
            "notes": "简洁总结，提出后续学习建议",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心知识点2条+下一步行动1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "培训内容要由浅入深，循序渐进",
        "概念讲解要准确，举例要贴切",
        "图文并茂，增强理解",
        "留出互动和练习时间",
        "结尾要有后续学习指引",
        "配套测试题，检验学习效果",
        "收集反馈，持续改进培训内容",
        "提供学习资料，方便后续复习"
    ],
    "training_materials": {
        "handouts": [
            "培训课件PDF",
            "操作手册",
            "制度汇编",
            "常见问题FAQ"
        ],
        "online_resources": [
            "内部知识库",
            "学习视频中心",
            "导师答疑群"
        ]
    },
    "assessment_methods": {
        "during_training": "课堂测验、练习表现",
        "after_training": "在线测试",
        "on_job": "1个月后绩效跟踪"
    }
}
