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
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "{周/月}报：{项目/团队名称}",
            "subtitle": "{时间段}",
            "author": "{汇报人}",
            "date": "{日期}",
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
                {"value": "{数值}", "label": "{指标说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{指标说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{指标说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"},
                {"value": "{数值}", "label": "{指标说明}", "delta": "{变化趋势}", "baseline": "{对比基准}"}
            ],
            "notes": "4个核心指标，展示本周期整体情况",
            "filling_prompt": "必须填入真实内容：提供4个本周期的核心指标数据（如'已完成10个功能点'、'进度完成68%'、'Bug修复率95%'），每个有 value（数字）、label（说明）、delta（变化趋势）、baseline（对比基准）。"
        },
        {
            "index": 4,
            "type": "content_slide",
            "title": "已完成事项",
            "content_type": "content_slide",
            "notes": "用 checklist 或已打勾的方式列出已完成事项",
            "filling_prompt": "必须填入真实内容：列出本周期的已完成事项（3-5条），每条说明具体工作内容（如'完成用户权限模块开发'、'上线A/B测试功能'）。"
        },
        {
            "index": 5,
            "type": "content_slide",
            "title": "进行中事项",
            "content_type": "content_slide",
            "notes": "当前进行中的工作及进度",
            "filling_prompt": "必须填入真实内容：列出当前进行中的事项（2-4条），每条说明工作内容和当前进度（如'订单系统重构：完成70%，预计下周完成'）。"
        },
        {
            "index": 6,
            "type": "image_text",
            "title": "关键进展：{项目/任务名称}",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "notes": "用图文混排展示关键进展，增强可信性",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少1个URL，如无外部数据需求可填入内部系统URL），再填入真实内容：kicker 为'关键进展'；title 中的 {项目/任务名称} 替换为具体项目或任务名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为进展标题（不超过35字）；sub_header 为进展概述（不超过35字）；bullets 列出3-4条关键成果或突破点，每条不超过35字。references 列出 URL。禁止虚构数据。"
        },
        {
            "index": 7,
            "type": "content_slide",
            "title": "问题与风险",
            "content_type": "content_slide",
            "notes": "列出当前面临的问题和风险，以及应对措施",
            "filling_prompt": "必须填入真实内容：列出1-3个当前面临的问题/风险，每条说明问题描述和应对措施。坦诚汇报，不回避问题。"
        },
        {
            "index": 8,
            "type": "content_slide",
            "title": "下周计划",
            "content_type": "content_slide",
            "notes": "下周的工作计划和目标",
            "filling_prompt": "必须填入真实内容：列出下周工作计划（3-5条），每条具体可执行（如'完成订单模块联调测试'、'提交发布申请'）。"
        },
        {
            "index": 9,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 {核心成果1}",
                "02 {核心成果2}",
                "03 {下周重点}"
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
        "计划要具体可执行"
    ]
}
