## Why

PPT Agent 当前模型创建逻辑硬编码 Ark provider 和 `ARK_*` 环境变量，用户切换到 OpenAI、DeepSeek、Qwen 或硅基流动时需要修改后端代码。近期硅基流动流式读 body 超时也暴露出 provider 差异没有隔离，fallback、timeout、thinking 和 tool-call 流式处理需要收敛到统一兼容层。

## What Changes

- 新增模型供应商兼容层，把上层 Agent 使用的 Eino `model.ToolCallingChatModel` 与具体 provider 构造方式解耦。
- 支持按 provider 构造 Ark、OpenAI-compatible、SiliconFlow 模型，并保留后续接入 DeepSeek/Qwen 专用 Eino 包的扩展点。
- 将 fallback 链从纯模型名升级为带 provider 的模型条目，同时保留旧 `ARK_MODEL`、`ARK_MODEL_BACKUP*`、`ARK_TEXT_MODEL`、`ARK_QA_MODEL` 配置兜底。
- 统一模型创建时的 timeout、API key 解析、thinking 禁用和日志显示名，避免跨 provider 行为漂移。
- 不改变 PPTPlanner、Reviewer、Fixer、Compressor 对模型的上层接口。

## Capabilities

### New Capabilities

- `model-provider-compatibility`: Defines how PPT Agent resolves provider-aware model configuration and creates Eino tool-calling chat models through a compatibility layer.

### Modified Capabilities

- `ppt-agent-runtime-harness`: Runtime model fallback and stream wrappers must preserve existing trace, compression, fallback and tool-call behavior after provider abstraction.

## Impact

- Affected backend code: `ppt-agent/backend/pkg/agent/utils/model.go`, new `ppt-agent/backend/pkg/agent/modelcompat`, and focused tests.
- Affected configuration: new provider-aware model env vars, with legacy `ARK_*` compatibility retained.
- Affected runtime behavior: model init, fallback model names, provider-specific API key/base URL/timeout handling, thinking controls.
- No frontend API or PPT generator contract change is expected in the first phase.
