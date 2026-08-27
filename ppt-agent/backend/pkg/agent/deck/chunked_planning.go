package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
)

const defaultChunkedPlanningThreshold = 8
const defaultSectionShardPageLimit = 4

type deckPlanningBlueprint struct {
	Title       string              `json:"title"`
	Theme       string              `json:"theme,omitempty"`
	Template    string              `json:"template,omitempty"`
	ContentBank map[string]any      `json:"content_bank,omitempty"`
	Sections    []DeckSection       `json:"sections,omitempty"`
	Pages       []deckBlueprintPage `json:"pages"`
}

type deckBlueprintPage struct {
	PageIndex        int          `json:"page_index"`
	SectionID        string       `json:"section_id,omitempty"`
	SectionTitle     string       `json:"section_title,omitempty"`
	Title            string       `json:"title"`
	ContentType      string       `json:"content_type"`
	LayoutVariant    string       `json:"layout_variant,omitempty"`
	PageIntent       string       `json:"page_intent,omitempty"`
	EvidenceRefs     []string     `json:"evidence_refs,omitempty"`
	DraftDescription string       `json:"draft_description,omitempty"`
	DraftContentPlan *ContentPlan `json:"draft_content_plan,omitempty"`
}

type sectionPlanShard struct {
	SectionID    string              `json:"section_id"`
	SectionTitle string              `json:"section_title,omitempty"`
	StartPage    int                 `json:"start_page"`
	EndPage      int                 `json:"end_page"`
	Tasks        []manifestTaskPatch `json:"tasks"`
}

type sectionPlanningJob struct {
	SectionID    string
	SectionTitle string
	StartPage    int
	EndPage      int
	Pages        []deckBlueprintPage
	Previous     string
	Next         string
}

func shouldUseChunkedPlanning(cfg *PPTTaskConfig) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PLANNER_CHUNKED_DISABLED"))) {
	case "1", "true", "yes", "on":
		return false
	}
	if cfg != nil && cfg.Outline != nil && len(cfg.Outline.Slides) > 0 {
		return true
	}
	threshold := agentutils.EnvInt("PLANNER_CHUNKED_PAGE_THRESHOLD", defaultChunkedPlanningThreshold)
	if threshold <= 0 {
		return true
	}
	suggested := 0
	if cfg != nil && cfg.IntentResult != nil {
		suggested = cfg.IntentResult.SuggestedPageCount
	}
	if suggested <= 0 {
		return true
	}
	return suggested > threshold
}

func runChunkedDeckPlanning(ctx context.Context, cfg *PPTTaskConfig, userQuery string, onEvent AgentEventCallback) (*TasksManifest, error) {
	if onEvent != nil {
		onEvent(AgentEvent{Type: AgentEventProgress, Phase: "planning", PhaseDetail: "正在建立整套演示蓝图"})
	}
	blueprint, err := generateDeckBlueprint(ctx, cfg, userQuery)
	if err != nil {
		return nil, err
	}
	normalizeDeckBlueprint(blueprint, cfg, userQuery)
	if len(blueprint.Pages) == 0 {
		return nil, fmt.Errorf("chunked blueprint produced no pages")
	}
	if onEvent != nil {
		onEvent(AgentEvent{Type: AgentEventProgress, Phase: "planning", PhaseDetail: fmt.Sprintf("蓝图已确定 %d 页，TaskExpander 正在按章节并行扩充页面内容", len(blueprint.Pages))})
	}
	shards, err := generateSectionPlanShards(ctx, cfg, userQuery, blueprint, onEvent)
	if err != nil {
		return nil, err
	}
	manifest, err := MergeSectionPlanShards(blueprint, shards)
	if err != nil {
		return nil, err
	}
	normalizeManifestLayoutVariants(manifest)
	normalizePlannerInitialManifest(manifest)
	if issues := plannerPreflightIssues(manifest); hasBlockingPlanReviewIssue(issues) {
		return nil, fmt.Errorf("chunked draft did not pass planner preflight: %d issues", len(issues))
	}
	if err := WriteTasksDraftManifest(cfg.WorkDir, manifest); err != nil {
		return nil, err
	}
	if cfg.RuntimeMeta != nil {
		cfg.RuntimeMeta.FreezePlan(runtimePlanSlidesFromTasks(manifest.Tasks))
		cfg.RuntimeMeta.RecordEvent("deck_spec_chunked_planned", "chunked_planner", "ok", fmt.Sprintf("merged %d slides from %d shards", len(manifest.Tasks), len(shards)), map[string]any{
			"slide_count": len(manifest.Tasks),
			"shard_count": len(shards),
			"template":    manifest.Template,
			"theme":       manifest.Theme,
		})
	}
	if onEvent != nil {
		onEvent(AgentEvent{Type: AgentEventAnswer, Content: fmt.Sprintf("已按章节分片完成 DeckSpec 草稿，共 %d 页，接下来进入一次统一 Reviewer。\n", len(manifest.Tasks))})
	}
	return manifest, nil
}

