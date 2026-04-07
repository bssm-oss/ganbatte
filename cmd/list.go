package cmd

import (
	"fmt"
	"os"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List aliases and workflows",
	Long: `List all aliases and workflows in the configuration.
Supports filtering by scope and tags (tags filtering to be implemented in v0.2).
Example:
  gnb list
  gnb list --scope global
  gnb list --scope project`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("=== Aliases ===")
		if len(cfg.Aliases) == 0 {
			fmt.Println("No aliases found")
		} else {
			for name, alias := range cfg.Aliases {
				fmt.Printf("- %s: %s\n", name, alias.Cmd)
			}
		}

		fmt.Println("\n=== Workflows ===")
		if len(cfg.Workflows) == 0 {
			fmt.Println("No workflows found")
		} else {
			for name, workflow := range cfg.Workflows {
				fmt.Printf("- %s: %s\n", name, workflow.Description)
				if len(workflow.Tags) > 0 {
					fmt.Printf("  Tags: %v\n", workflow.Tags)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
