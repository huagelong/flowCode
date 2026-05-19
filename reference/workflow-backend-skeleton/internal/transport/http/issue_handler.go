package http

import (
	"context"
	"net/http"

	executionapp "anserflow/internal/app/execution"

	"github.com/gin-gonic/gin"
)

type IssueHandler struct {
	app   executionapp.IssueService
	scope ScopeChecker
}

func NewIssueHandler(app executionapp.IssueService, scope ScopeChecker) *IssueHandler {
	return &IssueHandler{app: app, scope: scope}
}

func (h *IssueHandler) Get(c *gin.Context) {
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
	out, err := h.app.Get(c.Request.Context(), issueID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *IssueHandler) ensureIssueScope(ctx context.Context, orgID uint64, issueID uint64) error {
	ok, err := h.scope.IssueInOrg(ctx, orgID, issueID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
