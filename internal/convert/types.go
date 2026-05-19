package convert

type CandidateOptionData struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Supporters []string `json:"supporters"`
	Concerns   []string `json:"concerns"`
}

type DiscussionRiskData struct {
	Level string `json:"level"`
	Item  string `json:"item"`
	Owner string `json:"owner,omitempty"`
}

type ParticipantData struct {
	AgentID string `json:"agent_id"`
	Role    string `json:"role"`
	Stance  string `json:"stance,omitempty"`
}

type SelectedOptionData struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

type PlanRiskData struct {
	Level      string `json:"level"`
	Item       string `json:"item"`
	Mitigation string `json:"mitigation,omitempty"`
}

type ApprovalPolicyData struct {
	NeedsHumanReview bool   `json:"needs_human_review"`
	Reason           string `json:"reason,omitempty"`
}

type AcceptanceCheckData struct {
	Type          string `json:"type"`
	Command       string `json:"command,omitempty"`
	PassCondition string `json:"pass_condition"`
}

type RetryPolicyData struct {
	MaxAttempts     int  `json:"max_attempts"`
	AllowAutoRepair bool `json:"allow_auto_repair"`
}

type MergePolicyData struct {
	AllowAutoPR        bool `json:"allow_auto_pr"`
	AllowAutoMerge     bool `json:"allow_auto_merge"`
	RequireHumanReview bool `json:"require_human_review"`
}

type RollbackPolicyData struct {
	Strategy          string   `json:"strategy"`
	TriggerConditions []string `json:"trigger_conditions"`
}
