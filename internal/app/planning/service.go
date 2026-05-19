package planning

import (
	"context"
	"errors"
	"fmt"

	"anserflow/internal/app/discussion"
	"anserflow/internal/convert"
	"anserflow/internal/model"
	"anserflow/internal/store"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrNotPlannable      = errors.New("discussion state is not plannable")
	ErrPlanNotApproved   = errors.New("plan is not approved")
	ErrTaskNotReady      = errors.New("plan task is not ready")
	ErrTaskAlreadyIssued = errors.New("plan task already issued")
	ErrProjectRequired   = errors.New("conversation project is required")
)

type Planner interface {
	Generate(ctx context.Context, state *model.DiscussionState) (*GeneratedPlan, error)
}

type GeneratedPlan struct {
	Title             string
	Goal              string
	Scope             []string
	NonGoals          []string
	Constraints       []string
	SelectedOption    convert.SelectedOptionData
	ArchitectureNotes []string
	Risks             []convert.PlanRiskData
	Blockers          []string
	ApprovalPolicy    convert.ApprovalPolicyData
	RiskLevel         string
	Tasks             []GeneratedPlanTask
}

type GeneratedPlanTask struct {
	Seq               uint32
	Title             string
	Summary           string
	OwnerRole         string
	Priority          string
	RiskLevel         string
	DependsOn         []uint64
	AcceptanceOutline []string
}

type Compiler interface {
	Compile(ctx context.Context, plan *model.PlanSpec, task *model.PlanTask) (*CompiledIssueSpec, error)
}

type CompiledIssueSpec struct {
	Goal                string
	Scope               []string
	OutOfScope          []string
	TargetPaths         []string
	RelatedModules      []string
	Dependencies        []string
	Constraints         []string
	ImplementationNotes []string
	AcceptanceChecks    []convert.AcceptanceCheckData
	RetryPolicy         convert.RetryPolicyData
	MergePolicy         convert.MergePolicyData
	RollbackPolicy      convert.RollbackPolicyData
	ExecutionMode       string
	RiskLevel           string
}

type Dispatcher interface {
	Dispatch(ctx context.Context, issueID uint64) error
}

type Service interface {
	CreatePlan(ctx context.Context, in CreatePlanInput) (*PlanView, error)
	GetPlan(ctx context.Context, planID uint64) (*PlanView, error)
	ListPlans(ctx context.Context, sessionID string) ([]PlanView, error)
	ApprovePlan(ctx context.Context, in ApprovePlanInput) (*PlanView, error)
	RejectPlan(ctx context.Context, in RejectPlanInput) (*PlanView, error)
	ListTasks(ctx context.Context, planID uint64) ([]PlanTaskView, error)
	GetTask(ctx context.Context, planID uint64, planTaskID uint64) (*PlanTaskView, error)
	CompileTask(ctx context.Context, in CompileTaskInput) (*IssueSpecPreviewView, error)
	CreateIssue(ctx context.Context, in CreateIssueInput) (*model.Issue, *model.IssueSpec, error)
}

type service struct {
	discussions      discussion.Service
	discussionStates store.DiscussionStateStore
	conversations    store.ConversationStore
	plans            store.PlanStore
	planRead         store.PlanReadStore
	tasks            store.PlanTaskStore
	issues           store.IssueStore
	specs            store.IssueSpecStore
	tx               store.TxManager
	planner          Planner
	compiler         Compiler
	dispatcher       Dispatcher
}

func New(
	discussions discussion.Service,
	discussionStates store.DiscussionStateStore,
	conversations store.ConversationStore,
	plans store.PlanStore,
	planRead store.PlanReadStore,
	tasks store.PlanTaskStore,
	issues store.IssueStore,
	specs store.IssueSpecStore,
	tx store.TxManager,
	planner Planner,
	compiler Compiler,
	dispatcher Dispatcher,
) Service {
	return &service{
		discussions:      discussions,
		discussionStates: discussionStates,
		conversations:    conversations,
		plans:            plans,
		planRead:         planRead,
		tasks:            tasks,
		issues:           issues,
		specs:            specs,
		tx:               tx,
		planner:          planner,
		compiler:         compiler,
		dispatcher:       dispatcher,
	}
}

