package task

import (
	"fmt"
	"path"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
)

// CanonicalOutputFile converts absolute, relative, Windows, and POSIX paths to
// the public file name used by task APIs and download routes.
func CanonicalOutputFile(file string) string {
	file = strings.TrimSpace(strings.ReplaceAll(file, "\\", "/"))
	if file == "" || file == "." {
		return ""
	}
	return path.Base(path.Clean(file))
}

// StableSlideKey returns the identity shared by manifest progress and file
// events. Page index is the strongest contract; task id and basename are
// compatibility fallbacks for incomplete legacy records.
func StableSlideKey(pageIndex int, taskID, outputFile string) string {
	if pageIndex > 0 {
		return fmt.Sprintf("page:%d", pageIndex)
	}
	if taskID = strings.TrimSpace(taskID); taskID != "" {
		return "task:" + strings.ToLower(taskID)
	}
	if name := CanonicalOutputFile(outputFile); name != "" {
		return "file:" + strings.ToLower(name)
	}
	return ""
}

// DeduplicateOutputFiles returns canonical public file names in first-seen
// order. Empty and duplicate entries are omitted.
func DeduplicateOutputFiles(files []string) []string {
	result := make([]string, 0, len(files))
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		name := CanonicalOutputFile(file)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, name)
	}
	return result
}

// ManifestOutputFiles returns one canonical output file for each completed
// logical slide in page order.
func ManifestOutputFiles(manifest *deep.TasksManifest) []string {
	if manifest == nil {
		return nil
	}
	files := make([]string, 0, len(manifest.Tasks))
	seenSlides := make(map[string]struct{}, len(manifest.Tasks))
	for _, item := range manifest.Tasks {
		if item == nil || item.OutputFile == "" {
			continue
		}
		if item.Status != deep.StatusDone && item.Status != deep.StatusQADone && item.Status != deep.StatusFixed {
			continue
		}
		key := StableSlideKey(item.PageIndex, item.TaskID, item.OutputFile)
		if key != "" {
			if _, exists := seenSlides[key]; exists {
				continue
			}
			seenSlides[key] = struct{}{}
		}
		files = append(files, item.OutputFile)
	}
	return DeduplicateOutputFiles(files)
}
