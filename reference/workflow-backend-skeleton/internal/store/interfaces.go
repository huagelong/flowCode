package store

import (
	"context"
	"time"

	"anserflow/internal/model"
)

type Tx interface {
	DB() any
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(tx Tx) error) error
}

type ConversationStore interface {
	FindByID(ctx context.Context, id uint64) (*model.Conversation, error)
	FindByIDForUpdate(ctx context.Context, tx Tx, id uint64) (*model.Conversation, error)
	UpdateCurrentSession(ctx context.Context, tx Tx, conversationID uint64, sessionID string) error
}

type MessageStore interface {
	ListBySession(ctx context.Context, conversationID uint64, sessionID string, limit int) ([]*model.Message, error)
	ListAfterSeq(ctx context.Context, conversationID uint64, sessionID string, afterSeq uint64, limit int) ([]*model.Message, error)
	FindLastSeq(ctx context.Context, conversationID uint64, sessionID string) (uint64, error)
	Create(ctx context.Context, tx Tx, msg *model.Message) error
}

type DiscussionStateStore interface {
	FindBySession(ctx context.Context, sessionID string) (*model.DiscussionState, error)
	FindBySessionForUpdate(ctx context.Context, tx Tx, sessionID string) (*model.DiscussionState, error)
	Create(ctx context.Context, tx Tx, state *model.DiscussionState) error
	Update(ctx context.Context, tx Tx, state *model.DiscussionState) error
	Freeze(ctx context.Context, tx Tx, sessionID string, version uint32) error
}

type PlanStore interface {
	FindByID(ctx context.Context, id uint64) (*model.PlanSpec, error)
	FindByIDForUpdate(ctx context.Context, tx Tx, id uint64) (*model.PlanSpec, error)
	FindByIdempotencyKey(ctx context.Context, orgID uint64, key string) (*model.PlanSpec, error)
	ListBySession(ctx context.Context, sessionID string) ([]*model.PlanSpec, error)
	Create(ctx context.Context, tx Tx, plan *model.PlanSpec) error
	UpdateStatus(ctx context.Context, tx Tx, planID uint64, status model.PlanStatus) error
	SetApprovedBy(ctx context.Context, tx Tx, planID uint64, userID uint64) error
	SetSupersededBy(ctx context.Context, tx Tx, planID uint64, supersededByPlanID uint64) error
}

type PlanTaskStore interface {
	FindByID(ctx context.Context, id uint64) (*model.PlanTask, error)
	FindByIDForUpdate(ctx context.Context, tx Tx, id uint64) (*model.PlanTask, error)
	ListByPlanID(ctx context.Context, planID uint64) ([]*model.PlanTask, error)
	BatchCreate(ctx context.Context, tx Tx, tasks []*model.PlanTask) error
	MarkReady(ctx context.Context, tx Tx, taskID uint64) error
	MarkIssued(ctx context.Context, tx Tx, taskID uint64, issueID uint64) error
	MarkDone(ctx context.Context, tx Tx, taskID uint64) error
	MarkCancelled(ctx context.Context, tx Tx, taskID uint64) error
}

type PlanReadStore interface {
	FindPlanWithTasks(ctx context.Context, planID uint64) (*model.PlanSpec, []*model.PlanTask, error)
}

type IssueStore interface {
	FindByID(ctx context.Context, id uint64) (*model.Issue, error)
	FindByIDForUpdate(ctx context.Context, tx Tx, id uint64) (*model.Issue, error)
	FindBySourcePlanTaskID(ctx context.Context, sourcePlanTaskID uint64) (*model.Issue, error)
	Create(ctx context.Context, tx Tx, issue *model.Issue) error
	UpdateStatus(ctx context.Context, tx Tx, issueID uint64, status model.IssueStatus) error
	UpdateAutomationStatus(ctx context.Context, tx Tx, issueID uint64, status model.AutomationStatus) error
	UpdateReviewGateStatus(ctx context.Context, tx Tx, issueID uint64, status model.ReviewGateStatus) error
	UpdateCurrentSpec(ctx context.Context, tx Tx, issueID uint64, specID uint64) error
	IncrementAttemptNo(ctx context.Context, tx Tx, issueID uint64) (uint32, error)
	SetPRURL(ctx context.Context, tx Tx, issueID uint64, prURL string) error
}

type IssueSpecStore interface {
	FindByID(ctx context.Context, id uint64) (*model.IssueSpec, error)
	FindCurrentByIssueID(ctx context.Context, issueID uint64) (*model.IssueSpec, error)
	FindLatestByIssueID(ctx context.Context, issueID uint64) (*model.IssueSpec, error)
	ListByIssueID(ctx context.Context, issueID uint64) ([]*model.IssueSpec, error)
	FindByIdempotencyKey(ctx context.Context, issueID uint64, key string) (*model.IssueSpec, error)
	Create(ctx context.Context, tx Tx, spec *model.IssueSpec) error
}

type IssueReadStore interface {
	FindIssueWithCurrentSpec(ctx context.Context, issueID uint64) (*model.Issue, *model.IssueSpec, error)
	FindIssueWithLatestAttempt(ctx context.Context, issueID uint64) (*model.Issue, *model.AutomationAttempt, error)
}

type AutomationAttemptStore interface {
	FindByID(ctx context.Context, id uint64) (*model.AutomationAttempt, error)
	FindByIssueAndAttemptNo(ctx context.Context, issueID uint64, attemptNo uint32) (*model.AutomationAttempt, error)
	FindLatestByIssueID(ctx context.Context, issueID uint64) (*model.AutomationAttempt, error)
	FindRunningByIssueID(ctx context.Context, issueID uint64) (*model.AutomationAttempt, error)
	FindByQueueTaskID(ctx context.Context, queueTaskID string) (*model.AutomationAttempt, error)
	ListByIssueID(ctx context.Context, issueID uint64) ([]*model.AutomationAttempt, error)
	Create(ctx context.Context, tx Tx, attempt *model.AutomationAttempt) error
	MarkRunning(ctx context.Context, tx Tx, issueID uint64, attemptNo uint32, workerID string, startedAt time.Time) error
	MarkSuccess(ctx context.Context, tx Tx, issueID uint64, attemptNo uint32, endedAt time.Time, summary string) error
	MarkFailed(ctx context.Context, tx Tx, issueID uint64, attemptNo uint32, failureCategory string, endedAt time.Time, summary string) error
	MarkCancelled(ctx context.Context, tx Tx, issueID uint64, attemptNo uint32, endedAt time.Time, summary string) error
}
