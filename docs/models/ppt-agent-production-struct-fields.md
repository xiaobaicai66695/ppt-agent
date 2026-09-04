# PPT Agent 生产 Struct 逐字段参考

本文档自动依据当前 `backend` 生产 Go 源码生成，逐个列出具名 struct 的字段、Go 类型、tag 与字段说明。业务链路总览见 [生产结构体总索引](./ppt-agent-production-struct-catalog.md)，数据库模型的关联与约束细节见 [持久化模型](./ppt-agent-persistence-models.md)。`*_test.go` 不纳入。

> tag 列保留 JSON/GORM 等原始契约；空 struct 会明确标注为无字段。字段名称相同但出现在不同 struct 中时，以所在 struct 的使用场景为准。

## `options` — `cmd/pptbench/types.go:10`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `dataset` | `string` | `` | 承载 dataset 的 string 值；业务上下文见本 struct 的源码包。 |
| `suite` | `string` | `` | 承载 suite 的 string 值；业务上下文见本 struct 的源码包。 |
| `step` | `string` | `` | 承载 step 的 string 值；业务上下文见本 struct 的源码包。 |
| `casesPath` | `string` | `` | 承载 casesPath 的 string 值；业务上下文见本 struct 的源码包。 |
| `outPath` | `string` | `` | 承载 outPath 的 string 值；业务上下文见本 struct 的源码包。 |
| `limit` | `int` | `` | 承载 limit 的 int 值；业务上下文见本 struct 的源码包。 |
| `caseID` | `string` | `` | 承载 caseID 的 string 值；业务上下文见本 struct 的源码包。 |
| `timeout` | `time.Duration` | `` | 承载 timeout 的 time.Duration 值；业务上下文见本 struct 的源码包。 |

## `benchCase` — `cmd/pptbench/types.go:21`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `Name` | `string` | ``json:"name"`` | 承载 Name 的 string 值；业务上下文见本 struct 的源码包。 |
| `Input` | `json.RawMessage` | ``json:"input"`` | 承载 Input 的 json.RawMessage 值；业务上下文见本 struct 的源码包。 |
| `Expected` | `json.RawMessage` | ``json:"expected,omitempty"`` | 承载 Expected 的 json.RawMessage 值；业务上下文见本 struct 的源码包。 |
| `JudgeFocus` | `[]string` | ``json:"judge_focus,omitempty"`` | 承载 JudgeFocus 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Raw` | `json.RawMessage` | ``json:"-"`` | 承载 Raw 的 json.RawMessage 值；业务上下文见本 struct 的源码包。 |

## `caseInput` — `cmd/pptbench/types.go:30`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `UserRequest` | `string` | ``json:"user_request"`` | 承载 UserRequest 的 string 值；业务上下文见本 struct 的源码包。 |
| `UserMessage` | `string` | ``json:"user_message"`` | 承载 UserMessage 的 string 值；业务上下文见本 struct 的源码包。 |
| `HasOutline` | `bool` | ``json:"has_outline"`` | 承载 HasOutline 的 bool 值；业务上下文见本 struct 的源码包。 |
| `HasExistingTask` | `bool` | ``json:"has_existing_task"`` | 承载 HasExistingTask 的 bool 值；业务上下文见本 struct 的源码包。 |
| `TasksSummary` | `string` | ``json:"tasks_summary"`` | 承载 TasksSummary 的 string 值；业务上下文见本 struct 的源码包。 |
| `ConversationContext` | `[]string` | ``json:"conversation_context"`` | 承载 ConversationContext 的 []string 值；业务上下文见本 struct 的源码包。 |
| `DraftTasks` | `*deck.TasksManifest` | ``json:"draft_tasks"`` | 承载 DraftTasks 的 *deck.TasksManifest 值；业务上下文见本 struct 的源码包。 |
| `BaseTasks` | `*deck.TasksManifest` | ``json:"base_tasks"`` | 承载 BaseTasks 的 *deck.TasksManifest 值；业务上下文见本 struct 的源码包。 |
| `ReviewIssues` | `[]deck.PlanReviewIssue` | ``json:"review_issues"`` | 承载 ReviewIssues 的 []deck.PlanReviewIssue 值；业务上下文见本 struct 的源码包。 |
| `AllowedPageIndexes` | `[]int` | ``json:"allowed_page_indexes"`` | 承载 AllowedPageIndexes 的 []int 值；业务上下文见本 struct 的源码包。 |
| `SourceMaterials` | `[]any` | ``json:"source_materials"`` | 承载 SourceMaterials 的 []any 值；业务上下文见本 struct 的源码包。 |
| `Requirements` | `[]string` | ``json:"requirements"`` | 承载 Requirements 的 []string 值；业务上下文见本 struct 的源码包。 |

## `agentOutput` — `cmd/pptbench/types.go:45`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `CaseID` | `string` | ``json:"case_id"`` | 承载 CaseID 的 string 值；业务上下文见本 struct 的源码包。 |
| `Suite` | `string` | ``json:"suite"`` | 承载 Suite 的 string 值；业务上下文见本 struct 的源码包。 |
| `StartedAt` | `string` | ``json:"started_at"`` | 承载 StartedAt 的 string 值；业务上下文见本 struct 的源码包。 |
| `DurationMS` | `int64` | ``json:"duration_ms"`` | 承载 DurationMS 的 int64 值；业务上下文见本 struct 的源码包。 |
| `Output` | `any` | ``json:"output,omitempty"`` | 承载 Output 的 any 值；业务上下文见本 struct 的源码包。 |
| `Before` | `any` | ``json:"before,omitempty"`` | 承载 Before 的 any 值；业务上下文见本 struct 的源码包。 |
| `After` | `any` | ``json:"after,omitempty"`` | 承载 After 的 any 值；业务上下文见本 struct 的源码包。 |
| `Events` | `[]deck.AgentEvent` | ``json:"events,omitempty"`` | 承载 Events 的 []deck.AgentEvent 值；业务上下文见本 struct 的源码包。 |
| `DeterministicReview` | `*deck.PlanReviewReport` | ``json:"deterministic_review,omitempty"`` | 承载 DeterministicReview 的 *deck.PlanReviewReport 值；业务上下文见本 struct 的源码包。 |
| `ContentQuality` | `*contentQualityReport` | ``json:"content_quality,omitempty"`` | 承载 ContentQuality 的 *contentQualityReport 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |

## `modelOutput` — `cmd/pptbench/types.go:59`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `CaseID` | `string` | ``json:"case_id"`` | 承载 CaseID 的 string 值；业务上下文见本 struct 的源码包。 |
| `Suite` | `string` | ``json:"suite"`` | 承载 Suite 的 string 值；业务上下文见本 struct 的源码包。 |
| `Output` | `any` | ``json:"output,omitempty"`` | 承载 Output 的 any 值；业务上下文见本 struct 的源码包。 |
| `Before` | `any` | ``json:"before,omitempty"`` | 承载 Before 的 any 值；业务上下文见本 struct 的源码包。 |
| `After` | `any` | ``json:"after,omitempty"`` | 承载 After 的 any 值；业务上下文见本 struct 的源码包。 |
| `DeterministicReview` | `*deck.PlanReviewReport` | ``json:"deterministic_review,omitempty"`` | 承载 DeterministicReview 的 *deck.PlanReviewReport 值；业务上下文见本 struct 的源码包。 |
| `ContentQuality` | `*contentQualityReport` | ``json:"content_quality,omitempty"`` | 承载 ContentQuality 的 *contentQualityReport 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |

## `contentQualityReport` — `cmd/pptbench/types.go:73`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `DeckClaim` | `string` | ``json:"deck_claim,omitempty"`` | 承载 DeckClaim 的 string 值；业务上下文见本 struct 的源码包。 |
| `PageClaims` | `[]contentPageClaim` | ``json:"page_claims,omitempty"`` | 承载 PageClaims 的 []contentPageClaim 值；业务上下文见本 struct 的源码包。 |
| `MissingClaimPages` | `[]int` | ``json:"missing_claim_pages,omitempty"`` | 承载 MissingClaimPages 的 []int 值；业务上下文见本 struct 的源码包。 |
| `DuplicateClaimGroups` | `[][]int` | ``json:"duplicate_claim_groups,omitempty"`` | 承载 DuplicateClaimGroups 的 [][]int 值；业务上下文见本 struct 的源码包。 |
| `LongestRepeatedLayoutRun` | `int` | ``json:"longest_repeated_layout_run,omitempty"`` | 承载 LongestRepeatedLayoutRun 的 int 值；业务上下文见本 struct 的源码包。 |
| `RepeatedLayoutRunContentTypes` | `[]string` | ``json:"repeated_layout_run_content_types,omitempty"`` | 承载 RepeatedLayoutRunContentTypes 的 []string 值；业务上下文见本 struct 的源码包。 |
| `AgendaTOCItems` | `int` | ``json:"agenda_toc_items,omitempty"`` | 承载 AgendaTOCItems 的 int 值；业务上下文见本 struct 的源码包。 |
| `AgendaTOCSubtitles` | `int` | ``json:"agenda_toc_subtitles,omitempty"`` | 承载 AgendaTOCSubtitles 的 int 值；业务上下文见本 struct 的源码包。 |
| `AgendaSubtitleIssues` | `[]agendaSubtitleIssue` | ``json:"agenda_subtitle_issues,omitempty"`` | 承载 AgendaSubtitleIssues 的 []agendaSubtitleIssue 值；业务上下文见本 struct 的源码包。 |

## `contentPageClaim` — `cmd/pptbench/types.go:85`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PageIndex` | `int` | ``json:"page_index"`` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentType` | `string` | ``json:"content_type"`` | 承载 ContentType 的 string 值；业务上下文见本 struct 的源码包。 |
| `Claim` | `string` | ``json:"claim,omitempty"`` | 承载 Claim 的 string 值；业务上下文见本 struct 的源码包。 |

## `agendaSubtitleIssue` — `cmd/pptbench/types.go:94`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PageIndex` | `int` | ``json:"page_index"`` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `ComponentID` | `string` | ``json:"component_id,omitempty"`` | 承载 ComponentID 的 string 值；业务上下文见本 struct 的源码包。 |
| `Code` | `string` | ``json:"code"`` | 承载 Code 的 string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title,omitempty"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |

## `judgeInput` — `cmd/pptbench/types.go:101`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Case` | `any` | ``json:"case"`` | 承载 Case 的 any 值；业务上下文见本 struct 的源码包。 |
| `Suite` | `string` | ``json:"suite"`` | 承载 Suite 的 string 值；业务上下文见本 struct 的源码包。 |
| `Rubric` | `string` | ``json:"rubric"`` | 承载 Rubric 的 string 值；业务上下文见本 struct 的源码包。 |
| `ModelOutput` | `any` | ``json:"model_output"`` | 承载 ModelOutput 的 any 值；业务上下文见本 struct 的源码包。 |
| `RequiredOutputSchema` | `any` | ``json:"required_output_schema"`` | 承载 RequiredOutputSchema 的 any 值；业务上下文见本 struct 的源码包。 |
| `ScoringScale` | `[]string` | ``json:"scoring_scale"`` | 承载 ScoringScale 的 []string 值；业务上下文见本 struct 的源码包。 |

## `judgeResult` — `cmd/pptbench/types.go:110`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `CaseID` | `string` | ``json:"case_id"`` | 承载 CaseID 的 string 值；业务上下文见本 struct 的源码包。 |
| `Suite` | `string` | ``json:"suite"`` | 承载 Suite 的 string 值；业务上下文见本 struct 的源码包。 |
| `Score` | `int` | ``json:"score"`` | 承载 Score 的 int 值；业务上下文见本 struct 的源码包。 |
| `Pass` | `bool` | ``json:"pass"`` | 承载 Pass 的 bool 值；业务上下文见本 struct 的源码包。 |
| `DimensionScores` | `map[string]int` | ``json:"dimension_scores,omitempty"`` | 承载 DimensionScores 的 map[string]int 值；业务上下文见本 struct 的源码包。 |
| `Strengths` | `[]string` | ``json:"strengths,omitempty"`` | 承载 Strengths 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Weaknesses` | `[]string` | ``json:"weaknesses,omitempty"`` | 承载 Weaknesses 的 []string 值；业务上下文见本 struct 的源码包。 |
| `CriticalFailures` | `[]string` | ``json:"critical_failures,omitempty"`` | 承载 CriticalFailures 的 []string 值；业务上下文见本 struct 的源码包。 |
| `RecommendedFix` | `string` | ``json:"recommended_fix,omitempty"`` | 承载 RecommendedFix 的 string 值；业务上下文见本 struct 的源码包。 |
| `RawContent` | `string` | ``json:"raw_content,omitempty"`` | 正文或序列化内容。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |

## `caseSummary` — `cmd/pptbench/types.go:124`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `CaseID` | `string` | ``json:"case_id"`` | 承载 CaseID 的 string 值；业务上下文见本 struct 的源码包。 |
| `Suite` | `string` | ``json:"suite"`` | 承载 Suite 的 string 值；业务上下文见本 struct 的源码包。 |
| `Score` | `int` | ``json:"score,omitempty"`` | 承载 Score 的 int 值；业务上下文见本 struct 的源码包。 |
| `Pass` | `bool` | ``json:"pass,omitempty"`` | 承载 Pass 的 bool 值；业务上下文见本 struct 的源码包。 |
| `Judged` | `bool` | ``json:"judged,omitempty"`` | 承载 Judged 的 bool 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |

## `runSummary` — `cmd/pptbench/types.go:133`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Suite` | `string` | ``json:"suite"`` | 承载 Suite 的 string 值；业务上下文见本 struct 的源码包。 |
| `StartedAt` | `string` | ``json:"started_at"`` | 承载 StartedAt 的 string 值；业务上下文见本 struct 的源码包。 |
| `FinishedAt` | `string` | ``json:"finished_at"`` | 承载 FinishedAt 的 string 值；业务上下文见本 struct 的源码包。 |
| `Total` | `int` | ``json:"total"`` | 承载 Total 的 int 值；业务上下文见本 struct 的源码包。 |
| `Judged` | `int` | ``json:"judged"`` | 承载 Judged 的 int 值；业务上下文见本 struct 的源码包。 |
| `Passed` | `int` | ``json:"passed"`` | 承载 Passed 的 int 值；业务上下文见本 struct 的源码包。 |
| `Average` | `float64` | ``json:"average_score,omitempty"`` | 承载 Average 的 float64 值；业务上下文见本 struct 的源码包。 |
| `Cases` | `[]caseSummary` | ``json:"cases"`` | 承载 Cases 的 []caseSummary 值；业务上下文见本 struct 的源码包。 |

## `aiModelAdapter` — `main.go:313`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `model` | `model.ToolCallingChatModel` | `` | 承载 model 的 model.ToolCallingChatModel 值；业务上下文见本 struct 的源码包。 |

## `userModelCredential` — `main.go:400`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Provider` | `string` | `` | 承载 Provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `APIKey` | `string` | `` | 承载 APIKey 的 string 值；业务上下文见本 struct 的源码包。 |

## `WorkDirBackend` — `pkg/agent/command/operator.go:37`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `workDirFunc` | `WorkDirFunc` | `` | 承载 workDirFunc 的 WorkDirFunc 值；业务上下文见本 struct 的源码包。 |

## `LocalOperator` — `pkg/agent/command/operator.go:69`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.RWMutex` | `` | 承载 mu 的 sync.RWMutex 值；业务上下文见本 struct 的源码包。 |
| `workDir` | `string` | `` | 承载 workDir 的 string 值；业务上下文见本 struct 的源码包。 |

## `plannedBackgroundTarget` — `pkg/agent/deck/background_assets.go:30`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `taskID` | `string` | `` | 关联任务的唯一标识。 |
| `pageIndex` | `int` | `` | 承载 pageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `query` | `string` | `` | 用户或系统发起的查询文本。 |
| `subject` | `string` | `` | 承载 subject 的 string 值；业务上下文见本 struct 的源码包。 |
| `contentType` | `string` | `` | 承载 contentType 的 string 值；业务上下文见本 struct 的源码包。 |
| `slot` | `int` | `` | 承载 slot 的 int 值；业务上下文见本 struct 的源码包。 |
| `searchPage` | `int` | `` | 承载 searchPage 的 int 值；业务上下文见本 struct 的源码包。 |
| `visual` | `*VisualIntent` | `` | 承载 visual 的 *VisualIntent 值；业务上下文见本 struct 的源码包。 |
| `component` | `*PlanComponent` | `` | 承载 component 的 *PlanComponent 值；业务上下文见本 struct 的源码包。 |

## `assetQueryRevisionRequest` — `pkg/agent/deck/background_assets.go:44`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `TaskIDs` | `[]string` | `` | 承载 TaskIDs 的 []string 值；业务上下文见本 struct 的源码包。 |
| `PageIndexes` | `[]int` | `` | 承载 PageIndexes 的 []int 值；业务上下文见本 struct 的源码包。 |
| `Queries` | `[]string` | `` | 承载 Queries 的 []string 值；业务上下文见本 struct 的源码包。 |

## `assetQueryRevisionError` — `pkg/agent/deck/background_assets.go:50`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Requests` | `[]assetQueryRevisionRequest` | `` | 承载 Requests 的 []assetQueryRevisionRequest 值；业务上下文见本 struct 的源码包。 |
| `Cause` | `error` | `` | 承载 Cause 的 error 值；业务上下文见本 struct 的源码包。 |

## `resolvedBackgroundAsset` — `pkg/agent/deck/background_assets.go:78`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `photo` | `unsplash.Photo` | `` | 承载 photo 的 unsplash.Photo 值；业务上下文见本 struct 的源码包。 |
| `asset` | `*unsplash.DownloadedAsset` | `` | 承载 asset 的 *unsplash.DownloadedAsset 值；业务上下文见本 struct 的源码包。 |
| `provider` | `string` | `` | 承载 provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `searchStatus` | `string` | `` | 当前生命周期或处理状态。 |

## `plannedImageAssetTarget` — `pkg/agent/deck/background_assets.go:85`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `query` | `string` | `` | 用户或系统发起的查询文本。 |
| `component` | `*PlanComponent` | `` | 承载 component 的 *PlanComponent 值；业务上下文见本 struct 的源码包。 |

## `MaterializedDeckAssetCounts` — `pkg/agent/deck/background_assets.go:90`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Backgrounds` | `int` | `` | 承载 Backgrounds 的 int 值；业务上下文见本 struct 的源码包。 |
| `Images` | `int` | `` | 承载 Images 的 int 值；业务上下文见本 struct 的源码包。 |

## `DeckRenderEvent` — `pkg/agent/deck/deck_renderer.go:23`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Type` | `string` | `` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | `` | 关联任务的唯一标识。 |
| `PageIndex` | `int` | `` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `OutputFile` | `string` | `` | 承载 OutputFile 的 string 值；业务上下文见本 struct 的源码包。 |
| `Detail` | `string` | `` | 承载 Detail 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | `` | 失败原因或错误详情。 |

## `deckRenderInput` — `pkg/agent/deck/deck_renderer.go:34`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Config` | `*PPTTaskConfig` | `` | 对应组件的运行配置。 |
| `Callback` | `DeckRenderEventCallback` | `` | 承载 Callback 的 DeckRenderEventCallback 值；业务上下文见本 struct 的源码包。 |
| `Started` | `time.Time` | `` | 承载 Started 的 time.Time 值；业务上下文见本 struct 的源码包。 |

