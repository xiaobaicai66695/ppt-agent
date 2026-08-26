# 2026-08-26 多模型供应商兼容层

## 目标

将 PPT Agent 的模型创建从硬编码 Ark 调整为 provider-aware 兼容层，后续用户切换 OpenAI、硅基流动等 OpenAI-compatible 平台时不再需要改业务 Agent 代码。

## 本次实现

- 新增 `ppt-agent/backend/pkg/agent/modelcompat`：
  - 定义 `Provider`、`ModelSpec`、`DefaultFactory`。
  - 支持 Ark 专用 adapter。
  - 支持 OpenAI/OpenAI-compatible adapter。
  - 将 SiliconFlow 作为 OpenAI-compatible profile，默认 base URL 为 `https://api.siliconflow.cn/v1`。
  - 统一 provider display name 和 fallback tracker key。
- 调整 `ppt-agent/backend/pkg/agent/utils/model.go`：
  - `NewFallbackToolCallingChatModel` 改为解析 `[]modelcompat.ModelSpec`。
  - 支持 `MODEL_CHAIN` 与 `MODEL_<ENTRY>_*` provider-aware 配置。
  - 保留旧 `ARK_MODEL`、`ARK_MODEL_BACKUP*`、`ARK_TEXT_MODEL`、`ARK_QA_MODEL` 兜底。
  - `WithTextModel` 支持 `MODEL_TEXT_*`，QA 支持 `MODEL_QA_*`。
  - fallback 全局限流 key 改为优先使用 `provider:model`，避免跨平台同名模型互相暂停。
  - 保留 stream read fallback、tool-call delta sanitize、压缩器和 RuntimeStatus wrapper。

## 新配置示例

```env
MODEL_CHAIN=primary,backup1

MODEL_PRIMARY_PROVIDER=ark
MODEL_PRIMARY_NAME=...
MODEL_PRIMARY_API_KEY_ENV=ARK_API_KEY
MODEL_PRIMARY_BASE_URL=...
MODEL_PRIMARY_REGION=cn-beijing

MODEL_BACKUP1_PROVIDER=siliconflow
MODEL_BACKUP1_NAME=deepseek-ai/DeepSeek-V4-Flash
MODEL_BACKUP1_API_KEY_ENV=SILICONFLOW_API_KEY
MODEL_BACKUP1_BASE_URL=https://api.siliconflow.cn/v1

MODEL_TIMEOUT_SECONDS=600
```

如果不配置 `MODEL_CHAIN` 或 `MODEL_PRIMARY_NAME`，运行行为回落到旧 Ark 配置。

## 验证

- `go test ./pkg/agent/modelcompat ./pkg/agent/utils`
- `go test ./pkg/agent/modelcompat ./pkg/agent/utils ./pkg/agent/deck`
- `go build ./...`

## 遗留

- DeepSeek/Qwen 专用 Eino adapter 先保留扩展点，当前首轮不引入新依赖。
- 账号级多 provider key 的前端/数据库模型暂未扩展；短期仍由任务选中的 provider 消费当前 `ModelAPIKey`。
- 本地实现完成后仍需按运行变更闭环部署并执行线上冒烟。
