package ssh

import (
	"context"
	"os/exec"
)

// LocalExecutor executes commands locally for testing
type LocalExecutor struct{}

// Execute runs a command locally using shell
func (l *LocalExecutor) Execute(ctx context.Context, command string) (*ExecuteResult, error) {
	// Use shell to execute the command properly
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	
	stdout, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return &ExecuteResult{
				Command:  command,
				ExitCode: exitError.ExitCode(),
				Stdout:   string(stdout),
				Stderr:   string(exitError.Stderr),
			}, nil
		}
		return nil, err
	}
	
	return &ExecuteResult{
		Command:  command,
		ExitCode: 0,
		Stdout:   string(stdout),
		Stderr:   "",
	}, nil
}

// Connect is a no-op for local execution
func (l *LocalExecutor) Connect(ctx context.Context) error {
	return nil
}

// Close is a no-op for local execution
func (l *LocalExecutor) Close() error {
	return nil
}
