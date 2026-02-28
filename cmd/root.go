// Package cmd defines the stax-cli root and all subcommands (config, create, init, generate, run, build).
// Command trees and flags are structured for Cobra doc generation (e.g. GenMarkdownTree).
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	cmdconfig "github.com/stax-ai/stax-cli/cmd/config"
	cmdcreate "github.com/stax-ai/stax-cli/cmd/create"
	cmdgenerate "github.com/stax-ai/stax-cli/cmd/generate"
	cmdinit "github.com/stax-ai/stax-cli/cmd/init"
	"github.com/stax-ai/stax-cli/internal/config"
)

var configFile string

// rootCmd is the entrypoint when running stax-cli without a subcommand.
var rootCmd = &cobra.Command{
	Use:   "stax-cli",
	Short: "CLI for Stax automations (create, init, generate, run, build, deploy)",
	Long: `stax-cli manages Stax automation projects: scaffold Python projects with
config.yml, generate Pydantic models from config, run handlers locally
with CloudEvents, and build container images with buildpacks.

Commands:
  config      Read or write global CLI config (.stax.yml)
  create      Scaffold a new Python automation project (UV, stax-sdk, Ruff, pytest)
  init        Initialize an existing repo with UV, linting, GitHub Actions, config.yml
  legacy-init Initialize deployment config.yaml for legacy cloud run functions
  generate    Generate code from config.yml (e.g. Pydantic models)
  run         Invoke the automation handler once with a CloudEvent (one-shot)
  build       Build a container image from the project using Cloud Native Buildpacks
  deploy      Build and deploy a container image to Kubernetes with Knative and camel-k

Global flags (e.g. --config-file) apply to config get/set. Project commands
(create, init, generate, run, build, deploy) use config.yml in the project directory.`,
}

// SetVersion sets the version string shown by --version (injected at build time by GoReleaser).
func SetVersion(v string) {
	rootCmd.Version = v
}

// RootCmd returns the root command for use in tests and for Cobra doc generation.
func RootCmd() *cobra.Command {
	return rootCmd
}

// Execute runs the root command and exits with code 1 on error.
// Call this from main.main() after all subcommands are registered.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() { _ = config.Init(configFile) })
	rootCmd.PersistentFlags().StringVar(&configFile, "config-file", "", "path to config file (default: search .stax.yml/.stax.yaml in config dir, home, then current dir)")
	rootCmd.AddCommand(cmdconfig.Cmd, cmdcreate.Cmd, cmdinit.Cmd, cmdgenerate.Cmd)
}


