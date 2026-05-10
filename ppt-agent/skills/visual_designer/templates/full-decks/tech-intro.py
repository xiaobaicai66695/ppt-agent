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
            "title": "人工智能大模型技术介绍",
            "subtitle": "从GPT到AGI：AI技术演进与应用实践",
            "author": "张伟 | 技术研发部",
            "date": "2025年3月15日",
            "notes": "开场标题页，留白充足，标题字体有重量感",
            "filling_prompt": "必须填入真实内容：title 为本次演讲的主题名称（如'零代码平台技术介绍'），subtitle 为一句概括性副标题，author 为演讲者姓名或部门，date 为实际日期。禁止保留花括号占位符。",
            "visual_suggestions": "可添加抽象的数据流或网络图形作为背景装饰元素"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  什么是人工智能大模型",
                "02  技术发展历程",
                "03  核心原理",
                "04  能力与特点",
                "05  行业应用",
                "06  未来展望"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "必须填入真实内容：items[0] 中的 {主题} 替换为本次演讲的实际主题名称，其余章节名称根据主题适配（如'什么是Kubernetes'→'什么是容器编排'等）。禁止保留花括号。",
            "timing_hint": "约30秒，可快速翻过"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "什么是人工智能大模型",
            "subtitle": "AI新范式的崛起",
            "filling_prompt": "必须填入真实内容：title 中的 {主题} 替换为本次演讲的实际主题名称。",
            "design_notes": "章节分隔页使用大字号标题，配色与主色调一致，营造章节仪式感"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "人工智能大模型的定义与本质",
            "content_type": "example_detail",
            "kicker": "实例 · 人工智能大模型",
            "lede": "人工智能大模型通过海量数据训练，获得了接近人类的语言理解和生成能力",
            "context_block": "随着ChatGPT的爆火，全球掀起了大模型热潮。国内各大科技公司纷纷推出自己的大模型产品，如文心一言、通义千问等。用户数量在短短几个月内突破亿级。",
            "solution_block": "大模型的本质是通过深度学习技术，让计算机从海量文本数据中学习语言规律和知识。当用户输入问题时，模型会根据学到的知识生成相应的回答。就像人类通过阅读书籍学习知识一样，大模型通过'阅读'互联网上的数据来获得智能。",
            "metrics": [
                {"value": "GPT-4", "label": "最新模型版本", "trend": "↑ 性能提升显著"},
                {"value": "1.8万亿", "label": "模型参数规模", "trend": "↑ 较上代增长10倍"},
                {"value": "50+", "label": "支持语言数", "trend": "↑ 覆盖主流语言"}
            ],
            "takeaway": "启示：大模型代表着AI从专用走向通用的重要突破",
            "notes": "通过具体实例（可虚构但要合理）来解释主题的定义和本质，增强理解",
            "filling_prompt": "必须填入真实内容：kicker 中的 {主题} 替换为本次演讲的实际主题名称；lede 为一句话说明该主题的核心价值或影响力；context_block 描述该主题出现的行业背景或问题语境（1-2句话）；solution_block 用通俗语言描述该主题的核心原理或实现方式（2-3句话）；metrics_grid 提供3个具体指标（如用户量/性能提升/市场增速等），每个有 value（具体数字+单位）、label（指标名称）、trend（趋势方向）；takeaway 用一句话总结启示。禁止保留花括号占位符。",
            "speaking_notes": "讲解时可用'就像人类通过阅读学习知识一样'这样的类比来帮助理解",
            "transition_phrase": "那么，大模型是如何发展到今天的呢？让我们回顾一下技术演进历程。"
        },
        {
            "index": 5,
            "type": "kpi_dashboard",
            "title": "规模与影响力",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 规模与影响力",
            "kpis": [
                {"value": "100亿+", "label": "全球AI市场估值（美元）", "delta": "↑ 38% YoY", "baseline": "vs 2023年"},
                {"value": "1亿+", "label": "ChatGPT月活用户数", "delta": "↑ 持续增长", "baseline": "史上增长最快应用"},
                {"value": "300+", "label": "全球大模型数量", "delta": "↑ 新增加速", "baseline": "截至2025年初"},
                {"value": "60%", "label": "500强企业已部署AI", "delta": "↑ 显著提升", "baseline": "vs 2022年20%"}
            ],
            "notes": "用指标卡片展示规模数据，每个指标有具体数字",
            "filling_prompt": "必须填入真实内容（通过 web_search 获取权威数据，至少2个URL）：提供4个与主题相关的规模/影响力指标，每个有 value（具体数字+单位，如'12亿用户'、'4200亿美元市场'）、label（指标名称）、delta（变化趋势，如'↑38% YoY'）、baseline（对比基准，如'vs 2023年'）。指标可以是：用户量/下载量、市场规模、技术指标（如性能提升幅度）、社区活跃度等。如果某项数据确实无法获取，填'暂无公开数据'，不要虚构数字。references 列出 web_search 获取的 URL。禁止保留花括号占位符。",
            "data_source_tips": "数据来源建议：Gartner报告、IDC报告、企业财报、第三方调研机构数据",
            "visual_encouragement": "配合地图热力图或增长曲线图效果更佳"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "技术发展历程",
            "subtitle": "从概念提出到成熟应用",
            "filling_prompt": "本章为技术发展历程页，固定内容，无需额外填充。",
            "design_notes": "可用时间轴作为章节过渡的视觉元素"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "发展里程碑",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "timeline_items": [
                {"year": "2017年", "event": "Transformer架构提出", "desc": "谷歌发表《Attention is All You Need》论文"},
                {"year": "2018年", "event": "BERT发布", "desc": "谷歌推出预训练语言模型BERT"},
                {"year": "2020年", "event": "GPT-3诞生", "desc": "1750亿参数，few-shot学习能力"},
                {"year": "2022年11月", "event": "ChatGPT发布", "desc": "OpenAI推出对话式AI应用"},
                {"year": "2024年", "event": "GPT-4o发布", "desc": "多模态实时交互能力"}
            ],
            "notes": "时间轴展示关键技术节点，每个节点一句话",
            "filling_prompt": "必须填入真实内容：提供4-5个技术发展里程碑（年份+事件名称+一句话描述），如'2013年 Docker发布：容器技术正式诞生'。禁止虚构不存在的里程碑。",
            "visual_suggestions": "时间轴使用渐变色或图标区分不同时期",
            "speaking_tip": "讲到重要节点时可稍作停顿，让听众消化"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "核心原理",
            "subtitle": "技术实现的关键机制",
            "filling_prompt": "本章为核心原理页，固定内容，无需额外填充。",
            "transition_phrase": "接下来，让我们深入了解一下大模型的核心技术原理。"
        },
        {
            "index": 9,
            "type": "deep_dive",
            "title": "Transformer架构详解",
            "content_type": "deep_dive",
            "kicker": "详解 · Transformer架构",
            "lede": "Transformer是现代大模型的基础架构，革新了自然语言处理",
            "left_column": {
                "key_points": [
                    "自注意力机制：让模型'关注'输入的每个部分",
                    "并行计算：大幅提升训练效率",
                    "多头注意力：捕获不同类型的语义关系",
                    "位置编码：让模型理解词序信息",
                    "残差连接：便于深层网络训练"
                ],
                "analysis": [
                    "相比RNN：克服了长距离依赖问题",
                    "相比CNN：更好地捕捉全局依赖关系"
                ]
            },
            "right_column": {
                "case_example": [
                    "输入：'北京是中国的首都'",
                    "模型理解：'北京'与'首都'的强关联",
                    "注意力权重显示：'首都'高度关注'北京'",
                    "输出：准确识别实体关系"
                ],
                "data_evidence": [
                    "BERT-base: 1.1亿参数",
                    "GPT-3: 1750亿参数",
                    "训练数据: 数万亿token"
                ]
            },
            "notes": "双栏深入展开，左栏放核心要点和分析，右栏放案例和数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {核心概念} 替换为该章节要讲解的具体核心概念；title 同理；lede 为一句话说明核心价值；left_column.key_points 为该内容的核心要点（3-5条，每条不超过35字）；left_column.analysis 为2-3个深度分析维度；right_column.case_example 为具体案例说明（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。禁止保留花括号占位符。",
            "technical_level": "advanced",
            "speaking_tips": "如听众为非技术背景，可跳过技术细节，侧重讲解原理"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "关键能力：智能问答",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "关键能力",
            "header": "自然语言交互能力",
            "sub_header": "像与人对话一样与AI交流",
            "paragraph": "大模型最核心的能力之一是自然语言问答。用户可以用日常对话的方式向AI提问，AI能够理解问题的意图，并给出准确、有帮助的回答。与传统搜索引擎不同，大模型不仅能检索信息，还能理解和整合信息，用自然语言组织出完整的答案。例如，在客服场景中，大模型可以理解用户的复杂问题描述，自动生成个性化的解决方案，大幅提升客服效率和用户满意度。实测数据显示，引入大模型后，客服响应时间缩短60%，问题解决率提升45%。",
            "notes": "用图文混排展示关键能力，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'关键能力'；title 中的 {能力名称} 替换为具体能力名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为能力标题（如'智能弹性伸缩'，不超过35字）；sub_header 为能力简介（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该能力的技术原理、核心优势和典型应用场景，包含具体数据或使用效果，禁止罗列要点。references 列出 URL。禁止使用匿名实体；禁止虚构数据。",
            "demo_ideas": "可现场演示一个简单的问答示例",
            "use_case_examples": ["智能客服", "个人助手", "教育辅导", "代码助手"]
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "能力与特点",
            "subtitle": "人工智能大模型的核心优势",
            "filling_prompt": "必须填入真实内容：subtitle 中的 {主题} 替换为本次演讲的实际主题名称。"
        },
        {
            "index": 12,
            "type": "content_slide",
            "title": "核心能力",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {
                    "header": "涌现能力",
                    "body": "当模型规模达到一定量级后，会涌现出意想不到的能力，如逻辑推理、代码生成等，这些能力在小模型中几乎不存在"
                },
                {
                    "header": "泛化能力",
                    "body": "在一种任务上学到的知识可以迁移到其他相关任务，减少重复训练，大幅降低AI应用开发成本"
                },
                {
                    "header": "多模态能力",
                    "body": "最新的大模型支持文本、图像、音频、视频等多种模态的理解和生成，实现更丰富的人机交互"
                },
                {
                    "header": "持续学习",
                    "body": "通过人类反馈学习(RLHF)技术，大模型能够持续从用户反馈中学习，不断优化回答质量"
                }
            ],
            "notes": "4个核心能力卡片，每个一句话说明",
            "filling_prompt": "必须填入真实内容：提供4个核心能力，每个能力有 header（能力名称，不超过35字）和 body（详细描述该能力的核心价值和典型应用场景，100-120字，包含具体效果或数据）。能力名称要具体，如'弹性伸缩'、'灰度发布'、'故障自愈'，不能是'能力1'、'能力2'这类占位符。",
            "visual_suggestions": "每个卡片配以相应的图标，增强识别度"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "05",
            "title": "行业应用",
            "subtitle": "人工智能大模型正在改变各行各业",
            "filling_prompt": "必须填入真实内容：subtitle 中的 {主题} 替换为本次演讲的实际主题名称。"
        },
        {
            "index": 14,
            "type": "image_text",
            "title": "行业案例：金融行业",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "金融行业",
            "header": "智能投顾与风控",
            "sub_header": "招商银行AI助手案例",
            "paragraph": "招商银行推出的AI投顾助手，利用大模型技术为客户提供个性化的投资建议。用户只需描述自己的风险偏好和投资目标，AI就能自动分析市场数据，生成量身定制的投资组合方案。该系统上线半年内，服务客户超过500万人次，客户满意度达92%。与传统投顾相比，AI投顾的服务成本降低了70%，响应速度提升了10倍，真正实现了普惠金融的目标。",
            "references": [
                "https://finance.cmbchina.com/",
                "https://www.36kr.com/p/xxxxx"
            ],
            "notes": "用图文混排展示行业应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {行业} 替换为具体行业（如'金融'、'电商'、'医疗'）；title 中的 {行业名称} 替换为具体行业名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX银行智能风控系统'，不超过35字）；sub_header 为合作项目或应用名称（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该行业案例的背景、实施过程、应用效果和客户收益，包含具体数据，禁止罗列要点。references 列出 web_search 获取的 URL（至少2个）。禁止使用'某公司''某行业'等匿名实体；禁止虚构数据。",
            "more_case_examples": ["智能客服", "风险评估", "欺诈检测", "文档处理"]
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "应用效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 应用效果",
            "kpis": [
                {"value": "↑ 300%", "label": "客服效率提升", "delta": "↑ 3倍", "baseline": "vs 传统客服"},
                {"value": "↓ 60%", "label": "人工成本降低", "delta": "↓ 60%", "baseline": "vs 实施前"},
                {"value": "↑ 45%", "label": "问题解决率提升", "delta": "↑ 45%", "baseline": "vs 规则引擎"},
                {"value": "< 5秒", "label": "平均响应时间", "delta": "↓ 80%", "baseline": "vs 人工客服"}
            ],
            "notes": "展示4个核心效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心效果指标，每个 KPI 有 value（具体数值）、label（效果说明）、delta（变化趋势，如'↑ 30%'或'↓ 50%'）、baseline（对比基准，如'vs 传统方案'）。指标要具体且有代表性，如'处理效率提升 3 倍'、'故障恢复时间从 2 小时缩短至 5 分钟'。references 列出 web_search 获取的 URL（至少2个）。禁止保留花括号；禁止虚构数据。",
            "visual_tip": "配合对比柱状图效果更直观"
        },
        {
            "index": 16,
            "type": "section_divider",
            "number": "06",
            "title": "未来展望",
            "subtitle": "技术趋势与挑战",
            "filling_prompt": "本章为未来展望页，固定内容，无需额外填充。",
            "transition_phrase": "最后，让我们展望一下大模型技术的未来发展方向。"
        },
        {
            "index": 17,
            "type": "content_slide",
            "title": "发展趋势",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "AGI探索", "desc": "迈向通用人工智能的终极目标"},
                {"num": "02", "title": "端侧部署", "desc": "模型小型化，本地设备运行"},
                {"num": "03", "title": "多模态融合", "desc": "文本、图像、音频、视频统一处理"},
                {"num": "04", "title": "行业垂直化", "desc": "专注特定领域的专业大模型"},
                {"num": "05", "title": "AI Agent", "desc": "自主规划执行复杂任务的AI助手"},
                {"num": "06", "title": "安全可控", "desc": "大模型安全与对齐技术深化"}
            ],
            "notes": "6个发展趋势，zigzag排列，每个步骤不超过30字",
            "filling_prompt": "必须填入真实内容：提供6个该技术领域的未来发展趋势，每条有 title（趋势名称，如'AIOps自动化运维'）和 desc（一句话描述，不超过30字）。趋势要具体且基于行业观察，禁止虚构。禁止保留花括号。",
            "visual_style": "使用箭头或流程图连接各趋势点"
        },
        {
            "index": 18,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 大模型代表着AI从专用走向通用的重要突破",
                "02 Transformer架构是现代大模型的核心基础",
                "03 大模型正在金融、医疗、教育等领域广泛应用",
                "04 AGI是长期目标，安全可控是发展的前提"
            ],
            "thank_you": "感谢聆听",
            "contact": "联系方式：tech@company.com | 扫码入群交流",
            "filling_prompt": "必须填入真实内容：key_points 提供4个核心要点（每条30字以内，精炼概括本次演讲的核心内容）；contact 填写真实联系方式（如邮箱、微信号等）。禁止保留花括号。",
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
