package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/justn-hyeok/ganbatte/internal/shell"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration and environment",
	Long:  `Check configuration validity, shell integration status, and report issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		issues := 0

		// 1. Shell detection
		sh := shell.Detect()
		if sh == "unknown" {
			cmd.Println("[WARN] Shell not detected ($SHELL is empty)")
			issues++
		} else {
			cmd.Printf("[OK] Shell: %s\n", sh)
		}

		// 2. History file
		histPath := shell.HistoryPath(sh)
		if histPath == "" {
			cmd.Println("[WARN] History file path unknown for this shell")
			issues++
		} else if _, err := os.Stat(histPath); err != nil {
			cmd.Printf("[WARN] History file not found: %s\n", histPath)
			issues++
		} else {
			cmd.Printf("[OK] History file: %s\n", histPath)
		}

		// 3. Config file
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}

		configDir := filepath.Join(home, ".config", "ganbatte")
		configFound := false
		for _, ext := range []string{"toml", "yaml", "yml", "json"} {
			p := filepath.Join(configDir, "config."+ext)
			if _, err := os.Stat(p); err == nil {
				cmd.Printf("[OK] Global config: %s\n", p)
				configFound = true
				break
			}
		}
		if !configFound {
			cmd.Printf("[WARN] No global config found in %s\n", configDir)
			cmd.Println("       Run 'gnb init' to create one")
			issues++
		}

		// 4. Project config
		projectFound := false
		for _, name := range []string{".ganbatte.toml", ".ganbatte.yaml", ".ganbatte.yml", ".ganbatte.json"} {
			if _, err := os.Stat(name); err == nil {
				cmd.Printf("[OK] Project config: %s\n", name)
				projectFound = true
				break
			}
		}
		if !projectFound {
			cmd.Println("[INFO] No project config in current directory")
		}

		// 5. Config validity + duplicate check
		cfg, err := config.Load()
		if err != nil {
			cmd.Printf("[ERROR] Config load failed: %v\n", err)
			issues++
		} else {
			cmd.Printf("[OK] Config version: %s\n", cfg.Version)
			cmd.Printf("[OK] Aliases: %d, Workflows: %d\n", len(cfg.Aliases), len(cfg.Workflows))

			// Check for name collisions between aliases and workflows
			for name := range cfg.Aliases {
				if _, exists := cfg.Workflows[name]; exists {
					cmd.Printf("[WARN] Name collision: '%s' exists as both alias and workflow\n", name)
					issues++
				}
			}

			// Check for empty commands
			for name, alias := range cfg.Aliases {
				if alias.Cmd == "" {
					cmd.Printf("[WARN] Alias '%s' has empty command\n", name)
					issues++
				}
			}
			for name, wf := range cfg.Workflows {
				if len(wf.Steps) == 0 {
					cmd.Printf("[WARN] Workflow '%s' has no steps\n", name)
					issues++
				}
			}
		}

		// Summary
		cmd.Println()
		if issues == 0 {
			cmd.Println("No issues found. ganbatte!")
		} else {
			cmd.Printf("%d issue(s) found\n", issues)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}
