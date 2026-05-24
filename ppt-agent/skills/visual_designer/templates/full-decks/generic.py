TEMPLATE = {
    "name": "generic",
    "name_cn": "通用模板",
    "description": "通用演示模板，结构均衡，覆盖封面、目录、内容、图表、总结，每章节间有分割页过渡",
    "target_audience": "不限",
    "typical_slides": 24,
    "typical_duration": "30-45分钟",
    "palette": "ocean_soft",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "philosophy": "结构清晰，章节分明，分割页帮助观众理解整体结构",
        "key_principles": [
            "标题简洁有力，突出主题",
            "每个章节前用分割页预告内容",
            "图表配合文字，增强说服力",
            "金句收尾，加深印象",
            "总结页回顾全书核心要点"
        ]
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "",
            "subtitle": "",
            "author": "",
            "date": "",
            "notes": "开场封面",
            "filling_prompt": "必填字段：title（主题名称）、subtitle（副标题）、author（演讲者姓名）、date（日期）。禁止使用占位符。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "",
            "kicker": "",
            "items": [],
            "notes": "目录页，列出全部章节",
            "filling_prompt": "必填字段：items（章节列表，格式为'XX  章节名称'）。根据实际内容填写4-6个章节。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "",
            "subtitle": "",
            "notes": "第一章分割页——预告本章内容",
            "filling_prompt": "必填字段：number（章节编号，如01、02）、title（章节标题）、subtitle（副标题）。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "",
            "kicker": "",
            "section_header": "",
            "bullets": [],
            "highlight_stats": [],
            "notes": "标准内容页，展示核心观点",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、section_header（小节标题）、bullets（核心要点列表，4-6条）、highlight_stats（右侧数据卡片，可选）。"
        },
        {
            "index": 5,
            "type": "image_text",
            "title": "",
            "kicker": "",
            "image_placeholder": "",
            "text_alignment": "",
            "section_title": "",
            "bullets": [],
            "notes": "图文混排页",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、layout（布局方向：left-image或right-image）、section_title（小节标题）、bullets（要点列表）。"
        },
        {
            "index": 6,
            "type": "deep_dive",
            "title": "",
            "kicker": "",
            "lede": "",
            "left_header": "",
            "right_header": "",
            "key_points": [],
            "analysis": [],
            "case_example": [],
            "data_evidence": [],
            "notes": "双栏详解，左栏分析，右栏案例",
            "filling_prompt": "必填字段：title（主题）、kicker（领域标签）、lede（一句话概括）、left_header（左栏标题）、right_header（右栏标题）、key_points（要点列表）、case_example（案例列表）、data_evidence（数据指标列表）。"
        },
        {
            "index": 7,
            "type": "quote_slide",
            "quote": "",
            "attribution": "",
            "kicker": "",
            "notes": "金句/引用页——章节收尾",
            "filling_prompt": "必填字段：quote（金句或引用，40-80字）、attribution（出处/作者）、kicker（标签）。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "02",
            "title": "",
            "subtitle": "",
            "notes": "第二章分割页",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。"
        },
        {
            "index": 9,
            "type": "stat_slide",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "stats": [],
            "notes": "关键数据展示",
            "filling_prompt": "必填字段：stats（关键指标列表，每项包含number、unit、label、trend）。根据实际数据填写。"
        },
        {
            "index": 10,
            "type": "timeline",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "direction": "",
            "events": [],
            "notes": "时间轴展示发展历程",
            "filling_prompt": "必填字段：title（时间轴主题）、direction（方向：horizontal或vertical）、events（里程碑列表，每项包含year、title、desc）。4-6个里程碑事件。"
        },
        {
            "index": 11,
            "type": "process_flow",
            "title": "",
            "kicker": "",
            "direction": "",
            "steps": [],
            "notes": "流程图展示步骤或阶段",
            "filling_prompt": "必填字段：title（流程名称）、direction（方向）、steps（步骤列表，每项包含num、title、desc）。5-8个步骤。"
        },
        {
            "index": 12,
            "type": "two_column",
            "title": "",
            "kicker": "",
            "left_header": "",
            "right_header": "",
            "left_bullets": [],
            "right_bullets": [],
            "source": "",
            "notes": "左右对比展示",
            "filling_prompt": "必填字段：title（对比标题）、left_header（左栏标题）、right_header（右栏标题）、left_bullets（左栏要点）、right_bullets（右栏要点）、source（数据来源）。"
        },
        {
            "index": 13,
            "type": "section_divider",
            "number": "03",
            "title": "",
            "subtitle": "",
            "notes": "第三章分割页",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。"
        },
        {
            "index": 14,
            "type": "card_grid",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "layout": "",
            "cards": [],
            "notes": "卡片网格，展示多维度信息",
            "filling_prompt": "必填字段：title（页面标题）、layout（布局：2x2或2x3）、cards（卡片列表，每项包含header、body）。4-6张卡片。"
        },
        {
            "index": 15,
            "type": "three_column",
            "title": "",
            "kicker": "",
            "columns": [],
            "notes": "三栏并列展示",
            "filling_prompt": "必填字段：title（页面标题）、columns（三栏列表，每栏包含header和bullets）。每栏3-4个要点。"
        },
        {
            "index": 16,
            "type": "image_hero",
            "title": "",
            "subtitle": "",
            "kicker": "",
            "image_placeholder": "",
            "stats": [],
            "notes": "大图页，用视觉冲击强化核心理念",
            "filling_prompt": "必填字段：title（核心理念一句话）、subtitle（补充说明）、stats（3个关键数据，每项包含value和label）。"
        },
        {
            "index": 17,
            "type": "case_study",
            "title": "",
            "kicker": "",
            "context": "",
            "problem": "",
            "solution": "",
            "results": [],
            "notes": "案例详解：背景-痛点-方案-成果",
            "filling_prompt": "必填字段：kicker（标签）、title（案例名称）、context（背景）、problem（痛点）、solution（解决方案）、results（成果指标列表）。"
        },
        {
            "index": 18,
            "type": "section_divider",
            "number": "04",
            "title": "",
            "subtitle": "",
            "notes": "第四章分割页",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。"
        },
        {
            "index": 19,
            "type": "example_detail",
            "title": "",
            "kicker": "",
            "lede": "",
            "context_block": "",
            "solution_block": "",
            "metrics": [],
            "takeaway": "",
            "notes": "案例深度展开",
            "filling_prompt": "必填字段：kicker（标签）、title（案例名称）、lede（一句话概括）、context_block（背景描述）、solution_block（方案描述）、metrics（效果指标）、takeaway（总结语）。"
        },
        {
            "index": 20,
            "type": "bar_chart",
            "title": "",
            "subtitle": "",
            "notes": "柱状图展示数据对比",
            "filling_prompt": "必填字段：title（图表标题）、subtitle（副标题）。生成器会渲染图表。"
        },
        {
            "index": 21,
            "type": "comparison_table",
            "title": "",
            "subtitle": "",
            "headers": [],
            "rows": [],
            "notes": "表格对比展示",
            "filling_prompt": "必填字段：title（表格标题）、subtitle（副标题）、headers（表头列表）、rows（数据行列表）。"
        },
        {
            "index": 22,
            "type": "kpi_dashboard",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "layout_hint": "",
            "kpis": [],
            "notes": "关键指标看板",
            "filling_prompt": "必填字段：title（页面标题）、subtitle（副标题）、kpis（指标列表，每项包含value、label、delta、baseline）。4个指标。"
        },
        {
            "index": 23,
            "type": "section_divider",
            "number": "05",
            "title": "",
            "subtitle": "",
            "notes": "终章分割页——预告总结",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。"
        },
        {
            "index": 24,
            "type": "summary_slide",
            "title": "",
            "key_points": [],
            "thank_you": "",
            "contact": "",
            "notes": "最后一页：总结回顾全文核心要点，致谢",
            "filling_prompt": "必填字段：title（如'总结'或'回顾'）、key_points（核心要点列表，格式为'XX  核心内容'，4个要点）、thank_you（结束语）、contact（联系方式，可选）。禁止使用示例数据。"
        }
    ],
    "design_tips": [
        "每个章节前用分割页明确预告内容",
        "内容精炼，每页聚焦一个核心观点",
        "文字不宜过多，提炼关键词",
        "善用图表，让数据说话",
        "配色统一，保持视觉一致性",
        "总结页回顾全文，强化记忆"
    ]
}
