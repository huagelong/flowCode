package http

import (
	"context"
	"net/http"

	executionapp "anserflow/internal/app/execution"

	"github.com/gin-gonic/gin"
)

type IssueSpecHandler struct {
	app   executionapp.SpecService
	scope ScopeChecker
}

func NewIssueSpecHandler(app executionapp.SpecService, scope ScopeChecker) *IssueSpecHandler {
	return &IssueSpecHandler{app: app, scope: scope}
}

func (h *IssueSpecHandler) GetCurrent(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	issueID, ok := mustUint64Param(c, "issue_id")
	if !ok {
		return
	}
	if err := h.ensureIssueScope(c.Request.Context(), orgID, issueID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}
	out, err := h.app.GetCurrent(c.Request.Context(), issueID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *IssueSpecHandler) List(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	issueID, ok := mustUint64Param(c, "issue_id")
	if !ok {
		return
	}
	if err := h.ensureIssueScope(c.Request.Context(), orgID, issueID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}
	out, err := h.app.List(c.Request.Context(), issueID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

type rebuildSpecReq struct {
	TriggeredByID uint64 `json:"triggered_by_id"`
	Reason        string `json:"reason"`
}

func (h *IssueSpecHandler) Rebuild(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	issueID, ok := mustUint64Param(c, "issue_id")
	if !ok {
		return
	}
	if err := h.ensureIssueScope(c.Request.Context(), orgID, issueID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}

	var req rebuildSpecReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	out, err := h.app.Rebuild(c.Request.Context(), executionapp.RebuildSpecInput{
		IssueID:     issueID,
		TriggeredBy: req.TriggeredByID,
		Reason:      req.Reason,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondCreated(c, out)
}

type updateSpecReq struct {
	TriggeredByID       uint64                              `json:"triggered_by_id"`
	Goal                string                              `json:"goal"`
	Scope               []string                            `json:"scope"`
	OutOfScope          []string                            `json:"out_of_scope"`
	TargetPaths         []string                            `json:"target_paths"`
	RelatedModules      []string                            `json:"related_modules"`
	Dependencies        []string                            `json:"dependencies"`
	Constraints         []string                            `json:"constraints"`
	ImplementationNotes []string                            `json:"implementation_notes"`
	AcceptanceChecks    []executionapp.AcceptanceCheckInput `json:"acceptance_checks"`
	RetryPolicy         executionapp.RetryPolicyInput       `json:"retry_policy"`
	MergePolicy         executionapp.MergePolicyInput       `json:"merge_policy"`
	RollbackPolicy      executionapp.RollbackPolicyInput    `json:"rollback_policy"`
	ExecutionMode       string                              `json:"execution_mode"`
	RiskLevel           string                              `json:"risk_level"`
	Reason              string                              `json:"reason"`
}

func (h *IssueSpecHandler) UpdateCurrent(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	issueID, ok := mustUint64Param(c, "issue_id")
	if !ok {
		return
	}
	if err := h.ensureIssueScope(c.Request.Context(), orgID, issueID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}

	var req updateSpecReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	out, err := h.app.UpdateCurrent(c.Request.Context(), executionapp.UpdateSpecInput{
		IssueID:             issueID,
		TriggeredBy:         req.TriggeredByID,
		Goal:                req.Goal,
		Scope:               req.Scope,
		OutOfScope:          req.OutOfScope,
		TargetPaths:         req.TargetPaths,
		RelatedModules:      req.RelatedModules,
		Dependencies:        req.Dependencies,
		Constraints:         req.Constraints,
		ImplementationNotes: req.ImplementationNotes,
		AcceptanceChecks:    executionapp.ToAcceptanceCheckData(req.AcceptanceChecks),
		RetryPolicy:         executionapp.ToRetryPolicyData(req.RetryPolicy),
		MergePolicy:         executionapp.ToMergePolicyData(req.MergePolicy),
		RollbackPolicy:      executionapp.ToRollbackPolicyData(req.RollbackPolicy),
		ExecutionMode:       req.ExecutionMode,
		RiskLevel:           req.RiskLevel,
		Reason:              req.Reason,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondCreated(c, out)
}

func (h *IssueSpecHandler) ensureIssueScope(ctx context.Context, orgID uint64, issueID uint64) error {
	ok, err := h.scope.IssueInOrg(ctx, orgID, issueID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
