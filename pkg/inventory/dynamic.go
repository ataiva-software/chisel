package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/ataiva-software/forge/pkg/types"
)

// DynamicInventory represents a dynamic inventory source
type DynamicInventory interface {
	// Discover discovers targets based on selectors
	Discover(ctx context.Context, selector string) ([]types.Target, error)
	
	// Type returns the inventory type (e.g., "aws", "azure", "gcp")
	Type() string
	
	// Validate validates the inventory configuration
	Validate() error
}

// InventoryRegistry manages dynamic inventory providers
type InventoryRegistry struct {
	providers map[string]DynamicInventory
}

// NewInventoryRegistry creates a new inventory registry
func NewInventoryRegistry() *InventoryRegistry {
	return &InventoryRegistry{
		providers: make(map[string]DynamicInventory),
	}
}

// Register registers a dynamic inventory provider
func (r *InventoryRegistry) Register(provider DynamicInventory) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}
	
	providerType := provider.Type()
	if providerType == "" {
		return fmt.Errorf("provider type cannot be empty")
	}
	
	if err := provider.Validate(); err != nil {
		return fmt.Errorf("provider validation failed: %w", err)
	}
	
	r.providers[providerType] = provider
	return nil
}

// Discover discovers targets using the specified provider and selector
func (r *InventoryRegistry) Discover(ctx context.Context, providerType, selector string) ([]types.Target, error) {
	provider, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("no provider found for type: %s", providerType)
	}
	
	return provider.Discover(ctx, selector)
}

// ListProviders returns a list of registered provider types
func (r *InventoryRegistry) ListProviders() []string {
	var types []string
	for providerType := range r.providers {
		types = append(types, providerType)
	}
	return types
}

// ParseSelector parses a selector string into key-value pairs
// Format: "key1=value1,key2=value2"
func ParseSelector(selector string) (map[string]string, error) {
	if selector == "" {
		return map[string]string{}, nil
	}
	
	result := make(map[string]string)
	pairs := strings.Split(selector, ",")
	
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
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

// MatchesSelector checks if a target matches the given selector
func MatchesSelector(target types.Target, selector map[string]string) bool {
	for key, expectedValue := range selector {
		targetValue, exists := target.Labels[key]
		if !exists {
			return false
		}
		
		if targetValue != expectedValue {
			return false
		}
	}
	
	return true
}