func generateDeckBlueprint(ctx context.Context, cfg *PPTTaskConfig, userQuery string) (*deckPlanningBlueprint, error) {
	if cfg != nil && cfg.Outline != nil && len(cfg.Outline.Slides) > 0 {
		return deckBlueprintFromOutline(cfg.Outline, userQuery), nil
	}
	model, err := newPlanningChatModel(ctx, cfg, 8192)
	if err != nil {
		return nil, err
	}
	pageHint := 12
	if cfg != nil && cfg.IntentResult != nil && cfg.IntentResult.SuggestedPageCount > 0 {
		pageHint = cfg.IntentResult.SuggestedPageCount
	}
	prompt := buildDeckBlueprintPrompt(cfg, userQuery, pageHint)
	resp, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("你是 Deck Blueprint Planner。只输出 JSON 对象，不输出 Markdown、解释或代码块。"),
		schema.UserMessage(prompt),
	})
	if err != nil {
		return nil, fmt.Errorf("blueprint planner failed: %w", err)
	}
	var blueprint deckPlanningBlueprint
	if err := decodeJSONObject(resp.Content, &blueprint); err != nil {
		return nil, fmt.Errorf("parse blueprint JSON: %w", err)
	}
	return &blueprint, nil
}

func deckBlueprintFromOutline(outline *TaskOutline, userQuery string) *deckPlanningBlueprint {
	if outline == nil {
		return &deckPlanningBlueprint{}
	}
	pages := make([]deckBlueprintPage, 0, len(outline.Slides))
	for i, slide := range outline.Slides {
		pageIndex := i + 1
		title := strings.TrimSpace(slide.Title)
		if title == "" {
			title = fmt.Sprintf("第 %d 页", pageIndex)
		}
		contentType := strings.TrimSpace(slide.ContentType)
		if contentType == "" {
			contentType = "content_slide"
		}
		description := strings.TrimSpace(slide.Description)
		pages = append(pages, deckBlueprintPage{
			PageIndex:        pageIndex,
			Title:            title,
			ContentType:      contentType,
			LayoutVariant:    strings.TrimSpace(slide.LayoutVariant),
			PageIntent:       firstNonEmptyString(description, title),
			DraftDescription: description,
			DraftContentPlan: cloneContentPlan(slide.ContentPlan),
		})
	}
	sections := inferBlueprintSectionsFromPageTypes(pages)
	sectionsByID := map[string]DeckSection{}
	for _, section := range sections {
		sectionsByID[section.ID] = section
	}
	for i := range pages {
		page := &pages[i]
		if page.SectionID != "" {
			continue
		}
		for _, section := range sections {
			if page.PageIndex >= section.StartPage && page.PageIndex <= section.EndPage {
				page.SectionID = section.ID
				page.SectionTitle = section.Title
				break
			}
		}
		if section, ok := sectionsByID[page.SectionID]; ok && strings.TrimSpace(page.SectionTitle) == "" {
			page.SectionTitle = section.Title
		}
	}
	return &deckPlanningBlueprint{
		Title:    compactManifestTitle(firstNonEmptyString(outline.Title, userQuery)),
		Theme:    firstNonEmptyString(outline.Theme, defaultManifestTheme),
		Template: firstNonEmptyString(outline.Template, defaultManifestTemplate),
		Sections: sections,
		Pages:    pages,
	}
}

