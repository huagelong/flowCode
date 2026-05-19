package execution

type inMemoryQueueStub struct{}

func NewInMemoryQueueStub() Queue {
	return &inMemoryQueueStub{}
}

func (q *inMemoryQueueStub) EnqueueIssueExecution(issueID uint64, issueSpecID uint64, attemptNo uint32) error {
	return nil
}
