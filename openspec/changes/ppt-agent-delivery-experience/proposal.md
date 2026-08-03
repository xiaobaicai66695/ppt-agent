## Why

PPT Agent 已具备任务生成、SSE 事件和单页文件输出能力，但任务资源缺少归属校验，部分用户 API 无法闭环，缩略图按请求启动重型转换，进度状态在 SSE 丢失后可能失真，前端也仍是桌面固定布局。现在需要把这些分散能力收敛成可靠、快速、面向用户的 PPT 交付体验。

## What Changes

- 为任务详情、事件流、文件、缩略图、取消、删除、继续和会话接口增加任务归属校验，并统一用户身份读取方式。
- 补齐偏好总结路由，修复偏好编辑字段类型与 API 错误处理，确保反馈、洞察和推荐接口可正常鉴权。
- 建立明确的前后端任务/SSE 契约，覆盖系统步骤、阶段、页面进度、文件就绪、预览就绪和终态同步。
- 移除无消费者的完整 PPTX 预下载，采用低分辨率缩略图优先、按需下载原文件的渐进式交付策略。
- 为预设模板提供真实可访问的缩略图，并让单页布局元数据、字段和中文标签在编排页完整呈现。
- 将 Dashboard 和 Compose 调整为现代生产工作台：桌面保持高信息密度，窄屏使用抽屉、单栏或分段视图，关键操作满足键盘和触控要求。
- 将用户可理解的阶段进度置于主界面，开发者 RuntimeMeta/Timeline 改为可折叠诊断区域；移除已停用 QA 流程的默认用户文案。

## Capabilities

### New Capabilities

- `ppt-task-api-contract`: 任务资源授权、用户身份上下文、API 错误语义和前后端结构化契约。
- `ppt-progressive-delivery`: PPT 文件、缩略图和 SSE 进度的渐进式交付、终态恢复与快速预览行为。
- `ppt-workbench-ux`: Dashboard/Compose 响应式工作台、真实模板预览、布局字段呈现和可访问交互。

### Modified Capabilities

- `ppt-agent-developer-status-ui`: 将开发者状态面板调整为按需展开的诊断视图，并确保用户阶段进度始终优先展示。

## Impact

- 后端：`ppt-agent/backend/pkg/web`、`ppt-agent/backend/pkg/task`，以及相关 HTTP/SSE/缩略图测试。
- 前端：`ppt-agent/frontend/src/api.ts`、任务类型、Dashboard、Compose、Sidebar、进度和预览组件。
- 模板资源：`ppt-agent/frontend/public/templates/thumbs` 或等价的可公开访问真实缩略图资源。
- 接口行为：未归属当前用户的任务资源统一返回 404 或 403；前端必须显式处理非 2xx 响应。
- 不引入新的前端框架，不恢复默认在线 QA，不改变 PPT 生成器的内容与排版契约。
