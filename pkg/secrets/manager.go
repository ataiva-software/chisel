package secrets

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Secret represents a secret value
type Secret struct {
	Path      string            `json:"path"`
	Value     string            `json:"value"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

// SecretsProvider defines the interface for secret providers
type SecretsProvider interface {
	Type() string
	GetSecret(ctx context.Context, path string) (*Secret, error)
	SetSecret(ctx context.Context, path, value string) error
	DeleteSecret(ctx context.Context, path string) error
	ListSecrets(ctx context.Context, prefix string) ([]string, error)
}

// SecretsManager manages multiple secret providers
type SecretsManager struct {
	providers map[string]SecretsProvider
	enabled   bool
	mu        sync.RWMutex
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager() *SecretsManager {
	return &SecretsManager{
		providers: make(map[string]SecretsProvider),
		enabled:   true,
	}
}

// IsEnabled returns whether the secrets manager is enabled
func (m *SecretsManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// Enable enables the secrets manager
func (m *SecretsManager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables the secrets manager
func (m *SecretsManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// RegisterProvider registers a secrets provider
func (m *SecretsManager) RegisterProvider(provider SecretsProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}
	
	providerType := provider.Type()
	if providerType == "" {
		return fmt.Errorf("provider type cannot be empty")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.providers[providerType] = provider
	return nil
}

// GetProviders returns the list of registered provider types
func (m *SecretsManager) GetProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	types := make([]string, 0, len(m.providers))
	for providerType := range m.providers {
		types = append(types, providerType)
	}
	return types
}

// GetSecret retrieves a secret by path
func (m *SecretsManager) GetSecret(ctx context.Context, secretPath string) (*Secret, error) {
	if !m.IsEnabled() {
		return nil, fmt.Errorf("secrets manager is disabled")
	}
	
	providerType, path, err := parseSecretPath(secretPath)
	if err != nil {
		return nil, fmt.Errorf("invalid secret path: %w", err)
	}
	
	m.mu.RLock()
	provider, exists := m.providers[providerType]
	m.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no provider registered for type: %s", providerType)
	}
	
	return provider.GetSecret(ctx, path)
}

// SetSecret sets a secret value
func (m *SecretsManager) SetSecret(ctx context.Context, secretPath, value string) error {
	if !m.IsEnabled() {
		return fmt.Errorf("secrets manager is disabled")
	}
	
	providerType, path, err := parseSecretPath(secretPath)
	if err != nil {
		return fmt.Errorf("invalid secret path: %w", err)
	}
	
	m.mu.RLock()
	provider, exists := m.providers[providerType]
	m.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("no provider registered for type: %s", providerType)
	}
	
	return provider.SetSecret(ctx, path, value)
}

// DeleteSecret deletes a secret
func (m *SecretsManager) DeleteSecret(ctx context.Context, secretPath string) error {
	if !m.IsEnabled() {
		return fmt.Errorf("secrets manager is disabled")
	}
	
	providerType, path, err := parseSecretPath(secretPath)
	if err != nil {
		return fmt.Errorf("invalid secret path: %w", err)
	}
	
	m.mu.RLock()
	provider, exists := m.providers[providerType]
	m.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("no provider registered for type: %s", providerType)
	}
	
	return provider.DeleteSecret(ctx, path)
}

// ListSecrets lists secrets with optional prefix
func (m *SecretsManager) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	if !m.IsEnabled() {
		return nil, fmt.Errorf("secrets manager is disabled")
	}
	
	providerType, prefix, err := parseSecretPath(pathPrefix)
	if err != nil {
		return nil, fmt.Errorf("invalid path prefix: %w", err)
	}
	
	m.mu.RLock()
	provider, exists := m.providers[providerType]
	m.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no provider registered for type: %s", providerType)
	}
	
	return provider.ListSecrets(ctx, prefix)
}

// ResolveSecrets resolves secret references in a data structure
func (m *SecretsManager) ResolveSecrets(ctx context.Context, data interface{}) (interface{}, error) {
	if !m.IsEnabled() {
		return data, nil
	}
	
	return m.resolveValue(ctx, data)
}

// resolveValue recursively resolves secrets in any value
func (m *SecretsManager) resolveValue(ctx context.Context, value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return m.resolveString(ctx, v)
	case map[string]interface{}:
		return m.resolveMap(ctx, v)
	case []interface{}:
		return m.resolveSlice(ctx, v)
	default:
		return value, nil
	}
}

// resolveString resolves secret references in a string
func (m *SecretsManager) resolveString(ctx context.Context, s string) (string, error) {
	// Pattern to match ${secret:provider://path}
	secretPattern := regexp.MustCompile(`\$\{secret:([^}]+)\}`)
	
	result := s
	matches := secretPattern.FindAllStringSubmatch(s, -1)
	
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		
		placeholder := match[0]
		secretPath := match[1]
		
		secret, err := m.GetSecret(ctx, secretPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve secret %s: %w", secretPath, err)
		}
		
		result = strings.ReplaceAll(result, placeholder, secret.Value)
	}
	
	return result, nil
}

// resolveMap resolves secrets in a map
func (m *SecretsManager) resolveMap(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for key, value := range data {
		resolvedValue, err := m.resolveValue(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve value for key %s: %w", key, err)
		}
		result[key] = resolvedValue
	}
	
	return result, nil
}

// resolveSlice resolves secrets in a slice
func (m *SecretsManager) resolveSlice(ctx context.Context, data []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(data))
	
	for i, value := range data {
		resolvedValue, err := m.resolveValue(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve value at index %d: %w", i, err)
		}
		result[i] = resolvedValue
	}
	
	return result, nil
}

// parseSecretPath parses a secret path in the format "provider://path"
func parseSecretPath(secretPath string) (string, string, error) {
	if secretPath == "" {
		return "", "", fmt.Errorf("secret path cannot be empty")
	}
	
	parts := strings.SplitN(secretPath, "://", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid secret path format, expected 'provider://path'")
	}
	
	providerType := parts[0]
	path := parts[1]
	
	if providerType == "" {
		return "", "", fmt.Errorf("provider type cannot be empty")
	}
	
	// Allow empty path for listing operations
	return providerType, path, nil
}

// VaultProvider implements HashiCorp Vault integration
type VaultProvider struct {
	address string
	token   string
	// In a real implementation, this would use the Vault API client
}

// NewVaultProvider creates a new Vault secrets provider
func NewVaultProvider(address, token string) *VaultProvider {
	return &VaultProvider{
		address: address,
		token:   token,
	}
}

// Type returns the provider type
func (v *VaultProvider) Type() string {
	return "vault"
}

// GetSecret retrieves a secret from Vault
func (v *VaultProvider) GetSecret(ctx context.Context, path string) (*Secret, error) {
	// In a real implementation, this would use the Vault API
	return nil, fmt.Errorf("Vault integration not implemented in test environment")
}

// SetSecret sets a secret in Vault
func (v *VaultProvider) SetSecret(ctx context.Context, path, value string) error {
	// In a real implementation, this would use the Vault API
	return fmt.Errorf("Vault integration not implemented in test environment")
}

// DeleteSecret deletes a secret from Vault
func (v *VaultProvider) DeleteSecret(ctx context.Context, path string) error {
	// In a real implementation, this would use the Vault API
	return fmt.Errorf("Vault integration not implemented in test environment")
}

// ListSecrets lists secrets from Vault
func (v *VaultProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	// In a real implementation, this would use the Vault API
	return nil, fmt.Errorf("Vault integration not implemented in test environment")
}

// AWSSecretsProvider implements AWS Secrets Manager integration
type AWSSecretsProvider struct {
	region string
	// In a real implementation, this would use the AWS SDK
}

// NewAWSSecretsProvider creates a new AWS Secrets Manager provider
func NewAWSSecretsProvider(region string) *AWSSecretsProvider {
	return &AWSSecretsProvider{
		region: region,
	}
}

// Type returns the provider type
func (a *AWSSecretsProvider) Type() string {
	return "aws"
}

// GetSecret retrieves a secret from AWS Secrets Manager
func (a *AWSSecretsProvider) GetSecret(ctx context.Context, path string) (*Secret, error) {
	// In a real implementation, this would use the AWS SDK
	return nil, fmt.Errorf("AWS Secrets Manager integration not implemented in test environment")
}

// SetSecret sets a secret in AWS Secrets Manager
func (a *AWSSecretsProvider) SetSecret(ctx context.Context, path, value string) error {
	// In a real implementation, this would use the AWS SDK
	return fmt.Errorf("AWS Secrets Manager integration not implemented in test environment")
}

// DeleteSecret deletes a secret from AWS Secrets Manager
func (a *AWSSecretsProvider) DeleteSecret(ctx context.Context, path string) error {
	// In a real implementation, this would use the AWS SDK
	return fmt.Errorf("AWS Secrets Manager integration not implemented in test environment")
}

// ListSecrets lists secrets from AWS Secrets Manager
func (a *AWSSecretsProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	// In a real implementation, this would use the AWS SDK
	return nil, fmt.Errorf("AWS Secrets Manager integration not implemented in test environment")
}

// SecretsConfig represents secrets management configuration
type SecretsConfig struct {
	Enabled   bool                   `yaml:"enabled" json:"enabled"`
	Providers map[string]interface{} `yaml:"providers" json:"providers"`
}

// DefaultSecretsConfig returns default secrets configuration
func DefaultSecretsConfig() *SecretsConfig {
	return &SecretsConfig{
		Enabled:   true,
		Providers: make(map[string]interface{}),
	}
}
