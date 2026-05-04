package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	clc "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/coze-dev/cozeloop-go"

	"github.com/cloudwego/ppt-agent/pkg/agent"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	agentplan "github.com/cloudwego/ppt-agent/pkg/agent/planexecute"
	"github.com/cloudwego/ppt-agent/pkg/callback"
	"github.com/cloudwego/ppt-agent/pkg/human"
	"github.com/cloudwego/ppt-agent/pkg/server"
	"github.com/cloudwego/ppt-agent/pkg/store"
)

func main() {
	modeFlag := flag.String("mode", "cli", "启动模式: cli (命令行) 或 web (Web服务)")
	addrFlag := flag.String("addr", ":8080", "Web 模式监听地址")
	flag.Parse()

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	envPath := filepath.Join(pwd, ".env")
	_ = godotenv.Load(envPath)

	ctx := context.Background()

	// Callbacks
	logHandler := callback.NewLogHandler()
	callbacks.AppendGlobalHandlers(logHandler)
	fmt.Println("[Callback] 日志追踪 Handler 已注册")

	cozeLoopClient := setupCozeLoop()
	if cozeLoopClient != nil {
		callbacks.AppendGlobalHandlers(clc.NewLoopHandler(cozeLoopClient))
		fmt.Println("[Callback] CozeLoop Handler 已注册")
	}

	// Skill backend
	be, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		BaseDir: filepath.Join(pwd, "..", "skills"),
	})
	if err != nil {
		fmt.Printf("skill.NewBackendFromFilesystem failed, err: %v\n", err)
		return
	}
	_ = be

	skillsDir := filepath.Join(pwd, "..", "skills")
	loadedSkills, err := agent.LoadSkillsFromDir(ctx, skillsDir)
	if err != nil {
		fmt.Printf("agent.LoadSkillsFromDir failed, err: %v\n", err)
		return
	}
	skillsContent := agent.FormatSkillsForPrompt(loadedSkills)
	fmt.Printf("[启动] 加载了 %d 个 skills\n", len(loadedSkills))

	switch *modeFlag {
	case "web":
		runWebMode(pwd, skillsContent, skillsDir, *addrFlag)
	default:
		runCLIMode(ctx, pwd, skillsContent, skillsDir)
	}

	if cozeLoopClient != nil {
		fmt.Println("[Callback] 等待 CozeLoop 数据上报...")
		time.Sleep(5 * time.Second)
		cozeLoopClient.Close(ctx)
	}
}

// ---------------------------------------------------------------------------
// Web mode
// ---------------------------------------------------------------------------

func runWebMode(pwd, skillsContent, skillsDir, addr string) {
	outputBase := filepath.Join(pwd, "..", "output")

	// Shared operator (stateless, safe to reuse).
	operator := &command.LocalOperator{}

	concurrency := 5
	if c := os.Getenv("DEEP_AGENT_CONCURRENCY"); c != "" {
		if v, err := strconv.Atoi(c); err == nil && v > 0 {
			concurrency = v
		}
	}

	qaModelFn := func(ctx context.Context) (model.ToolCallingChatModel, error) {
		return agentutils.NewFallbackToolCallingChatModel(ctx,
			agentutils.WithMaxTokens(8192),
			agentutils.WithTemperature(0),
			agentutils.WithTopP(0),
		)
	}

	// Agent factory: creates a fresh agent per task with the right WorkDir/TaskID.
	agentFactory := func(ctx context.Context, cfg *deep.PPTTaskConfig) (adk.Agent, error) {
		cfg.Concurrency = concurrency
		cfg.Operator = operator
		cfg.QAModelFn = qaModelFn
		cfg.Skills = skillsContent
		cfg.SkillsDir = skillsDir
		return deep.NewPPTTaskDeepAgent(ctx, cfg)
	}

	srv := server.NewServer(&server.ServerConfig{
		Addr:         addr,
		BaseDir:      outputBase,
		FrontendDir:  filepath.Join(pwd, "..", "frontend", "dist"),
		AgentFactory: agentFactory,
		MakeTaskConfig: func(taskID string) *deep.PPTTaskConfig {
			return &deep.PPTTaskConfig{
				TaskID: taskID,
			}
		},
	})

	fmt.Printf("[Web] 启动模式: DeepAgent\n")
	fmt.Printf("[Web] 并发数: %d\n", concurrency)
	if err := srv.Start(); err != nil {
		log.Printf("[Web] 服务器错误: %v\n", err)
	}
}

