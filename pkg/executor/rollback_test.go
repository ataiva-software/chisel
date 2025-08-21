package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ataiva-software/chisel/pkg/types"
)

func TestRollbackExecutor_CreateRollbackPlan(t *testing.T) {
	executor := NewRollbackExecutor(3, time.Second)
	
	results := []ExecutionResult{
		{
			ResourceID: "pkg.nginx",
			Success:    true,
			Changes:    []string{"installed"},
		},
		{
			ResourceID: "file.config",
			Success:    true,
			Changes:    []string{"created"},
		},
		{
			ResourceID: "service.nginx",
			Success:    false,
			Error:      errors.New("failed to start"),
		},
	}
	
	prevStates := map[string]map[string]interface{}{
		"pkg.nginx": {
			"state": "absent",
		},
		"file.config": {
			"state": "absent",
		},
		"service.nginx": {
			"state": "stopped",
		},
	}
	
	plan := executor.CreateRollbackPlan(results, prevStates)
	
	// Should only create rollback actions for successful changes
	expectedActions := 2 // pkg.nginx and file.config
	if len(plan.Actions) != expectedActions {
		t.Errorf("Expected %d rollback actions, got %d", expectedActions, len(plan.Actions))
	}
	
	// Actions should be in reverse order
	if len(plan.Actions) >= 2 {
		if plan.Actions[0].ResourceID != "file.config" {
			t.Errorf("Expected first rollback action to be file.config, got %s", plan.Actions[0].ResourceID)
		}
		if plan.Actions[1].ResourceID != "pkg.nginx" {
			t.Errorf("Expected second rollback action to be pkg.nginx, got %s", plan.Actions[1].ResourceID)
		}
	}
}

func TestRollbackExecutor_ExecuteRollback(t *testing.T) {
	executor := NewRollbackExecutor(2, 10*time.Millisecond)
	
	plan := &RollbackPlan{
		Actions: []RollbackAction{
			{
				ResourceID:  "test.resource1",
				Action:      types.ActionUpdate,
				Description: "Test rollback 1",
			},
			{
				ResourceID:  "test.resource2",
				Action:      types.ActionUpdate,
				Description: "Test rollback 2",
			},
		},
		CreatedAt: time.Now(),
	}
	
	tests := []struct {
		name           string
		executorFunc   func(context.Context, *types.ResourceDiff) error
		expectSuccess  bool
		expectedErrors int
	}{
		{
			name: "successful rollback",
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				return nil
			},
			expectSuccess:  true,
			expectedErrors: 0,
		},
		{
			name: "partial rollback failure",
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				if diff.ResourceID == "test.resource1" {
					return errors.New("rollback failed")
				}
				return nil
			},
			expectSuccess:  false,
			expectedErrors: 1,
		},
		{
			name: "complete rollback failure",
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				return errors.New("rollback failed")
			},
			expectSuccess:  false,
			expectedErrors: 2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := executor.ExecuteRollback(ctx, plan, tt.executorFunc)
			
			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected successful rollback, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected rollback to fail, but it succeeded")
				}
			}
		})
	}
}

func TestEnhancedParallelExecutor_ExecuteWithRollback(t *testing.T) {
	executor := NewEnhancedParallelExecutor(2, 5*time.Second, true)
	
	plan := &ExecutionPlan{
		Batches: []ExecutionBatch{
			{
				BatchID: 0,
				Resources: []*types.ResourceDiff{
					{ResourceID: "test.resource1", Action: types.ActionCreate},
					{ResourceID: "test.resource2", Action: types.ActionCreate},
				},
			},
		},
	}
	
	tests := []struct {
		name         string
		executorFunc func(context.Context, *types.ResourceDiff) error
		readerFunc   func(context.Context, *types.ResourceDiff) (map[string]interface{}, error)
		expectError  bool
	}{
		{
			name: "successful execution",
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				return nil
			},
			readerFunc: func(ctx context.Context, diff *types.ResourceDiff) (map[string]interface{}, error) {
				return map[string]interface{}{"state": "absent"}, nil
			},
			expectError: false,
		},
		{
			name: "execution failure with rollback",
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				if diff.ResourceID == "test.resource2" {
					return errors.New("execution failed")
				}
				return nil
			},
			readerFunc: func(ctx context.Context, diff *types.ResourceDiff) (map[string]interface{}, error) {
				return map[string]interface{}{"state": "absent"}, nil
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			results, err := executor.ExecuteWithRollback(ctx, plan, tt.executorFunc, tt.readerFunc)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected execution to fail, but it succeeded")
				}
			} else {
				if err != nil {
					t.Errorf("Expected successful execution, got error: %v", err)
				}
			}
			
			// Should always get results, even on failure
			if len(results) == 0 && len(plan.Batches[0].Resources) > 0 {
				t.Errorf("Expected to get execution results")
			}
		})
	}
}

func TestErrorRecovery_ClassifyError(t *testing.T) {
	recovery := NewErrorRecovery()
	
	tests := []struct {
		name     string
		error    error
		expected string
	}{
		{
			name:     "permission denied error",
			error:    errors.New("permission denied: cannot access file"),
			expected: "permission_denied",
		},
		{
			name:     "access denied error",
			error:    errors.New("access denied to resource"),
			expected: "permission_denied",
		},
		{
			name:     "network timeout error",
			error:    errors.New("connection timeout after 30s"),
			expected: "network_timeout",
		},
		{
			name:     "connection refused error",
			error:    errors.New("connection refused by server"),
			expected: "network_timeout",
		},
		{
			name:     "resource conflict error",
			error:    errors.New("resource already exists"),
			expected: "resource_conflict",
		},
		{
			name:     "unknown error",
			error:    errors.New("some unknown error"),
			expected: "unknown",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := recovery.classifyError(tt.error)
			if result != tt.expected {
				t.Errorf("Expected error type %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestErrorRecovery_RecoverFromError(t *testing.T) {
	recovery := NewErrorRecovery()
	
	tests := []struct {
		name        string
		error       error
		expectError bool
	}{
		{
			name:        "permission denied error",
			error:       errors.New("permission denied"),
			expectError: true, // Recovery not implemented yet
		},
		{
			name:        "unknown error",
			error:       errors.New("unknown error"),
			expectError: true, // No strategy for unknown errors
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			diff := &types.ResourceDiff{ResourceID: "test.resource"}
			
			err := recovery.RecoverFromError(ctx, tt.error, diff)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected recovery to fail, but it succeeded")
				}
			} else {
				if err != nil {
					t.Errorf("Expected successful recovery, got error: %v", err)
				}
			}
		})
	}
}

func TestRollbackExecutor_executeWithRetry(t *testing.T) {
	executor := NewRollbackExecutor(3, 10*time.Millisecond)
	
	tests := []struct {
		name         string
		executorFunc func(context.Context, *types.ResourceDiff) error
		expectError  bool
		callCount    *int
	}{
		{
			name: "success on first try",
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "success on retry",
			executorFunc: func() func(context.Context, *types.ResourceDiff) error {
				attempts := 0
				return func(ctx context.Context, diff *types.ResourceDiff) error {
					attempts++
					if attempts < 2 {
						return errors.New("temporary failure")
					}
					return nil
				}
			}(),
			expectError: false,
		},
		{
			name: "failure after all retries",
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				return errors.New("persistent failure")
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			diff := &types.ResourceDiff{ResourceID: "test.resource"}
			
			err := executor.executeWithRetry(ctx, diff, tt.executorFunc)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected retry to fail, but it succeeded")
				}
			} else {
				if err != nil {
					t.Errorf("Expected successful retry, got error: %v", err)
				}
			}
		})
	}
}
