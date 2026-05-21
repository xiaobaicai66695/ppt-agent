TEMPLATE = {
    "name": "personal-summary",
    "name_cn": "述职报告",
    "description": "适合个人总结、述职报告、年终总结等场景。重点突出，成果可见，计划明确。",
    "target_audience": "领导、同事、评审",
    "typical_slides": 12,
    "typical_duration": "10-15分钟",
    "palette": "report_green",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "summary_strategy": "述职的核心是展示：你做了什么、做得怎么样，下一步怎么做",
        "key_principles": [
            "成果要用数据说话",
            "问题要坦诚面对",
            "计划要具体可执行",
            "态度要谦逊但不卑微"
        ],
        "evaluation_criteria": {
            "performance": "业绩达成情况",
            "capability": "能力提升情况",
            "collaboration": "团队协作情况",
            "growth": "成长与反思"
        }
    },
    "self_evaluation_tips": {
        "strengths": "突出自己的核心优势和亮点",
        "improvements": "诚实地分析不足，但不要自我贬低",
        "learnings": "强调从经历中学到了什么"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "年度/季度述职报告",
            "subtitle": "姓名 | 部门 | 岗位",
            "author": "姓名",
            "date": "述职日期",
            "notes": "标题页简洁正式",
            "filling_prompt": "必须填入真实内容：title 为时间段（如'2024年度'）+ 述职报告，subtitle 为姓名和岗位信息，author 为姓名，date 为述职日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  工作概述",
                "02  主要成果",
                "03  问题反思",
                "04  未来计划"
            ],
            "notes": "让领导快速了解汇报结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "工作概述",
            "subtitle": "岗位职责与工作回顾",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "工作概览",
            "content_type": "example_detail",
            "kicker": "实例 · 工作概览",
            "lede": "一句话概括本周期工作整体情况",
            "context_block": "描述述职人所在岗位职责和主要工作范围（1-2句话）。",
            "solution_block": "具体说明本周期内完成的主要工作和取得的成绩（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "指标1", "trend": "变化趋势"},
                {"value": "数字", "label": "指标2", "trend": "变化趋势"},
                {"value": "数字", "label": "指标3", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结本周期工作的价值。",
            "notes": "用数据展示工作量和成果",
            "filling_prompt": "必须填入真实内容：lede 一句话概括本周期工作整体情况；context_block 描述述职人所在岗位职责和主要工作范围（1-2句话）；solution_block 具体说明本周期内完成的主要工作和取得的成绩（2-3句话）；metrics_grid 提供3个量化指标，每个有 value、label、trend；takeaway 用一句话总结本周期工作的价值。禁止虚构数据。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "主要成果",
            "subtitle": "核心业绩与贡献",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "核心业绩",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "核心业绩",
            "header": "项目标题",
            "sub_header": "项目概述",
            "paragraph": "详细描述项目的实施过程、克服的困难和取得的成果，用流畅的段落形式呈现，禁止罗列要点。",
            "notes": "用图文混排展示核心业绩，增强可信性",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为项目标题（不超过35字）；sub_header 为项目概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述项目的实施过程、克服的困难和取得的成果，用流畅的段落形式呈现，禁止罗列要点。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "card_grid",
            "title": "其他成果",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "成果名称", "body": "成果描述"},
                {"header": "成果名称", "body": "成果描述"},
                {"header": "成果名称", "body": "成果描述"},
                {"header": "成果名称", "body": "成果描述"}
            ],
            "notes": "4个其他重要成果",
            "filling_prompt": "必须填入真实内容：提供4个其他工作成果，每个有 header（成果名称）和 body（一句话描述成果内容）。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "问题反思",
            "subtitle": "不足与改进",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "two_column",
            "title": "反思与改进",
            "content_type": "two_column",
            "kicker": "自我剖析 · 持续成长",
            "left_header": "不足与反思",
            "left_sections": {
                "key_points": [
                    "不足1：具体表现和影响",
                    "不足2：具体表现和影响",
                    "不足3：具体表现和影响"
                ],
                "analysis": [
                    "根因分析1",
                    "根因分析2"
                ]
            },
            "right_header": "改进计划",
            "right_sections": {
                "key_points": [
                    "改进措施1",
                    "改进措施2",
                    "改进措施3"
                ],
                "analysis": [
                    "预期效果",
                    "成长周期"
                ]
            },
            "notes": "坦诚分析不足，给出具体改进计划，展示自我驱动的成长意愿",
            "filling_prompt": "必须填入真实内容：left_header 为'不足与反思'，left_sections.key_points 列出2-4条工作中的不足（每条说明具体表现和影响），left_sections.analysis 列出2条根因分析；right_header 为'改进计划'，right_sections.key_points 列出对应的改进措施（每条说明具体行动、计划和时间节点），right_sections.analysis 列出预期效果。注意不足和改进一一对应。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "未来计划",
            "subtitle": "下阶段目标与规划",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "process_flow",
            "title": "下阶段计划",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "目标1", "desc": "具体行动"},
                {"num": "02", "title": "目标2", "desc": "具体行动"},
                {"num": "03", "title": "目标3", "desc": "具体行动"},
                {"num": "04", "title": "目标4", "desc": "具体行动"}
            ],
            "notes": "4个下阶段工作目标",
            "filling_prompt": "必须填入真实内容：提供4个下阶段工作目标，每个有 title（目标名称）和 desc（具体行动）。目标要具体可衡量。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 核心成果：成果描述",
                "02 主要收获：收获描述",
                "03 努力方向：方向描述"
            ],
            "thank_you": "感谢聆听，请领导批评指正！",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心成果+主要收获+努力方向）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "述职要实事求是，成果和问题都要有",
        "数据要真实，用具体数字说话",
        "反思要诚恳，改进计划要可行",
        "计划要具体，有时间节点",
        "PPT要简洁，不要堆砌文字",
        "用数据展示成果，而非罗列工作内容",
        "适当使用对比图展示进步",
        "准备Q&A，预判领导可能的问题"
    ],
    "self_review_template": {
        "performance": {
            "kpis": "KPI完成情况",
            "projects": "重点项目参与",
            "innovation": "创新贡献"
        },
        "capability": {
            "professional": "专业能力",
            "management": "管理能力",
            "leadership": "领导力"
        },
        "collaboration": {
            "teamwork": "团队协作",
            "cross_function": "跨部门协作",
            "stakeholder": "干系人管理"
        }
    },
    "common_questions": {
        "achievements": "最满意的成就是什么？",
        "challenges": "遇到的最大挑战是什么？",
        "improvements": "觉得自己还需要提升什么？",
        "next_year": "对下一年有什么规划？"
    }
}
