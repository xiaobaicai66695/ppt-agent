package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/schema"
)

type ContentElement struct {
	Type        string `json:"type"`                  // bullet_list, numbered_list, example_box, key_point_card, image_placeholder, table, chart_placeholder, callout, quote
	Items       []string `json:"items,omitempty"`     // 用于 bullet_list / numbered_list
	Text        string   `json:"text,omitempty"`      // 用于 callout / quote
	Title       string   `json:"title,omitempty"`     // 用于 example_box / key_point_card
	Description string   `json:"description,omitempty"` // 用于 example_box
	LayoutHint  string   `json:"layout_hint,omitempty"` // 布局提示：left-image, right-image, top-title, center
}

// ContentPlan 描述页面内部的内容结构，是 description 的结构化升级
type ContentPlan struct {
	Summary  string           `json:"summary,omitempty"`  // 页面核心内容的一句话概括
	Elements []ContentElement `json:"elements,omitempty"` // 内容元素列表
}

type Step struct {
	Index       int      `json:"index"`
	Title       string   `json:"title"`
	ContentType string   `json:"content_type"`
	Description string   `json:"description"`
	Desc        string   `json:"desc"`

	// SubSteps 分页组：若设置此字段，表示该幻灯片需要分多页生成。
	// SubSteps 中的每个子项对应一页幻灯片，页码依次递增。
	// 最多嵌套一层（SubSteps 中不能再有 SubSteps）。
	SubSteps []SubStep `json:"sub_steps,omitempty"`

	// ContentPlan 内容结构化描述（可选，Planner 可选择性填充）
	ContentPlan *ContentPlan `json:"content_plan,omitempty"`

	// LayoutHint 布局提示，补充 content_type 的不足
	LayoutHint string `json:"layout_hint,omitempty"`

	// SlideKey 内部字段，用于格式化输出文件名和 QA 查找（如 "4"、"4.1"）
	SlideKey string `json:"-"`
}

// SubStep 分页组的子页面，与 Step 结构相同但不再支持嵌套分页
type SubStep struct {
	Index       int           `json:"index"`
	Title       string        `json:"title"`
	ContentType string        `json:"content_type"`
	Description string        `json:"description"`
	ContentPlan *ContentPlan  `json:"content_plan,omitempty"`
	LayoutHint  string        `json:"layout_hint,omitempty"`
}

type Plan struct {
	Title  string  `json:"title"`
	Theme  string  `json:"theme"`
	Slides []Step  `json:"slides"`
	Steps  []Step  `json:"steps"`
}

func (p *Plan) FirstStep() string {
	steps := p.Slides
	if len(steps) == 0 {
		steps = p.Steps
	}
	if len(steps) == 0 {
		return ""
	}
	stepStr, _ := json.Marshal(steps[0])
	return string(stepStr)
}

func (p *Plan) GetSlides() []Step {
	if len(p.Slides) > 0 {
		return p.Slides
	}
	return p.Steps
}

// GetRemainingSlides 获取剩余未执行的幻灯片
// executedSteps 是已执行步骤的 JSON 字符串列表
func (p *Plan) GetRemainingSlides(executedSteps []string) []Step {
	allSlides := p.GetSlides()
	if len(allSlides) == 0 {
		return nil
	}

	executedIndexes := make(map[int]bool)
	for _, stepJSON := range executedSteps {
		var step Step
		if err := json.Unmarshal([]byte(stepJSON), &step); err == nil {
			executedIndexes[step.Index] = true
		}
	}

	var remaining []Step
	for _, slide := range allSlides {
		if !executedIndexes[slide.Index] {
			remaining = append(remaining, slide)
		}
	}
	return remaining
}

func (p *Plan) MarshalJSON() ([]byte, error) {
	type Alias Plan
	return json.Marshal((*Alias)(p))
}

func (p *Plan) UnmarshalJSON(bytes []byte) error {
	type Alias Plan
	a := (*Alias)(p)

	// 清理 JSON 字符串，去除 markdown 代码块格式和反引号包裹
	cleaned := strings.TrimSpace(string(bytes))

	// 去除 markdown 代码块标记 ```json\n...\n``` 或 ```\n...\n```
	cleaned = stripPlanFence(cleaned)

	// 去除首尾的反引号
	cleaned = strings.TrimPrefix(cleaned, "`")
	cleaned = strings.TrimSuffix(cleaned, "`")
	cleaned = strings.TrimSpace(cleaned)

	return json.Unmarshal([]byte(cleaned), a)
}

