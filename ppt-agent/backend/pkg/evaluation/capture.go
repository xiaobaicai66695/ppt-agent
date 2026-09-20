// Package evaluation owns private production evidence used to create reviewed
// benchmark cases. It never exposes operational task directories directly.
package evaluation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/cloudwego/ppt-agent/pkg/agent/ppt"
	"github.com/cloudwego/ppt-agent/pkg/db"
)

const (
	StagePlanner        = "planner"
	StageReviewer       = "reviewer"
	StagePlannerRefiner = "planner_refiner"
	StageFixer          = "fixer"
)

// Capturer persists stage evidence to a content-addressed root and database.
// A nil database is treated as disabled, keeping local/unit workflows intact.
type Capturer struct {
	root string
}

func NewCapturer(root string) *Capturer {
	root = strings.TrimSpace(root)
	if root == "" {
		root = strings.TrimSpace(os.Getenv("PPT_EVAL_ARTIFACT_ROOT"))
	}
	if root == "" {
		// The production deployment replaces backend source directories but
		// preserves weboutput. Keep private evaluation evidence beside delivery
		// artifacts by default; operators can override this with an absolute
		// protected volume through PPT_EVAL_ARTIFACT_ROOT.
		root = filepath.Join("..", "weboutput", "eval-data")
	}
	return &Capturer{root: root}
}

func (c *Capturer) Root() string {
	if c == nil {
		return ""
	}
	return c.root
}

func (c *Capturer) CaptureEvaluationStage(ctx context.Context, snapshot ppt.EvaluationStageSnapshot) error {
	if db.DB == nil {
		return nil
	}
	if c == nil || strings.TrimSpace(c.root) == "" {
		return fmt.Errorf("evaluation artifact root is not configured")
	}
	if strings.TrimSpace(snapshot.TaskID) == "" {
		return fmt.Errorf("evaluation stage snapshot requires task id")
	}
	if !validStage(snapshot.Stage) {
		return fmt.Errorf("unsupported evaluation stage %q", snapshot.Stage)
	}
	now := time.Now().UTC()
	sessionID := evaluationSessionID(snapshot.TaskID)
	if err := db.UpsertEvaluationSession(&db.EvaluationSession{
		ID:            sessionID,
		SourceTaskID:  snapshot.TaskID,
		UserID:        uint(snapshot.UserID),
		PrivacyStatus: "pending_redaction",
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		return fmt.Errorf("upsert evaluation session: %w", err)
	}
	if snapshot.Attempt < 1 {
		attempt, err := db.NextEvaluationStageAttempt(sessionID, snapshot.Stage)
		if err != nil {
			return fmt.Errorf("resolve evaluation stage attempt: %w", err)
		}
		snapshot.Attempt = attempt
	}

	canonicalInput := map[string]any{"stage_input": snapshot.Input}
	if messages, messageErr := db.ListConversationMessages(snapshot.TaskID); messageErr == nil && len(messages) > 0 {
		canonicalInput["conversation"] = messages
	}
	input, err := c.persistJSON(ctx, "stage_input", canonicalInput)
	if err != nil {
		return fmt.Errorf("persist stage input: %w", err)
	}
	output, err := c.persistJSON(ctx, "stage_output", snapshot.Output)
	if err != nil {
		return fmt.Errorf("persist stage output: %w", err)
	}
	provenance, err := json.Marshal(snapshot.Provenance)
	if err != nil {
		return fmt.Errorf("marshal stage provenance: %w", err)
	}
	stageID := evaluationStageID(snapshot.TaskID, snapshot.Stage, snapshot.Attempt)
	if err := db.UpsertEvaluationStageRun(&db.EvaluationStageRun{
		ID:               stageID,
		SessionID:        sessionID,
		Stage:            snapshot.Stage,
		Attempt:          snapshot.Attempt,
		ParentStageRunID: snapshot.ParentStageRunID,
		InputArtifactID:  input.ID,
		OutputArtifactID: output.ID,
		Status:           firstNonEmpty(snapshot.Status, "captured"),
		ProvenanceJSON:   string(provenance),
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		return fmt.Errorf("upsert evaluation stage run: %w", err)
	}
	if err := db.UpsertEvaluationCandidate(&db.EvaluationCandidate{
		ID:                  "candidate-" + stageID,
		SessionID:           sessionID,
		StageRunID:          stageID,
		Suite:               suiteForStage(snapshot.Stage),
		Status:              "pending_review",
		RedactionStatus:     "pending_redaction",
		SemanticFingerprint: input.SHA256,
		SplitGroup:          splitGroup(snapshot.TaskID),
		CreatedAt:           now,
		UpdatedAt:           now,
	}); err != nil {
		return fmt.Errorf("upsert evaluation candidate: %w", err)
	}
	return nil
}

func (c *Capturer) persistJSON(_ context.Context, kind string, value any) (*db.EvaluationArtifact, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	sha := hex.EncodeToString(digest[:])
	storageKey := filepath.ToSlash(filepath.Join("objects", sha[:2], sha+".json"))
	path := filepath.Join(c.root, filepath.FromSlash(storageKey))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		tmp, err := os.CreateTemp(filepath.Dir(path), ".artifact-*")
		if err != nil {
			return nil, err
		}
		tmpName := tmp.Name()
		if _, err = tmp.Write(data); err == nil {
			err = tmp.Chmod(0o600)
		}
		if closeErr := tmp.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(tmpName)
			return nil, err
		}
		if err := os.Rename(tmpName, path); err != nil && !os.IsExist(err) {
			_ = os.Remove(tmpName)
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	record := &db.EvaluationArtifact{
		ID:              sha,
		SHA256:          sha,
		StorageKey:      storageKey,
		Kind:            kind,
		ContentType:     "application/json",
		SizeBytes:       int64(len(data)),
		RedactionStatus: "pending_redaction",
		CreatedAt:       time.Now().UTC(),
	}
	if err := db.UpsertEvaluationArtifact(record); err != nil {
		return nil, err
	}
	return record, nil
}

func validStage(stage string) bool {
	switch stage {
	case StagePlanner, StageReviewer, StagePlannerRefiner, StageFixer:
		return true
	default:
		return false
	}
}

func suiteForStage(stage string) string {
	if stage == StageFixer {
		return "fixer"
	}
	if stage == StageReviewer || stage == StagePlannerRefiner {
		return "reviewer"
	}
	return "planner"
}

func splitGroup(taskID string) string {
	digest := sha256.Sum256([]byte(taskID))
	return hex.EncodeToString(digest[:])
}

func evaluationSessionID(taskID string) string {
	return "eval-" + uuid.NewSHA1(uuid.NameSpaceOID, []byte(taskID)).String()
}

func evaluationStageID(taskID, stage string, attempt int) string {
	return "eval-" + uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s:%s:%d", taskID, stage, attempt))).String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
