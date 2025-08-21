package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ataiva-software/chisel/pkg/core"
	"github.com/ataiva-software/chisel/pkg/types"
)

// PolicyEngine manages and evaluates policies
type PolicyEngine struct {
	policies map[string]string // policy name -> policy content
	enabled  bool
	mu       sync.RWMutex
}

// PolicyResult represents the result of a policy evaluation
type PolicyResult struct {
	Allowed    bool               `json:"allowed"`
	Violations []PolicyViolation  `json:"violations,omitempty"`
}

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	Policy   string `json:"policy"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Resource string `json:"resource"`
}

// String returns a string representation of the policy violation
func (v *PolicyViolation) String() string {
	return fmt.Sprintf("[%s] %s: %s (resource: %s)", v.Policy, v.Rule, v.Message, v.Resource)
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: make(map[string]string),
		enabled:  true,
	}
}

// IsEnabled returns whether the policy engine is enabled
func (e *PolicyEngine) IsEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.enabled
}

// Enable enables the policy engine
func (e *PolicyEngine) Enable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = true
}

// Disable disables the policy engine
func (e *PolicyEngine) Disable() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.enabled = false
}

// LoadPolicy loads a policy from content
func (e *PolicyEngine) LoadPolicy(name, content string) error {
	if name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}
	
	if content == "" {
		return fmt.Errorf("policy content cannot be empty")
	}
	
	// Basic validation - check if it looks like Rego
	if !strings.Contains(content, "package") {
		return fmt.Errorf("policy content must contain a package declaration")
	}
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.policies[name] = content
	return nil
}

// LoadPolicyFromFile loads a policy from a file
func (e *PolicyEngine) LoadPolicyFromFile(name, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}
	
	return e.LoadPolicy(name, string(content))
}

// RemovePolicy removes a policy
func (e *PolicyEngine) RemovePolicy(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if _, exists := e.policies[name]; !exists {
		return fmt.Errorf("policy '%s' not found", name)
	}
	
	delete(e.policies, name)
	return nil
}

// GetLoadedPolicies returns the names of all loaded policies
func (e *PolicyEngine) GetLoadedPolicies() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	names := make([]string, 0, len(e.policies))
	for name := range e.policies {
		names = append(names, name)
	}
	return names
}

// EvaluateResource evaluates a single resource against all loaded policies
func (e *PolicyEngine) EvaluateResource(ctx context.Context, resource *types.Resource) (*PolicyResult, error) {
	e.mu.RLock()
	enabled := e.enabled
	policies := make(map[string]string)
	for k, v := range e.policies {
		policies[k] = v
	}
	e.mu.RUnlock()
	
	// If disabled, allow everything
	if !enabled {
		return &PolicyResult{Allowed: true}, nil
	}
	
	result := &PolicyResult{
		Allowed:    true,
		Violations: make([]PolicyViolation, 0),
	}
	
	// Evaluate against each policy
	for policyName, policyContent := range policies {
		violations, err := e.evaluateResourceAgainstPolicy(ctx, resource, policyName, policyContent)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate policy %s: %w", policyName, err)
		}
		
		result.Violations = append(result.Violations, violations...)
	}
	
	// If there are violations, the resource is not allowed
	if len(result.Violations) > 0 {
		result.Allowed = false
	}
	
	return result, nil
}

// EvaluateModule evaluates all resources in a module against loaded policies
func (e *PolicyEngine) EvaluateModule(ctx context.Context, module *core.Module) (*PolicyResult, error) {
	e.mu.RLock()
	enabled := e.enabled
	policies := make(map[string]string)
	for k, v := range e.policies {
		policies[k] = v
	}
	e.mu.RUnlock()
	
	// If disabled, allow everything
	if !enabled {
		return &PolicyResult{Allowed: true}, nil
	}
	
	result := &PolicyResult{
		Allowed:    true,
		Violations: make([]PolicyViolation, 0),
	}
	
	// Evaluate each resource individually
	for _, resource := range module.Spec.Resources {
		resourceResult, err := e.EvaluateResource(ctx, &resource)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate resource %s: %w", resource.ResourceID(), err)
		}
		
		result.Violations = append(result.Violations, resourceResult.Violations...)
	}
	
	// Also evaluate module-level policies
	for policyName, policyContent := range policies {
		violations, err := e.evaluateModuleAgainstPolicy(ctx, module, policyName, policyContent)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate module policy %s: %w", policyName, err)
		}
		
		result.Violations = append(result.Violations, violations...)
	}
	
	// If there are violations, the module is not allowed
	if len(result.Violations) > 0 {
		result.Allowed = false
	}
	
	return result, nil
}

// evaluateResourceAgainstPolicy evaluates a resource against a specific policy
func (e *PolicyEngine) evaluateResourceAgainstPolicy(ctx context.Context, resource *types.Resource, policyName, policyContent string) ([]PolicyViolation, error) {
	// In a real implementation, this would use OPA (Open Policy Agent)
	// For now, we'll implement a simple rule engine that can handle basic policies
	
	violations := make([]PolicyViolation, 0)
	
	// Simple rule evaluation based on policy content
	if strings.Contains(policyContent, "chisel.security") {
		// Security policy evaluation
		if resource.Type == "user" && resource.Name == "root" {
			if strings.Contains(policyContent, "deny") && strings.Contains(policyContent, "root") {
				violations = append(violations, PolicyViolation{
					Policy:   policyName,
					Rule:     "deny_root_user",
					Message:  "Creating root user is not allowed",
					Resource: resource.ResourceID(),
				})
			}
		}
		
		if resource.Type == "file" {
			if mode, ok := resource.Properties["mode"].(string); ok {
				if mode == "0777" && strings.Contains(policyContent, "0777") {
					violations = append(violations, PolicyViolation{
						Policy:   policyName,
						Rule:     "insecure_permissions",
						Message:  fmt.Sprintf("File %s has insecure permissions: 0777", resource.Name),
						Resource: resource.ResourceID(),
					})
				}
			}
		}
	}
	
	if strings.Contains(policyContent, "chisel.compliance") {
		// Compliance policy evaluation
		if resource.Type == "service" {
			if enabled, ok := resource.Properties["enabled"].(bool); ok && !enabled {
				if strings.Contains(policyContent, "enabled == false") {
					violations = append(violations, PolicyViolation{
						Policy:   policyName,
						Rule:     "service_must_be_enabled",
						Message:  fmt.Sprintf("Service %s must be enabled for compliance", resource.Name),
						Resource: resource.ResourceID(),
					})
				}
			}
		}
	}
	
	return violations, nil
}

// evaluateModuleAgainstPolicy evaluates a module against a specific policy
func (e *PolicyEngine) evaluateModuleAgainstPolicy(ctx context.Context, module *core.Module, policyName, policyContent string) ([]PolicyViolation, error) {
	violations := make([]PolicyViolation, 0)
	
	// Module-level policy evaluation
	if strings.Contains(policyContent, "chisel.compliance") {
		// Check for required SSH service
		if strings.Contains(policyContent, "required_services") && strings.Contains(policyContent, "ssh") {
			hasSSH := false
			for _, resource := range module.Spec.Resources {
				if resource.Type == "service" && (resource.Name == "ssh" || resource.Name == "sshd") {
					hasSSH = true
					break
				}
			}
			
			if !hasSSH {
				violations = append(violations, PolicyViolation{
					Policy:   policyName,
					Rule:     "required_ssh_service",
					Message:  "SSH service is required for compliance",
					Resource: module.Metadata.Name,
				})
			}
		}
	}
	
	return violations, nil
}

// PolicyConfig represents policy engine configuration
type PolicyConfig struct {
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	PolicyPaths []string          `yaml:"policy_paths" json:"policy_paths"`
	Policies    map[string]string `yaml:"policies" json:"policies"`
}

// DefaultPolicyConfig returns default policy configuration
func DefaultPolicyConfig() *PolicyConfig {
	return &PolicyConfig{
		Enabled:     true,
		PolicyPaths: []string{},
		Policies:    make(map[string]string),
	}
}

// LoadFromConfig loads policies from configuration
func (e *PolicyEngine) LoadFromConfig(config *PolicyConfig) error {
	if config == nil {
		return fmt.Errorf("policy config cannot be nil")
	}
	
	if !config.Enabled {
		e.Disable()
		return nil
	}
	
	e.Enable()
	
	// Load policies from content
	for name, content := range config.Policies {
		if err := e.LoadPolicy(name, content); err != nil {
			return fmt.Errorf("failed to load policy %s: %w", name, err)
		}
	}
	
	// Load policies from files
	for _, path := range config.PolicyPaths {
		// Extract policy name from file path
		name := extractPolicyName(path)
		if err := e.LoadPolicyFromFile(name, path); err != nil {
			return fmt.Errorf("failed to load policy from %s: %w", path, err)
		}
	}
	
	return nil
}

// extractPolicyName extracts policy name from file path
func extractPolicyName(path string) string {
	// Simple extraction - use filename without extension
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[:idx]
	}
	
	return filename
}

// PolicyInput represents input data for policy evaluation
type PolicyInput struct {
	Resource  *types.Resource   `json:"resource,omitempty"`
	Resources []types.Resource  `json:"resources,omitempty"`
	Module    *core.Module      `json:"module,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ToJSON converts policy input to JSON for OPA evaluation
func (p *PolicyInput) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