// stripPlanFence 去除 markdown 代码块包裹，处理各种格式
func stripPlanFence(s string) string {
	// 处理多行代码块：```xxx\n...\n```
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) >= 2 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "```") {
			// 有语言标识的多行格式
			remaining := lines[1]
			// 去掉末尾的 ```
			for strings.HasSuffix(remaining, "\n") {
				remaining = strings.TrimSuffix(remaining, "\n")
			}
			if strings.HasSuffix(remaining, "```") {
				remaining = strings.TrimSuffix(remaining, "```")
				remaining = strings.TrimSpace(remaining)
				return remaining
			}
		}
	}

	// 处理单行格式：```{...}```
	if strings.HasPrefix(s, "```") && strings.HasSuffix(s, "```") && !strings.Contains(s[3:len(s)-3], "```") {
		s = s[3 : len(s)-3]
		s = strings.TrimSpace(s)
		// 去除语言标识
		for _, prefix := range []string{"json", "python", "shell"} {
			if strings.HasPrefix(s, prefix) {
				s = strings.TrimPrefix(s, prefix)
				s = strings.TrimSpace(s)
				break
			}
		}
	}

	return s
}

// contentPlanElementSubParams 是 content_plan.elements 中单个元素的子参数字段定义
var contentPlanElementSubParams = map[string]*schema.ParameterInfo{
	"type": {
		Type:    schema.String,
		Desc:    "元素类型：bullet_list(要点列表), numbered_list(编号列表), example_box(案例框), key_point_card(核心论断卡片), image_placeholder(配图占位), table(表格), chart_placeholder(图表占位), callout(突出引用), quote(金句引用)",
		Required: false,
	},
	"items": {
		Type:    schema.Array,
		ElemInfo: &schema.ParameterInfo{Type: schema.String},
		Desc:    "列表项（用于 bullet_list / numbered_list）",
		Required: false,
	},
	"text": {
		Type:    schema.String,
		Desc:    "文本内容（用于 callout / quote）",
		Required: false,
	},
	"title": {
		Type:    schema.String,
		Desc:    "标题（用于 example_box / key_point_card）",
		Required: false,
	},
	"description": {
		Type:    schema.String,
		Desc:    "详细描述（用于 example_box / key_point_card）",
		Required: false,
	},
	"layout_hint": {
		Type:    schema.String,
		Desc:    "布局提示：left-image, right-image, top-title, center",
		Required: false,
	},
}

// subStepSubParams 是 sub_steps 中单个子页面的参数字段定义
var subStepSubParams = map[string]*schema.ParameterInfo{
	"index": {
		Type:    schema.Integer,
		Desc:    "子页面序号，从1开始",
		Required: true,
	},
	"title": {
		Type:    schema.String,
		Desc:    "子页面标题",
		Required: true,
	},
	"content_type": {
		Type:    schema.String,
		Desc:    "内容类型：content_slide, two_column, image_text 等",
		Required: false,
	},
	"description": {
		Type:    schema.String,
		Desc:    "子页面内容描述",
		Required: true,
	},
	"content_plan": {
		Type:    schema.Object,
		Desc:    "子页面的内容结构化描述（可选）",
		Required: false,
		SubParams: map[string]*schema.ParameterInfo{
			"summary": {
				Type:    schema.String,
				Desc:    "页面核心内容的一句话概括",
				Required: false,
			},
			"elements": {
				Type:    schema.Array,
				Desc:    "内容元素列表",
				Required: false,
				ElemInfo: &schema.ParameterInfo{
					Type:        schema.Object,
					SubParams:   contentPlanElementSubParams,
					Desc: "单个内容元素",
				},
			},
		},
	},
	"layout_hint": {
		Type:    schema.String,
		Desc:    "布局补充提示",
		Required: false,
	},
}

