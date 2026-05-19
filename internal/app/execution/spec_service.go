package execution

import (
	"context"

	"anserflow/internal/convert"
	"anserflow/internal/model"
	"anserflow/internal/store"
)

type SpecBuilder interface {
	Rebuild(ctx context.Context, issue *model.Issue, latest *model.IssueSpec, reason string) (*BuiltSpec, error)
}

type BuiltSpec struct {
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

type SpecService interface {
	GetCurrent(ctx context.Context, issueID uint64) (*SpecView, error)
	List(ctx context.Context, issueID uint64) ([]SpecView, error)
	Rebuild(ctx context.Context, in RebuildSpecInput) (*SpecView, error)
	UpdateCurrent(ctx context.Context, in UpdateSpecInput) (*SpecView, error)
}

type specService struct {
	issues    store.IssueStore
	issueRead store.IssueReadStore
	specs     store.IssueSpecStore
	tx        store.TxManager
	builder   SpecBuilder
}

func NewSpecService(
	issues store.IssueStore,
	issueRead store.IssueReadStore,
	specs store.IssueSpecStore,
	tx store.TxManager,
	builder SpecBuilder,
) SpecService {
	return &specService{
		issues:    issues,
		issueRead: issueRead,
		specs:     specs,
		tx:        tx,
		builder:   builder,
	}
}

func (s *specService) GetCurrent(ctx context.Context, issueID uint64) (*SpecView, error) {
	issue, spec, err := s.issueRead.FindIssueWithCurrentSpec(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, ErrExecutionNotFound
	}
	if spec == nil {
		return nil, ErrSpecRequired
	}
	return toSpecView(spec), nil
}

func (s *specService) List(ctx context.Context, issueID uint64) ([]SpecView, error) {
	issue, _, err := s.issueRead.FindIssueWithCurrentSpec(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, ErrExecutionNotFound
	}
	items, err := s.specs.ListByIssueID(ctx, issueID)
	if err != nil {
		return nil, err
	}
	return toSpecViews(items), nil
}

func (s *specService) Rebuild(ctx context.Context, in RebuildSpecInput) (*SpecView, error) {
	var created *model.IssueSpec

	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		issue, err := s.issues.FindByIDForUpdate(ctx, tx, in.IssueID)
		if err != nil {
			return err
		}
		if issue == nil {
			return ErrExecutionNotFound
		}
		latest, err := s.specs.FindLatestByIssueID(ctx, in.IssueID)
		if err != nil {
			return err
		}
		if latest == nil {
			return ErrSpecRequired
		}
		built, err := s.builder.Rebuild(ctx, issue, latest, in.Reason)
		if err != nil {
			return err
		}
		spec := &model.IssueSpec{
			IssueID:     issue.ID,
			PlanTaskID:  latest.PlanTaskID,
			SpecVersion: latest.SpecVersion + 1,
		}
		if in.Reason != "" {
			spec.RebuildReason = &in.Reason
		}
		if err := convert.FillIssueSpecModel(
			spec,
			built.Goal,
			built.Scope,
			built.OutOfScope,
			built.TargetPaths,
			built.RelatedModules,
			built.Dependencies,
			built.Constraints,
			built.ImplementationNotes,
			built.AcceptanceChecks,
			built.RetryPolicy,
			built.MergePolicy,
			built.RollbackPolicy,
			built.ExecutionMode,
			built.RiskLevel,
		); err != nil {
			return err
		}
		if err := s.specs.Create(ctx, tx, spec); err != nil {
			return err
		}
		if err := s.issues.UpdateCurrentSpec(ctx, tx, issue.ID, spec.ID); err != nil {
			return err
		}
		created = spec
		return nil
	})
	if err != nil {
		return nil, err
	}

	return toSpecView(created), nil
}

func (s *specService) UpdateCurrent(ctx context.Context, in UpdateSpecInput) (*SpecView, error) {
	var created *model.IssueSpec

	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		issue, err := s.issues.FindByIDForUpdate(ctx, tx, in.IssueID)
		if err != nil {
			return err
		}
		if issue == nil {
			return ErrExecutionNotFound
		}
		latest, err := s.specs.FindLatestByIssueID(ctx, in.IssueID)
		if err != nil {
			return err
		}
		if latest == nil {
			return ErrSpecRequired
		}
		spec := &model.IssueSpec{
			IssueID:     issue.ID,
			PlanTaskID:  latest.PlanTaskID,
			SpecVersion: latest.SpecVersion + 1,
		}
		if in.Reason != "" {
			spec.RebuildReason = &in.Reason
		}
		if err := convert.FillIssueSpecModel(
			spec,
			in.Goal,
			in.Scope,
			in.OutOfScope,
			in.TargetPaths,
			in.RelatedModules,
			in.Dependencies,
			in.Constraints,
			in.ImplementationNotes,
			in.AcceptanceChecks,
			in.RetryPolicy,
			in.MergePolicy,
			in.RollbackPolicy,
			in.ExecutionMode,
			in.RiskLevel,
		); err != nil {
			return err
		}
		if err := s.specs.Create(ctx, tx, spec); err != nil {
			return err
		}
		if err := s.issues.UpdateCurrentSpec(ctx, tx, issue.ID, spec.ID); err != nil {
			return err
		}
		created = spec
		return nil
	})
	if err != nil {
		return nil, err
	}

	return toSpecView(created), nil
}