## `deckRenderContext` — `pkg/agent/deck/deck_renderer.go:40`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Config` | `*PPTTaskConfig` | `` | 对应组件的运行配置。 |
| `Callback` | `DeckRenderEventCallback` | `` | 承载 Callback 的 DeckRenderEventCallback 值；业务上下文见本 struct 的源码包。 |
| `Started` | `time.Time` | `` | 承载 Started 的 time.Time 值；业务上下文见本 struct 的源码包。 |
| `Manifest` | `*TasksManifest` | `` | 承载 Manifest 的 *TasksManifest 值；业务上下文见本 struct 的源码包。 |
| `Concurrency` | `int` | `` | 承载 Concurrency 的 int 值；业务上下文见本 struct 的源码包。 |
| `Tasks` | `[]*TaskItem` | `` | 承载 Tasks 的 []*TaskItem 值；业务上下文见本 struct 的源码包。 |

## `draftTasksPatchTool` — `pkg/agent/deck/fixer_manifest_tool.go:39`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `workDir` | `string` | `` | 承载 workDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `scoped` | `bool` | `` | 承载 scoped 的 bool 值；业务上下文见本 struct 的源码包。 |
| `allowed` | `map[int]bool` | `` | 承载 allowed 的 map[int]bool 值；业务上下文见本 struct 的源码包。 |

## `selectedTasksPatchTool` — `pkg/agent/deck/fixer_manifest_tool.go:92`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `workDir` | `string` | `` | 承载 workDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `allowed` | `map[string]bool` | `` | 承载 allowed 的 map[string]bool 值；业务上下文见本 struct 的源码包。 |

## `manifestTaskPatch` — `pkg/agent/deck/manifest_tool.go:79`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PageIndex` | `*int` | ``json:"page_index,omitempty"`` | 承载 PageIndex 的 *int 值；业务上下文见本 struct 的源码包。 |
| `SectionID` | `*string` | ``json:"section_id,omitempty"`` | 承载 SectionID 的 *string 值；业务上下文见本 struct 的源码包。 |
| `SectionTitle` | `*string` | ``json:"section_title,omitempty"`` | 承载 SectionTitle 的 *string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `*string` | ``json:"title,omitempty"`` | 承载 Title 的 *string 值；业务上下文见本 struct 的源码包。 |
| `ContentType` | `*string` | ``json:"content_type,omitempty"`` | 承载 ContentType 的 *string 值；业务上下文见本 struct 的源码包。 |
| `LayoutVariant` | `*string` | ``json:"layout_variant,omitempty"`` | 承载 LayoutVariant 的 *string 值；业务上下文见本 struct 的源码包。 |
| `PageIntent` | `*string` | ``json:"page_intent,omitempty"`` | 承载 PageIntent 的 *string 值；业务上下文见本 struct 的源码包。 |
| `EvidenceRefs` | `[]string` | ``json:"evidence_refs,omitempty"`` | 承载 EvidenceRefs 的 []string 值；业务上下文见本 struct 的源码包。 |
| `ContentPlan` | `*ContentPlan` | ``json:"content_plan,omitempty"`` | 承载 ContentPlan 的 *ContentPlan 值；业务上下文见本 struct 的源码包。 |

## `manifestToolInput` — `pkg/agent/deck/manifest_tool.go:91`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Mode` | `string` | ``json:"mode"`` | 承载 Mode 的 string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title,omitempty"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentBank` | `map[string]any` | ``json:"content_bank,omitempty"`` | 承载 ContentBank 的 map[string]any 值；业务上下文见本 struct 的源码包。 |
| `Sections` | `[]DeckSection` | ``json:"sections,omitempty"`` | 承载 Sections 的 []DeckSection 值；业务上下文见本 struct 的源码包。 |
| `VisualPolicy` | `*VisualPolicy` | ``json:"visual_policy,omitempty"`` | 承载 VisualPolicy 的 *VisualPolicy 值；业务上下文见本 struct 的源码包。 |
| `Tasks` | `[]manifestTaskPatch` | ``json:"tasks"`` | 承载 Tasks 的 []manifestTaskPatch 值；业务上下文见本 struct 的源码包。 |

## `manifestToolRawInput` — `pkg/agent/deck/manifest_tool.go:100`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Mode` | `string` | ``json:"mode"`` | 承载 Mode 的 string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title,omitempty"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentBank` | `map[string]any` | ``json:"content_bank,omitempty"`` | 承载 ContentBank 的 map[string]any 值；业务上下文见本 struct 的源码包。 |
| `Sections` | `[]DeckSection` | ``json:"sections,omitempty"`` | 承载 Sections 的 []DeckSection 值；业务上下文见本 struct 的源码包。 |
| `VisualPolicy` | `*VisualPolicy` | ``json:"visual_policy,omitempty"`` | 承载 VisualPolicy 的 *VisualPolicy 值；业务上下文见本 struct 的源码包。 |
| `Tasks` | `json.RawMessage` | ``json:"tasks"`` | 承载 Tasks 的 json.RawMessage 值；业务上下文见本 struct 的源码包。 |

## `manifestTool` — `pkg/agent/deck/manifest_tool.go:109`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `workDir` | `string` | `` | 承载 workDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `fallbackTitle` | `string` | `` | 承载 fallbackTitle 的 string 值；业务上下文见本 struct 的源码包。 |
| `draftFirst` | `bool` | `` | 承载 draftFirst 的 bool 值；业务上下文见本 struct 的源码包。 |

## `plannerManifestTool` — `pkg/agent/deck/manifest_tool.go:115`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `inner` | `*manifestTool` | `` | 承载 inner 的 *manifestTool 值；业务上下文见本 struct 的源码包。 |

## `planReviewRevisionPayload` — `pkg/agent/deck/plan_review_revision.go:10`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Round` | `int` | ``json:"round"`` | 承载 Round 的 int 值；业务上下文见本 struct 的源码包。 |
| `Summary` | `string` | ``json:"summary"`` | 承载 Summary 的 string 值；业务上下文见本 struct 的源码包。 |
| `Scope` | `planReviewScope` | ``json:"scope"`` | 承载 Scope 的 planReviewScope 值；业务上下文见本 struct 的源码包。 |
| `Issues` | `[]PlanReviewIssue` | ``json:"issues"`` | 承载 Issues 的 []PlanReviewIssue 值；业务上下文见本 struct 的源码包。 |
| `IncludedTasks` | `[]planReviewTask` | ``json:"included_tasks,omitempty"`` | 承载 IncludedTasks 的 []planReviewTask 值；业务上下文见本 struct 的源码包。 |
| `Instructions` | `[]string` | ``json:"instructions"`` | 承载 Instructions 的 []string 值；业务上下文见本 struct 的源码包。 |

## `planReviewScope` — `pkg/agent/deck/plan_review_revision.go:19`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PageIndexes` | `[]int` | ``json:"page_indexes,omitempty"`` | 承载 PageIndexes 的 []int 值；业务上下文见本 struct 的源码包。 |
| `SectionIDs` | `[]string` | ``json:"section_ids,omitempty"`` | 承载 SectionIDs 的 []string 值；业务上下文见本 struct 的源码包。 |
| `AllowedPageIndexes` | `[]int` | ``json:"allowed_page_indexes,omitempty"`` | 承载 AllowedPageIndexes 的 []int 值；业务上下文见本 struct 的源码包。 |
| `IncludesDeckLevel` | `bool` | ``json:"includes_deck_level,omitempty"`` | 承载 IncludesDeckLevel 的 bool 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |

## `planReviewTask` — `pkg/agent/deck/plan_review_revision.go:27`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PageIndex` | `int` | ``json:"page_index"`` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `SectionID` | `string` | ``json:"section_id,omitempty"`` | 承载 SectionID 的 string 值；业务上下文见本 struct 的源码包。 |
| `SectionTitle` | `string` | ``json:"section_title,omitempty"`` | 承载 SectionTitle 的 string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentType` | `string` | ``json:"content_type"`` | 承载 ContentType 的 string 值；业务上下文见本 struct 的源码包。 |
| `LayoutVariant` | `string` | ``json:"layout_variant,omitempty"`` | 承载 LayoutVariant 的 string 值；业务上下文见本 struct 的源码包。 |
| `PageIntent` | `string` | ``json:"page_intent,omitempty"`` | 承载 PageIntent 的 string 值；业务上下文见本 struct 的源码包。 |
| `EvidenceRefs` | `[]string` | ``json:"evidence_refs,omitempty"`` | 承载 EvidenceRefs 的 []string 值；业务上下文见本 struct 的源码包。 |
| `ContentPlan` | `*ContentPlan` | ``json:"content_plan,omitempty"`` | 承载 ContentPlan 的 *ContentPlan 值；业务上下文见本 struct 的源码包。 |

## `PlanReviewReport` — `pkg/agent/deck/plan_review_tool.go:23`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `OK` | `bool` | ``json:"ok"`` | 承载 OK 的 bool 值；业务上下文见本 struct 的源码包。 |
| `Passed` | `bool` | ``json:"passed"`` | 承载 Passed 的 bool 值；业务上下文见本 struct 的源码包。 |
| `Target` | `string` | ``json:"target"`` | 承载 Target 的 string 值；业务上下文见本 struct 的源码包。 |
| `Fingerprint` | `string` | ``json:"fingerprint"`` | 承载 Fingerprint 的 string 值；业务上下文见本 struct 的源码包。 |
| `Round` | `int` | ``json:"round"`` | 承载 Round 的 int 值；业务上下文见本 struct 的源码包。 |
| `TotalSlides` | `int` | ``json:"total_slides"`` | 承载 TotalSlides 的 int 值；业务上下文见本 struct 的源码包。 |
| `IssueCount` | `int` | ``json:"issue_count"`` | 承载 IssueCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `Issues` | `[]PlanReviewIssue` | ``json:"issues,omitempty"`` | 承载 Issues 的 []PlanReviewIssue 值；业务上下文见本 struct 的源码包。 |
| `Summary` | `string` | ``json:"summary"`` | 承载 Summary 的 string 值；业务上下文见本 struct 的源码包。 |
| `NextActions` | `[]string` | ``json:"next_actions,omitempty"`` | 承载 NextActions 的 []string 值；业务上下文见本 struct 的源码包。 |
| `ReviewedAt` | `string` | ``json:"reviewed_at"`` | 承载 ReviewedAt 的 string 值；业务上下文见本 struct 的源码包。 |
| `BackgroundPages` | `int` | ``json:"background_pages"`` | 承载 BackgroundPages 的 int 值；业务上下文见本 struct 的源码包。 |

## `recoveredThought` — `pkg/agent/deck/planner_recovery.go:13`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Thought` | `string` | ``json:"thought"`` | 承载 Thought 的 string 值；业务上下文见本 struct 的源码包。 |

## `recoveredSlideSpec` — `pkg/agent/deck/planner_recovery.go:82`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Title` | `string` | `` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentType` | `string` | `` | 承载 ContentType 的 string 值；业务上下文见本 struct 的源码包。 |

## `reviewCheckpoint` — `pkg/agent/deck/run.go:47`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `NextRound` | `int` | ``json:"next_round"`` | 承载 NextRound 的 int 值；业务上下文见本 struct 的源码包。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `AgentEvent` — `pkg/agent/deck/run.go:111`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Type` | `string` | ``json:"type"`` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `Content` | `string` | ``json:"content,omitempty"`` | 正文或序列化内容。 |
| `ToolName` | `string` | ``json:"tool_name,omitempty"`` | 承载 ToolName 的 string 值；业务上下文见本 struct 的源码包。 |
| `ToolArgs` | `string` | ``json:"tool_args,omitempty"`` | 承载 ToolArgs 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |
| `Phase` | `string` | ``json:"phase,omitempty"`` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhaseDetail` | `string` | ``json:"phase_detail,omitempty"`` | 承载 PhaseDetail 的 string 值；业务上下文见本 struct 的源码包。 |

## `PPTTaskConfig` — `pkg/agent/deck/types.go:40`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `WorkDir` | `string` | `` | 承载 WorkDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | `` | 关联任务的唯一标识。 |
| `Query` | `string` | `` | 用户或系统发起的查询文本。 |
| `Concurrency` | `int` | `` | 承载 Concurrency 的 int 值；业务上下文见本 struct 的源码包。 |
| `Operator` | `commandline.Operator` | `` | 承载 Operator 的 commandline.Operator 值；业务上下文见本 struct 的源码包。 |
| `SkillsDir` | `string` | `` | 承载 SkillsDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `ModelAPIKey` | `string` | `` | 承载 ModelAPIKey 的 string 值；业务上下文见本 struct 的源码包。 |
| `ModelProvider` | `string` | `` | 承载 ModelProvider 的 string 值；业务上下文见本 struct 的源码包。 |
| `CompressorOpt` | `CompressorOption` | `` | 承载 CompressorOpt 的 CompressorOption 值；业务上下文见本 struct 的源码包。 |
| `CompressorTracker` | `*agentutils.TokenTracker` | `` | 承载 CompressorTracker 的 *agentutils.TokenTracker 值；业务上下文见本 struct 的源码包。 |
| `RuntimeMeta` | `*agentutils.RuntimeMeta` | `` | 承载 RuntimeMeta 的 *agentutils.RuntimeMeta 值；业务上下文见本 struct 的源码包。 |
| `Outline` | `*TaskOutline` | `` | 承载 Outline 的 *TaskOutline 值；业务上下文见本 struct 的源码包。 |
| `Intent` | `string` | `` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `ConversationID` | `string` | `` | 关联逻辑会话的唯一标识。 |
| `SourceMessageID` | `string` | `` | 关联消息的唯一标识。 |
| `ParentTaskID` | `string` | `` | 关联任务的唯一标识。 |
| `OnFixerTriggered` | `func()` | `` | 承载 OnFixerTriggered 的 func() 值；业务上下文见本 struct 的源码包。 |
| `UserID` | `int` | `` | 关联用户的唯一标识。 |

## `TasksManifest` — `pkg/agent/deck/types.go:82`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentBank` | `map[string]any` | ``json:"content_bank,omitempty"`` | 承载 ContentBank 的 map[string]any 值；业务上下文见本 struct 的源码包。 |
| `Sections` | `[]DeckSection` | ``json:"sections,omitempty"`` | 承载 Sections 的 []DeckSection 值；业务上下文见本 struct 的源码包。 |
| `VisualPolicy` | `*VisualPolicy` | ``json:"visual_policy,omitempty"`` | 承载 VisualPolicy 的 *VisualPolicy 值；业务上下文见本 struct 的源码包。 |
| `Tasks` | `[]*TaskItem` | ``json:"tasks"`` | 承载 Tasks 的 []*TaskItem 值；业务上下文见本 struct 的源码包。 |

## `VisualPolicy` — `pkg/agent/deck/types.go:93`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Mode` | `string` | ``json:"mode,omitempty"`` | 承载 Mode 的 string 值；业务上下文见本 struct 的源码包。 |
| `MinImagePages` | `int` | ``json:"min_image_pages,omitempty"`` | 承载 MinImagePages 的 int 值；业务上下文见本 struct 的源码包。 |
| `RequiredRoles` | `[]string` | ``json:"required_roles,omitempty"`` | 承载 RequiredRoles 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason,omitempty"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |

## `TaskItem` — `pkg/agent/deck/types.go:100`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `TaskID` | `string` | ``json:"task_id"`` | 关联任务的唯一标识。 |
| `PageIndex` | `int` | ``json:"page_index"`` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `SectionID` | `string` | ``json:"section_id,omitempty"`` | 承载 SectionID 的 string 值；业务上下文见本 struct 的源码包。 |
| `SectionTitle` | `string` | ``json:"section_title,omitempty"`` | 承载 SectionTitle 的 string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentType` | `string` | ``json:"content_type"`` | 承载 ContentType 的 string 值；业务上下文见本 struct 的源码包。 |
| `LayoutVariant` | `string` | ``json:"layout_variant,omitempty"`` | 承载 LayoutVariant 的 string 值；业务上下文见本 struct 的源码包。 |
| `PageIntent` | `string` | ``json:"page_intent,omitempty"`` | 承载 PageIntent 的 string 值；业务上下文见本 struct 的源码包。 |
| `EvidenceRefs` | `[]string` | ``json:"evidence_refs,omitempty"`` | 承载 EvidenceRefs 的 []string 值；业务上下文见本 struct 的源码包。 |
| `OutputFile` | `string` | ``json:"output_file"`` | 承载 OutputFile 的 string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `string` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `ContentPlan` | `*ContentPlan` | ``json:"content_plan,omitempty"`` | 承载 ContentPlan 的 *ContentPlan 值；业务上下文见本 struct 的源码包。 |

## `ManifestValidationReport` — `pkg/agent/deck/types.go:115`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Total` | `int` | ``json:"total"`` | 承载 Total 的 int 值；业务上下文见本 struct 的源码包。 |
| `Done` | `int` | ``json:"done"`` | 承载 Done 的 int 值；业务上下文见本 struct 的源码包。 |
| `MissingFiles` | `[]string` | ``json:"missing_files,omitempty"`` | 承载 MissingFiles 的 []string 值；业务上下文见本 struct 的源码包。 |
| `PendingTasks` | `[]string` | ``json:"pending_tasks,omitempty"`` | 承载 PendingTasks 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Invalid` | `bool` | ``json:"invalid"`` | 承载 Invalid 的 bool 值；业务上下文见本 struct 的源码包。 |

## `TaskOutline` — `pkg/agent/deck/types.go:412`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentMode` | `string` | ``json:"content_mode,omitempty"`` | 承载 ContentMode 的 string 值；业务上下文见本 struct 的源码包。 |
| `RecommendationReason` | `string` | ``json:"recommendation_reason,omitempty"`` | 承载 RecommendationReason 的 string 值；业务上下文见本 struct 的源码包。 |
| `Slides` | `[]SlideOutline` | ``json:"slides"`` | 承载 Slides 的 []SlideOutline 值；业务上下文见本 struct 的源码包。 |

## `DeckSection` — `pkg/agent/deck/types.go:419`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `Summary` | `string` | ``json:"summary,omitempty"`` | 承载 Summary 的 string 值；业务上下文见本 struct 的源码包。 |
| `StartPage` | `int` | ``json:"start_page"`` | 承载 StartPage 的 int 值；业务上下文见本 struct 的源码包。 |
| `EndPage` | `int` | ``json:"end_page"`` | 承载 EndPage 的 int 值；业务上下文见本 struct 的源码包。 |
| `PageCount` | `int` | ``json:"page_count,omitempty"`` | 承载 PageCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `NeighborKey` | `string` | ``json:"neighbor_key,omitempty"`` | 承载 NeighborKey 的 string 值；业务上下文见本 struct 的源码包。 |

