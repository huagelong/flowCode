package model

import (
	"time"

	"gorm.io/datatypes"
)

type DiscussionReadinessStage string

const (
	DiscussionStageDiscussing DiscussionReadinessStage = "discussing"
	DiscussionStageConverging DiscussionReadinessStage = "converging"
	DiscussionStagePlannable  DiscussionReadinessStage = "plannable"
	DiscussionStageFrozen     DiscussionReadinessStage = "frozen"
)

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type PlanStatus string

const (
	PlanStatusDraft         PlanStatus = "draft"
	PlanStatusPendingReview PlanStatus = "pending_review"
	PlanStatusApproved      PlanStatus = "approved"
	PlanStatusRejected      PlanStatus = "rejected"
	PlanStatusSuperseded    PlanStatus = "superseded"
)

type PlanTaskStatus string

const (
	PlanTaskStatusProposed  PlanTaskStatus = "proposed"
	PlanTaskStatusReady     PlanTaskStatus = "ready"
	PlanTaskStatusIssued    PlanTaskStatus = "issued"
	PlanTaskStatusCancelled PlanTaskStatus = "cancelled"
	PlanTaskStatusDone      PlanTaskStatus = "done"
)

type AssigneeType string

const (
	AssigneeTypeAgent AssigneeType = "agent"
	AssigneeTypeHuman AssigneeType = "human"
)

type ExecutionMode string

const (
	ExecutionModeManual   ExecutionMode = "manual"
	ExecutionModeSemiAuto ExecutionMode = "semi_auto"
	ExecutionModeAuto     ExecutionMode = "auto"
)

type IssueStatus string

const (
	IssueStatusBacklog    IssueStatus = "backlog"
	IssueStatusTodo       IssueStatus = "todo"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusPaused     IssueStatus = "paused"
	IssueStatusInReview   IssueStatus = "in_review"
	IssueStatusDone       IssueStatus = "done"
)

type ReviewGateStatus string

const (
	ReviewGateStatusNone     ReviewGateStatus = "none"
	ReviewGateStatusPending  ReviewGateStatus = "pending"
	ReviewGateStatusApproved ReviewGateStatus = "approved"
	ReviewGateStatusRejected ReviewGateStatus = "rejected"
)

type AutomationStatus string

const (
	AutomationStatusIdle         AutomationStatus = "idle"
	AutomationStatusQueued       AutomationStatus = "queued"
	AutomationStatusRunning      AutomationStatus = "running"
	AutomationStatusRetryWaiting AutomationStatus = "retry_waiting"
	AutomationStatusEscalated    AutomationStatus = "escalated"
)

type AutomationTriggerType string

const (
	AutomationTriggerInitial    AutomationTriggerType = "initial"
	AutomationTriggerAutoRepair AutomationTriggerType = "auto_repair"
)

type SandboxStrategy string

const (
	SandboxStrategyFreshSnapshot  SandboxStrategy = "fresh_snapshot"
	SandboxStrategyReuseWorkspace SandboxStrategy = "reuse_workspace"
)

type AutomationResult string

const (
	AutomationResultQueued    AutomationResult = "queued"
	AutomationResultRunning   AutomationResult = "running"
	AutomationResultSuccess   AutomationResult = "success"
	AutomationResultFailed    AutomationResult = "failed"
	AutomationResultCancelled AutomationResult = "cancelled"
)