// slideSubParams 是 slides 数组中每个幻灯片的参数字段定义
var slideSubParams = map[string]*schema.ParameterInfo{
	"index": {
		Type:     schema.Integer,
		Desc:     "幻灯片序号，从1开始",
		Required: true,
	},
	"title": {
		Type:     schema.String,
		Desc:     "幻灯片标题",
		Required: true,
	},
	"content_type": {
		Type:    schema.String,
		Desc:    "内容类型：title_slide(标题页), content_slide(自由内容页), two_column(双栏对比), three_column(三栏多要点), image_text(图文混排), quote_slide(引用金句页), chart_slide(数据图表页), section_divider(分隔页), summary_slide(总结页)，也可留空或填 custom_layout 由 Executor 自行决定",
		Required: false,
	},
	"description": {
		Type:     schema.String,
		Desc:     "幻灯片内容描述",
		Required: true,
	},
	"sub_steps": {
		Type:    schema.Array,
		Desc:    "分页组：若设置此字段，表示该幻灯片需要分多页生成。每个子项对应一页，页码依次递增。最多嵌套一层（不能再有 sub_steps）。",
		Required: false,
		ElemInfo: &schema.ParameterInfo{
			Type:        schema.Object,
			SubParams:   subStepSubParams,
			Desc: "分页组的子页面",
		},
	},
	"content_plan": {
		Type:    schema.Object,
		Desc:    "内容结构化描述（可选），用于描述单页内的元素结构",
		Required: false,
		SubParams: map[string]*schema.ParameterInfo{
			"summary": {
				Type:    schema.String,
				Desc:    "页面核心内容的一句话概括",
				Required: false,
			},
			"elements": {
				Type:    schema.Array,
				Desc:    "内容元素列表",
				Required: false,
				ElemInfo: &schema.ParameterInfo{
					Type:        schema.Object,
					SubParams:   contentPlanElementSubParams,
					Desc: "单个内容元素",
				},
			},
		},
	},
	"layout_hint": {
		Type:    schema.String,
		Desc:    "布局补充提示：left-image, right-image, top-title, center 等",
		Required: false,
	},
}

var PlanToolInfo = &schema.ToolInfo{
	Name: "create_ppt_plan",
	Desc: "创建PPT制作计划，将任务分解为多个幻灯片步骤。",
	ParamsOneOf: schema.NewParamsOneOfByParams(
		map[string]*schema.ParameterInfo{
			"title": {
				Type:     schema.String,
				Desc:     "PPT的标题",
				Required: true,
			},
			"theme": {
				Type:     schema.String,
				Desc:     "PPT主题风格",
				Required: false,
			},
			"slides": {
				Type:     schema.Array,
				ElemInfo: &schema.ParameterInfo{Type: schema.Object, SubParams: slideSubParams, Desc: "幻灯片列表"},
				Desc:     "幻灯片列表",
				Required: true,
			},
		},
	),
}

// FullPlan 完整的计划结构
type FullPlan struct {
	TaskID     int           `json:"task_id,omitempty"`
	Status     PlanStatus    `json:"status,omitempty"`
	AgentName  string        `json:"agent_name,omitempty"`
	Desc       string        `json:"desc,omitempty"`
	ExecResult *SubmitResult `json:"exec_result,omitempty"`
}

// PlanStatus 计划状态
type PlanStatus string

const (
	PlanStatusTodo    PlanStatus = "todo"
	PlanStatusDoing   PlanStatus = "doing"
	PlanStatusDone    PlanStatus = "done"
	PlanStatusFailed  PlanStatus = "failed"
	PlanStatusSkipped PlanStatus = "skipped"
)

var PlanStatusMapping = map[PlanStatus]string{
	PlanStatusTodo:    "待执行",
	PlanStatusDoing:   "执行中",
	PlanStatusDone:    "已完成",
	PlanStatusFailed:  "执行失败",
	PlanStatusSkipped: "已跳过",
}

func (p *FullPlan) String() string {
	status, ok := PlanStatusMapping[p.Status]
	if !ok {
		status = string(p.Status)
	}
	res := fmt.Sprintf("%d. **[%s]** %s", p.TaskID, status, p.Desc)
	if p.ExecResult != nil {
		res += fmt.Sprintf("\n%s", p.ExecResult.String())
	}
	return res
}

func (p *FullPlan) PlanString(n int) string {
	if p.Status != PlanStatusDoing && p.Status != PlanStatusTodo {
		return fmt.Sprintf("- [x] %d. %s", n, p.Desc)
	}
	return fmt.Sprintf("- [ ] %d. %s", n, p.Desc)
}

func FullPlan2String(plan []*FullPlan) string {
	planStr := "### PPT 制作计划\n\n"
	for i, p := range plan {
		planStr += p.PlanString(i+1) + "\n"
	}
	return planStr
}

func Write2PlanMD(ctx context.Context, op commandline.Operator, wd string, plan []*FullPlan) error {
	planStr := FullPlan2String(plan)
	filePath := filepath.Join(wd, "plan.md")
	return op.WriteFile(ctx, filePath, planStr)
}

