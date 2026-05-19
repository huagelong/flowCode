package convert

import (
	"fmt"

	"anserflow/internal/model"
)

func FillDiscussionStateModel(
	state *model.DiscussionState,
	topic string,
	goal string,
	latestSummary string,
	confirmedFacts []string,
	assumptions []string,
	openQuestions []string,
	candidateOptions []CandidateOptionData,
	risks []DiscussionRiskData,
	constraints []string,
	participants []ParticipantData,
	missingFields []string,
	readinessStage string,
	confidence float64,
	basedOnMessageSeq uint64,
) error {
	var err error
	state.Topic = topic
	state.Goal = goal
	state.LatestSummary = latestSummary
	state.ReadinessStage = model.DiscussionReadinessStage(readinessStage)
	state.Confidence = confidence
	state.BasedOnMessageSeq = basedOnMessageSeq
	if state.ConfirmedFactsJSON, err = EncodeJSON(confirmedFacts); err != nil {
		return fmt.Errorf("encode confirmed facts: %w", err)
	}
	if state.AssumptionsJSON, err = EncodeJSON(assumptions); err != nil {
		return fmt.Errorf("encode assumptions: %w", err)
	}
	if state.OpenQuestionsJSON, err = EncodeJSON(openQuestions); err != nil {
		return fmt.Errorf("encode open questions: %w", err)
	}
	if state.CandidateOptionsJSON, err = EncodeJSON(candidateOptions); err != nil {
		return fmt.Errorf("encode candidate options: %w", err)
	}
	if state.RisksJSON, err = EncodeJSON(risks); err != nil {
		return fmt.Errorf("encode risks: %w", err)
	}
	if state.ConstraintsJSON, err = EncodeJSON(constraints); err != nil {
		return fmt.Errorf("encode constraints: %w", err)
	}
	if state.ParticipantsJSON, err = EncodeJSON(participants); err != nil {
		return fmt.Errorf("encode participants: %w", err)
	}
	if state.MissingFieldsJSON, err = EncodeJSON(missingFields); err != nil {
		return fmt.Errorf("encode missing fields: %w", err)
	}
	return nil
}

func FillPlanSpecModel(
	plan *model.PlanSpec,
	title string,
	goal string,
	scope []string,
	nonGoals []string,
	constraints []string,
	selectedOption SelectedOptionData,
	architectureNotes []string,
	risks []PlanRiskData,
	blockers []string,
	approvalPolicy ApprovalPolicyData,
	riskLevel string,
) error {
	var err error
	plan.Title = title
	plan.Goal = goal
	plan.RiskLevel = model.RiskLevel(riskLevel)
	if plan.ScopeJSON, err = EncodeJSON(scope); err != nil {
		return fmt.Errorf("encode scope: %w", err)
	}
	if plan.NonGoalsJSON, err = EncodeJSON(nonGoals); err != nil {
		return fmt.Errorf("encode non goals: %w", err)
	}
	if plan.ConstraintsJSON, err = EncodeJSON(constraints); err != nil {
		return fmt.Errorf("encode constraints: %w", err)
	}
	if plan.SelectedOptionJSON, err = EncodeJSON(selectedOption); err != nil {
		return fmt.Errorf("encode selected option: %w", err)
	}
	if plan.ArchitectureNotesJSON, err = EncodeJSON(architectureNotes); err != nil {
		return fmt.Errorf("encode architecture notes: %w", err)
	}
	if plan.RisksJSON, err = EncodeJSON(risks); err != nil {
		return fmt.Errorf("encode risks: %w", err)
	}
	if plan.BlockersJSON, err = EncodeJSON(blockers); err != nil {
		return fmt.Errorf("encode blockers: %w", err)
	}
	if plan.ApprovalPolicyJSON, err = EncodeJSON(approvalPolicy); err != nil {
		return fmt.Errorf("encode approval policy: %w", err)
	}
	return nil
}

func FillPlanTaskModel(
	task *model.PlanTask,
	seq uint32,
	title string,
	summary string,
	ownerRole string,
	priority string,
	riskLevel string,
	dependsOn []uint64,
	acceptanceOutline []string,
) error {
	var err error
	task.Seq = seq
	task.Title = title
	task.Summary = summary
	task.OwnerRole = ownerRole
	task.Priority = priority
	task.RiskLevel = model.RiskLevel(riskLevel)
	if task.DependsOnJSON, err = EncodeJSON(dependsOn); err != nil {
		return fmt.Errorf("encode depends_on: %w", err)
	}
	if task.AcceptanceOutlineJSON, err = EncodeJSON(acceptanceOutline); err != nil {
		return fmt.Errorf("encode acceptance outline: %w", err)
	}
	return nil
}

func FillIssueSpecModel(
	spec *model.IssueSpec,
	goal string,
	scope []string,
	outOfScope []string,
	targetPaths []string,
	relatedModules []string,
	dependencies []string,
	constraints []string,
	implementationNotes []string,
	acceptanceChecks []AcceptanceCheckData,
	retryPolicy RetryPolicyData,
	mergePolicy MergePolicyData,
	rollbackPolicy RollbackPolicyData,
	executionMode string,
	riskLevel string,
) error {
	var err error
	spec.Goal = goal
	spec.ExecutionMode = model.ExecutionMode(executionMode)
	spec.RiskLevel = model.RiskLevel(riskLevel)
	if spec.ScopeJSON, err = EncodeJSON(scope); err != nil {
		return fmt.Errorf("encode scope: %w", err)
	}
	if spec.OutOfScopeJSON, err = EncodeJSON(outOfScope); err != nil {
		return fmt.Errorf("encode out_of_scope: %w", err)
	}
	if spec.TargetPathsJSON, err = EncodeJSON(targetPaths); err != nil {
		return fmt.Errorf("encode target_paths: %w", err)
	}
	if spec.RelatedModulesJSON, err = EncodeJSON(relatedModules); err != nil {
		return fmt.Errorf("encode related_modules: %w", err)
	}
	if spec.DependenciesJSON, err = EncodeJSON(dependencies); err != nil {
		return fmt.Errorf("encode dependencies: %w", err)
	}
	if spec.ConstraintsJSON, err = EncodeJSON(constraints); err != nil {
		return fmt.Errorf("encode constraints: %w", err)
	}
	if spec.ImplementationNotesJSON, err = EncodeJSON(implementationNotes); err != nil {
		return fmt.Errorf("encode implementation_notes: %w", err)
	}
	if spec.AcceptanceChecksJSON, err = EncodeJSON(acceptanceChecks); err != nil {
		return fmt.Errorf("encode acceptance_checks: %w", err)
	}
	if spec.RetryPolicyJSON, err = EncodeJSON(retryPolicy); err != nil {
		return fmt.Errorf("encode retry_policy: %w", err)
	}
	if spec.MergePolicyJSON, err = EncodeJSON(mergePolicy); err != nil {
		return fmt.Errorf("encode merge_policy: %w", err)
	}
	if spec.RollbackPolicyJSON, err = EncodeJSON(rollbackPolicy); err != nil {
		return fmt.Errorf("encode rollback_policy: %w", err)
	}
	return nil
}
