TEMPLATE = {
    "name": "product-launch",
    "name_cn": "产品发布",
    "description": "适合新产品发布会、产品宣讲、客户演示等场景。强调价值主张、核心功能、差异化优势。",
    "target_audience": "客户、投资者、合作伙伴、媒体",
    "typical_slides": 14,
    "typical_duration": "15-20分钟",
    "palette": "warm_terracotta",
    "typography": {
        "header": "Arial Black",
        "body": "Arial",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "launch_strategy": "产品发布的核心是制造'哇'时刻",
        "key_moments": [
            "开场：用一个问题或痛点场景引发共鸣",
            "揭晓：产品亮相，配合视觉冲击",
            "演示：核心功能现场演示",
            "案例：真实用户证言",
            "CTA：明确的购买/试用号召"
        ],
        "messaging_hierarchy": [
            "一句话价值主张（最重要）",
            "三个核心卖点",
            "五个支撑点"
        ]
    },
    "launch_best_practices": {
        "pre_launch": [
            "预热传播，制造悬念",
            "邀请KOL和媒体",
            "准备产品演示和视频"
        ],
        "during_launch": [
            "控制节奏，高潮迭起",
            "现场演示增加可信度",
            "实时互动，收集反馈"
        ],
        "post_launch": [
            "发布新闻稿和评测",
            "跟进销售线索",
            "收集用户反馈"
        ]
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "云智AI平台3.0",
            "subtitle": "重新定义企业智能化 — 让每一家企业都能拥有AI能力",
            "author": "北京云智科技有限公司",
            "date": "2025年5月",
            "notes": "开场标题页要有冲击力，Slogan要简短有力，背景使用深蓝色渐变配合科技感线条动画",
            "filling_prompt": "必须填入真实内容：title 为实际产品名称，subtitle 为一句有冲击力的Slogan，author 为公司名称，date 为日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "发布会议程",
            "kicker": "目录",
            "items": [
                "01  市场痛点 — 企业智能化转型的困境",
                "02  解决方案 — 云智AI平台3.0概览",
                "03  核心功能 — 六大AI能力深度解析",
                "04  产品优势 — 为什么选择我们",
                "05  客户案例 — 真实客户的成功实践"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "目录页为固定结构，包含5个主要章节：市场痛点、解决方案、核心功能、产品优势、客户案例。无需额外填充。"
        },
        {
            "index": 3,
            "type": "stat_slide",
            "title": "市场数据",
            "content_type": "stat_slide",
            "kicker": "市场洞察",
            "stats": [
                {"value": "73%", "label": "的企业在AI落地过程中遇到重重困难", "trend": "上升"},
                {"value": "6-12个月", "label": "企业自建AI系统的平均交付周期", "trend": "居高不下"},
                {"value": "58%", "label": "的AI项目因缺乏专业人才而失败或搁置", "trend": "令人担忧"}
            ],
            "notes": "用3个大数字震撼开场，揭示企业智能化转型的严峻现实，激发观众共鸣",
            "filling_prompt": "必须先通过 web_search 获取权威市场研究报告（至少2个URL），再填入真实内容：提供3个市场数据指标，每个有 value（具体数值）、label（一句话说明）和 trend（变化方向）。数据要来自权威报告。禁止虚构数据。references 列出 URL。"
        },
        {
            "index": 4,
            "type": "section_divider",
            "number": "01",
            "title": "市场痛点",
            "subtitle": "企业智能化转型面临的真实困境",
            "notes": "进入第一部分，切换为深色背景章节页，配合紧张氛围的视觉元素",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 5,
            "type": "example_detail",
            "title": "企业AI落地的四大障碍",
            "content_type": "example_detail",
            "kicker": "实例 · 痛点分析",
            "lede": "技术门槛高、专业人才缺、部署周期长、运维成本高 — 这四座大山挡住了绝大多数企业的AI之路。",
            "context_block": "根据中国信通院2024年调研数据，超过70%的企业在推进AI应用时面临重重障碍。其中，62%的企业缺乏既懂AI又懂业务的复合型人才；55%的企业反映AI项目交付周期远超预期；48%的企业被高昂的运维成本所困扰。这些问题在中小企业中尤为突出。",
            "solution_block": "云智AI平台3.0正是为解决这些痛点而生。我们通过「零代码建模」「智能运维」「弹性伸缩」「一站式部署」四大核心能力，将AI落地的技术门槛降低80%，交付周期缩短60%，运维成本降低70%。即使没有AI专业背景的业务人员，也能通过拖拽式操作快速构建和部署AI模型。",
            "metrics": [
                {"value": "62%", "label": "企业缺乏复合型人才", "trend": "人才缺口持续扩大"},
                {"value": "6-12个月", "label": "平均AI项目交付周期", "trend": "远超企业预期"},
                {"value": "48%", "label": "企业被运维成本困扰", "trend": "成本居高不下"}
            ],
            "takeaway": "云智AI平台让AI不再是大企业的专利，中小企业同样可以拥有强大的AI能力。",
            "notes": "深入剖析企业在AI落地过程中面临的核心障碍，用数据说话",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：lede 一句话概括核心挑战；context_block 描述市场当前问题和困境（1-2句话）；solution_block 具体说明问题导致的后果和损失（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结紧迫性。禁止空泛描述。references 列出 URL。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "解决方案",
            "subtitle": "云智AI平台3.0 — 一站式解决企业AI难题",
            "notes": "进入第二部分，切换为明亮的渐变背景章节页，传递希望和信心",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "three_column",
            "title": "平台概览",
            "content_type": "three_column",
            "kicker": "解决方案",
            "columns": [
                {
                    "icon": "tech_stack",
                    "header": "技术架构",
                    "body": "基于云原生架构，支持公有云、私有云、混合云多种部署方式。集成AutoML自动建模、模型市场、MLOps全生命周期管理，单平台覆盖从数据处理到模型上线的完整流程。"
                },
                {
                    "icon": "cloud_deploy",
                    "header": "部署方式",
                    "body": "5分钟即可完成平台部署，开箱即用。支持一键导入数据集、自动训练模型、API快速接入业务系统。提供100+预置行业模型，企业无需从零开始，投入生产时间从数月缩短至数天。"
                },
                {
                    "icon": "value_deliver",
                    "header": "核心价值",
                    "body": "降低80%技术门槛，非AI专业人员也能快速上手。交付周期缩短60%，从数月到数周。运维成本降低70%，按需弹性计费。7x24小时专业运维支持，让企业专注核心业务创新。"
                }
            ],
            "notes": "三栏并排布局，清晰呈现解决方案的三个维度：技术架构、部署方式、核心价值",
            "filling_prompt": "必须填入真实内容：columns 提供3个维度的说明，每个有 icon（图标名称）、header（标题，不超过12字）和 body（150-200字的自然语言段落，详细展开该维度的内容）。禁止空洞描述。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "核心功能",
            "subtitle": "六大AI能力，重塑企业智能化体验",
            "notes": "进入第三部分，切换为深色科技感背景章节页",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "card_grid",
            "title": "核心功能",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {
                    "header": "零代码建模",
                    "body": "通过可视化拖拽界面，无需编写一行代码即可完成数据预处理、特征工程、模型训练和评估的全流程。系统自动调参AutoML，平均模型准确率提升15%以上。"
                },
                {
                    "header": "智能运维",
                    "body": "MLOps全生命周期管理，实时监控模型性能，自动预警模型漂移和效果衰减。内置A/B测试框架，支持模型的灰度发布和快速回滚，确保模型始终处于最佳状态。"
                },
                {
                    "header": "多模态分析",
                    "body": "支持文本、图像、语音、视频等多种数据类型的一体化处理。预置自然语言处理、计算机视觉、语音识别等领域的50+成熟模型，开箱即用，效果行业领先。"
                },
                {
                    "header": "企业知识库",
                    "body": "一键对接企业知识库和文档系统，通过RAG检索增强生成技术，让AI能够精准理解和回答与企业业务相关的专业问题。文档解析准确率达92%，问答响应时间低于2秒。"
                }
            ],
            "notes": "2x2卡片网格展示平台四大核心功能，每个卡片配有图标、标题和详细描述",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称，不超过8字）和 body（详细描述功能特性和客户价值，100-120字）。功能描述要具体、可信，禁止夸大宣传。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "产品优势",
            "subtitle": "为什么3000+企业选择云智AI",
            "notes": "进入第四部分，切换为高对比度背景章节页，传递信心和实力",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 产品价值",
            "kpis": [
                {"value": "3.2x", "label": "业务处理效率提升", "delta": "较传统方案+220%", "baseline": "vs 企业自建AI系统"},
                {"value": "68%", "label": "人力成本降低", "delta": "年节省IT支出约40万元", "baseline": "vs 行业平均水平"},
                {"value": "15分钟", "label": "模型上线时间", "delta": "从数月缩短至分钟级", "baseline": "vs 行业平均6-12个月"},
                {"value": "99.95%", "label": "平台可用性", "delta": "SLA保障", "baseline": "支持7x24小时运维"}
            ],
            "notes": "2x2仪表盘布局展示4个核心效果指标，每个配有具体数值、变化趋势和对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心效果指标，每个 KPI 有 value（具体数值）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。指标要具体且可验证。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "客户案例",
            "subtitle": "3000+企业的共同选择",
            "notes": "进入第五部分，切换为温暖色系背景章节页，传递信任和成功",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "image_text",
            "title": "招商银行智能风控项目",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "金融行业",
            "header": "招商银行智能风控平台",
            "sub_header": "全行级AI风控系统升级改造",
            "paragraph": "招商银行携手云智科技，对其全行级风控系统进行了全面智能化升级。基于云智AI平台3.0，招商银行构建了覆盖贷前、贷中、贷后全流程的智能风控体系。新系统上线后，贷款审批效率提升4倍，坏账率下降32%，风控模型的KS值从0.31提升至0.42，风险区分能力显著增强。同时，系统支持毫秒级实时决策，日均处理交易超过5000万笔，峰值并发能力达到10万TPS，为银行业务的稳健发展提供了强有力的技术支撑。",
            "references": [
                "https://www.cmbchina.com/about/tech/",
                "https://finance.sina.com.cn/tech/"
            ],
            "notes": "用图文混排展示招商银行智能风控项目案例，右侧为项目成果展示图占位",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为合作项目名称（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述客户背景、合作过程、实施效果和客户评价，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "立即体验云智AI平台3.0",
            "key_points": [
                "01  免费试用：30天全功能免费体验，零成本验证产品价值",
                "02  专业支持：专属方案顾问一对一服务，30分钟快速响应",
                "03  成功保障：达不到承诺效果，全额退款 — 让您零风险决策"
            ],
            "thank_you": "感谢关注，让AI成为每一家企业的核心竞争力！",
            "contact": "官网：www.yunzhicloud.com  |  热线：400-888-9999  |  邮箱：contact@yunzhicloud.com",
            "notes": "结尾页，明确三重行动号召和联系方式，以品牌主色调收尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个行动号召；contact 填写真实联系方式（可使用示例数据但格式要真实）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "产品发布要有冲击力，视觉上突出产品和价值",
        "痛点→方案→功能→优势→案例，逻辑链条清晰",
        "数据说话：转化率、效率提升、成本降低等",
        "结尾要明确行动号召（CTA）",
        "准备产品演示视频或现场演示",
        "邀请客户代表分享成功案例",
        "设置互动环节，增加参与感",
        "准备FAQ，应对媒体和客户提问"
    ],
    "launch_checklist": {
        "pre_event": [
            "确认场地、媒体邀请、嘉宾名单",
            "准备产品演示环境和备用方案",
            "制作产品视频和宣传物料",
            "培训销售团队应对咨询"
        ],
        "event_day": [
            "提前彩排演示流程",
            "安排专人接待媒体和VIP",
            "实时收集现场反馈",
            "做好图文直播准备"
        ],
        "post_event": [
            "发布新闻稿和评测文章",
            "跟进销售线索",
            "整理用户反馈",
            "复盘发布会效果"
        ]
    }
}
