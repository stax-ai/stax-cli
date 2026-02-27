package automation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	err := os.WriteFile(path, []byte(`
automation_id: my-automation
handler: main:handler
fields:
  - name: foo
    type: string
    required: true
  - name: count
    type: integer
    required: false
`), 0644)
	require.NoError(t, err)

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "my-automation", cfg.AutomationID)
	assert.Equal(t, "main:handler", cfg.Handler)
	require.Len(t, cfg.Fields, 2)
	assert.Equal(t, "foo", cfg.Fields[0].Name)
	assert.Equal(t, FieldTypeString, cfg.Fields[0].Type)
	assert.True(t, cfg.Fields[0].Required)
	assert.Equal(t, "count", cfg.Fields[1].Name)
	assert.Equal(t, FieldTypeInteger, cfg.Fields[1].Type)
	assert.False(t, cfg.Fields[1].Required)
}

func TestLoad_InvalidPath(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nonexistent.yml"))
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	err := os.WriteFile(path, []byte("invalid: [[[ yaml"), 0644)
	require.NoError(t, err)

	_, err = Load(path)
	require.Error(t, err)
}

func TestValidate_Valid(t *testing.T) {
	cfg := &Config{
		AutomationID: "id",
		Handler:      "main:handler",
		Fields: []Field{
			{Name: "x", Type: FieldTypeString, Required: true},
		},
	}
	err := Validate(cfg)
	assert.NoError(t, err)
}

func TestValidate_Nil(t *testing.T) {
	err := Validate(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestValidate_MissingAutomationID(t *testing.T) {
	cfg := &Config{Handler: "main:handler"}
	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "automation_id")
}

func TestValidate_MissingHandler(t *testing.T) {
	cfg := &Config{AutomationID: "id"}
	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler")
}

func TestValidate_InvalidHandlerFormat(t *testing.T) {
	cfg := &Config{AutomationID: "id", Handler: "nocolon"}
	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "module:function")
}

func TestValidate_InvalidFieldType(t *testing.T) {
	cfg := &Config{
		AutomationID: "id",
		Handler:      "main:handler",
		Fields:       []Field{{Name: "x", Type: "invalid_type", Required: true}},
	}
	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type")
}

func TestValidate_FieldMissingName(t *testing.T) {
	cfg := &Config{
		AutomationID: "id",
		Handler:      "main:handler",
		Fields:       []Field{{Name: "", Type: FieldTypeString, Required: true}},
	}
	err := Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidate_AllFieldTypes(t *testing.T) {
	for _, ft := range ValidFieldTypes {
		cfg := &Config{
			AutomationID: "id",
			Handler:      "main:run",
			Fields:       []Field{{Name: "f", Type: ft, Required: false}},
		}
		err := Validate(cfg)
		assert.NoError(t, err, "field type %q should be valid", ft)
	}
}

func TestFindAndLoad_FoundInDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	err := os.WriteFile(path, []byte(`
automation_id: found
handler: main:handler
fields: []
`), 0644)
	require.NoError(t, err)

	cfg, fpath, err := FindAndLoad(dir)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "found", cfg.AutomationID)
	assert.Equal(t, path, fpath)
}

func TestFindAndLoad_FoundConfigYaml(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`
automation_id: alt-ext
handler: main:handler
fields: []
`), 0644)
	require.NoError(t, err)

	cfg, fpath, err := FindAndLoad(dir)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "alt-ext", cfg.AutomationID)
	assert.Equal(t, path, fpath)
}

func TestFindAndLoad_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, _, err := FindAndLoad(dir)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

