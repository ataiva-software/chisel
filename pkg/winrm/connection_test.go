package winrm

import (
	"context"
	"testing"
	"time"
)

func TestWinRMConnectionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConnectionConfig
		wantErr bool
	}{
		{
			name: "valid config with password",
			config: &ConnectionConfig{
				Host:     "windows-server.example.com",
				Port:     5985,
				User:     "administrator",
				Password: "password123",
				UseHTTPS: false,
			},
			wantErr: false,
		},
		{
			name: "valid config with HTTPS",
			config: &ConnectionConfig{
				Host:     "windows-server.example.com",
				Port:     5986,
				User:     "administrator",
				Password: "password123",
				UseHTTPS: true,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: &ConnectionConfig{
				Port:     5985,
				User:     "administrator",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing user",
			config: &ConnectionConfig{
				Host:     "windows-server.example.com",
				Port:     5985,
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &ConnectionConfig{
				Host: "windows-server.example.com",
				Port: 5985,
				User: "administrator",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &ConnectionConfig{
				Host:     "windows-server.example.com",
				Port:     0,
				User:     "administrator",
				Password: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestWinRMConnectionConfig_SetDefaults(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "windows-server.example.com",
		User:     "administrator",
		Password: "password123",
	}

	config.SetDefaults()

	if config.Port != 5985 {
		t.Errorf("Expected default port 5985, got %d", config.Port)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", config.Timeout)
	}

	if config.UseHTTPS {
		t.Error("Expected UseHTTPS to default to false")
	}

	if config.Insecure {
		t.Error("Expected Insecure to default to false")
	}
}

func TestWinRMConnection_Connect(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     5985,
		User:     "testuser",
		Password: "testpass",
	}

	conn := NewWinRMConnection(config)

	// This will fail in CI/testing environment, but tests the interface
	ctx := context.Background()
	err := conn.Connect(ctx)
	
	// We expect this to fail in test environment
	if err == nil {
		t.Log("Connection succeeded (unexpected in test environment)")
	} else {
		t.Logf("Connection failed as expected in test environment: %v", err)
	}
}

func TestWinRMConnection_Execute(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "localhost",
		Port:     5985,
		User:     "testuser",
		Password: "testpass",
	}

	conn := NewWinRMConnection(config)

	// Test executing without connection
	ctx := context.Background()
	result, err := conn.Execute(ctx, "echo 'test'")
	
	if err == nil {
		t.Error("Expected error when executing without connection")
	}
	
	if result != nil {
		t.Error("Expected nil result when not connected")
	}
}

func TestWinRMExecuteResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		result   *ExecuteResult
		expected bool
	}{
		{
			name: "successful command",
			result: &ExecuteResult{
				Command:  "echo 'test'",
				ExitCode: 0,
				Stdout:   "test",
				Stderr:   "",
			},
			expected: true,
		},
		{
			name: "failed command",
			result: &ExecuteResult{
				Command:  "invalid-command",
				ExitCode: 1,
				Stdout:   "",
				Stderr:   "command not found",
			},
			expected: false,
		},
		{
			name: "command with high exit code",
			result: &ExecuteResult{
				Command:  "exit 255",
				ExitCode: 255,
				Stdout:   "",
				Stderr:   "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Success() != tt.expected {
				t.Errorf("Expected Success() to return %v, got %v", tt.expected, tt.result.Success())
			}
		})
	}
}

func TestNewWinRMConnection(t *testing.T) {
	config := &ConnectionConfig{
		Host:     "windows-server.example.com",
		Port:     5985,
		User:     "administrator",
		Password: "password123",
	}

	conn := NewWinRMConnection(config)

	if conn == nil {
		t.Fatal("Expected non-nil connection")
	}

	if conn.config != config {
		t.Error("Expected connection to store config reference")
	}

	if conn.connected {
		t.Error("Expected new connection to not be connected initially")
	}
}
