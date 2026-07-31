"""Test script to generate all 23 slides for the 百年孤独 PPT demo."""
import sys
from pathlib import Path

script_dir = Path(__file__).parent  # examples_output/
generators_pkg_dir = script_dir.parent.resolve()  # generators/
sys.path.insert(0, str(script_dir.parent.parent))  # visual_designer/

from generators import (
    new_presentation, save_slide,
    generate_title_slide, generate_section_divider, generate_content_slide,
    generate_stat_slide, generate_quote_slide, generate_card_grid,
    generate_timeline, generate_process_flow, generate_two_column,
    generate_three_column, generate_summary_slide, generate_image_text,
    generate_example_detail, generate_deep_dive, generate_agenda,
    generate_case_study, generate_kpi_dashboard, generate_chart_slide,
    generate_icon_grid, generate_swot_analysis, generate_comparison_table,
)

OUTPUT_DIR = script_dir  # examples_output/
OUTPUT_DIR.mkdir(exist_ok=True)
PALETTE = "ocean_soft"
BG = "artistic"

def prs_for_slide():
    return new_presentation(palette=PALETTE)


# ── Slide 1: 封面页 ──
prs = prs_for_slide()
prs = generate_title_slide(
    prs=prs, palette=PALETTE,
    title="百年孤独",
    subtitle="加西亚·马尔克斯的魔幻现实主义巅峰之作",
    author="文学赏析",
    date="2026年",
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "1_封面页.pptx"))
print("Slide 1 done")

# ── Slide 2: 目录页 ──
prs = prs_for_slide()
prs = generate_agenda(
    prs=prs, palette=PALETTE,
    kicker="目录",
    title="内容概览",
    items=[
        "01  作品背景与作者介绍",
        "02  布恩迪亚家族七代人物谱系",
        "03  魔幻现实主义手法解析",
        "04  核心主题与文学影响",
        "05  经典语录与总结回顾",
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "2_目录页.pptx"))
print("Slide 2 done")

# ── Slide 3: 概览页 ──
prs = prs_for_slide()
prs = generate_stat_slide(
    prs=prs, palette=PALETTE,
    title="关键数据一览",
    stats=[
        {"number": "1967", "unit": "年", "label": "首次出版", "trend": "哥伦比亚作家出版社首发"},
        {"number": "5000万", "unit": "册", "label": "全球销量", "trend": "西班牙语小说最高纪录"},
        {"number": "7", "unit": "代", "label": "布恩迪亚家族", "trend": "跨越144年家族史诗"},
        {"number": "1982", "unit": "年", "label": "诺贝尔文学奖", "trend": "马尔克斯获此殊荣"},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "3_概览页.pptx"))
print("Slide 3 done")

# ── Slide 4: 章节分隔页 01 ──
prs = prs_for_slide()
prs = generate_section_divider(
    prs=prs, palette=PALETTE,
    number="01",
    title="作品背景与作者介绍",
    subtitle="加西亚·马尔克斯与《百年孤独》的创作历程",
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "4_章节分隔页.pptx"))
print("Slide 4 done")

# ── Slide 5: 时间轴页 ──
prs = prs_for_slide()
prs = generate_timeline(
    prs=prs, palette=PALETTE,
    title="马尔克斯创作历程",
    direction="horizontal",
    nodes=[
        {"year": "1927", "event": "出生于阿拉卡塔卡", "icon": "01"},
        {"year": "1955", "event": "记者生涯转折", "icon": "02"},
        {"year": "1965", "event": "闭关写作", "icon": "03"},
        {"year": "1967", "event": "《百年孤独》出版", "icon": "04"},
        {"year": "1982", "event": "获诺贝尔文学奖", "icon": "05"},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "5_时间轴页.pptx"))
print("Slide 5 done")