## `PlanComponent` — `pkg/agent/deck/types.go:444`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id,omitempty"`` | 当前记录或对象的唯一标识。 |
| `Type` | `string` | ``json:"type"`` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title,omitempty"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `Text` | `string` | ``json:"text,omitempty"`` | 承载 Text 的 string 值；业务上下文见本 struct 的源码包。 |
| `Body` | `string` | ``json:"body,omitempty"`` | 承载 Body 的 string 值；业务上下文见本 struct 的源码包。 |
| `Items` | `[]string` | ``json:"items,omitempty"`` | 承载 Items 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Emphasis` | `string` | ``json:"emphasis,omitempty"`` | 承载 Emphasis 的 string 值；业务上下文见本 struct 的源码包。 |
| `Role` | `string` | ``json:"role,omitempty"`` | 承载 Role 的 string 值；业务上下文见本 struct 的源码包。 |
| `Relation` | `string` | ``json:"relation,omitempty"`` | 承载 Relation 的 string 值；业务上下文见本 struct 的源码包。 |
| `Target` | `string` | ``json:"target,omitempty"`` | 承载 Target 的 string 值；业务上下文见本 struct 的源码包。 |
| `Icon` | `string` | ``json:"icon,omitempty"`` | 承载 Icon 的 string 值；业务上下文见本 struct 的源码包。 |
| `Source` | `string` | ``json:"source,omitempty"`` | 承载 Source 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetPurpose` | `string` | ``json:"asset_purpose,omitempty"`` | 承载 AssetPurpose 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetSubject` | `string` | ``json:"asset_subject,omitempty"`` | 承载 AssetSubject 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetQuery` | `string` | ``json:"asset_query,omitempty"`` | 用户或系统发起的查询文本。 |
| `Composition` | `string` | ``json:"composition,omitempty"`` | 承载 Composition 的 string 值；业务上下文见本 struct 的源码包。 |
| `Provider` | `string` | ``json:"provider,omitempty"`` | 承载 Provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `SearchStatus` | `string` | ``json:"search_status,omitempty"`` | 当前生命周期或处理状态。 |
| `Caption` | `string` | ``json:"caption,omitempty"`` | 承载 Caption 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetID` | `string` | ``json:"asset_id,omitempty"`` | 承载 AssetID 的 string 值；业务上下文见本 struct 的源码包。 |
| `LocalPath` | `string` | ``json:"local_path,omitempty"`` | 承载 LocalPath 的 string 值；业务上下文见本 struct 的源码包。 |
| `ImageURL` | `string` | ``json:"image_url,omitempty"`` | 承载 ImageURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `PreviewURL` | `string` | ``json:"preview_url,omitempty"`` | 承载 PreviewURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `SourceURL` | `string` | ``json:"source_url,omitempty"`` | 承载 SourceURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Attribution` | `string` | ``json:"attribution,omitempty"`` | 承载 Attribution 的 string 值；业务上下文见本 struct 的源码包。 |
| `Data` | `map[string]any` | ``json:"data,omitempty"`` | 承载 Data 的 map[string]any 值；业务上下文见本 struct 的源码包。 |

## `PlanReviewIssue` — `pkg/agent/deck/types.go:532`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Code` | `string` | ``json:"code"`` | 承载 Code 的 string 值；业务上下文见本 struct 的源码包。 |
| `Severity` | `string` | ``json:"severity,omitempty"`` | 承载 Severity 的 string 值；业务上下文见本 struct 的源码包。 |
| `Message` | `string` | ``json:"message,omitempty"`` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |
| `PageIndex` | `int` | ``json:"page_index,omitempty"`` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `ComponentID` | `string` | ``json:"component_id,omitempty"`` | 承载 ComponentID 的 string 值；业务上下文见本 struct 的源码包。 |

## `ContentPlan` — `pkg/agent/deck/types.go:541`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Summary` | `string` | ``json:"summary,omitempty"`` | 承载 Summary 的 string 值；业务上下文见本 struct 的源码包。 |
| `SlideIntent` | `string` | ``json:"slide_intent,omitempty"`` | 承载 SlideIntent 的 string 值；业务上下文见本 struct 的源码包。 |
| `SectionNumber` | `string` | ``json:"section_number,omitempty"`` | 承载 SectionNumber 的 string 值；业务上下文见本 struct 的源码包。 |
| `EvidenceRefs` | `[]string` | ``json:"evidence_refs,omitempty"`` | 承载 EvidenceRefs 的 []string 值；业务上下文见本 struct 的源码包。 |
| `VisualIntent` | `*VisualIntent` | ``json:"visual_intent,omitempty"`` | 承载 VisualIntent 的 *VisualIntent 值；业务上下文见本 struct 的源码包。 |
| `Components` | `[]PlanComponent` | ``json:"components,omitempty"`` | 承载 Components 的 []PlanComponent 值；业务上下文见本 struct 的源码包。 |

## `VisualIntent` — `pkg/agent/deck/types.go:551`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Role` | `string` | ``json:"role,omitempty"`` | 承载 Role 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetPurpose` | `string` | ``json:"asset_purpose,omitempty"`` | 承载 AssetPurpose 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetSubject` | `string` | ``json:"asset_subject,omitempty"`` | 承载 AssetSubject 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetQuery` | `string` | ``json:"asset_query,omitempty"`` | 用户或系统发起的查询文本。 |
| `Composition` | `string` | ``json:"composition,omitempty"`` | 承载 Composition 的 string 值；业务上下文见本 struct 的源码包。 |
| `Provider` | `string` | ``json:"provider,omitempty"`` | 承载 Provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `SearchStatus` | `string` | ``json:"search_status,omitempty"`` | 当前生命周期或处理状态。 |
| `PreferredVariant` | `string` | ``json:"preferred_variant,omitempty"`` | 承载 PreferredVariant 的 string 值；业务上下文见本 struct 的源码包。 |
| `ImagePosition` | `string` | ``json:"image_position,omitempty"`` | 承载 ImagePosition 的 string 值；业务上下文见本 struct 的源码包。 |
| `Caption` | `string` | ``json:"caption,omitempty"`` | 承载 Caption 的 string 值；业务上下文见本 struct 的源码包。 |
| `AssetID` | `string` | ``json:"asset_id,omitempty"`` | 承载 AssetID 的 string 值；业务上下文见本 struct 的源码包。 |
| `LocalPath` | `string` | ``json:"local_path,omitempty"`` | 承载 LocalPath 的 string 值；业务上下文见本 struct 的源码包。 |
| `ImageURL` | `string` | ``json:"image_url,omitempty"`` | 承载 ImageURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `PreviewURL` | `string` | ``json:"preview_url,omitempty"`` | 承载 PreviewURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `SourceURL` | `string` | ``json:"source_url,omitempty"`` | 承载 SourceURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Attribution` | `string` | ``json:"attribution,omitempty"`` | 承载 Attribution 的 string 值；业务上下文见本 struct 的源码包。 |

## `SlideOutline` — `pkg/agent/deck/types.go:571`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentType` | `string` | ``json:"content_type"`` | 承载 ContentType 的 string 值；业务上下文见本 struct 的源码包。 |
| `LayoutVariant` | `string` | ``json:"layout_variant,omitempty"`` | 承载 LayoutVariant 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentPlan` | `*ContentPlan` | ``json:"content_plan,omitempty"`` | 承载 ContentPlan 的 *ContentPlan 值；业务上下文见本 struct 的源码包。 |

## `PPTTaskStart` — `pkg/agent/deck/types.go:578`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Runner` | `*adk.Runner` | `` | 承载 Runner 的 *adk.Runner 值；业务上下文见本 struct 的源码包。 |
| `Iter` | `*adk.AsyncIterator[*adk.AgentEvent]` | `` | 承载 Iter 的 *adk.AsyncIterator[*adk.AgentEvent] 值；业务上下文见本 struct 的源码包。 |
| `CheckpointID` | `string` | `` | 承载 CheckpointID 的 string 值；业务上下文见本 struct 的源码包。 |
| `StartTime` | `time.Time` | `` | 承载 StartTime 的 time.Time 值；业务上下文见本 struct 的源码包。 |

## `PPTTaskResult` — `pkg/agent/deck/types.go:585`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Message` | `string` | `` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |
| `TotalSlides` | `int` | `` | 承载 TotalSlides 的 int 值；业务上下文见本 struct 的源码包。 |
| `DoneSlides` | `int` | `` | 承载 DoneSlides 的 int 值；业务上下文见本 struct 的源码包。 |
| `Files` | `[]string` | `` | 承载 Files 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Duration` | `time.Duration` | `` | 承载 Duration 的 time.Duration 值；业务上下文见本 struct 的源码包。 |

## `ModelSpec` — `pkg/agent/modelcompat/modelcompat.go:34`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Provider` | `Provider` | `` | 承载 Provider 的 Provider 值；业务上下文见本 struct 的源码包。 |
| `Model` | `string` | `` | 承载 Model 的 string 值；业务上下文见本 struct 的源码包。 |
| `APIKey` | `string` | `` | 承载 APIKey 的 string 值；业务上下文见本 struct 的源码包。 |
| `BaseURL` | `string` | `` | 承载 BaseURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Region` | `string` | `` | 承载 Region 的 string 值；业务上下文见本 struct 的源码包。 |
| `Timeout` | `time.Duration` | `` | 承载 Timeout 的 time.Duration 值；业务上下文见本 struct 的源码包。 |
| `MaxTokens` | `*int` | `` | 模型调用的 token 统计或限制参数。 |
| `MaxCompletionTokens` | `*int` | `` | 模型调用的 token 统计或限制参数。 |
| `Temperature` | `*float32` | `` | 承载 Temperature 的 *float32 值；业务上下文见本 struct 的源码包。 |
| `TopP` | `*float32` | `` | 承载 TopP 的 *float32 值；业务上下文见本 struct 的源码包。 |
| `DisableThinking` | `*bool` | `` | 承载 DisableThinking 的 *bool 值；业务上下文见本 struct 的源码包。 |
| `JSONSchema` | `*openaiext.ChatCompletionResponseFormatJSONSchema` | `` | 承载 JSONSchema 的 *openaiext.ChatCompletionResponseFormatJSONSchema 值；业务上下文见本 struct 的源码包。 |
| `ExtraFields` | `map[string]any` | `` | 承载 ExtraFields 的 map[string]any 值；业务上下文见本 struct 的源码包。 |

## `DefaultFactory` — `pkg/agent/modelcompat/modelcompat.go:54`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `Input` — `pkg/agent/router/agent.go:25`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Query` | `string` | `` | 用户或系统发起的查询文本。 |
| `SelectedTaskID` | `string` | `` | 关联任务的唯一标识。 |
| `ConversationContext` | `string` | `` | 承载 ConversationContext 的 string 值；业务上下文见本 struct 的源码包。 |

## `Result` — `pkg/agent/router/agent.go:34`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Intent` | `string` | ``json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `Mode` | `string` | ``json:"mode"`` | 承载 Mode 的 string 值；业务上下文见本 struct 的源码包。 |
| `Confidence` | `float64` | ``json:"confidence"`` | 承载 Confidence 的 float64 值；业务上下文见本 struct 的源码包。 |
| `NeedsConfirmation` | `bool` | ``json:"needs_confirmation"`` | 承载 NeedsConfirmation 的 bool 值；业务上下文见本 struct 的源码包。 |
| `NormalizedRequest` | `string` | ``json:"normalized_request"`` | 承载 NormalizedRequest 的 string 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | ``json:"task_id"`` | 关联任务的唯一标识。 |
| `MissingFields` | `[]string` | ``json:"missing_fields"`` | 承载 MissingFields 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Action` | `string` | ``json:"action"`` | 承载 Action 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason,omitempty"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reply` | `string` | ``json:"reply,omitempty"`` | 承载 Reply 的 string 值；业务上下文见本 struct 的源码包。 |

## `ContinuationInput` — `pkg/agent/router/agent.go:49`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Message` | `string` | `` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |
| `TasksSummary` | `string` | `` | 承载 TasksSummary 的 string 值；业务上下文见本 struct 的源码包。 |

## `ContinuationResult` — `pkg/agent/router/agent.go:56`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Intent` | `string` | ``json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `TargetPages` | `[]int` | ``json:"target_pages,omitempty"`` | 承载 TargetPages 的 []int 值；业务上下文见本 struct 的源码包。 |
| `FixDetails` | `*FixDetails` | ``json:"fix_details,omitempty"`` | 承载 FixDetails 的 *FixDetails 值；业务上下文见本 struct 的源码包。 |
| `RegenerateScope` | `[]int` | ``json:"regenerate_scope,omitempty"`` | 承载 RegenerateScope 的 []int 值；业务上下文见本 struct 的源码包。 |
| `NeedsClarification` | `bool` | ``json:"needs_clarification,omitempty"`` | 承载 NeedsClarification 的 bool 值；业务上下文见本 struct 的源码包。 |
| `ClarificationQuestion` | `string` | ``json:"clarification_question,omitempty"`` | 承载 ClarificationQuestion 的 string 值；业务上下文见本 struct 的源码包。 |

## `FixDetails` — `pkg/agent/router/agent.go:66`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Aspect` | `string` | ``json:"aspect"`` | 承载 Aspect 的 string 值；业务上下文见本 struct 的源码包。 |
| `Detail` | `string` | ``json:"detail"`` | 承载 Detail 的 string 值；业务上下文见本 struct 的源码包。 |
| `TargetElements` | `string` | ``json:"target_elements,omitempty"`` | 承载 TargetElements 的 string 值；业务上下文见本 struct 的源码包。 |

## `Agent` — `pkg/agent/router/agent.go:77`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `generate` | `GenerateFunc` | `` | 承载 generate 的 GenerateFunc 值；业务上下文见本 struct 的源码包。 |
| `timeout` | `time.Duration` | `` | 承载 timeout 的 time.Duration 值；业务上下文见本 struct 的源码包。 |

## `jwtClaims` — `pkg/auth/auth.go:33`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `UserID` | `uint` | ``json:"user_id"`` | 关联用户的唯一标识。 |
| `Email` | `string` | ``json:"email"`` | 承载 Email 的 string 值；业务上下文见本 struct 的源码包。 |
| `IsAdmin` | `bool` | ``json:"is_admin"`` | 承载 IsAdmin 的 bool 值；业务上下文见本 struct 的源码包。 |
| `jwt.RegisteredClaims` | `嵌入字段` | `` | 承载 jwt.RegisteredClaims 的 嵌入字段 值；业务上下文见本 struct 的源码包。 |

## `callbackInputDetailsKey` — `pkg/callback/callback.go:60`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `httpErrorDetail` — `pkg/callback/callback.go:86`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `statusCode` | `int` | `` | 承载 statusCode 的 int 值；业务上下文见本 struct 的源码包。 |
| `requestID` | `string` | `` | 承载 requestID 的 string 值；业务上下文见本 struct 的源码包。 |

## `Event` — `pkg/chattrace/redis.go:25`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint64` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `SegmentID` | `string` | ``json:"segment_id,omitempty"`` | 承载 SegmentID 的 string 值；业务上下文见本 struct 的源码包。 |
| `Type` | `string` | ``json:"type"`` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `Phase` | `string` | ``json:"phase,omitempty"`` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `ToolName` | `string` | ``json:"tool_name,omitempty"`` | 承载 ToolName 的 string 值；业务上下文见本 struct 的源码包。 |
| `Detail` | `string` | ``json:"detail,omitempty"`` | 承载 Detail 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |
| `Preview` | `map[string]any` | ``json:"preview,omitempty"`` | 承载 Preview 的 map[string]any 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |

## `RedisStore` — `pkg/chattrace/redis.go:42`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `client` | `*redis.Client` | `` | 承载 client 的 *redis.Client 值；业务上下文见本 struct 的源码包。 |
| `ttl` | `time.Duration` | `` | 承载 ttl 的 time.Duration 值；业务上下文见本 struct 的源码包。 |

## `AdminMetrics` — `pkg/db/admin.go:7`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `RegisteredUserCount` | `int64` | ``json:"registered_user_count"`` | 承载 RegisteredUserCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `NonRootRegisteredUserCount` | `int64` | ``json:"non_root_registered_user_count"`` | 承载 NonRootRegisteredUserCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `PPTActiveUserCount` | `int64` | ``json:"ppt_active_user_count"`` | 承载 PPTActiveUserCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `CustomAPIKeyUserCount` | `int64` | ``json:"custom_api_key_user_count"`` | 承载 CustomAPIKeyUserCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `PPTGenerationCount` | `int64` | ``json:"ppt_generation_count"`` | 承载 PPTGenerationCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `NonRootPPTGenerationCount` | `int64` | ``json:"non_root_ppt_generation_count"`` | 承载 NonRootPPTGenerationCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `FeedbackCount` | `int64` | ``json:"feedback_count"`` | 承载 FeedbackCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `FeedbackSuggestionCount` | `int64` | ``json:"feedback_suggestion_count"`` | 承载 FeedbackSuggestionCount 的 int64 值；业务上下文见本 struct 的源码包。 |

## `AdminUserMetric` — `pkg/db/admin.go:19`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `User` | `嵌入字段` | `` | 承载 User 的 嵌入字段 值；业务上下文见本 struct 的源码包。 |
| `PPTGenerationCount` | `int64` | ``json:"ppt_generation_count"`` | 承载 PPTGenerationCount 的 int64 值；业务上下文见本 struct 的源码包。 |
| `CustomAPIKeyConfigured` | `bool` | ``json:"custom_api_key_configured"`` | 承载 CustomAPIKeyConfigured 的 bool 值；业务上下文见本 struct 的源码包。 |

## `AdminTaskFeedback` — `pkg/db/admin.go:27`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `TaskFeedback` | `嵌入字段` | `` | 承载 TaskFeedback 的 嵌入字段 值；业务上下文见本 struct 的源码包。 |
| `UserEmail` | `string` | ``json:"user_email"`` | 承载 UserEmail 的 string 值；业务上下文见本 struct 的源码包。 |
| `TaskQuery` | `string` | ``json:"task_query"`` | 用户或系统发起的查询文本。 |

## `User` — `pkg/db/internal/model/account.go:6`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint` | ``gorm:"primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `Email` | `string` | ``gorm:"size:120;uniqueIndex;not null" json:"email"`` | 承载 Email 的 string 值；业务上下文见本 struct 的源码包。 |
| `Password` | `string` | ``gorm:"size:255" json:"-"`` | 承载 Password 的 string 值；业务上下文见本 struct 的源码包。 |
| `IsAdmin` | `bool` | ``gorm:"default:false" json:"is_admin"`` | 承载 IsAdmin 的 bool 值；业务上下文见本 struct 的源码包。 |
| `GuestIPHash` | `string` | ``gorm:"size:64;index" json:"-"`` | 承载 GuestIPHash 的 string 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |

