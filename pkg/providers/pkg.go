package providers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ataiva-software/chisel/pkg/ssh"
	"github.com/ataiva-software/chisel/pkg/types"
)

// PkgProvider manages package resources
type PkgProvider struct {
	connection ssh.Executor
}

// NewPkgProvider creates a new package provider
func NewPkgProvider(connection ssh.Executor) *PkgProvider {
	return &PkgProvider{
		connection: connection,
	}
}

// Type returns the resource type this provider handles
func (p *PkgProvider) Type() string {
	return "pkg"
}

// Validate validates the package resource configuration
func (p *PkgProvider) Validate(resource *types.Resource) error {
	// Check state - can be in State field or Properties map
	var state string
	if resource.State != "" {
		state = string(resource.State)
	} else if stateInterface, ok := resource.Properties["state"]; ok {
		if stateStr, ok := stateInterface.(string); ok {
			state = stateStr
		} else {
			return fmt.Errorf("package 'state' must be a string")
		}
	} else {
		return fmt.Errorf("package resource must have 'state' property")
	}
	
	// Validate state values
	validStates := map[string]bool{
		"present": true,
		"absent":  true,
		"latest":  true,
	}
	
	if !validStates[state] {
		return fmt.Errorf("invalid package state '%s', must be one of: present, absent, latest", state)
	}
	
	// Validate version if provided
	if version, ok := resource.Properties["version"]; ok {
		if _, ok := version.(string); !ok {
			return fmt.Errorf("package 'version' must be a string")
		}
	}
	
	return nil
}

// Read reads the current state of the package
func (p *PkgProvider) Read(ctx context.Context, resource *types.Resource) (map[string]interface{}, error) {
	packageName := resource.Name
	
	// Detect package manager and check if package is installed
	isInstalled, version, err := p.isPackageInstalled(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to check package status: %w", err)
	}
	
	state := map[string]interface{}{}
	
	if isInstalled {
		state["state"] = "present"
		if version != "" {
			state["version"] = version
		}
	} else {
		state["state"] = "absent"
	}
	
	return state, nil
}

// Diff compares desired vs current state and returns the differences
func (p *PkgProvider) Diff(ctx context.Context, resource *types.Resource, current map[string]interface{}) (*types.ResourceDiff, error) {
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
	
	if desiredState == currentState {
		diff.Action = types.ActionNoop
		diff.Reason = "package already in desired state"
		return diff, nil
	}
	
	switch desiredState {
	case "present", "latest":
		if currentState == "absent" {
			diff.Action = types.ActionCreate
			diff.Reason = "package needs to be installed"
			diff.Changes["state"] = map[string]interface{}{
				"from": currentState,
				"to":   desiredState,
			}
		} else {
			// Check if version update is needed for "latest"
			if desiredState == "latest" {
				diff.Action = types.ActionUpdate
				diff.Reason = "package needs to be updated to latest version"
				diff.Changes["state"] = map[string]interface{}{
					"from": currentState,
					"to":   desiredState,
				}
			} else {
				diff.Action = types.ActionNoop
				diff.Reason = "package already installed"
			}
		}
	case "absent":
		if currentState == "present" {
			diff.Action = types.ActionDelete
			diff.Reason = "package needs to be removed"
			diff.Changes["state"] = map[string]interface{}{
				"from": currentState,
				"to":   desiredState,
			}
		} else {
			diff.Action = types.ActionNoop
			diff.Reason = "package already absent"
		}
	}
	
	return diff, nil
}

// Apply applies the changes to bring the package to desired state
func (p *PkgProvider) Apply(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	switch diff.Action {
	case types.ActionCreate:
		return p.installPackage(ctx, resource)
	case types.ActionUpdate:
		return p.updatePackage(ctx, resource)
	case types.ActionDelete:
		return p.removePackage(ctx, resource)
	case types.ActionNoop:
		return nil
	default:
		return fmt.Errorf("unsupported action: %s", diff.Action)
	}
}