// ---------------------------------------------------------------------------
// CLI mode (existing behaviour)
// ---------------------------------------------------------------------------

func runCLIMode(ctx context.Context, pwd, skillsContent, skillsDir string) {
	interactive := os.Getenv("INTERACTIVE") != "false"

	taskID := uuid.New().String()
	fmt.Printf("[启动] 任务ID: %s\n", taskID)

	outputDir := filepath.Join(pwd, "..", "output", taskID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("[错误] 创建输出目录失败: %v\n", err)
		return
	}
	fmt.Printf("[启动] 输出目录: %s\n", outputDir)

	operator := &command.LocalOperator{}
	ctx = operator.SetWorkDir(ctx, outputDir)
	fmt.Println("[启动] LocalOperator 创建成功")

	queryContent := getUserQuery()
	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("任务ID: %s\n", taskID)
	fmt.Printf("User query: %s\n", queryContent)
	fmt.Printf("交互模式: %s\n", map[bool]string{true: "启用", false: "禁用"}[interactive])
	fmt.Printf("输出目录: %s\n", outputDir)
	fmt.Println("========================================")
	fmt.Println()

	agentMode := os.Getenv("AGENT_MODE")
	hm := human.NewManager(interactive)

	if agentMode == "deep" {
		fmt.Println("[启动] 使用 DeepAgent 模式（eino prebuilt/deep）")
		runDeepAgentCLI(ctx, queryContent, taskID, outputDir, operator, skillsContent, skillsDir, interactive, hm)
	} else {
		query := schema.UserMessage(queryContent)
		runPlanExecuteCLI(ctx, query, taskID, operator, skillsContent, hm)
	}
}

