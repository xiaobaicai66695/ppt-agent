TEMPLATE = {
    "name": "tech-intro",
    "name_cn": "技术介绍/科普",
    "description": "适合新技术介绍、行业科普、知识分享等场景。内容全面，从基础概念到应用实践，循序渐进，适合非技术受众。",
    "target_audience": "非技术人员、业务人员、管理层、科普受众",
    "typical_slides": 20,
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
            "title": "大模型时代：AI的新纪元",
            "subtitle": "从GPT到AGI，AI如何重塑我们的世界",
            "author": "张明远 | 产品技术部",
            "date": "2026年5月",
            "notes": "开场标题页，留白充足，标题字体有重量感，背景可用抽象数据流图形装饰",
            "filling_prompt": "本模板已预填真实内容：title 为'大模型时代：AI的新纪元'，subtitle 为'从GPT到AGI，AI如何重塑我们的世界'，author 为演讲者姓名或部门，date 为实际日期。如需修改主题名称，请替换为本次演讲的实际主题。",
            "visual_suggestions": "可添加抽象的数据流或网络图形作为背景装饰元素"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "演讲大纲",
            "kicker": "目录",
            "items": [
                "01  什么是大模型",
                "02  技术演进历程",
                "03  核心能力解析",
                "04  行业应用案例",
                "05  未来发展趋势"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "本模板已预填真实内容：items 中的章节名称基于大模型技术介绍主题适配。如需修改章节名称，请替换为本次演讲的实际章节名称。",
            "timing_hint": "约30秒，可快速翻过"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "什么是大模型",
            "subtitle": "揭开人工智能的核心奥秘",
            "filling_prompt": "本模板已预填真实内容：number 为章节编号'01'，title 为'什么是大模型'，subtitle 为'揭开人工智能的核心奥秘'。如需修改章节主题，请替换为本次演讲的实际章节名称。",
            "design_notes": "章节分隔页使用大字号标题，配色与主色调一致，营造章节仪式感"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "大模型的核心定义",
            "content_type": "example_detail",
            "kicker": "实例 · 主题定义",
            "lede": "大模型就像一个读遍万卷书的孩子，经过海量数据的训练，学会了理解和生成人类语言",
            "context_block": "随着互联网数据的爆发式增长和GPU算力的大幅提升，2017年Transformer架构的诞生让训练超大规模语言模型成为可能，全球科技巨头纷纷投入大模型研发竞赛。",
            "solution_block": "大模型的核心原理是将海量文本数据压缩到一个巨大的神经网络中，通过自监督学习让模型学会预测下一个词。当模型规模足够大时，涌现出了令人惊讶的推理和泛化能力——就像孩子长大后突然'开窍'了一样。",
            "metrics": [
                {"value": "1750亿", "label": "GPT-3参数规模", "trend": "行业标杆"},
                {"value": "1.8万亿", "label": "GPT-4参数规模", "trend": "持续增长"},
                {"value": "100万亿", "label": "人脑突触数量", "trend": "对标参考"}
            ],
            "takeaway": "大模型的本质是用海量数据训练的超大神经网络，规模是能力涌现的关键。",
            "notes": "通过生活类比解释大模型本质",
            "filling_prompt": "本模板已预填关于大模型的真实内容。如需修改为其他技术主题，请替换 lede 为生活化类比语句，context_block 为技术出现的行业背景（2-3句话），solution_block 为通俗原理说明（3-4句话），metrics 为3个具体技术指标，takeaway 为一句话启示。"
        },
        {
            "index": 5,
            "type": "kpi_dashboard",
            "title": "规模与影响力",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 规模与影响力",
            "kpis": [
                {"value": "1亿+", "label": "ChatGPT月活用户", "delta": "+50%", "baseline": "2023年1月"},
                {"value": "1000亿美元", "label": "大模型市场预估规模", "delta": "2030年达峰", "baseline": "2023年150亿"},
                {"value": "500+", "label": "全球大模型数量", "delta": "持续增长", "baseline": "2023年3月"},
                {"value": "300万+", "label": "AI相关论文年产量", "delta": "+15%", "baseline": "2022年"}
            ],
            "notes": "用指标卡片展示规模数据，每个指标有具体数字和对比",
            "filling_prompt": "本模板已预填关于大模型的真实数据。如需修改为其他技术主题，请通过 web_search 获取权威数据（至少2个URL），替换 kpis 中的4个指标，每个有 value、label、delta、baseline。如某项数据确实无法获取，填'暂无公开数据'，不要虚构数字。references 列出 URL。",
            "data_source_tips": "数据来源建议：Gartner报告、IDC报告、企业财报、第三方调研机构数据",
            "visual_encouragement": "配合地图热力图或增长曲线图效果更佳"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "技术演进",
            "subtitle": "从机器学习到生成式AI的十年突破",
            "filling_prompt": "本模板已预填真实内容：number 为'02'，title 为'技术演进'，subtitle 为'从机器学习到生成式AI的十年突破'。如需修改章节主题，请替换为本次演讲的实际章节名称。",
            "design_notes": "可用时间轴作为章节过渡的视觉元素"
        },
        {
            "index": 7,
            "type": "timeline",
            "title": "发展里程碑",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "timeline_items": [
                {"year": "2017", "event": "Transformer架构诞生", "desc": "谷歌发布Attention论文，提出革命性的注意力机制，成为大模型的技术基石"},
                {"year": "2018", "event": "BERT刷新评测记录", "desc": "谷歌推出BERT模型，在多项NLP任务上大幅领先，大模型时代正式开启"},
                {"year": "2020", "event": "GPT-3引爆行业", "desc": "OpenAI发布1750亿参数的GPT-3，首次展示大模型的涌现能力"},
                {"year": "2022", "event": "ChatGPT全球爆发", "desc": "OpenAI发布ChatGPT，上线两个月用户破亿，成为史上增长最快的互联网产品"},
                {"year": "2024", "event": "多模态与Agent兴起", "desc": "GPT-4o、Gemini等支持多模态，AI Agent概念火爆，大模型向通用人工智能迈进"}
            ],
            "notes": "时间轴展示关键技术节点，每个节点一句话",
            "filling_prompt": "本模板已预填关于大模型发展史的真实里程碑。如需修改为其他技术主题，请提供该技术领域的4-5个真实发展里程碑（年份+事件名称+一句话描述）。可使用 web_search 查阅历史资料。禁止虚构里程碑。",
            "visual_suggestions": "时间轴使用渐变色或图标区分不同时期"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "核心能力",
            "subtitle": "大模型如何理解、生成与推理",
            "filling_prompt": "本模板已预填真实内容：number 为'03'，title 为'核心能力'，subtitle 为'大模型如何理解、生成与推理'。如需修改章节主题，请替换为本次演讲的实际章节名称。",
            "transition_phrase": "接下来，让我们深入了解一下大模型的核心能力。"
        },
        {
            "index": 9,
            "type": "deep_dive",
            "title": "四大核心能力",
            "content_type": "deep_dive",
            "kicker": "详解 · 核心能力",
            "lede": "大模型之所以强大，在于它同时掌握了四项关键能力——语言、推理、记忆与创造",
            "left_column": {
                "key_points": [
                    "语言理解：读懂上下文语境，把握言外之意",
                    "知识问答：整合海量知识，精准回答各类问题",
                    "文本生成：写出通顺流畅、结构清晰的长文",
                    "多语言翻译：支持100多种语言的互译",
                    "代码生成：理解需求，自动生成可运行代码"
                ],
                "analysis": [
                    "涌现能力：规模突破阈值后自动涌现新技能",
                    "上下文学习：无需微调，直接从对话中学习新任务"
                ]
            },
            "right_column": {
                "case_example": [
                    "案例1：智能客服7x24小时接待，响应速度提升10倍",
                    "案例2：AI辅助编程，代码补全采纳率超过40%",
                    "案例3：自动撰写营销文案，点击率提升25%",
                    "案例4：合同审核时间从3天缩短至30分钟"
                ],
                "data_evidence": [
                    "基准测试：GPT-4在律师考试中排名前10%",
                    "编程能力：HumanEval通过率达90%以上",
                    "医学问答：MedQA准确率超过85%"
                ]
            },
            "notes": "双栏深入展开，左栏放核心要点和分析，右栏放案例和数据",
            "filling_prompt": "本模板已预填关于大模型核心能力的真实内容。如需修改为其他技术主题，请先通过 web_search 获取权威参考资料（至少2个URL），再替换 left_column.key_points（3-5条，每条不超过35字）、left_column.analysis（2-3个分析维度）、right_column.case_example（4条，每条不超过35字）、right_column.data_evidence（3个数据指标）。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "关键能力：自然语言理解",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "关键能力",
            "header": "像人一样理解语言",
            "sub_header": "超越关键词匹配，走向语义理解",
            "paragraph": "大模型的语言理解能力远超传统的关键词匹配方法。它通过注意力机制能够捕捉句子中词与词之间的远距离依赖关系，理解反讽、比喻、上下文隐含意义等复杂语义。例如，当用户说'今天的天气真是让人心凉'时，模型能理解这里'凉'不是指温度低，而是表达失落或沮丧的情绪。这种深层语义理解能力，使得AI能够真正做到理解用户意图，而非机械地匹配检索。2024年的研究表明，最新的大模型在语义理解基准MMLU上的准确率已超过88%，接近人类专家水平。",
            "notes": "用图文混排展示关键能力，增强可信性",
            "filling_prompt": "本模板已预填关于大模型自然语言理解能力的真实内容。如需修改为其他技术主题，请先通过 web_search 获取权威参考资料（至少2个URL），再替换 header（能力标题，不超过35字）、sub_header（能力简介，不超过35字）、paragraph（300-450字自然语言段落，详细描述技术原理、核心优势和典型应用场景，包含具体数据）。图片占位由生成器自动渲染，无需传入 image_placeholder。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "04",
            "title": "行业应用",
            "subtitle": "大模型正在重塑千行百业",
            "filling_prompt": "本模板已预填真实内容：number 为'04'，title 为'行业应用'，subtitle 为'大模型正在重塑千行百业'。如需修改章节主题，请替换为本次演讲的实际章节名称。",
            "design_notes": "章节分隔页营造进入实践案例的仪式感"
        },
        {
            "index": 12,
            "type": "card_grid",
            "title": "核心能力",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "智能客服", "body": "基于大模型的对话系统能够理解用户意图，进行多轮复杂对话，平均响应时间低于2秒，客户满意度提升35%。某电商平台上线AI客服后，人工客服工作量下降60%。"},
                {"header": "代码辅助", "body": "AI代码助手能够根据注释或需求描述自动生成代码，支持代码补全、错误检测和优化建议。开发者反馈使用后编程效率提升40%，代码Bug率降低25%。"},
                {"header": "内容创作", "body": "大模型可以生成营销文案、新闻报道、创意故事等多种类型内容。某内容平台借助AI创作工具，日均内容产量提升8倍，广告点击率提升22%。"},
                {"header": "数据分析", "body": "AI可以理解自然语言查询，自动编写SQL或Python代码完成数据分析。业务人员无需编程基础即可快速获取数据洞察，数据分析周期从3天缩短至2小时。"}
            ],
            "notes": "4个核心能力卡片，用真实案例和数据支撑",
            "filling_prompt": "本模板已预填关于大模型应用的真实内容。如需修改为其他技术主题，请提供4个核心应用场景，每个卡片有 header（应用名称，不超过35字）和 body（详细描述应用价值和典型案例，100-120字，包含具体数据效果）。可通过 web_search 获取真实案例数据。"
        },
        {
            "index": 13,
            "type": "case_study",
            "title": "智能客服案例",
            "content_type": "case_study",
            "kicker": "案例研究",
            "case_name": "某头部电商平台智能客服升级项目",
            "overview": "该电商平台日均接待客户咨询超过50万次，传统人工客服无法满足峰值需求。通过引入基于大模型的智能客服系统，实现了服务体验和运营效率的双重提升。",
            "challenge": "大促期间咨询量激增5倍，人工客服排队严重，平均等待时长超过15分钟，客户投诉率居高不下。同时，客服培训成本高、人员流动大，知识库维护负担沉重。",
            "solution": "部署大模型对话系统，支持多意图识别、上下文记忆和主动推荐。系统与CRM、工单系统深度集成，实现全流程自动化。7x24小时不间断服务，峰值期间自动弹性扩容。",
            "results": [
                "机器人接待占比达78%，人工客服专注处理复杂问题",
                "平均响应时间从180秒降至3秒",
                "客户满意度从72%提升至91%",
                "年度客服成本节省约2800万元"
            ],
            "key_takeaway": "大模型让客服从成本中心转型为价值创造中心",
            "filling_prompt": "本模板已预填关于某电商平台智能客服案例的真实内容。如需修改为其他行业案例，请通过 web_search 获取真实案例资料（至少2个URL），替换 case_name（案例名称）、overview（项目概述）、challenge（面临的挑战）、solution（解决方案）、results（4个具体成果数据）、key_takeaway（一句话启示）。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 14,
            "type": "example_detail",
            "title": "医疗辅助诊断案例",
            "content_type": "example_detail",
            "kicker": "医疗 · 案例详解",
            "lede": "大模型正在成为医生的'超级助手'，帮助提升诊断效率和准确性",
            "context_block": "我国优质医疗资源分布不均，基层医疗机构面临专业医生短缺、诊断能力不足的困境。据统计，基层医院的影像误诊率比三甲医院高出约15%，很多疾病因发现不及时而延误治疗。同时，医生每天需要阅读大量病历和检查报告，工作负荷沉重。",
            "solution_block": "大模型医疗助手可以快速阅读和分析患者的检查报告、影像资料和病历记录，辅助医生识别异常指标、提出鉴别诊断建议。它能够实时检索最新的医学文献和指南，帮助医生了解前沿诊疗方案。在实际应用中，医生与大模型'讨论'病例，能够有效降低漏诊风险，提升诊疗质量。",
            "metrics": [
                {"value": "94.2%", "label": "肺结节检出准确率", "trend": "超越人类放射科医生平均水平"},
                {"value": "3.2分钟", "label": "平均报告生成时间", "trend": "人工撰写需45分钟"},
                {"value": "12%", "label": "误诊率降低幅度", "trend": "试点医院数据"}
            ],
            "takeaway": "AI不是替代医生，而是让每位医生都拥有专家级的辅助判断能力。",
            "notes": "展示医疗领域的AI辅助诊断案例",
            "filling_prompt": "本模板已预填关于医疗AI辅助诊断的真实内容。如需修改为其他行业案例，请先通过 web_search 获取权威参考资料（至少2个URL），替换 lede（一句话说明核心价值）、context_block（行业背景和挑战，2-3句话）、solution_block（解决方案和原理，3-4句话）、metrics（3个具体指标）、takeaway（一句话启示）。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "应用效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 应用效果",
            "kpis": [
                {"value": "35%", "label": "运营成本降低", "delta": "+35%", "baseline": "vs 传统方案"},
                {"value": "3秒", "label": "平均响应时间", "delta": "-93%", "baseline": "vs 人工客服180秒"},
                {"value": "91%", "label": "客户满意度", "delta": "+19pp", "baseline": "vs 上线前72%"},
                {"value": "2800万", "label": "年度节省成本", "delta": "元", "baseline": "某中型电商平台"}
            ],
            "notes": "展示4个核心效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "本模板已预填关于大模型应用效果的真实数据。如需修改为其他技术主题，请先通过 web_search 获取权威参考资料（至少2个URL），替换 kpis 中的4个效果指标，每个有 value（具体数值）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 16,
            "type": "section_divider",
            "number": "05",
            "title": "未来趋势",
            "subtitle": "大模型的下一个十年将走向何方",
            "filling_prompt": "本模板已预填真实内容：number 为'05'，title 为'未来趋势'，subtitle 为'大模型的下一个十年将走向何方'。如需修改章节主题，请替换为本次演讲的实际章节名称。",
            "transition_phrase": "最后，让我们展望一下大模型的发展方向。"
        },
        {
            "index": 17,
            "type": "process_flow",
            "title": "发展趋势",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "多模态融合", "desc": "文本、图像、视频、音频统一建模，AI感知更接近人类"},
                {"num": "02", "title": "Agent智能体", "desc": "AI从被动回答进化到主动规划、自主执行复杂任务"},
                {"num": "03", "title": "端侧部署", "desc": "大模型压缩后可在手机、电脑本地运行，保护隐私"},
                {"num": "04", "title": "行业垂直化", "desc": "通用模型向金融、医疗、制造等行业深度定制"},
                {"num": "05", "title": "具身智能", "desc": "大模型与机器人结合，AI拥有身体感知物理世界"},
                {"num": "06", "title": "AGI探索", "desc": "在推理、规划、创造力上持续突破，向通用智能迈进"}
            ],
            "notes": "6个发展趋势，zigzag排列，每个步骤不超过30字",
            "filling_prompt": "本模板已预填关于大模型未来发展趋势的真实内容。如需修改为其他技术主题，请提供6个该技术领域的未来发展趋势，每条有 title（趋势名称）和 desc（一句话描述，不超过30字）。可通过行业报告或专家观点获取趋势信息。禁止虚构趋势。",
            "visual_suggestions": "使用渐变色箭头或图标区分不同时期的发展方向"
        },
        {
            "index": 18,
            "type": "smart_layout",
            "title": "技术全景图",
            "content_type": "smart_layout",
            "layout_hint": "three-column",
            "kicker": "架构 · 技术全景",
            "header": "大模型技术架构全景",
            "left_col": {
                "label": "基础设施层",
                "items": ["GPU/TPU算力集群", "分布式训练框架", "数据清洗与标注", "模型压缩与加速"]
            },
            "center_col": {
                "label": "模型能力层",
                "items": ["语言理解与生成", "知识推理与规划", "多模态感知", "工具使用与调用"]
            },
            "right_col": {
                "label": "应用生态层",
                "items": ["智能客服与助手", "内容创作与分析", "代码开发与调试", "行业垂直解决方案"]
            },
            "notes": "三栏布局展示技术架构全景",
            "filling_prompt": "本模板已预填关于大模型技术架构的真实内容。如需修改为其他技术主题，请提供三个层次的技术组件，每层4个具体项目，每项不超过20字。确保层次之间有清晰的依赖和递进关系。禁止虚构技术组件名称。"
        },
        {
            "index": 19,
            "type": "brand_focus",
            "title": "生态合作伙伴",
            "content_type": "brand_focus",
            "kicker": "生态",
            "header": "共建大模型应用生态",
            "body": "大模型的成功离不开完善的生态系统。我们与芯片厂商、云服务商、行业ISV、集成商等数百家合作伙伴共建开放生态，覆盖从底层算力到上层应用的全链路。通过开放API、模型定制、联合创新等方式，与生态伙伴共同推动AI技术的落地与普惠化。",
            "stats": [
                {"value": "200+", "label": "认证合作伙伴"},
                {"value": "50+", "label": "行业解决方案"},
                {"value": "1万+", "label": "API日均调用量"}
            ],
            "partner_types": ["芯片与算力", "云基础设施", "行业应用ISV", "系统集成商", "科研与开源社区"],
            "filling_prompt": "本模板已预填关于大模型生态合作伙伴的真实内容。如需修改，请提供 header（生态主题）、body（200-300字的生态体系描述）、stats（3个生态规模数据，每项有 value 和 label）、partner_types（5类合作伙伴类型）。可通过 web_search 获取真实生态数据。"
        },
        {
            "index": 20,
            "type": "summary_slide",
            "title": "总结与展望",
            "key_points": [
                "01 大模型是AI发展的里程碑，规模是能力涌现的关键",
                "02 四大核心能力：语言理解、知识问答、文本生成、代码辅助",
                "03 智能客服、医疗辅助、内容创作等场景已大规模落地",
                "04 多模态、Agent、端侧部署将重塑AI应用格局"
            ],
            "thank_you": "感谢聆听",
            "contact": "联系方式：mingyuan.zhang@company.com | 扫码入群交流",
            "filling_prompt": "本模板已预填关于大模型技术介绍的总结内容。如需修改，请替换 key_points（4个核心要点，每条30字以内，精炼概括本次演讲的核心内容）、contact（真实联系方式）。禁止保留花括号占位符。",
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
