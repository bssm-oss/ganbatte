package cmd

import (
	"fmt"

	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/justn-hyeok/ganbatte/internal/history"
	"github.com/justn-hyeok/ganbatte/internal/shell"
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

		// Split by type for sectioned display
		var aliasSugs, wfSugs []history.Suggestion
		for _, s := range suggestions {
			switch s.Type {
			case "alias":
				aliasSugs = append(aliasSugs, s)
			case "workflow":
				wfSugs = append(wfSugs, s)
			}
		}

		if len(aliasSugs) > 0 {
			cmd.Println("=== Alias Suggestions (frequency) ===")
			for i, s := range aliasSugs {
				cmd.Printf("  %d. %s = %s\n", i+1, s.Name, s.Command)
				cmd.Printf("     %s\n", s.Reason)
			}
			cmd.Println()
		}

		if len(wfSugs) > 0 {
			cmd.Println("=== Workflow Suggestions (sequences) ===")
			for i, s := range wfSugs {
				cmd.Printf("  %d. %s\n", i+1, s.Name)
				for j, step := range s.Steps {
					cmd.Printf("     Step %d: %s\n", j+1, step)
				}
				cmd.Printf("     %s\n", s.Reason)
			}
			cmd.Println()
		}

		// Apply mode
		if apply {
			applied := 0

			for _, s := range aliasSugs {
				if cfg.Aliases == nil {
					cfg.Aliases = make(map[string]config.Alias)
				}
				if _, exists := cfg.Aliases[s.Name]; !exists {
					cfg.Aliases[s.Name] = config.Alias{Cmd: s.Command}
					cmd.Printf("Added alias '%s' = %s\n", s.Name, s.Command)
					applied++
				}
			}

			for _, s := range wfSugs {
				if cfg.Workflows == nil {
					cfg.Workflows = make(map[string]config.Workflow)
				}
				if _, exists := cfg.Workflows[s.Name]; !exists {
					steps := make([]config.Step, len(s.Steps))
					for j, step := range s.Steps {
						steps[j] = config.Step{Run: step}
					}
					cfg.Workflows[s.Name] = config.Workflow{
						Description: s.Reason,
						Steps:       steps,
					}
					cmd.Printf("Added workflow '%s' (%d steps)\n", s.Name, len(s.Steps))
					applied++
				}
			}

			if applied > 0 {
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("saving config: %w", err)
				}
				cmd.Printf("\n%d item(s) applied to config\n", applied)
			} else {
				cmd.Println("No new items to apply")
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
