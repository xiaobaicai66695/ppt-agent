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
	"unicode"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/schema"
)

type Step struct {
	Index       int    `json:"index"`
	Title       string `json:"title"`
	ContentType string `json:"content_type"`
	Description string `json:"description"`
	Desc        string `json:"desc"`
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
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
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
							Type:     schema.String,
							Desc:     "内容类型",
							Required: false,
						},
						"description": {
							Type:     schema.String,
							Desc:     "幻灯片内容描述",
							Required: true,
						},
					},
				},
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
	fileName := GetStepFileName(step)
	filePath := filepath.Join(workDir, fileName)
	// 构建请求
	request := fmt.Sprintf(`创建第%d个幻灯片
任务详情：
- 页码：%d
- 标题：%s
- 内容类型：%s
- 内容描述：%s

【重要】输出文件：
- 文件名：%s
- 完整路径：%s`, step.Index, step.Index, step.Title, step.ContentType, step.Description, fileName, filePath)
	return request
}

// FormatBatchSlidesForRequest 将一批幻灯片格式化为 CodeAgent 请求
func FormatBatchSlidesForRequest(slides []Step, batchNum, totalBatches int, workDir string) string {
	if len(slides) == 0 {
		return "[完成] 该批次幻灯片都已生成完毕。"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("【批量生成任务 - 第 %d/%d 批】请生成以下幻灯片：\n\n", batchNum, totalBatches))

	for i, slide := range slides {
		fileName := GetStepFileName(&slide)

		sb.WriteString(fmt.Sprintf("## 幻灯片 %d/%d (批内序号: %d)\n", slide.Index, len(slides), i+1))
		sb.WriteString(fmt.Sprintf("- 页码：%d\n", slide.Index))
		sb.WriteString(fmt.Sprintf("- 标题：%s\n", slide.Title))
		sb.WriteString(fmt.Sprintf("- 内容类型：%s\n", slide.ContentType))
		sb.WriteString(fmt.Sprintf("- 内容描述：%s\n", slide.Description))
		sb.WriteString(fmt.Sprintf("- 输出文件：%s\n\n", fileName))
	}

	sb.WriteString("【重要】\n")
	sb.WriteString("- 必须为每个幻灯片生成独立的 .pptx 文件\n")
	sb.WriteString("- 文件命名格式：{页码}_{标题}.pptx\n")
	sb.WriteString("- 所有文件保存到工作目录\n")
	sb.WriteString("- 完成后列出所有生成的文件\n")

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
		fileName := GetStepFileName(&slide)

		sb.WriteString(fmt.Sprintf("## 幻灯片 %d/%d\n", i+1, len(slides)))
		sb.WriteString(fmt.Sprintf("- 页码：%d\n", slide.Index))
		sb.WriteString(fmt.Sprintf("- 标题：%s\n", slide.Title))
		sb.WriteString(fmt.Sprintf("- 内容类型：%s\n", slide.ContentType))
		sb.WriteString(fmt.Sprintf("- 内容描述：%s\n", slide.Description))
		sb.WriteString(fmt.Sprintf("- 输出文件：%s\n\n", fileName))
	}

	sb.WriteString("【重要】\n")
	sb.WriteString("- 必须为每个幻灯片生成独立的 .pptx 文件\n")
	sb.WriteString("- 文件命名格式：{页码}_{标题}.pptx\n")
	sb.WriteString("- 所有文件保存到工作目录\n")
	sb.WriteString("- 完成后列出所有生成的文件\n")

	return sb.String()
}
