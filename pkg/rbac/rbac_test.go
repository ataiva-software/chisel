package rbac

import (
	"context"
	"testing"
)

func TestRBACManager_New(t *testing.T) {
	manager := NewRBACManager()
	
	if manager == nil {
		t.Fatal("Expected non-nil RBAC manager")
	}
	
	if !manager.IsEnabled() {
		t.Error("Expected RBAC manager to be enabled by default")
	}
}

func TestRBACManager_CreateRole(t *testing.T) {
	manager := NewRBACManager()
	
	role := &Role{
		Name:        "custom-admin",
		Description: "Custom administrator role with full access",
		Permissions: []Permission{
			PermissionModuleRead,
			PermissionModuleWrite,
			PermissionModuleDelete,
			PermissionResourceRead,
			PermissionResourceWrite,
			PermissionResourceDelete,
		},
	}
	
	err := manager.CreateRole(role)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
	
	// Verify role was created
	retrievedRole, err := manager.GetRole("custom-admin")
	if err != nil {
		t.Fatalf("Failed to get role: %v", err)
	}
	
	if retrievedRole.Name != "custom-admin" {
		t.Errorf("Expected role name 'custom-admin', got '%s'", retrievedRole.Name)
	}
	
	if len(retrievedRole.Permissions) != 6 {
		t.Errorf("Expected 6 permissions, got %d", len(retrievedRole.Permissions))
	}
}

func TestRBACManager_CreateUser(t *testing.T) {
	manager := NewRBACManager()
	
	// Create a role first
	role := &Role{
		Name:        "operator",
		Description: "Operator role",
		Permissions: []Permission{
			PermissionModuleRead,
			PermissionResourceRead,
		},
	}
	manager.CreateRole(role)
	
	user := &User{
		Username: "john.doe",
		Email:    "john.doe@example.com",
		Roles:    []string{"operator"},
		Active:   true,
	}
	
	err := manager.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Verify user was created
	retrievedUser, err := manager.GetUser("john.doe")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	
	if retrievedUser.Username != "john.doe" {
		t.Errorf("Expected username 'john.doe', got '%s'", retrievedUser.Username)
	}
	
	if len(retrievedUser.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(retrievedUser.Roles))
	}
}

