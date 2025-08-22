package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ataiva-software/forge/pkg/types"
)

func TestParallelExecutor_CreateExecutionPlan(t *testing.T) {
	executor := NewParallelExecutor(5, 30*time.Second)
	
	tests := []struct {
		name          string
		diffs         []*types.ResourceDiff
		expectedBatches int
		wantErr       bool
	}{
		{
			name:            "empty diffs",
			diffs:           []*types.ResourceDiff{},
			expectedBatches: 0,
			wantErr:         false,
		},
		{
			name: "single resource",
			diffs: []*types.ResourceDiff{
				{ResourceID: "pkg.git", Action: types.ActionCreate},
			},
			expectedBatches: 1,
			wantErr:         false,
		},
		{
			name: "independent resources",
			diffs: []*types.ResourceDiff{
				{ResourceID: "pkg.git", Action: types.ActionCreate},
				{ResourceID: "pkg.vim", Action: types.ActionCreate},
			},
			expectedBatches: 1, // Can run in parallel
			wantErr:         false,
		},
		{
			name: "dependent resources",
			diffs: []*types.ResourceDiff{
				{ResourceID: "user.testuser", Action: types.ActionCreate},
				{ResourceID: "file.testfile", Action: types.ActionCreate},
			},
			expectedBatches: 2, // File depends on user
			wantErr:         false,
		},
		{
			name: "complex dependencies",
			diffs: []*types.ResourceDiff{
				{ResourceID: "pkg.nginx", Action: types.ActionCreate},
				{ResourceID: "service.nginx", Action: types.ActionUpdate},
				{ResourceID: "user.webuser", Action: types.ActionCreate},
				{ResourceID: "file.config", Action: types.ActionCreate},
			},
			expectedBatches: 2, // Actually should be 2: [pkg.nginx, user.webuser] then [service.nginx, file.config]
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := executor.CreateExecutionPlan(tt.diffs)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParallelExecutor.CreateExecutionPlan() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParallelExecutor.CreateExecutionPlan() unexpected error = %v", err)
				return
			}
			
			if len(plan.Batches) != tt.expectedBatches {
				t.Errorf("ParallelExecutor.CreateExecutionPlan() batches = %d, want %d", len(plan.Batches), tt.expectedBatches)
			}
		})
	}
}

func TestParallelExecutor_Execute(t *testing.T) {
	executor := NewParallelExecutor(2, 5*time.Second)
	
	tests := []struct {
		name           string
		diffs          []*types.ResourceDiff
		executorFunc   func(context.Context, *types.ResourceDiff) error
		expectSuccess  bool
		expectedResults int
	}{
		{
			name: "successful execution",
			diffs: []*types.ResourceDiff{
				{ResourceID: "pkg.git", Action: types.ActionCreate},
				{ResourceID: "pkg.vim", Action: types.ActionCreate},
			},
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				// Simulate work
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			expectSuccess:   true,
			expectedResults: 2,
		},
		{
			name: "execution with failure",
			diffs: []*types.ResourceDiff{
				{ResourceID: "pkg.git", Action: types.ActionCreate},
				{ResourceID: "pkg.fail", Action: types.ActionCreate},
			},
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				if diff.ResourceID == "pkg.fail" {
					return errors.New("simulated failure")
				}
				return nil
			},
			expectSuccess:   false,
			expectedResults: 2,
		},
		{
			name: "empty execution",
			diffs: []*types.ResourceDiff{},
			executorFunc: func(ctx context.Context, diff *types.ResourceDiff) error {
				return nil
			},
			expectSuccess:   true,
			expectedResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := executor.CreateExecutionPlan(tt.diffs)
			if err != nil {
				t.Fatalf("Failed to create execution plan: %v", err)
			}
			
			ctx := context.Background()
			results, err := executor.Execute(ctx, plan, tt.executorFunc)
			
			if tt.expectSuccess {
				if err != nil {
					t.Errorf("ParallelExecutor.Execute() unexpected error = %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("ParallelExecutor.Execute() expected error but got none")
				}
			}
			
			if len(results) != tt.expectedResults {
				t.Errorf("ParallelExecutor.Execute() results count = %d, want %d", len(results), tt.expectedResults)
			}
		})
	}
}

