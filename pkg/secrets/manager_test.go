package secrets

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestSecretsManager_New(t *testing.T) {
	manager := NewSecretsManager()
	
	if manager == nil {
		t.Fatal("Expected non-nil secrets manager")
	}
	
	if !manager.IsEnabled() {
		t.Error("Expected secrets manager to be enabled by default")
	}
}

func TestSecretsManager_RegisterProvider(t *testing.T) {
	manager := NewSecretsManager()
	
	provider := &MockSecretsProvider{
		providerType: "mock",
	}
	
	err := manager.RegisterProvider(provider)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}
	
	providers := manager.GetProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providers))
	}
	
	if providers[0] != "mock" {
		t.Errorf("Expected provider type 'mock', got '%s'", providers[0])
	}
}

func TestSecretsManager_GetSecret(t *testing.T) {
	manager := NewSecretsManager()
	
	provider := &MockSecretsProvider{
		providerType: "mock",
		secrets: map[string]string{
			"database/password": "secret123",
			"api/key":          "abc123def456",
		},
	}
	
	manager.RegisterProvider(provider)
	
	tests := []struct {
		name        string
		secretPath  string
		expected    string
		expectError bool
	}{
		{
			name:        "existing secret",
			secretPath:  "mock://database/password",
			expected:    "secret123",
			expectError: false,
		},
		{
			name:        "another existing secret",
			secretPath:  "mock://api/key",
			expected:    "abc123def456",
			expectError: false,
		},
		{
			name:        "non-existent secret",
			secretPath:  "mock://nonexistent",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid provider",
			secretPath:  "invalid://database/password",
			expected:    "",
			expectError: true,
		},
		{
			name:        "malformed path",
			secretPath:  "invalid-path",
			expected:    "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			secret, err := manager.GetSecret(ctx, tt.secretPath)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if secret.Value != tt.expected {
				t.Errorf("Expected secret value '%s', got '%s'", tt.expected, secret.Value)
			}
			
			if secret.Path != tt.secretPath {
				t.Errorf("Expected secret path '%s', got '%s'", tt.secretPath, secret.Path)
			}
		})
	}
}

func TestSecretsManager_SetSecret(t *testing.T) {
	manager := NewSecretsManager()
	
	provider := &MockSecretsProvider{
		providerType: "mock",
		secrets:      make(map[string]string),
	}
	
	manager.RegisterProvider(provider)
	
	ctx := context.Background()
	secretPath := "mock://test/secret"
	secretValue := "test-value"
	
	err := manager.SetSecret(ctx, secretPath, secretValue)
	if err != nil {
		t.Fatalf("Failed to set secret: %v", err)
	}
	
	// Verify secret was set
	secret, err := manager.GetSecret(ctx, secretPath)
	if err != nil {
		t.Fatalf("Failed to get secret after setting: %v", err)
	}
	
	if secret.Value != secretValue {
		t.Errorf("Expected secret value '%s', got '%s'", secretValue, secret.Value)
	}
}

func TestSecretsManager_DeleteSecret(t *testing.T) {
	manager := NewSecretsManager()
	
	provider := &MockSecretsProvider{
		providerType: "mock",
		secrets: map[string]string{
			"test/secret": "test-value",
		},
	}
	
	manager.RegisterProvider(provider)
	
	ctx := context.Background()
	secretPath := "mock://test/secret"
	
	// Verify secret exists
	_, err := manager.GetSecret(ctx, secretPath)
	if err != nil {
		t.Fatalf("Secret should exist before deletion: %v", err)
	}
	
	// Delete secret
	err = manager.DeleteSecret(ctx, secretPath)
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}
	
	// Verify secret no longer exists
	_, err = manager.GetSecret(ctx, secretPath)
	if err == nil {
		t.Error("Expected error when getting deleted secret")
	}
}

func TestSecretsManager_EnableDisable(t *testing.T) {
	manager := NewSecretsManager()
	
	if !manager.IsEnabled() {
		t.Error("Expected secrets manager to be enabled by default")
	}
	
	manager.Disable()
	if manager.IsEnabled() {
		t.Error("Expected secrets manager to be disabled after Disable()")
	}
	
	// When disabled, getting secrets should return error
	ctx := context.Background()
	_, err := manager.GetSecret(ctx, "mock://test/secret")
	if err == nil {
		t.Error("Expected error when secrets manager is disabled")
	}
	
	manager.Enable()
	if !manager.IsEnabled() {
		t.Error("Expected secrets manager to be enabled after Enable()")
	}
}

