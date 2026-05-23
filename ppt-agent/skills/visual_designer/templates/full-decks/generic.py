TEMPLATE = {
    "name": "generic",
    "name_cn": "通用模板",
    "description": "通用演示模板，结构均衡，覆盖封面、目录、内容、图表、总结，适合各类场景自由编辑",
    "target_audience": "不限",
    "typical_slides": 31,
    "typical_duration": "45-60分钟",
    "palette": "ocean_soft",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "philosophy": "结构清晰，内容完整，适合作为各类演示的基础框架",
        "key_principles": [
            "标题简洁有力，突出主题",
            "内容层次分明，重点突出",
            "图表配合文字，增强说服力",
            "金句收尾，加深印象",
            "整体风格统一，视觉舒适"
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
            "notes": "开场封面，标题要简洁有力",
            "filling_prompt": "必填字段：title（演示主题名称）、subtitle（副标题）、author（演讲者姓名）、date（日期）。禁止使用占位符或示例数据。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "",
            "kicker": "",
            "items": [],
            "notes": "章节概览，帮助观众了解整体结构",
            "filling_prompt": "必填字段：items（章节列表，格式为'XX  章节名称'）。根据实际内容填写4-6个章节，不要使用示例数据。"
        },
        {
            "index": 3,
            "type": "image_hero",
            "title": "",
            "subtitle": "",
            "kicker": "",
            "image_placeholder": "",
            "stats": [],
            "notes": "大图页用主视觉图配合核心理念文字",
            "filling_prompt": "必填字段：title（核心理念一句话）、subtitle（补充说明）、kicker（标签）、stats（3个关键数据，每项包含value和label）。根据实际内容填写，禁止使用示例数据。"
        },
        {
            "index": 4,
            "type": "stat_slide",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "stats": [],
            "notes": "用关键数据吸引注意力",
            "filling_prompt": "必填字段：stats（关键指标列表，每项包含number、unit、label、trend）。根据实际业务数据填写，数字必须是真实数据，禁止使用示例数据。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "",
            "title": "",
            "subtitle": "",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "必填字段：number（章节编号，如01、02）、title（章节标题）、subtitle（副标题）。根据实际章节内容填写。"
        },
        {
            "index": 6,
            "type": "content_slide",
            "title": "",
            "kicker": "",
            "section_header": "",
            "bullets": [],
            "highlight_stats": [],
            "notes": "标准内容页，适合展示核心观点",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、section_header（小节标题）、bullets（核心要点列表，4-6条）、highlight_stats（右侧数据卡片，可选）。根据实际内容填写，禁止使用示例数据。"
        },
        {
            "index": 7,
            "type": "timeline",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "direction": "",
            "events": [],
            "notes": "时间轴展示技术发展历程",
            "filling_prompt": "必填字段：title（时间轴主题）、kicker（标签）、subtitle（副标题）、direction（方向：horizontal或vertical）、events（里程碑列表，每项包含year、title、desc）。根据实际发展历程填写，4-6个里程碑事件，禁止使用示例数据。"
        },
        {
            "index": 8,
            "type": "process_flow",
            "title": "",
            "kicker": "",
            "direction": "",
            "steps": [],
            "notes": "用流程图展示操作步骤或发展阶段",
            "filling_prompt": "必填字段：title（流程名称）、kicker（标签）、direction（方向）、steps（步骤列表，每项包含num、title、desc）。根据实际流程填写，5-8个步骤，禁止使用示例数据。"
        },
        {
            "index": 9,
            "type": "deep_dive",
            "title": "",
            "kicker": "",
            "lede": "",
            "left_header": "",
            "key_points": [],
            "analysis": [],
            "right_header": "",
            "case_example": [],
            "data_evidence": [],
            "notes": "双栏详解，左要点右案例",
            "filling_prompt": "必填字段：title（技术主题）、kicker（领域标签）、lede（一句话概括）、left_header（左栏标题）、right_header（右栏标题）、key_points（技术要点列表）、analysis（分析维度列表）、case_example（案例列表）、data_evidence（数据指标列表）。根据实际技术内容填写，禁止使用示例数据。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "",
            "title": "",
            "subtitle": "",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。根据实际章节内容填写。"
        },
        {
            "index": 11,
            "type": "card_grid",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "layout": "",
            "cards": [],
            "notes": "卡片网格适合展示多维度信息",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、layout（布局：2x2或2x3）、cards（卡片列表，每项包含header、icon、body、footer）。根据实际产品能力填写，4-8张卡片，禁止使用示例数据。"
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
            "notes": "左右对比，突显差异化优势",
            "filling_prompt": "必填字段：title（对比标题）、kicker（标签）、left_header（左栏标题）、right_header（右栏标题）、left_bullets（左栏要点列表）、right_bullets（右栏要点列表）、source（数据来源）。根据实际对比内容填写，禁止使用示例数据。"
        },
        {
            "index": 13,
            "type": "image_text",
            "title": "",
            "kicker": "",
            "image_placeholder": "",
            "text_alignment": "",
            "section_title": "",
            "bullets": [],
            "notes": "图文混排页，左图右文或左文右图",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、layout（布局方向：left-image或right-image）、section_title（小节标题）、bullets（功能说明列表）。根据实际内容填写，禁止使用示例数据。"
        },
        {
            "index": 14,
            "type": "quote_slide",
            "quote": "",
            "attribution": "",
            "kicker": "",
            "notes": "金句页用名人名言增强说服力和感染力",
            "filling_prompt": "必填字段：quote（金句内容，40-80字）、attribution（演讲者姓名、职位、机构）、kicker（标签）。使用真实金句或根据实际演讲内容填写，禁止使用示例数据。"
        },
        {
            "index": 15,
            "type": "section_divider",
            "number": "",
            "title": "",
            "subtitle": "",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。根据实际章节内容填写。"
        },
        {
            "index": 16,
            "type": "three_column",
            "title": "",
            "kicker": "",
            "columns": [],
            "notes": "三栏并列展示多维度信息",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、columns（三栏列表，每栏包含header和bullets）。根据实际内容填写，3栏，每栏3-4个要点，禁止使用示例数据。"
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
            "notes": "背景-痛点-方案-成果四段式",
            "filling_prompt": "必填字段：kicker（领域标签）、title（案例名称）、context（背景）、problem（痛点）、solution（解决方案）、results（成果指标列表，每项包含metric、value、comparison）。根据实际案例填写，禁止使用示例数据。"
        },
        {
            "index": 18,
            "type": "kpi_dashboard",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "layout_hint": "",
            "kpis": [],
            "notes": "多维度关键指标一目了然",
            "filling_prompt": "必填字段：title（页面标题）、subtitle（副标题）、kicker（标签）、kpis（指标列表，每项包含value、label、delta、baseline）。根据实际业务数据填写，4个指标，禁止使用示例数据。"
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
            "notes": "案例背景、方案、成果完整呈现",
            "filling_prompt": "必填字段：kicker（领域标签）、title（案例名称）、lede（一句话概括）、context_block（背景描述）、solution_block（解决方案描述）、metrics（效果指标列表，每项包含value、label、trend）、takeaway（总结语）。根据实际案例填写，禁止使用示例数据。"
        },
        {
            "index": 20,
            "type": "kanban",
            "title": "",
            "kicker": "",
            "columns": [],
            "progress": 0,
            "notes": "看板页展示项目实施进度",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、columns（列列表，每列包含header、color、cards，每张卡片包含text、tag、priority）、progress（整体进度百分比）。根据实际项目填写，禁止使用示例数据。"
        },
        {
            "index": 21,
            "type": "section_divider",
            "number": "",
            "title": "",
            "subtitle": "",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。根据实际章节内容填写。"
        },
        {
            "index": 22,
            "type": "kpi_dashboard",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "layout_hint": "",
            "kpis": [],
            "notes": "多维度关键指标一目了然",
            "filling_prompt": "必填字段：title（页面标题）、subtitle（副标题）、kicker（标签）、kpis（指标列表，每项包含value、label、delta、baseline）。根据实际业务数据填写，4个指标，禁止使用示例数据。"
        },
        {
            "index": 23,
            "type": "bar_chart",
            "title": "",
            "subtitle": "",
            "notes": "柱状图展示数据对比",
            "filling_prompt": "必填字段：title（图表标题）、subtitle（副标题，说明数据维度和单位）。生成器会渲染图表，数据由用户提供或根据实际业务填写，禁止使用示例数据。"
        },
        {
            "index": 24,
            "type": "line_chart",
            "title": "",
            "subtitle": "",
            "notes": "折线图展示趋势变化",
            "filling_prompt": "必填字段：title（图表标题）、subtitle（副标题，说明数据维度和单位）。生成器会渲染图表，数据由用户提供或根据实际业务填写，禁止使用示例数据。"
        },
        {
            "index": 25,
            "type": "pie_chart",
            "title": "",
            "subtitle": "",
            "notes": "饼图展示构成比例",
            "filling_prompt": "必填字段：title（图表标题）、subtitle（副标题，说明数据维度）。生成器会渲染图表，数据由用户提供或根据实际业务填写，禁止使用示例数据。"
        },
        {
            "index": 26,
            "type": "table",
            "title": "",
            "subtitle": "",
            "headers": [],
            "rows": [],
            "notes": "表格页展示对比数据",
            "filling_prompt": "必填字段：title（页面标题）、subtitle（副标题）、headers（表头列表，第1列为维度名称，后2-3列为各版本/方案）、rows（数据行列表，每行第1个单元格为维度名称，其余为该维度在各版本/方案中的值）。根据实际对比内容填写，禁止使用示例数据。"
        },
        {
            "index": 27,
            "type": "section_divider",
            "number": "",
            "title": "",
            "subtitle": "",
            "notes": "章节分隔页起过渡作用",
            "filling_prompt": "必填字段：number（章节编号）、title（章节标题）、subtitle（副标题）。根据实际章节内容填写。"
        },
        {
            "index": 28,
            "type": "brand_focus",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "center_text": "",
            "surrounding_points": [],
            "principles": [],
            "notes": "核心价值以视觉化方式呈现",
            "filling_prompt": "必填字段：title（页面标题）、subtitle（副标题）、kicker（标签）、center_text（中心圆圈文字，用换行分隔两行）、surrounding_points（围绕中心的价值点列表，每项包含title、description、angle）、principles（核心理念列表，每项包含title、description）。根据实际品牌战略填写，禁止使用示例数据。"
        },
        {
            "index": 29,
            "type": "image_text",
            "title": "",
            "kicker": "",
            "image_placeholder": "",
            "text_alignment": "",
            "section_title": "",
            "bullets": [],
            "notes": "图文混排展示内容",
            "filling_prompt": "必填字段：title（页面标题）、kicker（标签）、layout（布局方向：left-image或right-image）、section_title（小节标题）、bullets（要点列表）。根据实际内容填写，禁止使用示例数据。"
        },
        {
            "index": 30,
            "type": "region_map",
            "title": "",
            "kicker": "",
            "subtitle": "",
            "regions": [],
            "regions_detail": [],
            "notes": "区域版图展示全国业务布局",
            "filling_prompt": "必填字段：title（页面标题）、subtitle（副标题）、kicker（标签）、regions（区域列表，每项包含name、value、trend）、regions_detail（详细区域列表，每项包含name、value、trend、detail）。根据实际业务布局填写，禁止使用示例数据。"
        },
        {
            "index": 31,
            "type": "summary_slide",
            "title": "",
            "key_points": [],
            "thank_you": "",
            "contact": "",
            "notes": "总结回顾并留下联系方式",
            "filling_prompt": "必填字段：title（页面标题）、key_points（核心要点列表，格式为'XX  核心内容'）、thank_you（结束语）、contact（联系方式）。根据实际演示内容填写，禁止使用示例数据。"
        }
    ],
    "design_tips": [
        "内容精炼，每页聚焦一个核心观点",
        "文字不宜过多，提炼关键词",
        "善用图表，让数据说话",
        "配色统一，保持视觉一致性",
        "留白适度，避免信息过载",
        "字体大小适中，确保后排可读"
    ]
}
