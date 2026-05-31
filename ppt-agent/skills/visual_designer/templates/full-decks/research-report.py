TEMPLATE = {
    "name": "research-report",
    "name_cn": "调研报告",
    "description": "适合市场调研、行业分析、可行性研究等场景。数据详实，逻辑严密，结论明确。",
    "target_audience": "决策层、项目组、客户、评审专家",
    "typical_slides": 16,
    "typical_duration": "20-30分钟",
    "palette": "report_green",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "research_objectives": "调研的核心目的是为决策提供依据",
        "key_principles": [
            "数据说话：用数据支撑观点",
            "逻辑严密：论证过程清晰可追溯",
            "客观中立：如实呈现调研发现",
            "结论明确：给出可操作的建议"
        ],
        "methodology_tips": [
            "明确调研范围和方法",
            "多渠道收集数据，确保代表性",
            "交叉验证数据，确保准确性",
            "区分事实与观点"
        ]
    },
    "report_structure": {
        "executive_summary": "执行摘要",
        "background": "调研背景",
        "methodology": "调研方法",
        "findings": "现状分析",
        "diagnosis": "问题诊断",
        "recommendations": "对策建议",
        "conclusion": "结论展望"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "中国智能客服行业发展现状调研",
            "subtitle": "2024-2025年度行业深度研究报告",
            "author": "行业研究中心",
            "date": "2025年6月",
            "notes": "封面页正式庄重，注明调研时间和团队",
            "filling_prompt": "本模板为固定格式，title 填入调研报告的名称，subtitle 填入概括性副标题，author 填入调研团队或机构名称，date 填入报告完成日期。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  调研背景",
                "02  调研方法",
                "03  现状分析",
                "04  问题诊断",
                "05  对策建议",
                "06  结论展望"
            ],
            "notes": "清晰展示报告结构，六大章节一览无余",
            "filling_prompt": "目录页为固定结构，根据实际章节内容调整编号和章节名称。"
        },
        {
            "index": 3,
            "type": "example_detail",
            "title": "执行摘要",
            "content_type": "example_detail",
            "kicker": "摘要 · 核心发现",
            "lede": "中国智能客服市场规模已达487亿元，年复合增长率维持在23%，但企业渗透率不足40%，市场增长空间仍然广阔。",
            "context_block": "随着人力成本持续上升和客户体验要求不断提高，传统人工客服模式面临成本高、效率低、质量不稳定等多重压力。头部企业率先启动智能化转型，中小企业需求也在快速释放。",
            "solution_block": "本次调研覆盖全国31个省份327家企业，涵盖金融、零售、政务、医疗等12个行业。调研发现，技术成熟度不足、人才短缺和ROI难量化是制约行业渗透率提升的三大核心瓶颈。与此同时，多模态交互、大模型赋能和全渠道融合正成为下一阶段竞争的关键方向。",
            "metrics": [
                {"value": "487亿元", "label": "2024年市场规模", "trend": "同比增长31%"},
                {"value": "23%", "label": "年复合增长率", "trend": "近三年均值"},
                {"value": "38%", "label": "企业渗透率", "trend": "较2022年提升12%"},
                {"value": "4.2次", "label": "人均日处理会话", "trend": "AI辅助后效率提升4倍"}
            ],
            "takeaway": "智能客服已进入快速渗透期，但技术落地能力和商业价值验证仍是决定市场能否持续高增长的核心要素。",
            "notes": "执行摘要页，一页概括报告核心发现和建议方向",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含行业报告来源如艾瑞、IDC、Gartner等），再填入真实内容：lede 用一句话概括核心发现（不超过60字）；context_block 描述调研背景和问题现状（2-3句话）；solution_block 总结主要发现和建议方向（3-4句话）；metrics_grid 提供4个调研指标，每个有具体的 value、label、trend，数据必须与 references 来源一致；takeaway 用一句话总结调研对决策的意义。references 逐条列出 URL 并标注来源机构名称。禁止虚构数据。"
        },
        {
            "index": 4,
            "type": "section_divider",
            "number": "01",
            "title": "调研背景",
            "subtitle": "目的与范围",
            "notes": "章节分隔页，标注第一章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 5,
            "type": "image_text",
            "title": "调研背景与研究目的",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景",
            "header": "人力成本攀升与客户体验升级的双重驱动",
            "sub_header": "智能客服从「可选项」变为「必选项」",
            "paragraph": "近年来，我国劳动力成本年均增长约8.5%，客服行业人员流动率高达120%，企业客服部门年均离职率超过60%。与此同时，消费者对服务响应速度和解决问题能力的要求达到历史最高水平——调研显示，78%的消费者期望在3分钟内获得问题解答，67%的消费者会因一次不良体验而放弃该品牌。在此背景下，采用智能客服系统替代重复性人工劳动，已成为企业控制成本和提升体验的共同选择。",
            "references": [
                "https://www.iresearch.com.cn/Detail/Report?id=4497  艾瑞咨询《中国智能客服行业研究报告2024》",
                "https://www.idc.com/promo/customer-experience  IDC全球客户体验趋势报告2024",
                "https://www.mckinsey.com/business-functions/operations/our-insights  麦肯锡运营与客户体验白皮书"
            ],
            "notes": "右侧配行业趋势图或政策环境示意图，左侧呈现核心背景",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含行业报告来源），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最核心的一个调研背景（不超过35字）；sub_header 为背景的影响说明（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述调研背景（行业趋势、政策环境、企业痛点），用流畅的段落形式呈现，禁止罗列要点，必须包含具体数字。references 逐条列出 URL 并标注报告来源机构名称。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "调研方法",
            "subtitle": "数据来源与分析方法",
            "notes": "章节分隔页，标注第二章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "process_flow",
            "title": "数据来源与研究方法",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "问卷调研", "desc": "线上+线下发放问卷1200份，有效回收892份，覆盖一线至五线城市企业"},
                {"num": "02", "title": "深度访谈", "desc": "对32家企业的客服负责人进行60-90分钟一对一深度访谈"},
                {"num": "03", "title": "公开数据", "desc": "采集艾瑞、IDC、Gartner等机构公开报告及企业年报数据287份"},
                {"num": "04", "title": "交叉验证", "desc": "对关键数据点进行多源比对和专家评审，确保数据准确性"}
            ],
            "notes": "四步研究方法，确保数据可靠",
            "filling_prompt": "必须填入真实内容：提供4个研究方法步骤，每个有 title（方法名称）和 desc（具体描述，包括样本量、时间范围、数据来源等）。方法描述要具体，体现研究的科学性和严谨性。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "现状分析",
            "subtitle": "数据与发现",
            "notes": "章节分隔页，标注第三章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "kpi_dashboard",
            "title": "智能客服市场规模",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 市场规模",
            "kpis": [
                {"value": "487亿元", "label": "2024年市场规模", "delta": "+31% YoY", "baseline": "2023年: 372亿元"},
                {"value": "23%", "label": "近三年CAGR", "delta": "稳健增长", "baseline": "2021-2024年复合"},
                {"value": "38%", "label": "企业渗透率", "delta": "较2022年+12%", "baseline": "大型企业渗透率达62%"},
                {"value": "1.2万亿元", "label": "潜在市场空间", "delta": "2028年预测", "baseline": "基于企业数量估算"}
            ],
            "notes": "四个核心KPI展示市场规模和增长态势",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL，需包含艾瑞、IDC、Gartner等来源），再填入真实内容：提供4个关键数据指标，每个有 value、label、delta、baseline。这些数据要能直接支持调研结论，数字必须与 references 来源一致。references 逐条列出 URL 并标注来源机构名称。禁止虚构数据。"
        },
        {
            "index": 10,
            "type": "chart_slide",
            "title": "竞争格局分析",
            "content_type": "chart_slide",
            "layout_hint": "horizontal",
            "kicker": "竞争 · 市场份额",
            "chart_type": "bar",
            "header": "头部厂商市场份额（2024年）",
            "sub_header": "阿里、百度、科大讯飞位列前三，CR5超过65%",
            "categories": ["阿里小蜜", "百度智能云", "科大讯飞", "腾讯企点", "容联七陌", "其他"],
            "values": [22, 18, 15, 10, 7, 28],
            "unit": "%",
            "notes": "以水平条形图展示各厂商市场份额",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：header 填入图表标题（不超过35字）；sub_header 填入副标题说明（不超过35字）；categories 填入6个厂商名称；values 填入对应的百分比数值（总和应为100）；unit 填入单位（%）。references 逐条列出 URL 并标注来源机构。数据必须与 references 一致，禁止虚构。"
        },
        {
            "index": 11,
            "type": "case_study",
            "title": "典型企业案例",
            "content_type": "case_study",
            "kicker": "案例 · 最佳实践",
            "company": "招商银行信用卡中心",
            "industry": "金融行业",
            "challenge": "日均会话量超过200万次，人工客服团队超过3000人，年均人力成本超过8亿元，同时客户满意度面临下行压力。",
            "solution": "部署基于大语言模型的智能客服系统，接入知识库超过50万条FAQ，实现85%的常见问题由AI全自动处理，复杂问题无缝转人工。",
            "results": [
                {"metric": "人力成本节省", "value": "2.1亿元/年", "detail": "减少客服坐席1200人"},
                {"metric": "首响时间缩短", "value": "从47秒降至3秒", "detail": "AI即时回复"},
                {"metric": "客户满意度", "value": "提升至94%", "detail": "较上线前提升9个百分点"},
                {"metric": "问题解决率", "value": "92%", "detail": "AI+人工协同"}
            ],
            "takeaway": "金融行业头部企业已验证智能客服的高ROI路径，但需注重知识库质量和人机协同设计。",
            "notes": "标杆企业案例，展示最佳实践和可量化的ROI",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：company 填入企业名称；industry 填入所属行业；challenge 填入企业面临的核心痛点（100-150字，包含具体数字）；solution 填入采用的解决方案（100-150字）；results 填入4个量化结果指标，每个有 metric、value、detail，数据需与 references 一致；takeaway 填入一句话总结（不超过60字）。references 列出 URL 并标注来源。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "04",
            "title": "问题诊断",
            "subtitle": "挑战与风险",
            "notes": "章节分隔页，标注第四章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "card_grid",
            "title": "行业面临的主要挑战",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "kicker": "诊断 · 痛点分析",
            "cards": [
                {"header": "技术成熟度不足", "body": "当前NLP模型在复杂多轮对话、方言识别、专业领域知识理解上仍有明显短板。调研中62%的企业反映，智能客服在处理模糊意图和情感化表达时准确率低于70%，难以真正替代人工。"},
                {"header": "ROI量化困难", "body": "中小企业普遍缺乏数据基础设施，难以建立科学的ROI评估体系。47%的被调研企业表示无法准确衡量智能客服投入产出比，导致续费决策周期延长，平均决策周期达6-8个月。"},
                {"header": "人才缺口严重", "body": "既懂AI技术又懂客服业务的复合型人才严重短缺。调研显示，71%的企业表示AI训练师和知识工程师岗位招聘困难，平均每个岗位招聘周期超过90天，制约了系统的持续优化。"},
                {"header": "数据安全与合规", "body": "客服场景涉及大量用户隐私和商业敏感数据，合规要求日趋严格。38%的金融和医疗行业客户表示，数据本地化要求和跨系统对接的合规审计增加了至少3个月的部署周期。"}
            ],
            "notes": "四个核心痛点，客观呈现行业挑战",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL），再填入真实内容：提供4个痛点，每个有 header（不超过35字）和 body（详细描述，100-120字，包含具体数据或案例场景）。禁止虚构或使用通用化描述。references 列出 URL。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "05",
            "title": "对策建议",
            "subtitle": "解决方案与行动计划",
            "notes": "章节分隔页，标注第五章开始",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 15,
            "type": "deep_dive",
            "title": "战略建议与行动计划",
            "content_type": "deep_dive",
            "kicker": "建议 · 战略方向",
            "header": "三大战略方向推动行业持续健康发展",
            "sub_header": "技术深化、场景聚焦、生态协同",
            "context": "基于调研发现的痛点和行业趋势，本报告提出三条相互关联的战略建议，覆盖技术、场景和生态三个维度。三条建议形成递进关系：技术深化是基础，场景聚焦是路径，生态协同是放大器。",
            "findings": [
                {"num": "01", "title": "技术深化：加大NLP与领域知识图谱融合投入", "body": "建议企业将知识图谱构建作为智能客服的核心基础设施，优先在高频业务场景中建立结构化知识库。重点投入多模态交互能力（语音+文本+图像），并在合规前提下引入大模型进行意图理解和答案生成。调研显示，拥有完善知识图谱的企业，智能客服问题解决率平均高出28%。"},
                {"num": "02", "title": "场景聚焦：优先在高价值、高频场景中落地", "body": "建议从ROI最明确的场景切入：FAQ咨询、订单查询、退换货处理、预约挂号等高频标准化场景优先上线，再逐步扩展至复杂业务场景。聚焦场景可显著缩短部署周期，调研中场景聚焦策略的企业平均上线周期为3个月，而全面铺开的企业平均需要8个月。"},
                {"num": "03", "title": "生态协同：构建技术供应商+行业ISV+企业三方协作", "body": "建议形成分工明确的产业生态：基础AI能力由技术大厂提供，行业Know-How由ISV封装，企业专注业务运营和效果优化。生态协作模式可将企业智能客服的总体拥有成本（TCO）降低35%，同时将系统上线速度提升50%。"}
            ],
            "takeaway": "智能客服行业的下一轮增长，将由技术深度、场景精度和生态广度共同驱动，三者缺一不可。",
            "notes": "深入剖析三条战略建议及实施路径",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少3个URL），再填入真实内容：header 填入核心建议主题（不超过35字）；sub_header 填入三个关键词（不超过35字）；context 填入总体背景和三条建议的关系说明（100-150字）；findings 填入3条建议，每条有 num、title（不超过35字）、body（详细展开，150-200字，包含数据支撑）；takeaway 填入总结语（不超过60字）。references 逐条列出 URL。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "结论与展望",
            "content_type": "summary_slide",
            "kicker": "总结",
            "key_points": [
                "01 市场规模487亿元，年复合增长率23%，智能客服已进入快速渗透期",
                "02 技术成熟度、ROI量化、人才供给是制约渗透率提升的三大核心瓶颈",
                "03 建议路径：技术深化（知识图谱+大模型）→ 场景聚焦（高频标准化场景优先）→ 生态协同（三方协作降本增效）"
            ],
            "thank_you": "感谢聆听",
            "notes": "简洁总结三大核心结论和行动建议",
            "filling_prompt": "必须填入真实内容：key_points 提供3个核心结论，每条不超过60字，内容基于调研发现，禁止虚构。thank_you 填入感谢语。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "调研报告要数据说话，有据可查",
        "逻辑要严密，结论要有数据支撑",
        "建议要具体可执行",
        "引用来源要准确，注明出处",
        "问题诊断要客观，不回避",
        "图表要清晰，数据解读要到位",
        "结论要明确，便于决策参考"
    ],
    "data_visualization": {
        "recommended_charts": [
            "市场规模趋势图",
            "竞争格局饼图/条形图",
            "消费者画像雷达图",
            "问题分布帕累托图",
            "建议优先级矩阵图"
        ],
        "chart_design_tips": [
            "图表标题清晰，标注数据来源",
            "颜色统一，视觉协调",
            "数据标签清晰可读",
            "适当使用对比色突出重点"
        ]
    },
    "methodology_appendix": {
        "questionnaire_design": "问卷设计说明",
        "sampling_method": "抽样方法说明",
        "confidence_level": "置信水平说明",
        "data_validation": "数据验证方法说明"
    }
}
