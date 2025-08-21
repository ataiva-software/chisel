package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NotificationChannel represents a notification delivery channel
type NotificationChannel interface {
	Send(ctx context.Context, notification *Notification) error
	Type() string
	Name() string
}

// Notification represents a notification message
type Notification struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Level     NotificationLevel      `json:"level"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
}

// NotificationLevel represents the severity level
type NotificationLevel string

const (
	LevelInfo    NotificationLevel = "info"
	LevelWarning NotificationLevel = "warning"
	LevelError   NotificationLevel = "error"
	LevelCritical NotificationLevel = "critical"
)

// EmailChannel sends notifications via email
type EmailChannel struct {
	name     string
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
}

// NewEmailChannel creates a new email notification channel
func NewEmailChannel(name, smtpHost string, smtpPort int, username, password, from string, to []string) *EmailChannel {
	return &EmailChannel{
		name:     name,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

// Send sends a notification via email
func (e *EmailChannel) Send(ctx context.Context, notification *Notification) error {
	// Create email message
	subject := fmt.Sprintf("[Chisel %s] %s", strings.ToUpper(string(notification.Level)), notification.Title)
	body := fmt.Sprintf("Time: %s\nLevel: %s\nTitle: %s\n\nMessage:\n%s\n",
		notification.Timestamp.Format(time.RFC3339),
		notification.Level,
		notification.Title,
		notification.Message)
	
	if len(notification.Data) > 0 {
		body += "\nAdditional Data:\n"
		for k, v := range notification.Data {
			body += fmt.Sprintf("  %s: %v\n", k, v)
		}
	}
	
	// Create MIME message
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		strings.Join(e.to, ","), subject, body)
	
	// Send email
	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	addr := fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort)
	
	return smtp.SendMail(addr, auth, e.from, e.to, []byte(message))
}

// Type returns the channel type
func (e *EmailChannel) Type() string {
	return "email"
}

// Name returns the channel name
func (e *EmailChannel) Name() string {
	return e.name
}

// WebhookChannel sends notifications to a webhook URL
type WebhookChannel struct {
	name       string
	url        string
	method     string
	headers    map[string]string
	timeout    time.Duration
	httpClient *http.Client
}

// NewWebhookChannel creates a new webhook notification channel
func NewWebhookChannel(name, url, method string, headers map[string]string, timeout time.Duration) *WebhookChannel {
	if method == "" {
		method = "POST"
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &WebhookChannel{
		name:    name,
		url:     url,
		method:  method,
		headers: headers,
		timeout: timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Send sends a notification to the webhook
func (w *WebhookChannel) Send(ctx context.Context, notification *Notification) error {
	// Create payload
	payload := map[string]interface{}{
		"id":        notification.ID,
		"title":     notification.Title,
		"message":   notification.Message,
		"level":     notification.Level,
		"timestamp": notification.Timestamp.Format(time.RFC3339),
		"data":      notification.Data,
		"tags":      notification.Tags,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, w.method, w.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range w.headers {
		req.Header.Set(key, value)
	}
	
	// Send request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Type returns the channel type
func (w *WebhookChannel) Type() string {
	return "webhook"
}

// Name returns the channel name
func (w *WebhookChannel) Name() string {
	return w.name
}

// FileChannel writes notifications to a file
type FileChannel struct {
	name     string
	filePath string
	format   string // json, text, csv
}

// NewFileChannel creates a new file notification channel
func NewFileChannel(name, filePath, format string) *FileChannel {
	if format == "" {
		format = "json"
	}
	
	return &FileChannel{
		name:     name,
		filePath: filePath,
		format:   format,
	}
}

// Send writes a notification to the file
func (f *FileChannel) Send(ctx context.Context, notification *Notification) error {
	// Ensure directory exists
	dir := filepath.Dir(f.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Open file for appending
	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Format and write notification
	var content string
	switch f.format {
	case "json":
		jsonData, err := json.Marshal(notification)
		if err != nil {
			return fmt.Errorf("failed to marshal notification: %w", err)
		}
		content = string(jsonData) + "\n"
		
	case "text":
		content = fmt.Sprintf("[%s] %s - %s: %s\n",
			notification.Timestamp.Format(time.RFC3339),
			strings.ToUpper(string(notification.Level)),
			notification.Title,
			notification.Message)
		
	case "csv":
		// Escape quotes in CSV
		title := strings.ReplaceAll(notification.Title, "\"", "\"\"")
		message := strings.ReplaceAll(notification.Message, "\"", "\"\"")
		content = fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\"\n",
			notification.Timestamp.Format(time.RFC3339),
			notification.Level,
			title,
			message)
		
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}
	
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	
	return nil
}

// Type returns the channel type
func (f *FileChannel) Type() string {
	return "file"
}

// Name returns the channel name
func (f *FileChannel) Name() string {
	return f.name
}

// ConsoleChannel writes notifications to stdout/stderr
type ConsoleChannel struct {
	name      string
	useStderr bool
	colored   bool
}

// NewConsoleChannel creates a new console notification channel
func NewConsoleChannel(name string, useStderr, colored bool) *ConsoleChannel {
	return &ConsoleChannel{
		name:      name,
		useStderr: useStderr,
		colored:   colored,
	}
}

// Send writes a notification to the console
func (c *ConsoleChannel) Send(ctx context.Context, notification *Notification) error {
	var output io.Writer = os.Stdout
	if c.useStderr {
		output = os.Stderr
	}
	
	// Format message
	timestamp := notification.Timestamp.Format("15:04:05")
	level := strings.ToUpper(string(notification.Level))
	
	var message string
	if c.colored {
		// Add ANSI color codes
		var colorCode string
		switch notification.Level {
		case LevelInfo:
			colorCode = "\033[32m" // Green
		case LevelWarning:
			colorCode = "\033[33m" // Yellow
		case LevelError:
			colorCode = "\033[31m" // Red
		case LevelCritical:
			colorCode = "\033[35m" // Magenta
		default:
			colorCode = "\033[0m" // Reset
		}
		
		message = fmt.Sprintf("%s[%s] %s%s\033[0m - %s: %s\n",
			colorCode, timestamp, level, colorCode, notification.Title, notification.Message)
	} else {
		message = fmt.Sprintf("[%s] %s - %s: %s\n",
			timestamp, level, notification.Title, notification.Message)
	}
	
	_, err := fmt.Fprint(output, message)
	return err
}

// Type returns the channel type
func (c *ConsoleChannel) Type() string {
	return "console"
}

// Name returns the channel name
func (c *ConsoleChannel) Name() string {
	return c.name
}

// SlackChannel sends notifications to Slack via webhook
type SlackChannel struct {
	name       string
	webhookURL string
	channel    string
	username   string
	iconEmoji  string
	timeout    time.Duration
	httpClient *http.Client
}

// NewSlackChannel creates a new Slack notification channel
func NewSlackChannel(name, webhookURL, channel, username, iconEmoji string, timeout time.Duration) *SlackChannel {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &SlackChannel{
		name:       name,
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
		iconEmoji:  iconEmoji,
		timeout:    timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Send sends a notification to Slack
func (s *SlackChannel) Send(ctx context.Context, notification *Notification) error {
	// Create Slack payload
	payload := map[string]interface{}{
		"text": fmt.Sprintf("*%s*\n%s", notification.Title, notification.Message),
	}
	
	if s.channel != "" {
		payload["channel"] = s.channel
	}
	if s.username != "" {
		payload["username"] = s.username
	}
	if s.iconEmoji != "" {
		payload["icon_emoji"] = s.iconEmoji
	}
	
	// Add color based on level
	var color string
	switch notification.Level {
	case LevelInfo:
		color = "good"
	case LevelWarning:
		color = "warning"
	case LevelError, LevelCritical:
		color = "danger"
	}
	
	if color != "" {
		attachment := map[string]interface{}{
			"color": color,
			"text":  notification.Message,
			"ts":    notification.Timestamp.Unix(),
		}
		
		if len(notification.Data) > 0 {
			fields := make([]map[string]interface{}, 0, len(notification.Data))
			for k, v := range notification.Data {
				fields = append(fields, map[string]interface{}{
					"title": k,
					"value": fmt.Sprintf("%v", v),
					"short": true,
				})
			}
			attachment["fields"] = fields
		}
		
		payload["attachments"] = []map[string]interface{}{attachment}
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}
	
	// Send to Slack
	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack returned status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Type returns the channel type
func (s *SlackChannel) Type() string {
	return "slack"
}

// Name returns the channel name
func (s *SlackChannel) Name() string {
	return s.name
}

// NewNotification creates a new notification
func NewNotification(title, message string, level NotificationLevel) *Notification {
	return &Notification{
		ID:        generateNotificationID(),
		Title:     title,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
		Tags:      make([]string, 0),
	}
}

// generateNotificationID generates a unique notification ID
func generateNotificationID() string {
	return fmt.Sprintf("notif_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

// AddData adds data to the notification
func (n *Notification) AddData(key string, value interface{}) {
	if n.Data == nil {
		n.Data = make(map[string]interface{})
	}
	n.Data[key] = value
}

// AddTag adds a tag to the notification
func (n *Notification) AddTag(tag string) {
	if n.Tags == nil {
		n.Tags = make([]string, 0)
	}
	n.Tags = append(n.Tags, tag)
}
