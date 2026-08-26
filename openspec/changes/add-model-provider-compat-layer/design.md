## Context

PPT Agent 上层 Agent 当前依赖 Eino `model.ToolCallingChatModel`，但底层模型工厂仍在 `ppt-agent/backend/pkg/agent/utils/model.go` 中直接构造 Ark `ChatModelConfig`。这让模型供应商成为隐式全局假设：fallback 链只接受模型名，API key 只按 Ark key 解析，thinking 和 JSON schema 转换也绑定 Ark 类型。

2026-08-26 的预研文档 `docs/research/2026-08-26-model-provider-compat-layer.md` 已确认：OpenAI、Ark、DeepSeek、Qwen 均有 Eino/eino-ext 创建入口；硅基流动当前没有独立 Eino component，适合先作为 OpenAI-compatible profile 接入。

## Goals / Non-Goals

**Goals:**

- 抽出 provider-aware 模型兼容层，统一输出 `model.ToolCallingChatModel`。
- 第一阶段支持 Ark、OpenAI/OpenAI-compatible、SiliconFlow profile。
- 保持旧 `ARK_*` 环境变量和现有调用方行为兼容。
- 保留现有 fallback、stream read fallback、tool-call delta sanitize、压缩器和 RuntimeStatus wrapper。
- 为 DeepSeek/Qwen 专用包预留扩展点，但不强制首轮迁移。

**Non-Goals:**

- 不重写 PPTPlanner、Reviewer、Fixer 的 Agent 构造方式。
- 不在第一阶段改数据库结构或前端账号 key 管理 UI。
- 不在第一阶段强制启用 DeepSeek/Qwen 专用 adapter；若 OpenAI-compatible 模式暴露差异，再单独推进。
- 不替换现有 Eino ADK 或工具协议。

## Decisions

### Decision 1: 新增 `pkg/agent/modelcompat`

新增包承载 `Provider`、`ModelSpec`、`Factory` 和 provider adapter。`agentutils` 继续暴露 `NewFallbackToolCallingChatModel`，但内部把配置解析成 `[]modelcompat.ModelSpec` 后交给兼容层。

理由：上层 API 稳定，迁移面小；provider 细节不继续堆在 fallback、压缩和运行状态 wrapper 的同一个文件里。

备选方案：直接在 `utils/model.go` 里追加 switch。这个方案短期少文件，但会让模型创建、fallback、stream 包装、压缩器和运行时状态继续耦合，后续接 Qwen/DeepSeek 会更难维护。

### Decision 2: 第一阶段以 `ark` 和 `openai_compatible` 为主

Ark 保持专用 adapter，OpenAI、SiliconFlow 和通用兼容平台走 OpenAI adapter。SiliconFlow 作为 profile 填默认 base URL 和默认 key env。

理由：当前生产配置已经跑 Ark，必须零破坏；硅基流动没有独立 eino-ext 包，OpenAI-compatible 是最稳的接入面。DeepSeek/Qwen 虽有专用包，但 OpenAI-compatible 可以先满足“换平台不用改代码”的核心目标。

备选方案：首轮直接接 Ark/OpenAI/DeepSeek/Qwen/SiliconFlow 全部专用 adapter。这个方案覆盖更全，但会一次性引入更多依赖和行为差异，测试成本偏高。

### Decision 3: 配置先支持新链路，旧 `ARK_*` 全兜底

新配置使用 provider-aware 条目：

```env
MODEL_PROVIDER=ark
MODEL_CHAIN=primary,backup1,backup2
MODEL_PRIMARY_PROVIDER=ark
MODEL_PRIMARY_NAME=...
MODEL_PRIMARY_API_KEY_ENV=ARK_API_KEY
MODEL_PRIMARY_BASE_URL=...
MODEL_PRIMARY_REGION=cn-beijing
```

如果没有设置 `MODEL_CHAIN` 或 `MODEL_PRIMARY_NAME`，继续按旧 `ARK_MODEL`、`ARK_MODEL_BACKUP*`、`ARK_TEXT_MODEL`、`ARK_QA_MODEL` 解析。

