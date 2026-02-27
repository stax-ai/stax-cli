package create_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stax-ai/stax-cli/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand_CreatesProjectInPath(t *testing.T) {
	dir := t.TempDir()
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"create", "my-project", "--path", dir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	projectRoot := filepath.Join(dir, "my-project")
	_, err = os.Stat(projectRoot)
	require.NoError(t, err)

	for _, name := range []string{"config.yml", "pyproject.toml", "main.py", "README.md"} {
		_, err = os.Stat(filepath.Join(projectRoot, name))
		assert.NoError(t, err, "expected %s to exist", name)
	}

	data, _ := os.ReadFile(filepath.Join(projectRoot, "config.yml"))
	assert.Contains(t, string(data), "automation_id:")
	assert.Contains(t, string(data), "main:handler")
}

func TestCreateCommand_RequiresProjectName(t *testing.T) {
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"create"})
	err := rootCmd.Execute()
	require.Error(t, err)
}
