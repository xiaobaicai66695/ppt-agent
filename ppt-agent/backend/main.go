package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	clc "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/coze-dev/cozeloop-go"

	"github.com/cloudwego/ppt-agent/pkg/agent"
	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	agentplan "github.com/cloudwego/ppt-agent/pkg/agent/planexecute"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/callback"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/human"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/store"
	"github.com/cloudwego/ppt-agent/pkg/web"
)

func main() {
	modeFlag := flag.String("mode", "cli", "启动模式: cli (命令行) 或 web (Web服务)")
	addrFlag := flag.String("addr", ":8080", "Web 模式监听地址")
	flag.Parse()

	pwd, err := os.Getwd()
	if err != nil {
		logger.Error("get_cwd_failed", "error", err.Error())
		return
	}

	envPath := filepath.Join(pwd, ".env")
	if err := godotenv.Load(envPath); err != nil && !os.IsNotExist(err) {
		logger.Warn("dotenv_load_failed", "path", envPath, "error", err.Error())
	}
	// Initialize structured logger (JSON output for production, text for dev)
	logger.Init(os.Getenv("LOG_JSON") == "true")

	ctx := context.Background()

	// Callbacks
	logHandler := callback.NewLogHandler()
	callbacks.AppendGlobalHandlers(logHandler)
	logger.Info("callback_handler_registered", "type", "log")

	cozeLoopClient := setupCozeLoop()
	if cozeLoopClient != nil {
		callbacks.AppendGlobalHandlers(clc.NewLoopHandler(cozeLoopClient))
		logger.Info("callback_handler_registered", "type", "cozeloop")
	}

	// Skill backend
	localBE, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		logger.Error("local_backend_failed", "error", err.Error())
		return
	}

	skillsDir := filepath.Join(pwd, "..", "skills")
	_, err = skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: localBE,
		BaseDir: skillsDir,
	})
	if err != nil {
		logger.Error("skill_backend_failed", "error", err.Error())
		return
	}

	loadedSkills, err := agent.LoadSkillsFromDir(ctx, skillsDir)
	if err != nil {
		logger.Error("load_skills_failed", "error", err.Error())
		return
	}
	skillsContent := agent.FormatSkillsForPrompt(loadedSkills)
	logger.Info("skills_loaded", "count", len(loadedSkills), "dir", skillsDir)

	// MySQL 初始化
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:210618@tcp(127.0.0.1:3306)/myapp?charset=utf8mb4&parseTime=True"
	}
	if err := db.Init(dsn); err != nil {
		logger.Warn("mysql_init_failed", "error", err.Error())
	} else {
		// 创建默认 root 管理员
		rootEmail := getEnvDefault("ROOT_EMAIL", "root@ppt-agent.local")
		rootPass := getEnvDefault("ROOT_PASSWORD", "root")
		auth.SeedRootUser(rootEmail, rootPass)
	}

	switch *modeFlag {
	case "web":
		runWebMode(pwd, skillsContent, skillsDir, *addrFlag)
	default:
		runCLIMode(ctx, pwd, skillsContent, skillsDir)
	}

	if cozeLoopClient != nil {
		logger.Info("cozeloop_shutdown_waiting")
		time.Sleep(5 * time.Second)
		cozeLoopClient.Close(ctx)
	}
}

// ---------------------------------------------------------------------------
// Web mode
// ---------------------------------------------------------------------------

