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
        "knowledge": "掌握Python数据分析的核心工具NumPy和Pandas",
        "skills": "能够独立完成从数据加载、清洗到可视化的完整分析流程",
        "attitude": "建立数据驱动思维，善于从数据中发现规律和洞察"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "Python数据分析基础",
            "subtitle": "NumPy · Pandas · 数据可视化",
            "author": "讲师：陈宇明",
            "date": "2026年5月23日",
            "filling_prompt": "必须填入真实内容：title 为实际课程或章节名称，subtitle 为一句课程简介，author 为讲师姓名，date 为日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  基础知识 — 核心概念与工具",
                "02  进阶应用 — 典型场景与行业案例",
                "03  实践操作 — 动手完成数据分析全流程",
                "04  总结回顾 — 要点提炼与延伸方向"
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
            "subtitle": "核心概念与工具",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "学习目标",
            "content_type": "example_detail",
            "kicker": "实例 · 学习目标",
            "lede": "掌握NumPy和Pandas两大核心工具，独立完成数据加载、清洗、分析与可视化全流程",
            "context_block": "数据分析已成为各行各业的核心竞争力。然而很多初学者在学习过程中，常被庞杂的API参数淹没，学完后仍不知道如何从零开始完成一个完整的分析项目。传统课程往往将NumPy和Pandas割裂讲解，忽略了它们在实际项目中的协同应用。",
            "solution_block": "本课程采用「项目驱动」的教学设计，以一个完整的「电商用户行为分析」项目为载体，将NumPy和Pandas的知识点融入实际分析流程中。学员不仅学会单个工具的使用方法，更重要的是理解何时该用哪个工具、如何将两者结合使用。通过学练结合的方式，确保每个知识点都能落地。",
            "metrics": [
                {"value": "85%", "label": "预计掌握程度", "trend": "vs 课程前（20%）"},
                {"value": "90分钟", "label": "预计所需时间", "trend": "理论40%+实践60%"},
                {"value": "1个", "label": "配套实战项目", "trend": "电商用户行为分析"}
            ],
            "takeaway": "学完本章后，你将能够独立使用Python完成大多数日常数据分析任务。",
            "notes": "本章节需要掌握的3-4个核心知识点",
            "filling_prompt": "必须填入真实内容：lede 一句话说明本章的核心价值；context_block 描述学员常见的困惑或误区（1-2句话）；solution_block 说明本章内容如何帮助解决（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结学完本章的核心收获。禁止保留花括号。"
        },
        {
            "index": 5,
            "type": "deep_dive",
            "title": "核心机制详解",
            "content_type": "deep_dive",
            "kicker": "详解 · 核心机制",
            "lede": "NumPy与Pandas：数组运算与表格处理的底层逻辑",
            "left_column": {
                "key_points": [
                    "NumPy ndarray：同质多维数组，内存连续存储，计算效率比Python list高10-100倍",
                    "向量化运算：无需显式循环，对整列/整行数据同时执行数学运算",
                    "广播机制：形状不同的数组之间自动扩展对齐，支持灵活的数据操作",
                    "索引与切片：通过位置或条件快速筛选数据子集"
                ],
                "analysis": [
                    "性能维度：NumPy底层由C编写，GIL释放后真正并行计算",
                    "内存维度：连续内存布局配合SIMD指令，大幅减少缓存未命中"
                ]
            },
            "right_column": {
                "case_example": [
                    "场景：计算100万行数据的日环比增长率",
                    "原始写法：for循环逐行计算，约需8秒",
                    "向量化写法：df['pct_change']，约需0.02秒",
                    "性能提升：400倍"
                ],
                "data_evidence": [
                    "NumPy数组索引速度：比Python list快5倍",
                    "Pandas merge性能：10万行数据合并<1秒",
                    "Matplotlib渲染速度：1000个数据点绑图<0.1秒"
                ]
            },
            "notes": "双栏深入展开，左栏放核心要点和分析，右栏放案例和数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_column.key_points 为核心要点（3-4条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "two_column",
            "title": "知识点对比",
            "content_type": "two_column",
            "layout_hint": "left-numpy right-pandas",
            "left_header": "NumPy — 数值计算基石",
            "left_items": [
                {"role": "定位", "name": "基础数值计算库", "responsibility": "提供高性能多维数组和矩阵运算"},
                {"role": "核心数据结构", "name": "ndarray（同质数组）", "responsibility": "所有元素类型一致，内存连续存储"},
                {"role": "擅长场景", "name": "数值运算、矩阵计算、科学计算", "responsibility": "图像处理、信号处理、统计分析"},
                {"role": "API风格", "name": "函数式调用为主", "responsibility": "np.mean()、np.sum()、np.dot()"}
            ],
            "right_header": "Pandas — 表格数据分析",
            "right_items": [
                {"role": "定位", "name": "表格数据分析库", "responsibility": "提供DataFrame和Series，处理结构化数据"},
                {"role": "核心数据结构", "name": "DataFrame（表格）+ Series（一维）", "responsibility": "列可异质，带索引标签，灵活易用"},
                {"role": "擅长场景", "name": "数据清洗、探索性分析、时间序列", "responsibility": "数据合并、分组聚合、缺失值处理"},
                {"role": "API风格", "name": "链式方法调用", "responsibility": "df.groupby().agg().sort_values()"}
            ],
            "notes": "对比NumPy和Pandas的定位与适用场景，帮助学员理解何时用哪个工具",
            "filling_prompt": "必须填入真实内容：对比NumPy和Pandas的核心差异，每列列出3-4个对比维度，每个维度有role、name和responsibility。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "02",
            "title": "进阶应用",
            "subtitle": "典型场景与行业案例",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 8,
            "type": "card_grid",
            "title": "应用场景",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "用户行为分析", "body": "分析用户在产品中的点击、浏览、购买等行为路径，识别高价值用户群体与流失节点。"},
                {"header": "销售数据监控", "body": "实时追踪各地区、各品类的销售数据，发现异常波动，自动生成日报/周报。"},
                {"header": "A/B测试分析", "body": "对实验组与对照组数据进行统计检验，计算显著性水平，评估产品改动的实际效果。"},
                {"header": "金融风控建模", "body": "基于历史交易数据构建特征工程，通过Pandas进行数据清洗和聚合，为机器学习模型准备训练数据。"}
            ],
            "notes": "4个典型应用场景",
            "filling_prompt": "必须填入真实内容：提供4个典型应用场景，每个有 header（场景名称）和 body（一句话描述）。场景要具体。"
        },
        {
            "index": 9,
            "type": "image_text",
            "title": "行业案例",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "行业 · 电商数据分析",
            "header": "某头部电商平台用户复购分析实践",
            "sub_header": "基于Pandas的千万级用户行为数据挖掘项目",
            "paragraph": "某头部电商平台数据分析团队曾面临一个典型挑战：如何从每年数十亿条用户行为日志中，快速识别高潜力复购用户并制定差异化运营策略。该团队使用Pandas对用户的浏览-收藏-加购-下单全链路数据进行清洗和关联，剔除了机器人流量和测试账号后，保留了约1200万真实用户的完整行为数据。通过groupby+agg构建了超过200个用户特征，包括最近购买间隔、平均订单金额、品类偏好等维度。最终结合逻辑回归模型，将复购预测准确率从运营经验判断的40%提升至73%，使营销活动的投入产出比提升了2.3倍。这一案例充分展示了Pandas在大规模数据处理中的工程价值，以及数据驱动决策的巨大潜力。",
            "references": [
                "https://pandas.pydata.org/docs/user_guide/10min.html",
                "https://pandas.pydata.org/docs/cookbook.html"
            ],
            "notes": "用图文混排展示行业应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为项目名称；paragraph 为300-450字的自然语言段落，详细描述该行业案例的背景、实施方案和应用效果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "03",
            "title": "实践操作",
            "subtitle": "动手操作与代码示例",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 11,
            "type": "process_flow",
            "title": "操作步骤",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "1", "title": "环境准备", "desc": "安装Anaconda，创建虚拟环境，安装pandas、numpy、matplotlib"},
                {"num": "2", "title": "数据加载", "desc": "使用pd.read_csv()读取CSV文件，head()查看数据结构，info()了解字段信息"},
                {"num": "3", "title": "数据清洗", "desc": "处理缺失值(dropna/fillna)，数据类型转换，去除重复行，异常值过滤"},
                {"num": "4", "title": "特征工程", "desc": "通过apply()和lambda构建新特征，groupby+agg进行数据聚合统计"},
                {"num": "5", "title": "分析与可视化", "desc": "使用matplotlib/seaborn绑制趋势图、分布图、相关性热力图"},
                {"num": "6", "title": "导出报告", "desc": "将分析结果保存为Excel报表，图表导出为PNG图片"}
            ],
            "notes": "5-6个操作步骤",
            "filling_prompt": "必须填入真实内容：提供5-6个具体操作步骤，每步有名称和一句话描述。"
        },
        {
            "index": 12,
            "type": "kpi_dashboard",
            "title": "效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 实践效果验证",
            "kpis": [
                {"value": "73%", "label": "复购预测准确率", "delta": "+33pp", "baseline": "vs 经验判断（40%）"},
                {"value": "2.3x", "label": "营销ROI提升", "delta": "+130%", "baseline": "vs 随机投放（1.0x）"},
                {"value": "1200万", "label": "分析用户规模", "delta": "覆盖全年活跃用户", "baseline": "剔除机器人后"},
                {"value": "200+", "label": "用户特征维度", "delta": "覆盖行为全链路", "baseline": "浏览→收藏→购买"}
            ],
            "notes": "展示4个实践效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个实践效果指标，每个有 value、label、delta、baseline。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 13,
            "type": "example_detail",
            "title": "代码示例",
            "content_type": "example_detail",
            "kicker": "实例 · 核心代码",
            "lede": "使用Pandas chain式调用，完成从数据加载到分组聚合的完整操作",
            "context_block": "在日常数据分析中，最常见的操作是对表格数据按某个维度进行分组，然后对各组计算多个统计指标。传统写法需要多步临时变量，不仅代码冗长，还容易出错。Pandas的链式方法调用（method chaining）可以让代码更简洁、更易读，也更易于调试。",
            "solution_block": "以下代码展示了如何使用Pandas的链式调用完成用户复购分析的核心特征构建：首先用read_csv加载数据，然后用query过滤有效订单，接着通过groupby按用户ID分组并使用agg同时计算购买次数、消费总额和平均订单金额，最后用sort_values按消费总额降序排列，取出Top10高价值用户。全程无需创建任何中间变量，代码一气呵成。",
            "metrics": [
                {"value": "3行", "label": "核心代码量", "trend": "vs 传统写法（12行）"},
                {"value": "5x", "label": "代码简洁度提升", "trend": "方法链 vs 临时变量"},
                {"value": "95%", "label": "代码可复用性", "trend": "参数化后可跨项目使用"}
            ],
            "takeaway": "链式方法调用是Pandas最优雅的使用方式，能够大幅提升数据分析代码的可读性和维护效率。",
            "notes": "关键代码或配置示例，配注释说明",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心功能；context_block 描述要解决的问题和使用场景（1-2句话）；solution_block 详细解释代码结构和关键逻辑（2-3句话）；metrics_grid 提供3个代码相关指标；takeaway 用一句话总结代码示例的核心价值。"
        },
        {
            "index": 14,
            "type": "example_detail",
            "title": "注意事项",
            "content_type": "example_detail",
            "kicker": "实例 · 避坑指南",
            "lede": "初学者最常踩的坑：链式调用中的引用拷贝与内存爆炸",
            "context_block": "在使用Pandas处理大规模数据时，初学者最常遇到两个问题：一是链式调用中使用了视图而非拷贝，导致对数据的修改意外影响原始DataFrame；二是对大数据集进行不当操作（如不小心复制了整个DataFrame），导致内存溢出。这两个问题在开发环境中不易察觉，但在生产环境下会导致数据损坏或程序崩溃。",
            "solution_block": "针对第一个问题：Pandas默认的切片操作返回视图（view），而非拷贝。在链式调用中使用.assign()或.copy()显式创建拷贝，确保每个步骤操作的是独立数据副本。针对第二个问题：善用chunk参数分批读取大文件，或使用dtype指定字段类型减少内存占用。对于超过内存限制的场景，优先使用Dask或Polars等支持外存计算的库。",
            "metrics": [
                {"value": "67%", "label": "初学者踩坑率", "trend": "在处理10万行以上数据时"},
                {"value": "30分钟", "label": "平均排查时间", "trend": "视图vs拷贝问题定位"},
                {"value": "高", "label": "内存溢出风险", "trend": "未指定dtype时的内存膨胀"}
            ],
            "takeaway": "养成使用.copy()显式拷贝和指定dtype的习惯，是避免数据损坏和内存问题的关键。",
            "notes": "常见问题和避坑指南",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最容易踩的坑；context_block 描述问题的具体表现和发生场景（1-2句话）；solution_block 针对每个问题提供正确做法和解决方案（2-3句话）；metrics_grid 提供3个避坑相关指标；takeaway 用一句话总结如何避免常见错误。"
        },
        {
            "index": 15,
            "type": "section_divider",
            "number": "04",
            "title": "总结回顾",
            "subtitle": "要点与延伸",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 16,
            "type": "three_column",
            "title": "知识要点",
            "content_type": "three_column",
            "layout_hint": "3-column",
            "col1_header": "NumPy核心",
            "col1_items": [
                "ndarray多维数组，元素同质、内存连续",
                "向量化运算：无需for循环，高效批量计算",
                "广播机制：不同形状数组自动对齐扩展",
                "常用函数：np.mean/sum/dot/concatenate"
            ],
            "col2_header": "Pandas核心",
            "col2_items": [
                "DataFrame为二维表格，Series为一维序列",
                "链式方法：.query().groupby().agg().sort_values()",
                "数据清洗：dropna/fillna/drop_duplicates",
                "时间序列：pd.to_datetime + resample"
            ],
            "col3_header": "最佳实践",
            "col3_items": [
                "始终使用.copy()创建数据副本",
                "指定dtype减少内存占用",
                "优先链式调用，避免临时变量",
                "分批处理大数据集，使用chunk参数"
            ],
            "notes": "三栏对比总结本章核心知识点",
            "filling_prompt": "必须填入真实内容：col1_header 列出NumPy核心要点（3-4条）；col2_header 列出Pandas核心要点（3-4条）；col3_header 列出最佳实践（3-4条）。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "下课了",
            "key_points": [
                "01 核心收获：掌握NumPy数组运算与Pandas表格分析，完成数据加载→清洗→聚合→可视化的完整链路",
                "02 延伸方向：学习数据可视化（Seaborn/Plotly）、探索性数据分析（EDA）、特征工程与机器学习基础",
                "03 实战项目：完成「电商用户行为分析」项目，将课程所学应用到真实业务场景中"
            ],
            "thank_you": "感谢聆听，欢迎提问与交流！",
            "contact": "课程群：Python数据分析交流群 | 课件获取：data-course@company.com",
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
            "随堂测验",
            "代码演示",
            "小组讨论",
            "答疑环节"
        ],
        "assessment_methods": [
            "课堂表现：参与度和小测验",
            "实践作业：完成分析报告",
            "期末考试：综合能力测试"
        ]
    }
}