## `UserAPIKey` — `pkg/db/internal/model/account.go:17`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `UserID` | `uint` | ``gorm:"primaryKey" json:"user_id"`` | 关联用户的唯一标识。 |
| `Provider` | `string` | ``gorm:"size:40;not null;default:'ark'" json:"provider"`` | 承载 Provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `APIKey` | `string` | ``gorm:"type:text;not null" json:"-"`` | 承载 APIKey 的 string 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `VerificationCode` — `pkg/db/internal/model/account.go:26`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint` | ``gorm:"primaryKey" json:"-"`` | 当前记录或对象的唯一标识。 |
| `Email` | `string` | ``gorm:"size:120;index;not null" json:"email"`` | 承载 Email 的 string 值；业务上下文见本 struct 的源码包。 |
| `Code` | `string` | ``gorm:"size:10;not null" json:"-"`` | 承载 Code 的 string 值；业务上下文见本 struct 的源码包。 |
| `Used` | `bool` | ``gorm:"default:false" json:"-"`` | 承载 Used 的 bool 值；业务上下文见本 struct 的源码包。 |
| `ExpiresAt` | `time.Time` | ``gorm:"not null" json:"expires_at"`` | 承载 ExpiresAt 的 time.Time 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |

## `ConversationMessage` — `pkg/db/internal/model/conversation.go:6`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint` | ``gorm:"primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `TaskID` | `string` | ``gorm:"size:64;index;not null" json:"task_id"`` | 关联任务的唯一标识。 |
| `Role` | `string` | ``gorm:"size:20;not null" json:"role"`` | 承载 Role 的 string 值；业务上下文见本 struct 的源码包。 |
| `Content` | `string` | ``gorm:"type:longtext;not null" json:"content"`` | 正文或序列化内容。 |
| `Timestamp` | `time.Time` | ``json:"timestamp"`` | 事件或消息发生时间。 |
| `Chunks` | `[]ConversationMessageChunk` | ``gorm:"foreignKey:MessageID" json:"-"`` | 承载 Chunks 的 []ConversationMessageChunk 值；业务上下文见本 struct 的源码包。 |

## `ConversationMessageChunk` — `pkg/db/internal/model/conversation.go:17`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint` | ``gorm:"primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `MessageID` | `uint` | ``gorm:"uniqueIndex:idx_message_chunk_sequence;index;not null" json:"message_id"`` | 关联消息的唯一标识。 |
| `Sequence` | `int` | ``gorm:"uniqueIndex:idx_message_chunk_sequence;not null" json:"sequence"`` | 承载 Sequence 的 int 值；业务上下文见本 struct 的源码包。 |
| `Content` | `string` | ``gorm:"type:longtext;not null" json:"content"`` | 正文或序列化内容。 |

## `TaskErrorAnalysis` — `pkg/db/internal/model/runtime.go:6`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint` | ``gorm:"primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `TaskID` | `string` | ``gorm:"size:64;index;not null" json:"task_id"`` | 关联任务的唯一标识。 |
| `TriggerType` | `string` | ``gorm:"size:20;not null" json:"trigger_type"`` | 承载 TriggerType 的 string 值；业务上下文见本 struct 的源码包。 |
| `LogSnippet` | `string` | ``gorm:"type:longtext" json:"log_snippet"`` | 承载 LogSnippet 的 string 值；业务上下文见本 struct 的源码包。 |
| `Analysis` | `string` | ``gorm:"type:longtext" json:"analysis"`` | 承载 Analysis 的 string 值；业务上下文见本 struct 的源码包。 |
| `RootCause` | `string` | ``gorm:"type:text" json:"root_cause"`` | 承载 RootCause 的 string 值；业务上下文见本 struct 的源码包。 |
| `Suggestion` | `string` | ``gorm:"type:text" json:"suggestion"`` | 承载 Suggestion 的 string 值；业务上下文见本 struct 的源码包。 |
| `TokensUsed` | `int64` | ``gorm:"default:0" json:"tokens_used"`` | 模型调用的 token 统计或限制参数。 |
| `ModelUsed` | `string` | ``gorm:"size:100" json:"model_used"`` | 承载 ModelUsed 的 string 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |

## `RuntimeEventRecord` — `pkg/db/internal/model/runtime.go:20`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint` | ``gorm:"primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `TaskID` | `string` | ``gorm:"size:64;uniqueIndex:idx_runtime_event_task_seq;index;not null" json:"task_id"`` | 关联任务的唯一标识。 |
| `EventID` | `int64` | ``gorm:"uniqueIndex:idx_runtime_event_task_seq;not null" json:"event_id"`` | 承载 EventID 的 int64 值；业务上下文见本 struct 的源码包。 |
| `Timestamp` | `time.Time` | ``gorm:"index" json:"timestamp"`` | 事件或消息发生时间。 |
| `ElapsedMS` | `int64` | ``gorm:"default:0" json:"elapsed_ms"`` | 承载 ElapsedMS 的 int64 值；业务上下文见本 struct 的源码包。 |
| `Kind` | `string` | ``gorm:"size:64;index" json:"kind"`` | 承载 Kind 的 string 值；业务上下文见本 struct 的源码包。 |
| `Phase` | `string` | ``gorm:"size:64;index" json:"phase"`` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `Name` | `string` | ``gorm:"size:128;index" json:"name"`` | 承载 Name 的 string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `string` | ``gorm:"size:32;index" json:"status"`` | 当前生命周期或处理状态。 |
| `Detail` | `string` | ``gorm:"type:text" json:"detail"`` | 承载 Detail 的 string 值；业务上下文见本 struct 的源码包。 |
| `Metadata` | `string` | ``gorm:"type:longtext" json:"metadata"`` | 结构化补充元数据；输出前需按权限脱敏。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |

## `TaskRecord` — `pkg/db/internal/model/task.go:6`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``gorm:"size:64;primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `UserID` | `uint` | ``gorm:"index;not null" json:"user_id"`` | 关联用户的唯一标识。 |
| `Query` | `string` | ``gorm:"type:text" json:"query"`` | 用户或系统发起的查询文本。 |
| `Status` | `string` | ``gorm:"size:20;not null;default:'running'" json:"status"`` | 当前生命周期或处理状态。 |
| `WorkDir` | `string` | ``gorm:"size:512" json:"work_dir"`` | 承载 WorkDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `DoneCount` | `int` | ``gorm:"default:0" json:"done_count"`` | 承载 DoneCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `TotalCount` | `int` | ``gorm:"default:0" json:"total_count"`` | 承载 TotalCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `Duration` | `string` | ``gorm:"size:50" json:"duration"`` | 承载 Duration 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``gorm:"type:text" json:"error"`` | 失败原因或错误详情。 |
| `PromptTokens` | `int64` | ``gorm:"default:0" json:"prompt_tokens"`` | 模型调用的 token 统计或限制参数。 |
| `CompletionTokens` | `int64` | ``gorm:"default:0" json:"completion_tokens"`` | 模型调用的 token 统计或限制参数。 |
| `TotalTokens` | `int64` | ``gorm:"default:0" json:"total_tokens"`` | 模型调用的 token 统计或限制参数。 |
| `Files` | `string` | ``gorm:"type:text" json:"files"`` | 承载 Files 的 string 值；业务上下文见本 struct 的源码包。 |
| `ConversationContent` | `string` | ``gorm:"type:longtext" json:"conversation_content"`` | 正文或序列化内容。 |
| `FullAnswer` | `string` | ``gorm:"type:longtext" json:"full_answer"`` | 承载 FullAnswer 的 string 值；业务上下文见本 struct 的源码包。 |
| `Intent` | `string` | ``gorm:"size:32;index" json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `ConversationID` | `string` | ``gorm:"size:64;index" json:"conversation_id"`` | 关联逻辑会话的唯一标识。 |
| `SourceMessageID` | `string` | ``gorm:"size:64;index" json:"source_message_id"`` | 关联消息的唯一标识。 |
| `ParentTaskID` | `string` | ``gorm:"size:64;index" json:"parent_task_id"`` | 关联任务的唯一标识。 |
| `GenerationStartedAt` | `*time.Time` | ``gorm:"index" json:"generation_started_at,omitempty"`` | 承载 GenerationStartedAt 的 *time.Time 值；业务上下文见本 struct 的源码包。 |
| `GenerationFinishedAt` | `*time.Time` | ``json:"generation_finished_at,omitempty"`` | 承载 GenerationFinishedAt 的 *time.Time 值；业务上下文见本 struct 的源码包。 |
| `GenerationDurationMS` | `int64` | ``gorm:"default:0" json:"generation_duration_ms"`` | 承载 GenerationDurationMS 的 int64 值；业务上下文见本 struct 的源码包。 |
| `FixerRunCount` | `int` | ``gorm:"default:0" json:"fixer_run_count"`` | 承载 FixerRunCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `TaskFeedback` — `pkg/db/internal/model/task.go:35`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint` | ``gorm:"primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `TaskID` | `string` | ``gorm:"size:64;uniqueIndex:idx_task_feedback_task_user;index;not null" json:"task_id"`` | 关联任务的唯一标识。 |
| `UserID` | `uint` | ``gorm:"uniqueIndex:idx_task_feedback_task_user;index;not null" json:"user_id"`` | 关联用户的唯一标识。 |
| `Rating` | `int` | ``gorm:"not null" json:"rating"`` | 承载 Rating 的 int 值；业务上下文见本 struct 的源码包。 |
| `Suggestion` | `string` | ``gorm:"type:text" json:"suggestion"`` | 承载 Suggestion 的 string 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `PlanDraftRecord` — `pkg/db/internal/model/task.go:46`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``gorm:"size:64;primaryKey" json:"id"`` | 当前记录或对象的唯一标识。 |
| `UserID` | `uint` | ``gorm:"index;not null" json:"user_id"`` | 关联用户的唯一标识。 |
| `ConversationID` | `string` | ``gorm:"size:64;index" json:"conversation_id"`` | 关联逻辑会话的唯一标识。 |
| `SourceMessageID` | `string` | ``gorm:"size:64;index" json:"source_message_id"`` | 关联消息的唯一标识。 |
| `Query` | `string` | ``gorm:"type:text" json:"query"`` | 用户或系统发起的查询文本。 |
| `NormalizedRequest` | `string` | ``gorm:"type:text" json:"normalized_request"`` | 承载 NormalizedRequest 的 string 值；业务上下文见本 struct 的源码包。 |
| `DraftContent` | `string` | ``gorm:"type:longtext" json:"draft_content"`` | 正文或序列化内容。 |
| `Status` | `string` | ``gorm:"size:32;index;not null;default:'draft'" json:"status"`` | 当前生命周期或处理状态。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `Manager` — `pkg/human/manager.go:23`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `interactive` | `bool` | `` | 承载 interactive 的 bool 值；业务上下文见本 struct 的源码包。 |

## `Result` — `pkg/log_analysis/service.go:30`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Analysis` | `string` | `` | 承载 Analysis 的 string 值；业务上下文见本 struct 的源码包。 |
| `RootCause` | `string` | `` | 承载 RootCause 的 string 值；业务上下文见本 struct 的源码包。 |
| `Suggestion` | `string` | `` | 承载 Suggestion 的 string 值；业务上下文见本 struct 的源码包。 |
| `TokensUsed` | `int64` | `` | 模型调用的 token 统计或限制参数。 |
| `ModelUsed` | `string` | `` | 承载 ModelUsed 的 string 值；业务上下文见本 struct 的源码包。 |

## `LLMAnalyzer` — `pkg/log_analysis/service.go:41`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `modelFactory` | `func(ctx context.Context) (model.ToolCallingChatModel, error)` | `` | 承载 modelFactory 的 func(ctx context.Context) (model.ToolCallingChatModel, error) 值；业务上下文见本 struct 的源码包。 |
| `skillsDir` | `string` | `` | 承载 skillsDir 的 string 值；业务上下文见本 struct 的源码包。 |

## `readFileOperator` — `pkg/log_analysis/service.go:298`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `workDir` | `string` | `` | 承载 workDir 的 string 值；业务上下文见本 struct 的源码包。 |

## `Service` — `pkg/log_analysis/service.go:367`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `analyzer` | `Analyzer` | `` | 承载 analyzer 的 Analyzer 值；业务上下文见本 struct 的源码包。 |
| `modelFactory` | `func(ctx context.Context) (model.ToolCallingChatModel, error)` | `` | 承载 modelFactory 的 func(ctx context.Context) (model.ToolCallingChatModel, error) 值；业务上下文见本 struct 的源码包。 |
| `skillsDir` | `string` | `` | 承载 skillsDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `idleInterval` | `time.Duration` | `` | 承载 idleInterval 的 time.Duration 值；业务上下文见本 struct 的源码包。 |
| `logLines` | `int` | `` | 承载 logLines 的 int 值；业务上下文见本 struct 的源码包。 |
| `mu` | `sync.Mutex` | `` | 承载 mu 的 sync.Mutex 值；业务上下文见本 struct 的源码包。 |
| `stopCh` | `chan struct{}` | `` | 承载 stopCh 的 chan struct{} 值；业务上下文见本 struct 的源码包。 |
| `pendingTask` | `chan taskRequest` | `` | 承载 pendingTask 的 chan taskRequest 值；业务上下文见本 struct 的源码包。 |
| `running` | `bool` | `` | 承载 running 的 bool 值；业务上下文见本 struct 的源码包。 |
| `lastFileOffset` | `int64` | `` | 承载 lastFileOffset 的 int64 值；业务上下文见本 struct 的源码包。 |

## `taskRequest` — `pkg/log_analysis/service.go:382`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `TaskID` | `string` | `` | 关联任务的唯一标识。 |
| `TriggerType` | `string` | `` | 承载 TriggerType 的 string 值；业务上下文见本 struct 的源码包。 |
| `Logs` | `string` | `` | 承载 Logs 的 string 值；业务上下文见本 struct 的源码包。 |

## `ServiceConfig` — `pkg/log_analysis/service.go:389`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ModelFactory` | `func(ctx context.Context) (model.ToolCallingChatModel, error)` | `` | 承载 ModelFactory 的 func(ctx context.Context) (model.ToolCallingChatModel, error) 值；业务上下文见本 struct 的源码包。 |
| `IdleInterval` | `time.Duration` | `` | 承载 IdleInterval 的 time.Duration 值；业务上下文见本 struct 的源码包。 |
| `LogLines` | `int` | `` | 承载 LogLines 的 int 值；业务上下文见本 struct 的源码包。 |
| `SkillsDir` | `string` | `` | 承载 SkillsDir 的 string 值；业务上下文见本 struct 的源码包。 |

## `TemplateData` — `pkg/prompts/prompts.go:58`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `SkillsDir` | `string` | `` | 承载 SkillsDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `TasksJSON` | `string` | `` | 承载 TasksJSON 的 string 值；业务上下文见本 struct 的源码包。 |
| `FixerTaskSnapshot` | `string` | `` | 承载 FixerTaskSnapshot 的 string 值；业务上下文见本 struct 的源码包。 |
| `OutlineQuery` | `string` | `` | 用户或系统发起的查询文本。 |

## `LogAnalysisData` — `pkg/prompts/prompts.go:111`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `SkillsDir` | `string` | `` | 承载 SkillsDir 的 string 值；业务上下文见本 struct 的源码包。 |

## `Policy` — `pkg/retry/retry.go:25`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `MaxAttempts` | `int` | `` | 承载 MaxAttempts 的 int 值；业务上下文见本 struct 的源码包。 |
| `Cooldown` | `time.Duration` | `` | 承载 Cooldown 的 time.Duration 值；业务上下文见本 struct 的源码包。 |
| `MaxCooldown` | `time.Duration` | `` | 承载 MaxCooldown 的 time.Duration 值；业务上下文见本 struct 的源码包。 |

## `RetryableError` — `pkg/retry/retry.go:49`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Class` | `ErrorClass` | `` | 承载 Class 的 ErrorClass 值；业务上下文见本 struct 的源码包。 |
| `Cause` | `error` | `` | 承载 Cause 的 error 值；业务上下文见本 struct 的源码包。 |

## `FixedAttemptStrategy` — `pkg/retry/retry.go:130`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Operation` | `Operation` | `` | 承载 Operation 的 Operation 值；业务上下文见本 struct 的源码包。 |
| `StrategyName` | `string` | `` | 承载 StrategyName 的 string 值；业务上下文见本 struct 的源码包。 |
| `MaxAttempts` | `int` | `` | 承载 MaxAttempts 的 int 值；业务上下文见本 struct 的源码包。 |

## `Decision` — `pkg/retry/retry.go:148`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `StrategyName` | `string` | `` | 承载 StrategyName 的 string 值；业务上下文见本 struct 的源码包。 |
| `Policy` | `Policy` | `` | 承载 Policy 的 Policy 值；业务上下文见本 struct 的源码包。 |

## `Factory` — `pkg/retry/retry.go:154`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `strategies` | `[]Strategy` | `` | 承载 strategies 的 []Strategy 值；业务上下文见本 struct 的源码包。 |

## `HTTP410SearchTermRevisionStrategy` — `pkg/retry/retry.go:246`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `MaxAttempts` | `int` | `` | 承载 MaxAttempts 的 int 值；业务上下文见本 struct 的源码包。 |

## `RateLimitFallbackStrategy` — `pkg/retry/retry.go:269`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `TransientModelRetryStrategy` — `pkg/retry/retry.go:291`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `ModelStreamReadFallbackStrategy` — `pkg/retry/retry.go:305`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `CompressorConfig` — `pkg/runtime/model/compressor.go:36`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `MessageThreshold` | `int` | `` | 承载 MessageThreshold 的 int 值；业务上下文见本 struct 的源码包。 |
| `TokenThreshold` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `PreserveCount` | `int` | `` | 承载 PreserveCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `ToolResultPreserveCount` | `int` | `` | 承载 ToolResultPreserveCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `MinMessagesSinceLastCompression` | `int` | `` | 承载 MinMessagesSinceLastCompression 的 int 值；业务上下文见本 struct 的源码包。 |
| `MinTokensSinceLastCompression` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `MinCompressionInterval` | `time.Duration` | `` | 承载 MinCompressionInterval 的 time.Duration 值；业务上下文见本 struct 的源码包。 |

## `CompressionEvent` — `pkg/runtime/model/compressor.go:56`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Stage` | `string` | `` | 承载 Stage 的 string 值；业务上下文见本 struct 的源码包。 |
| `BeforeMessages` | `int` | `` | 承载 BeforeMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `AfterMessages` | `int` | `` | 承载 AfterMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `BeforeTokens` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `AfterTokens` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `Error` | `string` | `` | 失败原因或错误详情。 |

## `CompressionSummary` — `pkg/runtime/model/compressor.go:134`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `UserRequestSummary` | `string` | ``json:"user_request_summary"`` | 承载 UserRequestSummary 的 string 值；业务上下文见本 struct 的源码包。 |
| `PreservedRequirements` | `[]string` | ``json:"preserved_requirements"`` | 承载 PreservedRequirements 的 []string 值；业务上下文见本 struct 的源码包。 |
| `KeyDecisions` | `struct {` | `` | 承载 KeyDecisions 的 struct { 值；业务上下文见本 struct 的源码包。 |
| `Template` | `string` | ``json:"template,omitempty"`` | 承载 Template 的 string 值；业务上下文见本 struct 的源码包。 |
| `ColorScheme` | `string` | ``json:"color_scheme,omitempty"`` | 承载 ColorScheme 的 string 值；业务上下文见本 struct 的源码包。 |
| `Theme` | `string` | ``json:"theme,omitempty"`` | 承载 Theme 的 string 值；业务上下文见本 struct 的源码包。 |
| `TotalPages` | `int` | ``json:"total_pages,omitempty"`` | 承载 TotalPages 的 int 值；业务上下文见本 struct 的源码包。 |
| `SlideTypes` | `[]string` | ``json:"slide_types,omitempty"`` | 承载 SlideTypes 的 []string 值；业务上下文见本 struct 的源码包。 |
| `OtherDecisions` | `[]string` | ``json:"other_decisions,omitempty"`` | 承载 OtherDecisions 的 []string 值；业务上下文见本 struct 的源码包。 |
| `}` | `嵌入字段` | ``json:"key_decisions"`` | 承载 } 的 嵌入字段 值；业务上下文见本 struct 的源码包。 |
| `ProgressSummary` | `string` | ``json:"progress_summary"`` | 承载 ProgressSummary 的 string 值；业务上下文见本 struct 的源码包。 |
| `ConversationSummary` | `string` | ``json:"conversation_summary"`` | 承载 ConversationSummary 的 string 值；业务上下文见本 struct 的源码包。 |

