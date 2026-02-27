// Package init provides the "stax init" command for initializing an existing repo with Stax layout.
package init

import (
	"context"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stax-ai/stax-cli/internal/initrepo"
)

var (
	initPathFlag         string
	initAutomationIDFlag string
	initForceFlag        bool
)

// Cmd is the "stax init" command. It initializes the current (or --path) directory with Stax tooling.
var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an existing repo with UV, linting, GitHub Actions, and config.yml",
	Long: `Initialize the current directory (or --path) with Stax automation layout:

  - config.yml       Minimal automation config (automation_id, handler, empty fields)
  - pyproject.toml   UV project with stax-sdk, pydantic, ruff, pytest
  - main.py          Thin entrypoint importing handler from package (if missing)
  - src/<package>/   Package with stub handler (if missing)
  - .github/workflows/ci.yml   CI (UV, Ruff, pytest)
  - .python-version, .staxignore, Procfile

Does not overwrite existing files unless --force (applies to config.yml and CI only).`,
	Example: `  stax init
  stax init --path ./my-repo --automation-id my-id
  stax init --force`,
	Args: cobra.NoArgs,
	RunE: runInit,
}

func init() {
	Cmd.Flags().StringVar(&initPathFlag, "path", "", "directory to initialize (default: current directory)")
	Cmd.Flags().StringVar(&initAutomationIDFlag, "automation-id", "", "automation_id in config.yml (default: directory name)")
	Cmd.Flags().BoolVar(&initForceFlag, "force", false, "overwrite existing config.yml and .github/workflows/ci.yml")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectRoot := initPathFlag
	if projectRoot == "" {
		projectRoot = "."
	}
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return err
	}

	opts := initrepo.Options{
		ProjectRoot:  absRoot,
		AutomationID: initAutomationIDFlag,
		Force:        initForceFlag,
	}

	ctx := context.Background()
	root, err := initrepo.Init(ctx, opts)
	if err != nil {
		return err
	}

	cmd.Println("Initialized project at", root)
	return nil
}
