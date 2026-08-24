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

// SlideInfo 默认幻灯片信息
type SlideInfo struct {
	Title       string `json:"title"`
	ContentType string `json:"content_type"`
	Description string `json:"description"`
	Background  string `json:"background"`
}

// BackgroundOptions 背景选项配置
type BackgroundOptions struct {
	Themes []string `json:"themes"`
	Labels []string `json:"labels"`
}

// TemplateInfo 预设模板信息
type TemplateInfo struct {
	Name           string             `json:"name"`
	DisplayName    string             `json:"display_name"`
	Type           TemplateType       `json:"type"`
	Description    string             `json:"description"`
	Category       string             `json:"category"`
	DefaultPalette string             `json:"default_palette"`
	Tags           []string           `json:"tags"`
	Thumbnail      string             `json:"thumbnail"`
	SlideCount     int                `json:"slide_count"`
	DefaultSlides  []SlideInfo        `json:"default_slides"`
	BackgroundOpts *BackgroundOptions `json:"background_options,omitempty"`
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

// BackgroundThemeInfo is kept only for legacy API compatibility. New planning
// uses external image search metadata instead of local background themes.
type BackgroundThemeInfo struct {
	Name               string   `json:"name"`
	DisplayName        string   `json:"display_name"`
	Description        string   `json:"description"`
	Scenarios          []string `json:"scenarios"`
	RecommendedPalette string   `json:"recommended_palette,omitempty"`
	PreviewPath        string   `json:"preview_path"`
}

type backgroundManifest struct {
	Themes []struct {
		ID                 string   `json:"id"`
		NameCN             string   `json:"name_cn"`
		Description        string   `json:"description"`
		Scenarios          []string `json:"scenarios"`
		Priority           int      `json:"priority"`
		RecommendedPalette string   `json:"recommended_palette"`
	} `json:"themes"`
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
	presetsDir     string
	layoutsDir     string
	bgTemplatesDir string
	contractsPath  string
	presets        []TemplateInfo
	layouts        []LayoutInfo
	themes         []ThemeInfo
	backgrounds    []BackgroundThemeInfo
}

// NewLoader 创建模板加载器
func NewLoader(presetsDir, layoutsDir, bgTemplatesDir string) *Loader {
	l := &Loader{
		presetsDir:     presetsDir,
		layoutsDir:     layoutsDir,
		bgTemplatesDir: bgTemplatesDir,
		contractsPath:  filepath.Join(filepath.Dir(layoutsDir), "component_contracts.json"),
	}
	l.load()
	return l
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
	l.presets = l.loadPresets()
	l.layouts = l.loadLayouts()
	l.themes = l.loadThemes()
	l.backgrounds = l.loadBackgrounds()
}

func (l *Loader) loadPresets() []TemplateInfo {
	var result []TemplateInfo
	entries, err := os.ReadDir(l.presetsDir)
	if err != nil {
		return builtInPresets()
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
	if len(result) == 0 {
		return builtInPresets()
	}
	return result
}

func (l *Loader) loadLayouts() []LayoutInfo {
	var result []LayoutInfo
	entries, err := os.ReadDir(l.layoutsDir)
	if err != nil {
		return l.loadComponentContractLayouts()
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
	if len(result) == 0 {
		return l.loadComponentContractLayouts()
	}
	return result
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

// GetPresetByCategory 返回按类别筛选的模板（不区分大小写）
func (l *Loader) GetPresetByCategory(category string) []TemplateInfo {
	var result []TemplateInfo
	for i := range l.presets {
		if strings.EqualFold(l.presets[i].Category, category) {
			result = append(result, l.presets[i])
		}
	}
	return result
}

// ListBackgrounds returns legacy local background themes. The component-first
// loader intentionally returns none unless an old loader is constructed.
func (l *Loader) ListBackgrounds() []BackgroundThemeInfo {
	return l.backgrounds
}

// GetBackgroundTemplatesDir returns the legacy background template directory.
// Component-first runtime does not read this path.
func (l *Loader) GetBackgroundTemplatesDir() string {
	return l.bgTemplatesDir
}

func (l *Loader) loadBackgrounds() []BackgroundThemeInfo {
	return l.loadManifestBackgrounds()
}

func (l *Loader) loadManifestBackgrounds() []BackgroundThemeInfo {
	if strings.TrimSpace(l.bgTemplatesDir) == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(l.bgTemplatesDir, "manifest.json"))
	if err != nil {
		return nil
	}
	var manifest backgroundManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil
	}
	type rankedBackground struct {
		BackgroundThemeInfo
		priority int
	}
	var ranked []rankedBackground
	for _, theme := range manifest.Themes {
		name := strings.TrimSpace(theme.ID)
		if name == "" {
			continue
		}
		displayName := strings.TrimSpace(theme.NameCN)
		if displayName == "" {
			displayName = name
		}
		ranked = append(ranked, rankedBackground{
			BackgroundThemeInfo: BackgroundThemeInfo{
				Name:               name,
				DisplayName:        displayName,
				Description:        strings.TrimSpace(theme.Description),
				Scenarios:          theme.Scenarios,
				RecommendedPalette: strings.TrimSpace(theme.RecommendedPalette),
				PreviewPath:        "/api/backgrounds/" + name + "/preview",
			},
			priority: theme.Priority,
		})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].priority == ranked[j].priority {
			return ranked[i].Name < ranked[j].Name
		}
		return ranked[i].priority > ranked[j].priority
	})
	backgrounds := make([]BackgroundThemeInfo, 0, len(ranked))
	for _, item := range ranked {
		backgrounds = append(backgrounds, item.BackgroundThemeInfo)
	}
	return backgrounds
}

