package planning

import "time"

type CreatePlanInput struct {
	ConversationID              uint64
	SessionID                   string
	ForceRefreshDiscussionState bool
}

type ApprovePlanInput struct {
	PlanID     uint64
	ApprovedBy uint64
	Comment    string
}

type RejectPlanInput struct {
	PlanID     uint64
	RejectedBy uint64
	Reason     string
}

type CompileTaskInput struct {
	PlanID     uint64
	PlanTaskID uint64
}

type CreateIssueInput struct {
	PlanID       uint64
	PlanTaskID   uint64
	CreatedBy    uint64
	AutoDispatch bool
}

type PlanTaskView struct {
	ID                uint64    `json:"id"`
	PlanID            uint64    `json:"plan_id"`
	Seq               uint32    `json:"seq"`
	Title             string    `json:"title"`
	Summary           string    `json:"summary"`
	OwnerRole         string    `json:"owner_role"`
	Priority          string    `json:"priority"`
	RiskLevel         string    `json:"risk_level"`
	DependsOn         []uint64  `json:"depends_on"`
	AcceptanceOutline []string  `json:"acceptance_outline"`
	Status            string    `json:"status"`
	CompiledIssueID   *uint64   `json:"compiled_issue_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type PlanView struct {
	ID                      uint64         `json:"id"`
	OrgID                   uint64         `json:"org_id"`
	ConversationID          uint64         `json:"conversation_id"`
	SessionID               string         `json:"session_id"`
	DiscussionStateID       uint64         `json:"discussion_state_id"`
	Title                   string         `json:"title"`
	Goal                    string         `json:"goal"`
	Scope                   []string       `json:"scope"`
	NonGoals                []string       `json:"non_goals"`
	Constraints             []string       `json:"constraints"`
	SelectedOption          any            `json:"selected_option"`
	ArchitectureNotes       []string       `json:"architecture_notes"`
	Risks                   any            `json:"risks"`
	Blockers                []string       `json:"blockers"`
	ApprovalPolicy          any            `json:"approval_policy"`
	RiskLevel               string         `json:"risk_level"`
	Status                  string         `json:"status"`
	SourceDiscussionVersion uint32         `json:"source_discussion_version"`
	Version                 uint32         `json:"version"`
	Tasks                   []PlanTaskView `json:"tasks"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

type IssueSpecPreviewView struct {
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
