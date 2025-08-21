package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewNotification(t *testing.T) {
	title := "Test Notification"
	message := "This is a test message"
	level := LevelWarning
	
	notification := NewNotification(title, message, level)
	
	if notification.Title != title {
		t.Errorf("Expected title '%s', got '%s'", title, notification.Title)
	}
	
	if notification.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, notification.Message)
	}
	
	if notification.Level != level {
		t.Errorf("Expected level '%s', got '%s'", level, notification.Level)
	}
	
	if notification.ID == "" {
		t.Error("Expected non-empty ID")
	}
	
	if notification.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
	
	if notification.Data == nil {
		t.Error("Expected non-nil data map")
	}
	
	if notification.Tags == nil {
		t.Error("Expected non-nil tags slice")
	}
}

func TestNotification_AddData(t *testing.T) {
	notification := NewNotification("Test", "Message", LevelInfo)
	
	notification.AddData("key1", "value1")
	notification.AddData("key2", 42)
	
	if notification.Data["key1"] != "value1" {
		t.Errorf("Expected data key1 to be 'value1', got '%v'", notification.Data["key1"])
	}
	
	if notification.Data["key2"] != 42 {
		t.Errorf("Expected data key2 to be 42, got '%v'", notification.Data["key2"])
	}
}

func TestNotification_AddTag(t *testing.T) {
	notification := NewNotification("Test", "Message", LevelInfo)
	
	notification.AddTag("tag1")
	notification.AddTag("tag2")
	
	if len(notification.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(notification.Tags))
	}
	
	if notification.Tags[0] != "tag1" {
		t.Errorf("Expected first tag to be 'tag1', got '%s'", notification.Tags[0])
	}
	
	if notification.Tags[1] != "tag2" {
		t.Errorf("Expected second tag to be 'tag2', got '%s'", notification.Tags[1])
	}
}

func TestFileChannel_Send(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "chisel-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"text format", "text"},
		{"csv format", "csv"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, fmt.Sprintf("test-%s.log", tt.format))
			channel := NewFileChannel("test-file", filePath, tt.format)
			
			notification := NewNotification("Test Title", "Test message", LevelWarning)
			notification.AddData("resource_id", "test.resource")
			
			ctx := context.Background()
			err := channel.Send(ctx, notification)
			if err != nil {
				t.Fatalf("Failed to send notification: %v", err)
			}
			
			// Verify file was created and contains content
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			
			contentStr := string(content)
			if contentStr == "" {
				t.Error("Expected non-empty file content")
			}
			
			// Verify format-specific content
			switch tt.format {
			case "json":
				var parsed Notification
				lines := strings.Split(strings.TrimSpace(contentStr), "\n")
				err := json.Unmarshal([]byte(lines[0]), &parsed)
				if err != nil {
					t.Errorf("Failed to parse JSON: %v", err)
				}
				if parsed.Title != "Test Title" {
					t.Errorf("Expected title 'Test Title', got '%s'", parsed.Title)
				}
				
			case "text":
				if !strings.Contains(contentStr, "Test Title") {
					t.Error("Expected text format to contain title")
				}
				if !strings.Contains(contentStr, "WARNING") {
					t.Error("Expected text format to contain level")
				}
				
			case "csv":
				if !strings.Contains(contentStr, "\"Test Title\"") {
					t.Error("Expected CSV format to contain quoted title")
				}
			}
		})
	}
}

func TestConsoleChannel_Send(t *testing.T) {
	tests := []struct {
		name      string
		useStderr bool
		colored   bool
	}{
		{"stdout without color", false, false},
		{"stderr without color", true, false},
		{"stdout with color", false, true},
		{"stderr with color", true, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := NewConsoleChannel("test-console", tt.useStderr, tt.colored)
			
			if channel.Type() != "console" {
				t.Errorf("Expected type 'console', got '%s'", channel.Type())
			}
			
			if channel.Name() != "test-console" {
				t.Errorf("Expected name 'test-console', got '%s'", channel.Name())
			}
			
			notification := NewNotification("Test Title", "Test message", LevelError)
			
			ctx := context.Background()
			err := channel.Send(ctx, notification)
			if err != nil {
				t.Errorf("Failed to send notification: %v", err)
			}
		})
	}
}

func TestWebhookChannel_Send(t *testing.T) {
	// Create test server
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		
		err = json.Unmarshal(body, &receivedPayload)
		if err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
		
		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Expected JSON content type", http.StatusBadRequest)
			return
		}
		
		if r.Header.Get("X-Custom-Header") != "test-value" {
			http.Error(w, "Expected custom header", http.StatusBadRequest)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()
	
	headers := map[string]string{
		"X-Custom-Header": "test-value",
	}
	
	channel := NewWebhookChannel("test-webhook", server.URL, "POST", headers, 5*time.Second)
	
	notification := NewNotification("Test Title", "Test message", LevelInfo)
	notification.AddData("resource_id", "test.resource")
	notification.AddTag("test-tag")
	
	ctx := context.Background()
	err := channel.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to send webhook: %v", err)
	}
	
	// Verify received payload
	if receivedPayload["title"] != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%v'", receivedPayload["title"])
	}
	
	if receivedPayload["message"] != "Test message" {
		t.Errorf("Expected message 'Test message', got '%v'", receivedPayload["message"])
	}
	
	if receivedPayload["level"] != string(LevelInfo) {
		t.Errorf("Expected level '%s', got '%v'", LevelInfo, receivedPayload["level"])
	}
	
	// Verify data
	data, ok := receivedPayload["data"].(map[string]interface{})
	if !ok {
		t.Error("Expected data to be a map")
	} else if data["resource_id"] != "test.resource" {
		t.Errorf("Expected resource_id 'test.resource', got '%v'", data["resource_id"])
	}
	
	// Verify tags
	tags, ok := receivedPayload["tags"].([]interface{})
	if !ok {
		t.Error("Expected tags to be an array")
	} else if len(tags) != 1 || tags[0] != "test-tag" {
		t.Errorf("Expected tags ['test-tag'], got %v", tags)
	}
}

