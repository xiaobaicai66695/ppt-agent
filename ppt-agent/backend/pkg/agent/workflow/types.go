/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// tasksManifestMu protects tasks.json concurrent reads/writes.
var tasksManifestMu sync.Mutex

// TasksManifest represents the PPT task manifest.
type TasksManifest struct {
	Title    string      `json:"title"`
	Theme    string      `json:"theme"`
	Template string      `json:"template,omitempty"`
	Tasks    []*TaskItem `json:"tasks"`
}

// TaskItem represents a single slide task.
type TaskItem struct {
	TaskID      string       `json:"task_id"`
	PageIndex   int          `json:"page_index"`
	Title       string       `json:"title"`
	ContentType string       `json:"content_type"`
	Description string       `json:"description"`
	OutputFile  string       `json:"output_file"`
	Status      string       `json:"status"`
	QAReport    string       `json:"qa_report,omitempty"`
	FixAttempts int          `json:"fix_attempts,omitempty"`
	CreatedAt   string       `json:"created_at,omitempty"`
	Background  string       `json:"background,omitempty"`
}

// TaskOutline represents the AI-generated slide outline.
type TaskOutline struct {
	Template string         `json:"template,omitempty"`
	Theme   string         `json:"theme,omitempty"`
	Title   string         `json:"title"`
	Slides  []SlideOutline `json:"slides"`
}

// SlideOutline represents a single slide in the outline.
type SlideOutline struct {
	Title       string `json:"title"`
	ContentType string `json:"content_type"`
	Description string `json:"description"`
	Background  string `json:"background,omitempty"`
}

// TaskStatus constants.
const (
	StatusPending    = "pending"
	StatusGenerating  = "generating"
	StatusDone       = "done"
	StatusQADone     = "qa_done"
	StatusFixed      = "fixed"
)

// MustMarshalJSON serializes the manifest to JSON.
func (m *TasksManifest) MustMarshalJSON() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("tasks manifest marshal: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements custom JSON unmarshaling.
func (m *TasksManifest) UnmarshalJSON(data []byte) error {
	type alias TasksManifest
	return json.Unmarshal(data, (*alias)(m))
}

// CompletedCount returns the number of completed tasks.
func (m *TasksManifest) CompletedCount() int {
	count := 0
	for _, t := range m.Tasks {
		if t.Status == StatusDone || t.Status == StatusQADone || t.Status == StatusFixed {
			count++
		}
	}
	return count
}

// AllDone returns true if all tasks are completed.
func (m *TasksManifest) AllDone() bool {
	return m.CompletedCount() == len(m.Tasks) && len(m.Tasks) > 0
}

// NeedsFix returns tasks that have QA issues.
func (m *TasksManifest) NeedsFix() []*TaskItem {
	var result []*TaskItem
	for _, t := range m.Tasks {
		if t.Status == StatusQADone && t.QAReport != "" {
			result = append(result, t)
		}
	}
	return result
}

// PendingTasks returns tasks that are pending.
func (m *TasksManifest) PendingTasks() []*TaskItem {
	var result []*TaskItem
	for _, t := range m.Tasks {
		if t.Status == StatusPending {
			result = append(result, t)
		}
	}
	return result
}

// GetTask returns a task by ID.
func (m *TasksManifest) GetTask(taskID string) *TaskItem {
	for _, t := range m.Tasks {
		if t.TaskID == taskID {
			return t
		}
	}
	return nil
}

// WriteTasksManifest atomically writes the manifest to tasks.json.
func WriteTasksManifest(workDir string, manifest *TasksManifest) error {
	tasksManifestMu.Lock()
	defer tasksManifestMu.Unlock()

	filePath := filepath.Join(workDir, "tasks.json")
	tmpPath := filePath + ".tmp"

	content, err := manifest.MustMarshalJSON()
	if err != nil {
		return err
	}

	// Validate.
	if err := json.Unmarshal(content, &TasksManifest{}); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

// ReadTasksManifest reads the manifest from tasks.json.
func ReadTasksManifest(workDir string) (*TasksManifest, error) {
	data, err := os.ReadFile(filepath.Join(workDir, "tasks.json"))
	if err != nil {
		return nil, err
	}
	m := &TasksManifest{}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	return m, nil
}