# ── Slide 6: 卡片网格页 ──
prs = prs_for_slide()
prs = generate_card_grid(
    prs=prs, palette=PALETTE,
    kicker="文学成就",
    title="马尔克斯的文学成就",
    layout="2x3",
    cards=[
        {"header": "《百年孤独》", "body": "1967年出版，魔幻现实主义代表作，全球销量超5000万册"},
        {"header": "《霍乱时期的爱情》", "body": "1985年出版，被誉为人类有史以来最伟大的爱情小说"},
        {"header": "《族长的秋天》", "body": "1975年出版，实验性独白体小说，拉美文学巅峰之作"},
        {"header": "《一桩事先张扬的凶杀案》", "body": "1981年出版，基于真实事件，探讨集体沉默与命运"},
        {"header": "《苦妓回忆录》", "body": "2004年出版，马尔克斯最后一部长篇小说，80岁高龄完成"},
        {"header": "诺贝尔文学奖", "body": "1982年获奖，颁奖词称其小说世界中的丰富而想象的世界"},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "6_卡片网格页.pptx"))
print("Slide 6 done")

# ── Slide 7: 内容页 ──
prs = prs_for_slide()
prs = generate_content_slide(
    prs=prs, palette=PALETTE,
    kicker="故事背景",
    title="马孔多的诞生",
    section_header="虚构小镇的现实根基",
    bullets=[
        "马孔多源于马尔克斯童年故乡阿拉卡塔卡小镇，19世纪末香蕉公司进驻哥伦比亚的真实历史",
        "小镇由何塞·阿尔卡蒂奥·布恩迪亚带领一群追随者穿越沼泽建立，象征拉美殖民开拓史",
        "吉普赛人梅尔基亚德斯带来磁铁、望远镜等发明，映射西方文明对拉美的冲击",
        "香蕉公司带来繁荣后引发大屠杀，影射1928年哥伦比亚香蕉工人大屠杀真实事件",
        "马孔多最终被飓风抹去，象征拉美被遗忘的历史循环与孤独命运",
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "7_内容页.pptx"))
print("Slide 7 done")

# ── Slide 8: 章节分隔页 02 ──
prs = prs_for_slide()
prs = generate_section_divider(
    prs=prs, palette=PALETTE,
    number="02",
    title="布恩迪亚家族七代人物谱系",
    subtitle="跨越144年的家族命运轮回",
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "8_章节分隔页.pptx"))
print("Slide 8 done")

# ── Slide 9: 双栏对比页 ──
prs = prs_for_slide()
prs = generate_two_column(
    prs=prs, palette=PALETTE,
    title="家族命名循环与命运重复",
    left_header="男性角色（阿尔卡蒂奥/何塞）",
    right_header="女性角色（乌尔苏拉/阿玛兰妲）",
    left_bullets=[
        "何塞·阿尔卡蒂奥·布恩迪亚：家族创始人，痴迷科学实验，最终被绑在栗树下发疯",
        "阿尔卡蒂奥第二代：身材魁梧，性格冲动，在战争中崛起后被枪杀",
        "奥雷里亚诺上校：发动32场起义全部失败，晚年反复制作小金鱼又融化",
        "阿尔卡蒂奥第三代：继承祖父疯狂基因，在香蕉公司大屠杀中丧生",
    ],
    right_bullets=[
        "乌尔苏拉：家族支柱，活119岁，见证七代兴衰，是唯一保持清醒的人",
        "阿玛兰妲：终身未嫁，织裹尸布又拆掉，在孤独中度过一生",
        "雷梅苔丝（美人儿）：美丽到不真实，17岁升天消失，象征纯粹之美",
        "阿玛兰妲·乌尔苏拉：第六代，试图改革马孔多，与侄子乱伦生下猪尾巴孩子",
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "9_双栏对比页.pptx"))
print("Slide 9 done")

