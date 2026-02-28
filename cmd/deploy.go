package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/stax-ai/stax-cli/internal/deploy"
)

var (
	deployConfigFile string
	deployPath       string
	deployNamespace  string
	deployImage      string
	deployRegistry   string
	deployBuilder    string
	deployProduct    string
	deployEnvironment string
	deploySkipBuild  bool
	deploySkipPush   bool
	deployDryRun     bool
)

func init() {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy a container image to Kubernetes with Knative and camel-k",
		Long: `Deploy builds a container image using buildpacks, pushes it to a registry, and
deploys a Knative Service configured for single-concurrency processing along with
camel-k resources (Pipe) to connect pub/sub to the service via a Broker.

The deployment configuration is read from config.yaml (separate from project config.yml).
The command generates Kubernetes manifests and saves them to the deploy/ folder before applying.`,
		Example: `  stax deploy --config config.yaml
  stax deploy --config config.yaml --skip-build
  stax deploy --config config.yaml --dry-run`,
		RunE: runDeploy,
	}
	deployCmd.Flags().StringVar(&deployConfigFile, "config", "", "path to deployment config.yaml (default: search for config.yaml in current or parent directories)")
	deployCmd.Flags().StringVar(&deployPath, "path", ".", "project root (default: current directory)")
	deployCmd.Flags().StringVar(&deployNamespace, "namespace", "", "k8s namespace (overrides config.yaml and product-based default)")
	deployCmd.Flags().StringVar(&deployImage, "image", "", "override image name (overrides config.yaml)")
	deployCmd.Flags().StringVar(&deployRegistry, "registry", "", "container registry URL (overrides config.yaml)")
	deployCmd.Flags().StringVar(&deployBuilder, "builder", "gcr.io/buildpacks/builder", "buildpack builder image (default: gcr.io/buildpacks/builder)")
	deployCmd.Flags().StringVar(&deployProduct, "product", "", "product name: ta or cx (overrides config.yaml)")
	deployCmd.Flags().StringVar(&deployEnvironment, "environment", "", "environment name: prd or stg (overrides config.yaml)")
	deployCmd.Flags().BoolVar(&deploySkipBuild, "skip-build", false, "skip build step if image already exists")
	deployCmd.Flags().BoolVar(&deploySkipPush, "skip-push", false, "skip push step")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "generate YAML without applying (still saves to deploy/ folder)")
	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	projectRoot, err := filepath.Abs(deployPath)
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}
	if _, err := os.Stat(projectRoot); err != nil {
		return fmt.Errorf("project path %s: %w", projectRoot, err)
	}

	// Load deployment config
	var cfg *deploy.Config
	var configPath string
	if deployConfigFile != "" {
		cfg, err = deploy.LoadConfig(deployConfigFile)
		if err != nil {
			return fmt.Errorf("load deployment config: %w", err)
		}
		configPath = deployConfigFile
	} else {
		cfg, configPath, err = deploy.FindAndLoadConfig(projectRoot)
		if err != nil {
			return fmt.Errorf("find deployment config: %w", err)
		}
	}

	// Override config with flags if provided
	if deployProduct != "" {
		cfg.Product = deployProduct
	}
	if deployEnvironment != "" {
		cfg.Environment = deployEnvironment
	}
	if deployNamespace != "" {
		cfg.Namespace = deployNamespace
	}
	if deployImage != "" {
		cfg.ImageName = deployImage
	}
	if deployRegistry != "" {
		cfg.ImageRegistry = deployRegistry
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid deployment config: %w", err)
	}

	// Build deployment options
	opts := deploy.Options{
		ProjectRoot:  projectRoot,
		Config:       cfg,
		ConfigPath:   configPath,
		Builder:      deployBuilder,
		SkipBuild:    deploySkipBuild,
		SkipPush:     deploySkipPush,
		DryRun:       deployDryRun,
	}

	// Execute deployment
	if err := deploy.Deploy(ctx, opts); err != nil {
		return err
	}

	if deployDryRun {
		fmt.Fprintf(os.Stderr, "Dry run complete. Generated manifests saved to deploy/ folder.\n")
	} else {
		fmt.Fprintf(os.Stderr, "Deployment complete. Manifests saved to deploy/ folder.\n")
	}
	return nil
}
