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
    "scene_guidance": {
        "opening_hook": "用一个令人惊讶的数据或现象开场，引发听众好奇心",
        "key_moments": [
            "定义讲解时：使用生活化类比",
            "数据展示时：强调对比和趋势",
            "案例分享时：讲述具体故事",
            "展望未来时：描绘愿景画面"
        ],
        "closing_strength": "总结时呼应开场的数据或现象，形成闭环"
    },
    "audience_considerations": {
        "avoid_jargon": "尽量避免技术术语，如必须使用需立即解释",
        "use_analogies": "多用生活化类比帮助理解抽象概念",
        "visual_aids": "多用图表、流程图等可视化形式展示复杂内容",
        "interactive_hints": "适当设置互动问题，引导听众思考"
    },
    "content_depth_levels": {
        "beginner": "面向完全不了解该技术的人群，用最基础的概念解释",
        "intermediate": "面向有基本了解的人群，深入讲解核心原理和应用",
        "advanced": "面向有一定基础的人群，分享最新发展和深度思考"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "演讲主题名称",
            "subtitle": "一句概括性副标题",
            "author": "演讲者姓名 | 部门",
            "date": "实际日期",
            "notes": "开场标题页，留白充足，标题字体有重量感",
            "filling_prompt": "必须填入真实内容：title 为本次演讲的主题名称，subtitle 为一句概括性副标题，author 为演讲者姓名或部门，date 为实际日期。禁止保留花括号占位符。",
            "visual_suggestions": "可添加抽象的数据流或网络图形作为背景装饰元素"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  章节1",
                "02  章节2",
                "03  章节3",
                "04  章节4",
                "05  章节5",
                "06  章节6"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "必须填入真实内容：items 中的章节名称根据演讲主题适配。禁止保留花括号。",
            "timing_hint": "约30秒，可快速翻过"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "章节标题",
            "subtitle": "章节副标题",
            "filling_prompt": "必须填入真实内容：title 为本次演讲的章节标题。",
            "design_notes": "章节分隔页使用大字号标题，配色与主色调一致，营造章节仪式感"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "主题定义与本质",
            "content_type": "example_detail",
            "kicker": "实例 · 主题定义",
            "lede": "一句话说明该主题的核心价值或影响力",
            "context_block": "描述该主题出现的行业背景或问题语境（1-2句话）。",
            "solution_block": "用通俗语言描述该主题的核心原理或实现方式（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "指标1", "trend": "趋势方向"},
                {"value": "数字", "label": "指标2", "trend": "趋势方向"},
                {"value": "数字", "label": "指标3", "trend": "趋势方向"}
            ],
            "takeaway": "一句话总结启示。",
            "notes": "通过具体实例来解释主题的定义和本质，增强理解",
            "filling_prompt": "必须填入真实内容：lede 为一句话说明该主题的核心价值或影响力；context_block 描述该主题出现的行业背景或问题语境（1-2句话）；solution_block 用通俗语言描述该主题的核心原理或实现方式（2-3句话）；metrics_grid 提供3个具体指标，每个有 value（具体数字+单位）、label（指标名称）、trend（趋势方向）；takeaway 用一句话总结启示。禁止保留花括号占位符。"
        },
        {
            "index": 5,
            "type": "kpi_dashboard",
            "title": "规模与影响力",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 规模与影响力",
            "kpis": [
                {"value": "数字+单位", "label": "指标1", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字+单位", "label": "指标2", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字+单位", "label": "指标3", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字+单位", "label": "指标4", "delta": "变化趋势", "baseline": "对比基准"}
            ],
            "notes": "用指标卡片展示规模数据，每个指标有具体数字",
            "filling_prompt": "必须填入真实内容（通过 web_search 获取权威数据，至少2个URL）：提供4个与主题相关的规模/影响力指标，每个有 value（具体数字+单位）、label（指标名称）、delta（变化趋势）、baseline（对比基准）。指标可以是：用户量/下载量、市场规模、技术指标（如性能提升幅度）、社区活跃度等。如果某项数据确实无法获取，填'暂无公开数据'，不要虚构数字。references 列出 URL。禁止保留花括号占位符。",
            "data_source_tips": "数据来源建议：Gartner报告、IDC报告、企业财报、第三方调研机构数据",
            "visual_encouragement": "配合地图热力图或增长曲线图效果更佳"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "章节标题",
            "subtitle": "章节副标题",
            "filling_prompt": "固定内容，无需额外填充。",
            "design_notes": "可用时间轴作为章节过渡的视觉元素"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "发展里程碑",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "timeline_items": [
                {"year": "年份", "event": "事件名称", "desc": "事件描述"},
                {"year": "年份", "event": "事件名称", "desc": "事件描述"},
                {"year": "年份", "event": "事件名称", "desc": "事件描述"},
                {"year": "年份", "event": "事件名称", "desc": "事件描述"}
            ],
            "notes": "时间轴展示关键技术节点，每个节点一句话",
            "filling_prompt": "必须填入真实内容：提供4-5个技术发展里程碑（年份+事件名称+一句话描述）。禁止虚构不存在的里程碑。",
            "visual_suggestions": "时间轴使用渐变色或图标区分不同时期"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "章节标题",
            "subtitle": "章节副标题",
            "filling_prompt": "固定内容，无需额外填充。",
            "transition_phrase": "接下来，让我们深入了解一下核心内容。"
        },
        {
            "index": 9,
            "type": "deep_dive",
            "title": "核心内容详解",
            "content_type": "deep_dive",
            "kicker": "详解 · 核心内容",
            "lede": "一句话说明核心价值",
            "left_column": {
                "key_points": [
                    "核心要点1",
                    "核心要点2",
                    "核心要点3",
                    "核心要点4",
                    "核心要点5"
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
                    "数据指标1",
                    "数据指标2",
                    "数据指标3"
                ]
            },
            "notes": "双栏深入展开，左栏放核心要点和分析，右栏放案例和数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_column.key_points 为核心要点（3-5条，每条不超过35字）；left_column.analysis 为2-3个深度分析维度；right_column.case_example 为具体案例说明（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。禁止保留花括号占位符。",
            "technical_level": "advanced",
            "speaking_tips": "如听众为非技术背景，可跳过技术细节，侧重讲解原理"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "关键能力",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "关键能力",
            "header": "能力标题",
            "sub_header": "能力简介",
            "paragraph": "详细描述该能力的技术原理、核心优势和典型应用场景，包含具体数据或使用效果，禁止罗列要点。",
            "notes": "用图文混排展示关键能力，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为能力标题（不超过35字）；sub_header 为能力简介（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该能力的技术原理、核心优势和典型应用场景，包含具体数据或使用效果，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "章节标题",
            "subtitle": "章节副标题",
            "filling_prompt": "必须填入真实内容：subtitle 中的章节主题根据演讲内容适配。"
        },
        {
            "index": 12,
            "type": "content_slide",
            "title": "核心能力",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "能力名称（不超过35字）", "body": "详细描述该能力的核心价值和典型应用场景（100-120字），包含具体效果或数据。"},
                {"header": "能力名称（不超过35字）", "body": "详细描述该能力的核心价值和典型应用场景（100-120字），包含具体效果或数据。"},
                {"header": "能力名称（不超过35字）", "body": "详细描述该能力的核心价值和典型应用场景（100-120字），包含具体效果或数据。"},
                {"header": "能力名称（不超过35字）", "body": "详细描述该能力的核心价值和典型应用场景（100-120字），包含具体效果或数据。"}
            ],
            "notes": "4个核心能力卡片，每个一句话说明",
            "filling_prompt": "必须填入真实内容：提供4个核心能力，每个能力有 header（能力名称，不超过35字）和 body（详细描述该能力的核心价值和典型应用场景，100-120字，包含具体效果或数据）。能力名称要具体。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "章节标题",
            "subtitle": "章节副标题",
            "filling_prompt": "必须填入真实内容：subtitle 中的章节主题根据演讲内容适配。"
        },
        {
            "index": 14,
            "type": "image_text",
            "title": "行业案例",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "行业",
            "header": "案例标题",
            "sub_header": "合作项目名称",
            "paragraph": "详细描述该行业案例的背景、实施过程、应用效果和客户收益，包含具体数据，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示行业应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为合作项目名称（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该行业案例的背景、实施过程、应用效果和客户收益，包含具体数据，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "应用效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 应用效果",
            "kpis": [
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"}
            ],
            "notes": "展示4个核心效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心效果指标，每个 KPI 有 value（具体数值）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。指标要具体且有代表性。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 16,
            "type": "section_divider",
            "number": "06",
            "title": "章节标题",
            "subtitle": "章节副标题",
            "filling_prompt": "固定内容，无需额外填充。",
            "transition_phrase": "最后，让我们展望一下未来的发展方向。"
        },
        {
            "index": 17,
            "type": "content_slide",
            "title": "发展趋势",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "02", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "03", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "04", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "05", "title": "趋势名称", "desc": "趋势描述（不超过30字）"},
                {"num": "06", "title": "趋势名称", "desc": "趋势描述（不超过30字）"}
            ],
            "notes": "6个发展趋势，zigzag排列，每个步骤不超过30字",
            "filling_prompt": "必须填入真实内容：提供6个该技术领域的未来发展趋势，每条有 title（趋势名称）和 desc（一句话描述，不超过30字）。趋势要具体且基于行业观察，禁止虚构。禁止保留花括号。"
        },
        {
            "index": 18,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心要点1",
                "02 核心要点2",
                "03 核心要点3",
                "04 核心要点4"
            ],
            "thank_you": "感谢聆听",
            "contact": "联系方式：邮箱 | 扫码入群交流",
            "filling_prompt": "必须填入真实内容：key_points 提供4个核心要点（每条30字以内，精炼概括本次演讲的核心内容）；contact 填写真实联系方式。禁止保留花括号。",
            "q_and_a_hint": "建议预留5-10分钟回答听众问题"
        }
    ],
    "design_tips": [
        "技术介绍要通俗易懂，避免过度专业化",
        "多用大数字展示规模和效果",
        "案例要有具体数据和真实来源",
        "保持章节清晰，循序渐进",
        "结尾展望要结合实际，给出可行方向",
        "开场用一个令人惊讶的数据或现象引发好奇心",
        "用生活化类比帮助理解抽象概念",
        "每章之间使用过渡语，形成连贯叙事",
        "技术细节可选，避免在非专业受众面前过度展开"
    ],
    "presentation_flow": {
        "opening": {
            "duration": "1-2分钟",
            "goal": "建立悬念，引发兴趣",
            "tip": "用一个令人惊讶的数据、现象或问题开场"
        },
        "body": {
            "duration": "15-20分钟",
            "goal": "层层递进，讲解核心内容",
            "tip": "每个章节结尾回顾要点，帮助听众消化"
        },
        "closing": {
            "duration": "2-3分钟",
            "goal": "总结升华，呼应开头",
            "tip": "总结要点，展望未来，预留Q&A时间"
        }
    }
}
