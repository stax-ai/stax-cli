// Package generate provides the "stax generate" command and subcommands (e.g. pydantic).
package generate

import (
	"github.com/spf13/cobra"
)

// Cmd is the "stax generate" parent command. Subcommands emit code from config.yml.
var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code from config.yml",
	Long: `Generate artifacts from the project config.yml. Subcommands:

  pydantic   Write a Pydantic model (AutomationInput) whose fields match config.yml.

Config is resolved from the current directory or parent dirs, or from -c/--config
when running a subcommand.`,
}

func init() {
	Cmd.AddCommand(pydanticCmd)
}
