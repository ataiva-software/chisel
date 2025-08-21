package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ataiva-software/chisel/pkg/types"
)

func TestAuditLogger_New(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewAuditLogger(logPath)
	
	if logger == nil {
		t.Fatal("Expected non-nil audit logger")
	}
	
	if !logger.IsEnabled() {
		t.Error("Expected audit logger to be enabled by default")
	}
}

func TestAuditLogger_LogResourceChange(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewAuditLogger(logPath)
	
	resource := &types.Resource{
		Type: "file",
		Name: "test-file",
		Properties: map[string]interface{}{
			"path":    "/tmp/test.txt",
			"content": "test content",
		},
	}
	
	diff := &types.ResourceDiff{
		ResourceID: "file.test-file",
		Action:     types.ActionCreate,
		Changes: map[string]interface{}{
			"content": map[string]interface{}{
				"from": nil,
				"to":   "test content",
			},
		},
	}
	
	ctx := context.Background()
	err = logger.LogResourceChange(ctx, resource, diff, true, nil)
	if err != nil {
		t.Fatalf("Failed to log resource change: %v", err)
	}
	
	// Verify log file was created and contains entry
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Expected audit log to contain entries")
	}
	
	// Parse the log entry
	var entry AuditEntry
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one log line")
	}
	
	err = json.Unmarshal([]byte(lines[0]), &entry)
	if err != nil {
		t.Fatalf("Failed to parse audit entry: %v", err)
	}
	
	// Verify entry fields
	if entry.EventType != EventTypeResourceChange {
		t.Errorf("Expected event type %s, got %s", EventTypeResourceChange, entry.EventType)
	}
	
	if entry.ResourceID != "file.test-file" {
		t.Errorf("Expected resource ID 'file.test-file', got '%s'", entry.ResourceID)
	}
	
	if entry.Action != string(types.ActionCreate) {
		t.Errorf("Expected action 'create', got '%s'", entry.Action)
	}
	
	if !entry.Success {
		t.Error("Expected success to be true")
	}
}

func TestAuditLogger_LogPolicyViolation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewAuditLogger(logPath)
	
	resource := &types.Resource{
		Type: "user",
		Name: "root",
	}
	
	violation := PolicyViolation{
		Policy:   "security",
		Rule:     "deny_root_user",
		Message:  "Root user creation is not allowed",
		Resource: "user.root",
	}
	
	ctx := context.Background()
	err = logger.LogPolicyViolation(ctx, resource, violation)
	if err != nil {
		t.Fatalf("Failed to log policy violation: %v", err)
	}
	
	// Verify log entry
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	
	var entry AuditEntry
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	err = json.Unmarshal([]byte(lines[0]), &entry)
	if err != nil {
		t.Fatalf("Failed to parse audit entry: %v", err)
	}
	
	if entry.EventType != EventTypePolicyViolation {
		t.Errorf("Expected event type %s, got %s", EventTypePolicyViolation, entry.EventType)
	}
	
	if entry.PolicyViolation == nil {
		t.Fatal("Expected policy violation data")
	}
	
	if entry.PolicyViolation.Policy != "security" {
		t.Errorf("Expected policy 'security', got '%s'", entry.PolicyViolation.Policy)
	}
}

func TestAuditLogger_LogUserAction(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewAuditLogger(logPath)
	
	ctx := context.Background()
	err = logger.LogUserAction(ctx, "admin", "plan", "test-module", map[string]interface{}{
		"resources": 5,
	})
	if err != nil {
		t.Fatalf("Failed to log user action: %v", err)
	}
	
	// Verify log entry
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	
	var entry AuditEntry
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	err = json.Unmarshal([]byte(lines[0]), &entry)
	if err != nil {
		t.Fatalf("Failed to parse audit entry: %v", err)
	}
	
	if entry.EventType != EventTypeUserAction {
		t.Errorf("Expected event type %s, got %s", EventTypeUserAction, entry.EventType)
	}
	
	if entry.User != "admin" {
		t.Errorf("Expected user 'admin', got '%s'", entry.User)
	}
	
	if entry.Action != "plan" {
		t.Errorf("Expected action 'plan', got '%s'", entry.Action)
	}
}

func TestAuditLogger_LogSystemEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewAuditLogger(logPath)
	
	ctx := context.Background()
	err = logger.LogSystemEvent(ctx, "startup", "Chisel system started", map[string]interface{}{
		"version": "1.0.0",
		"pid":     12345,
	})
	if err != nil {
		t.Fatalf("Failed to log system event: %v", err)
	}
	
	// Verify log entry
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read audit log: %v", err)
	}
	
	var entry AuditEntry
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	err = json.Unmarshal([]byte(lines[0]), &entry)
	if err != nil {
		t.Fatalf("Failed to parse audit entry: %v", err)
	}
	
	if entry.EventType != EventTypeSystemEvent {
		t.Errorf("Expected event type %s, got %s", EventTypeSystemEvent, entry.EventType)
	}
	
	if entry.Message != "Chisel system started" {
		t.Errorf("Expected message 'Chisel system started', got '%s'", entry.Message)
	}
}

func TestAuditLogger_EnableDisable(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewAuditLogger(logPath)
	
	if !logger.IsEnabled() {
		t.Error("Expected logger to be enabled by default")
	}
	
	logger.Disable()
	if logger.IsEnabled() {
		t.Error("Expected logger to be disabled after Disable()")
	}
	
	// Test that logging is skipped when disabled
	ctx := context.Background()
	err = logger.LogUserAction(ctx, "admin", "test", "module", nil)
	if err != nil {
		t.Fatalf("Unexpected error when disabled: %v", err)
	}
	
	// Verify no log file was created
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("Expected no log file when disabled")
	}
	
	logger.Enable()
	if !logger.IsEnabled() {
		t.Error("Expected logger to be enabled after Enable()")
	}
}

func TestAuditLogger_Rotation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	logPath := filepath.Join(tempDir, "audit.log")
	logger := NewAuditLogger(logPath)
	
	// Set small max size for testing
	logger.SetMaxFileSize(100) // 100 bytes
	
	ctx := context.Background()
	
	// Write multiple entries to trigger rotation
	for i := 0; i < 10; i++ {
		err = logger.LogUserAction(ctx, "admin", "test", "module", map[string]interface{}{
			"iteration": i,
			"data":      "some long data to make the entry larger and trigger rotation",
		})
		if err != nil {
			t.Fatalf("Failed to log entry %d: %v", i, err)
		}
	}
	
	// Check that rotation occurred (backup file should exist)
	backupPath := logPath + ".1"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Expected backup file to be created during rotation")
	}
}

func TestAuditEntry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		entry   *AuditEntry
		wantErr bool
	}{
		{
			name: "valid entry",
			entry: &AuditEntry{
				Timestamp: time.Now(),
				EventType: EventTypeUserAction,
				User:      "admin",
				Action:    "plan",
			},
			wantErr: false,
		},
		{
			name: "missing timestamp",
			entry: &AuditEntry{
				EventType: EventTypeUserAction,
				User:      "admin",
				Action:    "plan",
			},
			wantErr: true,
		},
		{
			name: "missing event type",
			entry: &AuditEntry{
				Timestamp: time.Now(),
				User:      "admin",
				Action:    "plan",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
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
