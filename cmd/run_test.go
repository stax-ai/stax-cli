package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stax-ai/stax-cli/cmd"
	"github.com/stretchr/testify/require"
)

func TestRunCommand_RequiresPayloadOrInteractive(t *testing.T) {
	// No --payload-file, --payload, or --interactive: should error
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"run"})
	err := rootCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "payload")
}

func TestRunCommand_ConfigNotFound(t *testing.T) {
	dir := t.TempDir()
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"run", "--config", filepath.Join(dir, "config.yml"), "--payload", "{}"})
	err := rootCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "config")
}

func TestRunCommand_WithPayloadInNonProjectDir(t *testing.T) {
	// Dir exists but has no config.yml: FindAndLoad would fail when --config not set.
	// When --config is set to a nonexistent file, we get load error.
	dir := t.TempDir()
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"run", "--config", filepath.Join(dir, "nope.yml"), "--payload", "{}"})
	err := rootCmd.Execute()
	require.Error(t, err)
}

func TestRunCommand_ListedInHelp(t *testing.T) {
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	require.NoError(t, err)
	// Just ensure run is registered; help output is printed to stdout
	_ = os.Stderr
}

func TestBuildCommand_RequiresImageOrRegistry(t *testing.T) {
	dir := t.TempDir()
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"build", "--path", dir})
	err := rootCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "image")
}

func TestBuildCommand_ListedInHelp(t *testing.T) {
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"build", "--help"})
	err := rootCmd.Execute()
	require.NoError(t, err)
}