func cloneContentPlan(plan *ContentPlan) *ContentPlan {
	if plan == nil {
		return nil
	}
	copied := *plan
	if plan.EvidenceRefs != nil {
		copied.EvidenceRefs = append([]string(nil), plan.EvidenceRefs...)
	}
	if plan.Components != nil {
		copied.Components = append([]PlanComponent(nil), plan.Components...)
	}
	if plan.VisualIntent != nil {
		visual := *plan.VisualIntent
		copied.VisualIntent = &visual
	}
	if plan.CapacityHint != nil {
		capacity := *plan.CapacityHint
		copied.CapacityHint = &capacity
	}
	if plan.ReviewerStatus != nil {
		status := *plan.ReviewerStatus
		if plan.ReviewerStatus.Issues != nil {
			status.Issues = append([]PlanReviewIssue(nil), plan.ReviewerStatus.Issues...)
		}
		copied.ReviewerStatus = &status
	}
	return &copied
}

func inferBlueprintSectionsFromPageTypes(pages []deckBlueprintPage) []DeckSection {
	if len(pages) == 0 {
		return nil
	}
	type startMarker struct {
		index int
		title string
	}
	markers := []startMarker{}
	for _, page := range pages {
		if strings.TrimSpace(page.ContentType) != "section_divider" {
			continue
		}
		markers = append(markers, startMarker{index: page.PageIndex, title: firstNonEmptyString(page.Title, fmt.Sprintf("章节 %d", len(markers)+1))})
	}
	if len(markers) == 0 || markers[0].index > pages[0].PageIndex {
		markers = append([]startMarker{{index: pages[0].PageIndex, title: "开篇"}}, markers...)
	}
	sections := make([]DeckSection, 0, len(markers))
	for i, marker := range markers {
		endPage := pages[len(pages)-1].PageIndex
		if i+1 < len(markers) {
			endPage = markers[i+1].index - 1
		}
		id := fmt.Sprintf("section-%02d", i+1)
		sections = append(sections, DeckSection{
			ID:        id,
			Title:     marker.title,
			Summary:   fmt.Sprintf("围绕“%s”扩充对应页面内容", marker.title),
			StartPage: marker.index,
			EndPage:   endPage,
			PageCount: endPage - marker.index + 1,
		})
	}
	return sections
}

func buildDeckBlueprintPrompt(cfg *PPTTaskConfig, userQuery string, pageHint int) string {
	styleContext := ""
	if cfg != nil {
		styleContext = strings.TrimSpace(cfg.StyleContext)
	}
	return fmt.Sprintf(`根据用户主题创建 PPT 蓝图。只锁定全局结构，不写每页完整正文。

用户主题：
%s

风格/规模参考：
%s

目标页数：约 %d 页。输出 JSON：
{
  "title": "整套 PPT 标题",
  "theme": "ocean_soft",
  "template": "generic",
  "content_bank": {
    "entities": [{"id":"entity.xxx","name":"具体对象","note":"简短事实"}],
    "facts": [{"id":"fact.xxx","text":"具体事实、年份、人物、地点或数字"}],
    "themes": [{"id":"theme.xxx","claim":"可复述观点","evidence":"对应证据"}]
  },
  "sections": [{"id":"section-id","title":"章节标题","summary":"章节作用","start_page":1,"end_page":3}],
  "pages": [{"page_index":1,"section_id":"section-id","section_title":"章节标题","title":"页面标题","content_type":"title_slide","page_intent":"该页在章节中的作用","evidence_refs":["fact.xxx"]}]
}

要求：
- content_type 必须使用合法英文 ID，例如 title_slide、agenda、section_divider、content_slide、image_text、card_grid、two_column、timeline、stat_slide、summary_slide。
- page_index 必须从 1 连续递增。
- 章节范围必须覆盖所有页面。
- content_bank 要先给后续页面可引用的具体素材，避免只写抽象主题词。
- 每个非封面/目录/章节页至少引用 1-3 个 evidence_refs。`, userQuery, firstNonEmptyString(styleContext, "无额外风格参考"), pageHint)
}

