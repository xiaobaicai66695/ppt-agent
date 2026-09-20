# UX Contract

## Product context

- Audience: 需要制作、检查和迭代 PPT 的中文创作者；管理员可审阅跨用户任务记录。
- Primary jobs: 创建 PPT、查看规划和渲染进度、下载交付物、发起后续修订；管理员筛选和分页查看执行记录。
- Target market(s): 未单独定义；当前中文界面不意味着日本市场。
- Active locales: 简体中文界面；不声明 `ja` locale 或日本本地化行为。
- Language/content register and native-review policy: 产品工作台使用直接、可执行的简体中文；模型和 API 原样保留必要术语。
- Timezone/calendar policy: 任务创建时间按浏览器本地时区显示；不声明业务日历语义。
- Accessibility target: WCAG 2.2 AA。

## Business-context sources

| Domain / scope | Authoritative source | Source type | Reviewed date |
|---|---|---|---|
| PPT 生命周期与 SSE 交付 | `docs/architecture/ppt-agent-current-architecture-summary.md` | 架构基线 | 2026-09-04 |
| Planner / Reviewer / Fixer 边界 | `docs/architecture/pptspec-planner-reviewer-fixer-boundary.md` | 架构决策 | 2026-09-04 |
| 前端 API 与任务状态 | `frontend/src/api.ts`、`frontend/src/types.ts` | 已验证 API 契约 | 2026-09-04 |
| 管理员权限与执行记录 | `../backend/pkg/runtime/web/middleware.go`、`../backend/pkg/runtime/web/admin_handler.go` | 权限与 API 契约 | 2026-09-20 |

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
| Observable ReAct timeline | `conversationTimeline` + `ConversationTimelineItem.vue` | `src/utils/conversationTimeline.ts`、`src/components/ConversationTimelineItem.vue`、`DashboardPage.vue`、后端 `conversation_trace_events` | 活跃与历史会话的安全 thought / paired tool call-result / delivery event | unit + browser keyboard flow |
| Delivery preview and feedback | `TaskDeliveryPreview.vue`、`AppModal.vue` | `DashboardPage.vue`、任务缩略图/反馈 API | inline / side-rail / thumbnail unavailable / media preview / rated | build + browser flow |
| 管理执行记录表 | `ExecutionRecordsPage.vue` + `AppShell.vue` | `/api/admin/execution-records` | 服务端分页 / 用户 ID 精确筛选 | build + browser flow |

## Flow ledger

| Operation | Trigger | Pending | Success destination | Success feedback | Failure recovery | Focus outcome | Source ref |
|---|---|---|---|---|---|---|---|
| 创建 PPT | 发送 create 意图或手动 PPT 模式 | SSE 保持打开；显示规划和渲染阶段 | 当前会话 | 只要任务已有 PPT 文件即显示下载和缩略图；仅 `complete` 后开放评分 | 缩略图失败明确显示“暂不可用”；生成错误保留会话与已写入文件 | 评分对话框关闭后还原触发焦点 | 架构基线 §2 |
| 普通对话 | 发送 chat 意图 | assistant 流 | 当前会话 | `conversation_complete` 后结束回答 | 行内错误 | 输入区域保留 | `api.ts` |
| 继续修订 | 已交付任务中再次发送 | SSE 保持打开 | 当前会话 | `continue_complete` 后刷新任务 | 行内错误，保留原交付 | 输入区域保留 | `api.ts` |
| 停止任务 | 运行中的停止按钮 | 请求取消 | 当前会话 | 任务状态刷新为 cancelled | 显示 API 错误 | 停止按钮失效后输入恢复 | `api.ts` |
| 查阅执行记录 | 提交用户 ID 筛选、分页或更新 | 保留已有表格并标为 busy | 当前路由；查询参数保留 | 显示返回范围与更新时间 | 行内错误和重试；保留已加载数据 | 筛选输入保持值 | `/api/admin/execution-records` |

## Async and resilience

