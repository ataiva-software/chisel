package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// ConnectionConfig holds SSH connection configuration
type ConnectionConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	User            string        `yaml:"user" json:"user"`
	Password        string        `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKeyPath  string        `yaml:"private_key_path,omitempty" json:"private_key_path,omitempty"`
	PrivateKey      string        `yaml:"private_key,omitempty" json:"private_key,omitempty"`
	Timeout         time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout,omitempty" json:"connect_timeout,omitempty"`
	KeepAlive       time.Duration `yaml:"keep_alive,omitempty" json:"keep_alive,omitempty"`
	MaxRetries      int           `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
	StrictHostCheck bool          `yaml:"strict_host_check,omitempty" json:"strict_host_check,omitempty"`
}

// SetDefaults sets default values for connection config
func (c *ConnectionConfig) SetDefaults() {
	if c.Port == 0 {
		c.Port = 22
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 10 * time.Second
	}
	if c.KeepAlive == 0 {
		c.KeepAlive = 30 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
}

// Validate validates the connection configuration
func (c *ConnectionConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if c.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if c.Password == "" && c.PrivateKeyPath == "" && c.PrivateKey == "" {
		return fmt.Errorf("must provide either password, private_key_path, or private_key")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// Connection represents an SSH connection
type Connection struct {
	config *ConnectionConfig
	client *ssh.Client
}

// NewConnection creates a new SSH connection
func NewConnection(config *ConnectionConfig) *Connection {
	config.SetDefaults()
	return &Connection{
		config: config,
	}
}

// Connect establishes the SSH connection
func (c *Connection) Connect(ctx context.Context) error {
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid connection config: %w", err)
	}

	clientConfig, err := c.buildClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build SSH client config: %w", err)
	}

	address := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	
	// Create a dialer with timeout
	dialer := &net.Dialer{
		Timeout: c.config.ConnectTimeout,
	}

	// Dial with context
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to dial %s: %w", address, err)
	}

	// Create SSH connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, clientConfig)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SSH connection: %w", err)
	}

	c.client = ssh.NewClient(sshConn, chans, reqs)
	return nil
}

// Execute runs a command on the remote host
func (c *Connection) Execute(ctx context.Context, command string) (*ExecuteResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	result := &ExecuteResult{
		Command: command,
	}

	// Capture stdout and stderr
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := session.Start(command); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Read output in goroutines
	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)

	go func() {
		data, err := io.ReadAll(stdout)
		result.Stdout = string(data)
		stdoutDone <- err
	}()

	go func() {
		data, err := io.ReadAll(stderr)
		result.Stderr = string(data)
		stderrDone <- err
	}()

	// Wait for command completion or context cancellation
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- session.Wait()
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGTERM)
		return nil, ctx.Err()
	case err := <-cmdDone:
		// Wait for output to be read
		<-stdoutDone
		<-stderrDone
		
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
			} else {
				return nil, fmt.Errorf("command execution failed: %w", err)
			}
		}
	}

	return result, nil
}

// Close closes the SSH connection
func (c *Connection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ExecuteResult holds the result of command execution
type ExecuteResult struct {
	Command  string `json:"command"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// Success returns true if the command executed successfully
func (r *ExecuteResult) Success() bool {
	return r.ExitCode == 0
}

// buildClientConfig creates the SSH client configuration
func (c *Connection) buildClientConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            c.config.User,
		Timeout:         c.config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key checking
	}

	// Add authentication methods
	if c.config.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.config.Password))
	}

	if c.config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	if c.config.PrivateKeyPath != "" {
		key, err := c.loadPrivateKey(c.config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key from %s: %w", c.config.PrivateKeyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key from %s: %w", c.config.PrivateKeyPath, err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	if len(config.Auth) == 0 {
		return nil, fmt.Errorf("no authentication method configured")
	}

	return config, nil
}

// loadPrivateKey loads a private key from file
func (c *Connection) loadPrivateKey(path string) ([]byte, error) {
	// Expand home directory
	if path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	return key, nil
}
