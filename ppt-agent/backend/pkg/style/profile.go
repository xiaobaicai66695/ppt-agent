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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

// ContentTypeCount 追踪内容类型的使用频率
type ContentTypeCount map[string]int

// AnimationLevel 动画级别
type AnimationLevel int

const (
	AnimationNone AnimationLevel = iota
	AnimationMinimal
	AnimationModerate
	AnimationRich
)

func (a AnimationLevel) String() string {
	switch a {
	case AnimationNone:
		return "none"
	case AnimationMinimal:
		return "minimal"
	case AnimationModerate:
		return "moderate"
	case AnimationRich:
		return "rich"
	default:
		return "moderate"
	}
}

func ParseAnimationLevel(s string) AnimationLevel {
	switch strings.ToLower(s) {
	case "none", "无":
		return AnimationNone
	case "minimal", "简洁":
		return AnimationMinimal
	case "moderate", "适中":
		return AnimationModerate
	case "rich", "丰富":
		return AnimationRich
	default:
		return AnimationModerate
	}
}

// ChartPreference 图表偏好
type ChartPreference struct {
	PreferredTypes []string `json:"preferred_types"` // 偏好的图表类型: bar, line, pie, etc.
	Use3D         bool     `json:"use_3d"`         // 是否使用3D效果
	ColorScheme   string   `json:"color_scheme"`   // 配色方案
}

// BrandPreferences 品牌元素偏好
type BrandPreferences struct {
	LogoPosition   string `json:"logo_position"`   // Logo位置: top-left, bottom-right, etc.
	ShowFooter     bool   `json:"show_footer"`      // 是否显示页脚
	FooterText     string `json:"footer_text"`     // 页脚文字
	UseWatermark   bool   `json:"use_watermark"`    // 是否使用水印
	WatermarkText  string `json:"watermark_text"`   // 水印文字
}

// ContentTone 内容语调
type ContentTone struct {
	Formality    string `json:"formality"`    // 正式程度: formal, semi-formal, casual
	TechDensity  int    `json:"tech_density"` // 技术术语密度 1-10
	DetailLevel  int    `json:"detail_level"` // 详细程度 1-10
	HumorLevel   int    `json:"humor_level"`  // 幽默程度 1-10
}

// SuccessPattern 成功模式
type SuccessPattern struct {
	Domain         string   `json:"domain"`          // 应用领域
	Template       string   `json:"template"`       // 成功使用的模板
	Theme          string   `json:"theme"`          // 成功使用的配色
	PageCount      int      `json:"page_count"`     // 成功使用的页数
	AvgQualityScore float64 `json:"avg_quality_score"` // 平均质量评分
	SuccessCount   int      `json:"success_count"`  // 成功次数
}

// EnhancedProfile 增强版用户画像
type EnhancedProfile struct {
	UserProfile // 嵌入现有字段

	// 新增偏好维度
	BrandElements    BrandPreferences `json:"brand_elements"`
	ChartPreferences ChartPreference  `json:"chart_preferences"`
	AnimationLevel   AnimationLevel  `json:"animation_level"`
	ContentTone      ContentTone     `json:"content_tone"`
	SuccessPatterns  []SuccessPattern `json:"success_patterns"`

	// 学习统计
	FirstTaskTime   time.Time `json:"first_task_time"`
	LastActiveTime  time.Time `json:"last_active_time"`
	TotalTasks      int       `json:"total_tasks"`
	SuccessRate     float64   `json:"success_rate"` // 任务成功率

	// 领域偏好（基于任务历史）
	DomainPreferences map[string]int `json:"domain_preferences"` // 领域 -> 次数
}

// NewEnhancedProfile 创建增强版用户画像
func NewEnhancedProfile(userID int) *EnhancedProfile {
	return &EnhancedProfile{
		UserProfile: UserProfile{
			UserID:   userID,
			TaskCount: 0,
			UpdatedAt: time.Now(),
		},
		DomainPreferences: make(map[string]int),
		AnimationLevel:    AnimationModerate,
		ContentTone: ContentTone{
			Formality:   "semi-formal",
			TechDensity: 5,
			DetailLevel: 5,
			HumorLevel:  2,
		},
		SuccessRate: 0.0,
	}
}

