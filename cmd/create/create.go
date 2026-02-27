// Package create provides the "stax create" command for scaffolding new automation projects.
package create

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stax-ai/stax-cli/internal/scaffold"
)

var (
	pathFlag         string
	automationIDFlag string
)

// Cmd is the "stax create" command. It scaffolds a new Python automation project.
var Cmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Scaffold a new Python automation project",
	Long: `Create a new directory named project-name with a full Python automation layout:

  - config.yml       Stub automation config (automation_id, handler, fields)
  - pyproject.toml   UV project with pydantic, stax-sdk, ruff, pytest
  - main.py          Handler decorated with stax-sdk (main:handler)
  - tests/           Pytest tests
  - .github/workflows/ci.yml   CI (UV, Ruff, pytest)

Use --path to create the project in a specific parent directory; use
--automation-id to set the automation_id in config.yml (default: project name).`,
	Example: `  stax create my-automation
  stax create my-automation --path ./repos --automation-id my-id`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCreate,
}

func init() {
	Cmd.Flags().StringVar(&pathFlag, "path", "", "parent directory to create project in (default: current directory)")
	Cmd.Flags().StringVar(&automationIDFlag, "automation-id", "", "automation_id in config.yml (default: project name)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	}
	if projectName == "" {
		return fmt.Errorf("project name is required: stax create <project-name>")
	}

	targetDir := pathFlag
	if targetDir == "" {
		targetDir = "."
	}
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	opts := scaffold.Options{
		ProjectName:  projectName,
		AutomationID: automationIDFlag,
		TargetDir:    absTarget,
	}
	if err := scaffold.ValidateOptions(&opts); err != nil {
		return err
	}

	ctx := context.Background()
	projectRoot, err := scaffold.Scaffold(ctx, opts)
	if err != nil {
		return err
	}

	cmd.Println("Created project at", projectRoot)
	return nil
}
