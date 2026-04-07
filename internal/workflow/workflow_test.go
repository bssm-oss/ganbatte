package workflow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockExecutor is a mock executor for testing
type MockExecutor struct {
	ExecuteFunc  func(command string) error
	ExecuteCalls []string
}

func (m *MockExecutor) Execute(command string) error {
	m.ExecuteCalls = append(m.ExecuteCalls, command)
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(command)
	}
	return nil
}

func TestWorkflowRun_Success(t *testing.T) {
	// Arrange
	wf := Workflow{
		Description: "Test workflow",
		Params:      []string{"env"},
		Steps: []Step{
			{Run: "echo hello", OnFail: "stop", Confirm: false},
			{Run: "echo {env}", OnFail: "continue", Confirm: true},
		},
		Tags: []string{"test"},
	}

	executor := &MockExecutor{}
	args := []string{"production"}

	// Act
	err := Run(wf, args, executor)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 2)
	assert.Equal(t, "echo hello", executor.ExecuteCalls[0])
	assert.Equal(t, "echo production", executor.ExecuteCalls[1])
}

func TestWorkflowRun_WithStopOnFailure(t *testing.T) {
	// Arrange
	wf := Workflow{
		Description: "Test workflow",
		Steps: []Step{
			{Run: "fail command", OnFail: "stop", Confirm: false},
			{Run: "should not run", OnFail: "stop", Confirm: false},
		},
	}

	executor := &MockExecutor{
		ExecuteFunc: func(command string) error {
			if command == "fail command" {
				return errors.New("command failed")
			}
			return nil
		},
	}

	// Act
	err := Run(wf, []string{}, executor)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step 1 failed")
	assert.Len(t, executor.ExecuteCalls, 1) // Only first command should execute
}

func TestWorkflowRun_WithContinueOnFailure(t *testing.T) {
	// Arrange
	wf := Workflow{
		Description: "Test workflow",
		Steps: []Step{
			{Run: "fail command", OnFail: "continue", Confirm: false},
			{Run: "second command", OnFail: "stop", Confirm: false},
		},
	}

	executor := &MockExecutor{
		ExecuteFunc: func(command string) error {
			if command == "fail command" {
				return errors.New("command failed")
			}
			return nil
		},
	}

	// Act
	err := Run(wf, []string{}, executor)

	// Assert
	assert.NoError(t, err)                  // Should not error because continue
	assert.Len(t, executor.ExecuteCalls, 2) // Both commands should execute
}

func TestWorkflowRun_WithParameterSubstitution(t *testing.T) {
	// Arrange
	wf := Workflow{
		Description: "Test workflow",
		Params:      []string{"name", "value"},
		Steps: []Step{
			{Run: "echo Hello {name}", OnFail: "stop", Confirm: false},
			{Run: "echo Value is {value}", OnFail: "stop", Confirm: false},
			{Run: "echo No param here", OnFail: "stop", Confirm: false},
		},
	}

	executor := &MockExecutor{}
	args := []string{"World", "42"}

	// Act
	err := Run(wf, args, executor)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 3)
	assert.Equal(t, "echo Hello World", executor.ExecuteCalls[0])
	assert.Equal(t, "echo Value is 42", executor.ExecuteCalls[1])
	assert.Equal(t, "echo No param here", executor.ExecuteCalls[2])
}

func TestWorkflowRun_WithMissingParameters(t *testing.T) {
	// Arrange
	wf := Workflow{
		Description: "Test workflow",
		Params:      []string{"required", "optional"},
		Steps: []Step{
			{Run: "echo {required}-{optional}", OnFail: "stop", Confirm: false},
		},
	}

	executor := &MockExecutor{}
	args := []string{"given"} // Only one arg provided

	// Act
	err := Run(wf, args, executor)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 1)
	assert.Equal(t, "echo given-", executor.ExecuteCalls[0]) // Optional should be empty
}

func TestWorkflowRun_EmptyWorkflow(t *testing.T) {
	// Arrange
	wf := Workflow{
		Description: "Empty workflow",
		Steps:       []Step{}, // No steps
	}

	executor := &MockExecutor{}

	// Act
	err := Run(wf, []string{}, executor)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 0) // No commands should execute
}
