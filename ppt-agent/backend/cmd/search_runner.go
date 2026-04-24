package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/tools/search"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("      搜索工具独立测试入口")
	fmt.Println("========================================")
	fmt.Println("输入搜索关键词即可发起真实搜索（走 Bing）")
	fmt.Println("输入 q 退出程序")
	fmt.Println()

	tool := search.NewSearchTool()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("请输入搜索关键词> ")
		if !scanner.Scan() {
			break
		}
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			fmt.Println("[提示] 关键词不能为空，请重新输入")
			continue
		}
		if query == "q" || query == "quit" || query == "exit" {
			fmt.Println("退出程序")
			break
		}

		ctx := context.Background()
		input := fmt.Sprintf(`{"query": %s}`, jsonMarshal(query))

		fmt.Println()
		fmt.Println("--- 开始搜索 ---")
		result, err := tool.InvokableRun(ctx, input)
		if err != nil {
			fmt.Printf("[错误] 调用搜索工具失败: %v\n", err)
			continue
		}

		var resp map[string]interface{}
		if err := json.Unmarshal([]byte(result), &resp); err != nil {
			fmt.Printf("[原始输出]\n%s\n", result)
			continue
		}

		if errMsg, ok := resp["error"].(string); ok {
			fmt.Printf("[错误] %s\n", errMsg)
		} else {
			content, _ := resp["content"].(string)
			fmt.Printf("%s\n", content)
		}

		if results, ok := resp["results"].([]interface{}); ok {
			fmt.Printf("\n[统计] 共返回 %d 个可信来源结果\n", len(results))
		}
		fmt.Println()
		fmt.Println("--- 搜索结束 ---\n")
	}
}

func jsonMarshal(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
