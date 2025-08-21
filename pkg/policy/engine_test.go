package policy

import (
	"context"
	"testing"

	"github.com/ataiva-software/chisel/pkg/core"
	"github.com/ataiva-software/chisel/pkg/types"
)

func TestPolicyEngine_New(t *testing.T) {
	engine := NewPolicyEngine()
	
	if engine == nil {
		t.Fatal("Expected non-nil policy engine")
	}
	
	if !engine.IsEnabled() {
		t.Error("Expected policy engine to be enabled by default")
	}
}

func TestPolicyEngine_LoadPolicy(t *testing.T) {
	engine := NewPolicyEngine()
	
	// Test loading a simple policy
	policyContent := `
package chisel.security

# Deny root user creation
deny[msg] {
	input.resource.type == "user"
	input.resource.name == "root"
	msg := "Creating root user is not allowed"
}

# Require specific file permissions
deny[msg] {
	input.resource.type == "file"
	input.resource.properties.mode
	not startswith(input.resource.properties.mode, "0644")
	not startswith(input.resource.properties.mode, "0755")
	msg := sprintf("File %s has insecure permissions: %s", [input.resource.name, input.resource.properties.mode])
}
`
	
	err := engine.LoadPolicy("security", policyContent)
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}
	
	policies := engine.GetLoadedPolicies()
	if len(policies) != 1 {
		t.Errorf("Expected 1 loaded policy, got %d", len(policies))
	}
	
	if policies[0] != "security" {
		t.Errorf("Expected policy name 'security', got '%s'", policies[0])
	}
}

func TestPolicyEngine_LoadPolicyFromFile(t *testing.T) {
	engine := NewPolicyEngine()
	
	// This would fail in test environment without actual file
	err := engine.LoadPolicyFromFile("security", "/nonexistent/policy.rego")
	if err == nil {
		t.Error("Expected error for non-existent policy file")
	}
}

func TestPolicyEngine_EvaluateResource(t *testing.T) {
	engine := NewPolicyEngine()
	
	// Load security policy
	policyContent := `
package chisel.security

deny[msg] {
	input.resource.type == "user"
	input.resource.name == "root"
	msg := "Creating root user is not allowed"
}

deny[msg] {
	input.resource.type == "file"
	input.resource.properties.mode == "0777"
	msg := sprintf("File %s has insecure permissions: 0777", [input.resource.name])
}
`
	
	err := engine.LoadPolicy("security", policyContent)
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}
	
	tests := []struct {
		name           string
		resource       *types.Resource
		expectViolation bool
		expectedMsg     string
	}{
		{
			name: "allowed user creation",
			resource: &types.Resource{
				Type: "user",
				Name: "appuser",
				Properties: map[string]interface{}{
					"uid": 1001,
				},
			},
			expectViolation: false,
		},
		{
			name: "denied root user creation",
			resource: &types.Resource{
				Type: "user",
				Name: "root",
				Properties: map[string]interface{}{
					"uid": 0,
				},
			},
			expectViolation: true,
			expectedMsg:     "Creating root user is not allowed",
		},
		{
			name: "allowed file permissions",
			resource: &types.Resource{
				Type: "file",
				Name: "config.txt",
				Properties: map[string]interface{}{
					"path": "/etc/config.txt",
					"mode": "0644",
				},
			},
			expectViolation: false,
		},
		{
			name: "denied insecure file permissions",
			resource: &types.Resource{
				Type: "file",
				Name: "secret.txt",
				Properties: map[string]interface{}{
					"path": "/etc/secret.txt",
					"mode": "0777",
				},
			},
			expectViolation: true,
			expectedMsg:     "File secret.txt has insecure permissions: 0777",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := engine.EvaluateResource(ctx, tt.resource)
			if err != nil {
				t.Fatalf("Failed to evaluate resource: %v", err)
			}
			
			if tt.expectViolation {
				if result.Allowed {
					t.Error("Expected policy violation but resource was allowed")
				}
				
				if len(result.Violations) == 0 {
					t.Error("Expected policy violations but got none")
				}
				
				if tt.expectedMsg != "" {
					found := false
					for _, violation := range result.Violations {
						if violation.Message == tt.expectedMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected violation message '%s' not found in %v", tt.expectedMsg, result.Violations)
					}
				}
			} else {
				if !result.Allowed {
					t.Errorf("Expected resource to be allowed but got violations: %v", result.Violations)
				}
			}
		})
	}
}

