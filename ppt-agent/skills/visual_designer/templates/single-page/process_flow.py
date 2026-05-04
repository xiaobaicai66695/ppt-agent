TEMPLATE = {
    "type": "process_flow",
    "name": "步骤流程图",
    "description": "箭头/连线 + 步骤框，适合3~6步的操作流程、工作流、数据处理流水线等。",
    "layout_hint": "horizontal 或 vertical",
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
        "steps": [
            {
                "step_number": "序号（01/02/03）",
                "step_title": "步骤标题（16pt 加粗）",
                "step_desc": "步骤说明（14pt，不超过30字）"
            }
        ],
        "arrows": {
            "type": "单向箭头 或 双向箭头 或 虚线箭头（表示可选路径）",
            "color": "secondary",
            "width": "1.5pt",
            "arrow_head": "三角形填充箭头"
        },
        "steps_max": 6
    },
    "visual_elements": [
        "步骤框用 rounded rectangle（radius 0.08in），背景 light_bg 色，边框 primary 色（2pt）",
        "序号用大号字体（24pt）放在框内左上角或左侧（primary 色，bold）",
        "步骤之间用箭头连接，箭头颜色 secondary",
        "可交替上下排列（zigzag）以节省空间，箭头转向",
        "步骤框顶部可有 3pt 强调线（primary 色）",
        "可选：最后一个步骤用 accent 色边框表示完成状态",
        "流程上方或下方可添加流程名称标签（如「主流程」「异常处理」）",
        "整体背景可添加极淡的连接虚线网格（secondary 色，透明度 5%）"
    ],
    "example": {
        "kicker": "工程实践",
        "title": "模型训练流程",
        "subtitle": "端到端自动化训练流水线",
        "direction": "horizontal_zigzag",
        "steps": [
            {"num": "01", "title": "数据收集", "desc": "采集标注训练数据"},
            {"num": "02", "title": "数据清洗", "desc": "去噪标准化处理"},
            {"num": "03", "title": "特征工程", "desc": "提取有效特征表示"},
            {"num": "04", "title": "模型训练", "desc": "梯度下降迭代优化"},
            {"num": "05", "title": "评估验证", "desc": "测试集性能评测"},
            {"num": "06", "title": "部署上线", "desc": "容器化服务发布"}
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{流程名称}",
        "subtitle": "{流程说明}",
        "direction": "horizontal",
        "steps": [
            {"num": "01", "title": "{步骤1}", "desc": "{描述1}"},
            {"num": "02", "title": "{步骤2}", "desc": "{描述2}"},
            {"num": "03", "title": "{步骤3}", "desc": "{描述3}"}
        ]
    },
    "when_to_use": [
        "有明确步骤顺序的操作流程",
        "工作流/流水线展示",
        "方法论步骤说明"
    ],
    "never": [
        "禁止超过6个步骤",
        "禁止步骤框大小不一",
        "禁止箭头方向混乱",
        "禁止步骤描述超过30字"
    ]
}
