package gormstore

import (
	"context"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DiscussionStateStore struct {
	db *gorm.DB
}

func NewDiscussionStateStore(db *gorm.DB) *DiscussionStateStore {
	return &DiscussionStateStore{db: db}
}

func (s *DiscussionStateStore) FindBySession(ctx context.Context, sessionID string) (*model.DiscussionState, error) {
	var m model.DiscussionState
	err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *DiscussionStateStore) FindBySessionForUpdate(ctx context.Context, tx store.Tx, sessionID string) (*model.DiscussionState, error) {
	var m model.DiscussionState
	err := useTx(s.db, tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ?", sessionID).
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *DiscussionStateStore) Create(ctx context.Context, tx store.Tx, state *model.DiscussionState) error {
	return useTx(s.db, tx).WithContext(ctx).Create(state).Error
}

func (s *DiscussionStateStore) Update(ctx context.Context, tx store.Tx, state *model.DiscussionState) error {
	return useTx(s.db, tx).WithContext(ctx).Save(state).Error
}

func (s *DiscussionStateStore) Freeze(ctx context.Context, tx store.Tx, sessionID string, version uint32) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.DiscussionState{}).
		Where("session_id = ? AND version = ?", sessionID, version).
		Update("readiness_stage", model.DiscussionStageFrozen).Error
}
