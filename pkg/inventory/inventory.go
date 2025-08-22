package inventory

import (
	"fmt"
	"os"

	"github.com/ataiva-software/forge/pkg/ssh"
	"gopkg.in/yaml.v3"
)

// Inventory represents a Chisel inventory configuration
type Inventory struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Targets    map[string]TargetGroup `yaml:"targets"`
}

// TargetGroup represents a group of target hosts
type TargetGroup struct {
	Hosts      []string              `yaml:"hosts,omitempty"`
	Selector   string                `yaml:"selector,omitempty"`
	Connection ssh.ConnectionConfig  `yaml:"connection"`
}

// Validate validates the inventory configuration
func (i *Inventory) Validate() error {
	// Validate apiVersion
	if i.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if i.APIVersion != "ataiva.com/chisel/v1" {
		return fmt.Errorf("apiVersion must be ataiva.com/chisel/v1")
	}

	// Validate kind
	if i.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if i.Kind != "Inventory" {
		return fmt.Errorf("kind must be Inventory")
	}

	// Validate targets
	if len(i.Targets) == 0 {
		return fmt.Errorf("at least one target group is required")
	}

	for name, group := range i.Targets {
		if err := group.Validate(name); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates a target group
func (tg *TargetGroup) Validate(name string) error {
	// Must have either hosts or selector, but not both
	hasHosts := len(tg.Hosts) > 0
	hasSelector := tg.Selector != ""

	if hasHosts && hasSelector {
		return fmt.Errorf("target group '%s': cannot specify both hosts and selector", name)
	}

	if !hasHosts && !hasSelector {
		return fmt.Errorf("target group '%s': must specify either hosts or selector", name)
	}

	// Validate connection config
	if err := tg.Connection.Validate(); err != nil {
		return fmt.Errorf("target group '%s': %w", name, err)
	}

	return nil
}

// GetHosts returns the list of hosts for this target group
func (tg *TargetGroup) GetHosts() ([]string, error) {
	if len(tg.Hosts) > 0 {
		return tg.Hosts, nil
	}

	if tg.Selector != "" {
		// TODO: Implement selector-based host discovery
		// For now, return empty list
		return []string{}, nil
	}

	return []string{}, fmt.Errorf("no hosts or selector specified")
}

// LoadInventoryFromFile loads an inventory from a YAML file
func LoadInventoryFromFile(filename string) (*Inventory, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory file %s: %w", filename, err)
	}

	var inventory Inventory
	if err := yaml.Unmarshal(data, &inventory); err != nil {
		return nil, fmt.Errorf("failed to parse inventory file %s: %w", filename, err)
	}

	if err := inventory.Validate(); err != nil {
		return nil, fmt.Errorf("invalid inventory in file %s: %w", filename, err)
	}

	return &inventory, nil
}

// SaveToFile saves an inventory to a YAML file
func (i *Inventory) SaveToFile(filename string) error {
	if err := i.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid inventory: %w", err)
	}

	data, err := yaml.Marshal(i)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write inventory file %s: %w", filename, err)
	}

	return nil
}
