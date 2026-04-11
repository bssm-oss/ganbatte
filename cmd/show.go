package cmd

import (
	"fmt"

	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show detailed information about an alias or workflow",
	Long: `Show detailed information about an alias or workflow.
Example:
  gnb show gs
  gnb show deploy`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if alias, exists := cfg.Aliases[name]; exists {
			cmd.Printf("Alias: %s\n", name)
			cmd.Printf("  Command: %s\n", alias.Cmd)
			return nil
		}

		if wf, exists := cfg.Workflows[name]; exists {
			cmd.Printf("Workflow: %s\n", name)
			if wf.Description != "" {
				cmd.Printf("  Description: %s\n", wf.Description)
			}
			if len(wf.Params) > 0 {
				cmd.Printf("  Params: %v\n", wf.Params)
			}
			if len(wf.Steps) > 0 {
				cmd.Println("  Steps:")
				for i, step := range wf.Steps {
					cmd.Printf("    %d. %s\n", i+1, step.Run)
					if step.OnFail != "" {
						cmd.Printf("       on_fail: %s\n", step.OnFail)
					}
					if step.Confirm {
						cmd.Printf("       confirm: true\n")
					}
				}
			}
			if len(wf.Tags) > 0 {
				cmd.Printf("  Tags: %v\n", wf.Tags)
			}
			return nil
		}

		return fmt.Errorf("'%s' not found", name)
	},
}

func init() {
	RootCmd.AddCommand(showCmd)
}
