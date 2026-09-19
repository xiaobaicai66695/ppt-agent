package ppt

import (
	"fmt"
)

// PageCapacityDiagnostic records the capacity facts used by the Planner and
// deterministic review for one page.
type PageCapacityDiagnostic struct {
	PageIndex          int    `json:"page_index"`
	ContentType        string `json:"content_type"`
	ActualComponents   int    `json:"actual_components"`
	RecommendedMin     int    `json:"recommended_min"`
	RecommendedMax     int    `json:"recommended_max"`
	MaxComponents      int    `json:"max_components"`
	AboveRecommended   bool   `json:"above_recommended,omitempty"`
	OverLimit          bool   `json:"over_limit,omitempty"`
	OverflowComponents int    `json:"overflow_components,omitempty"`
}

// ManifestCapacityReport is benchmark-friendly evidence that the exact
// runtime contract was applied to every generated page.
type ManifestCapacityReport struct {
	ContractVersion       string                   `json:"contract_version"`
	ContractSHA256        string                   `json:"contract_sha256,omitempty"`
	ContractSource        string                   `json:"contract_source,omitempty"`
	Pages                 []PageCapacityDiagnostic `json:"pages,omitempty"`
	AboveRecommendedPages []int                    `json:"above_recommended_pages,omitempty"`
	OverLimitPages        []int                    `json:"over_limit_pages,omitempty"`
}

// InspectManifestCapacity evaluates a manifest with the same contract loader
// used by the production Planner. It does not mutate the manifest.
func InspectManifestCapacity(manifest *TasksManifest, skillsDir string) *ManifestCapacityReport {
	contract := contractForSkills(skillsDir)
	report := &ManifestCapacityReport{
		ContractVersion: fmt.Sprintf("%d", contract.Version),
		ContractSHA256:  contract.SHA256,
		ContractSource:  contract.Source,
	}
	if manifest == nil {
		return report
	}
	for _, task := range manifest.Tasks {
		if task == nil {
			continue
		}
		actual := 0
		if task.ContentPlan != nil {
			actual = len(task.ContentPlan.Components)
		}
		capacity := contract.CapacityFor(task.ContentType)
		diagnostic := PageCapacityDiagnostic{
			PageIndex:        task.PageIndex,
			ContentType:      task.ContentType,
			ActualComponents: actual,
			RecommendedMin:   capacity.RecommendedMin,
			RecommendedMax:   capacity.RecommendedMax,
			MaxComponents:    capacity.MaxComponents,
		}
		if actual > capacity.RecommendedMax {
			diagnostic.AboveRecommended = true
			report.AboveRecommendedPages = append(report.AboveRecommendedPages, task.PageIndex)
		}
		if actual > capacity.MaxComponents {
			diagnostic.OverLimit = true
			diagnostic.OverflowComponents = actual - capacity.MaxComponents
			report.OverLimitPages = append(report.OverLimitPages, task.PageIndex)
		}
		report.Pages = append(report.Pages, diagnostic)
	}
	return report
}
