package discussion

import "time"

type RefreshInput struct {
	ConversationID uint64
	SessionID      string
	Force          bool
}

type FreezeInput struct {
	ConversationID uint64
	SessionID      string
}

type StateView struct {
	ID                uint64    `json:"id"`
	OrgID             uint64    `json:"org_id"`
	ConversationID    uint64    `json:"conversation_id"`
	SessionID         string    `json:"session_id"`
	Topic             string    `json:"topic"`
	Goal              string    `json:"goal"`
	LatestSummary     string    `json:"latest_summary"`
	ReadinessStage    string    `json:"readiness_stage"`
	Confidence        float64   `json:"confidence"`
	BasedOnMessageSeq uint64    `json:"based_on_message_seq"`
	Version           uint32    `json:"version"`
	ConfirmedFacts    []string  `json:"confirmed_facts"`
	Assumptions       []string  `json:"assumptions"`
	OpenQuestions     []string  `json:"open_questions"`
	CandidateOptions  any       `json:"candidate_options"`
	Risks             any       `json:"risks"`
	Constraints       []string  `json:"constraints"`
	Participants      any       `json:"participants"`
	MissingFields     []string  `json:"missing_fields"`
	UpdatedAt         time.Time `json:"updated_at"`
}
