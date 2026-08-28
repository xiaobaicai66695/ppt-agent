# 2026-08-29 前端流式会话卡顿热修复

## 问题

长时间运行的任务持续接收 `runtime_meta` 快照和 `answer` SSE 片段时，前端会把滑动快照错误累积为完整历史，并且每个片段都会重新渲染整段 Markdown、执行一次滚动。运行轨迹中的 `assistant_output` 还会和实时流并行参与正文渲染，导致同一段内容被重复展示。

## 修复

- `frontend/src/utils/workbench.ts`：将合并后的运行事件限制为 120 条尾部事件；后端快照本身只提供最近事件，前端不再无界累积。
- `frontend/src/pages/DashboardPage.vue`：将 answer SSE 片段以 100ms 为批次合并，再触发响应式 Markdown 更新；任务结束、切换和断开连接时立即刷新或清空待处理片段。
- `frontend/src/components/ConversationComposer.vue`：实时流存在时只使用实时流作为 AI 正文，避免与运行轨迹重复；自动滚动合并到动画帧。
- `frontend/src/utils/workbench.test.ts`：新增运行事件尾部限制回归用例。

## 验证与上线

- 本地：`npm test -- --run src/utils/workbench.test.ts`，39 个测试通过；`npm run build` 通过。
- 部署：2026-08-29 05:00 +08:00，将提交 `d030abc` 的 `frontend/dist` 发布到 `remote-dev:/ppt/ppt-agent/frontend/dist`；旧版本保留为 `frontend/dist.bak.20260829-stream-hotfix`。
- 线上：后端进程 PID `174218` 未重启，仍监听 `:8080`；`/api/health`、`/`、`/dashboard` 及新 Dashboard 静态资源均返回 200；远端 Dashboard bundle SHA-256 与本地构建一致。

## 遗留边界

本修复消除前端重复渲染和无界事件累积。若模型本身持续产生语义上相似但不完全相同的内容，仍应另行从 Agent 的工具调用预算和 prompt 约束处理。
