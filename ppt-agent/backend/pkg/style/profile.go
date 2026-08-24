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
	Use3D          bool     `json:"use_3d"`          // 是否使用3D效果
	ColorScheme    string   `json:"color_scheme"`    // 配色方案
}

// BrandPreferences 品牌元素偏好
type BrandPreferences struct {
	LogoPosition  string `json:"logo_position"`  // Logo位置: top-left, bottom-right, etc.
	ShowFooter    bool   `json:"show_footer"`    // 是否显示页脚
	FooterText    string `json:"footer_text"`    // 页脚文字
	UseWatermark  bool   `json:"use_watermark"`  // 是否使用水印
	WatermarkText string `json:"watermark_text"` // 水印文字
}

// UserFacts 是用户画像中的确定性资料，适合直接作为当前任务上下文。
// 这些字段不是风格偏好，不需要按历史领域门控。
type UserFacts struct {
	DisplayName  string `json:"display_name"` // 用户姓名或常用称呼
	Organization string `json:"organization"` // 工作单位、学校或组织
	Department   string `json:"department"`   // 部门、学院或团队
	JobTitle     string `json:"job_title"`    // 职位或身份
	Industry     string `json:"industry"`     // 行业或业务领域
	Location     string `json:"location"`     // 常驻地区或服务区域
}

func (f UserFacts) IsEmpty() bool {
	return strings.TrimSpace(f.DisplayName) == "" &&
		strings.TrimSpace(f.Organization) == "" &&
		strings.TrimSpace(f.Department) == "" &&
		strings.TrimSpace(f.JobTitle) == "" &&
		strings.TrimSpace(f.Industry) == "" &&
		strings.TrimSpace(f.Location) == ""
}

func (f UserFacts) PromptLines() []string {
	fields := []struct {
		label string
		value string
	}{
		{"姓名/称呼", f.DisplayName},
		{"工作单位/组织", f.Organization},
		{"部门/团队", f.Department},
		{"职位/身份", f.JobTitle},
		{"行业/业务领域", f.Industry},
		{"地区", f.Location},
	}
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		if value := cleanFactValue(field.value); value != "" {
			lines = append(lines, fmt.Sprintf("- %s: %s", field.label, value))
		}
	}
	return lines
}

// ContentTone 内容语调
type ContentTone struct {
	Formality   string `json:"formality"`    // 正式程度: formal, semi-formal, casual
	TechDensity int    `json:"tech_density"` // 技术术语密度 1-10
	DetailLevel int    `json:"detail_level"` // 详细程度 1-10
	HumorLevel  int    `json:"humor_level"`  // 幽默程度 1-10
}

// SuccessPattern 是历史兼容字段，仅用于只读推荐。
// 新链路不再从任务完成结果自动写入成功模式。
type SuccessPattern struct {
	Domain          string  `json:"domain"`            // 应用领域
	Template        string  `json:"template"`          // 成功使用的模板
	Theme           string  `json:"theme"`             // 成功使用的配色
	PageCount       int     `json:"page_count"`        // 成功使用的页数
	AvgQualityScore float64 `json:"avg_quality_score"` // 平均质量评分
	SuccessCount    int     `json:"success_count"`     // 成功次数
}

// EnhancedProfile 增强版用户画像
type EnhancedProfile struct {
	UserProfile // 嵌入现有字段

	// 新增偏好维度
	BrandElements    BrandPreferences `json:"brand_elements"`
	ChartPreferences ChartPreference  `json:"chart_preferences"`
	AnimationLevel   AnimationLevel   `json:"animation_level"`
	ContentTone      ContentTone      `json:"content_tone"`
	SuccessPatterns  []SuccessPattern `json:"success_patterns"`

	// 历史统计，保留用于兼容已存储画像，不再由任务完成自动更新。
	FirstTaskTime  time.Time `json:"first_task_time"`
	LastActiveTime time.Time `json:"last_active_time"`
	TotalTasks     int       `json:"total_tasks"`
	SuccessRate    float64   `json:"success_rate"` // 任务成功率

	// 领域偏好，作为用户显式画像/历史兼容数据读取，不再自动学习。
	DomainPreferences map[string]int `json:"domain_preferences"` // 领域 -> 次数
}

