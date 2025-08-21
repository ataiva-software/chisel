package approval

import (
	"context"
	"testing"
	"time"

	"github.com/ataiva-software/chisel/pkg/core"
	"github.com/ataiva-software/chisel/pkg/types"
)

func TestApprovalManager_New(t *testing.T) {
	manager := NewApprovalManager()
	
	if manager == nil {
		t.Fatal("Expected non-nil approval manager")
	}
	
	if !manager.IsEnabled() {
		t.Error("Expected approval manager to be enabled by default")
	}
}

func TestApprovalManager_CreateWorkflow(t *testing.T) {
	manager := NewApprovalManager()
	
	workflow := &Workflow{
		Name:        "production-changes",
		Description: "Approval workflow for production changes",
		Stages: []Stage{
			{
				Name:      "security-review",
				Approvers: []string{"security-team"},
				Required:  1,
			},
			{
				Name:      "ops-review",
				Approvers: []string{"ops-lead", "ops-manager"},
				Required:  1,
			},
		},
		Conditions: []Condition{
			{
				Field:    "environment",
				Operator: "equals",
				Value:    "production",
			},
		},
	}
	
	err := manager.CreateWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}
	
	workflows := manager.GetWorkflows()
	if len(workflows) != 1 {
		t.Errorf("Expected 1 workflow, got %d", len(workflows))
	}
	
	if workflows[0].Name != "production-changes" {
		t.Errorf("Expected workflow name 'production-changes', got '%s'", workflows[0].Name)
	}
}

func TestApprovalManager_SubmitRequest(t *testing.T) {
	manager := NewApprovalManager()
	
	// Create workflow
	workflow := &Workflow{
		Name: "test-workflow",
		Stages: []Stage{
			{
				Name:      "review",
				Approvers: []string{"reviewer"},
				Required:  1,
			},
		},
	}
	manager.CreateWorkflow(workflow)
	
	// Create module for approval
	module := &core.Module{
		APIVersion: "ataiva.com/chisel/v1",
		Kind:       "Module",
		Metadata: core.ModuleMetadata{
			Name: "test-module",
			Labels: map[string]string{
				"environment": "production",
			},
		},
		Spec: core.ModuleSpec{
			Resources: []types.Resource{
				{
					Type: "file",
					Name: "test-file",
					Properties: map[string]interface{}{
						"path": "/tmp/test.txt",
					},
				},
			},
		},
	}
	
	ctx := context.Background()
	requestID, err := manager.SubmitRequest(ctx, "user1", "apply", module)
	if err != nil {
		t.Fatalf("Failed to submit approval request: %v", err)
	}
	
	if requestID == "" {
		t.Error("Expected non-empty request ID")
	}
	
	// Verify request was created
	request, err := manager.GetRequest(requestID)
	if err != nil {
		t.Fatalf("Failed to get approval request: %v", err)
	}
	
	if request.Submitter != "user1" {
		t.Errorf("Expected submitter 'user1', got '%s'", request.Submitter)
	}
	
	if request.Action != "apply" {
		t.Errorf("Expected action 'apply', got '%s'", request.Action)
	}
	
	if request.Status != StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", request.Status)
	}
}

func TestApprovalManager_ApproveRequest(t *testing.T) {
	manager := NewApprovalManager()
	
	// Create simple workflow
	workflow := &Workflow{
		Name: "simple-approval",
		Stages: []Stage{
			{
				Name:      "review",
				Approvers: []string{"approver1"},
				Required:  1,
			},
		},
	}
	manager.CreateWorkflow(workflow)
	
	// Submit request
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test"},
		Spec:     core.ModuleSpec{Resources: []types.Resource{}},
	}
	
	ctx := context.Background()
	requestID, _ := manager.SubmitRequest(ctx, "user1", "apply", module)
	
	// Approve request
	err := manager.ApproveRequest(ctx, requestID, "approver1", "Looks good to me")
	if err != nil {
		t.Fatalf("Failed to approve request: %v", err)
	}
	
	// Verify request is approved
	request, err := manager.GetRequest(requestID)
	if err != nil {
		t.Fatalf("Failed to get request: %v", err)
	}
	
	if request.Status != StatusApproved {
		t.Errorf("Expected status 'approved', got '%s'", request.Status)
	}
	
	if len(request.Approvals) != 1 {
		t.Errorf("Expected 1 approval, got %d", len(request.Approvals))
	}
	
	if request.Approvals[0].Approver != "approver1" {
		t.Errorf("Expected approver 'approver1', got '%s'", request.Approvals[0].Approver)
	}
}

