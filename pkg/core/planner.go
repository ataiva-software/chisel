package core

import (
	"context"
	"fmt"

	"github.com/ataiva-software/chisel/pkg/types"
)

// Action represents the type of action to be performed
type Action int

const (
	ActionCreate Action = iota
	ActionUpdate
	ActionDelete
	ActionNoOp
)

// String returns the string representation of an action
func (a Action) String() string {
	switch a {
	case ActionCreate:
		return "create"
	case ActionUpdate:
		return "update"
	case ActionDelete:
		return "delete"
	case ActionNoOp:
		return "no-op"
	default:
		return "unknown"
	}
}

// Change represents a planned change to a resource
type Change struct {
	Action   Action                `json:"action"`
	Resource types.Resource        `json:"resource"`
	Diff     *types.ResourceDiff   `json:"diff,omitempty"`
	Error    error                 `json:"error,omitempty"`
}

// Plan represents a collection of planned changes
type Plan struct {
	Changes []Change `json:"changes"`
}

// PlanSummary provides a summary of planned changes
type PlanSummary struct {
	ToCreate  int `json:"to_create"`
	ToUpdate  int `json:"to_update"`
	ToDelete  int `json:"to_delete"`
	NoChanges int `json:"no_changes"`
	Errors    int `json:"errors"`
}

// NewPlan creates a new empty plan
func NewPlan() *Plan {
	return &Plan{
		Changes: make([]Change, 0),
	}
}

// AddChange adds a change to the plan
func (p *Plan) AddChange(change Change) {
	p.Changes = append(p.Changes, change)
}

// Summary returns a summary of the plan
func (p *Plan) Summary() PlanSummary {
	summary := PlanSummary{}
	
	for _, change := range p.Changes {
		if change.Error != nil {
			summary.Errors++
			continue
		}
		
		switch change.Action {
		case ActionCreate:
			summary.ToCreate++
		case ActionUpdate:
			summary.ToUpdate++
		case ActionDelete:
			summary.ToDelete++
		case ActionNoOp:
			summary.NoChanges++
		}
	}
	
	return summary
}

// HasChanges returns true if the plan has any changes that need to be applied
func (p *Plan) HasChanges() bool {
	for _, change := range p.Changes {
		if change.Action != ActionNoOp && change.Error == nil {
			return true
		}
	}
	return false
}

// Planner creates execution plans for modules
type Planner struct {
	registry *types.ProviderRegistry
}

// NewPlanner creates a new planner with the given provider registry
func NewPlanner(registry *types.ProviderRegistry) *Planner {
	return &Planner{
		registry: registry,
	}
}

// CreatePlan creates an execution plan for the given module
func (p *Planner) CreatePlan(module *Module) (*Plan, error) {
	if err := module.Validate(); err != nil {
		return nil, fmt.Errorf("invalid module: %w", err)
	}
	
	plan := NewPlan()
	
	// Process each resource in the module
	for _, resource := range module.Spec.Resources {
		change, err := p.planResource(resource)
		if err != nil {
			change = Change{
				Action:   ActionNoOp,
				Resource: resource,
				Error:    err,
			}
		}
		plan.AddChange(change)
	}
	
	return plan, nil
}

// planResource creates a plan for a single resource
func (p *Planner) planResource(resource types.Resource) (Change, error) {
	// Get the provider for this resource type
	provider, err := p.registry.Get(resource.Type)
	if err != nil {
		return Change{}, fmt.Errorf("no provider found for resource type: %s", resource.Type)
	}
	
	// Validate the resource
	if err := provider.Validate(&resource); err != nil {
		return Change{}, fmt.Errorf("resource validation failed: %w", err)
	}
	
	// Read current state
	ctx := context.Background()
	currentState, err := provider.Read(ctx, &resource)
	if err != nil {
		return Change{}, fmt.Errorf("failed to read current state: %w", err)
	}
	
	// Calculate diff
	diff, err := provider.Diff(ctx, &resource, currentState)
	if err != nil {
		return Change{}, fmt.Errorf("failed to calculate diff: %w", err)
	}
	
	// Determine action based on diff
	action := p.determineAction(resource, currentState, diff)
	
	return Change{
		Action:   action,
		Resource: resource,
		Diff:     diff,
	}, nil
}

// determineAction determines what action should be taken based on the resource and diff
func (p *Planner) determineAction(resource types.Resource, currentState map[string]interface{}, diff *types.ResourceDiff) Action {
	// Check if resource should be absent
	if state, ok := resource.Properties["state"].(string); ok && state == "absent" {
		if currentState == nil {
			return ActionNoOp // Already absent
		}
		return ActionDelete
	}
	
	// Resource should be present
	if currentState == nil {
		return ActionCreate
	}
	
	// Check if there are any differences
	if diff == nil || diff.Action == types.ActionNoop {
		return ActionNoOp
	}
	
	// Map ResourceDiff action to our Action type
	switch diff.Action {
	case types.ActionCreate:
		return ActionCreate
	case types.ActionUpdate:
		return ActionUpdate
	case types.ActionDelete:
		return ActionDelete
	default:
		return ActionNoOp
	}
}
