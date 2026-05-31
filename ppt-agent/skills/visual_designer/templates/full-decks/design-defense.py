TEMPLATE = {
    "name": "design-defense",
    "name_cn": "答辩汇报",
    "description": "适合课程设计、毕业设计、项目答辩等场景。逻辑清晰，技术扎实，展示自信。",
    "target_audience": "答辩委员会、项目评审、导师、同学",
    "typical_slides": 16,
    "typical_duration": "15-20分钟（答辩）",
    "palette": "debate_purple",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "defense_strategy": "答辩的核心是展示：1）你做了什么；2）做得怎么样；3）为什么这样做",
        "key_principles": [
            "对工作了如指掌，技术细节要熟悉",
            "诚实面对不足，改进方向要清晰",
            "展示思考过程，而非仅仅展示结果",
            "自信但不自负，谦逊但不卑微"
        ],
        "time_allocation": {
            "presentation": "10-12分钟",
            "qa": "8-10分钟"
        }
    },
    "common_questions": {
        "technical": [
            "为什么选择这个技术方案？",
            "如何解决XX技术难点？",
            "系统的性能如何？",
            "有什么局限性？"
        ],
        "methodology": [
            "如何验证设计方案的合理性？",
            "参考了哪些相关工作？",
            "如何评估系统效果？"
        ],
        "future": [
            "如果继续完善，会做哪些改进？",
            "这个设计有什么扩展方向？"
        ]
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "基于深度学习的工业零件表面缺陷检测系统",
            "subtitle": "答辩人：王思远 | 指导老师：李明华教授",
            "author": "王思远",
            "date": "2025年6月",
            "notes": "标题页简洁，标题醒目，答辩信息完整",
            "filling_prompt": "必须填入真实内容：title 为论文或项目名称，subtitle 包含答辩人姓名和指导老师姓名，author 为答辩人姓名，date 为答辩日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "CONTENTS",
            "items": [
                "01  研究背景",
                "02  相关工作",
                "03  系统设计",
                "04  实验验证",
                "05  总结展望"
            ],
            "notes": "清晰展示答辩结构，5个章节让评委一目了然",
            "filling_prompt": "目录页为固定结构，可根据实际情况调整章节编号和名称。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "研究背景",
            "subtitle": "问题与动机",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "工业零件表面缺陷检测现状分析",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景 · 问题动机",
            "header": "质量检测是制造业的生命线",
            "sub_header": "零件表面缺陷直接影响产品安全与使用寿命",
            "paragraph": "在汽车、航空、精密仪器等行业，零件表面缺陷（如裂纹、划痕、凹坑、腐蚀等）可能导致严重安全事故。传统人工目检效率低、成本高、易疲劳，且难以保证检测一致性。据统计，我国制造业每年因表面缺陷造成的损失超过1200亿元，占总产值的1.5%以上。随着工业自动化程度提升，对高效、准确、低成本的缺陷检测方案需求日益迫切。深度学习技术的快速发展为解决这一难题提供了新的技术路径。",
            "notes": "右侧配工业零件缺陷场景图、工厂生产线图或缺陷样本示例",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为核心问题概述（不超过35字）；sub_header 为研究意义说明（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述研究背景、问题的严重性和研究的必要性，用流畅的段落形式呈现，禁止罗列要点，必须优先包含行业数据或技术指标。如果需要引用外部数据，通过 web_search 获取并在 references 列出 URL。禁止纯文字堆砌。"
        },
        {
            "index": 5,
            "type": "card_grid",
            "title": "传统检测方法的局限性",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "kicker": "现状分析 · 方法对比",
            "cards": [
                {
                    "header": "人工目检",
                    "body": "依赖经验、易疲劳、效率低，单件检测耗时3-5分钟，误检率达8-12%，无法满足批量生产需求"
                },
                {
                    "header": "传统机器视觉",
                    "body": "依赖人工设计特征，泛化能力差，换型号需重新调试，对复杂缺陷检出率不足70%"
                },
                {
                    "header": "传统机器学习方法",
                    "body": "需手工提取特征，特征选择依赖专家知识，面对新产品类别需重新训练模型，迁移能力弱"
                },
                {
                    "header": "深度学习方法（直接应用）",
                    "body": "计算资源消耗大，实时性差，直接部署成本高；公开数据集与实际工业场景存在域差异"
                }
            ],
            "notes": "2x2卡片展示四种方法及其局限性，形成问题→解决方案的铺垫",
            "filling_prompt": "必须填入真实内容：提供4个检测方法的局限性，每个有 header（方法名称）和 body（具体局限性描述，每条80-120字）。数据需通过 web_search 获取真实资料。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "相关工作",
            "subtitle": "文献综述与技术基础",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "timeline",
            "title": "缺陷检测算法技术演进",
            "content_type": "timeline",
            "kicker": "技术演进 · 发展脉络",
            "items": [
                {
                    "period": "2000-2010",
                    "title": "传统图像处理",
                    "desc": "边缘检测、形态学操作、纹理分析，依赖专家知识，泛化能力差"
                },
                {
                    "period": "2010-2015",
                    "title": "传统机器学习",
                    "desc": "SVM、随机森林、AdaBoost，配合HOG、LBP等特征提取方法"
                },
                {
                    "period": "2015-2019",
                    "title": "深度学习兴起",
                    "desc": "CNN在ImageNet取得突破，VGG、ResNet等骨干网络被引入工业检测"
                },
                {
                    "period": "2019-2022",
                    "title": "轻量化与实时性",
                    "desc": "MobileNet、EfficientNet、模型剪枝与量化技术使边缘部署成为可能"
                },
                {
                    "period": "2022-至今",
                    "title": "Transformer时代",
                    "desc": "Vision Transformer、Swin-Transformer在工业缺陷检测领域展现强大潜力"
                }
            ],
            "notes": "时间线从左到右或从上到下排列，每个节点标注年份范围、技术名称和简要描述",
            "filling_prompt": "必须填入真实内容：通过 web_search 获取缺陷检测领域的代表性方法和时间节点，每个节点有 period（年份范围）、title（技术名称）、desc（20-40字描述）。确保时间线逻辑连贯、数据真实。"
        },
        {
            "index": 8,
            "type": "two_column",
            "title": "主流方法对比分析",
            "content_type": "two_column",
            "kicker": "相关工作 · 技术调研",
            "left_header": "现有方案",
            "left_sections": {
                "analysis": [
                    "Faster R-CNN系列：检测精度较高，但推理速度慢（15-30FPS），对小缺陷不敏感",
                    "YOLO系列：速度快，但对小目标和重叠缺陷召回率低，难以处理多尺度问题",
                    "U-Net语义分割：像素级分割精度高，但计算开销大，边缘模糊问题突出"
                ],
                "data": [
                    "检测精度：78-92%（依场景而异）",
                    "推理速度：5-30 FPS",
                    "参数量：20-100M"
                ]
            },
            "right_header": "本方案优势",
            "right_sections": {
                "key_points": [
                    "针对工业零件缺陷定制特征提取网络，小目标检测能力提升40%",
                    "轻量化设计：参数量压缩至8.5M，支持边缘设备实时推理",
                    "域自适应策略：减少公开数据集与实际场景的域差异影响",
                    "多尺度特征融合：提升不同尺寸缺陷的检出率"
                ],
                "data": [
                    "检测精度：96.8% mAP",
                    "推理速度：52 FPS",
                    "参数量：8.5M（压缩率达91%）"
                ]
            },
            "references": [
                "https://arxiv.org/abs/1703.00151",
                "https://arxiv.org/abs/2005.09558"
            ],
            "notes": "左右对比现有技术方案与本设计的技术优势，每个维度用数据支撑",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_header 为'现有方案'，left_sections.analysis 列出2-3个现有方案及其技术局限，left_sections.data 列出量化指标；right_header 为'本方案优势'，right_sections.key_points 列出本设计的创新点和优势，right_sections.data 列出对应的技术指标。注意左右数据形成明确对比。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "系统设计",
            "subtitle": "系统架构与核心设计",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "process_flow",
            "title": "系统整体架构",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "kicker": "架构设计 · 流水线",
            "steps": [
                {"num": "01", "title": "图像采集", "desc": "工业相机采集零件表面图像，分辨率2048x2048，支持在线与离线两种模式"},
                {"num": "02", "title": "图像预处理", "desc": "自适应直方图均衡化、噪声滤波、亮度归一化，消除光照不均影响"},
                {"num": "03", "title": "特征提取", "desc": "轻量化骨干网络（自研SlimNet），多尺度特征金字塔融合"},
                {"num": "04", "title": "缺陷分类", "desc": "三级分类：有无缺陷 → 缺陷类型 → 缺陷等级，支持6种缺陷类别"},
                {"num": "05", "title": "结果输出", "desc": "可视化标注、缺陷定位热力图、置信度评分，支持JSON与图片双格式导出"}
            ],
            "notes": "横向或锯齿形流程图，5个步骤从左到右依次展示",
            "filling_prompt": "必须填入真实内容：提供5个系统处理步骤，每个有 num（编号）、title（步骤名称）、desc（具体描述，每条不超过60字）。流程要与实际系统设计一致。"
        },
        {
            "index": 11,
            "type": "deep_dive",
            "title": "核心检测模块详解",
            "content_type": "deep_dive",
            "kicker": "详解 · 核心模块",
            "lede": "专为工业小目标缺陷设计的轻量化特征提取网络",
            "left_column": {
                "key_points": [
                    "SlimNet骨干网络：3层深度可分离卷积替代标准卷积，参数量减少78%",
                    "多尺度特征金字塔：融合4个不同感受野的特征图，覆盖从8px到256px缺陷检测",
                    "注意力机制：CBAM模块增强缺陷区域特征响应，抑制背景干扰",
                    "数据增强策略：MixUp、CutMix、随机旋转与翻转扩充训练集至10倍",
                    "难例挖掘：在线难例挖掘策略提升对困难样本的检测能力"
                ],
                "analysis": [
                    "检测能力：从像素级缺陷（8px）到区域级缺陷全覆盖",
                    "泛化能力：域自适应训练减少场景迁移性能损失"
                ]
            },
            "right_column": {
                "case_example": [
                    "案例1：某汽车发动机缸体表面裂纹检测，检出率从87%提升至97.3%",
                    "案例2：精密轴承表面凹坑检测，误报率从12%降至2.1%",
                    "案例3：PCB板焊点缺陷检测，检测速度从0.3s/件提升至0.02s/件",
                    "案例4：注塑件飞边缺陷检测，准确率达98.7%，支持6种缺陷类型"
                ],
                "data_evidence": [
                    "mAP@0.5：96.8%（行业领先水平）",
                    "推理速度：52 FPS（RTX 3060实时推理）",
                    "模型大小：8.5MB（可部署至Jetson Nano边缘设备）"
                ]
            },
            "notes": "双栏展示，左栏讲设计要点，右栏放案例数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 和 title 中的核心模块替换为具体模块名称；lede 为一句话说明；left_column.key_points 为设计要点（3-5条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 URL。"
        },
        {
            "index": 12,
            "type": "content_slide",
            "title": "轻量化模型设计",
            "content_type": "content_slide",
            "kicker": "技术细节 · 模型压缩",
            "layout_hint": "icon_text_list",
            "header": "面向边缘部署的模型压缩技术",
            "sections": [
                {
                    "title": "深度可分离卷积",
                    "icon": "layers",
                    "body": "将标准卷积分解为逐通道卷积与逐点卷积，参数量减少78%，计算量降低88%，几乎不损失精度"
                },
                {
                    "title": "模型剪枝",
                    "icon": "scissors",
                    "body": "基于L1范数的通道剪枝策略，移除冗余通道，设计空间搜索自动确定最优剪枝率，整体剪枝率35%"
                },
                {
                    "title": "权重量化",
                    "icon": "hash",
                    "body": "FP32→INT8量化，采用KL散度校准方法，均值绝对误差控制在0.5%以内，模型体积缩小4倍"
                },
                {
                    "title": "知识蒸馏",
                    "icon": "graduation-cap",
                    "body": "以高精度大模型为教师，压缩后模型为学生，蒸馏温度T=4，软标签损失权重0.3，精度恢复至96.2%"
                }
            ],
            "notes": "带图标的列表式布局，每个技术点配图标，简明扼要展示4种模型压缩技术",
            "filling_prompt": "必须填入真实内容：通过 web_search 获取模型压缩技术的最新研究进展，header 为页面标题（不超过35字）；sections 列出4种轻量化技术，每种有 title（技术名称）、icon（图标名称）、body（30-60字描述）。确保技术描述准确，数据真实可信。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "04",
            "title": "实验验证",
            "subtitle": "开发过程与验证结果",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 14,
            "type": "kpi_dashboard",
            "title": "测试结果与性能分析",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 测试验证",
            "kpis": [
                {
                    "value": "96.8%",
                    "label": "mAP@0.5",
                    "delta": "+8.3%",
                    "baseline": "vs YOLOv8-S (88.5%)"
                },
                {
                    "value": "52 FPS",
                    "label": "推理速度",
                    "delta": "+3x",
                    "baseline": "vs Faster R-CNN (15 FPS)"
                },
                {
                    "value": "8.5MB",
                    "label": "模型大小",
                    "delta": "-92%",
                    "baseline": "vs ResNet50 (98MB)"
                },
                {
                    "value": "98.2%",
                    "label": "实际产线准确率",
                    "delta": "+11.2%",
                    "baseline": "vs 人工目检 (87%)"
                }
            ],
            "notes": "2x2布局，用数据证明系统效果，每个指标要有对比基准",
            "filling_prompt": "必须填入真实内容：提供4个测试指标（如准确率、速度、模型大小、产线实测数据等），每个有 value、label、delta、baseline。禁止虚构数据，数据需来自实际测试或 web_search 权威来源。"
        },
        {
            "index": 15,
            "type": "chart_slide",
            "title": "消融实验结果",
            "content_type": "chart_slide",
            "kicker": "消融实验 · 模块贡献",
            "chart_title": "各模块对检测性能的贡献度分析",
            "chart_type": "bar",
            "categories": [
                "基线模型",
                "+多尺度FPN",
                "+CBAM注意力",
                "+数据增强",
                "+难例挖掘",
                "+完整模型"
            ],
            "series": [
                {
                    "name": "mAP@0.5 (%)",
                    "data": [82.3, 88.7, 91.4, 93.8, 95.2, 96.8]
                },
                {
                    "name": "召回率 (%)",
                    "data": [78.5, 85.2, 88.9, 92.1, 94.3, 95.7]
                }
            ],
            "x_axis_label": "实验配置",
            "y_axis_label": "性能指标 (%)",
            "notes": "柱状图展示逐个添加模块后的性能提升，直观展示每个模块的边际贡献",
            "filling_prompt": "必须填入真实内容：chart_title 为图表标题；categories 列出6个实验配置（基线→最终模型）；series 包含2条数据线，每条有 name 和 6个 data 数值；x_axis_label 和 y_axis_label 标注坐标轴。数据必须来自实际消融实验结果，禁止虚构。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "答辩完毕",
            "key_points": [
                "01 提出了SlimNet轻量化骨干网络，参数量压缩至8.5MB，支持边缘设备实时推理",
                "02 设计了多尺度特征融合与CBAM注意力机制，mAP@0.5达96.8%，领先同类方法",
                "03 构建了完整的工业零件缺陷检测系统，在3家企业的实际产线上验证了有效性"
            ],
            "thank_you": "感谢各位老师的耐心聆听，欢迎批评指正！",
            "notes": "简洁有力的结尾，致谢评委，3条核心贡献要清晰有力",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心贡献2条+创新亮点1条），每条不超过60字。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "答辩要逻辑清晰，重点突出",
        "技术细节要扎实，数据要真实",
        "对自己方案的优缺点要有清晰认识",
        "预判评委问题，准备好应答",
        "PPT页数适中，留出答辩时间",
        "PPT配色统一，避免花哨",
        "每页只讲一个核心观点",
        "准备Q&A，预设问题清单"
    ],
    "qa_preparation": {
        "technical_prep": [
            "为什么选择这个算法？",
            "如何处理数据不平衡？",
            "模型轻量化方案？",
            "系统的局限性？"
        ],
        "methodology_prep": [
            "如何评估系统性能？",
            "训练数据来源？",
            "消融实验设计？"
        ],
        "tips": [
            "不会的问题诚实说不知道",
            "不确定的问题可以说'我认为...'",
            "适时引导问题到熟悉的领域",
            "保持冷静，不要争辩"
        ]
    }
}
