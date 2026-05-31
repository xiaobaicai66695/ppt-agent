TEMPLATE = {
    "type": "swot_analysis",
    "name": "SWOT分析页",
    "description": "经典四象限 SWOT 分析布局，左上 S（优势）、右上 W（劣势）、左下 O（机会）、右下 T（威胁）。适合战略分析、项目评估、个人规划等场景。",
    "layout_hint": "2x2 grid",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "30pt",
            "font_weight": "bold",
            "alignment": "left",
            "margin_bottom": "0.2in",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "14pt",
            "color": "text_muted",
            "margin_bottom": "0.25in",
            "max_chars": 40
        },
        "swot": {
            "strengths": {
                "label": "S",
                "label_full": "优势",
                "color": "primary",
                "items": ["{优势1}", "{优势2}", "{优势3}"]
            },
            "weaknesses": {
                "label": "W",
                "label_full": "劣势",
                "color": "secondary",
                "items": ["{劣势1}", "{劣势2}", "{劣势3}"]
            },
            "opportunities": {
                "label": "O",
                "label_full": "机会",
                "color": "accent",
                "items": ["{机会1}", "{机会2}", "{机会3}"]
            },
            "threats": {
                "label": "T",
                "label_full": "威胁",
                "color": "warm",
                "items": ["{威胁1}", "{威胁2}", "{威胁3}"]
            }
        }
    },
    "visual_elements": [
        "2x2 四象限布局，每个象限约 5.5x2.5in",
        "S（优势）左上：primary 色背景，浅色字",
        "W（劣势）右上：secondary 色背景，浅色字",
        "O（机会）左下：accent 色背景，深色字",
        "T（威胁）右下：warm 色背景，浅色字",
        "每个象限顶部有大号字母标签 S/W/O/T（72pt）",
        "字母下方有全称（优势/劣势/机会/威胁，14pt）",
        "象限内用 bullet 列表展示要点（每项最多8个字）",
        "四象限之间用分隔线（background 色，1pt）",
        "整体标题区在四象限上方"
    ],
    "example": {
        "kicker": "战略分析",
        "title": "AI产品战略SWOT分析",
        "subtitle": "基于市场与竞争格局的全面评估",
        "swot": {
            "strengths": {
                "items": [
                    "技术领先，算法准确率高",
                    "团队经验丰富，研发效率高",
                    "数据积累深厚，模型效果好",
                    "客户口碑好，复购率高"
                ]
            },
            "weaknesses": {
                "items": [
                    "成本高，定价缺乏竞争力",
                    "品牌知名度不足",
                    "销售渠道单一",
                    "国际化能力弱"
                ]
            },
            "opportunities": {
                "items": [
                    "市场需求快速增长",
                    "政策支持AI发展",
                    "行业标准尚未成熟",
                    "潜在合作伙伴多"
                ]
            },
            "threats": {
                "items": [
                    "大厂入局，竞争加剧",
                    "技术迭代快，研发压力大",
                    "数据合规要求趋严",
                    "经济下行影响IT预算"
                ]
            }
        }
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{SWOT分析标题}",
        "subtitle": "{副标题}",
        "swot": {
            "strengths": {"items": ["{优势1}", "{优势2}", "{优势3}"]},
            "weaknesses": {"items": ["{劣势1}", "{劣势2}", "{劣势3}"]},
            "opportunities": {"items": ["{机会1}", "{机会2}", "{机会3}"]},
            "threats": {"items": ["{威胁1}", "{威胁2}", "{威胁3}"]}
        }
    },
    "when_to_use": [
        "战略规划与商业分析",
        "项目可行性评估",
        "个人职业规划",
        "竞争分析"
    ],
    "never": [
        "禁止单个象限超过5个要点",
        "禁止四个象限颜色混乱",
        "禁止象限之间没有视觉分隔",
        "禁止要点文字过长（每项不超过12字）",
        "禁止缺少标题或标签"
    ]
}
