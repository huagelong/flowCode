package http

import (
	"context"
	"errors"

	"anserflow/internal/store"
)

var ErrForbidden = errors.New("resource does not belong to org")

type ScopeChecker interface {
	ConversationInOrg(ctx context.Context, orgID uint64, conversationID uint64) (bool, error)
	SessionInOrg(ctx context.Context, orgID uint64, sessionID string) (bool, error)
	PlanInOrg(ctx context.Context, orgID uint64, planID uint64) (bool, error)
	IssueInOrg(ctx context.Context, orgID uint64, issueID uint64) (bool, error)
}

type scopeChecker struct {
	conversations    store.ConversationStore
	discussionStates store.DiscussionStateStore
	plans            store.PlanStore
	issues           store.IssueStore
}

func NewScopeChecker(
	conversations store.ConversationStore,
	discussionStates store.DiscussionStateStore,
	plans store.PlanStore,
	issues store.IssueStore,
) ScopeChecker {
	return &scopeChecker{
		conversations:    conversations,
		discussionStates: discussionStates,
		plans:            plans,
		issues:           issues,
	}
}

func (s *scopeChecker) ConversationInOrg(ctx context.Context, orgID uint64, conversationID uint64) (bool, error) {
	item, err := s.conversations.FindByID(ctx, conversationID)
	if err != nil {
		return false, err
	}
	return item != nil && item.OrgID == orgID, nil
}

func (s *scopeChecker) SessionInOrg(ctx context.Context, orgID uint64, sessionID string) (bool, error) {
	item, err := s.discussionStates.FindBySession(ctx, sessionID)
	if err != nil {
		return false, err
	}
	return item != nil && item.OrgID == orgID, nil
}

func (s *scopeChecker) PlanInOrg(ctx context.Context, orgID uint64, planID uint64) (bool, error) {
	item, err := s.plans.FindByID(ctx, planID)
	if err != nil {
		return false, err
	}
	return item != nil && item.OrgID == orgID, nil
}

func (s *scopeChecker) IssueInOrg(ctx context.Context, orgID uint64, issueID uint64) (bool, error) {
	item, err := s.issues.FindByID(ctx, issueID)
	if err != nil {
		return false, err
	}
	return item != nil && item.OrgID == orgID, nil
}
