package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ataiva-software/chisel/pkg/types"
)

// EventType represents the type of audit event
type EventType string

const (
	EventTypeResourceChange   EventType = "resource_change"
	EventTypePolicyViolation  EventType = "policy_violation"
	EventTypeUserAction       EventType = "user_action"
	EventTypeSystemEvent      EventType = "system_event"
	EventTypeAuthentication   EventType = "authentication"
	EventTypeAuthorization    EventType = "authorization"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	Timestamp       time.Time              `json:"timestamp"`
	EventType       EventType              `json:"event_type"`
	User            string                 `json:"user,omitempty"`
	Action          string                 `json:"action,omitempty"`
	ResourceID      string                 `json:"resource_id,omitempty"`
	ModuleName      string                 `json:"module_name,omitempty"`
	Success         bool                   `json:"success"`
	Message         string                 `json:"message,omitempty"`
	Error           string                 `json:"error,omitempty"`
	Changes         map[string]interface{} `json:"changes,omitempty"`
	PolicyViolation *PolicyViolation       `json:"policy_violation,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	RemoteAddr      string                 `json:"remote_addr,omitempty"`
}

// PolicyViolation represents a policy violation in audit logs
type PolicyViolation struct {
	Policy   string `json:"policy"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Resource string `json:"resource"`
}

// Validate validates the audit entry
func (e *AuditEntry) Validate() error {
	if e.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	
	if e.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	
	return nil
}

// AuditLogger manages audit logging
type AuditLogger struct {
	filePath    string
	enabled     bool
	mu          sync.RWMutex
	file        *os.File
	maxFileSize int64 // Maximum file size in bytes before rotation
	maxFiles    int   // Maximum number of rotated files to keep
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(filePath string) *AuditLogger {
	return &AuditLogger{
		filePath:    filePath,
		enabled:     true,
		maxFileSize: 100 * 1024 * 1024, // 100MB default
		maxFiles:    10,                 // Keep 10 rotated files
	}
}

// IsEnabled returns whether audit logging is enabled
func (l *AuditLogger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled
}

// Enable enables audit logging
func (l *AuditLogger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = true
}

// Disable disables audit logging
func (l *AuditLogger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = false
	
	// Close file if open
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}

// SetMaxFileSize sets the maximum file size before rotation
func (l *AuditLogger) SetMaxFileSize(size int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxFileSize = size
}

// SetMaxFiles sets the maximum number of rotated files to keep
func (l *AuditLogger) SetMaxFiles(count int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxFiles = count
}

// LogResourceChange logs a resource change event
func (l *AuditLogger) LogResourceChange(ctx context.Context, resource *types.Resource, diff *types.ResourceDiff, success bool, err error) error {
	if !l.IsEnabled() {
		return nil
	}
	
	entry := &AuditEntry{
		Timestamp:  time.Now(),
		EventType:  EventTypeResourceChange,
		ResourceID: resource.ResourceID(),
		Action:     string(diff.Action),
		Success:    success,
		Changes:    diff.Changes,
	}
	
	if err != nil {
		entry.Error = err.Error()
	}
	
	// Extract user from context if available
	if user := getUserFromContext(ctx); user != "" {
		entry.User = user
	}
	
	// Extract session ID from context if available
	if sessionID := getSessionIDFromContext(ctx); sessionID != "" {
		entry.SessionID = sessionID
	}
	
	return l.writeEntry(entry)
}

// LogPolicyViolation logs a policy violation event
func (l *AuditLogger) LogPolicyViolation(ctx context.Context, resource *types.Resource, violation PolicyViolation) error {
	if !l.IsEnabled() {
		return nil
	}
	
	entry := &AuditEntry{
		Timestamp:       time.Now(),
		EventType:       EventTypePolicyViolation,
		ResourceID:      resource.ResourceID(),
		Success:         false,
		Message:         violation.Message,
		PolicyViolation: &violation,
	}
	
	// Extract user from context if available
	if user := getUserFromContext(ctx); user != "" {
		entry.User = user
	}
	
	return l.writeEntry(entry)
}

// LogUserAction logs a user action event
func (l *AuditLogger) LogUserAction(ctx context.Context, user, action, moduleName string, metadata map[string]interface{}) error {
	if !l.IsEnabled() {
		return nil
	}
	
	entry := &AuditEntry{
		Timestamp:  time.Now(),
		EventType:  EventTypeUserAction,
		User:       user,
		Action:     action,
		ModuleName: moduleName,
		Success:    true,
		Metadata:   metadata,
	}
	
	// Extract session ID from context if available
	if sessionID := getSessionIDFromContext(ctx); sessionID != "" {
		entry.SessionID = sessionID
	}
	
	// Extract remote address from context if available
	if remoteAddr := getRemoteAddrFromContext(ctx); remoteAddr != "" {
		entry.RemoteAddr = remoteAddr
	}
	
	return l.writeEntry(entry)
}

// LogSystemEvent logs a system event
func (l *AuditLogger) LogSystemEvent(ctx context.Context, action, message string, metadata map[string]interface{}) error {
	if !l.IsEnabled() {
		return nil
	}
	
	entry := &AuditEntry{
		Timestamp: time.Now(),
		EventType: EventTypeSystemEvent,
		Action:    action,
		Message:   message,
		Success:   true,
		Metadata:  metadata,
	}
	
	return l.writeEntry(entry)
}

// LogAuthentication logs an authentication event
func (l *AuditLogger) LogAuthentication(ctx context.Context, user string, success bool, method string, err error) error {
	if !l.IsEnabled() {
		return nil
	}
	
	entry := &AuditEntry{
		Timestamp: time.Now(),
		EventType: EventTypeAuthentication,
		User:      user,
		Action:    "authenticate",
		Success:   success,
		Metadata: map[string]interface{}{
			"method": method,
		},
	}
	
	if err != nil {
		entry.Error = err.Error()
	}
	
	// Extract remote address from context if available
	if remoteAddr := getRemoteAddrFromContext(ctx); remoteAddr != "" {
		entry.RemoteAddr = remoteAddr
	}
	
	return l.writeEntry(entry)
}

// writeEntry writes an audit entry to the log file
func (l *AuditLogger) writeEntry(entry *AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Validate entry
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid audit entry: %w", err)
	}
	
	// Ensure file is open
	if err := l.ensureFileOpen(); err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	
	// Check if rotation is needed
	if err := l.checkRotation(); err != nil {
		return fmt.Errorf("failed to rotate audit log: %w", err)
	}
	
	// Marshal entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}
	
	// Write to file
	_, err = l.file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}
	
	// Sync to disk
	return l.file.Sync()
}

