TEMPLATE = {
    "name": "tech-intro",
    "name_cn": "技术介绍/科普",
    "description": "适合新技术介绍、行业科普、知识分享等场景。内容全面，从基础概念到应用实践，循序渐进，适合非技术受众。",
    "target_audience": "非技术人员、业务人员、管理层、科普受众",
    "typical_slides": 18,
    "typical_duration": "20-30分钟",
    "palette": "ocean_soft",
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
            "title": "{主题名称}",
            "subtitle": "{副标题}",
            "author": "{演讲者}",
            "date": "{日期}",
            "notes": "开场标题页，留白充足，标题字体有重量感",
            "filling_prompt": "必须填入真实内容：title 为本次演讲的主题名称（如'零代码平台技术介绍'），subtitle 为一句概括性副标题，author 为演讲者姓名或部门，date 为实际日期。禁止保留花括号占位符。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  什么是{主题}",
                "02  技术发展历程",
                "03  核心原理",
                "04  能力与特点",
                "05  行业应用",
                "06  未来展望"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "必须填入真实内容：items[0] 中的 {主题} 替换为本次演讲的实际主题名称，其余章节名称根据主题适配（如'什么是Kubernetes'→'什么是容器编排'等）。禁止保留花括号。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "什么是{主题}",
            "subtitle": "{章节副标题}",
            "filling_prompt": "必须填入真实内容：title 中的 {主题} 替换为本次演讲的实际主题名称。禁止保留花括号。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "{主题}的定义",
            "content_type": "content_slide",
            "content_plan": {
                "summary": "{一句话概括主题的本质}",
                "elements": [
                    {"type": "bullet_list", "items": []},
                    {"type": "callout", "text": ""}
                ]
            },
            "notes": "用通俗易懂的语言解释核心概念，避免过多专业术语",
            "filling_prompt": "必须填入真实内容：title 中的 {主题} 替换为本次主题；section_header 为一个小节标题（如'什么是Kubernetes'中的'核心定义'）；bullets 为3-4条对主题核心定义的要点，每条不超过20个中文字符，必须通俗易懂。禁止保留花括号占位符。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "规模与影响力",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "用卡片展示规模数据，每个指标用一句话描述，可接受"-"或"暂无数据"",
            "filling_prompt": "必须填入真实内容：提供4个与主题相关的指标，每个有 header（指标名称）和 body（一句话描述）。指标可以是：用户量/下载量、市场规模、技术指标（如性能提升）、社区活跃度等。如果某项数据确实无法获取，填'暂无公开数据'，不要虚构数字。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "技术发展历程",
            "subtitle": "从概念提出到成熟应用",
            "filling_prompt": "本章为技术发展历程页，固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "发展里程碑",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "notes": "时间轴展示关键技术节点，每个节点一句话",
            "filling_prompt": "必须填入真实内容：提供4-5个技术发展里程碑（年份+事件名称+一句话描述），如'2013年 Docker发布：容器技术正式诞生'。禁止虚构不存在的里程碑。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "核心原理",
            "subtitle": "技术实现的关键机制",
            "filling_prompt": "本章为核心原理页，固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "deep_dive",
            "title": "{核心概念}详解",
            "content_type": "deep_dive",
            "kicker": "详解 · {核心概念}",
            "lede": "{一句话概括核心内容的核心价值}",
            "left_column": {
                "key_points": [
                    "{要点1}",
                    "{要点2}",
                    "{要点3}",
                    "{要点4}",
                    "{要点5}"
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
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {核心概念} 替换为该章节要讲解的具体核心概念；title 同理；lede 为一句话说明核心价值；left_column.key_points 为该内容的核心要点（3-5条，每条不超过35字）；left_column.analysis 为2-3个深度分析维度；right_column.case_example 为具体案例说明（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。禁止保留花括号占位符。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "关键能力：{能力名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示关键能力，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'关键能力'；title 中的 {能力名称} 替换为具体能力名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为能力标题（如'智能弹性伸缩'，不超过35字）；sub_header 为能力简介（不超过35字）；bullets 列出3-4条能力价值或应用效果，每条不超过35字。references 列出 URL。禁止使用匿名实体；禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "能力与特点",
            "subtitle": "{主题}的核心优势",
            "filling_prompt": "必须填入真实内容：subtitle 中的 {主题} 替换为本次演讲的实际主题名称。"
        },
        {
            "index": 12,
            "type": "content_slide",
            "title": "核心能力",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "notes": "4个核心能力卡片，每个一句话说明",
            "filling_prompt": "必须填入真实内容：提供4个核心能力，每个能力有 header（能力名称，不超过35字）和 body（一句话描述该能力的核心价值，不超过35字）。能力名称要具体，如'弹性伸缩'、'灰度发布'、'故障自愈'，不能是'能力1'、'能力2'这类占位符。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "行业应用",
            "subtitle": "{主题}正在改变各行各业",
            "filling_prompt": "必须填入真实内容：subtitle 中的 {主题} 替换为本次演讲的实际主题名称。"
        },
        {
            "index": 14,
            "type": "image_text",
            "title": "行业案例：{行业名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示行业应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {行业} 替换为具体行业（如'金融'、'电商'、'医疗'）；title 中的 {行业名称} 替换为具体行业名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX银行智能风控系统'，不超过35字）；sub_header 为合作项目或应用名称（不超过35字）；bullets 列出3-4条应用效果或客户评价，每条不超过35字。references 列出 web_search 获取的 URL（至少2个）。禁止使用'某公司''某行业'等匿名实体；禁止虚构数据。"
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "应用效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 应用效果",
            "kpis": [
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{效果说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"}
            ],
            "notes": "展示4个核心效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心效果指标，每个 KPI 有 value（具体数值）、label（效果说明）、delta（变化趋势，如'↑ 30%'或'↓ 50%'）、baseline（对比基准，如'vs 传统方案'）。指标要具体且有代表性，如'处理效率提升 3 倍'、'故障恢复时间从 2 小时缩短至 5 分钟'。references 列出 web_search 获取的 URL（至少2个）。禁止保留花括号；禁止虚构数据。"
        },
        {
            "index": 16,
            "type": "section_divider",
            "number": "06",
            "title": "未来展望",
            "subtitle": "技术趋势与挑战",
            "filling_prompt": "本章为未来展望页，固定内容，无需额外填充。"
        },
        {
            "index": 17,
            "type": "content_slide",
            "title": "发展趋势",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "{趋势1}", "desc": "{一句话描述}"},
                {"num": "02", "title": "{趋势2}", "desc": "{一句话描述}"},
                {"num": "03", "title": "{趋势3}", "desc": "{一句话描述}"},
                {"num": "04", "title": "{趋势4}", "desc": "{一句话描述}"},
                {"num": "05", "title": "{趋势5}", "desc": "{一句话描述}"},
                {"num": "06", "title": "{趋势6}", "desc": "{一句话描述}"}
            ],
            "notes": "6个发展趋势，zigzag排列，每个步骤不超过30字",
            "filling_prompt": "必须填入真实内容：提供6个该技术领域的未来发展趋势，每条有 title（趋势名称，如'AIOps自动化运维'）和 desc（一句话描述，不超过30字）。趋势要具体且基于行业观察，禁止虚构。禁止保留花括号。"
        },
        {
            "index": 18,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心要点1}",
                "02 {核心要点2}",
                "03 {核心要点3}",
                "04 {核心要点4}"
            ],
            "thank_you": "感谢聆听",
            "contact": "{联系方式}",
            "filling_prompt": "必须填入真实内容：key_points 提供4个核心要点（每条30字以内，精炼概括本次演讲的核心内容）；contact 填写真实联系方式（如邮箱、微信号等）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "技术介绍要通俗易懂，避免过度专业化",
        "多用大数字展示规模和效果",
        "案例要有具体数据和真实来源",
        "保持章节清晰，循序渐进",
        "结尾展望要结合实际，给出可行方向"
    ]
}
