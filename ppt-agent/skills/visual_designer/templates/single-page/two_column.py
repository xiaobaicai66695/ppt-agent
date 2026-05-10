TEMPLATE = {
    "type": "two_column",
    "name": "双栏对比",
    "description": "常见于 A vs B 分析，左右并置。可用于方案对比、功能对比、观点对比等场景。支持三种内容模式：①纯要点模式（bullets）；②带引言模式（intro + bullets）；③多区块模式（sections，分核心要点/深度分析/数据支撑等子区块）。",
    "layout_hint": "split",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "36pt",
            "font_weight": "bold",
            "alignment": "left",
            "max_chars": 30
        },
        "left_column": {
            "position": "left_half",
            "width": "45%",
            "header": {
                "font_size": "18pt",
                "font_weight": "bold",
                "color": "primary"
            },
            "header_bg": "light_bg",
            # 内容模式一：纯要点模式
            "bullets": {
                "font_size": "13pt",
                "items_max": 6,
                "char_per_item_max": 35,
                "bullet_style": "实心点"
            },
            # 内容模式二：开篇引言 + 要点
            "intro": {
                "font_size": "11pt",
                "color": "text",
                "max_chars": 80,
                "margin_bottom": "0.15in"
            },
            # 内容模式三：多区块结构
            "sections": {
                "key_points": {
                    "label": "核心要点",
                    "font_size": "12pt",
                    "max_items": 5,
                    "char_per_item_max": 35
                },
                "analysis": {
                    "label": "深度分析",
                    "font_size": "11pt",
                    "max_items": 3,
                    "char_per_item_max": 35
                },
                "data": {
                    "label": "数据支撑",
                    "font_size": "11pt",
                    "max_items": 3,
                    "char_per_item_max": 30
                },
                "quote": {
                    "label": "观点引用",
                    "font_size": "11pt",
                    "max_items": 2,
                    "char_per_item_max": 35
                }
            },
            "highlight_box": {
                "enabled": True,
                "label": "推荐",
                "color": "secondary"
            }
        },
        "right_column": {
            "position": "right_half",
            "width": "45%",
            "header": {
                "font_size": "18pt",
                "font_weight": "bold",
                "color": "accent"
            },
            "header_bg": "light_bg",
            "bullets": {
                "font_size": "13pt",
                "items_max": 6,
                "char_per_item_max": 35,
                "bullet_style": "实心点"
            },
            "intro": {
                "font_size": "11pt",
                "color": "text",
                "max_chars": 80,
                "margin_bottom": "0.15in"
            },
            "sections": {
                "key_points": {
                    "label": "核心要点",
                    "font_size": "12pt",
                    "max_items": 5,
                    "char_per_item_max": 35
                },
                "analysis": {
                    "label": "深度分析",
                    "font_size": "11pt",
                    "max_items": 3,
                    "char_per_item_max": 35
                },
                "data": {
                    "label": "数据支撑",
                    "font_size": "11pt",
                    "max_items": 3,
                    "char_per_item_max": 30
                },
                "quote": {
                    "label": "观点引用",
                    "font_size": "11pt",
                    "max_items": 2,
                    "char_per_item_max": 35
                }
            }
        },
        "divider": "两栏之间可用细线（secondary 色，1pt）或留白分隔（0.3in）"
    },
    "visual_elements": [
        "两栏顶部用色块条作为栏目标题背景（left: primary 色淡底，right: accent 色淡底）",
        "栏目标题下方用 2pt 强调线（与标题同色）",
        "可选：在某一栏右上角添加「推荐」标签徽章（secondary 色背景，白色文字）",
        "两栏背景均可使用 light_bg 浅底色，栏间留白 0.3in",
        "底部可添加小型装饰几何图形（primary 色，透明度 10%）",
        "如果对比的是时间维度，可在顶部添加小图标（时钟、箭头等文字描述）"
    ],
    # 模式一：纯要点模式（简单对比场景）
    "example": {
        "kicker": "方案对比",
        "title": "技术方案选型",
        "left_header": "传统方案",
        "left_bullets": [
            "部署周期 5-7 天",
            "运维人力 10 人",
            "故障恢复 > 2 小时",
            "扩容需提前申请"
        ],
        "left_highlight": "推荐",
        "right_header": "云原生方案",
        "right_bullets": [
            "部署周期 < 1 天",
            "运维人力 2 人",
            "故障恢复 < 5 分钟",
            "弹性扩容秒级响应"
        ]
    },
    # 模式二：带引言模式（需要背景说明的场景）
    "example_2": {
        "kicker": "自我剖析",
        "title": "反思与改进",
        "left_header": "不足与反思",
        "left_intro": "在过去一年的工作中，我在多个方面进行了深入反思，识别出以下需要改进的领域：",
        "left_bullets": [
            "技术深度不足：与研发沟通时理解不够深入，影响需求评审质量",
            "跨部门协调能力有待提升：复杂项目推进中存在明显短板",
            "文档规范性不足：部分需求文档细节描述不够完善"
        ],
        "right_header": "改进计划",
        "right_intro": "针对上述不足，我制定了明确的改进计划，并已开始执行：",
        "right_bullets": [
            "系统学习后端技术课程，已报名内部训练营，下季度完成3门课程",
            "向有经验的同事请教跨部门协作技巧，建立高效沟通话术模板",
            "建立需求文档Checklist，涵盖边界条件、异常处理等要素"
        ]
    },
    # 模式三：多区块模式（需要深度分析的对比场景）
    "example_3": {
        "kicker": "方案对比 · 技术选型",
        "title": "现有方案分析",
        "left_header": "传统方案局限性",
        "left_sections": {
            "analysis": [
                "部署效率：手动部署依赖人工操作，环境差异导致问题频发，部署周期长达5-7天",
                "资源利用：按峰值预留计算资源，平日利用率仅15-25%，造成大量资源浪费",
                "可用性保障：故障恢复依赖人工介入，平均恢复时间超过2小时，SLA难以保障"
            ],
            "data": [
                "部署周期：5-7天/次",
                "资源利用率：15-25%",
                "故障恢复时间：>2小时"
            ]
        },
        "right_header": "改进方向",
        "right_sections": {
            "key_points": [
                "声明式配置：用代码管理基础设施，环境一致性问题从源头解决，版本化管理支持回滚",
                "自动化编排：减少人工干预，CI/CD流水线实现一键部署，部署时间从数天缩短至分钟级",
                "弹性伸缩：根据负载自动调整资源，利用率提升至60%以上，峰值自动扩容、谷值自动缩容"
            ],
            "data": [
                "部署周期：<1小时",
                "资源利用率：60%+",
                "故障恢复时间：<5分钟"
            ]
        }
    },
    "when_to_use": [
        "A vs B 类对比分析",
        "两个方案/产品/观点并列呈现",
        "需要左右形成对照（如传统vs改进、不足vs计划、现状vs目标）",
        "对比需要深度分析或数据支撑时使用 sections 模式"
    ],
    "never": [
        "禁止两栏内容完全不对等（一边很多一边很少）",
        "禁止不用色块区分两栏，导致视觉模糊",
        "禁止每栏超过6条要点（过多时应使用多区块模式分层展示）",
        "禁止使用匿名实体（如'某公司''某系统'）—— 必须给出真实名称",
        "禁止虚构数据—— 必须通过 web_search 获取真实数据"
    ]
}
