package qa

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCheckDependencies 检查必要的依赖工具是否已安装
// 这是最快的方式来验证环境是否就绪
func TestCheckDependencies(t *testing.T) {
	tools := []string{"soffice", "pdftoppm"}

	for _, tool := range tools {
		path, err := exec.LookPath(tool)
		if err != nil {
			t.Errorf("依赖工具 %s 未找到: %v", tool, err)
		} else {
			t.Logf("依赖工具 %s 位置: %s", tool, path)
		}
	}
}

// TestPythonConverterScript 直接测试 Python 转换脚本
// 这是验证 QA 功能的最小化测试
func TestPythonConverterScript(t *testing.T) {
	// 配置：修改这些路径以匹配你的服务器环境
	pythonBin := "/root/pptx_env/bin/python"
	scriptPath := "/ppt/ppt-agent/backend/pkg/tools/qa/pptx_qa_converter.py"
	pptxDir := "/ppt/ppt-agent/output/018124c8-9c22-4135-a7c3-1627fb224193"

	// 检查脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("Python 转换脚本不存在，跳过测试")
	}

	// 检查 PPTX 文件是否存在
	pptxFiles, _ := filepath.Glob(filepath.Join(pptxDir, "*.pptx"))
	if len(pptxFiles) == 0 {
		t.Skip("未找到 PPTX 文件，跳过测试")
	}
	t.Logf("找到 %d 个 PPTX 文件", len(pptxFiles))

	// 创建临时输出目录
	tmpDir, err := os.MkdirTemp("", "qa-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	imgDir := filepath.Join(tmpDir, "images")
	os.MkdirAll(imgDir, 0755)

	// 执行转换脚本
	cmd := exec.Command(pythonBin, scriptPath,
		"--pptx-dir", pptxDir,
		"--output-dir", imgDir,
		"--dpi", "100") // 使用较低的 DPI 加快测试
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH"))

	output, err := cmd.CombinedOutput()
	t.Logf("脚本输出:\n%s", string(output))

	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	// 检查结果
	imgFiles, _ := filepath.Glob(filepath.Join(imgDir, "slide_*.jpg"))
	if len(imgFiles) == 0 {
		t.Error("未生成任何图片")
	} else {
		t.Logf("成功生成 %d 张图片", len(imgFiles))
		for _, f := range imgFiles {
			info, _ := os.Stat(f)
			t.Logf("  %s (%.1f KB)", filepath.Base(f), float64(info.Size())/1024)
		}
	}
}

// TestGoToolRunConverter 直接测试 Go 层的 runConverter 函数
// 这个测试验证 Go 代码调用 Python 脚本的完整链路
func TestGoToolRunConverter(t *testing.T) {
	// 这个测试需要完整的依赖注入，适合集成测试
	// 由于 NewTool 需要 model 和 operator，这里只做说明

	t.Log("如需测试完整的 Go 工具链路，请使用：")
	t.Log("  cd /ppt/ppt-agent/backend")
	t.Log("  go test -v ./pkg/tools/qa/ -run TestPythonConverterScript")
	t.Log("")
	t.Log("或手动验证：")
	t.Log("  python3 /ppt/ppt-agent/backend/pkg/tools/qa/pptx_qa_converter.py \\")
	t.Log("    --pptx-dir /ppt/ppt-agent/output/018124c8-9c22-4135-a7c3-1627fb224193 \\")
	t.Log("    --output-dir /tmp/qa-test")
}