- Mutation default: 任务创建、继续和取消均等待服务端确认，防止重复提交。
- Idempotency and duplicate-submit policy: `busy` 期间禁用发送；运行中的任务只允许取消。
- Long-running progress and return path: 用户可在会话间切换；重新选择运行中的任务时，先恢复服务端给出的完整持久化时间线，再从该快照的最后 SSE 事件后连接。新到达的 `thought`、工具调用/返回卡与 `final_answer` 必须以服务端事件 ID 的到达顺序嵌入当前对话时间线；`tool_call` 立即创建一张调用卡，`tool_result` 只能回填同一张卡的结果与状态，不能另起底部结果条目，也不能复制出第二张同名工具卡。
- Observable ReAct timeline: 界面仅显示后端发送的安全思考摘要、工具名称、已脱敏的完整调用参数、已脱敏的完整工具返回值和最终回答，绝不展示或伪造模型私有推理。每个 `llm_end` 到下一次 `llm_start` 之间的工具调用归为一组；旧轨迹未携带这些边界时，每段连续工具调用也归为一组。两种分组均默认折叠为一张“工具调用（N 项）”卡；展开后按原顺序展示每项的完整参数、结果和图片预览，不再要求用户逐张收起工具调用卡。完整工具参数和结果使用卡内的有限高度滚动文本区展示。图片预览仅消费已净化的 HTTPS 缩略图/来源字段。会话快照将 `conversation_messages` 与 `conversation_trace_events` 按时间顺序合并；旧任务缺少轨迹记录时，可从已保存的运行事件最佳努力恢复工具与阶段卡。`answer_end` / `llm_end` 到下一次 `llm_start` 后即使服务端复用同一 `segment_id`，前端也必须新建文本卡，不得继续追加到上一段。`complete`、`continue_complete`、`conversation_complete` 到达后结束流式状态但保留本次页面会话中的轨迹。
- Stale-request cancellation/invalidation and pending-state ownership: 重开流前关闭现有 `EventSource`；只有 `complete`、`continue_complete`、`conversation_complete` 是终态。`answer_end` 仅结束文本回答，不关闭流。执行记录的 URL 状态变更会触发最新页请求，旧请求不得写回新查询。
- Stream scroll behavior: 仅当消息容器已在底部 96px 范围内时，新增事件自动跟随到底部；用户手动向上滚动后不得被新事件拉回，回到该范围后自动跟随恢复。
- Failure recovery: 错误保留在当前会话中，随后刷新任务快照，不能把 SSE 连接意外关闭误报为任务交付。
- Delivery feedback: 仅在当前流式生成任务达到 `complete` 或 `continue_complete` 且没有既有评分时自动打开；用户关闭后可通过“评价这份演示”重新打开，不因浏览历史任务而反复打断。

## Navigation and responsive behavior

- Sidebar/drawer transformation: Dashboard 宽屏在右侧保留已有 PPT 文件的交付预览栏；中等宽度隐藏会话列表以保护主画布和预览栏，窄屏将预览栏下移。中栏必须允许收缩，长工具内容只能在其中换行或内部滚动，不得挤压、覆盖或遮挡交付栏。应用导航与会话列表在各自既有窄屏断点隐藏；主操作仍可见。
- Delivery thumbnail interaction: 可用缩略图是带可访问名称的按钮，点击后使用共享 `AppModal` 的 `media` 变体放大预览；Escape、点击遮罩和关闭按钮都会关闭，并把焦点还原到触发缩略图。缩略图加载失败显示不可用状态，下载操作保持独立可用。
- 执行记录表: 管理员从运营后台进入；表格在窄屏自身横向滚动，不约束应用主滚动；详情用原生 disclosure 展示完整链路 ID。
- Truncation/full-value access: 会话标题允许单行截断；消息正文和错误信息允许读取完整内容。
- Focus restoration and sticky-obstruction policy: native button/input 保持可见焦点；不在任务完成时抢夺焦点。

## Validation

- Required static commands: `npm test`、`npm run build`、Premium 静态审计。
- Browser/device/locale/theme matrix: Dashboard 的深/浅色、窄屏、PPT 创建与普通对话；执行记录页的加载、空态、无权限、筛选、分页、深/浅色和窄屏横向表格；当前不承诺日本本地化验证。
- Canonical sibling flow used for comparison: `ComposePage.vue` 的创建入口和 Dashboard 的继续生成入口；`AdminPage.vue` 的管理员权限加载模式。
