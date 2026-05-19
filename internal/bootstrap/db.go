package bootstrap

import (
	"fmt"

	"anserflow/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(cfg DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open gorm db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Conversation{},
		&model.Message{},
		&model.DiscussionState{},
		&model.PlanSpec{},
		&model.PlanTask{},
		&model.Issue{},
		&model.IssueSpec{},
		&model.AutomationAttempt{},
		&model.IssueTimeline{},
		&model.AgentLog{},
	)
}
