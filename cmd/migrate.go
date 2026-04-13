package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/justn-hyeok/ganbatte/internal/shell"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Import aliases from shell config files",
	Long: `Scan your shell config files (.zshrc, .bashrc, config.fish) for existing
alias definitions and import them into ganbatte.

This is the fastest way to get started — bring all your existing aliases
into ganbatte with one command.

Example:
  gnb migrate
  gnb migrate --shell zsh
  gnb migrate --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellFlag, _ := cmd.Flags().GetString("shell")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		sh := shellFlag
		if sh == "" {
			sh = shell.Detect()
		}

		// Find config files
		configs := shell.FindShellConfigs(sh)
		if len(configs) == 0 {
			return fmt.Errorf("no %s config files found", sh)
		}

		// Parse aliases from all config files
		var allAliases []shell.ShellAlias
		for _, path := range configs {
			aliases, err := shell.ParseShellAliases(path, sh)
			if err != nil {
				cmd.Printf("Warning: skipping %s: %v\n", path, err)
				continue
			}
			if len(aliases) > 0 {
				cmd.Printf("Found %d aliases in %s\n", len(aliases), path)
				allAliases = append(allAliases, aliases...)
			}
		}

		if len(allAliases) == 0 {
			cmd.Println("No aliases found in shell config files")
			return nil
		}

		// Load existing config to check duplicates
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Filter out duplicates
		var newAliases []shell.ShellAlias
		for _, a := range allAliases {
			if existing, ok := cfg.Aliases[a.Name]; ok {
				if existing.Cmd == a.Command {
					continue // exact duplicate
				}
				cmd.Printf("  skip: %s (already exists with different command)\n", a.Name)
				continue
			}
			newAliases = append(newAliases, a)
		}

		if len(newAliases) == 0 {
			cmd.Println("All aliases already exist in config")
			return nil
		}

		// Display what will be imported
		cmd.Printf("\n%d new alias(es) to import:\n", len(newAliases))
		for _, a := range newAliases {
			cmd.Printf("  %s = %q\n", a.Name, a.Command)
		}

		if dryRun {
			cmd.Println("\n(dry-run, no changes made)")
			return nil
		}

		// Confirm
		cmd.Printf("\nImport all? [Y/n] ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if input == "n" || input == "no" {
				cmd.Println("Cancelled")
				return nil
			}
		}

		// Apply
		if cfg.Aliases == nil {
			cfg.Aliases = make(map[string]config.Alias)
		}
		for _, a := range newAliases {
			cfg.Aliases[a.Name] = config.Alias{Cmd: a.Command}
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		cmd.Printf("✓ %d alias(es) imported\n", len(newAliases))
		cmd.Println("Run 'eval \"$(gnb shell-init)\"' to activate them")
		return nil
	},
}

func init() {
	migrateCmd.Flags().String("shell", "", "Shell type to scan (bash, zsh, fish). Auto-detected if omitted")
	migrateCmd.Flags().Bool("dry-run", false, "Preview without importing")
	RootCmd.AddCommand(migrateCmd)
}
