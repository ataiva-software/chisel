package core

import (
	"context"
	"fmt"
	"time"

	"github.com/ataiva-software/chisel/pkg/types"
)

// ChangeResult represents the result of executing a single change
type ChangeResult struct {
	Change    Change        `json:"change"`
	Success   bool          `json:"success"`
	Error     error         `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
}

// ExecutionResult represents the result of executing a plan
type ExecutionResult struct {
	Changes []ChangeResult    `json:"changes"`
	Summary ExecutionSummary `json:"summary"`
}

// ExecutionSummary provides a summary of execution results
type ExecutionSummary struct {
	Total     int           `json:"total"`
	Succeeded int           `json:"succeeded"`
	Failed    int           `json:"failed"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
}

// NewExecutionResult creates a new execution result
func NewExecutionResult() *ExecutionResult {
	return &ExecutionResult{
		Changes: make([]ChangeResult, 0),
		Summary: ExecutionSummary{
			StartTime: time.Now(),
		},
	}
}

// AddChangeResult adds a change result and updates the summary
func (er *ExecutionResult) AddChangeResult(result ChangeResult) {
	er.Changes = append(er.Changes, result)
	
	// Update summary
	er.Summary.Total++
	if result.Success {
		er.Summary.Succeeded++
	} else {
		er.Summary.Failed++
	}
}

// Finalize finalizes the execution result by setting end time and duration
func (er *ExecutionResult) Finalize() {
	er.Summary.EndTime = time.Now()
	er.Summary.Duration = er.Summary.EndTime.Sub(er.Summary.StartTime)
}

// Executor executes plans by applying changes
type Executor struct {
	registry *types.ProviderRegistry
}

// NewExecutor creates a new executor with the given provider registry
func NewExecutor(registry *types.ProviderRegistry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// ExecutePlan executes all changes in a plan
func (e *Executor) ExecutePlan(ctx context.Context, plan *Plan) (*ExecutionResult, error) {
	result := NewExecutionResult()
	
	// Execute each change in the plan
	for _, change := range plan.Changes {
		// Skip changes that have errors from planning phase
		if change.Error != nil {
			changeResult := ChangeResult{
				Change:    change,
				Success:   false,
				Error:     change.Error,
				StartTime: time.Now(),
			}
			changeResult.EndTime = changeResult.StartTime
			result.AddChangeResult(changeResult)
			continue
		}
		
		// Skip no-op changes
		if change.Action == ActionNoOp {
			changeResult := ChangeResult{
				Change:    change,
				Success:   true,
				StartTime: time.Now(),
			}
			changeResult.EndTime = changeResult.StartTime
			result.AddChangeResult(changeResult)
			continue
		}
		
		// Execute the change
		changeResult := e.executeChange(ctx, change)
		result.AddChangeResult(changeResult)
		
		// Stop execution on failure (fail-fast behavior)
		if !changeResult.Success {
			break
		}
	}
	
	result.Finalize()
	return result, nil
}

// executeChange executes a single change
func (e *Executor) executeChange(ctx context.Context, change Change) ChangeResult {
	startTime := time.Now()
	
	result := ChangeResult{
		Change:    change,
		StartTime: startTime,
	}
	
	// Get the provider for this resource type
	provider, err := e.registry.Get(change.Resource.Type)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("no provider found for resource type: %s", change.Resource.Type)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result
	}
	
	// Apply the change using the provider
	if err := provider.Apply(ctx, &change.Resource, change.Diff); err != nil {
		result.Success = false
		result.Error = fmt.Errorf("failed to apply change: %w", err)
	} else {
		result.Success = true
	}
	
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	return result
}

// ExecuteWithOptions executes a plan with additional options
type ExecuteOptions struct {
	DryRun    bool `json:"dry_run"`
	Parallel  bool `json:"parallel"`
	FailFast  bool `json:"fail_fast"`
	MaxRetries int `json:"max_retries"`
}

// ExecutePlanWithOptions executes a plan with the given options
func (e *Executor) ExecutePlanWithOptions(ctx context.Context, plan *Plan, options ExecuteOptions) (*ExecutionResult, error) {
	if options.DryRun {
		// For dry run, just return the plan as if it was executed successfully
		result := NewExecutionResult()
		for _, change := range plan.Changes {
			changeResult := ChangeResult{
				Change:    change,
				Success:   true,
				StartTime: time.Now(),
			}
			changeResult.EndTime = changeResult.StartTime
			result.AddChangeResult(changeResult)
		}
		result.Finalize()
		return result, nil
	}
	
	// For now, just use the regular execution
	// TODO: Implement parallel execution and retry logic
	return e.ExecutePlan(ctx, plan)
}