func normalizeDeckBlueprint(blueprint *deckPlanningBlueprint, cfg *PPTTaskConfig, userQuery string) {
	if blueprint == nil {
		return
	}
	blueprint.Title = compactManifestTitle(firstNonEmptyString(blueprint.Title, userQuery))
	if strings.TrimSpace(blueprint.Theme) == "" {
		blueprint.Theme = defaultManifestTheme
	}
	if strings.TrimSpace(blueprint.Template) == "" {
		blueprint.Template = defaultManifestTemplate
	}
	sort.SliceStable(blueprint.Pages, func(i, j int) bool {
		return blueprint.Pages[i].PageIndex < blueprint.Pages[j].PageIndex
	})
	for i := range blueprint.Pages {
		page := &blueprint.Pages[i]
		if page.PageIndex <= 0 {
			page.PageIndex = i + 1
		}
		if strings.TrimSpace(page.Title) == "" {
			page.Title = fmt.Sprintf("第 %d 页", page.PageIndex)
		}
		if strings.TrimSpace(page.ContentType) == "" {
			page.ContentType = "content_slide"
		}
		if strings.TrimSpace(page.SectionID) == "" {
			page.SectionID = fmt.Sprintf("section-%02d", max(1, (page.PageIndex-1)/defaultSectionShardPageLimit+1))
		}
		if strings.TrimSpace(page.SectionTitle) == "" {
			page.SectionTitle = page.SectionID
		}
	}
	if len(blueprint.Sections) == 0 {
		blueprint.Sections = inferBlueprintSections(blueprint.Pages)
	}
	sectionsByID := map[string]DeckSection{}
	for _, section := range blueprint.Sections {
		sectionsByID[section.ID] = section
	}
	for i := range blueprint.Pages {
		page := &blueprint.Pages[i]
		if section, ok := sectionsByID[page.SectionID]; ok && strings.TrimSpace(page.SectionTitle) == "" {
			page.SectionTitle = section.Title
		}
	}
}

func inferBlueprintSections(pages []deckBlueprintPage) []DeckSection {
	byID := map[string]*DeckSection{}
	var order []string
	for _, page := range pages {
		id := strings.TrimSpace(page.SectionID)
		if id == "" {
			id = fmt.Sprintf("section-%02d", max(1, (page.PageIndex-1)/defaultSectionShardPageLimit+1))
		}
		section := byID[id]
		if section == nil {
			title := firstNonEmptyString(page.SectionTitle, id)
			section = &DeckSection{ID: id, Title: title, StartPage: page.PageIndex, EndPage: page.PageIndex}
			byID[id] = section
			order = append(order, id)
		}
		if page.PageIndex < section.StartPage {
			section.StartPage = page.PageIndex
		}
		if page.PageIndex > section.EndPage {
			section.EndPage = page.PageIndex
		}
	}
	sections := make([]DeckSection, 0, len(order))
	for _, id := range order {
		section := *byID[id]
		section.PageCount = section.EndPage - section.StartPage + 1
		sections = append(sections, section)
	}
	return sections
}

