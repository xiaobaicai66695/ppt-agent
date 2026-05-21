TEMPLATE = {
    "name": "weekly-report",
    "name_cn": "周报/月报",
    "description": "适合团队周报、项目月报、工作汇报等场景。简洁高效，重点突出，数据驱动。",
    "target_audience": "团队负责人、项目经理、管理层",
    "typical_slides": 9,
    "typical_duration": "5-10分钟",
    "palette": "sage_calm",
    "typography": {
        "header": "Calibri",
        "body": "Calibri Light",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "reporting_principles": [
            "简洁：只汇报关键信息，不罗列细节",
            "数据：用数据说话，避免空泛描述",
            "重点：突出本周关键进展和问题",
            "计划：明确下周工作安排"
        ],
        "reader_mindset": "领导关心的是：进度如何？有什么问题？需要什么支持？"
    },
    "structure_guide": {
        "overview": "本周整体情况，一句话概括",
        "completed": "本周完成的工作",
        "in_progress": "进行中的工作",
        "issues": "遇到的问题和风险",
        "next_week": "下周工作计划"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "周报：项目名称",
            "subtitle": "时间段（如2025年第10周 3.3-3.9）",
            "author": "汇报人姓名",
            "date": "实际日期",
            "notes": "标题页简洁，注明时间段",
            "filling_prompt": "必须填入真实内容：title 中填入实际周/月报周期和项目名称，subtitle 为时间段，author 为汇报人姓名，date 为实际日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  本周期概览",
                "02  完成事项",
                "03  进行中事项",
                "04  关键进展",
                "05  问题与计划"
            ],
            "notes": "让观众快速了解本报告结构，每项一行即可",
            "filling_prompt": "目录页为固定结构，无需额外填充。"
        },
        {
            "index": 3,
            "type": "kpi_dashboard",
            "title": "本周期概览",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 周期概览",
            "kpis": [
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "数字", "label": "指标名称", "delta": "变化趋势", "baseline": "对比基准"},
                {"value": "状态", "label": "风险等级", "delta": "说明", "baseline": "说明"}
            ],
            "notes": "4个核心指标，展示本周期整体情况",
            "filling_prompt": "必须填入真实内容：提供4个本周期的核心指标数据，每个有 value（数字）、label（说明）、delta（变化趋势）、baseline（对比基准）。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "已完成事项",
            "content_type": "example_detail",
            "kicker": "实例 · 完成情况",
            "lede": "一句话概括本周期的核心成果",
            "context_block": "描述工作背景（1-2句话）。",
            "solution_block": "具体说明完成工作的过程和关键动作（2-3句话）。",
            "metrics": [
                {"value": "数字", "label": "指标1", "trend": "变化趋势"},
                {"value": "数字", "label": "指标2", "trend": "变化趋势"},
                {"value": "数字", "label": "指标3", "trend": "变化趋势"}
            ],
            "takeaway": "一句话总结成果意义。",
            "notes": "用具体数字说明完成情况",
            "filling_prompt": "必须填入真实内容：lede 一句话概括本周期的核心成果；context_block 描述工作背景（1-2句话）；solution_block 具体说明完成工作的过程和关键动作（2-3句话）；metrics_grid 提供3个量化指标，每个有 value（数字）、label（说明）、trend（对比上周期）；takeaway 用一句话总结成果意义。禁止虚构数据。"
        },
        {
            "index": 5,
            "type": "example_detail",
            "title": "进行中事项",
            "content_type": "example_detail",
            "kicker": "实例 · 进行中事项",
            "lede": "一句话概括当前的工作重点和挑战",
            "context_block": "描述当前事项的背景和目标（1-2句话）。",
            "solution_block": "具体说明当前进展、遇到的困难和解决思路（2-3句话）。",
            "metrics": [
                {"value": "百分比", "label": "完成度", "trend": "预计完成时间"},
                {"value": "百分比", "label": "完成度", "trend": "预计完成时间"},
                {"value": "数字", "label": "剩余工作量", "trend": "资源情况"}
            ],
            "takeaway": "一句话总结下一步关键行动。",
            "notes": "说明进行中事项的当前进度和挑战",
            "filling_prompt": "必须填入真实内容：lede 一句话概括当前的工作重点和挑战；context_block 描述当前事项的背景和目标（1-2句话）；solution_block 具体说明当前进展、遇到的困难和解决思路（2-3句话）；metrics_grid 提供3个进度指标，每个有 value、label、trend；takeaway 用一句话总结下一步关键行动。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "关键进展",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "关键进展",
            "header": "进展标题",
            "sub_header": "进展概述",
            "paragraph": "详细描述项目的关键进展、突破点和实际成果，用流畅的段落形式呈现，禁止罗列要点。",
            "references": [
                "内部系统URL或权威来源URL"
            ],
            "notes": "用图文混排展示关键进展，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少1个URL，如无外部数据需求可填入内部系统URL），再填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为进展标题（不超过35字）；sub_header 为进展概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述项目的关键进展、突破点和实际成果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "问题与风险",
            "content_type": "content_slide",
            "issues": [
                {
                    "issue": "问题名称",
                    "severity": "严重程度",
                    "description": "问题描述",
                    "solution": "解决方案",
                    "status": "状态"
                },
                {
                    "issue": "问题名称",
                    "severity": "严重程度",
                    "description": "问题描述",
                    "solution": "解决方案",
                    "status": "状态"
                }
            ],
            "notes": "列出当前面临的问题和风险，以及应对措施",
            "filling_prompt": "必须填入真实内容：列出1-3个当前面临的问题/风险，每条说明问题描述和应对措施。坦诚汇报，不回避问题。"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "下周计划",
            "content_type": "content_slide",
            "next_week_items": [
                {"task": "任务名称", "priority": "优先级", "due": "截止时间"},
                {"task": "任务名称", "priority": "优先级", "due": "截止时间"},
                {"task": "任务名称", "priority": "优先级", "due": "截止时间"},
                {"task": "任务名称", "priority": "优先级", "due": "截止时间"}
            ],
            "notes": "下周的工作计划和目标",
            "filling_prompt": "必须填入真实内容：列出下周工作计划（3-5条），每条具体可执行，每项标注优先级和截止时间。"
        },
        {
            "index": 9,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 本周期核心成果：描述",
                "02 下周工作重点：描述",
                "03 需要支持：描述"
            ],
            "thank_you": "感谢聆听",
            "notes": "简洁结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（本周期核心成果2条+下周工作重点1条）。"
        }
    ],
    "design_tips": [
        "周报要简洁，不要堆砌文字",
        "用数据说话，核心指标突出",
        "问题与风险要坦诚，不要回避",
        "计划要具体可执行",
        "每项行动标注优先级和截止时间",
        "善用颜色区分紧急程度",
        "图表比文字更直观",
        "保持格式一致，方便对比"
    ],
    "common_metrics": {
        "progress": ["完成率", "进度百分比", "里程碑达成"],
        "quality": ["Bug数", "测试通过率", "缺陷密度"],
        "team": ["人力投入", "人均产出", "协作情况"],
        "risk": ["风险数量", "风险等级", "应对措施"]
    },
    "anti_patterns": [
        "流水账式记录，缺乏重点",
        "报喜不报忧，隐藏问题",
        "计划太虚，缺乏具体行动",
        "数据不准确，前后矛盾"
    ]
}