// SubmitResult 提交结果
type SubmitResult struct {
	IsSuccess *bool               `json:"is_success,omitempty"`
	Result    string              `json:"result,omitempty"`
	Files     []*SubmitResultFile `json:"files,omitempty"`
}

// SubmitResultFile 提交结果文件
type SubmitResultFile struct {
	Path string `json:"path,omitempty"`
	Desc string `json:"desc,omitempty"`
}

func (s *SubmitResult) String() string {
	res := fmt.Sprintf("### 执行结果\n%s", s.Result)
	if len(s.Files) > 0 {
		res += "\n#### 生成的文件"
	}
	for _, f := range s.Files {
		res += fmt.Sprintf("\n- 描述：%s, 路径：%s", f.Desc, f.Path)
	}
	return res
}

func ListDir(dir string) ([]*SubmitResultFile, error) {
	var resp []*SubmitResultFile

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if path == dir {
			return nil
		}
		if d.IsDir() {
			next := filepath.Join(dir, d.Name())
			nextResp, err := ListDir(next)
			if err != nil {
				return err
			}
			resp = append(resp, nextResp...)
			return nil
		}
		resp = append(resp, &SubmitResultFile{
			Path: filepath.Join(filepath.Dir(dir), d.Name()),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SanitizeFilename 将标题转换为安全的文件名
// 规则：保留字母数字汉字，空格转下划线，移除特殊字符
func SanitizeFilename(title string) string {
	var result strings.Builder
	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			result.WriteRune(r)
		} else if unicode.IsSpace(r) {
			result.WriteRune('_')
		} else if unicode.Is(unicode.Han, r) {
			result.WriteRune(r) // 保留中文
		} else if r >= 0x4E00 && r <= 0x9FFF {
			result.WriteRune(r) // 中文Unicode范围
		}
		// 其他字符忽略
	}
	s := result.String()
	// 移除连续下划线
	re := regexp.MustCompile(`_+`)
	s = re.ReplaceAllString(s, "_")
	// 移除首尾下划线
	s = strings.Trim(s, "_")
	return s
}

// GetStepFileName 生成标准的幻灯片文件名：{页码}_{标题}.pptx
func GetStepFileName(step *Step) string {
	title := SanitizeFilename(step.Title)
	return fmt.Sprintf("%d_%s.pptx", step.Index, title)
}

// StepExists 检查指定页码的文件是否已存在
func StepExists(workDir string, pageIndex int) bool {
	pattern := fmt.Sprintf("%d_*.pptx", pageIndex)
	matches, _ := filepath.Glob(filepath.Join(workDir, pattern))
	return len(matches) > 0
}

// GetExistingStepFiles 获取已存在的幻灯片文件
func GetExistingStepFiles(workDir string) map[int]string {
	result := make(map[int]string)
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pptx") {
			continue
		}
		// 文件名格式：{页码}_{标题}.pptx
		name := strings.TrimSuffix(entry.Name(), ".pptx")
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 1 {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(parts[0], "%d", &idx); err == nil {
			result[idx] = entry.Name()
		}
	}
	return result
}

