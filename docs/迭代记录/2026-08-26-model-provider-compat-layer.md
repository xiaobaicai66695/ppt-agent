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
- 调整任务并发限制：
  - 移除 `TaskManager.CreateTask` 中“同一用户已有 running 任务就拒绝创建”的全局任务闸门。
  - 在 `FallbackChatModel.Generate` 和 `Stream` 内按 `provider:model:key-hash` 获取模型调用 slot。
  - 默认同一上游 API 资源一次只跑 1 个模型调用；不同 provider/model/API key 不互相阻塞。
  - 可通过 `MODEL_API_CONCURRENCY` 调整每个上游 API 资源的并发度；兼容 `MODEL_RESOURCE_CONCURRENCY`。

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

## 本地验证

- `go test ./pkg/agent/modelcompat ./pkg/agent/utils`
- `go test ./pkg/agent/modelcompat ./pkg/agent/utils ./pkg/task`
- `go test ./pkg/agent/modelcompat ./pkg/agent/utils ./pkg/agent/deck`
- `go test ./pkg/agent/modelcompat ./pkg/agent/utils ./pkg/task ./pkg/agent/deck ./pkg/web`
- `go build ./...`

## 上线记录

- 提交：`07668d0 feat: add model provider compat layer`
- 目标：`remote-dev:/ppt/ppt-agent`
- 时间：2026-08-26 17:41 Asia/Shanghai
- 新进程：PID `3502269`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
- 旧二进制备份：`/ppt/ppt-agent/ppt-agent-linux.bak.20260826174103`
- 启动确认：
  - `:8080` 正常监听
  - `/api/health` 返回 `{"status":"ok"}`
  - `/api/templates/layouts` 返回 17469 bytes
  - `/api/themes` 返回 2614 bytes
- 模型链冒烟：
  - `/api/ai/expand` 能触发模型链初始化，日志出现 `model_init_success model=ark/...`，说明兼容层按旧 Ark 配置成功创建 provider-aware fallback chain。
  - 上游模型调用返回 HTTP `402`，接口返回 500；该阻塞来自上游账户/计费侧，不记录凭据。完整 LLM 生成冒烟未通过，不能归档为完全 done。

## 并发修正上线记录

- 提交：`6974b8e feat: limit model concurrency by upstream resource`
- 目标：`remote-dev:/ppt/ppt-agent`
- 时间：2026-08-26 17:55 Asia/Shanghai
- 新进程：PID `3506041`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
- 旧二进制备份：`/ppt/ppt-agent/ppt-agent-linux.bak.20260826175547`
- 启动确认：
  - `:8080` 正常监听
  - `/api/health` 返回 `{"status":"ok"}`
  - `/api/templates/layouts` 和 `/api/themes` 静态接口正常
- 行为变化：
  - 任务创建入口不再因为同一用户已有 running 任务而直接返回 409。
  - 模型调用默认按 `provider:model:key-hash` 串行，避免同一上游 API 资源并发过高触发 429。
  - 不同 provider、不同模型或不同 API key 的任务不再被全局任务闸门互相阻塞。
- 模型链冒烟：
  - `/api/ai/expand` 能触发模型链初始化，日志出现 `model_init_success model=ark/...`。
  - 上游仍返回 HTTP `402`，完整 LLM 生成冒烟继续受账户/计费侧阻塞。

## 账号 Key 与执行轨迹修正

- 代码范围：
  - 后端：
    - `ppt-agent/backend/pkg/agent/modelcompat/modelcompat.go`
    - `ppt-agent/backend/pkg/agent/utils/model.go`
    - `ppt-agent/backend/pkg/agent/utils/runtime_meta.go`
    - `ppt-agent/backend/pkg/task/manager.go`
    - `ppt-agent/backend/pkg/web/handler.go`
    - `ppt-agent/backend/pkg/agent/deck/{types.go,agent.go}`
  - 前端：
    - `ppt-agent/frontend/src/components/AccountSettingsDialog.vue`
    - `ppt-agent/frontend/src/components/ConversationComposer.vue`
    - `ppt-agent/frontend/src/pages/DashboardPage.vue`
    - `ppt-agent/frontend/src/utils/workbench.ts`
    - `ppt-agent/frontend/src/api.ts`
- 行为变化：
  - 账户 API Key 配置从单 Key 升级为“厂商 + Key”，支持 Ark、OpenAI、硅基流动、DeepSeek、Qwen、OpenAI-compatible。
  - 账户 Key 只注入同 provider 的模型资源，避免把一个厂商的 Key 误传给其他上游。
  - 账户设置弹窗明确引导用户配置自己的厂商 Key，系统默认 Key 只作为共享兜底提示。
  - DeepSeek 和 Qwen 先作为 OpenAI-compatible profile 接入，默认 base URL 分别为 `https://api.deepseek.com` 和 `https://dashscope.aliyuncs.com/compatible-mode/v1`，可通过环境变量覆盖。
  - 已通过 SSE 展示给用户的可见 AI 正文会进入 RuntimeMeta `assistant_output` 事件；前端按 runtime event 顺序穿插渲染 AI 正文和工具调用卡片，不展示隐藏 chain-of-thought。
