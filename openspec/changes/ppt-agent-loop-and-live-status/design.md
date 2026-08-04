## Context

`TaskState` 已为每个 SSE 事件分配递增 ID，`/stream?after_id=` 也支持增量回放，但 Dashboard 切换任务时会清空游标，并在恢复缓存后从事件 0 重新连接。同时，会话接口会返回 `full_answer`，前端又把重放到的 answer chunk 组装成新消息，最终形成重复内容。

生成链路中，主 Agent 使用通用 `edit_file` 修改 `tasks.json`。模板脚手架模式需要一次重写多页内容，生成后又逐页写 status 并重新读取全文件，模型调用、工具调用和 JSON 转义失败都会放大闭环时间。后端已经具备 `WriteTasksManifest` 原子落盘和基于输出文件的 reconcile 能力，可以作为结构化工具的基础。

## Goals / Non-Goals

**Goals:**

- 页面切换、断线重连和历史恢复都不重复 AI 消息或工具事件。
- 用户始终能看到一个简短、实时、可理解的当前状态。
- 主 Agent 通过结构化工具一次性更新页面规划，通过文件落地自动收敛完成状态。
- 减少主 Agent 的读写往返、无效自然语言输出和 JSON 修复循环。

**Non-Goals:**

- 本轮不替换 Eino ADK 或现有 SlideExecutor。
- 本轮不重新启用在线 QA。
- 不改变已有任务文件格式，也不迁移历史任务目录。

## Decisions

1. **按任务缓存事件游标和未完成回答。** Dashboard 的任务缓存同时保存 `lastSeenEventID`、结构化消息和当前流式文本。切回运行中任务时从保存的 ID 增量订阅，不清空已有 UI 状态。相比内容字符串去重，事件游标能保留相同文本但属于不同轮次的合法消息。

2. **会话 API 暴露 `latest_event_id`。** 冷加载可先恢复已持久化消息，再明确知道事件日志的边界。前端对已缓存任务使用本地游标，对首次加载任务使用服务端边界或完整回放，避免会话快照和 SSE 的双写竞争。

3. **增加 DeepAgent 专用 manifest 工具。** 工具接受结构化页面数组或单页 patch，在 Go 内完成 schema 解析、合法性校验、字段保留与原子写入。主 Agent 不再拥有对 `tasks.json` 的通用 `edit_file` 权限；其他普通文件能力保持不变。

4. **页面状态由文件事实自动收敛。** SlideExecutor 只生成文件，后端进度轮询通过 `ReconcileTasksManifestOutputFiles` 将对应页面标记为 done。主 Agent 不再为每个成功页面执行 `edit_file + read_file`。

5. **聊天与运行状态分层。** answer 仍进入 Markdown 对话；`system_step`、`tool_call`、phase、runtime meta 和连接状态归一化为一个 `LiveActivity`，在会话框头部下方显示。工具原始参数只保留在开发者轨迹中，不直接占据对话。

6. **状态更新采用前端派生，后端事件保持兼容。** 现有事件已经覆盖阶段、工具调用、错误和文件进度，本轮优先在前端形成稳定文案，只为恢复边界补充 API 字段，降低部署风险。

## Risks / Trade-offs

- [首次加载运行中任务若只恢复持久化轮次，可能短暂看不到正在生成的半轮文本] → 首次进入仍允许从 0 回放；只有带本地游标的任务切换使用增量回放。
- [模型可能不按结构化工具 schema 输出] → 使用 Eino object/array schema、宽容的 `ContentElement` 解码和明确错误，不让无效输入触碰磁盘。
- [自动 reconcile 有轮询延迟] → 文件监测与现有 3 秒轮询继续广播 progress；比模型逐页写状态更快且可靠。
- [移除主 Agent 的通用 tasks.json 写能力会影响无大纲规划] → manifest 工具同时支持初始化和批量更新，覆盖新建与模板脚手架两条路径。

## Migration Plan

1. 先部署兼容字段和前端游标逻辑，旧客户端可忽略 `latest_event_id`。
2. 再启用 manifest 工具并更新 prompt；保留 `read_file` 用于诊断，但移除主 Agent 的 `edit_file`。
3. 通过既有任务与新任务分别验证切换、断线、生成和继续对话。
4. 如 manifest 工具出现回归，可回滚 prompt/tool 注册；任务 JSON 格式未变化，无数据回滚需求。

## Open Questions

- 真实生产 trace 中最优并发数仍需根据模型限流和服务器资源继续调优，本轮不硬编码提高并发。
