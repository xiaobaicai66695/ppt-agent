package web

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	agentintent "github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/templates"
)

func testTemplateServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("UNSPLASH_ACCESS_KEY", "")
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

func TestPresetSelectionSkipsLocalBackgroundWhenImageSearchConfigured(t *testing.T) {
	server := testTemplateServer(t)
	t.Setenv("UNSPLASH_ACCESS_KEY", "test-access-key")

	outline, strategy, err := server.resolveTemplateSelection("低空经济场景分析", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if strategy.UseBackground || strategy.Background != "" || outline.UseBackground || outline.RecommendedBackground != "" {
		t.Fatalf("local background should be disabled when image search is configured: strategy=%#v outline=%#v", strategy, outline)
	}
	for _, slide := range outline.Slides {
		if slide.Background != "" {
			t.Fatalf("slide %q unexpectedly has local background %q", slide.ContentType, slide.Background)
		}
	}
}

func TestRecommendTemplateStrategyUsesTopicAndRealResources(t *testing.T) {
	server := testTemplateServer(t)
	useBackground := true
	intent := &agentintent.ClassificationResult{
		IntentReasoning:    "行业研究需要数据驱动的叙事结构",
		SuggestedTemplates: []string{"missing", "research-report"},
		SuggestedTheme:     "report_green", SuggestedBackground: "minimalist_blue",
		SuggestedPageCount: 13, UseBackground: &useBackground,
	}
	strategy, preset, err := server.recommendTemplateStrategyWithIntent(intent)
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
	if strategy.PageCount != 13 || strategy.Background != "minimalist_blue" {
		t.Fatalf("recommendation did not reuse LLM result: %#v", strategy)
	}
}

func TestRecommendedTemplateUsesBackgroundPaletteOverIntentTheme(t *testing.T) {
	server := testTemplateServer(t)
	useBackground := true
	intent := &agentintent.ClassificationResult{
		Intent:              agentintent.IntentCreate,
		Domain:              agentintent.DomainGovernment,
		SuggestedTemplates:  []string{"current-affairs"},
		SuggestedTheme:      "patriotic_blue",
		SuggestedBackground: "party_government",
		SuggestedPageCount:  10,
		UseBackground:       &useBackground,
	}
	outline, strategy, err := server.resolveTemplateSelectionWithIntent("介绍延安", &TemplateSelection{Mode: "recommended"}, intent)
	if err != nil {
		t.Fatal(err)
	}
	if strategy.Theme != "government_red" || outline.Theme != "government_red" {
		t.Fatalf("theme = %q outline=%q, want government_red", strategy.Theme, outline.Theme)
	}
	if strategy.Background != "party_government" {
		t.Fatalf("background = %q, want party_government", strategy.Background)
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

func TestRecommendationProducesStyleGuidanceWithoutPresetSlides(t *testing.T) {
	server := testTemplateServer(t)
	useBackground := true
	intent := &agentintent.ClassificationResult{
		SuggestedTemplates: []string{"course-module"}, SuggestedTheme: "sage_calm",
		SuggestedBackground: "vintage_chinese", SuggestedPageCount: 9, UseBackground: &useBackground,
	}
	outline, strategy, err := server.resolveTemplateSelectionWithIntent("中国风传统文化主题分享", &TemplateSelection{Mode: "recommended"}, intent)
	if err != nil {
		t.Fatal(err)
	}
	if !strategy.UseBackground || strategy.Background == "" {
		t.Fatalf("strategy = %#v, want background", strategy)
	}
	if outline.ContentMode != deck.OutlineContentModeRecommendedStyle || len(outline.Slides) != 0 {
		t.Fatalf("recommended outline copied preset slides: %#v", outline)
	}
	if outline.SuggestedPageCount != 9 || outline.Template != "course-module" || outline.Theme != "warm_terracotta" {
		t.Fatalf("recommended style guidance = %#v", outline)
	}
}

func TestRecommendationUsesExplicitUserPageCountBeforeLLMSuggestion(t *testing.T) {
	server := testTemplateServer(t)
	useBackground := true
	intent := &agentintent.ClassificationResult{
		SuggestedTemplates: []string{"generic"}, SuggestedTheme: "ocean_soft",
		SuggestedPageCount: 5, UseBackground: &useBackground,
	}

	outline, strategy, err := server.resolveTemplateSelectionWithIntent(
		"请生成2页关于绿色数据中心节能实践的PPT", &TemplateSelection{Mode: "recommended"}, intent)
	if err != nil {
		t.Fatal(err)
	}
	if strategy.PageCount != 2 || outline.SuggestedPageCount != 2 {
		t.Fatalf("explicit page count was not preserved: strategy=%#v outline=%#v", strategy, outline)
	}
}

func TestExplicitPageCountDoesNotTreatPageReferenceAsDeckSize(t *testing.T) {
	if pageCount, ok := explicitRequestedPageCount("修改第3页的数据，并保持原有页数"); ok {
		t.Fatalf("page reference was treated as deck size: %d", pageCount)
	}
	if pageCount, ok := explicitRequestedPageCount("Create a 7-slide presentation about renewable energy"); !ok || pageCount != 7 {
		t.Fatalf("english page count = %d, %v", pageCount, ok)
	}
}

func TestPresetSelectionAssignsBackgroundToEverySlide(t *testing.T) {
	server := testTemplateServer(t)
	outline, _, err := server.resolveTemplateSelection("说明一些尚未分类的新想法", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range outline.Slides {
		if slide.Background == "" {
			t.Fatalf("slide %q should receive a background", slide.ContentType)
		}
		if backgroundTheme(slide.Background) != outline.RecommendedBackground {
			t.Fatalf("slide background = %q, want theme %q", slide.Background, outline.RecommendedBackground)
		}
	}
}

func TestRecommendationAvoidsAdjacentDuplicateBackgroundImages(t *testing.T) {
	server := testTemplateServer(t)
	outline, strategy, err := server.resolveTemplateSelection("党建政府工作汇报", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if strategy.Background != "party_government" {
		t.Fatalf("background = %q, want party_government", strategy.Background)
	}
	previous := ""
	for _, slide := range outline.Slides {
		if !strings.HasPrefix(slide.Background, "party_government/") {
			t.Fatalf("slide background = %q, want concrete party_government image ref", slide.Background)
		}
		if previous != "" && backgroundTheme(previous) == backgroundTheme(slide.Background) && previous == slide.Background {
			t.Fatalf("adjacent visual backgrounds repeated: %q", slide.Background)
		}
		previous = slide.Background
	}
}

func TestPresetSelectionUsesDefaultBackgroundUnlessSuppressed(t *testing.T) {
	server := testTemplateServer(t)
	outline, strategy, err := server.resolveTemplateSelection("说明一些尚未分类的新想法", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if !strategy.UseBackground || strategy.Background != "minimalist_blue" {
		t.Fatalf("strategy = %#v, want default minimalist background", strategy)
	}
	if outline.Slides[0].Background == "" {
		t.Fatal("expected title slide to receive background")
	}
}

func TestPresetSelectionCanSuppressBackground(t *testing.T) {
	server := testTemplateServer(t)
	outline, strategy, err := server.resolveTemplateSelection("说明一些尚未分类的新想法，不要背景", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if strategy.UseBackground {
		t.Fatalf("strategy = %#v, want background suppressed", strategy)
	}
	for _, slide := range outline.Slides {
		if slide.Background != "" {
			t.Fatalf("slide %q unexpectedly has background %q", slide.ContentType, slide.Background)
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
	outline := &deck.TaskOutline{
		Template: "custom", Theme: "ocean_soft", Title: "用户标题",
		Slides: []deck.SlideOutline{
			{Title: "已有标题", ContentType: "content_slide", Description: "用户已经填写的短内容"},
			{Title: "", ContentType: "summary_slide", Description: ""},
		},
	}
	got, err := server.prepareOutline(context.Background(), "主题", outline)
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentMode != deck.OutlineContentModeUserOutline {
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
