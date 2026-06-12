package video

import (
	"context"
	"fmt"
	"sync"
)

type Mock struct {
	mu    sync.Mutex
	jobs  map[string]int
	calls int
}

func NewMock() *Mock {
	return &Mock{jobs: map[string]int{}}
}

func (m *Mock) ID() string { return "mock" }

func (m *Mock) Submit(ctx context.Context, req SceneRequest) (*Job, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	id := fmt.Sprintf("mock-job-%s-%d", req.SceneID, m.calls)
	m.jobs[id] = 0
	return &Job{ID: id}, nil
}

func (m *Mock) Poll(ctx context.Context, jobID string) (*JobResult, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("unknown job %s", jobID)
	}
	n++
	m.jobs[jobID] = n
	if n < 2 {
		return &JobResult{Status: StatusRunning}, nil
	}
	return &JobResult{
		Status:   StatusCompleted,
		VideoURL: "https://example.com/mock-video.mp4",
	}, nil
}

func (m *Mock) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}
