TEMPLATE = {
    "type": "detail_expand",
    "name": "双栏详解页",
    "description": "双栏布局深入展开核心内容。左栏放要点阐述，右栏放案例/数据/补充信息。适用于讲解核心概念、关键机制、专题分析等场景。",
    "layout_hint": "two_column",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "font_size": "30pt",
            "font_weight": "bold",
            "alignment": "left",
            "format": "{主题}: {一句话概括核心内容}",
            "margin_bottom": "0.15in",
            "max_chars": 50
        },
        "lede": {
            "font_size": "16pt",
            "color": "text_muted",
            "max_chars": 60,
            "margin_bottom": "0.2in"
        },
        "left_column": {
            "width_ratio": "48%",
            "header": {
                "font_size_header": "18pt",
                "color": "primary"
            },
            "sections": {
                "key_points": {
                    "label": "核心要点",
                    "font_size": "14pt",
                    "max_items": 5,
                    "item_format": "{要点标题}：{说明}",
                    "example": [
                        "明确目标：设定清晰、可衡量的学习目标",
                        "循序渐进：从基础概念到实际应用逐步深入",
                        "实践结合：理论学习与动手实践相结合"
                    ]
                },
                "analysis": {
                    "label": "深度分析",
                    "font_size": "13pt",
                    "max_items": 3,
                    "item_format": "{分析维度}：{分析结论}",
                    "example": [
                        "优势：该方案能够有效解决核心问题",
                        "挑战：实施过程中可能面临资源约束"
                    ]
                }
            }
        },
        "right_column": {
            "width_ratio": "48%",
            "header": {
                "font_size_header": "18pt",
                "color": "accent"
            },
            "sections": {
                "case_example": {
                    "label": "案例说明",
                    "font_size": "13pt",
                    "max_items": 4,
                    "item_format": "{案例要素}",
                    "example": [
                        "案例：某企业实施数字化转型项目",
                        "背景：传统业务模式增长乏力",
                        "措施：引入智能管理系统",
                        "效果：运营效率提升 40%"
                    ]
                },
                "data_evidence": {
                    "label": "数据支撑",
                    "font_size": "13pt",
                    "max_items": 3,
                    "item_format": "{指标}：{数值}",
                    "example": [
                        "用户满意度：92%（↑15%）",
                        "处理效率：提升 3 倍",
                        "成本降低：约 30 万元/年"
                    ]
                },
                "supplement": {
                    "label": "补充信息",
                    "font_size": "13pt",
                    "max_items": 2,
                    "item_format": "{信息类型}：{内容}",
                    "example": [
                        "注意事项：需注意数据安全合规",
                        "适用条件：适用于中等规模团队"
                    ]
                }
            }
        },
        "references": {
            "label": "参考来源",
            "position": "bottom",
            "font_size": "10pt",
            "color": "text_muted",
            "data_source": "内容中的数据、案例、事实必须通过 web_search 获取真实来源；禁止虚构。"
        }
    },
    "visual_elements": [
        "左栏顶部 primary 色强调线（3pt），右栏顶部 accent 色强调线",
        "两栏之间有 0.3in 间隔",
        "各 section 的 label 使用浅底色背景标签",
        "steps/items 中每条前有数字序号（primary 色）",
        "data_evidence 中「↑」用绿色，「↓」用红色标注",
        "references 区域位于页面最底部，用细线与上方内容分隔"
    ],
    "example": {
        "kicker": "管理方法",
        "title": "OKR 管理法：让目标与结果紧密关联",
        "lede": "OKR 帮助团队明确目标、聚焦重点、对齐方向",
        "left_column": {
            "header": "核心要点",
            "key_points": [
                "目标设定：设定有挑战性、可衡量的目标",
                "关键结果：定义可量化的关键结果指标",
                "周期复盘：定期回顾 OKR 完成情况",
                "公开透明：全员可见，促进协作对齐"
            ],
            "analysis": [
                "优势：聚焦重点，避免资源分散",
                "挑战：需要全员理解和持续跟进"
            ]
        },
        "right_column": {
            "header": "实践案例",
            "case_example": [
                "案例：某互联网公司引入 OKR 管理",
                "背景：团队扩张后协调效率下降",
                "措施：全公司推行 OKR 季度管理",
                "效果：项目交付率提升 25%"
            ],
            "data_evidence": [
                "目标达成率：平均 70%",
                "员工满意度：提升 18%",
                "跨团队协作：效率提升 30%"
            ]
        },
        "references": [
            {"title": "OKR 起源与发展", "url": "https://en.wikipedia.org/wiki/OKR"},
            {"title": "Google OKR 实践", "url": "https://rework.withgoogle.com/guides/set-goals-with-okrs/"}
        ]
    },
    "example_2": {
        "kicker": "技术方案",
        "title": "{技术名称}: {一句话概括}",
        "lede": "{一句话说明该技术的核心价值}",
        "left_column": {
            "header": "核心要点",
            "key_points": [
                "1. {要点1}",
                "2. {要点2}",
                "3. {要点3}",
                "4. {要点4}"
            ],
            "analysis": [
                "{分析维度1}",
                "{分析维度2}"
            ]
        },
        "right_column": {
            "header": "案例/数据",
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
        "references": [
            {"title": "{参考标题1}", "url": "{web_search获取的URL1}"},
            {"title": "{参考标题2}", "url": "{web_search获取的URL2}"}
        ]
    },
    "when_to_use": [
        "深入讲解一个核心概念/方法/机制时",
        "需要左侧理论+右侧案例配合时",
        "展示有数据支撑的分析内容时",
        "通用场景均适用，不限于技术主题"
    ],
    "never": [
        "禁止左栏要点少于 3 条",
        "禁止两侧内容完全相同——左右要有区分",
        "禁止虚构数据——必须通过 web_search 获取真实数据",
        "禁止案例使用匿名实体（如'某公司'）——必须给出真实名称",
        "禁止 references 为空——必须列出 web_search 获取的参考资料"
    ]
}
