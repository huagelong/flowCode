package http

import (
	"context"
	"net/http"

	planningapp "anserflow/internal/app/planning"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	app   planningapp.Service
	scope ScopeChecker
}

func NewPlanHandler(app planningapp.Service, scope ScopeChecker) *PlanHandler {
	return &PlanHandler{app: app, scope: scope}
}

type createPlanReq struct {
	ForceRefreshDiscussionState bool `json:"force_refresh_discussion_state"`
}

func (h *PlanHandler) Create(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	conversationID, ok := mustUint64Param(c, "conversation_id")
	if !ok {
		return
	}
	sessionID := c.Param("session_id")
	if err := h.ensureConversationScope(c.Request.Context(), orgID, conversationID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}

	var req createPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	out, err := h.app.CreatePlan(c.Request.Context(), planningapp.CreatePlanInput{
		ConversationID:              conversationID,
		SessionID:                   sessionID,
		ForceRefreshDiscussionState: req.ForceRefreshDiscussionState,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondCreated(c, out)
}

func (h *PlanHandler) Get(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	planID, ok := mustUint64Param(c, "plan_id")
	if !ok {
		return
	}
	if err := h.ensurePlanScope(c.Request.Context(), orgID, planID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}
	out, err := h.app.GetPlan(c.Request.Context(), planID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *PlanHandler) ListBySession(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	sessionID := c.Query("session_id")
	if sessionID == "" {
		respondError(c, http.StatusBadRequest, errRequired("session_id"))
		return
	}
	if err := h.ensureSessionScope(c.Request.Context(), orgID, sessionID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}
	out, err := h.app.ListPlans(c.Request.Context(), sessionID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

type approvePlanReq struct {
	UserID  uint64 `json:"user_id"`
	Comment string `json:"comment"`
}

func (h *PlanHandler) Approve(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	planID, ok := mustUint64Param(c, "plan_id")
	if !ok {
		return
	}
	if err := h.ensurePlanScope(c.Request.Context(), orgID, planID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}

	var req approvePlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	out, err := h.app.ApprovePlan(c.Request.Context(), planningapp.ApprovePlanInput{
		PlanID:     planID,
		ApprovedBy: req.UserID,
		Comment:    req.Comment,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

type rejectPlanReq struct {
	UserID uint64 `json:"user_id"`
	Reason string `json:"reason"`
}

func (h *PlanHandler) Reject(c *gin.Context) {
	orgID, ok := mustUint64Param(c, "org_id")
	if !ok {
		return
	}
	planID, ok := mustUint64Param(c, "plan_id")
	if !ok {
		return
	}
	if err := h.ensurePlanScope(c.Request.Context(), orgID, planID); err != nil {
		respondError(c, http.StatusForbidden, err)
		return
	}

	var req rejectPlanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	out, err := h.app.RejectPlan(c.Request.Context(), planningapp.RejectPlanInput{
		PlanID:     planID,
		RejectedBy: req.UserID,
		Reason:     req.Reason,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *PlanHandler) ensureConversationScope(ctx context.Context, orgID uint64, conversationID uint64) error {
	ok, err := h.scope.ConversationInOrg(ctx, orgID, conversationID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (h *PlanHandler) ensureSessionScope(ctx context.Context, orgID uint64, sessionID string) error {
	ok, err := h.scope.SessionInOrg(ctx, orgID, sessionID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (h *PlanHandler) ensurePlanScope(ctx context.Context, orgID uint64, planID uint64) error {
	ok, err := h.scope.PlanInOrg(ctx, orgID, planID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