func runDeepAgentCLI(ctx context.Context, userQuery, taskID, outputDir string,
	operator *command.LocalOperator, skillsContent, skillsDir string, interactive bool, hm *human.Manager) {

	concurrency := 5
	if envConcurrency := os.Getenv("DEEP_AGENT_CONCURRENCY"); envConcurrency != "" {
		if c, err := strconv.Atoi(envConcurrency); err == nil && c > 0 {
			concurrency = c
		}
	}
	fmt.Printf("[启动] 并发数: %d\n", concurrency)

	qaModelFn := func(ctx context.Context) (model.ToolCallingChatModel, error) {
		return agentutils.NewFallbackToolCallingChatModel(ctx,
			agentutils.WithMaxTokens(8192),
			agentutils.WithTemperature(0),
			agentutils.WithTopP(0),
		)
	}

	fmt.Println("[启动] 创建 PPT Deep Agent（eino prebuilt/deep）...")
	agent, err := deep.NewPPTTaskDeepAgent(ctx, &deep.PPTTaskConfig{
		WorkDir:     outputDir,
		TaskID:      taskID,
		Concurrency: concurrency,
		Operator:    operator,
		QAModelFn:   qaModelFn,
		Skills:      skillsContent,
		SkillsDir:   skillsDir,
	})
	if err != nil {
		fmt.Printf("deep.NewPPTTaskDeepAgent failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] PPT Deep Agent 创建成功")

	cfg := &deep.PPTTaskConfig{
		WorkDir:  outputDir,
		TaskID:   taskID,
		Operator: operator,
	}

	var result *deep.PPTTaskResult
	if interactive && hm != nil {
		result, err = deep.RunPPTTaskDeepAgentWithHuman(ctx, agent, cfg, userQuery, hm)
	} else {
		result, err = deep.RunPPTTaskDeepAgent(ctx, agent, cfg, userQuery)
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if result != nil {
		fmt.Printf("\n=== 最终结果 ===\n%s\n", result.Message)
		fmt.Printf("生成文件数: %d / %d\n", result.DoneSlides, result.TotalSlides)
		if len(result.Files) > 0 {
			fmt.Println("生成的文件:")
			for _, f := range result.Files {
				fmt.Printf("  - %s\n", f)
			}
		}
	}
}

func runPlanExecuteCLI(ctx context.Context, query *schema.Message, taskID string,
	operator *command.LocalOperator, skillsContent string, hm *human.Manager) {

	fmt.Println("[启动] 创建 planner agent...")
	planAgent, err := agentplan.NewPlanner(ctx, operator, skillsContent)
	if err != nil {
		fmt.Printf("planner.NewPlanner failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] planner 创建成功")

	fmt.Println("[启动] 创建 executor agent...")
	executeAgent, err := agentplan.NewExecutor(ctx, operator, skillsContent)
	if err != nil {
		fmt.Printf("executor.NewExecutor failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] executor 创建成功")

	fmt.Println("[启动] 创建 replanner agent...")
	replanAgent, err := agentplan.NewReplanner(ctx, operator)
	if err != nil {
		fmt.Printf("replanner.NewReplanner failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] replanner 创建成功")

	fmt.Println("[启动] 创建 plan-execute agent...")
	entryAgent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planAgent,
		Executor:      executeAgent,
		Replanner:     replanAgent,
		MaxIterations: 150,
	})
	if err != nil {
		fmt.Printf("planexecute.New failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] plan-execute agent 创建成功")

	fmt.Println("[启动] 创建 runner...")
	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           entryAgent,
		EnableStreaming: true,
		CheckPointStore: store.NewInMemoryStore(),
	})
	fmt.Println("[启动] runner 创建成功")

	fmt.Println("[执行] 启动 ADK Query...")
	iter := r.Query(ctx, query.Content, adk.WithCheckPointID(taskID))
	fmt.Println("[执行] Query 已启动，开始处理事件...")

	event, err := hm.RunWithApproval(ctx, r, taskID, iter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if event != nil && event.Output != nil {
		if msg, _, getErr := adk.GetMessage(event); getErr == nil && msg != nil {
			fmt.Printf("\n=== 最终结果 ===\n%s\n", msg.Content)
		}
	}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

func setupCozeLoop() cozeloop.Client {
	apiToken := os.Getenv("COZELOOP_API_TOKEN")
	workspaceID := os.Getenv("COZELOOP_WORKSPACE_ID")

	if apiToken == "" || workspaceID == "" {
		log.Println("[Callback] CozeLoop 未配置 (COZELOOP_API_TOKEN 或 COZELOOP_WORKSPACE_ID 未设置)，跳过")
		return nil
	}

	client, err := cozeloop.NewClient(
		cozeloop.WithAPIToken(apiToken),
		cozeloop.WithWorkspaceID(workspaceID),
	)
	if err != nil {
		log.Printf("[Callback] CozeLoop 设置失败: %v\n", err)
		return nil
	}

	log.Printf("[Callback] CozeLoop 配置成功: workspaceID=%s", workspaceID)
	return client
}

func getUserQuery() string {
	if len(os.Args) > 1 {
		return strings.Join(os.Args[1:], " ")
	}

	fmt.Println("请输入您的PPT需求（如：帮我做一个关于AI大模型介绍的PPT）:")
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	query, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return ""
		}
		log.Printf("读取输入失败: %v\n", err)
		return ""
	}

	query = strings.TrimSpace(query)
	if query == "" {
		fmt.Println("未输入内容，将使用默认查询")
		return "帮我做一个关于AI大模型介绍的PPT"
	}

	return query
}
