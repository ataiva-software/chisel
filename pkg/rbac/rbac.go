package rbac

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Permission represents a specific permission
type Permission string

const (
	// Module permissions
	PermissionModuleRead   Permission = "module:read"
	PermissionModuleWrite  Permission = "module:write"
	PermissionModuleDelete Permission = "module:delete"
	
	// Resource permissions
	PermissionResourceRead   Permission = "resource:read"
	PermissionResourceWrite  Permission = "resource:write"
	PermissionResourceDelete Permission = "resource:delete"
	
	// System permissions
	PermissionSystemAdmin Permission = "system:admin"
	PermissionUserManage  Permission = "user:manage"
	PermissionRoleManage  Permission = "role:manage"
	
	// Audit permissions
	PermissionAuditRead Permission = "audit:read"
	
	// Policy permissions
	PermissionPolicyRead  Permission = "policy:read"
	PermissionPolicyWrite Permission = "policy:write"
)

// String returns the string representation of the permission
func (p Permission) String() string {
	return string(p)
}

// Role represents a role with a set of permissions
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permission Permission) bool {
	for _, p := range r.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// User represents a user in the RBAC system
type User struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

// RBACManager manages roles, users, and permissions
type RBACManager struct {
	roles   map[string]*Role
	users   map[string]*User
	enabled bool
	mu      sync.RWMutex
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager() *RBACManager {
	manager := &RBACManager{
		roles:   make(map[string]*Role),
		users:   make(map[string]*User),
		enabled: true,
	}
	
	// Create default roles
	manager.createDefaultRoles()
	
	return manager
}

// createDefaultRoles creates default system roles
func (m *RBACManager) createDefaultRoles() {
	// Admin role with all permissions
	adminRole := &Role{
		Name:        "admin",
		Description: "Administrator with full system access",
		Permissions: []Permission{
			PermissionModuleRead,
			PermissionModuleWrite,
			PermissionModuleDelete,
			PermissionResourceRead,
			PermissionResourceWrite,
			PermissionResourceDelete,
			PermissionSystemAdmin,
			PermissionUserManage,
			PermissionRoleManage,
			PermissionAuditRead,
			PermissionPolicyRead,
			PermissionPolicyWrite,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Operator role with read/write access to modules and resources
	operatorRole := &Role{
		Name:        "operator",
		Description: "Operator with read/write access to modules and resources",
		Permissions: []Permission{
			PermissionModuleRead,
			PermissionModuleWrite,
			PermissionResourceRead,
			PermissionResourceWrite,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Read-only role
	readOnlyRole := &Role{
		Name:        "readonly",
		Description: "Read-only access to modules and resources",
		Permissions: []Permission{
			PermissionModuleRead,
			PermissionResourceRead,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	m.roles["admin"] = adminRole
	m.roles["operator"] = operatorRole
	m.roles["readonly"] = readOnlyRole
}

// IsEnabled returns whether RBAC is enabled
func (m *RBACManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// Enable enables RBAC
func (m *RBACManager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables RBAC (allows all operations)
func (m *RBACManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// CreateRole creates a new role
func (m *RBACManager) CreateRole(role *Role) error {
	if role.Name == "" {
		return fmt.Errorf("role name cannot be empty")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.roles[role.Name]; exists {
		return fmt.Errorf("role '%s' already exists", role.Name)
	}
	
	// Set timestamps
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now
	
	m.roles[role.Name] = role
	return nil
}

// GetRole retrieves a role by name
func (m *RBACManager) GetRole(name string) (*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	role, exists := m.roles[name]
	if !exists {
		return nil, fmt.Errorf("role '%s' not found", name)
	}
	
	// Return a copy to prevent external modification
	roleCopy := *role
	return &roleCopy, nil
}

// UpdateRole updates an existing role
func (m *RBACManager) UpdateRole(role *Role) error {
	if role.Name == "" {
		return fmt.Errorf("role name cannot be empty")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	existingRole, exists := m.roles[role.Name]
	if !exists {
		return fmt.Errorf("role '%s' not found", role.Name)
	}
	
	// Preserve creation time
	role.CreatedAt = existingRole.CreatedAt
	role.UpdatedAt = time.Now()
	
	m.roles[role.Name] = role
	return nil
}

// DeleteRole deletes a role
func (m *RBACManager) DeleteRole(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.roles[name]; !exists {
		return fmt.Errorf("role '%s' not found", name)
	}
	
	// Check if any users have this role
	for _, user := range m.users {
		for _, roleName := range user.Roles {
			if roleName == name {
				return fmt.Errorf("cannot delete role '%s': still assigned to user '%s'", name, user.Username)
			}
		}
	}
	
	delete(m.roles, name)
	return nil
}

// ListRoles returns all roles
func (m *RBACManager) ListRoles() []*Role {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	roles := make([]*Role, 0, len(m.roles))
	for _, role := range m.roles {
		// Return copies to prevent external modification
		roleCopy := *role
		roles = append(roles, &roleCopy)
	}
	
	return roles
}

// CreateUser creates a new user
func (m *RBACManager) CreateUser(user *User) error {
	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.users[user.Username]; exists {
		return fmt.Errorf("user '%s' already exists", user.Username)
	}
	
	// Validate roles exist
	for _, roleName := range user.Roles {
		if _, exists := m.roles[roleName]; !exists {
			return fmt.Errorf("role '%s' does not exist", roleName)
		}
	}
	
	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	m.users[user.Username] = user
	return nil
}

// GetUser retrieves a user by username
func (m *RBACManager) GetUser(username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.users[username]
	if !exists {
		return nil, fmt.Errorf("user '%s' not found", username)
	}
	
	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// UpdateUser updates an existing user
func (m *RBACManager) UpdateUser(user *User) error {
	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	existingUser, exists := m.users[user.Username]
	if !exists {
		return fmt.Errorf("user '%s' not found", user.Username)
	}
	
	// Validate roles exist
	for _, roleName := range user.Roles {
		if _, exists := m.roles[roleName]; !exists {
			return fmt.Errorf("role '%s' does not exist", roleName)
		}
	}
	
	// Preserve creation time and last login
	user.CreatedAt = existingUser.CreatedAt
	user.LastLogin = existingUser.LastLogin
	user.UpdatedAt = time.Now()
	
	m.users[user.Username] = user
	return nil
}

// DeleteUser deletes a user
func (m *RBACManager) DeleteUser(username string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.users[username]; !exists {
		return fmt.Errorf("user '%s' not found", username)
	}
	
	delete(m.users, username)
	return nil
}

// ListUsers returns all users
func (m *RBACManager) ListUsers() []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		// Return copies to prevent external modification
		userCopy := *user
		users = append(users, &userCopy)
	}
	
	return users
}

// ListActiveUsers returns all active users
func (m *RBACManager) ListActiveUsers() []*User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	users := make([]*User, 0)
	for _, user := range m.users {
		if user.Active {
			// Return copies to prevent external modification
			userCopy := *user
			users = append(users, &userCopy)
		}
	}
	
	return users
}

// AssignRole assigns a role to a user
func (m *RBACManager) AssignRole(username, roleName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[username]
	if !exists {
		return fmt.Errorf("user '%s' not found", username)
	}
	
	if _, exists := m.roles[roleName]; !exists {
		return fmt.Errorf("role '%s' not found", roleName)
	}
	
	// Check if user already has the role
	for _, existingRole := range user.Roles {
		if existingRole == roleName {
			return fmt.Errorf("user '%s' already has role '%s'", username, roleName)
		}
	}
	
	user.Roles = append(user.Roles, roleName)
	user.UpdatedAt = time.Now()
	
	return nil
}

// RevokeRole revokes a role from a user
func (m *RBACManager) RevokeRole(username, roleName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[username]
	if !exists {
		return fmt.Errorf("user '%s' not found", username)
	}
	
	// Find and remove the role
	for i, existingRole := range user.Roles {
		if existingRole == roleName {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			user.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("user '%s' does not have role '%s'", username, roleName)
}

// CheckPermission checks if a user has a specific permission for a resource
func (m *RBACManager) CheckPermission(ctx context.Context, username string, permission Permission, resource string) bool {
	m.mu.RLock()
	enabled := m.enabled
	m.mu.RUnlock()
	
	// If RBAC is disabled, allow all operations
	if !enabled {
		return true
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Get user
	user, exists := m.users[username]
	if !exists || !user.Active {
		return false
	}
	
	// Check each role for the permission
	for _, roleName := range user.Roles {
		role, exists := m.roles[roleName]
		if !exists {
			continue
		}
		
		if role.HasPermission(permission) {
			return true
		}
	}
	
	return false
}

// UpdateLastLogin updates the last login time for a user
func (m *RBACManager) UpdateLastLogin(username string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, exists := m.users[username]
	if !exists {
		return fmt.Errorf("user '%s' not found", username)
	}
	
	user.LastLogin = time.Now()
	return nil
}

// GetUserPermissions returns all permissions for a user
func (m *RBACManager) GetUserPermissions(username string) ([]Permission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.users[username]
	if !exists {
		return nil, fmt.Errorf("user '%s' not found", username)
	}
	
	if !user.Active {
		return []Permission{}, nil
	}
	
	permissionSet := make(map[Permission]bool)
	
	// Collect permissions from all roles
	for _, roleName := range user.Roles {
		role, exists := m.roles[roleName]
		if !exists {
			continue
		}
		
		for _, permission := range role.Permissions {
			permissionSet[permission] = true
		}
	}
	
	// Convert to slice
	permissions := make([]Permission, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}
	
	return permissions, nil
}

// RBACConfig represents RBAC configuration
type RBACConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DefaultRBACConfig returns default RBAC configuration
func DefaultRBACConfig() *RBACConfig {
	return &RBACConfig{
		Enabled: true,
	}
}
