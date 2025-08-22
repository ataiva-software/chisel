package providers

import (
	"context"
	"testing"

	"github.com/ataiva-software/forge/pkg/ssh"
	"github.com/ataiva-software/forge/pkg/types"
)

func TestPkgProvider_Type(t *testing.T) {
	provider := NewPkgProvider(nil)
	if provider.Type() != "pkg" {
		t.Errorf("Expected type 'pkg', got '%s'", provider.Type())
	}
}

func TestPkgProvider_Validate(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		wantErr  bool
	}{
		{
			name: "valid package resource",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "present",
				},
			},
			wantErr: false,
		},
		{
			name: "valid package resource with version",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state":   "present",
					"version": "1.18.0",
				},
			},
			wantErr: false,
		},
		{
			name: "missing state",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "invalid state",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "state not string",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": 123,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewPkgProvider(nil)
			err := provider.Validate(&tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("PkgProvider.Validate() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("PkgProvider.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestPkgProvider_Read(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		mockCmd  string
		mockOut  string
		mockExit int
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name: "package installed",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "present",
				},
			},
			mockCmd:  "dpkg -l 'nginx' 2>/dev/null | grep '^ii' | wc -l",
			mockOut:  "1",
			mockExit: 0,
			want: map[string]interface{}{
				"state": "present",
			},
			wantErr: false,
		},
		{
			name: "package not installed",
			resource: types.Resource{
				Type: "pkg",
				Name: "nonexistent",
				Properties: map[string]interface{}{
					"state": "present",
				},
			},
			mockCmd:  "dpkg -l 'nonexistent' 2>/dev/null | grep '^ii' | wc -l",
			mockOut:  "0",
			mockExit: 0,
			want: map[string]interface{}{
				"state": "absent",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockSSHConnection{
				responses: map[string]*ssh.ExecuteResult{
					tt.mockCmd: {
						Command:  tt.mockCmd,
						Stdout:   tt.mockOut,
						ExitCode: tt.mockExit,
					},
				},
			}
			
			provider := NewPkgProvider(mockConn)
			ctx := context.Background()
			
			got, err := provider.Read(ctx, &tt.resource)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("PkgProvider.Read() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("PkgProvider.Read() unexpected error = %v", err)
				return
			}
			
			if got["state"] != tt.want["state"] {
				t.Errorf("PkgProvider.Read() state = %v, want %v", got["state"], tt.want["state"])
			}
		})
	}
}

func TestPkgProvider_Diff(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		current  map[string]interface{}
		want     types.DiffAction
	}{
		{
			name: "install package",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "present",
				},
			},
			current: map[string]interface{}{
				"state": "absent",
			},
			want: types.ActionCreate,
		},
		{
			name: "remove package",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "absent",
				},
			},
			current: map[string]interface{}{
				"state": "present",
			},
			want: types.ActionDelete,
		},
		{
			name: "package already in desired state",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "present",
				},
			},
			current: map[string]interface{}{
				"state": "present",
			},
			want: types.ActionNoop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewPkgProvider(nil)
			ctx := context.Background()
			
			diff, err := provider.Diff(ctx, &tt.resource, tt.current)
			if err != nil {
				t.Errorf("PkgProvider.Diff() unexpected error = %v", err)
				return
			}
			
			if diff.Action != tt.want {
				t.Errorf("PkgProvider.Diff() Action = %v, want %v", diff.Action, tt.want)
			}
		})
	}
}

func TestPkgProvider_Apply(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Resource
		diff     *types.ResourceDiff
		mockCmd  string
		mockExit int
		wantErr  bool
	}{
		{
			name: "install package",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "present",
				},
			},
			diff: &types.ResourceDiff{
				ResourceID: "pkg.nginx",
				Action:     types.ActionCreate,
				Reason:     "package needs to be installed",
			},
			mockCmd:  "apt-get update && apt-get install -y 'nginx'",
			mockExit: 0,
			wantErr:  false,
		},
		{
			name: "remove package",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "absent",
				},
			},
			diff: &types.ResourceDiff{
				ResourceID: "pkg.nginx",
				Action:     types.ActionDelete,
				Reason:     "package needs to be removed",
			},
			mockCmd:  "apt-get remove -y 'nginx'",
			mockExit: 0,
			wantErr:  false,
		},
		{
			name: "no action needed",
			resource: types.Resource{
				Type: "pkg",
				Name: "nginx",
				Properties: map[string]interface{}{
					"state": "present",
				},
			},
			diff: &types.ResourceDiff{
				ResourceID: "pkg.nginx",
				Action:     types.ActionNoop,
				Reason:     "package already in desired state",
			},
			mockCmd:  "",
			mockExit: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockSSHConnection{
				responses: make(map[string]*ssh.ExecuteResult),
			}
			
			if tt.mockCmd != "" {
				mockConn.responses[tt.mockCmd] = &ssh.ExecuteResult{
					Command:  tt.mockCmd,
					ExitCode: tt.mockExit,
				}
			}
			
			provider := NewPkgProvider(mockConn)
			ctx := context.Background()
			
			err := provider.Apply(ctx, &tt.resource, tt.diff)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("PkgProvider.Apply() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("PkgProvider.Apply() unexpected error = %v", err)
				}
			}
		})
	}
}