func (s *service) CreatePlan(ctx context.Context, in CreatePlanInput) (*PlanView, error) {
	if in.ForceRefreshDiscussionState {
		if _, err := s.discussions.Refresh(ctx, discussion.RefreshInput{
			ConversationID: in.ConversationID,
			SessionID:      in.SessionID,
			Force:          true,
		}); err != nil {
			return nil, err
		}
	}

	state, err := s.discussionStates.FindBySession(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if state == nil || state.ConversationID != in.ConversationID {
		return nil, ErrNotFound
	}
	if state.ReadinessStage != model.DiscussionStagePlannable {
		return nil, ErrNotPlannable
	}

	idemKey := buildPlanIdempotencyKey(state)
	existing, err := s.plans.FindByIdempotencyKey(ctx, state.OrgID, idemKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.GetPlan(ctx, existing.ID)
	}

	generated, err := s.planner.Generate(ctx, state)
	if err != nil {
		return nil, err
	}

	var createdPlanID uint64
	err = s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		plan := &model.PlanSpec{
			OrgID:                   state.OrgID,
			ConversationID:          state.ConversationID,
			SessionID:               state.SessionID,
			DiscussionStateID:       state.ID,
			SourceDiscussionVersion: state.Version,
			Version:                 1,
			IdempotencyKey:          &idemKey,
			Status:                  defaultPlanStatus(generated.RiskLevel),
		}
		if err := convert.FillPlanSpecModel(
			plan,
			generated.Title,
			generated.Goal,
			generated.Scope,
			generated.NonGoals,
			generated.Constraints,
			generated.SelectedOption,
			generated.ArchitectureNotes,
			generated.Risks,
			generated.Blockers,
			generated.ApprovalPolicy,
			generated.RiskLevel,
		); err != nil {
			return err
		}
		if err := s.plans.Create(ctx, tx, plan); err != nil {
			return err
		}

		tasks := make([]*model.PlanTask, 0, len(generated.Tasks))
		for _, item := range generated.Tasks {
			task := &model.PlanTask{
				PlanID: plan.ID,
				Status: model.PlanTaskStatusReady,
			}
			if err := convert.FillPlanTaskModel(
				task,
				item.Seq,
				item.Title,
				item.Summary,
				item.OwnerRole,
				item.Priority,
				item.RiskLevel,
				item.DependsOn,
				item.AcceptanceOutline,
			); err != nil {
				return err
			}
			tasks = append(tasks, task)
		}
		if err := s.tasks.BatchCreate(ctx, tx, tasks); err != nil {
			return err
		}
		createdPlanID = plan.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetPlan(ctx, createdPlanID)
}

func (s *service) GetPlan(ctx context.Context, planID uint64) (*PlanView, error) {
	plan, tasks, err := s.planRead.FindPlanWithTasks(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrNotFound
	}
	return toPlanView(plan, tasks), nil
}

func (s *service) ListPlans(ctx context.Context, sessionID string) ([]PlanView, error) {
	items, err := s.plans.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	out := make([]PlanView, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		full, err := s.GetPlan(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *full)
	}
	return out, nil
}

func (s *service) ApprovePlan(ctx context.Context, in ApprovePlanInput) (*PlanView, error) {
	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		plan, err := s.plans.FindByIDForUpdate(ctx, tx, in.PlanID)
		if err != nil {
			return err
		}
		if plan == nil {
			return ErrNotFound
		}
		if plan.Status != model.PlanStatusDraft && plan.Status != model.PlanStatusPendingReview {
			return errors.New("plan cannot be approved in current status")
		}
		if err := s.plans.UpdateStatus(ctx, tx, plan.ID, model.PlanStatusApproved); err != nil {
			return err
		}
		return s.plans.SetApprovedBy(ctx, tx, plan.ID, in.ApprovedBy)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPlan(ctx, in.PlanID)
}

func (s *service) RejectPlan(ctx context.Context, in RejectPlanInput) (*PlanView, error) {
	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		plan, err := s.plans.FindByIDForUpdate(ctx, tx, in.PlanID)
		if err != nil {
			return err
		}
		if plan == nil {
			return ErrNotFound
		}
		if plan.Status != model.PlanStatusDraft && plan.Status != model.PlanStatusPendingReview {
			return errors.New("plan cannot be rejected in current status")
		}
		return s.plans.UpdateStatus(ctx, tx, plan.ID, model.PlanStatusRejected)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPlan(ctx, in.PlanID)
}

func (s *service) ListTasks(ctx context.Context, planID uint64) ([]PlanTaskView, error) {
	plan, err := s.plans.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrNotFound
	}
	tasks, err := s.tasks.ListByPlanID(ctx, planID)
	if err != nil {
		return nil, err
	}
	return toPlanTaskViews(tasks), nil
}