# ── Slide 10: 三栏并列页 ──
prs = prs_for_slide()
prs = generate_three_column(
    prs=prs, palette=PALETTE,
    title="三代核心人物对比",
    columns=[
        {"header": "第一代：开创者", "bullets": ["何塞·阿尔卡蒂奥·布恩迪亚：建立马孔多", "乌尔苏拉：家族精神支柱", "代表开拓与理想主义"]},
        {"header": "第二代：战争与权力", "bullets": ["奥雷里亚诺上校：32场起义", "阿尔卡蒂奥：独裁统治马孔多", "代表暴力与权力异化"]},
        {"header": "第三代：衰落与毁灭", "bullets": ["奥雷里亚诺第二：奢靡放纵", "阿尔卡蒂奥第三：见证大屠杀", "代表堕落与历史遗忘"]},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "10_三栏并列页.pptx"))
print("Slide 10 done")

# ── Slide 11: 图表页 ──
prs = prs_for_slide()
prs = generate_chart_slide(
    prs=prs, palette=PALETTE,
    kicker="数据分析",
    title="布恩迪亚家族七代人口统计",
    subtitle="每代男性与女性成员数量对比",
    chart_type="bar",
    data={
        "labels": ["第1代", "第2代", "第3代", "第4代", "第5代", "第6代", "第7代"],
        "datasets": [
            {"name": "男性成员", "values": [5, 6, 8, 5, 4, 3, 2]},
            {"name": "女性成员", "values": [4, 4, 5, 4, 3, 2, 1]},
        ]
    },
    show_legend=True,
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "11_图表页.pptx"))
print("Slide 11 done")

# ── Slide 12: 章节分隔页 03 ──
prs = prs_for_slide()
prs = generate_section_divider(
    prs=prs, palette=PALETTE,
    number="03",
    title="魔幻现实主义手法解析",
    subtitle="幻想与现实的完美融合",
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "12_章节分隔页.pptx"))
print("Slide 12 done")

# ── Slide 13: 流程图页 ──
prs = prs_for_slide()
prs = generate_process_flow(
    prs=prs, palette=PALETTE,
    title="魔幻现实主义叙事结构",
    direction="horizontal",
    steps=[
        {"num": "1", "title": "循环时间观", "desc": "家族历史不断重复，名字与命运轮回"},
        {"num": "2", "title": "魔幻事件日常化", "desc": "升天、鬼魂、预言以平静语气叙述"},
        {"num": "3", "title": "现实事件魔幻化", "desc": "战争、屠杀、殖民以荒诞手法呈现"},
        {"num": "4", "title": "预言与宿命", "desc": "吉普赛人羊皮卷预言家族百年命运"},
        {"num": "5", "title": "孤独主题贯穿", "desc": "每个角色以不同方式体验孤独"},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "13_流程图页.pptx"))
print("Slide 13 done")

# ── Slide 14: 技术详解页 ──
prs = prs_for_slide()
prs = generate_deep_dive(
    prs=prs, palette=PALETTE,
    kicker="文学手法",
    title="魔幻现实主义核心技法",
    lede="马尔克斯将拉美现实与神话传说融合，创造独特的叙事风格",
    left_header="叙事技巧分析",
    key_points=[
        "以平静语气叙述荒诞事件，消除魔幻与现实的界限",
        "循环时间结构打破线性叙事，命运轮回贯穿全书",
        "大量使用象征与隐喻，如猪尾巴、黄蝴蝶、雨",
        "全知视角与多重视角切换，增强叙事层次感",
    ],
    analysis=[
        "时间维度：环形叙事打破线性传统",
        "空间维度：马孔多与拉美地理一一对应",
    ],
    right_header="经典案例",
    case_example=[
        "美人儿雷梅苔丝升天：抓着床单飞升，家人平静接受",
        "失眠症蔓延：全镇人失去记忆，必须给物品贴标签",
        "四年十一个月零两天大雨：象征马孔多的衰落与清洗",
        "梅尔基亚德斯鬼魂：死后仍在密室撰写羊皮卷预言",
    ],
    data_evidence=[
        "全书含魔幻元素场景约47处",
        "循环命名模式：7代中奥雷里亚诺出现11次",
        "时间跨度144年，但叙事呈环形结构",
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "14_技术详解页.pptx"))
print("Slide 14 done")

