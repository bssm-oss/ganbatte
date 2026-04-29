package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bssm-oss/ganbatte/internal/config"
	"github.com/bssm-oss/ganbatte/internal/shell"
	"github.com/bssm-oss/ganbatte/internal/track"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration and environment",
	Long:  `Check configuration validity, shell integration status, and report issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fix, _ := cmd.Flags().GetBool("fix")
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
		var foundConfigs []string
		for _, ext := range []string{"toml", "yaml", "yml", "json"} {
			p := filepath.Join(configDir, "config."+ext)
			if _, err := os.Stat(p); err == nil {
				foundConfigs = append(foundConfigs, p)
			}
		}
		switch len(foundConfigs) {
		case 0:
			cmd.Printf("[WARN] No global config found in %s\n", configDir)
			cmd.Println("       Run 'gnb init' to create one")
			issues++
		case 1:
			cmd.Printf("[OK] Global config: %s\n", foundConfigs[0])
		default:
			cmd.Printf("[WARN] Multiple config files found — only %s is used\n", foundConfigs[0])
			for _, p := range foundConfigs[1:] {
				cmd.Printf("       ignored: %s\n", p)
			}
			if fix {
				for _, p := range foundConfigs[1:] {
					os.Remove(p)
					cmd.Printf("[FIXED] Removed %s\n", p)
				}
			} else {
				cmd.Println("       Run 'gnb doctor --fix' to remove ignored files")
				issues++
			}
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

			// Check for alias names shadowing system commands
			for name := range cfg.Aliases {
				if path, err := exec.LookPath(name); err == nil {
					cmd.Printf("[WARN] Alias '%s' shadows system command: %s\n", name, path)
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

		// 6. Passive tracking (track.log)
		cmd.Println()
		cmd.Println("=== Passive Tracking ===")
		logPath, err := track.LogPath()
		if err != nil {
			cmd.Printf("[WARN] Could not determine track.log path: %v\n", err)
		} else {
			n, _ := track.Count(logPath)
			if n == 0 {
				cmd.Printf("[INFO] track.log: empty (%s)\n", logPath)
				cmd.Println("       Add 'eval \"$(gnb shell-init)\"' to your shell config to start tracking")
			} else if n < trackMinEntries {
				cmd.Printf("[INFO] track.log: %d entries (need %d more before 'gnb suggest' uses it)\n",
					n, trackMinEntries-n)
				cmd.Printf("       %s\n", logPath)
			} else {
				cmd.Printf("[OK] track.log: %d entries — 'gnb suggest' will use this\n", n)
				cmd.Printf("     %s\n", logPath)
			}
		}

		// 7. p10k instant prompt ordering (zsh only)
		if sh == "zsh" {
			cmd.Println()
			cmd.Println("=== Shell Integration ===")
			if checkP10kOrdering(cmd, home, fix) {
				issues++
			} else {
				p10kPresent := func() bool {
					data, err := os.ReadFile(filepath.Join(home, ".zshrc"))
					if err != nil {
						return false
					}
					lines := strings.Split(string(data), "\n")
					p10k, _ := findShellInitLines(lines)
					return p10k != -1
				}()
				if p10kPresent {
					cmd.Println("[OK] shell-init: before p10k instant prompt preamble")
				} else {
					cmd.Println("[OK] shell-init: no p10k detected")
				}
			}
		}

		// Summary
		cmd.Println()
		if issues == 0 {
			cmd.Println("No issues found. ganbatte!")
		} else {
			cmd.Printf("%d issue(s) found\n", issues)
			if !fix {
				cmd.Println("Run 'gnb doctor --fix' to automatically fix repairable issues")
			}
		}
		return nil
	},
}

// findShellInitLines returns 0-indexed line numbers of the p10k instant prompt preamble
// and the gnb shell-init eval line. Returns -1 for each if not found.
// Commented-out lines (trimmed prefix "#") are ignored.
func findShellInitLines(lines []string) (p10kLine, gnbLine int) {
	p10kLine, gnbLine = -1, -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if p10kLine == -1 && strings.Contains(line, "p10k-instant-prompt-") {
			p10kLine = i
		}
		if gnbLine == -1 && strings.Contains(line, "gnb shell-init") {
			gnbLine = i
		}
	}
	return
}

// applyP10kFix returns a new lines slice with the gnb shell-init line moved to just
// before the p10k instant prompt block (including its preceding comment lines).
// A blank line immediately before the gnb line is removed to avoid leftover whitespace.
func applyP10kFix(lines []string, gnbLine, p10kLine int) []string {
	gnbContent := lines[gnbLine]

	// Remove gnb line; also remove a preceding blank line to avoid leftover whitespace
	start := gnbLine
	if start > 0 && strings.TrimSpace(lines[start-1]) == "" {
		start--
	}
	fixed := make([]string, 0, len(lines)-1)
	fixed = append(fixed, lines[:start]...)
	fixed = append(fixed, lines[gnbLine+1:]...)

	// Walk back from p10kLine through consecutive comment lines to find block start.
	// p10kLine index is still valid because gnbLine > p10kLine (we only removed lines after it).
	insertAt := p10kLine
	for insertAt > 0 && strings.HasPrefix(strings.TrimSpace(fixed[insertAt-1]), "#") {
		insertAt--
	}

	result := make([]string, 0, len(fixed)+2)
	result = append(result, fixed[:insertAt]...)
	result = append(result, gnbContent, "")
	result = append(result, fixed[insertAt:]...)
	return result
}

// checkP10kOrdering detects and optionally fixes p10k instant prompt ordering in ~/.zshrc.
// Only prints output when p10k is detected. Returns true if an unfixed issue remains.
func checkP10kOrdering(cmd *cobra.Command, home string, fix bool) bool {
	zshrc := filepath.Join(home, ".zshrc")
	data, err := os.ReadFile(zshrc)
	if err != nil {
		return false
	}

	lines := strings.Split(string(data), "\n")
	p10kLine, gnbLine := findShellInitLines(lines)

	if p10kLine == -1 || gnbLine == -1 {
		return false // p10k or shell-init not configured — nothing to check
	}

	if gnbLine < p10kLine {
		return false
	}

	cmd.Printf("[WARN] shell-init is after p10k instant prompt preamble (line %d vs %d)\n", gnbLine+1, p10kLine+1)
	cmd.Println("       This causes a warning on every zsh start")

	if !fix {
		cmd.Println("       Run 'gnb doctor --fix' to automatically reorder")
		return true
	}

	result := applyP10kFix(lines, gnbLine, p10kLine)
	if err := os.WriteFile(zshrc, []byte(strings.Join(result, "\n")), 0o644); err != nil {
		cmd.Printf("[ERROR] Could not write %s: %v\n", zshrc, err)
		return true
	}
	cmd.Printf("[FIXED] Moved eval line before p10k preamble in %s\n", zshrc)
	return false
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Automatically fix detected issues")
	RootCmd.AddCommand(doctorCmd)
}
