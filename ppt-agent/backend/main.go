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
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/ppt-agent/pkg/agent"
	"github.com/cloudwego/ppt-agent/pkg/agent/command"
	"github.com/cloudwego/ppt-agent/pkg/agent/executor"
	"github.com/cloudwego/ppt-agent/pkg/agent/planner"
	"github.com/cloudwego/ppt-agent/pkg/agent/replanner"
	"github.com/cloudwego/ppt-agent/pkg/human"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 自动加载当前目录下的 .env 文件（WSL/Windows 都通用）
	envPath := filepath.Join(pwd, ".env")
	_ = godotenv.Load(envPath)

	ctx := context.Background()

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

	// 创建 agents
	fmt.Println("[启动] 创建 planner agent...")
	planAgent, err := planner.NewPlanner(ctx, operator, skillsContent)
	if err != nil {
		fmt.Printf("planner.NewPlanner failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] planner 创建成功")

	fmt.Println("[启动] 创建 executor agent...")
	executeAgent, err := executor.NewExecutor(ctx, operator, skillsContent)
	if err != nil {
		fmt.Printf("executor.NewExecutor failed, err: %v\n", err)
		return
	}
	fmt.Println("[启动] executor 创建成功")

	fmt.Println("[启动] 创建 replanner agent...")
	replanAgent, err := replanner.NewReplanner(ctx, operator)
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
		MaxIterations: 50,
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
	})
	fmt.Println("[启动] runner 创建成功")

	// 创建人机交互管理器
	hm := human.NewManager(interactive)

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

	fmt.Printf("\n[完成] 所有文件已保存到: %s\n", outputDir)
	time.Sleep(2 * time.Second)
}
