package ssh

import (
	"context"
	"testing"
	"time"
)

func TestConnectionConfig_SetDefaults(t *testing.T) {
	config := &ConnectionConfig{
		Host: "example.com",
		User: "testuser",
	}

	config.SetDefaults()

	if config.Port != 22 {
		t.Errorf("Expected default port 22, got %d", config.Port)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", config.Timeout)
	}
	if config.ConnectTimeout != 10*time.Second {
		t.Errorf("Expected default connect timeout 10s, got %v", config.ConnectTimeout)
	}
	if config.KeepAlive != 30*time.Second {
		t.Errorf("Expected default keep alive 30s, got %v", config.KeepAlive)
	}
	if config.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", config.MaxRetries)
	}
}

func TestConnectionConfig_SetDefaultsDoesNotOverride(t *testing.T) {
	config := &ConnectionConfig{
		Host:           "example.com",
		User:           "testuser",
		Port:           2222,
		Timeout:        60 * time.Second,
		ConnectTimeout: 20 * time.Second,
		KeepAlive:      60 * time.Second,
		MaxRetries:     5,
	}

	config.SetDefaults()

	if config.Port != 2222 {
		t.Errorf("Expected port 2222, got %d", config.Port)
	}
	if config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", config.Timeout)
	}
	if config.ConnectTimeout != 20*time.Second {
		t.Errorf("Expected connect timeout 20s, got %v", config.ConnectTimeout)
	}
	if config.KeepAlive != 60*time.Second {
		t.Errorf("Expected keep alive 60s, got %v", config.KeepAlive)
	}
	if config.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", config.MaxRetries)
	}
}

func TestConnectionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ConnectionConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with password",
			config: ConnectionConfig{
				Host:     "example.com",
				Port:     22,
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "valid config with private key",
			config: ConnectionConfig{
				Host:       "example.com",
				Port:       22,
				User:       "testuser",
				PrivateKey: "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----",
			},
			wantErr: false,
		},
		{
			name: "valid config with private key path",
			config: ConnectionConfig{
				Host:           "example.com",
				Port:           22,
				User:           "testuser",
				PrivateKeyPath: "~/.ssh/id_rsa",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: ConnectionConfig{
				Port:     22,
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: true,
			errMsg:  "host cannot be empty",
		},
		{
			name: "missing user",
			config: ConnectionConfig{
				Host:     "example.com",
				Port:     22,
				Password: "testpass",
			},
			wantErr: true,
			errMsg:  "user cannot be empty",
		},
		{
			name: "missing authentication",
			config: ConnectionConfig{
				Host: "example.com",
				Port: 22,
				User: "testuser",
			},
			wantErr: true,
			errMsg:  "must provide either password, private_key_path, or private_key",
		},
		{
			name: "invalid port - zero",
			config: ConnectionConfig{
				Host:     "example.com",
				Port:     0,
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "invalid port - negative",
			config: ConnectionConfig{
				Host:     "example.com",
				Port:     -1,
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: ConnectionConfig{
				Host:     "example.com",
				Port:     65536,
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ConnectionConfig.Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ConnectionConfig.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ConnectionConfig.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestNewConnection(t *testing.T) {
	config := &ConnectionConfig{
		Host: "example.com",
		User: "testuser",
	}

	conn := NewConnection(config)

	if conn == nil {
		t.Fatal("NewConnection() returned nil")
	}

	if conn.config == nil {
		t.Error("Connection config is nil")
	}

	// Check that defaults were set
	if conn.config.Port != 22 {
		t.Errorf("Expected default port 22, got %d", conn.config.Port)
	}
}

func TestExecuteResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		result   ExecuteResult
		expected bool
	}{
		{
			name: "successful command",
			result: ExecuteResult{
				ExitCode: 0,
			},
			expected: true,
		},
		{
			name: "failed command",
			result: ExecuteResult{
				ExitCode: 1,
			},
			expected: false,
		},
		{
			name: "command with high exit code",
			result: ExecuteResult{
				ExitCode: 127,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Success(); got != tt.expected {
				t.Errorf("ExecuteResult.Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConnection_ConnectWithInvalidConfig(t *testing.T) {
	config := &ConnectionConfig{
		Host: "", // Invalid: empty host
		User: "testuser",
	}

	conn := NewConnection(config)
	ctx := context.Background()

	err := conn.Connect(ctx)
	if err == nil {
		t.Error("Expected error when connecting with invalid config")
	}

	expectedErr := "invalid connection config: host cannot be empty"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestConnection_ExecuteWithoutConnection(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "example.com",
		User:     "testuser",
		Password: "testpass",
	}

	conn := NewConnection(config)
	ctx := context.Background()

	// Try to execute without connecting
	result, err := conn.Execute(ctx, "echo test")
	if err == nil {
		t.Error("Expected error when executing without connection")
	}

	if result != nil {
		t.Error("Expected nil result when executing without connection")
	}

	expectedErr := "not connected"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}
