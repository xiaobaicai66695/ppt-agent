package search

import (
	"context"
	"testing"
)

func TestInvokableRun_RealSearch(t *testing.T) {
	tool := NewSearchTool()
	ctx := context.Background()

	input := `{"query": "ReAct", "reason": "需要最新版本号"}`

	result, err := tool.InvokableRun(ctx, input)
	if err != nil {
		t.Fatalf("InvokableRun error = %v", err)
	}

	t.Logf("工具返回结果: %s", result)
}
