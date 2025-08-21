package inventory

import (
	"context"
	"testing"
)

func TestAzureInventoryProvider_Type(t *testing.T) {
	provider := NewAzureInventoryProvider("subscription-123", "resource-group", "eastus")
	
	if provider.Type() != "azure" {
		t.Errorf("Expected type 'azure', got '%s'", provider.Type())
	}
}

func TestAzureInventoryProvider_Validate(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		resourceGroup  string
		region         string
		wantErr        bool
	}{
		{
			name:           "valid config",
			subscriptionID: "subscription-123",
			resourceGroup:  "my-rg",
			region:         "eastus",
			wantErr:        false,
		},
		{
			name:           "missing subscription ID",
			subscriptionID: "",
			resourceGroup:  "my-rg",
			region:         "eastus",
			wantErr:        true,
		},
		{
			name:           "missing resource group",
			subscriptionID: "subscription-123",
			resourceGroup:  "",
			region:         "eastus",
			wantErr:        true,
		},
		{
			name:           "missing region",
			subscriptionID: "subscription-123",
			resourceGroup:  "my-rg",
			region:         "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewAzureInventoryProvider(tt.subscriptionID, tt.resourceGroup, tt.region)
			err := provider.Validate()
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestAzureInventoryProvider_Discover(t *testing.T) {
	provider := NewAzureInventoryProvider("subscription-123", "my-rg", "eastus")
	
	// Mock the Azure client for testing
	provider.SetMockMode(true)
	
	tests := []struct {
		name           string
		selector       string
		expectedCount  int
		expectedError  bool
	}{
		{
			name:          "discover all VMs",
			selector:      "",
			expectedCount: 2, // Mock will return 2 VMs
			expectedError: false,
		},
		{
			name:          "discover by environment tag",
			selector:      "Environment=production",
			expectedCount: 1, // Mock will return 1 production VM
			expectedError: false,
		},
		{
			name:          "discover by role tag",
			selector:      "Role=web",
			expectedCount: 1, // Mock will return 1 web VM
			expectedError: false,
		},
		{
			name:          "discover with multiple selectors",
			selector:      "Environment=production,Role=web",
			expectedCount: 1, // Mock will return 1 VM matching both
			expectedError: false,
		},
		{
			name:          "discover with no matches",
			selector:      "Environment=nonexistent",
			expectedCount: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			targets, err := provider.Discover(ctx, tt.selector)
			
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(targets) != tt.expectedCount {
				t.Errorf("Expected %d targets, got %d", tt.expectedCount, len(targets))
			}
			
			// Verify target structure
			for _, target := range targets {
				if target.Host == "" {
					t.Error("Expected target to have host")
				}
				
				if target.Port == 0 {
					t.Error("Expected target to have port")
				}
				
				if target.User == "" {
					t.Error("Expected target to have user")
				}
				
				if target.Labels == nil {
					t.Error("Expected target to have labels")
				}
			}
		})
	}
}

func TestAzureInventoryProvider_DiscoverWithInvalidSelector(t *testing.T) {
	provider := NewAzureInventoryProvider("subscription-123", "my-rg", "eastus")
	provider.SetMockMode(true)
	
	ctx := context.Background()
	_, err := provider.Discover(ctx, "invalid-selector-format")
	
	if err == nil {
		t.Error("Expected error for invalid selector format")
	}
}

func TestAzureInventoryProvider_DiscoverWithoutMockMode(t *testing.T) {
	provider := NewAzureInventoryProvider("subscription-123", "my-rg", "eastus")
	
	// Don't set mock mode - this should fail in test environment
	ctx := context.Background()
	_, err := provider.Discover(ctx, "")
	
	// We expect this to fail in test environment without Azure credentials
	if err == nil {
		t.Log("Azure discovery succeeded (unexpected in test environment without credentials)")
	} else {
		t.Logf("Azure discovery failed as expected in test environment: %v", err)
	}
}

func TestParseAzureSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "empty selector",
			selector: "",
			expected: map[string]string{},
			wantErr:  false,
		},
		{
			name:     "single tag",
			selector: "Environment=production",
			expected: map[string]string{"Environment": "production"},
			wantErr:  false,
		},
		{
			name:     "multiple tags",
			selector: "Environment=production,Role=web",
			expected: map[string]string{"Environment": "production", "Role": "web"},
			wantErr:  false,
		},
		{
			name:     "tags with spaces",
			selector: "Environment = production , Role = web",
			expected: map[string]string{"Environment": "production", "Role": "web"},
			wantErr:  false,
		},
		{
			name:     "invalid format - no equals",
			selector: "Environment",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid format - empty key",
			selector: "=production",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAzureSelector(tt.selector)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d", len(tt.expected), len(result))
				return
			}
			
			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected key '%s' not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Expected value '%s' for key '%s', got '%s'", expectedValue, key, actualValue)
				}
			}
		})
	}
}