func generateSectionPlanShards(ctx context.Context, cfg *PPTTaskConfig, userQuery string, blueprint *deckPlanningBlueprint, onEvent AgentEventCallback) ([]sectionPlanShard, error) {
	jobs := BuildSectionPlanningJobs(blueprint, agentutils.EnvInt("PLANNER_SECTION_SHARD_PAGE_LIMIT", defaultSectionShardPageLimit))
	if len(jobs) == 0 {
		return nil, fmt.Errorf("no section planning jobs")
	}
	concurrency := agentutils.EnvInt("PLANNER_SECTION_CONCURRENCY", getConcurrency(nil))
	if cfg != nil && cfg.RoutingDecision != nil {
		concurrency = min(concurrency, getConcurrency(cfg.RoutingDecision))
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(jobs) {
		concurrency = len(jobs)
	}
	shards := make([]sectionPlanShard, len(jobs))
	errs := make(chan error, len(jobs))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for idx, job := range jobs {
		idx, job := idx, job
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()
			if onEvent != nil {
				onEvent(AgentEvent{Type: AgentEventProgress, Phase: "planning", PhaseDetail: fmt.Sprintf("TaskExpander 正在扩充章节 %s（第 %d-%d 页）", job.SectionTitle, job.StartPage, job.EndPage)})
			}
			shard, err := generateOneSectionShard(ctx, cfg, userQuery, blueprint, job)
			if err != nil {
				errs <- err
				return
			}
			shards[idx] = *shard
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return shards, nil
}

func BuildSectionPlanningJobs(blueprint *deckPlanningBlueprint, pageLimit int) []sectionPlanningJob {
	if blueprint == nil || len(blueprint.Pages) == 0 {
		return nil
	}
	if pageLimit <= 0 {
		pageLimit = defaultSectionShardPageLimit
	}
	sections := blueprint.Sections
	if len(sections) == 0 {
		sections = inferBlueprintSections(blueprint.Pages)
	}
	pagesBySection := map[string][]deckBlueprintPage{}
	for _, page := range blueprint.Pages {
		pagesBySection[page.SectionID] = append(pagesBySection[page.SectionID], page)
	}
	jobs := make([]sectionPlanningJob, 0, len(sections))
	for i, section := range sections {
		pages := pagesBySection[section.ID]
		if len(pages) == 0 {
			continue
		}
		sort.SliceStable(pages, func(i, j int) bool { return pages[i].PageIndex < pages[j].PageIndex })
		previous, next := "", ""
		if i > 0 {
			previous = sections[i-1].Title
		}
		if i+1 < len(sections) {
			next = sections[i+1].Title
		}
		for start := 0; start < len(pages); start += pageLimit {
			end := min(start+pageLimit, len(pages))
			part := pages[start:end]
			jobs = append(jobs, sectionPlanningJob{
				SectionID: section.ID, SectionTitle: section.Title,
				StartPage: part[0].PageIndex, EndPage: part[len(part)-1].PageIndex,
				Pages: append([]deckBlueprintPage(nil), part...), Previous: previous, Next: next,
			})
		}
	}
	return jobs
}

func generateOneSectionShard(ctx context.Context, cfg *PPTTaskConfig, userQuery string, blueprint *deckPlanningBlueprint, job sectionPlanningJob) (*sectionPlanShard, error) {
	model, err := newPlanningChatModel(ctx, cfg, 12288)
	if err != nil {
		return nil, err
	}
	prompt := buildSectionPlannerPrompt(userQuery, blueprint, job)
	resp, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("你是 Section Task Expander。你的职责是把蓝图页扩写成完整 task 内容；只输出 JSON 对象，不输出 Markdown、解释或代码块。"),
		schema.UserMessage(prompt),
	})
	if err != nil {
		return nil, fmt.Errorf("section planner %s %d-%d failed: %w", job.SectionID, job.StartPage, job.EndPage, err)
	}
	var shard sectionPlanShard
	if err := decodeJSONObject(resp.Content, &shard); err != nil {
		return nil, fmt.Errorf("parse section shard %s %d-%d JSON: %w", job.SectionID, job.StartPage, job.EndPage, err)
	}
	if shard.SectionID == "" {
		shard.SectionID = job.SectionID
	}
	if shard.SectionTitle == "" {
		shard.SectionTitle = job.SectionTitle
	}
	if shard.StartPage == 0 {
		shard.StartPage = job.StartPage
	}
	if shard.EndPage == 0 {
		shard.EndPage = job.EndPage
	}
	return &shard, nil
}

