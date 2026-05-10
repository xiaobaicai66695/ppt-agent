TEMPLATE = {
    "name": "tech-sharing",
    "name_cn": "技术分享",
    "description": "适合内部技术分享、技术培训、架构讲解等场景。结构清晰，有章节划分，注重内容深度。",
    "target_audience": "工程师、技术管理者、技术爱好者",
    "typical_slides": 18,
    "typical_duration": "30-45分钟",
    "palette": "ocean_soft",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "opening_hook": "用一个实际的技术问题或挑战开场，引起工程师共鸣",
        "key_moments": [
            "问题分析时：展示具体的问题场景和代码",
            "原理讲解时：配合架构图和数据流图",
            "实践环节：展示真实的代码示例和运行结果",
            "Q&A环节：预留时间解答技术细节"
        ],
        "closing_strength": "总结核心知识点，提供延伸学习资源"
    },
    "prerequisites": {
        "required_knowledge": "本次分享假设听众具备基本的编程能力和相关技术基础",
        "optional_prep": "建议听众提前阅读相关技术文档",
        "materials_provided": "演讲结束后将分享演讲稿和代码示例"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "Kubernetes架构深度解析",
            "subtitle": "云原生时代的容器编排之道",
            "author": "李明 | 基础架构部",
            "date": "2025年3月10日",
            "notes": "开场标题页，留白充足，标题字体有重量感",
            "filling_prompt": "必须填入真实内容：title 为本次分享的实际技术主题名称（如'Kubernetes架构深度解析'），author 为演讲者姓名，date 为实际日期。禁止保留花括号占位符。",
            "visual_suggestions": "可添加抽象的集群拓扑图或容器示意图作为背景"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  背景与问题",
                "02  核心原理",
                "03  架构设计",
                "04  实践案例",
                "05  总结与展望"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "目录页为固定结构，无需额外填充。",
            "timing_hint": "约30秒"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "背景与问题",
            "subtitle": "为什么需要容器编排",
            "notes": "章节分隔页，仪式感强",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "design_notes": "章节标题使用大字号，与正文形成对比"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "问题背景",
            "content_type": "example_detail",
            "kicker": "实例 · 问题背景",
            "lede": "传统部署方式难以应对现代互联网业务的高并发、弹性伸缩需求",
            "context_block": "随着业务规模扩大，单机部署已无法满足需求。公司某核心系统曾因部署周期长达5-7天，错失了两次重要的业务推广窗口期。同时，生产环境故障恢复时间超过2小时，严重影响用户体验。运维成本已占IT预算的35%。",
            "solution_block": "通过引入Kubernetes，我们实现了：部署周期从5-7天缩短至小时级；故障自动恢复时间从2小时降至5分钟；运维效率提升60%，支撑了从10个服务到200个服务的规模增长。",
            "metrics": [
                {"value": "5-7天", "label": "传统部署周期", "trend": "↑ 严重拖累业务"},
                {"value": ">2小时", "label": "故障恢复时间", "trend": "↑ SLA难以保障"},
                {"value": "35%", "label": "运维成本占比", "trend": "↑ 资源浪费严重"}
            ],
            "takeaway": "启示：容器化和编排是现代云原生架构的必由之路",
            "notes": "用具体数字说明痛点，不要空泛描述",
            "filling_prompt": "必须填入真实内容（通过 web_search 获取权威数据，至少2个URL）：lede 一句话说明问题的严重性；context_block 描述行业普遍痛点（1-2句话）；solution_block 具体说明该痛点导致的后果或损失（2-3句话）；metrics_grid 提供3个量化指标（如'部署周期5-7天'、'故障恢复>2小时'、'运维成本占IT预算40%'），每个有 value（数字）、label（说明）、trend（趋势）；takeaway 用一句话总结启示。禁止空泛描述。references 列出 URL。禁止保留花括号。",
            "speaking_tip": "讲述时可以用亲身经历增加说服力"
        },
        {
            "index": 5,
            "type": "two_column",
            "title": "现有方案分析",
            "content_type": "two_column",
            "kicker": "方案对比 · 技术选型",
            "left_header": "传统方案局限性",
            "left_sections": {
                "analysis": [
                    "部署效率：手动部署依赖人工操作，环境差异导致'在我机器上能跑'问题频发，部署周期长达5-7天，错过业务推广窗口",
                    "资源利用：按峰值预留计算资源，平日利用率仅15-25%，造成大量资源浪费，云成本居高不下",
                    "可用性保障：故障恢复依赖人工介入，平均恢复时间超过2小时，SLA难以保障，影响用户体验和业务口碑"
                ],
                "data": [
                    "部署周期：5-7天/次",
                    "资源利用率：15-25%",
                    "故障恢复时间：>2小时"
                ]
            },
            "right_header": "改进方向",
            "right_sections": {
                "key_points": [
                    "声明式配置：用代码管理基础设施，环境一致性问题从源头解决，版本化管理支持回滚",
                    "自动化编排：减少人工干预，CI/CD流水线实现一键部署，部署时间从数天缩短至分钟级",
                    "弹性伸缩：根据负载自动调整资源，利用率提升至60%以上，峰值自动扩容、谷值自动缩容",
                    "服务治理：统一的流量管理、熔断限流、灰度发布，服务可用性从99.5%提升至99.9%"
                ],
                "data": [
                    "部署周期：<1小时",
                    "资源利用率：60%+",
                    "故障恢复时间：<5分钟"
                ]
            },
            "notes": "左右对比传统方案的不足与改进方向，每个维度用数据支撑",
            "filling_prompt": "必须填入真实内容：left_header 为'传统方案局限性'，left_sections.analysis 列出2-3个具体问题并说明其影响，left_sections.data 列出2-3个量化指标；right_header 为'改进方向'，right_sections.key_points 列出2-3条对应的改进方案并说明效果，right_sections.data 列出对应的改进后指标。注意左右数据形成对比。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "核心原理",
            "subtitle": "容器与编排的核心概念",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "transition_phrase": "接下来，让我们深入理解Kubernetes的核心概念。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "核心概念",
            "content_type": "content_slide",
            "concepts": [
                {"name": "Pod", "desc": "Kubernetes最小调度单元，一个Pod可包含一个或多个容器"},
                {"name": "Service", "desc": "服务的抽象，屏蔽Pod动态变化的IP"},
                {"name": "Namespace", "desc": "资源隔离的逻辑分组，便于多团队协作"},
                {"name": "Deployment", "desc": "声明式的Pod管理，支持滚动更新和回滚"}
            ],
            "notes": "用简洁的语言解释核心概念，配合示意图（文字描述即可）",
            "filling_prompt": "必须填入真实内容：用通俗语言解释3-4个核心概念，每条配合一句话说明。可用文字描述示意图内容（如'控制平面负责调度，所有节点上报状态'）。",
            "visual_suggestions": "配合简单的架构示意图效果更佳"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "调度流程",
            "content_type": "process_flow",
            "direction": "horizontal",
            "steps": [
                {"num": "1", "title": "提交请求", "desc": "用户通过kubectl提交Deployment"},
                {"num": "2", "title": "API Server接收", "desc": "请求经过认证授权入库etcd"},
                {"num": "3", "title": "调度决策", "desc": "Scheduler根据策略选择最优节点"},
                {"num": "4", "title": "分配执行", "desc": "Kubelet在目标节点创建容器"},
                {"num": "5", "title": "状态同步", "desc": "节点状态持续上报至控制平面"}
            ],
            "notes": "用流程图展示核心步骤，3-5步为宜",
            "filling_prompt": "必须填入真实内容：提供3-5个核心步骤，每步有名称和一句话描述，展示该技术的工作流程。",
            "technical_detail_level": "overview"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "架构设计",
            "subtitle": "系统整体架构与模块划分",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "content_slide",
            "title": "整体架构",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "architecture_description": "控制平面（Master）包含API Server、Scheduler、Controller Manager、etcd四个核心组件，负责整个集群的协调和管理。工作节点（Worker）运行Kubelet、kube-proxy和容器运行时（如Docker或containerd），负责实际的工作负载执行。用户通过kubectl与API Server交互，API Server是整个系统的唯一入口。",
            "notes": "用文字描述架构图（组件+关系），不要求真实图片",
            "filling_prompt": "必须填入真实内容：用文字描述系统整体架构（组件名称+组件之间的关系，如'API Server接收请求 → Scheduler分配节点 → Kubelet执行 → 状态同步至etcd'）。",
            "visual_placeholder": "架构图区域：展示Master节点和控制平面组件，以及Worker节点和工作负载的关系"
        },
        {
            "index": 11,
            "type": "content_slide",
            "title": "核心模块详解",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "etcd存储层", "body": "高可用的键值存储，保存集群所有状态数据"},
                {"header": "API Server", "body": "集群统一入口，处理所有REST请求"},
                {"header": "Scheduler", "body": "负责Pod调度，为新Pod选择最优节点"},
                {"header": "Controller Manager", "body": "运行各种控制器，维护期望状态"}
            ],
            "notes": "用卡片展示核心模块，每个模块一句话说明",
            "filling_prompt": "必须填入真实内容：提供4个核心模块，每个模块有 header（模块名称）和 body（一句话说明功能）。模块名称要具体，如'etcd存储层'、'API Server'、'Scheduler调度器'。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "04",
            "title": "实践案例",
            "subtitle": "真实项目中的应用与效果",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "image_text",
            "title": "应用案例：电商平台容器化改造",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "电商行业",
            "header": "某头部电商平台迁移实践",
            "sub_header": "日均订单量500万+的大促保障",
            "paragraph": "该电商平台原有架构采用传统的虚拟机部署方式，在双十一大促期间频繁出现扩容不及时、服务雪崩等问题。通过将核心交易系统迁移至Kubernetes平台，配合HPA自动扩缩容和熔断限流机制，成功应对了峰值流量的考验。实测数据显示，大促期间系统可承载的并发处理能力提升了8倍，资源利用率从25%提升至65%，每年节省云资源成本约800万元。",
            "references": [
                "https://www.kubernetes.org.cn/",
                "https://aws.amazon.com/cn/containers/"
            ],
            "notes": "用图文混排展示具体应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 填具体行业领域（如'金融'、'电商'、'医疗'）；title 中的 {客户/项目名称} 替换为真实客户或项目名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX公司智能客服系统'，不超过35字）；sub_header 为合作项目名称（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述案例背景、技术方案、实施过程和应用效果，用流畅的段落形式呈现，禁止罗列要点。references 列出 web_search 获取的 URL（至少2个）。禁止使用'某公司''某系统'等匿名实体；禁止虚构数据。",
            "key_metrics": ["并发处理能力提升8倍", "资源利用率65%", "年节省成本800万"]
        },
        {
            "index": 14,
            "type": "kpi_dashboard",
            "title": "关键数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 核心指标",
            "kpis": [
                {"value": "↑ 800%", "label": "并发处理能力提升", "delta": "↑ 8倍", "baseline": "vs 迁移前"},
                {"value": "↑ 160%", "label": "资源利用率提升", "delta": "↑ 2.6倍", "baseline": "vs 迁移前"},
                {"value": "↓ 800万/年", "label": "云资源成本节省", "delta": "↓ 40%", "baseline": "vs 迁移前"},
                {"value": "< 30秒", "label": "故障自愈时间", "delta": "↓ 95%", "baseline": "vs 迁移前"}
            ],
            "notes": "4个核心指标，delta 为变化比例，baseline 为对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心数据指标，每个有 value（具体数字）、label（效果说明）、delta（变化趋势，如'↑ 30%'或'↓ 50%'）、baseline（对比基准，如'vs 传统方案'）。指标要具体且有代表性。references 列出 web_search 获取的 URL（至少2个）。禁止虚构数据。"
        },
        {
            "index": 15,
            "type": "section_divider",
            "number": "05",
            "title": "总结与展望",
            "subtitle": "核心要点与未来方向",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。",
            "transition_phrase": "最后，让我们总结一下本次分享的核心要点。"
        },
        {
            "index": 16,
            "type": "example_detail",
            "title": "核心要点",
            "content_type": "example_detail",
            "kicker": "实例 · 核心要点",
            "lede": "Kubernetes是云原生时代的基础设施标准，容器编排是现代DevOps的核心能力",
            "context_block": "本次分享涵盖了Kubernetes的核心概念、架构设计和实践案例。通过理论结合实践的方式，帮助大家建立对容器编排技术的系统性认知。",
            "solution_block": "核心要点包括：1）Pod是最小调度单元，理解Pod的生命周期至关重要；2）声明式API是Kubernetes的核心设计理念；3）控制平面负责协调，工作节点负责执行；4）自动扩缩容和自愈能力是Kubernetes的核心优势。",
            "metrics": [
                {"value": "掌握核心概念", "label": "知识点覆盖", "trend": "vs 分享前"},
                {"value": "理解架构设计", "label": "架构认知", "trend": "vs 分享前"},
                {"value": "具备实践能力", "label": "动手能力", "trend": "vs 分享前"}
            ],
            "takeaway": "启示：容器化和编排能力将成为每个工程师的必备技能",
            "notes": "3-4条核心要点，用加粗序号",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的信息；context_block 回顾分享的核心内容（1-2句话）；solution_block 总结核心要点和关键结论（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结如何将分享应用到实际工作中。禁止保留花括号。"
        },
        {
            "index": 17,
            "type": "example_detail",
            "title": "未来方向",
            "content_type": "example_detail",
            "kicker": "实例 · 未来方向",
            "lede": "Kubernetes生态持续演进，Serverless和GitOps是重要发展方向",
            "context_block": "当前Kubernetes已在生产环境广泛采用，但技术仍在快速迭代。Operator模式、GitOps、Serveless等新范式不断涌现，Kubernetes正在变得更加智能和自动化。",
            "solution_block": "未来演进方向包括：1）Kubernetes Operators的普及让自定义资源管理更加便捷；2）GitOps将成为主流的部署运维模式；3）Serverless容器（如AWS Fargate）让用户无需管理节点；4）AIops结合让集群运维更加智能化。",
            "metrics": [
                {"value": "持续迭代", "label": "版本更新频率", "trend": "每季度大版本"},
                {"value": "生态繁荣", "label": "周边工具增长", "trend": "↑ 30%/年"},
                {"value": "广泛采用", "label": "企业采纳率", "trend": "↑ 覆盖80%云原生项目"}
            ],
            "takeaway": "启示：保持学习，持续跟进云原生技术演进",
            "notes": "技术演进方向或后续规划",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最重要演进趋势；context_block 说明当前现状和局限（1-2句话）；solution_block 详细描述未来演进方向和突破点（2-3句话）；metrics_grid 提供3个趋势指标；takeaway 用一句话总结如何把握未来趋势。禁止保留花括号。"
        },
        {
            "index": 18,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 Kubernetes是云原生时代的基础设施标准",
                "02 声明式API和自动化是核心设计理念",
                "03 通过实践案例验证了容器化的价值"
            ],
            "thank_you": "感谢聆听",
            "contact": "联系方式：liming@company.com | 技术交流群：xxx",
            "notes": "结尾页，核心回顾 + 感谢",
            "filling_prompt": "必须填入真实内容：key_points 提供3个核心回顾要点；contact 填写真实联系方式。禁止保留花括号。",
            "q_and_a_hint": "预留10-15分钟Q&A，欢迎提问",
            "materials_note": "演讲材料和代码示例将在会后通过邮件发送"
        }
    ],
    "design_tips": [
        "技术分享要注重内容深度，不要堆砌文字",
        "每章节用 section_divider 清晰划分",
        "用数字说明效果比文字描述更有说服力",
        "架构图用文字描述组件关系即可，不需要真实图片",
        "代码示例要精选，突出重点，避免大段代码",
        "预留Q&A时间，技术问题当场解答效果更好",
        "提供延伸学习资料，满足不同深度的学习需求",
        "结合自身实践经验，增加分享的真实性和说服力"
    ],
    "presentation_flow": {
        "opening": {
            "duration": "3-5分钟",
            "goal": "建立问题意识，引起工程师共鸣",
            "tip": "用一个实际的技术挑战或痛点开场"
        },
        "body": {
            "duration": "25-35分钟",
            "goal": "深入讲解原理和实践",
            "tip": "每个章节控制在5-8分钟，穿插互动问题"
        },
        "closing": {
            "duration": "5-10分钟",
            "goal": "总结要点，展望未来",
            "tip": "预留足够Q&A时间，解答技术细节"
        }
    },
    "code_examples_guidance": {
        "best_practices": [
            "代码示例要简短，突出核心逻辑",
            "添加注释解释关键步骤",
            "展示实际运行结果",
            "对比优化前后的差异"
        ],
        "common_pitfalls": [
            "避免一次性展示大量代码",
            "不要展示敏感的配置信息",
            "确保代码在演示环境下可以运行"
        ]
    }
}
