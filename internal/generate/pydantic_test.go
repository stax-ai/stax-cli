package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stax-ai/stax-cli/internal/automation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePydantic_BasicModel(t *testing.T) {
	cfg := &automation.Config{
		AutomationID: "test",
		Handler:      "main:handler",
		Fields: []automation.Field{
			{Name: "name", Type: automation.FieldTypeString, Required: true},
			{Name: "count", Type: automation.FieldTypeInteger, Required: false},
		},
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "models.py")

	err := GeneratePydantic(cfg, out)
	require.NoError(t, err)

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "class AutomationInput(BaseModel):")
	assert.Contains(t, body, "name: str")
	assert.Contains(t, body, "count: Optional[int] = None")
	assert.Contains(t, body, "from pydantic import BaseModel")
}

func TestGeneratePydantic_AllTypes(t *testing.T) {
	cfg := &automation.Config{
		Fields: []automation.Field{
			{Name: "s", Type: automation.FieldTypeString, Required: true},
			{Name: "i", Type: automation.FieldTypeInteger, Required: true},
			{Name: "n", Type: automation.FieldTypeNumber, Required: true},
			{Name: "b", Type: automation.FieldTypeBoolean, Required: true},
			{Name: "a", Type: automation.FieldTypeArray, Required: true},
			{Name: "o", Type: automation.FieldTypeObject, Required: true},
			{Name: "d", Type: automation.FieldTypeDatetime, Required: true},
		},
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "out.py")

	err := GeneratePydantic(cfg, out)
	require.NoError(t, err)

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "from datetime import datetime")
	assert.Contains(t, body, "s: str")
	assert.Contains(t, body, "i: int")
	assert.Contains(t, body, "n: float")
	assert.Contains(t, body, "b: bool")
	assert.Contains(t, body, "a: list[Any]")
	assert.Contains(t, body, "o: dict[str, Any]")
	assert.Contains(t, body, "d: datetime")
}

func TestGeneratePydantic_NilConfig(t *testing.T) {
	dir := t.TempDir()
	err := GeneratePydantic(nil, filepath.Join(dir, "x.py"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestGeneratePydantic_EmptyFields(t *testing.T) {
	cfg := &automation.Config{AutomationID: "id", Handler: "main:h", Fields: nil}
	dir := t.TempDir()
	out := filepath.Join(dir, "empty.py")

	err := GeneratePydantic(cfg, out)
	require.NoError(t, err)

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "class AutomationInput(BaseModel):"))
}
