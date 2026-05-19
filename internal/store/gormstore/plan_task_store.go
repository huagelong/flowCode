package gormstore

import (
	"context"

	"anserflow/internal/model"
	"anserflow/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlanTaskStore struct {
	db *gorm.DB
}

func NewPlanTaskStore(db *gorm.DB) *PlanTaskStore {
	return &PlanTaskStore{db: db}
}

func (s *PlanTaskStore) FindByID(ctx context.Context, id uint64) (*model.PlanTask, error) {
	var m model.PlanTask
	err := s.db.WithContext(ctx).First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *PlanTaskStore) FindByIDForUpdate(ctx context.Context, tx store.Tx, id uint64) (*model.PlanTask, error) {
	var m model.PlanTask
	err := useTx(s.db, tx).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&m, id).Error
	return nilIfNotFound(&m, err)
}

func (s *PlanTaskStore) ListByPlanID(ctx context.Context, planID uint64) ([]*model.PlanTask, error) {
	var items []*model.PlanTask
	err := s.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("seq ASC").
		Find(&items).Error
	return items, err
}

func (s *PlanTaskStore) BatchCreate(ctx context.Context, tx store.Tx, tasks []*model.PlanTask) error {
	if len(tasks) == 0 {
		return nil
	}
	return useTx(s.db, tx).WithContext(ctx).Create(&tasks).Error
}

func (s *PlanTaskStore) MarkReady(ctx context.Context, tx store.Tx, taskID uint64) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.PlanTask{}).
		Where("id = ?", taskID).
		Update("status", model.PlanTaskStatusReady).Error
}

func (s *PlanTaskStore) MarkIssued(ctx context.Context, tx store.Tx, taskID uint64, issueID uint64) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.PlanTask{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"status":            model.PlanTaskStatusIssued,
			"compiled_issue_id": issueID,
		}).Error
}

func (s *PlanTaskStore) MarkDone(ctx context.Context, tx store.Tx, taskID uint64) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.PlanTask{}).
		Where("id = ?", taskID).
		Update("status", model.PlanTaskStatusDone).Error
}

func (s *PlanTaskStore) MarkCancelled(ctx context.Context, tx store.Tx, taskID uint64) error {
	return useTx(s.db, tx).WithContext(ctx).
		Model(&model.PlanTask{}).
		Where("id = ?", taskID).
		Update("status", model.PlanTaskStatusCancelled).Error
}
