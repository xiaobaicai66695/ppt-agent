---
version: alpha
name: "Deckform PPT Agent"
description: "面向创作者的深色/浅色 PPT 生产工作台，以可核验的生成状态和真实交付文件为核心。"
colors:
  primary: "#6ce5ca"
  surface-base: "#0c1d2a"
  surface-raised: "#0b1a27"
  surface-accent: "#0a2631"
  text-strong: "#e9f1f8"
  text-muted: "#b5cbd0"
  accent: "#6ce5ca"
  accent-strong: "#9af1dc"
  info: "#79c6e5"
  danger: "#ef9892"
typography:
  body:
    fontFamily: "'Noto Sans SC', system-ui, sans-serif"
  display:
    fontFamily: "'Noto Serif SC', serif"
  mono:
    fontFamily: "'DM Mono', monospace"
rounded:
  control: "5px"
  panel: "8px"
  signature: "16px"
spacing:
  compact: "8px"
  control: "12px"
  section: "18px"
  panel: "30px"
components:
  workbench:
    backgroundColor: "{colors.surface-base}"
    textColor: "{colors.text-strong}"
  execution-trace:
    textColor: "{colors.text-muted}"
    rounded: "{rounded.panel}"
  message:
    backgroundColor: "{colors.surface-raised}"
    rounded: "{rounded.signature}"
---

# Deckform PPT Agent Design System

## Overview

### Creative North Star

这是一个“演示制作台”，不是营销页：像编辑室里的任务轨迹，主画布清楚呈现用户请求、已完成阶段和可下载交付物；每一次状态变化都可被核验。

### Product context and register

- **Audience and primary job:** 中文创作者提交 PPT 需求、查看规划与渲染进度、下载和继续修订交付物。
- **Target market(s) and evidence:** 目标市场未单独定义；当前界面和任务请求以简体中文为主，依据 `src/pages/DashboardPage.vue` 与根目录 `docs/architecture/ppt-agent-current-architecture-summary.md`。
- **Locale(s) and language policy:** 已有界面文案使用简体中文；模型、文件名和 API 标识可保留其原始语言。当前不是日本市场或 `ja` locale。
- **Usage scene:** 桌面优先的生产工作台，用户在生成期间需要辨别“规划文本已结束”与“PPT 正在渲染”的不同状态。
- **Register:** product。安静、紧凑、操作导向。
- **Memorable signature:** 仅在活跃会话中，对话时间线在可见“请求分析”阶段内嵌逐次工具调用；每次调用可独立展开结果与安全图片预览，而不展示模型私有推理。任务终态只保留对话与交付，不恢复工具轨迹。
- **Restraint:** 不用装饰性渐变、营销式卡片或伪造的百分比进度替代真实任务状态。
- **Anti-references:** 避免聊天应用将流式文本结束误写成任务完成，也避免深色面板堆叠造成的视觉噪声。
- **Token ownership/runtime mapping:** 既有运行时 token 是权威来源（Model B）：`src/App.vue` 的 `:root` 与 `html[data-theme='light']` 定义语义变量，`AppShell.vue` 和各页面消费这些变量。本文件镜像当前接受的 token 及意图；前端构建与 Premium 静态审计是漂移门禁。

## Colors

深色工作台的 `surface-base`、`surface-raised` 和 `surface-accent` 建立安静的层次；`accent` 只用于主操作和运行中状态，`info` 用于已交付，`danger` 只用于错误或破坏性操作。浅色主题由 `src/App.vue` 覆盖同一语义 token，不改变状态含义。焦点样式使用可见的浅绿色描边，状态绝不只靠颜色表达。

## Typography

`Noto Sans SC` 承载常规中文界面；`Noto Serif SC` 仅用于页面标题；`DM Mono` 用于短状态标记与技术元数据。正文保持 12–14px 的紧凑工作台密度，长任务名在自己的区域截断而不挤压相邻操作。

## Layout

`AppShell` 负责固定侧栏和主工作区；活跃 Dashboard 在主画布中按到达顺序排列“任务上下文 → 用户消息 → 可观察阶段与工具调用 → 助手文本 → 交付 → 输入”。每次工具调用以紧凑状态行出现，展开后显示调用说明、结果摘要及受限图片预览；收起不删除状态。终态会话移除工具阶段而保留消息和交付物。`answer_end` 只完成规划说明，不改变任务的运行态或关闭 SSE。窄屏隐藏任务侧栏，主画布保留输入与停止操作。

## Elevation & Depth

常规层次通过相邻表面色和细边框建立；阴影仅可用于浮层和亮色输入区。工作台不使用玻璃拟态或大面积渐变。

## Shapes

常规控制使用 5–8px 圆角；品牌和消息气泡可以使用不对称的签名圆角。图标采用 `lucide-vue-next`，图标按钮须有可访问名称。

## Components

### Foundational visual states

运行中的轨迹显示加载图标与明确的进行中文案；成功、失败和取消均保留文字描述。按钮禁用时不触发请求且保持尺寸；文本框禁用时不暗示仍可提交。

可展开的工具调用使用原生按钮，具有可见焦点、`aria-expanded` 与关联的详情区域；每一项独立收起或展开，状态图标之外必须保留“执行中 / 已完成 / 失败”文字。

### Buttons and actions

主操作使用 `accent` 实底，次级操作使用边框或透明表面。停止生成与删除属于风险操作，必须使用明确动词并与主要生成操作分开。

### Navigation and data display

`AppShell` 是认证工作区的统一导航外壳。会话列表的状态和页数是紧凑摘要；Dashboard 以真实文件和缩略图状态为交付依据，而不是单一“完成”文案。

### Forms and overlays

工作台提交表单使用 `novalidate`，Enter 发送、Shift+Enter 换行。错误就近显示并保留已有消息。需要确认的操作使用应用内对话框；不新增浏览器原生确认。

### Iconography

统一使用 `lucide-vue-next` 的线性图标；信息图标可配短文本，非通用图标按钮必须提供 `aria-label`。

### Motion

仅将转动加载图标和短页面淡入用于状态提示；在 `prefers-reduced-motion` 下不依赖位移动画表达任务状态。

### Content and data visualization

文案使用直接的中文动作词，例如“正在开始生成演示页面”“演示文件已生成”。“组织回答”“规划说明已完成”和“生成阶段结束”是不同事件，不能混为同一终态。

## Do's and Don'ts

- **Do:** 用后端的终态 SSE 事件确认任务完成，并持续呈现渲染进度和交付文件。
- **Do:** 复用 `src/App.vue` 的语义 token 和 `AppShell` 的工作台结构。
- **Don't:** 把 Planner 的文本结束事件视为 PPT 任务已经交付。
- **Don't:** 以营销装饰或模型推理替代任务真实状态。
