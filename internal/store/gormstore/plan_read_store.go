package gormstore

import (
	"context"

	"anserflow/internal/model"

	"gorm.io/gorm"
)

type PlanReadStore struct {
	db *gorm.DB
}

func NewPlanReadStore(db *gorm.DB) *PlanReadStore {
	return &PlanReadStore{db: db}
}

func (s *PlanReadStore) FindPlanWithTasks(ctx context.Context, planID uint64) (*model.PlanSpec, []*model.PlanTask, error) {
	var plan model.PlanSpec
	if err := s.db.WithContext(ctx).First(&plan, planID).Error; err != nil {
		p, err := nilIfNotFound(&plan, err)
		return p, nil, err
	}

	var tasks []*model.PlanTask
	if err := s.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("seq ASC").
		Find(&tasks).Error; err != nil {
		return nil, nil, err
	}

	return &plan, tasks, nil
}
