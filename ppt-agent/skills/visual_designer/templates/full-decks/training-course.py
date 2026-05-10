TEMPLATE = {
    "name": "training-course",
    "name_cn": "培训课件",
    "description": "适合内部培训、新人入职培训、技能培训等场景。知识系统，讲解清晰，互动引导。",
    "target_audience": "员工、新人、需要学习相关技能的人员",
    "typical_slides": 17,
    "typical_duration": "30-60分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "training_philosophy": "培训的目标是让学员'学会'而非'听过'",
        "key_principles": [
            "内容要由浅入深，循序渐进",
            "概念讲解要准确，举例要贴切",
            "图文并茂，增强理解",
            "留出互动和练习时间"
        ],
        "adult_learning": {
            "relevance": "培训内容要与工作实际相关",
            "experience": "善于利用学员已有经验",
            "problem_orientation": "聚焦解决实际问题",
            "autonomy": "给予学员一定选择权"
        }
    },
    "training_objectives": {
        "knowledge": "知识目标：理解XXX的概念和原理",
        "skill": "技能目标：能够独立完成XXX操作",
        "attitude": "态度目标：建立对XXX的正确认识"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "新员工入职培训",
            "subtitle": "快速融入，掌握基础，迈向卓越",
            "author": "人力资源部",
            "date": "2025年3月10日",
            "notes": "标题页正式，注明培训主题和讲师",
            "filling_prompt": "必须填入真实内容：title 为培训主题名称，subtitle 为培训副标题，author 为培训讲师姓名，date 为培训日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "content_slide",
            "title": "培训目标",
            "content_type": "content_slide",
            "objectives": [
                "了解公司发展历程、使命愿景价值观",
                "掌握公司组织架构和部门职责",
                "熟悉工作流程和常用工具系统",
                "理解公司文化和行为规范",
                "建立职业发展基本认知"
            ],
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
            "concepts": [
                {"name": "公司使命", "desc": "让商业更高效，让生活更美好"},
                {"name": "公司愿景", "desc": "成为最受尊敬的科技企业"},
                {"name": "核心价值观", "desc": "客户第一、团队协作、拥抱变化、追求卓越"},
                {"name": "人才理念", "desc": "德才兼备，以德为先；能上能下，能进能出"}
            ],
            "notes": "讲解核心概念和定义",
            "filling_prompt": "必须填入真实内容：列出3-5个核心概念，每个有名称和定义/说明（每条说明不超过35字）。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "背景介绍",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景",
            "header": "公司发展历程与行业地位",
            "sub_header": "10年深耕，成为行业领军者",
            "paragraph": "公司成立于2015年，是一家专注于企业级SaaS服务的科技公司。10年来，公司从最初的10人团队发展到如今的2000+人规模，业务覆盖全国30个省市自治区，服务企业客户超过10000家。公司先后获得红杉资本、高瓴资本等知名投资机构的多轮融资，估值超过50亿元。公司产品连续5年被评为'最受欢迎的企业服务产品'，客户NPS评分位居行业第一。",
            "references": [
                "https://www.company.com/about",
                "https://www.36kr.com/"
            ],
            "notes": "用图文混排讲解背景知识",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'背景'；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为背景主题（不超过35字）；sub_header 为背景说明；paragraph 为300-450字的自然语言段落，详细描述培训内容的背景、重要性和发展趋势，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
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
            "type": "example_detail",
            "title": "重点知识",
            "content_type": "example_detail",
            "kicker": "实例 · 重点知识",
            "lede": "掌握职场基本规范，快速融入团队",
            "context_block": "新人入职后，常见的困惑包括：不知道如何与同事相处，不了解公司的隐性规则，不清楚如何高效沟通。通过调研，新员工最希望了解的前3个问题是：1）如何快速建立人脉；2）如何高效汇报工作；3）如何处理跨部门协作。",
            "solution_block": "本部分将重点讲解：1）职场沟通技巧：金字塔原理、邮件规范、会议参与；2）高效工作方法：优先级矩阵、时间管理、工具使用；3）跨部门协作：角色定位、期望管理、资源协调。通过案例分析和互动练习，帮助新员工掌握这些实用技能。",
            "metrics": [
                {"value": "1周内", "label": "预计掌握时间", "trend": "vs 传统3个月"},
                {"value": "高频使用", "label": "实际应用场景", "trend": "日常工作"},
                {"value": "显著提升", "label": "预期效果", "trend": "沟通效率"}
            ],
            "takeaway": "启示：掌握这些技能，将帮助你在职场快速成长",
            "notes": "讲解本次培训的重点内容",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的知识点；context_block 描述常见误区或薄弱环节（1-2句话）；solution_block 详细讲解核心要点和实践技巧（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结核心价值。禁止保留花括号。"
        },
        {
            "index": 9,
            "type": "process_flow",
            "title": "操作流程",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "日报填写", "desc": "每日15:00前提交工作日报"},
                {"num": "02", "title": "周报汇总", "desc": "每周五提交周报给直属领导"},
                {"num": "03", "title": "月度复盘", "desc": "每月最后一周进行月度总结"},
                {"num": "04", "title": "季度评估", "desc": "每季度进行绩效评估"},
                {"num": "05", "title": "年度review", "desc": "每年进行年度总结与规划"}
            ],
            "notes": "用流程图展示操作步骤",
            "filling_prompt": "必须填入真实内容：提供5个操作步骤，每个步骤有 title（步骤名称，不超过35字）和 desc（操作说明，不超过35字）。"
        },
        {
            "index": 10,
            "type": "image_text",
            "title": "案例讲解：高效沟通",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "案例",
            "header": "如何向上级汇报工作",
            "sub_header": "结论先行，数据支撑",
            "paragraph": "小王是今年新入职的产品经理。在第一次向总监汇报工作时，他准备了30分钟的详细讲解，从背景、过程讲到结果，但由于没有提前明确汇报目的，总监听完一头雾水。后来主管指导他使用'金字塔原理'：先说结论，再说论据，最后补充细节。调整后，同样的汇报内容只需5分钟，总监立刻明白了重点，也给出了有针对性的反馈。这个案例告诉我们：职场沟通要有的放矢，结论先行，效率为王。",
            "references": [
                "https://www.mckinsey.com/",
                "https://www.bain.com/"
            ],
            "notes": "用图文混排讲解一个具体案例",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'案例'；title 中的 {案例名称} 替换为具体案例；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为案例简介；paragraph 为300-450字的自然语言段落，详细描述案例的背景、具体情况、处理过程和经验启示，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
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
            "type": "example_detail",
            "title": "练习任务",
            "content_type": "example_detail",
            "kicker": "实例 · 练习任务",
            "lede": "通过实战练习，巩固所学知识",
            "context_block": "学习知识只是第一步，更重要的是将知识转化为能力。本次练习将模拟真实工作场景，让大家在实践中运用所学技能。",
            "solution_block": "练习任务：1）小组讨论：针对指定的跨部门协作案例，分析问题并提出解决方案；2）角色扮演：模拟向上级汇报的场景，练习金字塔表达法；3）工具实操：在OA系统中完成请假、报销等流程操作。完成后各组分享，导师点评。",
            "metrics": [
                {"value": "30分钟", "label": "预计完成时间", "trend": "含讨论和分享"},
                {"value": "小组形式", "label": "评判标准", "trend": "团队协作+解决方案"},
                {"value": "85%", "label": "预计正确率", "trend": "vs 培训前"}
            ],
            "takeaway": "启示：实践出真知，多练才能熟练",
            "notes": "布置课堂练习任务",
            "filling_prompt": "必须填入真实内容：lede 一句话概括练习核心目标；context_block 描述练习目的和预期效果（1-2句话）；solution_block 具体说明操作步骤和关键要点（2-3句话）；metrics_grid 提供3个练习指标；takeaway 用一句话总结价值。"
        },
        {
            "index": 13,
            "type": "example_detail",
            "title": "常见问题",
            "content_type": "example_detail",
            "kicker": "实例 · 常见问题",
            "lede": "避开这些坑，少走弯路",
            "context_block": "根据对新员工的跟踪调研，以下问题出现频率最高。如果能提前了解并避免，将大大加速你的成长。",
            "solution_block": "常见问题及解决方案：1）不敢问问题→建立'提问清单'，定期向导师请教；2）被动等待→主动向领导确认优先级；3）埋头单干→遇到困难及时求助，团队协作更高效；4）只顾执行→定期复盘，总结经验教训；5）忽视健康→工作再忙也要注意休息，保持身心健康。",
            "metrics": [
                {"value": "60%", "label": "问题发生率", "trend": "vs 老员工"},
                {"value": "节省1个月", "label": "平均解决时间", "trend": "提前预防"},
                {"value": "高", "label": "问题影响级别", "trend": "影响转正评估"}
            ],
            "takeaway": "启示：多向前辈请教，少踩坑、快成长",
            "notes": "列出常见问题和解决方法",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最易遇到的问题；context_block 描述问题具体表现（1-2句话）；solution_block 针对每个问题提供解决方法和技术细节（2-3句话）；metrics_grid 提供3个问题相关指标；takeaway 用一句话总结如何避免。禁止保留花括号。"
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
            "highlights": [
                {"title": "公司文化", "desc": "使命愿景价值观，人才理念"},
                {"title": "沟通技巧", "desc": "金字塔原理，邮件规范"},
                {"title": "工作方法", "desc": "优先级矩阵，时间管理"},
                {"title": "协作模式", "desc": "跨部门协作，期望管理"},
                {"title": "成长路径", "desc": "绩效评估，职业发展"}
            ],
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
            "kpis": [
                {"value": "8分", "label": "知识掌握程度", "delta": "vs 培训前+5分", "baseline": "满分10分"},
                {"value": "9分", "label": "技能熟练程度", "delta": "vs 培训前+6分", "baseline": "满分10分"},
                {"value": "7分", "label": "文化认同程度", "delta": "vs 培训前+4分", "baseline": "满分10分"},
                {"value": "立即行动", "label": "下一步计划", "delta": "制定中", "baseline": "待确定"}
            ],
            "notes": "帮助学员自评学习效果",
            "filling_prompt": "必须填入真实内容：提供4个学习效果自评维度，每个有 value（如分数或完成度）、label（维度名称）、delta（提升情况）。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 掌握公司文化和职场基础",
                "02 学会高效沟通和工作方法",
                "03 持续学习，快速成长"
            ],
            "thank_you": "感谢聆听！",
            "contact": "HR邮箱：hr@company.com | 新人导师：各业务部门指定",
            "notes": "简洁总结，提出后续学习建议",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心知识点2条+下一步行动1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "培训内容要由浅入深，循序渐进",
        "概念讲解要准确，举例要贴切",
        "图文并茂，增强理解",
        "留出互动和练习时间",
        "结尾要有后续学习指引",
        "配套测试题，检验学习效果",
        "收集反馈，持续改进培训内容",
        "提供学习资料，方便后续复习"
    ],
    "training_materials": {
        "handouts": [
            "培训课件PDF",
            "常用工具操作手册",
            "公司制度汇编",
            "常见问题FAQ"
        ],
        "online_resources": [
            "内部知识库",
            "学习视频中心",
            "导师答疑群"
        ]
    },
    "assessment_methods": {
        "during_training": "课堂测验、练习表现",
        "after_training": "在线测试（80分及格）",
        "on_job": "1个月后绩效跟踪"
    }
}
