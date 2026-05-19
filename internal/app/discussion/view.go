package discussion

import (
	"anserflow/internal/convert"
	"anserflow/internal/model"
)

func toStateView(state *model.DiscussionState) *StateView {
	if state == nil {
		return nil
	}
	confirmedFacts, _ := convert.DecodeJSON[[]string](state.ConfirmedFactsJSON)
	assumptions, _ := convert.DecodeJSON[[]string](state.AssumptionsJSON)
	openQuestions, _ := convert.DecodeJSON[[]string](state.OpenQuestionsJSON)
	candidateOptions, _ := convert.DecodeJSON[[]convert.CandidateOptionData](state.CandidateOptionsJSON)
	risks, _ := convert.DecodeJSON[[]convert.DiscussionRiskData](state.RisksJSON)
	constraints, _ := convert.DecodeJSON[[]string](state.ConstraintsJSON)
	participants, _ := convert.DecodeJSON[[]convert.ParticipantData](state.ParticipantsJSON)
	missingFields, _ := convert.DecodeJSON[[]string](state.MissingFieldsJSON)

	return &StateView{
		ID:                state.ID,
		OrgID:             state.OrgID,
		ConversationID:    state.ConversationID,
		SessionID:         state.SessionID,
		Topic:             state.Topic,
		Goal:              state.Goal,
		LatestSummary:     state.LatestSummary,
		ReadinessStage:    string(state.ReadinessStage),
		Confidence:        state.Confidence,
		BasedOnMessageSeq: state.BasedOnMessageSeq,
		Version:           state.Version,
		ConfirmedFacts:    confirmedFacts,
		Assumptions:       assumptions,
		OpenQuestions:     openQuestions,
		CandidateOptions:  candidateOptions,
		Risks:             risks,
		Constraints:       constraints,
		Participants:      participants,
		MissingFields:     missingFields,
		UpdatedAt:         state.UpdatedAt,
	}
}
