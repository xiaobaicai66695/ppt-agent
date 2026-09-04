# PPT Agent 生产结构体总索引

本索引覆盖生产代码中的具名 struct。每一行的“字段职责”是字段组的解释；完整字段名、类型、JSON/GORM tag 和逐字段说明见[生产 struct 逐字段参考](./ppt-agent-production-struct-fields.md)。持久化模型的关联与约束细节见[持久化模型文档](./ppt-agent-persistence-models.md)。

## 入口、路由与 Web API

| 源码 | struct | 使用场景与字段职责 |
| --- | --- | --- |
| `main.go` | `aiModelAdapter`、`userModelCredential` | 启动层模型适配与按用户装配的模型凭据。 |
| `pkg/runtime/web/server.go` | `Server`、`ServerConfig` | HTTP 服务依赖、任务管理器、路由模型和服务配置。 |
| `pkg/runtime/web/request_router.go` | `createRequestRoute`、`MessageRouteResult`、`TaskCandidate` | 创建请求路由决策、消息路由输出和候选任务摘要。 |
| `pkg/runtime/web/continuation_handler.go` | `RouteResult`、`FixDetails` | 续办/修复请求的路由结果及修复范围。 |
| `pkg/runtime/web/benchmark_router.go` | `BenchmarkCreateRouteResult`、`BenchmarkMessageRouteResult`、`benchmarkModelAdapter` | 路由基准接口输出与测试模型适配。 |
| `pkg/runtime/web/benchmark_chat.go` | `ChatBenchmarkSearchResult`、`ChatBenchmarkImageResult`、`ChatBenchmarkInput` | 聊天基准的搜索、图片与输入载荷。 |
| `pkg/runtime/web/message_chat.go` | `chatImageResult`、`chatImageSearchResponse`、`chatAugmentations`、`chatTraceEvent` | 聊天附加图片、资料增强及流式轨迹事件。 |
| `pkg/runtime/web/plan_draft.go` | `planDraftResponse` | 规划草稿的 API 返回视图。 |
| `pkg/runtime/web/admin_handler.go` | `adminTaskResponse` | 管理端任务列表的安全投影视图。 |
| `pkg/runtime/web/credential_handler.go` | `modelCredential` | 用户模型凭据请求/响应的安全 DTO。 |
| `pkg/runtime/web/health.go` | `HealthStatus`、`HealthReport` | 服务、数据库与依赖项健康检查结果。 |

字段规则：Web DTO 应只带调用方需要的状态、ID、展示摘要和安全化数据；密钥、完整日志、原始模型载荷不可直接进入响应。

## 路由、DeckSpec、PPT 规划与渲染

| 源码 | struct | 使用场景与字段职责 |
| --- | --- | --- |
| `pkg/agent/router/agent.go` | `Input`、`Result`、`ContinuationInput`、`ContinuationResult`、`FixDetails`、`Agent` | 统一请求分类、续办判断、修复信息及 Router Agent 依赖。 |
| `pkg/agent/deck/types.go` | `PPTTaskConfig`、`TasksManifest`、`VisualPolicy`、`TaskItem`、`ManifestValidationReport` | 任务工作目录、Deck manifest、视觉政策、单页任务与校验报告。 |
| 同上 | `TaskOutline`、`DeckSection`、`PlanComponent`、`PlanReviewIssue` | 用户大纲、章节、组件级内容规划与 Reviewer 问题项。 |
| 同上 | `ContentPlan`、`VisualIntent`、`SlideOutline`、`PPTTaskStart`、`PPTTaskResult` | 每页内容容量、图片/背景意图、页级大纲、启动参数和交付结果。 |
| `pkg/agent/deck/run.go` | `reviewCheckpoint`、`AgentEvent` | 审查检查点与 Planner/Reviewer/Fixer 的运行事件。 |
| `pkg/agent/deck/deck_renderer.go` | `DeckRenderEvent`、`deckRenderInput`、`deckRenderContext` | 单页渲染进度、渲染输入与共享上下文。 |
| `pkg/agent/deck/manifest_tool.go` | `manifestTaskPatch`、`manifestToolInput`、`manifestToolRawInput`、`manifestTool`、`plannerManifestTool` | Agent 对 manifest 的受控 patch 协议与工具状态。 |
| `pkg/agent/deck/fixer_manifest_tool.go` | `draftTasksPatchTool`、`selectedTasksPatchTool` | Fixer 可修改的草稿/选中页 manifest 工具。 |
| `pkg/agent/deck/plan_review_tool.go` | `PlanReviewReport` | 审查器输出的结构化通过/问题报告。 |
| `pkg/agent/deck/plan_review_revision.go` | `planReviewRevisionPayload`、`planReviewScope`、`planReviewTask` | 审查后修订的请求、范围和单页任务。 |
| `pkg/agent/deck/planner_recovery.go` | `recoveredThought`、`recoveredSlideSpec` | 从中断运行恢复 Planner 思路与页规格。 |
| `pkg/agent/deck/background_assets.go` | `plannedBackgroundTarget`、`assetQueryRevisionRequest`、`assetQueryRevisionError`、`resolvedBackgroundAsset`、`plannedImageAssetTarget`、`MaterializedDeckAssetCounts` | 背景/图片检索词、改写失败、落盘素材与素材计数。 |
| `pkg/agent/command/operator.go` | `WorkDirBackend`、`LocalOperator` | Agent 文件与命令操作的工作目录边界。 |
| `pkg/templates/loader.go` | `LayoutInfo`、`LayoutContract`、`Field`、`componentContractsFile`、`Loader` | 模板布局、组件字段契约、加载后的索引和资源路径。 |

