package web

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
	"github.com/cloudwego/ppt-agent/pkg/db"
	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
)

func (s *Server) handleGetUserAPIKey(c *gin.Context) {
	uid := userIDGin(c)
	key, err := db.GetUserAPIKey(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询 API Key 配置失败: " + err.Error()})
		return
	}
	if key == nil {
		c.JSON(http.StatusOK, gin.H{
			"configured":         false,
			"provider":           string(defaultModelProvider()),
			"default_provider":   string(defaultModelProvider()),
			"masked_key":         "",
			"default_configured": systemProviderKeyConfigured(defaultModelProvider()),
		})
		return
	}
	provider := modelcompat.NormalizeProvider(key.Provider)
	c.JSON(http.StatusOK, gin.H{
		"configured":         true,
		"provider":           string(provider),
		"default_provider":   string(defaultModelProvider()),
		"masked_key":         maskAPIKey(key.APIKey),
		"default_configured": systemProviderKeyConfigured(provider),
		"updated_at":         key.UpdatedAt,
	})
}

func (s *Server) handleUpdateUserAPIKey(c *gin.Context) {
	var req struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}
	apiKey := strings.TrimSpace(req.APIKey)
	if len(apiKey) < 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key 长度过短"})
		return
	}
	provider := modelcompat.NormalizeProvider(req.Provider)
	if !isSupportedAccountProvider(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的模型厂商: " + strings.TrimSpace(req.Provider)})
		return
	}
	if err := db.UpsertUserAPIKey(uint(userIDGin(c)), string(provider), apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 API Key 失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configured":         true,
		"provider":           string(provider),
		"default_provider":   string(defaultModelProvider()),
		"masked_key":         maskAPIKey(apiKey),
		"default_configured": systemProviderKeyConfigured(provider),
		"updated_at":         time.Now(),
	})
}

func (s *Server) handleDeleteUserAPIKey(c *gin.Context) {
	if err := db.DeleteUserAPIKey(uint(userIDGin(c))); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除 API Key 配置失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"configured":         false,
		"provider":           string(defaultModelProvider()),
		"default_provider":   string(defaultModelProvider()),
		"default_configured": systemProviderKeyConfigured(defaultModelProvider()),
	})
}

type modelCredential struct {
	Provider string
	APIKey   string
}

func userModelCredential(userID int) modelCredential {
	provider := defaultModelProvider()
	accountKey := ""
	if userID > 0 && db.DB != nil {
		key, err := db.GetUserAPIKey(uint(userID))
		if err != nil {
			logger.Warn("user_api_key_lookup_failed", "user_id", userID, "error", err.Error())
		} else if key != nil {
			provider = modelcompat.NormalizeProvider(key.Provider)
			accountKey = key.APIKey
		}
	}
	return modelCredential{
		Provider: string(provider),
		APIKey:   modelcompat.ResolveProviderAPIKey(provider, accountKey),
	}
}

func defaultModelProvider() modelcompat.Provider {
	for _, entry := range strings.Split(os.Getenv("MODEL_CHAIN"), ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		entryKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(entry, "-", "_"), " ", "_"))
		if provider := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_PROVIDER")); provider != "" {
			return modelcompat.NormalizeProvider(provider)
		}
	}
	if provider := strings.TrimSpace(os.Getenv("MODEL_PRIMARY_PROVIDER")); provider != "" {
		return modelcompat.NormalizeProvider(provider)
	}
	return modelcompat.NormalizeProvider(os.Getenv("MODEL_PROVIDER"))
}

func systemProviderKeyConfigured(provider modelcompat.Provider) bool {
	provider = modelcompat.NormalizeProvider(string(provider))
	if modelcompat.ResolveProviderAPIKey(provider, "") != "" {
		return true
	}
	for _, entry := range append(strings.Split(os.Getenv("MODEL_CHAIN"), ","), "primary", "text", "qa") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		entryKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(entry, "-", "_"), " ", "_"))
		entryProvider := modelcompat.NormalizeProvider(firstNonEmpty(
			os.Getenv("MODEL_"+entryKey+"_PROVIDER"),
			os.Getenv("MODEL_PROVIDER"),
		))
		if entryProvider != provider {
			continue
		}
		if strings.TrimSpace(os.Getenv("MODEL_"+entryKey+"_API_KEY")) != "" {
			return true
		}
		if keyEnv := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_API_KEY_ENV")); keyEnv != "" && strings.TrimSpace(os.Getenv(keyEnv)) != "" {
			return true
		}
	}
	return false
}

func isSupportedAccountProvider(provider modelcompat.Provider) bool {
	switch modelcompat.NormalizeProvider(string(provider)) {
	case modelcompat.ProviderArk, modelcompat.ProviderOpenAI, modelcompat.ProviderOpenAICompat,
		modelcompat.ProviderSiliconFlow, modelcompat.ProviderDeepSeek, modelcompat.ProviderQwen:
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func maskAPIKey(apiKey string) string {
	apiKey = strings.TrimSpace(apiKey)
	runes := []rune(apiKey)
	if len(runes) <= 8 {
		return strings.Repeat("*", len(runes))
	}
	prefix := string(runes[:4])
	suffix := string(runes[len(runes)-4:])
	return prefix + strings.Repeat("*", 8) + suffix
}

func runtimeEventFromRecord(record db.RuntimeEventRecord, includeMetadata bool) agentutils.RuntimeEvent {
	event := agentutils.RuntimeEvent{
		ID:        record.EventID,
		TaskID:    record.TaskID,
		Timestamp: record.Timestamp.Format(time.RFC3339Nano),
		ElapsedMS: record.ElapsedMS,
		Kind:      record.Kind,
		Phase:     record.Phase,
		Name:      record.Name,
		Status:    record.Status,
		Detail:    record.Detail,
	}
	if strings.TrimSpace(record.Metadata) != "" {
		metadata := map[string]any{}
		if err := json.Unmarshal([]byte(record.Metadata), &metadata); err == nil {
			event.Metadata = metadata
		} else {
			event.Metadata = map[string]any{"raw": record.Metadata}
		}
	}
	if !includeMetadata {
		return agentutils.RuntimeEventSummary(event)
	}
	return event
}

func runtimeEventCounts(events []agentutils.RuntimeEvent) map[string]int {
	if len(events) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, event := range events {
		kind := strings.TrimSpace(event.Kind)
		if kind == "" {
			kind = "event"
		}
		counts[kind]++
	}
	return counts
}
