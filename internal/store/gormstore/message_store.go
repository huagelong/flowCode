package gormstore

import (
	"context"
	"errors"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
)

type MessageStore struct {
	db *gorm.DB
}

func NewMessageStore(db *gorm.DB) *MessageStore {
	return &MessageStore{db: db}
}

func (s *MessageStore) ListBySession(ctx context.Context, conversationID uint64, sessionID string, limit int) ([]*model.Message, error) {
	var items []*model.Message
	q := s.db.WithContext(ctx).
		Where("conversation_id = ? AND session_id = ?", conversationID, sessionID).
		Order("seq ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&items).Error
	return items, err
}

func (s *MessageStore) ListAfterSeq(ctx context.Context, conversationID uint64, sessionID string, afterSeq uint64, limit int) ([]*model.Message, error) {
	var items []*model.Message
	q := s.db.WithContext(ctx).
		Where("conversation_id = ? AND session_id = ? AND seq > ?", conversationID, sessionID, afterSeq).
		Order("seq ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&items).Error
	return items, err
}

func (s *MessageStore) FindLastSeq(ctx context.Context, conversationID uint64, sessionID string) (uint64, error) {
	var item model.Message
	err := s.db.WithContext(ctx).
		Where("conversation_id = ? AND session_id = ?", conversationID, sessionID).
		Order("seq DESC").
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return item.Seq, nil
}

func (s *MessageStore) Create(ctx context.Context, tx store.Tx, msg *model.Message) error {
	return useTx(s.db, tx).WithContext(ctx).Create(msg).Error
}