func TestRBACManager_CheckPermission(t *testing.T) {
	manager := NewRBACManager()
	
	// Create roles
	adminRole := &Role{
		Name: "admin",
		Permissions: []Permission{
			PermissionModuleRead,
			PermissionModuleWrite,
			PermissionModuleDelete,
		},
	}
	
	readOnlyRole := &Role{
		Name: "readonly",
		Permissions: []Permission{
			PermissionModuleRead,
		},
	}
	
	manager.CreateRole(adminRole)
	manager.CreateRole(readOnlyRole)
	
	// Create users
	adminUser := &User{
		Username: "admin",
		Roles:    []string{"admin"},
		Active:   true,
	}
	
	readOnlyUser := &User{
		Username: "readonly",
		Roles:    []string{"readonly"},
		Active:   true,
	}
	
	inactiveUser := &User{
		Username: "inactive",
		Roles:    []string{"admin"},
		Active:   false,
	}
	
	manager.CreateUser(adminUser)
	manager.CreateUser(readOnlyUser)
	manager.CreateUser(inactiveUser)
	
	tests := []struct {
		name       string
		username   string
		permission Permission
		resource   string
		expected   bool
	}{
		{
			name:       "admin can write modules",
			username:   "admin",
			permission: PermissionModuleWrite,
			resource:   "test-module",
			expected:   true,
		},
		{
			name:       "readonly cannot write modules",
			username:   "readonly",
			permission: PermissionModuleWrite,
			resource:   "test-module",
			expected:   false,
		},
		{
			name:       "readonly can read modules",
			username:   "readonly",
			permission: PermissionModuleRead,
			resource:   "test-module",
			expected:   true,
		},
		{
			name:       "inactive user cannot access anything",
			username:   "inactive",
			permission: PermissionModuleRead,
			resource:   "test-module",
			expected:   false,
		},
		{
			name:       "non-existent user cannot access anything",
			username:   "nonexistent",
			permission: PermissionModuleRead,
			resource:   "test-module",
			expected:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := manager.CheckPermission(ctx, tt.username, tt.permission, tt.resource)
			
			if result != tt.expected {
				t.Errorf("Expected permission check to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRBACManager_AssignRole(t *testing.T) {
	manager := NewRBACManager()
	
	// Create roles
	role1 := &Role{Name: "role1", Permissions: []Permission{PermissionModuleRead}}
	role2 := &Role{Name: "role2", Permissions: []Permission{PermissionModuleWrite}}
	
	manager.CreateRole(role1)
	manager.CreateRole(role2)
	
	// Create user
	user := &User{
		Username: "testuser",
		Roles:    []string{"role1"},
		Active:   true,
	}
	manager.CreateUser(user)
	
	// Assign additional role
	err := manager.AssignRole("testuser", "role2")
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}
	
	// Verify user has both roles
	retrievedUser, err := manager.GetUser("testuser")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	
	if len(retrievedUser.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(retrievedUser.Roles))
	}
	
	// Check permissions
	ctx := context.Background()
	if !manager.CheckPermission(ctx, "testuser", PermissionModuleRead, "test") {
		t.Error("Expected user to have read permission")
	}
	
	if !manager.CheckPermission(ctx, "testuser", PermissionModuleWrite, "test") {
		t.Error("Expected user to have write permission")
	}
}

func TestRBACManager_RevokeRole(t *testing.T) {
	manager := NewRBACManager()
	
	// Create roles
	role1 := &Role{Name: "role1", Permissions: []Permission{PermissionModuleRead}}
	role2 := &Role{Name: "role2", Permissions: []Permission{PermissionModuleWrite}}
	
	manager.CreateRole(role1)
	manager.CreateRole(role2)
	
	// Create user with both roles
	user := &User{
		Username: "testuser",
		Roles:    []string{"role1", "role2"},
		Active:   true,
	}
	manager.CreateUser(user)
	
	// Revoke one role
	err := manager.RevokeRole("testuser", "role2")
	if err != nil {
		t.Fatalf("Failed to revoke role: %v", err)
	}
	
	// Verify user has only one role
	retrievedUser, err := manager.GetUser("testuser")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	
	if len(retrievedUser.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(retrievedUser.Roles))
	}
	
	if retrievedUser.Roles[0] != "role1" {
		t.Errorf("Expected remaining role to be 'role1', got '%s'", retrievedUser.Roles[0])
	}
}

func TestRBACManager_EnableDisable(t *testing.T) {
	manager := NewRBACManager()
	
	if !manager.IsEnabled() {
		t.Error("Expected RBAC to be enabled by default")
	}
	
	manager.Disable()
	if manager.IsEnabled() {
		t.Error("Expected RBAC to be disabled after Disable()")
	}
	
	// When disabled, all permissions should be allowed
	ctx := context.Background()
	result := manager.CheckPermission(ctx, "nonexistent", PermissionModuleWrite, "test")
	if !result {
		t.Error("Expected all permissions to be allowed when RBAC is disabled")
	}
	
	manager.Enable()
	if !manager.IsEnabled() {
		t.Error("Expected RBAC to be enabled after Enable()")
	}
}

func TestRBACManager_DeleteUser(t *testing.T) {
	manager := NewRBACManager()
	
	// Create role and user
	role := &Role{Name: "test", Permissions: []Permission{PermissionModuleRead}}
	user := &User{Username: "testuser", Roles: []string{"test"}, Active: true}
	
	manager.CreateRole(role)
	manager.CreateUser(user)
	
	// Verify user exists
	_, err := manager.GetUser("testuser")
	if err != nil {
		t.Fatalf("User should exist: %v", err)
	}
	
	// Delete user
	err = manager.DeleteUser("testuser")
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	
	// Verify user no longer exists
	_, err = manager.GetUser("testuser")
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestRBACManager_DeleteRole(t *testing.T) {
	manager := NewRBACManager()
	
	// Create role
	role := &Role{Name: "test", Permissions: []Permission{PermissionModuleRead}}
	manager.CreateRole(role)
	
	// Verify role exists
	_, err := manager.GetRole("test")
	if err != nil {
		t.Fatalf("Role should exist: %v", err)
	}
	
	// Delete role
	err = manager.DeleteRole("test")
	if err != nil {
		t.Fatalf("Failed to delete role: %v", err)
	}
	
	// Verify role no longer exists
	_, err = manager.GetRole("test")
	if err == nil {
		t.Error("Expected error when getting deleted role")
	}
}

func TestRBACManager_ListUsers(t *testing.T) {
	manager := NewRBACManager()
	
	// Create role
	role := &Role{Name: "test", Permissions: []Permission{PermissionModuleRead}}
	manager.CreateRole(role)
	
	// Create users
	users := []*User{
		{Username: "user1", Roles: []string{"test"}, Active: true},
		{Username: "user2", Roles: []string{"test"}, Active: false},
		{Username: "user3", Roles: []string{"test"}, Active: true},
	}
	
	for _, user := range users {
		manager.CreateUser(user)
	}
	
	// List all users
	allUsers := manager.ListUsers()
	if len(allUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(allUsers))
	}
	
	// List active users only
	activeUsers := manager.ListActiveUsers()
	if len(activeUsers) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(activeUsers))
	}
}

func TestRBACManager_ListRoles(t *testing.T) {
	manager := NewRBACManager()
	
	// Create roles
	roles := []*Role{
		{Name: "admin", Permissions: []Permission{PermissionModuleWrite}},
		{Name: "readonly", Permissions: []Permission{PermissionModuleRead}},
		{Name: "operator", Permissions: []Permission{PermissionResourceWrite}},
	}
	
	for _, role := range roles {
		manager.CreateRole(role)
	}
	
	// List roles
	allRoles := manager.ListRoles()
	if len(allRoles) != 3 {
		t.Errorf("Expected 3 roles, got %d", len(allRoles))
	}
}

func TestPermission_String(t *testing.T) {
	tests := []struct {
		permission Permission
		expected   string
	}{
		{PermissionModuleRead, "module:read"},
		{PermissionModuleWrite, "module:write"},
		{PermissionResourceDelete, "resource:delete"},
	}
	
	for _, tt := range tests {
		if tt.permission.String() != tt.expected {
			t.Errorf("Expected permission string '%s', got '%s'", tt.expected, tt.permission.String())
		}
	}
}
