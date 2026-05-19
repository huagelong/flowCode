package http

import (
	discussionapp "anserflow/internal/app/discussion"
	executionapp "anserflow/internal/app/execution"
	planningapp "anserflow/internal/app/planning"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Discussion *DiscussionHandler
	Plan       *PlanHandler
	PlanTask   *PlanTaskHandler
	Issue      *IssueHandler
	IssueSpec  *IssueSpecHandler
	Automation *AutomationHandler
}

func NewHandlers(
	discussionApp discussionapp.Service,
	planningApp planningapp.Service,
	issueApp executionapp.IssueService,
	specApp executionapp.SpecService,
	automationApp executionapp.Service,
	scope ScopeChecker,
) *Handlers {
	return &Handlers{
		Discussion: NewDiscussionHandler(discussionApp, scope),
		Plan:       NewPlanHandler(planningApp, scope),
		PlanTask:   NewPlanTaskHandler(planningApp, scope),
		Issue:      NewIssueHandler(issueApp, scope),
		IssueSpec:  NewIssueSpecHandler(specApp, scope),
		Automation: NewAutomationHandler(automationApp, scope),
	}
}

func RegisterRoutes(r *gin.Engine, h *Handlers) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	api := r.Group("/api/orgs/:org_id")
	{
		conv := api.Group("/conversations/:conversation_id/sessions/:session_id")
		{
			conv.GET("/discussion-state", h.Discussion.Get)
			conv.POST("/discussion-state/refresh", h.Discussion.Refresh)
			conv.POST("/discussion-state/freeze", h.Discussion.Freeze)
			conv.POST("/plans", h.Plan.Create)
		}

		plans := api.Group("/plans")
		{
			plans.GET("", h.Plan.ListBySession)
			plans.GET("/:plan_id", h.Plan.Get)
			plans.POST("/:plan_id/approve", h.Plan.Approve)
			plans.POST("/:plan_id/reject", h.Plan.Reject)
			plans.GET("/:plan_id/tasks", h.PlanTask.List)
			plans.GET("/:plan_id/tasks/:task_id", h.PlanTask.Get)
			plans.POST("/:plan_id/tasks/:task_id/compile", h.PlanTask.Compile)
			plans.POST("/:plan_id/tasks/:task_id/issue", h.PlanTask.CreateIssue)
		}

		issues := api.Group("/issues")
		{
			issues.GET("/:issue_id", h.Issue.Get)
			issues.GET("/:issue_id/spec", h.IssueSpec.GetCurrent)
			issues.GET("/:issue_id/specs", h.IssueSpec.List)
			issues.POST("/:issue_id/spec/rebuild", h.IssueSpec.Rebuild)
			issues.PATCH("/:issue_id/spec", h.IssueSpec.UpdateCurrent)
			issues.POST("/:issue_id/dispatch", h.Automation.Dispatch)
			issues.POST("/:issue_id/auto-repair", h.Automation.AutoRepair)
			issues.GET("/:issue_id/automation-attempts", h.Automation.ListAttempts)
			issues.GET("/:issue_id/automation-attempts/latest", h.Automation.GetLatestAttempt)
			issues.POST("/:issue_id/escalate", h.Automation.Escalate)
		}
	}
}
