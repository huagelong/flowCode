package gormstore

import (
	"context"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IssueStore struct {
	db *gorm.DB
}

func NewIssueStore(db *gorm.DB) *IssueStore {
	return &IssueStore{db: db}
}

func (s *IssueStore) FindByID(ctx context.Context, id uint64) (*model.Issue, error) {
	var m model.Issue
	err := s.db.WithContext(ctx).First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *IssueStore) FindByIDForUpdate(ctx context.Context, tx store.Tx, id uint64) (*model.Issue, error) {
	var m model.Issue
	err := useTx(s.db, tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *IssueStore) FindBySourcePlanTaskID(ctx context.Context, sourcePlanTaskID uint64) (*model.Issue, error) {
	var m model.Issue
	err := s.db.WithContext(ctx).
		Where("source_plan_task_id = ?", sourcePlanTaskID).
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *IssueStore) Create(ctx context.Context, tx store.Tx, issue *model.Issue) error {
	return useTx(s.db, tx).WithContext(ctx).Create(issue).Error
}

func (s *IssueStore) UpdateStatus(ctx context.Context, tx store.Tx, issueID uint64, status model.IssueStatus) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.Issue{}).
		Where("id = ?", issueID).
		Update("status", status).Error
}

func (s *IssueStore) UpdateAutomationStatus(ctx context.Context, tx store.Tx, issueID uint64, status model.AutomationStatus) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.Issue{}).
		Where("id = ?", issueID).
		Update("automation_status", status).Error
}

func (s *IssueStore) UpdateReviewGateStatus(ctx context.Context, tx store.Tx, issueID uint64, status model.ReviewGateStatus) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.Issue{}).
		Where("id = ?", issueID).
		Update("review_gate_status", status).Error
}

func (s *IssueStore) UpdateCurrentSpec(ctx context.Context, tx store.Tx, issueID uint64, specID uint64) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.Issue{}).
		Where("id = ?", issueID).
		Update("current_spec_id", specID).Error
}

func (s *IssueStore) IncrementAttemptNo(ctx context.Context, tx store.Tx, issueID uint64) (uint32, error) {
	db := useTx(s.db, tx).WithContext(ctx)
	var issue model.Issue
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&issue, issueID).Error; err != nil {
		return 0, err
	}
	next := issue.LastAttemptNo + 1
	if err := db.Model(&model.Issue{}).Where("id = ?", issueID).Update("last_attempt_no", next).Error; err != nil {
		return 0, err
	}
	return next, nil
}

func (s *IssueStore) SetPRURL(ctx context.Context, tx store.Tx, issueID uint64, prURL string) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.Issue{}).
		Where("id = ?", issueID).
		Update("pr_url", prURL).Error
}
