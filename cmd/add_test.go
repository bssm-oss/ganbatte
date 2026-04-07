package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAliasCommand(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tmpDir)
	require.NoError(t, err)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()

	// Get the absolute path to the gnb binary
	gnbPath, err := filepath.Abs("./gnb")
	require.NoError(t, err)

	// Run gnb init first to create config
	initCmd := exec.Command(gnbPath, "init")
	initCmd.Dir = tmpDir
	var initOut bytes.Buffer
	initCmd.Stdout = &initOut
	initCmd.Stderr = &initOut
	err = initCmd.Run()
	require.NoError(t, err)

	// Now test add command
	addCmd := exec.Command(gnbPath, "add", "testalias", "echo hello")
	addCmd.Dir = tmpDir
	var addOut bytes.Buffer
	addCmd.Stdout = &addOut
	addCmd.Stderr = &addOut
	err = addCmd.Run()
	require.NoError(t, err)

	// Check output
	output := addOut.String()
	assert.Contains(t, output, "Added alias 'testalias'")
	assert.Contains(t, output, "echo hello")

	// Verify it was actually added by running list
	listCmd := exec.Command(gnbPath, "list")
	listCmd.Dir = tmpDir
	var listOut bytes.Buffer
	listCmd.Stdout = &listOut
	listCmd.Stderr = &listOut
	err = listCmd.Run()
	require.NoError(t, err)

	listOutput := listOut.String()
	assert.Contains(t, listOutput, "testalias")
	assert.Contains(t, listOutput, "echo hello")
}

func TestAddDuplicateAlias(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	err := os.Setenv("HOME", tmpDir)
	require.NoError(t, err)
	defer func() {
		if oldHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", oldHome)
		}
	}()

	// Get the absolute path to the gnb binary
	gnbPath, err := filepath.Abs("./gnb")
	require.NoError(t, err)

	// Run gnb init first to create config
	initCmd := exec.Command(gnbPath, "init")
	initCmd.Dir = tmpDir
	var initOut bytes.Buffer
	initCmd.Stdout = &initOut
	initCmd.Stderr = &initOut
	err = initCmd.Run()
	require.NoError(t, err)

	// Add an alias first time
	addCmd1 := exec.Command(gnbPath, "add", "duptest", "first command")
	addCmd1.Dir = tmpDir
	var addOut1 bytes.Buffer
	addCmd1.Stdout = &addOut1
	addCmd1.Stderr = &addOut1
	err = addCmd1.Run()
	require.NoError(t, err)

	// Try to add the same alias again - should fail
	addCmd2 := exec.Command(gnbPath, "add", "duptest", "second command")
	addCmd2.Dir = tmpDir
	var addOut2 bytes.Buffer
	addCmd2.Stdout = &addOut2
	addCmd2.Stderr = &addOut2
	err = addCmd2.Run()
	assert.Error(t, err) // Should fail

	// Check that error message indicates duplicate
	assert.Contains(t, addOut2.String(), "already exists")
	assert.Contains(t, addOut2.String(), "Use 'gnb edit'")
}