// isPackageInstalled checks if a package is installed and returns its version
func (p *PkgProvider) isPackageInstalled(ctx context.Context, packageName string) (bool, string, error) {
	// Try dpkg (Debian/Ubuntu) first
	cmd := fmt.Sprintf("dpkg -l %s 2>/dev/null | grep '^ii' | wc -l", shellEscape(packageName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, "", err
	}
	
	if result.ExitCode == 0 {
		count, err := strconv.Atoi(strings.TrimSpace(result.Stdout))
		if err == nil {
			if count > 0 {
				// Get version
				versionCmd := fmt.Sprintf("dpkg -l %s 2>/dev/null | grep '^ii' | awk '{print $3}'", shellEscape(packageName))
				versionResult, err := p.connection.Execute(ctx, versionCmd)
				if err == nil && versionResult.ExitCode == 0 {
					version := strings.TrimSpace(versionResult.Stdout)
					return true, version, nil
				}
				return true, "", nil
			} else {
				return false, "", nil
			}
		}
	}
	
	// If dpkg fails, try rpm (RedHat/CentOS/Fedora)
	cmd = fmt.Sprintf("rpm -q %s >/dev/null 2>&1 && echo 1 || echo 0", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, "", err
	}
	
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == "1" {
		// Get version
		versionCmd := fmt.Sprintf("rpm -q --queryformat '%%{VERSION}' %s 2>/dev/null", shellEscape(packageName))
		versionResult, err := p.connection.Execute(ctx, versionCmd)
		if err == nil && versionResult.ExitCode == 0 {
			version := strings.TrimSpace(versionResult.Stdout)
			return true, version, nil
		}
		return true, "", nil
	}
	
	// If rpm fails, try brew (macOS)
	cmd = fmt.Sprintf("brew list %s >/dev/null 2>&1 && echo 1 || echo 0", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, "", err
	}
	
	if result.ExitCode == 0 && strings.TrimSpace(result.Stdout) == "1" {
		return true, "", nil
	}
	
	return false, "", nil
}

// installPackage installs a package
func (p *PkgProvider) installPackage(ctx context.Context, resource *types.Resource) error {
	packageName := resource.Name
	version := ""
	
	if v, ok := resource.Properties["version"]; ok {
		version = v.(string)
	}
	
	// Try different package managers
	
	// Try apt (Debian/Ubuntu)
	var cmd string
	if version != "" {
		cmd = fmt.Sprintf("apt-get update && apt-get install -y %s=%s", shellEscape(packageName), shellEscape(version))
	} else {
		cmd = fmt.Sprintf("apt-get update && apt-get install -y %s", shellEscape(packageName))
	}
	
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try yum (RedHat/CentOS)
	if version != "" {
		cmd = fmt.Sprintf("yum install -y %s-%s", shellEscape(packageName), shellEscape(version))
	} else {
		cmd = fmt.Sprintf("yum install -y %s", shellEscape(packageName))
	}
	
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try dnf (Fedora)
	if version != "" {
		cmd = fmt.Sprintf("dnf install -y %s-%s", shellEscape(packageName), shellEscape(version))
	} else {
		cmd = fmt.Sprintf("dnf install -y %s", shellEscape(packageName))
	}
	
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try brew (macOS)
	cmd = fmt.Sprintf("brew install %s", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to install package %s: %s", packageName, result.Stderr)
	}
	
	return nil
}

// updatePackage updates a package to the latest version
func (p *PkgProvider) updatePackage(ctx context.Context, resource *types.Resource) error {
	packageName := resource.Name
	
	// Try different package managers
	
	// Try apt (Debian/Ubuntu)
	cmd := fmt.Sprintf("apt-get update && apt-get upgrade -y %s", shellEscape(packageName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try yum (RedHat/CentOS)
	cmd = fmt.Sprintf("yum update -y %s", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try dnf (Fedora)
	cmd = fmt.Sprintf("dnf upgrade -y %s", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try brew (macOS)
	cmd = fmt.Sprintf("brew upgrade %s", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update package %s: %w", packageName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to update package %s: %s", packageName, result.Stderr)
	}
	
	return nil
}

// removePackage removes a package
func (p *PkgProvider) removePackage(ctx context.Context, resource *types.Resource) error {
	packageName := resource.Name
	
	// Try different package managers
	
	// Try apt (Debian/Ubuntu)
	cmd := fmt.Sprintf("apt-get remove -y %s", shellEscape(packageName))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try yum (RedHat/CentOS)
	cmd = fmt.Sprintf("yum remove -y %s", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try dnf (Fedora)
	cmd = fmt.Sprintf("dnf remove -y %s", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove package %s: %w", packageName, err)
	}
	
	if result.ExitCode == 0 {
		return nil
	}
	
	// Try brew (macOS)
	cmd = fmt.Sprintf("brew uninstall %s", shellEscape(packageName))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove package %s: %w", packageName, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to remove package %s: %s", packageName, result.Stderr)
	}
	
	return nil
}