func buildSectionPlannerPrompt(userQuery string, blueprint *deckPlanningBlueprint, job sectionPlanningJob) string {
	deckContext, _ := json.Marshal(taskExpanderDeckContext(blueprint, job))
	sectionDraft, _ := json.Marshal(job.Pages)
	return fmt.Sprintf(`请以 TaskExpander 身份补全一个 PPT 章节的页面内容。不要改变页码、标题、content_type、章节顺序。

用户主题：
%s

整套蓝图摘要：
%s

本节草稿页面：
%s

相邻章节：上一节=%s；下一节=%s

输出 JSON：
{
  "section_id": "%s",
  "section_title": "%s",
  "start_page": %d,
  "end_page": %d,
  "tasks": [
    {
      "page_index": 4,
      "title": "沿用蓝图标题",
      "content_type": "沿用蓝图 content_type",
      "description": "80-160字，说明本页具体内容",
      "page_intent": "该页具体叙事任务",
      "evidence_refs": ["fact.xxx"],
      "content_plan": {
        "summary": "一句话核心判断",
        "slide_intent": "本页在整套 PPT 中的作用",
        "evidence_refs": ["fact.xxx"],
        "visual_intent": {"asset_purpose":"background","asset_query":"literature","asset_subject":"具体视觉主体","composition":"wide background"},
        "components": [{"id":"insight_1","type":"insight","title":"具体观点","body":"必须包含具体事实、对象、事件、文本细节或数字，避免空泛套话。"}]
      }
    }
  ]
}

内容要求：
- 只扩写“本节草稿页面”，不要重写整套 PPT 结构；草稿中的 draft_description/draft_content_plan 是输入素材，必须补成可执行正文后再输出。
- 每个非章节页至少写 2 个有信息密度的组件，不能只有概念词。
- fact_card/key_point/insight 的 body 目标 70-130 字，必须包含具体对象或证据。
- 严禁把模板脚手架、栏目名或大纲词当作正文；例如“左栏成就”“右栏短板”“卡片一”“要点一”“观点一”“内容一”“对比后的判断”“未来展望洞察”“核心观点”“补充说明”等都不是可上屏 body/items，必须改写为包含事实、时间、数字、影响或结论的完整表达。
- image_text 页必须包含 image 组件，asset_purpose 为 scene 或 evidence；背景仍写 visual_intent。
- 不写 task_id、output_file、status、created_at、capacity_hint.component_count，这些由系统合并。`, userQuery, deckContext, sectionDraft,
		firstNonEmptyString(job.Previous, "无"), firstNonEmptyString(job.Next, "无"),
		job.SectionID, job.SectionTitle, job.StartPage, job.EndPage)
}

func taskExpanderDeckContext(blueprint *deckPlanningBlueprint, job sectionPlanningJob) map[string]any {
	context := map[string]any{
		"title":            "",
		"theme":            defaultManifestTheme,
		"template":         defaultManifestTemplate,
		"sections":         []DeckSection{},
		"previous_section": firstNonEmptyString(job.Previous, "无"),
		"next_section":     firstNonEmptyString(job.Next, "无"),
	}
	if blueprint == nil {
		return context
	}
	context["title"] = blueprint.Title
	context["theme"] = firstNonEmptyString(blueprint.Theme, defaultManifestTheme)
	context["template"] = firstNonEmptyString(blueprint.Template, defaultManifestTemplate)
	context["sections"] = blueprint.Sections
	if len(blueprint.ContentBank) > 0 {
		context["content_bank"] = blueprint.ContentBank
	}
	return context
}

