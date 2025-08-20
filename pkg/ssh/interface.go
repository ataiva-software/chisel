package ssh

import "context"

// Executor defines the interface for executing commands remotely
type Executor interface {
	Execute(ctx context.Context, command string) (*ExecuteResult, error)
	Connect(ctx context.Context) error
	Close() error
}

// Ensure Connection implements Executor
var _ Executor = (*Connection)(nil)