- 本地验证：
  - `npm run test -- src/utils/workbench.test.ts`
  - `npm run build`
  - `go test ./pkg/agent/modelcompat ./pkg/agent/utils ./pkg/task ./pkg/web`
  - `go build ./...`
- 上线记录：
  - 提交：`2c4944e feat: add provider key UI and ordered runtime trace`
  - 目标：`remote-dev:/ppt/ppt-agent`
  - 时间：2026-08-26 19:31 Asia/Shanghai
  - 新进程：PID `3529816`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
  - 旧二进制备份：`/ppt/ppt-agent/ppt-agent-linux.bak.20260826193106`
  - 前端备份：`/ppt/ppt-agent/frontend/dist.bak.20260826193106`
  - 启动确认：
    - `:8080` 正常监听
    - `/api/health` 返回 `{"status":"ok"}`
    - `/` 返回 HTTP 200，709 bytes
    - `/api/templates/layouts` 返回 HTTP 200，17469 bytes
    - `/api/themes` 返回 HTTP 200，2614 bytes
    - 未登录访问 `/api/users/me/api-key` 返回 HTTP 401，鉴权保护正常
  - 清理：远端 `/tmp/ppt-agent-linux-2c4944e`、`/tmp/ppt-agent-frontend-dist-2c4944e.tar` 和冒烟临时响应文件已删除。
  - 遗留：完整 LLM 生成冒烟仍受上游 HTTP `402`/计费侧阻塞，未执行高成本生成任务。

## 遗留

- DeepSeek/Qwen 专用 Eino adapter 仍保留扩展点，当前先走 OpenAI-compatible profile。
- 账号级 key 当前仍是一组 provider/key；多 provider key 并存可作为后续增强。
- 线上完整 LLM 冒烟需在上游 HTTP `402` 解决后复测。

## 硅基流动超时诊断与上下文角色修正

- 代码范围：
  - 后端：
    - `ppt-agent/backend/pkg/agent/utils/model.go`
    - `ppt-agent/backend/pkg/agent/utils/model_test.go`
  - 前端：
    - `ppt-agent/frontend/src/components/RuntimeEventDetail.vue`
    - `ppt-agent/frontend/src/components/AccountSettingsDialog.vue`
    - `ppt-agent/frontend/src/pages/DashboardPage.vue`
    - `ppt-agent/frontend/src/utils/workbench.ts`
  - OpenSpec：
    - `openspec/changes/add-model-provider-compat-layer/tasks.md`
    - `openspec/changes/add-model-provider-compat-layer/specs/model-provider-compatibility/spec.md`
    - `openspec/changes/add-model-provider-compat-layer/specs/ppt-agent-runtime-harness/spec.md`
- 行为变化：
  - `FallbackChatModel.Generate/Stream` 在真正调用上游前记录 `model_request` 事件，包含 provider、model、timeout、message_count、role_counts、system/user 摘要、tool 名称和最近上下文 history。
  - 流式读取阶段如果出现 `context deadline exceeded ... while reading body`，会补充记录带 provider/model/timeout 的 `llm_error`，便于区分平台长读超时和任务流超时。
  - 前端运行详情不再从“模型上下文”中过滤 `tool` role，并新增“上下文概览”展示 system/user/tool 等可观测摘要；system 原文仍不放入 history。
  - 新增 provider 级 timeout 配置：`MODEL_SILICONFLOW_TIMEOUT_SECONDS=900` 可单独拉长硅基流动，不影响 Ark/OpenAI/DeepSeek/Qwen。
  - 账户 Key 弹窗进一步提示用户先配置自己的厂商 Key，系统兜底 Key 只作为临时共享额度。
- 本地验证：
  - `go test ./pkg/agent/modelcompat ./pkg/agent/utils ./pkg/task ./pkg/web`
  - `go build ./...`
  - `npm run test -- src/utils/workbench.test.ts`
  - `npm run build`
  - `openspec validate add-model-provider-compat-layer --strict`
