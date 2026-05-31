// test_pdf_qa 手动测试 QA 模型的 PDF 支持情况。
//
// 用法：
//   go run ./cmd/test_pdf_qa batch_00.pdf
//
// 测试内容：
//   1. image_url + data:application/pdf;base64,...  （当前 qa_tool.go 使用的方式）
//   2. SiliconFlow Responses API（可能支持 document 类型）
//
// 预期结果：
//   - 400 错误   → 模型不支持 PDF，需切换为 JPEG 方案
//   - 200 + 内容 → 模型支持 PDF，当前方案可用
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultRequestTimeout = 120 * time.Second

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancel()

	// SiliconFlow ARK 兼容接口
	apiKey := "sk-wgzzmxvotvtuvmieonrhzutpyxkuhvwuslnmtinxlmwddmib"
	baseURL := "https://api.siliconflow.cn/v1"
	modelName := "deepseek-ai/DeepSeek-OCR"

	log.Printf("模型: %s", modelName)
	log.Printf("APIKey: %s", mask(apiKey))
	log.Printf("BaseURL: %s", baseURL)

	// 读取 PDF 文件
	pdfPath := "batch_00.pdf"
	if len(os.Args) > 1 {
		pdfPath = os.Args[1]
	}
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Fatalf("读取 PDF 失败: %v", err)
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)
	log.Printf("PDF: %s, 大小: %d bytes, Base64: %d chars", pdfPath, len(pdfBytes), len(pdfBase64))

	// ── 测试 1: Chat Completions API，测试 image_url + PDF ──────────────────────────
	log.Println("\n══════════════════════════════════════════")
	log.Println("测试 1: Chat Completions — image_url + application/pdf")
	log.Println("══════════════════════════════════════════")
	testImageURLPDF(ctx, modelName, apiKey, baseURL, pdfBase64)

	// ── 测试 2: SiliconFlow Responses API ──────────────────────────────────────────
	log.Println("\n══════════════════════════════════════════")
	log.Println("测试 2: SiliconFlow Responses API — application/pdf")
	log.Println("══════════════════════════════════════════")
	testResponsesAPI(ctx, modelName, apiKey, baseURL, pdfBase64)

	// ── 提示 ──────────────────────────────────────────────────────────────────────
	log.Println("\n══════════════════════════════════════════")
	log.Println("提示")
	log.Println("══════════════════════════════════════════")
	log.Printf("如果测试 1 返回 400，说明模型不支持 PDF，改用 JPEG：")
	log.Printf("  pdftoppm -r 150 -jpeg %s batch", pdfPath)
}

// testImageURLPDF 直接调 Chat Completions API，测试 image_url + PDF
func testImageURLPDF(ctx context.Context, modelName, apiKey, baseURL, pdfBase64 string) {
	reqBody := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": "请描述这张 PDF 的内容（这是第 1 页）。"},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": "data:application/pdf;base64," + pdfBase64,
						},
					},
				},
			},
		},
		"max_tokens": 512,
	}

	status, body := sendChatRequest(ctx, modelName, apiKey, baseURL, reqBody)
	log.Printf("HTTP %d", status)

	if status == http.StatusOK {
		var resp map[string]interface{}
		if err := json.Unmarshal([]byte(body), &resp); err == nil {
			if choices, ok := resp["choices"].([]any); ok && len(choices) > 0 {
				if c, ok := choices[0].(map[string]any); ok {
					if msg, ok := c["message"].(map[string]any); ok {
						content := fmt.Sprintf("%v", msg["content"])
						if len(content) > 300 {
							log.Printf("✅ 成功! 响应前 300 字:\n%s", content[:300])
						} else {
							log.Printf("✅ 成功! 响应:\n%s", content)
						}
					}
				}
			}
		}
	} else {
		// 尝试解析 JSON 错误，提取 message 字段
		var errResp map[string]any
		if je := json.Unmarshal([]byte(body), &errResp); je == nil {
			if errMsg, ok := errResp["error"].(map[string]any); ok {
				log.Printf("❌ HTTP %d — 错误码: %v, 消息: %v",
					status, errMsg["code"], errMsg["message"])
			} else if msg, ok := errResp["error"].(string); ok {
				log.Printf("❌ HTTP %d — %s", status, msg)
			} else {
				log.Printf("❌ HTTP %d — %s", status, body)
			}
		} else {
			log.Printf("❌ HTTP %d — %s", status, body)
		}
	}
}

// testResponsesAPI 调 SiliconFlow Responses API
func testResponsesAPI(ctx context.Context, modelName, apiKey, baseURL, pdfBase64 string) {
	reqBody := map[string]interface{}{
		"model": modelName,
		"input": []map[string]interface{}{
			{
				"type": "input_file",
				"file": map[string]string{
					"url":       "data:application/pdf;base64," + pdfBase64,
					"mime_type": "application/pdf",
				},
			},
		},
	}

	responsesURL := getResponsesURL(baseURL)
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", responsesURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: defaultRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("HTTP %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		if je := json.Unmarshal(body, &errResp); je == nil {
			if errMsg, ok := errResp["error"].(map[string]any); ok {
				log.Printf("❌ 错误码: %v, 消息: %v", errMsg["code"], errMsg["message"])
			} else if msg, ok := errResp["error"].(string); ok {
				log.Printf("❌ %s", msg)
			} else {
				log.Printf("❌ %s", string(body))
			}
		} else {
			log.Printf("❌ %s", string(body))
		}
	} else {
		log.Printf("✅ 成功:\n%s", string(body))
	}
}

// sendChatRequest 发送 Chat Completions 请求
func sendChatRequest(ctx context.Context, modelName, apiKey, baseURL string, reqBody map[string]interface{}) (int, string) {
	chatURL := getChatURL(baseURL)
	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return 0, fmt.Sprintf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: defaultRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func getChatURL(baseURL string) string {
	return strings.TrimSuffix(baseURL, "/") + "/chat/completions"
}

func getResponsesURL(baseURL string) string {
	return strings.TrimSuffix(baseURL, "/") + "/responses"
}

func mask(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
