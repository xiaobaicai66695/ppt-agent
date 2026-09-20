package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

func (s *Server) handleAdminEvaluationCandidates(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	candidates, err := db.ListEvaluationCandidates(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询评测候选失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"candidates": candidates})
}

func (s *Server) handleAdminEvaluationCandidateEvidence(c *gin.Context) {
	evidence, err := s.evaluation.CandidateEvidence(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评测候选证据不存在或校验失败"})
		return
	}
	c.JSON(http.StatusOK, evidence)
}

func (s *Server) handleAdminApproveEvaluationCandidate(c *gin.Context) {
	var req struct {
		SelectionReason string `json:"selection_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的候选审核请求"})
		return
	}
	if err := db.ApproveEvaluationCandidate(c.Param("id"), req.SelectionReason); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleAdminWithdrawEvaluationCandidate(c *gin.Context) {
	if err := db.WithdrawEvaluationCandidate(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleAdminCreateEvaluationCase(c *gin.Context) {
	var req struct {
		Suite      string          `json:"suite"`
		Case       json.RawMessage `json:"case"`
		SplitGroup string          `json:"split_group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !json.Valid(req.Case) || strings.TrimSpace(req.Suite) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评测用例必须包含 suite 和合法 case JSON"})
		return
	}
	if req.Suite == "router" || (req.Suite != "planner" && req.Suite != "reviewer" && req.Suite != "fixer") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "线上评测仅支持 planner、reviewer 或 fixer"})
		return
	}
	record := &db.EvaluationCaseRevision{ID: "case-" + uuid.NewString(), CandidateID: c.Param("id"), Suite: req.Suite, CaseJSON: string(req.Case), SplitGroup: strings.TrimSpace(req.SplitGroup)}
	if err := db.CreateEvaluationCaseRevision(record); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, record)
}

func (s *Server) handleAdminPublishEvaluationDataset(c *gin.Context) {
	var req struct {
		Name            string   `json:"name"`
		Role            string   `json:"role"`
		ParentID        string   `json:"parent_id"`
		CaseRevisionIDs []string `json:"case_revision_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的数据集发布请求"})
		return
	}
	record, err := s.evaluation.PublishDataset(req.Name, req.Role, req.ParentID, req.CaseRevisionIDs)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, record)
}

func (s *Server) handleAdminCreateEvaluationExport(c *gin.Context) {
	record, err := s.evaluation.BuildExport(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"export": record, "download_path": "/api/admin/evaluations/exports/" + record.ID + "/download"})
}

func (s *Server) handleAdminDownloadEvaluationExport(c *gin.Context) {
	path, record, err := s.evaluation.ExportPath(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评测数据集导出包不存在或已撤销"})
		return
	}
	c.Header("X-Evaluation-Export-SHA256", record.SHA256)
	c.FileAttachment(path, record.ID+".zip")
}
