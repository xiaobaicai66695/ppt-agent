package learning

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cloudwego/eino/schema"
)

type learningRoutingModel struct {
	calls       atomic.Int32
	generateErr error
}

func (m *learningRoutingModel) Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error) {
	m.calls.Add(1)
	if m.generateErr != nil {
		return nil, m.generateErr
	}
	return schema.AssistantMessage(`{
		"intent":"create","intent_reasoning":"生成地域介绍演示","domain":"creative",
		"complexity_level":5,"page_count_estimate":12,"confidence":0.91,
		"suggested_theme":"ocean_soft",
		"agent_type":"planner","pipeline":["plan","generate"],"concurrency":5
	}`, nil), nil
}

func TestProcessTaskUsesConfiguredTextFactoryOnce(t *testing.T) {
	model := &learningRoutingModel{}
	factory := func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return model, nil
	}
	engine := NewEngine(&EngineConfig{EnableLLMClassification: true}, nil, factory)
	defer engine.Close()

	classification, err := engine.classifier.Classify(context.Background(), "介绍一下大兴安岭", 1)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := engine.router.Route(classification, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := model.calls.Load(); got != 1 {
		t.Fatalf("model calls=%d, want exactly one", got)
	}
	if classification.RoutingSource != "llm" {
		t.Fatalf("intent=%#v", classification)
	}
	if decision.AgentType != "planner" || len(decision.Pipeline) != 2 {
		t.Fatalf("routing=%#v", decision)
	}
}

func TestProcessTaskCanUseGenerateOnlyAIModelFactoryForRouting(t *testing.T) {
	model := &learningRoutingModel{}
	factory := func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return model, nil
	}
	engine := NewEngine(&EngineConfig{EnableLLMClassification: true}, factory, nil)
	defer engine.Close()

	classification, err := engine.classifier.Classify(context.Background(), "介绍延安", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := model.calls.Load(); got != 1 {
		t.Fatalf("model calls=%d, want exactly one", got)
	}
	if classification == nil || classification.RoutingSource != "llm" {
		t.Fatalf("intent=%#v", classification)
	}
	decision, err := engine.router.Route(classification, nil)
	if err != nil {
		t.Fatal(err)
	}
	if decision.AgentType != "planner" {
		t.Fatalf("routing=%#v", decision)
	}
}

func TestProcessTaskTextFailureDoesNotTryGenerateOnlyToolConversion(t *testing.T) {
	textModel := &learningRoutingModel{generateErr: errors.New("context deadline exceeded")}
	textFactory := func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return textModel, nil
	}
	aiFactory := func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return &learningRoutingModel{}, nil
	}
	engine := NewEngine(&EngineConfig{EnableLLMClassification: true}, aiFactory, textFactory)
	defer engine.Close()

	classification, err := engine.classifier.Classify(context.Background(), "介绍延安", 1)
	if err != nil {
		t.Fatal(err)
	}
	if classification == nil || classification.RoutingSource != "fallback" {
		t.Fatalf("intent=%#v", classification)
	}
	if strings.Contains(classification.IntentReasoning, "ToolCallingChatModel") ||
		strings.Contains(classification.IntentReasoning, "深度生成流程") {
		t.Fatalf("unexpected fallback reason: %q", classification.IntentReasoning)
	}
}
