package workflow

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Executor defines the interface for executing commands
type Executor interface {
	Execute(command string) error
}

// RealExecutor executes commands using os/exec
type RealExecutor struct{}

// Execute runs a command using os/exec and returns the error
func (e *RealExecutor) Execute(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Step represents a single step in a workflow
type Step struct {
	Run     string
	OnFail  string // stop, continue, prompt
	Confirm bool
}

// Workflow represents a sequence of steps
type Workflow struct {
	Description string
	Params      []string
	Steps       []Step
	Tags        []string
}

// RunOptions configures workflow execution behavior
type RunOptions struct {
	DryRun bool
	Writer io.Writer
	Reader io.Reader // for confirm prompts; defaults to os.Stdin
}

// Run executes a workflow with the given arguments
func Run(wf Workflow, args []string, ex Executor, opts RunOptions) error {
	w := opts.Writer
	if w == nil {
		w = os.Stdout
	}
	r := opts.Reader
	if r == nil {
		r = os.Stdin
	}

	// Parameter substitution
	paramMap := make(map[string]string)
	for i, param := range wf.Params {
		if i < len(args) {
			paramMap["{"+param+"}"] = args[i]
		} else {
			paramMap["{"+param+"}"] = ""
		}
	}

	for i, step := range wf.Steps {
		stepCmd := step.Run
		for placeholder, value := range paramMap {
			stepCmd = strings.ReplaceAll(stepCmd, placeholder, value)
		}

		// Dry-run: print steps without executing
		if opts.DryRun {
			prefix := fmt.Sprintf("  %d. ", i+1)
			if IsDestructive(stepCmd) {
				fmt.Fprintf(w, "%s[DESTRUCTIVE] %s\n", prefix, stepCmd)
			} else {
				fmt.Fprintf(w, "%s%s\n", prefix, stepCmd)
			}
			if step.Confirm {
				fmt.Fprintf(w, "     (requires confirmation)\n")
			}
			if step.OnFail != "" && step.OnFail != "stop" {
				fmt.Fprintf(w, "     on_fail: %s\n", step.OnFail)
			}
			continue
		}

		// Confirm prompt
		if step.Confirm {
			fmt.Fprintf(w, "Run '%s'? [y/N] ", stepCmd)
			scanner := bufio.NewScanner(r)
			if scanner.Scan() {
				input := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if input != "y" && input != "yes" {
					fmt.Fprintf(w, "Skipped\n")
					continue
				}
			} else {
				fmt.Fprintf(w, "Skipped (no input)\n")
				continue
			}
		}

		fmt.Fprintf(w, "Step %d/%d: %s\n", i+1, len(wf.Steps), stepCmd)

		err := ex.Execute(stepCmd)
		if err != nil {
			switch step.OnFail {
			case "stop":
				return fmt.Errorf("step %d failed: %w", i+1, err)
			case "continue":
				fmt.Fprintf(w, "[Step %d failed, continuing: %v]\n", i+1, err)
			case "prompt":
				fmt.Fprintf(w, "[Step %d failed: %v]\n", i+1, err)
				fmt.Fprintf(w, "Continue? [y/N] ")
				scanner := bufio.NewScanner(r)
				if scanner.Scan() {
					input := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if input == "y" || input == "yes" {
						continue
					}
				}
				return fmt.Errorf("step %d failed (aborted by user): %w", i+1, err)
			default:
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}
		}
	}

	return nil
}

// IsDestructive checks if a command contains potentially dangerous operations.
func IsDestructive(cmd string) bool {
	patterns := []string{
		"rm -rf", "rm -r", "rm -f",
		"git push -f", "git push --force",
		"git reset --hard",
		"drop table", "drop database",
		"delete from", "truncate",
	}
	lower := strings.ToLower(cmd)
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
