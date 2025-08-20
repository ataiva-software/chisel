package types

import (
	"context"
	"fmt"
)

// ResourceState represents the desired state of a resource
type ResourceState string

const (
	StatePresent ResourceState = "present"
	StateAbsent  ResourceState = "absent"
	StateRunning ResourceState = "running"
	StateStopped ResourceState = "stopped"
)

// Resource represents a unit of infrastructure state
type Resource struct {
	Type         string                 `yaml:"type" json:"type"`
	Name         string                 `yaml:"name" json:"name"`
	State        ResourceState          `yaml:"state,omitempty" json:"state,omitempty"`
	Properties   map[string]interface{} `yaml:",inline" json:",inline"`
	DependsOn    []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Notify       []string               `yaml:"notify,omitempty" json:"notify,omitempty"`
	OnlyIf       string                 `yaml:"only_if,omitempty" json:"only_if,omitempty"`
	NotIf        string                 `yaml:"not_if,omitempty" json:"not_if,omitempty"`
}

// ResourceID returns a unique identifier for the resource
func (r *Resource) ResourceID() string {
	return fmt.Sprintf("%s.%s", r.Type, r.Name)
}

// Validate checks if the resource configuration is valid
func (r *Resource) Validate() error {
	if r.Type == "" {
		return fmt.Errorf("resource type cannot be empty")
	}
	if r.Name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}
	return nil
}

// ResourceDiff represents the difference between current and desired state
type ResourceDiff struct {
	ResourceID string                 `json:"resource_id"`
	Action     DiffAction             `json:"action"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
}

// DiffAction represents the type of change needed
type DiffAction string

const (
	ActionCreate DiffAction = "create"
	ActionUpdate DiffAction = "update"
	ActionDelete DiffAction = "delete"
	ActionNoop   DiffAction = "noop"
)

// Provider defines the interface that all resource providers must implement
type Provider interface {
	// Type returns the resource type this provider handles
	Type() string

	// Validate validates the resource configuration
	Validate(resource *Resource) error

	// Read reads the current state of the resource
	Read(ctx context.Context, resource *Resource) (map[string]interface{}, error)

	// Diff compares desired vs current state and returns the differences
	Diff(ctx context.Context, resource *Resource, current map[string]interface{}) (*ResourceDiff, error)

	// Apply applies the changes to bring the resource to desired state
	Apply(ctx context.Context, resource *Resource, diff *ResourceDiff) error
}

// ProviderRegistry manages available providers
type ProviderRegistry struct {
	providers map[string]Provider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider for a resource type
func (pr *ProviderRegistry) Register(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}
	
	providerType := provider.Type()
	if providerType == "" {
		return fmt.Errorf("provider type cannot be empty")
	}
	
	if _, exists := pr.providers[providerType]; exists {
		return fmt.Errorf("provider for type %s already registered", providerType)
	}
	
	pr.providers[providerType] = provider
	return nil
}

// Get retrieves a provider for the given resource type
func (pr *ProviderRegistry) Get(resourceType string) (Provider, error) {
	provider, exists := pr.providers[resourceType]
	if !exists {
		return nil, fmt.Errorf("no provider registered for resource type: %s", resourceType)
	}
	return provider, nil
}

// Types returns all registered provider types
func (pr *ProviderRegistry) Types() []string {
	types := make([]string, 0, len(pr.providers))
	for t := range pr.providers {
		types = append(types, t)
	}
	return types
}
