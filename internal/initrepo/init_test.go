package initrepo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_CreatesLayout(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{
		ProjectRoot:  dir,
		AutomationID: "my-id",
	}

	root, err := Init(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, dir, root)

	expected := []string{
		"config.yml",
		"pyproject.toml",
		".python-version",
		"main.py",
		"Procfile",
		".staxignore",
		"tests/test_handler.py",
		".github/workflows/ci.yml",
	}
	baseName := filepath.Base(dir)
	if baseName == "." {
		baseName = "init_test"
	}
	packageName := baseName
	if packageName == "." {
		packageName = "init_test"
	}
	expected = append(expected, filepath.Join("src", packageName, "__init__.py"))
	for _, name := range expected {
		p := filepath.Join(dir, name)
		_, err := os.Stat(p)
		assert.NoError(t, err, "expected %s to exist", name)
	}
}

func TestInit_ConfigHasEmptyFields(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	opts := Options{
		ProjectRoot:  dir,
		AutomationID: "custom-id",
	}

	_, err := Init(ctx, opts)
	require.NoError(t, err)

	cfgPath := filepath.Join(dir, "config.yml")
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "automation_id: custom-id")
	assert.Contains(t, body, "handler: main:handler")
	assert.Contains(t, body, "fields: []")
}

func TestInit_DoesNotOverwriteExistingConfig(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	existing := []byte("automation_id: existing\nhandler: main:handler\nfields: []\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), existing, 0644))

	opts := Options{ProjectRoot: dir}
	_, err := Init(ctx, opts)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "config.yml"))
	require.NoError(t, err)
	assert.Equal(t, string(existing), string(data), "config should not be overwritten without --force")
}

func TestInit_ForceOverwritesConfig(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte("old"), 0644))

	opts := Options{ProjectRoot: dir, AutomationID: "forced-id", Force: true}
	_, err := Init(ctx, opts)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "config.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "automation_id: forced-id")
	assert.Contains(t, string(data), "fields: []")
}

func TestInit_ProjectRootRequired(t *testing.T) {
	ctx := context.Background()
	_, err := Init(ctx, Options{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project root is required")
}

func TestInitConfigYAML(t *testing.T) {
	out := initConfigYAML(templateData{AutomationID: "test-id"})
	assert.Contains(t, out, "automation_id: test-id")
	assert.Contains(t, out, "handler: main:handler")
	assert.Contains(t, out, "fields: []")
}
