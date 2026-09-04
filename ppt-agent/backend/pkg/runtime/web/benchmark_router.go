package web

import (
	"context"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
)

// BenchmarkCreateRouteResult is the stable evaluation vocabulary used by
// pptbench. It records the production route's downstream handoff explicitly,
// so benchmark cases remain comparable when the HTTP API uses shorter intent
// names such as "create" and "fix".
type BenchmarkCreateRouteResult struct {
	Intent                string  `json:"intent"`
	TargetAgent           string  `json:"target_agent,omitempty"`
	NormalizedRequest     string  `json:"normalized_request,omitempty"`
	Reason                string  `json:"reason"`
	ClarificationQuestion string  `json:"clarification_question,omitempty"`
	Confidence            float64 `json:"confidence,omitempty"`
}

// BenchmarkMessageRouteResult records the route selected for a persistent
// workbench conversation. Every first message receives a task ID before its
// intent is resolved, so this result keeps that identity observable.
type BenchmarkMessageRouteResult struct {
	Intent            string  `json:"intent"`
	TargetAgent       string  `json:"target_agent,omitempty"`
	TaskID            string  `json:"task_id"`
	NormalizedRequest string  `json:"normalized_request,omitempty"`
	Action            string  `json:"action,omitempty"`
	Reason            string  `json:"reason"`
	Confidence        float64 `json:"confidence,omitempty"`
}

// ClassifyCreateRequestForBenchmark runs the same create-entry router used by
// the HTTP task creation path. The API key is benchmark-only configuration and
// is never read by the production handler.
func ClassifyCreateRequestForBenchmark(ctx context.Context, query string, hasOutline bool, apiKey string) BenchmarkCreateRouteResult {
	server := &Server{}
	credential := modelCredential{Provider: "deepseek", APIKey: strings.TrimSpace(apiKey)}
	return benchmarkCreateRoute(server.routeCreateRequest(ctx, query, hasOutline, credential), query)
}

// ClassifyTaskMessageForBenchmark runs the production message router with the
// bounded task context used by the Dashboard. It verifies that a follow-up is
// prepared on the existing task rather than treated as a stateless request.
func ClassifyTaskMessageForBenchmark(ctx context.Context, message, taskID, conversationContext, apiKey string) BenchmarkMessageRouteResult {
	server := &Server{}
	credential := modelCredential{Provider: "deepseek", APIKey: strings.TrimSpace(apiKey)}
	route := server.routeTaskMessageRequest(ctx, message, taskID, conversationContext, credential)
	result := BenchmarkMessageRouteResult{
		Intent:            route.Intent,
		TaskID:            taskID,
		NormalizedRequest: route.NormalizedRequest,
		Action:            route.Action,
		Reason:            route.Reason,
		Confidence:        route.Confidence,
	}
	switch route.Intent {
	case messageIntentCreate:
		result.Intent = "create_deck"
		result.TargetAgent = "PPTPlanner"
	case messageIntentPlan:
		result.TargetAgent = "PPTPlanner (planning only)"
	case messageIntentFix:
		if route.Action == messageActionUpdateTask {
			result.TargetAgent = "PPTFixer"
		} else {
			result.TargetAgent = "PPTFixer after selecting an existing task"
		}
	}
	return result
}

// ClassifyContinueIntentForBenchmark runs the same continuation router prompt
// used by the task continue path. The API key is benchmark-only configuration.
// tasksSummary should describe the current deck in the same concise form that
// production builds from tasks.json.
func ClassifyContinueIntentForBenchmark(ctx context.Context, message string, tasksSummary string, apiKey string) RouteResult {
	for _, kw := range []string{"再加", "添加", "新增", "加一页", "加两页", "再加一页", "再加几页"} {
		if strings.Contains(message, kw) {
			return RouteResult{Intent: "add_page", Reason: "用户明确要求新增页面", TargetPages: extractTargetPages(message)}
		}
	}
	if tasksSummary == "" {
		tasksSummary = "无法读取任务清单"
	}
	model, err := agentutils.NewFallbackToolCallingChatModel(ctx,
		agentutils.WithTextModel(),
		agentutils.WithAPIKeyForProvider("deepseek", apiKey),
	)
	if err != nil || model == nil {
		return RouteResult{Intent: "unknown", Reason: "AI 模型初始化失败: " + errString(err), TargetPages: extractTargetPages(message)}
	}
	adapter := benchmarkModelAdapter{inner: model}
	server := &Server{textModelFactory: func(context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error) {
		return adapter, nil
	}}
	return server.classifyIntentByLLM(ctx, message, tasksSummary)
}

func benchmarkCreateRoute(route createRequestRoute, query string) BenchmarkCreateRouteResult {
	result := BenchmarkCreateRouteResult{
		Intent:                route.Intent,
		Reason:                route.Reason,
		ClarificationQuestion: route.ClarificationQuestion,
		Confidence:            route.Confidence,
		NormalizedRequest:     strings.TrimSpace(query),
	}
	switch route.Intent {
	case createIntentDeck:
		result.Intent = "create_deck"
		result.TargetAgent = "PPTPlanner"
	case createIntentFixExisting:
		result.Intent = "fix_existing"
		result.TargetAgent = "PPTFixer after selecting an existing task"
	}
	return result
}

type benchmarkModelAdapter struct {
	inner einomodel.ToolCallingChatModel
}

func (a benchmarkModelAdapter) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	return a.inner.Generate(ctx, messages)
}

func errString(err error) string {
	if err == nil {
		return "unknown"
	}
	return err.Error()
}
