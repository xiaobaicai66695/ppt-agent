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
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{课程/章节名称}",
            "subtitle": "{课程简介}",
            "author": "{讲师}",
            "date": "{日期}",
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
            "filling_prompt": "目录页为固定结构，无需额外填充。"
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
            "type": "content_slide",
            "title": "学习目标",
            "content_type": "content_slide",
            "notes": "本章节需要掌握的3-4个核心知识点",
            "filling_prompt": "必须填入真实内容：提供3-4个具体学习目标，每条说明要掌握的知识点（如'理解Kubernetes核心概念'、'掌握Pod调度机制'）。"
        },
        {
            "index": 5,
            "type": "deep_dive",
            "title": "核心原理",
            "content_type": "deep_dive",
            "kicker": "详解 · 核心机制",
            "lede": "深入理解关键机制的核心价值",
            "left_column": {
                "key_points": [
                    "{要点1}",
                    "{要点2}",
                    "{要点3}",
                    "{要点4}"
                ],
                "analysis": [
                    "{分析维度1}",
                    "{分析维度2}"
                ]
            },
            "right_column": {
                "case_example": [
                    "{案例要素1}",
                    "{案例要素2}",
                    "{案例要素3}",
                    "{案例要素4}"
                ],
                "data_evidence": [
                    "{数据指标1}",
                    "{数据指标2}",
                    "{数据指标3}"
                ]
            },
            "notes": "双栏深入展开，左栏放核心要点和分析，右栏放案例和数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_column.key_points 为核心要点（3-4条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 web_search 获取的 URL（至少2个）。禁止保留花括号；禁止虚构数据。"
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
            "notes": "4个典型应用场景",
            "filling_prompt": "必须填入真实内容：提供4个典型应用场景，每个有 header（场景名称）和 body（一句话描述）。场景要具体，如'微服务架构'、'弹性扩容'。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "行业案例：{行业名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示行业应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {行业} 替换为具体行业（如'金融'、'电商'、'医疗'）；title 中的 {行业名称} 替换为具体行业名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX银行智能客服系统'，不超过35字）；sub_header 为项目名称（不超过35字）；bullets 列出3-4条应用效果或客户评价，每条不超过35字。references 列出 URL。禁止使用匿名实体；禁止虚构数据。"
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
            "notes": "5-6个操作步骤",
            "filling_prompt": "必须填入真实内容：提供5-6个具体操作步骤，每步有名称和一句话描述（如'步骤1：安装 kubectl'）。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 效果验证",
            "kpis": [
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"}
            ],
            "notes": "展示4个实践效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个实践效果指标，每个有 value（具体数字）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。指标要具体，如'部署时间缩短 90%'、'资源利用率提升 40%'。references 列出 web_search 获取的 URL（至少2个）。禁止保留花括号；禁止虚构数据。"
        },
        {
            "index": 12,
            "type": "content_slide",
            "title": "代码示例",
            "content_type": "content_slide",
            "notes": "关键代码或配置示例，配注释说明",
            "filling_prompt": "必须填入真实内容：展示关键代码片段或配置示例，配注释说明每段代码的作用。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "注意事项",
            "content_type": "content_slide",
            "notes": "常见问题和避坑指南",
            "filling_prompt": "必须填入真实内容：列出2-4个常见问题和避坑指南，每条说明具体问题（如'注意：ConfigMap更新后需要重启Pod才能生效'）。"
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
            "type": "content_slide",
            "title": "知识要点",
            "content_type": "content_slide",
            "notes": "本章核心知识点回顾",
            "filling_prompt": "必须填入真实内容：回顾本章3-4个核心知识点，每条用一句话概括。"
        },
        {
            "index": 16,
            "type": "image_text",
            "title": "延伸学习：{学习方向}",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "notes": "用图文混排展示延伸学习方向",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'延伸学习'；title 中的 {学习方向} 替换为具体学习方向；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为学习方向标题（不超过35字）；sub_header 为学习概述（不超过35字）；bullets 列出3-4条学习要点或推荐资源，每条不超过35字。references 列出 URL。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "下课了",
            "key_points": [
                "01 {核心要点1}",
                "02 {核心要点2}",
                "03 {延伸方向}"
            ],
            "thank_you": "感谢聆听",
            "contact": "{联系方式}",
            "notes": "结尾页",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（2个核心总结+1个延伸方向）；contact 填入讲师联系方式或课程群信息。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "课件要系统化，每个章节有明确的学习目标",
        "内容要循序渐进，从基础到进阶",
        "重点内容要突出，不要平均用力",
        "案例和实践部分要具体可操作",
        "结尾要有延伸学习方向"
    ]
}