func runWebMode(pwd, skillsContent, skillsDir, addr string) {
	outputBase := filepath.Join(pwd, "..", "weboutput")

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

	srv := web.NewServer(&web.ServerConfig{
		Addr:         addr,
		BaseDir:      outputBase,
		FrontendDir:  filepath.Join(pwd, "..", "frontend", "dist"),
		SkillsDir:    skillsDir,
		AgentFactory: agentFactory,
		MakeTaskConfig: func(taskID string) *deep.PPTTaskConfig {
			return &deep.PPTTaskConfig{
				TaskID: taskID,
			}
		},
		AIModelFactory: func(ctx context.Context) (interface {
			Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
		}, error) {
			m, err := agentutils.NewFallbackToolCallingChatModel(ctx,
				agentutils.WithMaxTokens(4096),
				agentutils.WithTemperature(0),
			)
			if err != nil {
				return nil, err
			}
			return &aiModelAdapter{model: m}, nil
		},
		StyleModelFactory: func(ctx context.Context) (model.ToolCallingChatModel, error) {
			return agentutils.NewFallbackToolCallingChatModel(ctx,
				agentutils.WithModel(os.Getenv("ARK_TEXT_MODEL")),
				agentutils.WithMaxTokens(4096),
				agentutils.WithTemperature(0),
			)
		},
		// 意图分类 / 偏好总结使用轻量级模型，节省成本
		TextModelFactory: func(ctx context.Context) (interface {
			Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
		}, error) {
			m, err := agentutils.NewFallbackToolCallingChatModel(ctx,
				agentutils.WithModel(os.Getenv("ARK_TEXT_MODEL")),
				agentutils.WithMaxTokens(2048),
				agentutils.WithTemperature(0),
			)
			if err != nil {
				return nil, err
			}
			return &aiModelAdapter{model: m}, nil
		},
	})

	logger.Info("server_starting", "mode", "web", "addr", addr, "concurrency", concurrency)

	// Set health check metadata
	buildVersion := os.Getenv("APP_VERSION")
	if buildVersion == "" {
		buildVersion = "dev"
	}
	web.Version = buildVersion
	web.StartTime = time.Now()

	if err := srv.Start(); err != nil {
		logger.Error("server_error", "error", err.Error())
	}
}

// ---------------------------------------------------------------------------
// CLI mode (existing behaviour)
// ---------------------------------------------------------------------------

func runCLIMode(ctx context.Context, pwd, skillsContent, skillsDir string) {
	interactive := os.Getenv("INTERACTIVE") != "false"

	taskID := uuid.New().String()
	logger.Info("cli_task_starting", "task_id", taskID)

	outputDir := filepath.Join(pwd, "..", "output", taskID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("mkdir_output_dir_failed", "path", outputDir, "error", err.Error())
		return
	}
	logger.Info("output_dir_created", "path", outputDir)

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
		logger.Info("agent_mode_selected", "mode", "deep")
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
	logger.Info("cli_deep_agent_config", "concurrency", concurrency)

	qaModelFn := func(ctx context.Context) (model.ToolCallingChatModel, error) {
		return agentutils.NewFallbackToolCallingChatModel(ctx,
			agentutils.WithMaxTokens(8192),
			agentutils.WithTemperature(0),
			agentutils.WithTopP(0),
		)
	}

	logger.Info("deep_agent_creating")
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
		logger.Error("deep_agent_creation_failed", "error", err.Error())
		return
	}
	logger.Info("deep_agent_created")

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

type aiModelAdapter struct {
	model model.ToolCallingChatModel
}

func (a *aiModelAdapter) Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error) {
	return a.model.Generate(ctx, messages)
}

// ---------------------------------------------------------------------------

func setupCozeLoop() cozeloop.Client {
	apiToken := os.Getenv("COZELOOP_API_TOKEN")
	workspaceID := os.Getenv("COZELOOP_WORKSPACE_ID")

	if apiToken == "" || workspaceID == "" {
		logger.Debug("cozeloop_not_configured")
		return nil
	}

	client, err := cozeloop.NewClient(
		cozeloop.WithAPIToken(apiToken),
		cozeloop.WithWorkspaceID(workspaceID),
	)
	if err != nil {
		logger.Warn("cozeloop_setup_failed", "error", err.Error())
		return nil
	}

	logger.Info("cozeloop_configured", "workspace_id", workspaceID)
	return client
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
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
		fmt.Printf("读取输入失败: %v\n", err)
		return ""
	}

	query = strings.TrimSpace(query)
	if query == "" {
		fmt.Println("未输入内容，将使用默认查询")
		return "帮我做一个关于AI大模型介绍的PPT"
	}

	return query
}
