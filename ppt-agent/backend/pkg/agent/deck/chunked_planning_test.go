package deck

import (
	"strings"
	"testing"
)

func TestShouldUseChunkedPlanningKeepsOutlineOnTaskExpanderPath(t *testing.T) {
	t.Setenv("PLANNER_CHUNKED_DISABLED", "")
	cfg := &PPTTaskConfig{
		Outline: &TaskOutline{Slides: []SlideOutline{{Title: "用户大纲页", ContentType: "content_slide"}}},
	}

	if !shouldUseChunkedPlanning(cfg) {
		t.Fatal("outline tasks should still use chunked planning so TaskExpander, not monolithic Planner, fills page content")
	}
}

func TestDeckBlueprintFromOutlineCarriesSectionDraft(t *testing.T) {
	blueprint := deckBlueprintFromOutline(&TaskOutline{
		Title: "制造强国",
		Theme: "ocean_soft",
		Slides: []SlideOutline{
			{Title: "封面", ContentType: "title_slide", Description: "建立主题。"},
			{Title: "成就与短板", ContentType: "two_column", Description: "左栏成就，右栏短板。", ContentPlan: &ContentPlan{
				Components: []PlanComponent{{ID: "left", Type: "fact_card", Body: "左栏成就"}},
			}},
			{Title: "未来规划", ContentType: "section_divider", Description: "进入未来规划章节。"},
			{Title: "重点方向", ContentType: "content_slide", Description: "自主可控、智能制造、绿色低碳。"},
		},
	}, "制造强国汇报")

	if len(blueprint.Pages) != 4 {
		t.Fatalf("len(pages)=%d, want 4", len(blueprint.Pages))
	}
	if blueprint.Pages[1].DraftDescription == "" || blueprint.Pages[1].DraftContentPlan == nil {
		t.Fatalf("outline draft was not carried into blueprint page: %#v", blueprint.Pages[1])
	}
	if len(blueprint.Sections) != 2 {
		t.Fatalf("len(sections)=%d, want 2: %#v", len(blueprint.Sections), blueprint.Sections)
	}
	if blueprint.Pages[3].SectionTitle != "未来规划" {
		t.Fatalf("page 4 section title=%q, want future section", blueprint.Pages[3].SectionTitle)
	}
}

func TestTaskExpanderPromptUsesSectionDraftNotFullPageRewrite(t *testing.T) {
	blueprint := deckBlueprintFromOutline(&TaskOutline{
		Title: "制造强国",
		Slides: []SlideOutline{
			{Title: "成就与短板", ContentType: "two_column", Description: "左栏成就，右栏短板。"},
		},
	}, "制造强国汇报")
	jobs := BuildSectionPlanningJobs(blueprint, 4)
	if len(jobs) != 1 {
		t.Fatalf("len(jobs)=%d, want 1", len(jobs))
	}

	prompt := buildSectionPlannerPrompt("制造强国汇报", blueprint, jobs[0])

	if !strings.Contains(prompt, "本节草稿页面") || !strings.Contains(prompt, "draft_description") {
		t.Fatalf("TaskExpander prompt should carry section draft, got: %s", prompt)
	}
	if !strings.Contains(prompt, "只扩写“本节草稿页面”") {
		t.Fatalf("TaskExpander prompt should forbid rewriting whole deck, got: %s", prompt)
	}
}

func TestBuildSectionPlanningJobsSplitsLongSections(t *testing.T) {
	blueprint := &deckPlanningBlueprint{
		Sections: []DeckSection{{ID: "story", Title: "故事", StartPage: 1, EndPage: 5}},
		Pages: []deckBlueprintPage{
			{PageIndex: 1, SectionID: "story", SectionTitle: "故事", Title: "p1", ContentType: "title_slide"},
			{PageIndex: 2, SectionID: "story", SectionTitle: "故事", Title: "p2", ContentType: "content_slide"},
			{PageIndex: 3, SectionID: "story", SectionTitle: "故事", Title: "p3", ContentType: "content_slide"},
			{PageIndex: 4, SectionID: "story", SectionTitle: "故事", Title: "p4", ContentType: "content_slide"},
			{PageIndex: 5, SectionID: "story", SectionTitle: "故事", Title: "p5", ContentType: "summary_slide"},
		},
	}
	jobs := BuildSectionPlanningJobs(blueprint, 2)
	if len(jobs) != 3 {
		t.Fatalf("len(jobs) = %d, want 3", len(jobs))
	}
	if jobs[0].StartPage != 1 || jobs[0].EndPage != 2 || jobs[2].StartPage != 5 || jobs[2].EndPage != 5 {
		t.Fatalf("jobs were not split by page limit: %#v", jobs)
	}
}

func TestMergeSectionPlanShardsFillsDeterministicTaskFields(t *testing.T) {
	desc := "用具体事实说明页面内容"
	blueprint := &deckPlanningBlueprint{
		Title:    "百年孤独",
		Theme:    "ocean_soft",
		Template: "generic",
		Sections: []DeckSection{{ID: "author", Title: "作者与时代", StartPage: 1, EndPage: 2}},
		Pages: []deckBlueprintPage{
			{PageIndex: 1, SectionID: "author", SectionTitle: "作者与时代", Title: "作者与时代", ContentType: "section_divider", PageIntent: "进入作者背景章节"},
			{PageIndex: 2, SectionID: "author", SectionTitle: "作者与时代", Title: "加西亚·马尔克斯", ContentType: "content_slide", PageIntent: "说明作者背景", EvidenceRefs: []string{"fact.publish"}},
		},
	}
	shards := []sectionPlanShard{{
		SectionID: "author", StartPage: 1, EndPage: 2,
		Tasks: []manifestTaskPatch{
			{PageIndex: intPtr(1), Title: stringPtr("作者与时代"), ContentType: stringPtr("section_divider"), Description: &desc, ContentPlan: &ContentPlan{Components: []PlanComponent{{ID: "marker", Type: "section_marker", Text: "01"}}}},
			{PageIndex: intPtr(2), Description: &desc, ContentPlan: &ContentPlan{Components: []PlanComponent{{ID: "fact", Type: "fact_card", Title: "1967年出版", Body: "《百年孤独》于1967年出版，随后成为拉美文学爆炸的重要代表作品之一。"}}}},
		},
	}}
	manifest, err := MergeSectionPlanShards(blueprint, shards)
	if err != nil {
		t.Fatalf("MergeSectionPlanShards returned error: %v", err)
	}
	if len(manifest.Tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(manifest.Tasks))
	}
	got := manifest.Tasks[1]
	if got.TaskID != "slide-2" || got.OutputFile == "" || got.Status != StatusPending || got.SectionID != "author" {
		t.Fatalf("deterministic fields not filled: %#v", got)
	}
	if got.ContentPlan == nil || got.ContentPlan.CapacityHint == nil || got.ContentPlan.CapacityHint.ComponentCount != 1 {
		t.Fatalf("capacity hint not filled from components: %#v", got.ContentPlan)
	}
	if len(got.EvidenceRefs) != 1 || got.EvidenceRefs[0] != "fact.publish" {
		t.Fatalf("evidence refs not preserved: %#v", got.EvidenceRefs)
	}
}

func intPtr(value int) *int { return &value }

func stringPtr(value string) *string { return &value }
