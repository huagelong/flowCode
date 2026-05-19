package compiler

import (
	"context"

	"anserflow/internal/app/planning"
	"anserflow/internal/convert"
	"anserflow/internal/model"
)

type Stub struct{}

func NewStub() *Stub {
	return &Stub{}
}

func (s *Stub) Compile(ctx context.Context, plan *model.PlanSpec, task *model.PlanTask) (*planning.CompiledIssueSpec, error) {
	return &planning.CompiledIssueSpec{
		Goal:                task.Title,
		Scope:               []string{"stub-scope"},
		OutOfScope:          []string{},
		TargetPaths:         []string{"internal/"},
		RelatedModules:      []string{},
		Dependencies:        []string{},
		Constraints:         []string{},
		ImplementationNotes: []string{"stub note"},
		AcceptanceChecks: []convert.AcceptanceCheckData{
			{
				Type:          "custom",
				PassCondition: "manual stub pass",
			},
		},
		RetryPolicy: convert.RetryPolicyData{
			MaxAttempts:     3,
			AllowAutoRepair: true,
		},
		MergePolicy: convert.MergePolicyData{
			AllowAutoPR:        true,
			AllowAutoMerge:     false,
			RequireHumanReview: true,
		},
		RollbackPolicy: convert.RollbackPolicyData{
			Strategy:          "manual",
			TriggerConditions: []string{},
		},
		ExecutionMode: string(model.ExecutionModeSemiAuto),
		RiskLevel:     string(model.RiskLevelLow),
	}, nil
}
