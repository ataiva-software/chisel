package ssh

import (
	"context"
	"fmt"
)

// MockExecutor is a mock implementation of the Executor interface for testing
type MockExecutor struct {
	connected bool
}

// NewMockExecutor creates a new mock executor
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

// Execute executes a command (mock implementation)
func (m *MockExecutor) Execute(ctx context.Context, command string) (*ExecuteResult, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	
	// Mock successful execution
	return &ExecuteResult{
		ExitCode: 0,
		Stdout:   "mock output",
		Stderr:   "",
	}, nil
}

// Connect establishes a connection (mock implementation)
func (m *MockExecutor) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}

// Close closes the connection (mock implementation)
func (m *MockExecutor) Close() error {
	m.connected = false
	return nil
}
