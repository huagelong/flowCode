package gormstore

import (
	"context"

	"anserflow/internal/store"

	"gorm.io/gorm"
)

type tx struct {
	db *gorm.DB
}

func (t *tx) DB() any {
	return t.db
}

type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) WithinTransaction(ctx context.Context, fn func(tx store.Tx) error) error {
	return m.db.WithContext(ctx).Transaction(func(gdb *gorm.DB) error {
		return fn(&tx{db: gdb})
	})
}