## `ChatModelCompressor` — `pkg/runtime/model/compressor.go:352`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `inner` | `model.ToolCallingChatModel` | `` | 承载 inner 的 model.ToolCallingChatModel 值；业务上下文见本 struct 的源码包。 |
| `summarizer` | `model.ToolCallingChatModel` | `` | 承载 summarizer 的 model.ToolCallingChatModel 值；业务上下文见本 struct 的源码包。 |
| `cfg` | `*CompressorConfig` | `` | 承载 cfg 的 *CompressorConfig 值；业务上下文见本 struct 的源码包。 |
| `tracker` | `*TokenTracker` | `` | 承载 tracker 的 *TokenTracker 值；业务上下文见本 struct 的源码包。 |
| `runtime` | `*RuntimeMeta` | `` | 承载 runtime 的 *RuntimeMeta 值；业务上下文见本 struct 的源码包。 |
| `onEvent` | `CompressionEventCallback` | `` | 承载 onEvent 的 CompressionEventCallback 值；业务上下文见本 struct 的源码包。 |
| `state` | `*compressionState` | `` | 承载 state 的 *compressionState 值；业务上下文见本 struct 的源码包。 |

## `compressionState` — `pkg/runtime/model/compressor.go:362`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.Mutex` | `` | 承载 mu 的 sync.Mutex 值；业务上下文见本 struct 的源码包。 |
| `lastMessages` | `int` | `` | 承载 lastMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `lastTokens` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `lastAt` | `time.Time` | `` | 承载 lastAt 的 time.Time 值；业务上下文见本 struct 的源码包。 |

## `pairWithToolCalls` — `pkg/runtime/model/compressor.go:732`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `user` | `*schema.Message` | `` | 承载 user 的 *schema.Message 值；业务上下文见本 struct 的源码包。 |
| `assistant` | `*schema.Message` | `` | 承载 assistant 的 *schema.Message 值；业务上下文见本 struct 的源码包。 |
| `toolResults` | `[]*schema.Message` | `` | 承载 toolResults 的 []*schema.Message 值；业务上下文见本 struct 的源码包。 |

## `ChatModelConfig` — `pkg/runtime/model/model.go:46`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `MaxTokens` | `*int` | `` | 模型调用的 token 统计或限制参数。 |
| `MaxCompletionTokens` | `*int` | `` | 模型调用的 token 统计或限制参数。 |
| `Temperature` | `*float32` | `` | 承载 Temperature 的 *float32 值；业务上下文见本 struct 的源码包。 |
| `TopP` | `*float32` | `` | 承载 TopP 的 *float32 值；业务上下文见本 struct 的源码包。 |
| `DisableThinking` | `*bool` | `` | 承载 DisableThinking 的 *bool 值；业务上下文见本 struct 的源码包。 |
| `JsonSchema` | `*openai.ChatCompletionResponseFormatJSONSchema` | `` | 承载 JsonSchema 的 *openai.ChatCompletionResponseFormatJSONSchema 值；业务上下文见本 struct 的源码包。 |
| `Model` | `*string` | `` | 承载 Model 的 *string 值；业务上下文见本 struct 的源码包。 |
| `APIKey` | `*string` | `` | 承载 APIKey 的 *string 值；业务上下文见本 struct 的源码包。 |
| `APIKeyProvider` | `*modelcompat.Provider` | `` | 承载 APIKeyProvider 的 *modelcompat.Provider 值；业务上下文见本 struct 的源码包。 |
| `ModelRole` | `string` | `` | 承载 ModelRole 的 string 值；业务上下文见本 struct 的源码包。 |

## `globalRateLimitTracker` — `pkg/runtime/model/model.go:206`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.Mutex` | `` | 承载 mu 的 sync.Mutex 值；业务上下文见本 struct 的源码包。 |
| `pauseEndTimes` | `map[string]time.Time` | `` | 承载 pauseEndTimes 的 map[string]time.Time 值；业务上下文见本 struct 的源码包。 |

## `modelCallLimiter` — `pkg/runtime/model/model.go:215`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.Mutex` | `` | 承载 mu 的 sync.Mutex 值；业务上下文见本 struct 的源码包。 |
| `slots` | `map[string]chan struct{}` | `` | 承载 slots 的 map[string]chan struct{} 值；业务上下文见本 struct 的源码包。 |

## `FallbackChatModel` — `pkg/runtime/model/model.go:285`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `models` | `[]model.ToolCallingChatModel` | `` | 承载 models 的 []model.ToolCallingChatModel 值；业务上下文见本 struct 的源码包。 |
| `modelNames` | `[]string` | `` | 承载 modelNames 的 []string 值；业务上下文见本 struct 的源码包。 |
| `rawNames` | `[]string` | `` | 承载 rawNames 的 []string 值；业务上下文见本 struct 的源码包。 |
| `trackerNames` | `[]string` | `` | 承载 trackerNames 的 []string 值；业务上下文见本 struct 的源码包。 |
| `resourceKeys` | `[]string` | `` | 承载 resourceKeys 的 []string 值；业务上下文见本 struct 的源码包。 |
| `profiles` | `[]modelCallProfile` | `` | 承载 profiles 的 []modelCallProfile 值；业务上下文见本 struct 的源码包。 |
| `mu` | `sync.Mutex` | `` | 承载 mu 的 sync.Mutex 值；业务上下文见本 struct 的源码包。 |
| `pauseDuration` | `time.Duration` | `` | 承载 pauseDuration 的 time.Duration 值；业务上下文见本 struct 的源码包。 |
| `compressorCfg` | `*compressorConfig` | `` | 承载 compressorCfg 的 *compressorConfig 值；业务上下文见本 struct 的源码包。 |

## `modelCallProfile` — `pkg/runtime/model/model.go:299`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Provider` | `string` | `` | 承载 Provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `Model` | `string` | `` | 承载 Model 的 string 值；业务上下文见本 struct 的源码包。 |
| `Timeout` | `time.Duration` | `` | 承载 Timeout 的 time.Duration 值；业务上下文见本 struct 的源码包。 |

## `compressorConfig` — `pkg/runtime/model/model.go:1378`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `summarizerFactory` | `func() (model.ToolCallingChatModel, error)` | `` | 承载 summarizerFactory 的 func() (model.ToolCallingChatModel, error) 值；业务上下文见本 struct 的源码包。 |
| `messageThreshold` | `int` | `` | 承载 messageThreshold 的 int 值；业务上下文见本 struct 的源码包。 |
| `tokenThreshold` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `preserveCount` | `int` | `` | 承载 preserveCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `toolResultPreserveCount` | `int` | `` | 承载 toolResultPreserveCount 的 int 值；业务上下文见本 struct 的源码包。 |

## `RuntimeMeta` — `pkg/runtime/model/runtime_meta.go:48`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.RWMutex` | `` | 承载 mu 的 sync.RWMutex 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | `` | 关联任务的唯一标识。 |
| `WorkDir` | `string` | `` | 承载 WorkDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `StartedAt` | `time.Time` | `` | 承载 StartedAt 的 time.Time 值；业务上下文见本 struct 的源码包。 |
| `Phase` | `string` | `` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhaseDetail` | `string` | `` | 承载 PhaseDetail 的 string 值；业务上下文见本 struct 的源码包。 |
| `LastError` | `string` | `` | 失败原因或错误详情。 |
| `LastTool` | `string` | `` | 承载 LastTool 的 string 值；业务上下文见本 struct 的源码包。 |
| `LastToolArgs` | `string` | `` | 承载 LastToolArgs 的 string 值；业务上下文见本 struct 的源码包。 |
| `ToolCalls` | `map[string]int` | `` | 承载 ToolCalls 的 map[string]int 值；业务上下文见本 struct 的源码包。 |
| `ToolErrors` | `map[string]int` | `` | 承载 ToolErrors 的 map[string]int 值；业务上下文见本 struct 的源码包。 |
| `SameToolArgsStreak` | `int` | `` | 承载 SameToolArgsStreak 的 int 值；业务上下文见本 struct 的源码包。 |
| `PromptTokens` | `int64` | `` | 模型调用的 token 统计或限制参数。 |
| `CompletionTokens` | `int64` | `` | 模型调用的 token 统计或限制参数。 |
| `TotalTokens` | `int64` | `` | 模型调用的 token 统计或限制参数。 |
| `CompressionBeforeTokens` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `CompressionAfterTokens` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `CompressionSavedPct` | `string` | `` | 承载 CompressionSavedPct 的 string 值；业务上下文见本 struct 的源码包。 |
| `CompressionBeforeMessages` | `int` | `` | 承载 CompressionBeforeMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `CompressionAfterMessages` | `int` | `` | 承载 CompressionAfterMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `CompressionRemovedMessages` | `int` | `` | 承载 CompressionRemovedMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `CompressionSavedTokens` | `int` | `` | 模型调用的 token 统计或限制参数。 |
| `Budgets` | `RuntimeBudgets` | `` | 承载 Budgets 的 RuntimeBudgets 值；业务上下文见本 struct 的源码包。 |
| `BudgetWarnings` | `[]string` | `` | 承载 BudgetWarnings 的 []string 值；业务上下文见本 struct 的源码包。 |
| `DoneSlides` | `int` | `` | 承载 DoneSlides 的 int 值；业务上下文见本 struct 的源码包。 |
| `TotalSlides` | `int` | `` | 承载 TotalSlides 的 int 值；业务上下文见本 struct 的源码包。 |
| `MissingFiles` | `int` | `` | 承载 MissingFiles 的 int 值；业务上下文见本 struct 的源码包。 |
| `QAHighIssues` | `int` | `` | 承载 QAHighIssues 的 int 值；业务上下文见本 struct 的源码包。 |
| `QAMediumIssues` | `int` | `` | 承载 QAMediumIssues 的 int 值；业务上下文见本 struct 的源码包。 |
| `QALowIssues` | `int` | `` | 承载 QALowIssues 的 int 值；业务上下文见本 struct 的源码包。 |
| `TaskInput` | `TaskInputAnchor` | `` | 承载 TaskInput 的 TaskInputAnchor 值；业务上下文见本 struct 的源码包。 |
| `PlanSlides` | `[]PlanSlide` | `` | 承载 PlanSlides 的 []PlanSlide 值；业务上下文见本 struct 的源码包。 |
| `CurrentSlide` | `*PlanSlide` | `` | 承载 CurrentSlide 的 *PlanSlide 值；业务上下文见本 struct 的源码包。 |
| `AlignmentStatus` | `string` | `` | 当前生命周期或处理状态。 |
| `AlignmentWarnings` | `[]AlignmentWarning` | `` | 承载 AlignmentWarnings 的 []AlignmentWarning 值；业务上下文见本 struct 的源码包。 |
| `EventSeq` | `int64` | `` | 承载 EventSeq 的 int64 值；业务上下文见本 struct 的源码包。 |
| `EventCounts` | `map[string]int` | `` | 承载 EventCounts 的 map[string]int 值；业务上下文见本 struct 的源码包。 |
| `RecentEvents` | `[]RuntimeEvent` | `` | 承载 RecentEvents 的 []RuntimeEvent 值；业务上下文见本 struct 的源码包。 |
| `EventSink` | `RuntimeEventSink` | `` | 承载 EventSink 的 RuntimeEventSink 值；业务上下文见本 struct 的源码包。 |
| `lastManifestValidation` | `*manifestValidationState` | `` | 承载 lastManifestValidation 的 *manifestValidationState 值；业务上下文见本 struct 的源码包。 |

## `manifestValidationState` — `pkg/runtime/model/runtime_meta.go:101`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Done` | `int` | `` | 承载 Done 的 int 值；业务上下文见本 struct 的源码包。 |
| `Total` | `int` | `` | 承载 Total 的 int 值；业务上下文见本 struct 的源码包。 |
| `MissingFiles` | `[]string` | `` | 承载 MissingFiles 的 []string 值；业务上下文见本 struct 的源码包。 |
| `PendingTasks` | `[]string` | `` | 承载 PendingTasks 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `string` | `` | 当前生命周期或处理状态。 |

## `RuntimeBudgets` — `pkg/runtime/model/runtime_meta.go:109`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `SameToolArgsWarn` | `int` | ``json:"same_tool_args_warn,omitempty"`` | 承载 SameToolArgsWarn 的 int 值；业务上下文见本 struct 的源码包。 |
| `MaxToolCallsPerTool` | `int` | ``json:"max_tool_calls_per_tool,omitempty"`` | 承载 MaxToolCallsPerTool 的 int 值；业务上下文见本 struct 的源码包。 |
| `MaxTotalToolCalls` | `int` | ``json:"max_total_tool_calls,omitempty"`` | 承载 MaxTotalToolCalls 的 int 值；业务上下文见本 struct 的源码包。 |
| `TokenWarn` | `int` | ``json:"token_warn,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `PhaseDurationWarnSec` | `int` | ``json:"phase_duration_warn_sec,omitempty"`` | 承载 PhaseDurationWarnSec 的 int 值；业务上下文见本 struct 的源码包。 |

## `TaskInputAnchor` — `pkg/runtime/model/runtime_meta.go:117`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Summary` | `string` | ``json:"summary,omitempty"`` | 承载 Summary 的 string 值；业务上下文见本 struct 的源码包。 |
| `OriginalLength` | `int` | ``json:"original_length,omitempty"`` | 承载 OriginalLength 的 int 值；业务上下文见本 struct 的源码包。 |
| `Template` | `string` | ``json:"template,omitempty"`` | 承载 Template 的 string 值；业务上下文见本 struct 的源码包。 |
| `Theme` | `string` | ``json:"theme,omitempty"`` | 承载 Theme 的 string 值；业务上下文见本 struct 的源码包。 |
| `Recommendation` | `string` | ``json:"recommendation,omitempty"`` | 承载 Recommendation 的 string 值；业务上下文见本 struct 的源码包。 |

## `PlanSlide` — `pkg/runtime/model/runtime_meta.go:125`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PageIndex` | `int` | ``json:"page_index,omitempty"`` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | ``json:"task_id,omitempty"`` | 关联任务的唯一标识。 |
| `Title` | `string` | ``json:"title,omitempty"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentType` | `string` | ``json:"content_type,omitempty"`` | 承载 ContentType 的 string 值；业务上下文见本 struct 的源码包。 |
| `OutputFile` | `string` | ``json:"output_file,omitempty"`` | 承载 OutputFile 的 string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `string` | ``json:"status,omitempty"`` | 当前生命周期或处理状态。 |

## `AlignmentWarning` — `pkg/runtime/model/runtime_meta.go:134`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Code` | `string` | ``json:"code"`` | 承载 Code 的 string 值；业务上下文见本 struct 的源码包。 |
| `Step` | `string` | ``json:"step"`` | 承载 Step 的 string 值；业务上下文见本 struct 的源码包。 |
| `Severity` | `string` | ``json:"severity"`` | 承载 Severity 的 string 值；业务上下文见本 struct 的源码包。 |
| `Message` | `string` | ``json:"message"`` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |
| `PageIndex` | `int` | ``json:"page_index,omitempty"`` | 承载 PageIndex 的 int 值；业务上下文见本 struct 的源码包。 |
| `Expected` | `string` | ``json:"expected,omitempty"`` | 承载 Expected 的 string 值；业务上下文见本 struct 的源码包。 |
| `Observed` | `string` | ``json:"observed,omitempty"`` | 承载 Observed 的 string 值；业务上下文见本 struct 的源码包。 |

