package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// RealSSHConnection implements the Executor interface using real SSH connections
type RealSSHConnection struct {
	config     *ConnectionConfig
	client     *ssh.Client
	session    *ssh.Session
	connected  bool
	timeout    time.Duration
	retries    int
	retryDelay time.Duration
}

// NewRealSSHConnection creates a new real SSH connection
func NewRealSSHConnection(config *ConnectionConfig) *RealSSHConnection {
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &RealSSHConnection{
		config:     config,
		timeout:    timeout,
		retries:    3,
		retryDelay: 2 * time.Second,
	}
}

// Connect establishes the SSH connection
func (c *RealSSHConnection) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	// Set defaults
	c.config.SetDefaults()

	// Validate configuration
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid SSH configuration: %w", err)
	}

	// Create SSH client configuration
	sshConfig, err := c.createSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}

	// Connect with retries
	var lastErr error
	for attempt := 1; attempt <= c.retries; attempt++ {
		client, err := c.connectWithTimeout(ctx, sshConfig)
		if err == nil {
			c.client = client
			c.connected = true
			return nil
		}

		lastErr = err
		if attempt < c.retries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.retryDelay):
				// Continue to next attempt
			}
		}
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", c.retries, lastErr)
}

// Execute executes a command over SSH
func (c *RealSSHConnection) Execute(ctx context.Context, command string) (*ExecuteResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to SSH server")
	}

	// Create a new session for each command
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Set up pipes for stdout and stderr
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

	// Read output with timeout
	result := &ExecuteResult{
		Command: command,
	}

	// Create channels for output
	stdoutChan := make(chan string, 1)
	stderrChan := make(chan string, 1)
	doneChan := make(chan error, 1)

	// Read stdout
	go func() {
		output, err := io.ReadAll(stdout)
		if err != nil {
			stdoutChan <- ""
		} else {
			stdoutChan <- string(output)
		}
	}()

	// Read stderr
	go func() {
		output, err := io.ReadAll(stderr)
		if err != nil {
			stderrChan <- ""
		} else {
			stderrChan <- string(output)
		}
	}()

	// Wait for command completion
	go func() {
		doneChan <- session.Wait()
	}()

	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGTERM)
		return nil, ctx.Err()
	case err := <-doneChan:
		result.Stdout = <-stdoutChan
		result.Stderr = <-stderrChan

		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
			} else {
				result.ExitCode = 1
			}
		} else {
			result.ExitCode = 0
		}

		return result, nil
	case <-time.After(c.timeout):
		session.Signal(ssh.SIGTERM)
		return nil, fmt.Errorf("command timed out after %v", c.timeout)
	}
}

// Close closes the SSH connection
func (c *RealSSHConnection) Close() error {
	if c.session != nil {
		c.session.Close()
		c.session = nil
	}

	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		c.connected = false
		return err
	}

	return nil
}

// createSSHConfig creates the SSH client configuration
func (c *RealSSHConnection) createSSHConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            c.config.User,
		Timeout:         c.timeout,
		HostKeyCallback: c.createHostKeyCallback(),
	}

	// Set up authentication
	auth, err := c.createAuthMethods()
	if err != nil {
		return nil, fmt.Errorf("failed to create auth methods: %w", err)
	}
	config.Auth = auth

	return config, nil
}

// createAuthMethods creates authentication methods based on configuration
func (c *RealSSHConnection) createAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// Try SSH agent first
	if agentAuth := c.trySSHAgent(); agentAuth != nil {
		methods = append(methods, agentAuth)
	}

	// Try private key authentication
	if c.config.PrivateKey != "" {
		keyAuth, err := c.createKeyAuth(c.config.PrivateKey)
		if err == nil {
			methods = append(methods, keyAuth)
		}
	}

	// Try private key file authentication
	if c.config.PrivateKeyPath != "" {
		keyAuth, err := c.createKeyFileAuth(c.config.PrivateKeyPath)
		if err == nil {
			methods = append(methods, keyAuth)
		}
	}

	// Try password authentication (least secure)
	if c.config.Password != "" {
		methods = append(methods, ssh.Password(c.config.Password))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	return methods, nil
}

// trySSHAgent attempts to use SSH agent for authentication
func (c *RealSSHConnection) trySSHAgent() ssh.AuthMethod {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers)
}

// createKeyAuth creates authentication from private key string
func (c *RealSSHConnection) createKeyAuth(privateKey string) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// createKeyFileAuth creates authentication from private key file
func (c *RealSSHConnection) createKeyFileAuth(keyFile string) (ssh.AuthMethod, error) {
	// Expand home directory
	if strings.HasPrefix(keyFile, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		keyFile = filepath.Join(home, keyFile[2:])
	}

	key, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// createHostKeyCallback creates the host key verification callback
func (c *RealSSHConnection) createHostKeyCallback() ssh.HostKeyCallback {
	// Try to use known_hosts file
	knownHostsFile := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsFile); err == nil {
		callback, err := knownhosts.New(knownHostsFile)
		if err == nil {
			return callback
		}
	}

	// Fallback to insecure callback (for development/testing)
	// In production, you should always verify host keys
	return ssh.InsecureIgnoreHostKey()
}

// connectWithTimeout establishes SSH connection with timeout
func (c *RealSSHConnection) connectWithTimeout(ctx context.Context, config *ssh.ClientConfig) (*ssh.Client, error) {
	address := net.JoinHostPort(c.config.Host, fmt.Sprintf("%d", c.config.Port))

	// Create a dialer with timeout
	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	// Dial with context
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", address, err)
	}

	// Create SSH connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create SSH connection: %w", err)
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

// SSHConnectionPool manages a pool of SSH connections for efficiency
type SSHConnectionPool struct {
	connections map[string]*RealSSHConnection
	maxIdle     int
	maxActive   int
}

// NewSSHConnectionPool creates a new connection pool
func NewSSHConnectionPool(maxIdle, maxActive int) *SSHConnectionPool {
	return &SSHConnectionPool{
		connections: make(map[string]*RealSSHConnection),
		maxIdle:     maxIdle,
		maxActive:   maxActive,
	}
}

// GetConnection gets or creates a connection for the given configuration
func (p *SSHConnectionPool) GetConnection(config *ConnectionConfig) (*RealSSHConnection, error) {
	key := fmt.Sprintf("%s@%s:%d", config.User, config.Host, config.Port)

	if conn, exists := p.connections[key]; exists {
		return conn, nil
	}

	if len(p.connections) >= p.maxActive {
		return nil, fmt.Errorf("connection pool exhausted")
	}

	conn := NewRealSSHConnection(config)
	p.connections[key] = conn

	return conn, nil
}

// CloseAll closes all connections in the pool
func (p *SSHConnectionPool) CloseAll() error {
	var errors []error

	for key, conn := range p.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection %s: %w", key, err))
		}
		delete(p.connections, key)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	return nil
}
