package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/ataiva-software/chisel/pkg/types"
)

// RollbackAction represents an action that can be rolled back
type RollbackAction struct {
	ResourceID  string                 `json:"resource_id"`
	Action      types.DiffAction       `json:"action"`
	PrevState   map[string]interface{} `json:"prev_state"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
}

// RollbackPlan represents a plan for rolling back changes
type RollbackPlan struct {
	Actions   []RollbackAction `json:"actions"`
	CreatedAt time.Time        `json:"created_at"`
}

// RollbackExecutor handles rollback operations
type RollbackExecutor struct {
	maxRetries int
	retryDelay time.Duration
}

// NewRollbackExecutor creates a new rollback executor
func NewRollbackExecutor(maxRetries int, retryDelay time.Duration) *RollbackExecutor {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if retryDelay <= 0 {
		retryDelay = 5 * time.Second
	}
	
	return &RollbackExecutor{
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

// CreateRollbackPlan creates a rollback plan from execution results
func (r *RollbackExecutor) CreateRollbackPlan(results []ExecutionResult, prevStates map[string]map[string]interface{}) *RollbackPlan {
	plan := &RollbackPlan{
		CreatedAt: time.Now(),
	}
	
	// Process results in reverse order for rollback
	for i := len(results) - 1; i >= 0; i-- {
		result := results[i]
		
		// Only create rollback actions for successful changes
		if !result.Success {
			continue
		}
		
		prevState, exists := prevStates[result.ResourceID]
		if !exists {
			continue
		}
		
		action := RollbackAction{
			ResourceID:  result.ResourceID,
			PrevState:   prevState,
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Rollback %s", result.ResourceID),
		}
		
		// Determine rollback action based on what was done
		if len(result.Changes) > 0 {
			// If changes were made, we need to restore previous state
			action.Action = types.ActionUpdate
		}
		
		plan.Actions = append(plan.Actions, action)
	}
	
	return plan
}

// ExecuteRollback executes a rollback plan
func (r *RollbackExecutor) ExecuteRollback(ctx context.Context, plan *RollbackPlan, executor func(context.Context, *types.ResourceDiff) error) error {
	fmt.Printf("Starting rollback of %d actions...\n", len(plan.Actions))
	
	var rollbackErrors []error
	
	for i, action := range plan.Actions {
		fmt.Printf("Rolling back %d/%d: %s\n", i+1, len(plan.Actions), action.ResourceID)
		
		// Create a resource diff for rollback
		diff := &types.ResourceDiff{
			ResourceID: action.ResourceID,
			Action:     action.Action,
			Changes:    map[string]interface{}{
				"rollback": map[string]interface{}{
					"from": "current",
					"to":   "previous",
				},
			},
			Reason: fmt.Sprintf("Rollback to previous state: %s", action.Description),
		}
		
		// Execute rollback with retries
		err := r.executeWithRetry(ctx, diff, executor)
		if err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("failed to rollback %s: %w", action.ResourceID, err))
			fmt.Printf("✗ Failed to rollback %s: %v\n", action.ResourceID, err)
		} else {
			fmt.Printf("✓ Successfully rolled back %s\n", action.ResourceID)
		}
	}
	
	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback completed with %d errors: %v", len(rollbackErrors), rollbackErrors)
	}
	
	fmt.Printf("Rollback completed successfully\n")
	return nil
}

// executeWithRetry executes a rollback action with retry logic
func (r *RollbackExecutor) executeWithRetry(ctx context.Context, diff *types.ResourceDiff, executor func(context.Context, *types.ResourceDiff) error) error {
	var lastErr error
	
	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		err := executor(ctx, diff)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		if attempt < r.maxRetries {
			fmt.Printf("Rollback attempt %d/%d failed for %s, retrying in %v: %v\n", 
				attempt, r.maxRetries, diff.ResourceID, r.retryDelay, err)
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.retryDelay):
				// Continue to next attempt
			}
		}
	}
	
	return fmt.Errorf("rollback failed after %d attempts: %w", r.maxRetries, lastErr)
}

// EnhancedParallelExecutor extends ParallelExecutor with rollback capabilities
type EnhancedParallelExecutor struct {
	*ParallelExecutor
	rollbackExecutor *RollbackExecutor
	enableRollback   bool
}

// NewEnhancedParallelExecutor creates a new enhanced executor with rollback
func NewEnhancedParallelExecutor(maxConcurrency int, timeout time.Duration, enableRollback bool) *EnhancedParallelExecutor {
	return &EnhancedParallelExecutor{
		ParallelExecutor: NewParallelExecutor(maxConcurrency, timeout),
		rollbackExecutor: NewRollbackExecutor(3, 5*time.Second),
		enableRollback:   enableRollback,
	}
}

// ExecuteWithRollback executes a plan with automatic rollback on failure
func (e *EnhancedParallelExecutor) ExecuteWithRollback(ctx context.Context, plan *ExecutionPlan, executor func(context.Context, *types.ResourceDiff) error, reader func(context.Context, *types.ResourceDiff) (map[string]interface{}, error)) ([]ExecutionResult, error) {
	// Capture initial states if rollback is enabled
	var prevStates map[string]map[string]interface{}
	if e.enableRollback && reader != nil {
		prevStates = make(map[string]map[string]interface{})
		
		fmt.Println("Capturing initial states for rollback...")
		for _, batch := range plan.Batches {
			for _, resource := range batch.Resources {
				state, err := reader(ctx, resource)
				if err != nil {
					fmt.Printf("Warning: failed to capture state for %s: %v\n", resource.ResourceID, err)
					continue
				}
				prevStates[resource.ResourceID] = state
			}
		}
	}
	
	// Execute the plan
	results, err := e.ParallelExecutor.Execute(ctx, plan, executor)
	
	// If execution failed and rollback is enabled, attempt rollback
	if err != nil && e.enableRollback && len(results) > 0 {
		fmt.Printf("Execution failed: %v\n", err)
		fmt.Println("Attempting automatic rollback...")
		
		rollbackPlan := e.rollbackExecutor.CreateRollbackPlan(results, prevStates)
		rollbackErr := e.rollbackExecutor.ExecuteRollback(ctx, rollbackPlan, executor)
		
		if rollbackErr != nil {
			return results, fmt.Errorf("execution failed and rollback also failed: original error: %w, rollback error: %v", err, rollbackErr)
		}
		
		return results, fmt.Errorf("execution failed but rollback succeeded: %w", err)
	}
	
	return results, err
}

// ErrorRecovery handles different types of errors and suggests recovery actions
type ErrorRecovery struct {
	strategies map[string]RecoveryStrategy
}

// RecoveryStrategy defines how to recover from specific error types
type RecoveryStrategy struct {
	Name        string
	Description string
	Action      func(context.Context, error, *types.ResourceDiff) error
}

// NewErrorRecovery creates a new error recovery system
func NewErrorRecovery() *ErrorRecovery {
	recovery := &ErrorRecovery{
		strategies: make(map[string]RecoveryStrategy),
	}
	
	// Register default recovery strategies
	recovery.registerDefaultStrategies()
	
	return recovery
}

// registerDefaultStrategies registers built-in recovery strategies
func (e *ErrorRecovery) registerDefaultStrategies() {
	// Permission denied recovery
	e.strategies["permission_denied"] = RecoveryStrategy{
		Name:        "Permission Denied Recovery",
		Description: "Attempt to fix permission issues",
		Action: func(ctx context.Context, err error, diff *types.ResourceDiff) error {
			// In a real implementation, this might try to fix permissions
			return fmt.Errorf("permission recovery not implemented: %w", err)
		},
	}
	
	// Network timeout recovery
	e.strategies["network_timeout"] = RecoveryStrategy{
		Name:        "Network Timeout Recovery",
		Description: "Retry with exponential backoff",
		Action: func(ctx context.Context, err error, diff *types.ResourceDiff) error {
			// In a real implementation, this might retry with backoff
			return fmt.Errorf("network timeout recovery not implemented: %w", err)
		},
	}
	
	// Resource conflict recovery
	e.strategies["resource_conflict"] = RecoveryStrategy{
		Name:        "Resource Conflict Recovery",
		Description: "Resolve resource conflicts",
		Action: func(ctx context.Context, err error, diff *types.ResourceDiff) error {
			// In a real implementation, this might resolve conflicts
			return fmt.Errorf("resource conflict recovery not implemented: %w", err)
		},
	}
}

// RecoverFromError attempts to recover from an error using registered strategies
func (e *ErrorRecovery) RecoverFromError(ctx context.Context, err error, diff *types.ResourceDiff) error {
	errorType := e.classifyError(err)
	
	strategy, exists := e.strategies[errorType]
	if !exists {
		return fmt.Errorf("no recovery strategy for error type '%s': %w", errorType, err)
	}
	
	fmt.Printf("Attempting recovery using strategy: %s\n", strategy.Name)
	return strategy.Action(ctx, err, diff)
}

// classifyError classifies an error to determine the appropriate recovery strategy
func (e *ErrorRecovery) classifyError(err error) string {
	errStr := err.Error()
	
	// Simple error classification - in a real implementation,
	// this would be more sophisticated
	switch {
	case contains(errStr, "permission denied", "access denied"):
		return "permission_denied"
	case contains(errStr, "timeout", "connection refused"):
		return "network_timeout"
	case contains(errStr, "conflict", "already exists"):
		return "resource_conflict"
	default:
		return "unknown"
	}
}

// contains checks if any of the substrings are present in the main string
func contains(str string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
