package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <name> [args...]",
	Short: "Run an alias or workflow",
	Long: `Run an alias or workflow by name.
Supports parameter substitution for workflows.
Example:
  gnb run gs
  gnb run deploy main`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		runArgs := args[1:] // Additional arguments for parameter substitution

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Check if it's an alias
		if alias, exists := cfg.Aliases[name]; exists {
			fmt.Printf("Running alias: %s\n", alias.Cmd)
			// TODO: Actual command execution (will be implemented in workflow package)
			fmt.Printf("[SIMULATED] Executing: %s\n", alias.Cmd)
			return
		}

		// Check if it's a workflow
		if workflowDef, exists := cfg.Workflows[name]; exists {
			fmt.Printf("Running workflow: %s\n", workflowDef.Description)

			// Simple parameter substitution
			paramMap := make(map[string]string)
			for i, param := range workflowDef.Params {
				if i < len(runArgs) {
					paramMap["{"+param+"}"] = runArgs[i]
				} else {
					paramMap["{"+param+"}"] = "" // Empty if not provided
				}
			}

			// Execute each step
			for i, step := range workflowDef.Steps {
				stepCmd := step.Run
				// Apply parameter substitution
				for placeholder, value := range paramMap {
					stepCmd = strings.ReplaceAll(stepCmd, placeholder, value)
				}

				fmt.Printf("Step %d/%d: %s\n", i+1, len(workflowDef.Steps), stepCmd)

				// TODO: Actual command execution with proper error handling
				// For now, just simulate
				if step.Confirm {
					fmt.Printf("[SIMULATED] Would prompt for confirmation before: %s\n", stepCmd)
				} else {
					fmt.Printf("[SIMULATED] Executing: %s\n", stepCmd)
				}

				// Handle on_fail logic (simplified)
				if step.OnFail == "stop" {
					fmt.Printf("[SIMULATED] Would stop on failure\n")
				}
			}
			return
		}

		fmt.Printf("Alias or workflow '%s' not found\n", name)
		os.Exit(1)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
