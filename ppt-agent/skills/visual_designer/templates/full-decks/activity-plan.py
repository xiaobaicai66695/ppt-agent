TEMPLATE = {
    "name": "activity-plan",
    "name_cn": "活动策划",
    "description": "适合团建活动、校园活动、节日策划等场景。活泼有创意，执行清晰，预算可控。",
    "target_audience": "领导、团队成员、活动参与者",
    "typical_slides": 10,
    "typical_duration": "10-15分钟（策划汇报）",
    "palette": "activity_orange",
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
            "title": "{活动名称}",
            "subtitle": "{活动主题口号}",
            "author": "{策划团队}",
            "date": "{计划日期}",
            "notes": "标题页有活力，吸引眼球",
            "filling_prompt": "必须填入真实内容：title 为活动名称，subtitle 为一句概括性口号，author 为策划团队，date 为计划举办日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  活动概述",
                "02  活动详情",
                "03  人员分工",
                "04  预算安排",
                "05  风险预案"
            ],
            "notes": "让观众了解策划框架",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "活动概述",
            "subtitle": "背景与目标",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "活动背景与目标",
            "content_type": "content_slide",
            "notes": "说明为什么举办这个活动",
            "filling_prompt": "必须填入真实内容：说明活动背景（1-2段话）、活动目标（2-3条 SMART 目标：具体、可衡量、可实现、相关、有时限）。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "活动详情",
            "subtitle": "时间地点与流程",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "timeline",
            "title": "活动流程",
            "content_type": "timeline",
            "layout_hint": "horizontal",
            "notes": "展示活动当天的流程安排",
            "filling_prompt": "必须填入真实内容：提供活动当天的流程安排，列出4-6个关键时间节点，每个有时间+环节名称+持续时长+负责人。"
        },
        {
            "index": 7,
            "type": "image_text",
            "title": "活动亮点",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "亮点 · 核心创意",
            "notes": "右侧配往期活动照片/场地效果图/创意展示图，左侧列亮点",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为最具特色的亮点名称（不超过35字）；sub_header 为该亮点的预期效果（不超过35字）；bullets 列出3-4条活动亮点或特色环节，每条不超过35字且描述具体可感知的效果（如'沉浸式VR体验，预计参与率超90%'）。如果参考了外部案例，通过 web_search 获取相关资料并在 references 列出 URL。"
        },
        {
            "index": 8,
            "type": "section_divider",
            "number": "03",
            "title": "人员分工",
            "subtitle": "组织架构与职责",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 9,
            "type": "content_slide",
            "title": "组织架构与分工",
            "content_type": "content_slide",
            "notes": "展示团队分工和职责",
            "filling_prompt": "必须填入真实内容：说明活动组织架构，列出各小组（如策划组、执行组、后勤组、宣传组等）及每个小组的职责和负责人。"
        },
        {
            "index": 10,
            "type": "section_divider",
            "number": "04",
            "title": "预算安排",
            "subtitle": "费用预算与资源",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 11,
            "type": "kpi_dashboard",
            "title": "预算概览",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "预算 · 费用明细",
            "notes": "展示预算分配",
            "filling_prompt": "必须填入真实内容：提供4个预算类别（如场地费用、餐饮费用、物料费用、人员费用等），每个有 value（金额）、label（类别名称）、delta（占比）、baseline（总预算）。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "05",
            "title": "风险预案",
            "subtitle": "预案与措施",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "content_slide",
            "title": "风险识别与预案",
            "content_type": "content_slide",
            "notes": "识别风险并准备预案",
            "filling_prompt": "必须填入真实内容：列出3-5个可能的风险（如天气变化、人员缺席、设备故障等），每个说明风险描述、影响程度、预防措施和应急预案。"
        },
        {
            "index": 14,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {活动核心理念}",
                "02 {核心亮点}",
                "03 {预期效果}"
            ],
            "thank_you": "期待您的支持！",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心理念+核心亮点+预期效果）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "活动策划要有创意，亮点突出",
        "执行计划要具体，时间节点清晰",
        "预算要合理，有明细有依据",
        "风险预案要全面，有备无患",
        "PPT设计可以活泼一些，体现活动氛围"
    ]
}
