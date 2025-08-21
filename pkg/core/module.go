package core

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ataiva-software/chisel/pkg/types"
	"gopkg.in/yaml.v3"
)

// Module represents a Chisel configuration module
type Module struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   ModuleMetadata `yaml:"metadata"`
	Spec       ModuleSpec     `yaml:"spec"`
}

// ModuleMetadata contains metadata about the module
type ModuleMetadata struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Description string            `yaml:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

// ModuleSpec contains the module specification
type ModuleSpec struct {
	Resources []types.Resource `yaml:"resources"`
}

// Validate validates the module configuration
func (m *Module) Validate() error {
	// Validate apiVersion
	if m.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if m.APIVersion != "ataiva.com/chisel/v1" {
		return fmt.Errorf("apiVersion must be ataiva.com/chisel/v1")
	}

	// Validate kind
	if m.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if m.Kind != "Module" {
		return fmt.Errorf("kind must be Module")
	}

	// Validate metadata
	if m.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if m.Metadata.Version == "" {
		return fmt.Errorf("metadata.version is required")
	}

	// Validate version format (basic semver check)
	semverRegex := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	if !semverRegex.MatchString(m.Metadata.Version) {
		return fmt.Errorf("metadata.version must be valid semver")
	}

	// Validate resources
	for i, resource := range m.Spec.Resources {
		if err := resource.Validate(); err != nil {
			return fmt.Errorf("resource[%d]: %w", i, err)
		}
	}

	return nil
}

// LoadModuleFromFile loads a module from a YAML file
func LoadModuleFromFile(filename string) (*Module, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read module file %s: %w", filename, err)
	}

	var module Module
	if err := yaml.Unmarshal(data, &module); err != nil {
		return nil, fmt.Errorf("failed to parse module file %s: %w", filename, err)
	}

	if err := module.Validate(); err != nil {
		return nil, fmt.Errorf("invalid module in file %s: %w", filename, err)
	}

	return &module, nil
}

// SaveModuleToFile saves a module to a YAML file
func (m *Module) SaveToFile(filename string) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid module: %w", err)
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal module: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write module file %s: %w", filename, err)
	}

	return nil
}
