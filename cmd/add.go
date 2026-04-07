package cmd

import (
	"fmt"
	"os"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Add a new alias",
	Long: `Add a new alias to the configuration.
Example:
  gnb add gs "git status -sb"
  gnb add ll "ls -la"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		command := args[1]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if cfg.Aliases == nil {
			cfg.Aliases = make(map[string]config.Alias)
		}

		// Check if alias already exists
		if _, exists := cfg.Aliases[name]; exists {
			fmt.Printf("Alias '%s' already exists. Use 'gnb edit %s <command>' to modify it.\n", name, name)
			os.Exit(1)
		}

		cfg.Aliases[name] = config.Alias{
			Cmd: command,
		}

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Added alias '%s' = %s\n", name, command)
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}
