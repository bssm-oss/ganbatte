package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/justn-hyeok/ganbatte/internal/config"
	"github.com/justn-hyeok/ganbatte/internal/workflow"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <name> [args...]",
	Short: "Run an alias or workflow",
	Long: `Run an alias or workflow by name.
Supports parameter substitution for workflows.
Example:
  gnb run gs
  gnb run deploy main
  gnb run deploy main --dry-run`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		runArgs := args[1:]
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		scoped, err := config.LoadScoped()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		cfg := scoped.Merged

		yes, _ := cmd.Flags().GetBool("yes")

		// Check if it's an alias
		if alias, exists := cfg.Aliases[name]; exists {
			resolved := resolveAliasCmd(alias, runArgs)

			if dryRun {
				cmd.Printf("[dry-run] %s\n", resolved)
				if workflow.IsDestructive(resolved) {
					cmd.Printf("[DESTRUCTIVE] command detected\n")
				}
				if alias.Confirm {
					cmd.Printf("[requires confirmation]\n")
				}
				return nil
			}

			if alias.Confirm && !yes {
				if !confirmRun(cmd, fmt.Sprintf("⚠ Run %q? [y/N] ", resolved)) {
					return nil
				}
			}
			if projectOverrides(scoped, name, "alias") && !yes {
				if !confirmRun(cmd, fmt.Sprintf("Project alias '%s' overrides a global alias. Continue? [y/N] ", name)) {
					return nil
				}
			}

			cmd.Printf("Running: %s\n", resolved)
			ex := &workflow.RealExecutor{}
			return ex.Execute(resolved)
		}

		// Check if it's a workflow
		if wfDef, exists := cfg.Workflows[name]; exists {
			if projectOverrides(scoped, name, "workflow") && !yes && !dryRun {
				if !confirmRun(cmd, fmt.Sprintf("Project workflow '%s' overrides a global workflow. Continue? [y/N] ", name)) {
					return nil
				}
			}
			wf := workflow.Workflow{
				Description: wfDef.Description,
				Params:      wfDef.Params,
				Tags:        wfDef.Tags,
			}
			for _, s := range wfDef.Steps {
				wf.Steps = append(wf.Steps, workflow.Step{
					Run:     s.Run,
					OnFail:  s.OnFail,
					Confirm: s.Confirm,
				})
			}

			if dryRun {
				cmd.Printf("Workflow: %s (dry-run)\n", wfDef.Description)
			} else {
				cmd.Printf("Running workflow: %s\n", wfDef.Description)
			}

			return workflow.Run(wf, runArgs, &workflow.RealExecutor{}, workflow.RunOptions{
				DryRun:      dryRun,
				SkipConfirm: yes,
				Writer:      cmd.OutOrStdout(),
			})
		}

		return fmt.Errorf("alias or workflow '%s' not found", name)
	},
}

func projectOverrides(scoped *config.ScopedConfig, name, itemType string) bool {
	if scoped.Global == nil || scoped.Project == nil {
		return false
	}
	switch itemType {
	case "alias":
		_, inGlobal := scoped.Global.Aliases[name]
		_, inProject := scoped.Project.Aliases[name]
		return inGlobal && inProject
	case "workflow":
		_, inGlobal := scoped.Global.Workflows[name]
		_, inProject := scoped.Project.Workflows[name]
		return inGlobal && inProject
	default:
		return false
	}
}

func confirmRun(cmd *cobra.Command, prompt string) bool {
	return confirmPrompt(cmd.OutOrStdout(), prompt)
}

func confirmPrompt(w io.Writer, prompt string) bool {
	fmt.Fprint(w, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if input == "y" || input == "yes" {
			return true
		}
		fmt.Fprintln(w, "Cancelled")
		return false
	}
	fmt.Fprintln(w, "Cancelled (no input)")
	return false
}

// resolveAliasCmd substitutes parameters in alias command string.
func resolveAliasCmd(alias config.Alias, args []string) string {
	resolved := alias.Cmd

	if len(alias.Params) == 0 {
		return resolved
	}

	for i, param := range alias.Params {
		placeholder := "{" + param + "}"
		var value string
		if i < len(args) {
			value = args[i]
		} else if alias.DefaultParams != nil {
			value = alias.DefaultParams[param]
		}
		resolved = strings.ReplaceAll(resolved, placeholder, value)
	}

	return resolved
}

func init() {
	runCmd.Flags().Bool("dry-run", false, "Preview steps without executing")
	runCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	RootCmd.AddCommand(runCmd)
}