// FormatStepForRequest 将 Step 格式化为 CodeAgent 请求字符串
// 包含文件名约束指令
func FormatStepForRequest(step *Step, workDir string) string {
	// 若存在 SubSteps（分页组），则递归展开每个子页面
	if len(step.SubSteps) > 0 {
		var sb strings.Builder
		for i, sub := range step.SubSteps {
			if i > 0 {
				sb.WriteString("\n\n---\n\n")
			}
			fileName := GetSubStepFileName(step.Index, sub.Index, sub.Title)
			filePath := filepath.Join(workDir, fileName)

			var contentPlanStr string
			if sub.ContentPlan != nil {
				data, _ := json.Marshal(sub.ContentPlan)
				contentPlanStr = "\n- 内容结构规划（content_plan）：" + string(data)
			}
			var layoutHintStr string
			if sub.LayoutHint != "" {
				layoutHintStr = "\n- 布局提示：" + sub.LayoutHint
			}

			sb.WriteString(fmt.Sprintf(`创建第%d页（第%d/%d子页）
任务详情：
|- 页码：%d
|- 标题：%s
|- 内容类型：%s
|- 内容描述：%s%s%s

【重要】输出文件：
|- 文件名：%s
|- 完整路径：%s`,
				step.Index, sub.Index, len(step.SubSteps),
				step.Index, sub.Title, sub.ContentType, sub.Description,
				contentPlanStr, layoutHintStr,
				fileName, filePath))
		}
		return sb.String()
	}

	// 普通单页：如果有 slideKey 则用 slideKey 生成文件名
	var fileName string
	if step.SlideKey != "" {
		safeTitle := SanitizeFilename(step.Title)
		fileName = fmt.Sprintf("%s_%s.pptx", step.SlideKey, safeTitle)
	} else {
		fileName = GetStepFileName(step)
	}
	filePath := filepath.Join(workDir, fileName)

	var contentPlanStr string
	if step.ContentPlan != nil {
		data, _ := json.Marshal(step.ContentPlan)
		contentPlanStr = "\n- 内容结构规划（content_plan）：" + string(data)
	}

	var layoutHintStr string
	if step.LayoutHint != "" {
		layoutHintStr = "\n- 布局提示：" + step.LayoutHint
	}

	request := fmt.Sprintf(`创建幻灯片
任务详情：
|- 页码：%d
|- 标题：%s
|- 内容类型：%s
|- 内容描述：%s%s%s

【重要】输出文件：
|- 文件名：%s
|- 完整路径：%s`, step.Index, step.Title, step.ContentType, step.Description, contentPlanStr, layoutHintStr, fileName, filePath)
	return request
}

// GetSubStepFileName 生成分页组子页的文件名：{页码}.{子页码}_{标题}.pptx
func GetSubStepFileName(parentIdx, subIdx int, title string) string {
	safeTitle := SanitizeFilename(title)
	return fmt.Sprintf("%d.%d_%s.pptx", parentIdx, subIdx, safeTitle)
}

// FormatBatchSlidesForRequest 将一批幻灯片格式化为 CodeAgent 请求
func FormatBatchSlidesForRequest(slides []Step, batchNum, totalBatches int, workDir string) string {
	if len(slides) == 0 {
		return "[完成] 该批次幻灯片都已生成完毕。"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("【批量生成任务 - 第 %d/%d 批】请生成以下幻灯片：\n\n", batchNum, totalBatches))

	for i, slide := range slides {
		// 若存在分页组 SubSteps，递归展开
		if len(slide.SubSteps) > 0 {
			for j, sub := range slide.SubSteps {
				fileName := GetSubStepFileName(slide.Index, sub.Index, sub.Title)

				var contentPlanStr string
				if sub.ContentPlan != nil {
					data, _ := json.Marshal(sub.ContentPlan)
					contentPlanStr = fmt.Sprintf("\n- 内容结构规划（content_plan）：%s", string(data))
				}
				var layoutHintStr string
				if sub.LayoutHint != "" {
					layoutHintStr = fmt.Sprintf("\n- 布局提示：%s", sub.LayoutHint)
				}

				if i > 0 || j > 0 {
					sb.WriteString("\n\n")
				}
				sb.WriteString(fmt.Sprintf("## 幻灯片 %d.%d/%d\n", slide.Index, sub.Index, len(slide.SubSteps)))
				sb.WriteString(fmt.Sprintf("- 页码：%d（子页 %d）\n", slide.Index, sub.Index))
				sb.WriteString(fmt.Sprintf("- 标题：%s\n", sub.Title))
				sb.WriteString(fmt.Sprintf("- 内容类型：%s\n", sub.ContentType))
				sb.WriteString(fmt.Sprintf("- 内容描述：%s%s%s\n", sub.Description, contentPlanStr, layoutHintStr))
				sb.WriteString(fmt.Sprintf("- 输出文件：%s\n", fileName))
			}
			continue
		}

		// 普通单页
		fileName := GetStepFileName(&slide)

		var contentPlanStr string
		if slide.ContentPlan != nil {
			data, _ := json.Marshal(slide.ContentPlan)
			contentPlanStr = fmt.Sprintf("\n- 内容结构规划（content_plan）：%s", string(data))
		}
		var layoutHintStr string
		if slide.LayoutHint != "" {
			layoutHintStr = fmt.Sprintf("\n- 布局提示：%s", slide.LayoutHint)
		}
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("## 幻灯片 %d/%d\n", slide.Index, len(slides)))
		sb.WriteString(fmt.Sprintf("- 页码：%d\n", slide.Index))
		sb.WriteString(fmt.Sprintf("- 标题：%s\n", slide.Title))
		sb.WriteString(fmt.Sprintf("- 内容类型：%s\n", slide.ContentType))
		sb.WriteString(fmt.Sprintf("- 内容描述：%s%s%s\n", slide.Description, contentPlanStr, layoutHintStr))
		sb.WriteString(fmt.Sprintf("- 输出文件：%s\n", fileName))
	}

	sb.WriteString("\n【重要】\n")
	sb.WriteString("- 必须为每个幻灯片（包括分页组中的每个子页）生成独立的 .pptx 文件\n")
	sb.WriteString("- 分页组子页命名格式：{页码}.{子页码}_{标题}.pptx\n")
	sb.WriteString("- 普通页面命名格式：{页码}_{标题}.pptx\n")
	sb.WriteString("- 每生成一页后，必须调用 update_progress 和 single_qa_review\n")
	sb.WriteString("- 所有文件保存到工作目录\n")
	sb.WriteString("- 完成后列出所有生成的文件\n")

	return sb.String()
}

