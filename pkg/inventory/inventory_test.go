package inventory

import (
	"testing"

	"github.com/ataiva-software/chisel/pkg/ssh"
)

func TestInventory_Validate(t *testing.T) {
	tests := []struct {
		name      string
		inventory Inventory
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid inventory with hosts",
			inventory: Inventory{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Inventory",
				Targets: map[string]TargetGroup{
					"webservers": {
						Hosts: []string{"web1.example.com", "web2.example.com"},
						Connection: ssh.ConnectionConfig{
							Host: "web1.example.com",
							User: "ubuntu",
							Port: 22,
							PrivateKeyPath: "~/.ssh/id_rsa",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid inventory with selector",
			inventory: Inventory{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Inventory",
				Targets: map[string]TargetGroup{
					"webservers": {
						Selector: "role=web,env=prod",
						Connection: ssh.ConnectionConfig{
							Host: "web1.example.com",
							User: "ubuntu",
							Port: 22,
							PrivateKeyPath: "~/.ssh/id_rsa",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing apiVersion",
			inventory: Inventory{
				Kind: "Inventory",
			},
			wantErr: true,
			errMsg:  "apiVersion is required",
		},
		{
			name: "invalid apiVersion",
			inventory: Inventory{
				APIVersion: "invalid/v1",
				Kind:       "Inventory",
			},
			wantErr: true,
			errMsg:  "apiVersion must be ataiva.com/chisel/v1",
		},
		{
			name: "missing kind",
			inventory: Inventory{
				APIVersion: "ataiva.com/chisel/v1",
			},
			wantErr: true,
			errMsg:  "kind is required",
		},
		{
			name: "invalid kind",
			inventory: Inventory{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "InvalidKind",
			},
			wantErr: true,
			errMsg:  "kind must be Inventory",
		},
		{
			name: "empty targets",
			inventory: Inventory{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Inventory",
				Targets:    map[string]TargetGroup{},
			},
			wantErr: true,
			errMsg:  "at least one target group is required",
		},
		{
			name: "target group with both hosts and selector",
			inventory: Inventory{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Inventory",
				Targets: map[string]TargetGroup{
					"webservers": {
						Hosts:    []string{"web1.example.com"},
						Selector: "role=web",
						Connection: ssh.ConnectionConfig{
							Host: "web1.example.com",
							User: "ubuntu",
							Port: 22,
							PrivateKeyPath: "~/.ssh/id_rsa",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "target group 'webservers': cannot specify both hosts and selector",
		},
		{
			name: "target group with neither hosts nor selector",
			inventory: Inventory{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Inventory",
				Targets: map[string]TargetGroup{
					"webservers": {
						Connection: ssh.ConnectionConfig{
							Host: "web1.example.com",
							User: "ubuntu",
							Port: 22,
							PrivateKeyPath: "~/.ssh/id_rsa",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "target group 'webservers': must specify either hosts or selector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inventory.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Inventory.Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Inventory.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Inventory.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestLoadInventoryFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "load valid inventory",
			filename: "testdata/valid-inventory.yaml",
			wantErr:  false,
		},
		{
			name:     "file not found",
			filename: "testdata/nonexistent.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadInventoryFromFile(tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadInventoryFromFile() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("LoadInventoryFromFile() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTargetGroup_GetHosts(t *testing.T) {
	tests := []struct {
		name        string
		targetGroup TargetGroup
		want        []string
		wantErr     bool
	}{
		{
			name: "static hosts",
			targetGroup: TargetGroup{
				Hosts: []string{"web1.example.com", "web2.example.com"},
			},
			want:    []string{"web1.example.com", "web2.example.com"},
			wantErr: false,
		},
		{
			name: "selector (not implemented yet)",
			targetGroup: TargetGroup{
				Selector: "role=web,env=prod",
			},
			want:    []string{},
			wantErr: false, // For now, selectors return empty list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.targetGroup.GetHosts()
			if tt.wantErr {
				if err == nil {
					t.Errorf("TargetGroup.GetHosts() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("TargetGroup.GetHosts() unexpected error = %v", err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("TargetGroup.GetHosts() = %v, want %v", got, tt.want)
				return
			}
			for i, host := range got {
				if host != tt.want[i] {
					t.Errorf("TargetGroup.GetHosts()[%d] = %v, want %v", i, host, tt.want[i])
				}
			}
		})
	}
}
