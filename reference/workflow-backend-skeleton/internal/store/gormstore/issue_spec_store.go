package gormstore

import (
	"context"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
)

type IssueSpecStore struct {
	db *gorm.DB
}

func NewIssueSpecStore(db *gorm.DB) *IssueSpecStore {
	return &IssueSpecStore{db: db}
}

func (s *IssueSpecStore) FindByID(ctx context.Context, id uint64) (*model.IssueSpec, error) {
	var m model.IssueSpec
	err := s.db.WithContext(ctx).First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *IssueSpecStore) FindCurrentByIssueID(ctx context.Context, issueID uint64) (*model.IssueSpec, error) {
	var m model.IssueSpec
	err := s.db.WithContext(ctx).
		Table("issue_specs").
		Joins("JOIN issues ON issues.current_spec_id = issue_specs.id").
		Where("issues.id = ?", issueID).
		Select("issue_specs.*").
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *IssueSpecStore) FindLatestByIssueID(ctx context.Context, issueID uint64) (*model.IssueSpec, error) {
	var m model.IssueSpec
	err := s.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("spec_version DESC").
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *IssueSpecStore) ListByIssueID(ctx context.Context, issueID uint64) ([]*model.IssueSpec, error) {
	var items []*model.IssueSpec
	err := s.db.WithContext(ctx).
		Where("issue_id = ?", issueID).
		Order("spec_version ASC").
		Find(&items).Error
	return items, err
}

func (s *IssueSpecStore) FindByIdempotencyKey(ctx context.Context, issueID uint64, key string) (*model.IssueSpec, error) {
	var m model.IssueSpec
	err := s.db.WithContext(ctx).
		Where("issue_id = ? AND idempotency_key = ?", issueID, key).
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *IssueSpecStore) Create(ctx context.Context, tx store.Tx, spec *model.IssueSpec) error {
	return useTx(s.db, tx).WithContext(ctx).Create(spec).Error
}
