package execution

import (
	"context"
	"errors"
	"fmt"

	"anserflow/internal/convert"
	"anserflow/internal/model"
	"anserflow/internal/store"
)

var (
	ErrExecutionNotFound = errors.New("resource not found")
	ErrSpecRequired      = errors.New("current issue spec required")
	ErrAlreadyRunning    = errors.New("issue already has running attempt")
	ErrDispatchDenied    = errors.New("issue cannot be dispatched in current state")
	ErrRetryExhausted    = errors.New("retry policy exhausted")
	ErrAutoRepairDenied  = errors.New("auto repair is not allowed")
)

type Queue interface {
	EnqueueIssueExecution(issueID uint64, issueSpecID uint64, attemptNo uint32) error
}

type Service interface {
	Dispatch(ctx context.Context, in DispatchInput) (*AttemptView, error)
	AutoRepair(ctx context.Context, in AutoRepairInput) (*AttemptView, error)
	Escalate(ctx context.Context, in EscalateInput) error
	ListAttempts(ctx context.Context, issueID uint64) ([]AttemptView, error)
	GetLatestAttempt(ctx context.Context, issueID uint64) (*AttemptView, error)
}

type service struct {
	issues    store.IssueStore
	issueRead store.IssueReadStore
	specs     store.IssueSpecStore
	attempts  store.AutomationAttemptStore
	tx        store.TxManager
	queue     Queue
}

func NewAutomationService(
	issues store.IssueStore,
	issueRead store.IssueReadStore,
	specs store.IssueSpecStore,
	attempts store.AutomationAttemptStore,
	tx store.TxManager,
	queue Queue,
) Service {
	return &service{
		issues:    issues,
		issueRead: issueRead,
		specs:     specs,
		attempts:  attempts,
		tx:        tx,
		queue:     queue,
	}
}

func (s *service) Dispatch(ctx context.Context, in DispatchInput) (*AttemptView, error) {
	var createdAttempt *model.AutomationAttempt
	var issueSpecID uint64

	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		issue, err := s.issues.FindByIDForUpdate(ctx, tx, in.IssueID)
		if err != nil {
			return err
		}
		if issue == nil {
			return ErrExecutionNotFound
		}
		if !canDispatch(issue) {
			return ErrDispatchDenied
		}
		running, err := s.attempts.FindRunningByIssueID(ctx, in.IssueID)
		if err != nil {
			return err
		}
		if running != nil {
			return ErrAlreadyRunning
		}
		spec, err := s.specs.FindCurrentByIssueID(ctx, in.IssueID)
		if err != nil {
			return err
		}
		if spec == nil {
			return ErrSpecRequired
		}
		issueSpecID = spec.ID
		attemptNo, err := s.issues.IncrementAttemptNo(ctx, tx, in.IssueID)
		if err != nil {
			return err
		}
		attempt := &model.AutomationAttempt{
			IssueID:         issue.ID,
			IssueSpecID:     spec.ID,
			AttemptNo:       attemptNo,
			TriggerType:     model.AutomationTriggerInitial,
			SandboxStrategy: model.SandboxStrategyFreshSnapshot,
			Result:          model.AutomationResultQueued,
			Summary:         "queued for execution",
		}
		if err := s.attempts.Create(ctx, tx, attempt); err != nil {
			return err
		}
		if err := s.issues.UpdateAutomationStatus(ctx, tx, issue.ID, model.AutomationStatusQueued); err != nil {
			return err
		}
		createdAttempt = attempt
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.queue.EnqueueIssueExecution(createdAttempt.IssueID, issueSpecID, createdAttempt.AttemptNo); err != nil {
		return nil, fmt.Errorf("enqueue issue execution: %w", err)
	}
	return toAttemptView(createdAttempt), nil
}