func TestApprovalManager_RejectRequest(t *testing.T) {
	manager := NewApprovalManager()
	
	workflow := &Workflow{
		Name: "test-workflow",
		Stages: []Stage{
			{
				Name:      "review",
				Approvers: []string{"approver1"},
				Required:  1,
			},
		},
	}
	manager.CreateWorkflow(workflow)
	
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test"},
		Spec:     core.ModuleSpec{Resources: []types.Resource{}},
	}
	
	ctx := context.Background()
	requestID, _ := manager.SubmitRequest(ctx, "user1", "apply", module)
	
	// Reject request
	err := manager.RejectRequest(ctx, requestID, "approver1", "Security concerns")
	if err != nil {
		t.Fatalf("Failed to reject request: %v", err)
	}
	
	// Verify request is rejected
	request, err := manager.GetRequest(requestID)
	if err != nil {
		t.Fatalf("Failed to get request: %v", err)
	}
	
	if request.Status != StatusRejected {
		t.Errorf("Expected status 'rejected', got '%s'", request.Status)
	}
	
	if len(request.Approvals) != 1 {
		t.Errorf("Expected 1 approval record, got %d", len(request.Approvals))
	}
	
	if request.Approvals[0].Decision != DecisionReject {
		t.Errorf("Expected decision 'reject', got '%s'", request.Approvals[0].Decision)
	}
}

func TestApprovalManager_MultiStageWorkflow(t *testing.T) {
	manager := NewApprovalManager()
	
	// Create multi-stage workflow
	workflow := &Workflow{
		Name: "multi-stage",
		Stages: []Stage{
			{
				Name:      "security",
				Approvers: []string{"security-lead"},
				Required:  1,
			},
			{
				Name:      "operations",
				Approvers: []string{"ops-lead"},
				Required:  1,
			},
		},
	}
	manager.CreateWorkflow(workflow)
	
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test"},
		Spec:     core.ModuleSpec{Resources: []types.Resource{}},
	}
	
	ctx := context.Background()
	requestID, _ := manager.SubmitRequest(ctx, "user1", "apply", module)
	
	// First approval
	err := manager.ApproveRequest(ctx, requestID, "security-lead", "Security approved")
	if err != nil {
		t.Fatalf("Failed to approve first stage: %v", err)
	}
	
	// Check status - should still be pending
	request, _ := manager.GetRequest(requestID)
	if request.Status != StatusPending {
		t.Errorf("Expected status 'pending' after first approval, got '%s'", request.Status)
	}
	
	// Second approval
	err = manager.ApproveRequest(ctx, requestID, "ops-lead", "Operations approved")
	if err != nil {
		t.Fatalf("Failed to approve second stage: %v", err)
	}
	
	// Check status - should now be approved
	request, _ = manager.GetRequest(requestID)
	if request.Status != StatusApproved {
		t.Errorf("Expected status 'approved' after all approvals, got '%s'", request.Status)
	}
	
	if len(request.Approvals) != 2 {
		t.Errorf("Expected 2 approvals, got %d", len(request.Approvals))
	}
}