func (s *service) GetTask(ctx context.Context, planID uint64, planTaskID uint64) (*PlanTaskView, error) {
	plan, err := s.plans.FindByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrNotFound
	}
	task, err := s.tasks.FindByID(ctx, planTaskID)
	if err != nil {
		return nil, err
	}
	if task == nil || task.PlanID != plan.ID {
		return nil, ErrNotFound
	}
	return toPlanTaskView(task), nil
}

func (s *service) CompileTask(ctx context.Context, in CompileTaskInput) (*IssueSpecPreviewView, error) {
	plan, err := s.plans.FindByID(ctx, in.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrNotFound
	}
	task, err := s.tasks.FindByID(ctx, in.PlanTaskID)
	if err != nil {
		return nil, err
	}
	if task == nil || task.PlanID != plan.ID {
		return nil, ErrNotFound
	}
	compiled, err := s.compiler.Compile(ctx, plan, task)
	if err != nil {
		return nil, err
	}
	return toIssueSpecPreviewView(compiled), nil
}

func (s *service) CreateIssue(ctx context.Context, in CreateIssueInput) (*model.Issue, *model.IssueSpec, error) {
	var createdIssue *model.Issue
	var createdSpec *model.IssueSpec

	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		task, err := s.tasks.FindByIDForUpdate(ctx, tx, in.PlanTaskID)
		if err != nil {
			return err
		}
		if task == nil {
			return ErrNotFound
		}
		if task.CompiledIssueID != nil {
			return ErrTaskAlreadyIssued
		}
		if task.Status != model.PlanTaskStatusReady {
			return ErrTaskNotReady
		}

		plan, err := s.plans.FindByIDForUpdate(ctx, tx, in.PlanID)
		if err != nil {
			return err
		}
		if plan == nil {
			return ErrNotFound
		}
		if plan.Status != model.PlanStatusApproved {
			return ErrPlanNotApproved
		}

		conv, err := s.conversations.FindByID(ctx, plan.ConversationID)
		if err != nil {
			return err
		}
		if conv == nil || conv.ProjectID == nil {
			return ErrProjectRequired
		}

		compiled, err := s.compiler.Compile(ctx, plan, task)
		if err != nil {
			return err
		}

		issue := &model.Issue{
			OrgID:            plan.OrgID,
			ProjectID:        *conv.ProjectID,
			SourcePlanID:     plan.ID,
			SourcePlanTaskID: task.ID,
			Title:            task.Title,
			Summary:          task.Summary,
			AssigneeType:     model.AssigneeTypeAgent,
			RiskLevel:        model.RiskLevel(compiled.RiskLevel),
			ExecutionMode:    model.ExecutionMode(compiled.ExecutionMode),
			Status:           model.IssueStatusTodo,
			ReviewGateStatus: model.ReviewGateStatusNone,
			AutomationStatus: model.AutomationStatusIdle,
			CreatedByUserID:  &in.CreatedBy,
		}
		if err := s.issues.Create(ctx, tx, issue); err != nil {
			return err
		}

		spec := &model.IssueSpec{
			IssueID:     issue.ID,
			PlanTaskID:  task.ID,
			SpecVersion: 1,
		}
		if err := convert.FillIssueSpecModel(
			spec,
			compiled.Goal,
			compiled.Scope,
			compiled.OutOfScope,
			compiled.TargetPaths,
			compiled.RelatedModules,
			compiled.Dependencies,
			compiled.Constraints,
			compiled.ImplementationNotes,
			compiled.AcceptanceChecks,
			compiled.RetryPolicy,
			compiled.MergePolicy,
			compiled.RollbackPolicy,
			compiled.ExecutionMode,
			compiled.RiskLevel,
		); err != nil {
			return err
		}
		if err := s.specs.Create(ctx, tx, spec); err != nil {
			return err
		}
		if err := s.issues.UpdateCurrentSpec(ctx, tx, issue.ID, spec.ID); err != nil {
			return err
		}
		if err := s.tasks.MarkIssued(ctx, tx, task.ID, issue.ID); err != nil {
			return err
		}

		createdIssue = issue
		createdSpec = spec
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if in.AutoDispatch && s.dispatcher != nil && createdIssue != nil {
		if err := s.dispatcher.Dispatch(ctx, createdIssue.ID); err != nil {
			return createdIssue, createdSpec, fmt.Errorf("auto dispatch issue: %w", err)
		}
	}
	return createdIssue, createdSpec, nil
}

func buildPlanIdempotencyKey(state *model.DiscussionState) string {
	return fmt.Sprintf("plan:%d:%d", state.ID, state.Version)
}

func defaultPlanStatus(risk string) model.PlanStatus {
	if risk == string(model.RiskLevelHigh) {
		return model.PlanStatusPendingReview
	}
	return model.PlanStatusApproved
}
