/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/deep"
	agentplan "github.com/cloudwego/ppt-agent/pkg/agent/planexecute"
	"github.com/cloudwego/ppt-agent/pkg/callback"
	"github.com/cloudwego/ppt-agent/pkg/human"
	"github.com/cloudwego/ppt-agent/pkg/store"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	envPath := filepath.Join(pwd, ".env")
	_ = godotenv.Load(envPath)

	ctx := context.Background()

	// 设置 Callback（可观测性）
	// 1. 注册日志追踪 Handler
	logHandler := callback.NewLogHandler()
	callbacks.AppendGlobalHandlers(logHandler)
	fmt.Println("[Callback] 日志追踪 Handler 已注册")

	startTime := time.Now()

	// 2. 设置 CozeLoop 追踪（可选，需要配置环境变量）
	cozeLoopClient := setupCozeLoop(ctx)
	if cozeLoopClient != nil {
		callbacks.AppendGlobalHandlers(clc.NewLoopHandler(cozeLoopClient))
		fmt.Println("[Callback] CozeLoop Handler 已注册")
	}

	interactive := os.Getenv("INTERACTIVE") != "false"

	// 生成 UUID 作为会话标识
	taskID := uuid.New().String()
	fmt.Printf("[启动] 任务ID: %s\n", taskID)

	// 创建输出目录 output/{uuid}
	outputDir := filepath.Join(pwd, "..", "output", taskID)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Printf("[错误] 创建输出目录失败: %v\n", err)
		return
	}
	fmt.Printf("[启动] 输出目录: %s\n", outputDir)

	// 创建 LocalOperator
	fmt.Println("[启动] 创建 LocalOperator...")
	operator := &command.LocalOperator{}
	ctx = operator.SetWorkDir(ctx, outputDir)
	fmt.Println("[启动] LocalOperator 创建成功")

	// 创建 skill backend
	fmt.Println("[启动] 创建 skill backend...")
	be, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		fmt.Printf("local.NewBackend failed, err: %v\n", err)
		return
	}
	_ = be

	skillsDir := filepath.Join(pwd, "..", "skills")
	skillBackend, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: be,
		BaseDir: skillsDir,
	})
	if err != nil {
		fmt.Printf("skill.NewBackendFromFilesystem failed, err: %v\n", err)
		return
	}

	sm, err := skill.NewMiddleware(ctx, &skill.Config{
		Backend: skillBackend,
	})
	if err != nil {
		fmt.Printf("skill.NewMiddleware failed, err: %v\n", err)
		return
	}
	_ = sm
	fmt.Println("[启动] skill backend 创建成功")

	// 加载 skills 内容（用于注入 prompt）
	loadedSkills, err := agent.LoadSkillsFromDir(ctx, skillsDir)
	if err != nil {
		fmt.Printf("agent.LoadSkillsFromDir failed, err: %v\n", err)
		return
	}
	skillsContent := agent.FormatSkillsForPrompt(loadedSkills)
	fmt.Printf("[启动] 加载了 %d 个 skills\n", len(loadedSkills))

	query := schema.UserMessage("帮我做一个关于AI大模型介绍的PPT")
	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("任务ID: %s\n", taskID)
	fmt.Println("User query:", query.Content)
	fmt.Println("交互模式:", map[bool]string{true: "启用", false: "禁用"}[interactive])
	fmt.Printf("输出目录: %s\n", outputDir)
	fmt.Println("========================================")
	fmt.Println()

	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	select {
	case <-startupCtx.Done():
		fmt.Println("[错误] 启动超时 (30秒)")
		return
	default:
	}

	agentMode := os.Getenv("AGENT_MODE")

	// 创建人机交互管理器（两种模式都使用）
	hm := human.NewManager(interactive)

	if agentMode == "deep" {
		// DeepAgent 模式：使用 eino prebuilt/deep 并行执行
		fmt.Println("[启动] 使用 DeepAgent 模式（eino prebuilt/deep）")
		runDeepAgentMode(ctx, query.Content, taskID, outputDir, operator, skillsContent, interactive, hm)
	} else {
		// 原有 plan-execute 模式
		fmt.Println("[启动] 使用 Plan-Execute 模式（串行执行）")
		runPlanExecuteMode(ctx, query, taskID, outputDir, operator, skillsContent, hm)
	}

	// 打印执行摘要
	elapsed := time.Since(startTime).Round(time.Millisecond)
	fmt.Printf("\n[Callback] 任务执行完成，总耗时: %v\n", elapsed)

	// 关闭 CozeLoop（如果启用）
	if cozeLoopClient != nil {
		fmt.Println("[Callback] 等待 CozeLoop 数据上报...")
		time.Sleep(5 * time.Second)
		cozeLoopClient.Close(ctx)
	}

	fmt.Printf("\n[完成] 所有文件已保存到: %s\n", outputDir)
	time.Sleep(2 * time.Second)
}