// FormatBatchStepsForRequest 将批量步骤渲染为请求字符串
// 一次性输出多页的详细信息，供 Executor 在批量模式下使用
func FormatBatchStepsForRequest(slides []Step, workDir string) string {
	if len(slides) == 0 {
		return "[完成] 该批次幻灯片都已生成完毕。"
	}

	var sb strings.Builder
	for i, slide := range slides {
		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}

		// 若存在分页组 SubSteps，递归展开
		if len(slide.SubSteps) > 0 {
			for j, sub := range slide.SubSteps {
				if j > 0 {
					sb.WriteString("\n\n---\n\n")
				}
				fileName := GetSubStepFileName(slide.Index, sub.Index, sub.Title)
				filePath := filepath.Join(workDir, fileName)

				var contentPlanStr string
				if sub.ContentPlan != nil {
					data, _ := json.Marshal(sub.ContentPlan)
					contentPlanStr = "\n- 内容结构规划（content_plan）：" + string(data)
				}
				var layoutHintStr string
				if sub.LayoutHint != "" {
					layoutHintStr = "\n- 布局提示：" + sub.LayoutHint
				}

				sb.WriteString(fmt.Sprintf(`创建幻灯片
任务详情：
|- 页码：%d（子页 %d）
|- 标题：%s
|- 内容类型：%s
|- 内容描述：%s%s%s

【重要】输出文件：
|- 文件名：%s
|- 完整路径：%s`,
					slide.Index, sub.Index,
					sub.Title, sub.ContentType, sub.Description,
					contentPlanStr, layoutHintStr,
					fileName, filePath))
			}
			continue
		}

		// 普通单页
		fileName := GetStepFileName(&slide)
		filePath := filepath.Join(workDir, fileName)

		var contentPlanStr string
		if slide.ContentPlan != nil {
			data, _ := json.Marshal(slide.ContentPlan)
			contentPlanStr = "\n- 内容结构规划（content_plan）：" + string(data)
		}
		var layoutHintStr string
		if slide.LayoutHint != "" {
			layoutHintStr = "\n- 布局提示：" + slide.LayoutHint
		}

		sb.WriteString(fmt.Sprintf(`创建幻灯片
任务详情：
|- 页码：%d
|- 标题：%s
|- 内容类型：%s
|- 内容描述：%s%s%s

【重要】输出文件：
|- 文件名：%s
|- 完整路径：%s`,
			slide.Index, slide.Title, slide.ContentType, slide.Description,
			contentPlanStr, layoutHintStr,
			fileName, filePath))
	}

	sb.WriteString("\n\n【批量任务要求】")
	sb.WriteString("\n- 以上为本次需要生成的幻灯片列表（含分页组子页）")
	sb.WriteString("\n- 分页组子页命名格式：{页码}.{子页码}_{标题}.pptx")
	sb.WriteString("\n- 普通页面命名格式：{页码}_{标题}.pptx")
	sb.WriteString("\n- 每生成一页后，必须调用 update_progress 和 single_qa_review")

	return sb.String()
}

