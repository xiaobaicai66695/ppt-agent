package templates

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	Version      int `json:"version"`
	ContentTypes map[string]struct {
		BestFor               []string       `json:"best_for"`
		RecommendedComponents []string       `json:"recommended_components"`
		Capacity              map[string]any `json:"capacity"`
		Variants              []string       `json:"variants"`
		PPTRule               string         `json:"ppt_rule"`
	} `json:"content_types"`
}

// ComponentCapacity is the machine-readable capacity contract for one page
// type. It is intentionally small so it can be injected into agent prompts
// and carried through context compression without loading the whole skill.
type ComponentCapacity struct {
	ContentType           string   `json:"content_type"`
	RecommendedMin        int      `json:"recommended_min"`
	RecommendedMax        int      `json:"recommended_max"`
	MaxComponents         int      `json:"max_components"`
	RecommendedComponents []string `json:"recommended_components,omitempty"`
}

// ContractSummary is the runtime projection of component_contracts.json.
// SHA256 identifies the exact contract used by the Planner and validators.
type ContractSummary struct {
	Version      int                          `json:"version"`
	SHA256       string                       `json:"sha256"`
	Source       string                       `json:"source"`
	ContentTypes map[string]ComponentCapacity `json:"content_types"`
}

// LoadComponentContract reads the single skill contract used by Go runtime
// validation, prompt injection and benchmark diagnostics.
func LoadComponentContract(skillRoot string) (*ContractSummary, error) {
	path := filepath.Join(skillRoot, "templates", "component_contracts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var document componentContractsFile
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse component contract: %w", err)
	}
	if len(document.ContentTypes) == 0 {
		return nil, fmt.Errorf("component contract has no content_types")
	}

	result := &ContractSummary{
		Version:      document.Version,
		SHA256:       sha256Hex(data),
		Source:       "file",
		ContentTypes: make(map[string]ComponentCapacity, len(document.ContentTypes)),
	}
	for name, spec := range document.ContentTypes {
		capacity := ComponentCapacity{
			ContentType:           name,
			RecommendedComponents: append([]string(nil), spec.RecommendedComponents...),
		}
		capacity.RecommendedMin = contractInt(spec.Capacity["target_components_min"])
		capacity.RecommendedMax = contractInt(spec.Capacity["target_components_max"])
		capacity.MaxComponents = contractInt(spec.Capacity["max_components"])
		if capacity.MaxComponents <= 0 {
			return nil, fmt.Errorf("content type %s has invalid max_components", name)
		}
		if capacity.RecommendedMin < 0 || capacity.RecommendedMin > capacity.RecommendedMax || capacity.RecommendedMax > capacity.MaxComponents {
			return nil, fmt.Errorf("content type %s has invalid component capacity", name)
		}
		result.ContentTypes[name] = capacity
	}
	return result, nil
}

// BuiltInContractSummary is a compatibility fallback for tests and callers
// that do not have a skill directory. Production callers should load the
// contract from disk and preserve its SHA256.
func BuiltInContractSummary() *ContractSummary {
	capacities := map[string]ComponentCapacity{
		"title_slide":      {ContentType: "title_slide", RecommendedMin: 0, RecommendedMax: 3, MaxComponents: 4},
		"agenda":           {ContentType: "agenda", RecommendedMin: 3, RecommendedMax: 6, MaxComponents: 6},
		"section_divider":  {ContentType: "section_divider", RecommendedMin: 1, RecommendedMax: 2, MaxComponents: 3},
		"content_slide":    {ContentType: "content_slide", RecommendedMin: 3, RecommendedMax: 6, MaxComponents: 8},
		"image_text":       {ContentType: "image_text", RecommendedMin: 3, RecommendedMax: 5, MaxComponents: 6},
		"card_grid":        {ContentType: "card_grid", RecommendedMin: 4, RecommendedMax: 6, MaxComponents: 8},
		"timeline":         {ContentType: "timeline", RecommendedMin: 3, RecommendedMax: 5, MaxComponents: 6},
		"kpi_dashboard":    {ContentType: "kpi_dashboard", RecommendedMin: 3, RecommendedMax: 4, MaxComponents: 4},
		"chart_slide":      {ContentType: "chart_slide", RecommendedMin: 2, RecommendedMax: 4, MaxComponents: 6},
		"comparison_table": {ContentType: "comparison_table", RecommendedMin: 2, RecommendedMax: 4, MaxComponents: 5},
		"quote_slide":      {ContentType: "quote_slide", RecommendedMin: 1, RecommendedMax: 2, MaxComponents: 3},
		"swot_analysis":    {ContentType: "swot_analysis", RecommendedMin: 4, RecommendedMax: 8, MaxComponents: 8},
		"kanban":           {ContentType: "kanban", RecommendedMin: 4, RecommendedMax: 8, MaxComponents: 10},
		"brand_focus":      {ContentType: "brand_focus", RecommendedMin: 4, RecommendedMax: 7, MaxComponents: 8},
	}
	return &ContractSummary{Version: 1, Source: "builtin-fallback", ContentTypes: capacities}
}

// LoadComponentContractOrFallback keeps legacy tests and standalone callers
// usable while making the fallback visible in diagnostics.
func LoadComponentContractOrFallback(skillRoot string) *ContractSummary {
	if contract, err := LoadComponentContract(skillRoot); err == nil {
		return contract
	}
	return BuiltInContractSummary()
}

func (c *ContractSummary) CapacityFor(contentType string) ComponentCapacity {
	if c != nil {
		if capacity, ok := c.ContentTypes[strings.TrimSpace(contentType)]; ok {
			return capacity
		}
	}
	return ComponentCapacity{ContentType: strings.TrimSpace(contentType), RecommendedMin: 0, RecommendedMax: 8, MaxComponents: 8}
}

// PromptText returns a compact, deterministic projection suitable for a
// system instruction or a compression handoff.
func (c *ContractSummary) PromptText() string {
	if c == nil {
		return "容量契约不可用；未知页面类型默认最多 8 个组件。"
	}
	names := make([]string, 0, len(c.ContentTypes))
	for name := range c.ContentTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	fmt.Fprintf(&b, "version=%d", c.Version)
	if c.SHA256 != "" {
		fmt.Fprintf(&b, ", sha256=%s", c.SHA256)
	}
	for _, name := range names {
		capacity := c.ContentTypes[name]
		fmt.Fprintf(&b, "\n- %s: recommended=%d-%d, max=%d", name, capacity.RecommendedMin, capacity.RecommendedMax, capacity.MaxComponents)
		if len(capacity.RecommendedComponents) > 0 {
			fmt.Fprintf(&b, ", recommended_components=%s", strings.Join(capacity.RecommendedComponents, ","))
		}
	}
	return b.String()
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func contractInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, _ := strconv.Atoi(typed.String())
		return parsed
	default:
		return 0
	}
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
