package providers

import (
	"context"
	"testing"

	"github.com/ataiva-software/forge/pkg/ssh"
	"github.com/ataiva-software/forge/pkg/types"
)

func TestServiceProvider_Type(t *testing.T) {
	provider := NewServiceProvider(nil)
	if provider.Type() != "service" {
		t.Errorf("Expected type 'service', got '%s'", provider.Type())
	}
}

func TestServiceProvider_Validate(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		wantErr  bool
	}{
		{
			name: "valid service resource",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
			},
			wantErr: false,
		},
		{
			name: "valid service resource with enabled",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
				Properties: map[string]interface{}{
					"enabled": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid stopped service",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateStopped,
			},
			wantErr: false,
		},
		{
			name: "invalid state",
			resource: types.Resource{
				Type: "service",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "enabled not boolean",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
				Properties: map[string]interface{}{
					"enabled": "yes",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewServiceProvider(nil)
			err := provider.Validate(&tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ServiceProvider.Validate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ServiceProvider.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestServiceProvider_Read(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		mockCmds map[string]*ssh.ExecuteResult
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name: "service running and enabled",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"systemctl is-active 'nginx'": {
					Command:  "systemctl is-active 'nginx'",
					Stdout:   "active",
					ExitCode: 0,
				},
				"systemctl is-enabled 'nginx'": {
					Command:  "systemctl is-enabled 'nginx'",
					Stdout:   "enabled",
					ExitCode: 0,
				},
			},
			want: map[string]interface{}{
				"state":   "running",
				"enabled": true,
			},
			wantErr: false,
		},
		{
			name: "service stopped and disabled",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"systemctl is-active 'nginx'": {
					Command:  "systemctl is-active 'nginx'",
					Stdout:   "inactive",
					ExitCode: 3,
				},
				"systemctl is-enabled 'nginx'": {
					Command:  "systemctl is-enabled 'nginx'",
					Stdout:   "disabled",
					ExitCode: 1,
				},
			},
			want: map[string]interface{}{
				"state":   "stopped",
				"enabled": false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockSSHConnection{
				responses: tt.mockCmds,
			}
			
			provider := NewServiceProvider(mockConn)
			ctx := context.Background()
			
			got, err := provider.Read(ctx, &tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ServiceProvider.Read() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ServiceProvider.Read() unexpected error = %v", err)
				return
			}
			
			if got["state"] != tt.want["state"] {
				t.Errorf("ServiceProvider.Read() state = %v, want %v", got["state"], tt.want["state"])
			}
			
			if got["enabled"] != tt.want["enabled"] {
				t.Errorf("ServiceProvider.Read() enabled = %v, want %v", got["enabled"], tt.want["enabled"])
			}
		})
	}
}

func TestServiceProvider_Diff(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		current  map[string]interface{}
		want     types.DiffAction
	}{
		{
			name: "start service",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
			},
			current: map[string]interface{}{
				"state":   "stopped",
				"enabled": false,
			},
			want: types.ActionUpdate,
		},
		{
			name: "stop service",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateStopped,
			},
			current: map[string]interface{}{
				"state":   "running",
				"enabled": true,
			},
			want: types.ActionUpdate,
		},
		{
			name: "service already in desired state",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
			},
			current: map[string]interface{}{
				"state":   "running",
				"enabled": true,
			},
			want: types.ActionNoop,
		},
		{
			name: "enable service",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
				Properties: map[string]interface{}{
					"enabled": true,
				},
			},
			current: map[string]interface{}{
				"state":   "running",
				"enabled": false,
			},
			want: types.ActionUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewServiceProvider(nil)
			ctx := context.Background()
			
			diff, err := provider.Diff(ctx, &tt.resource, tt.current)
			if err != nil {
				t.Errorf("ServiceProvider.Diff() unexpected error = %v", err)
				return
			}
			
			if diff.Action != tt.want {
				t.Errorf("ServiceProvider.Diff() Action = %v, want %v", diff.Action, tt.want)
			}
		})
	}
}

func TestServiceProvider_Apply(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		diff     *types.ResourceDiff
		mockCmds map[string]*ssh.ExecuteResult
		wantErr  bool
	}{
		{
			name: "start service",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
			},
			diff: &types.ResourceDiff{
				ResourceID: "service.nginx",
				Action:     types.ActionUpdate,
				Reason:     "service needs to be started",
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"systemctl start 'nginx'": {
					Command:  "systemctl start 'nginx'",
					ExitCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "stop service",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateStopped,
			},
			diff: &types.ResourceDiff{
				ResourceID: "service.nginx",
				Action:     types.ActionUpdate,
				Reason:     "service needs to be stopped",
			},
			mockCmds: map[string]*ssh.ExecuteResult{
				"systemctl stop 'nginx'": {
					Command:  "systemctl stop 'nginx'",
					ExitCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "no action needed",
			resource: types.Resource{
				Type:  "service",
				Name:  "nginx",
				State: types.StateRunning,
			},
			diff: &types.ResourceDiff{
				ResourceID: "service.nginx",
				Action:     types.ActionNoop,
				Reason:     "service already in desired state",
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
			
			provider := NewServiceProvider(mockConn)
			ctx := context.Background()
			
			err := provider.Apply(ctx, &tt.resource, tt.diff)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ServiceProvider.Apply() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ServiceProvider.Apply() unexpected error = %v", err)
				}
			}
		})
	}
}