## `RuntimeMetaSnapshot` — `pkg/runtime/model/runtime_meta.go:145`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `TaskID` | `string` | ``json:"task_id,omitempty"`` | 关联任务的唯一标识。 |
| `WorkDir` | `string` | ``json:"work_dir,omitempty"`` | 承载 WorkDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `ElapsedMS` | `int64` | ``json:"elapsed_ms"`` | 承载 ElapsedMS 的 int64 值；业务上下文见本 struct 的源码包。 |
| `Phase` | `string` | ``json:"phase,omitempty"`` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhaseDetail` | `string` | ``json:"phase_detail,omitempty"`` | 承载 PhaseDetail 的 string 值；业务上下文见本 struct 的源码包。 |
| `LastError` | `string` | ``json:"last_error,omitempty"`` | 失败原因或错误详情。 |
| `LastTool` | `string` | ``json:"last_tool,omitempty"`` | 承载 LastTool 的 string 值；业务上下文见本 struct 的源码包。 |
| `ToolCalls` | `map[string]int` | ``json:"tool_calls,omitempty"`` | 承载 ToolCalls 的 map[string]int 值；业务上下文见本 struct 的源码包。 |
| `ToolErrors` | `map[string]int` | ``json:"tool_errors,omitempty"`` | 承载 ToolErrors 的 map[string]int 值；业务上下文见本 struct 的源码包。 |
| `SameToolArgsStreak` | `int` | ``json:"same_tool_args_streak,omitempty"`` | 承载 SameToolArgsStreak 的 int 值；业务上下文见本 struct 的源码包。 |
| `PromptTokens` | `int64` | ``json:"prompt_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `CompletionTokens` | `int64` | ``json:"completion_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `TotalTokens` | `int64` | ``json:"total_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `CompressionBeforeTokens` | `int` | ``json:"compression_before_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `CompressionAfterTokens` | `int` | ``json:"compression_after_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `CompressionSavedPct` | `string` | ``json:"compression_saved_pct,omitempty"`` | 承载 CompressionSavedPct 的 string 值；业务上下文见本 struct 的源码包。 |
| `CompressionBeforeMessages` | `int` | ``json:"compression_before_messages,omitempty"`` | 承载 CompressionBeforeMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `CompressionAfterMessages` | `int` | ``json:"compression_after_messages,omitempty"`` | 承载 CompressionAfterMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `CompressionRemovedMessages` | `int` | ``json:"compression_removed_messages,omitempty"`` | 承载 CompressionRemovedMessages 的 int 值；业务上下文见本 struct 的源码包。 |
| `CompressionSavedTokens` | `int` | ``json:"compression_saved_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `Budgets` | `RuntimeBudgets` | ``json:"budgets,omitempty"`` | 承载 Budgets 的 RuntimeBudgets 值；业务上下文见本 struct 的源码包。 |
| `BudgetWarnings` | `[]string` | ``json:"budget_warnings,omitempty"`` | 承载 BudgetWarnings 的 []string 值；业务上下文见本 struct 的源码包。 |
| `DoneSlides` | `int` | ``json:"done_slides,omitempty"`` | 承载 DoneSlides 的 int 值；业务上下文见本 struct 的源码包。 |
| `TotalSlides` | `int` | ``json:"total_slides,omitempty"`` | 承载 TotalSlides 的 int 值；业务上下文见本 struct 的源码包。 |
| `MissingFiles` | `int` | ``json:"missing_files,omitempty"`` | 承载 MissingFiles 的 int 值；业务上下文见本 struct 的源码包。 |
| `QAHighIssues` | `int` | ``json:"qa_high_issues,omitempty"`` | 承载 QAHighIssues 的 int 值；业务上下文见本 struct 的源码包。 |
| `QAMediumIssues` | `int` | ``json:"qa_medium_issues,omitempty"`` | 承载 QAMediumIssues 的 int 值；业务上下文见本 struct 的源码包。 |
| `QALowIssues` | `int` | ``json:"qa_low_issues,omitempty"`` | 承载 QALowIssues 的 int 值；业务上下文见本 struct 的源码包。 |
| `TaskInput` | `TaskInputAnchor` | ``json:"task_input,omitempty"`` | 承载 TaskInput 的 TaskInputAnchor 值；业务上下文见本 struct 的源码包。 |
| `PlanSlides` | `[]PlanSlide` | ``json:"plan_slides,omitempty"`` | 承载 PlanSlides 的 []PlanSlide 值；业务上下文见本 struct 的源码包。 |
| `CurrentSlide` | `*PlanSlide` | ``json:"current_slide,omitempty"`` | 承载 CurrentSlide 的 *PlanSlide 值；业务上下文见本 struct 的源码包。 |
| `AlignmentStatus` | `string` | ``json:"alignment_status,omitempty"`` | 当前生命周期或处理状态。 |
| `AlignmentWarnings` | `[]AlignmentWarning` | ``json:"alignment_warnings,omitempty"`` | 承载 AlignmentWarnings 的 []AlignmentWarning 值；业务上下文见本 struct 的源码包。 |
| `EventCounts` | `map[string]int` | ``json:"event_counts,omitempty"`` | 承载 EventCounts 的 map[string]int 值；业务上下文见本 struct 的源码包。 |
| `RecentEvents` | `[]RuntimeEvent` | ``json:"recent_events,omitempty"`` | 承载 RecentEvents 的 []RuntimeEvent 值；业务上下文见本 struct 的源码包。 |

## `RuntimeEvent` — `pkg/runtime/model/runtime_meta.go:191`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `int64` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `TaskID` | `string` | ``json:"task_id,omitempty"`` | 关联任务的唯一标识。 |
| `Timestamp` | `string` | ``json:"timestamp"`` | 事件或消息发生时间。 |
| `ElapsedMS` | `int64` | ``json:"elapsed_ms"`` | 承载 ElapsedMS 的 int64 值；业务上下文见本 struct 的源码包。 |
| `Kind` | `string` | ``json:"kind"`` | 承载 Kind 的 string 值；业务上下文见本 struct 的源码包。 |
| `Phase` | `string` | ``json:"phase,omitempty"`` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `Name` | `string` | ``json:"name,omitempty"`` | 承载 Name 的 string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `string` | ``json:"status,omitempty"`` | 当前生命周期或处理状态。 |
| `Detail` | `string` | ``json:"detail,omitempty"`` | 承载 Detail 的 string 值；业务上下文见本 struct 的源码包。 |
| `Metadata` | `map[string]any` | ``json:"metadata,omitempty"`` | 结构化补充元数据；输出前需按权限脱敏。 |

## `RuntimeReport` — `pkg/runtime/model/runtime_meta.go:204`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `TaskID` | `string` | ``json:"task_id,omitempty"`` | 关联任务的唯一标识。 |
| `WorkDir` | `string` | ``json:"work_dir,omitempty"`` | 承载 WorkDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `string` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `WrittenAt` | `string` | ``json:"written_at"`` | 承载 WrittenAt 的 string 值；业务上下文见本 struct 的源码包。 |
| `Snapshot` | `RuntimeMetaSnapshot` | ``json:"snapshot"`` | 承载 Snapshot 的 RuntimeMetaSnapshot 值；业务上下文见本 struct 的源码包。 |
| `EventCounts` | `map[string]int` | ``json:"event_counts,omitempty"`` | 承载 EventCounts 的 map[string]int 值；业务上下文见本 struct 的源码包。 |

## `runtimeMetaKey` — `pkg/runtime/model/runtime_meta.go:213`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `RuntimeStatusChatModel` — `pkg/runtime/model/runtime_model.go:29`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `inner` | `model.ToolCallingChatModel` | `` | 承载 inner 的 model.ToolCallingChatModel 值；业务上下文见本 struct 的源码包。 |
| `meta` | `*RuntimeMeta` | `` | 承载 meta 的 *RuntimeMeta 值；业务上下文见本 struct 的源码包。 |

## `TokenTracker` — `pkg/runtime/model/token_tracker.go:10`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PromptTokens` | `atomic.Int64` | `` | 模型调用的 token 统计或限制参数。 |
| `CompletionTokens` | `atomic.Int64` | `` | 模型调用的 token 统计或限制参数。 |
| `TotalTokens` | `atomic.Int64` | `` | 模型调用的 token 统计或限制参数。 |

## `tokenTrackerKey` — `pkg/runtime/model/token_tracker.go:28`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `DeliverySnapshot` — `pkg/runtime/task/delivery_metadata.go:14`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Total` | `int` | `` | 承载 Total 的 int 值；业务上下文见本 struct 的源码包。 |
| `Done` | `int` | `` | 承载 Done 的 int 值；业务上下文见本 struct 的源码包。 |
| `Files` | `[]string` | `` | 承载 Files 的 []string 值；业务上下文见本 struct 的源码包。 |
| `PendingTasks` | `[]string` | `` | 承载 PendingTasks 的 []string 值；业务上下文见本 struct 的源码包。 |

## `SSERichEvent` — `pkg/runtime/task/manager.go:42`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `uint64` | ``json:"id,omitempty"`` | 当前记录或对象的唯一标识。 |
| `SegmentID` | `string` | ``json:"segment_id,omitempty"`` | 承载 SegmentID 的 string 值；业务上下文见本 struct 的源码包。 |
| `SegmentBoundary` | `bool` | ``json:"segment_boundary,omitempty"`` | 承载 SegmentBoundary 的 bool 值；业务上下文见本 struct 的源码包。 |
| `ToolPreview` | `map[string]any` | ``json:"tool_preview,omitempty"`` | 承载 ToolPreview 的 map[string]any 值；业务上下文见本 struct 的源码包。 |
| `Type` | `string` | ``json:"type"`` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `Content` | `string` | ``json:"content,omitempty"`` | 正文或序列化内容。 |
| `ToolName` | `string` | ``json:"tool_name,omitempty"`` | 承载 ToolName 的 string 值；业务上下文见本 struct 的源码包。 |
| `ToolArgs` | `string` | ``json:"tool_args,omitempty"`` | 承载 ToolArgs 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |
| `Tasks` | `[]*deck.TaskItem` | ``json:"tasks,omitempty"`` | 承载 Tasks 的 []*deck.TaskItem 值；业务上下文见本 struct 的源码包。 |
| `Done` | `int` | ``json:"done,omitempty"`` | 承载 Done 的 int 值；业务上下文见本 struct 的源码包。 |
| `Total` | `int` | ``json:"total,omitempty"`` | 承载 Total 的 int 值；业务上下文见本 struct 的源码包。 |
| `Files` | `[]string` | ``json:"files,omitempty"`` | 承载 Files 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Message` | `string` | ``json:"message,omitempty"`` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |
| `Duration` | `string` | ``json:"duration,omitempty"`` | 承载 Duration 的 string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `TaskStatus` | ``json:"status,omitempty"`` | 当前生命周期或处理状态。 |
| `PromptTokens` | `int64` | ``json:"prompt_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `CompletionTokens` | `int64` | ``json:"completion_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `TotalTokens` | `int64` | ``json:"total_tokens,omitempty"`` | 模型调用的 token 统计或限制参数。 |
| `Phase` | `string` | ``json:"phase,omitempty"`` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhaseDetail` | `string` | ``json:"phase_detail,omitempty"`` | 承载 PhaseDetail 的 string 值；业务上下文见本 struct 的源码包。 |
| `RuntimeEvent` | `*utils.RuntimeEvent` | ``json:"runtime_event,omitempty"`` | 承载 RuntimeEvent 的 *utils.RuntimeEvent 值；业务上下文见本 struct 的源码包。 |

## `TaskInfo` — `pkg/runtime/task/manager.go:70`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `UserID` | `int` | ``json:"user_id"`` | 关联用户的唯一标识。 |
| `Query` | `string` | ``json:"query"`` | 用户或系统发起的查询文本。 |
| `Status` | `TaskStatus` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `WorkDir` | `string` | ``json:"work_dir"`` | 承载 WorkDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `DoneCount` | `int` | ``json:"done_count"`` | 承载 DoneCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `TotalCount` | `int` | ``json:"total_count"`` | 承载 TotalCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `Duration` | `string` | ``json:"duration,omitempty"`` | 承载 Duration 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |
| `Files` | `[]string` | ``json:"files,omitempty"`` | 承载 Files 的 []string 值；业务上下文见本 struct 的源码包。 |
| `PromptTokens` | `int64` | ``json:"prompt_tokens"`` | 模型调用的 token 统计或限制参数。 |
| `CompletionTokens` | `int64` | ``json:"completion_tokens"`` | 模型调用的 token 统计或限制参数。 |
| `TotalTokens` | `int64` | ``json:"total_tokens"`` | 模型调用的 token 统计或限制参数。 |
| `ConversationContent` | `string` | ``json:"conversation_content,omitempty"`` | 正文或序列化内容。 |
| `FullAnswer` | `string` | ``json:"full_answer,omitempty"`` | 承载 FullAnswer 的 string 值；业务上下文见本 struct 的源码包。 |
| `Intent` | `string` | ``json:"intent,omitempty"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `ConversationID` | `string` | ``json:"conversation_id,omitempty"`` | 关联逻辑会话的唯一标识。 |
| `SourceMessageID` | `string` | ``json:"source_message_id,omitempty"`` | 关联消息的唯一标识。 |
| `ParentTaskID` | `string` | ``json:"parent_task_id,omitempty"`` | 关联任务的唯一标识。 |
| `GenerationStartedAt` | `*time.Time` | ``json:"generation_started_at,omitempty"`` | 承载 GenerationStartedAt 的 *time.Time 值；业务上下文见本 struct 的源码包。 |
| `GenerationFinishedAt` | `*time.Time` | ``json:"generation_finished_at,omitempty"`` | 承载 GenerationFinishedAt 的 *time.Time 值；业务上下文见本 struct 的源码包。 |
| `GenerationDurationMS` | `int64` | ``json:"generation_duration_ms,omitempty"`` | 承载 GenerationDurationMS 的 int64 值；业务上下文见本 struct 的源码包。 |
| `FixerRunCount` | `int` | ``json:"fixer_run_count"`` | 承载 FixerRunCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `Feedback` | `*DeliveryFeedback` | ``json:"feedback,omitempty"`` | 承载 Feedback 的 *DeliveryFeedback 值；业务上下文见本 struct 的源码包。 |

## `DeliveryFeedback` — `pkg/runtime/task/manager.go:99`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Rating` | `int` | ``json:"rating"`` | 承载 Rating 的 int 值；业务上下文见本 struct 的源码包。 |
| `Suggestion` | `string` | ``json:"suggestion,omitempty"`` | 承载 Suggestion 的 string 值；业务上下文见本 struct 的源码包。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `TaskState` — `pkg/runtime/task/manager.go:106`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Info` | `TaskInfo` | `` | 承载 Info 的 TaskInfo 值；业务上下文见本 struct 的源码包。 |
| `Events` | `[]SSERichEvent` | `` | 承载 Events 的 []SSERichEvent 值；业务上下文见本 struct 的源码包。 |
| `nextEventID` | `uint64` | `` | 承载 nextEventID 的 uint64 值；业务上下文见本 struct 的源码包。 |
| `turnEventID` | `uint64` | `` | 承载 turnEventID 的 uint64 值；业务上下文见本 struct 的源码包。 |
| `listeners` | `map[string]chan SSERichEvent` | `` | 承载 listeners 的 map[string]chan SSERichEvent 值；业务上下文见本 struct 的源码包。 |
| `cancel` | `context.CancelFunc` | `` | 承载 cancel 的 context.CancelFunc 值；业务上下文见本 struct 的源码包。 |
| `result` | `*deck.PPTTaskResult` | `` | 承载 result 的 *deck.PPTTaskResult 值；业务上下文见本 struct 的源码包。 |
| `reportedFiles` | `map[string]bool` | `` | 承载 reportedFiles 的 map[string]bool 值；业务上下文见本 struct 的源码包。 |
| `runtimeMeta` | `*utils.RuntimeMeta` | `` | 承载 runtimeMeta 的 *utils.RuntimeMeta 值；业务上下文见本 struct 的源码包。 |
| `delivery` | `DeliverySnapshot` | `` | 承载 delivery 的 DeliverySnapshot 值；业务上下文见本 struct 的源码包。 |
| `Mu` | `sync.Mutex` | `` | 承载 Mu 的 sync.Mutex 值；业务上下文见本 struct 的源码包。 |
| `pendingContinueMsg` | `string` | `` | 承载 pendingContinueMsg 的 string 值；业务上下文见本 struct 的源码包。 |
| `pendingContinueQueued` | `bool` | `` | 承载 pendingContinueQueued 的 bool 值；业务上下文见本 struct 的源码包。 |
| `conversationStreamActive` | `bool` | `` | 承载 conversationStreamActive 的 bool 值；业务上下文见本 struct 的源码包。 |
| `fullAnswer` | `strings.Builder` | `` | 承载 fullAnswer 的 strings.Builder 值；业务上下文见本 struct 的源码包。 |
| `answerTurn` | `strings.Builder` | `` | 承载 answerTurn 的 strings.Builder 值；业务上下文见本 struct 的源码包。 |
| `assistantTurnFn` | `func(taskID, workDir, content string)` | `` | 承载 assistantTurnFn 的 func(taskID, workDir, content string) 值；业务上下文见本 struct 的源码包。 |

## `TaskManager` — `pkg/runtime/task/manager.go:532`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.RWMutex` | `` | 承载 mu 的 sync.RWMutex 值；业务上下文见本 struct 的源码包。 |
| `tasks` | `map[string]*TaskState` | `` | 承载 tasks 的 map[string]*TaskState 值；业务上下文见本 struct 的源码包。 |
| `baseDir` | `string` | `` | 承载 baseDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `onTaskComplete` | `func(userID int, workDir string, query string)` | `` | 承载 onTaskComplete 的 func(userID int, workDir string, query string) 值；业务上下文见本 struct 的源码包。 |
| `onTaskFailed` | `func(taskID string)` | `` | 承载 onTaskFailed 的 func(taskID string) 值；业务上下文见本 struct 的源码包。 |
| `onTaskContinue` | `func(taskID string)` | `` | 承载 onTaskContinue 的 func(taskID string) 值；业务上下文见本 struct 的源码包。 |
| `onFileReady` | `func(taskID string, workDir string, filename string)` | `` | 承载 onFileReady 的 func(taskID string, workDir string, filename string) 值；业务上下文见本 struct 的源码包。 |
| `onAssistantTurn` | `func(taskID string, workDir string, content string)` | `` | 承载 onAssistantTurn 的 func(taskID string, workDir string, content string) 值；业务上下文见本 struct 的源码包。 |

## `adminTaskResponse` — `pkg/runtime/web/admin_handler.go:57`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `UserID` | `uint` | ``json:"user_id"`` | 关联用户的唯一标识。 |
| `UserEmail` | `string` | ``json:"user_email"`` | 承载 UserEmail 的 string 值；业务上下文见本 struct 的源码包。 |
| `Query` | `string` | ``json:"query"`` | 用户或系统发起的查询文本。 |
| `Status` | `string` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `DoneCount` | `int` | ``json:"done_count"`` | 承载 DoneCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `TotalCount` | `int` | ``json:"total_count"`` | 承载 TotalCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `Duration` | `string` | ``json:"duration"`` | 承载 Duration 的 string 值；业务上下文见本 struct 的源码包。 |
| `GenerationStartedAt` | `*time.Time` | ``json:"generation_started_at,omitempty"`` | 承载 GenerationStartedAt 的 *time.Time 值；业务上下文见本 struct 的源码包。 |
| `GenerationFinishedAt` | `*time.Time` | ``json:"generation_finished_at,omitempty"`` | 承载 GenerationFinishedAt 的 *time.Time 值；业务上下文见本 struct 的源码包。 |
| `GenerationDurationMS` | `int64` | ``json:"generation_duration_ms"`` | 承载 GenerationDurationMS 的 int64 值；业务上下文见本 struct 的源码包。 |
| `FixerRunCount` | `int` | ``json:"fixer_run_count"`` | 承载 FixerRunCount 的 int 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error"`` | 失败原因或错误详情。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `ChatBenchmarkSearchResult` — `pkg/runtime/web/benchmark_chat.go:18`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `URL` | `string` | ``json:"url"`` | 承载 URL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Description` | `string` | ``json:"description"`` | 承载 Description 的 string 值；业务上下文见本 struct 的源码包。 |
| `Source` | `string` | ``json:"source,omitempty"`` | 承载 Source 的 string 值；业务上下文见本 struct 的源码包。 |

## `ChatBenchmarkImageResult` — `pkg/runtime/web/benchmark_chat.go:25`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PreviewURL` | `string` | ``json:"preview_url"`` | 承载 PreviewURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `ImageURL` | `string` | ``json:"image_url"`` | 承载 ImageURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `SourceURL` | `string` | ``json:"source_url"`` | 承载 SourceURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Photographer` | `string` | ``json:"photographer"`` | 承载 Photographer 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhotographerURL` | `string` | ``json:"photographer_url"`` | 承载 PhotographerURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Attribution` | `string` | ``json:"attribution"`` | 承载 Attribution 的 string 值；业务上下文见本 struct 的源码包。 |

## `ChatBenchmarkInput` — `pkg/runtime/web/benchmark_chat.go:37`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Message` | `string` | ``json:"message"`` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |
| `Fallback` | `string` | ``json:"fallback,omitempty"`` | 承载 Fallback 的 string 值；业务上下文见本 struct 的源码包。 |
| `ConversationContext` | `string` | ``json:"conversation_context,omitempty"`` | 承载 ConversationContext 的 string 值；业务上下文见本 struct 的源码包。 |
| `WebResults` | `[]ChatBenchmarkSearchResult` | ``json:"web_results,omitempty"`` | 承载 WebResults 的 []ChatBenchmarkSearchResult 值；业务上下文见本 struct 的源码包。 |
| `Images` | `[]ChatBenchmarkImageResult` | ``json:"images,omitempty"`` | 承载 Images 的 []ChatBenchmarkImageResult 值；业务上下文见本 struct 的源码包。 |
| `WebSearchError` | `string` | ``json:"web_search_error,omitempty"`` | 失败原因或错误详情。 |
| `ImageSearchError` | `string` | ``json:"image_search_error,omitempty"`` | 失败原因或错误详情。 |

