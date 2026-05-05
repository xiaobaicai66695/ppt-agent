package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	arkmodel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

func main() {
	pwd, _ := os.Getwd()
	envPath := filepath.Join(pwd, ".env")
	_ = godotenv.Load(envPath)

	// 优先从 backend 目录找
	_ = godotenv.Load(filepath.Join(pwd, "..", ".env"))

	apiKey := os.Getenv("ARK_API_KEY")
	baseURL := os.Getenv("ARK_BASE_URL")
	model := os.Getenv("ARK_MODEL")
	backup1 := os.Getenv("ARK_MODEL_BACKUP1")

	fmt.Println("========== 环境检测 ==========")
	fmt.Printf("ARK_API_KEY:       %s...\n", mask(apiKey, 12))
	fmt.Printf("ARK_BASE_URL:      %s\n", baseURL)
	fmt.Printf("ARK_MODEL:         %s\n", model)
	fmt.Printf("ARK_MODEL_BACKUP1: %s\n", backup1)
	fmt.Println()

	models := []string{}
	for _, m := range []string{model, backup1, os.Getenv("ARK_MODEL_BACKUP2"), os.Getenv("ARK_MODEL_BACKUP3"), os.Getenv("ARK_MODEL_BACKUP4")} {
		if m != "" {
			models = append(models, m)
		}
	}

	if len(models) == 0 {
		fmt.Println("❌ 没有配置任何模型 (ARK_MODEL 为空)")
		os.Exit(1)
	}

	fmt.Printf("将测试 %d 个模型\n\n", len(models))

	for i, m := range models {
		fmt.Printf("━━━ 测试 %d/%d: %s ━━━\n", i+1, len(models), m)
		testModel(m, apiKey, baseURL, i == 0)
		fmt.Println()
	}
}

func testModel(modelName, apiKey, baseURL string, testStream bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 自动禁用 thinking（与 app 逻辑一致）
	needThinkDisable := false
	for _, kw := range []string{"deepseek", "qwen3.5", "qwen3.6", "kimi-k2"} {
		if containsLower(modelName, kw) {
			needThinkDisable = true
			break
		}
	}

	conf := &ark.ChatModelConfig{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Model:       modelName,
		MaxTokens:   ptr(512),
		Temperature: ptr(float32(0)),
		TopP:        ptr(float32(0)),
	}
	if needThinkDisable {
		conf.Thinking = &arkmodel.Thinking{Type: arkmodel.ThinkingTypeDisabled}
		fmt.Println("  (thinking 已禁用)")
	}

	cm, err := ark.NewChatModel(ctx, conf)
	if err != nil {
		fmt.Printf("  ❌ 创建模型失败: %v\n", err)
		return
	}

	tools := []*schema.ToolInfo{
		{
			Name: "test_tool",
			Desc: "test tool for validation",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"input": {Type: schema.String, Required: true},
			}),
		},
	}

	// Test 1: 简单对话
	fmt.Println("  Test 1: 简单对话...")
	msg, err := cm.Generate(ctx, []*schema.Message{schema.UserMessage("回复 OK")})
	if err != nil {
		fmt.Printf("  ❌ 简单对话失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 简单对话: %s (tokens: %d)\n", msg.Content, msg.ResponseMeta.Usage.TotalTokens)
	}

	// Test 2: 工具调用
	fmt.Println("  Test 2: 工具调用...")
	cmWithTools, err := cm.WithTools(tools)
	if err != nil {
		fmt.Printf("  ❌ WithTools 失败: %v\n", err)
		return
	}

	msg2, err := cmWithTools.Generate(ctx, []*schema.Message{
		schema.UserMessage("帮我调用 test_tool，input 参数填 hello"),
	})
	if err != nil {
		fmt.Printf("  ❌ 工具调用失败: %v\n", err)
	} else {
		if len(msg2.ToolCalls) > 0 {
			fmt.Printf("  ✅ 工具调用成功: %d 个 tool call\n", len(msg2.ToolCalls))
			for _, tc := range msg2.ToolCalls {
				fmt.Printf("     - %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
			}
		} else {
			fmt.Printf("  ⚠️  无工具调用, 回复: %s\n", truncate(msg2.Content, 100))
		}
	}

	// Test 3: 流式（仅第一个模型）
	if testStream {
		fmt.Println("  Test 3: 流式输出...")
		stream, err := cm.Stream(ctx, []*schema.Message{schema.UserMessage("数1到5")})
		if err != nil {
			fmt.Printf("  ❌ 流式失败: %v\n", err)
		} else {
			var content string
			for {
				chunk, err := stream.Recv()
				if err != nil {
					break
				}
				content += chunk.Content
			}
			fmt.Printf("  ✅ 流式输出: %s\n", truncate(content, 200))
		}
	}
}

func mask(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "***"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func containsLower(s, sub string) bool {
	return len(s) >= len(sub) && stringContains(stringToLower(s), sub)
}

func stringToLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func ptr[T any](v T) *T { return &v }
