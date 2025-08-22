package compliance

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ataiva-software/forge/pkg/core"
	"github.com/ataiva-software/forge/pkg/types"
)

// Severity represents the severity level of a compliance violation
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	Framework   string   `json:"framework"`
	Version     string   `json:"version"`
	Control     string   `json:"control"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Resource    string   `json:"resource"`
	Message     string   `json:"message"`
}

// String returns a string representation of the violation
func (v *ComplianceViolation) String() string {
	return fmt.Sprintf("[%s %s] %s - %s: %s (Resource: %s)", 
		v.Framework, v.Severity, v.Control, v.Title, v.Message, v.Resource)
}

// ComplianceResult represents the result of a compliance check
type ComplianceResult struct {
	Framework  string                 `json:"framework"`
	Version    string                 `json:"version"`
	Compliant  bool                   `json:"compliant"`
	Violations []ComplianceViolation  `json:"violations"`
	Passed     int                    `json:"passed"`
	Failed     int                    `json:"failed"`
	Total      int                    `json:"total"`
}

// ComplianceModule defines a compliance framework module
type ComplianceModule interface {
	Framework() string
	Version() string
	CheckCompliance(ctx context.Context, module *core.Module) (*ComplianceResult, error)
}

// ComplianceManager manages compliance modules and checks
type ComplianceManager struct {
	modules map[string]ComplianceModule
	enabled bool
	mu      sync.RWMutex
}

// NewComplianceManager creates a new compliance manager
func NewComplianceManager() *ComplianceManager {
	return &ComplianceManager{
		modules: make(map[string]ComplianceModule),
		enabled: true,
	}
}

// IsEnabled returns whether compliance checking is enabled
func (m *ComplianceManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// Enable enables compliance checking
func (m *ComplianceManager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables compliance checking
func (m *ComplianceManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// LoadModule loads a compliance module by name
func (m *ComplianceManager) LoadModule(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var module ComplianceModule
	
	switch name {
	case "cis-ubuntu-20.04":
		module = NewCISUbuntu2004Module()
	case "nist-800-53":
		module = NewNIST80053Module()
	case "stig-rhel8":
		module = NewSTIGRHEL8Module()
	default:
		return fmt.Errorf("unknown compliance module: %s", name)
	}
	
	m.modules[name] = module
	return nil
}

// GetLoadedModules returns the names of loaded compliance modules
func (m *ComplianceManager) GetLoadedModules() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	names := make([]string, 0, len(m.modules))
	for name := range m.modules {
		names = append(names, name)
	}
	return names
}

// CheckCompliance checks compliance against the first loaded module
func (m *ComplianceManager) CheckCompliance(ctx context.Context, module *core.Module) (*ComplianceResult, error) {
	if !m.IsEnabled() {
		return &ComplianceResult{
			Framework: "DISABLED",
			Compliant: true,
			Passed:    len(module.Spec.Resources),
			Total:     len(module.Spec.Resources),
		}, nil
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.modules) == 0 {
		return nil, fmt.Errorf("no compliance modules loaded")
	}
	
	// Use the first loaded module
	for _, complianceModule := range m.modules {
		return complianceModule.CheckCompliance(ctx, module)
	}
	
	return nil, fmt.Errorf("no compliance modules available")
}

// CheckAllCompliance checks compliance against all loaded modules
func (m *ComplianceManager) CheckAllCompliance(ctx context.Context, module *core.Module) ([]*ComplianceResult, error) {
	if !m.IsEnabled() {
		return []*ComplianceResult{{
			Framework: "DISABLED",
			Compliant: true,
			Passed:    len(module.Spec.Resources),
			Total:     len(module.Spec.Resources),
		}}, nil
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.modules) == 0 {
		return nil, fmt.Errorf("no compliance modules loaded")
	}
	
	results := make([]*ComplianceResult, 0, len(m.modules))
	
	for _, complianceModule := range m.modules {
		result, err := complianceModule.CheckCompliance(ctx, module)
		if err != nil {
			return nil, fmt.Errorf("compliance check failed for %s: %w", complianceModule.Framework(), err)
		}
		results = append(results, result)
	}
	
	return results, nil
}

// CISUbuntu2004Module implements CIS Ubuntu 20.04 compliance checks
type CISUbuntu2004Module struct{}

// NewCISUbuntu2004Module creates a new CIS Ubuntu 20.04 compliance module
func NewCISUbuntu2004Module() *CISUbuntu2004Module {
	return &CISUbuntu2004Module{}
}

// Framework returns the compliance framework name
func (c *CISUbuntu2004Module) Framework() string {
	return "CIS"
}

// Version returns the compliance framework version
func (c *CISUbuntu2004Module) Version() string {
	return "Ubuntu 20.04"
}

// CheckCompliance checks CIS Ubuntu 20.04 compliance
func (c *CISUbuntu2004Module) CheckCompliance(ctx context.Context, module *core.Module) (*ComplianceResult, error) {
	result := &ComplianceResult{
		Framework:  "CIS",
		Version:    "Ubuntu 20.04",
		Compliant:  true,
		Violations: make([]ComplianceViolation, 0),
	}
	
	for _, resource := range module.Spec.Resources {
		violations := c.checkResource(&resource)
		result.Violations = append(result.Violations, violations...)
		
		if len(violations) > 0 {
			result.Failed++
		} else {
			result.Passed++
		}
		result.Total++
	}
	
	result.Compliant = len(result.Violations) == 0
	return result, nil
}

// checkResource checks a single resource against CIS controls
func (c *CISUbuntu2004Module) checkResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	switch resource.Type {
	case "file":
		violations = append(violations, c.checkFileResource(resource)...)
	case "user":
		violations = append(violations, c.checkUserResource(resource)...)
	case "service":
		violations = append(violations, c.checkServiceResource(resource)...)
	}
	
	return violations
}

// checkFileResource checks file resources against CIS controls
func (c *CISUbuntu2004Module) checkFileResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	path, _ := resource.Properties["path"].(string)
	mode, _ := resource.Properties["mode"].(string)
	
	// CIS 6.1.2 - Ensure permissions on /etc/passwd are configured
	if path == "/etc/passwd" && mode != "0644" {
		violations = append(violations, ComplianceViolation{
			Framework:   "CIS",
			Version:     "Ubuntu 20.04",
			Control:     "CIS-6.1.2",
			Title:       "Ensure permissions on /etc/passwd are configured",
			Description: "The /etc/passwd file should have 644 permissions",
			Severity:    SeverityHigh,
			Resource:    resource.ResourceID(),
			Message:     fmt.Sprintf("File has permissions %s, expected 0644", mode),
		})
	}
	
	// CIS 6.1.3 - Ensure permissions on /etc/shadow are configured
	if path == "/etc/shadow" && mode != "0640" && mode != "0000" {
		violations = append(violations, ComplianceViolation{
			Framework:   "CIS",
			Version:     "Ubuntu 20.04",
			Control:     "CIS-6.1.3",
			Title:       "Ensure permissions on /etc/shadow are configured",
			Description: "The /etc/shadow file should have 640 or 000 permissions",
			Severity:    SeverityHigh,
			Resource:    resource.ResourceID(),
			Message:     fmt.Sprintf("File has permissions %s, expected 0640 or 0000", mode),
		})
	}
	
	return violations
}

// checkUserResource checks user resources against CIS controls
func (c *CISUbuntu2004Module) checkUserResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	// CIS 5.4.2 - Ensure system accounts are secured
	if resource.Name == "root" {
		shell, _ := resource.Properties["shell"].(string)
		if shell == "/bin/bash" || shell == "/bin/sh" {
			violations = append(violations, ComplianceViolation{
				Framework:   "CIS",
				Version:     "Ubuntu 20.04",
				Control:     "CIS-5.4.2",
				Title:       "Ensure system accounts are secured",
				Description: "Root account should not have an interactive shell",
				Severity:    SeverityMedium,
				Resource:    resource.ResourceID(),
				Message:     fmt.Sprintf("Root user has interactive shell: %s", shell),
			})
		}
	}
	
	return violations
}

// checkServiceResource checks service resources against CIS controls
func (c *CISUbuntu2004Module) checkServiceResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	// CIS controls generally require certain services to be enabled
	// This is a simplified check - real CIS would have more complex rules
	
	return violations
}

// NIST80053Module implements NIST 800-53 compliance checks
type NIST80053Module struct{}

// NewNIST80053Module creates a new NIST 800-53 compliance module
func NewNIST80053Module() *NIST80053Module {
	return &NIST80053Module{}
}

// Framework returns the compliance framework name
func (n *NIST80053Module) Framework() string {
	return "NIST"
}

// Version returns the compliance framework version
func (n *NIST80053Module) Version() string {
	return "800-53"
}

// CheckCompliance checks NIST 800-53 compliance
func (n *NIST80053Module) CheckCompliance(ctx context.Context, module *core.Module) (*ComplianceResult, error) {
	result := &ComplianceResult{
		Framework:  "NIST",
		Version:    "800-53",
		Compliant:  true,
		Violations: make([]ComplianceViolation, 0),
	}
	
	for _, resource := range module.Spec.Resources {
		violations := n.checkResource(&resource)
		result.Violations = append(result.Violations, violations...)
		
		if len(violations) > 0 {
			result.Failed++
		} else {
			result.Passed++
		}
		result.Total++
	}
	
	result.Compliant = len(result.Violations) == 0
	return result, nil
}

// checkResource checks a single resource against NIST controls
func (n *NIST80053Module) checkResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	switch resource.Type {
	case "service":
		violations = append(violations, n.checkServiceResource(resource)...)
	case "file":
		violations = append(violations, n.checkFileResource(resource)...)
	}
	
	return violations
}

// checkServiceResource checks service resources against NIST controls
func (n *NIST80053Module) checkServiceResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	// NIST AU-2 - Audit Events
	if resource.Name == "auditd" {
		enabled, _ := resource.Properties["enabled"].(bool)
		if !enabled {
			violations = append(violations, ComplianceViolation{
				Framework:   "NIST",
				Version:     "800-53",
				Control:     "AU-2",
				Title:       "Audit Events",
				Description: "Audit service must be enabled for compliance",
				Severity:    SeverityHigh,
				Resource:    resource.ResourceID(),
				Message:     "Audit service is not enabled",
			})
		}
	}
	
	return violations
}

// checkFileResource checks file resources against NIST controls
func (n *NIST80053Module) checkFileResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	path, _ := resource.Properties["path"].(string)
	
	// NIST AU-4 - Audit Storage Capacity
	if strings.Contains(path, "audit") {
		// Simplified check - real NIST would verify audit configuration
	}
	
	return violations
}

// STIGRHELModule implements STIG RHEL 8 compliance checks
type STIGRHELModule struct{}

// NewSTIGRHEL8Module creates a new STIG RHEL 8 compliance module
func NewSTIGRHEL8Module() *STIGRHELModule {
	return &STIGRHELModule{}
}

// Framework returns the compliance framework name
func (s *STIGRHELModule) Framework() string {
	return "STIG"
}

// Version returns the compliance framework version
func (s *STIGRHELModule) Version() string {
	return "RHEL 8"
}

// CheckCompliance checks STIG RHEL 8 compliance
func (s *STIGRHELModule) CheckCompliance(ctx context.Context, module *core.Module) (*ComplianceResult, error) {
	result := &ComplianceResult{
		Framework:  "STIG",
		Version:    "RHEL 8",
		Compliant:  true,
		Violations: make([]ComplianceViolation, 0),
	}
	
	for _, resource := range module.Spec.Resources {
		violations := s.checkResource(&resource)
		result.Violations = append(result.Violations, violations...)
		
		if len(violations) > 0 {
			result.Failed++
		} else {
			result.Passed++
		}
		result.Total++
	}
	
	result.Compliant = len(result.Violations) == 0
	return result, nil
}

// checkResource checks a single resource against STIG controls
func (s *STIGRHELModule) checkResource(resource *types.Resource) []ComplianceViolation {
	var violations []ComplianceViolation
	
	// STIG checks would be implemented here
	// This is a placeholder for the actual STIG implementation
	
	return violations
}

// ComplianceConfig represents compliance configuration
type ComplianceConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Modules []string `yaml:"modules" json:"modules"`
}

// DefaultComplianceConfig returns default compliance configuration
func DefaultComplianceConfig() *ComplianceConfig {
	return &ComplianceConfig{
		Enabled: true,
		Modules: []string{},
	}
}
