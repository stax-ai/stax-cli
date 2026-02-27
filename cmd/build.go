package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/stax-ai/stax-cli/internal/automation"
	"github.com/stax-ai/stax-cli/internal/build"
)

var (
	buildPath        string
	buildImage       string
	buildRegistry    string
	buildBuilder     string
	buildTrustBuilder bool
	buildVerbose    bool
)

func init() {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build a container image from the project using buildpacks",
		Long: `Build produces an OCI container image from the current automation project
using Cloud Native Buildpacks (pack). Requires Docker (or a compatible daemon) to be running.

The project directory must contain at least main.py and pyproject.toml (and typically
config.yml). Optionally, .staxignore lists paths to exclude from the build context.
A Procfile defines the web process (e.g. "web: python main.py").`,
		Example: `  stax build --image myreg.io/my-automation:latest
  stax build --path ./my-project --registry myreg.io
  stax build --image localhost:5000/app:v1 --builder ghcr.io/knative/builder-jammy-base:v2 --verbose`,
		RunE: runBuild,
	}
	buildCmd.Flags().StringVar(&buildPath, "path", ".", "project root (default: current directory)")
	buildCmd.Flags().StringVar(&buildImage, "image", "", "full OCI image name to build (e.g. myreg.io/myapp:latest); required unless --registry is set with a project that has config.yml")
	buildCmd.Flags().StringVar(&buildRegistry, "registry", "", "registry for deriving image name when --image is not set (image = registry/automation_id:latest from config.yml)")
	buildCmd.Flags().StringVar(&buildBuilder, "builder", "", "builder image (default: "+build.DefaultBuilderImage+")")
	buildCmd.Flags().BoolVar(&buildTrustBuilder, "trust-builder", false, "trust any builder image (default: only trusted prefixes)")
	buildCmd.Flags().BoolVar(&buildVerbose, "verbose", false, "stream build logs to stderr")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	projectRoot, err := filepath.Abs(buildPath)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}
	if _, err := os.Stat(projectRoot); err != nil {
		return fmt.Errorf("project path %s: %w", projectRoot, err)
	}

	image := buildImage
	if image == "" {
		image, err = deriveImage(projectRoot)
		if err != nil {
			return err
		}
	}

	excludes, err := build.ReadStaxignore(projectRoot)
	if err != nil {
		return fmt.Errorf("read .staxignore: %w", err)
	}

	opts := build.Options{
		ProjectRoot:   projectRoot,
		Image:         image,
		Builder:       buildBuilder,
		Exclude:       excludes,
		TrustBuilder:  buildTrustBuilder,
		Verbose:       buildVerbose,
	}

	if err := build.Build(ctx, opts, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Built: %s\n", image)
	return nil
}

func deriveImage(projectRoot string) (string, error) {
	if buildRegistry == "" {
		return "", fmt.Errorf("either --image or --registry is required; with --registry, the project must have config.yml with automation_id")
	}
	cfg, _, err := automation.FindAndLoad(projectRoot)
	if err != nil {
		return "", fmt.Errorf("could not load config.yml to derive image name: %w", err)
	}
	if cfg.AutomationID == "" {
		return "", fmt.Errorf("config.yml has no automation_id; set --image explicitly")
	}
	registry := strings.TrimSuffix(buildRegistry, "/")
	return registry + "/" + cfg.AutomationID + ":latest", nil
}
