package cmd

import (
	"fmt"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/bssm-oss/ganbatte/internal/history"
	"github.com/bssm-oss/ganbatte/internal/shell"
	"github.com/spf13/cobra"
)

// suggestCmd represents the suggest command
var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Suggest aliases and workflows from shell history",
	Long: `Analyze shell history to suggest frequently used commands as aliases
and repeated command sequences as workflows.
Example:
  gnb suggest
  gnb suggest --apply`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")
		minFreq, _ := cmd.Flags().GetInt("min-frequency")
		minSeq, _ := cmd.Flags().GetInt("min-sequence")

		// Detect shell and history path
		sh := shell.Detect()
		histPath := shell.HistoryPath(sh)
		if histPath == "" {
			return fmt.Errorf("unsupported shell '%s' or history path not found", sh)
		}

		// Select parser
		var parser history.Parser
		switch sh {
		case "zsh":
			parser = &history.ZshParser{}
		case "bash":
			parser = &history.BashParser{}
		case "fish":
			parser = &history.FishParser{}
		default:
			return fmt.Errorf("no parser available for shell '%s'", sh)
		}

		// Parse history
		cmd.Printf("Analyzing %s history at %s...\n", sh, histPath)
		entries, err := parser.Parse(histPath)
		if err != nil {
			return fmt.Errorf("parsing history: %w", err)
		}
		cmd.Printf("Found %d history entries\n\n", len(entries))

		// Get existing aliases to avoid duplicates
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		existingCmds := make(map[string]string)
		for name, alias := range cfg.Aliases {
			existingCmds[name] = alias.Cmd
		}

		// Generate suggestions
		opts := history.SuggestOptions{
			MinFrequency:   minFreq,
			MinSequence:    minSeq,
			MaxSuggestions: 15,
		}
		suggestions := history.Suggest(entries, existingCmds, opts)

		if len(suggestions) == 0 {
			cmd.Println("No suggestions found. Try lowering --min-frequency or --min-sequence")
			return nil
		}

		// Display suggestions
		for i, s := range suggestions {
			switch s.Type {
			case "alias":
				cmd.Printf("%d. [alias] %s = %s\n", i+1, s.Name, s.Command)
				cmd.Printf("   %s\n\n", s.Reason)
			case "workflow":
				cmd.Printf("%d. [workflow] %s\n", i+1, s.Name)
				for j, step := range s.Steps {
					cmd.Printf("   Step %d: %s\n", j+1, step)
				}
				cmd.Printf("   %s\n\n", s.Reason)
			}
		}

		// Apply mode
		if apply {
			applied := 0
			for _, s := range suggestions {
				if s.Type == "alias" {
					if cfg.Aliases == nil {
						cfg.Aliases = make(map[string]config.Alias)
					}
					if _, exists := cfg.Aliases[s.Name]; !exists {
						cfg.Aliases[s.Name] = config.Alias{Cmd: s.Command}
						cmd.Printf("Added alias '%s' = %s\n", s.Name, s.Command)
						applied++
					}
				}
			}
			if applied > 0 {
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("saving config: %w", err)
				}
				cmd.Printf("\n%d alias(es) applied to config\n", applied)
			} else {
				cmd.Println("No new aliases to apply")
			}
		}

		return nil
	},
}

func init() {
	suggestCmd.Flags().Bool("apply", false, "Apply suggested aliases to config")
	suggestCmd.Flags().Int("min-frequency", 5, "Minimum command frequency for alias suggestion")
	suggestCmd.Flags().Int("min-sequence", 3, "Minimum sequence frequency for workflow suggestion")
	RootCmd.AddCommand(suggestCmd)
}
