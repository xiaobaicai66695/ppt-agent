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
            "title": "周报：CRM升级项目",
            "subtitle": "2025年第10周（3.3-3.9）",
            "author": "王芳",
            "date": "2025年3月10日",
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
                {"value": "85%", "label": "整体进度", "delta": "↑ 5%", "baseline": "vs 上周80%"},
                {"value": "8/10", "label": "功能点完成", "delta": "↑ 2", "baseline": "vs 上周6/10"},
                {"value": "3个", "label": "Bug修复", "delta": "↓ 2", "baseline": "vs 上周5个"},
                {"value": "正常", "label": "风险等级", "delta": "-", "baseline": "无重大风险"}
            ],
            "notes": "4个核心指标，展示本周期整体情况",
            "filling_prompt": "必须填入真实内容：提供4个本周期的核心指标数据（如'已完成10个功能点'、'进度完成68%'、'Bug修复率95%'），每个有 value（数字）、label（说明）、delta（变化趋势）、baseline（对比基准）。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "已完成事项",
            "content_type": "example_detail",
            "kicker": "实例 · 完成情况",
            "lede": "本周完成2个核心功能，系统原型基本成型",
            "context_block": "本周是项目冲刺阶段，重点推进核心功能开发。经过团队共同努力，完成了本周计划的所有功能点。",
            "solution_block": "本周完成的主要工作包括：1）完成用户管理模块开发，包含增删改查和权限配置；2）完成数据看板页面开发，支持多维度数据展示；3）修复上周遗留的3个Bug，包括登录超时和数据显示错误；4）完成与后台接口的对接测试，接口测试通过率100%。",
            "metrics": [
                {"value": "2个", "label": "功能模块完成", "trend": "vs 计划2个"},
                {"value": "100%", "label": "接口测试通过率", "trend": "全部通过"},
                {"value": "100%", "label": "计划完成率", "trend": "按期完成"}
            ],
            "takeaway": "启示：核心功能基本完成，下周重点推进集成测试",
            "notes": "用具体数字说明完成情况",
            "filling_prompt": "必须填入真实内容：lede 一句话概括本周期的核心成果；context_block 描述工作背景（1-2句话）；solution_block 具体说明完成工作的过程和关键动作（2-3句话）；metrics_grid 提供3个量化指标（如'完成10个功能点'、'进度68%'、'Bug修复率95%'），每个有 value（数字）、label（说明）、trend（对比上周期）；takeaway 用一句话总结成果意义。禁止虚构数据。"
        },
        {
            "index": 5,
            "type": "example_detail",
            "title": "进行中事项",
            "content_type": "example_detail",
            "kicker": "实例 · 进行中事项",
            "lede": "数据导出和报表功能开发中，预计下周一完成",
            "context_block": "本周有两项工作正在进行中，需要持续跟进。",
            "solution_block": "进行中的事项包括：1）数据导出功能开发（完成70%），遇到PDF生成库兼容性问题，正在调试，预计下周一完成；2）报表统计功能开发（完成50%），UI设计需要微调，预计下周三完成。已协调UI设计师周五支持。",
            "metrics": [
                {"value": "70%", "label": "数据导出完成度", "trend": "预计下周一完成"},
                {"value": "50%", "label": "报表功能完成度", "trend": "预计下周三完成"},
                {"value": "2人/天", "label": "剩余工作量", "trend": "资源充足"}
            ],
            "takeaway": "启示：下周一前完成数据导出，下周三完成报表",
            "notes": "说明进行中事项的当前进度和挑战",
            "filling_prompt": "必须填入真实内容：lede 一句话概括当前的工作重点和挑战；context_block 描述当前事项的背景和目标（1-2句话）；solution_block 具体说明当前进展、遇到的困难和解决思路（2-3句话）；metrics_grid 提供3个进度指标（如'完成70%'、'预计下周完成'、'剩余2人/天'），每个有 value、label、trend；takeaway 用一句话总结下一步关键行动。禁止虚构数据。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "关键进展：用户权限模块上线",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "关键进展",
            "header": "用户权限模块完成并上线",
            "sub_header": "支持细粒度权限控制",
            "paragraph": "本周最关键的进展是用户权限模块完成开发并成功上线。该模块支持细粒度的权限控制，可按角色、部门、数据范围等多维度配置权限。上线后，已完成20个用户账号的创建和权限分配，经过3天试运行，系统运行稳定。该模块是整个CRM系统的基础功能，后续所有功能都将基于此权限体系构建。上线过程得到了测试团队的大力支持，在此表示感谢。",
            "references": [
                "内部系统上线记录"
            ],
            "notes": "用图文混排展示关键进展，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少1个URL，如无外部数据需求可填入内部系统URL），再填入真实内容：kicker 为'关键进展'；title 中的 {项目/任务名称} 替换为具体项目或任务名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为进展标题（不超过35字）；sub_header 为进展概述（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述项目的关键进展、突破点和实际成果，用流畅的段落形式呈现，禁止罗列要点。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "问题与风险",
            "content_type": "content_slide",
            "issues": [
                {
                    "issue": "PDF生成库兼容性",
                    "severity": "中等",
                    "description": "PDF生成库与当前框架存在兼容性问题，导致数据导出功能开发延期1天",
                    "solution": "尝试其他PDF库，或使用前端PDF生成方案",
                    "status": "处理中"
                },
                {
                    "issue": "UI设计调整",
                    "severity": "低",
                    "description": "报表页面UI需要微调，可能影响1-2天进度",
                    "solution": "已协调UI设计师周五支持",
                    "status": "已协调"
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
                {"task": "完成数据导出功能", "priority": "P0", "due": "周一"},
                {"task": "完成报表统计功能", "priority": "P0", "due": "周三"},
                {"task": "启动集成测试", "priority": "P1", "due": "周五"},
                {"task": "编写用户手册", "priority": "P2", "due": "周五"}
            ],
            "notes": "下周的工作计划和目标",
            "filling_prompt": "必须填入真实内容：列出下周工作计划（3-5条），每条具体可执行（如'完成订单模块联调测试'、'提交发布申请'）。"
        },
        {
            "index": 9,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 本周完成2个核心功能，用户权限模块上线",
                "02 下周重点：完成数据导出和报表，启动集成测试",
                "03 需要支持：UI设计师周五支持"
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
