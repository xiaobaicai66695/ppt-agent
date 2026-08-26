# PPT Agent 多模型供应商兼容层预研

日期：2026-08-26

## 背景

当前 PPT Agent 的模型创建入口集中在 `ppt-agent/backend/pkg/agent/utils/model.go`。虽然上层 Agent 只依赖 Eino 的 `model.ToolCallingChatModel`，但实际创建逻辑仍硬编码为 Ark：

- `NewFallbackToolCallingChatModel` 只读取 `ARK_MODEL`、`ARK_MODEL_BACKUP1..4`。
- `WithTextModel`、`NewQAModel` 分别绑定 `ARK_TEXT_MODEL`、`ARK_QA_MODEL`。
- `newSingleModel` 固定构造 `ark.ChatModelConfig`，并调用 `ark.NewChatModel`。
- 账号级 API Key 只作为 Ark key 的覆盖值使用。

这会导致用户从 Ark 切到 OpenAI、DeepSeek、Qwen 或硅基流动时，不只是改环境变量，还需要改代码里的 provider 包、配置字段和部分模型特性开关。

## 目标

建立一层稳定的模型兼容层，让业务 Agent 不关心上游平台差异：

```text
PPTPlanner / Reviewer / Fixer / Compressor
        |
agentutils.NewFallbackToolCallingChatModel
        |
modelcompat.Factory
        |
ProviderAdapter
        |
Ark / OpenAI / DeepSeek / Qwen / SiliconFlow / OpenAI-compatible
```

兼容层输出仍然是 `model.ToolCallingChatModel`，保持 Eino ADK Agent、工具调用、压缩器、RuntimeStatus 包装和 fallback 逻辑的上层契约不变。

## Eino / eino-ext 创建方式

本次核查基于当前项目依赖和本地模块缓存：

- `github.com/cloudwego/eino v0.8.8`
- `github.com/cloudwego/eino-ext/components/model/ark v0.1.65`
- `github.com/cloudwego/eino-ext/components/model/openai v0.1.12`
- `github.com/cloudwego/eino-ext/components/model/deepseek v0.1.2`
- `github.com/cloudwego/eino-ext/components/model/qwen v0.1.9`

| 平台 | 推荐 Eino 包 | 创建函数 | 关键配置 | 备注 |
| --- | --- | --- | --- | --- |
| OpenAI | `github.com/cloudwego/eino-ext/components/model/openai` | `openai.NewChatModel(ctx, *openai.ChatModelConfig)` | `APIKey`、`Model`、`BaseURL`、`Timeout`、`MaxTokens`、`MaxCompletionTokens`、`Temperature`、`TopP`、`ResponseFormat`、`ReasoningEffort`、`ExtraFields` | 适合作为 OpenAI 原生和大部分 OpenAI-compatible provider 的通用适配器。 |
| DeepSeek | `github.com/cloudwego/eino-ext/components/model/deepseek` | `deepseek.NewChatModel(ctx, *deepseek.ChatModelConfig)` | `APIKey`、`Model`、`BaseURL`、`Path`、`Timeout`、`MaxTokens`、`Temperature`、`TopP`、`ResponseFormatType` | 官方 API 兼容 OpenAI 格式，也有独立 Eino 包。独立包更适合处理 DeepSeek reasoning/content 差异。 |
| Qwen / DashScope | `github.com/cloudwego/eino-ext/components/model/qwen` | `qwen.NewChatModel(ctx, *qwen.ChatModelConfig)` | `APIKey`、`Model`、`BaseURL`、`Timeout`、`MaxTokens`、`Temperature`、`TopP`、`ResponseFormat`、`EnableThinking` | DashScope 提供 OpenAI-compatible endpoint；独立包支持 `EnableThinking`，且有 Qwen tool-call 流式细节处理。 |
| 硅基流动 SiliconFlow | 当前未发现独立 `eino-ext/components/model/siliconflow` 版本 | 使用 `openai.NewChatModel` | `APIKey`、`Model`、`BaseURL=https://api.siliconflow.cn/v1`、OpenAI-compatible 参数 | `go list -m -json github.com/cloudwego/eino-ext/components/model/siliconflow@latest` 返回 no matching versions。应作为 OpenAI-compatible profile 接入。 |
| Ark / 火山方舟 | `github.com/cloudwego/eino-ext/components/model/ark` | `ark.NewChatModel(ctx, *ark.ChatModelConfig)` | `APIKey`、`BaseURL`、`Region`、`Model`、`MaxTokens`、`MaxCompletionTokens`、`Temperature`、`TopP`、`Thinking`、`ResponseFormat`、`RetryTimes` | 当前项目已使用。Ark 的 endpoint/model id、thinking 和 JSON schema 配置类型与 OpenAI 包不同。 |

