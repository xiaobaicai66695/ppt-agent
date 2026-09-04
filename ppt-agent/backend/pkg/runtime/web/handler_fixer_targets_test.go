package web

import (
	"reflect"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

func TestResolveRouteTaskIDsUsesCurrentManifestIdentity(t *testing.T) {
	manifest := &deck.TasksManifest{Tasks: []*deck.TaskItem{
		{TaskID: "task-intro", PageIndex: 1, Title: "引言"},
		{TaskID: "task-metrics", PageIndex: 2, Title: "指标"},
	}}
	route := resolveRouteTaskIDs(RouteResult{Intent: "fix", TargetPages: []int{2, 2, 99, 1}}, manifest)
	if want := []string{"task-metrics", "task-intro"}; !reflect.DeepEqual(route.TargetTaskIDs, want) {
		t.Fatalf("target task ids = %#v, want %#v", route.TargetTaskIDs, want)
	}
}

func TestResolveRouteTaskIDsOnlyAttachesIDsForFix(t *testing.T) {
	manifest := &deck.TasksManifest{Tasks: []*deck.TaskItem{{TaskID: "task-1", PageIndex: 1}}}
	route := resolveRouteTaskIDs(RouteResult{Intent: "regenerate", TargetPages: []int{1}}, manifest)
	if len(route.TargetTaskIDs) != 0 {
		t.Fatalf("non-fix route should not carry task ids: %#v", route.TargetTaskIDs)
	}
}