func TestPolicyEngine_EvaluateModule(t *testing.T) {
	engine := NewPolicyEngine()
	
	// Load compliance policy
	policyContent := `
package chisel.compliance

# All services must be enabled
deny[msg] {
	input.resource.type == "service"
	input.resource.properties.enabled == false
	msg := sprintf("Service %s must be enabled for compliance", [input.resource.name])
}

# SSH service is required
required_services := ["ssh", "sshd"]

deny[msg] {
	count([r | r := input.resources[_]; r.type == "service"; r.name == required_services[_]]) == 0
	msg := "SSH service is required for compliance"
}
`
	
	err := engine.LoadPolicy("compliance", policyContent)
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}
	
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name:    "test-module",
			Version: "1.0.0",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type: "service",
					Name: "nginx",
					Properties: map[string]interface{}{
						"enabled": true,
					},
				},
				{
					Type: "service",
					Name: "ssh",
					Properties: map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
	}
	
	ctx := context.Background()
	result, err := engine.EvaluateModule(ctx, module)
	if err != nil {
		t.Fatalf("Failed to evaluate module: %v", err)
	}
	
	if !result.Allowed {
		t.Errorf("Expected module to be allowed but got violations: %v", result.Violations)
	}
	
	// Test with policy violation
	module.Spec.Resources[0].Properties["enabled"] = false
	
	result, err = engine.EvaluateModule(ctx, module)
	if err != nil {
		t.Fatalf("Failed to evaluate module: %v", err)
	}
	
	if result.Allowed {
		t.Error("Expected policy violation but module was allowed")
	}
	
	if len(result.Violations) == 0 {
		t.Error("Expected policy violations but got none")
	}
}

func TestPolicyEngine_EnableDisable(t *testing.T) {
	engine := NewPolicyEngine()
	
	if !engine.IsEnabled() {
		t.Error("Expected engine to be enabled by default")
	}
	
	engine.Disable()
	if engine.IsEnabled() {
		t.Error("Expected engine to be disabled after Disable()")
	}
	
	engine.Enable()
	if !engine.IsEnabled() {
		t.Error("Expected engine to be enabled after Enable()")
	}
}

func TestPolicyEngine_EvaluateWhenDisabled(t *testing.T) {
	engine := NewPolicyEngine()
	engine.Disable()
	
	resource := &types.Resource{
		Type: "user",
		Name: "root",
	}
	
	ctx := context.Background()
	result, err := engine.EvaluateResource(ctx, resource)
	if err != nil {
		t.Fatalf("Unexpected error when engine disabled: %v", err)
	}
	
	if !result.Allowed {
		t.Error("Expected resource to be allowed when policy engine is disabled")
	}
}

func TestPolicyEngine_RemovePolicy(t *testing.T) {
	engine := NewPolicyEngine()
	
	policyContent := `
package chisel.test

deny[msg] {
	input.resource.name == "forbidden"
	msg := "Forbidden resource"
}
`
	
	err := engine.LoadPolicy("test", policyContent)
	if err != nil {
		t.Fatalf("Failed to load policy: %v", err)
	}
	
	policies := engine.GetLoadedPolicies()
	if len(policies) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(policies))
	}
	
	err = engine.RemovePolicy("test")
	if err != nil {
		t.Fatalf("Failed to remove policy: %v", err)
	}
	
	policies = engine.GetLoadedPolicies()
	if len(policies) != 0 {
		t.Errorf("Expected 0 policies after removal, got %d", len(policies))
	}
	
	// Test removing non-existent policy
	err = engine.RemovePolicy("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent policy")
	}
}

func TestPolicyViolation_String(t *testing.T) {
	violation := &PolicyViolation{
		Policy:   "security",
		Rule:     "deny_root_user",
		Message:  "Root user creation is not allowed",
		Resource: "user.root",
	}
	
	str := violation.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
	
	// Should contain key information
	if !contains(str, "security") {
		t.Error("Expected string to contain policy name")
	}
	
	if !contains(str, "Root user creation is not allowed") {
		t.Error("Expected string to contain message")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   (len(s) > len(substr) && containsHelper(s[1:], substr))
}

func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}
