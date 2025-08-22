package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/ataiva-software/forge/pkg/ssh"
	"github.com/ataiva-software/forge/pkg/types"
)

// ServiceProvider manages service resources
type ServiceProvider struct {
	connection ssh.Executor
}

// NewServiceProvider creates a new service provider
func NewServiceProvider(connection ssh.Executor) *ServiceProvider {
	return &ServiceProvider{
		connection: connection,
	}
}

// Type returns the resource type this provider handles
func (p *ServiceProvider) Type() string {
	return "service"
}

// Validate validates the service resource configuration
func (p *ServiceProvider) Validate(resource *types.Resource) error {
	// Check state - can be in State field or Properties map
	var state string
	if resource.State != "" {
		state = string(resource.State)
	} else if stateInterface, ok := resource.Properties["state"]; ok {
		if stateStr, ok := stateInterface.(string); ok {
			state = stateStr
		} else {
			return fmt.Errorf("service 'state' must be a string")
		}
	} else {
		return fmt.Errorf("service resource must have 'state' property")
	}
	
	// Validate state values
	validStates := map[string]bool{
		"running": true,
		"stopped": true,
	}
	
	if !validStates[state] {
		return fmt.Errorf("invalid service state '%s', must be one of: running, stopped", state)
	}
	
	// Validate enabled if provided
	if enabled, ok := resource.Properties["enabled"]; ok {
		if _, ok := enabled.(bool); !ok {
			return fmt.Errorf("service 'enabled' must be a boolean")
		}
	}
	
	return nil
}

// Read reads the current state of the service
func (p *ServiceProvider) Read(ctx context.Context, resource *types.Resource) (map[string]interface{}, error) {
	serviceName := resource.Name
	
	// Check if service is active
	isActive, err := p.isServiceActive(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to check service status: %w", err)
	}
	
	// Check if service is enabled
	isEnabled, err := p.isServiceEnabled(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to check service enabled status: %w", err)
	}
	
	state := map[string]interface{}{
		"enabled": isEnabled,
	}
	
	if isActive {
		state["state"] = "running"
	} else {
		state["state"] = "stopped"
	}
	
	return state, nil
}

// Diff compares desired vs current state and returns the differences
func (p *ServiceProvider) Diff(ctx context.Context, resource *types.Resource, current map[string]interface{}) (*types.ResourceDiff, error) {
	diff := &types.ResourceDiff{
		ResourceID: resource.ResourceID(),
		Changes:    make(map[string]interface{}),
	}
	
	// Get desired state - check State field first, then Properties
	var desiredState string
	if resource.State != "" {
		desiredState = string(resource.State)
	} else if stateInterface, ok := resource.Properties["state"]; ok {
		desiredState = stateInterface.(string)
	}
	
	currentState := current["state"].(string)
	currentEnabled := current["enabled"].(bool)
	
	hasChanges := false
	
	// Check state changes
	if desiredState != currentState {
		hasChanges = true
		diff.Changes["state"] = map[string]interface{}{
			"from": currentState,
			"to":   desiredState,
		}
	}
	
	// Check enabled changes
	if enabledInterface, ok := resource.Properties["enabled"]; ok {
		desiredEnabled := enabledInterface.(bool)
		if desiredEnabled != currentEnabled {
			hasChanges = true
			diff.Changes["enabled"] = map[string]interface{}{
				"from": currentEnabled,
				"to":   desiredEnabled,
			}
		}
	}
	
	if hasChanges {
		diff.Action = types.ActionUpdate
		diff.Reason = "service needs to be updated"
	} else {
		diff.Action = types.ActionNoop
		diff.Reason = "service already in desired state"
	}
	
	return diff, nil
}

// Apply applies the changes to bring the service to desired state
func (p *ServiceProvider) Apply(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	switch diff.Action {
	case types.ActionUpdate:
		return p.updateService(ctx, resource, diff)
	case types.ActionNoop:
		return nil
	default:
		return fmt.Errorf("unsupported action: %s", diff.Action)
	}
}

