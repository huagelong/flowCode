package gormstore

import (
	"context"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConversationStore struct {
	db *gorm.DB
}

func NewConversationStore(db *gorm.DB) *ConversationStore {
	return &ConversationStore{db: db}
}

func (s *ConversationStore) FindByID(ctx context.Context, id uint64) (*model.Conversation, error) {
	var m model.Conversation
	err := s.db.WithContext(ctx).First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *ConversationStore) FindByIDForUpdate(ctx context.Context, tx store.Tx, id uint64) (*model.Conversation, error) {
	var m model.Conversation
	err := useTx(s.db, tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *ConversationStore) UpdateCurrentSession(ctx context.Context, tx store.Tx, conversationID uint64, sessionID string) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.Conversation{}).
		Where("id = ?", conversationID).
		Update("current_session_id", sessionID).Error
}
