package deck

import "testing"

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
