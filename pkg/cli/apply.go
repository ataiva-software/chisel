package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ataiva-software/chisel/pkg/core"
	"github.com/ataiva-software/chisel/pkg/inventory"
	"github.com/ataiva-software/chisel/pkg/providers"
	"github.com/ataiva-software/chisel/pkg/ssh"
	"github.com/ataiva-software/chisel/pkg/types"
)

var (
	applyModuleFile    string
	applyInventoryFile string
	applyDryRun        bool
	applyAutoApprove   bool
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply changes to infrastructure",
	Long: `Apply changes to bring the infrastructure to the desired state
defined in the module.

The apply command first creates a plan, shows what changes will be made,
and then applies those changes (unless --dry-run is specified).`,
	RunE: runApply,
}

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().StringVarP(&applyModuleFile, "module", "m", "", "Path to module file (required)")
	applyCmd.Flags().StringVarP(&applyInventoryFile, "inventory", "i", "", "Path to inventory file")
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would be done without actually applying changes")
	applyCmd.Flags().BoolVar(&applyAutoApprove, "auto-approve", false, "Skip interactive approval of plan")
	
	applyCmd.MarkFlagRequired("module")
}

func runApply(cmd *cobra.Command, args []string) error {
	// Load the module
	module, err := core.LoadModuleFromFile(applyModuleFile)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Load inventory if specified
	var inv *inventory.Inventory
	if applyInventoryFile != "" {
		inv, err = inventory.LoadInventoryFromFile(applyInventoryFile)
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
	fmt.Println("Creating execution plan...")
	plan, err := planner.CreatePlan(module)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Display plan
	summary := plan.Summary()
	fmt.Printf("\nPlan: %d to add, %d to change, %d to destroy\n\n", 
		summary.ToCreate, summary.ToUpdate, summary.ToDelete)

	// Show changes
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

	// Check if there are any changes to apply
	if !plan.HasChanges() {
		fmt.Println("No changes. Infrastructure is up-to-date.")
		return nil
	}

	// Check for errors in plan
	if summary.Errors > 0 {
		fmt.Printf("Cannot apply plan due to %d error(s). Please fix the errors and try again.\n", summary.Errors)
		return fmt.Errorf("plan contains errors")
	}

	// Dry run mode
	if applyDryRun {
		fmt.Println("This was a dry run. No changes were actually applied.")
		return nil
	}

	// Ask for confirmation unless auto-approve is set
	if !applyAutoApprove {
		fmt.Print("\nDo you want to perform these actions? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			fmt.Println("Apply cancelled.")
			return nil
		}
	}

	// Apply the plan
	fmt.Println("\nApplying changes...")
	executor := core.NewExecutor(registry)
	
	result, err := executor.ExecutePlan(context.Background(), plan)
	if err != nil {
		return fmt.Errorf("failed to execute plan: %w", err)
	}

	// Display results
	fmt.Printf("\nApply complete! Resources: %d added, %d changed, %d destroyed.\n",
		countActionResults(result, core.ActionCreate),
		countActionResults(result, core.ActionUpdate),
		countActionResults(result, core.ActionDelete))

	fmt.Printf("Duration: %v\n", result.Summary.Duration)

	// Show any failures
	if result.Summary.Failed > 0 {
		fmt.Printf("\nFailed changes:\n")
		for _, changeResult := range result.Changes {
			if !changeResult.Success && changeResult.Error != nil {
				fmt.Printf("✗ %s.%s: %v\n", 
					changeResult.Change.Resource.Type, 
					changeResult.Change.Resource.Name, 
					changeResult.Error)
			}
		}
	}

	// Show inventory info if loaded
	if inv != nil {
		fmt.Printf("\nTargeted %d host groups\n", len(inv.Targets))
	}

	return nil
}

func countActionResults(result *core.ExecutionResult, action core.Action) int {
	count := 0
	for _, changeResult := range result.Changes {
		if changeResult.Success && changeResult.Change.Action == action {
			count++
		}
	}
	return count
}
