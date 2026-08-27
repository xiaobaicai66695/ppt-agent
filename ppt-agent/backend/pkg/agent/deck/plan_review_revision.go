package deck

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type planReviewRevisionPayload struct {
	Round         int               `json:"round"`
	Summary       string            `json:"summary"`
	Scope         planReviewScope   `json:"scope"`
	Issues        []PlanReviewIssue `json:"issues"`
	IncludedTasks []*TaskItem       `json:"included_tasks,omitempty"`
	Instructions  []string          `json:"instructions"`
}

type planReviewScope struct {
	PageIndexes       []int    `json:"page_indexes,omitempty"`
	SectionIDs        []string `json:"section_ids,omitempty"`
	AllowedTaskIDs    []string `json:"allowed_task_ids,omitempty"`
	IncludesDeckLevel bool     `json:"includes_deck_level,omitempty"`
	Reason            string   `json:"reason"`
}

func buildPlanReviewRevisionInput(workDir string, round int, report *PlanReviewReport) (string, []string, error) {
	if report == nil {
		return "", nil, fmt.Errorf("review report is nil")
	}
	manifest, err := ReadTasksDraftManifest(workDir)
	if err != nil {
		return "", nil, err
	}
	payload := buildPlanReviewRevisionPayload(manifest, round, report)
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", nil, err
	}
	input := fmt.Sprintf(`这是第 %d 轮 Task Reviewer 修正输入。

后端已经把完整 tasks.draft.json 按审查报告压缩为 item 切片。不要读取、复述或重写完整 tasks.draft.json；只基于 included_tasks 和 issues 修正 scope.allowed_task_ids 中的页面。若 issues 只有 deck 级字段问题，可以只 patch title/theme/template，不要附带无关页面。

调用 patch_tasks_draft 时只提交这些 task_id 的 patch，并把同一轮必要修正合并为一次工具调用：
%s`, round, string(data))
	return input, payload.Scope.AllowedTaskIDs, nil
}

func buildPlanReviewRevisionPayload(manifest *TasksManifest, round int, report *PlanReviewReport) planReviewRevisionPayload {
	issues := blockingPlanReviewIssues(report.Issues)
	if len(issues) == 0 {
		issues = append([]PlanReviewIssue(nil), report.Issues...)
	}

	tasksByPage := map[int]*TaskItem{}
	sectionByPage := map[int]string{}
	for _, task := range manifest.Tasks {
		if task == nil {
			continue
		}
		tasksByPage[task.PageIndex] = task
		if sectionID := strings.TrimSpace(task.SectionID); sectionID != "" {
			sectionByPage[task.PageIndex] = sectionID
		}
	}

	pageSet := map[int]bool{}
	sectionSet := map[string]bool{}
	deckLevel := false
	for _, issue := range issues {
		if issue.PageIndex <= 0 {
			deckLevel = true
			continue
		}
		if sectionID := sectionByPage[issue.PageIndex]; sectionID != "" {
			sectionSet[sectionID] = true
			continue
		}
		pageSet[issue.PageIndex] = true
	}

	var included []*TaskItem
	for _, task := range manifest.Tasks {
		if task == nil {
			continue
		}
		if sectionSet[strings.TrimSpace(task.SectionID)] || pageSet[task.PageIndex] {
			included = append(included, task)
		}
	}
	sort.SliceStable(included, func(i, j int) bool {
		return included[i].PageIndex < included[j].PageIndex
	})

	includedPages := map[int]bool{}
	allowedTaskIDs := make([]string, 0, len(included))
	for _, task := range included {
		includedPages[task.PageIndex] = true
		if taskID := strings.TrimSpace(task.TaskID); taskID != "" {
			allowedTaskIDs = append(allowedTaskIDs, taskID)
		}
	}
	sort.Strings(allowedTaskIDs)

	filteredIssues := make([]PlanReviewIssue, 0, len(report.Issues))
	for _, issue := range report.Issues {
		if issue.PageIndex <= 0 || includedPages[issue.PageIndex] {
			filteredIssues = append(filteredIssues, issue)
		}
	}

	sectionIDs := sortedStringKeys(sectionSet)
	pageIndexes := sortedIntKeys(includedPages)
	reason := "按失败页定位修正切片"
	if len(sectionIDs) > 0 {
		reason = "按失败页所属章节发送整节 item 切片"
	}
	if deckLevel && len(pageIndexes) == 0 {
		reason = "仅包含 deck 级字段问题，可只 patch 顶层字段"
	}

	return planReviewRevisionPayload{
		Round:   round,
		Summary: report.Summary,
		Scope: planReviewScope{
			PageIndexes:       pageIndexes,
			SectionIDs:        sectionIDs,
			AllowedTaskIDs:    allowedTaskIDs,
			IncludesDeckLevel: deckLevel,
			Reason:            reason,
		},
		Issues:        filteredIssues,
		IncludedTasks: included,
		Instructions: []string{
			"只修改 included_tasks 中列出的页面，禁止新增、删除、重排页面。",
			"patch_tasks_draft 必须使用 task_id 定位；不要修改 page_index、output_file、status 或 fix_attempts。",
			"不要输出完整 JSON；完成一次 patch 后用 1-3 句中文说明修改了哪些页和问题类别。",
		},
	}
}

func blockingPlanReviewIssues(issues []PlanReviewIssue) []PlanReviewIssue {
	blocking := make([]PlanReviewIssue, 0, len(issues))
	for _, issue := range issues {
		if strings.EqualFold(strings.TrimSpace(issue.Severity), "warning") {
			continue
		}
		blocking = append(blocking, issue)
	}
	return blocking
}

func sortedStringKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		if key = strings.TrimSpace(key); key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func sortedIntKeys(values map[int]bool) []int {
	keys := make([]int, 0, len(values))
	for key := range values {
		if key > 0 {
			keys = append(keys, key)
		}
	}
	sort.Ints(keys)
	return keys
}
