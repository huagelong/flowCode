package gormstore

import (
	"errors"

	"anserflow/internal/store"

	"gorm.io/gorm"
)

func useTx(db *gorm.DB, txn store.Tx) *gorm.DB {
	if txn == nil {
		return db
	}
	gtx, ok := txn.(*tx)
	if !ok || gtx.db == nil {
		return db
	}
	return gtx.db
}

func nilIfNotFound[T any](v *T, err error) (*T, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return v, err
}
