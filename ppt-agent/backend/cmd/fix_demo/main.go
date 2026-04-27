package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/generic"
)

// QAResult 是批量 QA 的审查结果（与 generic.QAResult 结构一致）
type QAResult struct {
	TotalSlides  int       `json:"total_slides"`
	Issues       []QAIssue `json:"issues"`
	Summary      string    `json:"summary"`
	HasIssues    bool      `json:"has_issues"`
	HasHighIssue bool      `json:"has_high_issue"`
	LastUpdated  string    `json:"last_updated"`
}

type QAIssue struct {
	Slide    int    `json:"slide"`
	Severity string `json:"severity"`
	Type     string `json:"type"`
	Desc     string `json:"description"`
	Fix      string `json:"recommendation"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/fix_demo <workDir>")
		fmt.Println("Example: go run ./cmd/fix_demo /ppt/ppt-agent/output/ea9c5c5f-bcd6-4f32-a6d3-034adaf46887")
		os.Exit(1)
	}
	workDir := os.Args[1]

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("QA Fix Demo - 读取 QA 结果并生成修复脚本")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("工作目录: %s\n\n", workDir)

	// Step 1: 读取 QA 结果
	qaPath := filepath.Join(workDir, ".qa_result.json")
	qaData, err := os.ReadFile(qaPath)
	if err != nil {
		fmt.Printf("错误: 无法读取 QA 结果文件: %v\n", err)
		os.Exit(1)
	}

	var qaResult QAResult
	if err := json.Unmarshal(qaData, &qaResult); err != nil {
		// 尝试使用 generic 包加载（支持中文引号等）
		qaResult2, err2 := generic.LoadQAResult(workDir)
		if err2 != nil {
			fmt.Printf("错误: 无法解析 QA 结果 JSON: %v\n", err)
			fmt.Printf("原始数据: %s\n", string(qaData))
			os.Exit(1)
		}
		qaResult = QAResult{
			TotalSlides:  qaResult2.TotalSlides,
			Issues:       make([]QAIssue, len(qaResult2.Issues)),
			Summary:      qaResult2.Summary,
			HasIssues:    qaResult2.HasIssues,
			HasHighIssue: qaResult2.HasHighIssue,
			LastUpdated:  qaResult2.LastUpdated,
		}
		for i, issue := range qaResult2.Issues {
			qaResult.Issues[i] = QAIssue{
				Slide:    issue.Slide,
				Severity: issue.Severity,
				Type:     issue.Type,
				Desc:     issue.Desc,
				Fix:      issue.Fix,
			}
		}
	}

	// Step 2: 列出所有 PPTX 文件
	pptxFiles, err := filepath.Glob(filepath.Join(workDir, "*.pptx"))
	if err != nil {
		fmt.Printf("错误: 无法列出 PPTX 文件: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("找到 %d 个 PPTX 文件:\n", len(pptxFiles))
	for _, f := range pptxFiles {
		info, _ := os.Stat(f)
		fmt.Printf("  - %s (%.1f KB)\n", filepath.Base(f), float64(info.Size())/1024)
	}
	fmt.Println()

	// Step 3: 打印 QA 结果摘要
	fmt.Printf("QA 审查结果:\n")
	fmt.Printf("  总页数: %d\n", qaResult.TotalSlides)
	fmt.Printf("  问题数: %d\n", len(qaResult.Issues))
	fmt.Printf("  有问题: %v\n", qaResult.HasIssues)
	fmt.Printf("  有高严重性问题: %v\n", qaResult.HasHighIssue)
	fmt.Printf("  摘要: %s\n\n", qaResult.Summary)

	// Step 4: 按严重性分组打印问题
	if len(qaResult.Issues) == 0 {
		fmt.Println("无问题，PPT 质量通过!")
		return
	}

	highIssues := []QAIssue{}
	mediumIssues := []QAIssue{}
	lowIssues := []QAIssue{}
	for _, issue := range qaResult.Issues {
		switch strings.ToLower(issue.Severity) {
		case "high":
			highIssues = append(highIssues, issue)
		case "medium":
			mediumIssues = append(mediumIssues, issue)
		default:
			lowIssues = append(lowIssues, issue)
		}
	}

	if len(highIssues) > 0 {
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("严重问题 (%d 个):\n", len(highIssues))
		for i, issue := range highIssues {
			fmt.Printf("  [%d] 第 %d 页 [%s] (%s)\n", i+1, issue.Slide, issue.Type, issue.Severity)
			fmt.Printf("      问题: %s\n", issue.Desc)
			fmt.Printf("      修复: %s\n", issue.Fix)
			fmt.Println()
		}
	}
	if len(mediumIssues) > 0 {
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("中等问题 (%d 个):\n", len(mediumIssues))
		for i, issue := range mediumIssues {
			fmt.Printf("  [%d] 第 %d 页 [%s] (%s)\n", i+1, issue.Slide, issue.Type, issue.Severity)
			fmt.Printf("      问题: %s\n", issue.Desc)
			fmt.Printf("      修复: %s\n", issue.Fix)
			fmt.Println()
		}
	}
	if len(lowIssues) > 0 {
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("低优先级问题 (%d 个):\n", len(lowIssues))
		for i, issue := range lowIssues {
			fmt.Printf("  [%d] 第 %d 页 [%s] (%s): %s\n", i+1, issue.Slide, issue.Type, issue.Severity, issue.Desc)
		}
		fmt.Println()
	}

	// Step 5: 为每个有问题的页生成修复脚本
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("生成修复脚本")
	fmt.Println(strings.Repeat("=", 70))

	// 找出有问题的页
	problemSlides := map[int][]QAIssue{}
	for _, issue := range qaResult.Issues {
		problemSlides[issue.Slide] = append(problemSlides[issue.Slide], issue)
	}

	// 找到每个问题页对应的 PPTX 文件
	// 策略: 按文件名匹配（第N页通常对应 数字_标题.pptx）
	for slideIdx, issues := range problemSlides {
		// 找最可能包含该页的 PPTX 文件
		var targetFile string
		for _, f := range pptxFiles {
			base := filepath.Base(f)
			// 文件名格式通常是 "N_标题.pptx"
			nameWithoutExt := strings.TrimSuffix(base, ".pptx")
			parts := strings.SplitN(nameWithoutExt, "_", 2)
			if len(parts) >= 1 {
				var fileSlideIdx int
				if _, err := fmt.Sscanf(parts[0], "%d", &fileSlideIdx); err == nil {
					if fileSlideIdx == slideIdx {
						targetFile = f
						break
					}
				}
			}
		}

		// 如果没精确匹配，尝试部分匹配或默认用第一个文件
		if targetFile == "" && len(pptxFiles) > 0 {
			targetFile = pptxFiles[0]
		}

		fmt.Printf("\n第 %d 页修复脚本 (%d 个问题):\n", slideIdx, len(issues))
		fmt.Printf("  目标文件: %s\n", filepath.Base(targetFile))

		// 生成 python-pptx 修复代码
		script := generateFixScript(slideIdx, targetFile, issues)
		fmt.Printf("  修复代码:\n%s\n", script)
	}

	// Step 6: 生成完整的汇总修复脚本
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("完整修复脚本 (修复所有问题)")
	fmt.Println(strings.Repeat("=", 70))

	fullScript := generateFullFixScript(workDir, qaResult.Issues)
	fmt.Println(fullScript)
}

// generateFixScript 为单个页面生成 python-pptx 修复脚本
func generateFixScript(slideIdx int, pptxPath string, issues []QAIssue) string {
	var fixInstructions []string
	for _, issue := range issues {
		fixInstructions = append(fixInstructions, fmt.Sprintf("    # [%s] %s", issue.Type, issue.Desc))
		fixInstructions = append(fixInstructions, fmt.Sprintf("    # 修复建议: %s", issue.Fix))
		fixInstructions = append(fixInstructions, "")
	}

	return fmt.Sprintf(`python3 << 'PYEOF'
from pptx import Presentation
from pptx.util import Inches, Pt, Emu
from pptx.dml.color import RGBColor
from pptx.enum.text import PP_ALIGN
import copy

prs = Presentation(r"%s")
slide = prs.slides[%d]

# 当前页面的问题及修复建议:
%s
# ---- 在此添加具体修复逻辑 ----

# 保存修复后的文件
output_path = r"%s"
prs.save(output_path)
print(f"已保存修复后的 PPT: {output_path}")
PYEOF`, filepath.Base(pptxPath), slideIdx-1, strings.Join(fixInstructions, "\n"),
		strings.ReplaceAll(pptxPath, ".pptx", "_fixed.pptx"))
}

// generateFullFixScript 生成一个完整的修复脚本
func generateFullFixScript(workDir string, issues []QAIssue) string {
	var issueBlocks []string
	for _, issue := range issues {
		issueBlocks = append(issueBlocks, fmt.Sprintf(`    {
        "slide": %d,
        "severity": "%s",
        "type": "%s",
        "description": "%s",
        "recommendation": "%s"
    }`, issue.Slide, issue.Severity, issue.Type,
			strings.ReplaceAll(issue.Desc, `"`, `\"`),
			strings.ReplaceAll(issue.Fix, `"`, `\"`)))
	}

	return fmt.Sprintf(`python3 << 'PYEOF'
"""
QA 修复脚本 - 基于视觉 QA 发现的问题自动修复 PPT
工作目录: %s
问题数: %d (高严重性: %d)
"""
from pptx import Presentation
from pptx.util import Inches, Pt, Emu
from pptx.dml.color import RGBColor
from pptx.enum.text import PP_ALIGN
import json
import copy
import os

work_dir = r"%s"

# QA 发现的问题列表
issues = [
%s
]

# 加载 PPTX 文件
pptx_files = sorted([f for f in os.listdir(work_dir) if f.endswith('.pptx')])
print(f"找到 {len(pptx_files)} 个 PPTX 文件: {pptx_files}")

for issue in issues:
    slide_idx = issue["slide"]
    severity = issue["severity"]
    issue_type = issue["type"]
    description = issue["description"]
    recommendation = issue["recommendation"]

    print(f"\n处理第 {slide_idx} 页 [{severity}] {issue_type}:")
    print(f"  问题: {description}")
    print(f"  修复: {recommendation}")

    # 找到包含该页的 PPTX 文件
    target_pptx = None
    for pptx_file in pptx_files:
        # 简单策略: 假设页码匹配文件名开头的数字
        name_part = pptx_file.split('_')[0]
        try:
            if int(name_part) == slide_idx:
                target_pptx = os.path.join(work_dir, pptx_file)
                break
        except ValueError:
            pass

    if target_pptx is None and pptx_files:
        # 默认用第一个文件
        target_pptx = os.path.join(work_dir, pptx_files[0])

    print(f"  目标文件: {target_pptx}")

    # ---- 在此处根据 issue_type 实现具体修复逻辑 ----
    # issue_type 可能的值:
    #   - contrast: 对比度问题 → 修改文字颜色
    #   - overflow: 内容溢出 → 调整布局或缩小字号
    #   - overlap: 元素重叠 → 调整元素位置
    #   - alignment: 对齐问题 → 调整对齐方式
    #   - spelling: 拼写错误 → 修正文字

    # 示例: 对比度问题修复（文字颜色与背景色对比度低）
    if issue_type == "contrast":
        try:
            prs = Presentation(target_pptx)
            slide = prs.slides[slide_idx - 1]  # 0-indexed
            # 将青色文字 (#00BCD4 等) 改为白色
            for shape in slide.shapes:
                if shape.has_text_frame:
                    for para in shape.text_frame.paragraphs:
                        for run in para.runs:
                            if run.font.color.rgb:
                                rgb = run.font.color.rgb
                                # 检测青色系 (G+B 高, R 低)
                                if rgb[1] > 180 and rgb[2] > 180 and rgb[0] < 100:
                                    run.font.color.rgb = RGBColor(255, 255, 255)
                                    print(f"    已修改文字颜色为白色")
            prs.save(target_pptx)
            print(f"    已保存修改")
        except Exception as e:
            print(f"    修复失败: {e}")

    # 示例: 溢出问题（缩小字号）
    elif issue_type == "overflow":
        try:
            prs = Presentation(target_pptx)
            slide = prs.slides[slide_idx - 1]
            for shape in slide.shapes:
                if shape.has_text_frame:
                    for para in shape.text_frame.paragraphs:
                        for run in para.runs:
                            if run.font.size and run.font.size > Pt(18):
                                original_size = run.font.size.pt
                                run.font.size = Pt(14)
                                print(f"    字号从 {original_size:.1f}pt 缩小到 14pt")
            prs.save(target_pptx)
            print(f"    已保存修改")
        except Exception as e:
            print(f"    修复失败: {e}")

    # 示例: 重叠问题（调整元素位置）
    elif issue_type == "overlap":
        print("    注意: 重叠问题需要手动调整元素位置")
        # 实际实现需要分析形状边界和 Z-order

    # 示例: 对齐问题
    elif issue_type == "alignment":
        try:
            prs = Presentation(target_pptx)
            slide = prs.slides[slide_idx - 1]
            for shape in slide.shapes:
                if shape.has_text_frame:
                    for para in shape.text_frame.paragraphs:
                        # 统一设置为左对齐
                        if para.alignment is None or para.alignment == PP_ALIGN.LEFT:
                            print(f"    已确认段落对齐: LEFT")
            prs.save(target_pptx)
        except Exception as e:
            print(f"    修复失败: {e}")

    else:
        print(f"    暂不支持修复类型: {issue_type}")

print("\n修复完成!")
PYEOF`, workDir, len(issues),
		countHighSeverity(issues),
		workDir, strings.Join(issueBlocks, ",\n"))
}

func countHighSeverity(issues []QAIssue) int {
	n := 0
	for _, issue := range issues {
		if strings.ToLower(issue.Severity) == "high" {
			n++
		}
	}
	return n
}
