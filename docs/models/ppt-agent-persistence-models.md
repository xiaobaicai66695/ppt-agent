# PPT Agent 持久化模型与结构体说明

本文覆盖 `ppt-agent/backend/pkg/db` 的全部持久化模型，以及面向管理端的聚合结构体。它按运行时的使用链路组织，而不是按 Go 文件名罗列。

实现来源：

- [`internal/model/`](../../ppt-agent/backend/pkg/db/internal/model/)：GORM 持久化模型。
- [`models.go`](../../ppt-agent/backend/pkg/db/models.go)：对外 `db.*` 类型别名；调用方无需导入 `internal` 包。
- [`admin.go`](../../ppt-agent/backend/pkg/db/admin.go)：管理端只读聚合结构体。

## 总览与关系

```text
账号与认证
User ── 1:1 ── UserAPIKey
 │
 ├── 1:N ── VerificationCode（按 Email 校验，无数据库外键）
 ├── 1:N ── TaskRecord ── 1:N ── ConversationMessage ── 1:N ── ConversationMessageChunk
 │                 │  └── 1:N ── RuntimeEventRecord
 │                 │  └── 1:N ── TaskErrorAnalysis
 │                 │  └── 1:1 ── TaskFeedback（与 UserID 联合唯一）
 │                 └── 任务可关联 ConversationID / ParentTaskID
 └── 1:N ── PlanDraftRecord（规划完成、尚未渲染）
```

除 `ConversationMessage.Chunks` 的 GORM 关联外，其余关系目前主要通过业务字段和索引表达，不依赖数据库级外键。因此删除或迁移记录时须由对应仓储操作维护关联数据。

## 1. 账号、认证与模型配置

### `User`（默认表：`users`）

注册账号、管理员身份和游客身份的统一主体。认证中间件、任务归属和管理端指标都以该模型为入口。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `uint`，主键 | 用户唯一标识；被任务、反馈和 API Key 使用。 |
| `Email` | `string`，最长 120、唯一索引、非空 | 注册邮箱；游客使用 `guest-…@guest.local` 格式。 |
| `Password` | `string`，最长 255，不输出 JSON | bcrypt 密码哈希；验证码登录流程不会返回该字段。 |
| `IsAdmin` | `bool`，默认 `false` | 管理端权限标记。 |
| `GuestIPHash` | `string`，最长 64、索引，不输出 JSON | 游客 IP 的单向摘要，用于匿名活跃用户去重；不保存原始 IP。 |
| `CreatedAt` | `time.Time` | 账号创建时间。 |

### `VerificationCode`（默认表：`verification_codes`）

邮箱注册、登录和密码相关流程使用的一次性验证码记录。查询通常按 `Email`、`Code`、`Used` 与有效期组合完成。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `uint`，主键，不输出 JSON | 验证码记录标识。 |
| `Email` | `string`，最长 120、索引、非空 | 验证码归属邮箱。 |
| `Code` | `string`，最长 10、非空，不输出 JSON | 验证码明文；仅供校验逻辑读取。 |
| `Used` | `bool`，默认 `false`，不输出 JSON | 是否已被成功消费，防止复用。 |
| `ExpiresAt` | `time.Time`，非空 | 失效时间。 |
| `CreatedAt` | `time.Time` | 签发时间。 |

### `UserAPIKey`（默认表：`user_api_keys`）

用户级模型供应商凭据覆盖。记录不存在表示使用服务端环境配置；空值不应长期保留为独立记录。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `UserID` | `uint`，主键 | 与 `User.ID` 一对一对应。 |
| `Provider` | `string`，最长 40、非空，默认 `ark` | 模型供应商稳定标识。 |
| `APIKey` | `string`，`text`、非空，不输出 JSON | 用户提供的密钥；管理端和普通 API 均不得回显。 |
| `CreatedAt` | `time.Time` | 首次设置时间。 |
| `UpdatedAt` | `time.Time` | 最近一次更新凭据或供应商的时间。 |

## 2. 任务创建、渲染与交付

### `TaskRecord`（默认表：`task_records`）

