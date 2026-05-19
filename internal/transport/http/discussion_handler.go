package http

import (
	"context"
	"net/http"

	discussionapp "anserflow/internal/app/discussion"

	"github.com/gin-gonic/gin"
)

type DiscussionHandler struct {
	app   discussionapp.Service
	scope ScopeChecker
}

func NewDiscussionHandler(app discussionapp.Service, scope ScopeChecker) *DiscussionHandler {
	return &DiscussionHandler{app: app, scope: scope}
}

func (h *DiscussionHandler) Get(c *gin.Context) {
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

	out, err := h.app.Get(c.Request.Context(), conversationID, sessionID)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

type refreshDiscussionReq struct {
	Force bool `json:"force"`
}

func (h *DiscussionHandler) Refresh(c *gin.Context) {
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

	var req refreshDiscussionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	out, err := h.app.Refresh(c.Request.Context(), discussionapp.RefreshInput{
		ConversationID: conversationID,
		SessionID:      sessionID,
		Force:          req.Force,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *DiscussionHandler) Freeze(c *gin.Context) {
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

	out, err := h.app.Freeze(c.Request.Context(), discussionapp.FreezeInput{
		ConversationID: conversationID,
		SessionID:      sessionID,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	respondOK(c, out)
}

func (h *DiscussionHandler) ensureConversationScope(ctx context.Context, orgID uint64, conversationID uint64) error {
	ok, err := h.scope.ConversationInOrg(ctx, orgID, conversationID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
