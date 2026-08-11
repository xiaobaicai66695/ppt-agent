package learning

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/cloudwego/eino/schema"
)

type learningRoutingModel struct {
	calls atomic.Int32
}

func (m *learningRoutingModel) Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error) {
	m.calls.Add(1)
	return schema.AssistantMessage(`{
		"intent":"create","intent_reasoning":"生成地域介绍演示","domain":"creative",
		"complexity_level":5,"page_count_estimate":12,"confidence":0.91,
		"suggested_theme":"ocean_soft","suggested_templates":[],
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