任务生命周期的核心持久化模型：对话任务可以原地提升为 PPT 生成任务，任务管理器可用其冷加载恢复状态。`Intent` 区分创建 PPT 与仅会话等场景，管理端 PPT 统计只计入 `create`。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `string`，最长 64、主键 | 任务唯一 ID，也是会话消息、运行事件和错误分析的关联键。 |
| `UserID` | `uint`，索引、非空 | 创建者 ID，用于鉴权和用户任务列表。 |
| `Query` | `string`，`text` | 用户原始请求。 |
| `Status` | `string`，最长 20、非空、默认 `running` | 任务状态；服务启动会将遗留 `running` 任务标记为失败。 |
| `WorkDir` | `string`，最长 512 | 任务工作目录。 |
| `DoneCount` / `TotalCount` | `int`，默认 0 | 已完成页数与计划总页数，用于进度展示。 |
| `Duration` | `string`，最长 50 | 人类可读的任务耗时。 |
| `Error` | `string`，`text` | 最终失败或中断原因。 |
| `PromptTokens` / `CompletionTokens` / `TotalTokens` | `int64`，默认 0 | 模型调用 token 使用量。 |
| `Files` | `string`，`text` | 交付文件元数据的序列化内容。 |
| `ConversationContent` | `string`，`longtext` | 历史兼容的拼接会话正文，用于恢复。 |
| `FullAnswer` | `string`，`longtext` | 拼接后的完整 LLM 回答，用于冷加载恢复。 |
| `Intent` | `string`，最长 32、索引 | 任务意图，例如 `create`；决定后续执行及统计归类。 |
| `ConversationID` | `string`，最长 64、索引 | 所属逻辑会话 ID。 |
| `SourceMessageID` | `string`，最长 64、索引 | 触发本任务的用户消息 ID。 |
| `ParentTaskID` | `string`，最长 64、索引 | 衍生或续办任务的父任务 ID。 |
| `GenerationStartedAt` | `*time.Time`，索引 | 实际开始渲染的时间；未开始时为 `nil`。 |
| `GenerationFinishedAt` | `*time.Time` | 渲染结束时间。 |
| `GenerationDurationMS` | `int64`，默认 0 | 渲染阶段的精确毫秒耗时。 |
| `FixerRunCount` | `int`，默认 0 | PPT Fixer 的执行次数。 |
| `CreatedAt` / `UpdatedAt` | `time.Time` | 创建与最近持久化更新时间。 |

### `TaskFeedback`（默认表：`task_feedbacks`）

用户对已交付 PPT 的评分与建议。`TaskID` 和 `UserID` 的联合唯一索引保证一个用户对一项任务最多保存一条反馈。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `uint`，主键 | 反馈记录 ID。 |
| `TaskID` | `string`，最长 64、索引、非空、联合唯一 | 被评价的任务。 |
| `UserID` | `uint`，索引、非空、联合唯一 | 反馈提交者；必须是任务拥有者。 |
| `Rating` | `int`，非空 | 用户评分，取值范围由 Web/API 层校验。 |
| `Suggestion` | `string`，`text` | 可选改进建议。 |
| `CreatedAt` / `UpdatedAt` | `time.Time` | 初次提交与最后更新的时间。 |

## 3. 会话、规划与任务启动

### `ConversationMessage`（默认表：`conversation_messages`）

单个任务的有序对话历史。较长正文会拆分到 `ConversationMessageChunk`，读取时仓储层按序拼回 `Content`。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `uint`，主键 | 消息 ID；分块记录通过它关联。 |
| `TaskID` | `string`，最长 64、索引、非空 | 所属任务 ID。 |
| `Role` | `string`，最长 20、非空 | 消息角色，当前约定为 `user` 或 `assistant`。 |
| `Content` | `string`，`longtext`、非空 | 短消息正文；长消息拆分后该字段存空串，读取时恢复为完整正文。 |
| `Timestamp` | `time.Time` | 对话排序时间。 |
| `Chunks` | `[]ConversationMessageChunk`，不输出 JSON | GORM 一对多关联，仅供持久化层装配。 |

### `ConversationMessageChunk`（默认表：`conversation_message_chunks`）

长对话消息的 UTF-8 安全分块，避免单行或单字段接近 MySQL 行长度限制。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `uint`，主键 | 分块记录 ID。 |
| `MessageID` | `uint`，索引、非空、与 `Sequence` 联合唯一 | 所属 `ConversationMessage.ID`。 |
| `Sequence` | `int`，非空、与 `MessageID` 联合唯一 | 分块顺序，从 0 开始。 |
| `Content` | `string`，`longtext`、非空 | 本分块正文。 |

### `PlanDraftRecord`（默认表：`plan_draft_records`）

Planner 的结果落库，但尚未触发 PPT 渲染时使用。它支持用户查看、确认或继续编辑规划，而不创建重复的渲染任务。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `string`，最长 64、主键 | 规划草稿 ID。 |
| `UserID` | `uint`，索引、非空 | 草稿拥有者。 |
| `ConversationID` | `string`，最长 64、索引 | 所属逻辑会话。 |
| `SourceMessageID` | `string`，最长 64、索引 | 生成此规划的来源消息。 |
| `Query` | `string`，`text` | 原始用户请求。 |
| `NormalizedRequest` | `string`，`text` | Planner 规范化后的可执行请求。 |
| `DraftContent` | `string`，`longtext` | 规划正文或结构化草稿内容。 |
| `Status` | `string`，最长 32、索引、非空、默认 `draft` | 草稿状态。 |
| `CreatedAt` / `UpdatedAt` | `time.Time` | 创建与修改时间。 |