// GetPreferredTemplates 获取用户偏好的模板列表
func (p *EnhancedProfile) GetPreferredTemplates() []string {
	// 从成功模式中提取
	templates := make(map[string]int)
	for _, sp := range p.SuccessPatterns {
		templates[sp.Template] += sp.SuccessCount
	}

	// 从 ContentTypes 中推断模板偏好
	contentTypeToTemplate := map[string][]string{
		"title_slide":      {"tech-intro", "pitch-deck", "course-module"},
		"content_slide":    {"tech-intro", "weekly-report"},
		"section_divider":  {"tech-intro", "course-module"},
		"chart_slide":     {"research-report", "pitch-deck"},
		"summary_slide":   {"tech-intro", "course-module"},
	}

	for ct, count := range p.ContentTypes {
		if ts, ok := contentTypeToTemplate[ct]; ok {
			for _, t := range ts {
				templates[t] += count
			}
		}
	}

	// 排序返回
	type templateCount struct {
		template string
		count    int
	}
	var tc []templateCount
	for t, c := range templates {
		tc = append(tc, templateCount{t, c})
	}
	sort.Slice(tc, func(i, j int) bool { return tc[i].count > tc[j].count })

	result := make([]string, 0, len(tc))
	for _, t := range tc {
		result = append(result, t.template)
	}
	return result
}

// GetTypicalPageCount 获取用户典型页数
func (p *EnhancedProfile) GetTypicalPageCount() int {
	if p.TypicalPageCount > 0 {
		return p.TypicalPageCount
	}
	// 从成功模式计算
	var totalPages, totalCount int
	for _, sp := range p.SuccessPatterns {
		totalPages += sp.PageCount * sp.SuccessCount
		totalCount += sp.SuccessCount
	}
	if totalCount > 0 {
		return totalPages / totalCount
	}
	return 12 // 默认值
}

// GetPreferredTheme 获取用户偏好的主题
func (p *EnhancedProfile) GetPreferredTheme() string {
	if len(p.PreferredThemes) > 0 {
		return p.PreferredThemes[0]
	}
	return ""
}

// GetTopDomain 获取用户最常用的领域
func (p *EnhancedProfile) GetTopDomain() string {
	var topDomain string
	var maxCount int
	for domain, count := range p.DomainPreferences {
		if count > maxCount {
			maxCount = count
			topDomain = domain
		}
	}
	return topDomain
}

// Recommend 生成个性化推荐
func (p *EnhancedProfile) Recommend(req *RecommendRequest) *RecommendResult {
	// 计算复杂度对应的领域字符串
	complexityStr := "medium"
	if req.Complexity >= 7 {
		complexityStr = "high"
	} else if req.Complexity <= 3 {
		complexityStr = "low"
	}

	result := &RecommendResult{
		Template:  p.suggestTemplate(req.Domain, complexityStr),
		Theme:     p.suggestTheme(req.Domain),
		PageCount: p.suggestPageCount(req.Complexity),
		Animation: p.AnimationLevel.String(),
		Tips:      []string{},
	}

	// 添加个性化提示
	if p.TaskCount > 5 {
		result.Tips = append(result.Tips, fmt.Sprintf("根据您过去%d个任务的经验进行了优化", p.TaskCount))
	}
	if p.SuccessRate > 0.8 {
		result.Tips = append(result.Tips, "您最近的任务完成率很高，继续保持!")
	}

	return result
}

func (p *EnhancedProfile) suggestTemplate(domain, complexity string) string {
	domainTemplates := map[string][]string{
		"business":   {"pitch-deck", "product-launch"},
		"technical":   {"tech-sharing", "tech-intro"},
		"academic":    {"course-module", "design-defense"},
		"government":  {"politics-ideology", "current-affairs"},
		"personal":    {"personal-summary", "weekly-report"},
		"creative":   {"activity-plan", "product-launch"},
	}

	if templates, ok := domainTemplates[domain]; ok {
		return templates[0]
	}
	return "tech-intro"
}

