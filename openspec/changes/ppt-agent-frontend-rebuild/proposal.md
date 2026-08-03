## Why

PPT Agent 的前端功能已经可以完成生成与交付，但各页面缺少统一应用结构，首页仍是营销落地页，Compose 与 Dashboard 又堆叠了过多同权重面板，导致产品显得死板、模板化，用户难以快速聚焦“创建、等待、预览、继续修改”四个核心动作。现在需要在稳定 API 和任务链路之上重建完整前端体验，让真实幻灯片内容成为视觉中心。

## What Changes

- 建立统一的应用壳、窄导航、上下文顶栏、页面标题和响应式导航规则，覆盖 Home、Compose、Dashboard 与 Admin。
- 将 Home 从营销页改成 prompt-first 创建工作区，首屏直接提供需求输入、模板入口和最近任务路径。
- 将 Compose 改造成编辑器式三段工作区：资源/模板、页面轨道、属性编辑，并保持移动端完整操作能力。
- 重排 Dashboard 的进度、预览、日志、聊天和诊断层级，使生成状态与已就绪幻灯片始终优先。
- 统一 Auth 与 Admin 的视觉语言、表单反馈、数据密度和空/错/加载状态。
- 引入 `lucide-vue-next` 作为统一图标集，移除触及页面中的装饰性渐变、光斑、emoji 和不一致手写图标。
- 建立可维护的全局设计 token 与轻量通用组件，保留现有 Vue、API、SSE、任务类型和后端契约。

## Capabilities

### New Capabilities

- `ppt-application-shell`: 统一应用导航、上下文顶栏、响应式壳层和跨页面可访问交互。
- `ppt-creation-workspace`: prompt-first 创建入口、编辑器式 Compose、生成交付 Dashboard 以及真实内容优先的视觉体验。

### Modified Capabilities

- `ppt-agent-developer-status-ui`: 进一步约束诊断信息在新 Dashboard 中保持次级和按需展开，不得压过用户进度与预览。

## Impact

- 前端：`ppt-agent/frontend/src/App.vue`、路由页面、Dashboard/Compose 组件、新增应用壳与基础 UI 组件。
- 依赖：新增 `lucide-vue-next`，不引入完整 UI 框架或新的状态管理库。
- 文档：新增 UI 重构预研和可维护设计系统说明。
- 后端与接口：不改变现有 REST、SSE、认证、任务和模板 API 契约。
