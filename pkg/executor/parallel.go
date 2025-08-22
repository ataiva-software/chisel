package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ataiva-software/forge/pkg/types"
)

// ExecutionResult represents the result of executing a resource
type ExecutionResult struct {
	ResourceID string        `json:"resource_id"`
	Success    bool          `json:"success"`
	Error      error         `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	Changes    []string      `json:"changes,omitempty"`
}

// ExecutionPlan represents a plan for parallel execution
type ExecutionPlan struct {
	Batches []ExecutionBatch `json:"batches"`
}

// ExecutionBatch represents a batch of resources that can be executed in parallel
type ExecutionBatch struct {
	Resources []*types.ResourceDiff `json:"resources"`
	BatchID   int                   `json:"batch_id"`
}

// ParallelExecutor executes resources in parallel while respecting dependencies
type ParallelExecutor struct {
	maxConcurrency int
	timeout        time.Duration
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxConcurrency int, timeout time.Duration) *ParallelExecutor {
	if maxConcurrency <= 0 {
		maxConcurrency = 10 // default
	}
	if timeout <= 0 {
		timeout = 30 * time.Minute // default
	}
	
	return &ParallelExecutor{
		maxConcurrency: maxConcurrency,
		timeout:        timeout,
	}
}

// CreateExecutionPlan creates an execution plan that respects dependencies
func (e *ParallelExecutor) CreateExecutionPlan(diffs []*types.ResourceDiff) (*ExecutionPlan, error) {
	if len(diffs) == 0 {
		return &ExecutionPlan{Batches: []ExecutionBatch{}}, nil
	}
	
	// Build dependency graph
	graph := NewDependencyGraph()
	for _, diff := range diffs {
		graph.AddNode(diff.ResourceID, diff)
	}
	
	// Add dependency edges (this would be enhanced with actual dependency parsing)
	// For now, we'll use a simple heuristic: files depend on directories, services depend on packages
	for _, diff := range diffs {
		for _, other := range diffs {
			if diff.ResourceID != other.ResourceID && e.hasDependency(diff, other) {
				graph.AddEdge(other.ResourceID, diff.ResourceID)
			}
		}
	}
	
	// Perform topological sort to get execution order
	batches, err := graph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}
	
	plan := &ExecutionPlan{
		Batches: make([]ExecutionBatch, len(batches)),
	}
	
	for i, batch := range batches {
		plan.Batches[i] = ExecutionBatch{
			BatchID:   i,
			Resources: batch,
		}
	}
	
	return plan, nil
}

// Execute executes the plan in parallel
func (e *ParallelExecutor) Execute(ctx context.Context, plan *ExecutionPlan, executor func(context.Context, *types.ResourceDiff) error) ([]ExecutionResult, error) {
	var allResults []ExecutionResult
	var mu sync.Mutex
	
	// Execute each batch sequentially, but resources within a batch in parallel
	for batchID, batch := range plan.Batches {
		fmt.Printf("Executing batch %d with %d resources...\n", batchID, len(batch.Resources))
		
		batchResults, err := e.executeBatch(ctx, batch, executor)
		if err != nil {
			return allResults, fmt.Errorf("batch %d failed: %w", batchID, err)
		}
		
		mu.Lock()
		allResults = append(allResults, batchResults...)
		mu.Unlock()
		
		// Check if any resource in this batch failed
		for _, result := range batchResults {
			if !result.Success {
				return allResults, fmt.Errorf("resource %s failed: %v", result.ResourceID, result.Error)
			}
		}
	}
	
	return allResults, nil
}

// executeBatch executes a single batch of resources in parallel
func (e *ParallelExecutor) executeBatch(ctx context.Context, batch ExecutionBatch, executor func(context.Context, *types.ResourceDiff) error) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, len(batch.Resources))
	
	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, e.maxConcurrency)
	var wg sync.WaitGroup
	
	// Execute resources in parallel
	for i, resource := range batch.Resources {
		wg.Add(1)
		go func(index int, res *types.ResourceDiff) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Create context with timeout
			execCtx, cancel := context.WithTimeout(ctx, e.timeout)
			defer cancel()
			
			// Execute the resource
			start := time.Now()
			err := executor(execCtx, res)
			duration := time.Since(start)
			
			results[index] = ExecutionResult{
				ResourceID: res.ResourceID,
				Success:    err == nil,
				Error:      err,
				Duration:   duration,
			}
			
			if err == nil {
				fmt.Printf("✓ %s (%.2fs)\n", res.ResourceID, duration.Seconds())
			} else {
				fmt.Printf("✗ %s (%.2fs): %v\n", res.ResourceID, duration.Seconds(), err)
			}
		}(i, resource)
	}
	
	// Wait for all resources to complete
	wg.Wait()
	
	return results, nil
}

// hasDependency determines if one resource depends on another
// This is a simple heuristic - in a real implementation, this would parse
// explicit dependencies from the resource configuration
func (e *ParallelExecutor) hasDependency(dependent, dependency *types.ResourceDiff) bool {
	// Extract resource type from ResourceID (format: "type.name")
	depType := e.getResourceType(dependent.ResourceID)
	depOnType := e.getResourceType(dependency.ResourceID)
	
	// Simple dependency rules:
	// 1. Files depend on users (for ownership)
	// 2. Services depend on packages
	// 3. Shell commands depend on files, packages, and users
	
	switch depType {
	case "file":
		// Files depend on users (for ownership)
		return depOnType == "user"
	case "service":
		// Services depend on packages
		return depOnType == "pkg"
	case "shell":
		// Shell commands depend on packages, files, and users
		return depOnType == "pkg" || depOnType == "file" || depOnType == "user"
	}
	
	return false
}

// getResourceType extracts the resource type from a ResourceID
func (e *ParallelExecutor) getResourceType(resourceID string) string {
	for i, char := range resourceID {
		if char == '.' {
			return resourceID[:i]
		}
	}
	return resourceID
}

// DependencyGraph represents a directed acyclic graph for dependency resolution
type DependencyGraph struct {
	nodes map[string]*types.ResourceDiff
	edges map[string][]string
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*types.ResourceDiff),
		edges: make(map[string][]string),
	}
}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(id string, resource *types.ResourceDiff) {
	g.nodes[id] = resource
	if g.edges[id] == nil {
		g.edges[id] = []string{}
	}
}

// AddEdge adds a directed edge from 'from' to 'to'
func (g *DependencyGraph) AddEdge(from, to string) {
	g.edges[from] = append(g.edges[from], to)
}

// TopologicalSort performs a topological sort and returns batches of resources
// that can be executed in parallel
func (g *DependencyGraph) TopologicalSort() ([][]*types.ResourceDiff, error) {
	// Calculate in-degrees
	inDegree := make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = 0
	}
	
	for _, neighbors := range g.edges {
		for _, neighbor := range neighbors {
			inDegree[neighbor]++
		}
	}
	
	var batches [][]*types.ResourceDiff
	visited := make(map[string]bool)
	
	// Process nodes in batches
	for len(visited) < len(g.nodes) {
		var currentBatch []*types.ResourceDiff
		
		// Find all nodes with in-degree 0
		for node := range g.nodes {
			if !visited[node] && inDegree[node] == 0 {
				currentBatch = append(currentBatch, g.nodes[node])
				visited[node] = true
			}
		}
		
		if len(currentBatch) == 0 {
			// Circular dependency detected
			return nil, fmt.Errorf("circular dependency detected")
		}
		
		batches = append(batches, currentBatch)
		
		// Update in-degrees for next iteration
		for _, resource := range currentBatch {
			for _, neighbor := range g.edges[resource.ResourceID] {
				inDegree[neighbor]--
			}
		}
	}
	
	return batches, nil
}