func TestDependencyGraph_TopologicalSort(t *testing.T) {
	tests := []struct {
		name          string
		setupGraph    func(*DependencyGraph)
		expectedBatches int
		wantErr       bool
	}{
		{
			name: "simple linear dependency",
			setupGraph: func(g *DependencyGraph) {
				g.AddNode("a", &types.ResourceDiff{ResourceID: "a"})
				g.AddNode("b", &types.ResourceDiff{ResourceID: "b"})
				g.AddEdge("a", "b") // a -> b
			},
			expectedBatches: 2,
			wantErr:         false,
		},
		{
			name: "parallel resources",
			setupGraph: func(g *DependencyGraph) {
				g.AddNode("a", &types.ResourceDiff{ResourceID: "a"})
				g.AddNode("b", &types.ResourceDiff{ResourceID: "b"})
				// No edges - can run in parallel
			},
			expectedBatches: 1,
			wantErr:         false,
		},
		{
			name: "diamond dependency",
			setupGraph: func(g *DependencyGraph) {
				g.AddNode("a", &types.ResourceDiff{ResourceID: "a"})
				g.AddNode("b", &types.ResourceDiff{ResourceID: "b"})
				g.AddNode("c", &types.ResourceDiff{ResourceID: "c"})
				g.AddNode("d", &types.ResourceDiff{ResourceID: "d"})
				g.AddEdge("a", "b") // a -> b
				g.AddEdge("a", "c") // a -> c
				g.AddEdge("b", "d") // b -> d
				g.AddEdge("c", "d") // c -> d
			},
			expectedBatches: 3, // a, then b&c in parallel, then d
			wantErr:         false,
		},
		{
			name: "circular dependency",
			setupGraph: func(g *DependencyGraph) {
				g.AddNode("a", &types.ResourceDiff{ResourceID: "a"})
				g.AddNode("b", &types.ResourceDiff{ResourceID: "b"})
				g.AddEdge("a", "b") // a -> b
				g.AddEdge("b", "a") // b -> a (creates cycle)
			},
			expectedBatches: 0,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := NewDependencyGraph()
			tt.setupGraph(graph)
			
			batches, err := graph.TopologicalSort()
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("DependencyGraph.TopologicalSort() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("DependencyGraph.TopologicalSort() unexpected error = %v", err)
				return
			}
			
			if len(batches) != tt.expectedBatches {
				t.Errorf("DependencyGraph.TopologicalSort() batches = %d, want %d", len(batches), tt.expectedBatches)
			}
		})
	}
}

func TestParallelExecutor_hasDependency(t *testing.T) {
	executor := NewParallelExecutor(5, 30*time.Second)
	
	tests := []struct {
		name       string
		dependent  *types.ResourceDiff
		dependency *types.ResourceDiff
		expected   bool
	}{
		{
			name:       "file depends on user",
			dependent:  &types.ResourceDiff{ResourceID: "file.config"},
			dependency: &types.ResourceDiff{ResourceID: "user.webuser"},
			expected:   true,
		},
		{
			name:       "service depends on package",
			dependent:  &types.ResourceDiff{ResourceID: "service.nginx"},
			dependency: &types.ResourceDiff{ResourceID: "pkg.nginx"},
			expected:   true,
		},
		{
			name:       "shell depends on package",
			dependent:  &types.ResourceDiff{ResourceID: "shell.setup"},
			dependency: &types.ResourceDiff{ResourceID: "pkg.git"},
			expected:   true,
		},
		{
			name:       "no dependency",
			dependent:  &types.ResourceDiff{ResourceID: "pkg.git"},
			dependency: &types.ResourceDiff{ResourceID: "pkg.vim"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.hasDependency(tt.dependent, tt.dependency)
			if result != tt.expected {
				t.Errorf("ParallelExecutor.hasDependency() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParallelExecutor_getResourceType(t *testing.T) {
	executor := NewParallelExecutor(5, 30*time.Second)
	
	tests := []struct {
		name       string
		resourceID string
		expected   string
	}{
		{
			name:       "package resource",
			resourceID: "pkg.nginx",
			expected:   "pkg",
		},
		{
			name:       "file resource",
			resourceID: "file.config",
			expected:   "file",
		},
		{
			name:       "service resource",
			resourceID: "service.nginx",
			expected:   "service",
		},
		{
			name:       "user resource",
			resourceID: "user.webuser",
			expected:   "user",
		},
		{
			name:       "shell resource",
			resourceID: "shell.setup-script",
			expected:   "shell",
		},
		{
			name:       "no dot in name",
			resourceID: "invalid",
			expected:   "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.getResourceType(tt.resourceID)
			if result != tt.expected {
				t.Errorf("ParallelExecutor.getResourceType() = %v, want %v", result, tt.expected)
			}
		})
	}
}
