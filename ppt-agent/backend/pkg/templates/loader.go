package templates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TemplateType 模板类型
type TemplateType string

const (
	TypePreset TemplateType = "preset"
	TypeAtomic TemplateType = "atomic"
)

// LayoutInfo 原子布局信息
type LayoutInfo struct {
	Name             string   `json:"name"`
	DisplayName      string   `json:"display_name"`
	Type             TemplateType `json:"type"`
	Description      string   `json:"description"`
	AllowedPalettes  []string `json:"allowed_palettes"`
	Fields           []Field  `json:"fields"`
}

// Field 布局字段定义
type Field struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// SlideInfo 默认幻灯片信息
type SlideInfo struct {
	Title       string `json:"title"`
	ContentType string `json:"content_type"`
	Description string `json:"description"`
}

// TemplateInfo 预设模板信息
type TemplateInfo struct {
	Name           string      `json:"name"`
	DisplayName   string      `json:"display_name"`
	Type           TemplateType `json:"type"`
	Description   string      `json:"description"`
	Category      string      `json:"category"`
	DefaultPalette string     `json:"default_palette"`
	Tags          []string    `json:"tags"`
	Thumbnail     string      `json:"thumbnail"`
	SlideCount    int         `json:"slide_count"`
	DefaultSlides []SlideInfo `json:"default_slides"`
}

// ThemeInfo 配色方案信息
type ThemeInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Primary     string   `json:"primary"`
	Secondary   string   `json:"secondary"`
	Accent      string   `json:"accent"`
	Background  string   `json:"background"`
	Tags        []string `json:"tags"`
}

// Loader 模板加载器
type Loader struct {
	presetsDir string
	layoutsDir string
	presets    []TemplateInfo
	layouts    []LayoutInfo
	themes     []ThemeInfo
}

// NewLoader 创建模板加载器
func NewLoader(presetsDir, layoutsDir string) *Loader {
	l := &Loader{
		presetsDir: presetsDir,
		layoutsDir: layoutsDir,
	}
	l.load()
	return l
}

func (l *Loader) load() {
	l.presets = l.loadPresets()
	l.layouts = l.loadLayouts()
	l.themes = l.loadThemes()
}

func (l *Loader) loadPresets() []TemplateInfo {
	var result []TemplateInfo
	entries, err := os.ReadDir(l.presetsDir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(l.presetsDir, entry.Name()))
		if err != nil {
			continue
		}
		var t TemplateInfo
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		t.Type = TypePreset
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SlideCount < result[j].SlideCount
	})
	return result
}

func (l *Loader) loadLayouts() []LayoutInfo {
	var result []LayoutInfo
	entries, err := os.ReadDir(l.layoutsDir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(l.layoutsDir, entry.Name()))
		if err != nil {
			continue
		}
		var t LayoutInfo
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		t.Type = TypeAtomic
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayName < result[j].DisplayName
	})
	return result
}

