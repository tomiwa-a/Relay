package executor

import (
	"context"
	"encoding/json"
)

// Executor defines the interface for executing jobs
type Executor interface {
	Execute(ctx context.Context, payload json.RawMessage) (*ExecutionResult, error)
}

// NewExecutor returns the appropriate executor based on payload type
func NewExecutor() Executor {
	return &ShellExecutor{}
}
