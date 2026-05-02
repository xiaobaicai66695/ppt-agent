# 模板目录说明

本目录包含 PPT 模板资产，分为两类：

## 目录结构

```
templates/
├── single-page/      # 单页布局模板（按类型分类）
│   ├── title_slide.py        # 标题页
│   ├── content_slide.py      # 普通内容页
│   ├── two_column.py         # 双栏对比
│   ├── three_column.py       # 三栏并列
│   ├── card_grid.py          # 卡片阵列
│   ├── timeline.py           # 时间轴
│   ├── process_flow.py       # 流程步骤
│   ├── stat_slide.py        # 大数字强调
│   ├── quote_slide.py        # 金句引言
│   ├── section_divider.py    # 章节分隔
│   ├── image_text.py         # 图文混排
│   └── summary_slide.py      # 总结页
└── full-decks/      # 完整PPT模板（Python 模板文件）
    ├── tech-sharing.py       # 技术分享模板
    ├── ai-intro.py          # AI大模型介绍
    ├── product-launch.py     # 产品发布
    ├── weekly-report.py      # 周报
    ├── pitch-deck.py         # 商业计划
    └── course-module.py       # 课程课件
```

> 模板以 Python 脚本形式存储（`TEMPLATE` 变量），AI 可直接读取源码作为示例参考，无需解析 JSON。

## 使用方式

Agent 在规划时会参考 `full-decks` 模板确定整体结构，
生成单页时会参考 `single-page` 模板确定具体布局和内容格式。

详见 `SKILL.md` 中的「模板系统」章节。
