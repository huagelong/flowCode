package discussion

import (
	"context"
	"errors"

	"anserflow/internal/convert"
	"anserflow/internal/model"
	"anserflow/internal/store"
)

var ErrNotFound = errors.New("resource not found")

type Orchestrator interface {
	Refresh(ctx context.Context, messages []*model.Message, previous *model.DiscussionState, force bool) (*OrchestratorResult, error)
}

type OrchestratorResult struct {
	Topic             string
	Goal              string
	LatestSummary     string
	ConfirmedFacts    []string
	Assumptions       []string
	OpenQuestions     []string
	CandidateOptions  []convert.CandidateOptionData
	Risks             []convert.DiscussionRiskData
	Constraints       []string
	Participants      []convert.ParticipantData
	MissingFields     []string
	ReadinessStage    string
	Confidence        float64
	BasedOnMessageSeq uint64
}

type Service interface {
	Get(ctx context.Context, conversationID uint64, sessionID string) (*StateView, error)
	Refresh(ctx context.Context, in RefreshInput) (*StateView, error)
	Freeze(ctx context.Context, in FreezeInput) (*StateView, error)
}

type service struct {
	conversations store.ConversationStore
	messages      store.MessageStore
	states        store.DiscussionStateStore
	tx            store.TxManager
	orchestrator  Orchestrator
}

func New(
	conversations store.ConversationStore,
	messages store.MessageStore,
	states store.DiscussionStateStore,
	tx store.TxManager,
	orchestrator Orchestrator,
) Service {
	return &service{
		conversations: conversations,
		messages:      messages,
		states:        states,
		tx:            tx,
		orchestrator:  orchestrator,
	}
}

func (s *service) Get(ctx context.Context, conversationID uint64, sessionID string) (*StateView, error) {
	state, err := s.states.FindBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if state == nil || state.ConversationID != conversationID {
		return nil, ErrNotFound
	}
	return toStateView(state), nil
}

func (s *service) Refresh(ctx context.Context, in RefreshInput) (*StateView, error) {
	var out *StateView

	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		conv, err := s.conversations.FindByIDForUpdate(ctx, tx, in.ConversationID)
		if err != nil {
			return err
		}
		if conv == nil {
			return ErrNotFound
		}

		current, err := s.states.FindBySessionForUpdate(ctx, tx, in.SessionID)
		if err != nil {
			return err
		}

		lastSeq, err := s.messages.FindLastSeq(ctx, in.ConversationID, in.SessionID)
		if err != nil {
			return err
		}
		if current != nil && lastSeq <= current.BasedOnMessageSeq && !in.Force {
			out = toStateView(current)
			return nil
		}

		var afterSeq uint64
		if current != nil {
			afterSeq = current.BasedOnMessageSeq
		}
		msgs, err := s.messages.ListAfterSeq(ctx, in.ConversationID, in.SessionID, afterSeq, 500)
		if err != nil {
			return err
		}

		refreshed, err := s.orchestrator.Refresh(ctx, msgs, current, in.Force)
		if err != nil {
			return err
		}

		if current == nil {
			current = &model.DiscussionState{
				OrgID:          conv.OrgID,
				ConversationID: in.ConversationID,
				SessionID:      in.SessionID,
				Version:        1,
			}
		} else {
			current.Version++
		}

		if err := convert.FillDiscussionStateModel(
			current,
			refreshed.Topic,
			refreshed.Goal,
			refreshed.LatestSummary,
			refreshed.ConfirmedFacts,
			refreshed.Assumptions,
			refreshed.OpenQuestions,
			refreshed.CandidateOptions,
			refreshed.Risks,
			refreshed.Constraints,
			refreshed.Participants,
			refreshed.MissingFields,
			refreshed.ReadinessStage,
			refreshed.Confidence,
			refreshed.BasedOnMessageSeq,
		); err != nil {
			return err
		}

		if current.ID == 0 {
			if err := s.states.Create(ctx, tx, current); err != nil {
				return err
			}
		} else {
			if err := s.states.Update(ctx, tx, current); err != nil {
				return err
			}
		}

		out = toStateView(current)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *service) Freeze(ctx context.Context, in FreezeInput) (*StateView, error) {
	var out *StateView

	err := s.tx.WithinTransaction(ctx, func(tx store.Tx) error {
		current, err := s.states.FindBySessionForUpdate(ctx, tx, in.SessionID)
		if err != nil {
			return err
		}
		if current == nil || current.ConversationID != in.ConversationID {
			return ErrNotFound
		}
		if err := s.states.Freeze(ctx, tx, in.SessionID, current.Version); err != nil {
			return err
		}
		current.ReadinessStage = model.DiscussionStageFrozen
		out = toStateView(current)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
