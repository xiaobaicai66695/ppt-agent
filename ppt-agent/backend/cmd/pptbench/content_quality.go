package main

import (
	"strings"
	"unicode"

	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
)

func assessContentQuality(manifest *deck.TasksManifest) *contentQualityReport {
	if manifest == nil {
		return nil
	}
	report := &contentQualityReport{}
	seenClaims := map[string][]int{}
	lastContentType := ""
	currentRun := 0
	longestRun := 0
	longestRunType := ""

	for _, task := range manifest.Tasks {
		if task == nil {
			continue
		}
		claim := extractPageClaim(task)
		report.PageClaims = append(report.PageClaims, contentPageClaim{
			PageIndex: task.PageIndex, Title: task.Title, ContentType: task.ContentType, Claim: claim,
		})
		if task.ContentType == "title_slide" && report.DeckClaim == "" {
			report.DeckClaim = claim
		}
		if requiresBenchmarkPageClaim(task.ContentType) && claim == "" {
			report.MissingClaimPages = append(report.MissingClaimPages, task.PageIndex)
		}
		if normalized := normalizeClaim(claim); normalized != "" {
			seenClaims[normalized] = append(seenClaims[normalized], task.PageIndex)
		}

		if isNarrativeContentType(task.ContentType) {
			if task.ContentType == lastContentType {
				currentRun++
			} else {
				lastContentType, currentRun = task.ContentType, 1
			}
			if currentRun > longestRun {
				longestRun, longestRunType = currentRun, task.ContentType
			}
		} else {
			lastContentType, currentRun = "", 0
		}
		assessAgendaSubtitles(task, report)
	}
	for _, pages := range seenClaims {
		if len(pages) > 1 {
			report.DuplicateClaimGroups = append(report.DuplicateClaimGroups, pages)
		}
	}
	report.LongestRepeatedLayoutRun = longestRun
	if longestRunType != "" && longestRun > 1 {
		report.RepeatedLayoutRunContentTypes = []string{longestRunType}
	}
	return report
}

func assessAgendaSubtitles(task *deck.TaskItem, report *contentQualityReport) {
	if task == nil || report == nil || strings.TrimSpace(task.ContentType) != "agenda" || task.ContentPlan == nil {
		return
	}
	for _, component := range task.ContentPlan.Components {
		if strings.TrimSpace(component.Type) != "toc_item" {
			continue
		}
		report.AgendaTOCItems++
		title := strings.TrimSpace(component.Title)
		body := strings.TrimSpace(component.Body)
		issue := agendaSubtitleIssue{PageIndex: task.PageIndex, ComponentID: component.ID, Title: title}
		switch {
		case title == "":
			issue.Code = "missing_toc_title"
		case body == "":
			issue.Code = "missing_toc_subtitle"
		case normalizeClaim(title) == normalizeClaim(body):
			issue.Code = "repeated_toc_title"
		default:
			report.AgendaTOCSubtitles++
			continue
		}
		report.AgendaSubtitleIssues = append(report.AgendaSubtitleIssues, issue)
	}
}

func extractPageClaim(task *deck.TaskItem) string {
	if task == nil || task.ContentPlan == nil {
		return ""
	}
	for _, preferredType := range []string{"argument_block", "insight", "recommendation", "key_point", "quote_block", "headline", "paragraph"} {
		for _, component := range task.ContentPlan.Components {
			if strings.TrimSpace(component.Type) != preferredType {
				continue
			}
			if text := firstNonEmpty(component.Body, component.Text, component.Title); text != "" {
				return text
			}
		}
	}
	return ""
}

func requiresBenchmarkPageClaim(contentType string) bool {
	switch strings.TrimSpace(contentType) {
	case "title_slide", "agenda", "section_divider":
		return false
	default:
		return true
	}
}

func isNarrativeContentType(contentType string) bool {
	switch strings.TrimSpace(contentType) {
	case "content_slide", "image_text", "card_grid", "three_column", "two_column", "timeline", "process_flow", "comparison_table", "case_study", "example_detail", "deep_dive", "swot_analysis", "kanban", "summary_slide":
		return true
	default:
		return false
	}
}

func normalizeClaim(value string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