func (s *service) AutoRepair(ctx context.Context, in AutoRepairInput) (*AttemptView, error) {
	var createdAttempt *model.AutomationAttempt
	var issueSpecID uint64

	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		issue, err := s.issues.FindByIDForUpdate(ctx, tx, in.IssueID)
		if err != nil {
			return err
		}
		if issue == nil {
			return ErrExecutionNotFound
		}
		if !canAutoRepair(issue) {
			return ErrAutoRepairDenied
		}
		spec, err := s.specs.FindCurrentByIssueID(ctx, in.IssueID)
		if err != nil {
			return err
		}
		if spec == nil {
			return ErrSpecRequired
		}
		issueSpecID = spec.ID
		retryPolicy, err := convert.DecodeJSON[convert.RetryPolicyData](spec.RetryPolicyJSON)
		if err != nil {
			return err
		}
		if !retryPolicy.AllowAutoRepair {
			return ErrAutoRepairDenied
		}
		if int(issue.LastAttemptNo) >= retryPolicy.MaxAttempts {
			return ErrRetryExhausted
		}
		attemptNo, err := s.issues.IncrementAttemptNo(ctx, tx, in.IssueID)
		if err != nil {
			return err
		}
		attempt := &model.AutomationAttempt{
			IssueID:         issue.ID,
			IssueSpecID:     spec.ID,
			AttemptNo:       attemptNo,
			TriggerType:     model.AutomationTriggerAutoRepair,
			SandboxStrategy: model.SandboxStrategyReuseWorkspace,
			Result:          model.AutomationResultQueued,
			Summary:         in.Reason,
		}
		if err := s.attempts.Create(ctx, tx, attempt); err != nil {
			return err
		}
		if err := s.issues.UpdateAutomationStatus(ctx, tx, issue.ID, model.AutomationStatusQueued); err != nil {
			return err
		}
		if err := s.issues.UpdateStatus(ctx, tx, issue.ID, model.IssueStatusTodo); err != nil {
			return err
		}
		createdAttempt = attempt
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.queue.EnqueueIssueExecution(createdAttempt.IssueID, issueSpecID, createdAttempt.AttemptNo); err != nil {
		return nil, fmt.Errorf("enqueue auto repair execution: %w", err)
	}
	return toAttemptView(createdAttempt), nil
}

func (s *service) Escalate(ctx context.Context, in EscalateInput) error {
	return s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		issue, err := s.issues.FindByIDForUpdate(ctx, tx, in.IssueID)
		if err != nil {
			return err
		}
		if issue == nil {
			return ErrExecutionNotFound
		}
		if err := s.issues.UpdateAutomationStatus(ctx, tx, issue.ID, model.AutomationStatusEscalated); err != nil {
			return err
		}
		return s.issues.UpdateStatus(ctx, tx, issue.ID, model.IssueStatusTodo)
	})
}

func (s *service) ListAttempts(ctx context.Context, issueID uint64) ([]AttemptView, error) {
	issue, _, err := s.issueRead.FindIssueWithLatestAttempt(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, ErrExecutionNotFound
	}

	items, err := s.attempts.ListByIssueID(ctx, issueID)
	if err != nil {
		return nil, err
	}
	return toAttemptViews(items), nil
}

func (s *service) GetLatestAttempt(ctx context.Context, issueID uint64) (*AttemptView, error) {
	issue, item, err := s.issueRead.FindIssueWithLatestAttempt(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if issue == nil || item == nil {
		return nil, ErrExecutionNotFound
	}
	return toAttemptView(item), nil
}

func canDispatch(issue *model.Issue) bool {
	if issue == nil {
		return false
	}
	if issue.Status != model.IssueStatusTodo {
		return false
	}
	return issue.AutomationStatus == model.AutomationStatusIdle || issue.AutomationStatus == model.AutomationStatusRetryWaiting
}

func canAutoRepair(issue *model.Issue) bool {
	if issue == nil {
		return false
	}
	if issue.Status != model.IssueStatusTodo {
		return false
	}
	return issue.AutomationStatus == model.AutomationStatusRetryWaiting || issue.AutomationStatus == model.AutomationStatusEscalated
}