func TestSecretsManager_ResolveSecrets(t *testing.T) {
	manager := NewSecretsManager()
	
	provider := &MockSecretsProvider{
		providerType: "mock",
		secrets: map[string]string{
			"database/password": "secret123",
			"api/key":          "abc123def456",
		},
	}
	
	manager.RegisterProvider(provider)
	
	// Test resolving secrets in a map
	input := map[string]interface{}{
		"database_url": "postgres://user:${secret:mock://database/password}@localhost/db",
		"api_key":      "${secret:mock://api/key}",
		"static_value": "no-secret-here",
		"nested": map[string]interface{}{
			"secret_key": "${secret:mock://database/password}",
		},
	}
	
	ctx := context.Background()
	resolved, err := manager.ResolveSecrets(ctx, input)
	if err != nil {
		t.Fatalf("Failed to resolve secrets: %v", err)
	}
	
	// Convert to map for easier access
	resolvedMap, ok := resolved.(map[string]interface{})
	if !ok {
		t.Fatal("Expected resolved to be a map")
	}
	
	// Verify secrets were resolved
	expectedURL := "postgres://user:secret123@localhost/db"
	if resolvedMap["database_url"] != expectedURL {
		t.Errorf("Expected database_url '%s', got '%s'", expectedURL, resolvedMap["database_url"])
	}
	
	if resolvedMap["api_key"] != "abc123def456" {
		t.Errorf("Expected api_key 'abc123def456', got '%s'", resolvedMap["api_key"])
	}
	
	if resolvedMap["static_value"] != "no-secret-here" {
		t.Errorf("Expected static_value to remain unchanged")
	}
	
	// Check nested resolution
	nested, ok := resolvedMap["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected nested to be a map")
	}
	
	if nested["secret_key"] != "secret123" {
		t.Errorf("Expected nested secret_key 'secret123', got '%s'", nested["secret_key"])
	}
}

func TestSecretsManager_ListSecrets(t *testing.T) {
	manager := NewSecretsManager()
	
	provider := &MockSecretsProvider{
		providerType: "mock",
		secrets: map[string]string{
			"database/password": "secret123",
			"database/username": "admin",
			"api/key":          "abc123def456",
		},
	}
	
	manager.RegisterProvider(provider)
	
	ctx := context.Background()
	
	// List all secrets
	secrets, err := manager.ListSecrets(ctx, "mock://")
	if err != nil {
		t.Fatalf("Failed to list secrets: %v", err)
	}
	
	if len(secrets) != 3 {
		t.Errorf("Expected 3 secrets, got %d", len(secrets))
	}
	
	// List secrets with prefix
	secrets, err = manager.ListSecrets(ctx, "mock://database/")
	if err != nil {
		t.Fatalf("Failed to list secrets with prefix: %v", err)
	}
	
	if len(secrets) != 2 {
		t.Errorf("Expected 2 secrets with database prefix, got %d", len(secrets))
	}
}

func TestParseSecretPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedType string
		expectedPath string
		expectError  bool
	}{
		{
			name:         "valid path",
			path:         "vault://secret/database/password",
			expectedType: "vault",
			expectedPath: "secret/database/password",
			expectError:  false,
		},
		{
			name:         "aws secrets manager path",
			path:         "aws://prod/database/credentials",
			expectedType: "aws",
			expectedPath: "prod/database/credentials",
			expectError:  false,
		},
		{
			name:         "invalid format - no protocol",
			path:         "secret/database/password",
			expectedType: "",
			expectedPath: "",
			expectError:  true,
		},
		{
			name:         "empty path after provider (allowed for listing)",
			path:         "vault://",
			expectedType: "vault",
			expectedPath: "",
			expectError:  false,
		},
		{
			name:         "empty string",
			path:         "",
			expectedType: "",
			expectedPath: "",
			expectError:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerType, secretPath, err := parseSecretPath(tt.path)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if providerType != tt.expectedType {
				t.Errorf("Expected provider type '%s', got '%s'", tt.expectedType, providerType)
			}
			
			if secretPath != tt.expectedPath {
				t.Errorf("Expected secret path '%s', got '%s'", tt.expectedPath, secretPath)
			}
		})
	}
}

// MockSecretsProvider for testing
type MockSecretsProvider struct {
	providerType string
	secrets      map[string]string
}

func (m *MockSecretsProvider) Type() string {
	return m.providerType
}

func (m *MockSecretsProvider) GetSecret(ctx context.Context, path string) (*Secret, error) {
	value, exists := m.secrets[path]
	if !exists {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	
	return &Secret{
		Path:  fmt.Sprintf("%s://%s", m.providerType, path),
		Value: value,
	}, nil
}

func (m *MockSecretsProvider) SetSecret(ctx context.Context, path, value string) error {
	if m.secrets == nil {
		m.secrets = make(map[string]string)
	}
	m.secrets[path] = value
	return nil
}

func (m *MockSecretsProvider) DeleteSecret(ctx context.Context, path string) error {
	if _, exists := m.secrets[path]; !exists {
		return fmt.Errorf("secret not found: %s", path)
	}
	delete(m.secrets, path)
	return nil
}

func (m *MockSecretsProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	var secrets []string
	for path := range m.secrets {
		if prefix == "" || strings.HasPrefix(path, prefix) {
			secrets = append(secrets, path)
		}
	}
	return secrets, nil
}
