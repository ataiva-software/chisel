package providers

import (
	"context"
	"testing"

	"github.com/ao/chisel/pkg/ssh"
	"github.com/ao/chisel/pkg/types"
)

func TestFileProvider_Type(t *testing.T) {
	provider := NewFileProvider(nil)
	if provider.Type() != "file" {
		t.Errorf("Expected type 'file', got '%s'", provider.Type())
	}
}

func TestFileProvider_Validate(t *testing.T) {
	provider := NewFileProvider(nil)

	tests := []struct {
		name     string
		resource *types.Resource
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid file resource",
			resource: &types.Resource{
				Type: "file",
				Name: "test-file",
				Properties: map[string]interface{}{
					"path": "/etc/test.conf",
				},
			},
			wantErr: false,
		},
		{
			name: "valid file resource with all properties",
			resource: &types.Resource{
				Type:  "file",
				Name:  "test-file",
				State: types.StatePresent,
				Properties: map[string]interface{}{
					"path":    "/etc/test.conf",
					"content": "test content",
					"mode":    "0644",
					"owner":   "root",
					"group":   "root",
				},
			},
			wantErr: false,
		},
		{
			name: "missing path",
			resource: &types.Resource{
				Type: "file",
				Name: "test-file",
				Properties: map[string]interface{}{
					"content": "test content",
				},
			},
			wantErr: true,
			errMsg:  "file resource must have a 'path' property",
		},
		{
			name: "empty path",
			resource: &types.Resource{
				Type: "file",
				Name: "test-file",
				Properties: map[string]interface{}{
					"path": "",
				},
			},
			wantErr: true,
			errMsg:  "file resource must have a 'path' property",
		},
		{
			name: "invalid mode format",
			resource: &types.Resource{
				Type: "file",
				Name: "test-file",
				Properties: map[string]interface{}{
					"path": "/etc/test.conf",
					"mode": "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid file mode 'invalid': strconv.ParseUint: parsing \"invalid\": invalid syntax",
		},
		{
			name: "mode not string",
			resource: &types.Resource{
				Type: "file",
				Name: "test-file",
				Properties: map[string]interface{}{
					"path": "/etc/test.conf",
					"mode": 644,
				},
			},
			wantErr: true,
			errMsg:  "file mode must be a string (e.g., '0644')",
		},
		{
			name: "invalid state",
			resource: &types.Resource{
				Type:  "file",
				Name:  "test-file",
				State: "invalid",
				Properties: map[string]interface{}{
					"path": "/etc/test.conf",
				},
			},
			wantErr: true,
			errMsg:  "file resource state must be 'present' or 'absent', got 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Validate(tt.resource)
			if tt.wantErr {
				if err == nil {
					t.Errorf("FileProvider.Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("FileProvider.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("FileProvider.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestFileProvider_Read_FileNotExists(t *testing.T) {
	mockConn := &MockSSHConnection{
		responses: map[string]*ssh.ExecuteResult{
			"test -f '/etc/test.conf'": {
				ExitCode: 1, // File doesn't exist
			},
		},
	}

	provider := NewFileProvider(mockConn)
	resource := &types.Resource{
		Type: "file",
		Name: "test-file",
		Properties: map[string]interface{}{
			"path": "/etc/test.conf",
		},
	}

	ctx := context.Background()
	current, err := provider.Read(ctx, resource)
	if err != nil {
		t.Fatalf("FileProvider.Read() unexpected error = %v", err)
	}

	if current["exists"] != false {
		t.Errorf("Expected exists=false, got %v", current["exists"])
	}
	if current["state"] != types.StateAbsent {
		t.Errorf("Expected state=absent, got %v", current["state"])
	}
	if current["path"] != "/etc/test.conf" {
		t.Errorf("Expected path='/etc/test.conf', got %v", current["path"])
	}
}

func TestFileProvider_Read_FileExists(t *testing.T) {
	mockConn := &MockSSHConnection{
		responses: map[string]*ssh.ExecuteResult{
			"test -f '/etc/test.conf'": {
				ExitCode: 0, // File exists
			},
			"stat -c '%s:%a:%U:%G' '/etc/test.conf'": {
				ExitCode: 0,
				Stdout:   "1024:644:root:root",
			},
		},
	}

	provider := NewFileProvider(mockConn)
	resource := &types.Resource{
		Type: "file",
		Name: "test-file",
		Properties: map[string]interface{}{
			"path": "/etc/test.conf",
		},
	}

	ctx := context.Background()
	current, err := provider.Read(ctx, resource)
	if err != nil {
		t.Fatalf("FileProvider.Read() unexpected error = %v", err)
	}

	if current["exists"] != true {
		t.Errorf("Expected exists=true, got %v", current["exists"])
	}
	if current["state"] != types.StatePresent {
		t.Errorf("Expected state=present, got %v", current["state"])
	}
	if current["size"] != "1024" {
		t.Errorf("Expected size='1024', got %v", current["size"])
	}
	if current["mode"] != "644" {
		t.Errorf("Expected mode='644', got %v", current["mode"])
	}
	if current["owner"] != "root" {
		t.Errorf("Expected owner='root', got %v", current["owner"])
	}
	if current["group"] != "root" {
		t.Errorf("Expected group='root', got %v", current["group"])
	}
}

func TestFileProvider_Diff_CreateFile(t *testing.T) {
	provider := NewFileProvider(nil)
	resource := &types.Resource{
		Type: "file",
		Name: "test-file",
		Properties: map[string]interface{}{
			"path":    "/etc/test.conf",
			"content": "test content",
		},
	}

	current := map[string]interface{}{
		"path":   "/etc/test.conf",
		"exists": false,
		"state":  types.StateAbsent,
	}

	ctx := context.Background()
	diff, err := provider.Diff(ctx, resource, current)
	if err != nil {
		t.Fatalf("FileProvider.Diff() unexpected error = %v", err)
	}

	if diff.Action != types.ActionCreate {
		t.Errorf("Expected action=create, got %v", diff.Action)
	}
	if diff.Reason != "file does not exist" {
		t.Errorf("Expected reason='file does not exist', got %v", diff.Reason)
	}
}

func TestFileProvider_Diff_DeleteFile(t *testing.T) {
	provider := NewFileProvider(nil)
	resource := &types.Resource{
		Type:  "file",
		Name:  "test-file",
		State: types.StateAbsent,
		Properties: map[string]interface{}{
			"path": "/etc/test.conf",
		},
	}

	current := map[string]interface{}{
		"path":   "/etc/test.conf",
		"exists": true,
		"state":  types.StatePresent,
	}

	ctx := context.Background()
	diff, err := provider.Diff(ctx, resource, current)
	if err != nil {
		t.Fatalf("FileProvider.Diff() unexpected error = %v", err)
	}

	if diff.Action != types.ActionDelete {
		t.Errorf("Expected action=delete, got %v", diff.Action)
	}
	if diff.Reason != "file should be absent but exists" {
		t.Errorf("Expected reason='file should be absent but exists', got %v", diff.Reason)
	}
}

func TestFileProvider_Diff_UpdateContent(t *testing.T) {
	provider := NewFileProvider(nil)
	resource := &types.Resource{
		Type: "file",
		Name: "test-file",
		Properties: map[string]interface{}{
			"path":    "/etc/test.conf",
			"content": "new content",
		},
	}

	current := map[string]interface{}{
		"path":    "/etc/test.conf",
		"exists":  true,
		"state":   types.StatePresent,
		"content": "old content",
	}

	ctx := context.Background()
	diff, err := provider.Diff(ctx, resource, current)
	if err != nil {
		t.Fatalf("FileProvider.Diff() unexpected error = %v", err)
	}

	if diff.Action != types.ActionUpdate {
		t.Errorf("Expected action=update, got %v", diff.Action)
	}

	contentChange, ok := diff.Changes["content"]
	if !ok {
		t.Fatal("Expected content change in diff")
	}

	changeMap, ok := contentChange.(map[string]interface{})
	if !ok {
		t.Fatal("Expected content change to be a map")
	}

	if changeMap["from"] != "old content" {
		t.Errorf("Expected from='old content', got %v", changeMap["from"])
	}
	if changeMap["to"] != "new content" {
		t.Errorf("Expected to='new content', got %v", changeMap["to"])
	}
}

func TestFileProvider_Diff_NoChanges(t *testing.T) {
	provider := NewFileProvider(nil)
	resource := &types.Resource{
		Type: "file",
		Name: "test-file",
		Properties: map[string]interface{}{
			"path":    "/etc/test.conf",
			"content": "same content",
			"mode":    "644",
		},
	}

	current := map[string]interface{}{
		"path":    "/etc/test.conf",
		"exists":  true,
		"state":   types.StatePresent,
		"content": "same content",
		"mode":    "644",
	}

	ctx := context.Background()
	diff, err := provider.Diff(ctx, resource, current)
	if err != nil {
		t.Fatalf("FileProvider.Diff() unexpected error = %v", err)
	}

	if diff.Action != types.ActionNoop {
		t.Errorf("Expected action=noop, got %v", diff.Action)
	}
	if diff.Reason != "file is already in desired state" {
		t.Errorf("Expected reason='file is already in desired state', got %v", diff.Reason)
	}
}

func TestShellEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "simple",
			expected: "'simple'",
		},
		{
			input:    "with spaces",
			expected: "'with spaces'",
		},
		{
			input:    "with'quote",
			expected: "'with'\"'\"'quote'",
		},
		{
			input:    "/path/to/file",
			expected: "'/path/to/file'",
		},
		{
			input:    "",
			expected: "''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shellEscape(tt.input)
			if result != tt.expected {
				t.Errorf("shellEscape(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// MockSSHConnection is a mock implementation of SSH connection for testing
type MockSSHConnection struct {
	responses map[string]*ssh.ExecuteResult
}

func (m *MockSSHConnection) Execute(ctx context.Context, command string) (*ssh.ExecuteResult, error) {
	if result, ok := m.responses[command]; ok {
		return result, nil
	}
	// Default response for unknown commands
	return &ssh.ExecuteResult{
		Command:  command,
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
	}, nil
}

func (m *MockSSHConnection) Connect(ctx context.Context) error {
	return nil
}

func (m *MockSSHConnection) Close() error {
	return nil
}

// Ensure MockSSHConnection implements ssh.Executor
var _ ssh.Executor = (*MockSSHConnection)(nil)
