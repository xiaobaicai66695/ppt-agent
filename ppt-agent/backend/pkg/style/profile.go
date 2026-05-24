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

package style

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// ContentTypeCount tracks usage frequency of content types.
type ContentTypeCount map[string]int

// UserProfile represents a user's style preferences learned from past tasks.
type UserProfile struct {
	UserID            int              `json:"user_id"`
	PreferredThemes   []string         `json:"preferred_themes"`
	PreferredColors   []string         `json:"preferred_colors"`
	ContentPatterns  []string         `json:"content_patterns"`
	LayoutPreferences []string        `json:"layout_preferences"`
	LanguageTone     string           `json:"language_tone"`
	TypicalPageCount int              `json:"typical_page_count"`
	ContentTypes     ContentTypeCount `json:"content_types"`
	SpecialNotes     []string         `json:"special_notes"`
	TaskCount        int              `json:"task_count"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// MergeWith updates the profile with new style data using weighted averaging.
func (p *UserProfile) MergeWith(new *ExtractedStyle) {
	p.TaskCount++

	if p.TaskCount <= 3 {
		p.mergeColdStart(new)
		return
	}

	weight := 1.0 / float64(p.TaskCount)
	existingWeight := 1.0 - weight

	p.mergeWeighted(new, existingWeight, weight)
}

func (p *UserProfile) mergeColdStart(new *ExtractedStyle) {
	if len(new.Themes) > 0 {
		p.PreferredThemes = deduplicate(append(p.PreferredThemes, new.Themes...))
	}
	if len(new.Colors) > 0 {
		p.PreferredColors = deduplicate(append(p.PreferredColors, new.Colors...))
	}
	if len(new.ContentPatterns) > 0 {
		p.ContentPatterns = deduplicate(append(p.ContentPatterns, new.ContentPatterns...))
	}
	if len(new.LayoutPreferences) > 0 {
		p.LayoutPreferences = deduplicate(append(p.LayoutPreferences, new.LayoutPreferences...))
	}
	if new.LanguageTone != "" {
		p.LanguageTone = new.LanguageTone
	}
	if new.PageCount > 0 {
		p.TypicalPageCount = new.PageCount
	}
	if p.ContentTypes == nil {
		p.ContentTypes = make(ContentTypeCount)
	}
	for ct, cnt := range new.ContentTypes {
		p.ContentTypes[ct] += cnt
	}
	if len(new.SpecialNotes) > 0 {
		p.SpecialNotes = deduplicate(append(p.SpecialNotes, new.SpecialNotes...))
	}
}

func (p *UserProfile) mergeWeighted(new *ExtractedStyle, existingWeight, newWeight float64) {
	p.PreferredThemes = mergeStringSliceWeighted(p.PreferredThemes, new.Themes, existingWeight, newWeight)
	p.PreferredColors = mergeStringSliceWeighted(p.PreferredColors, new.Colors, existingWeight, newWeight)
	p.ContentPatterns = mergeStringSliceWeighted(p.ContentPatterns, new.ContentPatterns, existingWeight, newWeight)
	p.LayoutPreferences = mergeStringSliceWeighted(p.LayoutPreferences, new.LayoutPreferences, existingWeight, newWeight)

	if new.LanguageTone != "" && p.LanguageTone == "" {
		p.LanguageTone = new.LanguageTone
	}

	if new.PageCount > 0 {
		p.TypicalPageCount = int(float64(p.TypicalPageCount)*existingWeight + float64(new.PageCount)*newWeight)
	}

	if p.ContentTypes == nil {
		p.ContentTypes = make(ContentTypeCount)
	}
	for ct, cnt := range new.ContentTypes {
		p.ContentTypes[ct] = int(float64(p.ContentTypes[ct])*existingWeight + float64(cnt)*newWeight)
	}

	if len(new.SpecialNotes) > 0 {
		p.SpecialNotes = mergeStringSliceWeighted(p.SpecialNotes, new.SpecialNotes, existingWeight, newWeight)
	}
}

// BuildStyleContext generates a human-readable style context string for prompt injection.
func (p *UserProfile) BuildStyleContext() string {
	if p.TaskCount == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n【用户风格偏好】（基于您过去 ")
	sb.WriteString(intToStr(p.TaskCount))
	sb.WriteString(" 次任务总结）\n")

	if len(p.PreferredThemes) > 0 {
		sb.WriteString("- 配色主题：")
		sb.WriteString(strings.Join(p.PreferredThemes, "、"))
		sb.WriteString("\n")
	}
	if len(p.PreferredColors) > 0 {
		sb.WriteString("- 色彩搭配：")
		sb.WriteString(strings.Join(p.PreferredColors, "、"))
		sb.WriteString("\n")
	}
	if len(p.ContentPatterns) > 0 {
		sb.WriteString("- 内容风格：")
		sb.WriteString(strings.Join(p.ContentPatterns, "、"))
		sb.WriteString("\n")
	}
	if len(p.LayoutPreferences) > 0 {
		sb.WriteString("- 布局偏好：")
		sb.WriteString(strings.Join(p.LayoutPreferences, "、"))
		sb.WriteString("\n")
	}
	if p.LanguageTone != "" {
		sb.WriteString("- 语言风格：")
		sb.WriteString(p.LanguageTone)
		sb.WriteString("\n")
	}
	if p.TypicalPageCount > 0 {
		sb.WriteString("- 常用页数：约 ")
		sb.WriteString(intToStr(p.TypicalPageCount))
		sb.WriteString(" 页\n")
	}
	if len(p.SpecialNotes) > 0 {
		for _, note := range p.SpecialNotes {
			sb.WriteString("- ")
			sb.WriteString(note)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("请在生成PPT时遵循上述偏好。\n")
	return sb.String()
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// ExtractedStyle represents style data extracted from a single task.
type ExtractedStyle struct {
	Themes           []string         `json:"themes,omitempty"`
	Colors           []string         `json:"colors,omitempty"`
	ContentPatterns  []string         `json:"content_patterns,omitempty"`
	LayoutPreferences []string        `json:"layout_preferences,omitempty"`
	LanguageTone     string           `json:"language_tone,omitempty"`
	PageCount        int              `json:"page_count,omitempty"`
	ContentTypes     ContentTypeCount `json:"content_types,omitempty"`
	SpecialNotes     []string         `json:"special_notes,omitempty"`
}

// TaskItemInfo is a simplified representation of a task item for style extraction.
type TaskItemInfo struct {
	ContentType string `json:"content_type"`
	Theme      string `json:"theme"`
}

// ExtractFromTasks extracts style preferences from a list of task items and theme.
func ExtractFromTasks(tasks []*TaskItemInfo, theme string) *ExtractedStyle {
	es := &ExtractedStyle{
		ContentTypes: make(ContentTypeCount),
	}

	es.PageCount = len(tasks)

	for _, t := range tasks {
		if t.ContentType != "" {
			es.ContentTypes[t.ContentType]++
		}
	}

	if theme != "" {
		es.Themes = append(es.Themes, theme)
	}

	return es
}

// ExtractFromQuery analyzes the user's query for style hints.
func ExtractFromQuery(query string) *ExtractedStyle {
	es := &ExtractedStyle{}

	queryLower := strings.ToLower(query)

	colorMap := map[string][]string{
		"科技":   {"ocean_soft", "sage_calm"},
		"技术":   {"ocean_soft"},
		"学术":   {"ocean_soft", "sage_calm"},
		"教学":   {"sage_calm", "simple_gray"},
		"商务":   {"charcoal_light", "warm_terracotta"},
		"路演":   {"charcoal_light", "berry_cream"},
		"政务":   {"government_red", "patriotic_blue"},
		"思政":   {"patriotic_blue", "government_red"},
		"答辩":   {"debate_purple", "charcoal_light"},
		"竞赛":   {"civic_gold", "innovation_green"},
		"创新":   {"civic_gold", "berry_cream"},
		"教育":   {"sage_calm", "simple_gray"},
		"环保":   {"sage_calm", "report_green"},
		"医疗":   {"ocean_soft", "medical_blue"},
		"金融":   {"charcoal_light", "finance_gold"},
		"简约":   {"simple_gray", "sage_calm"},
		"活泼":   {"activity_orange", "berry_cream"},
		"创意":   {"berry_cream", "lavender_mist"},
	}

	for keyword, themes := range colorMap {
		if strings.Contains(queryLower, keyword) {
			es.Themes = append(es.Themes, themes...)
			break
		}
	}

	if strings.Contains(queryLower, "图文") || strings.Contains(queryLower, "图片") {
		es.ContentPatterns = append(es.ContentPatterns, "图文并茂")
	}
	if strings.Contains(queryLower, "数据") || strings.Contains(queryLower, "图表") || strings.Contains(queryLower, "统计") {
		es.ContentPatterns = append(es.ContentPatterns, "数据图表多")
	}
	if strings.Contains(queryLower, "案例") || strings.Contains(queryLower, "实例") {
		es.ContentPatterns = append(es.ContentPatterns, "案例驱动")
	}
	if strings.Contains(queryLower, "对比") || strings.Contains(queryLower, "比较") {
		es.ContentPatterns = append(es.ContentPatterns, "对比分析")
	}

	if strings.Contains(queryLower, "正式") || strings.Contains(queryLower, "汇报") || strings.Contains(queryLower, "述职") {
		es.LanguageTone = "专业正式"
	} else if strings.Contains(queryLower, "分享") || strings.Contains(queryLower, "培训") {
		es.LanguageTone = "专业正式"
	} else if strings.Contains(queryLower, "活泼") || strings.Contains(queryLower, "有趣") {
		es.LanguageTone = "轻松活泼"
	} else if strings.Contains(queryLower, "技术") || strings.Contains(queryLower, "学术") {
		es.LanguageTone = "技术细节"
	}

	return es
}

// ProfileStore manages user style profiles backed by the database.
type ProfileStore struct{}

// NewProfileStore creates a new profile store backed by the database.
func NewProfileStore(_ string) *ProfileStore {
	return &ProfileStore{}
}

func (ps *ProfileStore) toDBRecord(p *UserProfile) *db.UserStyleProfile {
	themes, err := json.Marshal(p.PreferredThemes)
	if err != nil {
		logger.Warn("profile_marshal_failed", "field", "PreferredThemes", "error", err.Error())
	}
	colors, err := json.Marshal(p.PreferredColors)
	if err != nil {
		logger.Warn("profile_marshal_failed", "field", "PreferredColors", "error", err.Error())
	}
	patterns, err := json.Marshal(p.ContentPatterns)
	if err != nil {
		logger.Warn("profile_marshal_failed", "field", "ContentPatterns", "error", err.Error())
	}
	contentTypes, err := json.Marshal(p.ContentTypes)
	if err != nil {
		logger.Warn("profile_marshal_failed", "field", "ContentTypes", "error", err.Error())
	}
	notes, err := json.Marshal(p.SpecialNotes)
	if err != nil {
		logger.Warn("profile_marshal_failed", "field", "SpecialNotes", "error", err.Error())
	}

	return &db.UserStyleProfile{
		UserID:            uint(p.UserID),
		PreferredThemes:   string(themes),
		PreferredColors:   string(colors),
		ContentPatterns:   string(patterns),
		LanguageTone:      p.LanguageTone,
		TypicalPageCount:  p.TypicalPageCount,
		ContentTypes:      string(contentTypes),
		SpecialNotes:      string(notes),
		TaskCount:         p.TaskCount,
		UpdatedAt:        time.Now(),
	}
}

func (ps *ProfileStore) fromDBRecord(r *db.UserStyleProfile) *UserProfile {
	p := &UserProfile{
		UserID:            int(r.UserID),
		LanguageTone:      r.LanguageTone,
		TypicalPageCount: r.TypicalPageCount,
		TaskCount:        r.TaskCount,
		UpdatedAt:        r.UpdatedAt,
	}

	if err := json.Unmarshal([]byte(r.PreferredThemes), &p.PreferredThemes); err != nil {
		logger.Warn("profile_unmarshal_failed", "field", "PreferredThemes", "error", err.Error())
	}
	if err := json.Unmarshal([]byte(r.PreferredColors), &p.PreferredColors); err != nil {
		logger.Warn("profile_unmarshal_failed", "field", "PreferredColors", "error", err.Error())
	}
	if err := json.Unmarshal([]byte(r.ContentPatterns), &p.ContentPatterns); err != nil {
		logger.Warn("profile_unmarshal_failed", "field", "ContentPatterns", "error", err.Error())
	}
	if err := json.Unmarshal([]byte(r.ContentTypes), &p.ContentTypes); err != nil {
		logger.Warn("profile_unmarshal_failed", "field", "ContentTypes", "error", err.Error())
	}
	if err := json.Unmarshal([]byte(r.SpecialNotes), &p.SpecialNotes); err != nil {
		logger.Warn("profile_unmarshal_failed", "field", "SpecialNotes", "error", err.Error())
	}

	return p
}

// Get returns a user's profile, loading from the database.
func (ps *ProfileStore) Get(userID int) *UserProfile {
	r, err := db.GetUserStyleProfile(uint(userID))
	if err != nil || r == nil {
		return &UserProfile{UserID: userID, UpdatedAt: time.Now()}
	}
	return ps.fromDBRecord(r)
}

// Save persists a user's profile to the database.
func (ps *ProfileStore) Save(p *UserProfile) {
	r := ps.toDBRecord(p)
	db.UpsertUserStyleProfile(r)
}

// UpdateWithTask updates a user's profile with style data extracted from a task.
func (ps *ProfileStore) UpdateWithTask(userID int, extracted *ExtractedStyle) {
	p := ps.Get(userID)
	p.MergeWith(extracted)
	ps.Save(p)
}

// ── Helper functions ──────────────────────────────────────────────────────

func deduplicate(ss []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range ss {
		lower := strings.ToLower(s)
		if !seen[lower] {
			seen[lower] = true
			result = append(result, s)
		}
	}
	return result
}

func mergeStringSliceWeighted(existing, new []string, ew, nw float64) []string {
	if len(new) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return new
	}

	counts := make(map[string]int)
	for _, s := range existing {
		counts[strings.ToLower(s)]++
	}
	for _, s := range new {
		lower := strings.ToLower(s)
		counts[lower] += int(nw / ew)
	}

	type pair struct {
		s     string
		count int
	}
	var pairs []pair
	for s, c := range counts {
		pairs = append(pairs, pair{s, c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	result := make([]string, 0, min(len(pairs), 5))
	for i := 0; i < min(len(pairs), 5); i++ {
		result = append(result, pairs[i].s)
	}
	return result
}
