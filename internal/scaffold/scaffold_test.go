package scaffold

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaffold_CreatesLayout(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{
		ProjectName:  "my-automation",
		AutomationID: "my-automation-id",
		TargetDir:    dir,
	}

	projectRoot, err := Scaffold(ctx, opts)
	require.NoError(t, err)
	require.NotEmpty(t, projectRoot)

	// Required files and dirs
	expected := []string{
		"config.yml",
		"pyproject.toml",
		".python-version",
		"README.md",
		"main.py",
		"Procfile",
		".staxignore",
		"tests/test_handler.py",
		".github/workflows/ci.yml",
		"src", // dir
	}
	for _, name := range expected {
		p := filepath.Join(projectRoot, name)
		_, err := os.Stat(p)
		assert.NoError(t, err, "expected %s to exist", name)
	}
}

func TestScaffold_ConfigContent(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{
		ProjectName:  "proj",
		AutomationID: "custom-id",
		TargetDir:    dir,
	}

	projectRoot, err := Scaffold(ctx, opts)
	require.NoError(t, err)

	cfgPath := filepath.Join(projectRoot, "config.yml")
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "automation_id: custom-id")
	assert.Contains(t, body, "handler: main:handler")
	assert.Contains(t, body, "fields:")
}

func TestScaffold_PyprojectContent(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{
		ProjectName: "hello-world",
		TargetDir:   dir,
	}

	projectRoot, err := Scaffold(ctx, opts)
	require.NoError(t, err)

	pyPath := filepath.Join(projectRoot, "pyproject.toml")
	data, err := os.ReadFile(pyPath)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "stax-sdk")
	assert.Contains(t, body, "pydantic")
	assert.Contains(t, body, "ruff")
	assert.Contains(t, body, "pytest")
}

func TestScaffold_MainPyHasStaxHandler(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{ProjectName: "p", TargetDir: dir}

	projectRoot, err := Scaffold(ctx, opts)
	require.NoError(t, err)

	mainPath := filepath.Join(projectRoot, "main.py")
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "stax_sdk")
	assert.Contains(t, body, "def handler(")
	assert.Contains(t, body, "@stax_handler")
}

func TestScaffold_DefaultAutomationID(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{
		ProjectName: "default-id",
		TargetDir:   dir,
		// AutomationID left empty -> should default to project name
	}

	projectRoot, err := Scaffold(ctx, opts)
	require.NoError(t, err)

	cfgPath := filepath.Join(projectRoot, "config.yml")
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "automation_id: default-id")
}

func TestScaffold_RequiresProjectName(t *testing.T) {
	ctx := context.Background()
	_, err := Scaffold(ctx, Options{TargetDir: t.TempDir()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project name")
}

func TestValidateOptions(t *testing.T) {
	err := ValidateOptions(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")

	err = ValidateOptions(&Options{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project name")

	err = ValidateOptions(&Options{ProjectName: "ok"})
	assert.NoError(t, err)
}

func TestScaffold_PythonVersion(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{ProjectName: "p", TargetDir: dir}

	projectRoot, err := Scaffold(ctx, opts)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(projectRoot, ".python-version"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "3."))
}