# ── Slide 15: 案例研究页 ──
prs = prs_for_slide()
prs = generate_case_study(
    prs=prs, palette=PALETTE,
    kicker="经典场景分析",
    title="香蕉公司大屠杀——历史与魔幻的交织",
    context="1928年哥伦比亚苏克雷省香蕉种植园，数千工人罢工要求改善待遇，政府军镇压造成大量伤亡，但官方长期否认此事发生。",
    problem="真实历史被官方抹去，集体记忆被篡改，受害者后代无法确认祖先遭遇，形成历史虚无主义。",
    solution="马尔克斯以魔幻手法重写历史：大屠杀后只有一列火车运走尸体，但全镇人集体失忆，只有奥雷里亚诺第三从老兵口中得知真相。政府发布官方声明称只有一个人死亡，用谎言覆盖真实。",
    results=[
        {"metric": "历史真相", "value": "公开承认", "comparison": "小说出版后政府被迫承认"},
        {"metric": "文学影响", "value": "全球经典", "comparison": "成为拉美历史文学书写典范"},
        {"metric": "学术价值", "value": "持续研究", "comparison": "影响后殖民文学历史叙事反思"},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "15_案例研究页.pptx"))
print("Slide 15 done")

# ── Slide 16: 章节分隔页 04 ──
prs = prs_for_slide()
prs = generate_section_divider(
    prs=prs, palette=PALETTE,
    number="04",
    title="核心主题与文学影响",
    subtitle="孤独、命运与拉美身份认同",
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "16_章节分隔页.pptx"))
print("Slide 16 done")

# ── Slide 17: KPI仪表盘页 ──
prs = prs_for_slide()
prs = generate_kpi_dashboard(
    prs=prs, palette=PALETTE,
    kicker="数据 · 影响力",
    title="《百年孤独》全球影响力数据",
    subtitle="出版至今的文化传播指标",
    kpis=[
        {"value": "5000万+", "label": "全球累计销量", "delta": "↑", "baseline": "西班牙语小说最高纪录"},
        {"value": "46", "label": "翻译语言数", "delta": "↑", "baseline": "覆盖全球主要语种"},
        {"value": "1982", "label": "诺贝尔文学奖", "delta": "↑", "baseline": "马尔克斯获奖"},
        {"value": "100+", "label": "学术研究专著", "delta": "↑", "baseline": "全球高校必修文本"},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "17_KPI仪表盘页.pptx"))
print("Slide 17 done")

# ── Slide 18: 图标网格页 ──
prs = prs_for_slide()
prs = generate_icon_grid(
    prs=prs, palette=PALETTE,
    kicker="核心主题",
    title="《百年孤独》五大核心主题",
    subtitle="贯穿七代人命运的精神线索",
    layout="3x2",
    icons=[
        {"icon": "孤", "label": "孤独：每个角色以不同方式体验与他人的隔绝", "color": "#4A90D9"},
        {"icon": "命", "label": "宿命：羊皮卷预言不可逃避的家族命运", "color": "#D94A4A"},
        {"icon": "时", "label": "时间循环：历史不断重复，无法打破的轮回", "color": "#4AD9A0"},
        {"icon": "爱", "label": "爱情与乱伦：家族禁忌之恋导致猪尾巴孩子", "color": "#D9A04A"},
        {"icon": "权", "label": "权力异化：从起义领袖到独裁者的堕落", "color": "#9B59B6"},
        {"icon": "忘", "label": "遗忘与记忆：失眠症隐喻拉美历史被抹去", "color": "#1ABC9C"},
    ],
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "18_图标网格页.pptx"))
print("Slide 18 done")

