package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/ataiva-software/chisel/pkg/types"
)

// AzureInventoryProvider discovers targets from Azure VMs
type AzureInventoryProvider struct {
	subscriptionID string
	resourceGroup  string
	region         string
	mockMode       bool
}

// NewAzureInventoryProvider creates a new Azure inventory provider
func NewAzureInventoryProvider(subscriptionID, resourceGroup, region string) *AzureInventoryProvider {
	return &AzureInventoryProvider{
		subscriptionID: subscriptionID,
		resourceGroup:  resourceGroup,
		region:         region,
		mockMode:       false,
	}
}

// Type returns the provider type
func (a *AzureInventoryProvider) Type() string {
	return "azure"
}

// Validate validates the provider configuration
func (a *AzureInventoryProvider) Validate() error {
	if a.subscriptionID == "" {
		return fmt.Errorf("Azure subscription ID is required")
	}
	
	if a.resourceGroup == "" {
		return fmt.Errorf("Azure resource group is required")
	}
	
	if a.region == "" {
		return fmt.Errorf("Azure region is required")
	}
	
	return nil
}

// SetMockMode enables mock mode for testing
func (a *AzureInventoryProvider) SetMockMode(enabled bool) {
	a.mockMode = enabled
}

// Discover discovers Azure VMs based on the selector
func (a *AzureInventoryProvider) Discover(ctx context.Context, selector string) ([]types.Target, error) {
	if a.mockMode {
		return a.discoverMock(ctx, selector)
	}
	
	return a.discoverReal(ctx, selector)
}

// discoverMock provides mock data for testing
func (a *AzureInventoryProvider) discoverMock(ctx context.Context, selector string) ([]types.Target, error) {
	// Parse selector
	selectorMap, err := parseAzureSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid selector: %w", err)
	}
	
	// Mock Azure VMs
	mockVMs := []mockAzureVM{
		{
			Name:              "web-vm-1",
			ResourceGroup:     a.resourceGroup,
			Location:          a.region,
			PublicIPAddress:   "20.1.2.3",
			PrivateIPAddress:  "10.0.1.10",
			PowerState:        "VM running",
			Tags: map[string]string{
				"Environment": "production",
				"Role":        "web",
				"Team":        "platform",
			},
		},
		{
			Name:              "db-vm-1",
			ResourceGroup:     a.resourceGroup,
			Location:          a.region,
			PublicIPAddress:   "20.1.2.4",
			PrivateIPAddress:  "10.0.1.20",
			PowerState:        "VM running",
			Tags: map[string]string{
				"Environment": "staging",
				"Role":        "database",
				"Team":        "data",
			},
		},
	}
	
	// Filter VMs based on selector
	var matchingVMs []mockAzureVM
	for _, vm := range mockVMs {
		if a.vmMatchesSelector(vm, selectorMap) {
			matchingVMs = append(matchingVMs, vm)
		}
	}
	
	// Convert to targets
	var targets []types.Target
	for _, vm := range matchingVMs {
		target := types.Target{
			Host: vm.PublicIPAddress,
			Port: 22, // Default SSH port
			User: "azureuser", // Default Azure user
			Labels: make(map[string]string),
		}
		
		// Copy tags as labels
		for key, value := range vm.Tags {
			target.Labels[key] = value
		}
		
		// Add Azure-specific labels
		target.Labels["azure:vm-name"] = vm.Name
		target.Labels["azure:resource-group"] = vm.ResourceGroup
		target.Labels["azure:location"] = vm.Location
		target.Labels["azure:private-ip"] = vm.PrivateIPAddress
		target.Labels["azure:power-state"] = vm.PowerState
		target.Labels["azure:subscription-id"] = a.subscriptionID
		
		targets = append(targets, target)
	}
	
	return targets, nil
}

// discoverReal discovers real Azure VMs (would use Azure SDK)
func (a *AzureInventoryProvider) discoverReal(ctx context.Context, selector string) ([]types.Target, error) {
	// In a real implementation, this would use the Azure SDK:
	// - Create Azure client with subscription and credentials
	// - Call compute.VirtualMachinesClient.List with filters based on selector
	// - Convert Azure VMs to targets
	
	return nil, fmt.Errorf("Azure SDK integration not available in test environment")
}

// vmMatchesSelector checks if a VM matches the selector
func (a *AzureInventoryProvider) vmMatchesSelector(vm mockAzureVM, selector map[string]string) bool {
	// If no selector, match all running VMs
	if len(selector) == 0 {
		return vm.PowerState == "VM running"
	}
	
	// Check each selector condition
	for key, expectedValue := range selector {
		actualValue, exists := vm.Tags[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	
	// Only include running VMs
	return vm.PowerState == "VM running"
}

// parseAzureSelector parses Azure tag selector format
func parseAzureSelector(selector string) (map[string]string, error) {
	result := make(map[string]string)
	
	if selector == "" {
		return result, nil
	}
	
	// Split by comma
	pairs := strings.Split(selector, ",")
	
	for _, pair := range pairs {
		// Trim whitespace
		pair = strings.TrimSpace(pair)
		
		// Split by equals
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid selector format: %s (expected key=value)", pair)
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		if key == "" {
			return nil, fmt.Errorf("empty key in selector: %s", pair)
		}
		
		result[key] = value
	}
	
	return result, nil
}

// mockAzureVM represents a mock Azure VM for testing
type mockAzureVM struct {
	Name              string
	ResourceGroup     string
	Location          string
	PublicIPAddress   string
	PrivateIPAddress  string
	PowerState        string
	Tags              map[string]string
}

// AzureConfig represents Azure inventory configuration
type AzureConfig struct {
	SubscriptionID string `yaml:"subscription_id" json:"subscription_id"`
	ResourceGroup  string `yaml:"resource_group" json:"resource_group"`
	Region         string `yaml:"region" json:"region"`
	TenantID       string `yaml:"tenant_id" json:"tenant_id"`
	ClientID       string `yaml:"client_id" json:"client_id"`
	ClientSecret   string `yaml:"client_secret" json:"client_secret"`
}

// DefaultAzureConfig returns default Azure configuration
func DefaultAzureConfig() *AzureConfig {
	return &AzureConfig{
		SubscriptionID: "",
		ResourceGroup:  "",
		Region:         "eastus",
		TenantID:       "",
		ClientID:       "",
		ClientSecret:   "",
	}
}
