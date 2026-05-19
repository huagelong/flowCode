package gormstore

import (
	"context"
	"time"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
)

type AutomationAttemptStore struct {
	db *gorm.DB
}

func NewAutomationAttemptStore(db *gorm.DB) *AutomationAttemptStore {
	return &AutomationAttemptStore{db: db}
}

func (s *AutomationAttemptStore) FindByID(ctx context.Context, id uint64) (*model.AutomationAttempt, error) {
	var m model.AutomationAttempt
	err := s.db.WithContext(ctx).First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *AutomationAttemptStore) FindByIssueAndAttemptNo(ctx context.Context, issueID uint64, attemptNo uint32) (*model.AutomationAttempt, error) {
	var m model.AutomationAttempt
	err := s.db.WithContext(ctx).
		Where("issue_id = ? AND attempt_no = ?", issueID, attemptNo).
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *AutomationAttemptStore) FindLatestByIssueID(ctx context.Context, issueID uint64) (*model.AutomationAttempt, error) {
	var m model.AutomationAttempt
	err := s.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("attempt_no DESC").
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *AutomationAttemptStore) FindRunningByIssueID(ctx context.Context, issueID uint64) (*model.AutomationAttempt, error) {
	var m model.AutomationAttempt
	err := s.db.WithContext(ctx).
		Where("issue_id = ? AND result = ?", issueID, model.AutomationResultRunning).
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *AutomationAttemptStore) FindByQueueTaskID(ctx context.Context, queueTaskID string) (*model.AutomationAttempt, error) {
	var m model.AutomationAttempt
	err := s.db.WithContext(ctx).
		Where("queue_task_id = ?", queueTaskID).
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *AutomationAttemptStore) ListByIssueID(ctx context.Context, issueID uint64) ([]*model.AutomationAttempt, error) {
	var items []*model.AutomationAttempt
	err := s.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("attempt_no ASC").
		Find(&items).Error
	return items, err
}

func (s *AutomationAttemptStore) Create(ctx context.Context, tx store.Tx, attempt *model.AutomationAttempt) error {
	return useTx(s.db, tx).WithContext(ctx).Create(attempt).Error
}

func (s *AutomationAttemptStore) MarkRunning(ctx context.Context, tx store.Tx, issueID uint64, attemptNo uint32, workerID string, startedAt time.Time) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.AutomationAttempt{}).
		Where("issue_id = ? AND attempt_no = ?", issueID, attemptNo).
		Updates(map[string]any{
			"result":     model.AutomationResultRunning,
			"worker_id":  workerID,
			"started_at": startedAt,
		}).Error
}

func (s *AutomationAttemptStore) MarkSuccess(ctx context.Context, tx store.Tx, issueID uint64, attemptNo uint32, endedAt time.Time, summary string) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.AutomationAttempt{}).
		Where("issue_id = ? AND attempt_no = ?", issueID, attemptNo).
		Updates(map[string]any{
			"result":   model.AutomationResultSuccess,
			"ended_at": endedAt,
			"summary":  summary,
		}).Error
}

func (s *AutomationAttemptStore) MarkFailed(ctx context.Context, tx store.Tx, issueID uint64, attemptNo uint32, failureCategory string, endedAt time.Time, summary string) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.AutomationAttempt{}).
		Where("issue_id = ? AND attempt_no = ?", issueID, attemptNo).
		Updates(map[string]any{
			"result":           model.AutomationResultFailed,
			"failure_category": failureCategory,
			"ended_at":         endedAt,
			"summary":          summary,
		}).Error
}

func (s *AutomationAttemptStore) MarkCancelled(ctx context.Context, tx store.Tx, issueID uint64, attemptNo uint32, endedAt time.Time, summary string) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.AutomationAttempt{}).
		Where("issue_id = ? AND attempt_no = ?", issueID, attemptNo).
		Updates(map[string]any{
			"result":   model.AutomationResultCancelled,
			"ended_at": endedAt,
			"summary":  summary,
		}).Error
}
