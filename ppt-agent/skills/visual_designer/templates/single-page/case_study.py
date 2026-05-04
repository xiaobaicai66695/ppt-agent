TEMPLATE = {
    "type": "case_study",
    "name": "案例研究页",
    "description": "结构化展示 Context→Problem→Solution→Results 的完整案例链路。适用于技术方案落地、产品迭代、行业应用等场景。",
    "layout_hint": "left",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "案例 · {行业/领域}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "font_size": "30pt",
            "font_weight": "bold",
            "alignment": "left",
            "margin_bottom": "0.2in",
            "max_chars": 40
        },
        "context": {
            "label": "Context",
            "label_color": "secondary",
            "label_bg": "light_bg",
            "font_size": "14pt",
            "max_chars": 120,
            "content_placeholder": "{行业背景和现状，1-2 句}",
            "icon": "背景"
        },
        "problem": {
            "label": "Problem",
            "label_color": "accent",
            "label_bg": "light_bg",
            "font_size": "14pt",
            "font_weight": "bold",
            "color": "primary",
            "max_chars": 80,
            "content_placeholder": "{核心痛点，1 句话点明}",
            "icon": "痛点"
        },
        "solution": {
            "label": "Solution",
            "label_color": "primary",
            "label_bg": "light_bg",
            "font_size": "14pt",
            "max_lines": 3,
            "content_placeholder": "{技术方案描述，含技术栈、关键设计决策、与替代方案的对比。2-3 句。}",
            "icon": "方案"
        },
        "results": {
            "label": "Results",
            "label_color": "secondary",
            "label_bg": "light_bg",
            "layout": "horizontal cards or bullet list",
            "items_min": 3,
            "content_placeholder": "{3-4 个量化指标，带对比基准}",
            "items": [
                {"metric": "指标名", "value": "具体数值", "comparison": "对比基准"}
            ],
            "example": [
                {"metric": "推理延迟", "value": "18ms", "comparison": "优化前 320ms，降低 94%"},
                {"metric": "模型准确率", "value": "99.99%", "comparison": "超过人工标注(99.95%)"},
                {"metric": "运维成本", "value": "¥12万/月", "comparison": "旧方案 ¥80万/月，节省 85%"}
            ]
        }
    },
    "visual_elements": [
        "Context/Problem/Solution/Results 四个区域用浅底色卡片分隔（light_bg）",
        "每个区域左侧有对应颜色的竖线装饰（context: secondary, problem: accent, solution: primary, results: secondary）",
        "每个区域的 label 使用带背景的标签样式（label_bg 色）",
        "Problem 的卡片可用 warm/accent 边框突出（表示这是需要解决的问题）",
        "Results 区域的三项指标用横向卡片排列，每卡片顶部有主色强调线（primary 色，3pt）",
        "Results 中的数值用大号字体（20pt）加粗 primary 色，标签用小字（12pt）text_muted",
        "整体布局从标题到结果层层递进，视觉上体现「问题→方案→结果」逻辑"
    ],
    "example": {
        "kicker": "案例 · 云基础设施",
        "title": "从自建 IDC 迁移到云原生: 某出行平台的架构演进",
        "context": "某头部出行平台日订单量超 5000 万，原有自建 IDC 面临弹性不足、运维复杂、多活容灾困难等问题",
        "problem": "节假日峰值流量是日常 5 倍，IDC 扩容周期需 2 周，无法弹性响应，导致高峰时段用户打车超时率高达 12%",
        "solution": "采用 Kubernetes + Istio 构建统一编排层，数据库迁移至云原生 DB，消息队列从 Kafka 自建切换至云服务托管版。核心业务做单元化拆分，按城市维度隔离故障域。CI/CD 从 Jenkins 迁移至 GitOps(ArgoCD)",
        "results": [
            {"metric": "扩容周期", "value": "2 分钟", "comparison": "从 2 周缩短至 2 分钟，效率提升 10,000x"},
            {"metric": "打车超时率", "value": "0.3%", "comparison": "从 12% 降至 0.3%，用户体验大幅改善"},
            {"metric": "运维人力", "value": "3 人", "comparison": "从 25 人缩减至 3 人，成本降低 88%"},
            {"metric": "全年可用率", "value": "99.995%", "comparison": "从 99.9% 提升至 99.995%，全年停机 < 26 分钟"}
        ]
    },
    "example_2": {
        "kicker": "案例 · {行业领域}",
        "title": "{案例标题}",
        "context": "{行业背景和现状}",
        "problem": "{核心痛点}",
        "solution": "{技术方案描述}",
        "results": [
            {"metric": "{指标1}", "value": "{数值}", "comparison": "{对比说明}"},
            {"metric": "{指标2}", "value": "{数值}", "comparison": "{对比说明}"},
            {"metric": "{指标3}", "value": "{数值}", "comparison": "{对比说明}"}
        ]
    },
    "when_to_use": [
        "展示技术方案的实际落地效果",
        "说服听众采用某项技术或方案",
        "产品/技术迭代的 before/after 汇报"
    ],
    "never": [
        "禁止使用 emoji — 标签和装饰中禁止使用任何 emoji 字符",
        "禁止缺少量化结果——results 至少 3 条且每条含具体数字",
        "禁止 solution 只写技术名不写细节——必须展开 2-3 句",
        "禁止 context 和 problem 混在一起——必须分开展示"
    ]
}