// FormatAllSlidesForRequest 将所有剩余步骤格式化为 CodeAgent 批量请求
// 一次性传递所有幻灯片信息，支持批量生成
func FormatAllSlidesForRequest(slides []Step, workDir string) string {
	if len(slides) == 0 {
		return "[完成] 所有幻灯片都已生成完毕。"
	}

	var sb strings.Builder
	sb.WriteString("【批量生成任务】请一次性生成所有幻灯片：\n\n")

	for i, slide := range slides {
		// 若存在分页组 SubSteps，递归展开
		if len(slide.SubSteps) > 0 {
			for j, sub := range slide.SubSteps {
				if i > 0 || j > 0 {
					sb.WriteString("\n\n")
				}
				fileName := GetSubStepFileName(slide.Index, sub.Index, sub.Title)

				var contentPlanStr string
				if sub.ContentPlan != nil {
					data, _ := json.Marshal(sub.ContentPlan)
					contentPlanStr = fmt.Sprintf("\n- 内容结构规划（content_plan）：%s", string(data))
				}
				var layoutHintStr string
				if sub.LayoutHint != "" {
					layoutHintStr = fmt.Sprintf("\n- 布局提示：%s", sub.LayoutHint)
				}

				sb.WriteString(fmt.Sprintf("## 幻灯片 %d.%d/%d\n", slide.Index, sub.Index, len(slide.SubSteps)))
				sb.WriteString(fmt.Sprintf("- 页码：%d（子页 %d）\n", slide.Index, sub.Index))
				sb.WriteString(fmt.Sprintf("- 标题：%s\n", sub.Title))
				sb.WriteString(fmt.Sprintf("- 内容类型：%s\n", sub.ContentType))
				sb.WriteString(fmt.Sprintf("- 内容描述：%s%s%s\n", sub.Description, contentPlanStr, layoutHintStr))
				sb.WriteString(fmt.Sprintf("- 输出文件：%s\n", fileName))
			}
			continue
		}

		// 普通单页
		if i > 0 {
			sb.WriteString("\n\n")
		}
		fileName := GetStepFileName(&slide)

		var contentPlanStr string
		if slide.ContentPlan != nil {
			data, _ := json.Marshal(slide.ContentPlan)
			contentPlanStr = fmt.Sprintf("\n- 内容结构规划（content_plan）：%s", string(data))
		}
		var layoutHintStr string
		if slide.LayoutHint != "" {
			layoutHintStr = fmt.Sprintf("\n- 布局提示：%s", slide.LayoutHint)
		}

		sb.WriteString(fmt.Sprintf("## 幻灯片 %d/%d\n", i+1, len(slides)))
		sb.WriteString(fmt.Sprintf("- 页码：%d\n", slide.Index))
		sb.WriteString(fmt.Sprintf("- 标题：%s\n", slide.Title))
		sb.WriteString(fmt.Sprintf("- 内容类型：%s\n", slide.ContentType))
		sb.WriteString(fmt.Sprintf("- 内容描述：%s%s%s\n", slide.Description, contentPlanStr, layoutHintStr))
		sb.WriteString(fmt.Sprintf("- 输出文件：%s\n", fileName))
	}

	sb.WriteString("\n【重要】\n")
	sb.WriteString("- 必须为每个幻灯片（包括分页组子页）生成独立的 .pptx 文件\n")
	sb.WriteString("- 分页组子页命名格式：{页码}.{子页码}_{标题}.pptx\n")
	sb.WriteString("- 普通页面命名格式：{页码}_{标题}.pptx\n")
	sb.WriteString("- 每生成一页后，必须调用 update_progress 和 single_qa_review\n")
	sb.WriteString("- 所有文件保存到工作目录\n")
	sb.WriteString("- 完成后列出所有生成的文件\n")

	return sb.String()
}

// CheckpointFileName 是进度 checkpoint 文件名
const CheckpointFileName = ".slides_checkpoint.json"

// SlidesCheckpoint 存储已完成幻灯片的进度信息
type SlidesCheckpoint struct {
	CompletedSlides []string `json:"completed_slides"` // 已完成的页码标识列表（如 "4"、"4.1"）
	TotalSlides    int      `json:"total_slides"`    // 总幻灯片数
	LastUpdated    string   `json:"last_updated"`     // 最后更新时间
}

// SaveCheckpoint 保存进度到 checkpoint 文件
func SaveCheckpoint(workDir string, completedSlides []string, totalSlides int) error {
	checkpoint := SlidesCheckpoint{
		CompletedSlides: completedSlides,
		TotalSlides:    totalSlides,
		LastUpdated:    time.Now().Format("2006-01-02 15:04:05"),
	}
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return err
	}
	filePath := filepath.Join(workDir, CheckpointFileName)
	return os.WriteFile(filePath, data, 0644)
}

