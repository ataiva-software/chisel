package compliance

import (
	"context"
	"testing"

	"github.com/ataiva-software/chisel/pkg/core"
	"github.com/ataiva-software/chisel/pkg/types"
)

func TestComplianceManager_New(t *testing.T) {
	manager := NewComplianceManager()
	
	if manager == nil {
		t.Fatal("Expected non-nil compliance manager")
	}
	
	if !manager.IsEnabled() {
		t.Error("Expected compliance manager to be enabled by default")
	}
}

func TestComplianceManager_LoadCISModule(t *testing.T) {
	manager := NewComplianceManager()
	
	err := manager.LoadModule("cis-ubuntu-20.04")
	if err != nil {
		t.Fatalf("Failed to load CIS module: %v", err)
	}
	
	modules := manager.GetLoadedModules()
	if len(modules) != 1 {
		t.Errorf("Expected 1 loaded module, got %d", len(modules))
	}
	
	if modules[0] != "cis-ubuntu-20.04" {
		t.Errorf("Expected module 'cis-ubuntu-20.04', got '%s'", modules[0])
	}
}

func TestComplianceManager_CheckCISCompliance(t *testing.T) {
	manager := NewComplianceManager()
	manager.LoadModule("cis-ubuntu-20.04")
	
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
					Type: "file",
					Name: "passwd-permissions",
					Properties: map[string]interface{}{
						"path": "/etc/passwd",
						"mode": "0644",
					},
				},
				{
					Type: "file",
					Name: "shadow-permissions",
					Properties: map[string]interface{}{
						"path": "/etc/shadow",
						"mode": "0640",
					},
				},
				{
					Type: "service",
					Name: "ssh",
					Properties: map[string]interface{}{
						"enabled": true,
						"state":   "running",
					},
				},
			},
		},
	}
	
	ctx := context.Background()
	result, err := manager.CheckCompliance(ctx, module)
	if err != nil {
		t.Fatalf("Failed to check compliance: %v", err)
	}
	
	if !result.Compliant {
		t.Errorf("Expected module to be compliant, got violations: %v", result.Violations)
	}
	
	if result.Framework != "CIS" {
		t.Errorf("Expected framework 'CIS', got '%s'", result.Framework)
	}
	
	if result.Version != "Ubuntu 20.04" {
		t.Errorf("Expected version 'Ubuntu 20.04', got '%s'", result.Version)
	}
}

func TestComplianceManager_CheckCISViolations(t *testing.T) {
	manager := NewComplianceManager()
	manager.LoadModule("cis-ubuntu-20.04")
	
	// Module with CIS violations
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name: "non-compliant-module",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type: "file",
					Name: "passwd-bad-permissions",
					Properties: map[string]interface{}{
						"path": "/etc/passwd",
						"mode": "0777", // CIS violation - too permissive
					},
				},
				{
					Type: "user",
					Name: "root",
					Properties: map[string]interface{}{
						"shell": "/bin/bash", // CIS violation - root should not have interactive shell
					},
				},
			},
		},
	}
	
	ctx := context.Background()
	result, err := manager.CheckCompliance(ctx, module)
	if err != nil {
		t.Fatalf("Failed to check compliance: %v", err)
	}
	
	if result.Compliant {
		t.Error("Expected module to be non-compliant")
	}
	
	if len(result.Violations) == 0 {
		t.Error("Expected compliance violations")
	}
	
	// Check specific violations
	foundPasswdViolation := false
	foundRootShellViolation := false
	
	for _, violation := range result.Violations {
		if violation.Control == "CIS-6.1.2" && violation.Resource == "file.passwd-bad-permissions" {
			foundPasswdViolation = true
		}
		if violation.Control == "CIS-5.4.2" && violation.Resource == "user.root" {
			foundRootShellViolation = true
		}
	}
	
	if !foundPasswdViolation {
		t.Error("Expected passwd permissions violation")
	}
	
	if !foundRootShellViolation {
		t.Error("Expected root shell violation")
	}
}

func TestComplianceManager_LoadNISTModule(t *testing.T) {
	manager := NewComplianceManager()
	
	err := manager.LoadModule("nist-800-53")
	if err != nil {
		t.Fatalf("Failed to load NIST module: %v", err)
	}
	
	modules := manager.GetLoadedModules()
	found := false
	for _, module := range modules {
		if module == "nist-800-53" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected NIST module to be loaded")
	}
}