- 上线记录：
  - 提交：`feat: record provider model request context`
  - 目标：`remote-dev:/ppt/ppt-agent`
  - 时间：2026-08-26 21:28 Asia/Shanghai
  - 新进程：PID `3559464`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
  - 远端配置：`MODEL_SILICONFLOW_TIMEOUT_SECONDS=900`，未记录任何 Key 或凭据。
  - 启动确认：
    - `/api/health` 返回 HTTP 200，`{"status":"ok"}`
    - `/` 返回 HTTP 200，709 bytes
    - `/api/templates/layouts` 返回 HTTP 200，17469 bytes
    - `/api/themes` 返回 HTTP 200，2614 bytes
    - 未登录访问 `/api/users/me/api-key` 返回 HTTP 401，鉴权保护正常
  - 清理：远端 `/tmp/ppt-agent-linux-modelctx`、`/tmp/ppt-agent-frontend-dist-modelctx.tar` 和冒烟临时响应文件已删除。
  - 说明：未主动发起完整 LLM 生成任务，避免继续消耗共享/平台 Key；后续用户使用自备 Key 跑任务时，若硅基流动仍在长读阶段断开，运行详情会显示具体 provider/model/timeout 和真实消息角色摘要。

## 账户 API Key 保存 1292 修复

- 问题：
  - 用户在账户设置中保存 DeepSeek 等 provider Key 时，MySQL 返回 `Error 1292 (22007): Incorrect datetime value: '0000-00-00' for column 'created_at'`。
  - 根因是 `UpsertUserAPIKey` 使用 `DB.Save` 保存主键结构体，但没有填 `CreatedAt`，在 MySQL 严格模式下会把 Go 零值时间写成非法日期。
- 行为变化：
  - `UpsertUserAPIKey` 改为 `ON CONFLICT/ON DUPLICATE KEY` 风格 upsert。
  - 插入时显式写入 `created_at` 与 `updated_at`。
  - 已存在记录更新时只更新 `provider`、`api_key`、`updated_at`，不覆盖 `created_at`。
- 本地验证：
  - `go test ./pkg/db ./pkg/web`
  - `go build ./...`
- 上线记录：
  - 目标：`remote-dev:/ppt/ppt-agent`
  - 时间：2026-08-26 23:28 Asia/Shanghai
  - 新进程：PID `3587786`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
  - 启动确认：`/api/health` 返回 HTTP 200。
  - 线上接口复测：
    - 使用临时 smoke token 调用 `PUT /api/users/me/api-key`，provider=`deepseek`，假 Key，返回 HTTP 200，`configured=true`，仅返回 masked key。
    - 随后调用 `DELETE /api/users/me/api-key` 返回 HTTP 200，临时记录已清理。
  - 清理：远端 `/tmp/ppt-agent-linux-apikeyfix` 和冒烟临时响应文件已删除。

## DeepSeek 账号 Key provider 路由 401 修复

- 问题：
  - 用户保存的账号 Key 已标记为 `provider=deepseek`，但 Web 入口的部分模型工厂仍只传递裸 API Key。
  - 线上 legacy `ARK_MODEL` 配置仍存在时，裸 Key 会进入 Ark adapter，导致 DeepSeek Key 被误发给 Ark 上游并返回 HTTP `401`。
- 行为变化：
  - Web 入口统一解析账号凭据为 `provider + api_key`，`AgentFactory`、`AIModelFactory`、`TextModelFactory` 都使用 `WithAPIKeyForProvider`。
  - 非 Ark 账号 Key 在未配置 `MODEL_CHAIN` 的 legacy Ark 环境下，会自动插入对应 provider 的首选 `ModelSpec`；DeepSeek 默认模型为 `deepseek-chat`。
  - legacy Ark fallback 不再消费非 Ark 账号 Key，只使用 Ark 环境 Key。
- 本地验证：
  - `go test ./pkg/agent/utils ./pkg/agent/modelcompat ./pkg/db ./pkg/web`
  - `go test ./pkg/agent/deck`
  - `go build ./...`
- 上线记录：
  - 目标：`remote-dev:/ppt/ppt-agent`
  - 时间：2026-08-26 23:44 Asia/Shanghai
  - 新进程：PID `3591788`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
  - 启动确认：`/api/health` 返回 HTTP 200，`{"status":"ok"}`。
  - 线上账号状态：user_id=1 的账号 Key 返回 `configured=true`、`provider=deepseek`，仅检查脱敏状态。
  - 线上 DeepSeek 冒烟：
    - 创建 1 页临时任务 `4a3b7101-7eab-4912-a14a-56084aab83f6`。
    - 日志显示模型初始化为 `deepseek/deepseek-chat(backup-0)`，不再是 `ark/deepseek-ai/...`。
    - 任务终态 `completed`，交付 `1/1`。
    - 通过 `DELETE /api/tasks/{id}` 清理临时任务，返回 HTTP 200。
  - 清理：远端 `/tmp/ppt-agent-linux-providerkey` 已不再作为运行文件使用；本地 Linux 构建产物仅用于传输，不纳入提交。
