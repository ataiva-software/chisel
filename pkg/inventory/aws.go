package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/ataiva-software/chisel/pkg/types"
)

// AWSInventoryProvider discovers targets from AWS EC2 instances
type AWSInventoryProvider struct {
	region   string
	profile  string
	roleArn  string
	mockMode bool
}

// NewAWSInventoryProvider creates a new AWS inventory provider
func NewAWSInventoryProvider(region, profile, roleArn string) *AWSInventoryProvider {
	return &AWSInventoryProvider{
		region:   region,
		profile:  profile,
		roleArn:  roleArn,
		mockMode: false,
	}
}

// Type returns the provider type
func (a *AWSInventoryProvider) Type() string {
	return "aws"
}

// Validate validates the provider configuration
func (a *AWSInventoryProvider) Validate() error {
	if a.region == "" {
		return fmt.Errorf("AWS region is required")
	}
	
	return nil
}

// SetMockMode enables mock mode for testing
func (a *AWSInventoryProvider) SetMockMode(enabled bool) {
	a.mockMode = enabled
}

// Discover discovers EC2 instances based on the selector
func (a *AWSInventoryProvider) Discover(ctx context.Context, selector string) ([]types.Target, error) {
	if a.mockMode {
		return a.discoverMock(ctx, selector)
	}
	
	return a.discoverReal(ctx, selector)
}

// discoverMock provides mock data for testing
func (a *AWSInventoryProvider) discoverMock(ctx context.Context, selector string) ([]types.Target, error) {
	// Parse selector
	selectorMap, err := parseAWSSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid selector: %w", err)
	}
	
	// Mock EC2 instances
	mockInstances := []mockEC2Instance{
		{
			InstanceID:       "i-1234567890abcdef0",
			PublicIPAddress:  "203.0.113.1",
			PrivateIPAddress: "10.0.1.10",
			State:            "running",
			Tags: map[string]string{
				"Name":        "web-server-1",
				"Environment": "production",
				"Role":        "web",
			},
		},
		{
			InstanceID:       "i-0987654321fedcba0",
			PublicIPAddress:  "203.0.113.2",
			PrivateIPAddress: "10.0.1.20",
			State:            "running",
			Tags: map[string]string{
				"Name":        "db-server-1",
				"Environment": "staging",
				"Role":        "database",
			},
		},
	}
	
	// Filter instances based on selector
	var matchingInstances []mockEC2Instance
	for _, instance := range mockInstances {
		if a.instanceMatchesSelector(instance, selectorMap) {
			matchingInstances = append(matchingInstances, instance)
		}
	}
	
	// Convert to targets
	var targets []types.Target
	for _, instance := range matchingInstances {
		target := types.Target{
			Host: instance.PublicIPAddress,
			Port: 22, // Default SSH port
			User: "ec2-user", // Default EC2 user
			Labels: make(map[string]string),
		}
		
		// Copy tags as labels
		for key, value := range instance.Tags {
			target.Labels[key] = value
		}
		
		// Add AWS-specific labels
		target.Labels["aws:instance-id"] = instance.InstanceID
		target.Labels["aws:private-ip"] = instance.PrivateIPAddress
		target.Labels["aws:region"] = a.region
		target.Labels["aws:state"] = instance.State
		
		targets = append(targets, target)
	}
	
	return targets, nil
}

// discoverReal discovers real EC2 instances (would use AWS SDK)
func (a *AWSInventoryProvider) discoverReal(ctx context.Context, selector string) ([]types.Target, error) {
	// In a real implementation, this would use the AWS SDK:
	// - Create EC2 client with region and credentials
	// - Call DescribeInstances with filters based on selector
	// - Convert EC2 instances to targets
	
	return nil, fmt.Errorf("AWS SDK integration not available in test environment")
}

// instanceMatchesSelector checks if an instance matches the selector
func (a *AWSInventoryProvider) instanceMatchesSelector(instance mockEC2Instance, selector map[string]string) bool {
	// If no selector, match all running instances
	if len(selector) == 0 {
		return instance.State == "running"
	}
	
	// Check each selector condition
	for key, expectedValue := range selector {
		actualValue, exists := instance.Tags[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	
	// Only include running instances
	return instance.State == "running"
}

// parseAWSSelector parses AWS tag selector format
func parseAWSSelector(selector string) (map[string]string, error) {
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

// mockEC2Instance represents a mock EC2 instance for testing
type mockEC2Instance struct {
	InstanceID       string
	PublicIPAddress  string
	PrivateIPAddress string
	State            string
	Tags             map[string]string
}
