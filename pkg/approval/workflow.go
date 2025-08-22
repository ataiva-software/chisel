package approval

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ataiva-software/forge/pkg/core"
)

// Status represents the status of an approval request
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
	StatusExpired  Status = "expired"
)

// Decision represents an approval decision
type Decision string

const (
	DecisionApprove Decision = "approve"
	DecisionReject  Decision = "reject"
)

// Condition represents a workflow condition
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Stage represents a workflow stage
type Stage struct {
	Name      string   `json:"name"`
	Approvers []string `json:"approvers"`
	Required  int      `json:"required"`
}

// Workflow represents an approval workflow
type Workflow struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Stages      []Stage     `json:"stages"`
	Conditions  []Condition `json:"conditions"`
	Timeout     time.Duration `json:"timeout"`
}

// Approval represents a single approval decision
type Approval struct {
	Approver  string    `json:"approver"`
	Decision  Decision  `json:"decision"`
	Comment   string    `json:"comment"`
	Timestamp time.Time `json:"timestamp"`
	Stage     string    `json:"stage"`
}

// ApprovalRequest represents an approval request
type ApprovalRequest struct {
	ID          string        `json:"id"`
	Submitter   string        `json:"submitter"`
	Action      string        `json:"action"`
	Module      *core.Module  `json:"module"`
	Workflow    string        `json:"workflow"`
	Status      Status        `json:"status"`
	Approvals   []Approval    `json:"approvals"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
	CurrentStage int          `json:"current_stage"`
}

// IsExpired returns whether the approval request has expired
func (r *ApprovalRequest) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// ApprovalManager manages approval workflows and requests
type ApprovalManager struct {
	workflows map[string]*Workflow
	requests  map[string]*ApprovalRequest
	enabled   bool
	mu        sync.RWMutex
}

// NewApprovalManager creates a new approval manager
func NewApprovalManager() *ApprovalManager {
	return &ApprovalManager{
		workflows: make(map[string]*Workflow),
		requests:  make(map[string]*ApprovalRequest),
		enabled:   true,
	}
}

// IsEnabled returns whether approval workflows are enabled
func (m *ApprovalManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// Enable enables approval workflows
func (m *ApprovalManager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables approval workflows
func (m *ApprovalManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// CreateWorkflow creates a new approval workflow
func (m *ApprovalManager) CreateWorkflow(workflow *Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}
	
	if len(workflow.Stages) == 0 {
		return fmt.Errorf("workflow must have at least one stage")
	}
	
	// Set default timeout if not specified
	if workflow.Timeout == 0 {
		workflow.Timeout = 24 * time.Hour
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.workflows[workflow.Name] = workflow
	return nil
}

// GetWorkflows returns all workflows
func (m *ApprovalManager) GetWorkflows() []*Workflow {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	workflows := make([]*Workflow, 0, len(m.workflows))
	for _, workflow := range m.workflows {
		// Return copies to prevent external modification
		workflowCopy := *workflow
		workflows = append(workflows, &workflowCopy)
	}
	
	return workflows
}

// RequiresApproval checks if an action requires approval
func (m *ApprovalManager) RequiresApproval(ctx context.Context, user, action string, module *core.Module) bool {
	if !m.IsEnabled() {
		return false
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Check each workflow to see if it matches
	for _, workflow := range m.workflows {
		if m.matchesWorkflow(workflow, module, action) {
			return true
		}
	}
	
	return false
}

// SubmitRequest submits an approval request
func (m *ApprovalManager) SubmitRequest(ctx context.Context, user, action string, module *core.Module) (string, error) {
	if !m.IsEnabled() {
		return "", fmt.Errorf("approval workflows are disabled")
	}
	
	// Find matching workflow
	var matchedWorkflow *Workflow
	m.mu.RLock()
	for _, workflow := range m.workflows {
		if m.matchesWorkflow(workflow, module, action) {
			matchedWorkflow = workflow
			break
		}
	}
	m.mu.RUnlock()
	
	if matchedWorkflow == nil {
		return "", fmt.Errorf("no matching workflow found")
	}
	
	// Create approval request
	requestID := generateRequestID()
	request := &ApprovalRequest{
		ID:           requestID,
		Submitter:    user,
		Action:       action,
		Module:       module,
		Workflow:     matchedWorkflow.Name,
		Status:       StatusPending,
		Approvals:    make([]Approval, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(matchedWorkflow.Timeout),
		CurrentStage: 0,
	}
	
	m.mu.Lock()
	m.requests[requestID] = request
	m.mu.Unlock()
	
	return requestID, nil
}

// GetRequest retrieves an approval request by ID
func (m *ApprovalManager) GetRequest(requestID string) (*ApprovalRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	request, exists := m.requests[requestID]
	if !exists {
		return nil, fmt.Errorf("approval request not found: %s", requestID)
	}
	
	// Return a copy to prevent external modification
	requestCopy := *request
	return &requestCopy, nil
}

// ApproveRequest approves an approval request
func (m *ApprovalManager) ApproveRequest(ctx context.Context, requestID, approver, comment string) error {
	return m.processApproval(ctx, requestID, approver, DecisionApprove, comment)
}

// RejectRequest rejects an approval request
func (m *ApprovalManager) RejectRequest(ctx context.Context, requestID, approver, comment string) error {
	return m.processApproval(ctx, requestID, approver, DecisionReject, comment)
}

// processApproval processes an approval or rejection
func (m *ApprovalManager) processApproval(ctx context.Context, requestID, approver string, decision Decision, comment string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	request, exists := m.requests[requestID]
	if !exists {
		return fmt.Errorf("approval request not found: %s", requestID)
	}
	
	if request.Status != StatusPending {
		return fmt.Errorf("request is not pending: %s", request.Status)
	}
	
	if request.IsExpired() {
		request.Status = StatusExpired
		return fmt.Errorf("request has expired")
	}
	
	// Get workflow
	workflow, exists := m.workflows[request.Workflow]
	if !exists {
		return fmt.Errorf("workflow not found: %s", request.Workflow)
	}
	
	// Check if approver is authorized for current stage
	currentStage := workflow.Stages[request.CurrentStage]
	authorized := false
	for _, validApprover := range currentStage.Approvers {
		if validApprover == approver {
			authorized = true
			break
		}
	}
	
	if !authorized {
		return fmt.Errorf("approver %s is not authorized for stage %s", approver, currentStage.Name)
	}
	
	// Add approval
	approval := Approval{
		Approver:  approver,
		Decision:  decision,
		Comment:   comment,
		Timestamp: time.Now(),
		Stage:     currentStage.Name,
	}
	
	request.Approvals = append(request.Approvals, approval)
	request.UpdatedAt = time.Now()
	
	// If rejected, mark as rejected
	if decision == DecisionReject {
		request.Status = StatusRejected
		return nil
	}
	
	// Check if current stage is complete
	stageApprovals := m.countStageApprovals(request, currentStage.Name)
	if stageApprovals >= currentStage.Required {
		// Move to next stage or complete
		if request.CurrentStage+1 >= len(workflow.Stages) {
			// All stages complete
			request.Status = StatusApproved
		} else {
			// Move to next stage
			request.CurrentStage++
		}
	}
	
	return nil
}

// countStageApprovals counts approvals for a specific stage
func (m *ApprovalManager) countStageApprovals(request *ApprovalRequest, stageName string) int {
	count := 0
	for _, approval := range request.Approvals {
		if approval.Stage == stageName && approval.Decision == DecisionApprove {
			count++
		}
	}
	return count
}

// ListRequests returns all approval requests
func (m *ApprovalManager) ListRequests() []*ApprovalRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	requests := make([]*ApprovalRequest, 0, len(m.requests))
	for _, request := range m.requests {
		// Return copies to prevent external modification
		requestCopy := *request
		requests = append(requests, &requestCopy)
	}
	
	return requests
}

// ListPendingRequests returns all pending approval requests
func (m *ApprovalManager) ListPendingRequests() []*ApprovalRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	requests := make([]*ApprovalRequest, 0)
	for _, request := range m.requests {
		if request.Status == StatusPending && !request.IsExpired() {
			// Return copies to prevent external modification
			requestCopy := *request
			requests = append(requests, &requestCopy)
		}
	}
	
	return requests
}

// matchesWorkflow checks if a module and action match a workflow
func (m *ApprovalManager) matchesWorkflow(workflow *Workflow, module *core.Module, action string) bool {
	// Check conditions
	for _, condition := range workflow.Conditions {
		if !m.evaluateCondition(condition, module, action) {
			return false
		}
	}
	
	return true
}

// evaluateCondition evaluates a workflow condition
func (m *ApprovalManager) evaluateCondition(condition Condition, module *core.Module, action string) bool {
	var value string
	
	// Get value based on field
	switch condition.Field {
	case "action":
		value = action
	case "environment":
		if module.Metadata.Labels != nil {
			value = module.Metadata.Labels["environment"]
		}
	case "module_name":
		value = module.Metadata.Name
	default:
		// Check in module labels
		if module.Metadata.Labels != nil {
			value = module.Metadata.Labels[condition.Field]
		}
	}
	
	// Evaluate condition
	switch condition.Operator {
	case "equals":
		return value == condition.Value
	case "not_equals":
		return value != condition.Value
	case "contains":
		return contains(value, condition.Value)
	case "not_contains":
		return !contains(value, condition.Value)
	default:
		return false
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || 
		   (len(s) > len(substr) && containsHelper(s[1:], substr)))
}

func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}

// ApprovalConfig represents approval workflow configuration
type ApprovalConfig struct {
	Enabled   bool                   `yaml:"enabled" json:"enabled"`
	Workflows map[string]*Workflow   `yaml:"workflows" json:"workflows"`
}

// DefaultApprovalConfig returns default approval configuration
func DefaultApprovalConfig() *ApprovalConfig {
	return &ApprovalConfig{
		Enabled:   true,
		Workflows: make(map[string]*Workflow),
	}
}