字段规则：`content_type`、组件 kind、布局 ID 和 Agent 工具输入均是稳定机器契约；显示名、中文文案不得写回这些字段。

## 任务运行时、模型调用与可观测性

| 源码 | struct | 使用场景与字段职责 |
| --- | --- | --- |
| `pkg/runtime/task/manager.go` | `SSERichEvent`、`TaskInfo`、`DeliveryFeedback`、`TaskState`、`TaskManager` | SSE 事件、可持久化任务快照、交付反馈、并发状态和任务容器。 |
| `pkg/runtime/task/delivery_metadata.go` | `DeliverySnapshot` | 交付文件、缩略图与质量元数据的读取快照。 |
| `pkg/runtime/model/model.go` | `ChatModelConfig`、`globalRateLimitTracker`、`modelCallLimiter`、`FallbackChatModel`、`modelCallProfile`、`compressorConfig` | 模型供应商配置、全局限流、调用画像、fallback 模型和压缩参数。 |
| `pkg/runtime/model/runtime_model.go` | `RuntimeStatusChatModel` | 为模型调用附加运行状态的包装器。 |
| `pkg/runtime/model/compressor.go` | `CompressorConfig`、`CompressionEvent`、`CompressionSummary`、`ChatModelCompressor`、`compressionState`、`pairWithToolCalls` | 上下文压缩配置、事件、汇总、状态机与工具调用配对。 |
| `pkg/runtime/model/runtime_meta.go` | `RuntimeMeta`、`manifestValidationState`、`RuntimeBudgets`、`TaskInputAnchor`、`PlanSlide` | 任务运行元数据、manifest 校验、预算、输入锚点与计划页快照。 |
| 同上 | `AlignmentWarning`、`RuntimeMetaSnapshot`、`RuntimeEvent`、`RuntimeReport` | 对齐告警、持久化快照、Timeline 单事件与完整报告。 |
| `pkg/runtime/model/token_tracker.go` | `TokenTracker`、`tokenTrackerKey` | token 累计统计与 context 存取键。 |
| `pkg/chattrace/redis.go` | `Event`、`RedisStore` | Redis 中聊天轨迹的事件载荷及存储实现。 |
| `pkg/log_analysis/service.go` | `Result`、`LLMAnalyzer`、`readFileOperator`、`Service`、`taskRequest`、`ServiceConfig` | 失败日志分析结果、LLM 依赖、任务请求与后台服务配置。 |
| `pkg/retry/retry.go` | `Policy`、`RetryableError`、`FixedAttemptStrategy`、`Decision`、`Factory` | 重试上限、可重试错误、策略决策及策略工厂。 |
| 同上 | `HTTP410SearchTermRevisionStrategy`、`RateLimitFallbackStrategy`、`TransientModelRetryStrategy`、`ModelStreamReadFallbackStrategy` | 针对搜索词、限流、短暂模型失败与流读取失败的策略参数/状态。 |

字段规则：运行事件可保留诊断元数据，但向 Web 输出时应做长度限制和脱敏；限流、预算、token 与重试字段必须保持数值单位一致。

## 账号、会话、持久化与人工介入