func (p *EnhancedProfile) suggestTheme(domain string) string {
	if len(p.PreferredThemes) > 0 {
		return p.PreferredThemes[0]
	}
	domainThemes := map[string]string{
		"business":   "charcoal_light",
		"technical":  "ocean_soft",
		"academic":   "sage_calm",
		"government": "government_red",
		"personal":   "simple_gray",
		"creative":   "berry_cream",
	}
	if theme, ok := domainThemes[domain]; ok {
		return theme
	}
	return "simple_gray"
}

func (p *EnhancedProfile) suggestPageCount(complexity int) int {
	if p.TypicalPageCount > 0 {
		base := p.TypicalPageCount
		if complexity > 7 {
			return base + 4
		} else if complexity < 4 {
			return base - 4
		}
		return base
	}
	if complexity > 7 {
		return 18
	} else if complexity < 4 {
		return 8
	}
	return 12
}

// RecommendRequest 推荐请求
type RecommendRequest struct {
	Domain     string
	Complexity int
	PageCount  int
}

// RecommendResult 推荐结果
type RecommendResult struct {
	Template  string   `json:"template"`
	Theme     string   `json:"theme"`
	PageCount int      `json:"page_count"`
	Animation string   `json:"animation"`
	Tips      []string `json:"tips"`
}

// UserProfile 表示从过去任务中学习的用户风格偏好
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

// MergeWith 使用加权平均更新配置文件的样式数据
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

// BuildStyleContext 生成人类可读的风格上下文字符串，用于注入到提示词
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

// ExtractedStyle 表示从单个任务中提取的样式数据
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

// TaskItemInfo 是用于样式提取的任务项的简化表示
type TaskItemInfo struct {
	ContentType string `json:"content_type"`
	Theme      string `json:"theme"`
}

// ExtractFromTasks 从任务项列表和主题中提取样式偏好
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

// ExtractFromQuery 分析用户查询中的风格提示
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

// ProfileStore 管理以数据库为后端的用户风格配置文件
type ProfileStore struct{}

// NewProfileStore 创建一个以数据库为后端的新配置文件存储
func NewProfileStore(_ string) *ProfileStore {
	return &ProfileStore{}
}

// EnhancedProfileStore 增强版画像存储
type EnhancedProfileStore struct {
	*ProfileStore
}

// NewEnhancedProfileStore 创建增强版画像存储
func NewEnhancedProfileStore() *EnhancedProfileStore {
	return &EnhancedProfileStore{
		ProfileStore: NewProfileStore(""),
	}
}

// GetEnhanced 获取增强版用户画像
func (s *EnhancedProfileStore) GetEnhanced(userID int) *EnhancedProfile {
	// 先获取基础画像
	profile := s.ProfileStore.Get(userID)

	// 转换为增强版
	enhanced := &EnhancedProfile{
		UserProfile: *profile,
	}

	// 加载额外的偏好设置
	extPrefs := s.loadExtendedPreferences(userID)
	if extPrefs != nil {
		enhanced.BrandElements = extPrefs.BrandElements
		enhanced.ChartPreferences = extPrefs.ChartPreferences
		enhanced.AnimationLevel = extPrefs.AnimationLevel
		enhanced.ContentTone = extPrefs.ContentTone
		enhanced.SuccessPatterns = extPrefs.SuccessPatterns
		enhanced.SuccessRate = extPrefs.SuccessRate
		enhanced.DomainPreferences = extPrefs.DomainPreferences
	}

	return enhanced
}

// SaveEnhanced 保存增强版用户画像
func (s *EnhancedProfileStore) SaveEnhanced(p *EnhancedProfile) {
	// 保存基础画像
	s.ProfileStore.Save(&p.UserProfile)

	// 保存扩展偏好
	extPrefs := &extendedPreferences{
		BrandElements:    p.BrandElements,
		ChartPreferences: p.ChartPreferences,
		AnimationLevel:   p.AnimationLevel,
		ContentTone:      p.ContentTone,
		SuccessPatterns:  p.SuccessPatterns,
		SuccessRate:     p.SuccessRate,
		DomainPreferences: p.DomainPreferences,
	}
	s.saveExtendedPreferences(p.UserID, extPrefs)
}

