package cmd

import (
	"fmt"

	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/justn-hyeok/ganbatte/internal/history"
	"github.com/justn-hyeok/ganbatte/internal/shell"
	"github.com/justn-hyeok/ganbatte/internal/track"
	"github.com/spf13/cobra"
)

const trackMinEntries = 50 // prefer track.log over shell history once we have enough data

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Suggest aliases and workflows from shell history",
	Long: `Analyze shell history to suggest frequently used commands as aliases
and repeated command sequences as workflows.

ganbatte learns passively when eval "$(gnb shell-init)" is active. Once
enough commands are collected, suggest uses that data instead of raw shell
history for better accuracy.

Example:
  gnb suggest
  gnb suggest --apply
  gnb suggest --from-history   # force shell history source`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")
		minFreq, _ := cmd.Flags().GetInt("min-frequency")
		minSeq, _ := cmd.Flags().GetInt("min-sequence")
		forceHistory, _ := cmd.Flags().GetBool("from-history")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		existingCmds := make(map[string]string)
		for name, alias := range cfg.Aliases {
			existingCmds[name] = alias.Cmd
		}

		entries, source, err := loadEntries(forceHistory)
		if err != nil {
			return err
		}
		cmd.Printf("Analyzing %s (%d entries)...\n\n", source, len(entries))

		opts := history.SuggestOptions{
			MinFrequency:   minFreq,
			MinSequence:    minSeq,
			MaxSuggestions: 20,
		}
		suggestions := history.Suggest(entries, existingCmds, opts)

		if len(suggestions) == 0 {
			cmd.Println("No suggestions found. Try lowering --min-frequency or --min-sequence")
			return nil
		}

		var aliasSugs, paramSugs, wfSugs []history.Suggestion
		for _, s := range suggestions {
			switch s.Type {
			case "alias":
				aliasSugs = append(aliasSugs, s)
			case "param-alias":
				paramSugs = append(paramSugs, s)
			case "workflow":
				wfSugs = append(wfSugs, s)
			}
		}

		if len(aliasSugs) > 0 {
			cmd.Println("=== Alias Suggestions ===")
			for i, s := range aliasSugs {
				cmd.Printf("  %d. %-12s = %s\n", i+1, s.Name, s.Command)
				cmd.Printf("     %s\n", s.Reason)
			}
			cmd.Println()
		}

		if len(paramSugs) > 0 {
			cmd.Println("=== Parameterized Alias Suggestions ===")
			for i, s := range paramSugs {
				cmd.Printf("  %d. %s(%s) → %s\n", i+1, s.Name, s.Params[0], s.Command)
				cmd.Printf("     %s\n", s.Reason)
			}
			cmd.Println()
		}

		if len(wfSugs) > 0 {
			cmd.Println("=== Workflow Suggestions ===")
			for i, s := range wfSugs {
				cmd.Printf("  %d. %s\n", i+1, s.Name)
				for j, step := range s.Steps {
					cmd.Printf("     Step %d: %s\n", j+1, step)
				}
				cmd.Printf("     %s\n", s.Reason)
			}
			cmd.Println()
		}

		// Total impact summary
		totalSaved := 0
		for _, s := range suggestions {
			totalSaved += s.SavedKeystrokes
		}
		if totalSaved > 0 {
			cmd.Printf("Applying all suggestions would save ~%d keystrokes based on your history.\n\n", totalSaved)
		}

		if !apply {
			return nil
		}

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

		for _, s := range paramSugs {
			if cfg.Aliases == nil {
				cfg.Aliases = make(map[string]config.Alias)
			}
			if _, exists := cfg.Aliases[s.Name]; !exists {
				cfg.Aliases[s.Name] = config.Alias{
					Cmd:    s.Command,
					Params: s.Params,
				}
				cmd.Printf("Added param-alias '%s'(%s) = %s\n", s.Name, s.Params[0], s.Command)
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

		return nil
	},
}

// loadEntries returns history entries and a description of the source used.
// Prefers track.log when it has enough entries, falls back to shell history.
func loadEntries(forceHistory bool) ([]history.Entry, string, error) {
	if !forceHistory {
		logPath, err := track.LogPath()
		if err == nil {
			n, _ := track.Count(logPath)
			if n >= trackMinEntries {
				entries, err := track.Parse(logPath)
				if err == nil && len(entries) > 0 {
					return entries, fmt.Sprintf("ganbatte track log (%s)", logPath), nil
				}
			}
		}
	}

	// Fall back to shell history
	sh := shell.Detect()
	histPath := shell.HistoryPath(sh)
	if histPath == "" {
		return nil, "", fmt.Errorf("unsupported shell '%s' or history path not found", sh)
	}

	var parser history.Parser
	switch sh {
	case "zsh":
		parser = &history.ZshParser{}
	case "bash":
		parser = &history.BashParser{}
	case "fish":
		parser = &history.FishParser{}
	default:
		return nil, "", fmt.Errorf("no parser available for shell '%s'", sh)
	}

	entries, err := parser.Parse(histPath)
	if err != nil {
		return nil, "", fmt.Errorf("parsing history: %w", err)
	}
	return entries, fmt.Sprintf("%s shell history (%s)", sh, histPath), nil
}

func init() {
	suggestCmd.Flags().Bool("apply", false, "Apply suggested aliases and workflows to config")
	suggestCmd.Flags().Bool("from-history", false, "Force shell history source instead of track log")
	suggestCmd.Flags().Int("min-frequency", 5, "Minimum command frequency for alias suggestion")
	suggestCmd.Flags().Int("min-sequence", 3, "Minimum sequence frequency for workflow suggestion")
	RootCmd.AddCommand(suggestCmd)
}
