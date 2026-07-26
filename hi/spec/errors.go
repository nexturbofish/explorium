package spec

import "fmt"

type ErrorType string

const (
	ErrConfig       ErrorType = "config"
	ErrStore        ErrorType = "store"
	ErrInvalidInput ErrorType = "invalid_input"
	ErrToolExec     ErrorType = "tool_exec"
)

type HiError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *HiError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *HiError) Unwrap() error { return e.Err }