func TestComplianceManager_CheckNISTCompliance(t *testing.T) {
	manager := NewComplianceManager()
	manager.LoadModule("nist-800-53")
	
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name: "nist-test-module",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type: "service",
					Name: "auditd",
					Properties: map[string]interface{}{
						"enabled": true,
						"state":   "running",
					},
				},
				{
					Type: "file",
					Name: "audit-config",
					Properties: map[string]interface{}{
						"path":    "/etc/audit/auditd.conf",
						"content": "log_file = /var/log/audit/audit.log\n",
					},
				},
			},
		},
	}
	
	ctx := context.Background()
	result, err := manager.CheckCompliance(ctx, module)
	if err != nil {
		t.Fatalf("Failed to check NIST compliance: %v", err)
	}
	
	if result.Framework != "NIST" {
		t.Errorf("Expected framework 'NIST', got '%s'", result.Framework)
	}
	
	if result.Version != "800-53" {
		t.Errorf("Expected version '800-53', got '%s'", result.Version)
	}
}

func TestComplianceManager_LoadSTIGModule(t *testing.T) {
	manager := NewComplianceManager()
	
	err := manager.LoadModule("stig-rhel8")
	if err != nil {
		t.Fatalf("Failed to load STIG module: %v", err)
	}
	
	modules := manager.GetLoadedModules()
	found := false
	for _, module := range modules {
		if module == "stig-rhel8" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected STIG module to be loaded")
	}
}

func TestComplianceManager_CheckMultipleFrameworks(t *testing.T) {
	manager := NewComplianceManager()
	manager.LoadModule("cis-ubuntu-20.04")
	manager.LoadModule("nist-800-53")
	
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name: "multi-compliance-module",
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type: "service",
					Name: "ssh",
					Properties: map[string]interface{}{
						"enabled": true,
					},
				},
				{
					Type: "service",
					Name: "auditd",
					Properties: map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
	}
	
	ctx := context.Background()
	results, err := manager.CheckAllCompliance(ctx, module)
	if err != nil {
		t.Fatalf("Failed to check multi-framework compliance: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 compliance results, got %d", len(results))
	}
	
	// Verify we have both CIS and NIST results
	frameworks := make(map[string]bool)
	for _, result := range results {
		frameworks[result.Framework] = true
	}
	
	if !frameworks["CIS"] {
		t.Error("Expected CIS compliance result")
	}
	
	if !frameworks["NIST"] {
		t.Error("Expected NIST compliance result")
	}
}

func TestComplianceManager_EnableDisable(t *testing.T) {
	manager := NewComplianceManager()
	
	if !manager.IsEnabled() {
		t.Error("Expected compliance manager to be enabled by default")
	}
	
	manager.Disable()
	if manager.IsEnabled() {
		t.Error("Expected compliance manager to be disabled after Disable()")
	}
	
	// When disabled, compliance checks should pass
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test"},
		Spec:     core.ModuleSpec{Resources: []types.Resource{}},
	}
	
	ctx := context.Background()
	result, err := manager.CheckCompliance(ctx, module)
	if err != nil {
		t.Fatalf("Unexpected error when disabled: %v", err)
	}
	
	if !result.Compliant {
		t.Error("Expected compliance to pass when manager is disabled")
	}
	
	manager.Enable()
	if !manager.IsEnabled() {
		t.Error("Expected compliance manager to be enabled after Enable()")
	}
}

func TestComplianceViolation_String(t *testing.T) {
	violation := &ComplianceViolation{
		Framework:   "CIS",
		Version:     "Ubuntu 20.04",
		Control:     "CIS-6.1.2",
		Title:       "Ensure permissions on /etc/passwd are configured",
		Description: "The /etc/passwd file should have 644 permissions",
		Severity:    SeverityHigh,
		Resource:    "file.passwd",
		Message:     "File has permissions 0777, expected 0644",
	}
	
	str := violation.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
	
	// Should contain key information
	if !contains(str, "CIS-6.1.2") {
		t.Error("Expected string to contain control ID")
	}
	
	if !contains(str, "HIGH") {
		t.Error("Expected string to contain severity")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || 
		   (len(s) > len(substr) && containsHelper(s[1:], substr)))
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