func MergeSectionPlanShards(blueprint *deckPlanningBlueprint, shards []sectionPlanShard) (*TasksManifest, error) {
	if blueprint == nil || len(blueprint.Pages) == 0 {
		return nil, fmt.Errorf("blueprint pages are required")
	}
	shardTasks := map[int]manifestTaskPatch{}
	for _, shard := range shards {
		for _, task := range shard.Tasks {
			if task.PageIndex == nil || *task.PageIndex <= 0 {
				return nil, fmt.Errorf("section shard %s contains task without page_index", shard.SectionID)
			}
			page := *task.PageIndex
			if _, exists := shardTasks[page]; exists {
				return nil, fmt.Errorf("duplicate page_index %d in section shards", page)
			}
			shardTasks[page] = task
		}
	}
	manifest := &TasksManifest{
		Title:       compactManifestTitle(blueprint.Title),
		Theme:       firstNonEmptyString(blueprint.Theme, defaultManifestTheme),
		Template:    firstNonEmptyString(blueprint.Template, defaultManifestTemplate),
		ContentBank: blueprint.ContentBank,
		Sections:    append([]DeckSection(nil), blueprint.Sections...),
	}
	seenPages := map[int]bool{}
	for _, page := range blueprint.Pages {
		if page.PageIndex <= 0 {
			return nil, fmt.Errorf("blueprint contains invalid page_index %d", page.PageIndex)
		}
		if seenPages[page.PageIndex] {
			return nil, fmt.Errorf("duplicate page_index %d in blueprint", page.PageIndex)
		}
		seenPages[page.PageIndex] = true
		patch, ok := shardTasks[page.PageIndex]
		if !ok {
			return nil, fmt.Errorf("missing section shard output for page_index %d", page.PageIndex)
		}
		item := taskItemFromBlueprintAndPatch(page, patch)
		manifest.Tasks = append(manifest.Tasks, item)
	}
	sort.SliceStable(manifest.Tasks, func(i, j int) bool {
		return manifest.Tasks[i].PageIndex < manifest.Tasks[j].PageIndex
	})
	normalizeMergedBackgroundQueries(manifest)
	return manifest, nil
}