func (l *Loader) loadThemes() []ThemeInfo {
	return []ThemeInfo{
		{Name: "ocean_soft", DisplayName: "海洋蓝", Primary: "#0891b2", Secondary: "#06b6d4", Accent: "#22d3ee", Background: "#f0f9ff", Tags: []string{"清新", "专业"}},
		{Name: "sage_calm", DisplayName: "森系绿", Primary: "#16a34a", Secondary: "#22c55e", Accent: "#4ade80", Background: "#f0fdf4", Tags: []string{"自然", "健康"}},
		{Name: "charcoal_light", DisplayName: "商务灰", Primary: "#475569", Secondary: "#64748b", Accent: "#94a3b8", Background: "#f8fafc", Tags: []string{"商务", "正式"}},
		{Name: "government_red", DisplayName: "党政红", Primary: "#dc2626", Secondary: "#ef4444", Accent: "#f87171", Background: "#fef2f2", Tags: []string{"党政", "红色"}},
		{Name: "patriotic_blue", DisplayName: "爱国蓝", Primary: "#1d4ed8", Secondary: "#3b82f6", Accent: "#60a5fa", Background: "#eff6ff", Tags: []string{"爱国", "庄重"}},
		{Name: "warm_terracotta", DisplayName: "活力橙", Primary: "#ea580c", Secondary: "#f97316", Accent: "#fb923c", Background: "#fff7ed", Tags: []string{"活力", "创意"}},
		{Name: "berry_cream", DisplayName: "优雅紫", Primary: "#7c3aed", Secondary: "#8b5cf6", Accent: "#a78bfa", Background: "#faf5ff", Tags: []string{"优雅", "时尚"}},
		{Name: "lavender_mist", DisplayName: "梦幻紫", Primary: "#9333ea", Secondary: "#a855f7", Accent: "#c084fc", Background: "#faf5ff", Tags: []string{"梦幻", "柔和"}},
		{Name: "civic_gold", DisplayName: "金色典雅", Primary: "#ca8a04", Secondary: "#eab308", Accent: "#facc15", Background: "#fefce8", Tags: []string{"金贵", "典雅"}},
		{Name: "debate_purple", DisplayName: "辩论紫", Primary: "#6d28d9", Secondary: "#7c3aed", Accent: "#8b5cf6", Background: "#f5f3ff", Tags: []string{"学术", "辩论"}},
		{Name: "activity_orange", DisplayName: "活动橙", Primary: "#c2410c", Secondary: "#ea580c", Accent: "#f97316", Background: "#fff7ed", Tags: []string{"活动", "活泼"}},
		{Name: "report_green", DisplayName: "汇报绿", Primary: "#15803d", Secondary: "#16a34a", Accent: "#22c55e", Background: "#f0fdf4", Tags: []string{"汇报", "数据"}},
		{Name: "simple_gray", DisplayName: "简约灰", Primary: "#374151", Secondary: "#4b5563", Accent: "#6b7280", Background: "#ffffff", Tags: []string{"简约", "通用"}},
		{Name: "medical_blue", DisplayName: "医疗蓝", Primary: "#0284c7", Secondary: "#0ea5e9", Accent: "#38bdf8", Background: "#f0f9ff", Tags: []string{"医疗", "健康"}},
		{Name: "finance_gold", DisplayName: "金融金", Primary: "#b45309", Secondary: "#d97706", Accent: "#f59e0b", Background: "#fffbeb", Tags: []string{"金融", "专业"}},
		{Name: "education_blue", DisplayName: "教育蓝", Primary: "#1e40af", Secondary: "#2563eb", Accent: "#3b82f6", Background: "#eff6ff", Tags: []string{"教育", "学习"}},
	}
}

// ListPresets 返回所有预设模板
func (l *Loader) ListPresets() []TemplateInfo {
	return l.presets
}

// ListLayouts 返回所有原子布局
func (l *Loader) ListLayouts() []LayoutInfo {
	return l.layouts
}

// ListThemes 返回所有配色方案
func (l *Loader) ListThemes() []ThemeInfo {
	return l.themes
}

// GetPreset 根据名称获取预设模板
func (l *Loader) GetPreset(name string) *TemplateInfo {
	for i := range l.presets {
		if l.presets[i].Name == name {
			return &l.presets[i]
		}
	}
	return nil
}

// GetLayout 根据名称获取原子布局
func (l *Loader) GetLayout(name string) *LayoutInfo {
	for i := range l.layouts {
		if l.layouts[i].Name == name {
			return &l.layouts[i]
		}
	}
	return nil
}

// GetPresetByCategory 返回指定分类的模板
func (l *Loader) GetPresetByCategory(category string) []TemplateInfo {
	var result []TemplateInfo
	for i := range l.presets {
		if strings.EqualFold(l.presets[i].Category, category) {
			result = append(result, l.presets[i])
		}
	}
	return result
}