type builtInPresetSpec struct {
	name        string
	displayName string
	description string
	category    string
	palette     string
	tags        []string
	slideCount  int
}

var builtInPresetSpecs = []builtInPresetSpec{
	{"short-class-talk", "短讲分享", "5-10 分钟短时分享，适合少页高密度表达。", "edu", "education_blue", []string{"短讲", "课堂", "分享"}, 6},
	{"meeting-minutes", "会议纪要", "会议记录、项目评审和行动项同步，强调结论与责任。", "biz", "charcoal_light", []string{"会议", "行动项", "复盘"}, 8},
	{"weekly-report", "周期汇报", "团队周报、月报和项目进展汇报，适合状态、问题、计划结构。", "biz", "report_green", []string{"汇报", "进展", "数据"}, 9},
	{"activity-plan", "活动方案", "活动策划、校园活动和执行安排，适合目标、流程、资源和风险。", "creative", "activity_orange", []string{"活动", "策划", "执行"}, 10},
	{"personal-summary", "个人总结", "述职、年终总结和个人成长复盘，强调成果与反思。", "personal", "charcoal_light", []string{"述职", "总结", "成果"}, 10},
	{"product-intro", "产品介绍", "产品说明、客户演示和功能展示，突出场景、价值和信任。", "biz", "ocean_soft", []string{"产品", "功能", "演示"}, 12},
	{"project-proposal", "项目方案", "立项申请、资源申请和方案论证，强调目标、路径和收益。", "biz", "charcoal_light", []string{"项目", "方案", "资源"}, 12},
	{"design-defense", "答辩汇报", "课程设计、毕业设计和项目答辩，强调问题、方法、实现和结论。", "academic", "debate_purple", []string{"答辩", "设计", "项目"}, 12},
	{"product-launch", "产品发布", "发布会、宣讲和产品路演，强调价值主张与差异化。", "creative", "berry_cream", []string{"发布", "产品", "路线图"}, 14},
	{"research-report", "调研报告", "市场调研、行业分析和可行性研究，适合数据、证据和结论。", "biz", "report_green", []string{"调研", "分析", "数据"}, 14},
	{"current-affairs", "时政解读", "政策、时事和公共议题分析，适合背景、数据、影响和建议。", "gov", "patriotic_blue", []string{"政策", "时政", "公共"}, 14},
	{"pitch-deck", "路演方案", "商业计划、投融资路演和创业项目展示，强调说服力。", "biz", "finance_gold", []string{"路演", "商业计划", "融资"}, 16},
	{"politics-ideology", "思政教育", "思政、团课、党政和价值观教育，适合庄重叙事。", "gov", "government_red", []string{"思政", "党政", "教育"}, 16},
	{"innovation-compete", "科创竞赛", "大创、挑战杯和互联网+汇报，强调创新性和可行性。", "academic", "civic_gold", []string{"竞赛", "创新", "项目"}, 16},
	{"training-course", "培训课程", "培训、新人入职和技能课程，适合章节化教学。", "edu", "education_blue", []string{"培训", "课程", "教学"}, 16},
	{"course-module", "课程模块", "教学课件、知识讲解和系统化学习材料。", "edu", "sage_calm", []string{"课程", "教学", "知识"}, 17},
	{"tech-intro", "技术介绍", "新技术介绍、行业科普和知识分享，从概念到应用展开。", "tech", "ocean_soft", []string{"技术", "科普", "介绍"}, 18},
	{"tech-sharing", "技术分享", "内部技术分享、架构讲解和工程实践，注重深度和案例。", "tech", "ocean_soft", []string{"技术", "架构", "实践"}, 18},
	{"generic", "通用规划", "通用兜底模板，由 Planner 根据主题动态规划章节和组件。", "general", "ocean_soft", []string{"通用", "自适应"}, 12},
}

