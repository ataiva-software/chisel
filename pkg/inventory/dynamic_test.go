package inventory

import (
	"context"
	"reflect"
	"testing"

	"github.com/ataiva-software/chisel/pkg/types"
)

// MockDynamicInventory for testing
type MockDynamicInventory struct {
	inventoryType string
	targets       []types.Target
	validateErr   error
}

func (m *MockDynamicInventory) Type() string {
	return m.inventoryType
}

func (m *MockDynamicInventory) Validate() error {
	return m.validateErr
}

func (m *MockDynamicInventory) Discover(ctx context.Context, selector string) ([]types.Target, error) {
	selectorMap, err := ParseSelector(selector)
	if err != nil {
		return nil, err
	}
	
	var matchingTargets []types.Target
	for _, target := range m.targets {
		if MatchesSelector(target, selectorMap) {
			matchingTargets = append(matchingTargets, target)
		}
	}
	
	// Return empty slice instead of nil for no matches
	if matchingTargets == nil {
		return []types.Target{}, nil
	}
	
	return matchingTargets, nil
}

func TestInventoryRegistry_Register(t *testing.T) {
	tests := []struct {
		name        string
		provider    DynamicInventory
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid provider",
			provider: &MockDynamicInventory{
				inventoryType: "test",
				validateErr:   nil,
			},
			wantErr: false,
		},
		{
			name:        "nil provider",
			provider:    nil,
			wantErr:     true,
			expectedErr: "provider cannot be nil",
		},
		{
			name: "empty type",
			provider: &MockDynamicInventory{
				inventoryType: "",
				validateErr:   nil,
			},
			wantErr:     true,
			expectedErr: "provider type cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewInventoryRegistry()
			err := registry.Register(tt.provider)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("InventoryRegistry.Register() expected error but got none")
				} else if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("InventoryRegistry.Register() error = %v, want %v", err.Error(), tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("InventoryRegistry.Register() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestInventoryRegistry_Discover(t *testing.T) {
	targets := []types.Target{
		{
			Host: "web1.example.com",
			Labels: map[string]string{
				"role": "web",
				"env":  "prod",
			},
		},
		{
			Host: "web2.example.com",
			Labels: map[string]string{
				"role": "web",
				"env":  "staging",
			},
		},
		{
			Host: "db1.example.com",
			Labels: map[string]string{
				"role": "database",
				"env":  "prod",
			},
		},
	}

	registry := NewInventoryRegistry()
	mockProvider := &MockDynamicInventory{
		inventoryType: "test",
		targets:       targets,
	}
	
	err := registry.Register(mockProvider)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	tests := []struct {
		name         string
		providerType string
		selector     string
		want         []types.Target
		wantErr      bool
	}{
		{
			name:         "select web servers in prod",
			providerType: "test",
			selector:     "role=web,env=prod",
			want:         []types.Target{targets[0]},
			wantErr:      false,
		},
		{
			name:         "select all web servers",
			providerType: "test",
			selector:     "role=web",
			want:         []types.Target{targets[0], targets[1]},
			wantErr:      false,
		},
		{
			name:         "select prod environment",
			providerType: "test",
			selector:     "env=prod",
			want:         []types.Target{targets[0], targets[2]},
			wantErr:      false,
		},
		{
			name:         "no matches",
			providerType: "test",
			selector:     "role=nonexistent",
			want:         []types.Target{},
			wantErr:      false,
		},
		{
			name:         "unknown provider",
			providerType: "unknown",
			selector:     "role=web",
			want:         nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := registry.Discover(ctx, tt.providerType, tt.selector)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("InventoryRegistry.Discover() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("InventoryRegistry.Discover() unexpected error = %v", err)
				return
			}
			
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InventoryRegistry.Discover() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		want     map[string]string
		wantErr  bool
	}{
		{
			name:     "empty selector",
			selector: "",
			want:     map[string]string{},
			wantErr:  false,
		},
		{
			name:     "single key-value pair",
			selector: "role=web",
			want:     map[string]string{"role": "web"},
			wantErr:  false,
		},
		{
			name:     "multiple key-value pairs",
			selector: "role=web,env=prod",
			want:     map[string]string{"role": "web", "env": "prod"},
			wantErr:  false,
		},
		{
			name:     "with spaces",
			selector: " role = web , env = prod ",
			want:     map[string]string{"role": "web", "env": "prod"},
			wantErr:  false,
		},
		{
			name:     "invalid format - no equals",
			selector: "role",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid format - empty key",
			selector: "=web",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "value with equals sign",
			selector: "config=key=value",
			want:     map[string]string{"config": "key=value"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSelector(tt.selector)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSelector() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseSelector() unexpected error = %v", err)
				return
			}
			
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesSelector(t *testing.T) {
	target := types.Target{
		Host: "web1.example.com",
		Labels: map[string]string{
			"role": "web",
			"env":  "prod",
			"zone": "us-east-1a",
		},
	}

	tests := []struct {
		name     string
		selector map[string]string
		want     bool
	}{
		{
			name:     "empty selector matches all",
			selector: map[string]string{},
			want:     true,
		},
		{
			name:     "single matching label",
			selector: map[string]string{"role": "web"},
			want:     true,
		},
		{
			name:     "multiple matching labels",
			selector: map[string]string{"role": "web", "env": "prod"},
			want:     true,
		},
		{
			name:     "non-matching label value",
			selector: map[string]string{"role": "database"},
			want:     false,
		},
		{
			name:     "non-existent label",
			selector: map[string]string{"nonexistent": "value"},
			want:     false,
		},
		{
			name:     "partial match",
			selector: map[string]string{"role": "web", "env": "staging"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesSelector(target, tt.selector)
			if got != tt.want {
				t.Errorf("MatchesSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}
