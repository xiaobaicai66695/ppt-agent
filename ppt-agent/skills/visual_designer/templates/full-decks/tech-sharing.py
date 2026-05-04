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
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{主题}",
            "subtitle": "技术分享",
            "author": "{演讲者}",
            "date": "{日期}",
            "notes": "开场标题页，留白充足，标题字体有重量感",
            "filling_prompt": "必须填入真实内容：title 为本次分享的实际技术主题名称（如'Kubernetes架构深度解析'），author 为演讲者姓名，date 为实际日期。禁止保留花括号占位符。"
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
            "filling_prompt": "目录页为固定结构，无需额外填充。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "背景与问题",
            "subtitle": "为什么需要这项技术",
            "notes": "章节分隔页，仪式感强",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "问题背景",
            "content_type": "content_slide",
            "content_plan": {
                "summary": "{一句话概括本节的核心问题}",
                "elements": [
                    {"type": "bullet_list", "items": []},
                    {"type": "callout", "text": ""}
                ]
            },
            "notes": "用具体数字说明痛点，不要空泛描述",
            "filling_prompt": "必须填入真实内容：提供3-4条具体痛点，每条要有具体数字说明（如'部署周期平均5-7天'、'故障恢复时间超过2小时'、'运维人力成本占IT预算40%'等）。禁止空泛描述。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "现有方案分析",
            "content_type": "two_column",
            "left_header": "传统方案",
            "right_header": "改进方向",
            "notes": "双栏对比，指出传统方案的不足和改进方向",
            "filling_prompt": "必须填入真实内容：left_header 为"传统方案局限性"，left_bullets 列出2-3个具体问题；right_header 为"改进方向"，right_bullets 列出2-3条对应的改进方案。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "核心原理",
            "subtitle": "技术实现的关键机制",
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "核心概念",
            "content_type": "content_slide",
            "notes": "用简洁的语言解释核心概念，配合示意图（文字描述即可）",
            "filling_prompt": "必须填入真实内容：用通俗语言解释3-4个核心概念，每条配合一句话说明。可用文字描述示意图内容（如'控制平面负责调度，所有节点上报状态'）。"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "关键算法/流程",
            "content_type": "process_flow",
            "direction": "horizontal",
            "notes": "用流程图展示核心步骤，3-5步为宜",
            "filling_prompt": "必须填入真实内容：提供3-5个核心步骤，每步有名称和一句话描述，展示该技术的工作流程。"
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
            "notes": "用文字描述架构图（组件+关系），不要求真实图片",
            "filling_prompt": "必须填入真实内容：用文字描述系统整体架构（组件名称+组件之间的关系，如'API Server接收请求 → Scheduler分配节点 → Kubelet执行 → 状态同步至etcd'）。"
        },
        {
            "index": 11,
            "type": "content_slide",
            "title": "核心模块详解",
            "content_type": "card_grid",
            "layout_hint": "2x2",
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
            "title": "应用案例：{客户/项目名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示具体应用案例，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 填具体行业领域（如'金融'、'电商'、'医疗'）；title 中的 {客户/项目名称} 替换为真实客户或项目名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX公司智能客服系统'，不超过35字）；sub_header 为合作项目名称（不超过35字）；bullets 列出3-4条应用效果或客户评价，每条不超过35字。references 列出 web_search 获取的 URL（至少2个）。禁止使用'某公司''某系统'等匿名实体；禁止虚构数据。"
        },
        {
            "index": 14,
            "type": "kpi_dashboard",
            "title": "关键数据",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 核心指标",
            "kpis": [
                {"value": "{数值1}", "label": "{效果说明1}", "delta": "{变化趋势1}", "baseline": "{对比基准1}"},
                {"value": "{数值2}", "label": "{效果说明2}", "delta": "{变化趋势2}", "baseline": "{对比基准2}"},
                {"value": "{数值3}", "label": "{效果说明3}", "delta": "{变化趋势3}", "baseline": "{对比基准3}"},
                {"value": "{数值4}", "label": "{效果说明4}", "delta": "{变化趋势4}", "baseline": "{对比基准4}"}
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
            "filling_prompt": "章节分隔页，固定内容，无需额外填充。"
        },
        {
            "index": 16,
            "type": "content_slide",
            "title": "核心要点",
            "content_type": "content_slide",
            "notes": "3-4条核心要点，用加粗序号",
            "filling_prompt": "必须填入真实内容：提供3-4条本次分享的核心要点，每条15字以内，精炼概括关键信息。"
        },
        {
            "index": 17,
            "type": "content_slide",
            "title": "未来方向",
            "content_type": "content_slide",
            "notes": "技术演进方向或后续规划",
            "filling_prompt": "必须填入真实内容：描述2-3个该技术的未来演进方向或后续规划，每条一句话。"
        },
        {
            "index": 18,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心点1}",
                "02 {核心点2}",
                "03 {核心点3}"
            ],
            "thank_you": "感谢聆听",
            "contact": "{联系方式}",
            "notes": "结尾页，核心回顾 + 感谢",
            "filling_prompt": "必须填入真实内容：key_points 提供3个核心回顾要点；contact 填写真实联系方式。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "技术分享要注重内容深度，不要堆砌文字",
        "每章节用 section_divider 清晰划分",
        "用数字说明效果比文字描述更有说服力",
        "架构图用文字描述组件关系即可，不需要真实图片"
    ]
}
