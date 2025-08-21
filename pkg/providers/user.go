package providers

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ataiva-software/chisel/pkg/ssh"
	"github.com/ataiva-software/chisel/pkg/types"
)

// UserProvider manages user resources
type UserProvider struct {
	connection ssh.Executor
}

// NewUserProvider creates a new user provider
func NewUserProvider(connection ssh.Executor) *UserProvider {
	return &UserProvider{
		connection: connection,
	}
}

// Type returns the resource type this provider handles
func (p *UserProvider) Type() string {
	return "user"
}

// Validate validates the user resource configuration
func (p *UserProvider) Validate(resource *types.Resource) error {
	// Check state - can be in State field or Properties map
	var state string
	if resource.State != "" {
		state = string(resource.State)
	} else if stateInterface, ok := resource.Properties["state"]; ok {
		if stateStr, ok := stateInterface.(string); ok {
			state = stateStr
		} else {
			return fmt.Errorf("user 'state' must be a string")
		}
	} else {
		return fmt.Errorf("user resource must have 'state' property")
	}
	
	// Validate state values
	validStates := map[string]bool{
		"present": true,
		"absent":  true,
	}
	
	if !validStates[state] {
		return fmt.Errorf("invalid user state '%s', must be one of: present, absent", state)
	}
	
	// Validate optional properties
	if uid, ok := resource.Properties["uid"]; ok {
		if _, ok := uid.(int); !ok {
			return fmt.Errorf("user 'uid' must be an integer")
		}
	}
	
	if gid, ok := resource.Properties["gid"]; ok {
		if _, ok := gid.(int); !ok {
			return fmt.Errorf("user 'gid' must be an integer")
		}
	}
	
	if home, ok := resource.Properties["home"]; ok {
		if _, ok := home.(string); !ok {
			return fmt.Errorf("user 'home' must be a string")
		}
	}
	
	if shell, ok := resource.Properties["shell"]; ok {
		if _, ok := shell.(string); !ok {
			return fmt.Errorf("user 'shell' must be a string")
		}
	}
	
	if groups, ok := resource.Properties["groups"]; ok {
		if groupsSlice, ok := groups.([]interface{}); ok {
			for _, group := range groupsSlice {
				if _, ok := group.(string); !ok {
					return fmt.Errorf("user 'groups' must be an array of strings")
				}
			}
		} else if groupsStrSlice, ok := groups.([]string); ok {
			// Already valid
			_ = groupsStrSlice
		} else {
			return fmt.Errorf("user 'groups' must be an array of strings")
		}
	}
	
	return nil
}

// Read reads the current state of the user
func (p *UserProvider) Read(ctx context.Context, resource *types.Resource) (map[string]interface{}, error) {
	username := resource.Name
	
	// Check if user exists
	exists, err := p.userExists(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	state := map[string]interface{}{}
	
	if !exists {
		state["state"] = "absent"
		return state, nil
	}
	
	state["state"] = "present"
	
	// Get user details
	userInfo, err := p.getUserInfo(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	
	// Merge user info into state
	for k, v := range userInfo {
		state[k] = v
	}
	
	return state, nil
}

// Diff compares desired vs current state and returns the differences
func (p *UserProvider) Diff(ctx context.Context, resource *types.Resource, current map[string]interface{}) (*types.ResourceDiff, error) {
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
	
	// Handle state changes
	if desiredState != currentState {
		switch desiredState {
		case "present":
			if currentState == "absent" {
				diff.Action = types.ActionCreate
				diff.Reason = "user needs to be created"
				diff.Changes["state"] = map[string]interface{}{
					"from": currentState,
					"to":   desiredState,
				}
				return diff, nil
			}
		case "absent":
			if currentState == "present" {
				diff.Action = types.ActionDelete
				diff.Reason = "user needs to be removed"
				diff.Changes["state"] = map[string]interface{}{
					"from": currentState,
					"to":   desiredState,
				}
				return diff, nil
			}
		}
	}
	
	// If user exists and should exist, check for property changes
	if desiredState == "present" && currentState == "present" {
		hasChanges := false
		
		// Check each property for changes
		properties := []string{"uid", "gid", "home", "shell", "groups"}
		for _, prop := range properties {
			if desired, ok := resource.Properties[prop]; ok {
				if current, ok := current[prop]; ok {
					if !reflect.DeepEqual(desired, current) {
						hasChanges = true
						diff.Changes[prop] = map[string]interface{}{
							"from": current,
							"to":   desired,
						}
					}
				} else {
					hasChanges = true
					diff.Changes[prop] = map[string]interface{}{
						"from": nil,
						"to":   desired,
					}
				}
			}
		}
		
		if hasChanges {
			diff.Action = types.ActionUpdate
			diff.Reason = "user properties need to be updated"
		} else {
			diff.Action = types.ActionNoop
			diff.Reason = "user already in desired state"
		}
	} else {
		diff.Action = types.ActionNoop
		diff.Reason = "user already in desired state"
	}
	
	return diff, nil
}

// Apply applies the changes to bring the user to desired state
func (p *UserProvider) Apply(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	switch diff.Action {
	case types.ActionCreate:
		return p.createUser(ctx, resource)
	case types.ActionUpdate:
		return p.updateUser(ctx, resource, diff)
	case types.ActionDelete:
		return p.deleteUser(ctx, resource)
	case types.ActionNoop:
		return nil
	default:
		return fmt.Errorf("unsupported action: %s", diff.Action)
	}
}

// userExists checks if a user exists
func (p *UserProvider) userExists(ctx context.Context, username string) (bool, error) {
	cmd := fmt.Sprintf("id -u %s 2>/dev/null", shellEscape(username))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return false, err
	}
	
	return result.ExitCode == 0, nil
}

// getUserInfo gets detailed information about a user
func (p *UserProvider) getUserInfo(ctx context.Context, username string) (map[string]interface{}, error) {
	info := make(map[string]interface{})
	
	// Get passwd entry
	cmd := fmt.Sprintf("getent passwd %s", shellEscape(username))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("failed to get user info: %s", result.Stderr)
	}
	
	// Parse passwd entry: username:x:uid:gid:gecos:home:shell
	parts := strings.Split(strings.TrimSpace(result.Stdout), ":")
	if len(parts) >= 7 {
		if uid, err := strconv.Atoi(parts[2]); err == nil {
			info["uid"] = uid
		}
		if gid, err := strconv.Atoi(parts[3]); err == nil {
			info["gid"] = gid
		}
		info["home"] = parts[5]
		info["shell"] = parts[6]
	}
	
	// Get groups
	cmd = fmt.Sprintf("groups %s", shellEscape(username))
	result, err = p.connection.Execute(ctx, cmd)
	if err == nil && result.ExitCode == 0 {
		// Parse groups output: "username : group1 group2 group3"
		output := strings.TrimSpace(result.Stdout)
		if colonIndex := strings.Index(output, ":"); colonIndex != -1 {
			groupsStr := strings.TrimSpace(output[colonIndex+1:])
			if groupsStr != "" {
				groups := strings.Fields(groupsStr)
				info["groups"] = groups
			}
		}
	}
	
	return info, nil
}