理由：服务器 `.env` 可以渐进迁移，不需要一次性改运维配置。

### Decision 4: Fallback 全局暂停 key 使用 `provider:model`

旧逻辑只用模型名作为 429 暂停 key。跨 provider 后同名模型可能来自不同平台，必须用 provider 加模型名区分。

理由：避免 SiliconFlow 的 `deepseek-ai/...` 限流误伤 Ark 或 OpenAI-compatible 另一条链路。

### Decision 5: timeout 在兼容层统一设置

不同 Eino 包默认 timeout 不一致。兼容层应支持 `MODEL_TIMEOUT_SECONDS`，未设置时使用现有可接受的默认值，并允许 provider 条目覆盖。

理由：近期长时间卡在模型流式读 body 的故障说明 timeout 必须成为统一配置，而不是落在各 provider SDK 的默认值上。

## Risks / Trade-offs

- [Risk] OpenAI-compatible 平台的 tool-call streaming delta 与 OpenAI 原生不完全一致。→ 继续保留统一 `sanitizeStreamingToolCallDeltas` wrapper，并用 SiliconFlow 构造单测和线上冒烟覆盖。
- [Risk] JSON schema response format 类型在 Ark/OpenAI 包之间不同。→ 兼容层只接受项目内部 schema 表示，在 adapter 内做类型转换。
- [Risk] thinking/reasoning 控制差异导致 ReAct tool call 解析失败。→ 工具调用场景默认禁用 known thinking 模型；Ark/Qwen/OpenAI-compatible 分别映射 provider 字段或 ExtraFields。
- [Risk] 新配置变量过多。→ 第一阶段保留旧配置兜底，并把新配置解析集中在单测覆盖的 parser 中。
- [Risk] 只做 OpenAI-compatible 可能无法发挥 DeepSeek/Qwen 专用能力。→ 首轮目标是切换平台便利性，专用 adapter 作为后续增强，不阻塞主链路。

## Migration Plan

1. 新增兼容层和配置解析测试。
2. 修改 `agentutils.NewFallbackToolCallingChatModel` 使用 `ModelSpec` chain，但保持函数签名不变。
3. 验证旧 `ARK_*` 配置可构造同等 Ark 模型链。
4. 验证 `siliconflow` profile 可转成 OpenAI-compatible 模型配置。
5. 本地跑聚焦后端测试和 `go build ./...`。
6. 若本次进入运行闭环，再按项目部署流程上线并做 `/api/health` 与低成本生成冒烟。

Rollback：保留旧 `ARK_*` 配置分支；若 provider-aware 配置解析异常，可清空 `MODEL_CHAIN` 回到旧 Ark 行为。

## Open Questions

- 账号级 API key 已补充 provider 字段：UI 必须让用户选择厂商，后端只把该 Key 应用于同厂商模型资源，避免跨平台误用。
- SiliconFlow 上实际使用的默认模型名和 timeout 是否需要单独的 UI 配置入口？
- DeepSeek/Qwen 首轮按 OpenAI-compatible profile 创建，并保留后续专用 Eino adapter 优化空间。

## Follow-up Decisions

### Decision 6: 可见正文进入 RuntimeMeta timeline

前端最终展示的是可观测执行轨迹，不是隐藏思维链。后端把已经通过 SSE `answer` 事件发送给用户的可见正文同步记录为 `assistant_output` runtime event；工具调用、命令执行和可见正文共享同一条 RuntimeMeta 顺序号。前端据此按事件顺序穿插渲染，解决工具调用集中在上方、文本集中在下方的问题。

### Decision 7: 账号 Key 配置从“单 Key”升级为“厂商 + Key”

账户设置弹窗要求选择 provider 后保存 Key，并明确引导用户使用自己的 Key。系统默认 Key 只作为兜底提示。模型创建时，账户 Key 只注入 provider 匹配的 `ModelSpec`，不同 provider fallback entry 继续使用各自环境变量。
