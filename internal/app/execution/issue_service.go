package execution

import (
	"context"

	"anserflow/internal/store"
)

type IssueService interface {
	Get(ctx context.Context, issueID uint64) (*IssueDetailView, error)
}

type issueService struct {
	issueRead store.IssueReadStore
}

func NewIssueService(issueRead store.IssueReadStore) IssueService {
	return &issueService{issueRead: issueRead}
}

func (s *issueService) Get(ctx context.Context, issueID uint64) (*IssueDetailView, error) {
	issue, spec, err := s.issueRead.FindIssueWithCurrentSpec(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, ErrExecutionNotFound
	}
	return toIssueDetailView(issue, spec), nil
}
