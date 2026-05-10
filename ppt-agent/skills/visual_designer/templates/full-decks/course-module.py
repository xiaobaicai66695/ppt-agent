TEMPLATE = {
    "name": "course-module",
    "name_cn": "课程课件",
    "description": "适合教学课件、培训材料、知识分享等场景。内容系统化，重点清晰，便于学习和理解。",
    "target_audience": "学生、培训学员、自学者",
    "typical_slides": 17,
    "typical_duration": "45-90分钟（一节课或一个章节）",
    "palette": "sage_calm",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "teaching_philosophy": "以学生为中心，注重启发式教学",
        "key_principles": [
            "循序渐进：从已知到未知，从简单到复杂",
            "理论联系实际：通过案例帮助理解抽象概念",
            "互动引导：设置思考题和讨论环节",
            "及时巩固：每个知识点后有练习和反馈"
        ],
        "assessment_hints": "可配套准备课后作业、测验题目"
    },
    "learning_objectives": {
        "knowledge": "学员将理解XXX的基本概念和原理",
        "skills": "学员将能够运用XXX解决实际问题",
        "attitude": "学员将建立对XXX的正确认识和学习兴趣"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "Python数据分析基础",
            "subtitle": "NumPy与Pandas入门",
            "author": "李老师 | 数据科学学院",
            "date": "2025年3月10日",
            "filling_prompt": "必须填入真实内容：title 为实际课程或章节名称，subtitle 为一句课程简介，author 为讲师姓名，date 为日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  基础知识",
                "02  进阶应用",
                "03  实践操作",
                "04  总结回顾"
            ],
            "notes": "让学员快速了解课程结构，每项一行即可",
            "filling_prompt": "目录页为固定结构，无需额外填充。",
            "timing_hint": "约30秒"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "基础知识",
            "subtitle": "核心概念与原理",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "学习目标",
            "content_type": "example_detail",
            "kicker": "实例 · 学习目标",
            "lede": "掌握NumPy和Pandas的核心用法，能够独立完成数据清洗和分析任务",
            "context_block": "数据科学已成为当今最热门的技能之一。企业在招聘时越来越重视候选人的数据分析能力。然而，很多初学者在学习过程中常常感到困惑：NumPy和Pandas有什么区别？如何高效地进行数据处理？",
            "solution_block": "本章节将系统讲解NumPy数组和Pandas数据框的核心操作，通过大量实例帮助学员建立直观理解。学完本课程后，学员将能够：理解NumPy和Pandas的适用场景；掌握数组创建、索引、运算的基本方法；能够使用Pandas进行数据读取、清洗、统计分析。",
            "metrics": [
                {"value": "80%+", "label": "预计掌握程度", "trend": "vs 课程前"},
                {"value": "90分钟", "label": "预计所需时间", "trend": "理论+实践"},
                {"value": "3个", "label": "配套实践项目", "trend": "循序渐进"}
            ],
            "takeaway": "启示：动手实践是学习数据分析的最佳方式",
            "notes": "本章节需要掌握的3-4个核心知识点",
            "filling_prompt": "必须填入真实内容：lede 一句话说明本章节的核心价值；context_block 描述学员常见的困惑或误区（1-2句话）；solution_block 说明本章内容如何帮助解决（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结学完本章的核心收获。禁止保留花括号。"
        },
        {
            "index": 5,
            "type": "deep_dive",
            "title": "NumPy核心机制",
            "content_type": "deep_dive",
            "kicker": "详解 · NumPy数组",
            "lede": "NumPy是Python科学计算的基础库，提供高效的多维数组支持",
            "left_column": {
                "key_points": [
                    "ndarray：NumPy核心数据结构，支持多维数组",
                    "向量化运算：无需循环，直接对整个数组操作",
                    "广播机制：自动扩展数组维度进行运算",
                    "内存连续：连续的内存块确保高效的计算性能"
                ],
                "analysis": [
                    "相比Python list：NumPy数组性能提升10-100倍",
                    "内存布局：连续内存 vs 分散对象，更适合CPU缓存"
                ]
            },
            "right_column": {
                "case_example": [
                    "创建数组：np.array([1,2,3])",
                    "数组运算：arr * 2 对每个元素乘2",
                    "索引切片：arr[0:3] 取前3个元素",
                    "形状变换：arr.reshape(2,3) 改变维度"
                ],
                "data_evidence": [
                    "性能提升：100倍（vs Python list）",
                    "内存节省：50%（vs 等价list）",
                    "计算速度：向量化比循环快10倍"
                ]
            },
            "notes": "双栏深入展开，左栏放核心要点和分析，右栏放案例和数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_column.key_points 为核心要点（3-4条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 web_search 获取的 URL（至少2个）。禁止保留花括号；禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "进阶应用",
            "subtitle": "典型场景与案例",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "应用场景",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "数据清洗", "body": "处理缺失值、重复数据、异常值"},
                {"header": "数据转换", "body": "格式转换、编码处理、数据合并"},
                {"header": "统计分析", "body": "描述统计、分组聚合、相关性分析"},
                {"header": "数据可视化", "body": "配合Matplotlib绘制图表"}
            ],
            "notes": "4个典型应用场景",
            "filling_prompt": "必须填入真实内容：提供4个典型应用场景，每个有 header（场景名称）和 body（一句话描述）。场景要具体，如'微服务架构'、'弹性扩容'。"
        },
        {
            "index": 8,
            "type": "image_text",
            "title": "行业案例：电商用户分析",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "电商行业",
            "header": "电商平台用户行为分析",
            "sub_header": "从数据中发现业务洞察",
            "paragraph": "某电商平台希望分析用户的购买行为，优化推荐算法。数据团队使用Pandas对用户订单数据进行分析，发现了多个有价值的信息：通过RFM模型识别出高价值用户群体；通过复购周期分析优化了营销触达时机；通过商品关联分析改进了推荐系统。分析结果显示，针对高价值用户的定向营销活动，转化率提升了45%。",
            "references": [
                "https://pandas.pydata.org/",
                "https://numpy.org/"
            ],
            "notes": "用图文混排展示行业应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 中的 {行业} 替换为具体行业（如'金融'、'电商'、'医疗'）；title 中的 {行业名称} 替换为具体行业名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX银行智能客服系统'，不超过35字）；sub_header 为项目名称（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该行业案例的背景、实施方案和应用效果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止使用匿名实体；禁止虚构数据。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "实践操作",
            "subtitle": "动手操作与代码示例",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "操作步骤",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "1", "title": "环境准备", "desc": "安装Python和必要的库"},
                {"num": "2", "title": "数据加载", "desc": "使用Pandas读取CSV/Excel"},
                {"num": "3", "title": "数据清洗", "desc": "处理缺失值和异常数据"},
                {"num": "4", "title": "数据分析", "desc": "分组聚合和统计计算"},
                {"num": "5", "title": "结果输出", "desc": "导出分析结果"}
            ],
            "notes": "5-6个操作步骤",
            "filling_prompt": "必须填入真实内容：提供5-6个具体操作步骤，每步有名称和一句话描述（如'步骤1：安装 kubectl'）。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 效果验证",
            "kpis": [
                {"value": "↓ 90%", "label": "数据处理时间缩短", "delta": "↓ 90%", "baseline": "vs Excel手动处理"},
                {"value": "↑ 40%", "label": "分析效率提升", "delta": "↑ 40%", "baseline": "vs 基础方法"},
                {"value": "99%", "label": "数据准确性", "delta": "↑ 稳定性", "baseline": "vs 人工处理"},
                {"value": "可复用", "label": "分析脚本复用", "delta": "↑ 效率", "baseline": "vs 一次性脚本"}
            ],
            "notes": "展示4个实践效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个实践效果指标，每个有 value（具体数字）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。指标要具体，如'部署时间缩短 90%'、'资源利用率提升 40%'。references 列出 web_search 获取的 URL（至少2个）。禁止保留花括号；禁止虚构数据。"
        },
        {
            "index": 12,
            "type": "example_detail",
            "title": "代码示例",
            "content_type": "example_detail",
            "kicker": "实例 · 代码示例",
            "lede": "使用Pandas进行数据清洗的完整代码示例",
            "context_block": "实际工作中，原始数据往往存在各种问题：缺失值、重复记录、格式不一致等。如何快速有效地清洗这些数据，是每个数据分析师必须掌握的技能。",
            "solution_block": "本示例展示了一个典型的数据清洗流程：1）使用drop_duplicates()去除重复记录；2）使用fillna()填充缺失值，可选择均值填充、前向填充或固定值填充；3）使用astype()转换数据类型；4）使用正则表达式清洗文本数据。整个流程可通过链式调用优雅地实现，代码简洁易读。",
            "metrics": [
                {"value": "约30行", "label": "示例代码规模", "trend": "简洁易读"},
                {"value": "↑ 80%", "label": "处理效率提升", "trend": "vs 手动处理"},
                {"value": "高", "label": "代码复用性", "trend": "可应用于其他项目"}
            ],
            "takeaway": "启示：掌握数据清洗的标准流程，提高数据分析效率",
            "notes": "关键代码或配置示例，配注释说明",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心功能；context_block 描述要解决的问题和使用场景（1-2句话）；solution_block 详细解释代码结构和关键逻辑（2-3句话）；metrics_grid 提供3个代码相关指标；takeaway 用一句话总结代码示例的核心价值。"
        },
        {
            "index": 13,
            "type": "example_detail",
            "title": "注意事项",
            "content_type": "example_detail",
            "kicker": "实例 · 避坑指南",
            "lede": "数据分析中最常见的5个错误及规避方法",
            "context_block": "初学者在使用Pandas时，常常因为对API不够熟悉或缺乏经验，犯一些常见错误，导致分析结果不准确或代码运行效率低下。",
            "solution_block": "常见错误及解决方案：1）修改副本而非原数据框——使用inplace=True或在赋值时注意返回值；2）循环遍历DataFrame——使用向量化操作替代循环；3）忽略缺失值——明确缺失值的处理策略并执行；4）数据类型不当——在分析前检查并转换数据类型；5）内存溢出——分块读取大文件或优化数据类型。",
            "metrics": [
                {"value": "60%", "label": "初学者踩坑率", "trend": "vs 常见情况"},
                {"value": "节省2小时", "label": "平均解决时间", "trend": "浪费的时间"},
                {"value": "严重", "label": "问题影响程度", "trend": "影响结果准确性"}
            ],
            "takeaway": "启示：养成良好的编码习惯，避免常见错误",
            "notes": "常见问题和避坑指南",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最容易踩的坑；context_block 描述问题的具体表现和发生场景（1-2句话）；solution_block 针对每个问题提供正确做法和解决方案（2-3句话）；metrics_grid 提供3个避坑相关指标；takeaway 用一句话总结如何避免常见错误。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "04",
            "title": "总结回顾",
            "subtitle": "要点与延伸",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 15,
            "type": "example_detail",
            "title": "知识要点",
            "content_type": "example_detail",
            "kicker": "实例 · 知识要点",
            "lede": "NumPy提供高效数组操作，Pandas提供强大的数据处理能力",
            "context_block": "本课程系统讲解了NumPy和Pandas的核心知识，包括数组创建、索引运算、数据框操作、缺失值处理等关键技能。",
            "solution_block": "核心要点回顾：1）NumPy ndarray是高效的多维数组，支持向量化运算；2）Pandas DataFrame是表格数据的利器，提供了丰富的数据操作方法；3）数据清洗是分析的前提，需要系统性地处理各种数据问题；4）代码可复用性很重要，多用函数封装常用逻辑。",
            "metrics": [
                {"value": "核心方法", "label": "NumPy方法数", "trend": "提升幅度"},
                {"value": "核心方法", "label": "Pandas方法数", "trend": "使用频率"},
                {"value": "推荐", "label": "延伸学习方向", "trend": "推荐深度"}
            ],
            "takeaway": "启示：多加练习，将技能内化为本能",
            "notes": "本章核心知识点回顾",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最重要的知识点；context_block 回顾本章核心内容（1-2句话）；solution_block 总结学习要点和关键收获（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结如何将知识运用到实际工作中。"
        },
        {
            "index": 16,
            "type": "image_text",
            "title": "延伸学习：数据可视化",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "延伸学习",
            "header": "Matplotlib数据可视化",
            "sub_header": "让数据说话的艺术",
            "paragraph": "数据分析的最终目的是从数据中获取洞察，而可视化是传达洞察的最佳方式。Matplotlib是Python最流行的可视化库，可以创建各种静态、动态、交互式的图表。建议学员在掌握Pandas后，继续学习Matplotlib和Seaborn，能够将分析结果以图表形式直观地展示出来。进阶方向还包括Plotly交互式图表、Bokeh动态可视化等。",
            "references": [
                "https://matplotlib.org/",
                "https://seaborn.pydata.org/"
            ],
            "notes": "用图文混排展示延伸学习方向",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'延伸学习'；title 中的 {学习方向} 替换为具体学习方向；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为学习方向标题（不超过35字）；sub_header 为学习概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该学习方向的背景、推荐理由和实践方法，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "下课了",
            "key_points": [
                "01 NumPy和Pandas是Python数据分析的基础",
                "02 数据清洗是数据分析的关键步骤",
                "03 延伸学习：Matplotlib数据可视化"
            ],
            "thank_you": "感谢聆听",
            "contact": "课程群：xxx | 课件下载：xxx@example.com",
            "notes": "结尾页",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（2个核心总结+1个延伸方向）；contact 填入讲师联系方式或课程群信息。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "课件要系统化，每个章节有明确的学习目标",
        "内容要循序渐进，从基础到进阶",
        "重点内容要突出，不要平均用力",
        "案例和实践部分要具体可操作",
        "结尾要有延伸学习方向",
        "配合代码演示，增加互动性",
        "设置思考题，引导学员主动思考",
        "提供课后练习，巩固学习效果"
    ],
    "teaching_notes": {
        "time_allocation": {
            "theory": "40% 理论讲解",
            "practice": "40% 动手实践",
            "discussion": "20% 互动讨论"
        },
        "interactive_elements": [
            "随堂测验：每章节后的小测验",
            "代码演示：现场编写和运行代码",
            "小组讨论：分组完成实践任务",
            "答疑环节：解答学员疑问"
        ],
        "assessment_methods": [
            "课堂表现：参与度和小测验",
            "实践作业：完成数据分析报告",
            "期末考试：综合能力测试"
        ]
    }
}
