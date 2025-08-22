package providers

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ataiva-software/forge/pkg/ssh"
	"github.com/ataiva-software/forge/pkg/templating"
	"github.com/ataiva-software/forge/pkg/types"
)

// FileProvider manages file resources
type FileProvider struct {
	connection ssh.Executor
}

// NewFileProvider creates a new file provider
func NewFileProvider(connection ssh.Executor) *FileProvider {
	return &FileProvider{
		connection: connection,
	}
}

// Type returns the resource type this provider handles
func (p *FileProvider) Type() string {
	return "file"
}

// Validate validates the file resource configuration
func (p *FileProvider) Validate(resource *types.Resource) error {
	if err := resource.Validate(); err != nil {
		return err
	}

	// Check required properties
	path, ok := resource.Properties["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("file resource must have a 'path' property")
	}

	// Validate mode if provided
	if mode, exists := resource.Properties["mode"]; exists {
		if modeStr, ok := mode.(string); ok {
			if _, err := strconv.ParseUint(modeStr, 8, 32); err != nil {
				return fmt.Errorf("invalid file mode '%s': %w", modeStr, err)
			}
		} else {
			return fmt.Errorf("file mode must be a string (e.g., '0644')")
		}
	}

	// Validate state
	if resource.State != "" && resource.State != types.StatePresent && resource.State != types.StateAbsent {
		return fmt.Errorf("file resource state must be 'present' or 'absent', got '%s'", resource.State)
	}

	return nil
}

// Read reads the current state of the file resource
func (p *FileProvider) Read(ctx context.Context, resource *types.Resource) (map[string]interface{}, error) {
	path := resource.Properties["path"].(string)
	
	current := make(map[string]interface{})
	current["path"] = path

	// Check if file exists
	result, err := p.connection.Execute(ctx, fmt.Sprintf("test -f %s", shellEscape(path)))
	if err != nil {
		return nil, fmt.Errorf("failed to check file existence: %w", err)
	}

	if result.ExitCode != 0 {
		current["exists"] = false
		current["state"] = types.StateAbsent
		return current, nil
	}

	current["exists"] = true
	current["state"] = types.StatePresent

	// Get file stats
	statCmd := fmt.Sprintf("stat -c '%%s:%%a:%%U:%%G' %s", shellEscape(path))
	result, err = p.connection.Execute(ctx, statCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	if result.ExitCode == 0 {
		parts := strings.Split(strings.TrimSpace(result.Stdout), ":")
		if len(parts) == 4 {
			current["size"] = parts[0]
			current["mode"] = parts[1]
			current["owner"] = parts[2]
			current["group"] = parts[3]
		}
	}

	// Get file content if requested
	if _, needsContent := resource.Properties["content"]; needsContent {
		result, err = p.connection.Execute(ctx, fmt.Sprintf("cat %s", shellEscape(path)))
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}
		if result.ExitCode == 0 {
			current["content"] = result.Stdout
		}
	}

	// Get file checksum if content is provided
	if _, hasContent := resource.Properties["content"]; hasContent {
		result, err = p.connection.Execute(ctx, fmt.Sprintf("md5sum %s | cut -d' ' -f1", shellEscape(path)))
		if err != nil {
			return nil, fmt.Errorf("failed to get file checksum: %w", err)
		}
		if result.ExitCode == 0 {
			current["checksum"] = strings.TrimSpace(result.Stdout)
		}
	}

	return current, nil
}

