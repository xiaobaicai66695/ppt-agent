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

	clc "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/coze-dev/cozeloop-go"

	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/deck"
	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/auth"
	"github.com/cloudwego/ppt-agent/pkg/callback"
	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/human"
	"github.com/cloudwego/ppt-agent/pkg/logger"
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

	skillsDir := filepath.Join(pwd, "..", "skills")
	if _, err := os.Stat(filepath.Join(skillsDir, "ppt-deck-planner", "SKILL.md")); err != nil {
		logger.Error("deck_planner_skill_missing", "dir", skillsDir, "error", err.Error())
		return
	}
	logger.Info("deck_planner_skill_ready", "dir", skillsDir)

	// MySQL 初始化
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:210618@tcp(127.0.0.1:3306)/myapp?charset=utf8mb4&parseTime=True"
	}
	if err := db.Init(dsn); err != nil {
		logger.Warn("mysql_init_failed", "error", err.Error())
	} else {
		// 创建默认 root 管理员
		rootEmail := getEnvDefault("ROOT_EMAIL", "root@qq.com")
		rootPass := getEnvDefault("ROOT_PASSWORD", "root")
		auth.SeedRootUser(rootEmail, rootPass)
	}

	switch *modeFlag {
	case "web":
		runWebMode(pwd, skillsDir, *addrFlag)
	default:
		runCLIMode(ctx, pwd, skillsDir)
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

func runWebMode(pwd, skillsDir, addr string) {
	outputBase := filepath.Join(pwd, "..", "weboutput")

	// Shared operator (stateless, safe to reuse).
	operator := &command.LocalOperator{}

	concurrency := 5
	if c := os.Getenv("PLANNER_CONCURRENCY"); c != "" {
		if v, err := strconv.Atoi(c); err == nil && v > 0 {
			concurrency = v
		}
	}

	// Agent factory: creates a fresh agent per task with the right WorkDir/TaskID.
	agentFactory := func(ctx context.Context, cfg *deck.PPTTaskConfig) (adk.Agent, error) {
		if cfg.ModelAPIKey == "" && cfg.UserID > 0 {
			credential := resolveUserModelCredential(uint(cfg.UserID))
			cfg.ModelAPIKey = credential.APIKey
			cfg.ModelProvider = credential.Provider
		}
		cfg.Concurrency = concurrency
		cfg.Operator = operator
		cfg.SkillsDir = skillsDir
		return deck.NewPPTPlannerAgent(ctx, cfg)
	}

	logAnalysisModelFactory := func(ctx context.Context) (model.ToolCallingChatModel, error) {
		return agentutils.NewFallbackToolCallingChatModel(ctx,
			agentutils.WithMaxTokens(8192),
			agentutils.WithTemperature(0),
		)
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LOG_ANALYSIS_ENABLED")), "false") {
		logAnalysisModelFactory = nil
	}

	srv := web.NewServer(&web.ServerConfig{
		Addr:         addr,
		BaseDir:      outputBase,
		FrontendDir:  filepath.Join(pwd, "..", "frontend", "dist"),
		SkillsDir:    skillsDir,
		Operator:     operator,
		AgentFactory: agentFactory,
		MakeTaskConfig: func(taskID string) *deck.PPTTaskConfig {
			return &deck.PPTTaskConfig{
				TaskID: taskID,
			}
		},
		AIModelFactory: func(ctx context.Context) (interface {
			Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
		}, error) {
			opts := []agentutils.ChatModelOption{
				agentutils.WithMaxTokens(4096),
				agentutils.WithTemperature(0),
			}
			if credential := resolveUserModelCredentialFromContext(ctx); credential.APIKey != "" {
				opts = append(opts, agentutils.WithAPIKeyForProvider(credential.Provider, credential.APIKey))
			}
			m, err := agentutils.NewFallbackToolCallingChatModel(ctx, opts...)
			if err != nil {
				return nil, err
			}
			return &aiModelAdapter{model: m}, nil
		},
		// 继续对话分类等辅助任务使用轻量级模型，节省成本
		TextModelFactory: func(ctx context.Context) (interface {
			Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
		}, error) {
			routingModel := strings.TrimSpace(os.Getenv("ARK_ROUTER_MODEL"))
			if routingModel == "" {
				routingModel = strings.TrimSpace(os.Getenv("ARK_MODEL"))
			}
			credential := resolveUserModelCredentialFromContext(ctx)
			opts := []agentutils.ChatModelOption{
				agentutils.WithMaxTokens(1024),
				agentutils.WithTemperature(0),
				agentutils.WithDisableThinking(true),
			}
			if credential.Provider == "" || modelcompat.NormalizeProvider(credential.Provider) == modelcompat.ProviderArk {
				opts = append([]agentutils.ChatModelOption{agentutils.WithModel(routingModel)}, opts...)
			}
			if credential.APIKey != "" {
				opts = append(opts, agentutils.WithAPIKeyForProvider(credential.Provider, credential.APIKey))
			}
			m, err := agentutils.NewFallbackToolCallingChatModel(ctx, opts...)
			if err != nil {
				return nil, err
			}
			return &aiModelAdapter{model: m}, nil
		},
		LogAnalysisModelFactory: logAnalysisModelFactory,
		LogAnalysisIdleInterval: parseDurationEnv("LOG_ANALYSIS_IDLE_INTERVAL", 5*time.Minute),
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

func runCLIMode(ctx context.Context, pwd, skillsDir string) {
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

	logger.Info("agent_mode_selected", "mode", "planner")
	hm := human.NewManager(interactive)
	runPlannerCLI(ctx, queryContent, taskID, outputDir, operator, skillsDir, interactive, hm)
}

func runPlannerCLI(ctx context.Context, userQuery, taskID, outputDir string,
	operator *command.LocalOperator, skillsDir string, interactive bool, hm *human.Manager) {

	concurrency := 5
	if envConcurrency := os.Getenv("PLANNER_CONCURRENCY"); envConcurrency != "" {
		if c, err := strconv.Atoi(envConcurrency); err == nil && c > 0 {
			concurrency = c
		}
	}
	logger.Info("cli_planner_config", "concurrency", concurrency)

	logger.Info("planner_creating")
	agent, err := deck.NewPPTPlannerAgent(ctx, &deck.PPTTaskConfig{
		WorkDir:     outputDir,
		TaskID:      taskID,
		Concurrency: concurrency,
		Operator:    operator,
		SkillsDir:   skillsDir,
	})
	if err != nil {
		logger.Error("planner_creation_failed", "error", err.Error())
		return
	}
	logger.Info("planner_created")

	cfg := &deck.PPTTaskConfig{
		WorkDir:  outputDir,
		TaskID:   taskID,
		Operator: operator,
	}

	var result *deck.PPTTaskResult
	if interactive && hm != nil {
		result, err = deck.RunPPTPlannerWithHuman(ctx, agent, cfg, userQuery, hm)
	} else {
		result, err = deck.RunPPTPlanner(ctx, agent, cfg, userQuery)
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
	// 在 NewClient 之前设置，压制 SDK 包初始化时打印的 DEBUG 日志
	cozeloop.SetLogLevel(cozeloop.LogLevelWarn)

	apiToken := os.Getenv("COZELOOP_API_TOKEN")
	workspaceID := os.Getenv("COZELOOP_WORKSPACE_ID")

	if apiToken == "" || workspaceID == "" {
		logger.Info("cozeloop_not_configured")
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

// parseDurationEnv parses a duration string from an env var, returns defaultVal on error or if empty.
func parseDurationEnv(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		logger.Warn("invalid_duration_env", "key", key, "value", v, "error", err.Error())
		return defaultVal
	}
	return d
}

type userModelCredential struct {
	Provider string
	APIKey   string
}

func resolveUserModelCredentialFromContext(ctx context.Context) userModelCredential {
	if ctx == nil {
		return userModelCredential{}
	}
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok || userID <= 0 {
		return userModelCredential{}
	}
	return resolveUserModelCredential(uint(userID))
}

func resolveUserModelCredential(userID uint) userModelCredential {
	provider := modelcompat.ProviderArk
	accountKey := ""
	if userID > 0 && db.DB != nil {
		record, err := db.GetUserAPIKey(userID)
		if err != nil {
			logger.Warn("user_api_key_lookup_failed", "user_id", userID, "error", err.Error())
		} else if record != nil {
			provider = modelcompat.NormalizeProvider(record.Provider)
			accountKey = record.APIKey
		}
	}
	return userModelCredential{
		Provider: string(provider),
		APIKey:   modelcompat.ResolveProviderAPIKey(provider, accountKey),
	}
}
