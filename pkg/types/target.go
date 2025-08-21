package types

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Target represents a target host for configuration management
type Target struct {
	// Host is the hostname or IP address
	Host string `yaml:"host" json:"host"`
	
	// Port is the SSH port (default: 22)
	Port int `yaml:"port,omitempty" json:"port,omitempty"`
	
	// User is the SSH username
	User string `yaml:"user,omitempty" json:"user,omitempty"`
	
	// KeyFile is the path to the SSH private key
	KeyFile string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	
	// Password is the SSH password (not recommended)
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	
	// Labels are key-value pairs for target selection
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	
	// Facts are discovered properties about the target
	Facts map[string]interface{} `yaml:"facts,omitempty" json:"facts,omitempty"`
	
	// Connection configuration
	Connection *ConnectionConfig `yaml:"connection,omitempty" json:"connection,omitempty"`
}

// ConnectionConfig represents connection configuration for a target
type ConnectionConfig struct {
	Type     string `yaml:"type" json:"type"`         // ssh, winrm, local
	User     string `yaml:"user,omitempty" json:"user,omitempty"`
	KeyFile  string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	Port     int    `yaml:"port,omitempty" json:"port,omitempty"`
	Timeout  int    `yaml:"timeout,omitempty" json:"timeout,omitempty"` // seconds
}

// TargetGroup represents a group of targets
type TargetGroup struct {
	Name        string            `yaml:"name" json:"name"`
	Selector    string            `yaml:"selector,omitempty" json:"selector,omitempty"`
	Hosts       []string          `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	Connection  *ConnectionConfig `yaml:"connection,omitempty" json:"connection,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// Inventory represents the complete inventory configuration
type Inventory struct {
	Targets []Target      `yaml:"targets,omitempty" json:"targets,omitempty"`
	Groups  []TargetGroup `yaml:"groups,omitempty" json:"groups,omitempty"`
	
	// Dynamic inventory configuration
	Dynamic map[string]interface{} `yaml:"dynamic,omitempty" json:"dynamic,omitempty"`
}

// Validate validates the target configuration
func (t *Target) Validate() error {
	if t.Host == "" {
		return fmt.Errorf("target host cannot be empty")
	}
	
	// Validate host format (IP or hostname)
	if net.ParseIP(t.Host) == nil {
		// Not an IP, validate as hostname
		if !isValidHostname(t.Host) {
			return fmt.Errorf("invalid hostname: %s", t.Host)
		}
	}
	
	// Validate port
	if t.Port < 0 || t.Port > 65535 {
		return fmt.Errorf("invalid port: %d", t.Port)
	}
	
	// Set default port if not specified
	if t.Port == 0 {
		t.Port = 22
	}
	
	return nil
}

// Address returns the target address in host:port format
func (t *Target) Address() string {
	if t.Port == 0 || t.Port == 22 {
		return t.Host
	}
	return net.JoinHostPort(t.Host, strconv.Itoa(t.Port))
}

// GetUser returns the effective username for the target
func (t *Target) GetUser() string {
	if t.Connection != nil && t.Connection.User != "" {
		return t.Connection.User
	}
	if t.User != "" {
		return t.User
	}
	return "root" // default
}

// GetKeyFile returns the effective SSH key file for the target
func (t *Target) GetKeyFile() string {
	if t.Connection != nil && t.Connection.KeyFile != "" {
		return t.Connection.KeyFile
	}
	return t.KeyFile
}

// GetPassword returns the effective password for the target
func (t *Target) GetPassword() string {
	if t.Connection != nil && t.Connection.Password != "" {
		return t.Connection.Password
	}
	return t.Password
}

// GetPort returns the effective port for the target
func (t *Target) GetPort() int {
	if t.Connection != nil && t.Connection.Port > 0 {
		return t.Connection.Port
	}
	if t.Port > 0 {
		return t.Port
	}
	return 22 // default SSH port
}

// HasLabel checks if the target has a specific label with the given value
func (t *Target) HasLabel(key, value string) bool {
	if t.Labels == nil {
		return false
	}
	labelValue, exists := t.Labels[key]
	return exists && labelValue == value
}

// AddLabel adds or updates a label on the target
func (t *Target) AddLabel(key, value string) {
	if t.Labels == nil {
		t.Labels = make(map[string]string)
	}
	t.Labels[key] = value
}

// RemoveLabel removes a label from the target
func (t *Target) RemoveLabel(key string) {
	if t.Labels != nil {
		delete(t.Labels, key)
	}
}

// isValidHostname validates a hostname according to RFC standards
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}
	
	// Remove trailing dot if present
	if hostname[len(hostname)-1] == '.' {
		hostname = hostname[:len(hostname)-1]
	}
	
	// Split into labels
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		
		// Check label format
		for i, r := range label {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
				return false
			}
			// First and last character cannot be hyphen
			if (i == 0 || i == len(label)-1) && r == '-' {
				return false
			}
		}
	}
	
	return true
}