## `BenchmarkCreateRouteResult` — `pkg/runtime/web/benchmark_router.go:17`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Intent` | `string` | ``json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `TargetAgent` | `string` | ``json:"target_agent,omitempty"`` | 承载 TargetAgent 的 string 值；业务上下文见本 struct 的源码包。 |
| `NormalizedRequest` | `string` | ``json:"normalized_request,omitempty"`` | 承载 NormalizedRequest 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `ClarificationQuestion` | `string` | ``json:"clarification_question,omitempty"`` | 承载 ClarificationQuestion 的 string 值；业务上下文见本 struct 的源码包。 |
| `Confidence` | `float64` | ``json:"confidence,omitempty"`` | 承载 Confidence 的 float64 值；业务上下文见本 struct 的源码包。 |

## `BenchmarkMessageRouteResult` — `pkg/runtime/web/benchmark_router.go:29`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Intent` | `string` | ``json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `TargetAgent` | `string` | ``json:"target_agent,omitempty"`` | 承载 TargetAgent 的 string 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | ``json:"task_id"`` | 关联任务的唯一标识。 |
| `NormalizedRequest` | `string` | ``json:"normalized_request,omitempty"`` | 承载 NormalizedRequest 的 string 值；业务上下文见本 struct 的源码包。 |
| `Action` | `string` | ``json:"action,omitempty"`` | 承载 Action 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `Confidence` | `float64` | ``json:"confidence,omitempty"`` | 承载 Confidence 的 float64 值；业务上下文见本 struct 的源码包。 |

## `benchmarkModelAdapter` — `pkg/runtime/web/benchmark_router.go:127`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `inner` | `einomodel.ToolCallingChatModel` | `` | 承载 inner 的 einomodel.ToolCallingChatModel 值；业务上下文见本 struct 的源码包。 |

## `RouteResult` — `pkg/runtime/web/continuation_handler.go:275`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Intent` | `string` | ``json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `TargetPages` | `[]int` | ``json:"target_pages,omitempty"`` | 承载 TargetPages 的 []int 值；业务上下文见本 struct 的源码包。 |
| `TargetTaskIDs` | `[]string` | ``json:"target_task_ids,omitempty"`` | 承载 TargetTaskIDs 的 []string 值；业务上下文见本 struct 的源码包。 |
| `NeedsClarification` | `bool` | ``json:"needs_clarification,omitempty"`` | 承载 NeedsClarification 的 bool 值；业务上下文见本 struct 的源码包。 |
| `ClarificationQuestion` | `string` | ``json:"clarification_question,omitempty"`` | 承载 ClarificationQuestion 的 string 值；业务上下文见本 struct 的源码包。 |
| `FixDetails` | `*FixDetails` | ``json:"fix_details,omitempty"`` | 承载 FixDetails 的 *FixDetails 值；业务上下文见本 struct 的源码包。 |
| `RegenerateScope` | `[]int` | ``json:"regenerate_scope,omitempty"`` | 承载 RegenerateScope 的 []int 值；业务上下文见本 struct 的源码包。 |
| `SuggestFix` | `bool` | ``json:"suggest_fix,omitempty"`` | 承载 SuggestFix 的 bool 值；业务上下文见本 struct 的源码包。 |

## `FixDetails` — `pkg/runtime/web/continuation_handler.go:306`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Aspect` | `string` | ``json:"aspect"`` | 承载 Aspect 的 string 值；业务上下文见本 struct 的源码包。 |
| `Detail` | `string` | ``json:"detail"`` | 承载 Detail 的 string 值；业务上下文见本 struct 的源码包。 |
| `TargetElements` | `string` | ``json:"target_elements,omitempty"`` | 承载 TargetElements 的 string 值；业务上下文见本 struct 的源码包。 |

## `modelCredential` — `pkg/runtime/web/credential_handler.go:93`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Provider` | `string` | `` | 承载 Provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `APIKey` | `string` | `` | 承载 APIKey 的 string 值；业务上下文见本 struct 的源码包。 |

## `HealthStatus` — `pkg/runtime/web/health.go:33`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Status` | `string` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `Message` | `string` | ``json:"message,omitempty"`` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |

## `HealthReport` — `pkg/runtime/web/health.go:39`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Status` | `string` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `Version` | `string` | ``json:"version"`` | 承载 Version 的 string 值；业务上下文见本 struct 的源码包。 |
| `Uptime` | `string` | ``json:"uptime"`` | 承载 Uptime 的 string 值；业务上下文见本 struct 的源码包。 |
| `Components` | `map[string]HealthStatus` | ``json:"components"`` | 承载 Components 的 map[string]HealthStatus 值；业务上下文见本 struct 的源码包。 |

## `chatImageResult` — `pkg/runtime/web/message_chat.go:20`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PreviewURL` | `string` | ``json:"preview_url"`` | 承载 PreviewURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `ImageURL` | `string` | ``json:"image_url"`` | 承载 ImageURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `SourceURL` | `string` | ``json:"source_url"`` | 承载 SourceURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Photographer` | `string` | ``json:"photographer"`` | 承载 Photographer 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhotographerURL` | `string` | ``json:"photographer_url"`` | 承载 PhotographerURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Attribution` | `string` | ``json:"attribution"`` | 承载 Attribution 的 string 值；业务上下文见本 struct 的源码包。 |

## `chatImageSearchResponse` — `pkg/runtime/web/message_chat.go:29`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Photos` | `[]chatImageResult` | ``json:"photos"`` | 承载 Photos 的 []chatImageResult 值；业务上下文见本 struct 的源码包。 |

## `chatAugmentations` — `pkg/runtime/web/message_chat.go:33`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `promptParts` | `[]string` | `` | 承载 promptParts 的 []string 值；业务上下文见本 struct 的源码包。 |
| `webResults` | `[]search.SearchResult` | `` | 承载 webResults 的 []search.SearchResult 值；业务上下文见本 struct 的源码包。 |
| `images` | `[]chatImageResult` | `` | 承载 images 的 []chatImageResult 值；业务上下文见本 struct 的源码包。 |
| `query` | `string` | `` | 用户或系统发起的查询文本。 |

## `chatTraceEvent` — `pkg/runtime/web/message_chat.go:42`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Type` | `string` | `` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `Phase` | `string` | `` | 承载 Phase 的 string 值；业务上下文见本 struct 的源码包。 |
| `ToolName` | `string` | `` | 承载 ToolName 的 string 值；业务上下文见本 struct 的源码包。 |
| `Detail` | `string` | `` | 承载 Detail 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | `` | 失败原因或错误详情。 |
| `Preview` | `map[string]any` | `` | 承载 Preview 的 map[string]any 值；业务上下文见本 struct 的源码包。 |

## `planDraftResponse` — `pkg/runtime/web/plan_draft.go:14`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `UserID` | `uint` | ``json:"user_id"`` | 关联用户的唯一标识。 |
| `ConversationID` | `string` | ``json:"conversation_id"`` | 关联逻辑会话的唯一标识。 |
| `SourceMessageID` | `string` | ``json:"source_message_id"`` | 关联消息的唯一标识。 |
| `Query` | `string` | ``json:"query"`` | 用户或系统发起的查询文本。 |
| `NormalizedRequest` | `string` | ``json:"normalized_request"`` | 承载 NormalizedRequest 的 string 值；业务上下文见本 struct 的源码包。 |
| `DraftContent` | `string` | ``json:"draft_content"`` | 正文或序列化内容。 |
| `Status` | `string` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `createRequestRoute` — `pkg/runtime/web/request_router.go:40`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Intent` | `string` | ``json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `ClarificationQuestion` | `string` | ``json:"clarification_question,omitempty"`` | 承载 ClarificationQuestion 的 string 值；业务上下文见本 struct 的源码包。 |
| `Confidence` | `float64` | ``json:"confidence,omitempty"`` | 承载 Confidence 的 float64 值；业务上下文见本 struct 的源码包。 |

## `MessageRouteResult` — `pkg/runtime/web/request_router.go:47`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Intent` | `string` | ``json:"intent"`` | 承载 Intent 的 string 值；业务上下文见本 struct 的源码包。 |
| `Mode` | `string` | ``json:"mode"`` | 承载 Mode 的 string 值；业务上下文见本 struct 的源码包。 |
| `Confidence` | `float64` | ``json:"confidence"`` | 承载 Confidence 的 float64 值；业务上下文见本 struct 的源码包。 |
| `NeedsConfirmation` | `bool` | ``json:"needs_confirmation"`` | 承载 NeedsConfirmation 的 bool 值；业务上下文见本 struct 的源码包。 |
| `NormalizedRequest` | `string` | ``json:"normalized_request"`` | 承载 NormalizedRequest 的 string 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | ``json:"task_id"`` | 关联任务的唯一标识。 |
| `DraftID` | `string` | ``json:"draft_id,omitempty"`` | 承载 DraftID 的 string 值；业务上下文见本 struct 的源码包。 |
| `MissingFields` | `[]string` | ``json:"missing_fields"`` | 承载 MissingFields 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Action` | `string` | ``json:"action"`` | 承载 Action 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason,omitempty"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `Reply` | `string` | ``json:"reply,omitempty"`` | 承载 Reply 的 string 值；业务上下文见本 struct 的源码包。 |
| `TaskCandidates` | `[]TaskCandidate` | ``json:"task_candidates,omitempty"`` | 承载 TaskCandidates 的 []TaskCandidate 值；业务上下文见本 struct 的源码包。 |
| `Streaming` | `bool` | ``json:"streaming,omitempty"`` | 承载 Streaming 的 bool 值；业务上下文见本 struct 的源码包。 |
| `AfterEventID` | `uint64` | ``json:"after_event_id,omitempty"`` | 承载 AfterEventID 的 uint64 值；业务上下文见本 struct 的源码包。 |

## `TaskCandidate` — `pkg/runtime/web/request_router.go:64`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `Status` | `string` | ``json:"status"`` | 当前生命周期或处理状态。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |

## `Server` — `pkg/runtime/web/server.go:32`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `tasks` | `*task.TaskManager` | `` | 承载 tasks 的 *task.TaskManager 值；业务上下文见本 struct 的源码包。 |
| `sessionManager` | `*session.SessionManager` | `` | 承载 sessionManager 的 *session.SessionManager 值；业务上下文见本 struct 的源码包。 |
| `agentFactory` | `task.AgentFactory` | `` | 承载 agentFactory 的 task.AgentFactory 值；业务上下文见本 struct 的源码包。 |
| `makeTaskConfig` | `func(taskID string) *deck.PPTTaskConfig` | `` | 对应组件的运行配置。 |
| `taskIDGen` | `func() string` | `` | 承载 taskIDGen 的 func() string 值；业务上下文见本 struct 的源码包。 |
| `engine` | `*gin.Engine` | `` | 承载 engine 的 *gin.Engine 值；业务上下文见本 struct 的源码包。 |
| `addr` | `string` | `` | 承载 addr 的 string 值；业务上下文见本 struct 的源码包。 |
| `templateLoader` | `*templates.Loader` | `` | 承载 templateLoader 的 *templates.Loader 值；业务上下文见本 struct 的源码包。 |
| `skillDir` | `string` | `` | 承载 skillDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `operator` | `commandline.Operator` | `` | 承载 operator 的 commandline.Operator 值；业务上下文见本 struct 的源码包。 |
| `logAnalysis` | `*loganalysis.Service` | `` | 承载 logAnalysis 的 *loganalysis.Service 值；业务上下文见本 struct 的源码包。 |
| `chatTrace` | `chattrace.Store` | `` | 承载 chatTrace 的 chattrace.Store 值；业务上下文见本 struct 的源码包。 |
| `continueStarter` | `func(taskID string, ts *task.TaskState, message string, uid int, sess *session.ConversationSession)` | `` | 承载 continueStarter 的 func(taskID string, ts *task.TaskState, message string, uid int, sess *session.ConversationSession) 值；业务上下文见本 struct 的源码包。 |
| `aiModelFactory` | `func(ctx context.Context) (interface {` | `` | 承载 aiModelFactory 的 func(ctx context.Context) (interface { 值；业务上下文见本 struct 的源码包。 |
| `Generate(ctx` | `context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)` | `` | 承载 Generate(ctx 的 context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error) 值；业务上下文见本 struct 的源码包。 |
| `},` | `error)` | `` | 承载 }, 的 error) 值；业务上下文见本 struct 的源码包。 |
| `textModelFactory` | `func(ctx context.Context) (interface {` | `` | 承载 textModelFactory 的 func(ctx context.Context) (interface { 值；业务上下文见本 struct 的源码包。 |
| `Generate(ctx` | `context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)` | `` | 承载 Generate(ctx 的 context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error) 值；业务上下文见本 struct 的源码包。 |
| `},` | `error)` | `` | 承载 }, 的 error) 值；业务上下文见本 struct 的源码包。 |

## `ServerConfig` — `pkg/runtime/web/server.go:57`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Addr` | `string` | `` | 承载 Addr 的 string 值；业务上下文见本 struct 的源码包。 |
| `BaseDir` | `string` | `` | 承载 BaseDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `FrontendDir` | `string` | `` | 承载 FrontendDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `SkillsDir` | `string` | `` | 承载 SkillsDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `Operator` | `commandline.Operator` | `` | 承载 Operator 的 commandline.Operator 值；业务上下文见本 struct 的源码包。 |
| `AgentFactory` | `task.AgentFactory` | `` | 承载 AgentFactory 的 task.AgentFactory 值；业务上下文见本 struct 的源码包。 |
| `MakeTaskConfig` | `func(taskID string) *deck.PPTTaskConfig` | `` | 对应组件的运行配置。 |
| `AIModelFactory` | `func(ctx context.Context) (interface {` | `` | 承载 AIModelFactory 的 func(ctx context.Context) (interface { 值；业务上下文见本 struct 的源码包。 |
| `Generate(ctx` | `context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)` | `` | 承载 Generate(ctx 的 context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error) 值；业务上下文见本 struct 的源码包。 |
| `},` | `error)` | `` | 承载 }, 的 error) 值；业务上下文见本 struct 的源码包。 |
| `TextModelFactory` | `func(ctx context.Context) (interface {` | `` | 承载 TextModelFactory 的 func(ctx context.Context) (interface { 值；业务上下文见本 struct 的源码包。 |
| `Generate(ctx` | `context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)` | `` | 承载 Generate(ctx 的 context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error) 值；业务上下文见本 struct 的源码包。 |
| `},` | `error)` | `` | 承载 }, 的 error) 值；业务上下文见本 struct 的源码包。 |
| `LogAnalysisModelFactory` | `func(ctx context.Context) (model.ToolCallingChatModel, error)` | `` | 承载 LogAnalysisModelFactory 的 func(ctx context.Context) (model.ToolCallingChatModel, error) 值；业务上下文见本 struct 的源码包。 |
| `LogAnalysisIdleInterval` | `time.Duration` | `` | 承载 LogAnalysisIdleInterval 的 time.Duration 值；业务上下文见本 struct 的源码包。 |
| `ChatTraceStore` | `chattrace.Store` | `` | 承载 ChatTraceStore 的 chattrace.Store 值；业务上下文见本 struct 的源码包。 |

## `Message` — `pkg/session/session.go:27`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Role` | `string` | ``json:"role"`` | 承载 Role 的 string 值；业务上下文见本 struct 的源码包。 |
| `Content` | `string` | ``json:"content"`` | 正文或序列化内容。 |
| `Timestamp` | `time.Time` | ``json:"timestamp"`` | 事件或消息发生时间。 |

## `ConversationSession` — `pkg/session/session.go:34`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.RWMutex` | `` | 承载 mu 的 sync.RWMutex 值；业务上下文见本 struct 的源码包。 |
| `TaskID` | `string` | ``json:"task_id"`` | 关联任务的唯一标识。 |
| `WorkDir` | `string` | ``json:"work_dir"`` | 承载 WorkDir 的 string 值；业务上下文见本 struct 的源码包。 |
| `Messages` | `[]Message` | ``json:"messages"`` | 承载 Messages 的 []Message 值；业务上下文见本 struct 的源码包。 |
| `CreatedAt` | `time.Time` | ``json:"created_at"`` | 创建时间。 |
| `UpdatedAt` | `time.Time` | ``json:"updated_at"`` | 最近更新时间。 |

## `SessionManager` — `pkg/session/session.go:159`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.RWMutex` | `` | 承载 mu 的 sync.RWMutex 值；业务上下文见本 struct 的源码包。 |
| `sessions` | `map[string]*ConversationSession` | `` | 承载 sessions 的 map[string]*ConversationSession 值；业务上下文见本 struct 的源码包。 |

## `LayoutInfo` — `pkg/templates/loader.go:19`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Name` | `string` | ``json:"name"`` | 承载 Name 的 string 值；业务上下文见本 struct 的源码包。 |
| `DisplayName` | `string` | ``json:"display_name"`` | 承载 DisplayName 的 string 值；业务上下文见本 struct 的源码包。 |
| `Type` | `TemplateType` | ``json:"type"`` | 承载 Type 的 TemplateType 值；业务上下文见本 struct 的源码包。 |
| `Description` | `string` | ``json:"description"`` | 承载 Description 的 string 值；业务上下文见本 struct 的源码包。 |
| `Fields` | `[]Field` | ``json:"fields"`` | 承载 Fields 的 []Field 值；业务上下文见本 struct 的源码包。 |
| `Contract` | `*LayoutContract` | ``json:"contract,omitempty"`` | 承载 Contract 的 *LayoutContract 值；业务上下文见本 struct 的源码包。 |

## `LayoutContract` — `pkg/templates/loader.go:29`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Capacity` | `map[string]any` | ``json:"capacity,omitempty"`` | 承载 Capacity 的 map[string]any 值；业务上下文见本 struct 的源码包。 |
| `RequiredFields` | `[]string` | ``json:"required_fields,omitempty"`` | 承载 RequiredFields 的 []string 值；业务上下文见本 struct 的源码包。 |
| `BestFor` | `[]string` | ``json:"best_for,omitempty"`` | 承载 BestFor 的 []string 值；业务上下文见本 struct 的源码包。 |
| `AvoidFor` | `[]string` | ``json:"avoid_for,omitempty"`` | 承载 AvoidFor 的 []string 值；业务上下文见本 struct 的源码包。 |
| `OverflowStrategy` | `string` | ``json:"overflow_strategy,omitempty"`` | 承载 OverflowStrategy 的 string 值；业务上下文见本 struct 的源码包。 |
| `BackgroundPolicy` | `string` | ``json:"background_policy,omitempty"`` | 承载 BackgroundPolicy 的 string 值；业务上下文见本 struct 的源码包。 |
| `VisualPrimitives` | `[]string` | ``json:"visual_primitives,omitempty"`` | 承载 VisualPrimitives 的 []string 值；业务上下文见本 struct 的源码包。 |

## `Field` — `pkg/templates/loader.go:40`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Name` | `string` | ``json:"name"`` | 承载 Name 的 string 值；业务上下文见本 struct 的源码包。 |
| `Label` | `string` | ``json:"label"`` | 承载 Label 的 string 值；业务上下文见本 struct 的源码包。 |
| `Type` | `string` | ``json:"type"`` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `Required` | `bool` | ``json:"required"`` | 承载 Required 的 bool 值；业务上下文见本 struct 的源码包。 |