// runPlanExecuteMode 运行原有的 plan-execute 串行模式
func runPlanExecuteMode(ctx context.Context, query *schema.Message, taskID, outputDir string,
	operator *command.LocalOperator, skillsContent string, hm *human.Manager) {

	// 创建 agents
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

	// 创建 plan-execute agent
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

	// 创建 runner
	fmt.Println("[启动] 创建 runner...")
	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           entryAgent,
		EnableStreaming: true,
		CheckPointStore: store.NewInMemoryStore(),
	})
	fmt.Println("[启动] runner 创建成功")

	fmt.Println("[执行] 启动 ADK Query...")
	fmt.Println("[执行] 请等待，这可能需要几秒钟...")

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

// runDeepAgentMode 运行 DeepAgent 并行模式（使用 eino prebuilt/deep）
func runDeepAgentMode(ctx context.Context, userQuery, taskID, outputDir string,
	operator *command.LocalOperator, skillsContent string, interactive bool, hm *human.Manager) {

	// 从环境变量获取并发数，默认 3
	concurrency := 3
	if envConcurrency := os.Getenv("DEEP_AGENT_CONCURRENCY"); envConcurrency != "" {
		if c, err := strconv.Atoi(envConcurrency); err == nil && c > 0 {
			concurrency = c
		}
	}
	fmt.Printf("[启动] 并发数: %d\n", concurrency)

	// QA 模型创建函数
	qaModelFn := func(ctx context.Context) (model.ToolCallingChatModel, error) {
		return agentutils.NewFallbackToolCallingChatModel(ctx,
			agentutils.WithMaxTokens(8192),
			agentutils.WithTemperature(0),
			agentutils.WithTopP(0),
		)
	}

	// 创建 PPT Deep Agent
	fmt.Println("[启动] 创建 PPT Deep Agent（eino prebuilt/deep）...")
	agent, err := deep.NewPPTTaskDeepAgent(ctx, &deep.PPTTaskConfig{
		WorkDir:     outputDir,
		TaskID:      taskID,
		Concurrency: concurrency,
		Operator:    operator,
		QAModelFn:   qaModelFn,
		Skills:      skillsContent,
	})
	if err != nil {
		fmt.Printf("deep.NewPPTTaskDeepAgent failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] PPT Deep Agent 创建成功")

	fmt.Println("[执行] 启动 DeepAgent...")
	fmt.Println("[执行] 请等待，这可能需要较长时间...")

	cfg := &deep.PPTTaskConfig{
		WorkDir:  outputDir,
		TaskID:   taskID,
		Operator: operator,
	}

	var result *deep.PPTTaskResult
	if interactive && hm != nil {
		// 启用人机交互模式：网络搜索前会询问用户
		result, err = deep.RunPPTTaskDeepAgentWithHuman(ctx, agent, cfg, userQuery, hm)
	} else {
		// 普通模式：直接流式输出
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

// setupCozeLoop 设置 CozeLoop 追踪
// 如果未配置环境变量，返回 nil
func setupCozeLoop(ctx context.Context) cozeloop.Client {
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
