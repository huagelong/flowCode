package gormstore

import (
	"context"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlanStore struct {
	db *gorm.DB
}

func NewPlanStore(db *gorm.DB) *PlanStore {
	return &PlanStore{db: db}
}

func (s *PlanStore) FindByID(ctx context.Context, id uint64) (*model.PlanSpec, error) {
	var m model.PlanSpec
	err := s.db.WithContext(ctx).First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *PlanStore) FindByIDForUpdate(ctx context.Context, tx store.Tx, id uint64) (*model.PlanSpec, error) {
	var m model.PlanSpec
	err := useTx(s.db, tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *PlanStore) FindByIdempotencyKey(ctx context.Context, orgID uint64, key string) (*model.PlanSpec, error) {
	var m model.PlanSpec
	err := s.db.WithContext(ctx).
		Where("org_id = ? AND idempotency_key = ?", orgID, key).
		First(&m).Error
	return nilIfNotFound(&m, err)
}

func (s *PlanStore) ListBySession(ctx context.Context, sessionID string) ([]*model.PlanSpec, error) {
	var items []*model.PlanSpec
	err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (s *PlanStore) Create(ctx context.Context, tx store.Tx, plan *model.PlanSpec) error {
	return useTx(s.db, tx).WithContext(ctx).Create(plan).Error
}

func (s *PlanStore) UpdateStatus(ctx context.Context, tx store.Tx, planID uint64, status model.PlanStatus) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.PlanSpec{}).
		Where("id = ?", planID).
		Update("status", status).Error
}

func (s *PlanStore) SetApprovedBy(ctx context.Context, tx store.Tx, planID uint64, userID uint64) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.PlanSpec{}).
		Where("id = ?", planID).
		Update("approved_by_user_id", userID).Error
}

func (s *PlanStore) SetSupersededBy(ctx context.Context, tx store.Tx, planID uint64, supersededByPlanID uint64) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.PlanSpec{}).
		Where("id = ?", planID).
		Updates(map[string]any{
			"status":                model.PlanStatusSuperseded,
			"superseded_by_plan_id": supersededByPlanID,
		}).Error
}
