package core

import (
	"testing"

	"github.com/ataiva-software/forge/pkg/types"
)

func TestModule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		module  Module
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid module",
			module: Module{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Module",
				Metadata: ModuleMetadata{
					Name:    "test-module",
					Version: "1.0.0",
				},
				Spec: ModuleSpec{
					Resources: []types.Resource{
						{
							Type: "file",
							Name: "test-file",
							Properties: map[string]interface{}{
								"path":    "/tmp/test",
								"content": "test content",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing apiVersion",
			module: Module{
				Kind: "Module",
				Metadata: ModuleMetadata{
					Name:    "test-module",
					Version: "1.0.0",
				},
			},
			wantErr: true,
			errMsg:  "apiVersion is required",
		},
		{
			name: "invalid apiVersion",
			module: Module{
				APIVersion: "invalid/v1",
				Kind:       "Module",
				Metadata: ModuleMetadata{
					Name:    "test-module",
					Version: "1.0.0",
				},
			},
			wantErr: true,
			errMsg:  "apiVersion must be ataiva.com/chisel/v1",
		},
		{
			name: "missing kind",
			module: Module{
				APIVersion: "ataiva.com/chisel/v1",
				Metadata: ModuleMetadata{
					Name:    "test-module",
					Version: "1.0.0",
				},
			},
			wantErr: true,
			errMsg:  "kind is required",
		},
		{
			name: "invalid kind",
			module: Module{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "InvalidKind",
				Metadata: ModuleMetadata{
					Name:    "test-module",
					Version: "1.0.0",
				},
			},
			wantErr: true,
			errMsg:  "kind must be Module",
		},
		{
			name: "missing module name",
			module: Module{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Module",
				Metadata: ModuleMetadata{
					Version: "1.0.0",
				},
			},
			wantErr: true,
			errMsg:  "metadata.name is required",
		},
		{
			name: "missing module version",
			module: Module{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Module",
				Metadata: ModuleMetadata{
					Name: "test-module",
				},
			},
			wantErr: true,
			errMsg:  "metadata.version is required",
		},
		{
			name: "invalid version format",
			module: Module{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Module",
				Metadata: ModuleMetadata{
					Name:    "test-module",
					Version: "invalid-version",
				},
			},
			wantErr: true,
			errMsg:  "metadata.version must be valid semver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.module.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Module.Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Module.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Module.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestLoadModuleFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     *Module
		wantErr  bool
	}{
		{
			name:     "load valid module",
			filename: "testdata/valid-module.yaml",
			want: &Module{
				APIVersion: "ataiva.com/chisel/v1",
				Kind:       "Module",
				Metadata: ModuleMetadata{
					Name:        "test-module",
					Version:     "1.0.0",
					Description: "A test module",
				},
				Spec: ModuleSpec{
					Resources: []types.Resource{
						{
							Type: "file",
							Name: "test-file",
							Properties: map[string]interface{}{
								"path":    "/tmp/test",
								"content": "test content",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "file not found",
			filename: "testdata/nonexistent.yaml",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadModuleFromFile(tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Errorf("LoadModuleFromFile() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("LoadModuleFromFile() unexpected error = %v", err)
				return
			}
			if got.APIVersion != tt.want.APIVersion {
				t.Errorf("LoadModuleFromFile() APIVersion = %v, want %v", got.APIVersion, tt.want.APIVersion)
			}
			if got.Kind != tt.want.Kind {
				t.Errorf("LoadModuleFromFile() Kind = %v, want %v", got.Kind, tt.want.Kind)
			}
			if got.Metadata.Name != tt.want.Metadata.Name {
				t.Errorf("LoadModuleFromFile() Name = %v, want %v", got.Metadata.Name, tt.want.Metadata.Name)
			}
		})
	}
}
