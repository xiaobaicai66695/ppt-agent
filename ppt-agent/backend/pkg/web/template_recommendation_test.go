package web

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	skills := filepath.Join(projectRoot, "skills", "ppt-deck-planner")
	loader := templates.NewComponentLoader(skills)
	if len(loader.ListPresets()) == 0 || len(loader.ListLayouts()) == 0 {
		t.Fatal("template fixtures did not load")
	}
	return &Server{templateLoader: loader}
}

func TestPresetSelectionDoesNotExposeLocalBackgroundWhenImageSearchConfigured(t *testing.T) {
	server := testTemplateServer(t)
	t.Setenv("UNSPLASH_ACCESS_KEY", "test-access-key")

	outline, strategy, err := server.resolveTemplateSelection("低空经济场景分析", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if strategy.VisualHint != "" || strategy.UseVisualAssets {
		t.Fatalf("preset selection should not assign visual search policy: strategy=%#v outline=%#v", strategy, outline)
	}
	if len(outline.Slides) == 0 {
		t.Fatal("preset outline should keep template slides")
	}
}

func TestRecommendTemplateStrategyUsesTopicAndRealResources(t *testing.T) {
	server := testTemplateServer(t)
	useVisualAssets := true
	intent := &agentintent.ClassificationResult{
		IntentReasoning:    "行业研究需要数据驱动的叙事结构",
		SuggestedTemplates: []string{"missing", "research-report"},
		SuggestedTheme:     "report_green", VisualHint: "modern research workspace",
		SuggestedPageCount: 13, UseVisualAssets: &useVisualAssets,
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
	if strategy.PageCount != 13 || strategy.VisualHint != "modern research workspace" || !strategy.UseVisualAssets {
		t.Fatalf("recommendation did not reuse LLM result: %#v", strategy)
	}
}

func TestBuildTemplateRecommendationReturnsExplainableStrategy(t *testing.T) {
	server := testTemplateServer(t)
	useVisualAssets := true
	intent := &agentintent.ClassificationResult{
		IntentReasoning:    "技术分享需要先讲概念，再讲架构与案例",
		Domain:             agentintent.DomainTechnical,
		SuggestedTemplates: []string{"tech-sharing"},
		SuggestedTheme:     "ocean_soft",
		SuggestedPageCount: 8,
		UseVisualAssets:    &useVisualAssets,
	}

	rec, err := server.buildTemplateRecommendation("生成8页关于AI Agent工具调用机制的技术分享", intent)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Strategy.Template == "" || rec.PrimaryTemplate.Name == "" || len(rec.RankedTemplates) == 0 {
		t.Fatalf("recommendation missing template details: %#v", rec)
	}
	if rec.Strategy.PageCount != 8 {
		t.Fatalf("page count = %d, want 8", rec.Strategy.PageCount)
	}
	if rec.Theme == nil {
		t.Fatal("recommendation should include theme detail")
	}
	if len(rec.ComponentFocus) == 0 || rec.VisualPolicy == "" {
		t.Fatalf("recommendation missing explainable planning fields: %#v", rec)
	}
}

func TestRecommendTemplateRouteDoesNotFallThroughToFrontend(t *testing.T) {
	projectRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(&ServerConfig{
		BaseDir:     t.TempDir(),
		SkillsDir:   filepath.Join(projectRoot, "skills"),
		FrontendDir: filepath.Join(projectRoot, "frontend", "dist"),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/templates/recommend", strings.NewReader(`{"query":"生成3页AI产品方案汇报"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "<!DOCTYPE html>") {
		t.Fatalf("recommend route fell through to frontend: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "primary_template") {
		t.Fatalf("recommend route returned unexpected payload: %s", rec.Body.String())
	}
}

func TestRecommendedTemplateDoesNotUseLegacyBackgroundPalette(t *testing.T) {
	server := testTemplateServer(t)
	useVisualAssets := true
	intent := &agentintent.ClassificationResult{
		Intent:             agentintent.IntentCreate,
		Domain:             agentintent.DomainGovernment,
		SuggestedTemplates: []string{"current-affairs"},
		SuggestedTheme:     "patriotic_blue",
		VisualHint:         "formal government meeting",
		SuggestedPageCount: 10,
		UseVisualAssets:    &useVisualAssets,
	}
	outline, strategy, err := server.resolveTemplateSelectionWithIntent("介绍延安", &TemplateSelection{Mode: "recommended"}, intent)
	if err != nil {
		t.Fatal(err)
	}
	if strategy.Theme != "patriotic_blue" || outline.Theme != "patriotic_blue" {
		t.Fatalf("theme = %q outline=%q, want patriotic_blue", strategy.Theme, outline.Theme)
	}
	if strategy.VisualHint != "formal government meeting" || !strategy.UseVisualAssets {
		t.Fatalf("visual policy should be carried without legacy background: strategy=%#v outline=%#v", strategy, outline)
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
	useVisualAssets := true
	intent := &agentintent.ClassificationResult{
		SuggestedTemplates: []string{"course-module"}, SuggestedTheme: "sage_calm",
		VisualHint: "traditional Chinese culture detail", SuggestedPageCount: 9, UseVisualAssets: &useVisualAssets,
	}
	outline, strategy, err := server.resolveTemplateSelectionWithIntent("中国风传统文化主题分享", &TemplateSelection{Mode: "recommended"}, intent)
	if err != nil {
		t.Fatal(err)
	}
	if strategy.VisualHint == "" || !strategy.UseVisualAssets {
		t.Fatalf("strategy = %#v, want visual search policy", strategy)
	}
	if outline.ContentMode != deck.OutlineContentModeRecommendedStyle || len(outline.Slides) != 0 {
		t.Fatalf("recommended outline copied preset slides: %#v", outline)
	}
	if outline.SuggestedPageCount != 9 || outline.Template != "course-module" || outline.Theme != "sage_calm" {
		t.Fatalf("recommended style guidance = %#v", outline)
	}
}

func TestRecommendationUsesExplicitUserPageCountBeforeLLMSuggestion(t *testing.T) {
	server := testTemplateServer(t)
	useVisualAssets := true
	intent := &agentintent.ClassificationResult{
		SuggestedTemplates: []string{"generic"}, SuggestedTheme: "ocean_soft",
		SuggestedPageCount: 5, UseVisualAssets: &useVisualAssets,
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

func TestPresetSelectionDoesNotAssignLegacyBackgrounds(t *testing.T) {
	server := testTemplateServer(t)
	outline, _, err := server.resolveTemplateSelection("说明一些尚未分类的新想法", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if len(outline.Slides) == 0 {
		t.Fatal("preset outline should keep template slides")
	}
}

func TestPresetSelectionIgnoresLegacyBackgroundThemeMatches(t *testing.T) {
	server := testTemplateServer(t)
	outline, strategy, err := server.resolveTemplateSelection("党建政府工作汇报", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if strategy.VisualHint != "" || strategy.UseVisualAssets {
		t.Fatalf("strategy should not use legacy background: %#v", strategy)
	}
	if len(outline.Slides) == 0 {
		t.Fatal("preset outline should keep template slides")
	}
}

func TestPresetSelectionUsesTemplateWithoutDefaultBackground(t *testing.T) {
	server := testTemplateServer(t)
	outline, strategy, err := server.resolveTemplateSelection("说明一些尚未分类的新想法", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if strategy.VisualHint != "" || strategy.UseVisualAssets {
		t.Fatalf("strategy = %#v, want no legacy background", strategy)
	}
	if len(outline.Slides) == 0 {
		t.Fatal("expected preset slides")
	}
}

func TestPresetSelectionCanSuppressBackground(t *testing.T) {
	server := testTemplateServer(t)
	outline, strategy, err := server.resolveTemplateSelection("说明一些尚未分类的新想法，不要背景", &TemplateSelection{Mode: "preset", Template: "generic"})
	if err != nil {
		t.Fatal(err)
	}
	if strategy.UseVisualAssets {
		t.Fatalf("strategy = %#v, want visual assets suppressed", strategy)
	}
	if len(outline.Slides) == 0 {
		t.Fatal("preset outline should keep template slides")
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
