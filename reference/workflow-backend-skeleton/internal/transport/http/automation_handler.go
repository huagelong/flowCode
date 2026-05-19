package http

import (
	"context"
	"net/http"

	executionapp "anserflow/internal/app/execution"

	"github.com/gin-gonic/gin"
)

type AutomationHandler struct {
	app   executionapp.Service
	scope ScopeChecker
}

func NewAutomationHandler(app executionapp.Service, scope ScopeChecker) *AutomationHandler {
	return &AutomationHandler{app: app, scope: scope}
}

func (h *AutomationHandler) Dispatch(c *gin.Context) {
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
	out, err := h.app.Dispatch(c.Request.Context(), executionapp.DispatchInput{IssueID: issueID})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondCreated(c, out)
}

type autoRepairReq struct {
	Reason string `json:"reason"`
}

func (h *AutomationHandler) AutoRepair(c *gin.Context) {
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

	var req autoRepairReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	out, err := h.app.AutoRepair(c.Request.Context(), executionapp.AutoRepairInput{
		IssueID: issueID,
		Reason:  req.Reason,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondCreated(c, out)
}

func (h *AutomationHandler) ListAttempts(c *gin.Context) {
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
	out, err := h.app.ListAttempts(c.Request.Context(), issueID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *AutomationHandler) GetLatestAttempt(c *gin.Context) {
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
	out, err := h.app.GetLatestAttempt(c.Request.Context(), issueID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

type escalateReq struct {
	UserID uint64 `json:"user_id"`
	Reason string `json:"reason"`
}

func (h *AutomationHandler) Escalate(c *gin.Context) {
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

	var req escalateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.app.Escalate(c.Request.Context(), executionapp.EscalateInput{
		IssueID:     issueID,
		EscalatedBy: req.UserID,
		Reason:      req.Reason,
	}); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, gin.H{"ok": true})
}

func (h *AutomationHandler) ensureIssueScope(ctx context.Context, orgID uint64, issueID uint64) error {
	ok, err := h.scope.IssueInOrg(ctx, orgID, issueID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
