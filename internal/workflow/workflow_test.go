package workflow

import (
	"bytes"
	"errors"
	"strings"
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
	wf := Workflow{
		Description: "Test workflow",
		Params:      []string{"env"},
		Steps: []Step{
			{Run: "echo hello", OnFail: "stop"},
			{Run: "echo {env}", OnFail: "continue", Confirm: true},
		},
		Tags: []string{"test"},
	}

	executor := &MockExecutor{}
	args := []string{"production"}

	err := Run(wf, args, executor, RunOptions{
		Reader: strings.NewReader("y\n"),
	})
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 2)
	assert.Equal(t, "echo hello", executor.ExecuteCalls[0])
	assert.Equal(t, "echo production", executor.ExecuteCalls[1])
}

func TestWorkflowRun_WithStopOnFailure(t *testing.T) {
	wf := Workflow{
		Steps: []Step{
			{Run: "fail command", OnFail: "stop"},
			{Run: "should not run", OnFail: "stop"},
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

	err := Run(wf, []string{}, executor, RunOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step 1 failed")
	assert.Len(t, executor.ExecuteCalls, 1)
}

func TestWorkflowRun_WithContinueOnFailure(t *testing.T) {
	wf := Workflow{
		Steps: []Step{
			{Run: "fail command", OnFail: "continue"},
			{Run: "second command", OnFail: "stop"},
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

	err := Run(wf, []string{}, executor, RunOptions{})
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 2)
}

func TestWorkflowRun_WithParameterSubstitution(t *testing.T) {
	wf := Workflow{
		Params: []string{"name", "value"},
		Steps: []Step{
			{Run: "echo Hello {name}"},
			{Run: "echo Value is {value}"},
			{Run: "echo No param here"},
		},
	}

	executor := &MockExecutor{}
	args := []string{"World", "42"}

	err := Run(wf, args, executor, RunOptions{})
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 3)
	assert.Equal(t, "echo Hello World", executor.ExecuteCalls[0])
	assert.Equal(t, "echo Value is 42", executor.ExecuteCalls[1])
	assert.Equal(t, "echo No param here", executor.ExecuteCalls[2])
}

func TestWorkflowRun_WithMissingParameters(t *testing.T) {
	wf := Workflow{
		Params: []string{"required", "optional"},
		Steps: []Step{
			{Run: "echo {required}-{optional}"},
		},
	}

	executor := &MockExecutor{}
	args := []string{"given"}

	err := Run(wf, args, executor, RunOptions{})
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 1)
	assert.Equal(t, "echo given-", executor.ExecuteCalls[0])
}

func TestWorkflowRun_EmptyWorkflow(t *testing.T) {
	wf := Workflow{
		Description: "Empty workflow",
		Steps:       []Step{},
	}

	executor := &MockExecutor{}
	err := Run(wf, []string{}, executor, RunOptions{})
	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 0)
}

func TestWorkflowRun_DryRun(t *testing.T) {
	wf := Workflow{
		Description: "Deploy workflow",
		Params:      []string{"branch"},
		Steps: []Step{
			{Run: "pnpm lint"},
			{Run: "pnpm test", OnFail: "continue"},
			{Run: "git push -f origin {branch}", Confirm: true},
		},
	}

	buf := new(bytes.Buffer)
	executor := &MockExecutor{}

	err := Run(wf, []string{"main"}, executor, RunOptions{
		DryRun: true,
		Writer: buf,
	})

	assert.NoError(t, err)
	assert.Empty(t, executor.ExecuteCalls) // nothing should execute

	out := buf.String()
	assert.Contains(t, out, "1. pnpm lint")
	assert.Contains(t, out, "2. pnpm test")
	assert.Contains(t, out, "on_fail: continue")
	assert.Contains(t, out, "[DESTRUCTIVE]")
	assert.Contains(t, out, "git push -f origin main")
	assert.Contains(t, out, "(requires confirmation)")
}

func TestWorkflowRun_ConfirmYes(t *testing.T) {
	wf := Workflow{
		Steps: []Step{
			{Run: "echo safe"},
			{Run: "echo dangerous", Confirm: true},
		},
	}

	outBuf := new(bytes.Buffer)
	inBuf := strings.NewReader("y\n")
	executor := &MockExecutor{}

	err := Run(wf, []string{}, executor, RunOptions{
		Writer: outBuf,
		Reader: inBuf,
	})

	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 2)
	assert.Equal(t, "echo dangerous", executor.ExecuteCalls[1])
}

func TestWorkflowRun_ConfirmNo(t *testing.T) {
	wf := Workflow{
		Steps: []Step{
			{Run: "echo safe"},
			{Run: "echo dangerous", Confirm: true},
			{Run: "echo after"},
		},
	}

	outBuf := new(bytes.Buffer)
	inBuf := strings.NewReader("n\n")
	executor := &MockExecutor{}

	err := Run(wf, []string{}, executor, RunOptions{
		Writer: outBuf,
		Reader: inBuf,
	})

	assert.NoError(t, err)
	assert.Len(t, executor.ExecuteCalls, 2) // safe + after (dangerous skipped)
	assert.Equal(t, "echo safe", executor.ExecuteCalls[0])
	assert.Equal(t, "echo after", executor.ExecuteCalls[1])
	assert.Contains(t, outBuf.String(), "Skipped")
}

func TestWorkflowRun_OnFailPrompt_Continue(t *testing.T) {
	mock := &MockExecutor{
		ExecuteFunc: func(cmd string) error {
			if cmd == "failing-step" {
				return errors.New("step failed")
			}
			return nil
		},
	}

	wf := Workflow{
		Steps: []Step{
			{Run: "failing-step", OnFail: "prompt"},
			{Run: "next-step"},
		},
	}

	var out bytes.Buffer
	err := Run(wf, nil, mock, RunOptions{
		Writer: &out,
		Reader: strings.NewReader("y\n"),
	})

	assert.NoError(t, err)
	assert.Contains(t, out.String(), "Continue? [y/N]")
	assert.Equal(t, 2, len(mock.ExecuteCalls))
}

func TestWorkflowRun_OnFailPrompt_Abort(t *testing.T) {
	mock := &MockExecutor{
		ExecuteFunc: func(cmd string) error {
			if cmd == "failing-step" {
				return errors.New("step failed")
			}
			return nil
		},
	}

	wf := Workflow{
		Steps: []Step{
			{Run: "failing-step", OnFail: "prompt"},
			{Run: "next-step"},
		},
	}

	var out bytes.Buffer
	err := Run(wf, nil, mock, RunOptions{
		Writer: &out,
		Reader: strings.NewReader("n\n"),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "aborted by user")
	assert.Equal(t, 1, len(mock.ExecuteCalls))
}

func TestIsDestructive(t *testing.T) {
	assert.True(t, IsDestructive("rm -rf /tmp/foo"))
	assert.True(t, IsDestructive("git push -f origin main"))
	assert.True(t, IsDestructive("git push --force origin main"))
	assert.True(t, IsDestructive("git reset --hard HEAD~1"))
	assert.True(t, IsDestructive("DROP TABLE users"))
	assert.True(t, IsDestructive("DELETE FROM users"))
	assert.False(t, IsDestructive("echo hello"))
	assert.False(t, IsDestructive("git push origin main"))
	assert.False(t, IsDestructive("pnpm test"))
}
