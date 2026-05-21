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
            "title": "产品名称",
            "subtitle": "一句有冲击力的Slogan",
            "author": "公司名称",
            "date": "日期",
            "notes": "开场标题页要有冲击力，Slogan要简短有力",
            "filling_prompt": "必须填入真实内容：title 为实际产品名称，subtitle 为一句有冲击力的Slogan，author 为公司名称，date 为日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  市场痛点",
                "02  解决方案",
                "03  核心功能",
                "04  产品优势",
                "05  客户案例"
            ],
            "notes": "让观众建立心理地图，每项一行即可，不要展开内容",
            "filling_prompt": "目录页为固定结构，无需额外填充。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "市场痛点",
            "subtitle": "当前面临的挑战",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "痛点分析",
            "content_type": "example_detail",
            "kicker": "实例 · 痛点分析",
            "lede": "一句话概括核心挑战",
            "context_block": "描述市场当前问题和困境（1-2句话）。",
            "solution_block": "具体说明问题导致的后果和损失（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "指标1", "trend": "变化趋势"},
                {"value": "数字", "label": "指标2", "trend": "变化趋势"},
                {"value": "数字", "label": "指标3", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结紧迫性。",
            "notes": "用数据说明当前市场的痛点和问题，唤起共鸣",
            "filling_prompt": "必须填入真实内容（通过 web_search 获取权威数据，至少2个URL）：lede 一句话概括核心挑战；context_block 描述市场当前问题和困境（1-2句话）；solution_block 具体说明问题导致的后果和损失（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结紧迫性。禁止空泛描述。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "解决方案",
            "subtitle": "我们如何解决",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 6,
            "type": "content_slide",
            "title": "解决方案",
            "content_type": "content_slide",
            "value_proposition": "一句话概括解决方案的核心价值主张",
            "key_points": [
                "支撑要点1",
                "支撑要点2",
                "支撑要点3",
                "支撑要点4"
            ],
            "notes": "一句话说清楚我们的解决方案是什么",
            "filling_prompt": "必须填入真实内容：一句话概括解决方案的核心价值主张，并配合2-3条支撑要点。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "核心功能",
            "subtitle": "产品能做什么",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "核心功能概览",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "功能名称", "body": "功能描述和价值"},
                {"header": "功能名称", "body": "功能描述和价值"},
                {"header": "功能名称", "body": "功能描述和价值"},
                {"header": "功能名称", "body": "功能描述和价值"}
            ],
            "notes": "4个核心功能，用图标+标题+一句话描述",
            "filling_prompt": "必须填入真实内容：提供4个核心功能，每个有 header（功能名称）和 body（一句话描述功能价值）。功能名称要具体。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "04",
            "title": "产品优势",
            "subtitle": "为什么选择我们",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 10,
            "type": "kpi_dashboard",
            "title": "效果数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 产品价值",
            "kpis": [
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"},
                {"value": "数字", "label": "效果指标", "delta": "变化趋势", "baseline": "vs 对比基准"}
            ],
            "notes": "展示4个核心效果指标，须有 delta 趋势和 baseline 对比基准",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4个核心效果指标，每个 KPI 有 value（具体数值）、label（效果说明）、delta（变化趋势）、baseline（对比基准）。指标要具体。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "section_divider",
            "number": "05",
            "title": "客户案例",
            "subtitle": "真实客户的成功故事",
            "filling_prompt": "章节分隔页，固定内容。"
        },
        {
            "index": 12,
            "type": "image_text",
            "title": "客户案例",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "行业",
            "header": "客户名称或项目名称",
            "sub_header": "合作项目名称",
            "paragraph": "详细描述客户背景、合作过程、实施效果和客户评价，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示客户成功案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为合作项目名称（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述客户背景、合作过程、实施效果和客户评价，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 13,
            "type": "image_text",
            "title": "实施效果",
            "content_type": "image_text",
            "layout_hint": "left-image",
            "kicker": "实施效果",
            "header": "项目名称",
            "sub_header": "项目简介",
            "paragraph": "详细描述项目实施过程、克服的困难和取得的量化成果，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示实施效果，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为项目标题（不超过35字）；sub_header 为项目简介（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述项目实施过程、克服的困难和取得的量化成果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "联系我们",
            "key_points": [
                "01 立即体验：免费试用说明",
                "02 预约演示：服务说明"
            ],
            "thank_you": "感谢关注",
            "contact": "官网 | 热线 | 扫码咨询",
            "notes": "结尾页，明确行动号召和联系方式",
            "filling_prompt": "必须填入真实内容：key_points 提供2个行动号召；contact 填写真实联系方式。禁止保留花括号。"
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
