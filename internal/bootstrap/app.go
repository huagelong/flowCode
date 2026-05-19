package bootstrap

import (
	"context"
	"fmt"

	discussionapp "anserflow/internal/app/discussion"
	executionapp "anserflow/internal/app/execution"
	planningapp "anserflow/internal/app/planning"
	"anserflow/internal/domain/compiler"
	"anserflow/internal/domain/orchestrator"
	"anserflow/internal/domain/planner"
	"anserflow/internal/store/gormstore"
	httptransport "anserflow/internal/transport/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	cfg    *Config
	db     *gorm.DB
	router *gin.Engine
}

func (a *App) Run() error {
	return a.router.Run(a.cfg.HTTP.Addr)
}

func NewApp(cfg *Config) (*App, error) {
	db, err := NewDB(cfg.DB)
	if err != nil {
		return nil, err
	}
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	txManager := gormstore.NewTxManager(db)

	conversationStore := gormstore.NewConversationStore(db)
	messageStore := gormstore.NewMessageStore(db)
	discussionStateStore := gormstore.NewDiscussionStateStore(db)
	planStore := gormstore.NewPlanStore(db)
	planTaskStore := gormstore.NewPlanTaskStore(db)
	planReadStore := gormstore.NewPlanReadStore(db)
	issueStore := gormstore.NewIssueStore(db)
	issueReadStore := gormstore.NewIssueReadStore(db)
	issueSpecStore := gormstore.NewIssueSpecStore(db)
	automationAttemptStore := gormstore.NewAutomationAttemptStore(db)

	orchestratorDomain := orchestrator.NewStub()
	plannerDomain := planner.NewStub()
	compilerDomain := compiler.NewStub()

	discussionService := discussionapp.New(
		conversationStore,
		messageStore,
		discussionStateStore,
		txManager,
		orchestratorDomain,
	)

	automationService := executionapp.NewAutomationService(
		issueStore,
		issueReadStore,
		issueSpecStore,
		automationAttemptStore,
		txManager,
		executionapp.NewInMemoryQueueStub(),
	)

	specService := executionapp.NewSpecService(
		issueStore,
		issueReadStore,
		issueSpecStore,
		txManager,
		executionapp.NewSpecBuilderStub(),
	)

	issueService := executionapp.NewIssueService(issueReadStore)

	planningService := planningapp.New(
		discussionService,
		discussionStateStore,
		conversationStore,
		planStore,
		planReadStore,
		planTaskStore,
		issueStore,
		issueSpecStore,
		txManager,
		plannerDomain,
		compilerDomain,
		planningDispatcherAdapter{automation: automationService},
	)

	handlers := httptransport.NewHandlers(
		discussionService,
		planningService,
		issueService,
		specService,
		automationService,
		httptransport.NewScopeChecker(
			conversationStore,
			discussionStateStore,
			planStore,
			issueStore,
		),
	)

	return &App{
		cfg:    cfg,
		db:     db,
		router: NewHTTPServer(handlers),
	}, nil
}

type planningDispatcherAdapter struct {
	automation executionapp.Service
}

func (a planningDispatcherAdapter) Dispatch(ctx context.Context, issueID uint64) error {
	_, err := a.automation.Dispatch(ctx, executionapp.DispatchInput{IssueID: issueID})
	return err
}
