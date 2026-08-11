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

package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/style"
)

// GenerateModel 是 Updater 所需的轻量 LLM 接口（与 intent.Classifier 的 textModelFactory 兼容）。
type GenerateModel interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error)
}

// Updater 偏好更新器 —— 通过 LLM 分析累积的信号，输出结构化画像更新。
// 不再使用硬编码的 if-then 规则，而是将信号批量发送给 LLM 进行模式分析。
type Updater struct {
	profileStore *style.EnhancedProfileStore
	modelFactory func(ctx context.Context) (GenerateModel, error)

	// 信号缓冲：按 userID 分组累积，达到阈值或定时触发 LLM 分析
	mu         sync.Mutex
	buffers    map[int][]*LearningSignal // userID → signals
	bufferSize int
	flushTimer *time.Ticker
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// UpdaterConfig 更新器配置。
type UpdaterConfig struct {
	BufferSize    int           // 每用户累积多少条信号后触发 LLM 分析（默认 10）
	FlushInterval time.Duration // 定时刷新间隔（默认 5 分钟）
}

// NewUpdater 创建偏好更新器。
// modelFactory 为 nil 时仍可工作——此时只做简单的统计更新（total_tasks、success_rate），
// 不做 LLM 深度分析。
func NewUpdater(profileStore *style.EnhancedProfileStore, modelFactory func(ctx context.Context) (GenerateModel, error)) *Updater {
	return NewUpdaterWithConfig(profileStore, modelFactory, nil)
}

// NewUpdaterWithConfig 使用自定义配置创建偏好更新器。
func NewUpdaterWithConfig(profileStore *style.EnhancedProfileStore, modelFactory func(ctx context.Context) (GenerateModel, error), cfg *UpdaterConfig) *Updater {
	if cfg == nil {
		cfg = &UpdaterConfig{
			BufferSize:    10,
			FlushInterval: 5 * time.Minute,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	u := &Updater{
		profileStore: profileStore,
		modelFactory: modelFactory,
		buffers:      make(map[int][]*LearningSignal),
		bufferSize:   cfg.BufferSize,
		flushTimer:   time.NewTicker(cfg.FlushInterval),
		ctx:          ctx,
		cancel:       cancel,
	}

	// 后台定时刷新
	u.wg.Add(1)
	go u.flushLoop()

	return u
}

// flushLoop 定时将各用户缓冲的信号刷新到 LLM 分析。
func (u *Updater) flushLoop() {
	defer u.wg.Done()
	for {
		select {
		case <-u.ctx.Done():
			u.flushTimer.Stop()
			u.flushAll()
			return
		case <-u.flushTimer.C:
			u.flushAll()
		}
	}
}

// UpdateFromSignal 接收一条学习信号，累积到对应用户的缓冲区。
// 当缓冲区达到阈值时，异步触发 LLM 分析。
func (u *Updater) UpdateFromSignal(signal *LearningSignal) {
	if signal == nil {
		return
	}

	u.mu.Lock()
	u.buffers[signal.UserID] = append(u.buffers[signal.UserID], signal)
	count := len(u.buffers[signal.UserID])
	u.mu.Unlock()

	if count >= u.bufferSize {
		// 异步刷新该用户（不阻塞信号采集）
		go u.flushUser(signal.UserID)
	}
}

// flushAll 刷新所有用户的缓冲区。
func (u *Updater) flushAll() {
	u.mu.Lock()
	userIDs := make([]int, 0, len(u.buffers))
	for uid := range u.buffers {
		userIDs = append(userIDs, uid)
	}
	u.mu.Unlock()

	for _, uid := range userIDs {
		u.flushUser(uid)
	}
}

// flushUser 取出某用户的所有累积信号，调用 LLM 分析并更新画像。
func (u *Updater) flushUser(userID int) {
	u.mu.Lock()
	signals := u.buffers[userID]
	if len(signals) == 0 {
		u.mu.Unlock()
		return
	}
	delete(u.buffers, userID)
	u.mu.Unlock()

	profile := u.profileStore.GetEnhanced(userID)
	if profile == nil {
		profile = style.NewEnhancedProfile(userID)
	}

	// 无条件做基础统计（total_tasks、success_rate）
	u.applyBaseStats(profile, signals)

	// LLM 深度分析
	if u.modelFactory != nil {
		ctx, cancel := context.WithTimeout(u.ctx, 30*time.Second)
		defer cancel()
		if err := u.llmAnalyzeAndApply(ctx, profile, signals); err != nil {
			logger.Warn("updater_llm_analysis_failed", "user_id", userID, "error", err.Error())
		}
	}

	u.profileStore.SaveEnhanced(profile)
	logger.Debug("updater_profile_saved", "user_id", userID, "signal_count", len(signals))
}

// applyBaseStats 做无需 LLM 的基础统计（纯计数，不涉及语义判断）。
func (u *Updater) applyBaseStats(profile *style.EnhancedProfile, signals []*LearningSignal) {
	for _, s := range signals {
		profile.LastActiveTime = time.Now()

		switch s.Type {
		case SignalCompletion:
			profile.TotalTasks++
			if s.Context != nil && s.Context.QualityScore >= 3.5 {
				successCount := int(float64(profile.TotalTasks-1) * profile.SuccessRate)
				successCount++
				profile.SuccessRate = float64(successCount) / float64(profile.TotalTasks)
			} else {
				profile.TotalTasks++ // 未达质量阈值也算一次任务
				profile.SuccessRate = float64(int(float64(profile.TotalTasks-1)*profile.SuccessRate)) / float64(profile.TotalTasks)
			}
		case SignalAbandonTask:
			profile.TotalTasks++
			profile.SuccessRate = float64(int(float64(profile.TotalTasks-1)*profile.SuccessRate)) / float64(profile.TotalTasks)
		case SignalExplicitFeedback:
			if s.Context != nil && s.Context.QualityScore >= 3.5 {
				profile.TotalTasks++
				successCount := int(float64(profile.TotalTasks-1) * profile.SuccessRate)
				successCount++
				profile.SuccessRate = float64(successCount) / float64(profile.TotalTasks)
			}
		}
	}
}

// llmAnalyzeAndApply 将信号发送给 LLM，解析返回的画像更新并应用。
func (u *Updater) llmAnalyzeAndApply(ctx context.Context, profile *style.EnhancedProfile, signals []*LearningSignal) error {
	m, err := u.modelFactory(ctx)
	if err != nil {
		return fmt.Errorf("create model: %w", err)
	}

	prompt := buildProfileAnalysisPrompt(profile, signals)

	resp, err := m.Generate(ctx, []*schema.Message{
		{Role: schema.System, Content: profileAnalysisSystemPrompt},
		{Role: schema.User, Content: prompt},
	})
	if err != nil {
		return fmt.Errorf("llm generate: %w", err)
	}

	updates, err := parseProfileUpdates(resp.Content)
	if err != nil {
		return fmt.Errorf("parse llm response: %w", err)
	}

	applyProfileUpdates(profile, updates)
	return nil
}

// profileAnalysisSystemPrompt 是 LLM 分析用户画像的 system prompt。
const profileAnalysisSystemPrompt = `你是一个用户画像分析引擎，负责从用户的 PPT 生成行为中提取偏好模式。

你会收到：
1. 用户当前画像（JSON）
2. 最近的用户行为信号列表

你需要分析这些信号中隐含的模式，输出一个 JSON 描述需要更新哪些画像字段。

## 输出格式

严格按以下 JSON 输出（只输出 JSON，不要 markdown 代码块）：

{
  "analysis": "简短分析，1-2句话总结发现的模式",
  "updates": {
    "preferred_themes": ["ocean_soft"],
    "preferred_colors": ["#1a1a2e", "#e94560"],
    "language_tone": "专业严谨",
    "layout_preferences": ["two-column", "full-width"],
    "content_types": {"text": 5, "chart": 3, "image": 1},
    "special_notes": ["用户偏好深色背景", "经常要求数据图表"],
    "domain_preferences": {"technical": 5, "business": 2},
    "animation_level": "minimal" ,
    "content_tone": {
      "formality": "formal",
      "tech_density": 8,
      "detail_level": 7,
      "humor_level": 1
    }
  }
}

## 规则

1. 只输出有变化或新发现的字段，没变化的字段不要出现
2. preferred_themes / preferred_colors / layout_preferences 是追加式列表（新值追加到已有列表末尾）
3. language_tone 是单一字符串，后出现的覆盖前面的
4. content_types 是计数 map，正数表示偏好，负数表示不喜欢
5. special_notes 是追加式列表
6. domain_preferences 是计数 map（领域→累计次数）
7. 如果信号太少不足以判断模式，analysis 中说明"数据不足"，updates 置为 {}
8. animation_level 取值: "none" / "minimal" / "moderate" / "dynamic"
9. content_tone.formality 取值: "casual" / "semi-formal" / "formal"
10. 所有字段都是可选的，只输出确定有变化的`

// buildProfileAnalysisPrompt 构建发送给 LLM 的用户 prompt。
func buildProfileAnalysisPrompt(profile *style.EnhancedProfile, signals []*LearningSignal) string {
	profileJSON, _ := json.MarshalIndent(profileSummary(profile), "", "  ")

	var signalLines []string
	for _, s := range signals {
		entry := map[string]interface{}{
			"type":      s.Type.String(),
			"timestamp": s.Timestamp.Format("2006-01-02 15:04:05"),
		}
		if s.Context != nil {
			entry["phase"] = s.Context.TaskPhase
			entry["quality_score"] = s.Context.QualityScore
			entry["duration_sec"] = s.Context.Duration
			entry["action_type"] = s.Context.ActionType
			entry["page_index"] = s.Context.PageIndex
		}
		if s.Data != nil {
			entry["data"] = s.Data
		}
		b, _ := json.Marshal(entry)
		signalLines = append(signalLines, string(b))
	}

	return fmt.Sprintf(`## 当前用户画像
%s

## 最近行为信号 (%d 条)
%s

请分析以上信号，输出需要更新的画像字段。`, string(profileJSON), len(signals), strings.Join(signalLines, "\n"))
}

// profileSummary 提取画像中 LLM 需要感知的字段（去掉内部时间戳等噪音）。
func profileSummary(p *style.EnhancedProfile) map[string]interface{} {
	return map[string]interface{}{
		"user_id":            p.UserID,
		"total_tasks":        p.TotalTasks,
		"success_rate":       p.SuccessRate,
		"preferred_themes":   p.PreferredThemes,
		"preferred_colors":   p.PreferredColors,
		"language_tone":      p.LanguageTone,
		"layout_preferences": p.LayoutPreferences,
		"content_types":      p.ContentTypes,
		"domain_preferences": p.DomainPreferences,
		"special_notes":      p.SpecialNotes,
		"animation_level":    p.AnimationLevel.String(),
		"content_tone": map[string]interface{}{
			"formality":    p.ContentTone.Formality,
			"tech_density": p.ContentTone.TechDensity,
			"detail_level": p.ContentTone.DetailLevel,
			"humor_level":  p.ContentTone.HumorLevel,
		},
	}
}

// profileUpdates LLM 返回的画像更新指令。
type profileUpdates struct {
	Analysis string               `json:"analysis"`
	Updates  profileUpdatesFields `json:"updates"`
}

type profileUpdatesFields struct {
	PreferredThemes   []string           `json:"preferred_themes"`
	PreferredColors   []string           `json:"preferred_colors"`
	LanguageTone      string             `json:"language_tone"`
	LayoutPreferences []string           `json:"layout_preferences"`
	ContentTypes      map[string]int     `json:"content_types"`
	SpecialNotes      []string           `json:"special_notes"`
	DomainPreferences map[string]int     `json:"domain_preferences"`
	AnimationLevel    string             `json:"animation_level"`
	ContentTone       *contentToneUpdate `json:"content_tone"`
}

type contentToneUpdate struct {
	Formality   string `json:"formality"`
	TechDensity int    `json:"tech_density"`
	DetailLevel int    `json:"detail_level"`
	HumorLevel  int    `json:"humor_level"`
}

// parseProfileUpdates 从 LLM 响应中提取画像更新指令。
func parseProfileUpdates(content string) (*profileUpdates, error) {
	content = strings.TrimSpace(content)
	// 去除可能的 markdown 代码块包裹
	if idx := strings.Index(content, "```"); idx >= 0 {
		start := idx + 3
		if rest := content[start:]; len(rest) > 4 && rest[:4] == "json" {
			start += 4
		}
		end := strings.Index(content[start:], "```")
		if end >= 0 {
			content = content[start : start+end]
		}
	}
	var updates profileUpdates
	if err := json.Unmarshal([]byte(content), &updates); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w (content: %.200s)", err, content)
	}
	return &updates, nil
}

// applyProfileUpdates 将 LLM 输出的更新字段写入画像。
func applyProfileUpdates(p *style.EnhancedProfile, u *profileUpdates) {
	if u == nil {
		return
	}
	f := u.Updates

	// 追加式列表
	if len(f.PreferredThemes) > 0 {
		p.PreferredThemes = appendUnique(p.PreferredThemes, f.PreferredThemes...)
	}
	if len(f.PreferredColors) > 0 {
		p.PreferredColors = appendUnique(p.PreferredColors, f.PreferredColors...)
	}
	if len(f.LayoutPreferences) > 0 {
		p.LayoutPreferences = appendUnique(p.LayoutPreferences, f.LayoutPreferences...)
	}
	if len(f.SpecialNotes) > 0 {
		p.SpecialNotes = append(p.SpecialNotes, f.SpecialNotes...)
		if len(p.SpecialNotes) > 50 {
			p.SpecialNotes = p.SpecialNotes[len(p.SpecialNotes)-50:]
		}
	}

	// 覆盖式字段
	if f.LanguageTone != "" {
		p.LanguageTone = f.LanguageTone
	}

	// 计数 map
	if f.ContentTypes != nil {
		if p.ContentTypes == nil {
			p.ContentTypes = style.ContentTypeCount{}
		}
		for k, v := range f.ContentTypes {
			p.ContentTypes[k] += v
		}
	}
	if f.DomainPreferences != nil {
		if p.DomainPreferences == nil {
			p.DomainPreferences = make(map[string]int)
		}
		for k, v := range f.DomainPreferences {
			p.DomainPreferences[k] += v
		}
	}

	// 动画级别
	if f.AnimationLevel != "" {
		p.AnimationLevel = style.ParseAnimationLevel(f.AnimationLevel)
	}

	// 内容语调
	if f.ContentTone != nil {
		t := f.ContentTone
		if t.Formality != "" {
			p.ContentTone.Formality = t.Formality
		}
		if t.TechDensity > 0 {
			p.ContentTone.TechDensity = clamp(t.TechDensity, 1, 10)
		}
		if t.DetailLevel > 0 {
			p.ContentTone.DetailLevel = clamp(t.DetailLevel, 1, 10)
		}
		if t.HumorLevel > 0 {
			p.ContentTone.HumorLevel = clamp(t.HumorLevel, 0, 10)
		}
	}
}

// appendUnique 追加不重复的元素到切片末尾。
func appendUnique(slice []string, items ...string) []string {
	seen := make(map[string]bool, len(slice))
	for _, s := range slice {
		seen[s] = true
	}
	for _, item := range items {
		if !seen[item] {
			slice = append(slice, item)
			seen[item] = true
		}
	}
	// 限制长度
	if len(slice) > 30 {
		slice = slice[len(slice)-30:]
	}
	return slice
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// BatchUpdate 批量更新偏好（对外兼容接口）。
func (u *Updater) BatchUpdate(signals []*LearningSignal) {
	for _, signal := range signals {
		u.UpdateFromSignal(signal)
	}
}

// Close 关闭更新器，刷新所有缓冲。
func (u *Updater) Close() {
	u.cancel()
	u.wg.Wait()
}
