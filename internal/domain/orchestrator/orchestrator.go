package orchestrator

import (
	"context"

	"anserflow/internal/app/discussion"
	"anserflow/internal/convert"
	"anserflow/internal/model"
)

type Stub struct{}

func NewStub() *Stub {
	return &Stub{}
}

func (s *Stub) Refresh(ctx context.Context, messages []*model.Message, previous *model.DiscussionState, force bool) (*discussion.OrchestratorResult, error) {
	return &discussion.OrchestratorResult{
		Topic:             "stub topic",
		Goal:              "stub goal",
		LatestSummary:     "stub summary",
		ConfirmedFacts:    []string{},
		Assumptions:       []string{},
		OpenQuestions:     []string{},
		CandidateOptions:  []convert.CandidateOptionData{},
		Risks:             []convert.DiscussionRiskData{},
		Constraints:       []string{},
		Participants:      []convert.ParticipantData{},
		MissingFields:     []string{},
		ReadinessStage:    string(model.DiscussionStagePlannable),
		Confidence:        0.8,
		BasedOnMessageSeq: maxSeq(messages),
	}, nil
}

func maxSeq(messages []*model.Message) uint64 {
	var max uint64
	for _, m := range messages {
		if m != nil && m.Seq > max {
			max = m.Seq
		}
	}
	return max
}
