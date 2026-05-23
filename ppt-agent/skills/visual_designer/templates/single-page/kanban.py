TEMPLATE = {
    "type": "kanban",
    "name": "看板进度页",
    "description": "三列看板布局，适合展示项目进度、任务状态或流程阶段。列：待办/进行中/已完成，配合优先级色条和进度条。",
    "layout_hint": "three_column",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "32pt",
            "font_weight": "bold",
            "alignment": "left",
            "margin_bottom": "0.2in",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "14pt",
            "color": "text_muted",
            "margin_bottom": "0.3in",
            "max_chars": 40
        },
        "columns": {
            "count": 3,
            "width": "flexible (equal)",
            "headers": [
                {"title": "待办事项", "color": "text_muted"},
                {"title": "进行中", "color": "secondary"},
                {"title": "已完成", "color": "primary"}
            ]
        },
        "cards": {
            "card_bg": "background",
            "card_padding": "0.15in",
            "priority_indicator": {
                "high": "primary",
                "medium": "secondary",
                "low": "accent",
                "done": "divider",
                "position": "left edge",
                "width": "0.06in"
            },
            "tag": {
                "font_size": "9pt",
                "bg": "light_bg",
                "color": "text_muted",
                "position": "top-left"
            },
            "card_title": {
                "font_size": "12pt",
                "font_weight": "bold",
                "color": "text"
            }
        },
        "footer": {
            "progress_bar": True,
            "show_stats": "整体进度：X% | 待办 N 项 | 进行中 N 项 | 已完成 N 项"
        }
    },
    "visual_elements": [
        "三列等宽布局，间距 0.25in",
        "列标题栏用对应颜色填充（待办:text_muted，进行中:secondary，已完成:primary）",
        "列标题右侧显示数量徽章（圆角矩形，background色）",
        "每张卡片用 background 色背景，左侧有优先级色条（3pt宽）",
        "卡片内标签用圆角矩形小标签（light_bg底）",
        "底部进度条：总进度用 light_bg 底色，已完成部分用 primary 色填充",
        "底部统计文字居中显示（text_muted，12pt）",
        "卡片之间垂直间距 0.15in"
    ],
    "example": {
        "kicker": "项目管理",
        "title": "项目进度看板",
        "subtitle": "敏捷开发任务跟踪与可视化",
        "columns": [
            {
                "title": "待办事项",
                "color": "text_muted",
                "cards": [
                    {"text": "用户调研报告", "tag": "需求", "priority": "high"},
                    {"text": "竞品分析文档", "tag": "分析", "priority": "medium"},
                    {"text": "技术方案设计", "tag": "设计", "priority": "low"}
                ]
            },
            {
                "title": "进行中",
                "color": "secondary",
                "cards": [
                    {"text": "前端界面开发", "tag": "开发", "priority": "high"},
                    {"text": "API接口联调", "tag": "开发", "priority": "medium"}
                ]
            },
            {
                "title": "已完成",
                "color": "primary",
                "cards": [
                    {"text": "项目立项审批", "tag": "管理", "priority": "done"},
                    {"text": "团队组建", "tag": "管理", "priority": "done"},
                    {"text": "需求评审会", "tag": "需求", "priority": "done"}
                ]
            }
        ],
        "progress": 65,
        "stats": "待办 3 项 | 进行中 2 项 | 已完成 3 项"
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{看板标题}",
        "subtitle": "{说明}",
        "columns": [
            {
                "title": "{列1名称}",
                "color": "{text_muted}",
                "cards": [
                    {"text": "{任务1}", "tag": "{标签}", "priority": "{high/medium/low}"},
                    {"text": "{任务2}", "tag": "{标签}", "priority": "{high/medium/low}"}
                ]
            },
            {
                "title": "{列2名称}",
                "color": "{secondary}",
                "cards": [
                    {"text": "{任务3}", "tag": "{标签}", "priority": "{high/medium/low}"}
                ]
            },
            {
                "title": "{列3名称}",
                "color": "{primary}",
                "cards": [
                    {"text": "{任务4}", "tag": "{标签}", "priority": "done"}
                ]
            }
        ],
        "progress": {百分比},
        "stats": "待办 N 项 | 进行中 N 项 | 已完成 N 项"
    },
    "when_to_use": [
        "项目进度展示",
        "任务状态跟踪",
        "流程阶段可视化",
        " sprint 回顾"
    ],
    "never": [
        "禁止超过3列",
        "禁止卡片内容过长（每卡标题不超过20字）",
        "禁止列之间间距过大或过小",
        "禁止缺少进度统计"
    ]
}