type Conversation struct {
	ID               uint64    `gorm:"primaryKey"`
	OrgID            uint64    `gorm:"not null;index"`
	ProjectID        *uint64   `gorm:"index"`
	CurrentSessionID *string   `gorm:"size:64"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
}

type Message struct {
	ID             uint64    `gorm:"primaryKey"`
	ConversationID uint64    `gorm:"not null;index"`
	SessionID      string    `gorm:"size:64;not null;index"`
	Seq            uint64    `gorm:"not null;index"`
	Role           string    `gorm:"size:32;not null"`
	Content        string    `gorm:"type:text;not null"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

type DiscussionState struct {
	ID                   uint64                   `gorm:"primaryKey"`
	OrgID                uint64                   `gorm:"not null;index"`
	ConversationID       uint64                   `gorm:"not null;index"`
	SessionID            string                   `gorm:"size:64;not null;uniqueIndex"`
	Topic                string                   `gorm:"size:255;not null"`
	Goal                 string                   `gorm:"type:text;not null"`
	LatestSummary        string                   `gorm:"type:text;not null"`
	ReadinessStage       DiscussionReadinessStage `gorm:"size:32;not null;index"`
	Confidence           float64                  `gorm:"not null"`
	BasedOnMessageSeq    uint64                   `gorm:"not null"`
	Version              uint32                   `gorm:"not null"`
	ConfirmedFactsJSON   datatypes.JSON           `gorm:"type:json;not null"`
	AssumptionsJSON      datatypes.JSON           `gorm:"type:json;not null"`
	OpenQuestionsJSON    datatypes.JSON           `gorm:"type:json;not null"`
	CandidateOptionsJSON datatypes.JSON           `gorm:"type:json;not null"`
	RisksJSON            datatypes.JSON           `gorm:"type:json;not null"`
	ConstraintsJSON      datatypes.JSON           `gorm:"type:json;not null"`
	ParticipantsJSON     datatypes.JSON           `gorm:"type:json;not null"`
	MissingFieldsJSON    datatypes.JSON           `gorm:"type:json;not null"`
	CreatedAt            time.Time                `gorm:"not null"`
	UpdatedAt            time.Time                `gorm:"not null"`
}

type PlanSpec struct {
	ID                      uint64         `gorm:"primaryKey"`
	OrgID                   uint64         `gorm:"not null;index"`
	ConversationID          uint64         `gorm:"not null;index"`
	SessionID               string         `gorm:"size:64;not null;index"`
	DiscussionStateID       uint64         `gorm:"not null;index"`
	Title                   string         `gorm:"size:255;not null"`
	Goal                    string         `gorm:"type:text;not null"`
	ScopeJSON               datatypes.JSON `gorm:"type:json;not null"`
	NonGoalsJSON            datatypes.JSON `gorm:"type:json;not null"`
	ConstraintsJSON         datatypes.JSON `gorm:"type:json;not null"`
	SelectedOptionJSON      datatypes.JSON `gorm:"type:json;not null"`
	ArchitectureNotesJSON   datatypes.JSON `gorm:"type:json;not null"`
	RisksJSON               datatypes.JSON `gorm:"type:json;not null"`
	BlockersJSON            datatypes.JSON `gorm:"type:json;not null"`
	ApprovalPolicyJSON      datatypes.JSON `gorm:"type:json;not null"`
	RiskLevel               RiskLevel      `gorm:"size:32;not null"`
	Status                  PlanStatus     `gorm:"size:32;not null;index"`
	SourceDiscussionVersion uint32         `gorm:"not null"`
	Version                 uint32         `gorm:"not null"`
	IdempotencyKey          *string        `gorm:"size:128;index"`
	ApprovedByUserID        *uint64        `gorm:"index"`
	SupersededByPlanID      *uint64        `gorm:"index"`
	CreatedAt               time.Time      `gorm:"not null"`
	UpdatedAt               time.Time      `gorm:"not null"`
}

type PlanTask struct {
	ID                    uint64         `gorm:"primaryKey"`
	PlanID                uint64         `gorm:"not null;index"`
	Seq                   uint32         `gorm:"not null"`
	Title                 string         `gorm:"size:255;not null"`
	Summary               string         `gorm:"type:text;not null"`
	OwnerRole             string         `gorm:"size:64;not null"`
	Priority              string         `gorm:"size:32;not null"`
	RiskLevel             RiskLevel      `gorm:"size:32;not null"`
	DependsOnJSON         datatypes.JSON `gorm:"type:json;not null"`
	AcceptanceOutlineJSON datatypes.JSON `gorm:"type:json;not null"`
	Status                PlanTaskStatus `gorm:"size:32;not null;index"`
	CompiledIssueID       *uint64        `gorm:"index"`
	CreatedAt             time.Time      `gorm:"not null"`
	UpdatedAt             time.Time      `gorm:"not null"`
}

type Issue struct {
	ID               uint64           `gorm:"primaryKey"`
	OrgID            uint64           `gorm:"not null;index"`
	ProjectID        uint64           `gorm:"not null;index"`
	ParentID         *uint64          `gorm:"index"`
	SourcePlanID     uint64           `gorm:"not null;index"`
	SourcePlanTaskID uint64           `gorm:"not null;uniqueIndex"`
	Title            string           `gorm:"size:255;not null"`
	Summary          string           `gorm:"type:text;not null"`
	AssigneeType     AssigneeType     `gorm:"size:32;not null"`
	AssigneeID       *uint64          `gorm:"index"`
	RiskLevel        RiskLevel        `gorm:"size:32;not null"`
	ExecutionMode    ExecutionMode    `gorm:"size:32;not null"`
	Status           IssueStatus      `gorm:"size:32;not null;index"`
	ReviewGateStatus ReviewGateStatus `gorm:"size:32;not null"`
	AutomationStatus AutomationStatus `gorm:"size:32;not null;index"`
	CurrentSpecID    *uint64          `gorm:"index"`
	LastAttemptNo    uint32           `gorm:"not null"`
	PRURL            *string          `gorm:"size:1024"`
	CreatedByUserID  *uint64          `gorm:"index"`
	CreatedAt        time.Time        `gorm:"not null"`
	UpdatedAt        time.Time        `gorm:"not null"`
}

type IssueSpec struct {
	ID                      uint64         `gorm:"primaryKey"`
	IssueID                 uint64         `gorm:"not null;index"`
	PlanTaskID              uint64         `gorm:"not null;index"`
	SpecVersion             uint32         `gorm:"not null"`
	RebuildReason           *string        `gorm:"type:text"`
	Goal                    string         `gorm:"type:text;not null"`
	ScopeJSON               datatypes.JSON `gorm:"type:json;not null"`
	OutOfScopeJSON          datatypes.JSON `gorm:"type:json;not null"`
	TargetPathsJSON         datatypes.JSON `gorm:"type:json;not null"`
	RelatedModulesJSON      datatypes.JSON `gorm:"type:json;not null"`
	DependenciesJSON        datatypes.JSON `gorm:"type:json;not null"`
	ConstraintsJSON         datatypes.JSON `gorm:"type:json;not null"`
	ImplementationNotesJSON datatypes.JSON `gorm:"type:json;not null"`
	AcceptanceChecksJSON    datatypes.JSON `gorm:"type:json;not null"`
	RetryPolicyJSON         datatypes.JSON `gorm:"type:json;not null"`
	MergePolicyJSON         datatypes.JSON `gorm:"type:json;not null"`
	RollbackPolicyJSON      datatypes.JSON `gorm:"type:json;not null"`
	ExecutionMode           ExecutionMode  `gorm:"size:32;not null"`
	RiskLevel               RiskLevel      `gorm:"size:32;not null"`
	IdempotencyKey          *string        `gorm:"size:128;index"`
	CreatedAt               time.Time      `gorm:"not null"`
	UpdatedAt               time.Time      `gorm:"not null"`
}

type AutomationAttempt struct {
	ID              uint64                `gorm:"primaryKey"`
	IssueID         uint64                `gorm:"not null;index"`
	IssueSpecID     uint64                `gorm:"not null;index"`
	AttemptNo       uint32                `gorm:"not null"`
	TriggerType     AutomationTriggerType `gorm:"size:32;not null"`
	SandboxStrategy SandboxStrategy       `gorm:"size:32;not null"`
	Result          AutomationResult      `gorm:"size:32;not null;index"`
	FailureCategory *string               `gorm:"size:128"`
	QueueTaskID     *string               `gorm:"size:128;index"`
	WorkerID        *string               `gorm:"size:128"`
	Summary         string                `gorm:"type:text;not null"`
	StartedAt       *time.Time
	EndedAt         *time.Time
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

type IssueTimeline struct {
	ID        uint64    `gorm:"primaryKey"`
	IssueID   uint64    `gorm:"not null;index"`
	EventType string    `gorm:"size:64;not null"`
	Summary   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type AgentLog struct {
	ID        uint64    `gorm:"primaryKey"`
	IssueID   uint64    `gorm:"not null;index"`
	AttemptNo uint32    `gorm:"not null;index"`
	Level     string    `gorm:"size:32;not null"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
