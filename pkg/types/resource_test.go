package types

import (
	"context"
	"testing"
)

func TestResource_ResourceID(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
		expected string
	}{
		{
			name: "basic resource ID",
			resource: Resource{
				Type: "file",
				Name: "/etc/nginx/nginx.conf",
			},
			expected: "file./etc/nginx/nginx.conf",
		},
		{
			name: "service resource ID",
			resource: Resource{
				Type: "service",
				Name: "nginx",
			},
			expected: "service.nginx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resource.ResourceID(); got != tt.expected {
				t.Errorf("Resource.ResourceID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResource_Validate(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid resource",
			resource: Resource{
				Type: "file",
				Name: "/etc/nginx/nginx.conf",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			resource: Resource{
				Name: "/etc/nginx/nginx.conf",
			},
			wantErr: true,
			errMsg:  "resource type cannot be empty",
		},
		{
			name: "missing name",
			resource: Resource{
				Type: "file",
			},
			wantErr: true,
			errMsg:  "resource name cannot be empty",
		},
		{
			name: "empty type",
			resource: Resource{
				Type: "",
				Name: "test",
			},
			wantErr: true,
			errMsg:  "resource type cannot be empty",
		},
		{
			name: "empty name",
			resource: Resource{
				Type: "file",
				Name: "",
			},
			wantErr: true,
			errMsg:  "resource name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resource.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Resource.Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Resource.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Resource.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestProviderRegistry_Register(t *testing.T) {
	registry := NewProviderRegistry()
	mockProvider := &MockProvider{resourceType: "test"}

	tests := []struct {
		name     string
		provider Provider
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "register valid provider",
			provider: mockProvider,
			wantErr:  false,
		},
		{
			name:     "register nil provider",
			provider: nil,
			wantErr:  true,
			errMsg:   "provider cannot be nil",
		},
		{
			name:     "register provider with empty type",
			provider: &MockProvider{resourceType: ""},
			wantErr:  true,
			errMsg:   "provider type cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.provider)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ProviderRegistry.Register() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ProviderRegistry.Register() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ProviderRegistry.Register() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestProviderRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewProviderRegistry()
	provider1 := &MockProvider{resourceType: "test"}
	provider2 := &MockProvider{resourceType: "test"}

	// Register first provider
	err := registry.Register(provider1)
	if err != nil {
		t.Fatalf("Failed to register first provider: %v", err)
	}

	// Try to register duplicate
	err = registry.Register(provider2)
	if err == nil {
		t.Error("Expected error when registering duplicate provider")
	}

	expectedErr := "provider for type test already registered"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestProviderRegistry_Get(t *testing.T) {
	registry := NewProviderRegistry()
	mockProvider := &MockProvider{resourceType: "test"}
	
	// Register provider
	err := registry.Register(mockProvider)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	tests := []struct {
		name         string
		resourceType string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "get existing provider",
			resourceType: "test",
			wantErr:      false,
		},
		{
			name:         "get non-existent provider",
			resourceType: "nonexistent",
			wantErr:      true,
			errMsg:       "no provider registered for resource type: nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := registry.Get(tt.resourceType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ProviderRegistry.Get() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ProviderRegistry.Get() error = %v, want %v", err.Error(), tt.errMsg)
				}
				if provider != nil {
					t.Errorf("ProviderRegistry.Get() expected nil provider but got %v", provider)
				}
			} else {
				if err != nil {
					t.Errorf("ProviderRegistry.Get() unexpected error = %v", err)
				}
				if provider == nil {
					t.Error("ProviderRegistry.Get() expected provider but got nil")
				}
			}
		})
	}
}

func TestProviderRegistry_Types(t *testing.T) {
	registry := NewProviderRegistry()
	
	// Empty registry
	types := registry.Types()
	if len(types) != 0 {
		t.Errorf("Expected empty types slice, got %v", types)
	}
	
	// Add providers
	provider1 := &MockProvider{resourceType: "file"}
	provider2 := &MockProvider{resourceType: "service"}
	
	registry.Register(provider1)
	registry.Register(provider2)
	
	types = registry.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(types))
	}
	
	// Check that both types are present (order doesn't matter)
	typeMap := make(map[string]bool)
	for _, t := range types {
		typeMap[t] = true
	}
	
	if !typeMap["file"] {
		t.Error("Expected 'file' type to be present")
	}
	if !typeMap["service"] {
		t.Error("Expected 'service' type to be present")
	}
}

// MockProvider is a test implementation of the Provider interface
type MockProvider struct {
	resourceType string
}

func (m *MockProvider) Type() string {
	return m.resourceType
}

func (m *MockProvider) Validate(resource *Resource) error {
	return nil
}

func (m *MockProvider) Read(ctx context.Context, resource *Resource) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func (m *MockProvider) Diff(ctx context.Context, resource *Resource, current map[string]interface{}) (*ResourceDiff, error) {
	return &ResourceDiff{
		ResourceID: resource.ResourceID(),
		Action:     ActionNoop,
	}, nil
}

func (m *MockProvider) Apply(ctx context.Context, resource *Resource, diff *ResourceDiff) error {
	return nil
}