// LoadCheckpoint 从 checkpoint 文件加载进度
func LoadCheckpoint(workDir string) (*SlidesCheckpoint, error) {
	filePath := filepath.Join(workDir, CheckpointFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var checkpoint SlidesCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, err
	}
	return &checkpoint, nil
}

// AddCompletedSlide 向 checkpoint 添加一个已完成的幻灯片
func AddCompletedSlide(workDir string, slideKey string) error {
	checkpoint, err := LoadCheckpoint(workDir)
	if err != nil {
		return err
	}
	if checkpoint == nil {
		checkpoint = &SlidesCheckpoint{CompletedSlides: []string{}}
	}
	for _, k := range checkpoint.CompletedSlides {
		if k == slideKey {
			return nil
		}
	}
	checkpoint.CompletedSlides = append(checkpoint.CompletedSlides, slideKey)
	checkpoint.LastUpdated = time.Now().Format("2006-01-02 15:04:05")
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return err
	}
	filePath := filepath.Join(workDir, CheckpointFileName)
	return os.WriteFile(filePath, data, 0644)
}

// GetCompletedCountFromCheckpoint 从 checkpoint 获取已完成数量
func GetCompletedCountFromCheckpoint(workDir string) (int, error) {
	checkpoint, err := LoadCheckpoint(workDir)
	if err != nil {
		return 0, err
	}
	if checkpoint == nil {
		return 0, nil
	}
	return len(checkpoint.CompletedSlides), nil
}

// QAAttemptFileName 是 QA 尝试次数文件名
const QAAttemptFileName = ".qa_attempts.json"

// QAAttempts 记录每张幻灯片的 QA 尝试次数，key 为 PPTX 文件名
type QAAttempts struct {
	Attempts map[string]int `json:"attempts"` // pptx_filename -> attempt count
}

// LoadQAAttempts 加载 QA 尝试次数
func LoadQAAttempts(workDir string) (*QAAttempts, error) {
	filePath := filepath.Join(workDir, QAAttemptFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &QAAttempts{Attempts: make(map[string]int)}, nil
		}
		return nil, err
	}
	var attempts QAAttempts
	if err := json.Unmarshal(data, &attempts); err != nil {
		return nil, err
	}
	return &attempts, nil
}

// SaveQAAttempts 保存 QA 尝试次数
func SaveQAAttempts(workDir string, a *QAAttempts) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	filePath := filepath.Join(workDir, QAAttemptFileName)
	return os.WriteFile(filePath, data, 0644)
}

// GetQAAttempt 获取某页的 QA 尝试次数
func GetQAAttempt(workDir string, pptxFilename string) (int, error) {
	attempts, err := LoadQAAttempts(workDir)
	if err != nil {
		return 0, err
	}
	return attempts.Attempts[pptxFilename], nil
}

// IncrementQAAttempt 增加某页的 QA 尝试次数，返回增加后的值
func IncrementQAAttempt(workDir string, pptxFilename string) (int, error) {
	attempts, err := LoadQAAttempts(workDir)
	if err != nil {
		return 0, err
	}
	attempts.Attempts[pptxFilename]++
	if err := SaveQAAttempts(workDir, attempts); err != nil {
		return 0, err
	}
	return attempts.Attempts[pptxFilename], nil
}

// QAResultFileName 是 QA 结果文件名前缀
const QAResultFileName = ".qa_result.json"

// QAResult 是 QA 审查结果，直接使用自然语言报告，不再解析 JSON。
type QAResult struct {
	TotalSlides  int      `json:"total_slides"`
	Reports      []string `json:"reports"` // 每页的自然语言审查报告，格式为 "页码|报告内容"
	Summary      string   `json:"summary"` // 整体摘要
	HasIssues    bool     `json:"has_issues"`
	HasHighIssue bool     `json:"has_high_issue"`
	LastUpdated  string   `json:"last_updated"`
}

// SaveQAResult 将 QA 结果保存到文件
func SaveQAResult(workDir string, result *QAResult) error {
	result.LastUpdated = time.Now().Format("2006-01-02 15:04:05")
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	filePath := filepath.Join(workDir, QAResultFileName)
	return os.WriteFile(filePath, data, 0644)
}

// LoadQAResult 从文件加载 QA 结果
func LoadQAResult(workDir string) (*QAResult, error) {
	filePath := filepath.Join(workDir, QAResultFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result QAResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
