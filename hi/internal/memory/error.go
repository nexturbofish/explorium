package memory

import (
	"fmt"
	"hi/spec"
)

type MemoryStoreError struct {
	Op    string // save, load, delete, list, search, pin, unpin
	ID    string
	Scope spec.MemoryScope
	Err   error
}

func (e *MemoryStoreError) Error() string {
	return fmt.Sprintf("memory %s: %s: %v", e.Op, e.ID, e.Err)
}

func (e *MemoryStoreError) Unwrap() error { return e.Err }
