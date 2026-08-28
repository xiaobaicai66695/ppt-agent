package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

type planDraftResponse struct {
	ID                string    `json:"id"`
	UserID            uint      `json:"user_id"`
	ConversationID    string    `json:"conversation_id"`
	SourceMessageID   string    `json:"source_message_id"`
	Query             string    `json:"query"`
	NormalizedRequest string    `json:"normalized_request"`
	DraftContent      string    `json:"draft_content"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (s *Server) handleCreatePlanDraft(c *gin.Context) {
	var req struct {
		Query             string `json:"query"`
		NormalizedRequest string `json:"normalized_request,omitempty"`
		DraftContent      string `json:"draft_content,omitempty"`
		ConversationID    string `json:"conversation_id,omitempty"`
		SourceMessageID   string `json:"source_message_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}
	record := s.newPlanDraftRecord(userIDGin(c), req.Query, req.NormalizedRequest, req.DraftContent, req.ConversationID, req.SourceMessageID)
	if err := db.CreatePlanDraft(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存规划草稿失败"})
		return
	}
	c.JSON(http.StatusCreated, planDraftFromRecord(record))
}

func (s *Server) handleListPlanDrafts(c *gin.Context) {
	records, err := db.ListPlanDraftsByUser(uint(userIDGin(c)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取规划草稿失败"})
		return
	}
	items := make([]planDraftResponse, 0, len(records))
	for _, record := range records {
		items = append(items, planDraftFromRecord(&record))
	}
	c.JSON(http.StatusOK, gin.H{"drafts": items})
}

func (s *Server) handleGetPlanDraft(c *gin.Context) {
	record, err := db.GetPlanDraft(c.Param("id"))
	if err != nil || record == nil || int(record.UserID) != userIDGin(c) {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan draft not found"})
		return
	}
	c.JSON(http.StatusOK, planDraftFromRecord(record))
}

func (s *Server) newPlanDraftRecord(uid int, query, normalized, content, conversationID, sourceMessageID string) *db.PlanDraftRecord {
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		normalized = strings.TrimSpace(query)
	}
	content = strings.TrimSpace(content)
	if content == "" {
		content = draftPlanReply(normalized)
	}
	return &db.PlanDraftRecord{
		ID:                uuid.New().String(),
		UserID:            uint(uid),
		ConversationID:    strings.TrimSpace(conversationID),
		SourceMessageID:   strings.TrimSpace(sourceMessageID),
		Query:             strings.TrimSpace(query),
		NormalizedRequest: normalized,
		DraftContent:      content,
		Status:            "draft",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func planDraftFromRecord(record *db.PlanDraftRecord) planDraftResponse {
	if record == nil {
		return planDraftResponse{}
	}
	return planDraftResponse{
		ID:                record.ID,
		UserID:            record.UserID,
		ConversationID:    record.ConversationID,
		SourceMessageID:   record.SourceMessageID,
		Query:             record.Query,
		NormalizedRequest: record.NormalizedRequest,
		DraftContent:      record.DraftContent,
		Status:            record.Status,
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         record.UpdatedAt,
	}
}
