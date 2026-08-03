## Why

当前 PPT Agent 在真实交付中同时暴露了页面重叠、重复预览、会话历史丢失、消息不可读、长提示词破坏布局、图标资源缺失等问题，用户无法可靠判断产物是否完整，也无法看清 Agent 从用户意图到执行结果的偏离位置。生成入口与续聊入口分离还进一步割裂了任务上下文和模板选择，因此需要对会话、交付、可观测性和部署资源做一次跨层收敛。

## What Changes

- 修复完成态布局重叠，并对预览文件按稳定页面身份去重，保证每个逻辑页只显示一次。
- 将任务过程、历史对话和续聊统一为可恢复的 Markdown 消息流；SSE 增量按原始块合并，不再臆造分句。
- 建立用户意图、规划大纲、当前 Agent 阶段、页面执行结果与偏离告警之间的可视化追踪。
- 为任务标题提供确定性的摘要/截断展示，完整提示词仍可按需查看。
- 将生成任务与 AI 续聊合并到同一会话输入框，并允许在提交前选择本次任务使用的自定义模板。
- 使 Visual Designer 的图标和背景资源在仅部署 `ppt-agent` 目录时仍可解析，并为缺失图标提供可见降级。
- **BREAKING** 移除项目 CLI 的 Windows 专用入口和构建产物约定，仅保留 Linux 命令行脚本、路径和发布目标。

## Capabilities

### New Capabilities

- `ppt-delivery-integrity`: 定义完成态布局、逻辑页面去重、渐进预览和资源失败状态的交付一致性要求。
- `unified-agent-conversation`: 定义生成与续聊共用会话、历史恢复、Markdown 增量渲染及模板选择行为。
- `ppt-visual-assets-portability`: 定义仅部署 `ppt-agent` 时图标/背景资源的自包含解析和缺失资源降级。
- `linux-cli-distribution`: 定义项目 CLI 只面向 Linux 的入口、路径与发布约束。

### Modified Capabilities

- `ppt-agent-developer-status-ui`: 将仅展示运行元数据升级为可追踪“用户意图 -> 规划 -> 执行 -> 产物”的偏离可观测界面。

## Impact

- 前端：`DashboardPage.vue`、`Sidebar.vue`、`AppShell.vue`、`SlidePreviewCard.vue`、API 类型、Markdown 渲染和会话输入组件。
- 后端：任务详情、会话历史、SSE 事件、输出文件归一化、模板选择透传及 RuntimeMeta/事件报告。
- PPT skill：`skills/visual_designer` 的资源清单、解析器和图标生成器。
- 运维与 CLI：项目脚本、构建说明和 Windows 专用文件；服务器只需部署 `ppt-agent` 目录。
- 保持既有 REST/SSE 路径兼容，优先通过可选字段扩展契约；Linux-only CLI 属于明确的不兼容变更。
