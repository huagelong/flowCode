package gormstore

import (
	"context"

	"anserflow/internal/model"

	"gorm.io/gorm"
)

type IssueReadStore struct {
	db *gorm.DB
}

func NewIssueReadStore(db *gorm.DB) *IssueReadStore {
	return &IssueReadStore{db: db}
}

func (s *IssueReadStore) FindIssueWithCurrentSpec(ctx context.Context, issueID uint64) (*model.Issue, *model.IssueSpec, error) {
	var issue model.Issue
	if err := s.db.WithContext(ctx).First(&issue, issueID).Error; err != nil {
		item, err := nilIfNotFound(&issue, err)
		return item, nil, err
	}

	if issue.CurrentSpecID == nil {
		return &issue, nil, nil
	}

	var spec model.IssueSpec
	if err := s.db.WithContext(ctx).First(&spec, *issue.CurrentSpecID).Error; err != nil {
		item, err := nilIfNotFound(&spec, err)
		return &issue, item, err
	}

	return &issue, &spec, nil
}

func (s *IssueReadStore) FindIssueWithLatestAttempt(ctx context.Context, issueID uint64) (*model.Issue, *model.AutomationAttempt, error) {
	var issue model.Issue
	if err := s.db.WithContext(ctx).First(&issue, issueID).Error; err != nil {
		item, err := nilIfNotFound(&issue, err)
		return item, nil, err
	}

	var attempt model.AutomationAttempt
	err := s.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("attempt_no DESC").
		First(&attempt).Error
	if err != nil {
		item, err := nilIfNotFound(&attempt, err)
		return &issue, item, err
	}

	return &issue, &attempt, nil
}
