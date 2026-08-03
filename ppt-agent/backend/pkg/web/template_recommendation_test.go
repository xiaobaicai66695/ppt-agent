package web

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

func testTemplateServer(t *testing.T) *Server {
	t.Helper()
	projectRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	skills := filepath.Join(projectRoot, "skills", "visual_designer")
	loader := templates.NewLoader(
		filepath.Join(skills, "templates", "full-decks"),
		filepath.Join(skills, "templates", "single-page"),
		filepath.Join(skills, "background_templates"),
	)
	if len(loader.ListPresets()) == 0 || len(loader.ListLayouts()) == 0 {
		t.Fatal("template fixtures did not load")
	}
	return &Server{templateLoader: loader}
}

func TestRecommendTemplateStrategyUsesTopicAndRealResources(t *testing.T) {
	server := testTemplateServer(t)
	strategy, preset, err := server.recommendTemplateStrategy("制作新能源汽车行业市场调研和可行性研究报告")
	if err != nil {
		t.Fatal(err)
	}
	if preset.Name != "research-report" || strategy.Template != preset.Name {
		t.Fatalf("strategy = %#v, preset = %q", strategy, preset.Name)
	}
	if server.templateLoader.GetPreset(strategy.Template) == nil {
		t.Fatalf("recommended missing template %q", strategy.Template)
	}
	if _, err := server.getTheme(strategy.Theme); err != nil {
		t.Fatalf("recommended missing theme %q", strategy.Theme)
	}
}

func TestRecommendTemplateStrategyFallsBackToGeneric(t *testing.T) {
	server := testTemplateServer(t)
	strategy, _, err := server.recommendTemplateStrategy("说明一些尚未分类的新想法")
	if err != nil {
		t.Fatal(err)
	}
	if strategy.Template != "generic" {
		t.Fatalf("template = %q, want generic", strategy.Template)
	}
}

func TestRecommendationAppliesBackgroundOnlyToVisualPages(t *testing.T) {
	server := testTemplateServer(t)
	outline, strategy, err := server.resolveTemplateSelection("中国风传统文化主题分享", &TemplateSelection{Mode: "recommended"})
	if err != nil {
		t.Fatal(err)
	}
	if !strategy.UseBackground || strategy.Background == "" {
		t.Fatalf("strategy = %#v, want background", strategy)
	}
	for _, slide := range outline.Slides {
		if isVisualBackgroundPage(slide.ContentType) && slide.Background == "" {
			t.Fatalf("visual slide %q has no background", slide.ContentType)
		}
		if !isVisualBackgroundPage(slide.ContentType) && slide.Background == strategy.Background {
			t.Fatalf("dense slide %q received recommended background", slide.ContentType)
		}
	}
}

func TestResolveTemplateSelectionRejectsMissingPreset(t *testing.T) {
	server := testTemplateServer(t)
	if _, _, err := server.resolveTemplateSelection("主题", &TemplateSelection{Mode: "preset", Template: "missing"}); err == nil {
		t.Fatal("expected missing preset error")
	}
}

func TestPrepareOutlineDoesNotEnrichOrReplaceUserContent(t *testing.T) {
	server := testTemplateServer(t)
	outline := &deep.TaskOutline{
		Template: "custom", Theme: "ocean_soft", Title: "用户标题",
		Slides: []deep.SlideOutline{
			{Title: "已有标题", ContentType: "content_slide", Description: "用户已经填写的短内容"},
			{Title: "", ContentType: "summary_slide", Description: ""},
		},
	}
	got, err := server.prepareOutline(context.Background(), "主题", outline)
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentMode != deep.OutlineContentModeUserOutline {
		t.Fatalf("content mode = %q", got.ContentMode)
	}
	if got.Slides[0].Description != "用户已经填写的短内容" {
		t.Fatalf("user content was replaced: %#v", got.Slides[0])
	}
	if got.Slides[1].Description != "" {
		t.Fatalf("blank content was unexpectedly enriched: %#v", got.Slides[1])
	}
	if got.Slides[1].Title != "" {
		t.Fatalf("blank title was unexpectedly replaced: %#v", got.Slides[1])
	}
}
