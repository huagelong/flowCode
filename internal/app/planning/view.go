package planning

import (
	"anserflow/internal/convert"
	"anserflow/internal/model"
)

func toPlanTaskView(task *model.PlanTask) *PlanTaskView {
	if task == nil {
		return nil
	}
	dependsOn, _ := convert.DecodeJSON[[]uint64](task.DependsOnJSON)
	acceptanceOutline, _ := convert.DecodeJSON[[]string](task.AcceptanceOutlineJSON)
	return &PlanTaskView{
		ID:                task.ID,
		PlanID:            task.PlanID,
		Seq:               task.Seq,
		Title:             task.Title,
		Summary:           task.Summary,
		OwnerRole:         task.OwnerRole,
		Priority:          task.Priority,
		RiskLevel:         string(task.RiskLevel),
		DependsOn:         dependsOn,
		AcceptanceOutline: acceptanceOutline,
		Status:            string(task.Status),
		CompiledIssueID:   task.CompiledIssueID,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
	}
}

func toPlanTaskViews(tasks []*model.PlanTask) []PlanTaskView {
	out := make([]PlanTaskView, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		out = append(out, *toPlanTaskView(task))
	}
	return out
}

func toPlanView(plan *model.PlanSpec, tasks []*model.PlanTask) *PlanView {
	if plan == nil {
		return nil
	}
	scope, _ := convert.DecodeJSON[[]string](plan.ScopeJSON)
	nonGoals, _ := convert.DecodeJSON[[]string](plan.NonGoalsJSON)
	constraints, _ := convert.DecodeJSON[[]string](plan.ConstraintsJSON)
	selectedOption, _ := convert.DecodeJSON[convert.SelectedOptionData](plan.SelectedOptionJSON)
	architectureNotes, _ := convert.DecodeJSON[[]string](plan.ArchitectureNotesJSON)
	risks, _ := convert.DecodeJSON[[]convert.PlanRiskData](plan.RisksJSON)
	blockers, _ := convert.DecodeJSON[[]string](plan.BlockersJSON)
	approvalPolicy, _ := convert.DecodeJSON[convert.ApprovalPolicyData](plan.ApprovalPolicyJSON)
	return &PlanView{
		ID:                      plan.ID,
		OrgID:                   plan.OrgID,
		ConversationID:          plan.ConversationID,
		SessionID:               plan.SessionID,
		DiscussionStateID:       plan.DiscussionStateID,
		Title:                   plan.Title,
		Goal:                    plan.Goal,
		Scope:                   scope,
		NonGoals:                nonGoals,
		Constraints:             constraints,
		SelectedOption:          selectedOption,
		ArchitectureNotes:       architectureNotes,
		Risks:                   risks,
		Blockers:                blockers,
		ApprovalPolicy:          approvalPolicy,
		RiskLevel:               string(plan.RiskLevel),
		Status:                  string(plan.Status),
		SourceDiscussionVersion: plan.SourceDiscussionVersion,
		Version:                 plan.Version,
		Tasks:                   toPlanTaskViews(tasks),
		CreatedAt:               plan.CreatedAt,
		UpdatedAt:               plan.UpdatedAt,
	}
}

func toIssueSpecPreviewView(compiled *CompiledIssueSpec) *IssueSpecPreviewView {
	if compiled == nil {
		return nil
	}
	return &IssueSpecPreviewView{
		Goal:                compiled.Goal,
		Scope:               compiled.Scope,
		OutOfScope:          compiled.OutOfScope,
		TargetPaths:         compiled.TargetPaths,
		RelatedModules:      compiled.RelatedModules,
		Dependencies:        compiled.Dependencies,
		Constraints:         compiled.Constraints,
		ImplementationNotes: compiled.ImplementationNotes,
		AcceptanceChecks:    compiled.AcceptanceChecks,
		RetryPolicy:         compiled.RetryPolicy,
		MergePolicy:         compiled.MergePolicy,
		RollbackPolicy:      compiled.RollbackPolicy,
		ExecutionMode:       compiled.ExecutionMode,
		RiskLevel:           compiled.RiskLevel,
	}
}
