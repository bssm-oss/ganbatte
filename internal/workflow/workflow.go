package workflow

import (
	"fmt"
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
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = nil  // Don't take stdin from user
	cmd.Stdout = nil // Will be handled by caller
	cmd.Stderr = nil // Will be handled by caller

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

// Run executes a workflow with the given arguments
func Run(wf Workflow, args []string, ex Executor) error {
	// Simple parameter substitution
	paramMap := make(map[string]string)
	for i, param := range wf.Params {
		if i < len(args) {
			paramMap["{"+param+"}"] = args[i]
		} else {
			paramMap["{"+param+"}"] = "" // Empty if not provided
		}
	}

	// Execute each step
	for i, step := range wf.Steps {
		stepCmd := step.Run
		// Apply parameter substitution
		for placeholder, value := range paramMap {
			stepCmd = strings.ReplaceAll(stepCmd, placeholder, value)
		}

		if step.Confirm {
			// In a real implementation, we would prompt the user here
			// For now, we'll just continue (in v0.2 this will be proper)
			fmt.Printf("[Would prompt for confirmation: %s]\n", stepCmd)
		}

		err := ex.Execute(stepCmd)
		if err != nil {
			switch step.OnFail {
			case "stop":
				return fmt.Errorf("step %d failed: %w", i+1, err)
			case "continue":
				fmt.Printf("[Step %d failed, continuing: %v]\n", i+1, err)
			case "prompt":
				// In a real implementation, we would prompt the user here
				fmt.Printf("[Would prompt on failure: %v]\n", err)
				return fmt.Errorf("step %d failed: %w", i+1, err)
			default:
				// Default behavior is stop
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}
		}
	}

	return nil
}
