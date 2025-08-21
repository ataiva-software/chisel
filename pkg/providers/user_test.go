package providers

import (
	"context"
	"testing"

	"github.com/ataiva-software/chisel/pkg/ssh"
	"github.com/ataiva-software/chisel/pkg/types"
)

func TestUserProvider_Type(t *testing.T) {
	provider := NewUserProvider(nil)
	if provider.Type() != "user" {
		t.Errorf("Expected type 'user', got '%s'", provider.Type())
	}
}

func TestUserProvider_Validate(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		wantErr  bool
	}{
		{
			name: "valid user resource",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
			},
			wantErr: false,
		},
		{
			name: "valid user with all properties",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
				Properties: map[string]interface{}{
					"uid":    1001,
					"gid":    1001,
					"home":   "/home/testuser",
					"shell":  "/bin/bash",
					"groups": []string{"sudo", "docker"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid state",
			resource: types.Resource{
				Type: "user",
				Name: "testuser",
				Properties: map[string]interface{}{
					"state": "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid uid type",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
				Properties: map[string]interface{}{
					"uid": "not-a-number",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid groups type",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
				Properties: map[string]interface{}{
					"groups": "not-an-array",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewUserProvider(nil)
			err := provider.Validate(&tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("UserProvider.Validate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("UserProvider.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestUserProvider_Read(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		mockCmds map[string]*ssh.ExecuteResult
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name: "user exists",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"id -u 'testuser' 2>/dev/null": {
					Command:  "id -u 'testuser' 2>/dev/null",
					Stdout:   "1001",
					ExitCode: 0,
				},
				"getent passwd 'testuser'": {
					Command:  "getent passwd 'testuser'",
					Stdout:   "testuser:x:1001:1001:Test User:/home/testuser:/bin/bash",
					ExitCode: 0,
				},
				"groups 'testuser'": {
					Command:  "groups 'testuser'",
					Stdout:   "testuser : testuser sudo docker",
					ExitCode: 0,
				},
			},
			want: map[string]interface{}{
				"state":  "present",
				"uid":    1001,
				"gid":    1001,
				"home":   "/home/testuser",
				"shell":  "/bin/bash",
				"groups": []string{"testuser", "sudo", "docker"},
			},
			wantErr: false,
		},
		{
			name: "user does not exist",
			resource: types.Resource{
				Type:  "user",
				Name:  "nonexistent",
				State: types.StatePresent,
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"id -u 'nonexistent' 2>/dev/null": {
					Command:  "id -u 'nonexistent' 2>/dev/null",
					Stdout:   "",
					ExitCode: 1,
				},
			},
			want: map[string]interface{}{
				"state": "absent",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockSSHConnection{
				responses: tt.mockCmds,
			}
			
			provider := NewUserProvider(mockConn)
			ctx := context.Background()
			
			got, err := provider.Read(ctx, &tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("UserProvider.Read() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("UserProvider.Read() unexpected error = %v", err)
				return
			}
			
			if got["state"] != tt.want["state"] {
				t.Errorf("UserProvider.Read() state = %v, want %v", got["state"], tt.want["state"])
			}
			
			if tt.want["uid"] != nil && got["uid"] != tt.want["uid"] {
				t.Errorf("UserProvider.Read() uid = %v, want %v", got["uid"], tt.want["uid"])
			}
		})
	}
}

func TestUserProvider_Diff(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		current  map[string]interface{}
		want     types.DiffAction
	}{
		{
			name: "create user",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
			},
			current: map[string]interface{}{
				"state": "absent",
			},
			want: types.ActionCreate,
		},
		{
			name: "delete user",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StateAbsent,
			},
			current: map[string]interface{}{
				"state": "present",
				"uid":   1001,
			},
			want: types.ActionDelete,
		},
		{
			name: "user already in desired state",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
			},
			current: map[string]interface{}{
				"state": "present",
				"uid":   1001,
			},
			want: types.ActionNoop,
		},
		{
			name: "update user properties",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
				Properties: map[string]interface{}{
					"shell": "/bin/zsh",
				},
			},
			current: map[string]interface{}{
				"state": "present",
				"uid":   1001,
				"shell": "/bin/bash",
			},
			want: types.ActionUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewUserProvider(nil)
			ctx := context.Background()
			
			diff, err := provider.Diff(ctx, &tt.resource, tt.current)
			if err != nil {
				t.Errorf("UserProvider.Diff() unexpected error = %v", err)
				return
			}
			
			if diff.Action != tt.want {
				t.Errorf("UserProvider.Diff() Action = %v, want %v", diff.Action, tt.want)
			}
		})
	}
}

func TestUserProvider_Apply(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		diff     *types.ResourceDiff
		mockCmds map[string]*ssh.ExecuteResult
		wantErr  bool
	}{
		{
			name: "create user",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
			},
			diff: &types.ResourceDiff{
				ResourceID: "user.testuser",
				Action:     types.ActionCreate,
				Reason:     "user needs to be created",
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"useradd 'testuser'": {
					Command:  "useradd 'testuser'",
					ExitCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "delete user",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StateAbsent,
			},
			diff: &types.ResourceDiff{
				ResourceID: "user.testuser",
				Action:     types.ActionDelete,
				Reason:     "user needs to be removed",
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"userdel 'testuser'": {
					Command:  "userdel 'testuser'",
					ExitCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "no action needed",
			resource: types.Resource{
				Type:  "user",
				Name:  "testuser",
				State: types.StatePresent,
			},
			diff: &types.ResourceDiff{
				ResourceID: "user.testuser",
				Action:     types.ActionNoop,
				Reason:     "user already in desired state",
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
			
			provider := NewUserProvider(mockConn)
			ctx := context.Background()
			
			err := provider.Apply(ctx, &tt.resource, tt.diff)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("UserProvider.Apply() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("UserProvider.Apply() unexpected error = %v", err)
				}
			}
		})
	}
}
