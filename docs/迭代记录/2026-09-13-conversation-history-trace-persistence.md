# 2026-09-13 会话历史执行轨迹持久化

- 提交：`15bdb34 fix: persist conversation execution history`
- 范围：历史 PPT 会话回放。

## 改动

- 新增 `conversation_trace_events`，仅持久化已发送到工作台的安全思考摘要、工具调用/结果和关键进度；不保存模型私有推理。
- `/api/tasks/:id/conversation` 按时间合并会话消息与持久化轨迹，前端恢复可折叠的思考、工具调用/结果卡；对旧任务从 `runtime_event_records` 最佳努力回放工具与阶段卡。
- 删除任务时同步删除轨迹记录。

## 验证与发布

- 本地：`go test ./pkg/runtime/task ./pkg/runtime/web ./pkg/db`、`go build ./...`、`npm test -- --run src/utils/conversationTimeline.test.ts`、`npm run build` 均通过。
- UI 静态审计：Premium strict audit 0 findings；`DESIGN.md` lint 0 errors（保留 5 条既有 orphaned-token warnings）。
- 部署：`remote-dev:/ppt/ppt-agent`，替换后端源码、skills、Linux 二进制和前端 `dist`；保留远端 `backend/.env` 与 `weboutput`。
- 新进程：PID `1404404`，`/api/health` 返回 200，启动日志确认 MySQL 连接和单实例锁正常。
- 线上低成本冒烟：临时检索对话的历史接口返回 `message, thought, tool_call, tool_result`；数据库写入 4 条安全轨迹记录。临时任务、消息、轨迹和工作目录均已清理。
- 追加发布（`b75754e`，2026-09-13）：工具调用的已脱敏参数和返回内容改为完整传递与持久化，前端以最多约 196px 高的卡内滚动文本区查看。`remote-dev:/ppt/ppt-agent` 已按全量后端源码、skill、Linux 二进制和前端 `dist` 替换；保留远端 `backend/.env` 与 `weboutput`。二进制含 `args_display`，发布 CSS 含 `max-height:196px`。未创建消耗模型额度的线上对话，完整内容传递由本地超过 24KB 的单元回归覆盖。
- 追加前端发布（`116e0ba`，2026-09-13）：每个 `llm_end → llm_start` 窗口中的工具调用默认合并折叠为一张“工具调用（N 项）”卡，展开后显示各项完整参数、返回内容和图片预览；历史回放采用相同边界。发布替换 `frontend/dist` 后，服务改由 transient `ppt-agent.service` 的非阻塞 systemd 单元托管，最终 PID `1414137`，20 秒稳定检查无重启；内网与公网 `/api/health`、首页和 Dashboard 均为 200，入口 JS 为 `index-BM5n9jed.js`。
- 追加前端发布（`170ba95`，2026-09-13）：兼容没有 `llm_start` / `llm_end` 的旧任务回放，将连续的工具调用隐式合并为默认折叠的“工具调用（N 项）”卡；思考、回答或其他时间线事件会结束该组。已仅替换 `remote-dev:/ppt/ppt-agent/frontend/dist`，不改后端与任务数据；`ppt-agent.service` 为 active、`:8080` 正在监听，内网 `/api/health` 与 Dashboard 均为 200。该行为由前端 28 项单元回归、生产构建和严格设计审计覆盖，未发起消耗模型额度的线上任务。
- 追加后端发布（`c4161bd`，2026-09-13）：修复 RuntimeMeta 已记录但未被时间线转发的 `llm_start` / `llm_end`。新生成任务会通过 SSE 并写入 `conversation_trace_events` 保存这两个边界，因此工具调用可严格按“前一轮 `llm_end` 到下一轮 `llm_start`”分组；已完成的旧记录不会伪造边界，继续使用连续工具调用兼容分组。按发布规范完整替换远端后端源码和 `skills`、保留 `backend/.env` 与 `weboutput`，并安装 Linux 交付二进制。服务 transient 单元因停止后自动移除，已以非阻塞方式重新注册；PID `1424158`，20 秒稳定检查 `NRestarts=0`，内外网 `/api/health` 和 Dashboard 均为 200。未发起消耗模型额度的线上任务。
- 追加前端发布（`65192d7`，2026-09-13）：交付缩略图改为可访问按钮，点击后通过共享 `AppModal` 的媒体变体放大预览，并支持 Escape、遮罩关闭与触发焦点恢复；下载链接仍独立可用。Dashboard 中栏和消息时间线补齐可收缩的 Grid/overflow 约束，长工具详情不再撑开中栏并覆盖右侧交付预览。仅替换 `remote-dev:/ppt/ppt-agent/frontend/dist`；入口脚本为 `index-ELgD5bUK.js`，服务保持 active，内外网 `/api/health` 和 Dashboard 均为 200。前端 28 项单元回归、类型生产构建和严格 UI 审计通过；`DESIGN.md` lint 无 error（5 项既有 token 引用 warning），浏览器自动化 CLI 未返回可执行工作流，因此未自动截图验证。

## 遗留边界

- 本次发布前已完成的旧任务没有 `conversation_trace_events`，只能从已有运行事件恢复工具和阶段摘要；无法补回当时未持久化的完整中间文本。
