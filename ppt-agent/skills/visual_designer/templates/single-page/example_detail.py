TEMPLATE = {
    "type": "example_detail",
    "name": "实例详解页",
    "description": "通过一个命名、有数据的真实案例来解释概念。每介绍一个重要概念或技术时必须配此页型。",
    "layout_hint": "left",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "实例 · {领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "font_size": "30pt",
            "font_weight": "bold",
            "alignment": "left",
            "format": "{案例名称}: {一句话总结}",
            "margin_bottom": "0.15in",
            "max_chars": 40
        },
        "lede": {
            "font_size": "16pt",
            "color": "text_muted",
            "max_chars": 60,
            "margin_bottom": "0.25in"
        },
        "context_block": {
            "label": "背景",
            "label_color": "secondary",
            "label_bg": "light_bg",
            "font_size": "14pt",
            "max_lines": 2,
            "content_placeholder": "{行业背景或问题语境}"
        },
        "solution_block": {
            "label": "方案",
            "label_color": "primary",
            "label_bg": "light_bg",
            "font_size": "14pt",
            "max_lines": 3,
            "content_placeholder": "{技术方案描述，含具体技术栈名称（如 深度学习+图计算、Transformer+RLHF）}"
        },
        "metrics_grid": {
            "layout": "3-4 card grid",
            "card_spec": {
                "value_font_size": "36pt",
                "label_font_size": "11pt",
                "trend_font_size": "11pt",
                "card_bg": "light_bg"
            },
            "example": [
                {"value": "99.99%", "label": "欺诈识别准确率", "trend": "↑ 0.3%"},
                {"value": "50万", "label": "日均处理峰值(TPS)", "trend": "↑ 120%"},
                {"value": "100亿+", "label": "年减损金额(元)", "trend": "↓ 60%"}
            ]
        },
        "takeaway": {
            "font_size": "14pt",
            "font_weight": "bold",
            "color": "primary",
            "max_chars": 80,
            "format": "启示：{1句总结，点出可迁移的经验}",
            "icon": "💡"
        },
        "references": {
            "label": "参考来源",
            "position": "bottom",
            "font_size": "10pt",
            "color": "text_muted",
            "url_label": "参考资料：",
            "max_urls": 3,
            "format": "序号. 标题 | URL",
            "data_source": "内容中的数据、案例、指标必须来自 web_search 返回的真实 URL 页面；禁止虚构数据。Agent 在生成页面前必须先 web_search 获取权威来源，并在 references 区域列出所有使用到的 URL。"
        }
    },
    "visual_elements": [
        "左侧装饰竖线（primary 色，3pt）贯穿标题区域",
        "context_block 和 solution_block 使用浅底色卡片（light_bg），每个区域顶部有对应颜色的强调线",
        "context_block 顶部强调线用 secondary 色",
        "solution_block 顶部强调线用 primary 色",
        "metrics_grid 区域使用浅底色卡片（light_bg），每卡片顶部色条（primary 色 3pt）",
        "takeaway 区域可用浅底色背景衬托，左侧添加💡或类似图标（用文字「启示」代替 emoji）",
        "references 区域位于页面最底部，用细线（secondary 色，0.5pt）与上方内容分隔",
        "references 中的 URL 以链接文本形式呈现，hover 时显示完整 URL",
        "整体布局上下留白充足，内容紧凑但不拥挤"
    ],
    "example": {
        "kicker": "实例 · 金融风控",
        "title": "蚂蚁金服智能风控: 实时识别欺诈交易",
        "lede": "年交易额超万亿，欺诈损失率从 0.8% 降至 0.02%",
        "context_block": "传统规则引擎误报率高（15%），人工审核人力成本巨大，且无法识别新型欺诈模式",
        "solution_block": "基于深度学习构建实时风控模型，融合设备指纹、行为序列、关系图谱多维特征。采用流计算引擎实现 50ms 内响应，覆盖支付全链路",
        "metrics_grid": [
            {"value": "99.99%", "label": "欺诈识别准确率", "trend": "↑ 0.3%"},
            {"value": "50ms", "label": "平均响应时间", "trend": "↓ 80%"},
            {"value": "¥100亿+", "label": "年减损金额", "trend": "↓ 60%"}
        ],
        "takeaway": "启示：多模态特征融合+实时流计算是风控系统的核心竞争力",
        "references": [
            {"title": "蚂蚁金服智能风控技术白皮书", "url": "https://tech.antgroup.com/article/21000"},
            {"title": "深度学习在金融风控领域的应用实践", "url": "https://arxiv.org/abs/2103.00000"},
            {"title": "流计算引擎在实时风控中的实践", "url": "https://flink.apache.org/case-studies"}
        ]
    },
    "example_2": {
        "kicker": "实例 · {领域标签}",
        "title": "{案例名称}: {一句话总结}",
        "lede": "{一句话说明该案例的核心效果}",
        "context_block": "{行业背景或问题语境}",
        "solution_block": "{技术方案描述}",
        "metrics_grid": [
            {"value": "{指标值}", "label": "{指标名称}", "trend": "{趋势}"},
            {"value": "{指标值}", "label": "{指标名称}", "trend": "{趋势}"},
            {"value": "{指标值}", "label": "{指标名称}", "trend": "{趋势}"}
        ],
        "takeaway": "启示：{1句总结，点出可迁移的经验}",
        "references": [
            {"title": "{参考标题1}", "url": "{web_search获取的URL1}"},
            {"title": "{参考标题2}", "url": "{web_search获取的URL2}"}
        ]
    },
    "when_to_use": [
        "介绍一个技术概念/方案后，必须跟一页实例",
        "展示产品/技术在实际场景中的落地效果",
        "需要用具体数据说服听众时"
    ],
    "never": [
        "禁止使用'某公司''某系统'等匿名实体——必须给出真实名称",
        "禁止 metrics 少于 2 个具体数字",
        "禁止 solution_block 只有一句话——必须展开写 2-3 句",
        "禁止缺少 takeaway——每个实例必须有可迁移的启示",
        "禁止 references 为空或使用无法访问的 URL——必须通过 web_search 获取真实有效的参考资料",
        "禁止在 metrics 或 solution_block 中虚构数据——所有数据必须来源于 references 中的 URL"
    ]
}