// Diff compares desired vs current state and returns the differences
func (p *FileProvider) Diff(ctx context.Context, resource *types.Resource, current map[string]interface{}) (*types.ResourceDiff, error) {
	diff := &types.ResourceDiff{
		ResourceID: resource.ResourceID(),
		Changes:    make(map[string]interface{}),
	}

	desiredState := resource.State
	if desiredState == "" {
		desiredState = types.StatePresent
	}

	exists := current["exists"].(bool)

	// Handle state changes
	if desiredState == types.StateAbsent {
		if exists {
			diff.Action = types.ActionDelete
			diff.Reason = "file should be absent but exists"
		} else {
			diff.Action = types.ActionNoop
			diff.Reason = "file is already absent"
		}
		return diff, nil
	}

	// File should be present
	if !exists {
		diff.Action = types.ActionCreate
		diff.Reason = "file does not exist"
		diff.Changes["state"] = map[string]interface{}{
			"from": types.StateAbsent,
			"to":   types.StatePresent,
		}
		return diff, nil
	}

	// File exists, check for changes
	hasChanges := false

	// Check content
	if desiredContent, ok := resource.Properties["content"].(string); ok {
		currentContent, hasCurrentContent := current["content"].(string)
		if !hasCurrentContent || currentContent != desiredContent {
			hasChanges = true
			diff.Changes["content"] = map[string]interface{}{
				"from": currentContent,
				"to":   desiredContent,
			}
		}
	}

	// Check mode
	if desiredMode, ok := resource.Properties["mode"].(string); ok {
		currentMode, hasCurrentMode := current["mode"].(string)
		if !hasCurrentMode || currentMode != desiredMode {
			hasChanges = true
			diff.Changes["mode"] = map[string]interface{}{
				"from": currentMode,
				"to":   desiredMode,
			}
		}
	}

	// Check owner
	if desiredOwner, ok := resource.Properties["owner"].(string); ok {
		currentOwner, hasCurrentOwner := current["owner"].(string)
		if !hasCurrentOwner || currentOwner != desiredOwner {
			hasChanges = true
			diff.Changes["owner"] = map[string]interface{}{
				"from": currentOwner,
				"to":   desiredOwner,
			}
		}
	}

	// Check group
	if desiredGroup, ok := resource.Properties["group"].(string); ok {
		currentGroup, hasCurrentGroup := current["group"].(string)
		if !hasCurrentGroup || currentGroup != desiredGroup {
			hasChanges = true
			diff.Changes["group"] = map[string]interface{}{
				"from": currentGroup,
				"to":   desiredGroup,
			}
		}
	}

	if hasChanges {
		diff.Action = types.ActionUpdate
		diff.Reason = "file properties need to be updated"
	} else {
		diff.Action = types.ActionNoop
		diff.Reason = "file is already in desired state"
	}

	return diff, nil
}

// Apply applies the changes to bring the resource to desired state
func (p *FileProvider) Apply(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	path := resource.Properties["path"].(string)

	switch diff.Action {
	case types.ActionDelete:
		return p.deleteFile(ctx, path)
	case types.ActionCreate:
		return p.createFile(ctx, resource)
	case types.ActionUpdate:
		return p.updateFile(ctx, resource, diff)
	case types.ActionNoop:
		return nil
	default:
		return fmt.Errorf("unsupported action: %s", diff.Action)
	}
}

// createFile creates a new file
func (p *FileProvider) createFile(ctx context.Context, resource *types.Resource) error {
	path := resource.Properties["path"].(string)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		cmd := fmt.Sprintf("mkdir -p %s", shellEscape(dir))
		result, err := p.connection.Execute(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("failed to create directory %s: %s", dir, result.Stderr)
		}
	}

	// Handle content - check for template first, then regular content
	content, err := p.resolveContent(resource)
	if err != nil {
		return fmt.Errorf("failed to resolve content: %w", err)
	}

	if content != "" {
		if err := p.writeFileContent(ctx, path, content); err != nil {
			return err
		}
	} else {
		// Create empty file
		cmd := fmt.Sprintf("touch %s", shellEscape(path))
		result, err := p.connection.Execute(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("failed to create file %s: %s", path, result.Stderr)
		}
	}

	// Set permissions and ownership
	return p.setFileAttributes(ctx, resource)
}

