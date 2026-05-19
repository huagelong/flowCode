package execution

import (
	"context"

	"anserflow/internal/convert"
	"anserflow/internal/model"
)

type specBuilderStub struct{}

func NewSpecBuilderStub() SpecBuilder {
	return &specBuilderStub{}
}

func (b *specBuilderStub) Rebuild(ctx context.Context, issue *model.Issue, latest *model.IssueSpec, reason string) (*BuiltSpec, error) {
	scope, _ := convert.DecodeJSON[[]string](latest.ScopeJSON)
	outOfScope, _ := convert.DecodeJSON[[]string](latest.OutOfScopeJSON)
	targetPaths, _ := convert.DecodeJSON[[]string](latest.TargetPathsJSON)
	relatedModules, _ := convert.DecodeJSON[[]string](latest.RelatedModulesJSON)
	dependencies, _ := convert.DecodeJSON[[]string](latest.DependenciesJSON)
	constraints, _ := convert.DecodeJSON[[]string](latest.ConstraintsJSON)
	implementationNotes, _ := convert.DecodeJSON[[]string](latest.ImplementationNotesJSON)
	acceptanceChecks, _ := convert.DecodeJSON[[]convert.AcceptanceCheckData](latest.AcceptanceChecksJSON)
	retryPolicy, _ := convert.DecodeJSON[convert.RetryPolicyData](latest.RetryPolicyJSON)
	mergePolicy, _ := convert.DecodeJSON[convert.MergePolicyData](latest.MergePolicyJSON)
	rollbackPolicy, _ := convert.DecodeJSON[convert.RollbackPolicyData](latest.RollbackPolicyJSON)

	return &BuiltSpec{
		Goal:                latest.Goal,
		Scope:               scope,
		OutOfScope:          outOfScope,
		TargetPaths:         targetPaths,
		RelatedModules:      relatedModules,
		Dependencies:        dependencies,
		Constraints:         constraints,
		ImplementationNotes: implementationNotes,
		AcceptanceChecks:    acceptanceChecks,
		RetryPolicy:         retryPolicy,
		MergePolicy:         mergePolicy,
		RollbackPolicy:      rollbackPolicy,
		ExecutionMode:       string(latest.ExecutionMode),
		RiskLevel:           string(latest.RiskLevel),
	}, nil
}
