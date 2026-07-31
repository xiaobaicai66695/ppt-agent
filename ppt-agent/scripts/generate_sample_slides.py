"""
为所有单页模板生成模拟数据样例PPT（共24个）。
"""
import sys
from pathlib import Path

# 设置路径
script_dir = Path(r"d:\environment\codeGo\llm-examples\projects\ppt-agent\temp_samples")
script_dir.mkdir(parents=True, exist_ok=True)

generators_pkg_dir = Path(r"d:\environment\codeGo\llm-examples\projects\ppt-agent\skills\visual_designer")
sys.path.insert(0, str(generators_pkg_dir))

from generators import (
    new_presentation, save_slide,
    generate_title_slide, generate_section_divider, generate_content_slide,
    generate_stat_slide, generate_quote_slide, generate_card_grid,
    generate_timeline, generate_process_flow, generate_two_column,
    generate_three_column, generate_summary_slide, generate_image_text,
    generate_example_detail, generate_deep_dive, generate_agenda,
    generate_case_study, generate_kpi_dashboard,
    generate_chart_slide, generate_icon_grid, generate_swot_analysis,
    generate_comparison_table, generate_brand_focus, generate_region_map,
    generate_kanban,
)


def save(prs, filename):
    """保存最后一页"""
    output_path = script_dir / filename
    save_slide(prs.slides[-1], str(output_path))
    print(f"  -> {filename}")