func builtInPresets() []TemplateInfo {
	result := make([]TemplateInfo, 0, len(builtInPresetSpecs))
	for _, spec := range builtInPresetSpecs {
		result = append(result, TemplateInfo{
			Name:           spec.name,
			DisplayName:    spec.displayName,
			Type:           TypePreset,
			Description:    spec.description,
			Category:       spec.category,
			DefaultPalette: spec.palette,
			Tags:           append([]string(nil), spec.tags...),
			Thumbnail:      "/templates/thumbs/" + spec.name + ".jpg",
			SlideCount:     spec.slideCount,
			DefaultSlides:  defaultSlides(spec.slideCount),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SlideCount < result[j].SlideCount
	})
	return result
}

func defaultSlides(count int) []SlideInfo {
	if count <= 0 {
		count = 8
	}
	plan := []SlideInfo{
		{Title: "主题开场", ContentType: "title_slide", Description: "建立演示主题、受众语境和核心看点。"},
		{Title: "内容概览", ContentType: "agenda", Description: "列出主要章节和汇报路线。"},
		{Title: "背景与问题", ContentType: "section_divider", Description: "开启背景章节，说明为什么需要讨论该主题。"},
		{Title: "关键背景", ContentType: "image_text", Description: "用图文方式说明主题背景、场景和代表事实。"},
		{Title: "核心观点", ContentType: "content_slide", Description: "围绕主观点展开事实、解释和结论。"},
		{Title: "证据与案例", ContentType: "case_study", Description: "用命名案例或事实证据支撑前述判断。"},
		{Title: "数据支撑", ContentType: "chart_slide", Description: "用结构化数据、图表和来源说明关键趋势。"},
		{Title: "方案路径", ContentType: "process_flow", Description: "说明方法、流程或行动路径。"},
		{Title: "能力矩阵", ContentType: "card_grid", Description: "以卡片组织并列能力、要点或模块。"},
		{Title: "对比判断", ContentType: "comparison_table", Description: "从多个维度对比方案、对象或阶段，并给出推荐。"},
		{Title: "深入分析", ContentType: "deep_dive", Description: "展开原理、机制、架构或复杂论证。"},
		{Title: "关键指标", ContentType: "kpi_dashboard", Description: "呈现关键指标、变化和业务含义。"},
		{Title: "风险机会", ContentType: "two_column", Description: "并列说明机会、风险、约束和应对。"},
		{Title: "阶段计划", ContentType: "timeline", Description: "按时间或阶段说明推进节奏。"},
		{Title: "行动建议", ContentType: "content_slide", Description: "给出可执行建议和优先级。"},
		{Title: "总结展望", ContentType: "summary_slide", Description: "收束核心结论、后续行动和展望。"},
		{Title: "补充案例", ContentType: "example_detail", Description: "拆解一个具体案例、背景、做法和启示。"},
		{Title: "结束页", ContentType: "quote_slide", Description: "用核心观点或引用形成收尾。"},
	}
	if count > len(plan) {
		count = len(plan)
	}
	return append([]SlideInfo(nil), plan[:count]...)
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