// createUser creates a new user
func (p *UserProvider) createUser(ctx context.Context, resource *types.Resource) error {
	username := resource.Name
	
	// Build useradd command
	args := []string{"useradd"}
	
	if uid, ok := resource.Properties["uid"]; ok {
		args = append(args, "-u", fmt.Sprintf("%d", uid.(int)))
	}
	
	if gid, ok := resource.Properties["gid"]; ok {
		args = append(args, "-g", fmt.Sprintf("%d", gid.(int)))
	}
	
	if home, ok := resource.Properties["home"]; ok {
		args = append(args, "-d", home.(string))
	}
	
	if shell, ok := resource.Properties["shell"]; ok {
		args = append(args, "-s", shell.(string))
	}
	
	args = append(args, shellEscape(username))
	
	cmd := strings.Join(args, " ")
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", username, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to create user %s: %s", username, result.Stderr)
	}
	
	// Add user to additional groups if specified
	if groups, ok := resource.Properties["groups"]; ok {
		if err := p.setUserGroups(ctx, username, groups); err != nil {
			return err
		}
	}
	
	return nil
}

// updateUser updates an existing user
func (p *UserProvider) updateUser(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	username := resource.Name
	
	// Build usermod command for basic properties
	args := []string{"usermod"}
	hasChanges := false
	
	if change, ok := diff.Changes["uid"]; ok {
		changeMap := change.(map[string]interface{})
		newUID := changeMap["to"].(int)
		args = append(args, "-u", fmt.Sprintf("%d", newUID))
		hasChanges = true
	}
	
	if change, ok := diff.Changes["gid"]; ok {
		changeMap := change.(map[string]interface{})
		newGID := changeMap["to"].(int)
		args = append(args, "-g", fmt.Sprintf("%d", newGID))
		hasChanges = true
	}
	
	if change, ok := diff.Changes["home"]; ok {
		changeMap := change.(map[string]interface{})
		newHome := changeMap["to"].(string)
		args = append(args, "-d", newHome)
		hasChanges = true
	}
	
	if change, ok := diff.Changes["shell"]; ok {
		changeMap := change.(map[string]interface{})
		newShell := changeMap["to"].(string)
		args = append(args, "-s", newShell)
		hasChanges = true
	}
	
	if hasChanges {
		args = append(args, shellEscape(username))
		cmd := strings.Join(args, " ")
		result, err := p.connection.Execute(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update user %s: %w", username, err)
		}
		
		if result.ExitCode != 0 {
			return fmt.Errorf("failed to update user %s: %s", username, result.Stderr)
		}
	}
	
	// Handle groups changes separately
	if change, ok := diff.Changes["groups"]; ok {
		changeMap := change.(map[string]interface{})
		newGroups := changeMap["to"]
		if err := p.setUserGroups(ctx, username, newGroups); err != nil {
			return err
		}
	}
	
	return nil
}

// deleteUser removes a user
func (p *UserProvider) deleteUser(ctx context.Context, resource *types.Resource) error {
	username := resource.Name
	
	cmd := fmt.Sprintf("userdel %s", shellEscape(username))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", username, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to delete user %s: %s", username, result.Stderr)
	}
	
	return nil
}

// setUserGroups sets the groups for a user
func (p *UserProvider) setUserGroups(ctx context.Context, username string, groups interface{}) error {
	var groupList []string
	
	// Convert groups to string slice
	switch g := groups.(type) {
	case []string:
		groupList = g
	case []interface{}:
		for _, group := range g {
			if groupStr, ok := group.(string); ok {
				groupList = append(groupList, groupStr)
			}
		}
	default:
		return fmt.Errorf("invalid groups format")
	}
	
	if len(groupList) == 0 {
		return nil
	}
	
	// Use usermod -G to set supplementary groups
	groupsStr := strings.Join(groupList, ",")
	cmd := fmt.Sprintf("usermod -G %s %s", shellEscape(groupsStr), shellEscape(username))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to set groups for user %s: %w", username, err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to set groups for user %s: %s", username, result.Stderr)
	}
	
	return nil
}
