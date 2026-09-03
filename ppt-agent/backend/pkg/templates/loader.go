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
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Type        TemplateType    `json:"type"`
	Description string          `json:"description"`
	Fields      []Field         `json:"fields"`
	Contract    *LayoutContract `json:"contract,omitempty"`
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

// ListLayouts 返回所有原子布局
func (l *Loader) ListLayouts() []LayoutInfo {
	return l.layouts
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
		"timeline", "kpi_dashboard", "chart_slide", "comparison_table", "quote_slide",
		"swot_analysis", "kanban", "brand_focus",
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
		"timeline":         "时间线",
		"kpi_dashboard":    "指标看板",
		"chart_slide":      "图表页",
		"comparison_table": "对比表",
		"quote_slide":      "引用页",
		"swot_analysis":    "SWOT 分析",
		"kanban":           "看板",
		"brand_focus":      "品牌聚焦",
	}
	if label, ok := labels[name]; ok {
		return label
	}
	return name
}
