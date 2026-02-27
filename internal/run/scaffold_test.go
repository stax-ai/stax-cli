package run

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteRunScaffold_CreatesFiles(t *testing.T) {
	runDir := t.TempDir()
	projectRoot := t.TempDir()

	err := WriteRunScaffold(runDir, projectRoot, "main:handler")
	require.NoError(t, err)

	mainPy := filepath.Join(runDir, "service", "main.py")
	data, err := os.ReadFile(mainPy)
	require.NoError(t, err)
	assert.Contains(t, string(data), "func_python.cloudevent")
	assert.Contains(t, string(data), "serve(handle)")
	assert.Contains(t, string(data), "STAX_HANDLER")

	pyproject := filepath.Join(runDir, "pyproject.toml")
	data, err = os.ReadFile(pyproject)
	require.NoError(t, err)
	assert.Contains(t, string(data), "func-python")
	assert.Contains(t, string(data), "cloudevents")

	linkPath := filepath.Join(runDir, "f")
	dest, err := os.Readlink(linkPath)
	require.NoError(t, err)
	absDest, err := filepath.EvalSymlinks(linkPath)
	require.NoError(t, err)
	absRoot, err := filepath.Abs(projectRoot)
	require.NoError(t, err)
	absRoot, err = filepath.EvalSymlinks(absRoot) // canonical form so /var and /private/var match on macOS
	require.NoError(t, err)
	assert.Equal(t, absRoot, absDest, "symlink f should point to project root (resolved: %s)", dest)
}

func TestWriteRunScaffold_RequiresDirAndRoot(t *testing.T) {
	err := WriteRunScaffold("", "/tmp", "main:handler")
	require.Error(t, err)

	err = WriteRunScaffold("/tmp", "", "main:handler")
	require.Error(t, err)
}

func TestRunDir(t *testing.T) {
	root := filepath.FromSlash("/home/proj")
	port := "8080"
	expected := filepath.Join(root, ".stax", "runs", port)
	assert.Equal(t, expected, RunDir(root, port))
}