## 4. 运行追踪、失败分析与恢复

### `RuntimeEventRecord`（默认表：`runtime_event_records`）

任务 Timeline 的可回放事件。完整元数据供受权限保护的详情接口读取，列表接口只生成经裁剪的摘要。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `uint`，主键 | 数据库事件记录 ID。 |
| `TaskID` | `string`，最长 64、索引、非空、与 `EventID` 联合唯一 | 所属任务。 |
| `EventID` | `int64`，非空、与 `TaskID` 联合唯一 | 任务内的单调事件序号。 |
| `Timestamp` | `time.Time`，索引 | 事件发生时间。 |
| `ElapsedMS` | `int64`，默认 0 | 自任务开始的累计毫秒。 |
| `Kind` | `string`，最长 64、索引 | 事件类别，例如 LLM、工具或生命周期事件。 |
| `Phase` | `string`，最长 64、索引 | 所在执行阶段。 |
| `Name` | `string`，最长 128、索引 | 工具、模型或步骤名称。 |
| `Status` | `string`，最长 32、索引 | 事件状态。 |
| `Detail` | `string`，`text` | 面向 Timeline 的简短说明。 |
| `Metadata` | `string`，`longtext` | 完整结构化元数据或载荷；不得直接在公开列表中回显。 |
| `CreatedAt` | `time.Time` | 持久化时间。 |

### `TaskErrorAnalysis`（默认表：`task_error_analyses`）

任务空闲或失败后由分析器保存的诊断结论，供运维与后续修复迭代使用。

| 字段 | Go 类型 / 约束 | 含义 |
| --- | --- | --- |
| `ID` | `uint`，主键 | 分析记录 ID。 |
| `TaskID` | `string`，最长 64、索引、非空 | 被分析的任务。 |
| `TriggerType` | `string`，最长 20、非空 | 触发原因，当前约定为 `idle` 或 `failed`。 |
| `LogSnippet` | `string`，`longtext` | 被分析的原始日志片段。 |
| `Analysis` | `string`，`longtext` | LLM 给出的完整分析。 |
| `RootCause` | `string`，`text` | 提炼后的根因。 |
| `Suggestion` | `string`，`text` | 建议的修复动作。 |
| `TokensUsed` | `int64`，默认 0 | 诊断调用的 token 消耗。 |
| `ModelUsed` | `string`，最长 100 | 诊断使用的模型标识。 |
| `CreatedAt` | `time.Time` | 分析生成时间。 |

## 5. 管理端聚合结构体（不建表）

以下结构体仅承载管理 API 的查询结果，不参与 `AutoMigrate`。

### `AdminMetrics`

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `RegisteredUserCount` | `int64` | 非游客注册用户数。 |
| `NonRootRegisteredUserCount` | `int64` | 排除部署根账号后的注册用户数。 |
| `PPTActiveUserCount` | `int64` | 有 PPT 创建记录的去重活跃用户数；游客按 IP 摘要聚合。 |
| `CustomAPIKeyUserCount` | `int64` | 配置了用户级 API Key 的用户数。 |
| `PPTGenerationCount` | `int64` | `Intent=create` 的任务数。 |
| `NonRootPPTGenerationCount` | `int64` | 排除根账号后的 PPT 创建任务数。 |
| `FeedbackCount` | `int64` | 反馈记录总数。 |
| `FeedbackSuggestionCount` | `int64` | 含非空建议的反馈数。 |

### `AdminUserMetric`

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| 嵌入 `User` | `User` | 用户基础字段；API 层仍需避免输出密码等敏感字段。 |
| `PPTGenerationCount` | `int64` | 该用户的 PPT 创建任务数。 |
| `CustomAPIKeyConfigured` | `bool` | 是否存在用户级模型 Key；不携带 Key 原文。 |

### `AdminTaskFeedback`

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| 嵌入 `TaskFeedback` | `TaskFeedback` | 反馈主体。 |
| `UserEmail` | `string` | 反馈用户邮箱，来自 `users` 关联查询。 |
| `TaskQuery` | `string` | 对应任务请求，来自 `task_records` 关联查询。 |

## 使用与演进约束

- 对外代码继续导入 `github.com/cloudwego/ppt-agent/pkg/db` 并使用 `db.User` 等类型；不要导入 `internal/model`。
- `TaskRecord.Intent`、`ConversationMessage.Role`、`TaskErrorAnalysis.TriggerType` 都是稳定的机器字段，不能改为中文显示文案。
- `User.Password`、`UserAPIKey.APIKey`、验证码字段和运行事件的完整 `Metadata` 属于敏感或受限数据，接口层需要最小化返回。
- 为模型新增字段时，同时检查 `internal/schema.Migrate`、相关仓储选择列、任务恢复逻辑、管理统计和本文档。
