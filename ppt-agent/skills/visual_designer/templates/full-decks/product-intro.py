TEMPLATE = {
    "name": "product-intro",
    "name_cn": "产品介绍",
    "description": "适合产品介绍、客户演示、功能展示等场景。突出价值，展示功能，增强信任。",
    "target_audience": "客户、合作伙伴、销售团队",
    "typical_slides": 18,
    "typical_duration": "15-20分钟",
    "palette": "warm_terracotta",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "opening_hook": "用一个客户痛点或成功故事开场，建立情感连接",
        "key_moments": [
            "产品价值：强调能为客户解决什么问题",
            "核心功能：用具体场景展示功能价值",
            "客户案例：用真实故事增加可信度",
            "竞争优势：差异化要清晰明确"
        ],
        "closing_strength": "明确的行动号召（预约演示/获取试用）"
    },
    "value_proposition_tips": {
        "focus_on_outcomes": "聚焦客户收益，而非产品功能",
        "use_stories": "用故事而非数据来建立情感连接",
        "quantify_impact": "尽可能量化业务影响",
        "address_objections": "预见并回应潜在疑虑"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "DataSync Pro",
            "subtitle": "企业级实时数据同步平台 — 让数据流动更高效、更可靠",
            "author": "周海洋 | 售前顾问",
            "date": "2025年7月",
            "notes": "标题页醒目，展示产品名称和核心价值主张，配合品牌视觉",
            "filling_prompt": "必须填入真实内容：title 为产品名称，subtitle 为产品定位或一句话价值主张，author 为主讲人姓名，date 为演示日期。禁止保留花括号占位符。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  产品概述",
                "02  核心功能",
                "03  客户案例",
                "04  竞争优势"
            ],
            "notes": "让听众了解产品介绍结构，每项一行即可，不要展开内容",
            "filling_prompt": "目录页为固定结构，无需额外填充。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "产品概述",
            "subtitle": "产品定位与价值",
            "notes": "章节分隔页，仪式感强，用于划分不同内容板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "产品概览",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "产品定位",
            "header": "一键打通企业数据孤岛",
            "sub_header": "企业级实时数据同步平台",
            "paragraph": "DataSync Pro是一款面向中大型企业的数据同步与整合平台，支持跨数据库、跨云、跨地域的实时数据流转。产品采用无侵入式设计，无需改造现有系统即可实现分钟级数据同步，已服务于超过200家企业客户，累计处理数据量超过日均500亿条。我们的客户涵盖金融、制造、零售、医疗等多个行业，帮助他们在保障数据安全的前提下，打通信息孤岛，释放数据价值。",
            "notes": "用图文混排展示产品主视觉和核心价值主张",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为一句话核心价值主张（不超过35字）；sub_header 为产品定位说明；paragraph 为300-450字的自然语言段落，详细描述产品的核心价值、功能特点和适用场景，用流畅的段落形式呈现，禁止罗列要点。"
        },
        {
            "index": 5,
            "type": "stat_slide",
            "title": "产品核心数据",
            "content_type": "stat_slide",
            "kicker": "数据 · 产品表现",
            "stats": [
                {"value": "200+", "label": "企业客户", "desc": "覆盖金融、制造、零售、医疗等10余个行业"},
                {"value": "500亿条/日", "label": "日均处理量", "desc": "稳定运行，峰值处理能力达每秒百万级"},
                {"value": "99.99%", "label": "数据准确率", "desc": "端到端数据校验，保障数据一致性"},
                {"value": "分钟级", "label": "同步延迟", "desc": "支持秒级实时同步，满足业务时效要求"}
            ],
            "notes": "用4个核心数字展示产品的市场表现和技术实力",
            "filling_prompt": "必须填入真实内容：提供4个产品相关指标数据，每个有 value（具体数字）、label（说明）、desc（详细描述）。数字要真实可信，禁止虚构。可通过 web_search 获取参考数据。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "核心功能",
            "subtitle": "主要功能模块",
            "notes": "章节分隔页，用于划分核心功能板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "四大核心功能",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {
                    "header": "多源异构数据同步",
                    "body": "支持Oracle、MySQL、PostgreSQL、SQL Server、MongoDB等20余种数据源，支持结构化、半结构化和非结构化数据。提供可视化配置界面，5分钟即可完成一条同步链路配置，无需编写代码。"
                },
                {
                    "header": "实时监控与告警",
                    "body": "提供端到端的数据同步监控大屏，实时展示同步吞吐量、数据延迟、错误率等核心指标。支持多级别告警（短信、邮件、企业微信），告警响应时间小于30秒，确保问题早发现早处理。"
                },
                {
                    "header": "数据质量治理",
                    "body": "内置数据清洗、转换、脱敏、校验五大引擎，支持JSON/XML解析、字段映射、数据类型转换等200余种转换函数。数据质量报告自动生成，帮助客户持续优化数据资产。"
                },
                {
                    "header": "安全与合规保障",
                    "body": "数据传输全程TLS 1.3加密，数据落盘AES-256加密。支持行级权限控制、审计日志、数据脱敏，满足等保三级和GDPR合规要求。已通过ISO 27001和SOC 2 Type II认证。"
                }
            ],
            "notes": "展示4个核心功能模块，每个功能配具体描述",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称，不超过35字）和 body（详细描述该功能的核心价值和应用场景，100-120字，包含具体使用效果或数据）。禁止保留花括号占位符。"
        },
        {
            "index": 8,
            "type": "process_flow",
            "title": "同步链路配置流程",
            "content_type": "process_flow",
            "kicker": "流程 · 操作步骤",
            "direction": "horizontal",
            "steps": [
                {"num": "1", "title": "连接数据源", "desc": "选择源端数据库，填写连接信息，测试连通性"},
                {"num": "2", "title": "配置同步规则", "desc": "选择同步对象，配置映射关系和转换规则"},
                {"num": "3", "title": "设置监控告警", "desc": "配置监控指标和告警规则，绑定通知渠道"},
                {"num": "4", "title": "启动并验证", "desc": "启动同步任务，验证数据一致性和完整性"}
            ],
            "notes": "用流程图展示用户配置同步链路的4个核心步骤",
            "filling_prompt": "必须填入真实内容：提供4个配置步骤，每步有名称和一句话描述，展示用户从配置到上线的完整操作流程。"
        },
        {
            "index": 9,
            "type": "two_column",
            "title": "传统方案 vs DataSync Pro",
            "content_type": "two_column",
            "kicker": "方案对比 · 核心差异",
            "left_header": "传统方案痛点",
            "left_sections": {
                "analysis": [
                    "ETL批处理延迟高：T+1甚至T+2的数据时效，无法满足实时业务需求",
                    "开发运维成本高：需要专职DBA和开发团队，维护复杂，版本迭代慢",
                    "数据质量难保障：缺乏端到端校验，数据漂移难以发现，影响业务决策"
                ],
                "data": [
                    "平均上线周期：3-6个月",
                    "年运维成本：50-100万元",
                    "数据延迟：小时级到天级"
                ]
            },
            "right_header": "DataSync Pro优势",
            "right_sections": {
                "key_points": [
                    "实时同步：支持秒级数据同步，延迟从小时级降至分钟级甚至秒级",
                    "零代码配置：可视化界面，非技术人员也能独立完成配置和维护",
                    "质量内置：端到端数据校验，异常自动告警，数据质量实时可见"
                ],
                "data": [
                    "平均上线周期：1-2周",
                    "年运维成本：10-20万元",
                    "数据延迟：分钟级到秒级"
                ]
            },
            "notes": "左右对比展示传统方案的不足与DataSync Pro的优势，每个维度用数据支撑",
            "filling_prompt": "必须填入真实内容：left_header 为传统方案痛点，left_sections.analysis 列出2-3个具体问题并说明其影响，left_sections.data 列出对应的量化数据；right_header 为DataSync Pro优势，right_sections.key_points 列出2-3条对应的优势并说明效果，right_sections.data 列出对应的改进后指标。注意左右数据形成对比关系。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "03",
            "title": "客户案例",
            "subtitle": "成功案例展示",
            "notes": "章节分隔页，用于划分客户案例板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "case_study",
            "title": "客户案例：华通银行",
            "content_type": "case_study",
            "kicker": "金融行业 · 银行",
            "client": "华通商业银行",
            "industry": "金融·商业银行",
            "challenge": "华通银行原有数据仓库采用T+1批处理模式，核心系统与风控系统之间数据延迟高达4-6小时，无法满足监管报送的实时性要求和风控模型的时效性需求。",
            "solution": "DataSync Pro为其搭建了覆盖核心系统、信贷系统、风控系统、反欺诈系统等8个系统的实时数据同步网络，将数据延迟从平均4.2小时降至3分钟以内，满足了监管报送的实时性要求，同时为实时风控模型提供了数据基础。",
            "results": [
                "数据同步延迟：从4.2小时降至3分钟",
                "监管报送时效：满足人行实时报送要求",
                "风控模型效果：欺诈交易识别率提升67%",
                "运维人力节省：减少2名DBA专职投入"
            ],
            "testimonial": "DataSync Pro帮助我们真正实现了数据的实时流动，不仅满足了监管要求，更为我们的智能风控体系奠定了数据基础。",
            "testimonial_author": "张建华 · 华通银行信息科技部总经理",
            "references": [
                "https://www.example-bank.com/news",
                "https://industry-report.example.org/finance"
            ],
            "notes": "用案例详解模板展示华通银行从痛点到解决到成效的完整故事",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：client 为客户名称，industry 为行业分类，challenge 描述客户的业务痛点（2-3句话），solution 描述DataSync Pro的具体解决方案和实施过程（2-3句话），results 列出3-4个量化成果，testimonial 为客户的一句评价，testimonial_author 为评价人和职位。禁止虚构客户名称和数据。"
        },
        {
            "index": 12,
            "type": "kpi_dashboard",
            "title": "客户实施成效",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 客户成效",
            "kpis": [
                {"value": "平均62%", "label": "运营效率提升", "delta": "较实施前", "baseline": "vs 实施前"},
                {"value": "平均85%", "label": "数据质量改善", "delta": "端到端校验通过率", "baseline": "vs 实施前"},
                {"value": "平均45%", "label": "IT成本节省", "delta": "运维和人力成本", "baseline": "vs 实施前"},
                {"value": "99.99%", "label": "服务可用性", "delta": "年度 SLA 保证", "baseline": "行业平均 99.5%"}
            ],
            "notes": "展示客户实施后的综合成效数据",
            "filling_prompt": "必须填入真实内容：提供4个客户成效指标，每个有 value、label、delta、baseline。数据来源于实际客户案例和行业平均水平。"
        },
        {
            "index": 13,
            "type": "quote_slide",
            "title": "客户心声",
            "content_type": "quote_slide",
            "kicker": "金句",
            "quote": "数据同步的延迟从小时级到分钟级，这不是简单的技术指标提升，而是意味着我们的业务决策可以真正做到实时响应。DataSync Pro是我们数字化转型路上的关键基础设施。",
            "attribution": "李明哲 · 华通银行数据总监",
            "context": "在华通银行DataSync Pro项目验收会上，数据总监李明哲对产品给予高度评价",
            "notes": "摘录客户的关键引言，增强产品介绍的说服力和可信度",
            "filling_prompt": "必须填入真实内容：quote 为客户的一句评价引言（1-2句话），attribution 为客户姓名和职位，context 为引言的背景说明。如无现成引言，可从已有案例中提炼，禁止虚构。"
        },
        {
            "index": 14,
            "type": "section_divider",
            "number": "04",
            "title": "竞争优势",
            "subtitle": "为什么选择我们",
            "notes": "章节分隔页，用于划分竞争优势板块",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 15,
            "type": "three_column",
            "title": "三大核心优势",
            "content_type": "three_column",
            "kicker": "优势 · 差异化竞争力",
            "columns": [
                {
                    "header": "技术领先",
                    "icon_suggestion": "技术/架构图标",
                    "points": [
                        "自主研发的增量解析引擎，性能较开源方案提升10倍",
                        "支持水平扩展，单集群可处理百万级TPS",
                        "核心代码自主可控，拥有38项技术专利"
                    ]
                },
                {
                    "header": "服务专业",
                    "icon_suggestion": "服务/支持图标",
                    "points": [
                        "7×24小时专属技术支持，15分钟内响应",
                        "提供从咨询到交付到运维的全流程服务",
                        "成功案例丰富，金融客户数量行业第一"
                    ]
                },
                {
                    "header": "安全合规",
                    "icon_suggestion": "安全/合规图标",
                    "points": [
                        "通过等保三级、ISO 27001、SOC 2认证",
                        "数据全程加密，满足GDPR和国内数据安全法要求",
                        "支持私有化部署，满足金融级安全要求"
                    ]
                }
            ],
            "notes": "用三栏布局展示产品的三大核心差异化竞争优势",
            "filling_prompt": "必须填入真实内容：提供3列内容，每列有 header（优势名称）、icon_suggestion（图标的文字描述）和 points（2-3个要点，每点一句话）。禁止保留花括号占位符。"
        },
        {
            "index": 16,
            "type": "content_slide",
            "title": "竞品对比分析",
            "content_type": "content_slide",
            "kicker": "对比 · 市场定位",
            "comparison": {
                "headers": ["对比维度", "DataSync Pro", "竞品A（开源）", "竞品B（传统）"],
                "rows": [
                    ["同步延迟", "秒级/分钟级", "小时级", "小时级/天级"],
                    ["配置方式", "可视化零代码", "需写代码配置", "需开发对接"],
                    ["监控告警", "内置全链路", "需自行开发", "基础监控"],
                    ["数据质量", "端到端校验", "无内置", "无内置"],
                    ["技术支持", "7×24专属", "社区支持", "工作日支持"],
                    ["行业认证", "等保三级/SOC2", "无", "部分认证"]
                ]
            },
            "notes": "用表格形式横向对比DataSync Pro与主要竞品的差异",
            "filling_prompt": "必须填入真实内容：comparison 包含 headers（4列：维度、DataSync Pro、竞品A、竞品B）和 rows（5-6行对比数据）。数据要客观真实，竞品名称可用'竞品A''竞品B'代替。"
        },
        {
            "index": 17,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心价值：实时打通企业数据孤岛，从小时级到分钟级",
                "02 关键优势：技术领先、专业服务、安全合规",
                "03 下一步行动：预约产品演示，15分钟了解DataSync Pro"
            ],
            "thank_you": "感谢聆听",
            "contact": "官网：www.datasyncpro.com | 热线：400-888-6789 | 扫码预约演示",
            "notes": "简洁有力的结尾，包含核心要点回顾和明确的行动号召",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心价值+关键优势+行动号召）；contact 填写真实的联系方式和官网信息。禁止保留花括号占位符。"
        },
        {
            "index": 18,
            "type": "content_slide",
            "title": "立即体验DataSync Pro",
            "content_type": "content_slide",
            "kicker": "行动号召 · 下一步",
            "action_items": [
                {
                    "action": "预约产品演示",
                    "desc": "15分钟在线演示，了解产品核心功能和实际效果",
                    "cta": "扫码预约",
                    "contact": "微信：datasync_sales"
                },
                {
                    "action": "申请免费试用",
                    "desc": "30天全功能免费试用，无需绑定信用卡",
                    "cta": "前往官网申请",
                    "contact": "www.datasyncpro.com/trial"
                },
                {
                    "action": "获取行业方案",
                    "desc": "下载金融、制造、零售等行业的专属解决方案",
                    "cta": "下载资料包",
                    "contact": "www.datasyncpro.com/resources"
                }
            ],
            "notes": "结尾页提供明确的行动选项，让观众知道下一步可以做什么",
            "filling_prompt": "必须填入真实内容：提供3个行动项，每个有 action（行动名称）、desc（行动描述）、cta（按钮文案）、contact（联系方式）。禁止保留花括号占位符。"
        }
    ],
    "design_tips": [
        "产品介绍要突出价值而非功能",
        "案例要真实可信，注明来源",
        "图文混排增强说服力",
        "数据说话，展示市场认可",
        "结尾要有明确的行动号召",
        "用客户语言而非技术语言",
        "强调ROI和业务价值",
        "准备好应对常见异议"
    ],
    "presentation_flow": {
        "opening": {
            "duration": "1-2分钟",
            "goal": "建立连接，引起兴趣",
            "tip": "用一个客户痛点或成功故事开场"
        },
        "body": {
            "duration": "12-15分钟",
            "goal": "展示价值，建立信任",
            "tip": "产品概述→功能详解→案例验证→优势总结"
        },
        "closing": {
            "duration": "2-3分钟",
            "goal": "促成行动",
            "tip": "明确CTA，提供联系方式"
        }
    },
    "objection_handling": {
        "common_objections": [
            "异议1 → 回应方式",
            "异议2 → 回应方式",
            "异议3 → 回应方式",
            "异议4 → 回应方式"
        ],
        "proof_points": [
            "第三方评测报告",
            "客户推荐信",
            "POC测试数据",
            "行业白皮书"
        ]
    }
}
