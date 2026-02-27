package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stax-ai/stax-cli/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePydanticCommand_WithConfigFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yml")
	err := os.WriteFile(configPath, []byte(`
automation_id: gen-test
handler: main:handler
fields:
  - name: foo
    type: string
    required: true
  - name: n
    type: integer
    required: false
`), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(dir, "models.py")
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"generate", "pydantic", "--config", configPath, "--output", outPath})
	err = rootCmd.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "class AutomationInput(BaseModel):")
	assert.Contains(t, body, "foo: str")
	assert.Contains(t, body, "n: Optional[int] = None")
}
