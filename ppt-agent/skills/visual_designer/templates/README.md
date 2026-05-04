# 模板目录说明

本目录包含 PPT 模板资产，分为两类：

## 目录结构

```
templates/
├── single-page/      # 单页布局模板（按类型分类）
│   ├── title_slide.py        # 标题页
│   ├── content_slide.py      # 普通内容页
│   ├── two_column.py         # 双栏对比
│   ├── three_column.py        # 三栏并列
│   ├── card_grid.py          # 卡片阵列
│   ├── timeline.py            # 时间轴
│   ├── process_flow.py       # 流程步骤
│   ├── stat_slide.py         # 大数字强调
│   ├── quote_slide.py        # 金句引言
│   ├── quote_slide.py        # 金句引言
│   ├── section_divider.py    # 章节分隔
│   ├── image_text.py         # 图文混排
│   ├── summary_slide.py      # 总结页
│   ├── agenda.py             # 目录页
│   ├── example_detail.py      # 实例详解
│   ├── deep_dive.py          # 深入详解
│   ├── case_study.py         # 案例研究
│   └── kpi_dashboard.py      # 指标看板
└── full-decks/       # 完整PPT模板（Python 模板文件）
    ├── tech-intro.py         # 技术介绍（18页）
    ├── tech-sharing.py        # 技术分享（18页）
    ├── product-launch.py      # 产品发布（12页）
    ├── weekly-report.py       # 周报（9页）
    ├── pitch-deck.py         # 商业计划（16页）
    ├── course-module.py       # 课程课件（17页）
    ├── current-affairs.py     # 时政分享（14页）
    ├── politics-ideology.py    # 思政/团课（16页）
    ├── design-defense.py      # 答辩汇报（12页）
    ├── innovation-compete.py   # 科创竞赛（15页）
    ├── research-report.py     # 调研报告（16页）
    ├── activity-plan.py       # 活动策划（14页）
    ├── personal-summary.py    # 述职报告（12页）
    ├── short-class-talk.py    # 课堂分享（6页）
    ├── meeting-minutes.py     # 会议纪要（12页）
    ├── product-intro.py       # 产品介绍（15页）
    ├── training-course.py     # 培训课件（17页）
    └── project-proposal.py    # 项目提案（16页）
```

> 模板以 Python 脚本形式存储（`TEMPLATE` 变量），AI 可直接读取源码作为示例参考，无需解析 JSON。

## 使用方式

Agent 在规划时会参考 `full-decks` 模板确定整体结构，
生成单页时会参考 `single-page` 模板确定具体布局和内容格式。

详见 `SKILL.md` 中的「模板系统」章节。
