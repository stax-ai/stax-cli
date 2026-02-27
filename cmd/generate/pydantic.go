package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stax-ai/stax-cli/internal/automation"
	"github.com/stax-ai/stax-cli/internal/generate"
)

var pydanticOutput string

var pydanticCmd = &cobra.Command{
	Use:   "pydantic",
	Short: "Generate a Pydantic model from config.yml fields",
	Long: `Read config.yml (from cwd/parent dirs or -c/--config), validate it, and write
a Python file containing a Pydantic BaseModel (default class name: AutomationInput)
whose fields match the config's fields. Types are mapped (e.g. integer -> int,
datetime -> datetime). Optional fields become Optional[...] = None.

Output path is relative to the project root unless an absolute path is given.`,
	Example: `  stax generate pydantic
  stax generate pydantic -c ./project/config.yml --output src/models.py`,
	RunE: runPydantic,
}

func init() {
	pydanticCmd.Flags().StringVar(&pydanticOutput, "output", "src/generated_models.py", "output Python file path")
	pydanticCmd.Flags().StringP("config", "c", "", "path to config.yml (default: search config.yml in current and parent dirs)")
}

func runPydantic(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	var cfg *automation.Config
	var projectRoot string
	var err error
	if configPath != "" {
		cfg, err = automation.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		projectRoot = filepath.Dir(configPath)
	} else {
		var fpath string
		cfg, fpath, err = automation.FindAndLoad("")
		if err != nil {
			return fmt.Errorf("find config: %w", err)
		}
		projectRoot = filepath.Dir(fpath)
	}
	if err := automation.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	out := pydanticOutput
	if !filepath.IsAbs(out) {
		out = filepath.Join(projectRoot, out)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := generate.GeneratePydantic(cfg, out); err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	cmd.Println("Wrote", out)
	return nil
}
