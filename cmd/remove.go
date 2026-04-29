package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an alias or workflow",
	Long: `Remove an alias or workflow from the configuration.
Example:
  gnb remove gs
  gnb remove deploy`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Check if it's an alias
		if _, exists := cfg.Aliases[name]; exists {
			delete(cfg.Aliases, name)
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			cmd.Printf("Removed alias '%s'\n", name)
			return nil
		}

		// Check if it's a workflow
		if _, exists := cfg.Workflows[name]; exists {
			delete(cfg.Workflows, name)
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			cmd.Printf("Removed workflow '%s'\n", name)
			return nil
		}

		return fmt.Errorf("alias or workflow '%s' not found", name)
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
