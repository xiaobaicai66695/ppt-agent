TEMPLATE = {
    "name": "training-course",
    "name_cn": "培训课件",
    "description": "适合内部培训、新人入职培训、技能培训等场景。知识系统，讲解清晰，互动引导。",
    "target_audience": "员工、新人、需要学习相关技能的人员",
    "typical_slides": 16,
    "typical_duration": "30-60分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{培训主题}培训",
            "subtitle": "{培训副标题}",
            "author": "{培训讲师}",
            "date": "{培训日期}",
            "notes": "标题页正式，注明培训主题和讲师",
            "filling_prompt": "必须填入真实内容：title 为培训主题名称，subtitle 为培训副标题，author 为培训讲师姓名，date 为培训日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "培训目标",
            "content_type": "content_slide",
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
            "notes": "讲解核心概念和定义",
            "filling_prompt": "必须填入真实内容：列出3-5个核心概念，每个有名称和定义/说明（每条说明不超过35字）。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "背景介绍",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排讲解背景知识",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'背景'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景主题（不超过35字）；sub_header 为背景说明；bullets 列出3-4条背景要点，每条不超过35字。references 列出 URL。"
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
            "type": "content_slide",
            "title": "重点知识",
            "content_type": "content_slide",
            "notes": "讲解本次培训的重点内容",
            "filling_prompt": "必须填入真实内容：列出4-5个重点知识，每个有标题（不超过35字）和详细说明（每条说明不超过35字）。"
        },
        {
            "index": 9,
            "type": "process_flow",
            "title": "操作流程",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "{步骤1}", "desc": "{操作说明，不超过35字}"},
                {"num": "02", "title": "{步骤2}", "desc": "{操作说明，不超过35字}"},
                {"num": "03", "title": "{步骤3}", "desc": "{操作说明，不超过35字}"},
                {"num": "04", "title": "{步骤4}", "desc": "{操作说明，不超过35字}"},
                {"num": "05", "title": "{步骤5}", "desc": "{操作说明，不超过35字}"}
            ],
            "notes": "用流程图展示操作步骤",
            "filling_prompt": "必须填入真实内容：提供5个操作步骤，每个步骤有 title（步骤名称，不超过35字）和 desc（操作说明，不超过35字）。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "案例讲解：{案例名称}",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "notes": "用图文混排讲解一个具体案例",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'案例'；title 中的 {案例名称} 替换为具体案例；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为案例简介；bullets 列出3-4条案例要点，每条不超过35字。references 列出 URL。"
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
            "type": "content_slide",
            "title": "练习任务",
            "content_type": "content_slide",
            "notes": "布置课堂练习任务",
            "filling_prompt": "必须填入真实内容：说明练习任务的要求、目标和完成标准，列出3-5条操作要点或检查清单。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "常见问题",
            "content_type": "content_slide",
            "notes": "列出常见问题和解决方法",
            "filling_prompt": "必须填入真实内容：列出3-5个常见问题，每个问题有标题（不超过35字）和解决方法（每条说明不超过35字）。"
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
            "notes": "帮助学员自评学习效果",
            "filling_prompt": "必须填入真实内容：提供4个学习效果自评维度，每个有 value（如分数或完成度）、label（维度名称）、delta（提升情况）。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心知识点1}",
                "02 {核心知识点2}",
                "03 {下一步行动}"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁总结，提出后续学习建议",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心知识点2条+下一步行动1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "培训内容要由浅入深，循序渐进",
        "概念讲解要准确，举例要贴切",
        "图文并茂，增强理解",
        "留出互动和练习时间",
        "结尾要有后续学习指引"
    ]
}
