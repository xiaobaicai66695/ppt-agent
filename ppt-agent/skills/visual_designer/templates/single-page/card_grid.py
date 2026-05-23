TEMPLATE = {
    "type": "card_grid",
    "name": "卡片阵列",
    "description": "适合展示4~8个同等重要的事项（如功能特性、产品优势、案例展示）。卡片尺寸自适应。",
    "layout_hint": "grid",
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
        "subtitle": {
            "font_size": "16pt",
            "color": "text_muted",
            "margin_bottom": "0.25in",
            "max_chars": 40
        },
        "cards": {
            "layout": "2x2 或 2x3 或 3x2 或 1x4（根据内容数量选择）",
            "card_width": "flexible",
            "card_padding": "0.15in",
            "card_header": {
                "font_size": "16pt",
                "font_weight": "bold",
                "color": "primary",
                "icon": "可选：数字序号或小图标文字描述"
            },
            "card_body": {
                "font_size": "14pt",
                "color": "text",
                "char_per_item_max": 115,
                "line_spacing": "1.4x",
                "description": "每张卡片描述 100-120 字，展开说明要点细节、具体数据或应用场景，禁止过短或抽象。"
            },
            "card_footer": {
                "font_size": "12pt",
                "color": "secondary",
                "content": "可选的补充说明或趋势标签（如 ↑ 30%）"
            }
        }
    },
    "visual_elements": [
        "每个卡片有浅色背景（light_bg 色）和左侧优先级/重要度色条（3pt 宽）",
        "卡片顶部数字序号用圆形背景（primary 色，白色数字）",
        "卡片之间间距 0.2-0.3 英寸",
        "卡片左侧色条可用 secondary/accent 色表示不同类型",
        "如果卡片数量为奇数，空位用浅色占位卡片保持平衡",
        "可选：每张卡片右上角添加状态徽章（如「新」「推荐」「Beta」）",
        "可选：底部添加趋势标签（secondary 色，如 ↑ 30%）",
        "整体布局上方可有装饰性小圆点阵列（accent 色，透明度 10%）"
    ],
    "example": {
        "kicker": "产品能力",
        "title": "产品核心能力",
        "subtitle": "全方位赋能企业数字化转型",
        "layout": "2x2",
        "cards": [
            {
                "header": "智能问答",
                "icon": "01",
                "body": "基于大模型的自然语言交互系统，支持多轮对话、上下文记忆与精准意图识别，已服务超过千家企业的日常问答需求，显著提升客服效率。",
                "footer": "↑ 3倍 效率提升"
            },
            {
                "header": "多模态理解",
                "icon": "02",
                "body": "支持文本、图像、音视频的统一理解与跨模态信息融合处理，适用于智能客服、内容审核与知识图谱构建等多种复杂业务场景。",
                "footer": "全模态覆盖"
            },
            {
                "header": "知识推理",
                "icon": "03",
                "body": "复杂逻辑推理与知识关联网络构建能力，支持多跳问答与因果分析，在金融风控、医疗诊断等领域已验证显著的准确率提升效果。",
                "footer": "准确率 99.9%"
            },
            {
                "header": "个性化定制",
                "icon": "04",
                "body": "企业私有知识库无缝接入，支持文档解析、向量检索与增量持续学习，灵活满足不同行业客户的差异化需求与数据安全要求。",
                "footer": "私有化部署"
            }
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{主题名称}",
        "subtitle": "{副标题}",
        "layout": "2x2",
        "cards": [
            {"header": "{卡片1标题}", "body": "{卡片1描述：100-120字，展开细节与具体数据}"},
            {"header": "{卡片2标题}", "body": "{卡片2描述：100-120字，展开细节与具体数据}"},
            {"header": "{卡片3标题}", "body": "{卡片3描述：100-120字，展开细节与具体数据}"},
            {"header": "{卡片4标题}", "body": "{卡片4描述：100-120字，展开细节与具体数据}"}
        ]
    },
    "when_to_use": [
        "多个（4-8个）同等重要性的事项",
        "功能特性展示",
        "产品优势清单"
    ],
    "never": [
        "禁止卡片内容差距过大（有的很多字有的很少）",
        "禁止超过8个卡片",
        "禁止卡片之间没有间距",
        "禁止卡片描述少于100字或超过120字——必须展开细节，禁止抽象概括"
    ]
}
