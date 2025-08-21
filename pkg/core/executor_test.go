package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/ataiva-software/chisel/pkg/types"
)

func TestExecutor_ExecutePlan(t *testing.T) {
	// Create a mock provider registry
	registry := types.NewProviderRegistry()
	
	// Create a plan with some changes
	plan := NewPlan()
	plan.AddChange(Change{
		Action: ActionCreate,
		Resource: types.Resource{
			Type: "file",
			Name: "test-file",
			Properties: map[string]interface{}{
				"path":    "/tmp/test",
				"content": "test content",
			},
		},
		Diff: &types.ResourceDiff{
			Action: types.ActionCreate,
		},
	})
	
	executor := NewExecutor(registry)
	result, err := executor.ExecutePlan(context.Background(), plan)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if result == nil {
		t.Error("Expected execution result to be created")
	}
	
	// Since we don't have a real provider registered, we expect failures
	if result.Summary.Succeeded != 0 {
		t.Errorf("Expected 0 successes, got %d", result.Summary.Succeeded)
	}
	
	if result.Summary.Failed == 0 {
		t.Error("Expected at least 1 failure due to missing provider")
	}
}

func TestExecutor_ExecuteChange(t *testing.T) {
	registry := types.NewProviderRegistry()
	executor := NewExecutor(registry)
	
	change := Change{
		Action: ActionCreate,
		Resource: types.Resource{
			Type: "file",
			Name: "test-file",
			Properties: map[string]interface{}{
				"path":    "/tmp/test",
				"content": "test content",
			},
		},
		Diff: &types.ResourceDiff{
			Action: types.ActionCreate,
		},
	}
	
	result := executor.executeChange(context.Background(), change)
	
	// Should fail because no provider is registered
	if result.Error == nil {
		t.Error("Expected error due to missing provider")
	}
	
	if result.Success {
		t.Error("Expected failure due to missing provider")
	}
}

func TestExecutionResult_Summary(t *testing.T) {
	result := NewExecutionResult()
	
	// Add some change results
	result.AddChangeResult(ChangeResult{
		Change:  Change{Action: ActionCreate},
		Success: true,
	})
	result.AddChangeResult(ChangeResult{
		Change:  Change{Action: ActionUpdate},
		Success: true,
	})
	result.AddChangeResult(ChangeResult{
		Change:  Change{Action: ActionDelete},
		Success: false,
		Error:   fmt.Errorf("test error"),
	})
	result.AddChangeResult(ChangeResult{
		Change:  Change{Action: ActionNoOp},
		Success: true,
	})
	
	summary := result.Summary
	
	if summary.Total != 4 {
		t.Errorf("Expected 4 total, got %d", summary.Total)
	}
	if summary.Succeeded != 3 {
		t.Errorf("Expected 3 succeeded, got %d", summary.Succeeded)
	}
	if summary.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", summary.Failed)
	}
}
