package providers

import (
	"context"
	"testing"

	"github.com/ataiva-software/chisel/pkg/ssh"
	"github.com/ataiva-software/chisel/pkg/types"
)

func TestShellProvider_Type(t *testing.T) {
	provider := NewShellProvider(nil)
	if provider.Type() != "shell" {
		t.Errorf("Expected type 'shell', got '%s'", provider.Type())
	}
}

func TestShellProvider_Validate(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		wantErr  bool
	}{
		{
			name: "valid shell resource",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello world'",
				},
			},
			wantErr: false,
		},
		{
			name: "valid shell with creates",
			resource: types.Resource{
				Type: "shell",
				Name: "create-file",
				Properties: map[string]interface{}{
					"command": "touch /tmp/test.txt",
					"creates": "/tmp/test.txt",
				},
			},
			wantErr: false,
		},
		{
			name: "valid shell with unless",
			resource: types.Resource{
				Type: "shell",
				Name: "conditional-command",
				Properties: map[string]interface{}{
					"command": "echo 'running'",
					"unless":  "test -f /tmp/skip",
				},
			},
			wantErr: false,
		},
		{
			name: "missing command",
			resource: types.Resource{
				Type: "shell",
				Name: "no-command",
				Properties: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "command not string",
			resource: types.Resource{
				Type: "shell",
				Name: "invalid-command",
				Properties: map[string]interface{}{
					"command": 123,
				},
			},
			wantErr: true,
		},
		{
			name: "creates not string",
			resource: types.Resource{
				Type: "shell",
				Name: "invalid-creates",
				Properties: map[string]interface{}{
					"command": "echo test",
					"creates": 123,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewShellProvider(nil)
			err := provider.Validate(&tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ShellProvider.Validate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ShellProvider.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestShellProvider_Read(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		mockCmds map[string]*ssh.ExecuteResult
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name: "command should run - no creates file",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello'",
					"creates": "/tmp/nonexistent.txt",
				},
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"test -e '/tmp/nonexistent.txt'": {
					Command:  "test -e '/tmp/nonexistent.txt'",
					ExitCode: 1,
				},
			},
			want: map[string]interface{}{
				"should_run": true,
			},
			wantErr: false,
		},
		{
			name: "command should not run - creates file exists",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello'",
					"creates": "/tmp/existing.txt",
				},
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"test -e '/tmp/existing.txt'": {
					Command:  "test -e '/tmp/existing.txt'",
					ExitCode: 0,
				},
			},
			want: map[string]interface{}{
				"should_run": false,
			},
			wantErr: false,
		},
		{
			name: "command should not run - unless condition true",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello'",
					"unless":  "test -f /tmp/skip",
				},
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"test -f /tmp/skip": {
					Command:  "test -f /tmp/skip",
					ExitCode: 0,
				},
			},
			want: map[string]interface{}{
				"should_run": false,
			},
			wantErr: false,
		},
		{
			name: "command should run - unless condition false",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello'",
					"unless":  "test -f /tmp/nonexistent",
				},
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"test -f /tmp/nonexistent": {
					Command:  "test -f /tmp/nonexistent",
					ExitCode: 1,
				},
			},
			want: map[string]interface{}{
				"should_run": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockSSHConnection{
				responses: tt.mockCmds,
			}
			
			provider := NewShellProvider(mockConn)
			ctx := context.Background()
			
			got, err := provider.Read(ctx, &tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ShellProvider.Read() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ShellProvider.Read() unexpected error = %v", err)
				return
			}
			
			if got["should_run"] != tt.want["should_run"] {
				t.Errorf("ShellProvider.Read() should_run = %v, want %v", got["should_run"], tt.want["should_run"])
			}
		})
	}
}

func TestShellProvider_Diff(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		current  map[string]interface{}
		want     types.DiffAction
	}{
		{
			name: "command should run",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello'",
				},
			},
			current: map[string]interface{}{
				"should_run": true,
			},
			want: types.ActionUpdate,
		},
		{
			name: "command should not run",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello'",
				},
			},
			current: map[string]interface{}{
				"should_run": false,
			},
			want: types.ActionNoop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewShellProvider(nil)
			ctx := context.Background()
			
			diff, err := provider.Diff(ctx, &tt.resource, tt.current)
			if err != nil {
				t.Errorf("ShellProvider.Diff() unexpected error = %v", err)
				return
			}
			
			if diff.Action != tt.want {
				t.Errorf("ShellProvider.Diff() Action = %v, want %v", diff.Action, tt.want)
			}
		})
	}
}

func TestShellProvider_Apply(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		diff     *types.ResourceDiff
		mockCmds map[string]*ssh.ExecuteResult
		wantErr  bool
	}{
		{
			name: "execute command",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello world'",
				},
			},
			diff: &types.ResourceDiff{
				ResourceID: "shell.test-command",
				Action:     types.ActionUpdate,
				Reason:     "command needs to be executed",
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"echo 'hello world'": {
					Command:  "echo 'hello world'",
					Stdout:   "hello world",
					ExitCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "command fails",
			resource: types.Resource{
				Type: "shell",
				Name: "failing-command",
				Properties: map[string]interface{}{
					"command": "exit 1",
				},
			},
			diff: &types.ResourceDiff{
				ResourceID: "shell.failing-command",
				Action:     types.ActionUpdate,
				Reason:     "command needs to be executed",
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"exit 1": {
					Command:  "exit 1",
					ExitCode: 1,
					Stderr:   "command failed",
				},
			},
			wantErr: true,
		},
		{
			name: "no action needed",
			resource: types.Resource{
				Type: "shell",
				Name: "test-command",
				Properties: map[string]interface{}{
					"command": "echo 'hello'",
				},
			},
			diff: &types.ResourceDiff{
				ResourceID: "shell.test-command",
				Action:     types.ActionNoop,
				Reason:     "command should not run",
			},
			mockCmds: map[string]*ssh.ExecuteResult{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockSSHConnection{
				responses: tt.mockCmds,
			}
			
			provider := NewShellProvider(mockConn)
			ctx := context.Background()
			
			err := provider.Apply(ctx, &tt.resource, tt.diff)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ShellProvider.Apply() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ShellProvider.Apply() unexpected error = %v", err)
				}
			}
		})
	}
}
