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

package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// SkillInfo represents a loaded skill with its name, description and content
type SkillInfo struct {
	Name        string
	Description string
	Content     string
}

// LoadSkillsFromDir loads all skills from a directory
// Each skill is a subdirectory containing a SKILL.md file
func LoadSkillsFromDir(ctx context.Context, baseDir string) ([]SkillInfo, error) {
	var skills []SkillInfo

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return skills, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(baseDir, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}

		content := string(data)
		name, desc := parseSkillMeta(content)

		skills = append(skills, SkillInfo{
			Name:        name,
			Description: desc,
			Content:     content,
		})
	}

	return skills, nil
}

// parseSkillMeta extracts name and description from SKILL.md frontmatter
func parseSkillMeta(content string) (name, desc string) {
	lines := strings.Split(content, "\n")
	inFrontmatter := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			// End of frontmatter
			break
		}

		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			} else if strings.HasPrefix(line, "description:") {
				desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}

	return
}

// FormatSkillsForPrompt formats skills into a prompt section
func FormatSkillsForPrompt(skills []SkillInfo) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n## Available Skills\n")

	for _, skill := range skills {
		sb.WriteString("\n### ")
		sb.WriteString(skill.Name)
		sb.WriteString("\n")
		sb.WriteString(skill.Description)
		sb.WriteString("\n")
		sb.WriteString(skill.Content)
		sb.WriteString("\n")
	}

	return sb.String()
}