// isServiceActive checks if a service is currently active/running
func (p *ServiceProvider) isServiceActive(ctx context.Context, serviceName string) (bool, error) {
	// Try systemctl first (systemd)
	cmd := fmt.Sprintf("systemctl is-active %s", shellEscape(serviceName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	
	// systemctl is-active returns 0 if active, 3 if inactive
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == "active" {
		return true, nil
	}
	
	// Try service command (SysV init)
	cmd = fmt.Sprintf("service %s status >/dev/null 2>&1 && echo active || echo inactive", shellEscape(serviceName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == "active" {
		return true, nil
	}
	
	return false, nil
}

// isServiceEnabled checks if a service is enabled to start at boot
func (p *ServiceProvider) isServiceEnabled(ctx context.Context, serviceName string) (bool, error) {
	// Try systemctl first (systemd)
	cmd := fmt.Sprintf("systemctl is-enabled %s", shellEscape(serviceName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	
	// systemctl is-enabled returns 0 if enabled
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == "enabled" {
		return true, nil
	}
	
	// For SysV init, check if service exists in runlevel directories
	cmd = fmt.Sprintf("ls /etc/rc*.d/S*%s 2>/dev/null | wc -l", shellEscape(serviceName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	
	if result.ExitCode == 0 {
		count := strings.TrimSpace(result.Stdout)
		return count != "0", nil
	}
	
	return false, nil
}

// updateService updates the service state and enabled status
func (p *ServiceProvider) updateService(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	serviceName := resource.Name
	
	// Handle state changes
	if stateChange, ok := diff.Changes["state"]; ok {
		change := stateChange.(map[string]interface{})
		desiredState := change["to"].(string)
		
		switch desiredState {
		case "running":
			if err := p.startService(ctx, serviceName); err != nil {
				return err
			}
		case "stopped":
			if err := p.stopService(ctx, serviceName); err != nil {
				return err
			}
		}
	}
	
	// Handle enabled changes
	if enabledChange, ok := diff.Changes["enabled"]; ok {
		change := enabledChange.(map[string]interface{})
		desiredEnabled := change["to"].(bool)
		
		if desiredEnabled {
			if err := p.enableService(ctx, serviceName); err != nil {
				return err
			}
		} else {
			if err := p.disableService(ctx, serviceName); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// startService starts a service
func (p *ServiceProvider) startService(ctx context.Context, serviceName string) error {
	// Try systemctl first (systemd)
	cmd := fmt.Sprintf("systemctl start %s", shellEscape(serviceName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try service command (SysV init)
	cmd = fmt.Sprintf("service %s start", shellEscape(serviceName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to start service %s: %s", serviceName, result.Stderr)
	}
	
	return nil
}

// stopService stops a service
func (p *ServiceProvider) stopService(ctx context.Context, serviceName string) error {
	// Try systemctl first (systemd)
	cmd := fmt.Sprintf("systemctl stop %s", shellEscape(serviceName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try service command (SysV init)
	cmd = fmt.Sprintf("service %s stop", shellEscape(serviceName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to stop service %s: %s", serviceName, result.Stderr)
	}
	
	return nil
}

// enableService enables a service to start at boot
func (p *ServiceProvider) enableService(ctx context.Context, serviceName string) error {
	// Try systemctl first (systemd)
	cmd := fmt.Sprintf("systemctl enable %s", shellEscape(serviceName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to enable service %s: %w", serviceName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try chkconfig (SysV init)
	cmd = fmt.Sprintf("chkconfig %s on", shellEscape(serviceName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to enable service %s: %w", serviceName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to enable service %s: %s", serviceName, result.Stderr)
	}
	
	return nil
}

// disableService disables a service from starting at boot
func (p *ServiceProvider) disableService(ctx context.Context, serviceName string) error {
	// Try systemctl first (systemd)
	cmd := fmt.Sprintf("systemctl disable %s", shellEscape(serviceName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to disable service %s: %w", serviceName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try chkconfig (SysV init)
	cmd = fmt.Sprintf("chkconfig %s off", shellEscape(serviceName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to disable service %s: %w", serviceName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to disable service %s: %s", serviceName, result.Stderr)
	}
	
	return nil
}
