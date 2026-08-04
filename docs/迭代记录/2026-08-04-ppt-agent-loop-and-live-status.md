# 2026-08-04 PPT Agent 生成闭环与实时状态加固

## 迭代目标

修复任务页切换后 AI 回答重复、生成闭环中反复整文件写入 `tasks.json`、步骤输出冗长且用户无法判断是否卡住的问题。本次事项登记为 `PPT-UX-004`，进入 OpenSpec change：`openspec/changes/ppt-agent-loop-and-live-status/`。

## 已完成

- 会话 API 返回 `latest_event_id` 和 `replay_after_event_id`，Dashboard 按任务缓存 SSE 游标、结构化消息和未完成流式回答。
- 切回运行中任务时从已见事件边界增量恢复，并对持久化会话与实时事件做幂等合并，避免同一回答重复出现。
- AI 正文只进入统一会话框；任务活动保留阶段、工具、文件和错误等执行事件，避免同一内容展示两次。
- 会话框增加实时状态行，覆盖思考、读取模板、搜索、内容规划、并行生成、渲染、连接恢复、失败和完成。
- 新增 `update_tasks_manifest` 结构化原子工具，主 Agent 不再使用通用 `edit_file` 覆盖 `tasks.json`。
- 主 Agent 将多页内容规划合并为一次批量更新；页面完成状态由后端根据输出文件自动收敛，减少读写和模型往返。
- 收紧主 Agent 输出，去除机械工具叙述、JSON 修复过程和重复步骤播报。

## 验证与部署

- `go test ./pkg/agent/deep ./pkg/task ./pkg/web ./pkg/prompts` 通过。
- `go build ./...` 通过。
- `npm test -- --run` 通过，共 13 项；`npm run build` 通过。
- `openspec validate ppt-agent-loop-and-live-status --strict` 通过。
- 已同步到 `/ppt/ppt-agent` 并随本轮版本切换，服务 PID `4094435`，`/api/health` 返回 200，日志确认 MySQL 与模型初始化成功。

## 关联内容

- `docs/issues/todo.md`
- `openspec/changes/ppt-agent-loop-and-live-status/`
- `ppt-agent/backend/pkg/agent/deep/manifest_tool.go`
- `ppt-agent/backend/pkg/task/manager.go`
- `ppt-agent/backend/pkg/web/handler.go`
- `ppt-agent/frontend/src/pages/DashboardPage.vue`
- `ppt-agent/frontend/src/components/ConversationComposer.vue`
- `ppt-agent/frontend/src/utils/workbench.ts`
