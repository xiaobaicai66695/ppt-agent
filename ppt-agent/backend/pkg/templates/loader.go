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
	TypeAtomic TemplateType = "atomic"
)

// LayoutInfo 原子布局信息
type LayoutInfo struct {
	Name            string          `json:"name"`
	DisplayName     string          `json:"display_name"`
	Type            TemplateType    `json:"type"`
	Description     string          `json:"description"`
	AllowedPalettes []string        `json:"allowed_palettes"`
	Fields          []Field         `json:"fields"`
	Contract        *LayoutContract `json:"contract,omitempty"`
}

// LayoutContract 描述布局的内容容量和使用边界，供规划器和编排页消费。
type LayoutContract struct {
	Capacity         map[string]any `json:"capacity,omitempty"`
	RequiredFields   []string       `json:"required_fields,omitempty"`
	BestFor          []string       `json:"best_for,omitempty"`
	AvoidFor         []string       `json:"avoid_for,omitempty"`
	OverflowStrategy string         `json:"overflow_strategy,omitempty"`
	BackgroundPolicy string         `json:"background_policy,omitempty"`
	VisualPrimitives []string       `json:"visual_primitives,omitempty"`
}

// Field 布局字段定义
type Field struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
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

type componentContractsFile struct {
	ContentTypes map[string]struct {
		BestFor               []string       `json:"best_for"`
		RecommendedComponents []string       `json:"recommended_components"`
		Capacity              map[string]any `json:"capacity"`
		Variants              []string       `json:"variants"`
		DeckRule              string         `json:"deck_rule"`
	} `json:"content_types"`
}

// Loader 模板加载器
type Loader struct {
	contractsPath string
	layouts       []LayoutInfo
	themes        []ThemeInfo
}

// NewComponentLoader creates the current component-first loader from a skill root.
func NewComponentLoader(skillRoot string) *Loader {
	l := &Loader{
		contractsPath: filepath.Join(skillRoot, "templates", "component_contracts.json"),
	}
	l.load()
	return l
}

func (l *Loader) load() {
	l.layouts = l.loadComponentContractLayouts()
	l.themes = l.loadThemes()
}

func (l *Loader) loadComponentContractLayouts() []LayoutInfo {
	data, err := os.ReadFile(l.contractsPath)
	if err != nil {
		return builtInLayouts()
	}
	var contracts componentContractsFile
	if err := json.Unmarshal(data, &contracts); err != nil {
		return builtInLayouts()
	}
	result := make([]LayoutInfo, 0, len(contracts.ContentTypes))
	for name, spec := range contracts.ContentTypes {
		displayName := displayNameForContentType(name)
		description := strings.Join(spec.BestFor, "、")
		if description == "" {
			description = displayName
		}
		contract := &LayoutContract{
			Capacity:         spec.Capacity,
			RequiredFields:   []string{"title"},
			BestFor:          spec.BestFor,
			OverflowStrategy: "split_slide",
			VisualPrimitives: spec.RecommendedComponents,
		}
		if len(spec.Variants) > 0 {
			contract.VisualPrimitives = append(contract.VisualPrimitives, "layout_variant:"+strings.Join(spec.Variants, "|"))
		}
		result = append(result, LayoutInfo{
			Name:        name,
			DisplayName: displayName,
			Type:        TypeAtomic,
			Description: description,
			Fields: []Field{
				{Name: "title", Label: "标题", Type: "text", Required: true},
				{Name: "description", Label: "内容描述", Type: "textarea", Required: false},
				{Name: "content_plan", Label: "组件计划", Type: "json", Required: false},
			},
			Contract: contract,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayName < result[j].DisplayName
	})
	if len(result) == 0 {
		return builtInLayouts()
	}
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

// ListLayouts 返回所有原子布局
func (l *Loader) ListLayouts() []LayoutInfo {
	return l.layouts
}

// ListThemes 返回所有配色方案
func (l *Loader) ListThemes() []ThemeInfo {
	return l.themes
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

func builtInLayouts() []LayoutInfo {
	names := []string{
		"title_slide", "agenda", "section_divider", "content_slide", "image_text", "card_grid",
		"three_column", "icon_grid", "two_column", "timeline", "process_flow", "stat_slide",
		"kpi_dashboard", "chart_slide", "comparison_table", "quote_slide", "image_hero",
		"example_detail", "case_study", "deep_dive", "swot_analysis", "kanban", "brand_focus",
		"region_map", "summary_slide",
	}
	result := make([]LayoutInfo, 0, len(names))
	for _, name := range names {
		result = append(result, LayoutInfo{
			Name:        name,
			DisplayName: displayNameForContentType(name),
			Type:        TypeAtomic,
			Description: displayNameForContentType(name),
			Fields: []Field{
				{Name: "title", Label: "标题", Type: "text", Required: true},
				{Name: "description", Label: "内容描述", Type: "textarea", Required: false},
			},
			Contract: &LayoutContract{RequiredFields: []string{"title"}},
		})
	}
	return result
}

func displayNameForContentType(name string) string {
	labels := map[string]string{
		"title_slide":      "封面",
		"agenda":           "目录",
		"section_divider":  "章节页",
		"content_slide":    "内容页",
		"image_text":       "图文页",
		"card_grid":        "卡片矩阵",
		"three_column":     "三栏并列",
		"icon_grid":        "图标网格",
		"two_column":       "双栏对比",
		"timeline":         "时间线",
		"process_flow":     "流程图",
		"stat_slide":       "关键数字",
		"kpi_dashboard":    "指标看板",
		"chart_slide":      "图表页",
		"comparison_table": "对比表",
		"quote_slide":      "引用页",
		"image_hero":       "视觉页",
		"example_detail":   "实例详解",
		"case_study":       "案例研究",
		"deep_dive":        "深入分析",
		"swot_analysis":    "SWOT 分析",
		"kanban":           "看板",
		"brand_focus":      "品牌聚焦",
		"region_map":       "区域地图",
		"summary_slide":    "总结页",
	}
	if label, ok := labels[name]; ok {
		return label
	}
	return name
}
