TEMPLATE = {
    "name": "personal-summary",
    "name_cn": "述职报告",
    "description": "适合个人总结、述职报告、年终总结等场景。重点突出，成果可见，计划明确。",
    "target_audience": "领导、同事、评审",
    "typical_slides": 10,
    "typical_duration": "10-15分钟",
    "palette": "report_green",
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
            "title": "{时间段}述职报告",
            "subtitle": "{姓名} | {部门/岗位}",
            "author": "{姓名}",
            "date": "{日期}",
            "notes": "标题页简洁正式",
            "filling_prompt": "必须填入真实内容：title 为时间段（如'2024年度'）+ 述职报告，subtitle 为姓名和岗位信息，author 为姓名，date 为述职日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  工作概述",
                "02  主要成果",
                "03  问题反思",
                "04  未来计划"
            ],
            "notes": "让领导快速了解汇报结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "工作概述",
            "subtitle": "岗位职责与工作回顾",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "kpi_dashboard",
            "title": "工作概览",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 工作概览",
            "notes": "用数据展示工作量和成果",
            "filling_prompt": "必须填入真实内容：提供4个关键工作指标（如完成项目数、关键成果、参与会议次数、团队协作次数等），每个有 value、label、delta、baseline。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "主要成果",
            "subtitle": "核心业绩与贡献",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "核心业绩：{项目名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示核心业绩，增强可信性",
            "filling_prompt": "必须填入真实内容：kicker 为'核心业绩'；title 中的 {项目名称} 替换为具体项目或工作名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为项目标题（不超过35字）；sub_header 为项目概述（不超过35字）；bullets 列出3-4条主要成果或贡献，每条不超过35字。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "其他成果",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "4个其他重要成果",
            "filling_prompt": "必须填入真实内容：提供4个其他工作成果，每个有 header（成果名称）和 body（一句话描述成果内容）。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "问题反思",
            "subtitle": "不足与改进",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "two_column",
            "title": "反思与改进",
            "content_type": "two_column",
            "notes": "坦诚分析不足和改进计划",
            "filling_prompt": "必须填入真实内容：left_header 为'不足与反思'，left_bullets 列出2-4条工作中的不足和反思；right_header 为'改进计划'，right_bullets 列出对应的改进措施和计划。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "未来计划",
            "subtitle": "下阶段目标与规划",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "process_flow",
            "title": "下阶段计划",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "{目标1}", "desc": "{具体行动}"},
                {"num": "02", "title": "{目标2}", "desc": "{具体行动}"},
                {"num": "03", "title": "{目标3}", "desc": "{具体行动}"},
                {"num": "04", "title": "{目标4}", "desc": "{具体行动}"}
            ],
            "notes": "4个下阶段工作目标",
            "filling_prompt": "必须填入真实内容：提供4个下阶段工作目标，每个有 title（目标名称）和 desc（具体行动）。目标要具体可衡量。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心成果}",
                "02 {主要收获}",
                "03 {努力方向}"
            ],
            "thank_you": "感谢聆听，请领导批评指正！",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心成果+主要收获+努力方向）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "述职要实事求是，成果和问题都要有",
        "数据要真实，用具体数字说话",
        "反思要诚恳，改进计划要可行",
        "计划要具体，有时间节点",
        "PPT要简洁，不要堆砌文字"
    ]
}
