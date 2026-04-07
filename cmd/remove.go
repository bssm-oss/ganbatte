package cmd

import (
	"fmt"
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Check if it's an alias
		if _, exists := cfg.Aliases[name]; exists {
			delete(cfg.Aliases, name)
			if err := cfg.Save(); err != nil {
				fmt.Printf("Error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Removed alias '%s'\n", name)
			return
		}

		// Check if it's a workflow
		if _, exists := cfg.Workflows[name]; exists {
			delete(cfg.Workflows, name)
			if err := cfg.Save(); err != nil {
				fmt.Printf("Error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Removed workflow '%s'\n", name)
			return
		}

		fmt.Printf("Alias or workflow '%s' not found\n", name)
		os.Exit(1)
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