// NewEnhancedProfile 创建增强版用户画像
func NewEnhancedProfile(userID int) *EnhancedProfile {
	return &EnhancedProfile{
		UserProfile: UserProfile{
			UserID:    userID,
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
		"title_slide":     {"tech-intro", "pitch-deck", "course-module"},
		"content_slide":   {"tech-intro", "weekly-report"},
		"section_divider": {"tech-intro", "course-module"},
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

// GetPreferredTemplatesForDomain returns scene-sensitive template preferences
// only when the current task domain has matching historical evidence.
func (p *EnhancedProfile) GetPreferredTemplatesForDomain(domain string) []string {
	domain = normalizeDomainName(domain)
	if domain == "" {
		return nil
	}

	templates := make(map[string]int)
	for _, sp := range p.SuccessPatterns {
		if !sameDomain(domain, sp.Domain) || strings.TrimSpace(sp.Template) == "" {
			continue
		}
		templates[sp.Template] += maxInt(sp.SuccessCount, 1)
	}

	if len(templates) == 0 && p.hasExactDomainHistory(domain) {
		contentTypeToTemplate := map[string][]string{
			"title_slide":     {"tech-intro", "pitch-deck", "course-module"},
			"content_slide":   {"tech-intro", "weekly-report"},
			"section_divider": {"tech-intro", "course-module"},
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
	}

	return sortedTemplateNames(templates)
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

// GetPreferredThemeForDomain returns a scene-sensitive theme only for exact
// domain matches. Cross-domain colors/themes are intentionally not migrated.
func (p *EnhancedProfile) GetPreferredThemeForDomain(domain string) string {
	domain = normalizeDomainName(domain)
	if domain == "" {
		return ""
	}

	themeCounts := make(map[string]int)
	for _, sp := range p.SuccessPatterns {
		if !sameDomain(domain, sp.Domain) || strings.TrimSpace(sp.Theme) == "" {
			continue
		}
		themeCounts[sp.Theme] += maxInt(sp.SuccessCount, 1)
	}
	if len(themeCounts) > 0 {
		return sortedTemplateNames(themeCounts)[0]
	}

	if p.hasExactDomainHistory(domain) && len(p.PreferredThemes) > 0 {
		return p.PreferredThemes[0]
	}
	return ""
}

// HasExactDomainHistory reports whether the profile contains reliable history
// for the current domain. It is used to gate scene-sensitive preferences.
func (p *EnhancedProfile) HasExactDomainHistory(domain string) bool {
	return p.hasExactDomainHistory(domain)
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

	if p.hasExactDomainHistory(req.Domain) {
		result.Tips = append(result.Tips, "已找到同领域历史偏好，推荐仅作为弱参考")
	} else if hasSceneSensitivePreferenceFields(p) {
		result.Tips = append(result.Tips, "未使用跨领域模板、配色和布局偏好")
	}

	return result
}

func (p *EnhancedProfile) suggestTemplate(domain, complexity string) string {
	domainTemplates := map[string][]string{
		"business":   {"pitch-deck", "product-launch"},
		"technical":  {"tech-sharing", "tech-intro"},
		"academic":   {"course-module", "design-defense"},
		"government": {"politics-ideology", "current-affairs"},
		"personal":   {"personal-summary", "weekly-report"},
		"creative":   {"activity-plan", "product-launch"},
	}

	if templates, ok := domainTemplates[domain]; ok {
		return templates[0]
	}
	return "tech-intro"
}

func (p *EnhancedProfile) suggestTheme(domain string) string {
	if theme := p.GetPreferredThemeForDomain(domain); theme != "" {
		return theme
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

func hasSceneSensitivePreferenceFields(p *EnhancedProfile) bool {
	if p == nil {
		return false
	}
	return len(p.PreferredThemes) > 0 ||
		len(p.PreferredColors) > 0 ||
		len(p.LayoutPreferences) > 0 ||
		len(p.SuccessPatterns) > 0 ||
		len(p.SpecialNotes) > 0
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

func (p *EnhancedProfile) hasExactDomainHistory(domain string) bool {
	domain = normalizeDomainName(domain)
	if domain == "" {
		return false
	}
	for d, count := range p.DomainPreferences {
		if count > 0 && sameDomain(domain, d) {
			return true
		}
	}
	for _, sp := range p.SuccessPatterns {
		if sameDomain(domain, sp.Domain) && sp.SuccessCount > 0 {
			return true
		}
	}
	return false
}

func sortedTemplateNames(counts map[string]int) []string {
	type templateCount struct {
		template string
		count    int
	}
	var tc []templateCount
	for t, c := range counts {
		if strings.TrimSpace(t) == "" || c <= 0 {
			continue
		}
		tc = append(tc, templateCount{t, c})
	}
	sort.Slice(tc, func(i, j int) bool { return tc[i].count > tc[j].count })

	result := make([]string, 0, len(tc))
	for _, t := range tc {
		result = append(result, t.template)
	}
	return result
}

func normalizeDomainName(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "unknown" {
		return ""
	}
	switch domain {
	case "tech":
		return "technical"
	case "education":
		return "academic"
	case "politics":
		return "government"
	case "art":
		return "creative"
	default:
		return domain
	}
}

func sameDomain(a, b string) bool {
	return normalizeDomainName(a) != "" && normalizeDomainName(a) == normalizeDomainName(b)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// UserProfile 表示用户显式维护的偏好和确定性资料。
// 生成结果、任务输出和 QA 评分不再自动写回该画像。
type UserProfile struct {
	UserID            int              `json:"user_id"`
	PreferredThemes   []string         `json:"preferred_themes"`
	PreferredColors   []string         `json:"preferred_colors"`
	ContentPatterns   []string         `json:"content_patterns"`
	LayoutPreferences []string         `json:"layout_preferences"`
	LanguageTone      string           `json:"language_tone"`
	TypicalPageCount  int              `json:"typical_page_count"`
	ContentTypes      ContentTypeCount `json:"content_types"`
	SpecialNotes      []string         `json:"special_notes"`
	UserFacts         UserFacts        `json:"user_facts"`
	TaskCount         int              `json:"task_count"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

func cleanFactValue(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) > 120 {
		return string(runes[:120])
	}
	return value
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
		enhanced.UserFacts = extPrefs.UserFacts
	}

	return enhanced
}

// SaveEnhanced 保存增强版用户画像
func (s *EnhancedProfileStore) SaveEnhanced(p *EnhancedProfile) {
	// 保存基础画像
	s.ProfileStore.Save(&p.UserProfile)

	// 保存扩展偏好
	extPrefs := &extendedPreferences{
		BrandElements:     p.BrandElements,
		ChartPreferences:  p.ChartPreferences,
		AnimationLevel:    p.AnimationLevel,
		ContentTone:       p.ContentTone,
		SuccessPatterns:   p.SuccessPatterns,
		SuccessRate:       p.SuccessRate,
		DomainPreferences: p.DomainPreferences,
		UserFacts:         p.UserFacts,
	}
	s.saveExtendedPreferences(p.UserID, extPrefs)
}

// extendedPreferences 扩展偏好存储结构
type extendedPreferences struct {
	BrandElements     BrandPreferences `json:"brand_elements"`
	ChartPreferences  ChartPreference  `json:"chart_preferences"`
	AnimationLevel    AnimationLevel   `json:"animation_level"`
	ContentTone       ContentTone      `json:"content_tone"`
	SuccessPatterns   []SuccessPattern `json:"success_patterns"`
	SuccessRate       float64          `json:"success_rate"`
	DomainPreferences map[string]int   `json:"domain_preferences"`
	UserFacts         UserFacts        `json:"user_facts"`
}

// loadExtendedPreferences 从数据库加载扩展偏好
func (s *EnhancedProfileStore) loadExtendedPreferences(userID int) *extendedPreferences {
	r, err := db.GetUserStyleProfile(uint(userID))
	if err != nil || r == nil {
		return nil
	}

	ext := parseExtendedPreferences(r.ExtendedPreferences)
	if ext == nil {
		ext = &extendedPreferences{}
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
	extended := ps.extendedPreferencesForSave(p)

	return &db.UserStyleProfile{
		UserID:              uint(p.UserID),
		PreferredThemes:     string(themes),
		PreferredColors:     string(colors),
		ContentPatterns:     string(patterns),
		LanguageTone:        p.LanguageTone,
		TypicalPageCount:    p.TypicalPageCount,
		ContentTypes:        string(contentTypes),
		SpecialNotes:        string(notes),
		ExtendedPreferences: extended,
		TaskCount:           p.TaskCount,
		UpdatedAt:           time.Now(),
	}
}

func (ps *ProfileStore) fromDBRecord(r *db.UserStyleProfile) *UserProfile {
	p := &UserProfile{
		UserID:           int(r.UserID),
		LanguageTone:     r.LanguageTone,
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
	if ext := parseExtendedPreferences(r.ExtendedPreferences); ext != nil {
		p.UserFacts = ext.UserFacts
	}

	return p
}

func (ps *ProfileStore) extendedPreferencesForSave(p *UserProfile) string {
	ext := &extendedPreferences{}
	if existing, err := db.GetUserStyleProfile(uint(p.UserID)); err == nil && existing != nil {
		if loaded := parseExtendedPreferences(existing.ExtendedPreferences); loaded != nil {
			ext = loaded
		}
	}
	ext.UserFacts = p.UserFacts
	data, err := json.Marshal(ext)
	if err != nil {
		logger.Warn("profile_marshal_failed", "field", "ExtendedPreferences", "error", err.Error())
		return ""
	}
	return string(data)
}

func parseExtendedPreferences(raw string) *extendedPreferences {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var ext extendedPreferences
	if err := json.Unmarshal([]byte(raw), &ext); err != nil {
		logger.Warn("profile_unmarshal_failed", "field", "ExtendedPreferences", "error", err.Error())
		return nil
	}
	return &ext
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
