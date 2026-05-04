TEMPLATE = {
    "name": "product-intro",
    "name_cn": "产品介绍",
    "description": "适合产品介绍、客户演示、功能展示等场景。突出价值，展示功能，增强信任。",
    "target_audience": "客户、合作伙伴、销售团队",
    "typical_slides": 12,
    "typical_duration": "15-20分钟",
    "palette": "warm_terracotta",
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
            "title": "{产品名称}",
            "subtitle": "{产品定位/一句话价值主张}",
            "author": "{主讲人}",
            "date": "{日期}",
            "notes": "标题页醒目，展示产品名称和价值主张",
            "filling_prompt": "必须填入真实内容：title 为产品名称，subtitle 为产品定位或一句话价值主张，author 为主讲人姓名，date 为演示日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  产品概述",
                "02  核心功能",
                "03  应用场景",
                "04  客户案例",
                "05  产品优势"
            ],
            "notes": "让听众了解产品介绍结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "产品概述",
            "subtitle": "产品定位与价值",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "产品概览",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示产品主视觉和核心价值",
            "filling_prompt": "必须填入真实内容：kicker 为产品领域标签；title 为产品名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为一句话核心价值主张（不超过35字）；sub_header 为产品定位说明；bullets 列出3-4条产品核心价值点，每条不超过35字。"
        },
        {
            "index": 5,
            "type": "kpi_dashboard",
            "title": "产品数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 产品表现",
            "notes": "用数据展示产品市场表现",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个产品相关指标（如用户数、功能数、覆盖率、客户满意度等），每个有 value、label、delta、baseline。references 列出 URL。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "核心功能",
            "subtitle": "主要功能模块",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "核心功能",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "展示4个核心功能模块",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称，不超过35字）和 body（一句话描述功能价值，不超过35字）。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "功能详解：{功能名称}",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "notes": "用图文混排详解一个核心功能",
            "filling_prompt": "必须填入真实内容：kicker 为'功能详解'；title 中的 {功能名称} 替换为具体功能；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为功能名称（不超过35字）；sub_header 为功能简介；bullets 列出3-4条功能特点或使用步骤，每条不超过35字。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "应用场景",
            "subtitle": "典型使用场景",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "典型应用场景",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示应用场景",
            "filling_prompt": "必须填入真实内容：kicker 为'应用场景'；title 为场景名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为场景标题（不超过35字）；sub_header 为场景描述；bullets 列出3-4条该场景下的使用要点，每条不超过35字。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "客户案例",
            "subtitle": "成功案例展示",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 12,
            "type": "image_text",
            "title": "客户案例：{客户名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示一个成功案例",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {行业} 替换为客户所在行业；title 中的 {客户名称} 替换为真实客户名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为客户名称（不超过35字）；sub_header 为合作项目名称；bullets 列出3-4条合作成果或客户评价，每条不超过35字。references 列出 URL。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "产品优势",
            "subtitle": "为什么选择我们",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 14,
            "type": "content_slide",
            "title": "核心优势",
            "content_type": "content_slide",
            "notes": "总结产品核心竞争优势",
            "filling_prompt": "必须填入真实内容：列出4-5条核心竞争优势，每条有标题（不超过35字）和说明（一句话，不超过35字）。"
        },
        {
            "index": 15,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心价值}",
                "02 {关键优势}",
                "03 {联系方式/下一步}"
            ],
            "thank_you": "感谢聆听！",
            "notes": "简洁有力的结尾，附上联系方式",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心价值+关键优势+联系方式/下一步）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "产品介绍要突出价值而非功能",
        "案例要真实可信，注明来源",
        "图文混排增强说服力",
        "数据说话，展示市场认可",
        "结尾要有明确的行动号召"
    ]
}