// resolveContent resolves the content for a file, handling templates if specified
func (p *FileProvider) resolveContent(resource *types.Resource) (string, error) {
	// Check for template content first
	if templateStr, ok := resource.Properties["template"].(string); ok {
		vars := make(map[string]interface{})
		if varsInterface, ok := resource.Properties["vars"]; ok {
			if varsMap, ok := varsInterface.(map[string]interface{}); ok {
				vars = varsMap
			}
		}
		
		engine := templating.NewTemplateEngine()
		return engine.Render(templateStr, vars)
	}
	
	// Check for template file
	if templateFile, ok := resource.Properties["template_file"].(string); ok {
		vars := make(map[string]interface{})
		if varsInterface, ok := resource.Properties["vars"]; ok {
			if varsMap, ok := varsInterface.(map[string]interface{}); ok {
				vars = varsMap
			}
		}
		
		engine := templating.NewTemplateEngine()
		return engine.RenderFile(templateFile, vars)
	}
	
	// Fall back to regular content
	if content, ok := resource.Properties["content"].(string); ok {
		return content, nil
	}
	
	return "", nil
}

// updateFile updates an existing file
func (p *FileProvider) updateFile(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff) error {
	// Update content if changed
	if contentChange, ok := diff.Changes["content"]; ok {
		if changeMap, ok := contentChange.(map[string]interface{}); ok {
			if newContent, ok := changeMap["to"].(string); ok {
				path := resource.Properties["path"].(string)
				if err := p.writeFileContent(ctx, path, newContent); err != nil {
					return err
				}
			}
		}
	}

	// Update attributes if changed
	if _, hasMode := diff.Changes["mode"]; hasMode {
		if err := p.setFileAttributes(ctx, resource); err != nil {
			return err
		}
	}
	if _, hasOwner := diff.Changes["owner"]; hasOwner {
		if err := p.setFileAttributes(ctx, resource); err != nil {
			return err
		}
	}
	if _, hasGroup := diff.Changes["group"]; hasGroup {
		if err := p.setFileAttributes(ctx, resource); err != nil {
			return err
		}
	}

	return nil
}

// deleteFile removes a file
func (p *FileProvider) deleteFile(ctx context.Context, path string) error {
	cmd := fmt.Sprintf("rm -f %s", shellEscape(path))
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to delete file %s: %s", path, result.Stderr)
	}
	return nil
}

// writeFileContent writes content to a file
func (p *FileProvider) writeFileContent(ctx context.Context, path, content string) error {
	// Use a temporary file and atomic move for safety
	tempPath := path + ".chisel.tmp"
	
	// Write to temporary file
	cmd := fmt.Sprintf("cat > %s << 'CHISEL_EOF'\n%s\nCHISEL_EOF", shellEscape(tempPath), content)
	result, err := p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to write content to %s: %w", path, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to write content to %s: %s", path, result.Stderr)
	}

	// Atomic move
	cmd = fmt.Sprintf("mv %s %s", shellEscape(tempPath), shellEscape(path))
	result, err = p.connection.Execute(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to move temporary file to %s: %w", path, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to move temporary file to %s: %s", path, result.Stderr)
	}

	return nil
}

// setFileAttributes sets file mode, owner, and group
func (p *FileProvider) setFileAttributes(ctx context.Context, resource *types.Resource) error {
	path := resource.Properties["path"].(string)

	// Set mode
	if mode, ok := resource.Properties["mode"].(string); ok {
		cmd := fmt.Sprintf("chmod %s %s", mode, shellEscape(path))
		result, err := p.connection.Execute(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set mode on %s: %w", path, err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("failed to set mode on %s: %s", path, result.Stderr)
		}
	}

	// Set owner and group
	owner, hasOwner := resource.Properties["owner"].(string)
	group, hasGroup := resource.Properties["group"].(string)
	
	if hasOwner || hasGroup {
		chownArg := ""
		if hasOwner && hasGroup {
			chownArg = fmt.Sprintf("%s:%s", owner, group)
		} else if hasOwner {
			chownArg = owner
		} else if hasGroup {
			chownArg = fmt.Sprintf(":%s", group)
		}

		cmd := fmt.Sprintf("chown %s %s", chownArg, shellEscape(path))
		result, err := p.connection.Execute(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set ownership on %s: %w", path, err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("failed to set ownership on %s: %s", path, result.Stderr)
		}
	}

	return nil
}

// shellEscape escapes a string for safe use in shell commands
func shellEscape(s string) string {
	// Simple shell escaping - wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