def generate_all_samples():
    """为每个模板类型生成样例"""
    palette = "ocean_soft"

    print("=" * 60)
    print("开始生成单页模板样例（共24个）...")
    print(f"输出目录: {script_dir}")
    print("=" * 60)

    # 1. 标题页
    print("\n[1] 标题页 title_slide")
    prs = new_presentation(palette=palette)
    generate_title_slide(
        prs=prs, palette=palette,
        kicker="技术分享 · 2026",
        title="人工智能技术概述",
        subtitle="从机器学习到大模型的演进之路",
        author="张三 | 高级工程师",
        date="2026年5月"
    )
    save(prs, "01_title_slide.pptx")

    # 2. 目录页
    print("\n[2] 目录页 agenda")
    prs = new_presentation(palette=palette)
    generate_agenda(
        prs=prs, palette=palette,
        kicker="目录",
        title="内容概览",
        items=[
            "01  技术背景与现状",
            "02  核心原理讲解",
            "03  行业应用案例",
            "04  未来发展趋势",
            "05  总结与展望"
        ]
    )
    save(prs, "02_agenda.pptx")

    # 3. 章节分隔页
    print("\n[3] 章节分隔页 section_divider")
    prs = new_presentation(palette=palette)
    generate_section_divider(
        prs=prs, palette=palette,
        number="01",
        title="技术背景",
        subtitle="从感知机到深度学习的跨越",
        kicker="第一章"
    )
    save(prs, "03_section_divider.pptx")

    # 4. 内容页
    print("\n[4] 普通内容页 content_slide")
    prs = new_presentation(palette=palette)
    generate_content_slide(
        prs=prs, palette=palette,
        kicker="要点 · 核心技术",
        title="深度学习的三大核心要素",
        section_header="支撑现代AI的基础",
        bullets=[
            "海量数据：互联网时代每天产生数十亿条训练样本",
            "强大算力：GPU/TPU集群提供每秒千万亿次计算能力",
            "算法创新：从CNN到Transformer不断突破性能上限",
            "开源生态：TensorFlow/PyTorch降低技术门槛"
        ]
    )
    save(prs, "04_content_slide.pptx")

    # 5. 双栏对比页
    print("\n[5] 双栏对比页 two_column")
    prs = new_presentation(palette=palette)
    generate_two_column(
        prs=prs, palette=palette,
        kicker="方案对比",
        title="CNN vs Transformer 架构对比",
        left_header="CNN（卷积神经网络）",
        right_header="Transformer（自注意力网络）",
        left_intro="卷积神经网络是计算机视觉的基础架构，通过局部感受野和权重共享实现高效的图像特征提取。",
        right_intro="Transformer通过自注意力机制实现全局依赖建模，在NLP和CV领域都取得了突破性进展。",
        left_bullets=[
            "局部感受野，擅长提取局部特征",
            "参数共享，计算效率高",
            "平移不变性，特征提取稳定"
        ],
        right_bullets=[
            "全局注意力，任意位置建模依赖",
            "并行计算，训练效率高",
            "可解释性强，注意力可视化"
        ]
    )
    save(prs, "05_two_column.pptx")

    # 6. 三栏并列页
    print("\n[6] 三栏并列页 three_column")
    prs = new_presentation(palette=palette)
    generate_three_column(
        prs=prs, palette=palette,
        kicker="能力矩阵",
        title="AI技术三大应用场景",
        columns=[
            {"header": "计算机视觉", "bullets": ["图像分类准确率99%+", "目标检测实时30fps", "人脸识别精度99.99%"]},
            {"header": "自然语言处理", "bullets": ["机器翻译BLEU 42+", "情感分析准确率95%", "智能问答理解率90%+"]},
            {"header": "语音技术", "bullets": ["语音识别准确率98%", "语音合成自然度4.5/5", "声纹识别准确率99%+"]}
        ]
    )
    save(prs, "06_three_column.pptx")

    # 7. 卡片阵列
    print("\n[7] 卡片阵列 card_grid")
    prs = new_presentation(palette=palette)
    generate_card_grid(
        prs=prs, palette=palette,
        kicker="能力 · 核心模块",
        title="产品四大核心能力",
        subtitle="全方位赋能企业数字化转型",
        layout="2x2",
        cards=[
            {"header": "智能问答系统", "body": "基于大模型的自然语言交互系统，支持多轮对话、上下文记忆与精准意图识别，已服务超过千家企业的日常问答需求。", "footer": "↑ 3倍 效率提升"},
            {"header": "多模态理解", "body": "支持文本、图像、音视频的统一理解与跨模态信息融合处理，适用于智能客服、内容审核与知识图谱构建等多种复杂业务场景。", "footer": "全模态覆盖"},
            {"header": "知识推理引擎", "body": "复杂逻辑推理与知识关联网络构建能力，支持多跳问答与因果分析，在金融风控、医疗诊断等领域已验证显著的准确率提升效果。", "footer": "准确率 99.9%"},
            {"header": "个性化推荐", "body": "基于深度学习的实时推荐算法，通过用户行为分析和内容特征匹配，实现千人千面的个性化推荐，显著提升用户体验和转化率。", "footer": "转化率 +45%"}
        ]
    )
    save(prs, "07_card_grid.pptx")

    # 8. 时间轴
    print("\n[8] 时间轴 timeline")
    prs = new_presentation(palette=palette)
    generate_timeline(
        prs=prs, palette=palette,
        kicker="技术演进",
        title="AI技术发展里程碑",
        subtitle="从深度学习到大模型的时代跨越",
        direction="horizontal",
        nodes=[
            {"year": "2012", "event": "AlexNet\nImageNet突破", "icon": "01"},
            {"year": "2017", "event": "Transformer\n注意力机制", "icon": "02"},
            {"year": "2020", "event": "GPT-3\n超大模型涌现", "icon": "03"},
            {"year": "2023", "event": "GPT-4\n多模态融合", "icon": "04"},
            {"year": "2025", "event": "Agent\n自主执行", "icon": "05"}
        ]
    )
    save(prs, "08_timeline.pptx")

    # 9. 流程图
    print("\n[9] 流程图 process_flow")
    prs = new_presentation(palette=palette)
    generate_process_flow(
        prs=prs, palette=palette,
        kicker="工程实践",
        title="模型训练全流程",
        subtitle="端到端自动化训练流水线",
        direction="horizontal_zigzag",
        steps=[
            {"num": "01", "title": "数据收集", "desc": "采集多源训练数据"},
            {"num": "02", "title": "数据清洗", "desc": "去噪和标准化处理"},
            {"num": "03", "title": "特征工程", "desc": "提取有效特征表示"},
            {"num": "04", "title": "模型训练", "desc": "分布式梯度下降"},
            {"num": "05", "title": "评估验证", "desc": "测试集性能评测"},
            {"num": "06", "title": "部署上线", "desc": "容器化灰度发布"}
        ]
    )
    save(prs, "09_process_flow.pptx")

    # 10. 关键数字页
    print("\n[10] 关键数字页 stat_slide")
    prs = new_presentation(palette=palette)
    generate_stat_slide(
        prs=prs, palette=palette,
        kicker="年度成果",
        title="2025年度核心指标",
        subtitle="关键技术指标一览",
        stats=[
            {"number": "99.99", "unit": "%", "label": "系统可用性", "trend": "↑ 0.3%"},
            {"number": "3.2", "unit": "倍", "label": "性能提升", "trend": "↑ 220%"},
            {"number": "500", "unit": "万+", "label": "服务用户数", "trend": "↑ 40%"}
        ]
    )
    save(prs, "10_stat_slide.pptx")

    # 11. KPI仪表盘
    print("\n[11] KPI仪表盘 kpi_dashboard")
    prs = new_presentation(palette=palette)
    generate_kpi_dashboard(
        prs=prs, palette=palette,
        kicker="数据 · 季度总结",
        title="2026年Q1核心指标",
        subtitle="业务线关键绩效数据",
        kpis=[
            {"value": "99.99%", "label": "系统可用性", "delta": "↑ 0.3%", "baseline": "vs 2025Q4: 99.96%"},
            {"value": "5.2亿", "label": "月活跃用户", "delta": "↑ 40%", "baseline": "vs 2025Q4: 3.7亿"},
            {"value": "¥12亿", "label": "季度营收", "delta": "↑ 28%", "baseline": "vs 2025Q4: ¥9.4亿"},
            {"value": "8.6分", "label": "客户满意度", "delta": "↑ 0.5", "baseline": "vs 2025Q4: 8.1分"}
        ]
    )
    save(prs, "11_kpi_dashboard.pptx")

    # 12. 金句页
    print("\n[12] 金句页 quote_slide")
    prs = new_presentation(palette=palette)
    generate_quote_slide(
        prs=prs, palette=palette,
        kicker="金句",
        quote="预测未来的最好方式，就是创造未来。",
        attribution="—— Alan Kay，计算机科学家"
    )
    save(prs, "12_quote_slide.pptx")

    # 13. 图文混排
    print("\n[13] 图文混排 image_text")
    prs = new_presentation(palette=palette)
    generate_image_text(
        prs=prs, palette=palette,
        kicker="产品介绍",
        title="智能推荐引擎核心架构",
        layout="right-image",
        header="个性化推荐系统",
        sub_header="基于深度学习的实时推荐引擎",
        paragraph="该推荐引擎通过深度学习算法实时分析用户行为数据，构建多维度用户画像，实现千人千面的个性化推荐。系统支持秒级响应，能够在用户浏览商品的瞬间完成推荐计算。同时通过A/B测试框架持续优化推荐效果，目前已在电商、内容平台等多个场景验证，显著提升用户点击率和转化率。"
    )
    save(prs, "13_image_text.pptx")

    # 14. 品牌价值聚焦
    print("\n[14] 品牌价值聚焦 brand_focus")
    prs = new_presentation(palette=palette)
    generate_brand_focus(
        prs=prs, palette=palette,
        kicker="品牌战略",
        title="品牌价值主张",
        subtitle="以用户为中心的核心价值体系",
        center_text="用户\n至上",
        surrounding_points=[
            {"title": "创新", "description": "持续突破技术边界", "color": "secondary"},
            {"title": "品质", "description": "精益求精的产品", "color": "accent"},
            {"title": "服务", "description": "专业贴心的支持", "color": "secondary"},
            {"title": "责任", "description": "可持续发展的承诺", "color": "accent"}
        ],
        principles=[
            {"title": "以用户为中心", "description": "每一项决策都基于用户真实需求"},
            {"title": "长期主义", "description": "坚持做正确的事，而非容易的事"},
            {"title": "开放创新", "description": "拥抱变化，持续学习和迭代"},
            {"title": "共赢合作", "description": "与伙伴共同成长，共享成果"}
        ]
    )
    save(prs, "14_brand_focus.pptx")

    # 15. 看板进度页
    print("\n[15] 看板进度页 kanban")
    prs = new_presentation(palette=palette)
    generate_kanban(
        prs=prs, palette=palette,
        kicker="项目管理",
        title="项目进度看板",
        subtitle="敏捷开发任务跟踪与可视化",
        columns=[
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
        progress=65,
        stats="待办 3 项 | 进行中 2 项 | 已完成 3 项"
    )
    save(prs, "15_kanban.pptx")

    # 16. 区域版图页
    print("\n[16] 区域版图页 region_map")
    prs = new_presentation(palette=palette)
    generate_region_map(
        prs=prs, palette=palette,
        kicker="战略布局",
        title="全国业务版图",
        subtitle="战略布局覆盖全国主要经济区域",
        regions=[
            {"name": "华东", "value": "35%", "trend": "+18%", "detail": "长三角核心区"},
            {"name": "华北", "value": "28%", "trend": "+12%", "detail": "京津冀协同发展"},
            {"name": "华南", "value": "18%", "trend": "+25%", "detail": "粤港澳大湾区"},
            {"name": "西南", "value": "6%", "trend": "+10%", "detail": "成渝双城经济圈"},
            {"name": "东北", "value": "5%", "trend": "+3%", "detail": "老工业基地"},
            {"name": "华中", "value": "8%", "trend": "+15%", "detail": "中部崛起战略"}
        ]
    )
    save(prs, "16_region_map.pptx")

    # 17. 图表专页 - 柱状图
    print("\n[17] 图表专页 chart_slide (柱状图)")
    prs = new_presentation(palette=palette)
    generate_chart_slide(
        prs=prs, palette=palette,
        kicker="数据分析",
        title="季度营收对比",
        subtitle="2025年各季度营收数据一览",
        chart_type="bar",
        data={
            "labels": ["Q1", "Q2", "Q3", "Q4"],
            "datasets": [
                {"name": "2025年", "values": [1200, 1500, 1800, 2200]},
                {"name": "2024年", "values": [900, 1100, 1300, 1600]}
            ]
        },
        show_legend=True
    )
    save(prs, "17_chart_slide_bar.pptx")

    # 18. 图表专页 - 饼图
    print("\n[18] 图表专页 chart_slide (饼图)")
    prs = new_presentation(palette=palette)
    generate_chart_slide(
        prs=prs, palette=palette,
        kicker="数据分析",
        title="市场份额分布",
        subtitle="2025年各产品线占比",
        chart_type="pie",
        data={
            "labels": ["企业版", "专业版", "基础版", "免费版"],
            "datasets": [{"name": "市场份额", "values": [45, 30, 15, 10]}]
        },
        show_legend=True
    )
    save(prs, "18_chart_slide_pie.pptx")

    # 19. 图表专页 - 折线图
    print("\n[19] 图表专页 chart_slide (折线图)")
    prs = new_presentation(palette=palette)
    generate_chart_slide(
        prs=prs, palette=palette,
        kicker="数据分析",
        title="用户增长趋势",
        subtitle="近6个月活跃用户数变化",
        chart_type="line",
        data={
            "labels": ["1月", "2月", "3月", "4月", "5月", "6月"],
            "datasets": [
                {"name": "2025年", "values": [100, 120, 150, 180, 220, 280]},
                {"name": "2024年", "values": [80, 90, 100, 110, 130, 150]}
            ]
        },
        show_legend=True
    )
    save(prs, "19_chart_slide_line.pptx")

    # 20. 图标网格页
    print("\n[20] 图标网格页 icon_grid")
    prs = new_presentation(palette=palette)
    generate_icon_grid(
        prs=prs, palette=palette,
        kicker="核心能力",
        title="六大技术支柱",
        subtitle="构建完整AI技术体系",
        layout="3x2",
        icons=[
            {"icon": "研", "label": "基础研究", "color": "primary"},
            {"icon": "算", "label": "算力平台", "color": "secondary"},
            {"icon": "数", "label": "数据治理", "color": "accent"},
            {"icon": "模", "label": "模型训练", "color": "primary"},
            {"icon": "工", "label": "工程落地", "color": "secondary"},
            {"icon": "安", "label": "安全合规", "color": "accent"}
        ]
    )
    save(prs, "20_icon_grid.pptx")

    # 21. SWOT分析页
    print("\n[21] SWOT分析页 swot_analysis")
    prs = new_presentation(palette=palette)
    generate_swot_analysis(
        prs=prs, palette=palette,
        kicker="战略分析",
        title="AI产品战略SWOT分析",
        subtitle="基于市场与竞争格局的全面评估",
        swot={
            "strengths": {"items": ["技术领先，算法准确率高", "团队经验丰富", "数据积累深厚", "客户口碑好"]},
            "weaknesses": {"items": ["成本高，定价缺乏竞争力", "品牌知名度不足", "销售渠道单一", "国际化能力弱"]},
            "opportunities": {"items": ["市场需求快速增长", "政策支持AI发展", "行业标准尚未成熟", "潜在合作伙伴多"]},
            "threats": {"items": ["大厂入局，竞争加剧", "技术迭代快", "数据合规要求趋严", "经济下行影响IT预算"]}
        }
    )
    save(prs, "21_swot_analysis.pptx")

    # 22. 对比表格页
    print("\n[22] 对比表格页 comparison_table")
    prs = new_presentation(palette=palette)
    generate_comparison_table(
        prs=prs, palette=palette,
        kicker="选型对比",
        title="AI平台选型对比",
        subtitle="三大云厂商AI能力全面对比",
        headers=["对比维度", "AWS SageMaker", "Google Vertex", "Azure ML"],
        rows=[
            ["模型训练速度", "快", "最快", "中等"],
            ["价格", "较高", "中等", "较低"],
            ["易用性", "中等", "简单", "简单"],
            ["生态丰富度", "丰富", "一般", "一般"],
            ["中国区支持", "有限", "有限", "良好"]
        ],
        recommendation="综合考虑，建议选择 Azure ML"
    )
    save(prs, "22_comparison_table.pptx")

    # 23. 实例详解页
    print("\n[23] 实例详解页 example_detail")
    prs = new_presentation(palette=palette)
    generate_example_detail(
        prs=prs, palette=palette,
        kicker="实例 · 金融风控",
        title="蚂蚁金服AlphaRisk：实时风控系统",
        lede="年交易额超万亿，欺诈损失率从0.8%降至0.02%",
        context_block="传统规则引擎误报率高(15%)，人工审核人力成本巨大，且无法识别新型欺诈模式。",
        solution_block="基于深度学习构建实时风控模型，融合设备指纹、行为序列、关系图谱多维特征。采用流计算引擎实现50ms内响应，覆盖支付全链路。",
        metrics=[
            {"value": "99.99%", "label": "欺诈识别准确率", "trend": "↑ 0.3%"},
            {"value": "50ms", "label": "平均响应时间", "trend": "↓ 80%"},
            {"value": "¥100亿+", "label": "年减损金额", "trend": "↓ 60%"}
        ],
        takeaway="启示：多模态特征融合+实时流计算是风控系统的核心竞争力。"
    )
    save(prs, "23_example_detail.pptx")

    # 24. 总结页
    print("\n[24] 总结页 summary_slide")
    prs = new_presentation(palette=palette)
    generate_summary_slide(
        prs=prs, palette=palette,
        kicker="总结",
        title="核心要点回顾",
        key_points=[
            "01  深度学习从CNN到Transformer持续演进突破",
            "02  大模型通过海量参数实现涌现智能能力",
            "03  行业落地从单点应用走向系统化解决方案",
            "04  未来发展聚焦多模态、自主学习、安全可信"
        ],
        thank_you="感谢聆听",
        contact="联系邮箱: tech@company.com | 公众号: AI研究院"
    )
    save(prs, "24_summary_slide.pptx")

    print("\n" + "=" * 60)
    print("所有样例生成完成！")
    print(f"共生成 24 个样例文件")
    print("=" * 60)


if __name__ == "__main__":
    generate_all_samples()
