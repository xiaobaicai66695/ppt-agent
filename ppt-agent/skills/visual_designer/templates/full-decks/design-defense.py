TEMPLATE = {
    "name": "design-defense",
    "name_cn": "答辩汇报",
    "description": "适合课程设计、毕业设计、项目答辩等场景。逻辑清晰，技术扎实，展示自信。",
    "target_audience": "答辩委员会、项目评审、导师、同学",
    "typical_slides": 12,
    "typical_duration": "15-20分钟（答辩）",
    "palette": "debate_purple",
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
            "title": "{项目/论文名称}",
            "subtitle": "{答辩人姓名} | {指导老师}",
            "author": "{答辩人}",
            "date": "{答辩日期}",
            "notes": "标题页简洁，标题醒目",
            "filling_prompt": "必须填入真实内容：title 为项目或论文名称，subtitle 包含答辩人姓名和指导老师，author 为答辩人，date 为答辩日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  研究背景",
                "02  相关工作",
                "03  设计方案",
                "04  实现与测试",
                "05  总结展望"
            ],
            "notes": "清晰展示答辩结构",
            "filling_prompt": "目录页为固定结构，可根据实际情况调整章节。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "研究背景",
            "subtitle": "问题与动机",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "image_text",
            "title": "研究背景与问题",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "背景 · 问题动机",
            "notes": "右侧配问题场景图/现有方案对比图/技术趋势图，左侧阐述问题和动机",
            "filling_prompt": "必须填入真实内容：图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为核心问题概述（不超过35字）；sub_header 为研究意义说明（不超过35字）；bullets 列出3-4条背景要点，每条不超过35字且优先包含行业数据或技术指标（如'XX领域处理延迟从300ms降至18ms'）。如果需要引用外部数据，通过 web_search 获取并在 references 列出 URL。禁止纯文字堆砌。"
        },
        {
            "index": 5,
            "type": "section_divider",
            "number": "02",
            "title": "相关工作",
            "subtitle": "文献综述与技术基础",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 6,
            "type": "two_column",
            "title": "相关技术与方法",
            "content_type": "two_column",
            "notes": "左右对比相关技术方案",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：left_header 为'现有方案'，left_bullets 列出3-5条现有方案及其特点；right_header 为'本方案优势'，right_bullets 列出3-5条本设计的创新点或优势。references 列出 URL。"
        },
        {
            "index": 7,
            "type": "section_divider",
            "number": "03",
            "title": "设计方案",
            "subtitle": "系统架构与核心设计",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 8,
            "type": "deep_dive",
            "title": "{核心模块}设计",
            "content_type": "deep_dive",
            "kicker": "详解 · {核心模块}",
            "lede": "核心模块的设计思路与实现方案",
            "left_column": {
                "key_points": [
                    "{要点1}",
                    "{要点2}",
                    "{要点3}",
                    "{要点4}",
                    "{要点5}"
                ],
                "analysis": [
                    "{分析维度1}",
                    "{分析维度2}"
                ]
            },
            "right_column": {
                "case_example": [
                    "{案例要素1}",
                    "{案例要素2}",
                    "{案例要素3}",
                    "{案例要素4}"
                ],
                "data_evidence": [
                    "{数据指标1}",
                    "{数据指标2}",
                    "{数据指标3}"
                ]
            },
            "notes": "双栏展示，左栏讲设计要点，右栏放案例数据",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 和 title 中的 {核心模块} 替换为具体模块名称；lede 为一句话说明；left_column.key_points 为设计要点（3-5条，每条不超过35字）；left_column.analysis 为2个分析维度；right_column.case_example 为具体案例（4条，每条不超过35字）；right_column.data_evidence 为3个数据指标。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "04",
            "title": "实现与测试",
            "subtitle": "开发过程与验证结果",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "kpi_dashboard",
            "title": "测试结果",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "数据 · 测试验证",
            "notes": "用数据证明系统效果",
            "filling_prompt": "必须填入真实内容：提供4个测试指标（如功能测试覆盖率、性能指标、用户体验评分、错误率等），每个有 value、label、delta、baseline。禁止虚构数据。"
        },
        {
            "index": 11,
            "type": "two_column",
            "title": "总结与展望",
            "content_type": "two_column",
            "notes": "左右对比已完成工作和未来计划",
            "filling_prompt": "必须填入真实内容：left_header 为'已完成工作'，left_bullets 列出3-5条已完成的主要工作；right_header 为'未来工作'，right_bullets 列出2-3条后续改进方向。"
        },
        {
            "index": 12,
            "type": "summary_slide",
            "title": "答辩完毕",
            "key_points": [
                "01 {核心贡献1}",
                "02 {核心贡献2}",
                "03 {创新亮点}"
            ],
            "thank_you": "感谢各位老师！",
            "notes": "简洁有力的结尾，致谢评委",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心贡献2条+创新亮点1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "答辩要逻辑清晰，重点突出",
        "技术细节要扎实，数据要真实",
        "对自己方案的优缺点要有清晰认识",
        "预判评委问题，准备好应答",
        "PPT页数适中，留出答辩时间"
    ]
}
