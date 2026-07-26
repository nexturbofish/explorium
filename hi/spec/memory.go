package spec

import "context"

type MemoryStore interface {
	Save(memory *LoadedMemory) error
	Load(id string) (*LoadedMemory, error)
	Delete(id string) error
	List(scope MemoryScope) ([]*LoadedMemory, error)
	Search(ctx context.Context, query string, limit int) ([]ScoredMemory, error)
	Touch(memory *LoadedMemory) error
	ListPinned(ctx context.Context) ([]*LoadedMemory, error)
	ListActive(ctx context.Context) ([]*LoadedMemory, error)
	Pin(id string) error
	Unpin(id string) error
}

type ScoredMemory struct {
	Memory    *LoadedMemory
	Score     float64
	MatchedOn string
}
