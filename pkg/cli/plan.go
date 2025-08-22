package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ataiva-software/forge/pkg/core"
	"github.com/ataiva-software/forge/pkg/inventory"
	"github.com/ataiva-software/forge/pkg/providers"
	"github.com/ataiva-software/forge/pkg/ssh"
	"github.com/ataiva-software/forge/pkg/types"
)

var (
	planModuleFile    string
	planInventoryFile string
	planOutputFile    string
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create an execution plan",
	Long: `Create an execution plan showing what changes will be made to bring
the infrastructure to the desired state defined in the module.

The plan command reads a module file and optionally an inventory file,
then shows what actions will be taken without actually applying them.`,
	RunE: runPlan,
}

func init() {
	rootCmd.AddCommand(planCmd)

	planCmd.Flags().StringVarP(&planModuleFile, "module", "m", "", "Path to module file (required)")
	planCmd.Flags().StringVarP(&planInventoryFile, "inventory", "i", "", "Path to inventory file")
	planCmd.Flags().StringVarP(&planOutputFile, "output", "o", "", "Path to save plan output (JSON format)")
	
	planCmd.MarkFlagRequired("module")
}

func runPlan(cmd *cobra.Command, args []string) error {
	// Load the module
	module, err := core.LoadModuleFromFile(planModuleFile)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Load inventory if specified
	var inv *inventory.Inventory
	if planInventoryFile != "" {
		inv, err = inventory.LoadInventoryFromFile(planInventoryFile)
		if err != nil {
			return fmt.Errorf("failed to load inventory: %w", err)
		}
	}

	// Create provider registry and register core providers
	registry := types.NewProviderRegistry()
	mockExecutor := ssh.NewMockExecutor()
	// Connect the mock executor
	if err := mockExecutor.Connect(context.Background()); err != nil {
		return fmt.Errorf("failed to connect mock executor: %w", err)
	}
	if err := registry.Register(providers.NewFileProvider(mockExecutor)); err != nil {
		return fmt.Errorf("failed to register file provider: %w", err)
	}
	if err := registry.Register(providers.NewPkgProvider(mockExecutor)); err != nil {
		return fmt.Errorf("failed to register package provider: %w", err)
	}
	if err := registry.Register(providers.NewServiceProvider(mockExecutor)); err != nil {
		return fmt.Errorf("failed to register service provider: %w", err)
	}
	if err := registry.Register(providers.NewUserProvider(mockExecutor)); err != nil {
		return fmt.Errorf("failed to register user provider: %w", err)
	}
	if err := registry.Register(providers.NewShellProvider(mockExecutor)); err != nil {
		return fmt.Errorf("failed to register shell provider: %w", err)
	}

	// Create planner
	planner := core.NewPlanner(registry)

	// Create plan
	plan, err := planner.CreatePlan(module)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Display plan summary
	summary := plan.Summary()
	fmt.Printf("Plan: %d to add, %d to change, %d to destroy\n\n", 
		summary.ToCreate, summary.ToUpdate, summary.ToDelete)

	// Display changes
	for _, change := range plan.Changes {
		if change.Error != nil {
			fmt.Printf("✗ %s.%s\n", change.Resource.Type, change.Resource.Name)
			fmt.Printf("  Error: %v\n\n", change.Error)
			continue
		}

		symbol := getChangeSymbol(change.Action)
		fmt.Printf("%s %s.%s\n", symbol, change.Resource.Type, change.Resource.Name)
		
		if change.Action != core.ActionNoOp {
			displayChangeDiff(change)
		}
		fmt.Println()
	}

	// Show inventory info if loaded
	if inv != nil {
		fmt.Printf("Inventory: %d target groups\n", len(inv.Targets))
		for name, group := range inv.Targets {
			hosts, _ := group.GetHosts()
			if len(hosts) > 0 {
				fmt.Printf("  %s: %d hosts\n", name, len(hosts))
			} else {
				fmt.Printf("  %s: selector '%s'\n", name, group.Selector)
			}
		}
	}

	// Save plan to file if requested
	if planOutputFile != "" {
		if err := savePlanToFile(plan, planOutputFile); err != nil {
			return fmt.Errorf("failed to save plan: %w", err)
		}
		fmt.Printf("\nPlan saved to: %s\n", planOutputFile)
	}

	return nil
}

func getChangeSymbol(action core.Action) string {
	switch action {
	case core.ActionCreate:
		return "+"
	case core.ActionUpdate:
		return "~"
	case core.ActionDelete:
		return "-"
	case core.ActionNoOp:
		return "="
	default:
		return "?"
	}
}

func displayChangeDiff(change core.Change) {
	if change.Diff == nil {
		return
	}

	// Display the diff information
	switch change.Action {
	case core.ActionCreate:
		fmt.Printf("  (will be created)\n")
	case core.ActionUpdate:
		fmt.Printf("  (will be updated)\n")
	case core.ActionDelete:
		fmt.Printf("  (will be destroyed)\n")
	}

	// Show some key properties
	if path, ok := change.Resource.Properties["path"]; ok {
		fmt.Printf("  path: %v\n", path)
	}
	if state, ok := change.Resource.Properties["state"]; ok {
		fmt.Printf("  state: %v\n", state)
	}
}

func savePlanToFile(plan *core.Plan, filename string) error {
	// For now, just save as JSON
	// TODO: Implement proper plan serialization
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Simple JSON-like output for now
	fmt.Fprintf(file, "{\n")
	fmt.Fprintf(file, "  \"changes\": %d,\n", len(plan.Changes))
	fmt.Fprintf(file, "  \"summary\": {\n")
	summary := plan.Summary()
	fmt.Fprintf(file, "    \"to_create\": %d,\n", summary.ToCreate)
	fmt.Fprintf(file, "    \"to_update\": %d,\n", summary.ToUpdate)
	fmt.Fprintf(file, "    \"to_delete\": %d,\n", summary.ToDelete)
	fmt.Fprintf(file, "    \"no_changes\": %d,\n", summary.NoChanges)
	fmt.Fprintf(file, "    \"errors\": %d\n", summary.Errors)
	fmt.Fprintf(file, "  }\n")
	fmt.Fprintf(file, "}\n")

	return nil
}
