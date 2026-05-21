TEMPLATE = {
    "name": "politics-ideology",
    "name_cn": "思政/团课",
    "description": "适合思政教育、团课培训、爱国主义教育等场景。政治性强，价值观明确，结构清晰。",
    "target_audience": "学生、共青团员、党员、干部",
    "typical_slides": 16,
    "typical_duration": "20-30分钟",
    "palette": "patriotic_blue",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "teaching_principles": [
            "政治表达要准确规范，与中央精神保持一致",
            "理论联系实际，避免空洞说教",
            "以理服人，以情感人",
            "注重互动引导，激发学员思考"
        ],
        "key_elements": [
            "理论学习：核心思想与理论基础",
            "案例分析：典型案例与榜样力量",
            "实践要求：如何将思想落实到行动",
            "总结提升：巩固学习成果"
        ]
    },
    "teaching_methods": {
        "theory": "讲授法：系统讲解理论知识",
        "case": "案例法：通过故事引发共鸣",
        "discussion": "讨论法：引导学员思考交流",
        "practice": "实践法：布置具体行动任务"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "思政教育主题名称",
            "subtitle": "概括性副标题",
            "author": "演讲者姓名",
            "date": "实际日期",
            "notes": "标题页庄重大气，体现政治严肃性",
            "filling_prompt": "必须填入真实内容：title 为本次思政教育的主题名称，subtitle 为概括性副标题，author 为演讲者姓名，date 为实际日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  理论学习",
                "02  案例分析",
                "03  实践要求",
                "04  总结提升"
            ],
            "notes": "让学员快速了解课程结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "理论学习",
            "subtitle": "核心思想与理论基础",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "核心思想解读",
            "content_type": "example_detail",
            "kicker": "实例 · 理论学习",
            "lede": "一句话概括核心思想的价值",
            "context_block": "描述学习和践行的背景必要性（1-2句话）。",
            "solution_block": "展开核心思想的主要内涵和实践要求（2-3句话）。",
            "metrics": [
                {"value": "政策文件", "label": "重要文件", "trend": "发布时间"},
                {"value": "数字", "label": "规模指标", "trend": "变化趋势"},
                {"value": "数字", "label": "结构指标", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结如何将理论转化为行动。",
            "notes": "解读核心思想的关键要点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：lede 一句话概括核心思想的价值；context_block 描述学习和践行的背景必要性（1-2句话）；solution_block 展开核心思想的主要内涵和实践要求（2-3句话）；metrics_grid 提供3个指标；takeaway 用一句话总结如何将理论转化为行动。禁止虚构理论内容。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "quote_slide",
            "title": "重要论述",
            "content_type": "quote_slide",
            "quotes": [
                {
                    "text": "领导人重要论述内容（需准确引用）",
                    "source": "出处（讲话名称、时间、场合）",
                    "date": "日期"
                }
            ],
            "notes": "展示领导人重要论述",
            "filling_prompt": "必须填入真实内容：提供1-2条重要论述（需准确引用），注明出处。禁止虚构论述。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "案例分析",
            "subtitle": "典型案例与榜样力量",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "image_text",
            "title": "典型案例",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "榜样力量",
            "header": "案例标题",
            "sub_header": "身份简介",
            "paragraph": "详细描述该人物或事件的主要事迹、贡献和影响，包含具体事例，禁止罗列要点。",
            "references": [
                "https://权威来源URL1",
                "https://权威来源URL2"
            ],
            "notes": "用图文混排展示典型案例，增强可信性和感染力",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（不超过35字）；sub_header 为身份简介（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该人物或事件的主要事迹、贡献和影响，包含具体事例，禁止罗列要点。references 列出 URL。禁止虚构人物和事迹。"
        },
        {
            "index": 8,
            "type": "card_grid",
            "title": "榜样人物",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "姓名 | 身份", "body": "主要事迹描述（100-120字），包含具体数据和贡献。"},
                {"header": "姓名 | 身份", "body": "主要事迹描述（100-120字），包含具体数据和贡献。"},
                {"header": "姓名 | 身份", "body": "主要事迹描述（100-120字），包含具体数据和贡献。"},
                {"header": "姓名 | 身份", "body": "主要事迹描述（100-120字），包含具体数据和贡献。"}
            ],
            "notes": "展示2-4位榜样人物的简要事迹",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4位榜样人物，每人有 header（姓名+身份）和 body（详细描述其主要事迹、贡献和影响力，100-120字，包含具体事例或数据）。禁止虚构人物。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "实践要求",
            "subtitle": "如何将思想落实到行动",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "process_flow",
            "title": "实践路径",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "路径名称", "desc": "具体做法（不超过30字）"},
                {"num": "02", "title": "路径名称", "desc": "具体做法（不超过30字）"},
                {"num": "03", "title": "路径名称", "desc": "具体做法（不超过30字）"},
                {"num": "04", "title": "路径名称", "desc": "具体做法（不超过30字）"}
            ],
            "notes": "4个具体可行的实践路径",
            "filling_prompt": "必须填入真实内容：提供4个具体可操作的实践路径，每条有 title（行动名称）和 desc（具体做法，不超过30字）。禁止空泛口号。"
        },
        {
            "index": 11,
            "type": "example_detail",
            "title": "行动指南",
            "content_type": "example_detail",
            "kicker": "实例 · 实践要求",
            "lede": "一句话概括关键行动",
            "context_block": "描述实践中的主要差距和不足（1-2句话）。",
            "solution_block": "具体说明如何将思想认识转化为实际行动（2-3句话）。",
            "metrics": [
                {"value": "行动指标1", "label": "指标名称", "trend": "执行情况"},
                {"value": "行动指标2", "label": "指标名称", "trend": "执行情况"},
                {"value": "行动指标3", "label": "指标名称", "trend": "执行情况"}
            ],
            "takeaway": "一句话总结如何坚持知行合一。",
            "notes": "具体的行动建议",
            "filling_prompt": "必须填入真实内容：lede 一句话概括关键行动；context_block 描述实践中的主要差距和不足（1-2句话）；solution_block 具体说明如何将思想认识转化为实际行动（2-3句话）；metrics_grid 提供3个行动指标；takeaway 用一句话总结如何坚持知行合一。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "04",
            "title": "总结提升",
            "subtitle": "巩固学习成果",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "example_detail",
            "title": "学习要点回顾",
            "content_type": "example_detail",
            "kicker": "实例 · 总结提升",
            "lede": "一句话概括最核心的收获",
            "context_block": "回顾学习内容和方法（1-2句话）。",
            "solution_block": "总结核心要点和启示（2-3句话）。",
            "metrics": [
                {"value": "掌握程度", "label": "理论掌握", "trend": "提升"},
                {"value": "应用方向", "label": "实践应用", "trend": "清晰"},
                {"value": "学习计划", "label": "后续安排", "trend": "制定"}
            ],
            "takeaway": "一句话总结如何将学习成果转化为行动。",
            "notes": "回顾本节课的核心要点",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的收获；context_block 回顾学习内容和方法（1-2句话）；solution_block 总结核心要点和启示（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结如何将学习成果转化为行动。"
        },
        {
            "index": 14,
            "type": "example_detail",
            "title": "心得体会",
            "content_type": "example_detail",
            "kicker": "实例 · 心得体会",
            "lede": "一句话概括最深刻的感悟",
            "context_block": "描述学习中的思考和触动（1-2句话）。",
            "solution_block": "分享心得体会和行动计划（2-3句话）。",
            "metrics": [
                {"value": "触动程度", "label": "情感触动", "trend": "深刻"},
                {"value": "改进方向", "label": "明确方向", "trend": "清晰"},
                {"value": "带动作用", "label": "对团队的贡献", "trend": "带动更多人"}
            ],
            "takeaway": "一句话总结如何带动他人一起进步。",
            "notes": "引导学员思考和分享",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最深刻的感悟；context_block 描述学习中的思考和触动（1-2句话）；solution_block 分享心得体会和行动计划（2-3句话）；metrics_grid 提供3个反思指标；takeaway 用一句话总结如何带动他人一起进步。"
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "学习效果自评",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "自评 · 学习效果",
            "kpis": [
                {"value": "分数", "label": "理论理解程度", "delta": "vs 学习前+X分", "baseline": "满分10分"},
                {"value": "分数", "label": "情感认同程度", "delta": "vs 学习前+X分", "baseline": "满分10分"},
                {"value": "分数", "label": "行动意愿强度", "delta": "vs 学习前+X分", "baseline": "满分10分"},
                {"value": "计划", "label": "下一步计划", "delta": "制定中", "baseline": "待确定"}
            ],
            "notes": "帮助学员自我评估学习效果",
            "filling_prompt": "必须填入真实内容：提供4个学习效果自评维度，每个有 value（如分数或等级）、label（维度名称）、delta（变化趋势）。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心学习要点1",
                "02 核心学习要点2",
                "03 行动承诺：承诺内容"
            ],
            "thank_you": "感谢聆听",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心学习要点2条+行动承诺1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "政治表达要准确规范",
        "案例要有真实性和感染力",
        "实践建议要具体可操作",
        "避免空洞说教，注重以理服人",
        "设计要有仪式感和庄重感",
        "理论联系实际，引发共鸣",
        "设置互动环节，激发思考",
        "用故事而非说教传达价值观"
    ],
    "teaching_notes": {
        "interactive_activities": [
            "小组讨论：讨论主题",
            "情景剧：模拟场景",
            "演讲比赛：演讲主题",
            "志愿服务：服务内容"
        ],
        "assessment_methods": [
            "课堂参与度",
            "心得体会质量",
            "实践活动表现",
            "理论知识测试"
        ]
    }
}
