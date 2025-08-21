package providers

import (
	"context"
	"fmt"

	"github.com/ataiva-software/chisel/pkg/ssh"
	"github.com/ataiva-software/chisel/pkg/types"
)

// ShellProvider manages shell command execution resources
type ShellProvider struct {
	connection ssh.Executor
}

// NewShellProvider creates a new shell provider
func NewShellProvider(connection ssh.Executor) *ShellProvider {
	return &ShellProvider{
		connection: connection,
	}
}

// Type returns the resource type this provider handles
func (p *ShellProvider) Type() string {
	return "shell"
}

// Validate validates the shell resource configuration
func (p *ShellProvider) Validate(resource *types.Resource) error {
	// Check required command property
	command, ok := resource.Properties["command"]
	if !ok {
		return fmt.Errorf("shell resource must have 'command' property")
	}
	
	if _, ok := command.(string); !ok {
		return fmt.Errorf("shell 'command' must be a string")
	}
	
	// Validate optional properties
	if creates, ok := resource.Properties["creates"]; ok {
		if _, ok := creates.(string); !ok {
			return fmt.Errorf("shell 'creates' must be a string")
		}
	}
	
	if unless, ok := resource.Properties["unless"]; ok {
		if _, ok := unless.(string); !ok {
			return fmt.Errorf("shell 'unless' must be a string")
		}
	}
	
	if onlyIf, ok := resource.Properties["only_if"]; ok {
		if _, ok := onlyIf.(string); !ok {
			return fmt.Errorf("shell 'only_if' must be a string")
		}
	}
	
	if cwd, ok := resource.Properties["cwd"]; ok {
		if _, ok := cwd.(string); !ok {
			return fmt.Errorf("shell 'cwd' must be a string")
		}
	}
	
	if user, ok := resource.Properties["user"]; ok {
		if _, ok := user.(string); !ok {
			return fmt.Errorf("shell 'user' must be a string")
		}
	}
	
	if timeout, ok := resource.Properties["timeout"]; ok {
		if _, ok := timeout.(int); !ok {
			return fmt.Errorf("shell 'timeout' must be an integer")
		}
	}
	
	return nil
}

// Read reads the current state to determine if the command should run
func (p *ShellProvider) Read(ctx context.Context, resource *types.Resource) (map[string]interface{}, error) {
	state := map[string]interface{}{}
	
	shouldRun := true
	
	// Check 'creates' condition - if file exists, don't run
	if creates, ok := resource.Properties["creates"]; ok {
		createsPath := creates.(string)
		exists, err := p.fileExists(ctx, createsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check creates condition: %w", err)
		}
		if exists {
			shouldRun = false
		}
	}
	
	// Check 'unless' condition - if command succeeds, don't run
	if shouldRun {
		if unless, ok := resource.Properties["unless"]; ok {
			unlessCmd := unless.(string)
			success, err := p.commandSucceeds(ctx, unlessCmd)
			if err != nil {
				return nil, fmt.Errorf("failed to check unless condition: %w", err)
			}
			if success {
				shouldRun = false
			}
		}
	}
	
	// Check 'only_if' condition - if command fails, don't run
	if shouldRun {
		if onlyIf, ok := resource.Properties["only_if"]; ok {
			onlyIfCmd := onlyIf.(string)
			success, err := p.commandSucceeds(ctx, onlyIfCmd)
			if err != nil {
				return nil, fmt.Errorf("failed to check only_if condition: %w", err)
			}
			if !success {
				shouldRun = false
			}
		}
	}
	
	state["should_run"] = shouldRun
	
	return state, nil
}

// Diff compares desired vs current state and returns the differences
func (p *ShellProvider) Diff(ctx context.Context, resource *types.Resource, current map[string]interface{}) (*types.ResourceDiff, error) {
	diff := &types.ResourceDiff{
		ResourceID: resource.ResourceID(),
		Changes:    make(map[string]interface{}),
	}
	
	shouldRun := current["should_run"].(bool)
	
	if shouldRun {
		diff.Action = types.ActionUpdate
		diff.Reason = "command needs to be executed"
		diff.Changes["command"] = map[string]interface{}{
			"from": "not executed",
			"to":   "will execute",
		}
	} else {
		diff.Action = types.ActionNoop
		diff.Reason = "command should not run based on conditions"
	}
	
	return diff, nil
}

// Apply applies the changes by executing the shell command
func (p *ShellProvider) Apply(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	switch diff.Action {
	case types.ActionUpdate:
		return p.executeCommand(ctx, resource)
	case types.ActionNoop:
		return nil
	default:
		return fmt.Errorf("unsupported action: %s", diff.Action)
	}
}

// fileExists checks if a file or directory exists
func (p *ShellProvider) fileExists(ctx context.Context, path string) (bool, error) {
	cmd := fmt.Sprintf("test -e %s", shellEscape(path))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	
	return result.ExitCode == 0, nil
}

// commandSucceeds checks if a command executes successfully
func (p *ShellProvider) commandSucceeds(ctx context.Context, command string) (bool, error) {
	result, err := p.connection.Execute(ctx, command)
	if err != nil {
		return false, err
	}
	
	return result.ExitCode == 0, nil
}

// executeCommand executes the shell command with proper context
func (p *ShellProvider) executeCommand(ctx context.Context, resource *types.Resource) error {
	command := resource.Properties["command"].(string)
	
	// Build the full command with context
	fullCommand := p.buildCommand(resource, command)
	
	result, err := p.connection.Execute(ctx, fullCommand)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}
	
	return nil
}

// buildCommand builds the full command with user, cwd, and other context
func (p *ShellProvider) buildCommand(resource *types.Resource, command string) string {
	fullCommand := command
	
	// Add user context if specified
	if user, ok := resource.Properties["user"]; ok {
		userStr := user.(string)
		fullCommand = fmt.Sprintf("sudo -u %s %s", shellEscape(userStr), fullCommand)
	}
	
	// Add working directory if specified
	if cwd, ok := resource.Properties["cwd"]; ok {
		cwdStr := cwd.(string)
		fullCommand = fmt.Sprintf("cd %s && %s", shellEscape(cwdStr), fullCommand)
	}
	
	// Add timeout if specified
	if timeout, ok := resource.Properties["timeout"]; ok {
		timeoutInt := timeout.(int)
		fullCommand = fmt.Sprintf("timeout %d %s", timeoutInt, fullCommand)
	}
	
	return fullCommand
}
