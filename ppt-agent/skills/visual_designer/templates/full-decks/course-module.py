TEMPLATE = {
    "name": "course-module",
    "name_cn": "课程课件",
    "description": "适合教学课件、培训材料、知识分享等场景。内容系统化，重点清晰，便于学习和理解。",
    "target_audience": "学生、培训学员、自学者",
    "typical_slides": 17,
    "typical_duration": "45-90分钟（一节课或一个章节）",
    "palette": "sage_calm",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "teaching_philosophy": "以学生为中心，注重启发式教学",
        "key_principles": [
            "循序渐进：从已知到未知，从简单到复杂",
            "理论联系实际：通过案例帮助理解抽象概念",
            "互动引导：设置思考题和讨论环节",
            "及时巩固：每个知识点后有练习和反馈"
        ],
        "assessment_hints": "可配套准备课后作业、测验题目"
    },
    "learning_objectives": {
        "knowledge": "学员将理解XXX的基本概念和原理",
        "skills": "学员将能够运用XXX解决实际问题",
        "attitude": "学员将建立对XXX的正确认识和学习兴趣"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "课程/章节名称",
            "subtitle": "课程简介",
            "author": "讲师姓名",
            "date": "日期",
            "filling_prompt": "必须填入真实内容：title 为实际课程或章节名称，subtitle 为一句课程简介，author 为讲师姓名，date 为日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  基础知识",
                "02  进阶应用",
                "03  实践操作",
                "04  总结回顾"
            ],
            "notes": "让学员快速了解课程结构，每项一行即可",
            "filling_prompt": "目录页为固定结构，无需额外填充。",
            "timing_hint": "约30秒"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "基础知识",
            "subtitle": "核心概念与原理",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "学习目标",
            "content_type": "example_detail",
            "kicker": "实例 · 学习目标",
            "lede": "一句话说明本章的核心价值",
            "context_block": "描述学员常见的困惑或误区（1-2句话）。",
            "solution_block": "说明本章内容如何帮助解决（2-3句话）。",
            "metrics": [
                {"value": "百分比", "label": "预计掌握程度", "trend": "vs 课程前"},
                {"value": "时长", "label": "预计所需时间", "trend": "理论+实践"},
                {"value": "数量", "label": "配套实践项目", "trend": "循序渐进"}
            ],
            "takeaway": "一句话总结学完本章的核心收获。",
            "notes": "本章节需要掌握的3-4个核心知识点",
            "filling_prompt": "必须填入真实内容：lede 一句话说明本章的核心价值；context_block 描述学员常见的困惑或误区（1-2句话）；solution_block 说明本章内容如何帮助解决（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结学完本章的核心收获。禁止保留花括号。"
        },
        {
            "index": 5,
            "type": "deep_dive",
            "title": "核心机制详解",
            "content_type": "deep_dive",
            "kicker": "详解 · 核心机制",
            "lede": "一句话概括核心价值",
            "left_column": {
                "key_points": [
                    "核心要点1",
                    "核心要点2",
                    "核心要点3",
                    "核心要点4"
                ],
                "analysis": [
                    "分析维度1",
                    "分析维度2"
                ]
            },
            "right_column": {
                "case_example": [
                    "案例步骤1",
                    "案例步骤2",
                    "案例步骤3",
                    "案例步骤4"
                ],
                "data_evidence": [
                    "效果指标1",
                    "效果指标2",
                    "效果指标3"
                ]
            },
            "notes": "双栏深入展开，左栏放核心要点和分析，右栏放案例和数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_column.key_points 为核心要点（3-4条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "进阶应用",
            "subtitle": "典型场景与案例",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "应用场景",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "场景名称", "body": "场景描述"},
                {"header": "场景名称", "body": "场景描述"},
                {"header": "场景名称", "body": "场景描述"},
                {"header": "场景名称", "body": "场景描述"}
            ],
            "notes": "4个典型应用场景",
            "filling_prompt": "必须填入真实内容：提供4个典型应用场景，每个有 header（场景名称）和 body（一句话描述）。场景要具体。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "行业案例",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "行业",
            "header": "案例标题",
            "sub_header": "项目名称",
            "paragraph": "详细描述该行业案例的背景、实施方案和应用效果，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示行业应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为项目名称；paragraph 为300-450字的自然语言段落，详细描述该行业案例的背景、实施方案和应用效果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "实践操作",
            "subtitle": "动手操作与代码示例",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "操作步骤",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "1", "title": "步骤名称", "desc": "步骤描述"},
                {"num": "2", "title": "步骤名称", "desc": "步骤描述"},
                {"num": "3", "title": "步骤名称", "desc": "步骤描述"},
                {"num": "4", "title": "步骤名称", "desc": "步骤描述"},
                {"num": "5", "title": "步骤名称", "desc": "步骤描述"}
            ],
            "notes": "5-6个操作步骤",
            "filling_prompt": "必须填入真实内容：提供5-6个具体操作步骤，每步有名称和一句话描述。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 效果验证",
            "kpis": [
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 基准"}
            ],
            "notes": "展示4个实践效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个实践效果指标，每个有 value、label、delta、baseline。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 12,
            "type": "example_detail",
            "title": "代码示例",
            "content_type": "example_detail",
            "kicker": "实例 · 代码示例",
            "lede": "一句话概括核心功能",
            "context_block": "描述要解决的问题和使用场景（1-2句话）。",
            "solution_block": "详细解释代码结构和关键逻辑（2-3句话）。",
            "metrics": [
                {"value": "行数", "label": "代码规模", "trend": "简洁度"},
                {"value": "百分比", "label": "效率提升", "trend": "vs 基准"},
                {"value": "程度", "label": "代码复用性", "trend": "可应用性"}
            ],
            "takeaway": "一句话总结代码示例的核心价值。",
            "notes": "关键代码或配置示例，配注释说明",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心功能；context_block 描述要解决的问题和使用场景（1-2句话）；solution_block 详细解释代码结构和关键逻辑（2-3句话）；metrics_grid 提供3个代码相关指标；takeaway 用一句话总结代码示例的核心价值。"
        },
        {
            "index": 13,
            "type": "example_detail",
            "title": "注意事项",
            "content_type": "example_detail",
            "kicker": "实例 · 避坑指南",
            "lede": "一句话概括最容易踩的坑",
            "context_block": "描述问题的具体表现和发生场景（1-2句话）。",
            "solution_block": "针对每个问题提供正确做法和解决方案（2-3句话）。",
            "metrics": [
                {"value": "百分比", "label": "踩坑率", "trend": "常见程度"},
                {"value": "时长", "label": "平均解决时间", "trend": "浪费时间"},
                {"value": "程度", "label": "问题影响", "trend": "影响准确性"}
            ],
            "takeaway": "一句话总结如何避免常见错误。",
            "notes": "常见问题和避坑指南",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最容易踩的坑；context_block 描述问题的具体表现和发生场景（1-2句话）；solution_block 针对每个问题提供正确做法和解决方案（2-3句话）；metrics_grid 提供3个避坑相关指标；takeaway 用一句话总结如何避免常见错误。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "04",
            "title": "总结回顾",
            "subtitle": "要点与延伸",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 15,
            "type": "example_detail",
            "title": "知识要点",
            "content_type": "example_detail",
            "kicker": "实例 · 知识要点",
            "lede": "一句话概括最重要的知识点",
            "context_block": "回顾本章核心内容（1-2句话）。",
            "solution_block": "总结学习要点和关键收获（2-3句话）。",
            "metrics": [
                {"value": "数量", "label": "NumPy方法数", "trend": "提升幅度"},
                {"value": "数量", "label": "Pandas方法数", "trend": "使用频率"},
                {"value": "推荐", "label": "延伸学习方向", "trend": "推荐深度"}
            ],
            "takeaway": "一句话总结如何将知识运用到实际工作中。",
            "notes": "本章核心知识点回顾",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最重要的知识点；context_block 回顾本章核心内容（1-2句话）；solution_block 总结学习要点和关键收获（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结如何将知识运用到实际工作中。"
        },
        {
            "index": 16,
            "type": "image_text",
            "title": "延伸学习",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "延伸学习",
            "header": "学习方向标题",
            "sub_header": "学习概述",
            "paragraph": "详细描述该学习方向的背景、推荐理由和实践方法，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示延伸学习方向",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为学习方向标题（不超过35字）；sub_header 为学习概述；paragraph 为300-450字的自然语言段落，详细描述该学习方向的背景、推荐理由和实践方法，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "下课了",
            "key_points": [
                "01 核心总结1",
                "02 核心总结2",
                "03 延伸方向：描述"
            ],
            "thank_you": "感谢聆听",
            "contact": "课程群 | 课件下载邮箱",
            "notes": "结尾页",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（2个核心总结+1个延伸方向）；contact 填入讲师联系方式或课程群信息。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "课件要系统化，每个章节有明确的学习目标",
        "内容要循序渐进，从基础到进阶",
        "重点内容要突出，不要平均用力",
        "案例和实践部分要具体可操作",
        "结尾要有延伸学习方向",
        "配合代码演示，增加互动性",
        "设置思考题，引导学员主动思考",
        "提供课后练习，巩固学习效果"
    ],
    "teaching_notes": {
        "time_allocation": {
            "theory": "40% 理论讲解",
            "practice": "40% 动手实践",
            "discussion": "20% 互动讨论"
        },
        "interactive_elements": [
            "随堂测验",
            "代码演示",
            "小组讨论",
            "答疑环节"
        ],
        "assessment_methods": [
            "课堂表现：参与度和小测验",
            "实践作业：完成分析报告",
            "期末考试：综合能力测试"
        ]
    }
}
