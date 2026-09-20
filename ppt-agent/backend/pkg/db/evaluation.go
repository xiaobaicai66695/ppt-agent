package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpsertEvaluationSession(record *EvaluationSession) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "source_task_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"updated_at": time.Now(),
		}),
	}).Create(record).Error
}

func UpsertEvaluationCandidate(record *EvaluationCandidate) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "stage_run_id"}},
		DoNothing: true,
	}).Create(record).Error
}

func ListEvaluationCandidates(limit int) ([]EvaluationCandidate, error) {
	if DB == nil {
		return []EvaluationCandidate{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var records []EvaluationCandidate
	err := DB.WithContext(ctx).Order("updated_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

func GetEvaluationCandidate(id string) (*EvaluationCandidate, error) {
	if DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var record EvaluationCandidate
	if err := DB.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func GetEvaluationStageRun(id string) (*EvaluationStageRun, error) {
	if DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var record EvaluationStageRun
	if err := DB.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func GetEvaluationArtifact(id string) (*EvaluationArtifact, error) {
	if DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var record EvaluationArtifact
	if err := DB.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func ApproveEvaluationCandidate(id, selectionReason string) error {
	if DB == nil {
		return fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	result := DB.WithContext(ctx).Model(&EvaluationCandidate{}).Where("id = ? AND withdrawn_at IS NULL", id).Updates(map[string]any{
		"status": "approved", "redaction_status": "approved", "selection_reason": strings.TrimSpace(selectionReason), "updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("evaluation candidate not found or withdrawn")
	}
	return nil
}

func WithdrawEvaluationCandidate(id string) error {
	if DB == nil {
		return fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	now := time.Now()
	result := DB.WithContext(ctx).Model(&EvaluationCandidate{}).Where("id = ?", id).Updates(map[string]any{
		"withdrawn_at": now, "status": "withdrawn", "redaction_status": "withdrawn", "updated_at": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("evaluation candidate not found")
	}
	return nil
}

func CreateEvaluationCaseRevision(record *EvaluationCaseRevision) error {
	if DB == nil {
		return fmt.Errorf("database unavailable")
	}
	if !json.Valid([]byte(record.CaseJSON)) {
		return fmt.Errorf("case_json must be valid JSON")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var candidate EvaluationCandidate
		if err := tx.Where("id = ?", record.CandidateID).First(&candidate).Error; err != nil {
			return err
		}
		if candidate.Status != "approved" || candidate.RedactionStatus != "approved" || candidate.WithdrawnAt != nil {
			return fmt.Errorf("candidate must be approved and not withdrawn")
		}
		if record.SplitGroup == "" {
			record.SplitGroup = candidate.SplitGroup
		}
		if record.Suite == "" {
			record.Suite = candidate.Suite
		} else if record.Suite != candidate.Suite {
			return fmt.Errorf("case suite must match candidate suite")
		}
		var payload struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(record.CaseJSON), &payload); err != nil || strings.TrimSpace(payload.ID) == "" {
			return fmt.Errorf("case_json must contain a non-empty id")
		}
		if record.Revision <= 0 {
			var latest EvaluationCaseRevision
			if err := tx.Where("candidate_id = ?", record.CandidateID).Order("revision DESC").First(&latest).Error; err == nil {
				record.Revision = latest.Revision + 1
			} else if isNotFound(err) {
				record.Revision = 1
			} else {
				return err
			}
		}
		record.ApprovalState = "approved"
		return tx.Create(record).Error
	})
}

func CreateEvaluationDatasetVersion(record *EvaluationDatasetVersion, caseRevisionIDs []string) error {
	if DB == nil {
		return fmt.Errorf("database unavailable")
	}
	if record.Name == "" || !validEvaluationDatasetRole(record.Role) || len(caseRevisionIDs) == 0 {
		return fmt.Errorf("dataset requires name, valid role, and at least one case revision")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var revisions []EvaluationCaseRevision
		if err := tx.Where("id IN ? AND approval_state = ?", caseRevisionIDs, "approved").Find(&revisions).Error; err != nil {
			return err
		}
		if len(revisions) != len(caseRevisionIDs) {
			return fmt.Errorf("all dataset members must be approved case revisions")
		}
		if record.Role == "validation" {
			groups := make([]string, 0, len(revisions))
			for _, revision := range revisions {
				groups = append(groups, revision.SplitGroup)
			}
			var exposed int64
			if err := tx.Table("evaluation_dataset_members AS member").
				Joins("JOIN evaluation_dataset_versions AS dataset ON dataset.id = member.dataset_version_id").
				Where("member.split_group IN ? AND dataset.role IN ? AND dataset.state = ?", groups, []string{"core", "active-dev"}, "frozen").Count(&exposed).Error; err != nil {
				return err
			}
			if exposed > 0 {
				return fmt.Errorf("validation dataset cannot reuse an exposed development split group")
			}
		}
		record.State = "frozen"
		now := time.Now()
		record.FrozenAt = &now
		record.CreatedAt = now
		record.UpdatedAt = now
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		for ordinal, revision := range revisions {
			if err := tx.Create(&EvaluationDatasetMember{DatasetVersionID: record.ID, CaseRevisionID: revision.ID, SplitGroup: revision.SplitGroup, Ordinal: ordinal + 1, CreatedAt: now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func GetEvaluationDatasetVersion(id string) (*EvaluationDatasetVersion, error) {
	if DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var record EvaluationDatasetVersion
	if err := DB.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func ListEvaluationDatasetCases(datasetID string) ([]EvaluationCaseRevision, error) {
	if DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var records []EvaluationCaseRevision
	err := DB.WithContext(ctx).Table("evaluation_case_revisions AS revision").
		Select("revision.*").Joins("JOIN evaluation_dataset_members AS member ON member.case_revision_id = revision.id").
		Where("member.dataset_version_id = ?", datasetID).Order("member.ordinal ASC").Find(&records).Error
	return records, err
}

func CreateEvaluationExport(record *EvaluationExport) error {
	if DB == nil {
		return fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Create(record).Error
}

func GetEvaluationExport(id string) (*EvaluationExport, error) {
	if DB == nil {
		return nil, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var record EvaluationExport
	if err := DB.WithContext(ctx).Where("id = ? AND revoked_at IS NULL", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func EvaluationDatasetHasWithdrawnEvidence(datasetID string) (bool, error) {
	if DB == nil {
		return true, fmt.Errorf("database unavailable")
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var count int64
	err := DB.WithContext(ctx).Table("evaluation_dataset_members AS member").
		Joins("JOIN evaluation_case_revisions AS revision ON revision.id = member.case_revision_id").
		Joins("JOIN evaluation_candidates AS candidate ON candidate.id = revision.candidate_id").
		Joins("JOIN evaluation_sessions AS session ON session.id = candidate.session_id").
		Where("member.dataset_version_id = ? AND (candidate.withdrawn_at IS NOT NULL OR session.withdrawn_at IS NOT NULL)", datasetID).
		Count(&count).Error
	return count > 0, err
}

func validEvaluationDatasetRole(role string) bool {
	switch role {
	case "core", "active-dev", "validation":
		return true
	default:
		return false
	}
}

func UpsertEvaluationArtifact(record *EvaluationArtifact) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sha256"}},
		DoNothing: true,
	}).Create(record).Error
}

func UpsertEvaluationStageRun(record *EvaluationStageRun) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "session_id"}, {Name: "stage"}, {Name: "attempt"}},
		DoUpdates: clause.Assignments(map[string]any{
			"parent_stage_run_id": record.ParentStageRunID,
			"input_artifact_id":   record.InputArtifactID,
			"output_artifact_id":  record.OutputArtifactID,
			"status":              record.Status,
			"provenance_json":     record.ProvenanceJSON,
			"capture_error":       record.CaptureError,
			"updated_at":          time.Now(),
		}),
	}).Create(record).Error
}

func NextEvaluationStageAttempt(sessionID, stage string) (int, error) {
	if DB == nil {
		return 1, nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	var latest EvaluationStageRun
	err := DB.WithContext(ctx).Where("session_id = ? AND stage = ?", sessionID, stage).Order("attempt DESC").First(&latest).Error
	if isNotFound(err) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return latest.Attempt + 1, nil
}

func UpdateEvaluationSessionOutcome(taskID, outcomeJSON string) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session EvaluationSession
		if err := tx.Where("source_task_id = ?", taskID).First(&session).Error; err != nil {
			if isNotFound(err) {
				return nil
			}
			return err
		}
		merged := map[string]any{}
		_ = json.Unmarshal([]byte(session.OutcomeJSON), &merged)
		var update map[string]any
		if err := json.Unmarshal([]byte(outcomeJSON), &update); err != nil {
			return err
		}
		for key, value := range update {
			merged[key] = value
		}
		data, err := json.Marshal(merged)
		if err != nil {
			return err
		}
		return tx.Model(&EvaluationSession{}).Where("id = ?", session.ID).Updates(map[string]any{
			"outcome_json": string(data),
			"updated_at":   time.Now(),
		}).Error
	})
}

func WithdrawEvaluationSession(taskID string) error {
	if DB == nil {
		return nil
	}
	ctx, cancel := withOperationTimeout()
	defer cancel()
	now := time.Now()
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session EvaluationSession
		if err := tx.Where("source_task_id = ?", taskID).First(&session).Error; err != nil {
			if isNotFound(err) {
				return nil
			}
			return err
		}
		if err := tx.Model(&EvaluationSession{}).Where("id = ?", session.ID).Updates(map[string]any{"withdrawn_at": now, "privacy_status": "withdrawn", "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&EvaluationCandidate{}).Where("session_id = ?", session.ID).Updates(map[string]any{"withdrawn_at": now, "status": "withdrawn", "redaction_status": "withdrawn", "updated_at": now}).Error
	})
}

func isNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
