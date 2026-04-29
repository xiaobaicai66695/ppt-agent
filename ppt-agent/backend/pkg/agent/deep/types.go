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

package deep

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// PPTTaskConfig PPT 任务配置
type PPTTaskConfig struct {
	WorkDir     string
	TaskID      string
	Concurrency int
	Operator    commandline.Operator
	QAModelFn   func(ctx context.Context) (model.ToolCallingChatModel, error)
	Skills      string
}

// TasksManifest PPT 任务清单
type TasksManifest struct {
	Title string      `json:"title"`
	Theme string      `json:"theme"`
	Tasks []*TaskItem `json:"tasks"`
}

type TaskItem struct {
	TaskID      string `json:"task_id"`
	PageIndex   int    `json:"page_index"`
	Title       string `json:"title"`
	ContentType string `json:"content_type"`
	Description string `json:"description"`
	OutputFile  string `json:"output_file"`
	Status      string `json:"status"`
	QAReport    string `json:"qa_report,omitempty"`
	FixAttempts int    `json:"fix_attempts,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func (m *TasksManifest) MustMarshalJSON() string {
	data, _ := json.Marshal(m)
	return string(data)
}

func (m *TasksManifest) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *TasksManifest) CompletedCount() int {
	count := 0
	for _, t := range m.Tasks {
		if t.Status == "done" || t.Status == "qa_done" || t.Status == "fixed" {
			count++
		}
	}
	return count
}

func (m *TasksManifest) AllDone() bool {
	return m.CompletedCount() == len(m.Tasks) && len(m.Tasks) > 0
}

func (m *TasksManifest) NeedsFix() []*TaskItem {
	var result []*TaskItem
	for _, t := range m.Tasks {
		if t.Status == "qa_done" && t.QAReport != "" {
			result = append(result, t)
		}
	}
	return result
}

func (m *TasksManifest) PendingTasks() []*TaskItem {
	var result []*TaskItem
	for _, t := range m.Tasks {
		if t.Status == "pending" {
			result = append(result, t)
		}
	}
	return result
}

func (m *TasksManifest) DoneTasks() []*TaskItem {
	var result []*TaskItem
	for _, t := range m.Tasks {
		if t.Status == "done" {
			result = append(result, t)
		}
	}
	return result
}

func (m *TasksManifest) GetTask(taskID string) *TaskItem {
	for _, t := range m.Tasks {
		if t.TaskID == taskID {
			return t
		}
	}
	return nil
}

func WriteTasksManifest(workDir string, manifest *TasksManifest) error {
	data, err := os.ReadFile(filepath.Join(workDir, "tasks.json"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(data) > 0 {
		existing := &TasksManifest{}
		if err := existing.UnmarshalJSON(data); err == nil {
			for i, t := range manifest.Tasks {
				for _, et := range existing.Tasks {
					if t.TaskID == et.TaskID {
						manifest.Tasks[i].Status = et.Status
						manifest.Tasks[i].QAReport = et.QAReport
						manifest.Tasks[i].FixAttempts = et.FixAttempts
						break
					}
				}
			}
		}
	}

	filePath := filepath.Join(workDir, "tasks.json")
	return os.WriteFile(filePath, []byte(manifest.MustMarshalJSON()), 0644)
}

func ReadTasksManifest(workDir string) (*TasksManifest, error) {
	data, err := os.ReadFile(filepath.Join(workDir, "tasks.json"))
	if err != nil {
		return nil, err
	}
	m := &TasksManifest{}
	if err := m.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return m, nil
}

type PPTTaskStart struct {
	Runner       *adk.Runner
	Iter         *adk.AsyncIterator[*adk.AgentEvent]
	CheckpointID string
	StartTime   time.Time
}

type PPTTaskResult struct {
	Message     string
	TotalSlides int
	DoneSlides  int
	Files       []string
	Duration    time.Duration
}

// TaskStatus 常量
const (
	StatusPending    = "pending"
	StatusGenerating = "generating"
	StatusDone       = "done"
	StatusQADone     = "qa_done"
	StatusFixed      = "fixed"
)