// ensureFileOpen ensures the audit log file is open
func (l *AuditLogger) ensureFileOpen() error {
	if l.file != nil {
		return nil
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(l.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create audit log directory: %w", err)
	}
	
	// Open file for appending
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	
	l.file = file
	return nil
}

// checkRotation checks if log rotation is needed and performs it
func (l *AuditLogger) checkRotation() error {
	if l.file == nil {
		return nil
	}
	
	// Get file info
	info, err := l.file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	// Check if rotation is needed
	if info.Size() < l.maxFileSize {
		return nil
	}
	
	// Close current file
	l.file.Close()
	l.file = nil
	
	// Rotate files
	if err := l.rotateFiles(); err != nil {
		return fmt.Errorf("failed to rotate files: %w", err)
	}
	
	// Reopen file
	return l.ensureFileOpen()
}

// rotateFiles rotates the audit log files
func (l *AuditLogger) rotateFiles() error {
	// Remove oldest file if it exists
	oldestFile := fmt.Sprintf("%s.%d", l.filePath, l.maxFiles)
	os.Remove(oldestFile) // Ignore error if file doesn't exist
	
	// Rotate existing files
	for i := l.maxFiles - 1; i >= 1; i-- {
		oldName := fmt.Sprintf("%s.%d", l.filePath, i)
		newName := fmt.Sprintf("%s.%d", l.filePath, i+1)
		
		if _, err := os.Stat(oldName); err == nil {
			if err := os.Rename(oldName, newName); err != nil {
				return fmt.Errorf("failed to rotate %s to %s: %w", oldName, newName, err)
			}
		}
	}
	
	// Move current file to .1
	backupName := l.filePath + ".1"
	if err := os.Rename(l.filePath, backupName); err != nil {
		return fmt.Errorf("failed to rotate current file: %w", err)
	}
	
	return nil
}

// Close closes the audit logger
func (l *AuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	
	return nil
}

// Context helper functions
func getUserFromContext(ctx context.Context) string {
	if user, ok := ctx.Value("user").(string); ok {
		return user
	}
	return ""
}

func getSessionIDFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

func getRemoteAddrFromContext(ctx context.Context) string {
	if remoteAddr, ok := ctx.Value("remote_addr").(string); ok {
		return remoteAddr
	}
	return ""
}

// AuditConfig represents audit logging configuration
type AuditConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	FilePath    string `yaml:"file_path" json:"file_path"`
	MaxFileSize int64  `yaml:"max_file_size" json:"max_file_size"`
	MaxFiles    int    `yaml:"max_files" json:"max_files"`
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Enabled:     true,
		FilePath:    "/var/log/chisel/audit.log",
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		MaxFiles:    10,
	}
}
