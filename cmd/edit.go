package cmd

import (
	"fmt"
	"os"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit <name> <command>",
	Short: "Edit an existing alias",
	Long: `Edit an existing alias in the configuration.
Example:
  gnb edit gs "git status --short --branch"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		command := args[1]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Check if alias exists
		if _, exists := cfg.Aliases[name]; !exists {
			fmt.Printf("Alias '%s' not found. Use 'gnb add %s <command>' to create it.\n", name, name)
			os.Exit(1)
		}

		cfg.Aliases[name] = config.Alias{
			Cmd: command,
		}

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Updated alias '%s' = %s\n", name, command)
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}