func taskItemFromBlueprintAndPatch(page deckBlueprintPage, patch manifestTaskPatch) *TaskItem {
	title := firstPatchString(patch.Title, page.Title)
	contentType := firstPatchString(patch.ContentType, page.ContentType)
	description := firstPatchString(patch.Description, page.DraftDescription, page.PageIntent, page.Title)
	sectionID := firstPatchString(patch.SectionID, page.SectionID)
	sectionTitle := firstPatchString(patch.SectionTitle, page.SectionTitle)
	pageIntent := firstPatchString(patch.PageIntent, page.PageIntent)
	evidenceRefs := append([]string(nil), page.EvidenceRefs...)
	if len(patch.EvidenceRefs) > 0 {
		evidenceRefs = append([]string(nil), patch.EvidenceRefs...)
	}
	item := &TaskItem{
		TaskID:        fmt.Sprintf("slide-%d", page.PageIndex),
		PageIndex:     page.PageIndex,
		SectionID:     sectionID,
		SectionTitle:  sectionTitle,
		Title:         title,
		ContentType:   contentType,
		LayoutVariant: firstPatchString(patch.LayoutVariant, page.LayoutVariant),
		Description:   description,
		PageIntent:    pageIntent,
		EvidenceRefs:  evidenceRefs,
		OutputFile:    fmt.Sprintf("slide_%02d_%s.pptx", page.PageIndex, sanitizeTaskFilename(title)),
		Status:        StatusPending,
		ContentPlan:   firstNonNilContentPlan(patch.ContentPlan, page.DraftContentPlan),
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	if item.ContentPlan == nil {
		item.ContentPlan = &ContentPlan{}
	}
	if strings.TrimSpace(item.ContentPlan.Summary) == "" {
		item.ContentPlan.Summary = firstNonEmptyString(pageIntent, description)
	}
	if strings.TrimSpace(item.ContentPlan.SlideIntent) == "" {
		item.ContentPlan.SlideIntent = firstNonEmptyString(pageIntent, description)
	}
	if len(item.ContentPlan.EvidenceRefs) == 0 && len(evidenceRefs) > 0 {
		item.ContentPlan.EvidenceRefs = append([]string(nil), evidenceRefs...)
	}
	item.ContentPlan.CapacityHint = &CapacityHint{
		EstimatedDensity: firstNonEmptyString(item.ContentPlan.CapacityHintValue("density"), "normal"),
		OverflowRisk:     firstNonEmptyString(item.ContentPlan.CapacityHintValue("risk"), "low"),
		ComponentCount:   len(item.ContentPlan.Components),
	}
	return item
}

func firstNonNilContentPlan(values ...*ContentPlan) *ContentPlan {
	for _, value := range values {
		if value != nil {
			return cloneContentPlan(value)
		}
	}
	return nil
}

func (p *ContentPlan) CapacityHintValue(kind string) string {
	if p == nil || p.CapacityHint == nil {
		return ""
	}
	switch kind {
	case "density":
		return p.CapacityHint.EstimatedDensity
	case "risk":
		return p.CapacityHint.OverflowRisk
	default:
		return ""
	}
}

func firstPatchString(value *string, fallbacks ...string) string {
	if value != nil && strings.TrimSpace(*value) != "" {
		return strings.TrimSpace(*value)
	}
	return firstNonEmptyString(fallbacks...)
}

func sanitizeTaskFilename(title string) string {
	name := strings.TrimSpace(compactManifestTitle(title))
	if name == "" {
		return "slide"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_", "\n", "_", "\r", "_")
	name = replacer.Replace(name)
	runes := []rune(name)
	if len(runes) > 36 {
		name = string(runes[:36])
	}
	return strings.TrimSpace(name)
}

func normalizeMergedBackgroundQueries(manifest *TasksManifest) {
	if manifest == nil {
		return
	}
	queriesByType := map[string]string{}
	apply := func(contentType string, visual *VisualIntent) {
		if visual == nil || !isExternalBackgroundIntent(visual) {
			return
		}
		key := backgroundContentTypeKey(contentType)
		query := normalizeAssetQuery(firstNonEmptyString(visual.AssetSubject, visual.AssetQuery), true)
		if query == "" {
			return
		}
		if existing := queriesByType[key]; existing != "" {
			visual.AssetQuery = existing
			visual.AssetSubject = firstNonEmptyString(visual.AssetSubject, existing)
			return
		}
		queriesByType[key] = query
		visual.AssetQuery = query
	}
	for _, item := range manifest.Tasks {
		if item == nil || item.ContentPlan == nil {
			continue
		}
		apply(item.ContentType, item.ContentPlan.VisualIntent)
	}
}

func runtimePlanSlidesFromTasks(items []*TaskItem) []agentutils.PlanSlide {
	slides := make([]agentutils.PlanSlide, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		slides = append(slides, agentutils.PlanSlide{
			PageIndex: item.PageIndex, TaskID: item.TaskID, Title: item.Title,
			ContentType: item.ContentType, OutputFile: filepath.Base(item.OutputFile), Status: item.Status,
		})
	}
	return slides
}

func decodeJSONObject(raw string, target any) error {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	if err := json.Unmarshal([]byte(raw), target); err == nil {
		return nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return fmt.Errorf("no JSON object found")
	}
	return json.Unmarshal([]byte(raw[start:end+1]), target)
}
