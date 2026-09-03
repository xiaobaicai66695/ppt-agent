## Why

闲聊链路只有路由 benchmark，缺少对最终回答、检索整合和来源渲染的质量门。截图中的“先泛化引导做 PPT、再堆叠低质量来源”会在模型或搜索结果变化后反复出现，且无法由现有四类 suite 捕获。

## What Changes

- 新增独立 `chat` benchmark suite，并提供 test/validation 各至少十条独立样例。
- 对闲聊回答建立结构化、可离线验证的质量契约：直接回答、上下文补全、检索降级、来源可点击与图片引用。
- 收紧生产闲聊 fallback 与资料来源清洗，避免有资料时输出泛化 PPT 引导、将检索/图片失败包装为成功，或输出不可用链接。
- 将 chat suite 接入当前 Go test benchmark harness 的加载、真实模型执行与 coverage regression 检查。

## Capabilities

### New Capabilities

- `chat-agent-quality-evaluation`: 闲聊 Agent 的离线 case、真实模型/Judge 评测、覆盖和隔离门禁。

### Modified Capabilities

- `unified-task-conversations`: 闲聊回复在资料、图片和降级场景下的可用性与可追溯性要求。

## Impact

- `ppt-agent/backend/pkg/web/message_chat.go`、搜索结果清洗和 benchmark 适配层。
- `ppt-agent/backend/test/chat_benchmark`、`ppt-agent/test/chat_benchmark/{testdata,README.md}` 与现有 benchmark README。
- 现有 Dashboard 对话输出仅消费更可靠的 Markdown，无 API breaking change。