## 官方协议侧结论

- OpenAI Chat API 是 OpenAI 原生入口；Eino OpenAI 包也可配置 `BaseURL`。
- DeepSeek 官方文档说明其 API 兼容 OpenAI/Anthropic，OpenAI base URL 为 `https://api.deepseek.com`。
- 阿里云百炼/DashScope 官方文档提供 OpenAI-compatible endpoint，地址以 `/compatible-mode/v1` 结尾，且地域和业务空间会影响 base URL。
- SiliconFlow Chat Completions API 使用 `https://api.siliconflow.cn/v1/chat/completions`，因此给 OpenAI SDK/Eino OpenAI 包的 base URL 应是 `https://api.siliconflow.cn/v1`。
- 火山方舟既有 Eino Ark 专用包，也有 OpenAI-compatible API；本项目当前已走 Ark 专用包，短期应继续保留以减少迁移风险。

## 建议的兼容层设计

新增包建议放在：

```text
ppt-agent/backend/pkg/agent/modelcompat
```

核心数据结构：

```go
type Provider string

const (
    ProviderArk          Provider = "ark"
    ProviderOpenAI       Provider = "openai"
    ProviderDeepSeek     Provider = "deepseek"
    ProviderQwen         Provider = "qwen"
    ProviderSiliconFlow  Provider = "siliconflow"
    ProviderOpenAICompat Provider = "openai_compatible"
)

type ModelSpec struct {
    Provider     Provider
    Model        string
    APIKey       string
    BaseURL      string
    Region       string
    Timeout      time.Duration
    MaxTokens    *int
    Temperature  *float32
    TopP         *float32
    DisableThink *bool
    JSONSchema   *openai.ChatCompletionResponseFormatJSONSchema
    Extra        map[string]any
}

type Factory interface {
    NewToolCallingChatModel(ctx context.Context, spec ModelSpec) (model.ToolCallingChatModel, error)
}
```

适配器职责：

- `arkAdapter`：只负责把 `ModelSpec` 转成 `ark.ChatModelConfig`，包含 Ark `Thinking` 和 Ark JSON schema 类型转换。
- `openAIAdapter`：负责 OpenAI 原生和 OpenAI-compatible provider，按 `BaseURL`、`ExtraFields` 透传差异参数。
- `deepSeekAdapter`：优先用独立 `deepseek` 包，保留 `Path`、`ResponseFormatType` 和 reasoning 处理余地。
- `qwenAdapter`：优先用独立 `qwen` 包，显式控制 `EnableThinking`。
- `siliconFlowProfile`：不是独立 adapter，而是 `openAIAdapter` 的 profile，默认 base URL 为 `https://api.siliconflow.cn/v1`，API key 读取 `SILICONFLOW_API_KEY`。

## 配置建议

为兼容旧部署，第一阶段不删除 `ARK_*`：

1. 新配置优先：

```env
MODEL_PROVIDER=ark
MODEL_CHAIN=primary,backup1,backup2

MODEL_PRIMARY_PROVIDER=ark
MODEL_PRIMARY_NAME=...
MODEL_PRIMARY_API_KEY_ENV=ARK_API_KEY
MODEL_PRIMARY_BASE_URL=...
MODEL_PRIMARY_REGION=cn-beijing

MODEL_BACKUP1_PROVIDER=siliconflow
MODEL_BACKUP1_NAME=deepseek-ai/DeepSeek-V4-Flash
MODEL_BACKUP1_API_KEY_ENV=SILICONFLOW_API_KEY
MODEL_BACKUP1_BASE_URL=https://api.siliconflow.cn/v1
```

2. 旧配置兜底：

