TEMPLATE = {
    "name": "politics-ideology",
    "name_cn": "思政/团课",
    "description": "适合思政教育、团课培训、爱国主义教育等场景。政治性强，价值观明确，结构清晰。",
    "target_audience": "学生、共青团员、党员、干部",
    "typical_slides": 16,
    "typical_duration": "20-30分钟",
    "palette": "patriotic_blue",
    "typography": {
        "header": "Georgia",
        "body": "Calibri",
        "chinese_header": "微软雅黑",
        "chinese_body": "微软雅黑"
    },
    "scene_guidance": {
        "teaching_principles": [
            "政治表达要准确规范，与中央精神保持一致",
            "理论联系实际，避免空洞说教",
            "以理服人，以情感人",
            "注重互动引导，激发学员思考"
        ],
        "key_elements": [
            "理论学习：核心思想与理论基础",
            "案例分析：典型案例与榜样力量",
            "实践要求：如何将思想落实到行动",
            "总结提升：巩固学习成果"
        ]
    },
    "teaching_methods": {
        "theory": "讲授法：系统讲解理论知识",
        "case": "案例法：通过故事引发共鸣",
        "discussion": "讨论法：引导学员思考交流",
        "practice": "实践法：布置具体行动任务"
    },
    "slide_structure": [
        {
            "index": 1,
            "type": "title_slide",
            "title": "新时代青年的使命与担当",
            "subtitle": "学习贯彻党的二十大精神",
            "author": "团支部",
            "date": "2025年5月4日",
            "notes": "标题页庄重大气，体现政治严肃性",
            "filling_prompt": "必须填入真实内容：title 为本次思政教育的主题名称（如'新时代青年的使命与担当'），subtitle 为概括性副标题，author 为演讲者姓名，date 为实际日期。禁止保留花括号。"
        },
        {
            "index": 2,
            "type": "agenda",
            "title": "目录",
            "kicker": "目录",
            "items": [
                "01  理论学习",
                "02  案例分析",
                "03  实践要求",
                "04  总结提升"
            ],
            "notes": "让学员快速了解课程结构",
            "filling_prompt": "目录页为固定结构。"
        },
        {
            "index": 3,
            "type": "section_divider",
            "number": "01",
            "title": "理论学习",
            "subtitle": "核心思想与理论基础",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 4,
            "type": "example_detail",
            "title": "核心思想解读",
            "content_type": "example_detail",
            "kicker": "实例 · 理论学习",
            "lede": "青年兴则国家兴，青年强则国家强",
            "context_block": "党的二十大报告指出，'青年强，则国强。当代中国青年生逢其时，施展才干的舞台无比广阔，实现梦想的前景无比光明。'这是党中央对当代青年的殷切期望，也是时代赋予我们的历史使命。",
            "solution_block": "新时代青年的使命与担当包括：1）坚定理想信念：听党话、跟党走，做习近平新时代中国特色社会主义思想的坚定信仰者；2）练就过硬本领：珍惜韶华、不负青春，努力学习掌握科学知识；3）勇于创新创造：敢为人先、敢于突破，以聪明才智贡献国家；4）矢志艰苦奋斗：不畏艰难、勇毅前行，在实践中增长本领。",
            "metrics": [
                {"value": "二十大报告", "label": "重要政策文件", "trend": "2022年发布"},
                {"value": "1000万+", "label": "团员规模", "trend": "持续壮大"},
                {"value": "95后成主力", "label": "青年结构变化", "trend": "更具活力"}
            ],
            "takeaway": "启示：将个人理想融入国家发展，在民族复兴中成就自我",
            "notes": "解读核心思想的关键要点",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：lede 一句话概括核心思想的价值；context_block 描述学习和践行的背景必要性（1-2句话）；solution_block 展开核心思想的主要内涵和实践要求（2-3句话）；metrics_grid 提供3个指标；takeaway 用一句话总结如何将理论转化为行动。禁止虚构理论内容。references 列出 URL。"
        },
        {
            "index": 5,
            "type": "quote_slide",
            "title": "重要论述",
            "content_type": "quote_slide",
            "quotes": [
                {
                    "text": "广大青年要坚定不移听党话、跟党走，怀抱梦想又脚踏实地，敢想敢为又善作善成。",
                    "source": "习近平在中国共产党第二十次全国代表大会上的报告",
                    "date": "2022年10月16日"
                }
            ],
            "notes": "展示领导人重要论述",
            "filling_prompt": "必须填入真实内容：提供1-2条重要论述（需准确引用），注明出处（讲话名称、时间、场合）。禁止虚构论述。"
        },
        {
            "index": 6,
            "type": "section_divider",
            "number": "02",
            "title": "案例分析",
            "subtitle": "典型案例与榜样力量",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 7,
            "type": "image_text",
            "title": "典型案例：黄文秀",
            "content_type": "image_text",
            "layout_hint": "right-image",
            "kicker": "榜样力量",
            "header": "黄文秀：脱贫攻坚路上的青春之花",
            "sub_header": "用生命诠释初心与使命",
            "paragraph": "黄文秀同志是北京大学硕士毕业生，2018年主动请缨到广西百色市乐业县百坭村担任驻村第一书记。她遍访建档立卡贫困户，手绘地图，扎根基层，带领群众发展产业、脱贫致富。在她的努力下，百坭村实现整村脱贫。然而，2019年6月，她在扶贫路上遭遇山洪，因公殉职，年仅30岁。黄文秀同志用短暂而绚烂的一生，诠释了新时代青年的责任与担当，是广大青年学习的榜样。",
            "references": [
                "https://www.gmw.cn/",
                "https://www.xinhuanet.com/"
            ],
            "notes": "用图文混排展示典型案例，增强可信性和感染力",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：kicker 为'榜样力量'；title 中的 {人物/事件名称} 替换为真实人物姓名或事件名称；图片占位由生成器自动渲染（灰色虚线框+文字提示），无需传入 image_placeholder 参数；header 为案例标题（如'XX同志先进事迹'，不超过35字）；sub_header 为身份简介（不超过35字）；paragraph 为300-450字的自然语言段落，详细描述该人物或事件的主要事迹、贡献和影响，包含具体事例，禁止罗列要点。references 列出 URL。禁止虚构人物和事迹。"
        },
        {
            "index": 8,
            "type": "card_grid",
            "title": "榜样人物",
            "content_type": "card_grid",
            "layout_hint": "2x2",
            "cards": [
                {"header": "黄文秀", "body": "北师大硕士，放弃城市工作，回到家乡担任驻村第一书记，带领全村脱贫，年仅30岁牺牲在扶贫路上"},
                {"header": "陈祥榕", "body": "戍边战士，'清澈的爱，只为中国'，在加勒万河谷冲突中英勇牺牲，年仅19岁"},
                {"header": "张桂梅", "body": "丽江华坪女子高级中学校长，帮助近2000名贫困女孩走出大山，践行教育扶贫理念"},
                {"header": "航天青年团队", "body": "北斗、嫦娥、天问等航天任务中的青年科研人员，用青春托起中国航天梦"}
            ],
            "notes": "展示2-4位榜样人物的简要事迹",
            "filling_prompt": "必须先通过 web_search 获取权威参考资料（至少2个URL），再填入真实内容：提供4位榜样人物，每人有 header（姓名+身份，不超过35字）和 body（详细描述其主要事迹、贡献和影响力，100-120字，包含具体事例或数据）。禁止虚构人物。references 列出 URL。"
        },
        {
            "index": 9,
            "type": "section_divider",
            "number": "03",
            "title": "实践要求",
            "subtitle": "如何将思想落实到行动",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 10,
            "type": "process_flow",
            "title": "实践路径",
            "content_type": "process_flow",
            "direction": "horizontal_zigzag",
            "steps": [
                {"num": "01", "title": "坚定信念", "desc": "加强理论学习，提高政治素养"},
                {"num": "02", "title": "刻苦学习", "desc": "珍惜学习时光，掌握过硬本领"},
                {"num": "03", "title": "勇于实践", "desc": "积极参加社会实践，增长见识"},
                {"num": "04", "title": "服务奉献", "desc": "参与志愿服务，贡献青春力量"}
            ],
            "notes": "4个具体可行的实践路径",
            "filling_prompt": "必须填入真实内容：提供4个具体可操作的实践路径，每条有 title（行动名称）和 desc（具体做法，不超过30字）。禁止空泛口号。"
        },
        {
            "index": 11,
            "type": "example_detail",
            "title": "行动指南",
            "content_type": "example_detail",
            "kicker": "实例 · 实践要求",
            "lede": "知行合一，将学习成果转化为具体行动",
            "context_block": "当前，部分青年存在'躺平'心态，只说不做、知行脱节的问题。学习二十大精神，不能停留在口头上、纸面上，必须落实到具体行动中。",
            "solution_block": "具体行动建议包括：1）制定个人学习计划，每月读一本理论书籍；2）参加志愿服务，每年不少于50小时；3）立足本职岗位建功，争当业务能手；4）积极参与团组织活动，发挥模范带头作用。通过这些具体行动，真正将二十大精神内化于心、外化于行。",
            "metrics": [
                {"value": "每月1本", "label": "理论学习计划", "trend": "可执行"},
                {"value": "每年50h", "label": "志愿服务时长", "trend": "切实可行"},
                {"value": "季度1次", "label": "自我检视", "trend": "持续改进"}
            ],
            "takeaway": "启示：从小事做起，从现在做起，在实践中成长成才",
            "notes": "具体的行动建议",
            "filling_prompt": "必须填入真实内容：lede 一句话概括关键行动；context_block 描述实践中的主要差距和不足（1-2句话）；solution_block 具体说明如何将思想认识转化为实际行动（2-3句话）；metrics_grid 提供3个行动指标；takeaway 用一句话总结如何坚持知行合一。"
        },
        {
            "index": 12,
            "type": "section_divider",
            "number": "04",
            "title": "总结提升",
            "subtitle": "巩固学习成果",
            "filling_prompt": "固定内容，无需额外填充。"
        },
        {
            "index": 13,
            "type": "example_detail",
            "title": "学习要点回顾",
            "content_type": "example_detail",
            "kicker": "实例 · 总结提升",
            "lede": "新时代青年要在实现民族复兴的赛道上奋勇争先",
            "context_block": "本次团课系统学习了党的二十大报告中关于青年工作的重要论述，分析了黄文秀等榜样人物的先进事迹，探讨了将理论转化为实践的具体路径。",
            "solution_block": "核心要点回顾：1）坚定理想信念是中国青年的精神支柱；2）练就本领是中国青年的立身之本；3）勇于担当是中国青年的时代使命；4）接续奋斗是中国青年的优良传统。作为新时代青年，我们要以实际行动践行二十大精神。",
            "metrics": [
                {"value": "掌握核心", "label": "理论掌握程度", "trend": "提升"},
                {"value": "明确方向", "label": "实践应用场景", "trend": "清晰"},
                {"value": "持续学习", "label": "后续学习计划", "trend": "制定"}
            ],
            "takeaway": "启示：以榜样为镜，以奋斗为桨，在民族复兴中书写青春华章",
            "notes": "回顾本节课的核心要点",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最核心的收获；context_block 回顾学习内容和方法（1-2句话）；solution_block 总结核心要点和启示（2-3句话）；metrics_grid 提供3个学习效果指标；takeaway 用一句话总结如何将学习成果转化为行动。"
        },
        {
            "index": 14,
            "type": "example_detail",
            "title": "心得体会",
            "content_type": "example_detail",
            "kicker": "实例 · 心得体会",
            "lede": "学习榜样力量，感悟初心使命",
            "context_block": "通过今天的团课学习，我被黄文秀等榜样的事迹深深感动。他们用实际行动诠释了什么是责任、什么是担当、什么是奉献。",
            "solution_block": "我的心得体会是：1）要树立远大理想，将个人梦想与国家发展结合；2）要脚踏实地，从身边小事做起；3）要勇于挑战，在困难中磨练意志。作为新时代的青年，我要以他们为榜样，在自己的岗位上发光发热，为社会做出贡献。",
            "metrics": [
                {"value": "深受触动", "label": "情感触动程度", "trend": "深刻"},
                {"value": "明确方向", "label": "改进方向", "trend": "清晰"},
                {"value": "积极带动", "label": "对团队的贡献", "trend": "带动更多人"}
            ],
            "takeaway": "启示：以榜样力量激励自己，用实际行动影响他人",
            "notes": "引导学员思考和分享",
            "filling_prompt": "必须填入真实内容：lede 一句话概括最深刻的感悟；context_block 描述学习中的思考和触动（1-2句话）；solution_block 分享心得体会和行动计划（2-3句话）；metrics_grid 提供3个反思指标；takeaway 用一句话总结如何带动他人一起进步。引导性问题可作为参考（如'通过今天的学习，我最深的体会是...'），但最终内容应精炼为2-3句心得总结。"
        },
        {
            "index": 15,
            "type": "kpi_dashboard",
            "title": "学习效果自评",
            "content_type": "kpi_dashboard",
            "layout_hint": "2x2",
            "kicker": "自评 · 学习效果",
            "kpis": [
                {"value": "9分", "label": "理论理解程度", "delta": "vs 学习前+3分", "baseline": "满分10分"},
                {"value": "8分", "label": "情感认同程度", "delta": "vs 学习前+2分", "baseline": "满分10分"},
                {"value": "7分", "label": "行动意愿强度", "delta": "vs 学习前+3分", "baseline": "满分10分"},
                {"value": "立即行动", "label": "下一步计划", "delta": "制定中", "baseline": "待确定"}
            ],
            "notes": "帮助学员自我评估学习效果",
            "filling_prompt": "必须填入真实内容：提供4个学习效果自评维度，每个有 value（如分数或等级）、label（维度名称）、delta（变化趋势）。"
        },
        {
            "index": 16,
            "type": "summary_slide",
            "title": "总结",
            "key_points": [
                "01 坚定理想信念，听党话、跟党走",
                "02 练就过硬本领，在实践中增长才干",
                "03 我承诺：从今天起，积极参与志愿服务"
            ],
            "thank_you": "感谢聆听",
            "notes": "简洁有力的结尾",
            "filling_prompt": "必须填入真实内容：key_points 提供3个要点（核心学习要点2条+行动承诺1条）。禁止保留花括号。"
        }
    ],
    "design_tips": [
        "政治表达要准确规范",
        "案例要有真实性和感染力",
        "实践建议要具体可操作",
        "避免空洞说教，注重以理服人",
        "设计要有仪式感和庄重感",
        "理论联系实际，引发共鸣",
        "设置互动环节，激发思考",
        "用故事而非说教传达价值观"
    ],
    "teaching_notes": {
        "interactive_activities": [
            "小组讨论：'我眼中的青年担当'",
            "情景剧：模拟青年奋斗场景",
            "演讲比赛：'我的青春故事'",
            "志愿服务：组织一次实践活动"
        ],
        "assessment_methods": [
            "课堂参与度",
            "心得体会质量",
            "实践活动表现",
            "理论知识测试"
        ]
    }
}
