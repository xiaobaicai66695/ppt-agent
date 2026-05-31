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
	"sync"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/style"
)

// Updater 偏好更新器
type Updater struct {
	profileStore *style.EnhancedProfileStore
	analyzer    *Analyzer

	// 更新统计
	mu             sync.RWMutex
	updateCount    int
	lastUpdateTime time.Time
}

// NewUpdater 创建偏好更新器
func NewUpdater(profileStore *style.EnhancedProfileStore) *Updater {
	return &Updater{
		profileStore: profileStore,
		analyzer:    NewAnalyzer(),
	}
}

// UpdateFromSignal 根据学习信号更新用户偏好
func (u *Updater) UpdateFromSignal(signal *LearningSignal) {
	if signal == nil {
		return
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	switch signal.Type {
	case SignalExplicitFeedback:
		u.updateFromExplicitFeedback(signal)
	case SignalImplicitFeedback:
		u.updateFromImplicitFeedback(signal)
	case SignalEditAction:
		u.updateFromEditAction(signal)
	case SignalCompletion:
		u.updateFromCompletion(signal)
	case SignalAbandonTask:
		u.updateFromAbandon(signal)
	case SignalQAResult:
		u.updateFromQAResult(signal)
	}

	u.updateCount++
	u.lastUpdateTime = time.Now()
}

// updateFromExplicitFeedback 从显式反馈更新
func (u *Updater) updateFromExplicitFeedback(signal *LearningSignal) {
	profile := u.profileStore.GetEnhanced(signal.UserID)
	if profile == nil {
		return
	}

	// 更新成功率
	profile.TotalTasks++
	if signal.Context != nil && signal.Context.QualityScore > 0 {
		// 假设质量评分 1-5，>3.5 为成功
		if signal.Context.QualityScore >= 3.5 {
			successCount := int(float64(profile.TotalTasks-1) * profile.SuccessRate)
			successCount++
			profile.SuccessRate = float64(successCount) / float64(profile.TotalTasks)
		}
	}

	// 从反馈数据中提取偏好
	if data := signal.Data; data != nil {
		if template, ok := data["template"].(string); ok && template != "" {
			profile.PreferredThemes = u.addToPreference(profile.PreferredThemes, template)
		}
		if theme, ok := data["theme"].(string); ok && theme != "" {
			profile.PreferredThemes = u.addToPreference(profile.PreferredThemes, theme)
		}
		if tone, ok := data["language_tone"].(string); ok && tone != "" {
			profile.LanguageTone = tone
		}
	}

	u.profileStore.SaveEnhanced(profile)
}

// updateFromImplicitFeedback 从隐式反馈更新
func (u *Updater) updateFromImplicitFeedback(signal *LearningSignal) {
	profile := u.profileStore.GetEnhanced(signal.UserID)
	if profile == nil {
		return
	}

	// 根据行为类型更新偏好
	if data := signal.Data; data != nil {
		action, _ := data["action"].(string)

		switch action {
		case "adjust_color":
			if color, ok := data["color"].(string); ok {
				profile.PreferredColors = u.addToPreference(profile.PreferredColors, color)
			}
		case "adjust_layout":
			if layout, ok := data["layout"].(string); ok {
				profile.LayoutPreferences = u.addToPreference(profile.LayoutPreferences, layout)
			}
		case "skip_page":
			// 用户跳过某些页面类型，记录为不喜欢
			if pageType, ok := data["page_type"].(string); ok {
				// 可以添加到不喜欢列表
				logger.Debug("user_skipped_page_type", "type", pageType)
			}
		case "zoom_in", "zoom_out":
			// 记录用户对页面细节的关注度
		}
	}

	u.profileStore.SaveEnhanced(profile)
}

// updateFromEditAction 从编辑行为更新
func (u *Updater) updateFromEditAction(signal *LearningSignal) {
	profile := u.profileStore.GetEnhanced(signal.UserID)
	if profile == nil {
		return
	}

	// 编辑行为表示生成的初始内容不完全符合期望
	// 但仍然使用了基础模板/主题，所以稍微调整权重

	if data := signal.Data; data != nil {
		editType, _ := data["edit_type"].(string)

		switch editType {
		case "content_modification":
			// 内容修改，略微降低该模板的置信度
			profile.SpecialNotes = append(profile.SpecialNotes, "内容可能需要调整")
		case "style_modification":
			// 样式修改，可能需要更新配色偏好
			if style, ok := data["new_style"].(string); ok {
				profile.PreferredThemes = u.addToPreference(profile.PreferredThemes, style)
			}
		case "layout_modification":
			// 布局修改，更新布局偏好
			if layout, ok := data["new_layout"].(string); ok {
				profile.LayoutPreferences = u.addToPreference(profile.LayoutPreferences, layout)
			}
		}
	}

	u.profileStore.SaveEnhanced(profile)
}

// updateFromCompletion 从任务完成更新
func (u *Updater) updateFromCompletion(signal *LearningSignal) {
	profile := u.profileStore.GetEnhanced(signal.UserID)
	if profile == nil {
		return
	}

	// 任务成功完成，强化成功模式
	profile.TotalTasks++
	profile.LastActiveTime = time.Now()

	if signal.Context != nil {
		// 成功率更新
		if signal.Context.QualityScore >= 3.5 {
			successCount := int(float64(profile.TotalTasks-1) * profile.SuccessRate)
			successCount++
			profile.SuccessRate = float64(successCount) / float64(profile.TotalTasks)
		}

		// 如果是高质量完成，强化当前配置
		if signal.Context.QualityScore >= 4.0 {
			u.analyzer.RecordSuccess(signal.TaskID, profile)
		}
	}

	u.profileStore.SaveEnhanced(profile)
}

// updateFromAbandon 从放弃任务更新
func (u *Updater) updateFromAbandon(signal *LearningSignal) {
	profile := u.profileStore.GetEnhanced(signal.UserID)
	if profile == nil {
		return
	}

	// 任务被放弃，降低成功率
	profile.TotalTasks++
	successCount := int(float64(profile.TotalTasks-1) * profile.SuccessRate)
	profile.SuccessRate = float64(successCount) / float64(profile.TotalTasks)

	// 记录放弃原因
	if data := signal.Data; data != nil {
		if reason, ok := data["reason"].(string); ok {
			profile.SpecialNotes = append(profile.SpecialNotes, "放弃原因: "+reason)
		}
	}

	u.profileStore.SaveEnhanced(profile)
}

// updateFromQAResult 从QA结果更新
func (u *Updater) updateFromQAResult(signal *LearningSignal) {
	profile := u.profileStore.GetEnhanced(signal.UserID)
	if profile == nil {
		return
	}

	if data := signal.Data; data != nil {
		hasIssue, _ := data["has_issue"].(bool)
		severity, _ := data["severity"].(string)

		if hasIssue {
			// 根据问题严重程度调整
			switch severity {
			case "high":
				profile.SpecialNotes = append(profile.SpecialNotes, "存在严重质量问题，需关注")
			case "medium":
				profile.SpecialNotes = append(profile.SpecialNotes, "存在中等问题")
			case "low":
				profile.SpecialNotes = append(profile.SpecialNotes, "存在轻微问题")
			}
		}
	}

	u.profileStore.SaveEnhanced(profile)
}

// addToPreference 添加到偏好列表
func (u *Updater) addToPreference(list []string, value string) []string {
	if value == "" {
		return list
	}
	list = append(list, value)
	// 限制列表长度
	if len(list) > 20 {
		list = list[len(list)-20:]
	}
	return list
}

// BatchUpdate 批量更新偏好
func (u *Updater) BatchUpdate(signals []*LearningSignal) {
	for _, signal := range signals {
		u.UpdateFromSignal(signal)
	}
}

// GetStats 获取更新统计
func (u *Updater) GetStats() map[string]interface{} {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return map[string]interface{}{
		"update_count":    u.updateCount,
		"last_update":     u.lastUpdateTime,
	}
}