```env
ARK_MODEL=...
ARK_MODEL_BACKUP1=...
ARK_TEXT_MODEL=...
ARK_QA_MODEL=...
ARK_API_KEY=...
ARK_BASE_URL=...
ARK_REGION=...
```

3. 账号级 key 扩展：

当前 `WithAPIKey` 只有 key，没有 provider 语义。后续如果用户在界面保存多个平台 key，应让任务配置携带：

```go
ModelProvider string
ModelAPIKey   string
ModelBaseURL  string
```

短期可以先沿用 `ModelAPIKey` 覆盖当前选中 provider 的 API key，避免一次性扩数据库模型。

## 分阶段实施路线

第一阶段建议走 `opsx`，因为它涉及后端配置契约、模型 fallback 行为、账号 key 语义和上线验证。

1. 抽出 `modelcompat` 包，只实现 `ark` 和 `openai_compatible` 两条路径。
2. 把 SiliconFlow 作为内置 OpenAI-compatible profile 接入。
3. 保留 `ARK_*` 完全兼容旧环境；新增配置只在显式设置时生效。
4. 将 `NewFallbackToolCallingChatModel` 从 `[]string` 模型名升级为 `[]ModelSpec`。
5. 补充 provider 配置解析单测、Ark 旧配置回归单测、SiliconFlow/OpenAI-compatible 构造参数单测。
6. 再按需要引入 `deepseek`、`qwen` 专用包，解决 thinking、response_format 或 tool-call 流式差异。

## 风险点

- 不同 Eino 包的 `ResponseFormat` 类型不同，不能直接复用当前 Ark JSON schema 转换。
- Thinking/reasoning 控制分散在 Ark `Thinking`、Qwen `EnableThinking`、OpenAI `ReasoningEffort/ExtraFields`、DeepSeek reasoning content 中，需要在工具调用场景默认关闭或降级。
- 流式 tool call delta 在不同 OpenAI-compatible provider 上可能有细节差异，现有 `sanitizeStreamingToolCallDeltas` 应保留在统一 wrapper 层。
- Timeout 默认值不同：Ark 默认 10 分钟，DeepSeek 默认 5 分钟，OpenAI/Qwen 默认无 timeout；兼容层需要统一设置，避免再次出现长时间读 body 卡死。
- Fallback 的全局暂停 key 不能只用模型名；跨 provider 时应改成 `provider:model`，否则同名模型会互相影响。

## 结论

不要让业务 Agent 直接认识 Ark/OpenAI/DeepSeek/Qwen/SiliconFlow。下一步应先把模型创建收敛为：

```text
Agent -> NewFallbackToolCallingChatModel -> modelcompat.ModelSpec chain -> provider adapter -> Eino ToolCallingChatModel
```

首轮实现建议只强制完成 `ark` 与 `openai_compatible`，把 SiliconFlow 纳入 profile。DeepSeek 和 Qwen 可以先通过 OpenAI-compatible 跑通，再在发现 thinking、tool-call 或 response_format 差异影响稳定性时切到专用 Eino 包。

## 参考来源

- 本地 Eino ext 源码：`D:/tools/gopath/pkg/mod/github.com/cloudwego/eino-ext/components/model/ark@v0.1.65/chatmodel.go`
- 本地 Eino ext 源码：`D:/tools/gopath/pkg/mod/github.com/cloudwego/eino-ext/components/model/openai@v0.1.12/chatmodel.go`
- 本地 Eino ext 源码：`D:/tools/gopath/pkg/mod/github.com/cloudwego/eino-ext/components/model/deepseek@v0.1.2/deepseek.go`
- 本地 Eino ext 源码：`D:/tools/gopath/pkg/mod/github.com/cloudwego/eino-ext/components/model/qwen@v0.1.9/chatmodel.go`
- OpenAI Chat API：https://platform.openai.com/docs/api-reference/chat/create
- DeepSeek API Docs：https://api-docs.deepseek.com/
- 阿里云百炼 OpenAI-compatible：https://help.aliyun.com/zh/model-studio/compatibility-of-openai-with-dashscope
- SiliconFlow Chat Completions：https://docs.siliconflow.cn/cn/api-reference/chat-completions/chat-completions
- 火山方舟 OpenAI-compatible：https://docs.volcengine.com/docs/82379/1330626?lang=zh
