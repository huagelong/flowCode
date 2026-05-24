package execution

import (
	"time"

	"anserflow/internal/convert"
)

type DispatchInput struct {
	IssueID uint64
}

type AutoRepairInput struct {
	IssueID uint64
	Reason  string
}

type EscalateInput struct {
	IssueID     uint64
	EscalatedBy uint64
	Reason      string
}

type AttemptView struct {
	ID              uint64     `json:"id"`
	IssueID         uint64     `json:"issue_id"`
	IssueSpecID     uint64     `json:"issue_spec_id"`
	AttemptNo       uint32     `json:"attempt_no"`
	TriggerType     string     `json:"trigger_type"`
	SandboxStrategy string     `json:"sandbox_strategy"`
	Result          string     `json:"result"`
	FailureCategory *string    `json:"failure_category,omitempty"`
	QueueTaskID     *string    `json:"queue_task_id,omitempty"`
	WorkerID        *string    `json:"worker_id,omitempty"`
	Summary         string     `json:"summary"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type IssueView struct {
	ID               uint64  `json:"id"`
	OrgID            uint64  `json:"org_id"`
	ProjectID        uint64  `json:"project_id"`
	ParentID         *uint64 `json:"parent_id,omitempty"`
	SourcePlanID     uint64  `json:"source_plan_id"`
	SourcePlanTaskID uint64  `json:"source_plan_task_id"`
	Title            string  `json:"title"`
	Summary          string  `json:"summary"`
	AssigneeType     string  `json:"assignee_type"`
	AssigneeID       *uint64 `json:"assignee_id,omitempty"`
	RiskLevel        string  `json:"risk_level"`
	ExecutionMode    string  `json:"execution_mode"`
	Status           string  `json:"status"`
	ReviewGateStatus string  `json:"review_gate_status"`
	AutomationStatus string  `json:"automation_status"`
	CurrentSpecID    *uint64 `json:"current_spec_id,omitempty"`
	LastAttemptNo    uint32  `json:"last_attempt_no"`
	PRURL            *string `json:"pr_url,omitempty"`
	CreatedByUserID  *uint64 `json:"created_by_user_id,omitempty"`
}

type IssueDetailView struct {
	Issue       *IssueView `json:"issue"`
	CurrentSpec *SpecView  `json:"current_spec,omitempty"`
}

type SpecView struct {
	ID                  uint64   `json:"id"`
	IssueID             uint64   `json:"issue_id"`
	PlanTaskID          uint64   `json:"plan_task_id"`
	SpecVersion         uint32   `json:"spec_version"`
	RebuildReason       *string  `json:"rebuild_reason,omitempty"`
	Goal                string   `json:"goal"`
	Scope               []string `json:"scope"`
	OutOfScope          []string `json:"out_of_scope"`
	TargetPaths         []string `json:"target_paths"`
	RelatedModules      []string `json:"related_modules"`
	Dependencies        []string `json:"dependencies"`
	Constraints         []string `json:"constraints"`
	ImplementationNotes []string `json:"implementation_notes"`
	AcceptanceChecks    any      `json:"acceptance_checks"`
	RetryPolicy         any      `json:"retry_policy"`
	MergePolicy         any      `json:"merge_policy"`
	RollbackPolicy      any      `json:"rollback_policy"`
	ExecutionMode       string   `json:"execution_mode"`
	RiskLevel           string   `json:"risk_level"`
}

type RebuildSpecInput struct {
	IssueID     uint64
	TriggeredBy uint64
	Reason      string
}

type UpdateSpecInput struct {
	IssueID             uint64
	TriggeredBy         uint64
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
	Reason              string
}