func TestWebhookChannel_SendError(t *testing.T) {
	// Create server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()
	
	channel := NewWebhookChannel("test-webhook", server.URL, "POST", nil, 5*time.Second)
	notification := NewNotification("Test", "Message", LevelError)
	
	ctx := context.Background()
	err := channel.Send(ctx, notification)
	if err == nil {
		t.Error("Expected error when server returns 500, but got none")
	}
	
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to mention status 500, got: %v", err)
	}
}

func TestSlackChannel_Send(t *testing.T) {
	// Create test server to simulate Slack webhook
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		
		err = json.Unmarshal(body, &receivedPayload)
		if err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()
	
	channel := NewSlackChannel("test-slack", server.URL, "#general", "chisel-bot", ":robot_face:", 5*time.Second)
	
	notification := NewNotification("Deployment Complete", "All resources applied successfully", LevelInfo)
	notification.AddData("module", "web-server")
	notification.AddData("resources", 5)
	
	ctx := context.Background()
	err := channel.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to send Slack notification: %v", err)
	}
	
	// Verify Slack-specific payload
	if receivedPayload["channel"] != "#general" {
		t.Errorf("Expected channel '#general', got '%v'", receivedPayload["channel"])
	}
	
	if receivedPayload["username"] != "chisel-bot" {
		t.Errorf("Expected username 'chisel-bot', got '%v'", receivedPayload["username"])
	}
	
	if receivedPayload["icon_emoji"] != ":robot_face:" {
		t.Errorf("Expected icon_emoji ':robot_face:', got '%v'", receivedPayload["icon_emoji"])
	}
	
	// Verify text contains title and message
	text, ok := receivedPayload["text"].(string)
	if !ok {
		t.Error("Expected text to be a string")
	} else {
		if !strings.Contains(text, "Deployment Complete") {
			t.Error("Expected text to contain title")
		}
		if !strings.Contains(text, "All resources applied successfully") {
			t.Error("Expected text to contain message")
		}
	}
	
	// Verify attachments for info level (should have "good" color)
	attachments, ok := receivedPayload["attachments"].([]interface{})
	if !ok {
		t.Error("Expected attachments to be an array")
	} else if len(attachments) > 0 {
		attachment, ok := attachments[0].(map[string]interface{})
		if !ok {
			t.Error("Expected attachment to be a map")
		} else if attachment["color"] != "good" {
			t.Errorf("Expected color 'good' for info level, got '%v'", attachment["color"])
		}
	}
}

func TestEmailChannel_Properties(t *testing.T) {
	to := []string{"admin@example.com", "ops@example.com"}
	channel := NewEmailChannel("test-email", "smtp.example.com", 587, "user", "pass", "from@example.com", to)
	
	if channel.Type() != "email" {
		t.Errorf("Expected type 'email', got '%s'", channel.Type())
	}
	
	if channel.Name() != "test-email" {
		t.Errorf("Expected name 'test-email', got '%s'", channel.Name())
	}
	
	// Note: We can't easily test actual email sending without a real SMTP server
	// In a real test environment, you might use a test SMTP server or mock
}

func TestWebhookChannel_Timeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()
	
	// Create channel with short timeout
	channel := NewWebhookChannel("test-webhook", server.URL, "POST", nil, 100*time.Millisecond)
	notification := NewNotification("Test", "Message", LevelInfo)
	
	ctx := context.Background()
	err := channel.Send(ctx, notification)
	if err == nil {
		t.Error("Expected timeout error, but got none")
	}
	
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected timeout-related error, got: %v", err)
	}
}

func TestFileChannel_DirectoryCreation(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "chisel-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Use nested path that doesn't exist
	filePath := filepath.Join(tempDir, "nested", "dir", "test.log")
	channel := NewFileChannel("test-file", filePath, "text")
	
	notification := NewNotification("Test", "Message", LevelInfo)
	
	ctx := context.Background()
	err = channel.Send(ctx, notification)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected file to be created, but it doesn't exist")
	}
	
	// Verify directory was created
	dirPath := filepath.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Error("Expected directory to be created, but it doesn't exist")
	}
}

func TestFileChannel_AppendMode(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "chisel-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	filePath := filepath.Join(tempDir, "append-test.log")
	channel := NewFileChannel("test-file", filePath, "text")
	
	// Send first notification
	notification1 := NewNotification("First", "First message", LevelInfo)
	ctx := context.Background()
	err = channel.Send(ctx, notification1)
	if err != nil {
		t.Fatalf("Failed to send first notification: %v", err)
	}
	
	// Send second notification
	notification2 := NewNotification("Second", "Second message", LevelWarning)
	err = channel.Send(ctx, notification2)
	if err != nil {
		t.Fatalf("Failed to send second notification: %v", err)
	}
	
	// Verify both notifications are in the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	
	contentStr := string(content)
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")
	
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in file, got %d", len(lines))
	}
	
	if !strings.Contains(lines[0], "First message") {
		t.Error("Expected first line to contain first message")
	}
	
	if !strings.Contains(lines[1], "Second message") {
		t.Error("Expected second line to contain second message")
	}
}
