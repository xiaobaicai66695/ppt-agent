# 幻灯片类型体系

## 类型分类

### 结构引导类
- title_slide: 标题页
- agenda: 目录页
- section_divider: 章节分隔页
- summary_slide: 总结页

### 内容陈述类
- content_slide: 普通文字内容页（兜底类型）
- quote_slide: 金句/引言页

### 对比与并列类
- two_column: 双栏对比
- three_column: 三栏并列
- card_grid: 卡片阵列（4~8个要点）

### 流程与关系类
- timeline: 时间轴
- process_flow: 步骤流程图

### 数据与图表类
- stat_slide: 关键数字页
- kpi_dashboard: 指标看板

### 内容叙事类（必须优先使用）
- example_detail: 实例详解页（命名案例+背景+方案+数据+启示）
- deep_dive: 深入详解页（双栏：左原理右代码/架构/数据）
- case_study: 案例研究页（Context→Problem→Solution→Results）

## 内容深度决策树

| 场景 | 类型 |
|------|------|
| 介绍概念 | example_detail（命名具体案例） |
| 讲解原理 | deep_dive（左原理右代码） |
| 展示效果 | case_study（必须有量化结果） |
| 多个平等要点 | card_grid / three_column |
| 时间顺序 | timeline |
| 数据展示 | stat_slide / kpi_dashboard |
| 无特殊特征 | content_slide |
