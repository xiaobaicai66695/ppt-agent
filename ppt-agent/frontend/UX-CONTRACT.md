# UX Contract

## Product context

- Audience: 需要制作、检查和迭代 PPT 的中文创作者。
- Primary jobs: 创建 PPT、查看规划和渲染进度、下载交付物、发起后续修订。
- Target market(s): 未单独定义；当前中文界面不意味着日本市场。
- Active locales: 简体中文界面；不声明 `ja` locale 或日本本地化行为。
- Language/content register and native-review policy: 产品工作台使用直接、可执行的简体中文；模型和 API 原样保留必要术语。
- Timezone/calendar policy: 本次流程不显示需由前端格式化的业务日期或日历。
- Accessibility target: WCAG 2.2 AA。

## Business-context sources

| Domain / scope | Authoritative source | Source type | Reviewed date |
|---|---|---|---|
| PPT 生命周期与 SSE 交付 | `docs/architecture/ppt-agent-current-architecture-summary.md` | 架构基线 | 2026-09-04 |
| Planner / Reviewer / Fixer 边界 | `docs/architecture/deckspec-planner-reviewer-fixer-boundary.md` | 架构决策 | 2026-09-04 |
| 前端 API 与任务状态 | `frontend/src/api.ts`、`frontend/src/types.ts` | 已验证 API 契约 | 2026-09-04 |

## Visual contract

- Project `DESIGN.md`: `frontend/DESIGN.md`。
- Token ownership model: 既有运行时 token 为权威来源。
- Runtime design-system/token source: `frontend/src/App.vue`。
- Mapping/export/adapters: `App.vue` 语义 CSS variables → `AppShell.vue` / 页面 scoped CSS。
- Token drift gate: `npm run build` 与 Premium 静态审计；当前无生成型 token 工具。
- Supported themes: 深色、浅色。
- Design-context owner/review policy: 任何持久化 token 或状态呈现变更同时更新 `DESIGN.md`。

## Canonical UI Map

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Form | 页面 Vue 表单与 `src/api.ts` | `DashboardPage.vue`、`ComposePage.vue` | create / continue | unit + build |
| Select/Listbox | 原生 `<select>` | `AccountSettingsDialog.vue`、`ComposePage.vue` | native（接受浏览器/系统弹窗） | keyboard + build |
| Scrollbar | 应用全局样式 | `src/App.vue` | 仅几何例外 | visual QA |
| CRUD | 任务 API 与 Dashboard | `src/api.ts`、后端任务状态 | create / continue / cancel | API + browser flow |
| Long-running progress | SSE 状态机 | `DashboardPage.vue`、后端 `pkg/web/streamer.go` | conversation / PPT generation / continue | unit + browser flow |
| Observable tool trace | `conversationTimeline` 工具函数 | `src/utils/conversationTimeline.ts`、`DashboardPage.vue` | 仅活跃会话的 phase / tool invocation / delivery event | unit + browser keyboard flow |
| Delivery preview and feedback | `TaskDeliveryPreview.vue`、`AppModal.vue` | `DashboardPage.vue`、任务缩略图/反馈 API | completed / thumbnail unavailable / rated | build + browser flow |

## Flow ledger

| Operation | Trigger | Pending | Success destination | Success feedback | Failure recovery | Focus outcome | Source ref |
|---|---|---|---|---|---|---|---|
| 创建 PPT | 发送 create 意图或手动 PPT 模式 | SSE 保持打开；显示规划和渲染阶段 | 当前会话 | 仅 `complete` 后显示交付、文件下载、缩略图和页数；未评分时弹出评分对话框 | 缩略图失败明确显示“暂不可用”；生成错误保留会话 | 评分对话框关闭后还原触发焦点 | 架构基线 §2 |
| 普通对话 | 发送 chat 意图 | assistant 流 | 当前会话 | `conversation_complete` 后结束回答 | 行内错误 | 输入区域保留 | `api.ts` |
| 继续修订 | 已交付任务中再次发送 | SSE 保持打开 | 当前会话 | `continue_complete` 后刷新任务 | 行内错误，保留原交付 | 输入区域保留 | `api.ts` |
| 停止任务 | 运行中的停止按钮 | 请求取消 | 当前会话 | 任务状态刷新为 cancelled | 显示 API 错误 | 停止按钮失效后输入恢复 | `api.ts` |

## Async and resilience

- Mutation default: 任务创建、继续和取消均等待服务端确认，防止重复提交。
- Idempotency and duplicate-submit policy: `busy` 期间禁用发送；运行中的任务只允许取消。
- Long-running progress and return path: 用户可在会话间切换；重新选择运行中的任务时从最后事件后恢复 SSE。新到达的 `system_step`、工具调用与工具结果必须嵌入当前对话时间线；工具结果更新原调用位置，不能移动到页面另一处。工具调用以服务端 `segment_id` 为边界，断线恢复不从 MySQL 补回短期轨迹。
- Observable tool trace: 界面仅显示后端发送的可观察阶段、工具名称、调用说明和结果摘要，绝不展示或伪造模型私有推理。每次工具调用独立展开/收起，默认展开进行中的调用；收起后仍显示工具名称和文字状态。图片预览仅消费已净化的 HTTPS 缩略图/来源字段。`complete`、`continue_complete`、`conversation_complete` 到达后移除所有工具阶段；MySQL 会话快照只恢复用户与助手消息。
- Stale-request cancellation/invalidation and pending-state ownership: 重开流前关闭现有 `EventSource`；只有 `complete`、`continue_complete`、`conversation_complete` 是终态。`answer_end` 仅结束文本回答，不关闭流。
- Failure recovery: 错误保留在当前会话中，随后刷新任务快照，不能把 SSE 连接意外关闭误报为任务交付。
- Delivery feedback: 仅在当前流式生成任务达到 `complete` 或 `continue_complete` 且没有既有评分时自动打开；用户关闭后可通过“评价这份演示”重新打开，不因浏览历史任务而反复打断。

## Navigation and responsive behavior

- Sidebar/drawer transformation: 桌面保持侧栏，窄屏隐藏导航和会话列表；主操作仍可见。
- Truncation/full-value access: 会话标题允许单行截断；消息正文和错误信息允许读取完整内容。
- Focus restoration and sticky-obstruction policy: native button/input 保持可见焦点；不在任务完成时抢夺焦点。

## Validation

- Required static commands: `npm test`、`npm run build`、Premium 静态审计。
- Browser/device/locale/theme matrix: Dashboard 的深/浅色、窄屏、PPT 创建与普通对话；当前不承诺日本本地化验证。
- Canonical sibling flow used for comparison: `ComposePage.vue` 的创建入口和 Dashboard 的继续生成入口。