// UpdateDomainPreference 更新领域偏好
func (s *EnhancedProfileStore) UpdateDomainPreference(userID int, domain string) {
	p := s.GetEnhanced(userID)
	if p.DomainPreferences == nil {
		p.DomainPreferences = make(map[string]int)
	}
	p.DomainPreferences[domain]++
	s.SaveEnhanced(p)
}

// RecordSuccess 记录任务成功
func (s *EnhancedProfileStore) RecordSuccess(userID int, domain, template, theme string, pageCount int) {
	p := s.GetEnhanced(userID)
	p.TotalTasks++

	// 更新成功率
	successCount := int(float64(p.TotalTasks) * p.SuccessRate)
	successCount++
	p.SuccessRate = float64(successCount) / float64(p.TotalTasks)

	// 更新成功模式
	found := false
	for i := range p.SuccessPatterns {
		if p.SuccessPatterns[i].Domain == domain {
			p.SuccessPatterns[i].SuccessCount++
			p.SuccessPatterns[i].AvgQualityScore = (p.SuccessPatterns[i].AvgQualityScore*float64(p.SuccessPatterns[i].SuccessCount-1) + 4.5) / float64(p.SuccessPatterns[i].SuccessCount)
			found = true
			break
		}
	}
	if !found {
		p.SuccessPatterns = append(p.SuccessPatterns, SuccessPattern{
			Domain:          domain,
			Template:        template,
			Theme:           theme,
			PageCount:       pageCount,
			AvgQualityScore: 4.5,
			SuccessCount:   1,
		})
	}

	// 更新活跃时间
	p.LastActiveTime = time.Now()
	s.SaveEnhanced(p)
}

// extendedPreferences 扩展偏好存储结构
type extendedPreferences struct {
	BrandElements     BrandPreferences   `json:"brand_elements"`
	ChartPreferences  ChartPreference    `json:"chart_preferences"`
	AnimationLevel    AnimationLevel     `json:"animation_level"`
	ContentTone       ContentTone        `json:"content_tone"`
	SuccessPatterns   []SuccessPattern   `json:"success_patterns"`
	SuccessRate       float64           `json:"success_rate"`
	DomainPreferences map[string]int     `json:"domain_preferences"`
}

// loadExtendedPreferences 从数据库加载扩展偏好
func (s *EnhancedProfileStore) loadExtendedPreferences(userID int) *extendedPreferences {
	r, err := db.GetUserStyleProfile(uint(userID))
	if err != nil || r == nil {
		return nil
	}

	ext := &extendedPreferences{}
	if r.ExtendedPreferences != "" {
		if err := json.Unmarshal([]byte(r.ExtendedPreferences), ext); err != nil {
			logger.Warn("extended_prefs_unmarshal_failed", "error", err.Error())
		}
	}

	// 兼容旧数据：回退到基础字段
	if ext.DomainPreferences == nil {
		ext.DomainPreferences = make(map[string]int)
	}

	return ext
}

// saveExtendedPreferences 保存扩展偏好到数据库
func (s *EnhancedProfileStore) saveExtendedPreferences(userID int, ext *extendedPreferences) {
	data, err := json.Marshal(ext)
	if err != nil {
		logger.Warn("extended_prefs_marshal_failed", "error", err.Error())
		return
	}

	// 更新数据库记录
	r, err := db.GetUserStyleProfile(uint(userID))
	if err != nil || r == nil {
		return
	}
	r.ExtendedPreferences = string(data)
	db.UpsertUserStyleProfile(r)
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

// Get 返回用户配置文件，从数据库加载
func (ps *ProfileStore) Get(userID int) *UserProfile {
	r, err := db.GetUserStyleProfile(uint(userID))
	if err != nil || r == nil {
		return &UserProfile{UserID: userID, UpdatedAt: time.Now()}
	}
	return ps.fromDBRecord(r)
}

// Save 将用户配置文件持久化到数据库
func (ps *ProfileStore) Save(p *UserProfile) {
	r := ps.toDBRecord(p)
	db.UpsertUserStyleProfile(r)
}

// UpdateWithTask 使用从任务中提取的样式数据更新用户配置文件
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
