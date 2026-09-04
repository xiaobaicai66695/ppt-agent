package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
	"github.com/cloudwego/ppt-agent/pkg/db"
	agentutils "github.com/cloudwego/ppt-agent/pkg/runtime/model"
	webmodel "github.com/cloudwego/ppt-agent/pkg/runtime/web/model"
	webservice "github.com/cloudwego/ppt-agent/pkg/runtime/web/service"
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
	var req webmodel.UpdateAPIKeyRequest
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

type modelCredential = webmodel.ModelCredential

func userModelCredential(userID int) modelCredential {
	return webservice.ResolveUserModelCredential(userID)
}

func defaultModelProvider() modelcompat.Provider {
	return webservice.DefaultModelProvider()
}

func systemProviderKeyConfigured(provider modelcompat.Provider) bool {
	return webservice.SystemProviderKeyConfigured(provider)
}

func isSupportedAccountProvider(provider modelcompat.Provider) bool {
	return webservice.IsSupportedAccountProvider(provider)
}

func firstNonEmpty(values ...string) string {
	return webservice.FirstNonEmpty(values...)
}

func maskAPIKey(apiKey string) string {
	return webservice.MaskAPIKey(apiKey)
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
