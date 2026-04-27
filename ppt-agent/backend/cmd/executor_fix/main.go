package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/callbacks"

	"github.com/cloudwego/ppt-agent/pkg/agent"
	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/executor"
	"github.com/cloudwego/ppt-agent/pkg/agent/planner"
	"github.com/cloudwego/ppt-agent/pkg/agent/replanner"
	"github.com/cloudwego/ppt-agent/pkg/callback"
	"github.com/cloudwego/ppt-agent/pkg/generic"
	"github.com/cloudwego/ppt-agent/pkg/human"
	"github.com/cloudwego/ppt-agent/pkg/params"
	"github.com/cloudwego/ppt-agent/pkg/store"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/executor_fix <workDir>")
		fmt.Println("Example: go run ./cmd/executor_fix /ppt/ppt-agent/output/ea9c5c5f-bcd6-4f32-a6d3-034adaf46887")
		os.Exit(1)
	}
	workDir := os.Args[1]
	fmt.Printf("[启动] 工作目录: %s\n", workDir)

	// 加载 .env
	pwd, _ := os.Getwd()
	_ = godotenv.Load(filepath.Join(pwd, ".env"))

	// Step 1: 扫描已有的 PPTX 文件
	fmt.Println("\n=== Step 1: 扫描已有幻灯片 ===")
	existingFiles := generic.GetExistingStepFiles(workDir)
	fmt.Printf("[文件] 找到 %d 个幻灯片文件\n", len(existingFiles))

	// Step 2: 注册日志回调
	fmt.Println("\n=== Step 2: 注册日志回调 ===")
	logHandler := callback.NewLogHandler()
	callbacks.AppendGlobalHandlers(logHandler)
	fmt.Println("[Callback] 日志 Handler 已注册")

	// Step 3: 创建 LocalOperator 和 skill backend
	fmt.Println("\n=== Step 3: 初始化 Agent 环境 ===")
	operator := &command.LocalOperator{}
	ctx = operator.SetWorkDir(ctx, workDir)

	be, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		fmt.Printf("[错误] local.NewBackend 失败: %v\n", err)
		os.Exit(1)
	}

	skillsDir := filepath.Join(pwd, "..", "skills")
	skillBackend, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: be,
		BaseDir: skillsDir,
	})
	if err != nil {
		fmt.Printf("[错误] skill.NewBackendFromFilesystem 失败: %v\n", err)
		os.Exit(1)
	}

	sm, err := skill.NewMiddleware(ctx, &skill.Config{
		Backend: skillBackend,
	})
	if err != nil {
		fmt.Printf("[错误] skill.NewMiddleware 失败: %v\n", err)
		os.Exit(1)
	}
	_ = sm
	fmt.Println("[Skill] skill backend 创建成功")

	// 加载 skills 内容
	loadedSkills, err := agent.LoadSkillsFromDir(ctx, skillsDir)
	if err != nil {
		fmt.Printf("[错误] LoadSkillsFromDir 失败: %v\n", err)
		os.Exit(1)
	}
	skillsContent := agent.FormatSkillsForPrompt(loadedSkills)
	fmt.Printf("[Skill] 加载了 %d 个 skills\n", len(loadedSkills))

	// Step 4: 设置工作目录到 context
	ctx = params.SetTypedContextParams(ctx, params.WorkDirSessionKey, workDir)

	// Step 5: 创建 Planner
	fmt.Println("\n=== Step 4: 创建 Planner Agent ===")
	planAgent, err := planner.NewPlanner(ctx, operator, skillsContent)
	if err != nil {
		fmt.Printf("[错误] NewPlanner 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[Planner] 创建成功")

	// Step 6: 创建 Executor
	fmt.Println("\n=== Step 5: 创建 Executor Agent ===")
	executeAgent, err := executor.NewExecutor(ctx, operator, skillsContent)
	if err != nil {
		fmt.Printf("[错误] NewExecutor 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[Executor] 创建成功")

	// Step 7: 创建 Replanner
	fmt.Println("\n=== Step 6: 创建 Replanner Agent ===")
	replanAgent, err := replanner.NewReplanner(ctx, operator)
	if err != nil {
		fmt.Printf("[错误] NewReplanner 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[Replanner] 创建成功")

	// Step 8: 创建 plan-execute agent
	fmt.Println("\n=== Step 7: 创建 Plan-Execute Agent ===")
	entryAgent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planAgent,
		Executor:      executeAgent,
		Replanner:     replanAgent,
		MaxIterations: 50,
	})
	if err != nil {
		fmt.Printf("[错误] planexecute.New 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[Plan-Execute] 创建成功")

	// Step 9: 创建 Runner
	fmt.Println("\n=== Step 8: 创建 Runner ===")
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           entryAgent,
		EnableStreaming: false,
		CheckPointStore: store.NewInMemoryStore(),
	})
	fmt.Println("[Runner] 创建成功")

	// Step 10: 从 stdin 读取用户输入
	fmt.Println("\n========================================")
	fmt.Printf("[测试] 工作目录: %s\n", workDir)
	fmt.Printf("[测试] 已存在幻灯片数量: %d\n", len(existingFiles))
	fmt.Println("========================================")
	fmt.Println("\n请输入 PPT 制作需求（输入完成后按 Ctrl+D 结束）：")

	var userInput string
	fmt.Scanln(&userInput)
	if userInput == "" {
		buf := make([]byte, 4096)
		n, _ := os.Stdin.Read(buf)
		userInput = string(buf[:n])
	}
	userInput = fmt.Sprintf("%s\n\n工作目录: %s\n已有幻灯片文件: %d 个", userInput, workDir, len(existingFiles))

	fmt.Printf("\n--- 用户查询 ---\n%s\n---\n", userInput)

	// Step 11: 运行
	startTime := time.Now()
	fmt.Println("\n[执行] 启动 ADK Query...")

	iter := runner.Query(ctx, userInput, adk.WithCheckPointID("ppt-gen"))

	fmt.Println("[执行] Query 已启动，开始处理事件...")

	hm := human.NewManager(false)

	event, err := hm.RunWithApproval(ctx, runner, "ppt-gen", iter)
	if err != nil {
		fmt.Printf("[错误] 执行失败: %v\n", err)
	}

	if event != nil && event.Output != nil {
		if msg, _, getErr := adk.GetMessage(event); getErr == nil && msg != nil {
			fmt.Printf("\n=== 最终结果 ===\n%s\n", msg.Content)
		}
	}

	elapsed := time.Since(startTime).Round(time.Millisecond)
	fmt.Printf("\n[完成] 执行完成，总耗时: %v\n", elapsed)
}
