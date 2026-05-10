TEMPLATE = {
    "name": "design-defense",
    "name_cn": "答辩汇报",
    "description": "适合课程设计、毕业设计、项目答辩等场景。逻辑清晰，技术扎实，展示自信。",
    "target_audience": "答辩委员会、项目评审、导师、同学",
    "typical_slides": 12,
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
            "title": "基于深度学习的图像识别系统设计",
            "subtitle": "张三 | 指导老师：李教授",
            "author": "张三",
            "date": "2025年5月20日",
            "notes": "标题页简洁，标题醒目",
            "filling_prompt": "必须填入真实内容：title 为项目或论文名称，subtitle 包含答辩人姓名和指导老师，author 为答辩人，date 为答辩日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  研究背景",
                "02  相关工作",
                "03  设计方案",
                "04  实现与测试",
                "05  总结展望"
            ],
            "notes": "清晰展示答辩结构",
            "filling_prompt": "目录页为固定结构，可根据实际情况调整章节。"
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
            "title": "研究背景与问题",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景 · 问题动机",
            "header": "工业质检场景下的缺陷检测挑战",
            "sub_header": "人工质检效率低、漏检率高、成本高昂",
            "paragraph": "随着制造业向智能化转型，产品质量检测成为关键环节。传统的人工质检面临诸多挑战：1）效率低：人工检测速度有限，难以满足高速生产线的需求；2）漏检率高：人眼长时间工作容易疲劳，漏检率可达5-10%；3）成本高：企业需雇佣大量质检员，人力成本持续上升。本研究旨在利用深度学习技术，开发一套自动化缺陷检测系统，提高质检效率和准确率。",
            "notes": "右侧配问题场景图/现有方案对比图/技术趋势图，左侧阐述问题和动机",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为核心问题概述（不超过35字）；sub_header 为研究意义说明（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述研究背景、问题的严重性和研究的必要性，用流畅的段落形式呈现，禁止罗列要点，必须优先包含行业数据或技术指标。如果需要引用外部数据，通过 web_search 获取并在 references 列出 URL。禁止纯文字堆砌。",
            "research_significance": "本研究的意义在于：1）提高质检效率，降低企业成本；2）减少人为误差，提高产品质量；3）为制造业智能化转型提供技术支撑"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "相关工作",
            "subtitle": "文献综述与技术基础",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "two_column",
            "title": "相关技术与方法",
            "content_type": "two_column",
            "kicker": "相关工作 · 技术调研",
            "left_header": "现有方案",
            "left_sections": {
                "analysis": [
                    "传统机器学习方法（如SVM、随机森林等）：依赖人工设计特征提取算法（如HOG、LBP等），在光照变化、遮挡等复杂场景下鲁棒性差，工业缺陷检测准确率通常在80-85%之间，难以满足现代生产线99%以上的品控要求",
                    "传统CNN方法：模型层数较浅（通常5-10层），表达能力有限，对细微缺陷（如微裂纹、轻微色差等）识别能力弱，且需要大量标注数据，训练成本高、周期长",
                    "传统目标检测算法（如R-CNN系列）：推理速度慢（通常>100ms/张），无法满足高速生产线的实时检测需求（通常要求<50ms/张）"
                ],
                "data": [
                    "传统方法准确率：80-85%",
                    "传统CNN推理速度：>100ms/张",
                    "所需训练样本：>5000张/类"
                ]
            },
            "right_header": "本方案优势",
            "right_sections": {
                "key_points": [
                    "深度迁移学习策略：利用ImageNet预训练的ResNet50作为backbone，只用500张/类的样本即可达到96%以上的准确率，大幅降低数据采集成本",
                    "针对性模型微调：冻结浅层特征提取器，复用通用视觉表征，仅微调高层语义特征层，训练周期从传统方法的数周缩短至数小时",
                    "实时推理优化：采用YOLOv8架构结合模型剪枝和量化技术，单图推理时间压缩至25ms，满足生产线实时检测需求",
                    "数据增强策略：引入随机旋转、翻转、亮度调整、噪声注入等增强手段，提升模型对复杂工业环境的泛化能力"
                ],
                "data": [
                    "本方案准确率：96.8%",
                    "本方案推理速度：25ms/张",
                    "所需训练样本：500张/类"
                ]
            },
            "references": [
                "https://arxiv.org/abs/1512.03385",
                "https://arxiv.org/abs/2003.05664"
            ],
            "notes": "左右对比现有技术方案与本设计的技术优势，每个维度用数据支撑",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_header 为'现有方案'，left_sections.analysis 列出2-3个现有方案及其技术局限，left_sections.data 列出量化指标；right_header 为'本方案优势'，right_sections.key_points 列出本设计的创新点和优势，right_sections.data 列出对应的技术指标。注意左右数据形成明确对比。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "设计方案",
            "subtitle": "系统架构与核心设计",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "deep_dive",
            "title": "缺陷检测模块设计",
            "content_type": "deep_dive",
            "kicker": "详解 · 缺陷检测模块",
            "lede": "基于ResNet50的迁移学习+数据增强策略，实现高精度缺陷检测",
            "left_column": {
                "key_points": [
                    "迁移学习：使用ImageNet预训练模型",
                    "模型微调：冻结浅层，训练高层",
                    "数据增强：旋转、翻转、亮度调整",
                    "损失函数：Focal Loss处理类别不平衡",
                    "迁移能力：只需少量样本即可适应新品"
                ],
                "analysis": [
                    "相比从头训练：收敛更快，效果更好",
                    "相比传统方法：特征自动学习，准确率更高"
                ]
            },
            "right_column": {
                "case_example": [
                    "输入：工业产品图片（256x256）",
                    "预处理：归一化、数据增强",
                    "特征提取：ResNet50 backbone",
                    "分类头：全连接层+Softmax"
                ],
                "data_evidence": [
                    "Top-1准确率：96.8%",
                    "推理速度：25ms/张",
                    "训练数据需求：500张/类"
                ]
            },
            "notes": "双栏展示，左栏讲设计要点，右栏放案例数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 和 title 中的 {核心模块} 替换为具体模块名称；lede 为一句话说明；left_column.key_points 为设计要点（3-5条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "04",
            "title": "实现与测试",
            "subtitle": "开发过程与验证结果",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "kpi_dashboard",
            "title": "测试结果",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 测试验证",
            "kpis": [
                {"value": "96.8%", "label": "准确率", "delta": "↑ 显著", "baseline": "vs 传统方法85%"},
                {"value": "99.2%", "label": "召回率", "delta": "↑ 优秀", "baseline": "vs 传统方法90%"},
                {"value": "25ms", "label": "单图推理时间", "delta": "↓ 满足实时", "baseline": "目标<50ms"},
                {"value": "4类", "label": "可检测缺陷类型", "delta": "扩展性强", "baseline": "支持新品类"}
            ],
            "notes": "用数据证明系统效果",
            "filling_prompt": "必须填入真实内容：提供4个测试指标（如功能测试覆盖率、性能指标、用户体验评分、错误率等），每个有 value、label、delta、baseline。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "example_detail",
            "title": "总结与展望",
            "content_type": "example_detail",
            "kicker": "实例 · 总结展望",
            "lede": "基于深度学习的缺陷检测系统，准确率达96.8%，满足工业应用需求",
            "context_block": "本研究完成了缺陷检测系统的设计与实现，包括数据采集、模型训练、系统部署等完整流程。系统在某工厂试运行3个月，取得了良好效果。",
            "solution_block": "主要成果包括：1）设计了基于迁移学习的缺陷检测模型，准确率比传统方法提升10%以上；2）开发了完整的检测系统，支持实时检测和结果可视化；3）在实际工业场景中验证了系统的有效性。未来计划：1）扩展更多缺陷类型；2）优化模型轻量化；3）开发WEB端管理界面。",
            "metrics": [
                {"value": "100%", "label": "功能完成度", "trend": "vs 计划100%"},
                {"value": "3项", "label": "创新点", "trend": "方法创新"},
                {"value": "良好", "label": "评审评分", "trend": "答辩表现"}
            ],
            "takeaway": "启示：本研究为工业质检智能化提供了可行方案，具有一定的应用价值",
            "notes": "左右对比已完成工作和未来计划",
            "filling_prompt": "必须填入真实内容：lede 一句话概括核心成果；context_block 回顾完成内容（1-2句话）；solution_block 总结核心贡献和未来方向（2-3句话）；metrics_grid 提供3个总结指标；takeaway 用一句话总结对后续工作的指导意义。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "答辩完毕",
            "key_points": [
                "01 提出了基于迁移学习的缺陷检测方法",
                "02 系统准确率达96.8%，满足工业应用需求",
                "03 在实际场景验证了方案有效性"
            ],
            "thank_you": "感谢各位老师！",
            "notes": "简洁有力的结尾，致谢评委",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心贡献2条+创新亮点1条）。禁止保留花括号。"
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
