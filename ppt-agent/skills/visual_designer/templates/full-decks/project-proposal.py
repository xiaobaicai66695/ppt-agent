TEMPLATE = {
    "name": "project-proposal",
    "name_cn": "项目提案",
    "description": "适合新项目立项、项目申请、资源申请等场景。理由充分，方案可行，预算清晰。",
    "target_audience": "领导、评审委员会、项目审批方",
    "typical_slides": 12,
    "typical_duration": "15-20分钟",
    "palette": "charcoal_light",
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
            "title": "{项目名称}项目提案",
            "subtitle": "{项目定位/一句话说明}",
            "author": "{提案人/团队}",
            "date": "{提案日期}",
            "notes": "标题页正式，清晰标注项目名称",
            "filling_prompt": "必须填入真实内容：title 为项目名称+项目提案，subtitle 为项目定位或一句话说明，author 为提案人或团队名称，date 为提案日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  项目背景",
                "02  需求分析",
                "03  解决方案",
                "04  项目计划",
                "05  预算估算",
                "06  预期效果"
            ],
            "notes": "让评审快速了解提案结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "项目背景",
            "subtitle": "为什么需要这个项目",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "项目背景",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排说明项目背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'背景'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景主题（不超过35字）；sub_header 为背景说明；bullets 列出3-4条背景要点，每条不超过35字。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "需求分析",
            "subtitle": "要解决什么问题",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "card_grid",
            "title": "需求分析",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "列出4个主要需求或痛点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个需求/痛点，每个有 header（需求名称，不超过35字）和 body（一句话描述，不超过35字）。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "解决方案",
            "subtitle": "如何解决这个问题",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "解决方案",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "notes": "用图文混排说明解决方案",
            "filling_prompt": "必须填入真实内容：kicker 为'方案'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为方案名称（不超过35字）；sub_header 为方案简介；bullets 列出3-4条方案要点，每条不超过35字。"
        },
        {
            "index": 9,
            "type": "process_flow",
            "title": "实施路径",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "{阶段1}", "desc": "{阶段说明，不超过35字}"},
                {"num": "02", "title": "{阶段2}", "desc": "{阶段说明，不超过35字}"},
                {"num": "03", "title": "{阶段3}", "desc": "{阶段说明，不超过35字}"},
                {"num": "04", "title": "{阶段4}", "desc": "{阶段说明，不超过35字}"}
            ],
            "notes": "用流程图展示实施阶段",
            "filling_prompt": "必须填入真实内容：提供4个实施阶段，每个有 title（阶段名称，不超过35字）和 desc（阶段说明，不超过35字）。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "项目计划",
            "subtitle": "时间表与里程碑",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "timeline",
            "title": "项目里程碑",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "notes": "展示项目关键里程碑",
            "filling_prompt": "必须填入真实内容：提供4-5个里程碑节点，每个有具体时间、里程碑名称和交付物说明。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "预算估算",
            "subtitle": "资源需求",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "kpi_dashboard",
            "title": "预算概览",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "预算 · 费用明细",
            "notes": "展示预算分配",
            "filling_prompt": "必须填入真实内容：提供4个预算类别，每个有 value（金额）、label（类别名称）、delta（占比）、baseline（总预算）。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "06",
            "title": "预期效果",
            "subtitle": "项目价值",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 15,
            "type": "content_slide",
            "title": "预期收益",
            "content_type": "content_slide",
            "notes": "列出项目的预期收益",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：列出3-5条预期收益，每条有标题（不超过35字）和说明（不超过35字）。references 列出 URL。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心价值}",
                "02 {关键优势}",
                "03 {资源请求}"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁总结，明确资源请求",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心价值+关键优势+资源请求）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "提案理由要充分，数据说话",
        "方案要可行，避免过度承诺",
        "预算要合理，有明细有依据",
        "预期效果要可衡量可验证",
        "PPT设计正式稳重，体现专业性"
    ]
}