## `componentContractsFile` — `pkg/templates/loader.go:47`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ContentTypes` | `map[string]struct {` | `` | 承载 ContentTypes 的 map[string]struct { 值；业务上下文见本 struct 的源码包。 |
| `BestFor` | `[]string` | ``json:"best_for"`` | 承载 BestFor 的 []string 值；业务上下文见本 struct 的源码包。 |
| `RecommendedComponents` | `[]string` | ``json:"recommended_components"`` | 承载 RecommendedComponents 的 []string 值；业务上下文见本 struct 的源码包。 |
| `Capacity` | `map[string]any` | ``json:"capacity"`` | 承载 Capacity 的 map[string]any 值；业务上下文见本 struct 的源码包。 |
| `Variants` | `[]string` | ``json:"variants"`` | 承载 Variants 的 []string 值；业务上下文见本 struct 的源码包。 |
| `DeckRule` | `string` | ``json:"deck_rule"`` | 承载 DeckRule 的 string 值；业务上下文见本 struct 的源码包。 |
| `}` | `嵌入字段` | ``json:"content_types"`` | 承载 } 的 嵌入字段 值；业务上下文见本 struct 的源码包。 |

## `Loader` — `pkg/templates/loader.go:58`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `contractsPath` | `string` | `` | 承载 contractsPath 的 string 值；业务上下文见本 struct 的源码包。 |
| `layouts` | `[]LayoutInfo` | `` | 承载 layouts 的 []LayoutInfo 值；业务上下文见本 struct 的源码包。 |

## `SearchApprovalInfo` — `pkg/tools/human_in_the_loop.go:38`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ToolName` | `string` | ``json:"tool_name"`` | 承载 ToolName 的 string 值；业务上下文见本 struct 的源码包。 |
| `Query` | `string` | ``json:"query"`` | 用户或系统发起的查询文本。 |
| `Reason` | `string` | ``json:"reason,omitempty"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |
| `Result` | `*SearchApprovalResult` | ``json:"result,omitempty"`` | 承载 Result 的 *SearchApprovalResult 值；业务上下文见本 struct 的源码包。 |

## `SearchApprovalResult` — `pkg/tools/human_in_the_loop.go:45`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Option` | `int` | ``json:"option"`` | 承载 Option 的 int 值；业务上下文见本 struct 的源码包。 |
| `EditedQuery` | `*string` | ``json:"edited_query,omitempty"`` | 用户或系统发起的查询文本。 |

## `InvokableSearchApprovalTool` — `pkg/tools/human_in_the_loop.go:58`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `tool.InvokableTool` | `嵌入字段` | `` | 承载 tool.InvokableTool 的 嵌入字段 值；业务上下文见本 struct 的源码包。 |

## `imageSearchTool` — `pkg/tools/image/image_search_tool.go:41`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `client` | `*unsplash.Client` | `` | 承载 client 的 *unsplash.Client 值；业务上下文见本 struct 的源码包。 |

## `imageSearchInput` — `pkg/tools/image/image_search_tool.go:43`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Query` | `string` | ``json:"query"`` | 用户或系统发起的查询文本。 |
| `Orientation` | `string` | ``json:"orientation,omitempty"`` | 承载 Orientation 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentFilter` | `string` | ``json:"content_filter,omitempty"`` | 承载 ContentFilter 的 string 值；业务上下文见本 struct 的源码包。 |
| `OrderBy` | `string` | ``json:"order_by,omitempty"`` | 承载 OrderBy 的 string 值；业务上下文见本 struct 的源码包。 |
| `PerPage` | `int` | ``json:"per_page,omitempty"`` | 承载 PerPage 的 int 值；业务上下文见本 struct 的源码包。 |
| `Reason` | `string` | ``json:"reason,omitempty"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |

## `ImageSearchResponse` — `pkg/tools/image/image_search_tool.go:52`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Provider` | `string` | ``json:"provider"`` | 承载 Provider 的 string 值；业务上下文见本 struct 的源码包。 |
| `Photos` | `[]ImagePhoto` | ``json:"photos"`` | 承载 Photos 的 []ImagePhoto 值；业务上下文见本 struct 的源码包。 |

## `ImagePhoto` — `pkg/tools/image/image_search_tool.go:57`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `ImageURL` | `string` | ``json:"image_url"`` | 承载 ImageURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `PreviewURL` | `string` | ``json:"preview_url,omitempty"`` | 承载 PreviewURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `SourceURL` | `string` | ``json:"source_url"`` | 承载 SourceURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Photographer` | `string` | ``json:"photographer,omitempty"`` | 承载 Photographer 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhotographerURL` | `string` | ``json:"photographer_url,omitempty"`` | 承载 PhotographerURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Attribution` | `string` | ``json:"attribution"`` | 承载 Attribution 的 string 值；业务上下文见本 struct 的源码包。 |

## `readFileTool` — `pkg/tools/read_file.go:56`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `op` | `commandline.Operator` | `` | 承载 op 的 commandline.Operator 值；业务上下文见本 struct 的源码包。 |

## `options` — `pkg/tools/read_file.go:60`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `op` | `commandline.Operator` | `` | 承载 op 的 commandline.Operator 值；业务上下文见本 struct 的源码包。 |

## `readInput` — `pkg/tools/read_file.go:68`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Path` | `string` | ``json:"path"`` | 承载 Path 的 string 值；业务上下文见本 struct 的源码包。 |
| `StartRow` | `*int` | ``json:"start_row"`` | 承载 StartRow 的 *int 值；业务上下文见本 struct 的源码包。 |
| `NRows` | `*int` | ``json:"n_rows"`` | 承载 NRows 的 *int 值；业务上下文见本 struct 的源码包。 |

## `tokenBucket` — `pkg/tools/search/search_tool.go:56`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `mu` | `sync.Mutex` | `` | 承载 mu 的 sync.Mutex 值；业务上下文见本 struct 的源码包。 |
| `rate` | `float64` | `` | 承载 rate 的 float64 值；业务上下文见本 struct 的源码包。 |
| `burst` | `float64` | `` | 承载 burst 的 float64 值；业务上下文见本 struct 的源码包。 |
| `tokens` | `float64` | `` | 模型调用的 token 统计或限制参数。 |
| `lastFill` | `time.Time` | `` | 承载 lastFill 的 time.Time 值；业务上下文见本 struct 的源码包。 |

## `searchTool` — `pkg/tools/search/search_tool.go:131`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `summarizer` | `ContentSummarizer` | `` | 承载 summarizer 的 ContentSummarizer 值；业务上下文见本 struct 的源码包。 |
| `urlReader` | `URLContentReader` | `` | 承载 urlReader 的 URLContentReader 值；业务上下文见本 struct 的源码包。 |

## `URLContent` — `pkg/tools/search/search_tool.go:140`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Title` | `string` | `` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `Text` | `string` | `` | 承载 Text 的 string 值；业务上下文见本 struct 的源码包。 |
| `Source` | `string` | `` | 承载 Source 的 string 值；业务上下文见本 struct 的源码包。 |

## `searchInput` — `pkg/tools/search/search_tool.go:160`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Query` | `string` | ``json:"query"`` | 用户或系统发起的查询文本。 |
| `Reason` | `string` | ``json:"reason,omitempty"`` | 承载 Reason 的 string 值；业务上下文见本 struct 的源码包。 |

## `SearchResponse` — `pkg/tools/search/search_tool.go:165`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Results` | `[]SearchResult` | ``json:"results"`` | 承载 Results 的 []SearchResult 值；业务上下文见本 struct 的源码包。 |
| `Content` | `string` | ``json:"content,omitempty"`` | 正文或序列化内容。 |
| `Mode` | `string` | ``json:"mode,omitempty"`` | 承载 Mode 的 string 值；业务上下文见本 struct 的源码包。 |
| `Error` | `string` | ``json:"error,omitempty"`` | 失败原因或错误详情。 |

## `SearchResult` — `pkg/tools/search/search_tool.go:172`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `URL` | `string` | ``json:"url"`` | 承载 URL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Description` | `string` | ``json:"description"`` | 承载 Description 的 string 值；业务上下文见本 struct 的源码包。 |
| `Source` | `string` | ``json:"source,omitempty"`` | 承载 Source 的 string 值；业务上下文见本 struct 的源码包。 |
| `Date` | `string` | ``json:"date,omitempty"`` | 承载 Date 的 string 值；业务上下文见本 struct 的源码包。 |

## `qianfanRequest` — `pkg/tools/search/search_tool.go:182`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Messages` | `[]qianfanMessage` | ``json:"messages"`` | 承载 Messages 的 []qianfanMessage 值；业务上下文见本 struct 的源码包。 |
| `SearchSource` | `string` | ``json:"search_source"`` | 承载 SearchSource 的 string 值；业务上下文见本 struct 的源码包。 |
| `SearchFilter` | `*qianfanSearchFilter` | ``json:"search_filter,omitempty"`` | 承载 SearchFilter 的 *qianfanSearchFilter 值；业务上下文见本 struct 的源码包。 |
| `ResourceTypeFilter` | `[]qianfanResource` | ``json:"resource_type_filter"`` | 承载 ResourceTypeFilter 的 []qianfanResource 值；业务上下文见本 struct 的源码包。 |

## `qianfanMessage` — `pkg/tools/search/search_tool.go:189`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Content` | `string` | ``json:"content"`` | 正文或序列化内容。 |
| `Role` | `string` | ``json:"role"`` | 承载 Role 的 string 值；业务上下文见本 struct 的源码包。 |

## `qianfanResource` — `pkg/tools/search/search_tool.go:194`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Type` | `string` | ``json:"type"`` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `TopK` | `int` | ``json:"top_k"`` | 承载 TopK 的 int 值；业务上下文见本 struct 的源码包。 |

## `qianfanSearchFilter` — `pkg/tools/search/search_tool.go:199`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Match` | `*qianfanMatch` | ``json:"match,omitempty"`` | 承载 Match 的 *qianfanMatch 值；业务上下文见本 struct 的源码包。 |

## `qianfanMatch` — `pkg/tools/search/search_tool.go:203`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Site` | `[]string` | ``json:"site,omitempty"`` | 承载 Site 的 []string 值；业务上下文见本 struct 的源码包。 |

## `qianfanResponse` — `pkg/tools/search/search_tool.go:207`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `RequestID` | `string` | ``json:"request_id"`` | 承载 RequestID 的 string 值；业务上下文见本 struct 的源码包。 |
| `Code` | `string` | ``json:"code"`` | 承载 Code 的 string 值；业务上下文见本 struct 的源码包。 |
| `Message` | `string` | ``json:"message"`` | 承载 Message 的 string 值；业务上下文见本 struct 的源码包。 |
| `References` | `[]qianfanRef` | ``json:"references"`` | 承载 References 的 []qianfanRef 值；业务上下文见本 struct 的源码包。 |

## `qianfanRef` — `pkg/tools/search/search_tool.go:214`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `int` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `URL` | `string` | ``json:"url"`` | 承载 URL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Title` | `string` | ``json:"title"`` | 承载 Title 的 string 值；业务上下文见本 struct 的源码包。 |
| `Date` | `string` | ``json:"date"`` | 承载 Date 的 string 值；业务上下文见本 struct 的源码包。 |
| `Content` | `string` | ``json:"content"`` | 正文或序列化内容。 |
| `Snippet` | `string` | ``json:"snippet"`` | 承载 Snippet 的 string 值；业务上下文见本 struct 的源码包。 |
| `Icon` | `string` | ``json:"icon"`` | 承载 Icon 的 string 值；业务上下文见本 struct 的源码包。 |
| `WebAnchor` | `string` | ``json:"web_anchor"`` | 承载 WebAnchor 的 string 值；业务上下文见本 struct 的源码包。 |
| `Type` | `string` | ``json:"type"`` | 承载 Type 的 string 值；业务上下文见本 struct 的源码包。 |
| `Website` | `string` | ``json:"website"`` | 承载 Website 的 string 值；业务上下文见本 struct 的源码包。 |
| `RerankScore` | `float64` | ``json:"rerank_score"`` | 承载 RerankScore 的 float64 值；业务上下文见本 struct 的源码包。 |
| `AuthorityScore` | `float64` | ``json:"authority_score"`` | 承载 AuthorityScore 的 float64 值；业务上下文见本 struct 的源码包。 |

## `evidenceBuilder` — `pkg/tools/search/search_tool.go:396`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `builder` | `strings.Builder` | `` | 承载 builder 的 strings.Builder 值；业务上下文见本 struct 的源码包。 |

## `contextParams` — `pkg/utils/params/context.go:8`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `typedContextParams` — `pkg/utils/params/context.go:12`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| — | — | — | 无字段；作为类型标记或零状态载体使用。 |

## `Client` — `pkg/utils/unsplash/client.go:156`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `accessKey` | `string` | `` | 承载 accessKey 的 string 值；业务上下文见本 struct 的源码包。 |
| `baseURL` | `*url.URL` | `` | 承载 baseURL 的 *url.URL 值；业务上下文见本 struct 的源码包。 |
| `httpClient` | `*http.Client` | `` | 承载 httpClient 的 *http.Client 值；业务上下文见本 struct 的源码包。 |
| `maxDownloadBytes` | `int64` | `` | 承载 maxDownloadBytes 的 int64 值；业务上下文见本 struct 的源码包。 |
| `allowedDownloadHosts` | `map[string]struct{}` | `` | 承载 allowedDownloadHosts 的 map[string]struct{} 值；业务上下文见本 struct 的源码包。 |

## `DownloadedAsset` — `pkg/utils/unsplash/helpers.go:15`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `PhotoID` | `string` | ``json:"photo_id"`` | 承载 PhotoID 的 string 值；业务上下文见本 struct 的源码包。 |
| `LocalPath` | `string` | ``json:"local_path"`` | 承载 LocalPath 的 string 值；业务上下文见本 struct 的源码包。 |
| `ImageURL` | `string` | ``json:"image_url"`` | 承载 ImageURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `SourceURL` | `string` | ``json:"source_url"`` | 承载 SourceURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Photographer` | `string` | ``json:"photographer"`` | 承载 Photographer 的 string 值；业务上下文见本 struct 的源码包。 |
| `PhotographerURL` | `string` | ``json:"photographer_url"`` | 承载 PhotographerURL 的 string 值；业务上下文见本 struct 的源码包。 |
| `Attribution` | `string` | ``json:"attribution"`` | 承载 Attribution 的 string 值；业务上下文见本 struct 的源码包。 |
| `Width` | `int` | ``json:"width"`` | 承载 Width 的 int 值；业务上下文见本 struct 的源码包。 |
| `Height` | `int` | ``json:"height"`` | 承载 Height 的 int 值；业务上下文见本 struct 的源码包。 |

## `SearchOptions` — `pkg/utils/unsplash/types.go:4`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Query` | `string` | `` | 用户或系统发起的查询文本。 |
| `Orientation` | `string` | `` | 承载 Orientation 的 string 值；业务上下文见本 struct 的源码包。 |
| `ContentFilter` | `string` | `` | 承载 ContentFilter 的 string 值；业务上下文见本 struct 的源码包。 |
| `Color` | `string` | `` | 承载 Color 的 string 值；业务上下文见本 struct 的源码包。 |
| `OrderBy` | `string` | `` | 承载 OrderBy 的 string 值；业务上下文见本 struct 的源码包。 |
| `Page` | `int` | `` | 承载 Page 的 int 值；业务上下文见本 struct 的源码包。 |
| `PerPage` | `int` | `` | 承载 PerPage 的 int 值；业务上下文见本 struct 的源码包。 |

## `SearchResponse` — `pkg/utils/unsplash/types.go:15`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Total` | `int` | ``json:"total"`` | 承载 Total 的 int 值；业务上下文见本 struct 的源码包。 |
| `TotalPages` | `int` | ``json:"total_pages"`` | 承载 TotalPages 的 int 值；业务上下文见本 struct 的源码包。 |
| `Results` | `[]Photo` | ``json:"results"`` | 承载 Results 的 []Photo 值；业务上下文见本 struct 的源码包。 |

## `Photo` — `pkg/utils/unsplash/types.go:22`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `ID` | `string` | ``json:"id"`` | 当前记录或对象的唯一标识。 |
| `Width` | `int` | ``json:"width"`` | 承载 Width 的 int 值；业务上下文见本 struct 的源码包。 |
| `Height` | `int` | ``json:"height"`` | 承载 Height 的 int 值；业务上下文见本 struct 的源码包。 |
| `Color` | `string` | ``json:"color"`` | 承载 Color 的 string 值；业务上下文见本 struct 的源码包。 |
| `Description` | `string` | ``json:"description"`` | 承载 Description 的 string 值；业务上下文见本 struct 的源码包。 |
| `AltDescription` | `string` | ``json:"alt_description"`` | 承载 AltDescription 的 string 值；业务上下文见本 struct 的源码包。 |
| `BlurHash` | `string` | ``json:"blur_hash"`` | 承载 BlurHash 的 string 值；业务上下文见本 struct 的源码包。 |
| `URLs` | `PhotoURLs` | ``json:"urls"`` | 承载 URLs 的 PhotoURLs 值；业务上下文见本 struct 的源码包。 |
| `Links` | `PhotoLinks` | ``json:"links"`` | 承载 Links 的 PhotoLinks 值；业务上下文见本 struct 的源码包。 |
| `User` | `User` | ``json:"user"`` | 承载 User 的 User 值；业务上下文见本 struct 的源码包。 |

## `PhotoURLs` — `pkg/utils/unsplash/types.go:35`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Raw` | `string` | ``json:"raw"`` | 承载 Raw 的 string 值；业务上下文见本 struct 的源码包。 |
| `Full` | `string` | ``json:"full"`` | 承载 Full 的 string 值；业务上下文见本 struct 的源码包。 |
| `Regular` | `string` | ``json:"regular"`` | 承载 Regular 的 string 值；业务上下文见本 struct 的源码包。 |
| `Small` | `string` | ``json:"small"`` | 承载 Small 的 string 值；业务上下文见本 struct 的源码包。 |
| `Thumb` | `string` | ``json:"thumb"`` | 承载 Thumb 的 string 值；业务上下文见本 struct 的源码包。 |

## `PhotoLinks` — `pkg/utils/unsplash/types.go:43`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `HTML` | `string` | ``json:"html"`` | 承载 HTML 的 string 值；业务上下文见本 struct 的源码包。 |
| `Download` | `string` | ``json:"download"`` | 承载 Download 的 string 值；业务上下文见本 struct 的源码包。 |
| `DownloadLocation` | `string` | ``json:"download_location"`` | 承载 DownloadLocation 的 string 值；业务上下文见本 struct 的源码包。 |

## `User` — `pkg/utils/unsplash/types.go:49`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `Name` | `string` | ``json:"name"`` | 承载 Name 的 string 值；业务上下文见本 struct 的源码包。 |
| `Username` | `string` | ``json:"username"`` | 承载 Username 的 string 值；业务上下文见本 struct 的源码包。 |
| `Links` | `UserLinks` | ``json:"links"`` | 承载 Links 的 UserLinks 值；业务上下文见本 struct 的源码包。 |

## `UserLinks` — `pkg/utils/unsplash/types.go:55`

| 字段 | Go 类型 | Tag / 约束 | 说明 |
| --- | --- | --- | --- |
| `HTML` | `string` | ``json:"html"`` | 承载 HTML 的 string 值；业务上下文见本 struct 的源码包。 |

