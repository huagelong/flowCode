package execution

import (
	"anserflow/internal/convert"
	"anserflow/internal/model"
)

func toAttemptView(attempt *model.AutomationAttempt) *AttemptView {
	if attempt == nil {
		return nil
	}
	return &AttemptView{
		ID:              attempt.ID,
		IssueID:         attempt.IssueID,
		IssueSpecID:     attempt.IssueSpecID,
		AttemptNo:       attempt.AttemptNo,
		TriggerType:     string(attempt.TriggerType),
		SandboxStrategy: string(attempt.SandboxStrategy),
		Result:          string(attempt.Result),
		FailureCategory: attempt.FailureCategory,
		QueueTaskID:     attempt.QueueTaskID,
		WorkerID:        attempt.WorkerID,
		Summary:         attempt.Summary,
		StartedAt:       attempt.StartedAt,
		EndedAt:         attempt.EndedAt,
		CreatedAt:       attempt.CreatedAt,
	}
}

func toIssueView(issue *model.Issue) *IssueView {
	if issue == nil {
		return nil
	}
	return &IssueView{
		ID:               issue.ID,
		OrgID:            issue.OrgID,
		ProjectID:        issue.ProjectID,
		ParentID:         issue.ParentID,
		SourcePlanID:     issue.SourcePlanID,
		SourcePlanTaskID: issue.SourcePlanTaskID,
		Title:            issue.Title,
		Summary:          issue.Summary,
		AssigneeType:     string(issue.AssigneeType),
		AssigneeID:       issue.AssigneeID,
		RiskLevel:        string(issue.RiskLevel),
		ExecutionMode:    string(issue.ExecutionMode),
		Status:           string(issue.Status),
		ReviewGateStatus: string(issue.ReviewGateStatus),
		AutomationStatus: string(issue.AutomationStatus),
		CurrentSpecID:    issue.CurrentSpecID,
		LastAttemptNo:    issue.LastAttemptNo,
		PRURL:            issue.PRURL,
		CreatedByUserID:  issue.CreatedByUserID,
	}
}

func toIssueDetailView(issue *model.Issue, spec *model.IssueSpec) *IssueDetailView {
	return &IssueDetailView{
		Issue:       toIssueView(issue),
		CurrentSpec: toSpecView(spec),
	}
}

func toAttemptViews(items []*model.AutomationAttempt) []AttemptView {
	out := make([]AttemptView, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, *toAttemptView(item))
	}
	return out
}

func toSpecView(spec *model.IssueSpec) *SpecView {
	if spec == nil {
		return nil
	}
	scope, _ := convert.DecodeJSON[[]string](spec.ScopeJSON)
	outOfScope, _ := convert.DecodeJSON[[]string](spec.OutOfScopeJSON)
	targetPaths, _ := convert.DecodeJSON[[]string](spec.TargetPathsJSON)
	relatedModules, _ := convert.DecodeJSON[[]string](spec.RelatedModulesJSON)
	dependencies, _ := convert.DecodeJSON[[]string](spec.DependenciesJSON)
	constraints, _ := convert.DecodeJSON[[]string](spec.ConstraintsJSON)
	implementationNotes, _ := convert.DecodeJSON[[]string](spec.ImplementationNotesJSON)
	acceptanceChecks, _ := convert.DecodeJSON[[]convert.AcceptanceCheckData](spec.AcceptanceChecksJSON)
	retryPolicy, _ := convert.DecodeJSON[convert.RetryPolicyData](spec.RetryPolicyJSON)
	mergePolicy, _ := convert.DecodeJSON[convert.MergePolicyData](spec.MergePolicyJSON)
	rollbackPolicy, _ := convert.DecodeJSON[convert.RollbackPolicyData](spec.RollbackPolicyJSON)
	return &SpecView{
		ID:                  spec.ID,
		IssueID:             spec.IssueID,
		PlanTaskID:          spec.PlanTaskID,
		SpecVersion:         spec.SpecVersion,
		RebuildReason:       spec.RebuildReason,
		Goal:                spec.Goal,
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
		ExecutionMode:       string(spec.ExecutionMode),
		RiskLevel:           string(spec.RiskLevel),
	}
}

func toSpecViews(items []*model.IssueSpec) []SpecView {
	out := make([]SpecView, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, *toSpecView(item))
	}
	return out
}
