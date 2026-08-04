package intent_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/agent/router"
)

type routingTextModel struct {
	calls       atomic.Int32
	content     string
	generateErr error
}

func (m *routingTextModel) Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error) {
	m.calls.Add(1)
	if m.generateErr != nil {
		return nil, m.generateErr
	}
	return schema.AssistantMessage(m.content, nil), nil
}

func newTextClassifier(model *routingTextModel) *intent.Classifier {
	return intent.NewClassifier(nil, func(context.Context) (interface {
		Generate(context.Context, []*schema.Message, ...interface{}) (*schema.Message, error)
	}, error) {
		return model, nil
	})
}

func TestClassifierUsesOneLLMRouteAndRouterReusesIt(t *testing.T) {
	model := &routingTextModel{content: `{
		"intent":"create","intent_reasoning":"用户要求制作地域介绍演示",
		"domain":"creative","complexity_level":6,"page_count_estimate":18,
		"confidence":0.94,"suggested_theme":"ocean_soft","suggested_templates":["tech-intro"],
		"agent_type":"deep","pipeline":["plan","generate"],"concurrency":5
	}`}
	classifier := newTextClassifier(model)
	classification, err := classifier.Classify(context.Background(), "介绍一下大兴安岭", 1)
	if err != nil {
		t.Fatal(err)
	}
	if classification.Intent != intent.IntentCreate || classification.Domain != intent.DomainCreative {
		t.Fatalf("classification=%#v", classification)
	}
	if classification.RoutingSource != "llm" || classification.SuggestedPageCount != 18 {
		t.Fatalf("route metadata=%#v", classification)
	}

	decision, err := router.NewEngine(classifier).Route(classification, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := model.calls.Load(); got != 1 {
		t.Fatalf("model calls=%d, want one shared classification", got)
	}
	if decision.AgentType != "deep" || len(decision.Pipeline) != 2 || decision.Pipeline[0] != "plan" || decision.Pipeline[1] != "generate" {
		t.Fatalf("decision=%#v", decision)
	}
	if !decision.SkipQA || !decision.SkipFix || decision.Source != "llm" {
		t.Fatalf("decision flags=%#v", decision)
	}
}

func TestClassifierUsesDeterministicFallbackWithoutRules(t *testing.T) {
	model := &routingTextModel{generateErr: errors.New("model unavailable")}
	classification, err := newTextClassifier(model).Classify(context.Background(), "修改第3页", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := model.calls.Load(); got != 1 {
		t.Fatalf("model calls=%d, want 1", got)
	}
	// The fallback is operational, not a keyword interpretation of "修改".
	if classification.Intent != intent.IntentCreate || classification.AgentType != "deep" || classification.RoutingSource != "fallback" {
		t.Fatalf("fallback=%#v", classification)
	}
	if classification.SuggestedPageCount <= 0 || classification.Complexity.PageCountEstimate <= 0 {
		t.Fatalf("fallback contains zero delivery estimates: %#v", classification)
	}
}

func TestRouterNormalizesUnsupportedAgentAndRemovesQA(t *testing.T) {
	classification := &intent.ClassificationResult{
		Intent: intent.IntentCreate, AgentType: "quick", Pipeline: []string{"quick_generate", "qa"},
		Concurrency: 99, RoutingSource: "llm", IntentReasoning: "short deck",
	}
	decision, err := router.NewEngine(nil).Route(classification, nil)
	if err != nil {
		t.Fatal(err)
	}
	if decision.AgentType != "deep" || decision.Concurrency != 10 {
		t.Fatalf("decision=%#v", decision)
	}
	if !decision.SkipQA || !decision.SkipFix || len(decision.Pipeline) != 2 {
		t.Fatalf("QA-free pipeline not enforced: %#v", decision)
	}
}
