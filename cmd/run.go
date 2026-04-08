package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/bssm-oss/ganbatte/internal/workflow"
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
  gnb run deploy main
  gnb run deploy main --dry-run`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		runArgs := args[1:]
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Check if it's an alias
		if alias, exists := cfg.Aliases[name]; exists {
			if dryRun {
				cmd.Printf("[dry-run] %s\n", alias.Cmd)
				if workflow.IsDestructive(alias.Cmd) {
					cmd.Printf("[DESTRUCTIVE] command detected\n")
				}
				return nil
			}
			cmd.Printf("Running: %s\n", alias.Cmd)
			ex := &workflow.RealExecutor{}
			return ex.Execute(alias.Cmd)
		}

		// Check if it's a workflow
		if wfDef, exists := cfg.Workflows[name]; exists {
			wf := workflow.Workflow{
				Description: wfDef.Description,
				Params:      wfDef.Params,
				Tags:        wfDef.Tags,
			}
			for _, s := range wfDef.Steps {
				wf.Steps = append(wf.Steps, workflow.Step{
					Run:     s.Run,
					OnFail:  s.OnFail,
					Confirm: s.Confirm,
				})
			}

			if dryRun {
				cmd.Printf("Workflow: %s (dry-run)\n", wfDef.Description)
			} else {
				cmd.Printf("Running workflow: %s\n", wfDef.Description)
			}

			return workflow.Run(wf, runArgs, &workflow.RealExecutor{}, workflow.RunOptions{
				DryRun: dryRun,
				Writer: cmd.OutOrStdout(),
			})
		}

		return fmt.Errorf("alias or workflow '%s' not found", name)
	},
}

func init() {
	runCmd.Flags().Bool("dry-run", false, "Preview steps without executing")
	RootCmd.AddCommand(runCmd)
}
