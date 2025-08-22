package core

import (
	"testing"

	"github.com/ataiva-software/forge/pkg/types"
)

func TestPlan_AddChange(t *testing.T) {
	plan := NewPlan()
	
	change := Change{
		Action:   ActionCreate,
		Resource: types.Resource{Type: "file", Name: "test"},
		Diff:     &types.ResourceDiff{Action: types.ActionCreate},
	}
	
	plan.AddChange(change)
	
	if len(plan.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(plan.Changes))
	}
	
	if plan.Changes[0].Action != ActionCreate {
		t.Errorf("Expected ActionCreate, got %v", plan.Changes[0].Action)
	}
}

func TestPlan_Summary(t *testing.T) {
	plan := NewPlan()
	
	// Add different types of changes
	plan.AddChange(Change{Action: ActionCreate, Resource: types.Resource{Type: "file", Name: "file1"}})
	plan.AddChange(Change{Action: ActionCreate, Resource: types.Resource{Type: "file", Name: "file2"}})
	plan.AddChange(Change{Action: ActionUpdate, Resource: types.Resource{Type: "file", Name: "file3"}})
	plan.AddChange(Change{Action: ActionDelete, Resource: types.Resource{Type: "file", Name: "file4"}})
	plan.AddChange(Change{Action: ActionNoOp, Resource: types.Resource{Type: "file", Name: "file5"}})
	
	summary := plan.Summary()
	
	if summary.ToCreate != 2 {
		t.Errorf("Expected 2 creates, got %d", summary.ToCreate)
	}
	if summary.ToUpdate != 1 {
		t.Errorf("Expected 1 update, got %d", summary.ToUpdate)
	}
	if summary.ToDelete != 1 {
		t.Errorf("Expected 1 delete, got %d", summary.ToDelete)
	}
	if summary.NoChanges != 1 {
		t.Errorf("Expected 1 no-op, got %d", summary.NoChanges)
	}
}

func TestPlanner_CreatePlan(t *testing.T) {
	// Create a mock provider registry
	registry := types.NewProviderRegistry()
	
	// Create a simple module
	module := &Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: ModuleMetadata{
			Name:    "test-module",
			Version: "1.0.0",
		},
		Spec: ModuleSpec{
			Resources: []types.Resource{
				{
					Type: "file",
					Name: "test-file",
					Properties: map[string]interface{}{
						"path":    "/tmp/test",
						"content": "test content",
						"state":   "present",
					},
				},
			},
		},
	}
	
	planner := NewPlanner(registry)
	plan, err := planner.CreatePlan(module)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if plan == nil {
		t.Error("Expected plan to be created")
	}
	
	// Since we don't have a real provider registered, we expect an error
	// but the plan should still be created with an error change
	if len(plan.Changes) == 0 {
		t.Error("Expected at least one change in plan")
	}
	
	// The change should have an error since no provider is registered
	if plan.Changes[0].Error == nil {
		t.Error("Expected error in change due to missing provider")
	}
}

func TestAction_String(t *testing.T) {
	tests := []struct {
		action Action
		want   string
	}{
		{ActionCreate, "create"},
		{ActionUpdate, "update"},
		{ActionDelete, "delete"},
		{ActionNoOp, "no-op"},
		{Action(999), "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.action.String(); got != tt.want {
				t.Errorf("Action.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
