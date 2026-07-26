package guard

import (
	"fmt"
	"sync"
	"time"
)

type ConfirmationRequest struct {
	ID        string
	Tool      string
	Args      map[string]any
	Reason    string
	CreatedAt time.Time
	Timeout   time.Duration
	done      chan struct{}
	approved  bool
}

type ConfirmationQueue struct {
	pending map[string]*ConfirmationRequest
	timeout time.Duration
	mu      sync.Mutex
}

func NewConfirmationQueue(timeout time.Duration) *ConfirmationQueue {
	return &ConfirmationQueue{
		pending: make(map[string]*ConfirmationRequest),
		timeout: timeout,
	}
}

func (q *ConfirmationQueue) Enqueue(tool string, args map[string]any, reason string) *ConfirmationRequest {
	req := &ConfirmationRequest{
		ID:        fmt.Sprintf("cfm-%d", time.Now().UnixNano()),
		Tool:      tool,
		Args:      args,
		Reason:    reason,
		CreatedAt: time.Now(),
		Timeout:   q.timeout,
		done:      make(chan struct{}),
	}
	q.mu.Lock()
	q.pending[req.ID] = req
	q.mu.Unlock()
	return req
}

func (q *ConfirmationQueue) Approve(id string) bool {
	q.mu.Lock()
	req, ok := q.pending[id]
	if ok {
		delete(q.pending, id)
	}
	q.mu.Unlock()
	if !ok {
		return false
	}
	req.approved = true
	close(req.done)
	return true
}

func (q *ConfirmationQueue) Deny(id string) bool {
	q.mu.Lock()
	req, ok := q.pending[id]
	if ok {
		delete(q.pending, id)
	}
	q.mu.Unlock()
	if !ok {
		return false
	}
	req.approved = false
	close(req.done)
	return true
}

func (q *ConfirmationQueue) Wait(req *ConfirmationRequest) (bool, error) {
	select {
	case <-req.done:
		return req.approved, nil
	case <-time.After(req.Timeout):
		q.mu.Lock()
		delete(q.pending, req.ID)
		q.mu.Unlock()
		return false, fmt.Errorf("confirmation %q timed out", req.ID)
	}
}