| 源码 | struct | 使用场景与字段职责 |
| --- | --- | --- |
| `pkg/db/internal/model/account.go` | `User`、`UserAPIKey`、`VerificationCode` | 账号身份、用户级模型 Key、验证码。逐字段见持久化模型文档。 |
| `pkg/db/internal/model/task.go` | `TaskRecord`、`TaskFeedback`、`PlanDraftRecord` | 任务历史、用户反馈与未渲染规划。 |
| `pkg/db/internal/model/conversation.go` | `ConversationMessage`、`ConversationMessageChunk` | 对话消息和 UTF-8 安全大正文分块。 |
| `pkg/db/internal/model/runtime.go` | `RuntimeEventRecord`、`TaskErrorAnalysis` | 持久化 Timeline 与失败分析。 |
| `pkg/db/admin.go` | `AdminMetrics`、`AdminUserMetric`、`AdminTaskFeedback` | 管理端安全聚合。 |
| `pkg/auth/auth.go` | `jwtClaims` | JWT 中的用户 ID、邮箱、管理员身份和标准注册声明。 |
| `pkg/session/session.go` | `Message`、`ConversationSession`、`SessionManager` | 内存会话消息、会话视图与并发会话管理。 |
| `pkg/human/manager.go` | `Manager` | 人工审批/恢复的等待器和并发控制。 |
| `pkg/tools/human_in_the_loop.go` | `SearchApprovalInfo`、`SearchApprovalResult`、`InvokableSearchApprovalTool` | 搜索授权所需信息、用户决定与可调用工具包装。 |
| `pkg/callback/callback.go` | `callbackInputDetailsKey`、`httpErrorDetail` | callback context 键和 HTTP 错误详细信息。 |

## 外部工具、素材与基础设施

| 源码 | struct | 使用场景与字段职责 |
| --- | --- | --- |
| `pkg/tools/search/search_tool.go` | `tokenBucket`、`searchTool`、`URLContent`、`searchInput`、`SearchResponse`、`SearchResult` | 检索限流、工具状态、网页正文、工具输入与标准检索结果。 |
| 同上 | `qianfanRequest`、`qianfanMessage`、`qianfanResource`、`qianfanSearchFilter`、`qianfanMatch`、`qianfanResponse`、`qianfanRef`、`evidenceBuilder` | 千帆搜索的请求/响应映射及证据正文构建器。 |
| `pkg/tools/image/image_search_tool.go` | `imageSearchTool`、`imageSearchInput`、`ImageSearchResponse`、`ImagePhoto` | 图片搜索工具、输入条件与标准化图片结果。 |
| `pkg/tools/read_file.go` | `readFileTool`、`options`、`readInput` | 受工作目录限制的读文件工具及输入选项。 |
| `pkg/utils/unsplash/types.go` | `SearchOptions`、`SearchResponse`、`Photo`、`PhotoURLs`、`PhotoLinks`、`User`、`UserLinks` | Unsplash API 的搜索条件和上游响应映射。 |
| `pkg/utils/unsplash/helpers.go`、`pkg/utils/unsplash/client.go` | `DownloadedAsset`、`Client` | 已物化图片资产及 Unsplash 客户端配置。 |
| `pkg/prompts/prompts.go` | `TemplateData`、`LogAnalysisData` | Prompt 模板渲染所需任务/日志变量。 |
| `pkg/agent/modelcompat/modelcompat.go` | `ModelSpec`、`DefaultFactory` | 供应商能力声明与默认模型工厂。 |
| `pkg/utils/params/context.go` | `contextParams`、`typedContextParams` | context 参数访问的私有键类型。 |

## 运维与评估命令

| 源码 | struct | 使用场景与字段职责 |
| --- | --- | --- |
| `cmd/pptbench/types.go` | `options`、`benchCase`、`caseInput`、`agentOutput`、`modelOutput` | 基准命令配置、样例输入及 Agent/模型输出。 |
| 同上 | `contentQualityReport`、`contentPageClaim`、`agendaSubtitleIssue`、`judgeInput`、`judgeResult`、`caseSummary`、`runSummary` | 内容质量判定、页级主张、议程问题、裁判输入输出和汇总结果。 |

## 字段维护检查

1. 新增数据库字段：更新 `internal/model`、`internal/schema.Migrate`、相关仓储 select/update、持久化模型文档。
2. 新增 Web DTO：明确 JSON tag、权限边界、错误字段和前端消费者。
3. 新增 DeckSpec 字段：同步 Planner prompt、Reviewer 规则、生成器契约与前端标签。
4. 新增外部 API 载荷：将上游原始字段与系统标准字段分开，避免凭据或不可信正文穿透到用户接口。
5. 新增内部状态：说明并发归属、生命周期和是否可安全持久化；避免将 mutex、channel、context 直接序列化。