# ── Slide 19: SWOT分析页 ──
prs = prs_for_slide()
prs = generate_swot_analysis(
    prs=prs, palette=PALETTE,
    kicker="战略分析",
    title="魔幻现实主义文学流派分析",
    subtitle="以《百年孤独》为代表的拉美文学爆炸",
    swot={
        "strengths": {
            "label": "优势",
            "items": [
                "将拉美现实与神话融合，创造独特叙事风格",
                "打破西方线性叙事传统，提供全新时间观",
                "语言丰富华丽，充满诗意与想象力",
                "深刻反映拉美殖民、独裁、被遗忘的历史",
            ],
        },
        "weaknesses": {
            "label": "劣势",
            "items": [
                "人物众多且命名重复，读者容易混淆",
                "魔幻元素过多可能削弱现实批判力度",
                "循环结构导致情节推进缓慢",
                "文化背景差异造成非拉美读者理解障碍",
            ],
        },
        "opportunities": {
            "label": "机遇",
            "items": [
                "全球后殖民研究兴起，重新发现其历史价值",
                "影视改编潜力巨大（Netflix已宣布改编计划）",
                "跨学科研究（历史学、人类学、心理学）拓展解读空间",
                "数字人文技术帮助读者理解复杂人物关系",
            ],
        },
        "threats": {
            "label": "威胁",
            "items": [
                "当代读者注意力下降，长篇小说阅读率降低",
                "拉美文学爆炸经典被过度学术化，远离大众",
                "文化挪用争议（西方对拉美叙事的消费）",
                "政治正确审查对魔幻元素真实性的质疑",
            ],
        },
    },
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "19_SWOT分析页.pptx"))
print("Slide 19 done")

# ── Slide 20: 金句页 ──
prs = prs_for_slide()
prs = generate_quote_slide(
    prs=prs, palette=PALETTE,
    kicker="经典开篇",
    quote="多年以后，面对行刑队，奥雷里亚诺·布恩迪亚上校将会回想起父亲带他去见识冰块的那个遥远的下午。",
    attribution="——加西亚·马尔克斯《百年孤独》开篇名句",
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "20_金句页.pptx"))
print("Slide 20 done")

# ── Slide 21: 图文混排页 ──
prs = prs_for_slide()
prs = generate_image_text(
    prs=prs, palette=PALETTE,
    kicker="作者自述",
    title="马尔克斯的创作理念",
    layout="right-image",
    header="魔幻现实主义的根源",
    sub_header="创作自述",
    paragraph="我小说中的任何事物都没有经过预先思考，所有人物在我开始写作时都带着他们自己的思想、语言和行为方式。《百年孤独》中的一切都有真实原型。我外祖父家就有吉普赛人带来新发明，我外祖母讲述故事时语气就像事情刚刚发生。孤独不是寂寞，而是缺乏爱情与团结的能力。布恩迪亚家族的悲剧在于他们学会了孤独，却从未学会团结。马孔多就是阿拉卡塔卡，我童年记忆中的那个被香蕉公司改变的小镇。魔幻不是虚构，而是拉美真实的现实。",
    background=BG,
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "21_图文混排页.pptx"))
print("Slide 21 done")

# ── Slide 22: 总结页 ──
prs = prs_for_slide()
prs = generate_summary_slide(
    prs=prs, palette=PALETTE,
    kicker="总结",
    title="总结回顾",
    key_points=[
        "01  《百年孤独》是魔幻现实主义巅峰之作，1967年出版后全球销量超5000万册",
        "02  布恩迪亚家族七代人跨越144年，命名循环与命运轮回构成独特叙事结构",
        "03  马尔克斯以平静语气叙述魔幻事件，将拉美历史创伤转化为永恒文学经典",
        "04  孤独、宿命、遗忘三大主题深刻反映拉美身份认同与历史困境",
    ],
    thank_you="感谢观看",
    contact="《百年孤独》——一部属于全人类的文学经典",
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "22_总结页.pptx"))
print("Slide 22 done")

# ── Slide 23: 结束页 ──
prs = prs_for_slide()
prs = generate_title_slide(
    prs=prs, palette=PALETTE,
    title="百年孤独",
    subtitle="加西亚·马尔克斯",
    author="文学赏析",
    date="2026年",
)
save_slide(prs.slides[-1], str(OUTPUT_DIR / "23_结束页.pptx"))
print("Slide 23 done")

print(f"\nAll 23 slides saved to: {OUTPUT_DIR}")