func TestApprovalManager_RequiresApproval(t *testing.T) {
	manager := NewApprovalManager()
	
	// Create workflow with conditions
	workflow := &Workflow{
		Name: "production-only",
		Stages: []Stage{
			{
				Name:      "review",
				Approvers: []string{"reviewer"},
				Required:  1,
			},
		},
		Conditions: []Condition{
			{
				Field:    "environment",
				Operator: "equals",
				Value:    "production",
			},
		},
	}
	manager.CreateWorkflow(workflow)
	
	tests := []struct {
		name        string
		module      *core.Module
		action      string
		expected    bool
	}{
		{
			name: "production module requires approval",
			module: &core.Module{
				Metadata: core.ModuleMetadata{
					Name: "prod-module",
					Labels: map[string]string{
						"environment": "production",
					},
				},
			},
			action:   "apply",
			expected: true,
		},
		{
			name: "development module does not require approval",
			module: &core.Module{
				Metadata: core.ModuleMetadata{
					Name: "dev-module",
					Labels: map[string]string{
						"environment": "development",
					},
				},
			},
			action:   "apply",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := manager.RequiresApproval(ctx, "user1", tt.action, tt.module)
			
			if result != tt.expected {
				t.Errorf("Expected RequiresApproval to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestApprovalManager_EnableDisable(t *testing.T) {
	manager := NewApprovalManager()
	
	if !manager.IsEnabled() {
		t.Error("Expected approval manager to be enabled by default")
	}
	
	manager.Disable()
	if manager.IsEnabled() {
		t.Error("Expected approval manager to be disabled after Disable()")
	}
	
	// When disabled, no approval should be required
	module := &core.Module{
		Metadata: core.ModuleMetadata{
			Name: "test",
			Labels: map[string]string{
				"environment": "production",
			},
		},
	}
	
	ctx := context.Background()
	result := manager.RequiresApproval(ctx, "user1", "apply", module)
	if result {
		t.Error("Expected no approval required when manager is disabled")
	}
	
	manager.Enable()
	if !manager.IsEnabled() {
		t.Error("Expected approval manager to be enabled after Enable()")
	}
}

func TestApprovalManager_ListRequests(t *testing.T) {
	manager := NewApprovalManager()
	
	workflow := &Workflow{
		Name: "test-workflow",
		Stages: []Stage{
			{
				Name:      "review",
				Approvers: []string{"reviewer"},
				Required:  1,
			},
		},
	}
	manager.CreateWorkflow(workflow)
	
	module := &core.Module{
		Metadata: core.ModuleMetadata{Name: "test"},
		Spec:     core.ModuleSpec{Resources: []types.Resource{}},
	}
	
	ctx := context.Background()
	
	// Submit multiple requests
	requestID1, _ := manager.SubmitRequest(ctx, "user1", "apply", module)
	requestID2, _ := manager.SubmitRequest(ctx, "user2", "plan", module)
	
	// List all requests
	requests := manager.ListRequests()
	if len(requests) != 2 {
		t.Errorf("Expected 2 requests, got %d", len(requests))
	}
	
	// List pending requests
	pendingRequests := manager.ListPendingRequests()
	if len(pendingRequests) != 2 {
		t.Errorf("Expected 2 pending requests, got %d", len(pendingRequests))
	}
	
	// Approve one request
	manager.ApproveRequest(ctx, requestID1, "reviewer", "Approved")
	
	// List pending requests again
	pendingRequests = manager.ListPendingRequests()
	if len(pendingRequests) != 1 {
		t.Errorf("Expected 1 pending request after approval, got %d", len(pendingRequests))
	}
	
	if pendingRequests[0].ID != requestID2 {
		t.Errorf("Expected pending request ID '%s', got '%s'", requestID2, pendingRequests[0].ID)
	}
}

func TestApprovalRequest_IsExpired(t *testing.T) {
	request := &ApprovalRequest{
		ID:        "test-request",
		CreatedAt: time.Now().Add(-25 * time.Hour), // 25 hours ago
		ExpiresAt: time.Now().Add(-1 * time.Hour),  // 1 hour ago
	}
	
	if !request.IsExpired() {
		t.Error("Expected request to be expired")
	}
	
	request.ExpiresAt = time.Now().Add(1 * time.Hour) // 1 hour from now
	if request.IsExpired() {
		t.Error("Expected request to not be expired")
	}
}
