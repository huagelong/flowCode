package planner

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

func (s *Stub) Generate(ctx context.Context, state *model.DiscussionState) (*planning.GeneratedPlan, error) {
	return &planning.GeneratedPlan{
		Title:             "stub plan",
		Goal:              state.Goal,
		Scope:             []string{"scope-a"},
		NonGoals:          []string{"non-goal-a"},
		Constraints:       []string{},
		SelectedOption:    convert.SelectedOptionData{ID: "opt-1", Reason: "stub"},
		ArchitectureNotes: []string{},
		Risks:             []convert.PlanRiskData{},
		Blockers:          []string{},
		ApprovalPolicy:    convert.ApprovalPolicyData{NeedsHumanReview: false},
		RiskLevel:         string(model.RiskLevelLow),
		Tasks: []planning.GeneratedPlanTask{
			{
				Seq:               1,
				Title:             "stub task",
				Summary:           "stub task summary",
				OwnerRole:         "backend",
				Priority:          "p2",
				RiskLevel:         string(model.RiskLevelLow),
				DependsOn:         []uint64{},
				AcceptanceOutline: []string{"should work"},
			},
		},
	}, nil
}
